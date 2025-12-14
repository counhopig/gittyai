package errors

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorCode represents a structured error code with category and specific code
type ErrorCode struct {
	Category string // High-level category (e.g., "validation", "config", "api")
	Code     string // Specific error code (e.g., "required_field", "invalid_format")
}

// String returns the string representation of the error code
func (ec ErrorCode) String() string {
	return fmt.Sprintf("%s.%s", ec.Category, ec.Code)
}

// Error categories
const (
	CategoryValidation  = "validation"
	CategoryConfig      = "config"
	CategoryAPI         = "api"
	CategoryNetwork     = "network"
	CategoryInternal    = "internal"
	CategoryUnsupported = "unsupported"
	CategoryNotFound    = "notfound"
	CategoryAuth        = "auth"
	CategoryTimeout     = "timeout"
	CategoryRateLimit   = "ratelimit"
)

// Predefined error codes
var (
	// Validation errors
	ErrRequiredField = ErrorCode{CategoryValidation, "required_field"}
	ErrInvalidField  = ErrorCode{CategoryValidation, "invalid_field"}
	ErrInvalidFormat = ErrorCode{CategoryValidation, "invalid_format"}
	ErrOutOfRange    = ErrorCode{CategoryValidation, "out_of_range"}

	// Configuration errors
	ErrMissingConfig  = ErrorCode{CategoryConfig, "missing_config"}
	ErrInvalidConfig  = ErrorCode{CategoryConfig, "invalid_config"}
	ErrProviderConfig = ErrorCode{CategoryConfig, "provider_config"}

	// API errors
	ErrAPICall       = ErrorCode{CategoryAPI, "call_failed"}
	ErrAPIResponse   = ErrorCode{CategoryAPI, "invalid_response"}
	ErrAPIStatusCode = ErrorCode{CategoryAPI, "bad_status_code"}

	// Network errors
	ErrNetworkTimeout = ErrorCode{CategoryNetwork, "timeout"}
	ErrNetworkRefused = ErrorCode{CategoryNetwork, "connection_refused"}
	ErrNetworkUnavail = ErrorCode{CategoryNetwork, "unavailable"}

	// Internal errors
	ErrInternal       = ErrorCode{CategoryInternal, "internal"}
	ErrNotImplemented = ErrorCode{CategoryInternal, "not_implemented"}
	ErrUnexpected     = ErrorCode{CategoryInternal, "unexpected"}

	// Not found errors
	ErrNotFound      = ErrorCode{CategoryNotFound, "resource"}
	ErrAgentNotFound = ErrorCode{CategoryNotFound, "agent"}
	ErrTaskNotFound  = ErrorCode{CategoryNotFound, "task"}

	// Unsupported errors
	ErrUnsupported     = ErrorCode{CategoryUnsupported, "feature"}
	ErrUnsupportedType = ErrorCode{CategoryUnsupported, "type"}

	// Auth errors
	ErrUnauthorized  = ErrorCode{CategoryAuth, "unauthorized"}
	ErrInvalidAPIKey = ErrorCode{CategoryAuth, "invalid_api_key"}

	// Rate limit and timeout
	ErrRateLimitExceeded = ErrorCode{CategoryRateLimit, "exceeded"}
	ErrTimeout           = ErrorCode{CategoryTimeout, "exceeded"}
)

// Severity levels for errors
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// StackFrame represents a single frame in the call stack
type StackFrame struct {
	File     string
	Line     int
	Function string
}

// Error represents a structured error with rich context
type Error struct {
	// Core error information
	Code    ErrorCode
	Message string
	Err     error // Wrapped error

	// Context and metadata
	Context   map[string]interface{} // Additional context information
	Severity  Severity               // Error severity
	Timestamp time.Time              // When the error occurred
	Stack     []StackFrame           // Call stack

	// Classification
	Retryable bool // Whether the operation can be retried
	Temporary bool // Whether the error is temporary
}

// Error implements the error interface
func (e *Error) Error() string {
	var parts []string

	// Add error code
	parts = append(parts, fmt.Sprintf("[%s]", e.Code.String()))

	// Add severity if not low
	if e.Severity > SeverityLow {
		parts = append(parts, fmt.Sprintf("[%s]", e.Severity))
	}

	// Add message
	parts = append(parts, e.Message)

	// Add wrapped error if present
	if e.Err != nil {
		parts = append(parts, fmt.Sprintf("caused by: %v", e.Err))
	}

	// Add context if present
	if len(e.Context) > 0 {
		var ctxParts []string
		for k, v := range e.Context {
			ctxParts = append(ctxParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("context: {%s}", strings.Join(ctxParts, ", ")))
	}

	return strings.Join(parts, " | ")
}

// Unwrap returns the underlying error for error chain traversal
func (e *Error) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target error code
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Code == t.Code
	}
	return false
}

