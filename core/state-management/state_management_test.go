package statemanagement

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
		t.Fatalf("Expected 1 message before clear, got %d", hm.Length())
	}

	hm.Clear()
	if hm.Length() != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", hm.Length())
	}
}

// TestHistoryManager_GetRecentMessages tests getting recent messages
func TestHistoryManager_GetRecentMessages(t *testing.T) {
	hm := NewHistoryManager()

	for i := 0; i < 5; i++ {
		hm.Add(common.Message{Role: "user", Content: "Message " + string(rune(i))})
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
	copy.Clear()

	if hm.Length() != 1 {
		t.Errorf("Original history should have 1 message, got %d", hm.Length())
	}

	if copy.Length() != 0 {
		t.Errorf("Copied history should be empty, got %d", copy.Length())
	}
}

// TestExecutionState_RecordRound tests recording round metrics
func TestExecutionState_RecordRound(t *testing.T) {
	es := NewExecutionState()

	es.RecordRound("agent1", time.Duration(100*time.Millisecond), true)

	if es.RoundCount != 1 {
		t.Errorf("Expected 1 round, got %d", es.RoundCount)
	}

	metrics := es.GetMetrics()
	if metrics.RoundCount != 1 {
		t.Errorf("Expected 1 round in metrics, got %d", metrics.RoundCount)
	}

	if metrics.SuccessfulRounds != 1 {
		t.Errorf("Expected 1 successful round, got %d", metrics.SuccessfulRounds)
	}
}

// TestExecutionState_RecordHandoff tests recording handoffs
func TestExecutionState_RecordHandoff(t *testing.T) {
	es := NewExecutionState()

	es.RecordHandoff()
	es.RecordHandoff()

	if es.HandoffCount != 2 {
		t.Errorf("Expected 2 handoffs, got %d", es.HandoffCount)
	}
}

// TestExecutionState_Finish tests marking execution as complete
func TestExecutionState_Finish(t *testing.T) {
	es := NewExecutionState()

	if !es.IsRunning() {
		t.Errorf("Expected execution to be running initially")
	}

	time.Sleep(10 * time.Millisecond)
	es.Finish()

	if es.IsRunning() {
		t.Errorf("Expected execution to be finished")
	}

	if es.EndTime.IsZero() {
		t.Errorf("Expected EndTime to be set")
	}
}

// TestExecutionState_Reset tests resetting state
func TestExecutionState_Reset(t *testing.T) {
	es := NewExecutionState()

	es.RecordRound("agent1", time.Duration(100*time.Millisecond), true)
	es.RecordHandoff()
	es.Finish()

	if es.RoundCount != 1 {
		t.Fatalf("Expected 1 round before reset, got %d", es.RoundCount)
	}

	es.Reset()

	if es.RoundCount != 0 {
		t.Errorf("Expected 0 rounds after reset, got %d", es.RoundCount)
	}

	if es.HandoffCount != 0 {
		t.Errorf("Expected 0 handoffs after reset, got %d", es.HandoffCount)
	}

	if !es.IsRunning() {
		t.Errorf("Expected execution to be running after reset")
	}
}

// TestExecutionState_Copy tests copying execution state
func TestExecutionState_Copy(t *testing.T) {
	es := NewExecutionState()
	es.RecordRound("agent1", time.Duration(100*time.Millisecond), true)

	copy := es.Copy()

	copy.RecordRound("agent2", time.Duration(50*time.Millisecond), true)

	if es.RoundCount != 1 {
		t.Errorf("Original should have 1 round, got %d", es.RoundCount)
	}

	if copy.RoundCount != 2 {
		t.Errorf("Copy should have 2 rounds, got %d", copy.RoundCount)
	}
}

// TestExecutionState_GetRoundMetric tests retrieving round metric
func TestExecutionState_GetRoundMetric(t *testing.T) {
	es := NewExecutionState()
	es.RecordRound("agent1", time.Duration(100*time.Millisecond), true)

	metric := es.GetRoundMetric(1)
	if metric == nil {
		t.Fatalf("Expected round metric, got nil")
	}

	if metric.AgentID != "agent1" {
		t.Errorf("Expected agentID 'agent1', got %s", metric.AgentID)
	}

	if !metric.Success {
		t.Errorf("Expected successful round")
	}
}

// TestHistoryManager_TotalSize tests calculating total size
func TestHistoryManager_TotalSize(t *testing.T) {
	hm := NewHistoryManager()

	hm.Add(common.Message{Role: "user", Content: "Hello"})      // 5 chars
	hm.Add(common.Message{Role: "assistant", Content: "World"}) // 5 chars

	if hm.TotalSize() != 10 {
		t.Errorf("Expected total size 10, got %d", hm.TotalSize())
	}
}

// TestHistoryManager_Trimming tests history auto-trimming
func TestHistoryManager_Trimming(t *testing.T) {
	// Create with low threshold for testing
	hm := NewHistoryManagerWithConfig(10000, 100, 0.25)

	// Add messages that exceed threshold
	for i := 0; i < 50; i++ {
		hm.Add(common.Message{
			Role:    "user",
			Content: "Message with some content that takes up space",
		})
	}

	// History should be trimmed but not empty
	if hm.Length() == 0 {
		t.Errorf("Expected some messages after trimming, got 0")
	}

	if hm.Length() >= 50 {
		t.Errorf("Expected trimming to reduce message count, but got %d messages", hm.Length())
	}
}
