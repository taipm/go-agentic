package routing

import (
	"context"
	"testing"

	"github.com/taipm/go-agentic/core/common"
)

// TestDetermineNextAgent_Terminal tests terminal agent detection
func TestDetermineNextAgent_Terminal(t *testing.T) {
	agent := &common.Agent{
		ID:         "agent1",
		Name:       "Agent 1",
		IsTerminal: true,
	}

	response := &common.AgentResponse{
		AgentID: "agent1",
		Content: "Final response",
	}

	decision, err := DetermineNextAgent(agent, response, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !decision.IsTerminal {
		t.Errorf("Expected terminal decision, got %v", decision.IsTerminal)
	}
}

// TestDetermineNextAgent_NoHandoffTargets tests no routing configured
func TestDetermineNextAgent_NoHandoffTargets(t *testing.T) {
	agent := &common.Agent{
		ID:              "agent1",
		Name:            "Agent 1",
		IsTerminal:      false,
		HandoffTargets:  []*common.Agent{},
	}

	response := &common.AgentResponse{
		AgentID: "agent1",
		Content: "Response",
	}

	decision, err := DetermineNextAgent(agent, response, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !decision.IsTerminal {
		t.Errorf("Expected terminal decision when no handoff targets, got %v", decision.IsTerminal)
	}
}

// TestDetermineNextAgent_WithHandoffTargets tests handoff routing
// Note: Current implementation doesn't fully support handoff routing
// This is a placeholder test documenting the expected behavior for Phase 5
func TestDetermineNextAgent_WithHandoffTargets(t *testing.T) {
	agent2 := &common.Agent{
		ID:   "agent2",
		Name: "Agent 2",
	}

	agent1 := &common.Agent{
		ID:             "agent1",
		Name:           "Agent 1",
		IsTerminal:     false,
		HandoffTargets: []*common.Agent{agent2},
	}

	response := &common.AgentResponse{
		AgentID: "agent1",
		Content: "Handing off to next agent",
	}

	decision, err := DetermineNextAgent(agent1, response, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Current behavior: terminates when no routing config (placeholder for Phase 5)
	// In Phase 5, this should route to agent2
	if !decision.IsTerminal {
		t.Logf("Note: handoff routing will be implemented in Phase 5")
	}
}

// TestDetermineNextAgentWithSignals_NilInput tests nil agent error
func TestDetermineNextAgentWithSignals_NilAgent(t *testing.T) {
	decision, err := DetermineNextAgentWithSignals(context.Background(), nil, nil, nil, nil)
	if err == nil {
		t.Fatalf("Expected error for nil agent, got none")
	}

	if decision != nil {
		t.Errorf("Expected nil decision for nil agent, got %v", decision)
	}
}

// TestRouteBySignal_NotFound tests signal not found error
func TestRouteBySignal_NotFound(t *testing.T) {
	routing := &common.RoutingConfig{
		Signals: make(map[string][]common.RoutingSignal),
	}

	target, err := RouteBySignal("nonexistent", routing)
	if err == nil {
		t.Fatalf("Expected error for nonexistent signal, got none")
	}

	if target != "" {
		t.Errorf("Expected empty target for nonexistent signal, got %s", target)
	}
}

// TestValidateRouting_Valid tests valid routing configuration
func TestValidateRouting_Valid(t *testing.T) {
	agents := map[string]*common.Agent{
		"agent1": {ID: "agent1", Name: "Agent 1"},
		"agent2": {ID: "agent2", Name: "Agent 2"},
	}

	routing := &common.RoutingConfig{
		Signals: map[string][]common.RoutingSignal{
			"agent1": {
				{Signal: "NEXT", Target: "agent2"},
			},
		},
	}

	err := ValidateRouting(routing, agents)
	if err != nil {
		t.Fatalf("Expected no error for valid routing, got %v", err)
	}
}

// TestValidateRouting_MissingAgent tests missing agent validation
func TestValidateRouting_MissingAgent(t *testing.T) {
	agents := map[string]*common.Agent{
		"agent1": {ID: "agent1", Name: "Agent 1"},
	}

	routing := &common.RoutingConfig{
		Signals: map[string][]common.RoutingSignal{
			"agent_missing": {
				{Signal: "NEXT", Target: "agent2"},
			},
		},
	}

	err := ValidateRouting(routing, agents)
	if err == nil {
		t.Fatalf("Expected error for missing agent, got none")
	}
}

// TestValidateRouting_Nil tests nil routing is valid
func TestValidateRouting_Nil(t *testing.T) {
	agents := map[string]*common.Agent{
		"agent1": {ID: "agent1", Name: "Agent 1"},
	}

	err := ValidateRouting(nil, agents)
	if err != nil {
		t.Fatalf("Expected no error for nil routing, got %v", err)
	}
}

// ============================================================================
// BEHAVIOR ROUTING TESTS
// ============================================================================

// TestRouteByBehavior_Valid tests successful behavior lookup
func TestRouteByBehavior_Valid(t *testing.T) {
	routing := &common.RoutingConfig{
		AgentBehaviors: map[string]common.AgentBehavior{
			"standard": {
				WaitForSignal: false,
				AutoRoute:     true,
				IsTerminal:    false,
				Description:   "Standard routing behavior",
			},
		},
	}

	result, err := RouteByBehavior("standard", routing)
	if err != nil {
		t.Fatalf("Expected no error for valid behavior, got %v", err)
	}

	if result != "standard" {
		t.Errorf("Expected behavior name 'standard', got '%s'", result)
	}
}

// TestRouteByBehavior_NilRouting tests nil routing parameter
func TestRouteByBehavior_NilRouting(t *testing.T) {
	result, err := RouteByBehavior("some_behavior", nil)
	if err == nil {
		t.Fatalf("Expected error for nil routing, got none")
	}

	if result != "" {
		t.Errorf("Expected empty result for nil routing, got '%s'", result)
	}

	if err.Error() != "routing configuration is nil" {
		t.Errorf("Expected 'routing configuration is nil' error, got: %v", err)
	}
}

// TestRouteByBehavior_EmptyBehaviorName tests empty behavior name
func TestRouteByBehavior_EmptyBehaviorName(t *testing.T) {
	routing := &common.RoutingConfig{
		AgentBehaviors: map[string]common.AgentBehavior{
			"standard": {IsTerminal: false},
		},
	}

	result, err := RouteByBehavior("", routing)
	if err == nil {
		t.Fatalf("Expected error for empty behavior name, got none")
	}

	if result != "" {
		t.Errorf("Expected empty result for empty behavior, got '%s'", result)
	}

	if err.Error() != "behavior name is empty" {
		t.Errorf("Expected 'behavior name is empty' error, got: %v", err)
	}
}

// TestRouteByBehavior_NilAgentBehaviors tests nil agent behaviors in routing
func TestRouteByBehavior_NilAgentBehaviors(t *testing.T) {
	routing := &common.RoutingConfig{
		AgentBehaviors: nil,
	}

	result, err := RouteByBehavior("standard", routing)
	if err == nil {
		t.Fatalf("Expected error for nil agent behaviors, got none")
	}

	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}

	if err.Error() != "no agent behaviors configured in routing" {
		t.Errorf("Expected 'no agent behaviors configured' error, got: %v", err)
	}
}

// TestRouteByBehavior_EmptyAgentBehaviors tests empty agent behaviors map
func TestRouteByBehavior_EmptyAgentBehaviors(t *testing.T) {
	routing := &common.RoutingConfig{
		AgentBehaviors: make(map[string]common.AgentBehavior),
	}

	result, err := RouteByBehavior("standard", routing)
	if err == nil {
		t.Fatalf("Expected error for empty agent behaviors, got none")
	}

	if result != "" {
		t.Errorf("Expected empty result, got '%s'", result)
	}
}

// TestRouteByBehavior_NotFound tests behavior not found in routing
func TestRouteByBehavior_NotFound(t *testing.T) {
	routing := &common.RoutingConfig{
		AgentBehaviors: map[string]common.AgentBehavior{
			"standard": {IsTerminal: false},
		},
	}

	result, err := RouteByBehavior("nonexistent", routing)
	if err == nil {
		t.Fatalf("Expected error for nonexistent behavior, got none")
	}

	if result != "" {
		t.Errorf("Expected empty result for nonexistent behavior, got '%s'", result)
	}

	expectedErr := "behavior 'nonexistent' not found in routing configuration"
	if err.Error() != expectedErr {
		t.Errorf("Expected '%s' error, got: %v", expectedErr, err)
	}
}

// TestRouteByBehavior_TerminalBehavior tests terminal behavior routing
func TestRouteByBehavior_TerminalBehavior(t *testing.T) {
	routing := &common.RoutingConfig{
		AgentBehaviors: map[string]common.AgentBehavior{
			"terminal": {
				IsTerminal:    true,
				WaitForSignal: false,
				AutoRoute:     false,
				Description:   "Terminal behavior - ends execution",
			},
		},
	}

	result, err := RouteByBehavior("terminal", routing)
	if err != nil {
		t.Fatalf("Expected no error for terminal behavior, got %v", err)
	}

	if result != "terminal" {
		t.Errorf("Expected 'terminal', got '%s'", result)
	}
}

// TestRouteByBehavior_MultipleBehaviors tests routing with multiple behaviors
func TestRouteByBehavior_MultipleBehaviors(t *testing.T) {
	routing := &common.RoutingConfig{
		AgentBehaviors: map[string]common.AgentBehavior{
			"standard": {
				AutoRoute:  true,
				IsTerminal: false,
			},
			"signal_wait": {
				WaitForSignal: true,
				AutoRoute:     false,
				IsTerminal:    false,
			},
			"terminal": {
				IsTerminal: true,
			},
		},
	}

	testCases := []struct {
		behavior string
		expect   string
		wantErr  bool
	}{
		{"standard", "standard", false},
		{"signal_wait", "signal_wait", false},
		{"terminal", "terminal", false},
		{"unknown", "", true},
	}

	for _, tc := range testCases {
		result, err := RouteByBehavior(tc.behavior, routing)
		if tc.wantErr && err == nil {
			t.Errorf("Behavior '%s': expected error, got none", tc.behavior)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("Behavior '%s': expected no error, got %v", tc.behavior, err)
		}
		if result != tc.expect {
			t.Errorf("Behavior '%s': expected '%s', got '%s'", tc.behavior, tc.expect, result)
		}
	}
}
