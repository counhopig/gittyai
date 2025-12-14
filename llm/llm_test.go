package llm

import (
	"context"
	"testing"

	"github.com/counhopig/gittyai/errors"
)

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				APIKey:      "test-key",
				Model:       "test-model",
				Temperature: 0.7,
				MaxTokens:   1000,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: Config{
				APIKey:      "",
				Model:       "test-model",
				Temperature: 0.7,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with NewOpenAI (requires API key)
			_, err := NewOpenAI(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewOpenAI() expected error, got nil")
				} else if errObj, ok := err.(*errors.Error); ok {
					if errObj.Code != errors.ErrRequiredField {
						t.Errorf("NewOpenAI() error code = %v, want %v", errObj.Code, errors.ErrRequiredField)
					}
				}
			} else {
				if err != nil {
					t.Errorf("NewOpenAI() unexpected error: %v", err)
				}
			}

			// Test with NewAnthropic (requires API key)
			_, err = NewAnthropic(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAnthropic() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("NewAnthropic() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestOpenAILikeConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  OpenAILikeConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: OpenAILikeConfig{
				BaseURL:     "https://api.example.com/v1",
				Model:       "test-model",
				Temperature: 0.7,
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			config: OpenAILikeConfig{
				BaseURL:     "",
				Model:       "test-model",
				Temperature: 0.7,
			},
			wantErr: true,
		},
		{
			name: "missing model",
			config: OpenAILikeConfig{
				BaseURL:     "https://api.example.com/v1",
				Model:       "",
				Temperature: 0.7,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOpenAILike(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewOpenAILike() expected error, got nil")
				} else if errObj, ok := err.(*errors.Error); ok {
					if errObj.Code != errors.ErrRequiredField {
						t.Errorf("NewOpenAILike() error code = %v, want %v", errObj.Code, errors.ErrRequiredField)
					}
				}
			} else {
				if err != nil {
					t.Errorf("NewOpenAILike() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAzureOpenAIConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  AzureOpenAIConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: AzureOpenAIConfig{
				Endpoint:       "https://test.openai.azure.com",
				APIKey:         "test-key",
				DeploymentName: "gpt-4o",
				APIVersion:     "2024-02-15-preview",
				Temperature:    0.7,
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			config: AzureOpenAIConfig{
				Endpoint:       "",
				APIKey:         "test-key",
				DeploymentName: "gpt-4o",
				APIVersion:     "2024-02-15-preview",
			},
			wantErr: true,
		},
		{
			name: "missing API key",
			config: AzureOpenAIConfig{
				Endpoint:       "https://test.openai.azure.com",
				APIKey:         "",
				DeploymentName: "gpt-4o",
				APIVersion:     "2024-02-15-preview",
			},
			wantErr: true,
		},
		{
			name: "missing deployment name",
			config: AzureOpenAIConfig{
				Endpoint:       "https://test.openai.azure.com",
				APIKey:         "test-key",
				DeploymentName: "",
				APIVersion:     "2024-02-15-preview",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAzureOpenAI(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewAzureOpenAI() expected error, got nil")
				} else if errObj, ok := err.(*errors.Error); ok {
					if errObj.Code != errors.ErrRequiredField {
						t.Errorf("NewAzureOpenAI() error code = %v, want %v", errObj.Code, errors.ErrRequiredField)
					}
				}
			} else {
				if err != nil {
					t.Errorf("NewAzureOpenAI() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConstructorFunctions(t *testing.T) {
	ctx := context.Background()

	// Note: These tests don't actually make API calls, they just verify
	// that the constructors work without errors when given valid inputs

	t.Run("NewOllama", func(t *testing.T) {
		llm, err := NewOllama("llama3.2")
		if err != nil {
			t.Errorf("NewOllama() error = %v", err)
			return
		}
		if llm == nil {
			t.Error("NewOllama() returned nil")
		}
	})

	t.Run("NewGroq", func(t *testing.T) {
		llm, err := NewGroq("test-key", "llama-3.1-70b-versatile")
		if err != nil {
			t.Errorf("NewGroq() error = %v", err)
			return
		}
		if llm == nil {
			t.Error("NewGroq() returned nil")
		}
	})

	t.Run("NewDeepseek", func(t *testing.T) {
		llm, err := NewDeepseek("test-key", "deepseek-chat")
		if err != nil {
			t.Errorf("NewDeepseek() error = %v", err)
			return
		}
		if llm == nil {
			t.Error("NewDeepseek() returned nil")
		}
	})

	t.Run("NewOpenRouter", func(t *testing.T) {
		llm, err := NewOpenRouter("test-key", "openai/gpt-4o-mini")
		if err != nil {
			t.Errorf("NewOpenRouter() error = %v", err)
			return
		}
		if llm == nil {
			t.Error("NewOpenRouter() returned nil")
		}
	})

	t.Run("NewTogether", func(t *testing.T) {
		llm, err := NewTogether("test-key", "meta-llama/Llama-3-70b-chat-hf")
		if err != nil {
			t.Errorf("NewTogether() error = %v", err)
			return
		}
		if llm == nil {
			t.Error("NewTogether() returned nil")
		}
	})

	t.Run("NewLMStudio", func(t *testing.T) {
		llm, err := NewLMStudio("local-model")
		if err != nil {
			t.Errorf("NewLMStudio() error = %v", err)
			return
		}
		if llm == nil {
			t.Error("NewLMStudio() returned nil")
		}
	})

	// Test that LLM interface is implemented
	var _ LLM = (*OpenAI)(nil)
	var _ LLM = (*Anthropic)(nil)
	var _ LLM = (*OpenAILike)(nil)

	// Test context cancellation
	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		// Create a mock LLM that would respect context cancellation
		config := Config{
			APIKey: "test-key",
			Model:  "test-model",
		}

		openai, err := NewOpenAI(config)
		if err != nil {
			t.Skipf("Skipping context test: %v", err)
		}

		// The actual API call would fail due to cancelled context
		// This just verifies the function signature
		_, _ = openai.Generate(ctx, "test prompt")
		// We don't check the error since it depends on network/API
	})
}

func TestDefaultModels(t *testing.T) {
	tests := []struct {
		name     string
		function func() (*OpenAILike, error)
		expected string
	}{
		{"Ollama default", func() (*OpenAILike, error) { return NewOllama("llama3.2") }, "llama3.2"},
		{"Groq default", func() (*OpenAILike, error) { return NewGroq("key", "") }, "llama-3.1-70b-versatile"},
		{"Together default", func() (*OpenAILike, error) { return NewTogether("key", "") }, "meta-llama/Llama-3-70b-chat-hf"},
		{"Deepseek default", func() (*OpenAILike, error) { return NewDeepseek("key", "") }, "deepseek-chat"},
		{"OpenRouter default", func() (*OpenAILike, error) { return NewOpenRouter("key", "") }, "openai/gpt-4o-mini"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := tt.function()
			if err != nil {
				t.Errorf("%s error = %v", tt.name, err)
				return
			}
			if llm.config.Model != tt.expected {
				t.Errorf("%s model = %v, want %v", tt.name, llm.config.Model, tt.expected)
			}
		})
	}
}
