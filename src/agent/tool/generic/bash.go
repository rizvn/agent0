package generic

import (
	"agent0/agent/tool"
	"agent0/util"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Bash struct {
	def openai.Tool
}

// Ensure Bash implements the Tool interface
var _ tool.Tool = (*Bash)(nil)

func NewBash() *Bash {
	b := &Bash{}
	b.def = openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "Bash",
			Description: "Execute a shell command",
			Parameters: map[string]any{
				"type":     "object",
				"required": []string{"command"},
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "The command to execute",
					},
				},
			},
		},
	}
	return b
}

func (b *Bash) Definition() openai.Tool {
	return b.def
}

func (b *Bash) Call(
	ctx context.Context,
	toolCall *openai.ToolCall,
	messages *[]openai.ChatCompletionMessage) error {

	args := make(map[string]any)
	err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
	if err != nil {
		return util.DetailedError("unable to parse args", err)
	}
	command := args["command"].(string)
	output, err := bash(command)
	if err != nil {
		output = fmt.Sprintf("error runnning bash command: %s\nerror:%v", command, err)
	}

	// add response to chat history
	toolResponse := openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    output,
		ToolCallID: toolCall.ID,
	}

	*messages = append(*messages, toolResponse)
	return nil
}

func bash(cmdString string) (string, error) {
	cmd := exec.CommandContext(context.TODO(), "bash", "-lc", cmdString)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		if stderr.Len() != 0 {
			return "", util.DetailedError(fmt.Sprintf("Error running bash command, stderr:%s", stderr.String()), err)
		}
	}

	output := strings.TrimSpace(stdout.String())

	if len(output) == 0 {
		output = "command completed"
	}
	return output, nil
}
