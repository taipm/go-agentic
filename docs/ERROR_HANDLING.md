# Error Handling Guidelines

This document provides comprehensive guidance on error handling patterns in the go-agentic framework.

## Overview

The go-agentic framework uses typed errors instead of plain strings for better error handling, debugging, and observability. All error types include automatic context capture (file:line information) and support for distributed request tracing via correlation IDs.

## Error Types

### 1. ValidationError

Used for validation failures on input data or configuration.

**When to use:**
- Missing required fields
- Invalid field values
- Constraint violations
- Configuration errors

**Example:**
```go
if agent == nil {
    return nil, common.NewValidationError("agent", "agent cannot be nil")
}

if agentID == "" {
    return nil, common.NewMissingFieldError("agent_id")
}

if value < 0 {
    return nil, common.NewInvalidValueError("count", "-5")
}
```

**Error Message Format:**
```
validation error for field 'agent_id': field is required [core/routing/parallel.go:120]
```

---

### 2. ExecutionError

Used for failures during agent or workflow execution.

**When to use:**
- Agent execution failures
- Task execution failures
- Workflow engine errors
- Recoverable and non-recoverable failures

**Error Types:**
- `ErrorTypeTimeout` - Execution exceeded time limit
- `ErrorTypePanic` - Execution panicked
- `ErrorTypeNetwork` - Network-related failure
- `ErrorTypeTemporary` - Retryable error
- `ErrorTypePermanent` - Non-retryable error

**Example:**
```go
// Simple execution error
return nil, common.NewAgentExecutionError(agent.ID, "agent failed to complete", err)

// With correlation ID for distributed tracing
return nil, common.NewExecutionErrorWithCorrelation(
    correlationID,
    agent.ID,
    "failed to invoke LLM provider",
    common.ErrorTypePermanent,
    err,
)

// Formatted message
return nil, common.NewAgentExecutionErrorf(agent.ID, "agent '%s' timed out after 30s", agent.ID)
```

**Error Message Format:**
```
execution error (agent=agent-1, task=task-5, type=permanent)(corr=550e8400-e29b-41d4-a716-446655440000): failed to invoke LLM provider (original error) [core/agent/execution.go:245]
```

---

### 3. TimeoutError

Used when operations exceed their time limit.

**When to use:**
- Agent execution timeout
- Tool execution timeout
- API call timeout
- Workflow timeout

**Example:**
```go
if timeExceeded {
    return nil, common.NewTimeoutError("agent execution", 30*time.Second)
}

// With wrapped cause
return nil, common.NewTimeoutErrorWithCause("api call", duration, originalErr)
```

**Error Message Format:**
```
timeout: agent execution exceeded 30s [core/workflow/execution.go:150]
```

---

### 4. QuotaExceededError

Used when resource limits are exceeded.

**When to use:**
- Token limits exceeded
- Cost limits exceeded
- Max rounds exceeded
- Max handoffs exceeded
- Rate limit exceeded

**Example:**
```go
// Token quota
return nil, common.NewTokensExceededError(100000, 150000)

// Cost quota
return nil, common.NewCostExceededError(10.0, 12.5)

// Generic quota
return nil, common.NewQuotaExceededError("max_rounds", maxRounds, currentRound)
```

**Error Message Format:**
```
quota exceeded: tokens limit=100000 current=150000 [core/workflow/execution.go:280]
```

---

### 5. ProviderError

Used for LLM provider communication failures.

**When to use:**
- OpenAI API errors
- Ollama API errors
- Provider connection failures
- Provider authentication failures

**Example:**
```go
return nil, common.NewOpenAIError("failed to generate completion", apiErr)
return nil, common.NewOllamaError("ollama server unreachable", connectionErr)
return nil, common.NewProviderErrorf("azure", "authentication failed: %v", err)
```

**Error Message Format:**
```
provider error (openai): failed to generate completion (api error details) [core/providers/openai/provider.go:180]
```

---

### 6. ToolExecutionError

Used for tool execution failures.

**When to use:**
- Tool invocation failures
- Tool timeout
- Tool execution errors

**Example:**
```go
return nil, common.NewToolError(toolName, "tool execution failed", err)
return nil, common.NewToolTimeoutError(toolName, 30*time.Second)
```

**Error Message Format:**
```
tool execution error (database_query): tool execution failed (underlying error) [core/tools/executor.go:95]
```

---

### 7. RoutingError

Used for workflow routing and agent selection failures.

**When to use:**
- Agent not found
- Behavior not found
- Signal routing failures
- Invalid routing configuration

**Example:**
```go
return "", common.NewAgentNotFoundError("reporter")
return "", common.NewBehaviorNotFoundError("analysis")
return "", common.NewRoutingErrorf("signal '%s' not configured", signalName)
```

**Error Message Format:**
```
validation error for field 'agent_id': agent 'reporter' not found [core/routing/parallel.go:85]
```

---

### 8. ConfigurationError

Used for configuration file or setup errors.

