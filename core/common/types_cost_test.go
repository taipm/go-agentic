package common

import (
	"testing"
	"time"
)

// ============================================================================
// CALCULATE COST TESTS
// ============================================================================

// TestCalculateCost_NilAgent tests nil agent handling
func TestCalculateCost_NilAgent(t *testing.T) {
	var agent *Agent
	cost := agent.CalculateCost(100, 50)

	if cost != 0 {
		t.Errorf("Expected cost 0 for nil agent, got %f", cost)
	}
}

// TestCalculateCost_ZeroTokens tests zero tokens cost
func TestCalculateCost_ZeroTokens(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	cost := agent.CalculateCost(0, 0)

	if cost != 0 {
		t.Errorf("Expected cost 0 for zero tokens, got %f", cost)
	}
}

// TestCalculateCost_InputOnly tests cost with input tokens only
func TestCalculateCost_InputOnly(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	// 1M input tokens should cost $30
	cost := agent.CalculateCost(1_000_000, 0)

	expected := 30.0
	if cost != expected {
		t.Errorf("Expected cost $%.6f for 1M input tokens, got $%.6f", expected, cost)
	}
}

// TestCalculateCost_OutputOnly tests cost with output tokens only
func TestCalculateCost_OutputOnly(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	// 1M output tokens should cost $60
	cost := agent.CalculateCost(0, 1_000_000)

	expected := 60.0
	if cost != expected {
		t.Errorf("Expected cost $%.6f for 1M output tokens, got $%.6f", expected, cost)
	}
}

// TestCalculateCost_Combined tests cost with both input and output
func TestCalculateCost_Combined(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	// 1M input ($30) + 1M output ($60) = $90
	cost := agent.CalculateCost(1_000_000, 1_000_000)

	expected := 90.0
	if cost != expected {
		t.Errorf("Expected cost $%.6f for 1M input + 1M output, got $%.6f", expected, cost)
	}
}

// TestCalculateCost_SmallTokens tests cost with small token counts
func TestCalculateCost_SmallTokens(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	// 1000 input tokens + 500 output tokens
	cost := agent.CalculateCost(1000, 500)

	// (1000 * 30/1M) + (500 * 60/1M) = 0.03 + 0.03 = 0.06
	expected := (1000 * 30.0 / 1_000_000.0) + (500 * 60.0 / 1_000_000.0)

	// Use approximate comparison for floating point
	if cost < expected-0.000001 || cost > expected+0.000001 {
		t.Errorf("Expected cost $%.6f, got $%.6f", expected, cost)
	}
}

// TestCalculateCost_TypicalRequest tests cost for typical LLM request
func TestCalculateCost_TypicalRequest(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	// Typical: 5000 input, 2000 output
	cost := agent.CalculateCost(5000, 2000)

	// (5000 * 30/1M) + (2000 * 60/1M)
	inputCost := 5000 * 30.0 / 1_000_000.0
	outputCost := 2000 * 60.0 / 1_000_000.0
	expected := inputCost + outputCost

	if cost != expected {
		t.Errorf("Expected cost $%.6f, got $%.6f", expected, cost)
	}
}

// ============================================================================
// CHECK COST LIMITS TESTS
// ============================================================================

// TestCheckCostLimits_NilAgent tests nil agent handling
func TestCheckCostLimits_NilAgent(t *testing.T) {
	var agent *Agent
	err := agent.CheckCostLimits(1000)

	if err != nil {
		t.Errorf("Expected no error for nil agent, got %v", err)
	}
}

// TestCheckCostLimits_NoCostMetrics tests missing cost metrics
func TestCheckCostLimits_NoCostMetrics(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	err := agent.CheckCostLimits(1000)

	if err != nil {
		t.Errorf("Expected no error when cost metrics nil, got %v", err)
	}
}

// TestCheckCostLimits_NoQuota tests missing quota configuration
func TestCheckCostLimits_NoQuota(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		CostMetrics: &AgentCostMetrics{
			CallCount:     0,
			TotalTokens:   0,
			DailyCost:     0,
			LastResetTime: time.Now(),
		},
	}

	err := agent.CheckCostLimits(1000)

	if err != nil {
		t.Errorf("Expected no error when quota is nil, got %v", err)
	}
}

// TestCheckCostLimits_ExceedsMaxTokensPerCall tests token per call limit
func TestCheckCostLimits_ExceedsMaxTokensPerCall(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Quota: &AgentQuotaLimits{
			MaxTokensPerCall: 1000,
		},
		CostMetrics: &AgentCostMetrics{
			CallCount:     0,
			TotalTokens:   0,
			DailyCost:     0,
			LastResetTime: time.Now(),
		},
	}

	err := agent.CheckCostLimits(2000)

	if err == nil {
		t.Fatalf("Expected error for exceeding MaxTokensPerCall, got none")
	}

	if err.Error() != "agent 'test-agent': estimated tokens (2000) exceeds max per call (1000)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestCheckCostLimits_ExceedsMaxTokensPerDay tests daily token limit
