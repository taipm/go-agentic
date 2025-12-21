package crewai

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
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

// ===== Issue #5: Panic Recovery Tests =====

// TestSafeExecuteToolNormalExecution verifies safeExecuteTool works normally without panics
func TestSafeExecuteToolNormalExecution(t *testing.T) {
	// Create a tool that returns normally (no panic)
	tool := &Tool{
		Name: "test_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "success result", nil
		},
	}

	// Execute through safeExecuteTool
	output, err := safeExecuteTool(nil, tool, map[string]interface{}{})

	// Should succeed without error
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if output != "success result" {
		t.Errorf("Expected 'success result', got: %s", output)
	}
}

// TestSafeExecuteToolErrorHandling verifies safeExecuteTool passes through normal errors
func TestSafeExecuteToolErrorHandling(t *testing.T) {
	// Create a tool that returns an error (not a panic)
	tool := &Tool{
		Name: "error_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "", fmt.Errorf("tool error: something went wrong")
		},
	}

	// Execute through safeExecuteTool
	output, err := safeExecuteTool(nil, tool, map[string]interface{}{})

	// Should return the error
	if err == nil {
		t.Error("Expected error from tool, but got nil")
	}

	if err.Error() != "tool error: something went wrong" {
		t.Errorf("Expected original error message, got: %v", err)
	}

	if output != "" {
		t.Errorf("Expected empty output on error, got: %s", output)
	}
}

// TestSafeExecuteToolPanicRecovery verifies safeExecuteTool catches panics
func TestSafeExecuteToolPanicRecovery(t *testing.T) {
	// Create a tool that panics
	tool := &Tool{
		Name: "panicking_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			// This will panic - simulating a buggy tool
			panic("nil pointer dereference in tool")
		},
	}

	// Execute through safeExecuteTool - should NOT panic, should return error
	output, err := safeExecuteTool(nil, tool, map[string]interface{}{})

	// Should recover from panic and return error
	if err == nil {
		t.Error("Expected panic to be caught and converted to error")
	}

	// Error message should contain the panic information
	if !strings.Contains(err.Error(), "panicked") {
		t.Errorf("Expected error to mention panic, got: %v", err)
	}

	if !strings.Contains(err.Error(), "nil pointer dereference in tool") {
		t.Errorf("Expected error to contain panic message, got: %v", err)
	}

	// Output should be empty on panic
	if output != "" {
		t.Errorf("Expected empty output on panic, got: %s", output)
	}
}

// TestSafeExecuteToolPanicWithRuntimeError verifies recovery from runtime panics
func TestSafeExecuteToolPanicWithRuntimeError(t *testing.T) {
	// Create a tool that causes a runtime panic (array index out of bounds)
	tool := &Tool{
		Name: "runtime_panic_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			// This will panic at runtime
			arr := []int{1, 2, 3}
			_ = arr[10]  // Index out of bounds → runtime panic
			return "should not reach here", nil
		},
	}

	// Execute through safeExecuteTool - should catch runtime panic
	_, err := safeExecuteTool(nil, tool, map[string]interface{}{})

	// Should recover from runtime panic
	if err == nil {
		t.Error("Expected runtime panic to be caught")
	}

	if !strings.Contains(err.Error(), "panicked") {
		t.Errorf("Expected error to mention panic, got: %v", err)
	}
}

