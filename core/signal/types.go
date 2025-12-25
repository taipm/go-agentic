package signal

import (
	"context"
	"time"

	"github.com/taipm/go-agentic/core/common"
)

// Signal represents a named event that can be emitted by agents
type Signal struct {
	Name        string                 // Unique signal identifier
	AgentID     string                 // Agent that emitted the signal
	Data        interface{}            // Signal payload data
	Timestamp   time.Time              // When signal was emitted
	Metadata    map[string]interface{} // Additional metadata
}

// SignalHandler defines the interface for handling signals
type SignalHandler struct {
	ID           string
	Name         string
	Description  string
	TargetAgent  string // Agent to route to when signal matches
	Signals      []string
	Condition    func(signal *Signal) bool // Optional condition for matching
	OnSignal     func(ctx context.Context, signal *Signal) error
}

// SignalMatch represents a successful signal match result
type SignalMatch struct {
	Handler      *SignalHandler
	Signal       *Signal
	Matched      bool
	Confidence   float64 // 0.0-1.0 confidence level
	MatchedAt    time.Time
}

// RoutingDecision represents the decision to route to next agent
type RoutingDecision struct {
	NextAgentID string
	Reason      string
	IsTerminal  bool
	Metadata    map[string]interface{}
}

// SignalConfig defines configuration for signal handling
type SignalConfig struct {
	Enabled            bool
	Timeout            time.Duration
	BufferSize         int
	MaxSignalsPerAgent int
	PersistSignals     bool
}

// DefaultSignalConfig returns default signal configuration
func DefaultSignalConfig() *SignalConfig {
	return &SignalConfig{
		Enabled:            true,
		Timeout:            time.Second * 30,
		BufferSize:         100,
		MaxSignalsPerAgent: 10,
		PersistSignals:     false,
	}
}

// Signal Event Types (constants for standard signals)
const (
	// Agent lifecycle signals
	SignalAgentStart   = "agent:start"
	SignalAgentEnd     = "agent:end"
	SignalAgentError   = "agent:error"
	SignalAgentPause   = "agent:pause"
	SignalAgentResume  = "agent:resume"

	// Tool signals
	SignalToolStart  = "tool:start"
	SignalToolEnd    = "tool:end"
	SignalToolError  = "tool:error"

	// Routing signals
	SignalHandoff       = "route:handoff"
	SignalTerminal      = "route:terminal"
	SignalRoute         = "route:custom"
	SignalWaitForInput  = "route:wait_input"

	// Control signals
	SignalCancel     = "control:cancel"
	SignalTimeout    = "control:timeout"
	SignalAbort      = "control:abort"
	SignalRetry      = "control:retry"
)

// PredefinedHandlers provides factory methods for common handlers
type PredefinedHandlers struct{}

// NewAgentStartHandler creates a handler for agent start signals
func (ph *PredefinedHandlers) NewAgentStartHandler(targetAgent string) *SignalHandler {
	return &SignalHandler{
		ID:          "handler-agent-start",
		Name:        "Agent Start Handler",
		Description: "Handles agent start signals",
		TargetAgent: targetAgent,
		Signals:     []string{SignalAgentStart},
		Condition: func(signal *Signal) bool {
			return signal.Name == SignalAgentStart
		},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			// Default implementation: just log the signal
			return nil
		},
	}
}

// NewAgentErrorHandler creates a handler for agent error signals
func (ph *PredefinedHandlers) NewAgentErrorHandler(targetAgent string) *SignalHandler {
	return &SignalHandler{
		ID:          "handler-agent-error",
		Name:        "Agent Error Handler",
		Description: "Handles agent error signals",
		TargetAgent: targetAgent,
		Signals:     []string{SignalAgentError},
		Condition: func(signal *Signal) bool {
			return signal.Name == SignalAgentError
		},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			// Default implementation: propagate error handling to target agent
			return nil
		},
	}
}

// NewToolErrorHandler creates a handler for tool error signals
func (ph *PredefinedHandlers) NewToolErrorHandler(targetAgent string) *SignalHandler {
	return &SignalHandler{
		ID:          "handler-tool-error",
		Name:        "Tool Error Handler",
		Description: "Handles tool error signals",
		TargetAgent: targetAgent,
		Signals:     []string{SignalToolError},
		Condition: func(signal *Signal) bool {
			return signal.Name == SignalToolError
		},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			// Default implementation: error recovery
			return nil
		},
	}
}

