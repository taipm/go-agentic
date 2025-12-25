// Package tools provides utility functions for tool execution and argument parsing
package tools

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ParseArguments parses tool arguments from a string, supporting multiple formats
// Priority:
// 1. JSON format: {key: value, ...}
// 2. Key=value format: key1=value1, key2=value2
// 3. Positional arguments: arg1, arg2, arg3
//
// Supports type conversion for numbers, booleans, and strings
func ParseArguments(argsStr string) map[string]interface{} {
	result := make(map[string]interface{})

	if argsStr == "" {
		return result
	}

	// Try to parse as JSON first (handles complex types like nested objects/arrays)
	var jsonArgs map[string]interface{}
	if err := json.Unmarshal([]byte("{"+argsStr+"}"), &jsonArgs); err == nil {
		return jsonArgs
	}

	// Try to parse key=value format (e.g., question_number=1, question="Q")
	parts := SplitArguments(argsStr)
	hasKeyValue := false
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if idx := strings.Index(part, "="); idx > 0 {
			hasKeyValue = true
			key := strings.TrimSpace(part[:idx])
			value := strings.TrimSpace(part[idx+1:])

			// Remove quotes from string values
			value = strings.Trim(value, `"'`)

			// Try to parse as number or boolean
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				result[key] = v
			} else if v, err := strconv.ParseFloat(value, 64); err == nil {
				result[key] = v
			} else if v, err := strconv.ParseBool(value); err == nil {
				result[key] = v
			} else {
				result[key] = value
			}
		}
	}

	if hasKeyValue {
		return result
	}

	// Fallback: parse as comma-separated positional arguments
	for i, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"'`)
		result[fmt.Sprintf("arg%d", i)] = part
	}

	return result
}

// SplitArguments splits arguments respecting nested brackets and quotes
// Handles comma-separated values while preserving structure in brackets and quoted strings
func SplitArguments(argsStr string) []string {
	var parts []string
	var current strings.Builder
	bracketDepth := 0
	quoteChar := rune(0)

	for _, ch := range argsStr {
		switch ch {
		case '"', '\'':
			if quoteChar == 0 {
				quoteChar = ch
			} else if quoteChar == ch {
				quoteChar = 0
			}
			current.WriteRune(ch)
		case '[', '{':
			if quoteChar == 0 {
				bracketDepth++
			}
			current.WriteRune(ch)
		case ']', '}':
			if quoteChar == 0 {
				bracketDepth--
			}
			current.WriteRune(ch)
		case ',':
			if bracketDepth == 0 && quoteChar == 0 {
				part := strings.TrimSpace(current.String())
				if part != "" {
					parts = append(parts, part)
				}
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if part := strings.TrimSpace(current.String()); part != "" {
		parts = append(parts, part)
	}

	return parts
}

// IsAlphanumeric checks if a rune is alphanumeric or underscore
// Used for validating argument names and identifiers
func IsAlphanumeric(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}
