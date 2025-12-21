package crewai

import (
	"testing"
)

// ===== Issue #9: Tool Call Extraction Tests =====

// TestExtractFromOpenAIToolCalls verifies OpenAI native tool_calls extraction
func TestExtractFromOpenAIToolCalls(t *testing.T) {
	// Create test agent with tools
	agent := &Agent{
		ID:   "test_agent",
		Name: "Test Agent",
		Tools: []*Tool{
			{
				Name:        "GetCPUUsage",
				Description: "Get CPU usage",
				Parameters:  map[string]interface{}{},
			},
			{
				Name:        "CheckDisk",
				Description: "Check disk space",
				Parameters:  map[string]interface{}{},
			},
		},
	}

	// Test 1: Valid OpenAI tool_calls format
	t.Run("ValidOpenAIToolCalls", func(t *testing.T) {
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_123",
				"function": map[string]interface{}{
					"name":      "GetCPUUsage",
					"arguments": `{"timeout": 30}`,
				},
			},
		}

		calls := extractFromOpenAIToolCalls(toolCalls, agent)

		if len(calls) != 1 {
			t.Errorf("Expected 1 tool call, got %d", len(calls))
		}

		if calls[0].ToolName != "GetCPUUsage" {
			t.Errorf("Expected ToolName 'GetCPUUsage', got '%s'", calls[0].ToolName)
		}

		if calls[0].ID != "call_123" {
			t.Errorf("Expected ID 'call_123', got '%s'", calls[0].ID)
		}

		// Verify arguments were parsed from JSON
		if args, ok := calls[0].Arguments["timeout"]; !ok {
			t.Error("Arguments not parsed correctly from JSON")
		} else if timeout, ok := args.(float64); !ok || timeout != 30 {
			t.Errorf("Expected timeout=30, got %v", args)
		}
	})

	// Test 2: Multiple tool calls
	t.Run("MultipleToolCalls", func(t *testing.T) {
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_1",
				"function": map[string]interface{}{
					"name":      "GetCPUUsage",
					"arguments": `{}`,
				},
			},
			map[string]interface{}{
				"id": "call_2",
				"function": map[string]interface{}{
					"name":      "CheckDisk",
					"arguments": `{"path": "/tmp"}`,
				},
			},
		}

		calls := extractFromOpenAIToolCalls(toolCalls, agent)

		if len(calls) != 2 {
			t.Errorf("Expected 2 tool calls, got %d", len(calls))
		}

		if calls[0].ToolName != "GetCPUUsage" || calls[1].ToolName != "CheckDisk" {
			t.Error("Tool names not extracted correctly")
		}
	})

	// Test 3: Unknown tool in OpenAI response
	t.Run("UnknownToolFiltered", func(t *testing.T) {
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_1",
				"function": map[string]interface{}{
					"name":      "UnknownTool",
					"arguments": `{}`,
				},
			},
			map[string]interface{}{
				"id": "call_2",
				"function": map[string]interface{}{
					"name":      "GetCPUUsage",
					"arguments": `{}`,
				},
			},
		}

		calls := extractFromOpenAIToolCalls(toolCalls, agent)

		// Should only have 1 call (UnknownTool filtered out)
		if len(calls) != 1 {
			t.Errorf("Expected 1 valid tool call, got %d", len(calls))
		}

		if calls[0].ToolName != "GetCPUUsage" {
			t.Error("Valid tool not extracted")
		}
	})

	// Test 4: Empty arguments
	t.Run("EmptyArguments", func(t *testing.T) {
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_1",
				"function": map[string]interface{}{
					"name":      "GetCPUUsage",
					"arguments": ``,
				},
			},
		}

		calls := extractFromOpenAIToolCalls(toolCalls, agent)

		if len(calls) != 1 {
			t.Errorf("Expected 1 tool call, got %d", len(calls))
		}

		if len(calls[0].Arguments) != 0 {
			t.Error("Arguments should be empty map")
		}
	})

	// Test 5: Complex JSON arguments
	t.Run("ComplexArguments", func(t *testing.T) {
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_1",
				"function": map[string]interface{}{
					"name": "CheckDisk",
					"arguments": `{
						"path": "/tmp",
						"recursive": true,
						"thresholds": [10, 20, 30],
						"metadata": {"key": "value"}
					}`,
				},
			},
		}

		calls := extractFromOpenAIToolCalls(toolCalls, agent)

		if len(calls) != 1 {
			t.Fatalf("Expected 1 tool call, got %d", len(calls))
		}

		// Verify complex arguments parsed correctly
		if path, ok := calls[0].Arguments["path"].(string); !ok || path != "/tmp" {
			t.Error("String argument not parsed correctly")
		}

		if recursive, ok := calls[0].Arguments["recursive"].(bool); !ok || !recursive {
			t.Error("Boolean argument not parsed correctly")
		}

		if thresholds, ok := calls[0].Arguments["thresholds"].([]interface{}); !ok || len(thresholds) != 3 {
			t.Error("Array argument not parsed correctly")
		}
	})

	// Test 6: Invalid format handling
	t.Run("InvalidFormatHandling", func(t *testing.T) {
		// Not a slice
		calls := extractFromOpenAIToolCalls("invalid", agent)
		if len(calls) != 0 {
			t.Error("Invalid format should return empty slice")
		}

		// Malformed tool call
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_1",
				// Missing "function" field
			},
		}
		calls = extractFromOpenAIToolCalls(toolCalls, agent)
		if len(calls) != 0 {
			t.Error("Malformed tool call should be filtered")
		}
	})
}

