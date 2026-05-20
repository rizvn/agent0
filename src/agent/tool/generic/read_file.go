package generic

import (
	"agent0/agent/tool"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

type ReadFile struct {
	def openai.Tool
}

// ensure ReadFile implements the Tool interface
var _ tool.Tool = (*ReadFile)(nil)

// NewReadFile Constructor for the ReadFile tool
func NewReadFile() *ReadFile {
	r := &ReadFile{}

	// define tool definition
	r.def = openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "Read",
			Description: "Read and return the content of a file",
			Parameters: map[string]any{
				"type":        "object",
				"description": "Read and return the contents of a file",
				"properties": map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "The path to the file to read",
					},
				},
				"required": []string{"file_path"},
			},
		},
	}
	return r
}

// Definition returns the tool definition
func (r *ReadFile) Definition() openai.Tool {
	return r.def
}

// Call is called when the tool is invoked by the agent
// It receives the context, the tool call information,
// and a pointer to the messages array.
func (r *ReadFile) Call(
	ctx context.Context,
	toolCall *openai.ToolCall,
	messages *[]openai.ChatCompletionMessage) error {
	// define args struct
	type Args struct {
		FilePath string `json:"file_path"`
	}

	// create args instance
	args := Args{}

	// unmarshall args from string
	err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
	if err != nil {
		return fmt.Errorf("unable to parse function arguments: %w", err)
	}

	content, err := readFile(args.FilePath)
	if err != nil {
		return fmt.Errorf("unable to read file, %w", err)
	}

	toolResponse := openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    content,
		ToolCallID: toolCall.ID,
	}

	*messages = append(*messages, toolResponse)
	return nil
}

// readFile is a helper function that reads the content
// of a file given its path and returns it as a string.
func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("unable to read file, %w", err)
	}
	return string(content), nil
}
