package crewai

import (
	"fmt"
	"time"
)

// UpdateMemoryMetrics updates memory usage metrics after a call
// [METRIC|MEMORY|RUNTIME] Tracks current, peak, average memory with trend calculation
func (a *Agent) UpdateMemoryMetrics(memoryUsedMB int, callDurationMs int64) {
	if a.Metadata == nil {
		return
	}

	a.Metadata.Mutex.Lock()
	defer a.Metadata.Mutex.Unlock()

	// Update current memory
	a.Metadata.Memory.CurrentMemoryMB = memoryUsedMB

	// Update peak memory
	if memoryUsedMB > a.Metadata.Memory.PeakMemoryMB {
		a.Metadata.Memory.PeakMemoryMB = memoryUsedMB
	}

	// Update average memory (rolling average)
	if a.Metadata.Cost.CallCount > 0 {
		// Average = (previous_avg * call_count + current_usage) / (call_count + 1)
		callCount := int(a.Metadata.Cost.CallCount)
		newAvg := (a.Metadata.Memory.AverageMemoryMB*callCount + memoryUsedMB) / (callCount + 1)
		a.Metadata.Memory.AverageMemoryMB = newAvg

		// Calculate trend
		if callCount > 1 {
			oldAvg := a.Metadata.Memory.AverageMemoryMB
			if oldAvg > 0 {
				a.Metadata.Memory.MemoryTrendPercent = float64((memoryUsedMB - oldAvg)) / float64(oldAvg) * 100
			}
		}
	}

	// Update call duration
	callDuration := time.Duration(callDurationMs) * time.Millisecond
	if a.Metadata.Performance.AverageResponseTime == 0 {
		a.Metadata.Performance.AverageResponseTime = callDuration
	} else {
		// Rolling average for response time
		callCount := int(a.Metadata.Cost.CallCount)
		oldAvg := a.Metadata.Performance.AverageResponseTime
		newAvg := time.Duration((int64(oldAvg)*int64(callCount) + callDurationMs) / int64(callCount+1))
		a.Metadata.Performance.AverageResponseTime = newAvg
	}
}

// UpdatePerformanceMetrics updates performance metrics after a call
// [METRIC|PERFORMANCE|RUNTIME] Tracks success rate, error patterns, and response times
func (a *Agent) UpdatePerformanceMetrics(success bool, errorMsg string) {
	if a.Metadata == nil {
		return
	}

	a.Metadata.Mutex.Lock()
	defer a.Metadata.Mutex.Unlock()

	if success {
		a.Metadata.Performance.SuccessfulCalls++
		a.Metadata.Performance.ConsecutiveErrors = 0 // Reset consecutive errors

		// Recalculate success rate
		total := a.Metadata.Performance.SuccessfulCalls + a.Metadata.Performance.FailedCalls
		if total > 0 {
			a.Metadata.Performance.SuccessRate = float64(a.Metadata.Performance.SuccessfulCalls) / float64(total) * 100
		}
	} else {
		a.Metadata.Performance.FailedCalls++
		a.Metadata.Performance.ConsecutiveErrors++
		a.Metadata.Performance.ErrorCountToday++
		a.Metadata.Performance.LastError = errorMsg
		a.Metadata.Performance.LastErrorTime = time.Now()

		// Recalculate success rate
		total := a.Metadata.Performance.SuccessfulCalls + a.Metadata.Performance.FailedCalls
		if total > 0 {
			a.Metadata.Performance.SuccessRate = float64(a.Metadata.Performance.SuccessfulCalls) / float64(total) * 100
		}
	}
}

// CheckMemoryQuota checks if memory usage exceeds quota
// [QUOTA|MEMORY|ENFORCEMENT] Validates memory usage against per-call and daily limits
func (a *Agent) CheckMemoryQuota() error {
	if a.Metadata == nil {
		return nil
	}

	a.Metadata.Mutex.RLock()
	defer a.Metadata.Mutex.RUnlock()

	// Check per-call memory limit
	if a.Metadata.Memory.CurrentMemoryMB > a.Metadata.Quotas.MaxMemoryPerCall {
		if a.Metadata.Quotas.EnforceQuotas {
			return fmt.Errorf(
				"agent '%s': memory quota exceeded - used %d MB, max %d MB per call",
				a.ID, a.Metadata.Memory.CurrentMemoryMB, a.Metadata.Quotas.MaxMemoryPerCall)
		}
		fmt.Printf("[MEMORY ALERT] Agent '%s': Used %d MB, exceeds limit of %d MB\n",
			a.Name, a.Metadata.Memory.CurrentMemoryMB, a.Metadata.Quotas.MaxMemoryPerCall)
	}

	// Note: Daily memory tracking would require tracking reset time
	// Implementation in future phase

	return nil
}

