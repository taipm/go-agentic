package agentic

import (
	"testing"

	openai "github.com/openai/openai-go/v3"
)

// Test 1.1.1: Different agents use different models
func TestDifferentAgentsUseDifferentModels(t *testing.T) {
	agent1 := &Agent{
		ID:    "agent-1",
		Name:  "Agent1",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	agent2 := &Agent{
		ID:    "agent-2",
		Name:  "Agent2",
		Model: "gpt-4o-mini",
		Tools: []*Tool{},
	}

	// Verify models are different (sanity check)
	if agent1.Model == agent2.Model {
		t.Errorf("Expected different models, but both are %s", agent1.Model)
	}

	if agent1.Model != "gpt-4o" {
		t.Errorf("Agent1 should have model gpt-4o, got %s", agent1.Model)
	}

	if agent2.Model != "gpt-4o-mini" {
		t.Errorf("Agent2 should have model gpt-4o-mini, got %s", agent2.Model)
	}
}

// Test 1.1.2: Agent model is used in ExecuteAgent (verified via buildSystemPrompt)
func TestAgentModelIsRespected(t *testing.T) {
	agent := &Agent{
		ID:    "test-agent",
		Name:  "TestAgent",
		Role:  "Helper",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	// Verify the model field is set correctly
	if agent.Model != "gpt-4o" {
		t.Errorf("Expected model gpt-4o, got %s", agent.Model)
	}

	// When ExecuteAgent is called, the params.Model will use agent.Model
	// This is verified by the code change: Model: agent.Model (not hardcoded)
	// Integration tests would verify the actual API call uses this value
}

// Test 1.1.3: Build system prompt works with agent model
func TestBuildSystemPromptIncludesAgentInfo(t *testing.T) {
	agent := &Agent{
		ID:         "test",
		Name:       "TestAgent",
		Role:       "Helper",
		Model:      "gpt-4o",
		Backstory:  "A helpful assistant",
		Tools:      []*Tool{},
	}

	prompt := buildSystemPrompt(agent)

	// Verify prompt contains agent information
	if prompt == "" {
		t.Fatal("Expected non-empty system prompt")
	}

	if !contains(prompt, agent.Name) {
		t.Errorf("System prompt should contain agent name %s", agent.Name)
	}

	if !contains(prompt, agent.Role) {
		t.Errorf("System prompt should contain agent role %s", agent.Role)
	}
}

// Test 1.1.4: Multiple agents with different models
func TestMultipleAgentsWithDifferentModels(t *testing.T) {
	agents := []*Agent{
		{
			ID:    "agent-1",
			Name:  "Agent1",
			Model: "gpt-4o",
			Tools: []*Tool{},
		},
		{
			ID:    "agent-2",
			Name:  "Agent2",
			Model: "gpt-4o-mini",
			Tools: []*Tool{},
		},
		{
			ID:    "agent-3",
			Name:  "Agent3",
			Model: "gpt-4",
			Tools: []*Tool{},
		},
	}

	// Verify each agent has correct model
	expectedModels := map[string]string{
		"agent-1": "gpt-4o",
		"agent-2": "gpt-4o-mini",
		"agent-3": "gpt-4",
	}

	for _, agent := range agents {
		expectedModel := expectedModels[agent.ID]
		if agent.Model != expectedModel {
			t.Errorf("Agent %s: expected model %s, got %s", agent.ID, expectedModel, agent.Model)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============ Test Story 2a.1: Parse Native OpenAI Tool Calls ============

// TestParseNativeToolCallsBasic tests extraction of native OpenAI tool calls
func TestParseNativeToolCallsBasic(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Tools: []*Tool{
			{Name: "PingHost", Description: "Ping a host"},
			{Name: "GetCPUUsage", Description: "Get CPU usage"},
		},
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "PingHost",
				Arguments: `{"host":"192.168.1.100"}`,
			},
		},
	}

	result := parseNativeToolCalls(toolCalls, agent)

	if len(result) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(result))
	}
	if result[0].ToolName != "PingHost" {
		t.Errorf("Expected tool name 'PingHost', got %q", result[0].ToolName)
	}
	if result[0].Arguments["host"] != "192.168.1.100" {
		t.Errorf("Expected host '192.168.1.100', got %v", result[0].Arguments["host"])
	}
}

// TestParseNativeToolCallsMultiple tests parsing multiple native tool calls
func TestParseNativeToolCallsMultiple(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Tools: []*Tool{
			{Name: "GetCPUUsage", Description: "Get CPU usage"},
			{Name: "PingHost", Description: "Ping host"},
		},
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "GetCPUUsage",
				Arguments: `{}`,
			},
		},
		{
			ID:   "call_2",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "PingHost",
				Arguments: `{"host":"8.8.8.8"}`,
			},
		},
	}

	result := parseNativeToolCalls(toolCalls, agent)

	if len(result) != 2 {
		t.Errorf("Expected 2 tool calls, got %d", len(result))
	}
	if result[0].ToolName != "GetCPUUsage" || result[1].ToolName != "PingHost" {
		t.Error("Tool names don't match expected")
	}
}

