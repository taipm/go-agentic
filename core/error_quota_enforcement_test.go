package crewai

import (
	"testing"
)

// TestCheckErrorQuota_ConsecutiveErrorsBlock tests blocking on consecutive errors
func TestCheckErrorQuota_ConsecutiveErrorsBlock(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Metadata: &AgentMetadata{
			Quotas: AgentQuotaLimits{
				BlockOnQuotaExceed: true,
			},
			Performance: AgentPerformanceMetrics{
				ConsecutiveErrors:    5, // At limit
				MaxConsecutiveErrors: 5, // Block at 5
			},
		},
	}

	err := agent.CheckErrorQuota()
	if err == nil {
		t.Error("CheckErrorQuota should return error when consecutive errors hit limit")
	}
}

// TestCheckErrorQuota_DailyErrorLimit tests blocking on daily error count
func TestCheckErrorQuota_DailyErrorLimit(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Metadata: &AgentMetadata{
			Quotas: AgentQuotaLimits{
				BlockOnQuotaExceed: true,
			},
			Performance: AgentPerformanceMetrics{
				ErrorCountToday:      51, // Exceeds daily limit
				MaxErrorsPerDay:      50, // Block at 50
				MaxConsecutiveErrors: 10,
			},
		},
	}

	err := agent.CheckErrorQuota()
	if err == nil {
		t.Error("CheckErrorQuota should return error when daily errors exceed limit")
	}
}

// TestCheckErrorQuota_WarnMode tests warn-only behavior
func TestCheckErrorQuota_WarnMode(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Metadata: &AgentMetadata{
			Quotas: AgentQuotaLimits{
				BlockOnQuotaExceed: false, // Warn mode
			},
			Performance: AgentPerformanceMetrics{
				ConsecutiveErrors:    10, // Exceeds limit
				MaxConsecutiveErrors: 5,
			},
		},
	}

	err := agent.CheckErrorQuota()
	if err != nil {
		t.Errorf("CheckErrorQuota should not return error in warn mode, got: %v", err)
	}
}

// TestUpdatePerformanceMetrics_ErrorTracking tests error tracking
func TestUpdatePerformanceMetrics_ErrorTracking(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Metadata: &AgentMetadata{
			Performance: AgentPerformanceMetrics{},
		},
	}

	// Simulate 3 errors
	agent.UpdatePerformanceMetrics(false, "error 1")
	agent.UpdatePerformanceMetrics(false, "error 2")
	agent.UpdatePerformanceMetrics(false, "error 3")

	agent.Metadata.Mutex.RLock()
	if agent.Metadata.Performance.FailedCalls != 3 {
		t.Errorf("FailedCalls should be 3, got %d", agent.Metadata.Performance.FailedCalls)
	}
	if agent.Metadata.Performance.ConsecutiveErrors != 3 {
		t.Errorf("ConsecutiveErrors should be 3, got %d", agent.Metadata.Performance.ConsecutiveErrors)
	}
	agent.Metadata.Mutex.RUnlock()

	// Simulate recovery with success
	agent.UpdatePerformanceMetrics(true, "")

	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	if agent.Metadata.Performance.ConsecutiveErrors != 0 {
		t.Errorf("ConsecutiveErrors should reset to 0 on success, got %d",
			agent.Metadata.Performance.ConsecutiveErrors)
	}
	if agent.Metadata.Performance.SuccessfulCalls != 1 {
		t.Errorf("SuccessfulCalls should be 1, got %d", agent.Metadata.Performance.SuccessfulCalls)
	}
}

// TestUpdatePerformanceMetrics_NilMetadata tests graceful handling of nil Metadata
func TestUpdatePerformanceMetrics_NilMetadata(t *testing.T) {
	agent := &Agent{
		ID:       "test-agent",
		Name:     "Test Agent",
		Metadata: nil,
	}

	// Should not panic
	agent.UpdatePerformanceMetrics(false, "error message")

	// Metadata should still be nil
	if agent.Metadata != nil {
		t.Error("Metadata should remain nil after update with nil metadata")
	}
}