// TestSafeExecuteToolMultipleCalls verifies repeated calls don't leak panic state
func TestSafeExecuteToolMultipleCalls(t *testing.T) {
	// Tool 1: Normal
	tool1 := &Tool{
		Name: "normal_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result1", nil
		},
	}

	// Tool 2: Panics
	tool2 := &Tool{
		Name: "panic_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			panic("tool panic")
		},
	}

	// Tool 3: Normal (after panic)
	tool3 := &Tool{
		Name: "normal_after_panic",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result3", nil
		},
	}

	// Call tool 1 (should succeed)
	output1, err1 := safeExecuteTool(nil, tool1, map[string]interface{}{})
	if err1 != nil || output1 != "result1" {
		t.Errorf("Tool 1 failed: err=%v, output=%s", err1, output1)
	}

	// Call tool 2 (should catch panic)
	output2, err2 := safeExecuteTool(nil, tool2, map[string]interface{}{})
	if err2 == nil {
		t.Error("Tool 2 panic not caught")
	}
	if output2 != "" {
		t.Errorf("Tool 2 should return empty output, got: %s", output2)
	}

	// Call tool 3 (should succeed - panic state didn't leak)
	output3, err3 := safeExecuteTool(nil, tool3, map[string]interface{}{})
	if err3 != nil || output3 != "result3" {
		t.Errorf("Tool 3 failed: err=%v, output=%s", err3, output3)
	}
}

// TestExecuteCallsWithPanicingTool verifies executeCalls handles panicking tools
func TestExecuteCallsWithPanicingTool(t *testing.T) {
	// Create agent with 3 tools: normal, panicking, normal
	agent := &Agent{
		ID:   "test_agent",
		Name: "Test Agent",
		Tools: []*Tool{
			{
				Name: "working_tool",
				Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
					return "working result", nil
				},
			},
			{
				Name: "buggy_tool",
				Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
					panic("buggy tool crashed")
				},
			},
			{
				Name: "another_tool",
				Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
					return "another result", nil
				},
			},
		},
	}

	// Create executor
	executor := &CrewExecutor{
		crew:       &Crew{Agents: []*Agent{agent}},
		entryAgent: agent,
	}

	// Call tools: working, buggy (panics), working
	toolCalls := []ToolCall{
		{ToolName: "working_tool", Arguments: map[string]interface{}{}},
		{ToolName: "buggy_tool", Arguments: map[string]interface{}{}},
		{ToolName: "another_tool", Arguments: map[string]interface{}{}},
	}

	results := executor.executeCalls(nil, toolCalls, agent)

	// Should have 3 results
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Result 1: Success
	if results[0].Status != "success" || results[0].Output != "working result" {
		t.Errorf("Tool 1 result incorrect: %v", results[0])
	}

	// Result 2: Error (panic caught and converted to error)
	if results[1].Status != "error" {
		t.Errorf("Tool 2 should be error status, got: %s", results[1].Status)
	}
	if !strings.Contains(results[1].Output, "panicked") {
		t.Errorf("Tool 2 error should mention panic, got: %s", results[1].Output)
	}

	// Result 3: Success (not affected by previous panic)
	if results[2].Status != "success" || results[2].Output != "another result" {
		t.Errorf("Tool 3 result incorrect: %v", results[2])
	}
}

// TestParallelExecutionWithPanicingTools verifies parallel tool execution handles panics
func TestParallelExecutionWithPanicingTools(t *testing.T) {
	// Simulate 5 tools executed in parallel, with 2 panicking
	agent := &Agent{
		ID:   "parallel_agent",
		Name: "Parallel Agent",
		Tools: []*Tool{
			{
				Name: "tool_1",
				Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
					return "result_1", nil
				},
			},
			{
				Name: "tool_2",
				Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
					panic("tool_2 panic")
				},
			},
			{
				Name: "tool_3",
				Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
					return "result_3", nil
				},
			},
			{
				Name: "tool_4",
				Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
					panic("tool_4 panic")
				},
			},
			{
				Name: "tool_5",
				Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
					return "result_5", nil
				},
			},
		},
	}

	executor := &CrewExecutor{
		crew:       &Crew{Agents: []*Agent{agent}},
		entryAgent: agent,
	}

	// Execute all 5 tools
	toolCalls := []ToolCall{
		{ToolName: "tool_1", Arguments: map[string]interface{}{}},
		{ToolName: "tool_2", Arguments: map[string]interface{}{}},
		{ToolName: "tool_3", Arguments: map[string]interface{}{}},
		{ToolName: "tool_4", Arguments: map[string]interface{}{}},
		{ToolName: "tool_5", Arguments: map[string]interface{}{}},
	}

	results := executor.executeCalls(nil, toolCalls, agent)

	// Should have 5 results (not crash despite panics)
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	// Count successful vs error results
	successCount := 0
	errorCount := 0

	for _, result := range results {
		if result.Status == "success" {
			successCount++
		} else if result.Status == "error" {
			errorCount++
		}
	}

	// 3 tools succeeded, 2 panicked (now errors)
	if successCount != 3 {
		t.Errorf("Expected 3 successful tools, got %d", successCount)
	}

	if errorCount != 2 {
		t.Errorf("Expected 2 error tools (panics), got %d", errorCount)
	}

	// Verify all results are accounted for
	if successCount+errorCount != 5 {
		t.Errorf("Not all results accounted for: %d + %d != 5", successCount, errorCount)
	}
}

