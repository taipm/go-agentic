package crewai

import (
	"testing"
)

// Helper function to create a test agent
func createTestAgent(id, name, role string) *Agent {
	return &Agent{
		ID:   id,
		Name: name,
		Role: role,
		Primary: &ModelConfig{
			Model:    "gpt-4",
			Provider: "openai",
		},
		Tools: []*Tool{},
	}
}

// TestCrewExecutorWithRegistry tests CrewExecutor with signal registry integration
func TestCrewExecutorWithRegistry(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			createTestAgent("teacher", "Teacher", "Quiz Master"),
			createTestAgent("reporter", "Reporter", "Report Handler"),
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"teacher": {
					{Signal: "[NEXT]", Target: "reporter"},
					{Signal: "[END_EXAM]", Target: ""},
				},
				"reporter": {
					{Signal: "[DONE]", Target: ""},
				},
			},
		},
	}

	// Create executor
	executor := NewCrewExecutor(crew, "test-api-key")
	if executor == nil {
		t.Fatal("Failed to create executor")
	}

	// Test 1: Validation without registry should pass (Phase 2)
	if err := executor.ValidateSignals(); err != nil {
		t.Errorf("Phase 2 validation should pass: %v", err)
	}

	// Test 2: Set registry
	registry := LoadDefaultSignals()
	executor.SetSignalRegistry(registry)

	// Test 3: Validation with registry should pass
	if err := executor.ValidateSignals(); err != nil {
		t.Errorf("Phase 3.5 validation with registry should pass: %v", err)
	}

	// Test 4: Registry should be set
	if executor.signalRegistry == nil {
		t.Error("Signal registry should be set after SetSignalRegistry()")
	}
}

// TestCrewExecutorWithoutRegistry tests CrewExecutor without signal registry
// This ensures backward compatibility - Phase 2 validation alone should work
func TestCrewExecutorWithoutRegistry(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			createTestAgent("agent1", "Agent 1", "Role"),
			createTestAgent("agent2", "Agent 2", "Role"),
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"agent1": {
					{Signal: "[NEXT]", Target: "agent2"},
					{Signal: "[END]", Target: ""},
				},
			},
		},
	}

	executor := NewCrewExecutor(crew, "test-api-key")

	// Should validate successfully without registry (Phase 2 validation)
	if err := executor.ValidateSignals(); err != nil {
		t.Errorf("Should validate without registry: %v", err)
	}

	// Registry should be nil
	if executor.signalRegistry != nil {
		t.Error("Signal registry should be nil when not set")
	}
}

// TestCrewExecutorRegistryWithInvalidSignal tests that registry validation catches invalid signals
func TestCrewExecutorRegistryWithInvalidSignal(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			createTestAgent("agent1", "Agent 1", "Role"),
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"agent1": {
					// Using invalid signal not in registry
					{Signal: "[UNKNOWN_SIGNAL]", Target: ""},
				},
			},
		},
	}

	executor := NewCrewExecutor(crew, "test-api-key")
	registry := LoadDefaultSignals()
	executor.SetSignalRegistry(registry)

	// Validation should fail because [UNKNOWN_SIGNAL] is not in registry
	err := executor.ValidateSignals()
	if err == nil {
		t.Error("Should fail validation for unknown signal in registry")
	}
}

// TestCrewExecutorRegistryWithTerminationSignalError tests termination signal target validation
func TestCrewExecutorRegistryWithTerminationSignalError(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			createTestAgent("agent1", "Agent 1", "Role"),
			createTestAgent("agent2", "Agent 2", "Role"),
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"agent1": {
					// Termination signal with target (should fail)
					{Signal: "[END]", Target: "agent2"},
				},
			},
		},
	}

	executor := NewCrewExecutor(crew, "test-api-key")
	registry := LoadDefaultSignals()
	executor.SetSignalRegistry(registry)

	// Validation should fail: termination signal must have empty target
	err := executor.ValidateSignals()
	if err == nil {
		t.Error("Should fail: termination signal with non-empty target")
	}
}

