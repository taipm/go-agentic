package crewai

import (
	"fmt"
	"testing"
)

// ===== Issue #4: History Mutation Bug Tests =====

// TestCopyHistoryEdgeCases verifies copyHistory handles all edge cases correctly
func TestCopyHistoryEdgeCases(t *testing.T) {
	// Test 1: Empty slice
	empty := copyHistory([]Message{})
	if len(empty) != 0 {
		t.Error("Empty history not handled correctly")
	}

	// Test 2: Nil slice (should return empty, not nil)
	nilHistory := copyHistory(nil)
	if nilHistory == nil {
		t.Error("Nil history should return empty slice, not nil")
	}
	if len(nilHistory) != 0 {
		t.Errorf("Nil history should return 0-length slice, got %d", len(nilHistory))
	}

	// Test 3: Single message
	single := copyHistory([]Message{{Role: "user", Content: "test"}})
	if len(single) != 1 {
		t.Error("Single message not copied correctly")
	}
	if single[0].Content != "test" {
		t.Error("Message content corrupted during copy")
	}

	// Test 4: Multiple messages
	original := []Message{
		{Role: "user", Content: "msg1"},
		{Role: "assistant", Content: "msg2"},
		{Role: "user", Content: "msg3"},
	}
	copied := copyHistory(original)
	if len(copied) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(copied))
	}

	// Test 5: Modification of copy doesn't affect original
	copied[0].Content = "modified"
	if original[0].Content != "msg1" {
		t.Error("Modifying copy affected original - not a true copy!")
	}

	// Test 6: Different slice instances
	if &copied[0] == &original[0] {
		t.Error("Copy shares memory with original - not a deep copy!")
	}
}

// TestExecuteStreamHistoryImmutability verifies concurrent requests don't corrupt history
func TestExecuteStreamHistoryImmutability(t *testing.T) {
	originalHistory := []Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi there"},
	}

	// Simulate what StreamHandler does (http.go line 107)
	history1 := copyHistory(originalHistory)
	history2 := copyHistory(originalHistory)

	// Modify first copy (simulate Request A's execution)
	history1 = append(history1, Message{
		Role:    "user",
		Content: "new message from request A",
	})

	// Modify second copy (simulate Request B's execution)
	history2 = append(history2, Message{
		Role:    "user",
		Content: "new message from request B",
	})

	// Verify copies are independent
	if len(history1) != 3 {
		t.Errorf("history1 should have 3 messages, got %d", len(history1))
	}

	if len(history2) != 3 {
		t.Errorf("history2 should have 3 messages, got %d", len(history2))
	}

	// Verify original is unchanged
	if len(originalHistory) != 2 {
		t.Errorf("Original should still have 2 messages, got %d", len(originalHistory))
	}

	// Verify they don't have each other's new messages
	if history1[2].Content != "new message from request A" {
		t.Error("history1 lost its own message!")
	}

	if history2[2].Content != "new message from request B" {
		t.Error("history2 lost its own message!")
	}

	// Most important: they should be different
	if history1[2].Content == history2[2].Content {
		t.Error("Copies share the same appended message - not isolated!")
	}
}

// TestExecuteStreamConcurrentRequests verifies no race on history under concurrent load
func TestExecuteStreamConcurrentRequests(t *testing.T) {
	originalHistory := []Message{
		{Role: "user", Content: "initial query"},
		{Role: "assistant", Content: "initial response"},
	}

	successCount := 0
	failureCount := 0
	resultsChan := make(chan bool, 10)

	// Simulate 10 concurrent requests (like StreamHandler being called 10 times)
	for i := 0; i < 10; i++ {
		go func(index int) {
			// Each "request" gets its own copy (like StreamHandler line 107)
			localHistory := copyHistory(originalHistory)

			// Simulate request-specific mutations
			localHistory = append(localHistory, Message{
				Role:    "user",
				Content: fmt.Sprintf("request %d query", index),
			})

			// Simulate agent response
			localHistory = append(localHistory, Message{
				Role:    "assistant",
				Content: fmt.Sprintf("request %d response", index),
			})

			// Verify local history integrity
			if len(localHistory) != 4 {
				resultsChan <- false
				return
			}

			// Verify original is still intact (wasn't modified by concurrent goroutine)
			if len(originalHistory) != 2 {
				resultsChan <- false
				return
			}

			// Success: concurrent request didn't corrupt state
			resultsChan <- true
		}(i)
	}

	// Collect results
	for i := 0; i < 10; i++ {
		if <-resultsChan {
			successCount++
		} else {
			failureCount++
		}
	}

	// All requests should succeed
	if failureCount > 0 {
		t.Errorf("Concurrent requests had failures: %d failures, %d successes", failureCount, successCount)
	}

	if successCount != 10 {
		t.Errorf("Expected 10 successful requests, got %d", successCount)
	}

	// Original should be completely untouched by any concurrent request
	if len(originalHistory) != 2 {
		t.Errorf("Original history was corrupted: expected 2, got %d", len(originalHistory))
	}
}
