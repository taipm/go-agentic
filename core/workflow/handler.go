// Package workflow provides workflow orchestration and execution functionality.
package workflow

import (
	"github.com/taipm/go-agentic/core/common"
)

// OutputHandler defines the interface for handling workflow output
type OutputHandler interface {
	// HandleStreamEvent processes a stream event during execution
	HandleStreamEvent(event *StreamEvent) error

	// HandleAgentResponse processes an agent response
	HandleAgentResponse(response *common.AgentResponse) error

	// HandleError processes an error that occurred during execution
	HandleError(err error) error

	// GetFinalResponse returns the final response
	GetFinalResponse() interface{}
}

// StreamEvent represents an event streamed during workflow execution
type StreamEvent struct {
	Type      string // "agent_response", "tool_result", "error", "status", etc.
	AgentName string
	Message   string
	Timestamp int64
}

// SyncHandler handles synchronous execution
type SyncHandler struct {
	finalResponse interface{}
	lastError     error
}

// NewSyncHandler creates a new synchronous handler
func NewSyncHandler() *SyncHandler {
	return &SyncHandler{}
}

// HandleStreamEvent processes a stream event synchronously
func (sh *SyncHandler) HandleStreamEvent(event *StreamEvent) error {
	if event == nil {
		return nil
	}

	// Store the event for retrieval
	sh.finalResponse = event
	return nil
}

// HandleAgentResponse processes an agent response synchronously
func (sh *SyncHandler) HandleAgentResponse(response *common.AgentResponse) error {
	if response == nil {
		return nil
	}

	sh.finalResponse = response
	return nil
}

// HandleError processes an error synchronously
func (sh *SyncHandler) HandleError(err error) error {
	sh.lastError = err
	return err
}

// GetFinalResponse returns the final response
func (sh *SyncHandler) GetFinalResponse() interface{} {
	return sh.finalResponse
}

// StreamHandler handles streaming execution with a channel
type StreamHandler struct {
	streamChan    chan *StreamEvent
	finalResponse interface{}
	lastError     error
}

// NewStreamHandler creates a new streaming handler
func NewStreamHandler(streamChan chan *StreamEvent) *StreamHandler {
	return &StreamHandler{
		streamChan: streamChan,
	}
}

// HandleStreamEvent sends a stream event to the channel
func (sh *StreamHandler) HandleStreamEvent(event *StreamEvent) error {
	if sh.streamChan == nil {
		return nil
	}

	select {
	case sh.streamChan <- event:
		return nil
	default:
		return nil // Non-blocking send
	}
}

// HandleAgentResponse sends an agent response as a stream event
func (sh *StreamHandler) HandleAgentResponse(response *common.AgentResponse) error {
	if response == nil {
		return nil
	}

	event := &StreamEvent{
		Type:      "agent_response",
		AgentName: response.AgentName,
		Message:   response.Content,
	}

	return sh.HandleStreamEvent(event)
}

// HandleError sends an error as a stream event
func (sh *StreamHandler) HandleError(err error) error {
	sh.lastError = err

	if err == nil {
		return nil
	}

	event := &StreamEvent{
		Type:    "error",
		Message: err.Error(),
	}

	return sh.HandleStreamEvent(event)
}

// GetFinalResponse returns the final response
func (sh *StreamHandler) GetFinalResponse() interface{} {
	return sh.finalResponse
}

// NoOpHandler is a handler that does nothing (for testing)
type NoOpHandler struct{}

// HandleStreamEvent does nothing
func (nh *NoOpHandler) HandleStreamEvent(event *StreamEvent) error {
	return nil
}

// HandleAgentResponse does nothing
func (nh *NoOpHandler) HandleAgentResponse(response *common.AgentResponse) error {
	return nil
}

// HandleError does nothing
func (nh *NoOpHandler) HandleError(err error) error {
	return nil
}

// GetFinalResponse returns nil
func (nh *NoOpHandler) GetFinalResponse() interface{} {
	return nil
}
