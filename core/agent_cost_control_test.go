package crewai

import (
	"testing"
	"time"
)

// TestEstimateTokens verifies token estimation accuracy
// Tests the formula: 1 token ≈ 4 characters
func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "empty content",
			content:  "",
			expected: 0,
		},
		{
			name:     "single character",
			content:  "a",
			expected: 1, // (1+3)/4 = 1
		},
		{
			name:     "four characters (exact)",
			content:  "abcd",
			expected: 1, // (4+3)/4 = 1
		},
		{
			name:     "five characters (round up)",
			content:  "abcde",
			expected: 2, // (5+3)/4 = 2
		},
		{
			name:     "eight characters",
			content:  "abcdefgh",
			expected: 2, // (8+3)/4 = 2
		},
		{
			name:     "nine characters",
			content:  "abcdefghi",
			expected: 3, // (9+3)/4 = 3
		},
		{
			name:     "typical message (11 chars)",
			content:  "Hello world", // "Hello " (6) + "world" (5) = 11
			expected: 3, // (11+3)/4 = 3
		},
		{
			name:     "large content (1000 chars)",
			content:  string(make([]byte, 1000)),
			expected: 250, // (1000+3)/4 = 250
		},
	}

	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.EstimateTokens(tt.content)
			if got != tt.expected {
				t.Errorf("EstimateTokens(%q) = %d, want %d", tt.content, got, tt.expected)
			}
		})
	}
}

// TestCalculateCost verifies cost calculation accuracy
// Uses OpenAI pricing: $0.15 per 1M input tokens
func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name     string
		tokens   int
		expected float64
	}{
		{
			name:     "zero tokens",
			tokens:   0,
			expected: 0.0,
		},
		{
			name:     "single token",
			tokens:   1,
			expected: 0.00000015, // 1 * 0.00000015
		},
		{
			name:     "1000 tokens",
			tokens:   1000,
			expected: 0.00015, // 1000 * 0.00000015 = 0.00015 (0.15 cents)
		},
		{
			name:     "1 million tokens",
			tokens:   1000000,
			expected: 0.15, // 1M tokens = $0.15
		},
		{
			name:     "10 million tokens",
			tokens:   10000000,
			expected: 1.5, // 10M tokens = $1.50
		},
		{
			name:     "large amount (100M tokens)",
			tokens:   100000000,
			expected: 15.0, // 100M tokens = $15.00
		},
	}

	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.CalculateCost(tt.tokens)
			// Use small epsilon for float comparison (allow 0.00000001 error)
			epsilon := 0.00000001
			if diff := got - tt.expected; diff < -epsilon || diff > epsilon {
				t.Errorf("CalculateCost(%d) = %.8f, want %.8f", tt.tokens, got, tt.expected)
			}
		})
	}
}

