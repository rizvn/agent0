package generic

import (
	"agent0/agent/tool"
	"agent0/util"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

type WriteFile struct {
	def openai.Tool
}

// ensure WriteFile implements the Tool interface
var _ tool.Tool = (*WriteFile)(nil)

func NewWriteFile() *WriteFile {
	w := &WriteFile{}
	w.def = openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "Write",
			Description: "Write content to a file",
			Parameters: map[string]any{
				"type":     "object",
				"required": []string{"file_path", "content"},
				"properties": map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "The path of the file to write to",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "The content to write to the file",
					},
				},
			},
		},
	}

	return w
}

func (w *WriteFile) Definition() openai.Tool {
	return w.def
}

func (w *WriteFile) Call(ctx context.Context, toolCall *openai.ToolCall, messages *[]openai.ChatCompletionMessage) error {
	type Args struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}

	args := Args{}
	err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
	if err != nil {
		return util.DetailedError("Unable to parse function arguments", nil)
	}

	err = writeFile(args.FilePath, args.Content)
	if err != nil {
		return util.DetailedError(fmt.Sprintf("unable to write file %s", args.FilePath), err)
	}

	toolResponse := openai.ChatCompletionMessage{
		Role:       openai.ChatMessageRoleTool,
		Content:    "File written",
		ToolCallID: toolCall.ID,
	}

	*messages = append(*messages, toolResponse)
	return nil
}

func writeFile(path, content string) error {
	err := os.WriteFile(path, []byte(content), 0777)
	if err != nil {
		return fmt.Errorf("unable to write file %s", path)
	}
	return nil
}
