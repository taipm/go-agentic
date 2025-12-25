package signal

import (
	"context"
	"testing"
	"time"

	"github.com/taipm/go-agentic/core/common"
)

// TestSignalHandler_NewHandler tests creating a new handler
func TestSignalHandler_NewHandler(t *testing.T) {
	handler := NewHandler()

	if handler == nil {
		t.Fatal("Expected handler, got nil")
	}

	if handler.GetHandlerCount() != 0 {
		t.Error("Expected 0 handlers initially")
	}
}

// TestSignalHandler_Register tests registering a handler
func TestSignalHandler_Register(t *testing.T) {
	handler := NewHandler()

	sh := &SignalHandler{
		ID:          "test-handler",
		Name:        "Test Handler",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
	}

	err := handler.Register(sh)
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	if handler.GetHandlerCount() != 1 {
		t.Error("Expected 1 handler after registration")
	}

	retrieved := handler.GetHandler("test-handler")
	if retrieved == nil {
		t.Error("Expected to retrieve handler")
	}

	if retrieved.ID != "test-handler" {
		t.Error("Retrieved handler has wrong ID")
	}
}

// TestSignalHandler_Unregister tests unregistering a handler
func TestSignalHandler_Unregister(t *testing.T) {
	handler := NewHandler()

	sh := &SignalHandler{
		ID:          "test-handler",
		Name:        "Test Handler",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
	}

	handler.Register(sh)

	err := handler.Unregister("test-handler")
	if err != nil {
		t.Fatalf("Failed to unregister handler: %v", err)
	}

	if handler.GetHandlerCount() != 0 {
		t.Error("Expected 0 handlers after unregistration")
	}
}

// TestSignalHandler_FindHandlers tests finding handlers for a signal
func TestSignalHandler_FindHandlers(t *testing.T) {
	handler := NewHandler()

	sh := &SignalHandler{
		ID:          "handler-1",
		Name:        "Handler 1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
	}

	handler.Register(sh)

	signal := &Signal{
		Name:    SignalHandoff,
		AgentID: "agent-0",
	}

	found := handler.FindHandlers(signal)
	if len(found) != 1 {
		t.Errorf("Expected 1 handler, found %d", len(found))
	}

	if found[0].ID != "handler-1" {
		t.Error("Found wrong handler")
	}
}

// TestSignalHandler_ProcessSignal tests processing a signal
func TestSignalHandler_ProcessSignal(t *testing.T) {
	handler := NewHandler()

	sh := &SignalHandler{
		ID:          "handler-1",
		Name:        "Handler 1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			return nil
		},
	}

	handler.Register(sh)

	signal := &Signal{
		Name:    SignalHandoff,
		AgentID: "agent-0",
	}

	decision, err := handler.ProcessSignal(context.Background(), signal)
	if err != nil {
		t.Fatalf("Failed to process signal: %v", err)
	}

	if decision == nil {
		t.Fatal("Expected routing decision")
	}

	if decision.NextAgentID != "agent-1" {
		t.Errorf("Expected target agent 'agent-1', got '%s'", decision.NextAgentID)
	}

	if decision.IsTerminal {
		t.Error("Expected non-terminal decision")
	}
}

// TestSignalHandler_ValidateHandlers tests handler validation
func TestSignalHandler_ValidateHandlers(t *testing.T) {
	handler := NewHandler()

	// Invalid handler (no ID)
	sh := &SignalHandler{
		Name:        "Invalid",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
	}

	handler.Register(sh)

	err := handler.ValidateHandlers()
	if err == nil {
		t.Error("Expected validation error for missing ID")
	}
}

// TestSignalRegistry_NewSignalRegistry tests creating registry
func TestSignalRegistry_NewSignalRegistry(t *testing.T) {
	registry := NewSignalRegistry()

	if registry == nil {
		t.Fatal("Expected registry, got nil")
	}

	stats := registry.GetStats()
	if stats["handler_count"] != 0 {
		t.Error("Expected 0 handlers initially")
	}
}

// TestSignalRegistry_RegisterHandler tests registering handler
func TestSignalRegistry_RegisterHandler(t *testing.T) {
	registry := NewSignalRegistry()

	sh := &SignalHandler{
		ID:          "handler-1",
		Name:        "Handler 1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
	}

	err := registry.RegisterHandler(sh)
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	stats := registry.GetStats()
	if stats["handler_count"] != 1 {
		t.Error("Expected 1 handler")
	}
}

