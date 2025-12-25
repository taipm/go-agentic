// Package tools provides utility functions for tool extraction and argument parsing
package tools

import (
	"fmt"
	"strings"
)

// ExtractToolCallsFromText extracts tool calls from unstructured response text
// Looks for patterns like: ToolName(arg1, arg2) or tool_name(...)
// Handles various argument formats: JSON, key=value, positional
// Returns unique tool calls with parsed arguments
//
// Example patterns matched:
//   SearchDatabase(query="python", limit=10)
//   GetWeather(city="New York")
//   calculate(x=5, y=10)
//
// Arguments are parsed using ParseArguments() which supports:
//   - JSON format: {key: value}
//   - Key=value format: key1=val1, key2=val2
//   - Positional arguments: arg1, arg2, arg3
func ExtractToolCallsFromText(text string) []ExtractedToolCall {
	var calls []ExtractedToolCall
	toolCallPattern := make(map[string]bool) // Track unique tool calls for deduplication

	if text == "" {
		return calls
	}

	// Split response into lines for line-by-line scanning
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Scan line for opening parentheses
		for i := 0; i < len(line); i++ {
			if line[i] == '(' {
				// Found opening paren, look back for tool name
				toolName, found := extractToolNameBackward(line, i)
				if !found {
					continue
				}

				// Look for closing paren to extract arguments
				endIdx := strings.Index(line[i:], ")")
				if endIdx == -1 {
					continue
				}

				// Extract argument string between parentheses
				endIdx += i
				argsStr := line[i+1 : endIdx]

				// Create unique tool call entry (deduplicate by toolname:args)
				callKey := fmt.Sprintf("%s:%s", toolName, argsStr)
				if toolCallPattern[callKey] {
					continue // Skip duplicate
				}

				toolCallPattern[callKey] = true

				// Parse arguments using unified argument parser
				parsedArgs := ParseArguments(argsStr)

				// Append tool call
				calls = append(calls, ExtractedToolCall{
					ToolName:  toolName,
					Arguments: parsedArgs,
				})
			}
		}
	}

	return calls
}

// ExtractedToolCall represents a tool/function call extracted from response text
// This is a simplified version used internally for text extraction
// Should be converted to providers.ToolCall when returning from provider
type ExtractedToolCall struct {
	ToolName  string
	Arguments map[string]interface{}
}

// extractToolNameBackward extracts function name by scanning backwards from opening paren
// Returns (toolName, found)
// The tool name must match identifier rules: start with letter/underscore,
// followed by alphanumeric or underscore characters
func extractToolNameBackward(line string, parenIdx int) (string, bool) {
	// Start position is just before the opening paren
	if parenIdx == 0 {
		return "", false
	}

	// Scan backwards over identifier characters
	j := parenIdx - 1

	// Skip backwards over alphanumeric and underscore
	for j >= 0 && IsAlphanumeric(rune(line[j])) {
		j--
	}

	// Check if we found a valid identifier (at least one character)
	if j >= parenIdx-1 {
		return "", false // No tool name found before paren
	}

	toolName := line[j+1 : parenIdx]

	// Validate tool name
	if !isValidToolName(toolName) {
		return "", false
	}

	return toolName, true
}

// isValidToolName validates if string is a valid tool identifier
// Must start with letter or underscore, contain only alphanumeric + underscore
// Examples:
//   SearchDatabase ✅
//   get_weather ✅
//   _private ✅
//   123invalid ❌ (starts with number)
//   invalid-name ❌ (contains hyphen)
func isValidToolName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// First character must be letter or underscore
	firstChar := rune(name[0])
	if !((firstChar >= 'a' && firstChar <= 'z') ||
		(firstChar >= 'A' && firstChar <= 'Z') ||
		firstChar == '_') {
		return false
	}

	// Remaining characters must be alphanumeric or underscore
	for i := 1; i < len(name); i++ {
		if !IsAlphanumeric(rune(name[i])) {
			return false
		}
	}

	return true
}