func TestCheckCostLimits_ExceedsMaxTokensPerDay(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Quota: &AgentQuotaLimits{
			MaxTokensPerDay: 100_000,
		},
		CostMetrics: &AgentCostMetrics{
			CallCount:     10,
			TotalTokens:   98_000,
			DailyCost:     0,
			LastResetTime: time.Now(),
		},
	}

	// Try to add 3000 more tokens (would exceed 100k limit)
	err := agent.CheckCostLimits(3000)

	if err == nil {
		t.Fatalf("Expected error for exceeding MaxTokensPerDay, got none")
	}

	if err.Error() != "agent 'test-agent': total daily tokens (101000) would exceed limit (100000)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestCheckCostLimits_ExceedsMaxCostPerDay tests daily cost limit
func TestCheckCostLimits_ExceedsMaxCostPerDay(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Quota: &AgentQuotaLimits{
			MaxCostPerDay: 1.0, // $1.00 per day
		},
		CostMetrics: &AgentCostMetrics{
			CallCount:     5,
			TotalTokens:   30_000_000, // 30M tokens already used
			DailyCost:     0.90,        // $0.90 already spent
			LastResetTime: time.Now(),
		},
	}

	// Estimate: 60% of 1M tokens = 600k input + 400k output
	// Cost: (600k * $30/1M) + (400k * $60/1M) = $18 + $24 = $42 estimated
	// This should exceed the $1.00 limit
	err := agent.CheckCostLimits(1_000_000)

	if err == nil {
		t.Fatalf("Expected error for exceeding MaxCostPerDay, got none")
	}

	if err.Error() != "agent 'test-agent': daily cost ($42.9000) would exceed limit ($1.0000)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestCheckCostLimits_WithinLimits tests that valid requests pass
func TestCheckCostLimits_WithinLimits(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Quota: &AgentQuotaLimits{
			MaxTokensPerCall: 100_000,
			MaxTokensPerDay:  1_000_000,
			MaxCostPerDay:    100.0,
		},
		CostMetrics: &AgentCostMetrics{
			CallCount:     5,
			TotalTokens:   100_000,
			DailyCost:     5.0,
			LastResetTime: time.Now(),
		},
	}

	err := agent.CheckCostLimits(50_000)

	if err != nil {
		t.Errorf("Expected no error for within limits, got %v", err)
	}
}

// TestCheckCostLimits_ZeroLimits tests that zero limits are ignored
func TestCheckCostLimits_ZeroLimits(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Quota: &AgentQuotaLimits{
			MaxTokensPerCall: 0, // No limit
			MaxTokensPerDay:  0, // No limit
			MaxCostPerDay:    0, // No limit
		},
		CostMetrics: &AgentCostMetrics{
			CallCount:     0,
			TotalTokens:   0,
			DailyCost:     0,
			LastResetTime: time.Now(),
		},
	}

	// Even massive token count should pass
	err := agent.CheckCostLimits(100_000_000)

	if err != nil {
		t.Errorf("Expected no error with zero limits, got %v", err)
	}
}

// TestCheckCostLimits_DailyResetAfter24Hours tests daily cost reset
func TestCheckCostLimits_DailyResetAfter24Hours(t *testing.T) {
	// Note: CheckCostLimits doesn't reset - that happens in UpdateCostMetrics
	// This test verifies that CheckCostLimits works correctly with small estimations
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Quota: &AgentQuotaLimits{
			MaxCostPerDay: 1.0, // Very small limit
		},
		CostMetrics: &AgentCostMetrics{
			CallCount:     0,
			TotalTokens:   0,
			DailyCost:     0.5, // $0.50 already used
			LastResetTime: time.Now(),
		},
	}

	// Small estimation that fits: 1000 tokens = ~$0.03
	err := agent.CheckCostLimits(1000)

	if err != nil {
		t.Errorf("Expected no error for small estimation, got %v", err)
	}
}

// TestCheckCostLimits_MultipleChecks tests sequential limit checks
func TestCheckCostLimits_MultipleChecks(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Quota: &AgentQuotaLimits{
			MaxTokensPerCall: 100_000,
			MaxTokensPerDay:  500_000,
			MaxCostPerDay:    50.0,
		},
		CostMetrics: &AgentCostMetrics{
			CallCount:     0,
			TotalTokens:   0,
			DailyCost:     0,
			LastResetTime: time.Now(),
		},
	}

	// First request: should pass (50k < 100k per call, < 500k per day)
	err1 := agent.CheckCostLimits(50_000)
	if err1 != nil {
		t.Errorf("First check failed: %v", err1)
	}

	// Simulate update
	agent.CostMetrics.TotalTokens += 50_000
	agent.CostMetrics.DailyCost += (50_000 * 30.0 / 1_000_000.0)

	// Second request: should pass (100k per call limit still satisfied)
	err2 := agent.CheckCostLimits(80_000)
	if err2 != nil {
		t.Errorf("Second check failed: %v", err2)
	}

	// Simulate update
	agent.CostMetrics.TotalTokens += 80_000
	agent.CostMetrics.DailyCost += (80_000 * 30.0 / 1_000_000.0)

	// Third request: should pass (total 130k, still under 500k)
	err3 := agent.CheckCostLimits(50_000)
	if err3 != nil {
		t.Errorf("Third check failed: %v", err3)
	}

	// Simulate update
	agent.CostMetrics.TotalTokens += 50_000

	// Fourth request: should fail (total would be 180k + 250k = 430k still ok, but let's exceed)
	// Total is 180k + 350k = 530k > 500k
	err4 := agent.CheckCostLimits(350_000)
	if err4 == nil {
		t.Fatalf("Fourth check should have failed but didn't")
	}
}
