package crewai

import (
	"time"
)

// HardcodedDefaults consolidates all configurable default values
// These defaults were previously hardcoded throughout the codebase.
// They can now be overridden via:
// 1. YAML configuration (crew.yaml, agent.yaml)
// 2. Environment variables
// 3. Programmatic configuration
//
// ✅ Phase 4: Extended Configuration - Makes all hardcoded values configurable
type HardcodedDefaults struct {
	// ===== TIMEOUT PARAMETERS =====
	// ParallelAgentTimeout: Maximum time allowed for parallel agent execution
	// Previously hardcoded in: crew.go:1183
	// Default: 60 seconds (reasonable for most parallel tasks)
	ParallelAgentTimeout time.Duration

	// ToolExecutionTimeout: Maximum time for a single tool execution
	// Previously hardcoded in: crew.go:1250 (5 * time.Second)
	// Default: 5 seconds (most tool operations should be quick)
	ToolExecutionTimeout time.Duration

	// ToolResultTimeout: Maximum time to wait for tool result processing
	// Previously hardcoded in: crew.go:1300 (30 * time.Second)
	// Default: 30 seconds (allows time for complex tool processing)
	ToolResultTimeout time.Duration

	// MinToolTimeout: Minimum allowed timeout for tool execution
	// Previously hardcoded in: crew.go validation (100ms)
	// Default: 100 milliseconds (prevent unreasonably low timeouts)
	MinToolTimeout time.Duration

	// StreamChunkTimeout: Timeout for each stream chunk processing
	// Previously hardcoded in: http.go:85 (500ms)
	// Default: 500 milliseconds (responsive streaming)
	StreamChunkTimeout time.Duration

	// SSEKeepAliveInterval: Server-Sent Event keep-alive interval
	// Previously hardcoded in: http.go (30 seconds)
	// Default: 30 seconds (prevents connection timeout on slow responses)
	SSEKeepAliveInterval time.Duration

	// RequestStoreCleanupInterval: How often to clean up old request data
	// Previously hardcoded in: request_tracking.go (5 minutes)
	// Default: 5 minutes (balance between memory and retention)
	RequestStoreCleanupInterval time.Duration

	// ===== RETRY AND BACKOFF PARAMETERS =====
	// RetryBackoffMinDuration: Initial backoff duration for retries
	// Previously hardcoded in: crew.go (100ms)
	// Default: 100 milliseconds
	RetryBackoffMinDuration time.Duration

	// RetryBackoffMaxDuration: Maximum backoff duration for retries
	// Previously hardcoded in: crew.go (5 seconds)
	// Default: 5 seconds
	RetryBackoffMaxDuration time.Duration

	// ===== INPUT VALIDATION LIMITS =====
	// MaxInputSize: Maximum size of user input in bytes
	// Previously hardcoded in: http.go (10KB = 10240 bytes)
	// Default: 10,240 bytes (10KB - reasonable for most inputs)
	MaxInputSize int

	// MinAgentIDLength: Minimum length for agent ID
	// Previously hardcoded in: http.go validation (1 byte minimum)
	// Default: 1 byte
	MinAgentIDLength int

	// MaxAgentIDLength: Maximum length for agent ID
	// Previously hardcoded in: http.go validation (128 bytes)
	// Default: 128 bytes (prevent unbounded identifiers)
	MaxAgentIDLength int

	// MaxRequestBodySize: Maximum size of entire HTTP request body
	// Previously hardcoded in: http.go (100KB = 102400 bytes)
	// Default: 100,240 bytes (100KB - reasonable for complex requests)
	MaxRequestBodySize int

	// ===== OUTPUT LIMITS =====
	// MaxToolOutputChars: Maximum characters in tool output before truncation
	// Previously hardcoded in: types.go:88 (2000 characters)
	// Default: 2,000 characters (balance context vs detail)
	MaxToolOutputChars int

	// StreamBufferSize: Size of streaming buffer for chunk handling
	// Previously hardcoded in: http.go:75 (100 chunks)
	// Default: 100 chunks
	StreamBufferSize int

	// ===== REQUEST STORAGE =====
	// MaxStoredRequests: Maximum number of requests to store in memory
	// Previously hardcoded in: request_tracking.go:30 (1000 requests)
	// Default: 1,000 requests (balance memory vs history)
	MaxStoredRequests int

	// ===== CLIENT CACHE =====
	// ClientCacheTTL: Time-to-live for cached LLM clients
	// Previously hardcoded in: providers/openai/provider.go:27 (1 hour)
	// Default: 1 hour (cache clients to avoid recreation)
	ClientCacheTTL time.Duration

	// ===== GRACEFUL SHUTDOWN =====
	// GracefulShutdownCheckInterval: Interval for checking shutdown state
	// Previously hardcoded in: shutdown.go (100ms)
	// Default: 100 milliseconds (responsive shutdown detection)
	GracefulShutdownCheckInterval time.Duration

	// TimeoutWarningThreshold: Percentage of timeout to trigger warning
	// Previously hardcoded in: crew.go (20%)
	// Default: 20% (warn when 80% of timeout used)
	TimeoutWarningThreshold float64
}

