package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicProvider struct {
	client *anthropic.Client
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &AnthropicProvider{
		client: &client,
	}
}

func (p *AnthropicProvider) GenerateCommand(description string, ctx context.Context) (string, error) {
	prompt := buildPrompt(description)

	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-3-5-haiku-latest",
		MaxTokens: 100,
		Messages: []anthropic.MessageParam{{
			Role: anthropic.MessageParamRoleUser,
			Content: []anthropic.ContentBlockParamUnion{
				anthropic.NewTextBlock(prompt),
			},
		}},
	})

	if err != nil {
		return "", err
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("no suggestions received")
	}

	for _, block := range message.Content {
		if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
			return strings.TrimSpace(textBlock.Text), nil
		}
	}

	return "", fmt.Errorf("no text content found")
}

func (p *AnthropicProvider) GenerateCommandWithContext(description, errorContext string, ctx context.Context) (string, error) {
	prompt := buildPromptWithContext(description, errorContext)

	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-3-5-haiku-latest",
		MaxTokens: 100,
		Messages: []anthropic.MessageParam{{
			Role: anthropic.MessageParamRoleUser,
			Content: []anthropic.ContentBlockParamUnion{
				anthropic.NewTextBlock(prompt),
			},
		}},
	})

	if err != nil {
		return "", err
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("no suggestions received")
	}

	for _, block := range message.Content {
		if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
			return strings.TrimSpace(textBlock.Text), nil
		}
	}

	return "", fmt.Errorf("no text content found")
}
