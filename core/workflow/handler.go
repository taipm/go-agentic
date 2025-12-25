// Package workflow provides workflow orchestration and execution functionality.
package workflow

import (
	"github.com/taipm/go-agentic/core/common"
)

// OutputHandler defines the interface for handling workflow output
type OutputHandler interface {
	// HandleStreamEvent processes a stream event during execution
	HandleStreamEvent(event *common.StreamEvent) error

	// HandleAgentResponse processes an agent response
	HandleAgentResponse(response *common.AgentResponse) error

	// HandleError processes an error that occurred during execution
	HandleError(err error) error

	// GetFinalResponse returns the final response
	GetFinalResponse() interface{}
}

// Handler is the unified implementation for all output handling strategies
type Handler struct {
	finalResponse interface{}
	lastError     error
	streamChan    chan *common.StreamEvent
	strategy      handlerStrategy
}

// handlerStrategy defines how to handle stream events
type handlerStrategy interface {
	handleStreamEvent(*Handler, *common.StreamEvent) error
}

// syncStrategy stores events in memory
type syncStrategy struct{}

func (s *syncStrategy) handleStreamEvent(h *Handler, event *common.StreamEvent) error {
	h.finalResponse = event
	return nil
}

// streamStrategy sends events to a channel
type streamStrategy struct{}

func (s *streamStrategy) handleStreamEvent(h *Handler, event *common.StreamEvent) error {
	if h.streamChan == nil {
		return nil
	}
	select {
	case h.streamChan <- event:
		return nil
	default:
		return nil // Non-blocking send
	}
}

// noOpStrategy ignores all events
type noOpStrategy struct{}

func (s *noOpStrategy) handleStreamEvent(h *Handler, event *common.StreamEvent) error {
	return nil
}

// HandleStreamEvent processes a stream event
func (h *Handler) HandleStreamEvent(event *common.StreamEvent) error {
	if event == nil {
		return nil
	}
	return h.strategy.handleStreamEvent(h, event)
}

// HandleAgentResponse processes an agent response
func (h *Handler) HandleAgentResponse(response *common.AgentResponse) error {
	if response == nil {
		return nil
	}

	// For memory-based handlers, store the response
	if _, ok := h.strategy.(*syncStrategy); ok || h.strategy == nil {
		h.finalResponse = response
	}

	// For stream-based handlers, convert to stream event
	if _, ok := h.strategy.(*streamStrategy); ok {
		event := &common.StreamEvent{
			Type:    "agent_response",
			Agent:   response.AgentName,
			Content: response.Content,
		}
		return h.HandleStreamEvent(event)
	}

	return nil
}

// HandleError processes an error
func (h *Handler) HandleError(err error) error {
	h.lastError = err

	if err == nil {
		return nil
	}

	// For noOp handler, do nothing
	if _, ok := h.strategy.(*noOpStrategy); ok {
		return nil
	}

	event := &common.StreamEvent{
		Type:    "error",
		Content: err.Error(),
	}

	return h.HandleStreamEvent(event)
}

// GetFinalResponse returns the final response
func (h *Handler) GetFinalResponse() interface{} {
	return h.finalResponse
}

// NewSyncHandler creates a new synchronous handler that stores responses in memory
func NewSyncHandler() OutputHandler {
	return &Handler{
		strategy: &syncStrategy{},
	}
}

// NewStreamHandler creates a new streaming handler that sends events to a channel
func NewStreamHandler(streamChan chan *common.StreamEvent) OutputHandler {
	return &Handler{
		streamChan: streamChan,
		strategy:   &streamStrategy{},
	}
}

// NewNoOpHandler creates a handler that discards all output (for testing)
func NewNoOpHandler() OutputHandler {
	return &Handler{
		strategy: &noOpStrategy{},
	}
}
