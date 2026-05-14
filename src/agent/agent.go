package agent

import (
	"agent0/agent/tool"
	"agent0/agent/tool/generic"
	"agent0/util"
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Agent struct {
	messages   []openai.ChatCompletionMessage
	tools      []tool.Tool
	toolDefs   []openai.Tool
	config     *Config
	client     *openai.Client
	tenantId   string
	userId     string
	dataCtxKey DataCtxKey
}

type DataCtxKey struct{}

func NewAgent(config *Config) *Agent {
	a := &Agent{
		config:     config,
		dataCtxKey: DataCtxKey{},
		messages:   make([]openai.ChatCompletionMessage, 1),
	}

	a.setupLLMClient()
	a.registerTools()

	// default system message
	a.SetSystemMessage("You are a helpful assistant.")
	return a
}

func (a *Agent) setupLLMClient() {
	llmConf := openai.DefaultConfig(a.config.LLMApiKey)
	llmConf.BaseURL = a.config.LLMApiBaseUrl
	a.client = openai.NewClientWithConfig(llmConf)
}

func (a *Agent) registerTools() {
	a.tools = []tool.Tool{
		generic.NewReadFile(),
		generic.NewBash(),
		generic.NewWriteFile(),
	}

	// generate tool defs from to pass to openai api
	var toolDefs []openai.Tool
	for _, t := range a.tools {
		toolDefs = append(toolDefs, t.Definition())
	}
	a.toolDefs = toolDefs
}

func (a *Agent) SetSystemMessage(message string) {
	system := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: message,
	}
	a.messages[0] = system
}

func (a *Agent) addUserMessage(prompt string) {
	// add user message to chat history
	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}
	a.messages = append(a.messages, message)
}

func (a *Agent) generateContext() context.Context {
	data := map[string]any{
		"tenant_id": a.tenantId,
		"user_id":   a.userId,
	}
	ctx := context.WithValue(context.Background(), a.dataCtxKey, data)
	return ctx
}

func (a *Agent) Loop() error {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Hi, How can I help you today?:")
	for {
		fmt.Printf("\n>> ")

		// read line
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return util.NewErr("Unable to read input", err)
			}
		}

		prompt := strings.TrimSpace(scanner.Text())

		//handle empty prompt
		if prompt == "" {
			continue
		}

		// exit loop
		if prompt == "exit" || prompt == "quit" {
			return nil
		}

		out := make(chan string)
		util.ChannelToStdOut(out)

		if err := a.GenerateResponse(prompt, out, true); err != nil {
			slog.Error("failed to generate response", "error", err)
		}
	}
}

func (a *Agent) GenerateResponse(prompt string, out chan<- string, streamIntermediateMessages bool) error {
	defer func() {
		// close channel when retuning
		// from this functon
		close(out)
	}()

	a.addUserMessage(prompt)
	ctx := a.generateContext()

	// agent loop
	for {
		// get next message
		var resp *openai.ChatCompletionResponse
		var err error

		if streamIntermediateMessages {
			resp, err = a.streamNextMessage(out)
			if err != nil {
				return util.NewErr("Unable stream next message", err)
			}
		} else {
			resp, err = a.getNextMessage()
		}

		if err != nil {
			return util.NewErr("Failed whilst receiving chat completion", err)
		}

		if len(resp.Choices) == 0 {
			// break out of loop if there are no tool calls
			break
		}

		// resp choices
		for _, choice := range resp.Choices {

			if len(choice.Message.ToolCalls) == 0 || choice.FinishReason == "stop" {
				//note: finish reason can also be tool_calls
				// no tool calls, continue to next message
				if !streamIntermediateMessages {
					fmt.Println(choice.Message.Content)
				}
				return nil
			}

			// record response
			a.messages = append(a.messages, choice.Message)

			for _, call := range choice.Message.ToolCalls {
				toolName := call.Function.Name
				for _, agentTool := range a.tools {
					if agentTool.Definition().Function.Name == toolName {
						slog.Debug("tool call", "tool", toolName, "arguments", call.Function.Arguments)
						slog.Debug("using", "tool", toolName)

						err := agentTool.Call(ctx, &call, &a.messages)
						if err != nil {
							return err
						}

						//end  of tool call exit
						break
					}
				}
			}
		}
	}
	return nil
}

func (a *Agent) getNextMessage() (*openai.ChatCompletionResponse, error) {
	// get next message
	resp, err := a.client.CreateChatCompletion(context.Background(),
		openai.ChatCompletionRequest{
			Model:    a.config.LLMModel,
			Messages: a.messages,
			Tools:    a.toolDefs,
		},
	)
	if err != nil {
		return nil, util.NewErr("Failed whilst receiving chat completion", err)
	}
	return &resp, nil
}

func (a *Agent) streamNextMessage(out chan<- string) (*openai.ChatCompletionResponse, error) {
	stream, err := a.client.CreateChatCompletionStream(context.Background(),
		openai.ChatCompletionRequest{
			Model:    a.config.LLMModel,
			Messages: a.messages,
			Tools:    a.toolDefs,
			Stream:   true,
		},
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = stream.Close()
	}()

	// make final response to append chunks too
	finalResponse := &openai.ChatCompletionResponse{}
	finalResponse.Choices = make([]openai.ChatCompletionChoice, 1)

	//var assistantContent strings.Builder
	for {
		// wait to recieve chuck
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		err = mergeStreamResponse(finalResponse, &chunk)
		if err != nil {
			return nil, util.NewErr("Unable to mergeStreamResponse chunk response to chat completion", nil)
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta
		// if chunk contains textual llm response
		if delta.Content != "" {
			//echo back the textual response to caller
			out <- delta.Content
		}
	}
	return finalResponse, nil
}

func mergeStreamResponse(finalResponse *openai.ChatCompletionResponse, chunk *openai.ChatCompletionStreamResponse) error {
	finalResponse.Model = chunk.Model
	finalResponse.ID = chunk.ID
	finalResponse.Object = chunk.Model
	finalResponse.Created = chunk.Created
	finalResponse.Model = chunk.Model

	finalChoice := &finalResponse.Choices[0]
	deltaChoice := &chunk.Choices[0]

	finalChoice.FinishReason = deltaChoice.FinishReason

	finalMessage := &finalChoice.Message
	deltaMessage := &deltaChoice.Delta

	finalMessage.Role += deltaMessage.Role
	finalMessage.Content += deltaMessage.Content

	if len(deltaMessage.ToolCalls) > 0 {
		deltaToolCall := &deltaMessage.ToolCalls[0]

		if finalMessage.ToolCalls == nil {
			finalMessage.ToolCalls = make([]openai.ToolCall, 1)
		}

		if len(finalMessage.ToolCalls) < *deltaToolCall.Index+1 {
			finalMessage.ToolCalls = append(finalMessage.ToolCalls, openai.ToolCall{})
		}

		finalToolCall := &finalMessage.ToolCalls[*deltaToolCall.Index]

		finalToolCall.Index = deltaToolCall.Index

		if deltaToolCall.ID != "" {
			finalToolCall.ID = deltaToolCall.ID
		}

		if deltaToolCall.Type != "" {
			finalToolCall.Type = deltaToolCall.Type
		}

		finalFunction := &finalToolCall.Function
		deltaFunction := &deltaToolCall.Function

		if deltaFunction.Name != "" {
			finalFunction.Name = deltaFunction.Name
		}

		if deltaFunction.Arguments != "" {
			finalFunction.Arguments += deltaFunction.Arguments
		}
	}

	return nil
}
