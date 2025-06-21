package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenaiProvider struct {
	client *openai.Client
}

func NewOpenaiProvider(apiKey string) *OpenaiProvider {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &OpenaiProvider{
		client: &client,
	}
}

func (p *OpenaiProvider) GenerateCommand(description string, ctx context.Context) (string, error) {
	prompt := fmt.Sprintf(COMMAND_PROMPT, description)

	chat, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: "gpt-4.1-mini-2025-04-14",
	})

	if err != nil {
		return "", err
	}

	if len(chat.Choices) == 0 {
		return "", fmt.Errorf("no suggestions received")
	}

	return strings.TrimSpace(chat.Choices[0].Message.Content), nil
}

func (p *OpenaiProvider) GenerateCommandWithContext(description, errorContext string, ctx context.Context) (string, error) {
	prompt := fmt.Sprintf(COMMAND_WITH_ERROR_PROMPT, description, errorContext)

	chat, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: "gpt-4.1-mini-2025-04-14",
	})

	if err != nil {
		return "", err
	}

	if len(chat.Choices) == 0 {
		return "", fmt.Errorf("no suggestions received")
	}

	return strings.TrimSpace(chat.Choices[0].Message.Content), nil
}
