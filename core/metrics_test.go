package crewai

import (
	"testing"
	"time"
)

// ============================================================================
// TOOL EXECUTION METRICS TESTS
// ============================================================================

// TestRecordToolExecution_Basic tests basic tool execution recording
func TestRecordToolExecution_Basic(t *testing.T) {
	mc := NewMetricsCollector()
	duration := 100 * time.Millisecond
	inputTokens := 1000
	outputTokens := 500
	cost := 0.05

	mc.RecordToolExecution("agent1", "web-search", duration, true, inputTokens, outputTokens, cost)

	metrics := mc.GetSystemMetrics()
	if metrics.TotalTokens != (inputTokens + outputTokens) {
		t.Errorf("Expected total tokens %d, got %d", inputTokens+outputTokens, metrics.TotalTokens)
	}
	if metrics.TotalCost != cost {
		t.Errorf("Expected total cost %.6f, got %.6f", cost, metrics.TotalCost)
	}
}

// TestRecordToolExecution_MultipleTools tests recording multiple tools
func TestRecordToolExecution_MultipleTools(t *testing.T) {
	mc := NewMetricsCollector()

	// First tool
	mc.RecordToolExecution("agent1", "web-search", 100*time.Millisecond, true, 1000, 500, 0.05)

	// Second tool
	mc.RecordToolExecution("agent1", "calculator", 50*time.Millisecond, true, 500, 200, 0.02)

	// Third tool (same agent, different tool)
	mc.RecordToolExecution("agent2", "web-search", 150*time.Millisecond, true, 2000, 1000, 0.10)

	metrics := mc.GetSystemMetrics()

	// Check agent1 has 2 tools
	agent1 := metrics.AgentMetrics["agent1"]
	if agent1 == nil {
		t.Fatalf("Expected agent1 metrics, got nil")
	}
	if len(agent1.ToolMetrics) != 2 {
		t.Errorf("Expected 2 tools for agent1, got %d", len(agent1.ToolMetrics))
	}

	// Check agent2 has 1 tool
	agent2 := metrics.AgentMetrics["agent2"]
	if agent2 == nil {
		t.Fatalf("Expected agent2 metrics, got nil")
	}
	if len(agent2.ToolMetrics) != 1 {
		t.Errorf("Expected 1 tool for agent2, got %d", len(agent2.ToolMetrics))
	}

	// Check total tokens
	expectedTokens := (1000 + 500) + (500 + 200) + (2000 + 1000)
	if metrics.TotalTokens != expectedTokens {
		t.Errorf("Expected total tokens %d, got %d", expectedTokens, metrics.TotalTokens)
	}

	// Check total cost
	expectedCost := 0.05 + 0.02 + 0.10
	if metrics.TotalCost != expectedCost {
		t.Errorf("Expected total cost %.6f, got %.6f", expectedCost, metrics.TotalCost)
	}
}

// TestRecordToolExecution_SuccessAndFailure tests recording both success and failure
func TestRecordToolExecution_SuccessAndFailure(t *testing.T) {
	mc := NewMetricsCollector()

	// Successful execution
	mc.RecordToolExecution("agent1", "web-search", 100*time.Millisecond, true, 1000, 500, 0.05)

	// Failed execution
	mc.RecordToolExecution("agent1", "web-search", 50*time.Millisecond, false, 500, 0, 0.02)

	metrics := mc.GetSystemMetrics()
	agent1 := metrics.AgentMetrics["agent1"]
	tool := agent1.ToolMetrics["web-search"]

	if tool.ExecutionCount != 2 {
		t.Errorf("Expected 2 executions, got %d", tool.ExecutionCount)
	}
	if tool.SuccessCount != 1 {
		t.Errorf("Expected 1 success, got %d", tool.SuccessCount)
	}
	if tool.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", tool.ErrorCount)
	}
}

// TestRecordToolExecution_TokenTracking tests input/output token tracking
func TestRecordToolExecution_TokenTracking(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordToolExecution("agent1", "api-call", 100*time.Millisecond, true, 5000, 2000, 0.25)

	metrics := mc.GetSystemMetrics()
	agent1 := metrics.AgentMetrics["agent1"]
	tool := agent1.ToolMetrics["api-call"]

	if tool.TotalInputTokens != 5000 {
		t.Errorf("Expected 5000 input tokens, got %d", tool.TotalInputTokens)
	}
	if tool.TotalOutputTokens != 2000 {
		t.Errorf("Expected 2000 output tokens, got %d", tool.TotalOutputTokens)
	}
	if tool.TotalCost != 0.25 {
		t.Errorf("Expected cost 0.25, got %.6f", tool.TotalCost)
	}
}

