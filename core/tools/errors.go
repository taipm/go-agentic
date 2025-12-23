// Package tools provides tool execution utilities including error handling,
// validation, and timeout management.
package tools

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// ErrorType classifies tool execution errors for smart recovery decisions
type ErrorType int

const (
	ErrorTypeUnknown    ErrorType = iota
	ErrorTypeTimeout              // Transient: exceeded deadline
	ErrorTypePanic                // Non-transient: panic in tool
	ErrorTypeValidation           // Non-transient: invalid arguments
	ErrorTypeNetwork              // Transient: connection issues
	ErrorTypeTemporary            // Transient: temporary failures
	ErrorTypePermanent            // Non-transient: permanent failure
)

// ClassifyError determines if an error is transient (retryable) or permanent
func ClassifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errStr := err.Error()

	// Timeout errors are transient
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorTypeTimeout
	}

	// Check for panic signature
	if strings.Contains(errStr, "panicked:") {
		return ErrorTypePanic
	}

	// Check for validation errors (non-transient)
	if strings.Contains(errStr, "required field") || strings.Contains(errStr, "parameter") {
		return ErrorTypeValidation
	}

	// Check for network-like errors (transient)
	networkPatterns := []string{
		"connection refused", "connection reset", "broken pipe",
		"network unreachable", "host unreachable", "no such host",
		"temporary failure", "i/o timeout",
	}
	for _, pattern := range networkPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return ErrorTypeNetwork
		}
	}

	// Default to temporary (transient) for unknown errors
	return ErrorTypeTemporary
}

// IsRetryable determines if an error type should trigger a retry
func IsRetryable(errType ErrorType) bool {
	switch errType {
	case ErrorTypeTimeout, ErrorTypeNetwork, ErrorTypeTemporary:
		return true
	case ErrorTypePanic, ErrorTypeValidation, ErrorTypePermanent, ErrorTypeUnknown:
		return false
	default:
		return false
	}
}

// CalculateBackoffDuration returns exponential backoff duration
// Start with 100ms, double each attempt: 100ms, 200ms, 400ms, 800ms...
// Capped at 5 seconds
func CalculateBackoffDuration(attempt int) time.Duration {
	baseDelay := time.Duration(100<<uint(attempt)) * time.Millisecond

	if baseDelay > 5*time.Second {
		baseDelay = 5 * time.Second
	}

	return baseDelay
}

// ToolHandler is the function signature for tool execution
type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries int
	Validator  func(tool interface{}, args map[string]interface{}) error
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 2, // 3 total attempts
	}
}

// ExecuteWithRetry executes a tool handler with exponential backoff retry logic
func ExecuteWithRetry(ctx context.Context, toolName string, handler ToolHandler, args map[string]interface{}, config RetryConfig) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		output, err := executeOnce(ctx, toolName, handler, args)

		if err == nil {
			if attempt > 0 {
				log.Printf("[TOOL RETRY] %s succeeded on attempt %d", toolName, attempt+1)
			}
			return output, nil
		}

		lastErr = err
		errType := ClassifyError(err)

		// If non-retryable, return immediately
		if !IsRetryable(errType) {
			log.Printf("[TOOL ERROR] %s failed with non-retryable error: %v", toolName, err)
			return "", err
		}

		// If this was the last attempt, return the error
		if attempt == config.MaxRetries {
			log.Printf("[TOOL ERROR] %s failed after %d retries: %v", toolName, config.MaxRetries+1, err)
			return "", err
		}

		// Calculate backoff and wait before retry
		backoff := CalculateBackoffDuration(attempt)
		log.Printf("[TOOL RETRY] %s failed on attempt %d (type: %v), retrying in %v: %v",
			toolName, attempt+1, errType, backoff, err)

		select {
		case <-ctx.Done():
			log.Printf("[TOOL RETRY] %s context cancelled during backoff", toolName)
			return "", ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	return "", lastErr
}

// executeOnce executes a tool once with panic recovery
func executeOnce(ctx context.Context, toolName string, handler ToolHandler, args map[string]interface{}) (output string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tool %s panicked: %v", toolName, r)
		}
	}()

	return handler(ctx, args)
}
