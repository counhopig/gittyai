package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAILikeConfig represents the configuration for OpenAI-compatible providers
type OpenAILikeConfig struct {
	// BaseURL is the API endpoint (e.g., "https://api.openai.com/v1", "http://localhost:11434/v1")
	BaseURL string
	// APIKey for authentication (optional for some local providers)
	APIKey string
	// Model name to use
	Model string
	// Temperature for generation (0.0 to 2.0)
	Temperature float32
	// MaxTokens limits the response length
	MaxTokens int
	// Headers allows custom HTTP headers
	Headers map[string]string
	// SystemPrompt is an optional system message
	SystemPrompt string
}

// OpenAILike implements the LLM interface for any OpenAI-compatible API
// This includes: Azure OpenAI, Ollama, LM Studio, vLLM, LocalAI, Together AI,
// Groq, Fireworks AI, Deepseek, and many other providers
type OpenAILike struct {
	config OpenAILikeConfig
	client *http.Client
}

// NewOpenAILike creates a new OpenAI-compatible LLM provider
func NewOpenAILike(cfg OpenAILikeConfig) (*OpenAILike, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	if cfg.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	return &OpenAILike{
		config: cfg,
		client: &http.Client{},
	}, nil
}

// Generate sends a prompt to the OpenAI-compatible API and returns the response
func (o *OpenAILike) Generate(ctx context.Context, prompt string) (string, error) {
	messages := make([]openAIMessage, 0, 2)

	// Add system prompt if provided
	if o.config.SystemPrompt != "" {
		messages = append(messages, openAIMessage{
			Role:    "system",
			Content: o.config.SystemPrompt,
		})
	}

	// Add user message
	messages = append(messages, openAIMessage{
		Role:    "user",
		Content: prompt,
	})

	reqBody := openAIRequest{
		Model:       o.config.Model,
		Temperature: o.config.Temperature,
		MaxTokens:   o.config.MaxTokens,
		Messages:    messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the endpoint URL
	endpoint := o.config.BaseURL
	// Ensure URL ends properly
	if endpoint[len(endpoint)-1] != '/' {
		endpoint += "/"
	}
	endpoint += "chat/completions"

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")

	// Set Authorization header if API key is provided
	if o.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.config.APIKey)
	}

	// Apply custom headers
	for key, value := range o.config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp openAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("API error: %s", apiResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return apiResp.Choices[0].Message.Content, nil
}

// Common preset constructors for popular providers

// NewOllama creates a new LLM provider for Ollama
func NewOllama(model string, baseURL ...string) (*OpenAILike, error) {
	url := "http://localhost:11434/v1"
	if len(baseURL) > 0 && baseURL[0] != "" {
		url = baseURL[0]
	}
	return NewOpenAILike(OpenAILikeConfig{
		BaseURL: url,
		Model:   model,
	})
}

// NewLMStudio creates a new LLM provider for LM Studio
func NewLMStudio(model string, baseURL ...string) (*OpenAILike, error) {
	url := "http://localhost:1234/v1"
	if len(baseURL) > 0 && baseURL[0] != "" {
		url = baseURL[0]
	}
	return NewOpenAILike(OpenAILikeConfig{
		BaseURL: url,
		Model:   model,
	})
}

// NewAzureOpenAI creates a new LLM provider for Azure OpenAI Service
func NewAzureOpenAI(cfg AzureOpenAIConfig) (*OpenAILike, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("azure endpoint is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key is required")
	}
	if cfg.DeploymentName == "" {
		return nil, fmt.Errorf("deployment name is required")
	}

	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = "2024-02-15-preview"
	}

	// Azure OpenAI uses a different URL structure
	baseURL := fmt.Sprintf("%s/openai/deployments/%s", cfg.Endpoint, cfg.DeploymentName)

	return NewOpenAILike(OpenAILikeConfig{
		BaseURL:     baseURL,
		APIKey:      cfg.APIKey,
		Model:       cfg.DeploymentName,
		Temperature: cfg.Temperature,
		MaxTokens:   cfg.MaxTokens,
		Headers: map[string]string{
			"api-key": cfg.APIKey,
		},
	})
}

// AzureOpenAIConfig represents configuration for Azure OpenAI Service
type AzureOpenAIConfig struct {
	// Endpoint is the Azure OpenAI resource endpoint
	Endpoint string
	// APIKey is the Azure OpenAI API key
	APIKey string
	// DeploymentName is the name of the deployed model
	DeploymentName string
	// APIVersion is the API version (default: "2024-02-15-preview")
	APIVersion string
	// Temperature for generation
	Temperature float32
	// MaxTokens limits the response length
	MaxTokens int
}

// NewGroq creates a new LLM provider for Groq
func NewGroq(apiKey, model string) (*OpenAILike, error) {
	if model == "" {
		model = "llama-3.1-70b-versatile"
	}
	return NewOpenAILike(OpenAILikeConfig{
		BaseURL: "https://api.groq.com/openai/v1",
		APIKey:  apiKey,
		Model:   model,
	})
}

// NewTogether creates a new LLM provider for Together AI
func NewTogether(apiKey, model string) (*OpenAILike, error) {
	if model == "" {
		model = "meta-llama/Llama-3-70b-chat-hf"
	}
	return NewOpenAILike(OpenAILikeConfig{
		BaseURL: "https://api.together.xyz/v1",
		APIKey:  apiKey,
		Model:   model,
	})
}

// NewDeepseek creates a new LLM provider for Deepseek
func NewDeepseek(apiKey, model string) (*OpenAILike, error) {
	if model == "" {
		model = "deepseek-chat"
	}
	return NewOpenAILike(OpenAILikeConfig{
		BaseURL: "https://api.deepseek.com/v1",
		APIKey:  apiKey,
		Model:   model,
	})
}

// NewFireworks creates a new LLM provider for Fireworks AI
func NewFireworks(apiKey, model string) (*OpenAILike, error) {
	if model == "" {
		model = "accounts/fireworks/models/llama-v3p1-70b-instruct"
	}
	return NewOpenAILike(OpenAILikeConfig{
		BaseURL: "https://api.fireworks.ai/inference/v1",
		APIKey:  apiKey,
		Model:   model,
	})
}