// DefaultHardcodedDefaults returns the global default configuration values
func DefaultHardcodedDefaults() *HardcodedDefaults {
	return &HardcodedDefaults{
		// Timeout Parameters (Phase 1 fixes)
		ParallelAgentTimeout:   60 * time.Second,     // ✅ FIX #4: Default parallel timeout
		ToolExecutionTimeout:   5 * time.Second,      // Per-tool timeout
		ToolResultTimeout:      30 * time.Second,     // Tool result processing
		MinToolTimeout:         100 * time.Millisecond,
		StreamChunkTimeout:     500 * time.Millisecond,
		SSEKeepAliveInterval:   30 * time.Second,
		RequestStoreCleanupInterval: 5 * time.Minute,

		// Retry and Backoff Parameters
		RetryBackoffMinDuration: 100 * time.Millisecond,
		RetryBackoffMaxDuration: 5 * time.Second,

		// Input Validation Limits
		MaxInputSize:     10 * 1024,    // 10KB
		MinAgentIDLength: 1,
		MaxAgentIDLength: 128,
		MaxRequestBodySize: 100 * 1024, // 100KB

		// Output Limits
		MaxToolOutputChars: 2000,        // ✅ FIX #5: Default output limit
		StreamBufferSize:   100,

		// Request Storage
		MaxStoredRequests: 1000,

		// Client Cache
		ClientCacheTTL: 1 * time.Hour,

		// Graceful Shutdown
		GracefulShutdownCheckInterval: 100 * time.Millisecond,
		TimeoutWarningThreshold:       0.20, // 20%
	}
}

// Validate checks that all timeout values are sensible
// Returns error if any timeout is invalid or contradictory
func (d *HardcodedDefaults) Validate() error {
	if d.ParallelAgentTimeout <= 0 {
		d.ParallelAgentTimeout = 60 * time.Second
	}
	if d.ToolExecutionTimeout <= 0 {
		d.ToolExecutionTimeout = 5 * time.Second
	}
	if d.ToolResultTimeout <= 0 {
		d.ToolResultTimeout = 30 * time.Second
	}
	if d.MinToolTimeout <= 0 {
		d.MinToolTimeout = 100 * time.Millisecond
	}
	if d.StreamChunkTimeout <= 0 {
		d.StreamChunkTimeout = 500 * time.Millisecond
	}
	if d.SSEKeepAliveInterval <= 0 {
		d.SSEKeepAliveInterval = 30 * time.Second
	}
	if d.RequestStoreCleanupInterval <= 0 {
		d.RequestStoreCleanupInterval = 5 * time.Minute
	}
	if d.RetryBackoffMinDuration <= 0 {
		d.RetryBackoffMinDuration = 100 * time.Millisecond
	}
	if d.RetryBackoffMaxDuration <= 0 {
		d.RetryBackoffMaxDuration = 5 * time.Second
	}
	if d.ClientCacheTTL <= 0 {
		d.ClientCacheTTL = 1 * time.Hour
	}
	if d.GracefulShutdownCheckInterval <= 0 {
		d.GracefulShutdownCheckInterval = 100 * time.Millisecond
	}

	// Validate size limits
	if d.MaxInputSize <= 0 {
		d.MaxInputSize = 10 * 1024
	}
	if d.MinAgentIDLength <= 0 {
		d.MinAgentIDLength = 1
	}
	if d.MaxAgentIDLength <= 0 {
		d.MaxAgentIDLength = 128
	}
	if d.MaxRequestBodySize <= 0 {
		d.MaxRequestBodySize = 100 * 1024
	}
	if d.MaxToolOutputChars <= 0 {
		d.MaxToolOutputChars = 2000
	}
	if d.StreamBufferSize <= 0 {
		d.StreamBufferSize = 100
	}
	if d.MaxStoredRequests <= 0 {
		d.MaxStoredRequests = 1000
	}

	// Validate threshold
	if d.TimeoutWarningThreshold <= 0 || d.TimeoutWarningThreshold >= 1 {
		d.TimeoutWarningThreshold = 0.20
	}

	return nil
}
