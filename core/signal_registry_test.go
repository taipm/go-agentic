package crewai

import (
	"testing"
)

// TestSignalRegistryBasics tests basic registry operations
func TestSignalRegistryBasics(t *testing.T) {
	registry := NewSignalRegistry()

	// Test empty registry
	if registry.Count() != 0 {
		t.Errorf("New registry should be empty, got %d signals", registry.Count())
	}

	// Register a signal
	def := &SignalDefinition{
		Name:           "[TEST]",
		Description:    "Test signal",
		AllowAllAgents: true,
		Behavior:       SignalBehaviorRoute,
	}
	err := registry.Register(def)
	if err != nil {
		t.Errorf("Failed to register signal: %v", err)
	}

	// Check count
	if registry.Count() != 1 {
		t.Errorf("After register, count should be 1, got %d", registry.Count())
	}

	// Check exists
	if !registry.Exists("[TEST]") {
		t.Error("Signal [TEST] should exist after registration")
	}

	// Get signal
	retrieved := registry.Get("[TEST]")
	if retrieved == nil {
		t.Error("Get should return the signal")
	}
	if retrieved.Name != "[TEST]" {
		t.Errorf("Retrieved signal should be [TEST], got %s", retrieved.Name)
	}
}

// TestSignalRegistryDuplicate tests that duplicate registration fails
func TestSignalRegistryDuplicate(t *testing.T) {
	registry := NewSignalRegistry()
	def := &SignalDefinition{Name: "[DUP]", Behavior: SignalBehaviorRoute}

	err1 := registry.Register(def)
	if err1 != nil {
		t.Errorf("First registration should succeed: %v", err1)
	}

	err2 := registry.Register(def)
	if err2 == nil {
		t.Error("Duplicate registration should fail")
	}
}

// TestLoadDefaultSignals tests loading default signal set
func TestLoadDefaultSignals(t *testing.T) {
	registry := LoadDefaultSignals()

	// Check some default signals exist
	defaultSignals := []string{"[END]", "[END_EXAM]", "[DONE]", "[NEXT]", "[OK]"}
	for _, signal := range defaultSignals {
		if !registry.Exists(signal) {
			t.Errorf("Default signal %s should exist", signal)
		}
	}

	// Check count is reasonable
	if registry.Count() < len(defaultSignals) {
		t.Errorf("Should have at least %d signals, got %d", len(defaultSignals), registry.Count())
	}
}

// TestSignalValidatorEmission tests signal emission validation
func TestSignalValidatorEmission(t *testing.T) {
	registry := LoadDefaultSignals()
	validator := NewSignalValidator(registry)

	tests := []struct {
		signal    string
		agentID   string
		wantError bool
		errMsg    string
	}{
		{"[END]", "agent1", false, ""},
		{"[UNKNOWN]", "agent1", true, "not registered"},
		{"INVALID", "agent1", true, "invalid format"},
		{"[]", "agent1", true, "invalid format"},
	}

	for _, tt := range tests {
		err := validator.ValidateSignalEmission(tt.signal, tt.agentID)
		if tt.wantError && err == nil {
			t.Errorf("Signal %s: expected error, got nil", tt.signal)
		}
		if !tt.wantError && err != nil {
			t.Errorf("Signal %s: unexpected error: %v", tt.signal, err)
		}
	}
}

// TestSignalValidatorTarget tests signal target validation
func TestSignalValidatorTarget(t *testing.T) {
	registry := LoadDefaultSignals()
	validator := NewSignalValidator(registry)

	agents := map[string]bool{
		"teacher":  true,
		"reporter": true,
	}

	tests := []struct {
		signal      string
		agentID     string
		target      string
		wantError   bool
		description string
	}{
		// Termination signals require empty target
		{"[END]", "teacher", "", false, "Termination with empty target"},
		{"[END]", "teacher", "reporter", true, "Termination with target"},
		// Routing signals require valid target
		{"[NEXT]", "teacher", "reporter", false, "Routing with valid target"},
		{"[NEXT]", "teacher", "", true, "Routing with empty target"},
		{"[NEXT]", "teacher", "unknown", true, "Routing with unknown target"},
	}

	for _, tt := range tests {
		err := validator.ValidateSignalTarget(tt.signal, tt.agentID, tt.target, agents)
		if tt.wantError && err == nil {
			t.Errorf("%s: expected error, got nil", tt.description)
		}
		if !tt.wantError && err != nil {
			t.Errorf("%s: unexpected error: %v", tt.description, err)
		}
	}
}