// WithContext adds context information to the error
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithSeverity sets the severity of the error
func (e *Error) WithSeverity(severity Severity) *Error {
	e.Severity = severity
	return e
}

// WithRetryable marks the error as retryable
func (e *Error) WithRetryable(retryable bool) *Error {
	e.Retryable = retryable
	return e
}

// WithTemporary marks the error as temporary
func (e *Error) WithTemporary(temporary bool) *Error {
	e.Temporary = temporary
	return e
}

// captureStack captures the current call stack
func captureStack(skip int) []StackFrame {
	var frames []StackFrame
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		fnName := "unknown"
		if fn != nil {
			fnName = fn.Name()
		}
		frames = append(frames, StackFrame{
			File:     file,
			Line:     line,
			Function: fnName,
		})
	}
	return frames
}

// New creates a new structured error
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Severity:  SeverityMedium,
		Stack:     captureStack(2),
	}
}

// Newf creates a new structured error with formatted message
func Newf(code ErrorCode, format string, args ...interface{}) *Error {
	return New(code, fmt.Sprintf(format, args...))
}

// Wrap wraps an existing error with additional context
func Wrap(code ErrorCode, message string, err error) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Err:       err,
		Timestamp: time.Now(),
		Severity:  SeverityMedium,
		Stack:     captureStack(2),
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(code ErrorCode, err error, format string, args ...interface{}) *Error {
	return Wrap(code, fmt.Sprintf(format, args...), err)
}

// FromError converts a standard error to a structured error
func FromError(err error, code ErrorCode) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return Wrap(code, "error occurred", err)
}

// Validation errors

// RequiredField returns a validation error for a required field
func RequiredField(fieldName string) *Error {
	return Newf(ErrRequiredField, "field '%s' is required", fieldName)
}

// InvalidField returns a validation error for an invalid field
func InvalidField(fieldName, reason string) *Error {
	return Newf(ErrInvalidField, "field '%s' is invalid: %s", fieldName, reason)
}

// InvalidFormat returns a validation error for invalid format
func InvalidFormat(fieldName, expected string) *Error {
	return Newf(ErrInvalidFormat, "field '%s' has invalid format, expected: %s", fieldName, expected)
}

// OutOfRange returns a validation error for out of range value
func OutOfRange(fieldName string, min, max interface{}) *Error {
	return Newf(ErrOutOfRange, "field '%s' is out of range [%v, %v]", fieldName, min, max)
}

// Validation creates a generic validation error
func Validation(message string) *Error {
	return New(ErrorCode{CategoryValidation, "error"}, message)
}

// Validationf creates a generic validation error with formatting
func Validationf(format string, args ...interface{}) *Error {
	return Newf(ErrorCode{CategoryValidation, "error"}, format, args...)
}

// Configuration errors

// MissingConfig returns a configuration missing error
func MissingConfig(configName string) *Error {
	return Newf(ErrMissingConfig, "configuration '%s' is missing", configName)
}

// InvalidConfig returns an invalid configuration error
func InvalidConfig(configName, reason string) *Error {
	return Newf(ErrInvalidConfig, "configuration '%s' is invalid: %s", configName, reason)
}

// ProviderError returns an error for provider-specific issues
func ProviderError(provider, message string) *Error {
	return Newf(ErrProviderConfig, "%s: %s", provider, message)
}

// Config creates a generic configuration error
func Config(message string) *Error {
	return New(ErrorCode{CategoryConfig, "error"}, message)
}

// Configf creates a generic configuration error with formatting
func Configf(format string, args ...interface{}) *Error {
	return Newf(ErrorCode{CategoryConfig, "error"}, format, args...)
}

// API errors

// APICallError returns an error for API call failures
func APICallError(operation string, err error) *Error {
	return Wrap(ErrAPICall, fmt.Sprintf("failed to %s", operation), err).
		WithRetryable(true).
		WithTemporary(true)
}

// APIResponseError returns an error for invalid API responses
func APIResponseError(message string) *Error {
	return New(ErrAPIResponse, message)
}

