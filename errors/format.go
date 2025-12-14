package errors

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatOption defines formatting options for errors
type FormatOption int

const (
	// FormatSimple outputs a simple error message
	FormatSimple FormatOption = iota
	// FormatDetailed outputs a detailed error with all context
	FormatDetailed
	// FormatJSON outputs error as JSON
	FormatJSON
	// FormatWithStack outputs error with stack trace
	FormatWithStack
)

// Format formats an error according to the specified option
func Format(err error, option FormatOption) string {
	if err == nil {
		return ""
	}

	e, ok := err.(*Error)
	if !ok {
		return err.Error()
	}

	switch option {
	case FormatSimple:
		return formatSimple(e)
	case FormatDetailed:
		return formatDetailed(e)
	case FormatJSON:
		return formatJSON(e)
	case FormatWithStack:
		return formatWithStack(e)
	default:
		return e.Error()
	}
}

// formatSimple returns a simple error message
func formatSimple(e *Error) string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// formatDetailed returns a detailed error message
func formatDetailed(e *Error) string {
	var parts []string

	// Header with code and severity
	header := fmt.Sprintf("Error [%s] [%s]", e.Code.String(), e.Severity)
	parts = append(parts, header)
	parts = append(parts, strings.Repeat("-", len(header)))

	// Message
	parts = append(parts, fmt.Sprintf("Message: %s", e.Message))

	// Timestamp
	parts = append(parts, fmt.Sprintf("Time: %s", e.Timestamp.Format("2006-01-02 15:04:05")))

	// Retryable and Temporary flags
	if e.Retryable || e.Temporary {
		flags := []string{}
		if e.Retryable {
			flags = append(flags, "Retryable")
		}
		if e.Temporary {
			flags = append(flags, "Temporary")
		}
		parts = append(parts, fmt.Sprintf("Flags: %s", strings.Join(flags, ", ")))
	}

	// Context
	if len(e.Context) > 0 {
		parts = append(parts, "Context:")
		for k, v := range e.Context {
			parts = append(parts, fmt.Sprintf("  %s: %v", k, v))
		}
	}

	// Wrapped error
	if e.Err != nil {
		parts = append(parts, fmt.Sprintf("Caused by: %v", e.Err))
	}

	return strings.Join(parts, "\n")
}

// formatJSON returns error as JSON
func formatJSON(e *Error) string {
	data := map[string]interface{}{
		"code":      e.Code.String(),
		"category":  e.Code.Category,
		"message":   e.Message,
		"severity":  e.Severity.String(),
		"timestamp": e.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		"retryable": e.Retryable,
		"temporary": e.Temporary,
	}

	if len(e.Context) > 0 {
		data["context"] = e.Context
	}

	if e.Err != nil {
		data["cause"] = e.Err.Error()
	}

	if len(e.Stack) > 0 {
		stack := make([]map[string]interface{}, len(e.Stack))
		for i, frame := range e.Stack {
			stack[i] = map[string]interface{}{
				"file":     frame.File,
				"line":     frame.Line,
				"function": frame.Function,
			}
		}
		data["stack"] = stack
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal error: %v"}`, err)
	}

	return string(jsonData)
}

// formatWithStack returns error with full stack trace
func formatWithStack(e *Error) string {
	var parts []string

	// Basic error info
	parts = append(parts, formatDetailed(e))

	// Stack trace
	if len(e.Stack) > 0 {
		parts = append(parts, "\nStack trace:")
		for i, frame := range e.Stack {
			parts = append(parts, fmt.Sprintf("  %d. %s", i+1, frame.Function))
			parts = append(parts, fmt.Sprintf("     %s:%d", frame.File, frame.Line))
		}
	}

	return strings.Join(parts, "\n")
}

// MarshalJSON implements json.Marshaler
func (e *Error) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"code":      e.Code.String(),
		"category":  e.Code.Category,
		"message":   e.Message,
		"severity":  e.Severity.String(),
		"timestamp": e.Timestamp,
		"retryable": e.Retryable,
		"temporary": e.Temporary,
	}

	if len(e.Context) > 0 {
		data["context"] = e.Context
	}

	if e.Err != nil {
		data["cause"] = e.Err.Error()
	}

	return json.Marshal(data)
}

// ToMap converts the error to a map for structured logging
func (e *Error) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"error_code":     e.Code.String(),
		"error_category": e.Code.Category,
		"error_message":  e.Message,
		"error_severity": e.Severity.String(),
		"error_time":     e.Timestamp,
		"retryable":      e.Retryable,
		"temporary":      e.Temporary,
	}

	// Add context fields
	for k, v := range e.Context {
		m[fmt.Sprintf("ctx_%s", k)] = v
	}

	// Add cause if present
	if e.Err != nil {
		m["error_cause"] = e.Err.Error()
	}

	return m
}

// ErrorChain returns all errors in the error chain
func ErrorChain(err error) []*Error {
	var chain []*Error

	for err != nil {
		if e, ok := err.(*Error); ok {
			chain = append(chain, e)
			err = e.Err
		} else {
			break
		}
	}

	return chain
}

// RootCause returns the root cause of an error
func RootCause(err error) error {
	for {
		if e, ok := err.(*Error); ok && e.Err != nil {
			err = e.Err
		} else {
			break
		}
	}
	return err
}