// TestRecordToolExecution_DurationTracking tests duration min/max/average
func TestRecordToolExecution_DurationTracking(t *testing.T) {
	mc := NewMetricsCollector()

	// First execution: 100ms
	mc.RecordToolExecution("agent1", "db-query", 100*time.Millisecond, true, 1000, 500, 0.05)

	// Second execution: 200ms (max)
	mc.RecordToolExecution("agent1", "db-query", 200*time.Millisecond, true, 2000, 1000, 0.10)

	// Third execution: 50ms (min)
	mc.RecordToolExecution("agent1", "db-query", 50*time.Millisecond, true, 500, 250, 0.03)

	metrics := mc.GetSystemMetrics()
	agent1 := metrics.AgentMetrics["agent1"]
	tool := agent1.ToolMetrics["db-query"]

	if tool.MinDuration != 50*time.Millisecond {
		t.Errorf("Expected min duration 50ms, got %v", tool.MinDuration)
	}
	if tool.MaxDuration != 200*time.Millisecond {
		t.Errorf("Expected max duration 200ms, got %v", tool.MaxDuration)
	}

	// Average: (100 + 200 + 50) / 3 = 350ms / 3 = 116.666...ms
	// The actual calculation is: tool.TotalDuration / time.Duration(tool.ExecutionCount)
	// which is (350ms) / 3 = 116.666...ms
	expectedAvg := (100 + 200 + 50) * time.Millisecond / time.Duration(3)
	if tool.AverageDuration != expectedAvg {
		t.Errorf("Expected average duration %v, got %v", expectedAvg, tool.AverageDuration)
	}
}

// TestRecordToolExecution_Disabled tests disabled metrics collector
func TestRecordToolExecution_Disabled(t *testing.T) {
	mc := NewMetricsCollector()
	mc.Disable()

	mc.RecordToolExecution("agent1", "test-tool", 100*time.Millisecond, true, 1000, 500, 0.05)

	metrics := mc.GetSystemMetrics()
	if metrics.TotalTokens != 0 {
		t.Errorf("Expected 0 tokens when disabled, got %d", metrics.TotalTokens)
	}
	if metrics.TotalCost != 0 {
		t.Errorf("Expected 0 cost when disabled, got %.6f", metrics.TotalCost)
	}
}

// TestRecordToolExecution_CostAggregation tests cost aggregation across tools
func TestRecordToolExecution_CostAggregation(t *testing.T) {
	mc := NewMetricsCollector()

	// Record multiple tool executions with different costs
	tools := []struct {
		name  string
		cost  float64
		input int
		output int
	}{
		{"web-search", 0.05, 1000, 500},
		{"calculator", 0.02, 500, 200},
		{"email", 0.10, 2000, 1000},
		{"file-ops", 0.01, 100, 50},
	}

	for _, tool := range tools {
		mc.RecordToolExecution("agent1", tool.name, 100*time.Millisecond, true, tool.input, tool.output, tool.cost)
	}

	metrics := mc.GetSystemMetrics()

	// Check total cost is sum of all tool costs (with floating point tolerance)
	expectedCost := 0.05 + 0.02 + 0.10 + 0.01
	if metrics.TotalCost < expectedCost-0.000001 || metrics.TotalCost > expectedCost+0.000001 {
		t.Errorf("Expected total cost %.6f, got %.6f", expectedCost, metrics.TotalCost)
	}

	// Check per-tool costs
	agent1 := metrics.AgentMetrics["agent1"]
	for _, tool := range tools {
		toolMetrics := agent1.ToolMetrics[tool.name]
		if toolMetrics.TotalCost < tool.cost-0.000001 || toolMetrics.TotalCost > tool.cost+0.000001 {
			t.Errorf("Tool %s: expected cost %.6f, got %.6f", tool.name, tool.cost, toolMetrics.TotalCost)
		}
	}
}

