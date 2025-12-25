package common

import (
	"fmt"
	"time"
)

// ============================================================================
// VALIDATION ERROR HELPERS
// ============================================================================

// NewValidationError creates a validation error with automatic context capture
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
		Context: captureContext(1),
	}
}

// NewValidationErrorf creates a validation error with formatted message
func NewValidationErrorf(field, format string, args ...interface{}) error {
	return &ValidationError{
		Field:   field,
		Message: stringf(format, args...),
		Context: captureContext(1),
	}
}

// NewValidationErrorWithCause creates a validation error wrapping another error
func NewValidationErrorWithCause(field, message string, cause error) error {
	return &ValidationError{
		Field:   field,
		Message: message,
		Context: captureContext(1),
		Err:     cause,
	}
}

// NewMissingFieldError creates a validation error for missing required field
func NewMissingFieldError(field string) error {
	return NewValidationError(field, "field is required")
}

// NewInvalidValueError creates a validation error for invalid field value
func NewInvalidValueError(field, value string) error {
	return NewValidationErrorf(field, "invalid value: %s", value)
}

// ============================================================================
// EXECUTION ERROR HELPERS
// ============================================================================

// NewExecutionError creates an execution error with automatic context capture
func NewExecutionError(agentID, message string, errType ErrorType, cause error) error {
	return &ExecutionError{
		AgentID:   agentID,
		Message:   message,
		ErrorType: errType,
		Context:   captureContext(1),
		Err:       cause,
	}
}

// NewExecutionErrorWithCorrelation creates an execution error with correlation ID
func NewExecutionErrorWithCorrelation(correlationID, agentID, message string, errType ErrorType, cause error) error {
	return &ExecutionError{
		AgentID:       agentID,
		CorrelationID: correlationID,
		Message:       message,
		ErrorType:     errType,
		Context:       captureContext(1),
		Err:           cause,
	}
}

// NewAgentExecutionError creates an execution error for agent failure
func NewAgentExecutionError(agentID, message string, cause error) error {
	return NewExecutionError(agentID, message, ErrorTypePermanent, cause)
}

// NewAgentExecutionErrorf creates a formatted agent execution error
func NewAgentExecutionErrorf(agentID, format string, args ...interface{}) error {
	return NewExecutionError(agentID, stringf(format, args...), ErrorTypePermanent, nil)
}

// NewTaskExecutionError creates an execution error for task failure
func NewTaskExecutionError(agentID, taskID, message string, cause error) error {
	err := &ExecutionError{
		AgentID:   agentID,
		TaskID:    taskID,
		Message:   message,
		ErrorType: ErrorTypePermanent,
		Context:   captureContext(1),
		Err:       cause,
	}
	return err
}

// ============================================================================
// TIMEOUT ERROR HELPERS
// ============================================================================

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string, duration time.Duration) error {
	return &TimeoutError{
		Operation: operation,
		Duration:  duration.String(),
		Context:   captureContext(1),
	}
}

// NewTimeoutErrorWithCause creates a timeout error wrapping another error
func NewTimeoutErrorWithCause(operation string, duration time.Duration, cause error) error {
	return &TimeoutError{
		Operation: operation,
		Duration:  duration.String(),
		Context:   captureContext(1),
		Err:       cause,
	}
}

// NewTimeoutErrorf creates a timeout error with custom duration string
func NewTimeoutErrorf(operation, duration string) error {
	return &TimeoutError{
		Operation: operation,
		Duration:  duration,
		Context:   captureContext(1),
	}
}

// ============================================================================
// QUOTA EXCEEDED ERROR HELPERS
// ============================================================================

// NewQuotaExceededError creates a quota exceeded error
func NewQuotaExceededError(quotaType string, limit, current interface{}) error {
	return &QuotaExceededError{
		QuotaType: quotaType,
		Limit:     limit,
		Current:   current,
		Context:   captureContext(1),
	}
}

// NewQuotaExceededErrorWithCause creates a quota error wrapping another error
func NewQuotaExceededErrorWithCause(quotaType string, limit, current interface{}, cause error) error {
	return &QuotaExceededError{
		QuotaType: quotaType,
		Limit:     limit,
		Current:   current,
		Context:   captureContext(1),
		Err:       cause,
	}
}

// NewTokensExceededError creates an error for token quota exceeded
func NewTokensExceededError(limit, current int) error {
	return NewQuotaExceededError("tokens", limit, current)
}

// NewCostExceededError creates an error for cost quota exceeded
func NewCostExceededError(limit, current float64) error {
	return NewQuotaExceededError("cost", limit, current)
}

// ============================================================================
// PROVIDER ERROR HELPERS
// ============================================================================

// NewProviderError creates a provider error
func NewProviderError(provider, message string, cause error) error {
	return &ProviderError{
		Provider: provider,
		Message:  message,
		Context:  captureContext(1),
		Err:      cause,
	}
}

// NewProviderErrorf creates a formatted provider error
func NewProviderErrorf(provider, format string, args ...interface{}) error {
	return &ProviderError{
		Provider: provider,
		Message:  stringf(format, args...),
		Context:  captureContext(1),
	}
}

// NewOpenAIError creates an OpenAI provider error
func NewOpenAIError(message string, cause error) error {
	return NewProviderError("openai", message, cause)
}

// NewOllamaError creates an Ollama provider error
func NewOllamaError(message string, cause error) error {
	return NewProviderError("ollama", message, cause)
}

// ============================================================================
// TOOL EXECUTION ERROR HELPERS
// ============================================================================

// NewToolError creates a tool execution error
func NewToolError(toolName, message string, cause error) error {
	return &ToolExecutionError{
		ToolName: toolName,
		Message:  message,
		Context:  captureContext(1),
		Err:      cause,
	}
}

// NewToolErrorf creates a formatted tool execution error
func NewToolErrorf(toolName, format string, args ...interface{}) error {
	return &ToolExecutionError{
		ToolName: toolName,
		Message:  stringf(format, args...),
		Context:  captureContext(1),
	}
}

// NewToolTimeoutError creates a tool timeout error
func NewToolTimeoutError(toolName string, duration time.Duration) error {
	return &ToolExecutionError{
		ToolName: toolName,
		Message:  "tool execution exceeded " + duration.String(),
		Context:  captureContext(1),
	}
}

// ============================================================================
// ROUTING ERROR HELPERS
// ============================================================================

// NewRoutingError creates a routing validation error
func NewRoutingError(message string) error {
	return NewValidationError("routing", message)
}

// NewRoutingErrorf creates a formatted routing error
func NewRoutingErrorf(format string, args ...interface{}) error {
	return NewValidationErrorf("routing", format, args...)
}

// NewAgentNotFoundError creates an error for missing agent
func NewAgentNotFoundError(agentID string) error {
	return NewValidationErrorf("agent_id", "agent '%s' not found", agentID)
}

// NewBehaviorNotFoundError creates an error for missing behavior
func NewBehaviorNotFoundError(behavior string) error {
	return NewValidationErrorf("behavior", "behavior '%s' not found", behavior)
}

// ============================================================================
// CONFIGURATION ERROR HELPERS
// ============================================================================

// NewConfigError creates a configuration error
func NewConfigError(field, message string) error {
	return &ConfigurationError{
		Field:   field,
		Message: message,
		Err:     nil,
	}
}

// NewConfigErrorf creates a formatted configuration error
func NewConfigErrorf(field, format string, args ...interface{}) error {
	return &ConfigurationError{
		Field:   field,
		Message: stringf(format, args...),
	}
}

// ============================================================================
// HELPER UTILITIES
// ============================================================================

// stringf is a string formatter helper
func stringf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