// NewHandoffHandler creates a handler for explicit handoff signals
func (ph *PredefinedHandlers) NewHandoffHandler(targetAgent string) *SignalHandler {
	return &SignalHandler{
		ID:          "handler-handoff",
		Name:        "Handoff Handler",
		Description: "Handles explicit handoff signals",
		TargetAgent: targetAgent,
		Signals:     []string{SignalHandoff},
		Condition: func(signal *Signal) bool {
			return signal.Name == SignalHandoff
		},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			// Default implementation: route to target agent
			return nil
		},
	}
}

// SignalEmitter defines the interface for objects that can emit signals
type SignalEmitter interface {
	EmitSignal(signal *Signal) error
	CanEmitSignal(signalName string) bool
}

// AgentWithSignals extends Agent with signal emission capability
type AgentWithSignals struct {
	*common.Agent
	signals          map[string]*Signal
	allowedSignals   map[string]bool
	signalRegistry   *SignalRegistry
}

// NewAgentWithSignals creates an agent with signal emission capability
func NewAgentWithSignals(agent *common.Agent, registry *SignalRegistry) *AgentWithSignals {
	return &AgentWithSignals{
		Agent:          agent,
		signals:        make(map[string]*Signal),
		allowedSignals: make(map[string]bool),
		signalRegistry: registry,
	}
}

// AllowSignal permits an agent to emit a specific signal
func (aws *AgentWithSignals) AllowSignal(signalName string) {
	aws.allowedSignals[signalName] = true
}

// DisallowSignal prevents an agent from emitting a specific signal
func (aws *AgentWithSignals) DisallowSignal(signalName string) {
	aws.allowedSignals[signalName] = false
}

// EmitSignal emits a signal from the agent
func (aws *AgentWithSignals) EmitSignal(signal *Signal) error {
	if !aws.CanEmitSignal(signal.Name) {
		return &SignalError{
			Code:    "SIGNAL_NOT_ALLOWED",
			Message: "Agent is not allowed to emit this signal: " + signal.Name,
		}
	}

	signal.AgentID = aws.Agent.ID
	signal.Timestamp = time.Now()

	aws.signals[signal.Name] = signal

	// Notify registry if available
	if aws.signalRegistry != nil {
		return aws.signalRegistry.Emit(signal)
	}

	return nil
}

// CanEmitSignal checks if agent can emit a signal
func (aws *AgentWithSignals) CanEmitSignal(signalName string) bool {
	return aws.allowedSignals[signalName]
}

// GetSignal retrieves a previously emitted signal
func (aws *AgentWithSignals) GetSignal(signalName string) *Signal {
	return aws.signals[signalName]
}

// GetAllSignals returns all emitted signals
func (aws *AgentWithSignals) GetAllSignals() map[string]*Signal {
	return aws.signals
}

// ClearSignals clears all emitted signals
func (aws *AgentWithSignals) ClearSignals() {
	aws.signals = make(map[string]*Signal)
}

// SignalError represents a signal-related error
type SignalError struct {
	Code    string
	Message string
	Signal  *Signal
	Context string
}

// Error implements the error interface
func (se *SignalError) Error() string {
	if se.Context != "" {
		return se.Code + ": " + se.Message + " (" + se.Context + ")"
	}
	return se.Code + ": " + se.Message
}

// IsSignalError checks if an error is a SignalError
func IsSignalError(err error) bool {
	_, ok := err.(*SignalError)
	return ok
}

// SignalErrors enum
var (
	ErrSignalNotFound    = &SignalError{Code: "SIGNAL_NOT_FOUND", Message: "Signal not found"}
	ErrHandlerNotFound   = &SignalError{Code: "HANDLER_NOT_FOUND", Message: "Handler not found"}
	ErrSignalTimeout     = &SignalError{Code: "SIGNAL_TIMEOUT", Message: "Signal processing timeout"}
	ErrInvalidSignal     = &SignalError{Code: "INVALID_SIGNAL", Message: "Invalid signal"}
	ErrDuplicateHandler  = &SignalError{Code: "DUPLICATE_HANDLER", Message: "Handler already registered"}
)
