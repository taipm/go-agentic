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
