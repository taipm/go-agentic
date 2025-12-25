package signal

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Handler provides core signal handling functionality
type Handler struct {
	config         *SignalConfig
	handlers       map[string]*SignalHandler
	mu             sync.RWMutex
	signalHistory  []*SignalMatch
	maxHistorySize int
}

// NewHandler creates a new signal handler with default configuration
func NewHandler() *Handler {
	return NewHandlerWithConfig(DefaultSignalConfig())
}

// NewHandlerWithConfig creates a new signal handler with custom configuration
func NewHandlerWithConfig(config *SignalConfig) *Handler {
	if config == nil {
		config = DefaultSignalConfig()
	}

	return &Handler{
		config:         config,
		handlers:       make(map[string]*SignalHandler),
		signalHistory:  make([]*SignalMatch, 0),
		maxHistorySize: 100,
	}
}

// Register registers a signal handler
func (h *Handler) Register(handler *SignalHandler) error {
	if handler == nil {
		return &SignalError{
			Code:    "INVALID_HANDLER",
			Message: "Handler cannot be nil",
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.handlers[handler.ID]; exists {
		return ErrDuplicateHandler
	}

	h.handlers[handler.ID] = handler
	return nil
}

// Unregister removes a signal handler
func (h *Handler) Unregister(handlerID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.handlers[handlerID]; !exists {
		return ErrHandlerNotFound
	}

	delete(h.handlers, handlerID)
	return nil
}

// GetHandler retrieves a handler by ID
func (h *Handler) GetHandler(handlerID string) *SignalHandler {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.handlers[handlerID]
}

// FindHandlers finds all handlers that can handle a signal
func (h *Handler) FindHandlers(signal *Signal) []*SignalHandler {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var matching []*SignalHandler

	for _, handler := range h.handlers {
		if h.handlerMatchesSignal(handler, signal) {
			matching = append(matching, handler)
		}
	}

	return matching
}

// handlerMatchesSignal checks if a handler matches a signal
func (h *Handler) handlerMatchesSignal(handler *SignalHandler, signal *Signal) bool {
	// Check if handler is listening for this signal
	for _, s := range handler.Signals {
		if s == signal.Name {
			// If handler has a condition, check it
			if handler.Condition != nil {
				return handler.Condition(signal)
			}
			return true
		}
	}

	// If handler says "*", it matches any signal
	for _, s := range handler.Signals {
		if s == "*" {
			if handler.Condition != nil {
				return handler.Condition(signal)
			}
			return true
		}
	}

	return false
}

// ProcessSignal processes a signal and returns routing decision
func (h *Handler) ProcessSignal(ctx context.Context, signal *Signal) (*RoutingDecision, error) {
	if signal == nil {
		return nil, ErrInvalidSignal
	}

	// Find matching handlers
	handlers := h.FindHandlers(signal)
	if len(handlers) == 0 {
		return nil, ErrSignalNotFound
	}

	// Execute first handler
	handler := handlers[0]
	if handler.OnSignal != nil {
		err := handler.OnSignal(ctx, signal)
		if err != nil {
			return nil, err
		}
	}

	// Record match in history
	h.recordMatch(&SignalMatch{
		Handler:    handler,
		Signal:     signal,
		Matched:    true,
		Confidence: 1.0,
		MatchedAt:  time.Now(),
	})

	// Return routing decision
	return &RoutingDecision{
		NextAgentID: handler.TargetAgent,
		Reason:      fmt.Sprintf("Signal '%s' matched handler '%s'", signal.Name, handler.Name),
		IsTerminal:  signal.Name == SignalTerminal,
		Metadata: map[string]interface{}{
			"signal_name":  signal.Name,
			"handler_id":   handler.ID,
			"handler_name": handler.Name,
		},
	}, nil
}

// ProcessSignalWithPriority processes signal using priority-based handler selection
func (h *Handler) ProcessSignalWithPriority(ctx context.Context, signal *Signal, priority []string) (*RoutingDecision, error) {
	if signal == nil {
		return nil, ErrInvalidSignal
	}

	// Find handlers matching the signal
	handlers := h.FindHandlers(signal)
	if len(handlers) == 0 {
		return nil, ErrSignalNotFound
	}

	// Sort handlers by priority
	priorityMap := make(map[string]int)
	for i, id := range priority {
		priorityMap[id] = i
	}

	// Select highest priority handler
	var selectedHandler *SignalHandler
	minPriority := len(priority)

	for _, handler := range handlers {
		if p, ok := priorityMap[handler.ID]; ok && p < minPriority {
			minPriority = p
			selectedHandler = handler
		} else if !ok && selectedHandler == nil {
			// Use first unprioritized handler if no prioritized found
			selectedHandler = handler
		}
	}

	if selectedHandler == nil {
		selectedHandler = handlers[0]
	}

	// Execute handler
	if selectedHandler.OnSignal != nil {
		err := selectedHandler.OnSignal(ctx, signal)
		if err != nil {
			return nil, err
		}
	}

	// Record match
	h.recordMatch(&SignalMatch{
		Handler:    selectedHandler,
		Signal:     signal,
		Matched:    true,
		Confidence: 1.0,
		MatchedAt:  time.Now(),
	})

	// Return routing decision
	return &RoutingDecision{
		NextAgentID: selectedHandler.TargetAgent,
		Reason:      fmt.Sprintf("Signal '%s' matched handler '%s' (priority)", signal.Name, selectedHandler.Name),
		IsTerminal:  signal.Name == SignalTerminal,
		Metadata: map[string]interface{}{
			"signal_name":  signal.Name,
			"handler_id":   selectedHandler.ID,
			"priority":     minPriority,
		},
	}, nil
}

// recordMatch records a signal match in history
func (h *Handler) recordMatch(match *SignalMatch) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.signalHistory = append(h.signalHistory, match)

	// Maintain max history size
	if len(h.signalHistory) > h.maxHistorySize {
		h.signalHistory = h.signalHistory[1:]
	}
}

// GetSignalHistory returns the signal match history
func (h *Handler) GetSignalHistory() []*SignalMatch {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy
	history := make([]*SignalMatch, len(h.signalHistory))
	copy(history, h.signalHistory)
	return history
}

// ClearSignalHistory clears the signal history
func (h *Handler) ClearSignalHistory() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.signalHistory = make([]*SignalMatch, 0)
}