// TestResetDailyMetricsIfNeeded verifies daily reset mechanism
func TestResetDailyMetricsIfNeeded(t *testing.T) {
	t.Run("first call initializes LastResetTime", func(t *testing.T) {
		agent := &Agent{
			ID:   "test-agent",
			Name: "Test Agent",
		}

		// First call should initialize LastResetTime
		agent.ResetDailyMetricsIfNeeded()

		if agent.CostMetrics.LastResetTime.IsZero() {
			t.Error("Expected LastResetTime to be set, but got zero value")
		}
	})

	t.Run("same day no reset", func(t *testing.T) {
		agent := &Agent{
			ID:   "test-agent",
			Name: "Test Agent",
		}

		// Initialize
		agent.ResetDailyMetricsIfNeeded()
		resetTime1 := agent.CostMetrics.LastResetTime

		// Set some metrics
		agent.CostMetrics.Mutex.Lock()
		agent.CostMetrics.CallCount = 5
		agent.CostMetrics.TotalTokens = 1000
		agent.CostMetrics.DailyCost = 0.15
		agent.CostMetrics.Mutex.Unlock()

		// Call reset again (same moment)
		agent.ResetDailyMetricsIfNeeded()

		// Metrics should NOT be reset (same day)
		agent.CostMetrics.Mutex.RLock()
		callCount := agent.CostMetrics.CallCount
		totalTokens := agent.CostMetrics.TotalTokens
		dailyCost := agent.CostMetrics.DailyCost
		resetTime2 := agent.CostMetrics.LastResetTime
		agent.CostMetrics.Mutex.RUnlock()

		if callCount != 5 || totalTokens != 1000 || dailyCost != 0.15 {
			t.Errorf("Metrics should not reset on same day. Got: callCount=%d, tokens=%d, cost=%.2f", callCount, totalTokens, dailyCost)
		}

		if resetTime1 != resetTime2 {
			t.Error("LastResetTime should not change within 24 hours")
		}
	})

	t.Run("24+ hours later triggers reset", func(t *testing.T) {
		agent := &Agent{
			ID:   "test-agent",
			Name: "Test Agent",
		}

		// Initialize
		agent.ResetDailyMetricsIfNeeded()

		// Set metrics
		agent.CostMetrics.Mutex.Lock()
		agent.CostMetrics.CallCount = 10
		agent.CostMetrics.TotalTokens = 5000
		agent.CostMetrics.DailyCost = 0.75
		oldTime := time.Now().Add(-25 * time.Hour) // 25 hours ago
		agent.CostMetrics.LastResetTime = oldTime
		agent.CostMetrics.Mutex.Unlock()

		// Trigger reset
		agent.ResetDailyMetricsIfNeeded()

		// Metrics should be reset
		agent.CostMetrics.Mutex.RLock()
		callCount := agent.CostMetrics.CallCount
		totalTokens := agent.CostMetrics.TotalTokens
		dailyCost := agent.CostMetrics.DailyCost
		newTime := agent.CostMetrics.LastResetTime
		agent.CostMetrics.Mutex.RUnlock()

		if callCount != 0 || totalTokens != 0 || dailyCost != 0 {
			t.Errorf("Metrics should reset after 24+ hours. Got: callCount=%d, tokens=%d, cost=%.2f", callCount, totalTokens, dailyCost)
		}

		if newTime.Before(oldTime) {
			t.Error("LastResetTime should be updated to current time")
		}
	})
}

// TestCheckCostLimits verifies enforcement of cost limits
func TestCheckCostLimits(t *testing.T) {
	t.Run("block mode - under limit", func(t *testing.T) {
		agent := &Agent{
			ID:                "test-agent",
			Name:              "Test Agent",
			MaxTokensPerCall:  1000,
			MaxTokensPerDay:   50000,
			MaxCostPerDay:     10.0,
			EnforceCostLimits: true, // Block mode
		}

		// Check 500 tokens (under 1000 limit)
		err := agent.CheckCostLimits(500)
		if err != nil {
			t.Errorf("CheckCostLimits(500) should succeed in block mode, got error: %v", err)
		}
	})

	t.Run("block mode - exceeds per-call limit", func(t *testing.T) {
		agent := &Agent{
			ID:                "test-agent",
			Name:              "Test Agent",
			MaxTokensPerCall:  1000,
			MaxTokensPerDay:   50000,
			MaxCostPerDay:     10.0,
			EnforceCostLimits: true, // Block mode
		}

		// Check 2000 tokens (exceeds 1000 limit)
		err := agent.CheckCostLimits(2000)
		if err == nil {
			t.Error("CheckCostLimits(2000) should fail when exceeds MaxTokensPerCall limit")
		}
	})

	t.Run("block mode - exceeds daily cost limit", func(t *testing.T) {
		agent := &Agent{
			ID:                "test-agent",
			Name:              "Test Agent",
			MaxTokensPerCall:  100000,
			MaxTokensPerDay:   50000,
			MaxCostPerDay:     10.0,
			EnforceCostLimits: true, // Block mode
		}

		// Set up metrics: $9.50 already spent today
		agent.CostMetrics.Mutex.Lock()
		agent.CostMetrics.DailyCost = 9.50
		agent.CostMetrics.LastResetTime = time.Now() // Today
		agent.CostMetrics.Mutex.Unlock()

		// Try to use 20M more tokens (~$3.00)
		// This would push total to $12.50, exceeding $10 limit
		err := agent.CheckCostLimits(20000000)
		if err == nil {
			t.Error("CheckCostLimits should fail when daily cost would exceed limit")
		}
	})

	t.Run("warn mode - no error returned", func(t *testing.T) {
		agent := &Agent{
			ID:                "test-agent",
			Name:              "Test Agent",
			MaxTokensPerCall:  1000,
			MaxTokensPerDay:   50000,
			MaxCostPerDay:     10.0,
			EnforceCostLimits: false, // Warn mode
		}

		// Check 2000 tokens (exceeds limit but in warn mode)
		err := agent.CheckCostLimits(2000)
		if err != nil {
			t.Errorf("CheckCostLimits in warn mode should not return error, got: %v", err)
		}
	})

	t.Run("warn mode - allows execution", func(t *testing.T) {
		agent := &Agent{
			ID:                "test-agent",
			Name:              "Test Agent",
			MaxTokensPerCall:  1000,
			MaxTokensPerDay:   1.0, // Very low limit
			MaxCostPerDay:     0.01,
			CostAlertThreshold: 0.80,
			EnforceCostLimits: false, // Warn mode
		}

		// Set up metrics: already at 80% of daily limit
		agent.CostMetrics.Mutex.Lock()
		agent.CostMetrics.DailyCost = 0.008 // 80% of $0.01
		agent.CostMetrics.LastResetTime = time.Now()
		agent.CostMetrics.Mutex.Unlock()

		// Check high tokens (would be blocked in block mode)
		err := agent.CheckCostLimits(100000000)
		if err != nil {
			t.Errorf("CheckCostLimits in warn mode should allow execution, got error: %v", err)
		}
	})
}