// CheckErrorQuota checks if error rate exceeds quota
// [QUOTA|ERROR|ENFORCEMENT] Validates error rate against consecutive and daily limits
func (a *Agent) CheckErrorQuota() error {
	if a.Metadata == nil {
		return nil
	}

	a.Metadata.Mutex.RLock()
	defer a.Metadata.Mutex.RUnlock()

	// Check consecutive errors
	if a.Metadata.Performance.ConsecutiveErrors >= a.Metadata.Performance.MaxConsecutiveErrors {
		if a.Metadata.Quotas.EnforceQuotas {
			return fmt.Errorf(
				"agent '%s': maximum consecutive errors exceeded - %d errors, max %d allowed",
				a.ID, a.Metadata.Performance.ConsecutiveErrors, a.Metadata.Performance.MaxConsecutiveErrors)
		}
		fmt.Printf("[ERROR ALERT] Agent '%s': %d consecutive errors (max: %d)\n",
			a.Name, a.Metadata.Performance.ConsecutiveErrors, a.Metadata.Performance.MaxConsecutiveErrors)
	}

	// Check daily error limit
	if a.Metadata.Performance.ErrorCountToday >= a.Metadata.Performance.MaxErrorsPerDay {
		if a.Metadata.Quotas.EnforceQuotas {
			return fmt.Errorf(
				"agent '%s': daily error limit exceeded - %d errors, max %d allowed",
				a.ID, a.Metadata.Performance.ErrorCountToday, a.Metadata.Performance.MaxErrorsPerDay)
		}
		fmt.Printf("[ERROR ALERT] Agent '%s': %d errors today (daily max: %d)\n",
			a.Name, a.Metadata.Performance.ErrorCountToday, a.Metadata.Performance.MaxErrorsPerDay)
	}

	return nil
}

// CheckSlowCall checks if call took longer than threshold
// [THRESHOLD|PERFORMANCE] Alerts when execution exceeds slow call threshold
func (a *Agent) CheckSlowCall(duration time.Duration) {
	if a.Metadata == nil {
		return
	}

	a.Metadata.Mutex.RLock()
	defer a.Metadata.Mutex.RUnlock()

	if duration > a.Metadata.Memory.SlowCallThreshold {
		fmt.Printf("[SLOW CALL] Agent '%s': took %.2fs (threshold: %.2fs)\n",
			a.Name, duration.Seconds(), a.Metadata.Memory.SlowCallThreshold.Seconds())
	}
}

// ResetDailyPerformanceMetrics resets daily counters at midnight
// [METRIC|PERFORMANCE|RUNTIME] Resets daily error and quota counters
func (a *Agent) ResetDailyPerformanceMetrics() {
	if a.Metadata == nil {
		return
	}

	a.Metadata.Mutex.Lock()
	defer a.Metadata.Mutex.Unlock()

	a.Metadata.Performance.ErrorCountToday = 0
	// Note: LastError and LastErrorTime are not reset (historical tracking)
}

// GetMemoryStatus returns current memory status summary
// [METRIC|MEMORY] Returns memory metrics: current, peak, average, and trend
func (a *Agent) GetMemoryStatus() (current, peak, average int, trend float64) {
	if a.Metadata == nil {
		return 0, 0, 0, 0
	}

	a.Metadata.Mutex.RLock()
	defer a.Metadata.Mutex.RUnlock()

	return a.Metadata.Memory.CurrentMemoryMB,
		a.Metadata.Memory.PeakMemoryMB,
		a.Metadata.Memory.AverageMemoryMB,
		a.Metadata.Memory.MemoryTrendPercent
}

// GetPerformanceStatus returns current performance status summary
// [METRIC|PERFORMANCE] Returns performance metrics: success rate, call counts, error count
func (a *Agent) GetPerformanceStatus() (successRate float64, successCount, failCount int, errorToday int) {
	if a.Metadata == nil {
		return 100, 0, 0, 0
	}

	a.Metadata.Mutex.RLock()
	defer a.Metadata.Mutex.RUnlock()

	return a.Metadata.Performance.SuccessRate,
		a.Metadata.Performance.SuccessfulCalls,
		a.Metadata.Performance.FailedCalls,
		a.Metadata.Performance.ErrorCountToday
}
