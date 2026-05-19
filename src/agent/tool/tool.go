package tool

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// Tool is an interface that defines the structure of a tool
type Tool interface {

	// Definition returns the tool definition that will be passed to the OpenAI API when registering the tool.
	Definition() openai.Tool

	// Call is called when the tool is invoked by the agent. It receives the context, the tool call information,
	Call(context context.Context, toolCall *openai.ToolCall, messages *[]openai.ChatCompletionMessage) error
}
