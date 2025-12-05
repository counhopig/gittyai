package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAI request/response types
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float32         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// OpenAI implements the LLM interface for OpenAI
type OpenAI struct {
	apiKey string
	config Config
	client *http.Client
}

// NewOpenAI creates a new OpenAI LLM provider
func NewOpenAI(cfg Config) (*OpenAI, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	return &OpenAI{
		apiKey: cfg.APIKey,
		config: cfg,
		client: &http.Client{},
	}, nil
}

// Generate sends a prompt to OpenAI and returns the response
func (o *OpenAI) Generate(ctx context.Context, prompt string) (string, error) {
	model := o.config.Model
	if model == "" {
		model = "gpt-4-turbo-preview"
	}

	reqBody := openAIRequest{
		Model:       model,
		Temperature: o.config.Temperature,
		MaxTokens:   o.config.MaxTokens,
		Messages: []openAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("openai API error: %s", openAIResp.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai API error (status %d): %s", resp.StatusCode, string(body))
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from openai")
	}

	return openAIResp.Choices[0].Message.Content, nil
}
