package errors

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	baseErr := New(ErrAPICall, "API request failed").
		WithContext("url", "https://api.example.com").
		WithContext("status", 500).
		WithSeverity(SeverityHigh)

	tests := []struct {
		name     string
		err      error
		option   FormatOption
		contains []string
	}{
		{
			name:     "simple format",
			err:      baseErr,
			option:   FormatSimple,
			contains: []string{"API request failed"},
		},
		{
			name:     "detailed format",
			err:      baseErr,
			option:   FormatDetailed,
			contains: []string{"api.call_failed", "HIGH", "API request failed", "url:", "status:"},
		},
		{
			name:     "JSON format",
			err:      baseErr,
			option:   FormatJSON,
			contains: []string{`"code"`, `"message"`, `"severity"`, `"context"`},
		},
		{
			name:     "with stack format",
			err:      baseErr,
			option:   FormatWithStack,
			contains: []string{"Stack trace:", "api.call_failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.err, tt.option)
			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("Format() = %v, should contain %v", got, substr)
				}
			}
		})
	}
}

func TestFormat_NilError(t *testing.T) {
	result := Format(nil, FormatSimple)
	if result != "" {
		t.Errorf("Format(nil) should return empty string, got %v", result)
	}
}

func TestFormat_StandardError(t *testing.T) {
	stdErr := errors.New("standard error")
	result := Format(stdErr, FormatDetailed)
	if result != "standard error" {
		t.Errorf("Format(standard error) = %v, want %v", result, "standard error")
	}
}

func TestError_MarshalJSON(t *testing.T) {
	err := New(ErrAPICall, "test error").
		WithContext("key", "value").
		WithSeverity(SeverityHigh)

	data, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Fatalf("MarshalJSON() error = %v", marshalErr)
	}

	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(data, &result); unmarshalErr != nil {
		t.Fatalf("Unmarshal() error = %v", unmarshalErr)
	}

	if result["code"] != "api.call_failed" {
		t.Errorf("MarshalJSON() code = %v, want %v", result["code"], "api.call_failed")
	}
	if result["message"] != "test error" {
		t.Errorf("MarshalJSON() message = %v, want %v", result["message"], "test error")
	}
	if result["severity"] != "HIGH" {
		t.Errorf("MarshalJSON() severity = %v, want %v", result["severity"], "HIGH")
	}
}

func TestError_ToMap(t *testing.T) {
	err := New(ErrAPICall, "test error").
		WithContext("request_id", "123").
		WithSeverity(SeverityMedium)

	m := err.ToMap()

	if m["error_code"] != "api.call_failed" {
		t.Errorf("ToMap() error_code = %v, want %v", m["error_code"], "api.call_failed")
	}
	if m["error_message"] != "test error" {
		t.Errorf("ToMap() error_message = %v, want %v", m["error_message"], "test error")
	}
	if m["error_severity"] != "MEDIUM" {
		t.Errorf("ToMap() error_severity = %v, want %v", m["error_severity"], "MEDIUM")
	}
	if m["ctx_request_id"] != "123" {
		t.Errorf("ToMap() ctx_request_id = %v, want %v", m["ctx_request_id"], "123")
	}
}

func TestErrorChain(t *testing.T) {
	err1 := New(ErrInternal, "level 1")
	err2 := Wrap(ErrAPICall, "level 2", err1)
	err3 := Wrap(ErrNetworkTimeout, "level 3", err2)

	chain := ErrorChain(err3)

	if len(chain) != 3 {
		t.Errorf("ErrorChain() length = %v, want %v", len(chain), 3)
	}

	if chain[0].Code != ErrNetworkTimeout {
		t.Errorf("ErrorChain()[0] code = %v, want %v", chain[0].Code, ErrNetworkTimeout)
	}
	if chain[1].Code != ErrAPICall {
		t.Errorf("ErrorChain()[1] code = %v, want %v", chain[1].Code, ErrAPICall)
	}
	if chain[2].Code != ErrInternal {
		t.Errorf("ErrorChain()[2] code = %v, want %v", chain[2].Code, ErrInternal)
	}
}

func TestRootCause(t *testing.T) {
	rootErr := errors.New("root cause")
	err1 := Wrap(ErrInternal, "level 1", rootErr)
	err2 := Wrap(ErrAPICall, "level 2", err1)

	root := RootCause(err2)

	if root != rootErr {
		t.Errorf("RootCause() = %v, want %v", root, rootErr)
	}
}

func TestRootCause_SingleError(t *testing.T) {
	err := New(ErrAPICall, "single error")
	root := RootCause(err)

	if root != err {
		t.Errorf("RootCause() should return the error itself when there's no wrapped error")
	}
}