// TestCrewExecutorRegistryWithRoutingSignal tests routing signal with registry
func TestCrewExecutorRegistryWithRoutingSignal(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			createTestAgent("teacher", "Teacher", "Role"),
			createTestAgent("student", "Student", "Role"),
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"teacher": {
					// Routing signal with valid target
					{Signal: "[NEXT]", Target: "student"},
					{Signal: "[END]", Target: ""},
				},
				"student": {
					{Signal: "[DONE]", Target: ""},
				},
			},
		},
	}

	executor := NewCrewExecutor(crew, "test-api-key")
	registry := LoadDefaultSignals()
	executor.SetSignalRegistry(registry)

	// Validation should pass
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Valid routing signals should pass: %v", err)
	}
}

// TestSetSignalRegistryNilExecutor tests SetSignalRegistry with nil executor (safety check)
func TestSetSignalRegistryNilExecutor(t *testing.T) {
	var executor *CrewExecutor
	registry := LoadDefaultSignals()

	// Should not panic even with nil executor
	executor.SetSignalRegistry(registry)
	// If we reach here without panic, test passes
}

// TestCrewExecutorMultipleSignalsWithRegistry tests multiple signals validation
func TestCrewExecutorMultipleSignalsWithRegistry(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			createTestAgent("agent1", "Agent 1", "Role"),
			createTestAgent("agent2", "Agent 2", "Role"),
			createTestAgent("agent3", "Agent 3", "Role"),
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"agent1": {
					{Signal: "[NEXT]", Target: "agent2"},
					{Signal: "[ANSWER]", Target: "agent3"},
					{Signal: "[END]", Target: ""},
				},
				"agent2": {
					{Signal: "[OK]", Target: "agent1"},
					{Signal: "[ERROR]", Target: "agent3"},
					{Signal: "[DONE]", Target: ""},
				},
				"agent3": {
					{Signal: "[RETRY]", Target: "agent1"},
					{Signal: "[DONE]", Target: ""},
				},
			},
		},
	}

	executor := NewCrewExecutor(crew, "test-api-key")
	registry := LoadDefaultSignals()
	executor.SetSignalRegistry(registry)

	// All these signals are in the default registry
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Multiple built-in signals should validate: %v", err)
	}
}

// TestCrewExecutorCustomSignalsWithRegistry tests custom signals with registry
func TestCrewExecutorCustomSignalsWithRegistry(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			createTestAgent("agent1", "Agent 1", "Role"),
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"agent1": {
					{Signal: "[CUSTOM]", Target: ""},
				},
			},
		},
	}

	executor := NewCrewExecutor(crew, "test-api-key")

	// Create custom registry with the custom signal
	registry := NewSignalRegistry()
	registry.Register(&SignalDefinition{
		Name:           "[CUSTOM]",
		Description:    "Custom signal",
		AllowAllAgents: true,
		Behavior:       SignalBehaviorTerminate,
	})
	executor.SetSignalRegistry(registry)

	// Should validate with custom signal in registry
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Custom signal should validate when in registry: %v", err)
	}
}

// TestCrewExecutorBackwardCompatibility tests Phase 2 validation still works
func TestCrewExecutorBackwardCompatibility(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			createTestAgent("a1", "A1", "Role"),
		},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"a1": {
					{Signal: "[TEST]", Target: ""},
				},
			},
		},
	}

	// Old code (Phase 2 only - no registry)
	executor := NewCrewExecutor(crew, "api-key")
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Phase 2 validation should work without registry: %v", err)
	}

	// New code (Phase 3.5 - with registry)
	executor.SetSignalRegistry(LoadDefaultSignals())
	// Registry should be set
	if executor.signalRegistry == nil {
		t.Error("Registry should be set")
	}
}

// TestCrewExecutorNoSignalsNoRegistry tests with empty signals
func TestCrewExecutorNoSignalsNoRegistry(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			createTestAgent("agent", "Agent", "Role"),
		},
		// No routing configured
	}

	executor := NewCrewExecutor(crew, "api-key")

	// Should pass even with no signals and no registry
	err := executor.ValidateSignals()
	if err != nil {
		t.Errorf("Should pass with no signals: %v", err)
	}
}
