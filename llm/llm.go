package llm

import "context"

// LLM is the interface for Language Model providers
type LLM interface {
	// Generate sends a prompt to the LLM and returns the response
	Generate(ctx context.Context, prompt string) (string, error)
}

// StructuredLLM extends LLM with structured output support
type StructuredLLM interface {
	LLM
	// GenerateStructured sends a prompt and returns structured JSON response
	// schema is a JSON schema that defines the expected output format
	GenerateStructured(ctx context.Context, prompt string, schema *JSONSchema) (string, error)
}

// JSONSchema represents a JSON Schema for structured output
type JSONSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Schema      *SchemaDefinition      `json:"schema"`
	Strict      bool                   `json:"strict,omitempty"`
}

// SchemaDefinition defines the structure of the expected output
type SchemaDefinition struct {
	Type                 string                        `json:"type"`
	Properties           map[string]*SchemaDefinition  `json:"properties,omitempty"`
	Items                *SchemaDefinition             `json:"items,omitempty"`
	Required             []string                      `json:"required,omitempty"`
	AdditionalProperties bool                          `json:"additionalProperties,omitempty"`
	Enum                 []string                      `json:"enum,omitempty"`
	Description          string                        `json:"description,omitempty"`
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

