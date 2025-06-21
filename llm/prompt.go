package llm

import (
	"context"
	"fmt"
	"os"
	"runtime"
)

const BASE_PROMPT = `Given this description: '%s', suggest a single CLI command that would accomplish this task.

SYSTEM CONTEXT:
- Operating System: %s
%s
You MUST respond with ONLY valid JSON in this EXACT format:
{"command": "the raw command", "description": "brief 1-2 line explanation of what this command does"}

IMPORTANT REQUIREMENTS:
- Return ONLY the JSON object, no other text
- Do not wrap in markdown code blocks
- Do not include backticks, explanations, or any other formatting
- The command field must contain the exact command that can be executed
- The description field must be 1-2 lines maximum
- Ensure the JSON is properly formatted and valid
- Consider the operating system when suggesting commands%s`

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

func getSystemInfo() string {
	return runtime.GOOS
}

func buildPrompt(description string) string {
	osName := getSystemInfo()
	return fmt.Sprintf(BASE_PROMPT, description, osName, "", "")
}

func buildPromptWithContext(description, errorContext string) string {
	osName := getSystemInfo()
	errorSection := fmt.Sprintf("\nIMPORTANT CONTEXT: %s\n\nPlease suggest an alternative command that addresses the error or takes a different approach.\n", errorContext)
	additionalRequirement := "\n- Consider the error context and suggest a different approach"
	return fmt.Sprintf(BASE_PROMPT, description, osName, errorSection, additionalRequirement)
}

func GenerateCommand(description string, ctx context.Context) (string, error) {
	provider := getProvider(ctx)
	if provider == nil {
		return "", fmt.Errorf("no API key found. Set OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY")
	}
	return provider.GenerateCommand(description, ctx)
}

func GenerateCommandWithContext(description, errorContext string, ctx context.Context) (string, error) {
	provider := getProvider(ctx)
	if provider == nil {
		return "", fmt.Errorf("no API key found. Set OPENAI_API_KEY, ANTHROPIC_API_KEY, or GEMINI_API_KEY")
	}
	return provider.GenerateCommandWithContext(description, errorContext, ctx)
}