// TestSignalRegistry_Emit tests emitting a signal
func TestSignalRegistry_Emit(t *testing.T) {
	registry := NewSignalRegistry()

	sh := &SignalHandler{
		ID:          "handler-1",
		Name:        "Handler 1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			return nil
		},
	}

	registry.RegisterHandler(sh)

	signal := &Signal{
		Name:    SignalHandoff,
		AgentID: "agent-0",
	}

	err := registry.Emit(signal)
	if err != nil {
		t.Fatalf("Failed to emit signal: %v", err)
	}

	// Give registry time to process
	time.Sleep(100 * time.Millisecond)

	// Check agent info was recorded
	info := registry.GetAgentSignalInfo("agent-0")
	if info == nil {
		t.Error("Expected agent signal info")
	}
}

// TestSignalRegistry_ProcessSignal tests processing signal
func TestSignalRegistry_ProcessSignal(t *testing.T) {
	registry := NewSignalRegistry()

	sh := &SignalHandler{
		ID:          "handler-1",
		Name:        "Handler 1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			return nil
		},
	}

	registry.RegisterHandler(sh)

	signal := &Signal{
		Name:    SignalHandoff,
		AgentID: "agent-0",
	}

	decision, err := registry.ProcessSignal(context.Background(), signal)
	if err != nil {
		t.Fatalf("Failed to process signal: %v", err)
	}

	if decision == nil {
		t.Fatal("Expected routing decision")
	}

	if decision.NextAgentID != "agent-1" {
		t.Errorf("Expected target 'agent-1', got '%s'", decision.NextAgentID)
	}
}

// TestSignalRegistry_AllowAgentSignal tests allowing agent signals
func TestSignalRegistry_AllowAgentSignal(t *testing.T) {
	registry := NewSignalRegistry()

	err := registry.AllowAgentSignal("agent-1", SignalHandoff)
	if err != nil {
		t.Fatalf("Failed to allow signal: %v", err)
	}

	allowed := registry.IsAgentSignalAllowed("agent-1", SignalHandoff)
	if !allowed {
		t.Error("Expected signal to be allowed")
	}
}

// TestSignalRegistry_GetAgentEmittedSignals tests retrieving signals
func TestSignalRegistry_GetAgentEmittedSignals(t *testing.T) {
	registry := NewSignalRegistry()

	registry.AllowAgentSignal("agent-1", SignalHandoff)

	sh := &SignalHandler{
		ID:          "handler-1",
		Name:        "Handler 1",
		TargetAgent: "agent-2",
		Signals:     []string{SignalHandoff},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			return nil
		},
	}

	registry.RegisterHandler(sh)

	signal := &Signal{
		Name:    SignalHandoff,
		AgentID: "agent-1",
	}

	registry.Emit(signal)
	time.Sleep(100 * time.Millisecond)

	emitted := registry.GetAgentEmittedSignals("agent-1")
	if len(emitted) == 0 {
		t.Error("Expected emitted signals")
	}
}

// TestSignalRegistry_Listen tests listening for signals
func TestSignalRegistry_Listen(t *testing.T) {
	registry := NewSignalRegistry()

	received := false

	registry.Listen([]string{SignalHandoff}, func(signal *Signal) {
		received = true
	})

	sh := &SignalHandler{
		ID:          "handler-1",
		Name:        "Handler 1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			return nil
		},
	}

	registry.RegisterHandler(sh)

	signal := &Signal{
		Name:    SignalHandoff,
		AgentID: "agent-0",
	}

	registry.Emit(signal)
	time.Sleep(100 * time.Millisecond)

	if !received {
		t.Error("Expected listener callback")
	}
}

// TestSignalRegistry_ValidateConfiguration tests validation
func TestSignalRegistry_ValidateConfiguration(t *testing.T) {
	registry := NewSignalRegistry()

	sh := &SignalHandler{
		ID:          "handler-1",
		Name:        "Handler 1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
	}

	registry.RegisterHandler(sh)

	err := registry.ValidateConfiguration()
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
}

// TestSignalRegistry_GetStats tests stats retrieval
func TestSignalRegistry_GetStats(t *testing.T) {
	registry := NewSignalRegistry()

	sh := &SignalHandler{
		ID:          "handler-1",
		Name:        "Handler 1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
	}

	registry.RegisterHandler(sh)

	stats := registry.GetStats()
	if stats["handler_count"] != 1 {
		t.Error("Expected 1 handler in stats")
	}

	if stats["enabled"] != true {
		t.Error("Expected signals to be enabled")
	}
}

