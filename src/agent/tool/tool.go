package tool

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// Tool is an interface that defines the structure of a tool that can be registered with the agent.
// Each tool must implement the Definition method, which returns the tool definition that will be passed
// to the OpenAI API when registering the tool, and the Call method, which will be called when the tool
// is invoked by the agent.
type Tool interface {
	Definition() openai.Tool
	Call(context context.Context, toolCall *openai.ToolCall, messages *[]openai.ChatCompletionMessage) error
}
