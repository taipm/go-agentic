package signal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/taipm/go-agentic/core/common"
)

// Error message constant to avoid duplication
const errSignalsDisabled = "Signal handling is disabled"

// SignalRegistry manages signal emission, registration, and routing
type SignalRegistry struct {
	handler       *Handler
	signals       chan *Signal
	listeners     map[string][]func(*Signal)
	mu            sync.RWMutex
	config        *SignalConfig
	closed        bool
	agentRegistry map[string]AgentSignalInfo
}

// AgentSignalInfo tracks signal information per agent
type AgentSignalInfo struct {
	AgentID        string
	EmittedSignals []string
	AllowedSignals []string
	LastSignalTime time.Time
}

// checkEnabled validates if signal handling is enabled
func (sr *SignalRegistry) checkEnabled() error {
	if !sr.config.Enabled {
		return &SignalError{
			Code:    "SIGNALS_DISABLED",
			Message: errSignalsDisabled,
		}
	}
	return nil
}

// NewSignalRegistry creates a new signal registry with default configuration
func NewSignalRegistry() *SignalRegistry {
	return NewSignalRegistryWithConfig(nil)
}

// NewSignalRegistryWithConfig creates a registry with custom configuration
func NewSignalRegistryWithConfig(config *SignalConfig) *SignalRegistry {
	if config == nil {
		config = DefaultSignalConfig()
	}

	return &SignalRegistry{
		handler:       NewHandlerWithConfig(config),
		signals:       make(chan *Signal, config.BufferSize),
		listeners:     make(map[string][]func(*Signal)),
		config:        config,
		agentRegistry: make(map[string]AgentSignalInfo),
	}
}

// RegisterHandler registers a signal handler
func (sr *SignalRegistry) RegisterHandler(handler *SignalHandler) error {
	if err := sr.checkEnabled(); err != nil {
		return err
	}
	return sr.handler.Register(handler)
}

// UnregisterHandler removes a signal handler
func (sr *SignalRegistry) UnregisterHandler(handlerID string) error {
	return sr.handler.Unregister(handlerID)
}

// GetHandler retrieves a handler by ID
func (sr *SignalRegistry) GetHandler(handlerID string) *SignalHandler {
	return sr.handler.GetHandler(handlerID)
}

// Emit broadcasts a signal to all listeners and returns routing decision
func (sr *SignalRegistry) Emit(signal *Signal) error {
	if signal == nil {
		return ErrInvalidSignal
	}

	if err := sr.checkEnabled(); err != nil {
		return err
	}

	sr.mu.RLock()
	if sr.closed {
		sr.mu.RUnlock()
		return &SignalError{
			Code:    "REGISTRY_CLOSED",
			Message: "Signal registry is closed",
		}
	}
	sr.mu.RUnlock()

	signal.Timestamp = time.Now()

	// Send to internal channel (non-blocking)
	select {
	case sr.signals <- signal:
	default:
		// Buffer full, skip queuing
	}

	// Notify listeners
	sr.notifyListeners(signal)

	// Update agent registry
	sr.recordAgentSignal(signal.AgentID, signal.Name)

	return nil
}

// Listen registers a listener for specific signal types
func (sr *SignalRegistry) Listen(signalNames []string, callback func(*Signal)) string {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	listenerID := fmt.Sprintf("listener_%d", time.Now().UnixNano())

	for _, signalName := range signalNames {
		sr.listeners[signalName] = append(sr.listeners[signalName], callback)
	}

	return listenerID
}

// notifyListeners notifies all registered listeners
func (sr *SignalRegistry) notifyListeners(signal *Signal) {
	sr.mu.RLock()
	listeners := sr.listeners[signal.Name]
	sr.mu.RUnlock()

	for _, listener := range listeners {
		if listener != nil {
			listener(signal)
		}
	}
}

// ProcessSignal processes a signal through registered handlers
func (sr *SignalRegistry) ProcessSignal(ctx context.Context, signal *Signal) (*common.RoutingDecision, error) {
	if err := sr.checkEnabled(); err != nil {
		return nil, err
	}
	return sr.handler.ProcessSignal(ctx, signal)
}

// ProcessSignalWithPriority processes signal with priority handling
func (sr *SignalRegistry) ProcessSignalWithPriority(ctx context.Context, signal *Signal, priority []string) (*common.RoutingDecision, error) {
	if err := sr.checkEnabled(); err != nil {
		return nil, err
	}
	return sr.handler.ProcessSignalWithPriority(ctx, signal, priority)
}

