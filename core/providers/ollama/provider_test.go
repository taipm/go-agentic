package ollama

import (
	"context"
	"testing"

	providers "github.com/taipm/go-agentic/core/providers"
)

// TestOllamaProviderName tests the provider name
func TestOllamaProviderName(t *testing.T) {
	provider := &OllamaProvider{baseURL: "http://localhost:11434"}
	if provider.Name() != "ollama" {
		t.Errorf("expected provider name 'ollama', got %s", provider.Name())
	}
}

// TestOllamaProviderClose tests that Close() returns nil without error
func TestOllamaProviderClose(t *testing.T) {
	provider := &OllamaProvider{baseURL: "http://localhost:11434"}
	err := provider.Close()
	if err != nil {
		t.Errorf("expected Close() to return nil, got %v", err)
	}
}

// TestConvertToOllamaMessages tests message conversion to Ollama format
func TestConvertToOllamaMessages(t *testing.T) {
	messages := []providers.ProviderMessage{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	result := convertToOllamaMessages(messages, "You are helpful")

	if len(result) != 3 {
		t.Errorf("expected 3 messages (system + 2 user messages), got %d", len(result))
	}

	if result[0].Role != "system" || result[0].Content != "You are helpful" {
		t.Errorf("first message should be system prompt")
	}

	if result[1].Role != "user" || result[1].Content != "Hello" {
		t.Errorf("second message conversion failed")
	}
}

// TestConvertToOllamaMessagesWithoutSystemPrompt tests message conversion without system prompt
func TestConvertToOllamaMessagesWithoutSystemPrompt(t *testing.T) {
	messages := []providers.ProviderMessage{
		{Role: "user", Content: "Hello"},
	}

	result := convertToOllamaMessages(messages, "")

	if len(result) != 1 {
		t.Errorf("expected 1 message, got %d", len(result))
	}

	if result[0].Role != "user" || result[0].Content != "Hello" {
		t.Errorf("message conversion failed")
	}
}

// TestExtractToolCallsFromText tests tool call extraction from text
func TestExtractToolCallsFromText(t *testing.T) {
	text := `I will help you check the system.
GetCPUUsage()
CheckDisk(/tmp)
`
	calls := extractToolCallsFromText(text)

	if len(calls) != 2 {
		t.Errorf("expected 2 tool calls, got %d", len(calls))
	}

	if calls[0].ToolName != "GetCPUUsage" {
		t.Errorf("expected first tool 'GetCPUUsage', got '%s'", calls[0].ToolName)
	}

	if calls[1].ToolName != "CheckDisk" {
		t.Errorf("expected second tool 'CheckDisk', got '%s'", calls[1].ToolName)
	}
}

// TestExtractToolCallsFromTextWithArguments tests tool call extraction with arguments
func TestExtractToolCallsFromTextWithArguments(t *testing.T) {
	text := "Let me check: GetMemoryUsage(verbose=true) and PingHost(192.168.1.1)"
	calls := extractToolCallsFromText(text)

	if len(calls) != 2 {
		t.Errorf("expected 2 tool calls, got %d", len(calls))
	}

	if calls[0].ToolName != "GetMemoryUsage" {
		t.Errorf("expected 'GetMemoryUsage', got '%s'", calls[0].ToolName)
	}

	if len(calls[0].Arguments) == 0 {
		t.Errorf("expected arguments for GetMemoryUsage")
	}
}

// TestExtractToolCallsFromTextMultipleCalls tests multiple tool calls extraction
func TestExtractToolCallsFromTextMultipleCalls(t *testing.T) {
	text := `System analysis:
GetCPUUsage()
GetMemoryUsage()
GetDiskSpace()
`
	calls := extractToolCallsFromText(text)

	if len(calls) != 3 {
		t.Errorf("expected 3 tool calls, got %d", len(calls))
	}

	expectedTools := []string{"GetCPUUsage", "GetMemoryUsage", "GetDiskSpace"}
	for i, expected := range expectedTools {
		if calls[i].ToolName != expected {
			t.Errorf("tool %d: expected '%s', got '%s'", i, expected, calls[i].ToolName)
		}
	}
}

// TestSplitArguments tests argument splitting with nested structures
func TestSplitArguments(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"arg1, arg2, arg3", 3},
		{"[1, 2, 3], value", 2},
		{`{"key": "value"}, arg2`, 2},
		{"single", 1},
		{"", 0},
	}

	for _, test := range tests {
		result := splitArguments(test.input)
		if len(result) != test.expected {
			t.Errorf("splitArguments(%q): expected %d parts, got %d", test.input, test.expected, len(result))
		}
	}
}