// APIStatusCodeError returns an error for bad status codes
func APIStatusCodeError(statusCode int, body string) *Error {
	return Newf(ErrAPIStatusCode, "unexpected status code %d: %s", statusCode, body)
}

// API creates a generic API error
func API(message string) *Error {
	return New(ErrorCode{CategoryAPI, "error"}, message)
}

// APIf creates a generic API error with formatting
func APIf(format string, args ...interface{}) *Error {
	return Newf(ErrorCode{CategoryAPI, "error"}, format, args...)
}

// Network errors

// NetworkTimeout returns a network timeout error
func NetworkTimeout(operation string) *Error {
	return Newf(ErrNetworkTimeout, "network timeout during %s", operation).
		WithRetryable(true).
		WithTemporary(true)
}

// NetworkUnavailable returns a network unavailable error
func NetworkUnavailable(service string) *Error {
	return Newf(ErrNetworkUnavail, "service '%s' is unavailable", service).
		WithRetryable(true).
		WithTemporary(true)
}

// Network creates a generic network error
func Network(message string) *Error {
	return New(ErrorCode{CategoryNetwork, "error"}, message)
}

// Networkf creates a generic network error with formatting
func Networkf(format string, args ...interface{}) *Error {
	return Newf(ErrorCode{CategoryNetwork, "error"}, format, args...)
}

// Internal errors

// Internal creates a generic internal error
func Internal(message string) *Error {
	return New(ErrInternal, message).WithSeverity(SeverityHigh)
}

// Internalf creates a generic internal error with formatting
func Internalf(format string, args ...interface{}) *Error {
	return Newf(ErrInternal, format, args...).WithSeverity(SeverityHigh)
}

// NotImplemented returns a not implemented error
func NotImplemented(feature string) *Error {
	return Newf(ErrNotImplemented, "feature '%s' is not implemented", feature)
}

// Unexpected returns an unexpected error
func Unexpected(message string) *Error {
	return New(ErrUnexpected, message).WithSeverity(SeverityHigh)
}

// NotFound errors

// NotFound creates a generic not found error
func NotFound(resourceType, identifier string) *Error {
	return Newf(ErrNotFound, "%s '%s' not found", resourceType, identifier)
}

// AgentNotFound returns an agent not found error
func AgentNotFound(agentName string) *Error {
	return Newf(ErrAgentNotFound, "agent '%s' not found", agentName)
}

// TaskNotFound returns a task not found error
func TaskNotFound(taskID string) *Error {
	return Newf(ErrTaskNotFound, "task '%s' not found", taskID)
}

// Unsupported errors

// Unsupported creates a generic unsupported error
func Unsupported(feature string) *Error {
	return Newf(ErrUnsupported, "feature '%s' is not supported", feature)
}

// Unsupportedf creates a generic unsupported error with formatting
func Unsupportedf(format string, args ...interface{}) *Error {
	return Newf(ErrUnsupported, format, args...)
}

// UnsupportedType returns an unsupported type error
func UnsupportedType(typeName string) *Error {
	return Newf(ErrUnsupportedType, "type '%s' is not supported", typeName)
}

// Auth errors

// Unauthorized returns an unauthorized error
func Unauthorized(message string) *Error {
	return New(ErrUnauthorized, message)
}

// InvalidAPIKey returns an invalid API key error
func InvalidAPIKey(provider string) *Error {
	return Newf(ErrInvalidAPIKey, "invalid API key for provider '%s'", provider)
}

// Rate limit and timeout

// RateLimitExceeded returns a rate limit exceeded error
func RateLimitExceeded(resource string, limit int) *Error {
	return Newf(ErrRateLimitExceeded, "rate limit exceeded for '%s' (limit: %d)", resource, limit).
		WithRetryable(true).
		WithTemporary(true)
}

// Timeout returns a timeout error
func Timeout(operation string, duration time.Duration) *Error {
	return Newf(ErrTimeout, "operation '%s' timed out after %v", operation, duration).
		WithRetryable(true).
		WithTemporary(true)
}

// Helper functions for error checking

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Retryable
	}
	return false
}

// IsTemporary checks if an error is temporary
func IsTemporary(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Temporary
	}
	return false
}

// HasCode checks if an error has the specified code
func HasCode(err error, code ErrorCode) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// GetSeverity returns the severity of an error
func GetSeverity(err error) Severity {
	if e, ok := err.(*Error); ok {
		return e.Severity
	}
	return SeverityLow
}