// TestRecordToolExecution_SessionCostTracking tests session-level cost tracking
func TestRecordToolExecution_SessionCostTracking(t *testing.T) {
	mc := NewMetricsCollector()

	// Record first tool
	mc.RecordToolExecution("agent1", "tool1", 100*time.Millisecond, true, 1000, 500, 0.05)

	// Record second tool
	mc.RecordToolExecution("agent1", "tool2", 100*time.Millisecond, true, 2000, 1000, 0.10)

	// Check session cost
	sessionTokens, sessionCost := mc.GetSessionCost()
	expectedTokens := 1000 + 500 + 2000 + 1000
	expectedCost := 0.05 + 0.10

	if sessionTokens != expectedTokens {
		t.Errorf("Expected session tokens %d, got %d", expectedTokens, sessionTokens)
	}
	if sessionCost < expectedCost-0.000001 || sessionCost > expectedCost+0.000001 {
		t.Errorf("Expected session cost %.6f, got %.6f", expectedCost, sessionCost)
	}

	// Reset session cost
	mc.ResetSessionCost()
	sessionTokens, sessionCost = mc.GetSessionCost()

	if sessionTokens != 0 {
		t.Errorf("Expected session tokens 0 after reset, got %d", sessionTokens)
	}
	if sessionCost != 0 {
		t.Errorf("Expected session cost 0 after reset, got %.6f", sessionCost)
	}

	// Verify total cost is unchanged
	metrics := mc.GetSystemMetrics()
	expectedTotal := 0.15
	if metrics.TotalCost < expectedTotal-0.000001 || metrics.TotalCost > expectedTotal+0.000001 {
		t.Errorf("Expected total cost %.6f (unchanged), got %.6f", expectedTotal, metrics.TotalCost)
	}
}

// TestRecordToolExecution_ZeroTokens tests recording with zero tokens
func TestRecordToolExecution_ZeroTokens(t *testing.T) {
	mc := NewMetricsCollector()

	// Tool that doesn't use tokens (e.g., file operation)
	mc.RecordToolExecution("agent1", "file-read", 50*time.Millisecond, true, 0, 0, 0)

	metrics := mc.GetSystemMetrics()
	agent1 := metrics.AgentMetrics["agent1"]
	tool := agent1.ToolMetrics["file-read"]

	if tool.ExecutionCount != 1 {
		t.Errorf("Expected 1 execution, got %d", tool.ExecutionCount)
	}
	if tool.TotalInputTokens != 0 {
		t.Errorf("Expected 0 input tokens, got %d", tool.TotalInputTokens)
	}
	if tool.TotalOutputTokens != 0 {
		t.Errorf("Expected 0 output tokens, got %d", tool.TotalOutputTokens)
	}
	if tool.TotalCost != 0 {
		t.Errorf("Expected 0 cost, got %.6f", tool.TotalCost)
	}
}

// TestRecordToolExecution_RepeatCalls tests repeated calls to same tool
func TestRecordToolExecution_RepeatCalls(t *testing.T) {
	mc := NewMetricsCollector()

	// Record same tool 5 times
	for i := 0; i < 5; i++ {
		mc.RecordToolExecution("agent1", "search", 100*time.Millisecond, true, 1000, 500, 0.05)
	}

	metrics := mc.GetSystemMetrics()
	agent1 := metrics.AgentMetrics["agent1"]
	tool := agent1.ToolMetrics["search"]

	if tool.ExecutionCount != 5 {
		t.Errorf("Expected 5 executions, got %d", tool.ExecutionCount)
	}
	if tool.SuccessCount != 5 {
		t.Errorf("Expected 5 successes, got %d", tool.SuccessCount)
	}

	// Total tokens: 5 * (1000 + 500) = 7500
	expectedTokens := 5 * (1000 + 500)
	if tool.TotalInputTokens+tool.TotalOutputTokens != expectedTokens {
		t.Errorf("Expected %d total tokens, got %d", expectedTokens, tool.TotalInputTokens+tool.TotalOutputTokens)
	}

	// Total cost: 5 * 0.05 = 0.25
	expectedCost := 5 * 0.05
	if tool.TotalCost != expectedCost {
		t.Errorf("Expected cost %.6f, got %.6f", expectedCost, tool.TotalCost)
	}
}
