package crewai

import (
	"testing"
)

// TestQuotaEnforcementFlow_MemoryBlock tests full flow with memory blocking
func TestQuotaEnforcementFlow_MemoryBlock(t *testing.T) {
	// Setup: Create mock crew with memory quota
	agent := &Agent{
		ID:   "agent1",
		Name: "Agent 1",
		Primary: &ModelConfig{
			Model:    "gpt-4",
			Provider: "openai",
		},
		Metadata: &AgentMetadata{
			Quotas: AgentQuotaLimits{
				MaxMemoryPerCall:  256, // Strict limit
				BlockOnQuotaExceed: true,
			},
			Memory: AgentMemoryMetrics{},
			Performance: AgentPerformanceMetrics{
				MaxConsecutiveErrors: 5,
				MaxErrorsPerDay:      50,
			},
			Cost: AgentCostMetrics{},
		},
	}

	crew := &Crew{
		Agents: []*Agent{agent},
	}

	executor := NewCrewExecutor(crew, "")

	// Simulate execution with high memory usage
	agent.Metadata.Memory.CurrentMemoryMB = 512 // Exceeds 256 MB limit

	// Should trigger memory quota error
	err := agent.CheckMemoryQuota()
	if err == nil {
		t.Fatal("Expected memory quota error")
	}

	expectedMsg := "agent 'agent1': memory quota exceeded - used 512 MB, max 256 MB per call"
	if err.Error() != expectedMsg {
		t.Errorf("Unexpected error: got %v, want %v", err.Error(), expectedMsg)
	}

	_ = executor // Use executor to avoid unused warning
}

// TestQuotaEnforcementFlow_ErrorQuotaBlock tests full flow with error quota blocking
func TestQuotaEnforcementFlow_ErrorQuotaBlock(t *testing.T) {
	agent := &Agent{
		ID:   "agent1",
		Name: "Agent 1",
		Metadata: &AgentMetadata{
			Quotas: AgentQuotaLimits{
				BlockOnQuotaExceed: true,
			},
			Performance: AgentPerformanceMetrics{
				MaxConsecutiveErrors: 3,
				MaxErrorsPerDay:      10,
				ConsecutiveErrors:    0,
			},
		},
	}

	// Simulate 3 consecutive errors
	agent.UpdatePerformanceMetrics(false, "error 1")
	agent.UpdatePerformanceMetrics(false, "error 2")
	agent.UpdatePerformanceMetrics(false, "error 3")

	// Should now trigger error quota blocking
	err := agent.CheckErrorQuota()
	if err == nil {
		t.Fatal("Expected error quota error")
	}

	expectedMsg := "agent 'agent1': maximum consecutive errors exceeded - 3 errors, max 3 allowed"
	if err.Error() != expectedMsg {
		t.Errorf("Unexpected error: got %v, want %v", err.Error(), expectedMsg)
	}
}

// TestQuotaEnforcementFlow_MemoryWarning tests warn-only behavior
func TestQuotaEnforcementFlow_MemoryWarning(t *testing.T) {
	agent := &Agent{
		ID:   "agent1",
		Name: "Agent 1",
		Metadata: &AgentMetadata{
			Quotas: AgentQuotaLimits{
				MaxMemoryPerCall:  256,
				BlockOnQuotaExceed: false, // Warn mode
			},
			Memory: AgentMemoryMetrics{
				CurrentMemoryMB: 512, // Exceeds limit but warn-only
			},
		},
	}

	// In warn mode, no error should be returned
	err := agent.CheckMemoryQuota()
	if err != nil {
		t.Errorf("CheckMemoryQuota in warn mode should not return error, got: %v", err)
	}
}

// TestQuotaEnforcementFlow_ErrorWarning tests warn-only error behavior
func TestQuotaEnforcementFlow_ErrorWarning(t *testing.T) {
	agent := &Agent{
		ID:   "agent1",
		Name: "Agent 1",
		Metadata: &AgentMetadata{
			Quotas: AgentQuotaLimits{
				BlockOnQuotaExceed: false, // Warn mode
			},
			Performance: AgentPerformanceMetrics{
				MaxConsecutiveErrors: 3,
				MaxErrorsPerDay:      10,
				ConsecutiveErrors:    5, // Exceeds limit
			},
		},
	}

	// In warn mode, no error should be returned
	err := agent.CheckErrorQuota()
	if err != nil {
		t.Errorf("CheckErrorQuota in warn mode should not return error, got: %v", err)
	}
}

// TestQuotaEnforcementFlow_MetricsUpdateOnly tests that metrics only update when quota passes
func TestQuotaEnforcementFlow_MetricsUpdateOnly(t *testing.T) {
	agent := &Agent{
		ID:   "agent1",
		Name: "Agent 1",
		Metadata: &AgentMetadata{
			Quotas: AgentQuotaLimits{
				MaxMemoryPerCall:  256,
				BlockOnQuotaExceed: true,
			},
			Memory: AgentMemoryMetrics{},
			Performance: AgentPerformanceMetrics{
				MaxConsecutiveErrors: 5,
			},
			Cost: AgentCostMetrics{},
		},
	}

	// Update metrics with valid memory
	agent.Metadata.Memory.CurrentMemoryMB = 100 // Under limit
	memoryPassed := agent.CheckMemoryQuota()
	if memoryPassed != nil {
		t.Errorf("CheckMemoryQuota should pass with 100 MB limit, got error: %v", memoryPassed)
	}

	// Update performance metrics
	agent.UpdatePerformanceMetrics(true, "")

	agent.Metadata.Mutex.RLock()
	successfulCalls := agent.Metadata.Performance.SuccessfulCalls
	agent.Metadata.Mutex.RUnlock()

	if successfulCalls != 1 {
		t.Errorf("SuccessfulCalls should be 1 after passing quota, got %d", successfulCalls)
	}

	// Now exceed quota and verify metrics aren't double-updated
	agent.Metadata.Memory.CurrentMemoryMB = 512 // Exceeds 256 limit
	memoryFailed := agent.CheckMemoryQuota()
	if memoryFailed == nil {
		t.Error("CheckMemoryQuota should fail with 512 MB (limit 256)")
	}

	agent.Metadata.Mutex.RLock()
	successfulCallsAfter := agent.Metadata.Performance.SuccessfulCalls
	agent.Metadata.Mutex.RUnlock()

	// Should still be 1, not incremented due to failed quota check
	if successfulCallsAfter != 1 {
		t.Errorf("SuccessfulCalls should remain 1 after quota failure, got %d", successfulCallsAfter)
	}
}
