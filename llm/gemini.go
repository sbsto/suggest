package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
}

func NewGeminiProvider(apiKey string, ctx context.Context) *GeminiProvider {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil
	}

	return &GeminiProvider{
		client: client,
	}
}

func (p *GeminiProvider) GenerateCommand(description string, ctx context.Context) (string, error) {
	model := p.client.GenerativeModel("gemini-2.5-flash")
	osName := getSystemInfo()
	prompt := fmt.Sprintf(COMMAND_PROMPT, description, osName)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no suggestions received")
	}

	return strings.TrimSpace(fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])), nil
}

func (p *GeminiProvider) GenerateCommandWithContext(description, errorContext string, ctx context.Context) (string, error) {
	model := p.client.GenerativeModel("gemini-2.5-flash")
	osName := getSystemInfo()
	prompt := fmt.Sprintf(COMMAND_WITH_ERROR_PROMPT, description, osName, errorContext)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no suggestions received")
	}

	return strings.TrimSpace(fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])), nil
}