**When to use:**
- Invalid configuration structure
- Missing configuration values
- Configuration parsing errors

**Example:**
```go
return common.NewConfigError("agents", "agents section is required")
return common.NewConfigErrorf("agent[0]", "name is required but missing")
```

**Error Message Format:**
```
configuration error for 'agents': agents section is required
```

---

## Error Helper Functions

### Validation Helpers

```go
// Create validation error
NewValidationError(field, message string) error

// Create validation error with format
NewValidationErrorf(field, format string, args ...interface{}) error

// Create validation error wrapping a cause
NewValidationErrorWithCause(field, message string, cause error) error

// Create error for missing field
NewMissingFieldError(field string) error

// Create error for invalid value
NewInvalidValueError(field, value string) error
```

---

### Execution Helpers

```go
// Create execution error with error type
NewExecutionError(agentID, message string, errType ErrorType, cause error) error

// Create execution error with correlation ID (for distributed tracing)
NewExecutionErrorWithCorrelation(correlationID, agentID, message string, errType ErrorType, cause error) error

// Create agent execution error
NewAgentExecutionError(agentID, message string, cause error) error

// Create agent execution error with format
NewAgentExecutionErrorf(agentID, format string, args ...interface{}) error

// Create task execution error
NewTaskExecutionError(agentID, taskID, message string, cause error) error
```

---

### Timeout Helpers

```go
// Create timeout error
NewTimeoutError(operation string, duration time.Duration) error

// Create timeout error with cause
NewTimeoutErrorWithCause(operation string, duration time.Duration, cause error) error

// Create timeout error with custom duration string
NewTimeoutErrorf(operation, duration string) error
```

---

### Quota Helpers

```go
// Create generic quota exceeded error
NewQuotaExceededError(quotaType string, limit, current interface{}) error

// Create quota error with cause
NewQuotaExceededErrorWithCause(quotaType string, limit, current interface{}, cause error) error

// Create token quota error
NewTokensExceededError(limit, current int) error

// Create cost quota error
NewCostExceededError(limit, current float64) error
```

---

### Provider Helpers

```go
// Create provider error
NewProviderError(provider, message string, cause error) error

// Create provider error with format
NewProviderErrorf(provider, format string, args ...interface{}) error

// Create OpenAI error
NewOpenAIError(message string, cause error) error

// Create Ollama error
NewOllamaError(message string, cause error) error
```

---

### Tool Helpers

```go
// Create tool error
NewToolError(toolName, message string, cause error) error

// Create tool error with format
NewToolErrorf(toolName, format string, args ...interface{}) error

// Create tool timeout error
NewToolTimeoutError(toolName string, duration time.Duration) error
```

---

### Routing Helpers

```go
// Create routing error
NewRoutingError(message string) error

// Create routing error with format
NewRoutingErrorf(format string, args ...interface{}) error

// Create agent not found error
NewAgentNotFoundError(agentID string) error

// Create behavior not found error
NewBehaviorNotFoundError(behavior string) error
```

---

## Error Handling Patterns

### 1. Type Assertion with errors.As()

The recommended way to handle typed errors:

```go
import "errors"

err := someFunction()

var validationErr *common.ValidationError
if errors.As(err, &validationErr) {
    log.Printf("Validation failed for field '%s': %s", validationErr.Field, validationErr.Message)
    // Handle validation error
}

var executionErr *common.ExecutionError
if errors.As(err, &executionErr) {
    if common.IsRetryable(executionErr.ErrorType) {
        // Retry the operation
    } else {
        // Log and fail permanently
    }
}
```

---

### 2. Error Unwrapping

All typed errors support error wrapping using Go's standard error interface:

```go
// All typed errors implement Unwrap()
err := common.NewExecutionError(agentID, "execution failed", common.ErrorTypePermanent, originalErr)

// Unwrap the original error
if unwrapped := errors.Unwrap(err); unwrapped != nil {
    log.Printf("Original error: %v", unwrapped)
}

// Using errors.Is() to check for specific wrapped errors
if errors.Is(err, io.EOF) {
    // Handle EOF specifically
}
```

---

### 3. Error Context Capture

All typed errors automatically capture file and line information:

```go
// No need to manually capture context - it's automatic!
return common.NewValidationError("field", "invalid value")

// Error message will include file:line automatically:
// validation error for field 'field': invalid value [core/routing/parallel.go:250]
```

---

### 4. Correlation IDs for Distributed Tracing

For distributed systems, correlation IDs enable request tracing:

```go
// At workflow entry point, generate correlation ID
correlationID := uuid.New().String()

// Thread through execution context
execCtx := &ExecutionContext{
    CorrelationID: correlationID,
    // ...
}

// Include in error when creating
err := common.NewExecutionErrorWithCorrelation(
    correlationID,
    agentID,
    "agent execution failed",
    common.ErrorTypePermanent,
    nil,
)

// Error message includes correlation ID:
// execution error (agent=agent-1, task=, type=permanent)(corr=550e8400-e29b-41d4-a716-446655440000): ...
```