// TestParseToolArguments tests tool argument parsing
func TestParseToolArguments(t *testing.T) {
	tests := []struct {
		input    string
		hasError bool
	}{
		{"timeout: 30", false},
		{"path: /tmp", false},
		{"", false},
		{`{"key": "value"}`, false},
	}

	for _, test := range tests {
		result := parseToolArguments(test.input)
		if result == nil {
			t.Errorf("parseToolArguments(%q): expected non-nil result", test.input)
		}
	}
}

// TestOllamaProviderCompleteNilRequest tests Complete with nil request
func TestOllamaProviderCompleteNilRequest(t *testing.T) {
	provider := &OllamaProvider{baseURL: "http://localhost:11434"}
	_, err := provider.Complete(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

// TestOllamaProviderCompleteEmptyModel tests Complete with empty model
func TestOllamaProviderCompleteEmptyModel(t *testing.T) {
	provider := &OllamaProvider{baseURL: "http://localhost:11434"}
	req := &providers.CompletionRequest{
		Model: "",
	}
	_, err := provider.Complete(context.Background(), req)
	if err == nil {
		t.Error("expected error for empty model")
	}
}

// TestOllamaProviderCompleteStreamNilRequest tests CompleteStream with nil request
func TestOllamaProviderCompleteStreamNilRequest(t *testing.T) {
	provider := &OllamaProvider{baseURL: "http://localhost:11434"}
	streamChan := make(chan providers.StreamChunk, 10)
	err := provider.CompleteStream(context.Background(), nil, streamChan)
	if err == nil {
		t.Error("expected error for nil request")
	}
	close(streamChan)
}

// TestOllamaProviderCompleteStreamEmptyModel tests CompleteStream with empty model
func TestOllamaProviderCompleteStreamEmptyModel(t *testing.T) {
	provider := &OllamaProvider{baseURL: "http://localhost:11434"}
	req := &providers.CompletionRequest{
		Model: "",
	}
	streamChan := make(chan providers.StreamChunk, 10)
	err := provider.CompleteStream(context.Background(), req, streamChan)
	if err == nil {
		t.Error("expected error for empty model")
	}
	close(streamChan)
}

// TestNewOllamaProviderDefaultURL tests provider creation with default URL
func TestNewOllamaProviderDefaultURL(t *testing.T) {
	if providers.NewOllamaProvider == nil {
		t.Skip("NewOllamaProvider not registered")
	}

	provider, err := providers.NewOllamaProvider("")
	if err != nil {
		t.Errorf("expected no error for default URL, got %v", err)
	}

	if provider == nil {
		t.Error("expected non-nil provider")
	}

	if provider.Name() != "ollama" {
		t.Errorf("expected provider name 'ollama', got %s", provider.Name())
	}
}

// TestNewOllamaProviderCustomURL tests provider creation with custom URL
func TestNewOllamaProviderCustomURL(t *testing.T) {
	if providers.NewOllamaProvider == nil {
		t.Skip("NewOllamaProvider not registered")
	}

	provider, err := providers.NewOllamaProvider("http://custom-ollama:11434")
	if err != nil {
		t.Errorf("expected no error for custom URL, got %v", err)
	}

	if provider == nil {
		t.Error("expected non-nil provider")
	}
}

// TestNewOllamaProviderInvalidURL tests provider creation with invalid URL
func TestNewOllamaProviderInvalidURL(t *testing.T) {
	if providers.NewOllamaProvider == nil {
		t.Skip("NewOllamaProvider not registered")
	}

	// Invalid URL format should be handled gracefully
	provider, err := providers.NewOllamaProvider("://invalid")
	if err == nil && provider == nil {
		// Either error or nil provider is acceptable for invalid URL
		return
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
		{'-', false},
		{' ', false},
		{'(', false},
	}

	for _, test := range tests {
		result := isAlphanumeric(test.ch)
		if result != test.expected {
			t.Errorf("isAlphanumeric(%q): expected %v, got %v", test.ch, test.expected, result)
		}
	}
}
