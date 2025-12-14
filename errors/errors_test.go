package errors

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		contains []string
	}{
		{
			name: "simple error",
			err: &Error{
				Code:      ErrRequiredField,
				Message:   "field is required",
				Timestamp: time.Now(),
				Severity:  SeverityMedium,
			},
			contains: []string{"validation.required_field", "field is required"},
		},
		{
			name: "error with wrapped error",
			err: &Error{
				Code:      ErrAPICall,
				Message:   "API call failed",
				Err:       errors.New("network timeout"),
				Timestamp: time.Now(),
				Severity:  SeverityMedium,
			},
			contains: []string{"api.call_failed", "API call failed", "network timeout"},
		},
		{
			name: "error with context",
			err: &Error{
				Code:      ErrAPICall,
				Message:   "request failed",
				Context:   map[string]interface{}{"url": "https://api.example.com", "status": 500},
				Timestamp: time.Now(),
				Severity:  SeverityHigh,
			},
			contains: []string{"api.call_failed", "request failed", "url=", "status="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(got, substr) {
					t.Errorf("Error() = %v, should contain %v", got, substr)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	err := New(ErrRequiredField, "test error")
	if err.Code != ErrRequiredField {
		t.Errorf("New() code = %v, want %v", err.Code, ErrRequiredField)
	}
	if err.Message != "test error" {
		t.Errorf("New() message = %v, want %v", err.Message, "test error")
	}
	if err.Err != nil {
		t.Errorf("New() Err should be nil")
	}
	if err.Severity != SeverityMedium {
		t.Errorf("New() severity = %v, want %v", err.Severity, SeverityMedium)
	}
}

func TestNewf(t *testing.T) {
	err := Newf(ErrInvalidField, "field %s has value %d", "age", 150)
	expected := "field age has value 150"
	if err.Message != expected {
		t.Errorf("Newf() message = %v, want %v", err.Message, expected)
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(ErrAPICall, "operation failed", originalErr)

	if wrappedErr.Code != ErrAPICall {
		t.Errorf("Wrap() code = %v, want %v", wrappedErr.Code, ErrAPICall)
	}
	if wrappedErr.Message != "operation failed" {
		t.Errorf("Wrap() message = %v, want %v", wrappedErr.Message, "operation failed")
	}
	if !errors.Is(wrappedErr, originalErr) {
		t.Errorf("Wrap() should wrap the original error")
	}
}

func TestError_WithContext(t *testing.T) {
	err := New(ErrAPICall, "test").
		WithContext("request_id", "123").
		WithContext("user_id", 456)

	if err.Context["request_id"] != "123" {
		t.Errorf("WithContext() request_id = %v, want %v", err.Context["request_id"], "123")
	}
	if err.Context["user_id"] != 456 {
		t.Errorf("WithContext() user_id = %v, want %v", err.Context["user_id"], 456)
	}
}

func TestError_WithSeverity(t *testing.T) {
	err := New(ErrInternal, "test").WithSeverity(SeverityCritical)
	if err.Severity != SeverityCritical {
		t.Errorf("WithSeverity() = %v, want %v", err.Severity, SeverityCritical)
	}
}

func TestError_WithRetryable(t *testing.T) {
	err := New(ErrNetworkTimeout, "test").WithRetryable(true)
	if !err.Retryable {
		t.Error("WithRetryable(true) should set Retryable to true")
	}
}

func TestError_WithTemporary(t *testing.T) {
	err := New(ErrNetworkTimeout, "test").WithTemporary(true)
	if !err.Temporary {
		t.Error("WithTemporary(true) should set Temporary to true")
	}
}

func TestRequiredField(t *testing.T) {
	err := RequiredField("username")
	if err.Code != ErrRequiredField {
		t.Errorf("RequiredField() code = %v, want %v", err.Code, ErrRequiredField)
	}
	if !strings.Contains(err.Message, "username") {
		t.Errorf("RequiredField() message should contain field name")
	}
}

func TestInvalidField(t *testing.T) {
	err := InvalidField("email", "invalid format")
	if err.Code != ErrInvalidField {
		t.Errorf("InvalidField() code = %v, want %v", err.Code, ErrInvalidField)
	}
	if !strings.Contains(err.Message, "email") || !strings.Contains(err.Message, "invalid format") {
		t.Errorf("InvalidField() message should contain field name and reason")
	}
}

func TestAPICallError(t *testing.T) {
	originalErr := errors.New("connection refused")
	err := APICallError("fetch data", originalErr)

	if err.Code != ErrAPICall {
		t.Errorf("APICallError() code = %v, want %v", err.Code, ErrAPICall)
	}
	if !err.Retryable {
		t.Error("APICallError() should be retryable")
	}
	if !err.Temporary {
		t.Error("APICallError() should be temporary")
	}
	if !errors.Is(err, originalErr) {
		t.Error("APICallError() should wrap the original error")
	}
}

func TestNetworkTimeout(t *testing.T) {
	err := NetworkTimeout("API request")
	if err.Code != ErrNetworkTimeout {
		t.Errorf("NetworkTimeout() code = %v, want %v", err.Code, ErrNetworkTimeout)
	}
	if !err.Retryable {
		t.Error("NetworkTimeout() should be retryable")
	}
	if !err.Temporary {
		t.Error("NetworkTimeout() should be temporary")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "retryable error",
			err:      NetworkTimeout("test"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      RequiredField("test"),
			expected: false,
		},
		{
			name:     "standard error",
			err:      errors.New("test"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.expected {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsTemporary(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "temporary error",
			err:      NetworkTimeout("test"),
			expected: true,
		},
		{
			name:     "non-temporary error",
			err:      InvalidField("test", "reason"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTemporary(tt.err); got != tt.expected {
				t.Errorf("IsTemporary() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHasCode(t *testing.T) {
	err := RequiredField("test")
	if !HasCode(err, ErrRequiredField) {
		t.Error("HasCode() should return true for matching code")
	}
	if HasCode(err, ErrAPICall) {
		t.Error("HasCode() should return false for non-matching code")
	}
}

func TestGetSeverity(t *testing.T) {
	err := Internal("test")
	if GetSeverity(err) != SeverityHigh {
		t.Errorf("GetSeverity() = %v, want %v", GetSeverity(err), SeverityHigh)
	}

	standardErr := errors.New("standard")
	if GetSeverity(standardErr) != SeverityLow {
		t.Errorf("GetSeverity() for standard error should return SeverityLow")
	}
}

func TestErrorCode_String(t *testing.T) {
	code := ErrorCode{Category: "test", Code: "example"}
	expected := "test.example"
	if code.String() != expected {
		t.Errorf("ErrorCode.String() = %v, want %v", code.String(), expected)
	}
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityLow, "LOW"},
		{SeverityMedium, "MEDIUM"},
		{SeverityHigh, "HIGH"},
		{SeverityCritical, "CRITICAL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.expected {
				t.Errorf("Severity.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
