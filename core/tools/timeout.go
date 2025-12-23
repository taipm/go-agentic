package tools

import (
	"sync"
	"time"
)

// ExecutionMetrics tracks execution time and status for tools
type ExecutionMetrics struct {
	ToolName  string        // Name of the tool executed
	Duration  time.Duration // Time taken to execute
	Status    string        // "success", "timeout", "error"
	TimedOut  bool          // True if tool execution exceeded timeout
	StartTime time.Time     // When tool execution started
	EndTime   time.Time     // When tool execution completed
}

// TimeoutTracker tracks sequence execution time and manages per-tool budgets
type TimeoutTracker struct {
	sequenceStartTime time.Time     // When sequence started
	sequenceDeadline  time.Time     // When sequence must complete
	overheadBudget    time.Duration // Estimated overhead per tool
	usedTime          time.Duration // Time already consumed in sequence
	mu                sync.Mutex    // Protect concurrent access
}

// NewTimeoutTracker creates a timeout tracker for a sequence
func NewTimeoutTracker(sequenceTimeout time.Duration, overheadBudget time.Duration) *TimeoutTracker {
	now := time.Now()
	return &TimeoutTracker{
		sequenceStartTime: now,
		sequenceDeadline:  now.Add(sequenceTimeout),
		overheadBudget:    overheadBudget,
		usedTime:          0,
	}
}

// GetRemainingTime returns how much time is left in the sequence
func (tt *TimeoutTracker) GetRemainingTime() time.Duration {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	remaining := time.Until(tt.sequenceDeadline)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CalculateToolTimeout calculates the appropriate timeout for the next tool
// accounting for: per-tool timeout, remaining sequence time, and overhead budget
func (tt *TimeoutTracker) CalculateToolTimeout(defaultTimeout, perToolTimeout time.Duration) time.Duration {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	// Start with per-tool timeout, fallback to default
	toolTimeout := perToolTimeout
	if toolTimeout <= 0 {
		toolTimeout = defaultTimeout
	}

	// Get remaining sequence time and subtract overhead budget
	remaining := time.Until(tt.sequenceDeadline)
	if remaining <= tt.overheadBudget {
		// Not enough time even for overhead
		return 100 * time.Millisecond // Minimal timeout to signal urgency
	}

	// Available time is remaining minus overhead
	availableTime := remaining - tt.overheadBudget

	// Use the minimum: per-tool timeout or available time
	if toolTimeout > availableTime {
		return availableTime
	}
	return toolTimeout
}

// RecordToolExecution records that a tool has finished and updates used time
func (tt *TimeoutTracker) RecordToolExecution(duration time.Duration) {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	tt.usedTime += duration
}

// IsTimeoutWarning returns true if we're within 20% of sequence deadline
func (tt *TimeoutTracker) IsTimeoutWarning() bool {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	remaining := time.Until(tt.sequenceDeadline)
	totalDuration := tt.sequenceDeadline.Sub(tt.sequenceStartTime)
	warnThreshold := totalDuration / 5 // 20%
	return remaining < warnThreshold && remaining > 0
}

// GetUsedTime returns total time used so far
func (tt *TimeoutTracker) GetUsedTime() time.Duration {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	return tt.usedTime
}

// TimeoutConfig defines timeout behavior for tools
type TimeoutConfig struct {
	DefaultToolTimeout time.Duration            // Default timeout per tool (e.g., 5s)
	SequenceTimeout    time.Duration            // Max total time for all tools in sequence
	PerToolTimeout     map[string]time.Duration // Per-tool overrides for specific tools
	OverheadBudget     time.Duration            // Estimated overhead per tool call
	CollectMetrics     bool                     // If true, collect execution metrics
	ExecutionMetrics   []ExecutionMetrics       // Collected metrics from last execution
}

// NewTimeoutConfig creates a timeout config with recommended defaults
func NewTimeoutConfig() *TimeoutConfig {
	return &TimeoutConfig{
		DefaultToolTimeout: 5 * time.Second,
		SequenceTimeout:    30 * time.Second,
		OverheadBudget:     500 * time.Millisecond,
		PerToolTimeout:     make(map[string]time.Duration),
		CollectMetrics:     true,
		ExecutionMetrics:   []ExecutionMetrics{},
	}
}

// GetToolTimeout gets the timeout for a specific tool (checks per-tool overrides first)
func (tc *TimeoutConfig) GetToolTimeout(toolName string) time.Duration {
	if timeout, exists := tc.PerToolTimeout[toolName]; exists {
		return timeout
	}
	return tc.DefaultToolTimeout
}

// SetToolTimeout sets a custom timeout for a specific tool
func (tc *TimeoutConfig) SetToolTimeout(toolName string, timeout time.Duration) {
	if tc.PerToolTimeout == nil {
		tc.PerToolTimeout = make(map[string]time.Duration)
	}
	tc.PerToolTimeout[toolName] = timeout
}

// ResetMetrics clears collected execution metrics
func (tc *TimeoutConfig) ResetMetrics() {
	tc.ExecutionMetrics = []ExecutionMetrics{}
}

// AddMetric adds an execution metric to the collection
func (tc *TimeoutConfig) AddMetric(metric ExecutionMetrics) {
	if tc.CollectMetrics {
		tc.ExecutionMetrics = append(tc.ExecutionMetrics, metric)
	}
}
