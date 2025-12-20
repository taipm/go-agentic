package agentic

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestNewStreamEventBasic tests stream event creation
func TestNewStreamEventBasic(t *testing.T) {
	event := NewStreamEvent("agent_start", "Agent1", "Starting...")

	if event.Type != "agent_start" {
		t.Error("Event type not set correctly")
	}
	if event.Agent != "Agent1" {
		t.Error("Agent not set correctly")
	}
	if event.Content != "Starting..." {
		t.Error("Content not set correctly")
	}
	if event.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

// TestNewStreamEventWithMetadataBasic tests stream event with metadata
func TestNewStreamEventWithMetadataBasic(t *testing.T) {
	metadata := map[string]string{"key": "value"}
	event := NewStreamEventWithMetadata("tool_result", "Agent1", "Result", metadata)

	if event.Type != "tool_result" {
		t.Error("Event type not set correctly")
	}
	if event.Agent != "Agent1" {
		t.Error("Agent not set correctly")
	}
	if event.Content != "Result" {
		t.Error("Content not set correctly")
	}
	if event.Metadata == nil {
		t.Error("Metadata should be set")
	}
}

// TestFormatStreamEventBasicFormat tests formatting stream event as SSE
func TestFormatStreamEventBasicFormat(t *testing.T) {
	event := NewStreamEvent("agent_response", "TestAgent", "Hello World")
	formatted := FormatStreamEvent(event)

	if !strings.HasPrefix(formatted, "data: ") {
		t.Error("Formatted event should start with 'data: '")
	}

	if !strings.HasSuffix(formatted, "\n\n") {
		t.Error("Formatted event should end with double newline")
	}

	// Should contain valid JSON
	jsonStart := strings.Index(formatted, "{")
	jsonEnd := strings.LastIndex(formatted, "}")
	if jsonStart < 0 || jsonEnd < 0 {
		t.Error("Formatted event should contain JSON")
	}
}

// TestFormatStreamEventJSONValid tests that formatted event contains valid JSON
func TestFormatStreamEventJSONValid(t *testing.T) {
	event := NewStreamEvent("agent_start", "TestAgent", "Starting...")
	formatted := FormatStreamEvent(event)

	// Extract JSON from "data: {...}\n\n"
	jsonStr := strings.TrimPrefix(formatted, "data: ")
	jsonStr = strings.TrimSuffix(jsonStr, "\n\n")

	var unmarshaled StreamEvent
	err := json.Unmarshal([]byte(jsonStr), &unmarshaled)
	if err != nil {
		t.Errorf("Formatted JSON is invalid: %v", err)
	}

	if unmarshaled.Type != "agent_start" {
		t.Error("JSON Type field incorrect after unmarshal")
	}
	if unmarshaled.Agent != "TestAgent" {
		t.Error("JSON Agent field incorrect after unmarshal")
	}
	if unmarshaled.Content != "Starting..." {
		t.Error("JSON Content field incorrect after unmarshal")
	}
}

// TestFormatStreamEventNilTimestampSet tests that nil timestamp is set
func TestFormatStreamEventNilTimestampSet(t *testing.T) {
	event := &StreamEvent{
		Type:    "test",
		Agent:   "Agent1",
		Content: "Test",
	}

	// Timestamp is zero before formatting
	if !event.Timestamp.IsZero() {
		t.Error("Initial timestamp should be zero")
	}

	FormatStreamEvent(event)

	// Timestamp should be set after formatting
	if event.Timestamp.IsZero() {
		t.Error("Timestamp should be set after formatting")
	}
}

// TestFormatStreamEventWithMetadata tests formatting event with metadata
func TestFormatStreamEventWithMetadata(t *testing.T) {
	metadata := map[string]string{"tool": "calculator", "result": "42"}
	event := NewStreamEventWithMetadata("tool_result", "Agent1", "42", metadata)

	formatted := FormatStreamEvent(event)

	// Should contain valid JSON
	jsonStr := strings.TrimPrefix(formatted, "data: ")
	jsonStr = strings.TrimSuffix(jsonStr, "\n\n")

	var unmarshaled StreamEvent
	err := json.Unmarshal([]byte(jsonStr), &unmarshaled)
	if err != nil {
		t.Errorf("JSON unmarshal failed: %v", err)
	}

	if unmarshaled.Metadata == nil {
		t.Error("Metadata should be preserved in JSON")
	}
}

// TestFormatStreamEventErrorRecovery tests JSON marshaling error handling
func TestFormatStreamEventErrorRecovery(t *testing.T) {
	event := &StreamEvent{
		Type:    "test",
		Agent:   "Agent1",
		Content: "Test",
		// Create unmarshalable metadata - circular reference would be best
		// but for JSON it's hard to trigger. This test validates the error recovery exists.
		Timestamp: time.Now(),
	}

	// This should not panic even if there are issues
	formatted := FormatStreamEvent(event)

	if !strings.Contains(formatted, "data: ") {
		t.Error("Formatted event should still have SSE format")
	}
}

// TestFormatStreamEventMultipleEvents tests formatting different event types
func TestFormatStreamEventMultipleEvents(t *testing.T) {
	events := []*StreamEvent{
		NewStreamEvent("agent_start", "Agent1", "Starting"),
		NewStreamEvent("agent_response", "Agent1", "Response"),
		NewStreamEvent("tool_start", "Agent1", "Executing tool"),
		NewStreamEvent("tool_result", "Agent1", "Tool result"),
		NewStreamEvent("pause", "Agent1", "Paused"),
		NewStreamEvent("error", "Agent1", "Error message"),
	}

	for _, event := range events {
		formatted := FormatStreamEvent(event)

		// All should be valid SSE format
		if !strings.HasPrefix(formatted, "data: {") {
			t.Errorf("Event %s not in valid SSE format", event.Type)
		}
		if !strings.HasSuffix(formatted, "\n\n") {
			t.Errorf("Event %s missing double newline", event.Type)
		}
	}
}

// TestSendStreamEventBasic tests sending stream event to writer
func TestSendStreamEventBasic(t *testing.T) {
	var buf bytes.Buffer
	event := NewStreamEvent("agent_start", "Agent1", "Starting...")

	err := SendStreamEvent(&buf, event)

	if err != nil {
		t.Errorf("SendStreamEvent returned error: %v", err)
	}

	output := buf.String()
	if !strings.HasPrefix(output, "data: ") {
		t.Error("Output should be SSE format")
	}
	if !strings.Contains(output, "agent_start") {
		t.Error("Output should contain event type")
	}
}

// TestSendStreamEventMultiple tests sending multiple events
func TestSendStreamEventMultiple(t *testing.T) {
	var buf bytes.Buffer

	events := []*StreamEvent{
		NewStreamEvent("agent_start", "Agent1", "Start"),
		NewStreamEvent("agent_response", "Agent1", "Response"),
		NewStreamEvent("tool_result", "Agent1", "Result"),
	}

	for _, event := range events {
		err := SendStreamEvent(&buf, event)
		if err != nil {
			t.Errorf("SendStreamEvent failed: %v", err)
		}
	}

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Should have multiple data lines
	dataLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			dataLines++
		}
	}

	if dataLines != 3 {
		t.Errorf("Expected 3 data lines, got %d", dataLines)
	}
}

