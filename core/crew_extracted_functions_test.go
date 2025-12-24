package crewai

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// ===== PHASE 2: TESTS FOR EXTRACTED HELPER FUNCTIONS =====

// TestSendStreamEvent tests the sendStreamEvent helper function
func TestSendStreamEvent(t *testing.T) {
	executor := NewCrewExecutor(&Crew{Agents: []*Agent{}}, "test-key")
	if executor == nil {
		t.Fatal("Failed to create CrewExecutor")
	}

	t.Run("sends event to non-nil channel", func(t *testing.T) {
		streamChan := make(chan *StreamEvent, 10)
		executor.sendStreamEvent(streamChan, "test_event", "TestAgent", "test message")

		// Check if event was sent
		if len(streamChan) != 1 {
			t.Fatalf("Expected 1 event, got %d", len(streamChan))
		}

		event := <-streamChan
		if event.Type != "test_event" {
			t.Fatalf("Expected event type 'test_event', got '%s'", event.Type)
		}
		if event.Agent != "TestAgent" {
			t.Fatalf("Expected agent 'TestAgent', got '%s'", event.Agent)
		}
		if event.Content != "test message" {
			t.Fatalf("Expected content 'test message', got '%s'", event.Content)
		}
	})

	t.Run("handles nil channel gracefully", func(t *testing.T) {
		// Should not panic with nil channel
		executor.sendStreamEvent(nil, "test_event", "TestAgent", "test message")
		// If we get here, no panic occurred - test passes
	})

	t.Run("handles full channel with timeout", func(t *testing.T) {
		// Create a channel with capacity 1 and fill it
		streamChan := make(chan *StreamEvent, 1)
		streamChan <- NewStreamEvent("existing", "Agent", "message")

		// This should timeout and log warning, but not panic
		executor.sendStreamEvent(streamChan, "test_event", "TestAgent", "test message")
		// If we get here, no panic occurred - test passes
	})
}

// TestHandleAgentError tests the handleAgentError helper function
func TestHandleAgentError(t *testing.T) {
	executor := NewCrewExecutor(&Crew{Agents: []*Agent{}}, "test-key")
	if executor == nil {
		t.Fatal("Failed to create CrewExecutor")
	}

	t.Run("handles nil error gracefully", func(t *testing.T) {
		streamChan := make(chan *StreamEvent, 10)
		err := executor.handleAgentError(context.Background(), &Agent{ID: "test"}, nil, streamChan)

		if err != nil {
			t.Fatal("Expected nil error for nil input")
		}
		if len(streamChan) != 0 {
			t.Fatal("Expected no stream events for nil error")
		}
	})

	t.Run("logs error and sends event", func(t *testing.T) {
		agent := &Agent{
			ID:   "agent1",
			Name: "TestAgent",
		}
		streamChan := make(chan *StreamEvent, 10)
		testErr := fmt.Errorf("test error")

		returnedErr := executor.handleAgentError(context.Background(), agent, testErr, streamChan)

		// Check return value
		if returnedErr != testErr {
			t.Fatal("Expected same error returned")
		}

		// Check stream event was sent
		if len(streamChan) != 1 {
			t.Fatalf("Expected 1 stream event, got %d", len(streamChan))
		}

		event := <-streamChan
		if event.Type != EventTypeError {
			t.Fatalf("Expected error event type, got '%s'", event.Type)
		}
	})

	t.Run("updates metrics when metadata exists", func(t *testing.T) {
		agent := &Agent{
			ID:   "agent2",
			Name: "TestAgent2",
			Metadata: &AgentMetadata{
				AgentID:   "agent2",
				AgentName: "TestAgent2",
			},
		}
		streamChan := make(chan *StreamEvent, 10)
		testErr := fmt.Errorf("test error 2")

		executor.handleAgentError(context.Background(), agent, testErr, streamChan)

		// Metadata should be updated (Performance metrics updated)
		// If we get here without panic, test passes
	})
}

