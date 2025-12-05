package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AnthropicMessage defines the request format for Anthropic API
type AnthropicMessage struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature,omitempty"`
	Messages    []Message `json:"messages"`
	System      string    `json:"system,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse defines the response from Anthropic API
type AnthropicResponse struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Model        string    `json:"model"`
	StopReason   string    `json:"stop_reason"`
	Content      []Content `json:"content"`
	Usage        Usage     `json:"usage"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Anthropic implements the LLM interface for Anthropic Claude
type Anthropic struct {
	apiKey string
	config Config
	client *http.Client
}

// NewAnthropic creates a new Anthropic LLM provider
func NewAnthropic(cfg Config) (*Anthropic, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Anthropic")
	}

	return &Anthropic{
		apiKey: cfg.APIKey,
		config: cfg,
		client: &http.Client{},
	}, nil
}

// Generate sends a prompt to Anthropic and returns the response
func (a *Anthropic) Generate(ctx context.Context, prompt string) (string, error) {
	model := a.config.Model
	if model == "" {
		model = "claude-3-sonnet-20240229"
	}

	maxTokens := a.config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	message := AnthropicMessage{
		Model:       model,
		MaxTokens:   maxTokens,
		Temperature: a.config.Temperature,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic API error (status %d): %s", resp.StatusCode, string(body))
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return anthropicResp.Content[0].Text, nil
}
