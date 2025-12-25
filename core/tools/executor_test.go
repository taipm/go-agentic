package tools

import (
	"context"
	"fmt"
	"testing"

	"github.com/taipm/go-agentic/core/common"
)

// TestExecuteTool tests single tool execution with various input types
func TestExecuteTool(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		tool      interface{}
		args      map[string]interface{}
		expectErr bool
		checkMsg  func(string) bool
	}{
		{
			name:     "Execute valid tool by pointer",
			toolName: "test_tool",
			tool: &common.Tool{
				Name:        "test_tool",
				Description: "Test tool",
				Func: ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
					return "success", nil
				}),
			},
			args:      map[string]interface{}{"arg1": "value1"},
			expectErr: false,
			checkMsg: func(msg string) bool {
				return msg == "success"
			},
		},
		{
			name:     "Execute valid tool by value",
			toolName: "test_tool",
			tool: common.Tool{
				Name:        "test_tool",
				Description: "Test tool",
				Func: ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
					return "success", nil
				}),
			},
			args:      map[string]interface{}{},
			expectErr: false,
			checkMsg: func(msg string) bool {
				return msg == "success"
			},
		},
		{
			name:      "Tool is nil",
			toolName:  "test_tool",
			tool:      nil,
			args:      map[string]interface{}{},
			expectErr: true,
		},
		{
			name:      "Tool name is empty",
			toolName:  "",
			tool:      &common.Tool{Name: "test_tool"},
			args:      map[string]interface{}{},
			expectErr: true,
		},
		{
			name:      "Tool with nil Func",
			toolName:  "test_tool",
			tool:      &common.Tool{Name: "test_tool", Func: nil},
			args:      map[string]interface{}{},
			expectErr: true,
		},
		{
			name:      "Tool with wrong Func type",
			toolName:  "test_tool",
			tool:      &common.Tool{Name: "test_tool", Func: "not_a_function"},
			args:      map[string]interface{}{},
			expectErr: true,
		},
		{
			name:     "Tool function returns error",
			toolName: "error_tool",
			tool: &common.Tool{
				Name: "error_tool",
				Func: ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
					return "", fmt.Errorf("tool execution failed")
				}),
			},
			args:      map[string]interface{}{},
			expectErr: true,
		},
		{
			name:     "Tool with arguments",
			toolName: "arg_tool",
			tool: &common.Tool{
				Name: "arg_tool",
				Func: ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
					val, ok := args["key"]
					if !ok {
						return "", fmt.Errorf("missing key")
					}
					return fmt.Sprintf("got %v", val), nil
				}),
			},
			args:      map[string]interface{}{"key": "test_value"},
			expectErr: false,
			checkMsg: func(msg string) bool {
				return msg == "got test_value"
			},
		},
		{
			name:      "Invalid tool type",
			toolName:  "invalid_tool",
			tool:      123, // Wrong type
			args:      map[string]interface{}{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExecuteTool(context.Background(), tt.toolName, tt.tool, tt.args)

			if (err != nil) != tt.expectErr {
				t.Errorf("ExecuteTool() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr && tt.checkMsg != nil {
				if !tt.checkMsg(result) {
					t.Errorf("ExecuteTool() result = %q, check failed", result)
				}
			}
		})
	}
}

