package llm

import "context"

// LLM is the interface for Language Model providers
type LLM interface {
	// Generate sends a prompt to the LLM and returns the response
	Generate(ctx context.Context, prompt string) (string, error)
}

// Config represents the base configuration for an LLM
type Config struct {
	// API key for the provider
	APIKey string
	// Model name (e.g. "gpt-4o", "claude-3-opus-20240229")
	Model string
	// Temperature for generation (0.0 to 2.0)
	Temperature float32
	// MaxTokens limits the response length
	MaxTokens int
}