// TestUpdateAgentMetrics tests the updateAgentMetrics helper function
func TestUpdateAgentMetrics(t *testing.T) {
	executor := NewCrewExecutor(&Crew{Agents: []*Agent{}}, "test-key")
	if executor == nil {
		t.Fatal("Failed to create CrewExecutor")
	}

	t.Run("handles nil agent gracefully", func(t *testing.T) {
		err := executor.updateAgentMetrics(nil, true, 100*time.Millisecond, 10, "")
		if err != nil {
			t.Fatal("Expected nil error for nil agent")
		}
	})

	t.Run("handles nil metadata gracefully", func(t *testing.T) {
		agent := &Agent{ID: "test", Name: "TestAgent", Metadata: nil}
		err := executor.updateAgentMetrics(agent, true, 100*time.Millisecond, 10, "")
		if err != nil {
			t.Fatal("Expected nil error for nil metadata")
		}
	})

	t.Run("updates metrics correctly", func(t *testing.T) {
		agent := &Agent{
			ID:   "agent3",
			Name: "TestAgent3",
			Metadata: &AgentMetadata{
				AgentID:   "agent3",
				AgentName: "TestAgent3",
			},
		}

		duration := 100 * time.Millisecond
		memory := 25
		executor.updateAgentMetrics(agent, true, duration, memory, "")

		// Verify metrics were updated (specific check depends on implementation)
		// If we get here without panic, test passes
	})

	t.Run("handles error message correctly", func(t *testing.T) {
		agent := &Agent{
			ID:   "agent4",
			Name: "TestAgent4",
			Metadata: &AgentMetadata{
				AgentID:   "agent4",
				AgentName: "TestAgent4",
			},
		}

		executor.updateAgentMetrics(agent, false, 50*time.Millisecond, 5, "test error")
		// If we get here without panic, test passes
	})
}

// TestCalculateMessageTokens tests the calculateMessageTokens helper function
func TestCalculateMessageTokens(t *testing.T) {
	t.Run("calculates tokens for empty message", func(t *testing.T) {
		msg := Message{Role: RoleUser, Content: ""}
		tokens := calculateMessageTokens(msg)

		// Should be just base value for empty content
		if tokens != TokenBaseValue {
			t.Fatalf("Expected %d tokens for empty message, got %d", TokenBaseValue, tokens)
		}
	})

	t.Run("calculates tokens for short message", func(t *testing.T) {
		msg := Message{Role: RoleUser, Content: "hello"}
		tokens := calculateMessageTokens(msg)

		// Formula: base + (len + padding) / divisor
		// = 4 + (5 + 3) / 4 = 4 + 2 = 6
		expected := TokenBaseValue + (5+TokenPaddingValue)/TokenDivisor
		if tokens != expected {
			t.Fatalf("Expected %d tokens, got %d", expected, tokens)
		}
	})

	t.Run("calculates tokens for longer message", func(t *testing.T) {
		msg := Message{Role: RoleAssistant, Content: "This is a longer message with multiple words"}
		tokens := calculateMessageTokens(msg)

		content := msg.Content
		expected := TokenBaseValue + (len(content)+TokenPaddingValue)/TokenDivisor
		if tokens != expected {
			t.Fatalf("Expected %d tokens, got %d", expected, tokens)
		}
	})

	t.Run("produces consistent results", func(t *testing.T) {
		msg := Message{Role: RoleUser, Content: "test content"}

		tokens1 := calculateMessageTokens(msg)
		tokens2 := calculateMessageTokens(msg)

		if tokens1 != tokens2 {
			t.Fatal("Expected consistent token calculation")
		}
	})
}

// TestAddUserMessageToHistory tests the addUserMessageToHistory helper
func TestAddUserMessageToHistory(t *testing.T) {
	executor := NewCrewExecutor(&Crew{Agents: []*Agent{}}, "test-key")
	if executor == nil {
		t.Fatal("Failed to create CrewExecutor")
	}

	executor.ClearHistory()

	t.Run("adds user message correctly", func(t *testing.T) {
		content := "test user message"
		executor.addUserMessageToHistory(content)

		history := executor.GetHistory()
		if len(history) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(history))
		}

		msg := history[0]
		if msg.Role != RoleUser {
			t.Fatalf("Expected role %s, got %s", RoleUser, msg.Role)
		}
		if msg.Content != content {
			t.Fatalf("Expected content '%s', got '%s'", content, msg.Content)
		}
	})

	t.Run("multiple calls append messages", func(t *testing.T) {
		executor.ClearHistory()

		executor.addUserMessageToHistory("message 1")
		executor.addUserMessageToHistory("message 2")
		executor.addUserMessageToHistory("message 3")

		history := executor.GetHistory()
		if len(history) != 3 {
			t.Fatalf("Expected 3 messages, got %d", len(history))
		}
	})
}

