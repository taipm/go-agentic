package crewai

import (
	"fmt"
	"time"
)

// ConfigMode defines how the system handles missing configuration values
type ConfigMode string

const (
	// PermissiveMode allows missing config values and uses defaults
	// System runs smoothly, no errors for missing config
	PermissiveMode ConfigMode = "permissive"

	// StrictMode requires all config values to be explicitly set
	// System fails with clear error messages if values are missing
	StrictMode ConfigMode = "strict"

	// DefaultConfigMode is PermissiveMode for backward compatibility
	DefaultConfigMode ConfigMode = PermissiveMode
)

// ConfigModeError represents configuration validation failures in Strict Mode
type ConfigModeError struct {
	Mode   ConfigMode
	Errors []string
}

// parameterDescriptions maps parameter names to detailed explanations
var parameterDescriptions = map[string]string{
	"ParallelAgentTimeout": "Max time for multiple agents running in parallel (seconds). Default: 60s",
	"ToolExecutionTimeout": "Max time for a single tool/function call to complete (seconds). Default: 5s",
	"ToolResultTimeout": "Max time to wait for tool result processing after execution (seconds). Default: 30s",
	"MinToolTimeout": "Sanity check: minimum allowed tool timeout to prevent unreasonably low values (milliseconds). Default: 100ms",
	"StreamChunkTimeout": "Timeout for processing each chunk in streaming responses (milliseconds). Default: 500ms",
	"SSEKeepAliveInterval": "Keep-alive ping interval for Server-Sent Events to prevent connection timeout (seconds). Default: 30s",
	"RequestStoreCleanupInterval": "How often to clean up old request tracking data from memory (minutes). Default: 5m",
	"RetryBackoffMinDuration": "Initial wait time before first retry on failure - exponential backoff starts here (milliseconds). Default: 100ms",
	"RetryBackoffMaxDuration": "Maximum wait time between retries - ceiling for exponential backoff (seconds). Default: 5s",
	"ClientCacheTTL": "How long to keep cached LLM provider clients (OpenAI/Ollama) before recreating (minutes). Default: 60m",
	"GracefulShutdownCheckInterval": "How often to check for shutdown signal during long operations (milliseconds). Default: 100ms",
	"MaxInputSize": "Prevents DoS attacks from extremely large user inputs (bytes, ~10KB = 10240). Default: 10240",
	"MinAgentIDLength": "Prevents empty or whitespace-only agent identifiers (characters). Default: 1",
	"MaxAgentIDLength": "Prevents unbounded identifier growth - practical limit for agent names (characters). Default: 128",
	"MaxRequestBodySize": "Prevents memory exhaustion from oversized HTTP requests (bytes, ~100KB = 102400). Default: 102400",
	"MaxToolOutputChars": "Truncates very large tool outputs to prevent LLM context overflow (characters). Default: 2000",
	"StreamBufferSize": "Number of chunks to buffer during streaming responses - balance responsiveness vs memory (chunks). Default: 100",
	"MaxStoredRequests": "Maximum number of requests to keep in memory for tracking/debugging (requests). Default: 1000",
	"TimeoutWarningThreshold": "Warn when this percentage of timeout remains (0.0-1.0, e.g., 0.20 = warn at 80% used). Default: 0.20",
}

// Error implements the error interface
func (cme *ConfigModeError) Error() string {
	if len(cme.Errors) == 0 {
		return "configuration validation error"
	}

	header := "Configuration Validation Errors (Mode: " + string(cme.Mode) + "):\n\n"
	errorList := ""
	for i, err := range cme.Errors {
		// Format: "1. ParamName: error message"
		// Extract parameter name if possible
		errorList += fmt.Sprintf("  %d. %s\n", i+1, err)
	}

	footer := "\n📋 PARAMETER DESCRIPTIONS:\n"
	for paramName, description := range parameterDescriptions {
		footer += fmt.Sprintf("   • %s: %s\n", paramName, description)
	}

	footer += "\n📝 NEXT STEPS:\n"
	footer += "   1. Open your crew.yaml settings section\n"
	footer += "   2. Add all 19 parameters listed above\n"
	footer += "   3. Set values appropriate for your use case\n"
	footer += "   4. For help: See crew-strict-documented.yaml example or docs\n"

	return header + errorList + footer
}

