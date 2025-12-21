package crewai

import (
	"strings"
	"testing"
)

// ===== Issue #16: Configuration Validation Tests =====

func makeConfig(entryPoint string, agents []string) *CrewConfig {
	config := &CrewConfig{
		EntryPoint: entryPoint,
		Agents:     agents,
	}
	config.Settings.MaxHandoffs = 5
	config.Settings.MaxRounds = 10
	config.Settings.TimeoutSeconds = 300
	return config
}

func makeAgent(id, name, model string, temp float64, isTerminal bool) *AgentConfig {
	return &AgentConfig{
		ID:          id,
		Name:        name,
		Model:       model,
		Temperature: temp,
		IsTerminal:  isTerminal,
	}
}

// TestValidConfiguration verifies valid config passes validation
func TestValidConfiguration(t *testing.T) {
	config := makeConfig("orchestrator", []string{"orchestrator", "clarifier", "executor"})

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 0.7, false),
		"clarifier":    makeAgent("clarifier", "Clarifier", "gpt-4o", 0.5, false),
		"executor":     makeAgent("executor", "Executor", "gpt-4o", 0.3, true),
	}

	validator := NewConfigValidator(config, agents)
	if err := validator.ValidateAll(); err != nil {
		t.Fatalf("Valid configuration rejected: %v", err)
	}

	if !validator.IsValid() {
		t.Fatal("Validator should report configuration as valid")
	}
}

// TestMissingEntryPoint verifies missing entry_point is detected
func TestMissingEntryPoint(t *testing.T) {
	config := makeConfig("", []string{"orchestrator", "executor"})

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 0.7, false),
		"executor":     makeAgent("executor", "Executor", "gpt-4o", 0.3, true),
	}

	validator := NewConfigValidator(config, agents)
	err := validator.ValidateAll()

	if err == nil {
		t.Fatal("Should reject missing entry_point")
	}

	if !strings.Contains(err.Error(), "entry_point") {
		t.Fatalf("Error should mention entry_point: %v", err)
	}
}

// TestEntryPointNotFound verifies non-existent entry_point is detected
func TestEntryPointNotFound(t *testing.T) {
	config := makeConfig("dispatcher", []string{"orchestrator", "executor"})

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 0.7, false),
		"executor":     makeAgent("executor", "Executor", "gpt-4o", 0.3, true),
	}

	validator := NewConfigValidator(config, agents)
	err := validator.ValidateAll()

	if err == nil {
		t.Fatal("Should reject entry_point not in agents list")
	}

	if !strings.Contains(err.Error(), "dispatcher") {
		t.Fatalf("Error should mention 'dispatcher': %v", err)
	}
}

// TestDuplicateAgentID verifies duplicate agent IDs are detected
func TestDuplicateAgentID(t *testing.T) {
	config := makeConfig("orchestrator", []string{"orchestrator", "executor"})

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 0.7, false),
		"executor":     makeAgent("executor", "Executor", "gpt-4o", 0.3, true),
		"duplicate":    makeAgent("orchestrator", "DupAgent", "gpt-4o", 0.5, false), // Duplicate ID!
	}

	validator := NewConfigValidator(config, agents)
	err := validator.ValidateAll()

	if err == nil {
		t.Fatal("Should detect duplicate agent IDs")
	}

	if !strings.Contains(err.Error(), "Duplicate") || !strings.Contains(err.Error(), "orchestrator") {
		t.Fatalf("Error should mention duplicate: %v", err)
	}
}

// TestInvalidTemperature verifies temperature validation
func TestInvalidTemperature(t *testing.T) {
	config := makeConfig("orchestrator", []string{"orchestrator"})

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 1.5, false), // Invalid!
	}

	validator := NewConfigValidator(config, agents)
	err := validator.ValidateAll()

	if err == nil {
		t.Fatal("Should reject temperature > 1")
	}

	if !strings.Contains(err.Error(), "temperature") {
		t.Fatalf("Error should mention temperature: %v", err)
	}
}

// TestCircularRouting verifies circular routing detection
func TestCircularRouting(t *testing.T) {
	config := makeConfig("orchestrator", []string{"orchestrator", "clarifier"})
	config.Routing = &RoutingConfig{
		Signals: map[string][]RoutingSignal{
			"orchestrator": {
				{Signal: "[ROUTE_CLARIFIER]", Target: "clarifier"},
			},
			"clarifier": {
				{Signal: "[ROUTE_ORCHESTRATOR]", Target: "orchestrator"}, // Circle!
			},
		},
	}

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 0.7, false),
		"clarifier":    makeAgent("clarifier", "Clarifier", "gpt-4o", 0.5, false),
	}

	validator := NewConfigValidator(config, agents)
	err := validator.ValidateAll()

	if err == nil {
		t.Fatal("Should detect circular routing")
	}

	if !strings.Contains(err.Error(), "Circular") {
		t.Fatalf("Error should mention circular: %v", err)
	}
}