// TestExtractToolCallsFromText verifies text parsing fallback method
func TestExtractToolCallsFromText(t *testing.T) {
	// Create test agent with tools
	agent := &Agent{
		ID:   "test_agent",
		Name: "Test Agent",
		Tools: []*Tool{
			{
				Name:        "GetCPUUsage",
				Description: "Get CPU usage",
				Parameters: map[string]interface{}{
					"required": []string{"timeout"},
				},
			},
			{
				Name:        "CheckDisk",
				Description: "Check disk",
				Parameters:  map[string]interface{}{},
			},
		},
	}

	// Test 1: Simple tool call
	t.Run("SimpleToolCall", func(t *testing.T) {
		text := "Let me check the CPU usage with GetCPUUsage(30)"
		calls := extractToolCallsFromText(text, agent)

		if len(calls) != 1 {
			t.Errorf("Expected 1 tool call, got %d", len(calls))
		}

		if calls[0].ToolName != "GetCPUUsage" {
			t.Errorf("Expected 'GetCPUUsage', got '%s'", calls[0].ToolName)
		}
	})

	// Test 2: Multiple tool calls
	t.Run("MultipleToolCalls", func(t *testing.T) {
		text := `
		First check GetCPUUsage() then check CheckDisk(/tmp)
		`
		calls := extractToolCallsFromText(text, agent)

		if len(calls) != 2 {
			t.Errorf("Expected 2 tool calls, got %d", len(calls))
		}
	})

	// Test 3: False positive from comments (known limitation of text parsing)
	t.Run("CommentMentionFalsePositive", func(t *testing.T) {
		text := "// Note: GetCPUUsage() should be called first"
		calls := extractToolCallsFromText(text, agent)

		// This is a known limitation - text parsing has false positives
		// We expect this to extract even though it's in a comment
		// This is why OpenAI native tool_calls are preferred
		if len(calls) > 0 {
			// Document known limitation
			// t.Logf("Known limitation: False positive from comment: %v", calls[0])
		}
	})

	// Test 4: No tool calls found
	t.Run("NoToolCallsFound", func(t *testing.T) {
		text := "Just a normal message with no tool calls"
		calls := extractToolCallsFromText(text, agent)

		if len(calls) != 0 {
			t.Errorf("Expected 0 tool calls, got %d", len(calls))
		}
	})
}

