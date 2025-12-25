package crewai

import (
	"fmt"
	"log"
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
	// ===== PHASE 4: Core Configuration =====
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
	"MaxToolOutputChars": "Truncates per-tool output to prevent LLM context overflow (characters). Default: 2000",
	"MaxTotalToolOutputChars": "Maximum TOTAL characters for all tools combined - summarizes excess tools (characters). Default: 4000",
	"StreamBufferSize": "Number of chunks to buffer during streaming responses - balance responsiveness vs memory (chunks). Default: 100",
	"MaxStoredRequests": "Maximum number of requests to keep in memory for tracking/debugging (requests). Default: 1000",
	"TimeoutWarningThreshold": "Warn when this percentage of timeout remains (0.0-1.0, e.g., 0.20 = warn at 80% used). Default: 0.20",

	// ===== WEEK 1: Cost Control =====
	"MaxTokensPerCall": "Maximum tokens per single request (e.g., 1000). Default: 4000",
	"MaxTokensPerDay": "Maximum tokens per 24-hour period (e.g., 50000). Default: 100000",
	"MaxCostPerDay": "Maximum daily budget in USD (e.g., 10.00). Default: 50.00",
	"CostAlertThreshold": "Warn when this percentage of daily budget is used (0.0-1.0, e.g., 0.80 = warn at 80%). Default: 0.80",

	// ===== WEEK 2: Memory Management =====
	"MaxMemoryMB": "Maximum memory per request in MB. Default: 512",
	"MaxDailyMemoryGB": "Maximum total memory per 24-hour period in GB. Default: 10",
	"MemoryAlertPercent": "Alert when memory usage exceeds this percentage (0.0-100.0). Default: 80.0",
	"MaxContextWindow": "Maximum context window size in tokens (e.g., 32000 for gpt-4). Default: 32000",
	"ContextTrimPercent": "Percentage of context to trim when window is full (0.0-100.0). Default: 20.0",
	"SlowCallThreshold": "Alert if call exceeds this duration (seconds). Default: 30",

	// ===== WEEK 2: Performance & Reliability =====
	"MaxErrorsPerHour": "Maximum errors per hour before alerting. Default: 10",
	"MaxErrorsPerDay": "Maximum errors per day before blocking. Default: 50",
	"MaxConsecutiveErrors": "Maximum consecutive errors before blocking. Default: 5",

	// ===== WEEK 2: Rate Limiting & Quotas =====
	"MaxCallsPerMinute": "Rate limit: maximum calls per minute. Default: 60",
	"MaxCallsPerHour": "Rate limit: maximum calls per hour. Default: 1000",
	"MaxCallsPerDay": "Rate limit: maximum calls per 24 hours. Default: 10000",
	"BlockOnQuotaExceed": "Block requests when quota exceeded (true=BLOCK, false=WARN only). Default: true",
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
	footer += "   2. Add all required parameters (Phase 4 core + WEEK 1 cost control + WEEK 2 quotas)\n"
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
	// MaxToolOutputChars: Maximum characters per tool output before truncation
	// Previously hardcoded in: types.go:88 (2000 characters)
	// Default: 2,000 characters (balance context vs detail)
	MaxToolOutputChars int

	// MaxTotalToolOutputChars: Maximum TOTAL characters for all tool outputs combined
	// ✅ FIX for MEDIUM Issue: Tool results not summarized - prevents unbounded token usage
	// When total exceeds this, later tools get summarized to save context
	// Default: 4,000 characters (~1,000 tokens)
	MaxTotalToolOutputChars int

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

	// ===== WEEK 1: COST CONTROL PARAMETERS =====
	// MaxTokensPerCall: Maximum tokens per single request
	// Previously hardcoded in: agent.go (no hardcode, but needs default)
	// Default: 4000 tokens per request
	MaxTokensPerCall int

	// MaxTokensPerDay: Maximum tokens per 24-hour period
	// Previously hardcoded in: agent.go (no hardcode, but needs default)
	// Default: 100000 tokens per day
	MaxTokensPerDay int

	// MaxCostPerDay: Maximum daily budget in USD
	// Previously hardcoded in: agent.go (no hardcode, but needs default)
	// Default: 50.00 USD per day
	MaxCostPerDay float64

	// CostAlertThreshold: Warn when this percentage of daily budget is used
	// Previously hardcoded in: agent.go (no hardcode, but needs default)
	// Default: 0.80 (warn at 80% usage)
	CostAlertThreshold float64

	// ===== WEEK 2: MEMORY MANAGEMENT PARAMETERS =====
	// MaxMemoryMB: Maximum memory per request in MB
	// Previously hardcoded in: types.go (512 MB mentioned in comment)
	// Default: 512 MB per request
	MaxMemoryMB int

	// MaxDailyMemoryGB: Maximum total memory per 24-hour period in GB
	// Previously hardcoded in: types.go (10 GB mentioned in comment)
	// Default: 10 GB per day
	MaxDailyMemoryGB int

	// MemoryAlertPercent: Alert when memory usage exceeds this percentage
	// Previously hardcoded in: types.go (80% mentioned in comment)
	// Default: 80.0% alert threshold
	MemoryAlertPercent float64

	// MaxContextWindow: Maximum context window size in tokens
	// Previously hardcoded in: types.go (32000 for gpt-4 mentioned in comment)
	// Default: 32000 tokens (typical for gpt-4)
	MaxContextWindow int

	// ContextTrimPercent: Percentage of context to trim when window is full
	// Previously hardcoded in: types.go (20% mentioned in comment)
	// Default: 20.0% trim
	ContextTrimPercent float64

	// SlowCallThreshold: Alert if call exceeds this duration
	// Previously hardcoded in: types.go (30s mentioned in comment)
	// Default: 30 seconds
	SlowCallThreshold time.Duration

	// ===== WEEK 2: PERFORMANCE & RELIABILITY PARAMETERS =====
	// MaxErrorsPerHour: Maximum errors per hour before alerting
	// Previously hardcoded in: types.go (10 mentioned in comment)
	// Default: 10 errors per hour
	MaxErrorsPerHour int

	// MaxErrorsPerDay: Maximum errors per day before blocking
	// Previously hardcoded in: types.go (50 mentioned in comment)
	// Default: 50 errors per day
	MaxErrorsPerDay int

	// MaxConsecutiveErrors: Maximum consecutive errors before blocking
	// Previously hardcoded in: types.go (5 mentioned in comment)
	// Default: 5 consecutive errors
	MaxConsecutiveErrors int

	// ===== WEEK 2: RATE LIMITING & QUOTA PARAMETERS =====
	// MaxCallsPerMinute: Rate limit - maximum calls per minute
	// Default: 60 calls per minute
	MaxCallsPerMinute int

	// MaxCallsPerHour: Rate limit - maximum calls per hour
	// Default: 1000 calls per hour
	MaxCallsPerHour int

	// MaxCallsPerDay: Rate limit - maximum calls per 24 hours
	// Default: 10000 calls per day
	MaxCallsPerDay int

	// BlockOnQuotaExceed: Block requests when quota exceeded (true=block, false=warn only)
	// Previously hardcoded in: types.go (true mentioned in comment)
	// Default: true (block on exceed - production-safe)
	BlockOnQuotaExceed bool
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
		MaxToolOutputChars:      2000,   // ✅ FIX #5: Default per-tool output limit
		MaxTotalToolOutputChars: 4000,   // ✅ FIX: Total limit for all tools combined (~1K tokens)
		StreamBufferSize:        100,

		// Request Storage
		MaxStoredRequests: 1000,

		// Client Cache
		ClientCacheTTL: 1 * time.Hour,

		// Graceful Shutdown
		GracefulShutdownCheckInterval: 100 * time.Millisecond,
		TimeoutWarningThreshold:       0.20, // 20%

		// Cost Control (WEEK 1)
		MaxTokensPerCall:   4000,    // 4K tokens per request
		MaxTokensPerDay:    100000,  // 100K tokens per day
		MaxCostPerDay:      50.0,    // $50/day budget
		CostAlertThreshold: 0.80,    // Alert at 80% usage

		// Memory Management (WEEK 2)
		MaxMemoryMB:        512,     // 512 MB per request
		MaxDailyMemoryGB:   10,      // 10 GB per day
		MemoryAlertPercent: 80.0,    // Alert at 80%
		MaxContextWindow:   32000,   // 32K tokens (gpt-4)
		ContextTrimPercent: 20.0,    // Trim 20% when full
		SlowCallThreshold:  30 * time.Second,

		// Performance & Reliability (WEEK 2)
		MaxErrorsPerHour:     10,  // 10 errors/hour alert
		MaxErrorsPerDay:      50,  // 50 errors/day block
		MaxConsecutiveErrors: 5,   // 5 consecutive block

		// Rate Limiting (WEEK 2)
		MaxCallsPerMinute:  60,    // 60 calls/minute
		MaxCallsPerHour:    1000,  // 1000 calls/hour
		MaxCallsPerDay:     10000, // 10000 calls/day
		BlockOnQuotaExceed: true,  // ✅ Default: BLOCK mode (production-safe)
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

// validateFloat checks a float value in range [0, 1] and applies defaults or collects errors
func (d *HardcodedDefaults) validateFloatRange(name string, value *float64, defaultVal float64, minVal, maxVal float64, errors *[]string) {
	if *value < minVal || *value > maxVal {
		if d.Mode == StrictMode {
			*errors = append(*errors, fmt.Sprintf("%s must be between %.2f and %.2f", name, minVal, maxVal))
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

// validatePhase4Timeouts validates Phase 4 timeout parameters
func (d *HardcodedDefaults) validatePhase4Timeouts(errors *[]string) {
	d.validateDuration("ParallelAgentTimeout", &d.ParallelAgentTimeout, 60*time.Second, errors)
	d.validateDuration("ToolExecutionTimeout", &d.ToolExecutionTimeout, 5*time.Second, errors)
	d.validateDuration("ToolResultTimeout", &d.ToolResultTimeout, 30*time.Second, errors)
	d.validateDuration("MinToolTimeout", &d.MinToolTimeout, 100*time.Millisecond, errors)
	d.validateDuration("StreamChunkTimeout", &d.StreamChunkTimeout, 500*time.Millisecond, errors)
	d.validateDuration("SSEKeepAliveInterval", &d.SSEKeepAliveInterval, 30*time.Second, errors)
	d.validateDuration("RequestStoreCleanupInterval", &d.RequestStoreCleanupInterval, 5*time.Minute, errors)
	d.validateDuration("RetryBackoffMinDuration", &d.RetryBackoffMinDuration, 100*time.Millisecond, errors)
	d.validateDuration("RetryBackoffMaxDuration", &d.RetryBackoffMaxDuration, 5*time.Second, errors)
	d.validateDuration("ClientCacheTTL", &d.ClientCacheTTL, 1*time.Hour, errors)
	d.validateDuration("GracefulShutdownCheckInterval", &d.GracefulShutdownCheckInterval, 100*time.Millisecond, errors)
}

// validatePhase4SizeLimits validates Phase 4 size limit parameters
func (d *HardcodedDefaults) validatePhase4SizeLimits(errors *[]string) {
	d.validateInt("MaxInputSize", &d.MaxInputSize, 10*1024, errors)
	d.validateInt("MinAgentIDLength", &d.MinAgentIDLength, 1, errors)
	d.validateInt("MaxAgentIDLength", &d.MaxAgentIDLength, 128, errors)
	d.validateInt("MaxRequestBodySize", &d.MaxRequestBodySize, 100*1024, errors)
	d.validateInt("MaxToolOutputChars", &d.MaxToolOutputChars, 2000, errors)
	d.validateInt("MaxTotalToolOutputChars", &d.MaxTotalToolOutputChars, 4000, errors)
	d.validateInt("StreamBufferSize", &d.StreamBufferSize, 100, errors)
	d.validateInt("MaxStoredRequests", &d.MaxStoredRequests, 1000, errors)
	d.validateFloatRange("TimeoutWarningThreshold", &d.TimeoutWarningThreshold, 0.20, 0.0, 1.0, errors)
}

// validateWeek1CostControl validates Week 1 cost control parameters
func (d *HardcodedDefaults) validateWeek1CostControl(errors *[]string) {
	d.validateInt("MaxTokensPerCall", &d.MaxTokensPerCall, 4000, errors)
	d.validateInt("MaxTokensPerDay", &d.MaxTokensPerDay, 100000, errors)
	if d.MaxCostPerDay < 0 {
		if d.Mode == StrictMode {
			*errors = append(*errors, "MaxCostPerDay must be >= 0")
		} else {
			d.MaxCostPerDay = 50.0
		}
	}
	d.validateFloatRange("CostAlertThreshold", &d.CostAlertThreshold, 0.80, 0.0, 1.0, errors)
}

// validateWeek2Memory validates Week 2 memory management parameters
func (d *HardcodedDefaults) validateWeek2Memory(errors *[]string) {
	d.validateInt("MaxMemoryMB", &d.MaxMemoryMB, 512, errors)
	d.validateInt("MaxDailyMemoryGB", &d.MaxDailyMemoryGB, 10, errors)
	d.validateFloatRange("MemoryAlertPercent", &d.MemoryAlertPercent, 80.0, 0.0, 100.0, errors)
	d.validateInt("MaxContextWindow", &d.MaxContextWindow, 32000, errors)
	d.validateFloatRange("ContextTrimPercent", &d.ContextTrimPercent, 20.0, 0.0, 100.0, errors)
	d.validateDuration("SlowCallThreshold", &d.SlowCallThreshold, 30*time.Second, errors)
}

// validateWeek2Performance validates Week 2 performance & reliability parameters
func (d *HardcodedDefaults) validateWeek2Performance(errors *[]string) {
	d.validateInt("MaxErrorsPerHour", &d.MaxErrorsPerHour, 10, errors)
	d.validateInt("MaxErrorsPerDay", &d.MaxErrorsPerDay, 50, errors)
	d.validateInt("MaxConsecutiveErrors", &d.MaxConsecutiveErrors, 5, errors)
}

// validateWeek2RateLimiting validates Week 2 rate limiting parameters
func (d *HardcodedDefaults) validateWeek2RateLimiting(errors *[]string) {
	d.validateInt("MaxCallsPerMinute", &d.MaxCallsPerMinute, 60, errors)
	d.validateInt("MaxCallsPerHour", &d.MaxCallsPerHour, 1000, errors)
	d.validateInt("MaxCallsPerDay", &d.MaxCallsPerDay, 10000, errors)
}

// Validate checks that all configuration values are sensible
// Returns error if any value is invalid
// In StrictMode: fails if values are missing/invalid
// In PermissiveMode: silently applies defaults for missing/invalid values
func (d *HardcodedDefaults) Validate() error {
	var errors []string

	// Ensure mode is set
	if d.Mode == "" {
		d.Mode = PermissiveMode
	}

	// Validate Phase 4 (core) parameters
	d.validatePhase4Timeouts(&errors)
	d.validatePhase4SizeLimits(&errors)

	// Validate Week 1 (cost control) parameters
	d.validateWeek1CostControl(&errors)

	// Validate Week 2 (memory, performance, rate limiting) parameters
	d.validateWeek2Memory(&errors)
	d.validateWeek2Performance(&errors)
	d.validateWeek2RateLimiting(&errors)

	// Return accumulated errors if any
	if len(errors) > 0 {
		return &ConfigModeError{
			Mode:   d.Mode,
			Errors: errors,
		}
	}

	return nil
}

// ConfigToHardcodedDefaults converts CrewConfig settings to HardcodedDefaults struct
// ✅ Phase 4: Extended Configuration - Maps YAML values to runtime defaults
// Returns defaults with YAML overrides applied; validation is performed after conversion
// ✅ Phase 5.1: In STRICT MODE, missing values are NOT defaulted and remain 0, causing validation to fail
func ConfigToHardcodedDefaults(config *CrewConfig) *HardcodedDefaults {
	// In PERMISSIVE MODE: Start with all defaults
	// In STRICT MODE: Start with all 0 values (except mode), require explicit YAML config
	var defaults *HardcodedDefaults

	// ✅ Phase 5.1: Check config mode FIRST
	configMode := PermissiveMode
	if config.Settings.ConfigMode != "" {
		configMode = ConfigMode(config.Settings.ConfigMode)
	}

	// In STRICT MODE, don't use defaults - start with empty values
	if configMode == StrictMode {
		defaults = &HardcodedDefaults{
			Mode: StrictMode,
			// All timeout/int fields default to 0
			// All duration fields default to 0 (0 seconds)
			// All float fields default to 0
			// Validation will catch these as errors
		}
	} else {
		// In PERMISSIVE MODE, start with defaults
		defaults = DefaultHardcodedDefaults()
		defaults.Mode = PermissiveMode
	}

	// Phase 1 configurations
	if config.Settings.ParallelTimeoutSeconds > 0 {
		defaults.ParallelAgentTimeout = time.Duration(config.Settings.ParallelTimeoutSeconds) * time.Second
	}
	if config.Settings.MaxToolOutputChars > 0 {
		defaults.MaxToolOutputChars = config.Settings.MaxToolOutputChars
	}
	if config.Settings.MaxTotalToolOutputChars > 0 {
		defaults.MaxTotalToolOutputChars = config.Settings.MaxTotalToolOutputChars
	}

	// Phase 4 timeout configurations
	if config.Settings.ToolExecutionTimeoutSeconds > 0 {
		defaults.ToolExecutionTimeout = time.Duration(config.Settings.ToolExecutionTimeoutSeconds) * time.Second
	}
	if config.Settings.ToolResultTimeoutSeconds > 0 {
		defaults.ToolResultTimeout = time.Duration(config.Settings.ToolResultTimeoutSeconds) * time.Second
	}
	if config.Settings.MinToolTimeoutMillis > 0 {
		defaults.MinToolTimeout = time.Duration(config.Settings.MinToolTimeoutMillis) * time.Millisecond
	}
	if config.Settings.StreamChunkTimeoutMillis > 0 {
		defaults.StreamChunkTimeout = time.Duration(config.Settings.StreamChunkTimeoutMillis) * time.Millisecond
	}
	if config.Settings.SSEKeepAliveSeconds > 0 {
		defaults.SSEKeepAliveInterval = time.Duration(config.Settings.SSEKeepAliveSeconds) * time.Second
	}
	if config.Settings.RequestStoreCleanupMinutes > 0 {
		defaults.RequestStoreCleanupInterval = time.Duration(config.Settings.RequestStoreCleanupMinutes) * time.Minute
	}

	// Phase 4 retry and backoff configurations
	if config.Settings.RetryBackoffMinMillis > 0 {
		defaults.RetryBackoffMinDuration = time.Duration(config.Settings.RetryBackoffMinMillis) * time.Millisecond
	}
	if config.Settings.RetryBackoffMaxSeconds > 0 {
		defaults.RetryBackoffMaxDuration = time.Duration(config.Settings.RetryBackoffMaxSeconds) * time.Second
	}

	// Phase 4 input validation limits
	if config.Settings.MaxInputSizeKB > 0 {
		defaults.MaxInputSize = config.Settings.MaxInputSizeKB * 1024
	}
	if config.Settings.MinAgentIDLength > 0 {
		defaults.MinAgentIDLength = config.Settings.MinAgentIDLength
	}
	if config.Settings.MaxAgentIDLength > 0 {
		defaults.MaxAgentIDLength = config.Settings.MaxAgentIDLength
	}
	if config.Settings.MaxRequestBodySizeKB > 0 {
		defaults.MaxRequestBodySize = config.Settings.MaxRequestBodySizeKB * 1024
	}

	// Phase 4 output and storage
	if config.Settings.StreamBufferSize > 0 {
		defaults.StreamBufferSize = config.Settings.StreamBufferSize
	}
	if config.Settings.MaxStoredRequests > 0 {
		defaults.MaxStoredRequests = config.Settings.MaxStoredRequests
	}

	// Phase 4 client cache
	if config.Settings.ClientCacheTTLMinutes > 0 {
		defaults.ClientCacheTTL = time.Duration(config.Settings.ClientCacheTTLMinutes) * time.Minute
	}

	// Phase 4 graceful shutdown
	if config.Settings.GracefulShutdownCheckMillis > 0 {
		defaults.GracefulShutdownCheckInterval = time.Duration(config.Settings.GracefulShutdownCheckMillis) * time.Millisecond
	}
	if config.Settings.TimeoutWarningThresholdPct > 0 && config.Settings.TimeoutWarningThresholdPct <= 100 {
		defaults.TimeoutWarningThreshold = float64(config.Settings.TimeoutWarningThresholdPct) / 100.0
	}

	// ✅ WEEK 1: Cost Control configurations
	if config.Settings.MaxTokensPerCall > 0 {
		defaults.MaxTokensPerCall = config.Settings.MaxTokensPerCall
	}
	if config.Settings.MaxTokensPerDay > 0 {
		defaults.MaxTokensPerDay = config.Settings.MaxTokensPerDay
	}
	if config.Settings.MaxCostPerDay > 0 {
		defaults.MaxCostPerDay = config.Settings.MaxCostPerDay
	}
	if config.Settings.CostAlertThreshold > 0 {
		defaults.CostAlertThreshold = config.Settings.CostAlertThreshold
	}

	// ✅ WEEK 2: Memory Management configurations
	if config.Settings.MaxMemoryMB > 0 {
		defaults.MaxMemoryMB = config.Settings.MaxMemoryMB
	}
	if config.Settings.MaxDailyMemoryGB > 0 {
		defaults.MaxDailyMemoryGB = config.Settings.MaxDailyMemoryGB
	}
	if config.Settings.MemoryAlertPercent > 0 {
		defaults.MemoryAlertPercent = config.Settings.MemoryAlertPercent
	}
	if config.Settings.MaxContextWindow > 0 {
		defaults.MaxContextWindow = config.Settings.MaxContextWindow
	}
	if config.Settings.ContextTrimPercent > 0 {
		defaults.ContextTrimPercent = config.Settings.ContextTrimPercent
	}
	if config.Settings.SlowCallThresholdSec > 0 {
		defaults.SlowCallThreshold = time.Duration(config.Settings.SlowCallThresholdSec) * time.Second
	}

	// ✅ WEEK 2: Performance & Reliability configurations
	if config.Settings.MaxErrorsPerHour > 0 {
		defaults.MaxErrorsPerHour = config.Settings.MaxErrorsPerHour
	}
	if config.Settings.MaxErrorsPerDay > 0 {
		defaults.MaxErrorsPerDay = config.Settings.MaxErrorsPerDay
	}
	if config.Settings.MaxConsecutiveErrors > 0 {
		defaults.MaxConsecutiveErrors = config.Settings.MaxConsecutiveErrors
	}

	// ✅ WEEK 2: Rate Limiting & Quotas configurations
	if config.Settings.MaxCallsPerMinute > 0 {
		defaults.MaxCallsPerMinute = config.Settings.MaxCallsPerMinute
	}
	if config.Settings.MaxCallsPerHour > 0 {
		defaults.MaxCallsPerHour = config.Settings.MaxCallsPerHour
	}
	if config.Settings.MaxCallsPerDay > 0 {
		defaults.MaxCallsPerDay = config.Settings.MaxCallsPerDay
	}
	// BlockOnQuotaExceed: true=BLOCK requests when quota exceeded, false=WARN only
	// Default is true (production-safe), only override if explicitly set in YAML
	if config.Settings.BlockOnQuotaExceed {
		defaults.BlockOnQuotaExceed = true
	}

	// Validate all converted values
	if err := defaults.Validate(); err != nil {
		// ✅ Phase 5.1: In STRICT MODE, validation errors are FATAL - no fallback
		if defaults.Mode == StrictMode {
			log.Printf("[CONFIG ERROR] STRICT MODE validation failed: %v", err)
			// Return nil will be caught by caller
			return nil
		}
		// In PERMISSIVE MODE, fallback to defaults
		log.Printf("[CONFIG WARNING] Failed to validate defaults after conversion: %v - using fallback defaults", err)
		return DefaultHardcodedDefaults()
	}

	return defaults
}
