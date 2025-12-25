// Package statemanagement provides state tracking for workflow execution.
package statemanagement

import (
	"sync"
	"time"
)

// RoundMetric tracks metrics for a single execution round.
type RoundMetric struct {
	AgentID      string
	Duration     time.Duration
	StartTime    time.Time
	EndTime      time.Time
	Success      bool
	ToolCalls    int
	ErrorCount   int
}

// ExecutionMetrics contains aggregated metrics for entire execution.
type ExecutionMetrics struct {
	TotalDuration    time.Duration
	RoundCount       int
	AverageDuration  time.Duration
	MaxDuration      time.Duration
	MinDuration      time.Duration
	TotalToolCalls   int
	TotalErrors      int
	SuccessfulRounds int
}

// ExecutionState tracks execution state including timing and metrics.
type ExecutionState struct {
	StartTime      time.Time
	EndTime        time.Time
	RoundCount     int
	HandoffCount   int
	TotalDuration  time.Duration
	RoundMetrics   map[int]*RoundMetric
	lastAgentTime  time.Duration
	mu             sync.RWMutex
}

// NewExecutionState creates a new ExecutionState for tracking execution.
func NewExecutionState() *ExecutionState {
	return &ExecutionState{
		StartTime:    time.Now(),
		RoundMetrics: make(map[int]*RoundMetric),
		RoundCount:   0,
		HandoffCount: 0,
	}
}

// RecordRound records metrics for a single round of execution.
func (es *ExecutionState) RecordRound(agentID string, duration time.Duration, success bool) {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.RoundCount++
	es.lastAgentTime = duration
	es.TotalDuration += duration

	metric := &RoundMetric{
		AgentID:   agentID,
		Duration:  duration,
		StartTime: time.Now().Add(-duration),
		EndTime:   time.Now(),
		Success:   success,
	}

	es.RoundMetrics[es.RoundCount] = metric
}

// RecordHandoff increments the handoff counter.
func (es *ExecutionState) RecordHandoff() {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.HandoffCount++
}

// GetMetrics returns aggregated execution metrics.
func (es *ExecutionState) GetMetrics() *ExecutionMetrics {
	es.mu.RLock()
	defer es.mu.RUnlock()

	if es.RoundCount == 0 {
		return &ExecutionMetrics{
			TotalDuration:    es.TotalDuration,
			RoundCount:       0,
			SuccessfulRounds: 0,
		}
	}

	metrics := &ExecutionMetrics{
		TotalDuration:    es.TotalDuration,
		RoundCount:       es.RoundCount,
		AverageDuration:  es.TotalDuration / time.Duration(es.RoundCount),
		SuccessfulRounds: 0,
		TotalToolCalls:   0,
		TotalErrors:      0,
	}

	// Calculate aggregate metrics from round metrics
	var minDuration, maxDuration time.Duration
	for i := 1; i <= es.RoundCount; i++ {
		if metric, ok := es.RoundMetrics[i]; ok {
			if metric.Success {
				metrics.SuccessfulRounds++
			}
			metrics.TotalToolCalls += metric.ToolCalls
			metrics.TotalErrors += metric.ErrorCount

			if i == 1 {
				minDuration = metric.Duration
				maxDuration = metric.Duration
			} else {
				if metric.Duration < minDuration {
					minDuration = metric.Duration
				}
				if metric.Duration > maxDuration {
					maxDuration = metric.Duration
				}
			}
		}
	}

	metrics.MinDuration = minDuration
	metrics.MaxDuration = maxDuration

	return metrics
}

// GetLastAgentTime returns the duration of the last agent execution.
func (es *ExecutionState) GetLastAgentTime() time.Duration {
	es.mu.RLock()
	defer es.mu.RUnlock()

	return es.lastAgentTime
}

// Finish marks execution as complete and records end time.
func (es *ExecutionState) Finish() {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.EndTime = time.Now()
	es.TotalDuration = es.EndTime.Sub(es.StartTime)
}

// IsRunning returns true if execution is still running.
func (es *ExecutionState) IsRunning() bool {
	es.mu.RLock()
	defer es.mu.RUnlock()

	return es.EndTime.IsZero()
}

// Reset clears all state for a new execution.
func (es *ExecutionState) Reset() {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.StartTime = time.Now()
	es.EndTime = time.Time{}
	es.RoundCount = 0
	es.HandoffCount = 0
	es.TotalDuration = 0
	es.lastAgentTime = 0
	es.RoundMetrics = make(map[int]*RoundMetric)
}

// GetRoundMetric returns metrics for a specific round.
func (es *ExecutionState) GetRoundMetric(roundNumber int) *RoundMetric {
	es.mu.RLock()
	defer es.mu.RUnlock()

	if metric, ok := es.RoundMetrics[roundNumber]; ok {
		return metric
	}
	return nil
}

// Copy creates a copy of the current execution state.
func (es *ExecutionState) Copy() *ExecutionState {
	es.mu.RLock()
	defer es.mu.RUnlock()

	newState := &ExecutionState{
		StartTime:     es.StartTime,
		EndTime:       es.EndTime,
		RoundCount:    es.RoundCount,
		HandoffCount:  es.HandoffCount,
		TotalDuration: es.TotalDuration,
		lastAgentTime: es.lastAgentTime,
		RoundMetrics:  make(map[int]*RoundMetric),
	}

	// Deep copy round metrics
	for k, v := range es.RoundMetrics {
		metricCopy := *v
		newState.RoundMetrics[k] = &metricCopy
	}

	return newState
}