// getOrCreateAgentInfo retrieves or creates agent signal info
func (sr *SignalRegistry) getOrCreateAgentInfo(agentID string) AgentSignalInfo {
	info, exists := sr.agentRegistry[agentID]
	if !exists {
		info = AgentSignalInfo{
			AgentID:        agentID,
			EmittedSignals: []string{},
			AllowedSignals: []string{},
		}
	}
	return info
}

// recordAgentSignal records signal emission by agent
func (sr *SignalRegistry) recordAgentSignal(agentID, signalName string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	info := sr.getOrCreateAgentInfo(agentID)

	info.EmittedSignals = append(info.EmittedSignals, signalName)
	info.LastSignalTime = time.Now()

	// Limit history per agent
	if len(info.EmittedSignals) > sr.config.MaxSignalsPerAgent {
		info.EmittedSignals = info.EmittedSignals[len(info.EmittedSignals)-sr.config.MaxSignalsPerAgent:]
	}

	sr.agentRegistry[agentID] = info
}

// GetAgentSignalInfo retrieves signal information for an agent
func (sr *SignalRegistry) GetAgentSignalInfo(agentID string) *AgentSignalInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	info, exists := sr.agentRegistry[agentID]
	if !exists {
		return nil
	}

	return &info
}

// GetSignalHistory returns signal history
func (sr *SignalRegistry) GetSignalHistory() []*SignalMatch {
	return sr.handler.GetSignalHistory()
}

// ClearSignalHistory clears the history
func (sr *SignalRegistry) ClearSignalHistory() {
	sr.handler.ClearSignalHistory()
}

// AllowAgentSignal allows an agent to emit a signal
func (sr *SignalRegistry) AllowAgentSignal(agentID, signalName string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	info := sr.getOrCreateAgentInfo(agentID)
	info.AllowedSignals = append(info.AllowedSignals, signalName)
	sr.agentRegistry[agentID] = info

	return nil
}

// IsAgentSignalAllowed checks if agent can emit signal
func (sr *SignalRegistry) IsAgentSignalAllowed(agentID, signalName string) bool {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	info, exists := sr.agentRegistry[agentID]
	if !exists {
		return false
	}

	for _, allowed := range info.AllowedSignals {
		if allowed == signalName || allowed == "*" {
			return true
		}
	}

	return false
}

// GetAgentEmittedSignals returns all signals emitted by an agent
func (sr *SignalRegistry) GetAgentEmittedSignals(agentID string) []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	info, exists := sr.agentRegistry[agentID]
	if !exists {
		return []string{}
	}

	// Return copy
	signals := make([]string, len(info.EmittedSignals))
	copy(signals, info.EmittedSignals)
	return signals
}

// Drain processes all queued signals (blocking)
func (sr *SignalRegistry) Drain(ctx context.Context) error {
	for {
		select {
		case signal := <-sr.signals:
			if signal == nil {
				return nil
			}
			// Process signal if needed
			_, _ = sr.ProcessSignal(ctx, signal)

		case <-ctx.Done():
			return ctx.Err()

		default:
			return nil
		}
	}
}

// Start starts processing signals (runs in goroutine)
func (sr *SignalRegistry) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case signal := <-sr.signals:
				if signal == nil {
					return
				}
				_, _ = sr.ProcessSignal(ctx, signal)

			case <-ctx.Done():
				sr.Close()
				return
			}
		}
	}()
}

// Close closes the signal registry
func (sr *SignalRegistry) Close() error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.closed {
		return nil
	}

	sr.closed = true
	close(sr.signals)
	return nil
}

// IsClosed checks if registry is closed
func (sr *SignalRegistry) IsClosed() bool {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	return sr.closed
}

// ValidateConfiguration validates the registry configuration
func (sr *SignalRegistry) ValidateConfiguration() error {
	// Validate handlers
	if err := sr.handler.ValidateHandlers(); err != nil {
		return err
	}

	// Validate config
	if sr.config.BufferSize <= 0 {
		return &SignalError{
			Code:    "INVALID_CONFIG",
			Message: "Buffer size must be positive",
		}
	}

	if sr.config.Timeout <= 0 {
		return &SignalError{
			Code:    "INVALID_CONFIG",
			Message: "Timeout must be positive",
		}
	}

	return nil
}

// GetStats returns registry statistics
func (sr *SignalRegistry) GetStats() map[string]interface{} {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	return map[string]interface{}{
		"handler_count":    sr.handler.GetHandlerCount(),
		"agent_count":      len(sr.agentRegistry),
		"listener_count":   len(sr.listeners),
		"signal_history":   len(sr.handler.GetSignalHistory()),
		"enabled":          sr.config.Enabled,
		"closed":           sr.closed,
		"buffer_size":      sr.config.BufferSize,
	}
}
