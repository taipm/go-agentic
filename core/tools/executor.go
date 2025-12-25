// Package tools provides tool execution utilities including error handling,
// validation, timeout management, and orchestration of tool calls.
package tools

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/logging"
)

// ExecuteTool executes a single tool with the provided arguments and returns the result.
// It handles multiple tool format types (pointer or by-value common.Tool).
func ExecuteTool(ctx context.Context, toolName string, tool interface{}, args map[string]interface{}) (string, error) {
	// Validate inputs
	if tool == nil {
		return "", fmt.Errorf("tool is nil for '%s'", toolName)
	}
	if toolName == "" {
		return "", fmt.Errorf("tool name is empty")
	}

	// Type assert to common.Tool (try pointer first)
	var commonTool *common.Tool
	switch t := tool.(type) {
	case *common.Tool:
		commonTool = t
	case common.Tool:
		commonTool = &t
	default:
		return "", fmt.Errorf("tool '%s' is not common.Tool type: got %T", toolName, tool)
	}

	// Validate tool
	if commonTool == nil {
		return "", fmt.Errorf("tool '%s' pointer is nil", toolName)
	}

	// Extract and validate function
	handler, ok := commonTool.Func.(ToolHandler)
	if !ok {
		return "", fmt.Errorf("tool '%s' does not have ToolHandler function: got %T", toolName, commonTool.Func)
	}
	if handler == nil {
		return "", fmt.Errorf("tool '%s' has nil ToolHandler function", toolName)
	}

	// Execute with retry logic
	config := DefaultRetryConfig()
	result, err := ExecuteWithRetry(ctx, toolName, handler, args, config)
	if err != nil {
		return "", fmt.Errorf("tool '%s' execution failed: %w", toolName, err)
	}

	return result, nil
}

// ExecuteToolCalls executes all tool calls from an agent response and collects results.
// It continues execution even if individual tools fail (partial failure tolerance).
// Returns a map of tool name to result, and an error if any tools failed.
func ExecuteToolCalls(ctx context.Context, toolCalls []common.ToolCall, agentTools []interface{}) (map[string]string, error) {
	results := make(map[string]string)

	// Handle empty tool calls
	if len(toolCalls) == 0 {
		return results, nil
	}

	// Build tool map for efficient lookup
	toolMap := buildToolMap(agentTools)
	var executionErrors []string

	// Execute each tool call
	for _, call := range toolCalls {
		// Find tool in agent's tool list
		tool, exists := toolMap[call.ToolName]
		if !exists {
			errMsg := fmt.Sprintf("tool '%s' not found in agent tools", call.ToolName)
			executionErrors = append(executionErrors, errMsg)
			log.Printf("[TOOL] %s", errMsg)
			continue
		}

		// Log: tool.start
		logging.GetLogger().InfoContext(ctx, "tool.start",
			slog.String("event", "tool.start"),
			slog.String("trace_id", logging.GetTraceID(ctx)),
			slog.String("tool_name", call.ToolName),
		)

		// Execute the tool
		result, err := ExecuteTool(ctx, call.ToolName, tool, call.Arguments)
		if err != nil {
			errMsg := fmt.Sprintf("tool '%s' failed: %v", call.ToolName, err)
			executionErrors = append(executionErrors, errMsg)
			log.Printf("[TOOL] %s", errMsg)

			// Log: tool.error
			logging.GetLogger().InfoContext(ctx, "tool.error",
				slog.String("event", "tool.error"),
				slog.String("trace_id", logging.GetTraceID(ctx)),
				slog.String("tool_name", call.ToolName),
				slog.String("error", err.Error()),
			)
			continue
		}

		// Log: tool.end
		logging.GetLogger().InfoContext(ctx, "tool.end",
			slog.String("event", "tool.end"),
			slog.String("trace_id", logging.GetTraceID(ctx)),
			slog.String("tool_name", call.ToolName),
			slog.String("status", "success"),
		)

		// Store successful result
		results[call.ToolName] = result
	}

	// Return results with any accumulated errors
	if len(executionErrors) > 0 {
		return results, fmt.Errorf("tool execution errors: %v", executionErrors)
	}

	return results, nil
}

// buildToolMap creates a map of tool names to tool objects for efficient lookup.
// Handles both pointer and by-value common.Tool types.
func buildToolMap(agentTools []interface{}) map[string]interface{} {
	toolMap := make(map[string]interface{})

	for _, tool := range agentTools {
		if tool == nil {
			continue
		}

		// Handle pointer type
		if toolPtr, ok := tool.(*common.Tool); ok {
			if toolPtr.Name != "" {
				toolMap[toolPtr.Name] = toolPtr
			}
			continue
		}

		// Handle by-value type
		if toolVal, ok := tool.(common.Tool); ok {
			if toolVal.Name != "" {
				toolMap[toolVal.Name] = toolVal
			}
			continue
		}

		// Log warning for unknown tool type
		log.Printf("[TOOL] Unknown tool type in agent tools: %T", tool)
	}

	return toolMap
}

// FormatToolResults formats tool execution results into a readable message
// suitable for adding to conversation history.
func FormatToolResults(results map[string]string) string {
	if len(results) == 0 {
		return "No tool results available"
	}

	message := "Tool Results:\n"
	for toolName, result := range results {
		message += fmt.Sprintf("- %s: %s\n", toolName, result)
	}

	return message
}

// FindToolByName searches for a tool by name in the agent's tool list.
// Returns a pointer to the tool if found, or nil if not found.
func FindToolByName(agentTools []interface{}, toolName string) *common.Tool {
	if toolName == "" {
		return nil
	}

	for _, tool := range agentTools {
		if tool == nil {
			continue
		}

		// Check pointer type
		if toolPtr, ok := tool.(*common.Tool); ok {
			if toolPtr.Name == toolName {
				return toolPtr
			}
			continue
		}

		// Check by-value type
		if toolVal, ok := tool.(common.Tool); ok {
			if toolVal.Name == toolName {
				return &toolVal
			}
			continue
		}
	}

	return nil
}

// ValidateToolCall validates that a tool call has all required fields
// and that the tool exists in the agent's tool list.
func ValidateToolCall(toolCall common.ToolCall, agentTools []interface{}) error {
	if toolCall.ToolName == "" {
		return fmt.Errorf("tool call has empty tool name")
	}

	if toolCall.Arguments == nil {
		toolCall.Arguments = make(map[string]interface{})
	}

	// Check if tool exists
	tool := FindToolByName(agentTools, toolCall.ToolName)
	if tool == nil {
		return fmt.Errorf("tool '%s' not found in agent tools", toolCall.ToolName)
	}

	return nil
}
