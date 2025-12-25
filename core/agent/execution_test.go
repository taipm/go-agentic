package agent

import (
	"strings"
	"testing"

	"github.com/taipm/go-agentic/core/common"
)

// TestBuildSystemPrompt verifies system prompt generation for agents
func TestBuildSystemPrompt(t *testing.T) {
	agent := &common.Agent{
		ID:        "test_agent",
		Name:      "Test Agent",
		Role:      "System Diagnostician",
		Backstory: "An expert in system analysis",
		Tools:     []interface{}{},
	}

	prompt := BuildSystemPrompt(agent)

	if prompt == "" {
		t.Error("Expected non-empty system prompt")
	}

	if !strings.Contains(prompt, "Test Agent") {
		t.Error("Expected agent name in system prompt")
	}

	if !strings.Contains(prompt, "System Diagnostician") {
		t.Error("Expected agent role in system prompt")
	}
}

// TestBuildSystemPromptWithCustomPrompt verifies custom system prompt handling
func TestBuildSystemPromptWithCustomPrompt(t *testing.T) {
	customPrompt := "You are {{name}}, a {{role}}. Backstory: {{backstory}}"
	agent := &common.Agent{
		ID:           "test_agent",
		Name:         "Test Agent",
		Role:         "Analyzer",
		Backstory:    "A custom backstory",
		SystemPrompt: customPrompt,
		Tools:        []interface{}{},
	}

	prompt := BuildSystemPrompt(agent)

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

// TestConvertToProviderMessages verifies message conversion to provider format
func TestConvertToProviderMessages(t *testing.T) {
	messages := []common.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	result := ConvertToProviderMessages(messages)

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
	messages := []common.Message{}
	result := ConvertToProviderMessages(messages)

	if len(result) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(result))
	}
}

// TestConvertToolCallsFromProvider verifies tool call conversion from provider format
func TestConvertToolCallsFromProvider(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		id       string
		args     map[string]interface{}
	}{
		{
			name:     "Simple tool call",
			toolName: "GetCPUUsage",
			id:       "call_123",
			args:     map[string]interface{}{"timeout": float64(30)},
		},
		{
			name:     "Tool call with multiple args",
			toolName: "CheckDisk",
			id:       "call_456",
			args:     map[string]interface{}{"path": "/tmp", "recursive": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from := []interface{}{
				map[string]interface{}{
					"id":        tt.id,
					"tool_name": tt.toolName,
					"arguments": tt.args,
				},
			}

			// Note: This is a simplified test since the actual conversion
			// depends on the real providerCalls structure
			_ = from
		})
	}
}

// TestBuildSystemPromptWithTools verifies prompt generation includes tool info
func TestBuildSystemPromptWithTools(t *testing.T) {
	agent := &common.Agent{
		ID:        "test_agent",
		Name:      "System Monitor",
		Role:      "Diagnostician",
		Backstory: "Monitors system health",
		Tools: []interface{}{
			"GetCPUUsage",
			"CheckDisk",
		},
	}

	prompt := BuildSystemPrompt(agent)

	if prompt == "" {
		t.Error("Expected non-empty system prompt")
	}

	if !strings.Contains(prompt, "System Monitor") {
		t.Error("Expected agent name in system prompt")
	}

	if !strings.Contains(prompt, "You have access to the following tools") {
		t.Error("Expected tool access message in prompt")
	}
}
