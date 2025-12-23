package crewai

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestHistoryThreadSafety verifies that concurrent access to history is safe
// This test will catch race conditions if run with -race flag
func TestHistoryThreadSafety(t *testing.T) {
	executor := NewCrewExecutor(&Crew{
		Agents: []*Agent{},
	}, "test-key")

	if executor == nil {
		t.Fatal("Failed to create CrewExecutor")
	}

	// Test 1: Multiple goroutines writing to history simultaneously
	t.Run("ConcurrentWrites", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 10
		messagesPerGoroutine := 100

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < messagesPerGoroutine; j++ {
					executor.appendMessage(Message{
						Role:    "user",
						Content: fmt.Sprintf("Message from goroutine %d, iteration %d", id, j),
					})
				}
			}(i)
		}

		wg.Wait()

		// Verify all messages were added
		expectedCount := numGoroutines * messagesPerGoroutine
		history := executor.GetHistory()
		if len(history) != expectedCount {
			t.Errorf("Expected %d messages, got %d", expectedCount, len(history))
		}
	})

	// Test 2: Concurrent reads and writes
	t.Run("ConcurrentReadsAndWrites", func(t *testing.T) {
		executor.ClearHistory()

		var wg sync.WaitGroup
		numReaders := 5
		numWriters := 5
		iterations := 50

		// Writers
		wg.Add(numWriters)
		for i := 0; i < numWriters; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					executor.appendMessage(Message{
						Role:    "assistant",
						Content: fmt.Sprintf("Writer %d, iteration %d", id, j),
					})
					time.Sleep(time.Millisecond) // Small delay to interleave with readers
				}
			}(i)
		}

		// Readers
		wg.Add(numReaders)
		for i := 0; i < numReaders; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					_ = executor.GetHistory()
					_ = executor.estimateHistoryTokens()
					time.Sleep(time.Millisecond)
				}
			}(i)
		}

		wg.Wait()
		// If we get here without panic, the test passes
	})

	// Test 3: ClearHistory while reading
	t.Run("ClearWhileReading", func(t *testing.T) {
		executor.ClearHistory()
		executor.appendMessage(Message{Role: "user", Content: "Test message"})

		var wg sync.WaitGroup

		// Writer: clearing history repeatedly
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				executor.ClearHistory()
				executor.appendMessage(Message{
					Role:    "user",
					Content: fmt.Sprintf("Message %d", i),
				})
				time.Sleep(time.Millisecond)
			}
		}()

		// Readers: reading history
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				_ = executor.GetHistory()
				time.Sleep(time.Millisecond / 10)
			}
		}()

		wg.Wait()
	})

	// Test 4: trimHistoryIfNeeded while appending
	t.Run("TrimWhileAppending", func(t *testing.T) {
		executor.ClearHistory()
		executor.defaults = &HardcodedDefaults{
			MaxContextWindow:  500,
			ContextTrimPercent: 20,
		}

		var wg sync.WaitGroup

		// Writer: appending messages
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				executor.appendMessage(Message{
					Role:    "user",
					Content: fmt.Sprintf("This is a long message that takes up some tokens. Message number %d", i),
				})
			}
		}()

		// Trimmer: trimming history
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				executor.trimHistoryIfNeeded()
				time.Sleep(time.Millisecond)
			}
		}()

		wg.Wait()
	})
}

// TestHistoryDataIntegrity verifies that messages are not lost or corrupted
func TestHistoryDataIntegrity(t *testing.T) {
	executor := NewCrewExecutor(&Crew{
		Agents: []*Agent{},
	}, "test-key")

	executor.ClearHistory()

	// Add messages in multiple goroutines
	var wg sync.WaitGroup
	numGoroutines := 5
	messagesPerGoroutine := 20
	messages := make(map[string]int)
	var mu sync.Mutex

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := fmt.Sprintf("G%d-M%d", id, j)
				executor.appendMessage(Message{
					Role:    "user",
					Content: msg,
				})
				mu.Lock()
				messages[msg]++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify all messages are present
	history := executor.GetHistory()
	if len(history) != numGoroutines*messagesPerGoroutine {
		t.Errorf("Expected %d messages, got %d", numGoroutines*messagesPerGoroutine, len(history))
	}

	// Verify message integrity
	for _, msg := range history {
		if messages[msg.Content] == 0 {
			t.Errorf("Unexpected message found: %s", msg.Content)
		}
	}
}

// BenchmarkConcurrentHistory benchmarks history operations under concurrency
func BenchmarkConcurrentHistory(b *testing.B) {
	executor := NewCrewExecutor(&Crew{
		Agents: []*Agent{},
	}, "test-key")

	executor.ClearHistory()

	b.Run("AppendMessage", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				executor.appendMessage(Message{
					Role:    "user",
					Content: fmt.Sprintf("Message %d", i),
				})
				i++
			}
		})
	})

	b.Run("GetHistory", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = executor.GetHistory()
			}
		})
	})

	b.Run("EstimateTokens", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = executor.estimateHistoryTokens()
			}
		})
	})
}

// TestRaceDetector is a simple test to ensure -race flag catches issues
// This test should pass, but would fail without proper mutex protection
func TestRaceDetector(t *testing.T) {
	executor := NewCrewExecutor(&Crew{
		Agents: []*Agent{},
	}, "test-key")

	executor.ClearHistory()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simulate concurrent operations that would race without mutex
	var wg sync.WaitGroup

	// Simulate ExecuteStream operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			executor.appendMessage(Message{Role: "user", Content: "input"})
			executor.appendMessage(Message{Role: "assistant", Content: "response"})
			executor.appendMessage(Message{Role: "user", Content: "tool results"})
		}
	}()

	// Simulate GetHistory calls (like from external monitoring)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = executor.GetHistory()
		}
	}()

	// Simulate trimming
	wg.Add(1)
	go func() {
		defer wg.Done()
		executor.defaults = &HardcodedDefaults{
			MaxContextWindow:  5000,
			ContextTrimPercent: 20,
		}
		for i := 0; i < 30; i++ {
			executor.trimHistoryIfNeeded()
		}
	}()

	// Simulate clearing
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			executor.ClearHistory()
			time.Sleep(time.Millisecond)
		}
	}()

	// Wait for all operations to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-ctx.Done():
		t.Fatal("Test timed out - possible deadlock")
	}
}
