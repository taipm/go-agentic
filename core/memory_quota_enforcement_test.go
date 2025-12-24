package crewai

import (
	"testing"
)

// TestCheckMemoryQuotaEnforcement_BlockMode tests blocking on memory overflow
func TestCheckMemoryQuotaEnforcement_BlockMode(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Metadata: &AgentMetadata{
			Quotas: AgentQuotaLimits{
				MaxMemoryPerCall:  512, // 512 MB limit
				BlockOnQuotaExceed: true, // Block mode
			},
			Memory: AgentMemoryMetrics{
				CurrentMemoryMB: 600, // 600 MB - exceeds limit
			},
			Performance: AgentPerformanceMetrics{
				MaxConsecutiveErrors: 5,
			},
		},
	}

	err := agent.CheckMemoryQuota()
	if err == nil {
		t.Error("CheckMemoryQuota should return error when memory exceeds limit in block mode")
	}
	if err.Error() != "agent 'test-agent': memory quota exceeded - used 600 MB, max 512 MB per call" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestCheckMemoryQuotaEnforcement_WarnMode tests warning in memory overflow
func TestCheckMemoryQuotaEnforcement_WarnMode(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Metadata: &AgentMetadata{
			Quotas: AgentQuotaLimits{
				MaxMemoryPerCall:  512,
				BlockOnQuotaExceed: false, // Warn mode
			},
			Memory: AgentMemoryMetrics{
				CurrentMemoryMB: 600, // Exceeds limit but warn-only
			},
		},
	}

	err := agent.CheckMemoryQuota()
	if err != nil {
		t.Errorf("CheckMemoryQuota should not return error in warn mode, got: %v", err)
	}
}

// TestUpdateMemoryMetrics_TrackingAccuracy tests metric updates
func TestUpdateMemoryMetrics_TrackingAccuracy(t *testing.T) {
	agent := &Agent{
		ID:   "test-agent",
		Name: "Test Agent",
		Metadata: &AgentMetadata{
			Memory: AgentMemoryMetrics{},
			Cost:   AgentCostMetrics{CallCount: 1},
		},
	}

	// First call
	agent.UpdateMemoryMetrics(256, 1500) // 256 MB, 1500ms

	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	if agent.Metadata.Memory.CurrentMemoryMB != 256 {
		t.Errorf("CurrentMemory should be 256, got %d", agent.Metadata.Memory.CurrentMemoryMB)
	}
	if agent.Metadata.Memory.PeakMemoryMB != 256 {
		t.Errorf("PeakMemory should be 256, got %d", agent.Metadata.Memory.PeakMemoryMB)
	}
}

// TestUpdateMemoryMetrics_NilMetadata tests graceful handling of nil Metadata
func TestUpdateMemoryMetrics_NilMetadata(t *testing.T) {
	agent := &Agent{
		ID:       "test-agent",
		Name:     "Test Agent",
		Metadata: nil,
	}

	// Should not panic
	agent.UpdateMemoryMetrics(256, 1500)

	// Metadata should still be nil
	if agent.Metadata != nil {
		t.Error("Metadata should remain nil after update with nil metadata")
	}
}
