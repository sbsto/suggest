package llm

import "context"

type LLMProvider interface {
	GenerateCommand(description string, ctx context.Context) (string, error)
	GenerateCommandWithContext(description, errorContext string, ctx context.Context) (string, error)
}
