// Package tools provides utility functions for tool execution and argument parsing
package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseArguments parses tool arguments from a string, supporting multiple formats
// Tries JSON parsing first, falls back to comma-separated key-value pairs
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

	// Fallback: parse as comma-separated positional arguments
	parts := SplitArguments(argsStr)
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
