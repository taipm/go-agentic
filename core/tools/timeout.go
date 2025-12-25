package tools

import (
	"time"

	"github.com/taipm/go-agentic/core/internal"
)

// Type aliases for backwards compatibility
// Deprecated: Use internal.ExecutionMetrics instead
type ExecutionMetrics = internal.ExecutionMetrics

// Deprecated: Use internal.TimeoutTracker instead
type TimeoutTracker = internal.TimeoutTracker

// Deprecated: Use internal.TimeoutConfig instead
type TimeoutConfig = internal.TimeoutConfig

// Deprecated: Use internal.NewTimeoutTracker instead
func NewTimeoutTracker(sequenceTimeout time.Duration, overheadBudget time.Duration) *TimeoutTracker {
	return internal.NewTimeoutTracker(sequenceTimeout, overheadBudget)
}

// Deprecated: Use internal.NewTimeoutConfig instead
func NewTimeoutConfig() *TimeoutConfig {
	return internal.NewTimeoutConfig()
}
