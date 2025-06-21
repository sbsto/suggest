package llm

import (
	"context"
	"fmt"
	"os"
)

const COMMAND_PROMPT = "Given this description: '%s', suggest a single CLI command that would accomplish this task.\n\nYou MUST respond with ONLY valid JSON in this EXACT format:\n{\"command\": \"the raw command\", \"description\": \"brief 1-2 line explanation of what this command does\"}\n\nIMPORTANT REQUIREMENTS:\n- Return ONLY the JSON object, no other text\n- Do not wrap in markdown code blocks\n- Do not include backticks, explanations, or any other formatting\n- The command field must contain the exact command that can be executed\n- The description field must be 1-2 lines maximum\n- Ensure the JSON is properly formatted and valid"

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
