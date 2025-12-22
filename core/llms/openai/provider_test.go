package openai

import (
	"context"
	"testing"

	llms "github.com/taipm/go-agentic/core/llms"
)

// TestOpenAIProviderName tests the provider name
func TestOpenAIProviderName(t *testing.T) {
	provider := &OpenAIProvider{apiKey: "test-key"}
	if provider.Name() != "openai" {
		t.Errorf("expected provider name 'openai', got %s", provider.Name())
	}
}

// TestOpenAIProviderClose tests that Close() returns nil without error
func TestOpenAIProviderClose(t *testing.T) {
	provider := &OpenAIProvider{apiKey: "test-key"}
	err := provider.Close()
	if err != nil {
		t.Errorf("expected Close() to return nil, got %v", err)
	}
}

// TestConvertToOpenAIMessages tests message conversion to OpenAI format
func TestConvertToOpenAIMessages(t *testing.T) {
	messages := []llms.ProviderMessage{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	result := convertToOpenAIMessages(messages, "You are helpful")

	if len(result) != 3 {
		t.Errorf("expected 3 messages (system + 2 user messages), got %d", len(result))
	}
}

// TestConvertToOpenAIMessagesWithoutSystemPrompt tests message conversion without system prompt
func TestConvertToOpenAIMessagesWithoutSystemPrompt(t *testing.T) {
	messages := []llms.ProviderMessage{
		{Role: "user", Content: "Hello"},
	}

	result := convertToOpenAIMessages(messages, "")

	if len(result) != 1 {
		t.Errorf("expected 1 message, got %d", len(result))
	}
}

// TestExtractToolCallsFromText tests text-based tool call extraction
func TestExtractToolCallsFromText(t *testing.T) {
	text := `I will use the GetCPUUsage tool.
GetCPUUsage()
This will help us monitor the system.`

	calls := extractToolCallsFromText(text)

	if len(calls) == 0 {
		t.Errorf("expected to extract tool calls from text, got 0")
	}

	found := false
	for _, call := range calls {
		if call.ToolName == "GetCPUUsage" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find GetCPUUsage tool call")
	}
}

// TestExtractToolCallsFromTextWithArguments tests extraction with arguments
func TestExtractToolCallsFromTextWithArguments(t *testing.T) {
	text := `Using the tool with arguments:
CheckServiceStatus(nginx)
That's the command.`

	calls := extractToolCallsFromText(text)

	found := false
	for _, call := range calls {
		if call.ToolName == "CheckServiceStatus" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find CheckServiceStatus tool call")
	}
}

// TestExtractToolCallsFromTextMultipleCalls tests extraction of multiple tool calls
func TestExtractToolCallsFromTextMultipleCalls(t *testing.T) {
	text := `First I'll check CPU:
GetCPUUsage()

Then check memory:
GetMemoryUsage()

Finally check disk:
GetDiskSpace()`

	calls := extractToolCallsFromText(text)

	expectedTools := map[string]bool{
		"GetCPUUsage":    false,
		"GetMemoryUsage": false,
		"GetDiskSpace":   false,
	}

	for _, call := range calls {
		if _, exists := expectedTools[call.ToolName]; exists {
			expectedTools[call.ToolName] = true
		}
	}

	for toolName, found := range expectedTools {
		if !found {
			t.Errorf("expected to find %s tool call", toolName)
		}
	}
}

// TestSplitArguments tests argument splitting with various formats
func TestSplitArguments(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"arg1, arg2, arg3", 3},
		{"192.168.1.1", 1},
		{"[1.0, 2.0, 3.0], 5", 2},
		{"", 0},
	}

	for _, tt := range tests {
		result := splitArguments(tt.input)
		if len(result) != tt.expected {
			t.Errorf("splitArguments(%q) expected %d args, got %d: %v", tt.input, tt.expected, len(result), result)
		}
	}
}

// TestParseToolArguments tests tool argument parsing
func TestParseToolArguments(t *testing.T) {
	args := parseToolArguments("arg1, arg2")
	if len(args) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(args))
	}
}

// TestOpenAIProviderCompleteNilRequest tests Complete with nil request
func TestOpenAIProviderCompleteNilRequest(t *testing.T) {
	provider := &OpenAIProvider{apiKey: "test-key"}
	_, err := provider.Complete(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

// TestOpenAIProviderCompleteEmptyModel tests Complete with empty model
func TestOpenAIProviderCompleteEmptyModel(t *testing.T) {
	provider := &OpenAIProvider{apiKey: "test-key"}
	req := &llms.CompletionRequest{
		Model: "",
	}
	_, err := provider.Complete(context.Background(), req)
	if err == nil {
		t.Error("expected error for empty model")
	}
}

// TestOpenAIProviderCompleteStreamNilRequest tests CompleteStream with nil request
func TestOpenAIProviderCompleteStreamNilRequest(t *testing.T) {
	provider := &OpenAIProvider{apiKey: "test-key"}
	streamChan := make(chan llms.StreamChunk, 10)
	err := provider.CompleteStream(context.Background(), nil, streamChan)
	if err == nil {
		t.Error("expected error for nil request")
	}
	close(streamChan)
}

// TestOpenAIProviderCompleteStreamEmptyModel tests CompleteStream with empty model
func TestOpenAIProviderCompleteStreamEmptyModel(t *testing.T) {
	provider := &OpenAIProvider{apiKey: "test-key"}
	req := &llms.CompletionRequest{
		Model: "",
	}
	streamChan := make(chan llms.StreamChunk, 10)
	err := provider.CompleteStream(context.Background(), req, streamChan)
	if err == nil {
		t.Error("expected error for empty model")
	}
	close(streamChan)
}

// TestNewOpenAIProviderEmpty tests provider creation with empty API key
func TestNewOpenAIProviderEmpty(t *testing.T) {
	// Manually test the factory function
	if llms.NewOpenAIProvider == nil {
		t.Skip("NewOpenAIProvider not registered")
	}

	_, err := llms.NewOpenAIProvider("")
	if err == nil {
		t.Error("expected error for empty API key")
	}
}

// TestIsAlphanumeric tests the isAlphanumeric helper
func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		ch       rune
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'_', true},
		{' ', false},
		{'(', false},
		{'-', false},
	}

	for _, tt := range tests {
		result := isAlphanumeric(tt.ch)
		if result != tt.expected {
			t.Errorf("isAlphanumeric(%q) expected %v, got %v", tt.ch, tt.expected, result)
		}
	}
}
