package crewai

import (
	"strings"
	"testing"

	llms "github.com/taipm/go-agentic/core/llms"
)

// ===== Message Conversion Tests =====

// TestConvertToProviderMessages verifies message conversion to provider format
func TestConvertToProviderMessages(t *testing.T) {
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	result := convertToProviderMessages(messages)

	if len(result) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(result))
	}

	if result[0].Role != "user" || result[0].Content != "Hello" {
		t.Errorf("Message conversion failed for first message")
	}

	if result[1].Role != "assistant" || result[1].Content != "Hi there" {
		t.Errorf("Message conversion failed for second message")
	}
}

// TestConvertToProviderMessagesEmpty verifies empty message list handling
func TestConvertToProviderMessagesEmpty(t *testing.T) {
	messages := []Message{}
	result := convertToProviderMessages(messages)

	if len(result) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(result))
	}
}

// TestConvertToolsToProvider verifies tool conversion to provider format
func TestConvertToolsToProvider(t *testing.T) {
	tools := []*Tool{
		{
			Name:        "GetCPUUsage",
			Description: "Get CPU usage",
			Parameters:  map[string]interface{}{"timeout": 30},
		},
		{
			Name:        "CheckDisk",
			Description: "Check disk space",
			Parameters:  map[string]interface{}{},
		},
	}

	result := convertToolsToProvider(tools)

	if len(result) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(result))
	}

	if result[0].Name != "GetCPUUsage" {
		t.Errorf("Expected tool name 'GetCPUUsage', got %s", result[0].Name)
	}

	if result[1].Name != "CheckDisk" {
		t.Errorf("Expected tool name 'CheckDisk', got %s", result[1].Name)
	}
}

// TestConvertToolsToProviderEmpty verifies empty tool list handling
func TestConvertToolsToProviderEmpty(t *testing.T) {
	tools := []*Tool{}
	result := convertToolsToProvider(tools)

	if len(result) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(result))
	}
}

// TestConvertToolCallsFromProvider verifies tool call conversion from provider format
func TestConvertToolCallsFromProvider(t *testing.T) {
	providerCalls := []llms.ToolCall{
		{
			ID:       "call_123",
			ToolName: "GetCPUUsage",
			Arguments: map[string]interface{}{"timeout": 30},
		},
		{
			ID:       "call_456",
			ToolName: "CheckDisk",
			Arguments: map[string]interface{}{"path": "/tmp"},
		},
	}

	calls := convertToolCallsFromProvider(providerCalls)

	if len(calls) != 2 {
		t.Errorf("Expected 2 tool calls, got %d", len(calls))
	}

	if calls[0].ToolName != "GetCPUUsage" {
		t.Errorf("Expected ToolName 'GetCPUUsage', got '%s'", calls[0].ToolName)
	}

	if calls[0].ID != "call_123" {
		t.Errorf("Expected ID 'call_123', got '%s'", calls[0].ID)
	}

	timeoutVal := calls[0].Arguments["timeout"]
	timeoutOK := false
	switch v := timeoutVal.(type) {
	case float64:
		timeoutOK = v == 30
	case int:
		timeoutOK = v == 30
	}
	if !timeoutOK {
		t.Errorf("Expected timeout=30, got %v (type: %T)", timeoutVal, timeoutVal)
	}
}

// TestConvertToolCallsFromProviderEmpty verifies empty tool call list handling
func TestConvertToolCallsFromProviderEmpty(t *testing.T) {
	providerCalls := []llms.ToolCall{}
	calls := convertToolCallsFromProvider(providerCalls)

	if len(calls) != 0 {
		t.Errorf("Expected 0 tool calls, got %d", len(calls))
	}
}

// TestBuildSystemPrompt verifies system prompt generation for agents
func TestBuildSystemPrompt(t *testing.T) {
	agent := &Agent{
		ID:   "test_agent",
		Name: "Test Agent",
		Role: "System Diagnostician",
		Backstory: "An expert in system analysis",
		Tools: []*Tool{
			{
				Name:        "GetCPUUsage",
				Description: "Get CPU usage percentage",
				Parameters:  map[string]interface{}{},
			},
		},
	}

	prompt := buildSystemPrompt(agent)

	if prompt == "" {
		t.Error("Expected non-empty system prompt")
	}

	if !strings.Contains(prompt, "Test Agent") {
		t.Error("Expected agent name in system prompt")
	}

	if !strings.Contains(prompt, "System Diagnostician") {
		t.Error("Expected agent role in system prompt")
	}

	if !strings.Contains(prompt, "GetCPUUsage") {
		t.Error("Expected tool name in system prompt")
	}
}

// TestBuildSystemPromptWithCustomPrompt verifies custom system prompt handling
func TestBuildSystemPromptWithCustomPrompt(t *testing.T) {
	customPrompt := "You are {{name}}, a {{role}}. Backstory: {{backstory}}"
	agent := &Agent{
		ID:           "test_agent",
		Name:         "Test Agent",
		Role:         "Analyzer",
		Backstory:    "A custom backstory",
		SystemPrompt: customPrompt,
	}

	prompt := buildSystemPrompt(agent)

	if !strings.Contains(prompt, "Test Agent") {
		t.Error("Expected agent name substitution")
	}

	if !strings.Contains(prompt, "Analyzer") {
		t.Error("Expected agent role substitution")
	}

	if !strings.Contains(prompt, "custom backstory") {
		t.Error("Expected backstory substitution")
	}
}
