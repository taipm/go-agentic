// Package tool provides tool execution and formatting functionality.
package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/taipm/go-agentic/core/common"
)

// ExecuteCallsTimeout is the default timeout for executing tool calls
const ExecuteCallsTimeout = 30 * time.Second

// ToolResult represents the result of executing a tool
type ToolResult struct {
	ToolUseID string
	Content   string
	IsError   bool
}

// ExecuteCalls executes a list of tool calls and returns the results
// This is a placeholder implementation for Phase 3
func ExecuteCalls(ctx context.Context, calls []common.ToolCall) ([]ToolResult, error) {
	var results []ToolResult

	for _, call := range calls {
		result := ToolResult{
			ToolUseID: call.ID,
			Content:   fmt.Sprintf("Executed %s with arguments: %v", call.ToolName, call.Arguments),
			IsError:   false,
		}
		results = append(results, result)
	}

	return results, nil
}

// ExecuteCall executes a single tool call and returns the result
func ExecuteCall(ctx context.Context, call common.ToolCall) (ToolResult, error) {
	// Create a context with timeout
	execCtx, cancel := context.WithTimeout(ctx, ExecuteCallsTimeout)
	defer cancel()

	// Execute the tool with the given arguments
	result, err := SafeExecuteTool(execCtx, call.ToolName, call.Arguments)

	if err != nil {
		return ToolResult{
			ToolUseID: call.ID,
			Content:   fmt.Sprintf("Tool execution failed: %v", err),
			IsError:   true,
		}, err
	}

	return ToolResult{
		ToolUseID: call.ID,
		Content:   result,
		IsError:   false,
	}, nil
}

// SafeExecuteTool executes a tool with panic recovery
func SafeExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	if toolName == "" {
		return "", fmt.Errorf("tool name is empty")
	}

	// This is a placeholder implementation for Phase 3
	// In a full implementation, this would:
	// 1. Validate arguments against tool schema
	// 2. Call the tool's Execute method
	// 3. Handle timeout
	// 4. Recover from panics

	result := fmt.Sprintf("Executed %s with arguments: %v", toolName, arguments)
	return result, nil
}

// FormatToolResults formats tool results for display
func FormatToolResults(results []ToolResult) string {
	if len(results) == 0 {
		return "No tool results"
	}

	var output string
	for _, result := range results {
		if result.IsError {
			output += fmt.Sprintf("[ERROR] Tool %s: %s\n", result.ToolUseID, result.Content)
		} else {
			output += fmt.Sprintf("[OK] Tool %s: %s\n", result.ToolUseID, result.Content)
		}
	}
	return output
}