// ===== Issue #11: Sequential Tool Timeout Tests =====

// TestToolTimeoutConfig verifies timeout configuration works correctly
func TestToolTimeoutConfig(t *testing.T) {
	cfg := NewToolTimeoutConfig()

	// Verify defaults
	if cfg.DefaultToolTimeout != 5*time.Second {
		t.Errorf("Expected 5s default timeout, got %v", cfg.DefaultToolTimeout)
	}
	if cfg.SequenceTimeout != 30*time.Second {
		t.Errorf("Expected 30s sequence timeout, got %v", cfg.SequenceTimeout)
	}
	if !cfg.CollectMetrics {
		t.Error("CollectMetrics should be true by default")
	}

	// Test per-tool override
	cfg.PerToolTimeout["slow_tool"] = 10 * time.Second
	timeout := cfg.GetToolTimeout("slow_tool")
	if timeout != 10*time.Second {
		t.Errorf("Expected 10s for slow_tool, got %v", timeout)
	}

	// Test default for unknown tool
	timeout = cfg.GetToolTimeout("unknown_tool")
	if timeout != 5*time.Second {
		t.Errorf("Expected 5s default for unknown tool, got %v", timeout)
	}
}

// TestExecuteCallsWithTimeout verifies per-tool timeouts work
func TestExecuteCallsWithTimeout(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{
				ID:   "test_agent",
				Name: "Test Agent",
				Tools: []*Tool{
					{
						Name: "fast_tool",
						Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
							return "fast result", nil
						},
					},
					{
						Name: "slow_tool",
						Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
							// Respect context cancellation
							select {
							case <-ctx.Done():
								return "", ctx.Err()
							case <-time.After(2 * time.Second):
								return "slow result", nil
							}
						},
					},
				},
			},
		},
	}

	executor := NewCrewExecutor(crew, "test-key")
	executor.ToolTimeouts = NewToolTimeoutConfig()
	executor.ToolTimeouts.DefaultToolTimeout = 100 * time.Millisecond // Very short timeout

	agent := crew.Agents[0]
	calls := []ToolCall{
		{ToolName: "fast_tool", Arguments: map[string]interface{}{}},
		{ToolName: "slow_tool", Arguments: map[string]interface{}{}},
	}

	results := executor.executeCalls(context.Background(), calls, agent)

	// Fast tool should succeed
	if results[0].Status != "success" {
		t.Errorf("Fast tool should succeed, got status: %s", results[0].Status)
	}

	// Slow tool should timeout
	if results[1].Status != "error" {
		t.Errorf("Slow tool should error on timeout, got status: %s", results[1].Status)
	}
	if !strings.Contains(results[1].Output, "context deadline exceeded") {
		t.Errorf("Slow tool error should mention context deadline, got: %s", results[1].Output)
	}
}