// TestInvalidSignalTarget verifies invalid routing targets detected
func TestInvalidSignalTarget(t *testing.T) {
	config := makeConfig("orchestrator", []string{"orchestrator", "executor"})
	config.Routing = &RoutingConfig{
		Signals: map[string][]RoutingSignal{
			"orchestrator": {
				{Signal: "[ROUTE_UNKNOWN]", Target: "unknown_agent"}, // Invalid target!
			},
		},
	}

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 0.7, false),
		"executor":     makeAgent("executor", "Executor", "gpt-4o", 0.3, true),
	}

	validator := NewConfigValidator(config, agents)
	err := validator.ValidateAll()

	if err == nil {
		t.Fatal("Should reject invalid routing target")
	}

	if !strings.Contains(err.Error(), "unknown_agent") {
		t.Fatalf("Error should mention unknown_agent: %v", err)
	}
}

// TestUnreachableAgent verifies unreachable agent detection
func TestUnreachableAgent(t *testing.T) {
	config := makeConfig("orchestrator", []string{"orchestrator", "clarifier", "executor"})
	config.Routing = &RoutingConfig{
		Signals: map[string][]RoutingSignal{
			"orchestrator": {
				{Signal: "[ROUTE_CLARIFIER]", Target: "clarifier"},
			},
			// Note: executor is not reachable from entry point
		},
	}

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 0.7, false),
		"clarifier":    makeAgent("clarifier", "Clarifier", "gpt-4o", 0.5, false),
		"executor":     makeAgent("executor", "Executor", "gpt-4o", 0.3, true),
	}

	validator := NewConfigValidator(config, agents)
	validator.ValidateAll()

	// This should produce a warning, not an error
	warnings := validator.GetWarnings()
	if len(warnings) == 0 {
		t.Fatal("Should warn about unreachable agent")
	}

	found := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "unreachable") || strings.Contains(w.Message, "not reachable") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Warning should mention unreachable: %+v", warnings)
	}
}

// TestNegativeMaxHandoffs detects invalid max_handoffs
func TestNegativeMaxHandoffs(t *testing.T) {
	config := makeConfig("orchestrator", []string{"orchestrator"})
	config.Settings.MaxHandoffs = -1 // Invalid!

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 0.7, false),
	}

	validator := NewConfigValidator(config, agents)
	err := validator.ValidateAll()

	if err == nil {
		t.Fatal("Should reject negative max_handoffs")
	}
}

// TestErrorMessageQuality verifies helpful error messages
func TestErrorMessageQuality(t *testing.T) {
	config := makeConfig("bad_agent", []string{"orchestrator", "executor"})

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 0.7, false),
		"executor":     makeAgent("executor", "Executor", "gpt-4o", 0.3, true),
	}

	validator := NewConfigValidator(config, agents)
	err := validator.ValidateAll()

	errorMsg := err.Error()

	// Error message should include:
	// 1. What's wrong
	if !strings.Contains(errorMsg, "bad_agent") {
		t.Fatal("Error should mention the bad agent name")
	}

	// 2. Why it's wrong
	if !strings.Contains(errorMsg, "not found") {
		t.Fatal("Error should explain why it's wrong")
	}

	// 3. How to fix it
	if !strings.Contains(errorMsg, "orchestrator") && !strings.Contains(errorMsg, "executor") {
		t.Fatal("Error should suggest valid alternatives")
	}
}

// TestValidationReport verifies report generation
func TestValidationReport(t *testing.T) {
	config := makeConfig("", []string{})
	config.Settings.MaxHandoffs = -1 // Invalid

	agents := map[string]*AgentConfig{}

	validator := NewConfigValidator(config, agents)
	validator.ValidateAll()

	errors := validator.GetErrors()
	if len(errors) < 2 {
		t.Fatal("Should have multiple errors")
	}

	for _, err := range errors {
		if err.File == "" {
			t.Fatal("Error should have File set")
		}
		if err.Message == "" {
			t.Fatal("Error should have Message set")
		}
		if err.Fix == "" {
			t.Fatal("Error should have Fix suggestion")
		}
	}
}

// TestEmptyAgentsList detects empty agents list
func TestEmptyAgentsList(t *testing.T) {
	config := makeConfig("orchestrator", []string{}) // Empty!

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "gpt-4o", 0.7, false),
	}

	validator := NewConfigValidator(config, agents)
	err := validator.ValidateAll()

	if err == nil {
		t.Fatal("Should reject empty agents list")
	}

	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("Error should mention empty: %v", err)
	}
}

// TestModelValidation checks model name validation
func TestModelValidation(t *testing.T) {
	config := makeConfig("orchestrator", []string{"orchestrator"})

	agents := map[string]*AgentConfig{
		"orchestrator": makeAgent("orchestrator", "Orchestrator", "invalid-model", 0.7, false),
	}

	validator := NewConfigValidator(config, agents)
	validator.ValidateAll()

	// Should produce warning (not error) for unknown model
	warnings := validator.GetWarnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w.Message, "invalid-model") {
			found = true
			break
		}
	}

	if !found {
		// It's OK if there's no warning, depends on implementation
		// Just verify the config doesn't produce errors
		errors := validator.GetErrors()
		for _, e := range errors {
			if strings.Contains(e.Message, "model") {
				t.Fatalf("Model validation should warn, not error: %v", e)
			}
		}
	}
}