// TestSignalValidatorFullConfig tests full configuration validation
func TestSignalValidatorFullConfig(t *testing.T) {
	registry := LoadDefaultSignals()
	validator := NewSignalValidator(registry)

	// Valid config
	validSignals := map[string][]RoutingSignal{
		"teacher": {
			{Signal: "[NEXT]", Target: "reporter"},
			{Signal: "[END]", Target: ""},
		},
		"reporter": {
			{Signal: "[DONE]", Target: ""},
		},
	}

	agents := map[string]bool{
		"teacher":  true,
		"reporter": true,
	}

	errors := validator.ValidateConfiguration(validSignals, agents)
	if len(errors) > 0 {
		t.Errorf("Valid config should have no errors, got: %v", errors)
	}

	// Invalid config - unknown agent
	invalidSignals := map[string][]RoutingSignal{
		"unknown_agent": {
			{Signal: "[NEXT]", Target: "teacher"},
		},
	}

	errors = validator.ValidateConfiguration(invalidSignals, agents)
	if len(errors) == 0 {
		t.Error("Invalid config should have errors")
	}
}

// TestSignalMatchingInContent tests signal detection in text
func TestSignalMatchingInContent(t *testing.T) {
	registry := LoadDefaultSignals()
	validator := NewSignalValidator(registry)

	tests := []struct {
		signal   string
		content  string
		wantMatch bool
		method   string
	}{
		{"[END]", "Task complete. [END]", true, "exact"},
		{"[END]", "Task complete. [end]", true, "case_insensitive"},
		{"[END]", "Task complete.", false, ""},
		{"[NEXT]", "Ready for next step. [NEXT]", true, "exact"},
	}

	for _, tt := range tests {
		matches, method := validator.ValidateSignalInContent(tt.signal, tt.content)
		if matches != tt.wantMatch {
			t.Errorf("Signal %s in %q: expected match=%v, got %v", tt.signal, tt.content, tt.wantMatch, matches)
		}
		if matches && method != tt.method {
			t.Errorf("Signal %s: expected method %s, got %s", tt.signal, tt.method, method)
		}
	}
}

// TestSignalRegistryBulk tests bulk registration
func TestSignalRegistryBulk(t *testing.T) {
	registry := NewSignalRegistry()

	signals := []*SignalDefinition{
		{Name: "[A]", Behavior: SignalBehaviorRoute},
		{Name: "[B]", Behavior: SignalBehaviorRoute},
		{Name: "[C]", Behavior: SignalBehaviorTerminate},
	}

	err := registry.RegisterBulk(signals)
	if err != nil {
		t.Errorf("Bulk registration failed: %v", err)
	}

	if registry.Count() != 3 {
		t.Errorf("After bulk register 3 signals, count should be 3, got %d", registry.Count())
	}
}

// TestSignalBehaviorGrouping tests retrieving signals by behavior
func TestSignalBehaviorGrouping(t *testing.T) {
	registry := LoadDefaultSignals()

	termSignals := registry.GetTerminationSignals()
	if len(termSignals) == 0 {
		t.Error("Should have termination signals")
	}

	routingSignals := registry.GetRoutingSignals()
	if len(routingSignals) == 0 {
		t.Error("Should have routing signals")
	}

	// Verify correct behavior
	for _, sig := range termSignals {
		if sig.Behavior != SignalBehaviorTerminate {
			t.Errorf("Termination signal has wrong behavior: %s", sig.Behavior)
		}
	}
}

// TestSignalReportGeneration tests report generation
func TestSignalReportGeneration(t *testing.T) {
	registry := LoadDefaultSignals()
	validator := NewSignalValidator(registry)

	report := validator.GenerateSignalReport()
	if len(report) == 0 {
		t.Error("Report should not be empty")
	}

	// Check report contains expected content
	if !stringContains(report, "SIGNAL REGISTRY REPORT") {
		t.Error("Report should contain title")
	}

	if !stringContains(report, "Signals Registered") {
		t.Error("Report should show count")
	}
}

// stringContains checks if string contains substring
func stringContains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
