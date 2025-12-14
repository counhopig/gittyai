package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/counhopig/gittyai/errors"
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
		return nil, errors.RequiredField("API key")
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
		return "", errors.Wrap(errors.ErrInternal, "failed to marshal request", err).WithContext("model", model)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", errors.APICallError("call OpenAI API", err).WithContext("model", model).WithContext("prompt_length", len(prompt))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(errors.ErrNetworkUnavail, "failed to read response", err).WithRetryable(true).WithTemporary(true)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", errors.Wrap(errors.ErrInternal, "failed to unmarshal response", err).WithContext("response_length", len(body))
	}

	if openAIResp.Error != nil {
		return "", errors.APIResponseError(openAIResp.Error.Message).WithContext("type", openAIResp.Error.Type).WithContext("code", openAIResp.Error.Code)
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.APIStatusCodeError(resp.StatusCode, string(body)).WithContext("model", model)
	}

	if len(openAIResp.Choices) == 0 {
		return "", errors.API("no response from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}
