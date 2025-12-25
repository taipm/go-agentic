package common

import "fmt"

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
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("validation error for field '%s': %s (%v)", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// ExecutionError represents an execution failure
type ExecutionError struct {
	AgentID   string
	TaskID    string
	Message   string
	ErrorType ErrorType
	Err       error
}

func (e *ExecutionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("execution error (agent=%s, task=%s, type=%s): %s (%v)",
			e.AgentID, e.TaskID, e.ErrorType, e.Message, e.Err)
	}
	return fmt.Sprintf("execution error (agent=%s, task=%s, type=%s): %s",
		e.AgentID, e.TaskID, e.ErrorType, e.Message)
}

func (e *ExecutionError) Unwrap() error {
	return e.Err
}

// TimeoutError represents a timeout failure
type TimeoutError struct {
	Operation string
	Duration  string
	Err       error
}

func (e *TimeoutError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("timeout: %s exceeded %s (%v)", e.Operation, e.Duration, e.Err)
	}
	return fmt.Sprintf("timeout: %s exceeded %s", e.Operation, e.Duration)
}

func (e *TimeoutError) Unwrap() error {
	return e.Err
}

// QuotaExceededError represents a quota boundary violation
type QuotaExceededError struct {
	QuotaType string
	Limit     interface{}
	Current   interface{}
	Err       error
}

func (e *QuotaExceededError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("quota exceeded: %s limit=%v current=%v (%v)",
			e.QuotaType, e.Limit, e.Current, e.Err)
	}
	return fmt.Sprintf("quota exceeded: %s limit=%v current=%v",
		e.QuotaType, e.Limit, e.Current)
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
	Err      error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("provider error (%s): %s (%v)", e.Provider, e.Message, e.Err)
	}
	return fmt.Sprintf("provider error (%s): %s", e.Provider, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// ToolExecutionError represents a tool execution failure
type ToolExecutionError struct {
	ToolName string
	Message  string
	Err      error
}

func (e *ToolExecutionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("tool execution error (%s): %s (%v)", e.ToolName, e.Message, e.Err)
	}
	return fmt.Sprintf("tool execution error (%s): %s", e.ToolName, e.Message)
}

func (e *ToolExecutionError) Unwrap() error {
	return e.Err
}