// TestExecuteToolCalls tests batch tool execution with multiple tools
func TestExecuteToolCalls(t *testing.T) {
	// Helper to create tool
	makeTool := func(name string, result string) interface{} {
		return &common.Tool{
			Name: name,
			Func: ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
				return result, nil
			}),
		}
	}

	// Helper to create failing tool
	makeFailingTool := func(name string) interface{} {
		return &common.Tool{
			Name: name,
			Func: ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
				return "", fmt.Errorf("%s failed", name)
			}),
		}
	}

	tests := []struct {
		name           string
		toolCalls      []common.ToolCall
		agentTools     []interface{}
		expectErr      bool
		expectResults  map[string]string
		checkResults   func(map[string]string) bool
	}{
		{
			name:           "Single tool execution",
			toolCalls:      []common.ToolCall{{ToolName: "tool1", Arguments: map[string]interface{}{}}},
			agentTools:     []interface{}{makeTool("tool1", "result1")},
			expectErr:      false,
			expectResults:  map[string]string{"tool1": "result1"},
		},
		{
			name: "Multiple tools execution",
			toolCalls: []common.ToolCall{
				{ToolName: "tool1", Arguments: map[string]interface{}{}},
				{ToolName: "tool2", Arguments: map[string]interface{}{}},
				{ToolName: "tool3", Arguments: map[string]interface{}{}},
			},
			agentTools: []interface{}{
				makeTool("tool1", "result1"),
				makeTool("tool2", "result2"),
				makeTool("tool3", "result3"),
			},
			expectErr: false,
			checkResults: func(results map[string]string) bool {
				return len(results) == 3 &&
					results["tool1"] == "result1" &&
					results["tool2"] == "result2" &&
					results["tool3"] == "result3"
			},
		},
		{
			name: "Partial failure - one tool fails",
			toolCalls: []common.ToolCall{
				{ToolName: "tool1", Arguments: map[string]interface{}{}},
				{ToolName: "tool2", Arguments: map[string]interface{}{}},
			},
			agentTools: []interface{}{
				makeTool("tool1", "result1"),
				makeFailingTool("tool2"),
			},
			expectErr: true, // Should return error but also partial results
			checkResults: func(results map[string]string) bool {
				// Should have result from successful tool
				return len(results) == 1 && results["tool1"] == "result1"
			},
		},
		{
			name: "All tools fail",
			toolCalls: []common.ToolCall{
				{ToolName: "tool1", Arguments: map[string]interface{}{}},
				{ToolName: "tool2", Arguments: map[string]interface{}{}},
			},
			agentTools: []interface{}{
				makeFailingTool("tool1"),
				makeFailingTool("tool2"),
			},
			expectErr: true,
			checkResults: func(results map[string]string) bool {
				return len(results) == 0
			},
		},
		{
			name:           "Empty tool calls",
			toolCalls:      []common.ToolCall{},
			agentTools:     []interface{}{makeTool("tool1", "result1")},
			expectErr:      false,
			expectResults:  map[string]string{},
		},
		{
			name: "Tool not found",
			toolCalls: []common.ToolCall{
				{ToolName: "missing_tool", Arguments: map[string]interface{}{}},
			},
			agentTools: []interface{}{
				makeTool("tool1", "result1"),
			},
			expectErr: true,
			checkResults: func(results map[string]string) bool {
				return len(results) == 0
			},
		},
		{
			name: "Mixed valid and invalid tools",
			toolCalls: []common.ToolCall{
				{ToolName: "tool1", Arguments: map[string]interface{}{}},
				{ToolName: "missing", Arguments: map[string]interface{}{}},
				{ToolName: "tool2", Arguments: map[string]interface{}{}},
			},
			agentTools: []interface{}{
				makeTool("tool1", "result1"),
				makeTool("tool2", "result2"),
			},
			expectErr: true, // One tool missing
			checkResults: func(results map[string]string) bool {
				// Should have results from found tools
				return len(results) == 2 &&
					results["tool1"] == "result1" &&
					results["tool2"] == "result2"
			},
		},
		{
			name: "Nil tools in agent tools list",
			toolCalls: []common.ToolCall{
				{ToolName: "tool1", Arguments: map[string]interface{}{}},
			},
			agentTools: []interface{}{
				nil,
				makeTool("tool1", "result1"),
				nil,
			},
			expectErr: false,
			checkResults: func(results map[string]string) bool {
				return len(results) == 1 && results["tool1"] == "result1"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := ExecuteToolCalls(context.Background(), tt.toolCalls, tt.agentTools)

			if (err != nil) != tt.expectErr {
				t.Errorf("ExecuteToolCalls() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if tt.expectResults != nil {
				if len(results) != len(tt.expectResults) {
					t.Errorf("ExecuteToolCalls() results length = %d, expected %d", len(results), len(tt.expectResults))
					return
				}
				for key, val := range tt.expectResults {
					if results[key] != val {
						t.Errorf("ExecuteToolCalls()[%q] = %q, expected %q", key, results[key], val)
					}
				}
			}

			if tt.checkResults != nil && !tt.checkResults(results) {
				t.Errorf("ExecuteToolCalls() results check failed: %v", results)
			}
		})
	}
}

// TestBuildToolMap tests the tool map builder
func TestBuildToolMap(t *testing.T) {
	tests := []struct {
		name         string
		agentTools   []interface{}
		expectCount  int
		expectNames  []string
	}{
		{
			name:        "Empty tools",
			agentTools:  []interface{}{},
			expectCount: 0,
			expectNames: []string{},
		},
		{
			name: "Single tool by pointer",
			agentTools: []interface{}{
				&common.Tool{Name: "tool1"},
			},
			expectCount: 1,
			expectNames: []string{"tool1"},
		},
		{
			name: "Single tool by value",
			agentTools: []interface{}{
				common.Tool{Name: "tool1"},
			},
			expectCount: 1,
			expectNames: []string{"tool1"},
		},
		{
			name: "Multiple tools mixed types",
			agentTools: []interface{}{
				&common.Tool{Name: "tool1"},
				common.Tool{Name: "tool2"},
				&common.Tool{Name: "tool3"},
			},
			expectCount: 3,
			expectNames: []string{"tool1", "tool2", "tool3"},
		},
		{
			name: "Tools with nil entries",
			agentTools: []interface{}{
				&common.Tool{Name: "tool1"},
				nil,
				common.Tool{Name: "tool2"},
				nil,
			},
			expectCount: 2,
			expectNames: []string{"tool1", "tool2"},
		},
		{
			name: "Tool with empty name (should be skipped)",
			agentTools: []interface{}{
				&common.Tool{Name: "tool1"},
				&common.Tool{Name: ""},
				&common.Tool{Name: "tool2"},
			},
			expectCount: 2,
			expectNames: []string{"tool1", "tool2"},
		},
		{
			name: "Invalid tool type (should be skipped)",
			agentTools: []interface{}{
				&common.Tool{Name: "tool1"},
				"invalid",
				123,
				&common.Tool{Name: "tool2"},
			},
			expectCount: 2,
			expectNames: []string{"tool1", "tool2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolMap := buildToolMap(tt.agentTools)

			if len(toolMap) != tt.expectCount {
				t.Errorf("buildToolMap() returned %d tools, expected %d", len(toolMap), tt.expectCount)
			}

			for _, name := range tt.expectNames {
				if _, exists := toolMap[name]; !exists {
					t.Errorf("buildToolMap() missing expected tool: %s", name)
				}
			}
		})
	}
}