// TestUpdateCostMetrics verifies metric tracking
func TestUpdateCostMetrics(t *testing.T) {
	t.Run("single update", func(t *testing.T) {
		agent := &Agent{
			ID:   "test-agent",
			Name: "Test Agent",
		}

		// Update with 1000 tokens, $0.15 cost
		agent.UpdateCostMetrics(1000, 0.15)

		agent.CostMetrics.Mutex.RLock()
		callCount := agent.CostMetrics.CallCount
		totalTokens := agent.CostMetrics.TotalTokens
		dailyCost := agent.CostMetrics.DailyCost
		agent.CostMetrics.Mutex.RUnlock()

		if callCount != 1 {
			t.Errorf("CallCount should be 1, got %d", callCount)
		}
		if totalTokens != 1000 {
			t.Errorf("TotalTokens should be 1000, got %d", totalTokens)
		}
		epsilon := 0.00000001
		if diff := dailyCost - 0.15; diff < -epsilon || diff > epsilon {
			t.Errorf("DailyCost should be 0.15, got %.8f", dailyCost)
		}
	})

	t.Run("multiple updates accumulate", func(t *testing.T) {
		agent := &Agent{
			ID:   "test-agent",
			Name: "Test Agent",
		}

		// First update: 1000 tokens, $0.15
		agent.UpdateCostMetrics(1000, 0.15)

		// Second update: 2000 tokens, $0.30
		agent.UpdateCostMetrics(2000, 0.30)

		// Third update: 500 tokens, $0.075
		agent.UpdateCostMetrics(500, 0.075)

		agent.CostMetrics.Mutex.RLock()
		callCount := agent.CostMetrics.CallCount
		totalTokens := agent.CostMetrics.TotalTokens
		dailyCost := agent.CostMetrics.DailyCost
		agent.CostMetrics.Mutex.RUnlock()

		if callCount != 3 {
			t.Errorf("CallCount should be 3, got %d", callCount)
		}
		if totalTokens != 3500 { // 1000 + 2000 + 500
			t.Errorf("TotalTokens should be 3500, got %d", totalTokens)
		}
		expectedCost := 0.15 + 0.30 + 0.075 // 0.525
		epsilon := 0.00000001
		if diff := dailyCost - expectedCost; diff < -epsilon || diff > epsilon {
			t.Errorf("DailyCost should be %.3f, got %.8f", expectedCost, dailyCost)
		}
	})

	t.Run("metrics are thread safe", func(t *testing.T) {
		agent := &Agent{
			ID:   "test-agent",
			Name: "Test Agent",
		}

		// Simulate concurrent updates from 10 goroutines
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				// Each goroutine updates 100 times
				for j := 0; j < 100; j++ {
					agent.UpdateCostMetrics(10, 0.0000015)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		agent.CostMetrics.Mutex.RLock()
		callCount := agent.CostMetrics.CallCount
		totalTokens := agent.CostMetrics.TotalTokens
		agent.CostMetrics.Mutex.RUnlock()

		if callCount != 1000 { // 10 goroutines * 100 updates
			t.Errorf("CallCount should be 1000 (10*100), got %d", callCount)
		}
		if totalTokens != 10000 { // 10 goroutines * 100 * 10 tokens
			t.Errorf("TotalTokens should be 10000, got %d", totalTokens)
		}
	})
}

// TestCostControlIntegration verifies the complete cost control workflow
func TestCostControlIntegration(t *testing.T) {
	t.Run("complete workflow - block mode", func(t *testing.T) {
		agent := &Agent{
			ID:                "test-agent",
			Name:              "Test Agent",
			MaxTokensPerCall:  2000,
			MaxTokensPerDay:   50000,
			MaxCostPerDay:     10.0,
			CostAlertThreshold: 0.80,
			EnforceCostLimits: true, // Block mode
		}

		// Scenario 1: Normal request (should pass)
		request1 := "What is 2+2?" // ~3 tokens
		tokens1 := agent.EstimateTokens(request1)
		err1 := agent.CheckCostLimits(tokens1)
		if err1 != nil {
			t.Errorf("Normal request should be allowed, got error: %v", err1)
		}

		// After successful execution, update metrics
		cost1 := agent.CalculateCost(tokens1)
		agent.UpdateCostMetrics(tokens1, cost1)

		// Scenario 2: Request that exceeds per-call limit (should block)
		largeRequest := string(make([]byte, 8004)) // ~2001 tokens (exceeds 2000 limit)
		tokens2 := agent.EstimateTokens(largeRequest)
		err2 := agent.CheckCostLimits(tokens2)
		if err2 == nil {
			t.Error("Request exceeding MaxTokensPerCall should be blocked")
		}

		// Verify metrics only include first request
		agent.CostMetrics.Mutex.RLock()
		callCount := agent.CostMetrics.CallCount
		totalTokens := agent.CostMetrics.TotalTokens
		agent.CostMetrics.Mutex.RUnlock()

		if callCount != 1 {
			t.Errorf("Only successful request should update metrics, got %d calls", callCount)
		}
		if totalTokens != tokens1 {
			t.Errorf("Metrics should only include first request: got %d, want %d", totalTokens, tokens1)
		}
	})

	t.Run("complete workflow - warn mode", func(t *testing.T) {
		agent := &Agent{
			ID:                "test-agent",
			Name:              "Test Agent",
			MaxTokensPerCall:  2000,
			MaxTokensPerDay:   50000,
			MaxCostPerDay:     10.0,
			CostAlertThreshold: 0.80,
			EnforceCostLimits: false, // Warn mode
		}

		// Large request that would be blocked in block mode
		largeRequest := string(make([]byte, 8000)) // ~2000 tokens
		tokens := agent.EstimateTokens(largeRequest)

		// Check should pass (warn mode)
		err := agent.CheckCostLimits(tokens)
		if err != nil {
			t.Errorf("Warn mode should allow execution, got error: %v", err)
		}

		// Metrics still update
		cost := agent.CalculateCost(tokens)
		agent.UpdateCostMetrics(tokens, cost)

		agent.CostMetrics.Mutex.RLock()
		callCount := agent.CostMetrics.CallCount
		agent.CostMetrics.Mutex.RUnlock()

		if callCount != 1 {
			t.Errorf("Warn mode should update metrics, got %d calls", callCount)
		}
	})
}
