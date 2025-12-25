// Package agent provides agent execution and related functionality.
package agent

import (
	"fmt"
	"strings"

	"github.com/taipm/go-agentic/core/common"
)

// ParseToolArguments splits tool arguments respecting nested brackets
// Handles cases like: collection_name, [1.0, 2.0, 3.0], 5
func ParseToolArguments(argsStr string) []string {
	var parts []string
	var current strings.Builder
	bracketDepth := 0
	parenDepth := 0

	for _, ch := range argsStr {
		switch ch {
		case '[':
			bracketDepth++
			current.WriteRune(ch)
		case ']':
			bracketDepth--
			current.WriteRune(ch)
		case '(':
			parenDepth++
			current.WriteRune(ch)
		case ')':
			parenDepth--
			current.WriteRune(ch)
		case ',':
			if bracketDepth == 0 && parenDepth == 0 {
				// This is a top-level comma, so split here
				part := strings.TrimSpace(current.String())
				if part != "" {
					parts = append(parts, part)
				}
				current.Reset()
			} else {
				// Comma is inside brackets, keep it
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	// Add last part
	if part := strings.TrimSpace(current.String()); part != "" {
		parts = append(parts, part)
	}

	return parts
}

// ExtractToolCallsFromText extracts tool calls from the response text
// This is a fallback method for models without tool_use support
func ExtractToolCallsFromText(text string, agent *common.Agent) []common.ToolCall {
	var calls []common.ToolCall

	// Get valid tool names from agent tools
	validToolNames := make(map[string]bool)
	for _, toolObj := range agent.Tools {
		// Handle interface{} type
		if toolObj != nil {
			validToolNames[fmt.Sprintf("%v", toolObj)] = true
		}
	}

	// Look for patterns like: ToolName(...)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to find tool calls in this line
		for toolName := range validToolNames {
			if strings.Contains(line, toolName+"(") {
				// Extract the arguments
				startIdx := strings.Index(line, toolName+"(")
				if startIdx != -1 {
					endIdx := strings.Index(line[startIdx:], ")")
					if endIdx != -1 {
						endIdx += startIdx
						argsStr := line[startIdx+len(toolName)+1 : endIdx]

						// Parse arguments - handle nested arrays/objects
						args := make(map[string]interface{})
						if argsStr != "" {
							// Split arguments respecting nested brackets
							argParts := ParseToolArguments(argsStr)

							for i, part := range argParts {
								part = strings.TrimSpace(part)
								part = strings.Trim(part, `"'`)

								// Use arg0, arg1, etc. for positional arguments
								args[fmt.Sprintf("arg%d", i)] = part
							}
						}

						calls = append(calls, common.ToolCall{
							ID:        fmt.Sprintf("%s_%d", toolName, len(calls)),
							ToolName:  toolName,
							Arguments: args,
						})
					}
				}
			}
		}
	}

	return calls
}

// GetToolParameterNames extracts parameter names from tool definition in order
// Tool parameter interface for compatibility
type toolLike interface {
	GetParameters() map[string]interface{}
}

func GetToolParameterNames(toolObj interface{}) []string {
	var paramNames []string

	// Handle map[string]interface{} (for tool parameters)
	params, ok := toolObj.(map[string]interface{})
	if !ok {
		return paramNames
	}

	// Extract properties from the tool definition
	if props, ok := params["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			// Get required parameters first (in order)
			if required, ok := params["required"]; ok {
				if requiredList, ok := required.([]interface{}); ok {
					for _, param := range requiredList {
						if paramName, ok := param.(string); ok {
							if _, exists := propsMap[paramName]; exists {
								paramNames = append(paramNames, paramName)
							}
						}
					}
				}
			}

			// Add optional parameters (those not in required list)
			requiredSet := make(map[string]bool)
			if required, ok := params["required"]; ok {
				if requiredList, ok := required.([]interface{}); ok {
					for _, param := range requiredList {
						if paramName, ok := param.(string); ok {
							requiredSet[paramName] = true
						}
					}
				}
			}

			// Go through properties in iteration order (maps are unordered, but this is best effort)
			for paramName := range propsMap {
				if !requiredSet[paramName] {
					paramNames = append(paramNames, paramName)
				}
			}
		}
	}

	return paramNames
}