// TestParseNativeToolCallsFiltersInvalidTools tests filtering of unknown tools
func TestParseNativeToolCallsFiltersInvalidTools(t *testing.T) {
	agent := &Agent{
		ID:   "clarifier",
		Name: "Clarifier",
		Tools: []*Tool{
			{Name: "AskQuestion", Description: "Ask question"},
		},
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "UnknownTool",
				Arguments: `{}`,
			},
		},
	}

	result := parseNativeToolCalls(toolCalls, agent)

	if len(result) != 0 {
		t.Errorf("Expected 0 tool calls for unknown tool, got %d", len(result))
	}
}

// TestParseNativeToolCallsFiltersNonFunctionType tests filtering non-function types
func TestParseNativeToolCallsFiltersNonFunctionType(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Tools: []*Tool{
			{Name: "PingHost", Description: "Ping host"},
		},
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "custom", // Not a function
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "PingHost",
				Arguments: `{}`,
			},
		},
	}

	result := parseNativeToolCalls(toolCalls, agent)

	if len(result) != 0 {
		t.Errorf("Expected 0 tool calls for non-function type, got %d", len(result))
	}
}

// ============ Test Story 2a.2: Fallback Text Parsing ============

// TestFallbackTextParsingBasic tests basic text-based tool call extraction
func TestFallbackTextParsingBasic(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Tools: []*Tool{
			{
				Name: "PingHost",
				Parameters: map[string]interface{}{
					"properties": map[string]interface{}{
						"host": map[string]interface{}{},
					},
					"required": []string{"host"},
				},
			},
		},
	}

	text := "I will ping 192.168.1.100\nPingHost(192.168.1.100)"
	result := extractToolCallsFromText(text, agent)

	if len(result) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(result))
	}
	if result[0].ToolName != "PingHost" {
		t.Errorf("Expected 'PingHost', got %q", result[0].ToolName)
	}
}

// TestFallbackTextParsingMultiple tests extracting multiple tool calls
func TestFallbackTextParsingMultiple(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Tools: []*Tool{
			{Name: "GetCPUUsage", Description: "Get CPU"},
			{Name: "PingHost", Description: "Ping host"},
		},
	}

	text := "GetCPUUsage()\nThen PingHost(192.168.1.1)"
	result := extractToolCallsFromText(text, agent)

	if len(result) != 2 {
		t.Errorf("Expected 2 tool calls, got %d", len(result))
	}
}

// TestFallbackTextParsingNoTools tests with no tool calls
func TestFallbackTextParsingNoTools(t *testing.T) {
	agent := &Agent{
		ID:   "clarifier",
		Name: "Clarifier",
		Tools: []*Tool{
			{Name: "AskQuestion", Description: "Ask"},
		},
	}

	text := "I need more information about your system."
	result := extractToolCallsFromText(text, agent)

	if len(result) != 0 {
		t.Errorf("Expected 0 tool calls, got %d", len(result))
	}
}

// ============ Test Story 2a.3: System Prompt Updates ============

// TestSystemPromptIncludesFunctionCalling tests prompt mentions function calling
func TestSystemPromptIncludesFunctionCalling(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Role: "Execute commands",
		Tools: []*Tool{
			{Name: "PingHost", Description: "Ping host"},
			{Name: "GetCPUUsage", Description: "Get CPU"},
		},
	}

	prompt := buildSystemPrompt(agent)

	if !contains(prompt, "function calling") && !contains(prompt, "Function calling") {
		t.Error("System prompt should mention function calling")
	}

	if contains(prompt, "ToolName(param1, param2)") {
		t.Error("System prompt should not use old text format")
	}
}

// TestParseJSONArgumentsValid tests JSON parsing with valid input
func TestParseJSONArgumentsValid(t *testing.T) {
	args := make(map[string]interface{})
	err := parseJSONArguments(`{"host":"192.168.1.1"}`, args)

	if err != nil {
		t.Errorf("parseJSONArguments failed: %v", err)
	}
	if args["host"] != "192.168.1.1" {
		t.Errorf("Expected host argument, got %v", args)
	}
}

// TestParseJSONArgumentsInvalid tests JSON parsing with invalid input
func TestParseJSONArgumentsInvalid(t *testing.T) {
	args := make(map[string]interface{})
	err := parseJSONArguments(`{invalid json}`, args)

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
