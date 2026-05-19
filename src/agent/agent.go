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

// Agent is the main struct that holds the state of the agent
// including the conversation history, available tools, and configuration.
type Agent struct {
	messages []openai.ChatCompletionMessage
	tools    []tool.Tool
	toolDefs []openai.Tool
	config   *Config
	client   *openai.Client
}

// NewAgent creates a new instance of the Agent struct,
// initializes the LLM client and registers the available tools.
func NewAgent(config *Config) *Agent {
	a := &Agent{
		config:   config,
		messages: make([]openai.ChatCompletionMessage, 1),
	}

	// setup llm client
	a.setupLLMClient()

	// register tools
	a.registerTools()

	// default system message
	a.SetSystemMessage("You are a helpful assistant.")
	return a
}

// setupLLMClient initializes the OpenAI client with the
// provided API key and base URL from the configuration.
func (a *Agent) setupLLMClient() {
	llmConf := openai.DefaultConfig(a.config.LLMApiKey)
	llmConf.BaseURL = a.config.LLMApiBaseUrl
	a.client = openai.NewClientWithConfig(llmConf)
}

// registerTools initializes the list of tools that the agent can use,
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

// SetSystemMessage sets the system message for the agent,
// which provides context and instructions to the LLM.
// it may be used provide a planning prompt or to set the behavior of the agent.
func (a *Agent) SetSystemMessage(message string) {
	system := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: message,
	}
	a.messages[0] = system
}

// addUserMessage adds a user message to the conversation history.
func (a *Agent) addUserMessage(prompt string) {
	// add user message to chat history
	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}
	a.messages = append(a.messages, message)
}

// Loop starts the main interaction loop of the agent
// where it continuously prompts the user for input,
func (a *Agent) Loop() error {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Hi, How can I help you today?:")
	for {
		fmt.Printf("\n\n>> ")

		// read line
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return util.DetailedError("Unable to read input", err)
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
		channelToStdOut(out)

		if err := a.GenerateResponse(prompt, out, true); err != nil {
			slog.Error("failed to generate response", "error", err)
		}
	}
}

// GenerateResponse takes a user prompt and generates a response using the LLM,
// it handles tool calls and updates the conversation history accordingly.
// if streamIntermediateMessages is true, it will stream intermediate messages
// back to the caller as they are generated, otherwise it will wait
// until the final response is generated before returning.
func (a *Agent) GenerateResponse(prompt string, out chan<- string, streamIntermediateMessages bool) error {
	defer func() {
		// close channel when retuning
		// from this functon
		close(out)
	}()

	// add user message to chat history
	a.addUserMessage(prompt)

	// agent loop
	for {
		// get next message
		var resp *openai.ChatCompletionResponse
		var err error

		if streamIntermediateMessages {
			// stream next message and stream intermediate messages back to caller
			resp, err = a.streamNextMessage(out)
			if err != nil {
				return util.DetailedError("Unable stream next message", err)
			}
		} else {
			// if not streaming intermediate messages, just get the next message
			resp, err = a.getNextMessage()
		}

		if err != nil {
			return util.DetailedError("Failed whilst receiving chat completion", err)
		}

		// if there are no choices, break out of loop
		if len(resp.Choices) == 0 {
			// break out of loop if there are no tool calls
			break
		}

		// resp choices
		for _, choice := range resp.Choices {

			// stop if there are no tool calls or finish reason is stop
			if len(choice.Message.ToolCalls) == 0 || choice.FinishReason == "stop" {
				if !streamIntermediateMessages {
					fmt.Println(choice.Message.Content)
				}
				return nil
			}

			// record response message in conversation history
			a.messages = append(a.messages, choice.Message)

			// handle tool calls
			for _, call := range choice.Message.ToolCalls {
				err = a.handleToolCall(context.Background(), call, &a.messages)
				if err != nil {
					slog.Error("failed to handle tool call", "error", err, "tool_call", call)
				}
			}
		}
	}
	return nil
}

// handleToolCall takes a tool call from the LLM and executes the corresponding tool,
func (a *Agent) handleToolCall(ctx context.Context, toolCall openai.ToolCall, messages *[]openai.ChatCompletionMessage) error {
	toolName := toolCall.Function.Name
	for _, agentTool := range a.tools {
		if agentTool.Definition().Function.Name == toolName {
			slog.Debug("tool call", "tool", toolName, "arguments", toolCall.Function.Arguments)
			err := agentTool.Call(ctx, &toolCall, messages)
			if err != nil {
				return err
			}

			//end  of tool call exit
			return nil
		}
	}
	return nil
}

// getNextMessage gets the next message from the LLM based on the current
// conversation history and tool definitions.
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
		return nil, util.DetailedError("Failed whilst receiving chat completion", err)
	}
	return &resp, nil
}

// streamNextMessage streams the next message from the LLM, sending intermediate
// messages back to the caller as they are generated.
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
		// wait to recieve chunk
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		err = mergeStreamResponse(finalResponse, &chunk)
		if err != nil {
			return nil, util.DetailedError("Unable to mergeStreamResponse chunk response to chat completion", nil)
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

// DirectResponse sends a prompt to the agent and prints the response to stdout
func (a *Agent) DirectResponse(ctx context.Context, prompt string) (string, error) {
	// make channel for output
	out := make(chan string)

	channelToStdOut(out)

	if err := a.GenerateResponse(prompt, out, false); err != nil {
		return "", util.DetailedError("Unable to generate response", err)
	}

	outStr := <-out

	return outStr, nil
}

// ChannelToStdOut writes the contents of a channel to stdout.
// - out is the channel to read from. It is expected that the channel will be closed when done,
// and this function will return at that point.
func channelToStdOut(out <-chan string) {
	// write output streamNextMessage
	go func() {
		for s := range out {
			_, err := fmt.Fprint(os.Stdout, s)
			if err != nil {
				slog.Error("Unable to write to stdout", "err", err)
			}
		}
	}()
}

// mergeStreamResponse merges a chunk response from the stream into the final chat completion response,
// this is necessary for tools as they can be streamed back in chunks and we need to merge them together
// to get the full tool call information.
func mergeStreamResponse(finalResponse *openai.ChatCompletionResponse, chunk *openai.ChatCompletionStreamResponse) error {
	finalResponse.Model = chunk.Model
	finalResponse.ID = chunk.ID
	finalResponse.Object = chunk.Object
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
