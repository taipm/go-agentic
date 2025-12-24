package crewai

import (
	"testing"
	"time"
)

// TestEstimateTokens verifies token estimation accuracy
// Uses heuristic rules based on OpenAI's BPE tokenization patterns:
//   - English text: ~4 chars/token
//   - Numbers: ~2 chars/token
//   - Punctuation: ~1 char/token
//   - CJK: ~1.5 chars/token
//   - Other Unicode: ~2 chars/token
func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name    string
		content string
		min     int // minimum expected tokens
		max     int // maximum expected tokens (allows for 10% safety margin)
	}{
		{
			name:    "empty content",
			content: "",
			min:     0,
			max:     0,
		},
		{
			name:    "single English letter",
			content: "a",
			min:     1,
			max:     1, // minimum 1 token
		},
		{
			name:    "English word (hello)",
			content: "hello",
			min:     1,
			max:     2, // 5 letters × 0.25 = 1.25, +10% = ~2
		},
		{
			name:    "English sentence",
			content: "Hello world", // 10 letters + 1 space
			min:     3,
			max:     5, // letters: 10×0.25=2.5, space: 0.5, total ~3, +10%
		},
		{
			name:    "Numbers only",
			content: "12345678",
			min:     4,
			max:     6, // 8 digits × 0.5 = 4, +10%
		},
		{
			name:    "Punctuation heavy",
			content: "{}[]().,!?",
			min:     10,
			max:     12, // 10 punctuation × 1.0 = 10, +10%
		},
		{
			name:    "JSON structure",
			content: `{"name":"test","value":123}`,
			min:     15,
			max:     25, // mix of punctuation (high) + letters (low) + numbers
		},
		{
			name:    "Code snippet",
			content: "func main() { fmt.Println(42) }",
			min:     15,
			max:     25, // letters + punctuation + spaces
		},
		{
			name:    "Chinese text",
			content: "你好世界", // 4 CJK characters
			min:     6,
			max:     8, // 4 × 1.5 = 6, +10%
		},
		{
			name:    "Japanese text",
			content: "こんにちは", // 5 Hiragana characters
			min:     7,
			max:     10, // 5 × 1.5 = 7.5, +10%
		},
		{
			name:    "Korean text",
			content: "안녕하세요", // 5 Hangul syllables
			min:     7,
			max:     10, // 5 × 1.5 = 7.5, +10%
		},
		{
			name:    "Vietnamese text",
			content: "Xin chào thế giới", // Vietnamese with diacritics
			min:     6,
			max:     12, // 13 letters (ASCII+diacritics) × ~0.35 + 3 spaces × 0.5 = ~6-7, +10%
		},
		{
			name:    "Russian text",
			content: "Привет мир", // Cyrillic
			min:     5,
			max:     10, // 9 Cyrillic × 0.5 + 1 space × 0.5 = 5, +10%
		},
		{
			name:    "Mixed content",
			content: "Hello 世界! 123",
			min:     8,
			max:     15, // English + CJK + punctuation + numbers
		},
	}

	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.EstimateTokens(tt.content)
			if got < tt.min || got > tt.max {
				t.Errorf("EstimateTokens(%q) = %d, want between %d and %d", tt.content, got, tt.min, tt.max)
			}
		})
	}
}

// TestCalculateCost verifies cost calculation with configurable pricing
// Default: gpt-4o-mini pricing ($0.15 per 1M input tokens)
func TestCalculateCost(t *testing.T) {
	t.Run("default pricing (gpt-4o-mini)", func(t *testing.T) {
		agent := &Agent{
			ID:   "test-agent",
			Name: "Test Agent",
			// No pricing configured - should use default $0.15/1M
		}

		tests := []struct {
			tokens   int
			expected float64
		}{
			{0, 0.0},
			{1, 0.00000015},           // 1 × ($0.15 / 1M)
			{1000, 0.00015},           // 1K tokens = $0.00015
			{1000000, 0.15},           // 1M tokens = $0.15
			{10000000, 1.5},           // 10M tokens = $1.50
		}

		for _, tt := range tests {
			got := agent.CalculateCost(tt.tokens)
			epsilon := 0.00000001
			if diff := got - tt.expected; diff < -epsilon || diff > epsilon {
				t.Errorf("CalculateCost(%d) = %.8f, want %.8f", tt.tokens, got, tt.expected)
			}
		}
	})

	t.Run("custom pricing (gpt-4o)", func(t *testing.T) {
		agent := &Agent{
			ID:                        "test-agent",
			Name:                      "Test Agent",
			InputTokenPricePerMillion: 2.50, // gpt-4o pricing
		}

		tests := []struct {
			tokens   int
			expected float64
		}{
			{0, 0.0},
			{1000000, 2.50},  // 1M tokens = $2.50
			{10000000, 25.0}, // 10M tokens = $25.00
		}

		for _, tt := range tests {
			got := agent.CalculateCost(tt.tokens)
			epsilon := 0.00000001
			if diff := got - tt.expected; diff < -epsilon || diff > epsilon {
				t.Errorf("CalculateCost(%d) with gpt-4o pricing = %.8f, want %.8f", tt.tokens, got, tt.expected)
			}
		}
	})

	t.Run("ollama (free/local)", func(t *testing.T) {
		// Note: When price is 0, it defaults to gpt-4o-mini pricing
		// To truly disable cost tracking, check if provider is "ollama"
		agent := &Agent{
			ID:                        "test-agent",
			Name:                      "Test Agent",
			InputTokenPricePerMillion: 0.0001, // Near-zero for local
		}

		got := agent.CalculateCost(1000000)
		if got > 0.001 {
			t.Errorf("Local model should have near-zero cost, got %.8f", got)
		}
	})
}

// TestCalculateOutputCost verifies output token cost calculation
func TestCalculateOutputCost(t *testing.T) {
	t.Run("default pricing (gpt-4o-mini)", func(t *testing.T) {
		agent := &Agent{ID: "test"}

		got := agent.CalculateOutputCost(1000000)
		expected := 0.60 // $0.60 per 1M output tokens
		epsilon := 0.00000001
		if diff := got - expected; diff < -epsilon || diff > epsilon {
			t.Errorf("CalculateOutputCost(1M) = %.8f, want %.8f", got, expected)
		}
	})

	t.Run("custom pricing (gpt-4o)", func(t *testing.T) {
		agent := &Agent{
			ID:                         "test",
			OutputTokenPricePerMillion: 10.0, // gpt-4o output pricing
		}

		got := agent.CalculateOutputCost(1000000)
		expected := 10.0 // $10.00 per 1M output tokens
		epsilon := 0.00000001
		if diff := got - expected; diff < -epsilon || diff > epsilon {
			t.Errorf("CalculateOutputCost(1M) with gpt-4o = %.8f, want %.8f", got, expected)
		}
	})
}

// TestCalculateTotalCost verifies combined input + output cost
func TestCalculateTotalCost(t *testing.T) {
	agent := &Agent{
		ID:                         "test",
		InputTokenPricePerMillion:  2.50,  // gpt-4o input
		OutputTokenPricePerMillion: 10.00, // gpt-4o output
	}

	// 1M input + 500K output
	got := agent.CalculateTotalCost(1000000, 500000)
	expected := 2.50 + 5.00 // $2.50 input + $5.00 output
	epsilon := 0.00000001
	if diff := got - expected; diff < -epsilon || diff > epsilon {
		t.Errorf("CalculateTotalCost(1M, 500K) = %.8f, want %.8f", got, expected)
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