// TestAgentWithSignals tests agent signal emission
func TestAgentWithSignals(t *testing.T) {
	registry := NewSignalRegistry()

	agent := &common.Agent{
		ID:   "agent-1",
		Name: "Test Agent",
	}

	agentWithSignals := NewAgentWithSignals(agent, registry)

	agentWithSignals.AllowSignal(SignalHandoff)

	if !agentWithSignals.CanEmitSignal(SignalHandoff) {
		t.Error("Expected agent to be able to emit signal")
	}

	signal := &Signal{
		Name: SignalHandoff,
	}

	err := agentWithSignals.EmitSignal(signal)
	if err != nil {
		t.Fatalf("Failed to emit signal: %v", err)
	}

	retrieved := agentWithSignals.GetSignal(SignalHandoff)
	if retrieved == nil {
		t.Error("Expected to retrieve emitted signal")
	}

	if retrieved.AgentID != agent.ID {
		t.Error("Signal should have agent ID")
	}
}

// TestSignalHandler_ProcessSignalWithPriority tests priority handling
func TestSignalHandler_ProcessSignalWithPriority(t *testing.T) {
	handler := NewHandler()

	sh1 := &SignalHandler{
		ID:          "handler-1",
		Name:        "Handler 1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			return nil
		},
	}

	sh2 := &SignalHandler{
		ID:          "handler-2",
		Name:        "Handler 2",
		TargetAgent: "agent-2",
		Signals:     []string{SignalHandoff},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			return nil
		},
	}

	handler.Register(sh1)
	handler.Register(sh2)

	signal := &Signal{
		Name:    SignalHandoff,
		AgentID: "agent-0",
	}

	// Set priority: handler-2 is higher priority
	priority := []string{"handler-2", "handler-1"}

	decision, err := handler.ProcessSignalWithPriority(context.Background(), signal, priority)
	if err != nil {
		t.Fatalf("Failed to process signal: %v", err)
	}

	if decision.NextAgentID != "agent-2" {
		t.Errorf("Expected agent-2 (higher priority), got %s", decision.NextAgentID)
	}
}

// TestSignalRegistry_Close tests closing registry
func TestSignalRegistry_Close(t *testing.T) {
	registry := NewSignalRegistry()

	err := registry.Close()
	if err != nil {
		t.Fatalf("Failed to close registry: %v", err)
	}

	if !registry.IsClosed() {
		t.Error("Expected registry to be closed")
	}

	// Try to emit after close
	signal := &Signal{Name: SignalHandoff}
	err = registry.Emit(signal)
	if err == nil {
		t.Error("Expected error when emitting to closed registry")
	}
}

// TestSignalHandler_ListHandlers tests listing all handlers
func TestSignalHandler_ListHandlers(t *testing.T) {
	handler := NewHandler()

	sh1 := &SignalHandler{
		ID:          "handler-1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
	}

	sh2 := &SignalHandler{
		ID:          "handler-2",
		TargetAgent: "agent-2",
		Signals:     []string{SignalTerminal},
	}

	handler.Register(sh1)
	handler.Register(sh2)

	handlers := handler.ListHandlers()
	if len(handlers) != 2 {
		t.Errorf("Expected 2 handlers, got %d", len(handlers))
	}
}

// TestSignalRegistry_ClearSignalHistory tests clearing history
func TestSignalRegistry_ClearSignalHistory(t *testing.T) {
	registry := NewSignalRegistry()

	sh := &SignalHandler{
		ID:          "handler-1",
		TargetAgent: "agent-1",
		Signals:     []string{SignalHandoff},
		OnSignal: func(ctx context.Context, signal *Signal) error {
			return nil
		},
	}

	registry.RegisterHandler(sh)

	signal := &Signal{Name: SignalHandoff, AgentID: "agent-0"}
	registry.ProcessSignal(context.Background(), signal)

	history := registry.GetSignalHistory()
	if len(history) == 0 {
		t.Error("Expected signal in history")
	}

	registry.ClearSignalHistory()
	history = registry.GetSignalHistory()
	if len(history) != 0 {
		t.Error("Expected empty history after clear")
	}
}