// TestHybridToolCallExtraction verifies the hybrid approach preference
func TestHybridToolCallExtraction(t *testing.T) {
	// When both OpenAI tool_calls and text parsing could work,
	// the hybrid approach should prefer OpenAI tool_calls (primary)

	agent := &Agent{
		ID:   "test_agent",
		Name: "Test Agent",
		Tools: []*Tool{
			{
				Name:        "Calculate",
				Description: "Simple math",
				Parameters:  map[string]interface{}{},
			},
		},
	}

	t.Run("PreferOpenAIToolCalls", func(t *testing.T) {
		// Simulate OpenAI response with native tool_calls
		openaiToolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_123",
				"function": map[string]interface{}{
					"name":      "Calculate",
					"arguments": `{"x": 5, "y": 3}`,
				},
			},
		}

		calls := extractFromOpenAIToolCalls(openaiToolCalls, agent)

		// Should extract from OpenAI format (structured, validated)
		if len(calls) != 1 {
			t.Error("Should extract from OpenAI native tool_calls")
		}

		if calls[0].ID != "call_123" {
			t.Error("Should preserve OpenAI call ID")
		}
	})

	t.Run("FallbackToTextParsing", func(t *testing.T) {
		// When OpenAI tool_calls is nil or empty, fallback to text parsing
		text := "I will use Calculate(5, 3) to solve this"

		// In hybrid approach, if message.ToolCalls is nil, use text parsing
		calls := extractToolCallsFromText(text, agent)

		if len(calls) != 1 {
			t.Error("Should fallback to text parsing when tool_calls unavailable")
		}
	})
}

// TestToolCallArgumentParsing verifies argument parsing edge cases
func TestToolCallArgumentParsing(t *testing.T) {
	agent := &Agent{
		ID:   "test_agent",
		Name: "Test Agent",
		Tools: []*Tool{
			{
				Name:        "Execute",
				Description: "Execute command",
				Parameters:  map[string]interface{}{},
			},
		},
	}

	t.Run("JSONWithEscapedStrings", func(t *testing.T) {
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_1",
				"function": map[string]interface{}{
					"name":      "Execute",
					"arguments": `{"path": "C:\\Users\\name\\file.txt", "mode": "read"}`,
				},
			},
		}

		calls := extractFromOpenAIToolCalls(toolCalls, agent)

		if len(calls) != 1 {
			t.Fatalf("Expected 1 call, got %d", len(calls))
		}

		if path, ok := calls[0].Arguments["path"].(string); !ok || path != `C:\Users\name\file.txt` {
			t.Errorf("Escaped strings not handled correctly: %v", calls[0].Arguments["path"])
		}
	})

	t.Run("JSONWithUnicodeCharacters", func(t *testing.T) {
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_1",
				"function": map[string]interface{}{
					"name":      "Execute",
					"arguments": `{"message": "Xin chào, thế giới! 你好"}`,
				},
			},
		}

		calls := extractFromOpenAIToolCalls(toolCalls, agent)

		if len(calls) != 1 {
			t.Fatalf("Expected 1 call, got %d", len(calls))
		}

		if msg, ok := calls[0].Arguments["message"].(string); !ok {
			t.Error("Unicode not handled correctly")
		} else if msg != "Xin chào, thế giới! 你好" {
			t.Errorf("Unicode characters corrupted: %s", msg)
		}
	})

	t.Run("InvalidJSONArgumentsSkipped", func(t *testing.T) {
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_1",
				"function": map[string]interface{}{
					"name":      "Execute",
					"arguments": `{"invalid json}`, // Malformed JSON
				},
			},
		}

		calls := extractFromOpenAIToolCalls(toolCalls, agent)

		// Malformed JSON should be skipped (logged but not added)
		if len(calls) != 0 {
			t.Error("Malformed JSON arguments should be skipped")
		}
	})
}