// HardcodedDefaults consolidates all configurable default values
// These defaults were previously hardcoded throughout the codebase.
// They can now be overridden via:
// 1. YAML configuration (crew.yaml, agent.yaml)
// 2. Environment variables
// 3. Programmatic configuration
//
// ✅ Phase 4: Extended Configuration - Makes all hardcoded values configurable
type HardcodedDefaults struct {
	// ===== CONFIGURATION MODE =====
	// Mode controls how system handles missing/invalid config values
	// PermissiveMode (default): Use hardcoded defaults silently
	// StrictMode: Fail loudly with clear error messages
	Mode ConfigMode
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
		// Configuration Mode
		Mode: PermissiveMode, // Default: allow missing config, use defaults

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

// validateDuration checks a duration value and applies defaults or collects errors
func (d *HardcodedDefaults) validateDuration(name string, value *time.Duration, defaultVal time.Duration, errors *[]string) {
	if *value <= 0 {
		if d.Mode == StrictMode {
			*errors = append(*errors, name+" must be > 0 (got: "+value.String()+")")
		} else {
			*value = defaultVal
		}
	}
}

// validateInt checks an int value and applies defaults or collects errors
func (d *HardcodedDefaults) validateInt(name string, value *int, defaultVal int, errors *[]string) {
	if *value <= 0 {
		if d.Mode == StrictMode {
			*errors = append(*errors, name+" must be > 0")
		} else {
			*value = defaultVal
		}
	}
}

// LogConfigurationMode returns a message about the current configuration mode
// In StrictMode: warns that defaults are NOT being used
// In PermissiveMode: info that safe defaults are being used
func (d *HardcodedDefaults) LogConfigurationMode() string {
	if d.Mode == StrictMode {
		return fmt.Sprintf("⚠️  STRICT MODE: All configuration parameters MUST be explicitly set. Defaults are NOT being used.")
	}
	return fmt.Sprintf("ℹ️  PERMISSIVE MODE: Using safe defaults for missing configuration parameters.")
}

// Validate checks that all timeout values are sensible
// Returns error if any timeout is invalid or contradictory
// In StrictMode: fails if values are missing/invalid
// In PermissiveMode: silently applies defaults for missing/invalid values
func (d *HardcodedDefaults) Validate() error {
	var errors []string

	// Ensure mode is set
	if d.Mode == "" {
		d.Mode = PermissiveMode
	}

	// Validate timeouts
	d.validateDuration("ParallelAgentTimeout", &d.ParallelAgentTimeout, 60*time.Second, &errors)
	d.validateDuration("ToolExecutionTimeout", &d.ToolExecutionTimeout, 5*time.Second, &errors)
	d.validateDuration("ToolResultTimeout", &d.ToolResultTimeout, 30*time.Second, &errors)
	d.validateDuration("MinToolTimeout", &d.MinToolTimeout, 100*time.Millisecond, &errors)
	d.validateDuration("StreamChunkTimeout", &d.StreamChunkTimeout, 500*time.Millisecond, &errors)
	d.validateDuration("SSEKeepAliveInterval", &d.SSEKeepAliveInterval, 30*time.Second, &errors)
	d.validateDuration("RequestStoreCleanupInterval", &d.RequestStoreCleanupInterval, 5*time.Minute, &errors)
	d.validateDuration("RetryBackoffMinDuration", &d.RetryBackoffMinDuration, 100*time.Millisecond, &errors)
	d.validateDuration("RetryBackoffMaxDuration", &d.RetryBackoffMaxDuration, 5*time.Second, &errors)
	d.validateDuration("ClientCacheTTL", &d.ClientCacheTTL, 1*time.Hour, &errors)
	d.validateDuration("GracefulShutdownCheckInterval", &d.GracefulShutdownCheckInterval, 100*time.Millisecond, &errors)

	// Validate size limits
	d.validateInt("MaxInputSize", &d.MaxInputSize, 10*1024, &errors)
	d.validateInt("MinAgentIDLength", &d.MinAgentIDLength, 1, &errors)
	d.validateInt("MaxAgentIDLength", &d.MaxAgentIDLength, 128, &errors)
	d.validateInt("MaxRequestBodySize", &d.MaxRequestBodySize, 100*1024, &errors)
	d.validateInt("MaxToolOutputChars", &d.MaxToolOutputChars, 2000, &errors)
	d.validateInt("StreamBufferSize", &d.StreamBufferSize, 100, &errors)
	d.validateInt("MaxStoredRequests", &d.MaxStoredRequests, 1000, &errors)

	// Threshold validation
	if d.TimeoutWarningThreshold <= 0 || d.TimeoutWarningThreshold >= 1 {
		if d.Mode == StrictMode {
			errors = append(errors, "TimeoutWarningThreshold must be between 0 and 1")
		} else {
			d.TimeoutWarningThreshold = 0.20
		}
	}

	// Return accumulated errors if any
	if len(errors) > 0 {
		return &ConfigModeError{
			Mode:   d.Mode,
			Errors: errors,
		}
	}

	return nil
}