// TestFormatToolResults tests result formatting
func TestFormatToolResults(t *testing.T) {
	tests := []struct {
		name      string
		results   map[string]string
		checkMsg  func(string) bool
	}{
		{
			name:    "Empty results",
			results: map[string]string{},
			checkMsg: func(msg string) bool {
				return msg == "No tool results available"
			},
		},
		{
			name: "Single result",
			results: map[string]string{
				"tool1": "result1",
			},
			checkMsg: func(msg string) bool {
				return len(msg) > 0 && msg != "No tool results available"
			},
		},
		{
			name: "Multiple results",
			results: map[string]string{
				"tool1": "result1",
				"tool2": "result2",
				"tool3": "result3",
			},
			checkMsg: func(msg string) bool {
				return len(msg) > 0 &&
					len(msg) > len("No tool results available")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatToolResults(tt.results)

			if !tt.checkMsg(result) {
				t.Errorf("FormatToolResults() = %q, check failed", result)
			}
		})
	}
}

// TestFindToolByName tests finding tools by name
func TestFindToolByName(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		agentTools   []interface{}
		shouldFind   bool
		checkTool    func(*common.Tool) bool
	}{
		{
			name:       "Find by exact name",
			toolName:   "tool1",
			agentTools: []interface{}{&common.Tool{Name: "tool1"}},
			shouldFind: true,
			checkTool: func(t *common.Tool) bool {
				return t != nil && t.Name == "tool1"
			},
		},
		{
			name:       "Tool not found",
			toolName:   "missing",
			agentTools: []interface{}{&common.Tool{Name: "tool1"}},
			shouldFind: false,
		},
		{
			name:       "Empty tool name",
			toolName:   "",
			agentTools: []interface{}{&common.Tool{Name: "tool1"}},
			shouldFind: false,
		},
		{
			name:       "Nil tools",
			toolName:   "tool1",
			agentTools: []interface{}{},
			shouldFind: false,
		},
		{
			name:     "Find in mixed types",
			toolName: "tool2",
			agentTools: []interface{}{
				&common.Tool{Name: "tool1"},
				common.Tool{Name: "tool2"},
				&common.Tool{Name: "tool3"},
			},
			shouldFind: true,
			checkTool: func(t *common.Tool) bool {
				return t != nil && t.Name == "tool2"
			},
		},
		{
			name:     "Find with nil entries",
			toolName: "tool2",
			agentTools: []interface{}{
				&common.Tool{Name: "tool1"},
				nil,
				common.Tool{Name: "tool2"},
				nil,
			},
			shouldFind: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := FindToolByName(tt.agentTools, tt.toolName)

			if tt.shouldFind && tool == nil {
				t.Errorf("FindToolByName() returned nil, expected to find tool")
				return
			}

			if !tt.shouldFind && tool != nil {
				t.Errorf("FindToolByName() found %v, expected not to find tool", tool)
				return
			}

			if tt.shouldFind && tt.checkTool != nil && !tt.checkTool(tool) {
				t.Errorf("FindToolByName() check failed for tool: %v", tool)
			}
		})
	}
}

// TestValidateToolCall tests tool call validation
func TestValidateToolCall(t *testing.T) {
	tests := []struct {
		name      string
		toolCall  common.ToolCall
		agentTools []interface{}
		expectErr bool
	}{
		{
			name:     "Valid tool call",
			toolCall: common.ToolCall{ToolName: "tool1", Arguments: map[string]interface{}{}},
			agentTools: []interface{}{
				&common.Tool{Name: "tool1"},
			},
			expectErr: false,
		},
		{
			name:       "Empty tool name",
			toolCall:   common.ToolCall{ToolName: "", Arguments: map[string]interface{}{}},
			agentTools: []interface{}{&common.Tool{Name: "tool1"}},
			expectErr:  true,
		},
		{
			name:     "Tool not found",
			toolCall: common.ToolCall{ToolName: "missing", Arguments: map[string]interface{}{}},
			agentTools: []interface{}{
				&common.Tool{Name: "tool1"},
			},
			expectErr: true,
		},
		{
			name:     "Nil arguments (should be ok)",
			toolCall: common.ToolCall{ToolName: "tool1", Arguments: nil},
			agentTools: []interface{}{
				&common.Tool{Name: "tool1"},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolCall(tt.toolCall, tt.agentTools)

			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateToolCall() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
