package crewai

import (
	"testing"
)

// TestNewCrewExecutorNilCrew verifies that NewCrewExecutor handles nil crew gracefully
func TestNewCrewExecutorNilCrew(t *testing.T) {
	t.Run("with nil crew", func(t *testing.T) {
		executor := NewCrewExecutor(nil, "test-key")

		// Should return nil gracefully, not panic
		if executor != nil {
			t.Fatal("Expected nil executor when crew is nil")
		}
	})

	t.Run("with valid crew but no agents", func(t *testing.T) {
		crew := &Crew{
			Agents: []*Agent{},
		}

		executor := NewCrewExecutor(crew, "test-key")

		// Should return executor but with nil entryAgent
		if executor == nil {
			t.Fatal("Expected non-nil executor")
		}
		if executor.entryAgent != nil {
			t.Fatal("Expected nil entryAgent when crew has no agents")
		}
	})

	t.Run("with valid crew and agents", func(t *testing.T) {
		agent1 := &Agent{
			ID:   "agent1",
			Name: "Agent 1",
		}
		crew := &Crew{
			Agents: []*Agent{agent1},
		}

		executor := NewCrewExecutor(crew, "test-key")

		// Should return executor with entryAgent set to first agent
		if executor == nil {
			t.Fatal("Expected non-nil executor")
		}
		if executor.entryAgent == nil {
			t.Fatal("Expected non-nil entryAgent when crew has agents")
		}
		if executor.entryAgent.ID != "agent1" {
			t.Fatalf("Expected entryAgent to be first agent, got %s", executor.entryAgent.ID)
		}
	})
}

// TestExecuteStreamNilEntryAgent verifies ExecuteStream handles nil entry agent
func TestExecuteStreamNilEntryAgent(t *testing.T) {
	t.Run("with nil entry agent", func(t *testing.T) {
		executor := &CrewExecutor{
			crew:        &Crew{Agents: []*Agent{}},
			apiKey:      "test-key",
			entryAgent:  nil, // ← nil entry agent
			history:     NewHistoryManager(),
			Verbose:     false,
			ResumeAgentID: "",
			ToolTimeouts: NewToolTimeoutConfig(),
			Metrics:     NewMetricsCollector(),
			defaults:    DefaultHardcodedDefaults(),
		}

		// Create a simple streaming test
		streamChan := make(chan *StreamEvent, 100)
		defer close(streamChan)

		err := executor.ExecuteStream(nil, "test input", streamChan)

		// Should return error, not panic
		if err == nil {
			t.Fatal("Expected error when entry agent is nil")
		}
		if err.Error() != "no entry agent found" {
			t.Fatalf("Expected 'no entry agent found' error, got %v", err)
		}
	})
}

// BenchmarkNewCrewExecutor benchmarks the nil check overhead
func BenchmarkNewCrewExecutor(b *testing.B) {
	crew := &Crew{
		Agents: []*Agent{
			{ID: "agent1", Name: "Agent 1"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCrewExecutor(crew, "test-key")
	}
}