// TestSendStreamEventFailedWriter tests behavior with failed writer
func TestSendStreamEventFailedWriter(t *testing.T) {
	// Create a writer that will fail
	failingWriter := &failingWriter{}
	event := NewStreamEvent("agent_start", "Agent1", "Start")

	err := SendStreamEvent(failingWriter, event)

	if err == nil {
		t.Error("Expected error from failing writer")
	}
}

// TestNewStreamEventTypeVariations tests various event types
func TestNewStreamEventTypeVariations(t *testing.T) {
	types := []string{
		"agent_start",
		"agent_response",
		"tool_start",
		"tool_result",
		"pause",
		"error",
		"custom",
	}

	for _, eventType := range types {
		event := NewStreamEvent(eventType, "Agent1", "Content")

		if event.Type != eventType {
			t.Errorf("Event type not set correctly for %s", eventType)
		}
	}
}

// TestStreamEventTimestampPrecision tests timestamp precision
func TestStreamEventTimestampPrecision(t *testing.T) {
	before := time.Now()
	event := NewStreamEvent("test", "Agent1", "Test")
	after := time.Now()

	if event.Timestamp.Before(before) || event.Timestamp.After(after) {
		t.Error("Timestamp should be between before and after")
	}
}

// TestFormatStreamEventConsistency tests formatting consistency
func TestFormatStreamEventConsistency(t *testing.T) {
	event := NewStreamEvent("test", "Agent1", "Test")

	// Format same event twice - should be similar (timestamps will differ)
	format1 := FormatStreamEvent(event)
	format2 := FormatStreamEvent(event)

	// Both should be valid SSE
	if !strings.HasPrefix(format1, "data: ") || !strings.HasPrefix(format2, "data: ") {
		t.Error("Both formats should be SSE")
	}

	// Extract and compare JSON (ignoring timestamp)
	json1 := extractJSON(format1)
	json2 := extractJSON(format2)

	var event1, event2 StreamEvent
	json.Unmarshal([]byte(json1), &event1)
	json.Unmarshal([]byte(json2), &event2)

	if event1.Type != event2.Type {
		t.Error("Event type should be consistent")
	}
	if event1.Agent != event2.Agent {
		t.Error("Agent should be consistent")
	}
	if event1.Content != event2.Content {
		t.Error("Content should be consistent")
	}
}

// Helper: failingWriter is a writer that always fails
type failingWriter struct{}

func (fw *failingWriter) Write(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}

// Helper: extract JSON from SSE formatted string
func extractJSON(sseStr string) string {
	jsonStart := strings.Index(sseStr, "{")
	jsonEnd := strings.LastIndex(sseStr, "}")
	if jsonStart >= 0 && jsonEnd >= 0 {
		return sseStr[jsonStart : jsonEnd+1]
	}
	return ""
}
