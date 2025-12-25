package executor

import (
	"testing"
	"time"

	"github.com/taipm/go-agentic/core/common"
)

// TestHistoryManager_Add tests adding messages to history
func TestHistoryManager_Add(t *testing.T) {
	hm := NewHistoryManager()

	msg := common.Message{
		Role:    "user",
		Content: "Hello",
	}

	hm.Add(msg)

	if hm.Length() != 1 {
		t.Errorf("Expected length 1, got %d", hm.Length())
	}

	messages := hm.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Content != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", messages[0].Content)
	}
}

// TestHistoryManager_AddMessages tests adding multiple messages
func TestHistoryManager_AddMessages(t *testing.T) {
	hm := NewHistoryManager()

	msgs := []common.Message{
		{Role: "user", Content: "First"},
		{Role: "assistant", Content: "Second"},
		{Role: "user", Content: "Third"},
	}

	hm.AddMessages(msgs)

	if hm.Length() != 3 {
		t.Errorf("Expected length 3, got %d", hm.Length())
	}
}

// TestHistoryManager_Clear tests clearing history
func TestHistoryManager_Clear(t *testing.T) {
	hm := NewHistoryManager()

	hm.Add(common.Message{Role: "user", Content: "Test"})
	if hm.Length() != 1 {
		t.Fatal("Expected length 1 after adding")
	}

	hm.Clear()

	if hm.Length() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", hm.Length())
	}
}

// TestHistoryManager_GetRecentMessages tests getting recent messages
func TestHistoryManager_GetRecentMessages(t *testing.T) {
	hm := NewHistoryManager()

	for i := 0; i < 5; i++ {
		msg := common.Message{Role: "user", Content: "Message " + string(rune(i))}
		hm.Add(msg)
	}

	recent := hm.GetRecentMessages(3)

	if len(recent) != 3 {
		t.Errorf("Expected 3 recent messages, got %d", len(recent))
	}
}

// TestHistoryManager_Copy tests copying history
func TestHistoryManager_Copy(t *testing.T) {
	hm := NewHistoryManager()
	hm.Add(common.Message{Role: "user", Content: "Test"})

	copy := hm.Copy()

	// Modify original
	hm.Add(common.Message{Role: "assistant", Content: "Response"})

	if copy.Length() != 1 {
		t.Errorf("Copy should have 1 message, got %d", copy.Length())
	}

	if hm.Length() != 2 {
		t.Errorf("Original should have 2 messages, got %d", hm.Length())
	}
}

// TestExecutionState_RecordRound tests recording round metrics
func TestExecutionState_RecordRound(t *testing.T) {
	es := NewExecutionState()

	es.RecordRound("agent-1", time.Millisecond*100, true)

	if es.RoundCount != 1 {
		t.Errorf("Expected 1 round, got %d", es.RoundCount)
	}

	metric := es.GetRoundMetric(1)
	if metric == nil {
		t.Fatal("Expected metric for round 1")
	}

	if metric.AgentID != "agent-1" {
		t.Errorf("Expected agent-1, got %s", metric.AgentID)
	}

	if !metric.Success {
		t.Error("Expected success=true")
	}
}

// TestExecutionState_GetMetrics tests getting aggregated metrics
func TestExecutionState_GetMetrics(t *testing.T) {
	es := NewExecutionState()

	es.RecordRound("agent-1", time.Millisecond*100, true)
	es.RecordRound("agent-2", time.Millisecond*200, true)

	metrics := es.GetMetrics()

	if metrics.RoundCount != 2 {
		t.Errorf("Expected 2 rounds, got %d", metrics.RoundCount)
	}

	if metrics.SuccessfulRounds != 2 {
		t.Errorf("Expected 2 successful rounds, got %d", metrics.SuccessfulRounds)
	}
}

// TestExecutionState_RecordHandoff tests handoff recording
func TestExecutionState_RecordHandoff(t *testing.T) {
	es := NewExecutionState()

	es.RecordHandoff()
	es.RecordHandoff()

	if es.HandoffCount != 2 {
		t.Errorf("Expected 2 handoffs, got %d", es.HandoffCount)
	}
}

// TestExecutionState_IsRunning tests running state
func TestExecutionState_IsRunning(t *testing.T) {
	es := NewExecutionState()

	if !es.IsRunning() {
		t.Error("Expected state to be running initially")
	}

	es.Finish()

	if es.IsRunning() {
		t.Error("Expected state to not be running after finish")
	}
}

// TestExecutionState_Reset tests resetting state
func TestExecutionState_Reset(t *testing.T) {
	es := NewExecutionState()

	es.RecordRound("agent-1", time.Millisecond*100, true)
	es.RecordHandoff()

	es.Reset()

	if es.RoundCount != 0 {
		t.Errorf("Expected 0 rounds after reset, got %d", es.RoundCount)
	}

	if es.HandoffCount != 0 {
		t.Errorf("Expected 0 handoffs after reset, got %d", es.HandoffCount)
	}
}

