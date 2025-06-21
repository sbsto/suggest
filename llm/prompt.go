package llm

import (
	"context"
	"fmt"
	"os"
)

const COMMAND_PROMPT = "Given this description: '%s', suggest a single CLI command that would accomplish this task. Return only the raw command that can be executed directly - no backticks, no code blocks, no markdown formatting, no quotes, no explanations. Just the plain command."

func getProvider(ctx context.Context) LLMProvider {
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		provider := NewGeminiProvider(key, ctx)
		if provider != nil {
			return provider
		}
	}

	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		provider := NewOpenaiProvider(key)
		if provider != nil {
			return provider
		}
	}

	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		provider := NewAnthropicProvider(key)
		if provider != nil {
			return provider
		}
	}

	return nil
}

func GenerateCommand(description string, ctx context.Context) (string, error) {
	provider := getProvider(ctx)
	if provider == nil {
		return "", fmt.Errorf("no API key found. Set OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY")
	}
	return provider.GenerateCommand(description, ctx)
}
