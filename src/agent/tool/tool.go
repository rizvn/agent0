package tool

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// Tool is an interface that defines the structure of a tool
type Tool interface {

	// Definition returns the tool definition of the tool
	Definition() openai.Tool

	// Call is used to call the tool by the agent
	Call(context context.Context,
		toolCall *openai.ToolCall,
		messages *[]openai.ChatCompletionMessage) error
}