// GetHandlerCount returns the number of registered handlers
func (h *Handler) GetHandlerCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.handlers)
}

// ListHandlers returns all registered handlers
func (h *Handler) ListHandlers() []*SignalHandler {
	h.mu.RLock()
	defer h.mu.RUnlock()

	handlers := make([]*SignalHandler, 0, len(h.handlers))
	for _, handler := range h.handlers {
		handlers = append(handlers, handler)
	}
	return handlers
}

// ValidateHandlers validates all registered handlers
func (h *Handler) ValidateHandlers() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, handler := range h.handlers {
		if handler.ID == "" {
			return &SignalError{
				Code:    "INVALID_HANDLER_ID",
				Message: "Handler ID cannot be empty",
			}
		}

		if handler.Name == "" {
			return &SignalError{
				Code:    "INVALID_HANDLER_NAME",
				Message: "Handler name cannot be empty",
			}
		}

		if handler.TargetAgent == "" {
			return &SignalError{
				Code:    "INVALID_TARGET_AGENT",
				Message: "Handler target agent cannot be empty",
			}
		}

		if len(handler.Signals) == 0 {
			return &SignalError{
				Code:    "INVALID_SIGNALS",
				Message: "Handler must listen for at least one signal",
			}
		}
	}

	return nil
}

// WithTimeout wraps signal processing with timeout
func (h *Handler) WithTimeout(duration time.Duration) func(context.Context, *Signal) (*RoutingDecision, error) {
	return func(baseCtx context.Context, signal *Signal) (*RoutingDecision, error) {
		ctx, cancel := context.WithTimeout(baseCtx, duration)
		defer cancel()

		return h.ProcessSignal(ctx, signal)
	}
}