---

## Error Decision Tree

Use this decision tree to determine which error type to use:

```
Is it a validation failure?
├─ YES → ValidationError (use helpers: NewValidationError, NewMissingFieldError, etc.)
└─ NO

Did execution timeout?
├─ YES → TimeoutError
└─ NO

Is it a resource limit exceeded?
├─ YES → QuotaExceededError
└─ NO

Is it a provider (LLM) failure?
├─ YES → ProviderError
└─ NO

Is it a tool execution failure?
├─ YES → ToolExecutionError
└─ NO

Is it a routing/agent selection failure?
├─ YES → RoutingError (agent not found, behavior not found)
└─ NO

Is it an agent or workflow execution failure?
├─ YES → ExecutionError
└─ NO

Is it a configuration error?
├─ YES → ConfigurationError
└─ NO
```

---

## Best Practices

### 1. Use Error Helpers

Always use provided error helper functions instead of creating error types directly:

```go
// GOOD
return nil, common.NewValidationError("agent_id", "agent not found")

// ALSO GOOD (when you need control over error details)
return nil, &common.ValidationError{
    Field:   "agent_id",
    Message: "agent not found",
    Context: common.captureContext(0),
}

// AVOID
return nil, fmt.Errorf("validation error for field 'agent_id': agent not found")
```

---

### 2. Include Relevant Context

Error messages should include enough context to debug the issue:

```go
// GOOD - includes specific details
return nil, common.NewAgentNotFoundError(agentID)

// AVOID - too generic
return nil, common.NewValidationError("agent", "not found")
```

---

### 3. Wrap Root Causes

When an error is caused by another error, wrap it:

```go
if err != nil {
    return nil, common.NewExecutionError(
        agentID,
        "failed to invoke provider",
        common.ErrorTypePermanent,
        err,  // Wrap the original error
    )
}
```

---

### 4. Use Correlation IDs

For distributed systems, propagate correlation IDs through the execution chain:

```go
func executeWorkflow(execCtx *ExecutionContext) error {
    // Correlation ID is already set when ExecutionContext is created
    correlationID := execCtx.CorrelationID

    // Use it when creating errors
    if err != nil {
        return common.NewExecutionErrorWithCorrelation(
            correlationID,
            agentID,
            "execution failed",
            common.ErrorTypePermanent,
            err,
        )
    }
}
```

---

### 5. Check Retryability

Use `IsRetryable()` to determine if an error should trigger a retry:

```go
var executionErr *common.ExecutionError
if errors.As(err, &executionErr) {
    if common.IsRetryable(executionErr.ErrorType) {
        // Retry the operation
        return retryOperation()
    } else {
        // Log and fail
        return err
    }
}
```

---

## Error Type vs. Error Message

- **Error Type** (ValidationError, ExecutionError, etc.) determines how the error is handled programmatically
- **Error Message** provides human-readable details
- **Error Context** (file:line) helps locate where the error was created

```go
// All three work together:
err := common.NewValidationError("agent_id", "agent 'reporter' not found")

// Type: ValidationError → Handle as validation failure
// Message: "agent 'reporter' not found" → Understand the issue
// Context: "core/routing/parallel.go:85" → Debug location
```

---

## Migration from fmt.Errorf

If you find `fmt.Errorf` in the codebase, migrate it to typed errors:

```go
// BEFORE
return nil, fmt.Errorf("agent '%s' not found", agentID)

// AFTER
return nil, common.NewAgentNotFoundError(agentID)

// BEFORE
return nil, fmt.Errorf("execution failed: %v", err)

// AFTER
return nil, common.NewExecutionError(agentID, "execution failed", common.ErrorTypePermanent, err)

// BEFORE
return nil, fmt.Errorf("validation error for field '%s': %s", field, message)

// AFTER
return nil, common.NewValidationError(field, message)
```

---

## Testing with Typed Errors

Use `errors.As()` to assert on error types in tests:

```go
import "errors"

func TestAgentNotFound(t *testing.T) {
    _, err := routing.DetermineNextAgent(nil, response, routingConfig)

    var validationErr *common.ValidationError
    if !errors.As(err, &validationErr) {
        t.Errorf("Expected ValidationError, got %T", err)
    }

    if validationErr.Field != "agent" {
        t.Errorf("Expected field 'agent', got '%s'", validationErr.Field)
    }
}
```

---

## Summary

The go-agentic error handling system provides:

✅ **Typed Errors** - Structured error handling instead of plain strings
✅ **Automatic Context** - File:line information captured automatically
✅ **Helper Functions** - Reduced boilerplate with constructor functions
✅ **Correlation IDs** - Request tracing for distributed systems
✅ **Error Wrapping** - Proper error cause tracking with Unwrap()
✅ **Retryability** - Built-in classification of retryable errors

This enables better debugging, error recovery, and observability throughout the system.
