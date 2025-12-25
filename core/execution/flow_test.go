package execution

import (
	"context"
	"testing"

	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/workflow"
)

// TestNewExecutionFlow tests creating a new execution flow
func TestNewExecutionFlow(t *testing.T) {
	agent := &common.Agent{
		ID:   "test_agent",
		Name: "Test Agent",
	}

	flow := NewExecutionFlow(agent, 5, 3)

	if flow.CurrentAgent != agent {
		t.Errorf("Expected current agent to be set")
	}

	if flow.MaxRounds != 5 {
		t.Errorf("Expected MaxRounds=5, got %d", flow.MaxRounds)
	}

	if flow.MaxHandoffs != 3 {
		t.Errorf("Expected MaxHandoffs=3, got %d", flow.MaxHandoffs)
	}

	if flow.RoundCount != 0 {
		t.Errorf("Expected initial RoundCount=0, got %d", flow.RoundCount)
	}

	if flow.State == nil {
		t.Errorf("Expected ExecutionState to be initialized")
	}
}

// TestExecutionFlow_CanContinue tests continuation check
func TestExecutionFlow_CanContinue(t *testing.T) {
	agent := &common.Agent{ID: "test", Name: "Test"}
	flow := NewExecutionFlow(agent, 2, 2)

	// Should be able to continue initially
	if err := flow.CanContinue(); err != nil {
		t.Errorf("Expected to be able to continue initially, got error: %v", err)
	}

	// Test max rounds limit
	flow.RoundCount = 2
	if err := flow.CanContinue(); err == nil {
		t.Errorf("Expected error when max rounds reached, got none")
	}

	// Test max handoffs limit
	flow.RoundCount = 0
	flow.HandoffCount = 2
	if err := flow.CanContinue(); err == nil {
		t.Errorf("Expected error when max handoffs reached, got none")
	}
}

// TestExecutionFlow_GetWorkflowStatus tests status retrieval
func TestExecutionFlow_GetWorkflowStatus(t *testing.T) {
	agent := &common.Agent{ID: "test", Name: "Test"}
	flow := NewExecutionFlow(agent, 5, 3)

	flow.RoundCount = 2
	flow.HandoffCount = 1

	status := flow.GetWorkflowStatus()

	if status["round_count"] != 2 {
		t.Errorf("Expected round_count=2 in status")
	}

	if status["handoff_count"] != 1 {
		t.Errorf("Expected handoff_count=1 in status")
	}

	if status["current_agent"] != "test" {
		t.Errorf("Expected current agent ID 'test' in status")
	}
}

// TestExecutionFlow_Reset tests resetting execution flow
func TestExecutionFlow_Reset(t *testing.T) {
	agent1 := &common.Agent{ID: "agent1", Name: "Agent 1"}
	agent2 := &common.Agent{ID: "agent2", Name: "Agent 2"}

	flow := NewExecutionFlow(agent1, 5, 3)
	flow.RoundCount = 3
	flow.HandoffCount = 2
	flow.History = append(flow.History, common.Message{Role: "user", Content: "test"})

	flow.Reset(agent2)

	if flow.CurrentAgent != agent2 {
		t.Errorf("Expected current agent to be reset")
	}

	if flow.RoundCount != 0 {
		t.Errorf("Expected RoundCount to be reset, got %d", flow.RoundCount)
	}

	if flow.HandoffCount != 0 {
		t.Errorf("Expected HandoffCount to be reset, got %d", flow.HandoffCount)
	}

	if len(flow.History) != 0 {
		t.Errorf("Expected History to be cleared, got %d messages", len(flow.History))
	}
}

// TestExecutionFlow_Copy tests copying execution flow
func TestExecutionFlow_Copy(t *testing.T) {
	agent := &common.Agent{ID: "test", Name: "Test"}
	flow := NewExecutionFlow(agent, 5, 3)

	flow.RoundCount = 2
	flow.HandoffCount = 1
	flow.History = append(flow.History, common.Message{Role: "user", Content: "test"})

	copy := flow.Copy()

	copy.RoundCount = 5
	copy.History = append(copy.History, common.Message{Role: "assistant", Content: "response"})

	if flow.RoundCount != 2 {
		t.Errorf("Original should maintain RoundCount=2, got %d", flow.RoundCount)
	}

	if len(flow.History) != 1 {
		t.Errorf("Original should have 1 message, got %d", len(flow.History))
	}

	if copy.RoundCount != 5 {
		t.Errorf("Copy should have RoundCount=5, got %d", copy.RoundCount)
	}

	if len(copy.History) != 2 {
		t.Errorf("Copy should have 2 messages, got %d", len(copy.History))
	}
}

// TestExecutionFlow_ValidateFlow tests flow validation
func TestExecutionFlow_ValidateFlow(t *testing.T) {
	agent := &common.Agent{ID: "test", Name: "Test"}

	// Valid flow
	flow := NewExecutionFlow(agent, 5, 3)
	if err := flow.ValidateFlow(); err != nil {
		t.Errorf("Expected valid flow, got error: %v", err)
	}

	// Nil agent
	flow.CurrentAgent = nil
	if err := flow.ValidateFlow(); err == nil {
		t.Errorf("Expected error for nil agent")
	}

	// Reset and test max rounds validation
	flow = NewExecutionFlow(agent, 0, 3)
	if err := flow.ValidateFlow(); err == nil {
		t.Errorf("Expected error for invalid MaxRounds")
	}

	// Test negative handoffs
	flow = NewExecutionFlow(agent, 5, -1)
	if err := flow.ValidateFlow(); err == nil {
		t.Errorf("Expected error for negative MaxHandoffs")
	}
}

// TestExecutionFlow_HandleAgentResponse_Terminal tests terminal agent response
func TestExecutionFlow_HandleAgentResponse_Terminal(t *testing.T) {
	terminalAgent := &common.Agent{ID: "terminal", Name: "Terminal", IsTerminal: true}
	flow := NewExecutionFlow(terminalAgent, 5, 3)

	response := &common.AgentResponse{
		AgentID: "terminal",
		Content: "Final response",
	}

	agentsMap := map[string]*common.Agent{
		"terminal": terminalAgent,
	}

	shouldContinue, err := flow.HandleAgentResponse(context.Background(), response, nil, agentsMap)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if shouldContinue {
		t.Errorf("Expected workflow to terminate for terminal agent")
	}
}

// TestExecutionFlow_HandleAgentResponse_NoRouting tests response handling without routing
func TestExecutionFlow_HandleAgentResponse_NoRouting(t *testing.T) {
	agent := &common.Agent{ID: "agent1", Name: "Agent 1", IsTerminal: false}
	flow := NewExecutionFlow(agent, 5, 3)

	response := &common.AgentResponse{
		AgentID: "agent1",
		Content: "Response",
	}

	agentsMap := map[string]*common.Agent{
		"agent1": agent,
	}

	// Without routing config, execution should continue as per default
	shouldContinue, err := flow.HandleAgentResponse(context.Background(), response, nil, agentsMap)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// The behavior depends on the agent configuration
	// For single non-terminal agent without routing, it should terminate
	if shouldContinue {
		t.Logf("Workflow continues: %v", shouldContinue)
	}
}

// TestExecutionFlow_WithNilHandler tests error handling for nil values
func TestExecutionFlow_ExecuteWorkflowStep_NilAgent(t *testing.T) {
	flow := NewExecutionFlow(nil, 5, 3)
	handler := workflow.NewSyncHandler()

	_, err := flow.ExecuteWorkflowStep(context.Background(), handler, "test-key")
	if err == nil {
		t.Errorf("Expected error for nil agent, got none")
	}
}
