package agentic

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// FormatStreamEvent formats a StreamEvent as SSE data
// Returns both JSON and plain text formats separated by newlines
func FormatStreamEvent(event *StreamEvent) string {
	// Set timestamp if not already set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Format as JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		jsonData = []byte(`{"type":"error","content":"Failed to marshal event"}`)
	}

	// SSE format: "data: {content}\n\n"
	return fmt.Sprintf("data: %s\n\n", string(jsonData))
}

// SendStreamEvent sends a StreamEvent to the response writer
func SendStreamEvent(w io.Writer, event *StreamEvent) error {
	formatted := FormatStreamEvent(event)
	_, err := fmt.Fprint(w, formatted)
	return err
}

// NewStreamEvent creates a new StreamEvent
func NewStreamEvent(eventType, agent, content string) *StreamEvent {
	return &StreamEvent{
		Type:      eventType,
		Agent:     agent,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// NewStreamEventWithMetadata creates a new StreamEvent with metadata
func NewStreamEventWithMetadata(eventType, agent, content string, metadata interface{}) *StreamEvent {
	return &StreamEvent{
		Type:      eventType,
		Agent:     agent,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
}