// TestExecutionState_Copy tests copying state
func TestExecutionState_Copy(t *testing.T) {
	es := NewExecutionState()

	es.RecordRound("agent-1", time.Millisecond*100, true)
	es.RecordHandoff()

	copy := es.Copy()

	if copy.RoundCount != es.RoundCount {
		t.Errorf("Copy should have same round count")
	}

	if copy.HandoffCount != es.HandoffCount {
		t.Errorf("Copy should have same handoff count")
	}
}

// TestExecutionFlow_NewExecutionFlow tests creating new flow
func TestExecutionFlow_NewExecutionFlow(t *testing.T) {
	agent := &common.Agent{ID: "test-agent"}

	flow := NewExecutionFlow(agent, 10, 5)

	if flow.CurrentAgent != agent {
		t.Error("Expected same agent")
	}

	if flow.MaxRounds != 10 {
		t.Errorf("Expected max rounds 10, got %d", flow.MaxRounds)
	}

	if flow.MaxHandoffs != 5 {
		t.Errorf("Expected max handoffs 5, got %d", flow.MaxHandoffs)
	}
}

// TestExecutionFlow_CanContinue tests execution limit checks
func TestExecutionFlow_CanContinue(t *testing.T) {
	agent := &common.Agent{ID: "test-agent"}
	flow := NewExecutionFlow(agent, 2, 1)

	// Should be able to continue initially
	if err := flow.CanContinue(); err != nil {
		t.Error("Should be able to continue initially")
	}

	// Exceed rounds
	flow.RoundCount = 2
	if err := flow.CanContinue(); err == nil {
		t.Error("Should not be able to continue after reaching max rounds")
	}

	// Exceed handoffs
	flow.RoundCount = 0
	flow.HandoffCount = 1
	if err := flow.CanContinue(); err == nil {
		t.Error("Should not be able to continue after reaching max handoffs")
	}
}

// TestExecutionFlow_GetWorkflowStatus tests status reporting
func TestExecutionFlow_GetWorkflowStatus(t *testing.T) {
	agent := &common.Agent{ID: "test-agent"}
	flow := NewExecutionFlow(agent, 10, 5)

	status := flow.GetWorkflowStatus()

	if status["current_agent"] != "test-agent" {
		t.Error("Expected correct agent in status")
	}

	if status["round_count"] != 0 {
		t.Error("Expected 0 rounds in status")
	}
}

// TestExecutionFlow_Reset tests resetting flow
func TestExecutionFlow_Reset(t *testing.T) {
	agent1 := &common.Agent{ID: "agent-1"}
	agent2 := &common.Agent{ID: "agent-2"}

	flow := NewExecutionFlow(agent1, 10, 5)
	flow.RoundCount = 5
	flow.HandoffCount = 2

	flow.Reset(agent2)

	if flow.CurrentAgent != agent2 {
		t.Error("Expected agent to be reset")
	}

	if flow.RoundCount != 0 {
		t.Error("Expected rounds to be reset")
	}

	if flow.HandoffCount != 0 {
		t.Error("Expected handoffs to be reset")
	}
}

// TestExecutionFlow_Copy tests copying flow
func TestExecutionFlow_Copy(t *testing.T) {
	agent := &common.Agent{ID: "test-agent"}
	flow := NewExecutionFlow(agent, 10, 5)
	flow.RoundCount = 2

	copy := flow.Copy()

	if copy.RoundCount != flow.RoundCount {
		t.Error("Copy should have same round count")
	}

	// Modify original
	flow.RoundCount = 5

	if copy.RoundCount != 2 {
		t.Error("Copy should be independent")
	}
}

// TestExecutionFlow_ValidateFlow tests flow validation
func TestExecutionFlow_ValidateFlow(t *testing.T) {
	agent := &common.Agent{ID: "test-agent"}
	flow := NewExecutionFlow(agent, 10, 5)

	if err := flow.ValidateFlow(); err != nil {
		t.Errorf("Valid flow should not error: %v", err)
	}

	// Test invalid flow
	invalidFlow := &ExecutionFlow{
		CurrentAgent: nil,
		MaxRounds:    10,
		MaxHandoffs:  5,
		State:        NewExecutionState(),
	}

	if err := invalidFlow.ValidateFlow(); err == nil {
		t.Error("Invalid flow should error")
	}
}

// TestHistoryManager_TrimConfig tests updating trim configuration
func TestHistoryManager_TrimConfig(t *testing.T) {
	hm := NewHistoryManager()

	hm.SetTrimConfig(5000, 0.5)

	if hm.trimThreshold != 5000 {
		t.Errorf("Expected trim threshold 5000, got %d", hm.trimThreshold)
	}

	if hm.trimPercentage != 0.5 {
		t.Errorf("Expected trim percentage 0.5, got %f", hm.trimPercentage)
	}
}

// TestExecutionState_TotalSize tests total character count
func TestExecutionState_GetLastAgentTime(t *testing.T) {
	es := NewExecutionState()

	es.RecordRound("agent-1", time.Millisecond*123, true)

	lastTime := es.GetLastAgentTime()
	if lastTime != time.Millisecond*123 {
		t.Errorf("Expected 123ms, got %v", lastTime)
	}
}