// TestExecutionMetricsCollection verifies metrics are collected correctly
func TestExecutionMetricsCollection(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{
				ID:   "test_agent",
				Name: "Test Agent",
				Tools: []*Tool{
					{
						Name: "tool_1",
						Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
							time.Sleep(10 * time.Millisecond)
							return "result 1", nil
						},
					},
					{
						Name: "tool_2",
						Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
							time.Sleep(20 * time.Millisecond)
							return "result 2", nil
						},
					},
				},
			},
		},
	}

	executor := NewCrewExecutor(crew, "test-key")
	executor.ToolTimeouts = NewToolTimeoutConfig()
	executor.ToolTimeouts.CollectMetrics = true

	agent := crew.Agents[0]
	calls := []ToolCall{
		{ToolName: "tool_1", Arguments: map[string]interface{}{}},
		{ToolName: "tool_2", Arguments: map[string]interface{}{}},
	}

	results := executor.executeCalls(context.Background(), calls, agent)

	// Should have 2 results
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Should have 2 metrics
	if len(executor.ToolTimeouts.ExecutionMetrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(executor.ToolTimeouts.ExecutionMetrics))
	}

	// Verify metrics
	if executor.ToolTimeouts.ExecutionMetrics[0].ToolName != "tool_1" {
		t.Errorf("Expected metric for tool_1, got %s", executor.ToolTimeouts.ExecutionMetrics[0].ToolName)
	}
	if executor.ToolTimeouts.ExecutionMetrics[0].Status != "success" {
		t.Errorf("Expected success status, got %s", executor.ToolTimeouts.ExecutionMetrics[0].Status)
	}
	if executor.ToolTimeouts.ExecutionMetrics[0].TimedOut {
		t.Error("Tool 1 should not have timed out")
	}

	// Verify duration is roughly correct (should be at least 10ms)
	if executor.ToolTimeouts.ExecutionMetrics[0].Duration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", executor.ToolTimeouts.ExecutionMetrics[0].Duration)
	}
}

// TestSequenceTimeoutStopsRemaining verifies sequence timeout stops remaining tools
func TestSequenceTimeoutStopsRemaining(t *testing.T) {
	crew := &Crew{
		Agents: []*Agent{
			{
				ID:   "test_agent",
				Name: "Test Agent",
				Tools: []*Tool{
					{
						Name: "tool_1",
						Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
							select {
							case <-ctx.Done():
								return "", ctx.Err()
							case <-time.After(30 * time.Millisecond):
								return "result 1", nil
							}
						},
					},
					{
						Name: "tool_2",
						Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
							select {
							case <-ctx.Done():
								return "", ctx.Err()
							case <-time.After(30 * time.Millisecond):
								return "result 2", nil
							}
						},
					},
					{
						Name: "tool_3",
						Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
							select {
							case <-ctx.Done():
								return "", ctx.Err()
							case <-time.After(30 * time.Millisecond):
								return "result 3", nil
							}
						},
					},
				},
			},
		},
	}

	executor := NewCrewExecutor(crew, "test-key")
	executor.ToolTimeouts = NewToolTimeoutConfig()
	executor.ToolTimeouts.SequenceTimeout = 60 * time.Millisecond // Total timeout for sequence (3 tools * 30ms each would exceed this)

	agent := crew.Agents[0]
	calls := []ToolCall{
		{ToolName: "tool_1", Arguments: map[string]interface{}{}},
		{ToolName: "tool_2", Arguments: map[string]interface{}{}},
		{ToolName: "tool_3", Arguments: map[string]interface{}{}},
	}

	results := executor.executeCalls(context.Background(), calls, agent)

	// First 1-2 tools might succeed, but at least 1 should fail due to sequence timeout
	successCount := 0
	errorCount := 0
	for _, result := range results {
		if result.Status == "success" {
			successCount++
		} else if result.Status == "error" {
			errorCount++
		}
	}

	// At least one should have failed
	if errorCount == 0 {
		t.Error("Expected at least one tool to timeout due to sequence limit")
	}

	// Total results count should be 3 (even if some failed)
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}
