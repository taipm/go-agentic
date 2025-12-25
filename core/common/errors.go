package common

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	// contextFormat is used for appending file/line info to error messages
	contextFormat = " [%s:%d]"
)

// ErrorContext captures file and line information for debugging
type ErrorContext struct {
	File string // Relative path from project root
	Line int    // Line number where error was created
}

// captureContext captures the caller's context (file and line number)
// skip parameter: 0 = captureContext, 1 = immediate caller, 2+ = higher up the stack
func captureContext(skip int) ErrorContext {
	_, file, line, ok := runtime.Caller(skip + 1) // +1 because we're one level deep
	if !ok {
		return ErrorContext{File: "unknown", Line: 0}
	}
	// Get relative path from project root
	parts := strings.Split(file, "/go-agentic/")
	if len(parts) > 1 {
		file = parts[1]
	}
	return ErrorContext{File: file, Line: line}
}

// ErrorType classifies error categories for smart recovery decisions
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeTimeout
	ErrorTypePanic
	ErrorTypeValidation
	ErrorTypeNetwork
	ErrorTypeTemporary
	ErrorTypePermanent
)

// String returns a string representation of the ErrorType
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypePanic:
		return "panic"
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeTemporary:
		return "temporary"
	case ErrorTypePermanent:
		return "permanent"
	default:
		return "unknown"
	}
}

// IsRetryable determines if an error type should trigger a retry
func IsRetryable(errType ErrorType) bool {
	switch errType {
	case ErrorTypeTimeout, ErrorTypeNetwork, ErrorTypeTemporary:
		return true
	default:
		return false
	}
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
	Context ErrorContext
	Err     error
}

func (e *ValidationError) Error() string {
	contextStr := ""
	if e.Context.File != "" && e.Context.File != "unknown" {
		contextStr = fmt.Sprintf(contextFormat, e.Context.File, e.Context.Line)
	}
	if e.Err != nil {
		return fmt.Sprintf("validation error for field '%s': %s (%v)%s", e.Field, e.Message, e.Err, contextStr)
	}
	return fmt.Sprintf("validation error for field '%s': %s%s", e.Field, e.Message, contextStr)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// ExecutionError represents an execution failure
type ExecutionError struct {
	AgentID       string
	TaskID        string
	CorrelationID string       // For distributed request tracing
	Context       ErrorContext // For debugging (file/line info)
	Message       string
	ErrorType     ErrorType
	Err           error
}

func (e *ExecutionError) Error() string {
	contextStr := ""
	if e.Context.File != "" && e.Context.File != "unknown" {
		contextStr = fmt.Sprintf(contextFormat, e.Context.File, e.Context.Line)
	}
	corrStr := ""
	if e.CorrelationID != "" {
		corrStr = fmt.Sprintf(" (corr=%s)", e.CorrelationID)
	}
	if e.Err != nil {
		return fmt.Sprintf("execution error (agent=%s, task=%s, type=%s)%s: %s (%v)%s",
			e.AgentID, e.TaskID, e.ErrorType, corrStr, e.Message, e.Err, contextStr)
	}
	return fmt.Sprintf("execution error (agent=%s, task=%s, type=%s)%s: %s%s",
		e.AgentID, e.TaskID, e.ErrorType, corrStr, e.Message, contextStr)
}

func (e *ExecutionError) Unwrap() error {
	return e.Err
}

// TimeoutError represents a timeout failure
type TimeoutError struct {
	Operation string
	Duration  string
	Context   ErrorContext
	Err       error
}

func (e *TimeoutError) Error() string {
	contextStr := ""
	if e.Context.File != "" && e.Context.File != "unknown" {
		contextStr = fmt.Sprintf(contextFormat, e.Context.File, e.Context.Line)
	}
	if e.Err != nil {
		return fmt.Sprintf("timeout: %s exceeded %s (%v)%s", e.Operation, e.Duration, e.Err, contextStr)
	}
	return fmt.Sprintf("timeout: %s exceeded %s%s", e.Operation, e.Duration, contextStr)
}

func (e *TimeoutError) Unwrap() error {
	return e.Err
}

// QuotaExceededError represents a quota boundary violation
type QuotaExceededError struct {
	QuotaType string
	Limit     interface{}
	Current   interface{}
	Context   ErrorContext
	Err       error
}

func (e *QuotaExceededError) Error() string {
	contextStr := ""
	if e.Context.File != "" && e.Context.File != "unknown" {
		contextStr = fmt.Sprintf(contextFormat, e.Context.File, e.Context.Line)
	}
	if e.Err != nil {
		return fmt.Sprintf("quota exceeded: %s limit=%v current=%v (%v)%s",
			e.QuotaType, e.Limit, e.Current, e.Err, contextStr)
	}
	return fmt.Sprintf("quota exceeded: %s limit=%v current=%v%s",
		e.QuotaType, e.Limit, e.Current, contextStr)
}

func (e *QuotaExceededError) Unwrap() error {
	return e.Err
}

// ConfigurationError represents a configuration error
type ConfigurationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ConfigurationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("configuration error for '%s': %s (%v)", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("configuration error for '%s': %s", e.Field, e.Message)
}

func (e *ConfigurationError) Unwrap() error {
	return e.Err
}

// ProviderError represents a provider communication error
type ProviderError struct {
	Provider string
	Message  string
	Context  ErrorContext
	Err      error
}

func (e *ProviderError) Error() string {
	contextStr := ""
	if e.Context.File != "" && e.Context.File != "unknown" {
		contextStr = fmt.Sprintf(contextFormat, e.Context.File, e.Context.Line)
	}
	if e.Err != nil {
		return fmt.Sprintf("provider error (%s): %s (%v)%s", e.Provider, e.Message, e.Err, contextStr)
	}
	return fmt.Sprintf("provider error (%s): %s%s", e.Provider, e.Message, contextStr)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// ToolExecutionError represents a tool execution failure
type ToolExecutionError struct {
	ToolName string
	Message  string
	Context  ErrorContext
	Err      error
}

func (e *ToolExecutionError) Error() string {
	contextStr := ""
	if e.Context.File != "" && e.Context.File != "unknown" {
		contextStr = fmt.Sprintf(contextFormat, e.Context.File, e.Context.Line)
	}
	if e.Err != nil {
		return fmt.Sprintf("tool execution error (%s): %s (%v)%s", e.ToolName, e.Message, e.Err, contextStr)
	}
	return fmt.Sprintf("tool execution error (%s): %s%s", e.ToolName, e.Message, contextStr)
}

func (e *ToolExecutionError) Unwrap() error {
	return e.Err
}