// TestAddAssistantMessageToHistory tests the addAssistantMessageToHistory helper
func TestAddAssistantMessageToHistory(t *testing.T) {
	executor := NewCrewExecutor(&Crew{Agents: []*Agent{}}, "test-key")
	if executor == nil {
		t.Fatal("Failed to create CrewExecutor")
	}

	executor.ClearHistory()

	t.Run("adds assistant message correctly", func(t *testing.T) {
		content := "test assistant response"
		executor.addAssistantMessageToHistory(content)

		history := executor.GetHistory()
		if len(history) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(history))
		}

		msg := history[0]
		if msg.Role != RoleAssistant {
			t.Fatalf("Expected role %s, got %s", RoleAssistant, msg.Role)
		}
		if msg.Content != content {
			t.Fatalf("Expected content '%s', got '%s'", content, msg.Content)
		}
	})

	t.Run("works with user messages for conversation flow", func(t *testing.T) {
		executor.ClearHistory()

		executor.addUserMessageToHistory("user question")
		executor.addAssistantMessageToHistory("assistant answer")
		executor.addUserMessageToHistory("follow-up question")
		executor.addAssistantMessageToHistory("follow-up answer")

		history := executor.GetHistory()
		if len(history) != 4 {
			t.Fatalf("Expected 4 messages, got %d", len(history))
		}

		// Verify alternating pattern
		if history[0].Role != RoleUser {
			t.Fatal("Expected first message to be from user")
		}
		if history[1].Role != RoleAssistant {
			t.Fatal("Expected second message to be from assistant")
		}
	})
}

// TestRecordAgentExecution tests the recordAgentExecution helper
func TestRecordAgentExecution(t *testing.T) {
	executor := NewCrewExecutor(&Crew{Agents: []*Agent{}}, "test-key")
	if executor == nil {
		t.Fatal("Failed to create CrewExecutor")
	}

	t.Run("handles nil agent gracefully", func(t *testing.T) {
		// Should not panic
		executor.recordAgentExecution(nil, 100*time.Millisecond, true)
	})

	t.Run("handles nil metrics gracefully", func(t *testing.T) {
		agent := &Agent{ID: "test", Name: "TestAgent"}
		executor.Metrics = nil

		// Should not panic
		executor.recordAgentExecution(agent, 100*time.Millisecond, true)
	})

	t.Run("records execution with metrics", func(t *testing.T) {
		// Create executor with metrics
		executor.Metrics = NewMetricsCollector()
		agent := &Agent{ID: "test", Name: "TestAgent"}

		executor.recordAgentExecution(agent, 100*time.Millisecond, true)

		// If we get here without panic, test passes
		// (actual metrics validation would depend on metrics implementation)
	})

	t.Run("records both success and failure", func(t *testing.T) {
		executor.Metrics = NewMetricsCollector()
		agent := &Agent{ID: "test2", Name: "TestAgent2"}

		executor.recordAgentExecution(agent, 100*time.Millisecond, true)
		executor.recordAgentExecution(agent, 150*time.Millisecond, false)

		// If we get here without panic, test passes
	})
}

// BenchmarkExtractedFunctions benchmarks the extracted helper functions
func BenchmarkExtractedFunctions(b *testing.B) {
	executor := NewCrewExecutor(&Crew{Agents: []*Agent{}}, "test-key")
	executor.Metrics = NewMetricsCollector()
	agent := &Agent{
		ID:   "benchmark",
		Name: "BenchmarkAgent",
		Metadata: &AgentMetadata{
			AgentID:   "benchmark",
			AgentName: "BenchmarkAgent",
		},
	}

	b.Run("SendStreamEvent", func(b *testing.B) {
		streamChan := make(chan *StreamEvent, 100)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			executor.sendStreamEvent(streamChan, "test", "Agent", "message")
		}
	})

	b.Run("CalculateMessageTokens", func(b *testing.B) {
		msg := Message{
			Role:    RoleUser,
			Content: "This is a benchmark message for testing token calculation",
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = calculateMessageTokens(msg)
		}
	})

	b.Run("AddUserMessageToHistory", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			executor.addUserMessageToHistory(fmt.Sprintf("message %d", i))
		}
	})

	b.Run("RecordAgentExecution", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			executor.recordAgentExecution(agent, 100*time.Millisecond, true)
		}
	})
}