// TestOpenAIValidationFiltersInvalidTools verifies that unknown tools are rejected
func TestOpenAIValidationFiltersInvalidTools(t *testing.T) {
	agent := &Agent{
		ID:   "test_agent",
		Name: "Test Agent",
		Tools: []*Tool{
			{
				Name:        "ValidTool",
				Description: "Valid tool",
				Parameters:  map[string]interface{}{},
			},
		},
	}

	t.Run("UnknownToolsFiltered", func(t *testing.T) {
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_1",
				"function": map[string]interface{}{
					"name":      "UnknownTool",
					"arguments": `{}`,
				},
			},
			map[string]interface{}{
				"id": "call_2",
				"function": map[string]interface{}{
					"name":      "AnotherUnknownTool",
					"arguments": `{}`,
				},
			},
			map[string]interface{}{
				"id": "call_3",
				"function": map[string]interface{}{
					"name":      "ValidTool",
					"arguments": `{}`,
				},
			},
		}

		calls := extractFromOpenAIToolCalls(toolCalls, agent)

		// Should only have 1 valid call
		if len(calls) != 1 {
			t.Errorf("Expected 1 valid call, got %d", len(calls))
		}

		if calls[0].ToolName != "ValidTool" {
			t.Error("Only valid tool should be extracted")
		}
	})
}

// TestToolCallExtractorRobustness verifies edge cases are handled gracefully
func TestToolCallExtractorRobustness(t *testing.T) {
	agent := &Agent{
		ID:    "test_agent",
		Name:  "Test Agent",
		Tools: []*Tool{},
	}

	t.Run("EmptyToolsList", func(t *testing.T) {
		toolCalls := []interface{}{
			map[string]interface{}{
				"id": "call_1",
				"function": map[string]interface{}{
					"name":      "AnyTool",
					"arguments": `{}`,
				},
			},
		}

		calls := extractFromOpenAIToolCalls(toolCalls, agent)

		// No tools available, should return empty
		if len(calls) != 0 {
			t.Error("Should return empty when no tools available")
		}
	})

	t.Run("NilToolCalls", func(t *testing.T) {
		calls := extractFromOpenAIToolCalls(nil, agent)

		if calls != nil && len(calls) != 0 {
			t.Error("Nil tool_calls should return empty slice")
		}
	})

	t.Run("EmptyToolCalls", func(t *testing.T) {
		calls := extractFromOpenAIToolCalls([]interface{}{}, agent)

		if len(calls) != 0 {
			t.Error("Empty tool_calls should return empty slice")
		}
	})
}

// TestParseToolArguments verifies argument parsing with various formats
func TestParseToolArguments(t *testing.T) {
	t.Run("SimpleArguments", func(t *testing.T) {
		args := parseToolArguments("arg1, arg2, arg3")
		if len(args) != 3 {
			t.Errorf("Expected 3 args, got %d", len(args))
		}
	})

	t.Run("ArgumentsWithBrackets", func(t *testing.T) {
		args := parseToolArguments("query, [1.0, 2.0, 3.0], timeout")
		if len(args) != 3 {
			t.Errorf("Expected 3 args, got %d", len(args))
		}
		if args[1] != "[1.0, 2.0, 3.0]" {
			t.Errorf("Array not preserved: %s", args[1])
		}
	})

	t.Run("ArgumentsWithNestedParens", func(t *testing.T) {
		args := parseToolArguments("a, func(b, c), d")
		if len(args) != 3 {
			t.Errorf("Expected 3 args, got %d", len(args))
		}
		if args[1] != "func(b, c)" {
			t.Errorf("Nested parens not preserved: %s", args[1])
		}
	})

	t.Run("EmptyArguments", func(t *testing.T) {
		args := parseToolArguments("")
		if len(args) != 0 {
			t.Errorf("Empty should result in 0 args, got %d", len(args))
		}
	})
}
