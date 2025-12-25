// Package common provides core types, constants, and utilities shared across the system.
// This package consolidates foundational types from types.go, config_types.go, and agent_types.go
// to establish a clear base layer with zero dependencies on other core packages.
package common

import (
	"context"
	"sync"
	"time"
)

// ============================================================================
// CORE DOMAIN TYPES
// ============================================================================

// Task represents a task to be executed by an agent
type Task struct {
	ID          string
	Description string
	Agent       *Agent
	Expected    string
}

// Message represents a message in the conversation
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// ToolCall represents a tool call made by the agent
type ToolCall struct {
	ID        string
	ToolName  string
	Arguments map[string]interface{}
}

// AgentResponse represents a response from an agent
type AgentResponse struct {
	AgentID   string
	AgentName string
	Content   string
	ToolCalls []ToolCall
}

// CrewResponse represents the final response from the crew
type CrewResponse struct {
	AgentID       string
	AgentName     string
	Content       string
	ToolCalls     []ToolCall
	IsTerminal    bool
	PausedAgentID string // Agent ID that paused, used for resume functionality
}

// Crew represents a group of agents working together
type Crew struct {
	Agents                  []*Agent
	Tasks                   []*Task
	MaxRounds               int
	MaxHandoffs             int
	ParallelAgentTimeout    time.Duration  // Timeout for parallel agent execution (default: 60s)
	MaxToolOutputChars      int            // Max characters per tool output (default: 2000)
	MaxTotalToolOutputChars int            // Max TOTAL characters for all tools combined (default: 4000)
	Routing                 *RoutingConfig // Routing configuration from crew.yaml
}

// StreamEvent represents a streaming event sent to the client
type StreamEvent struct {
	Type      string      `json:"type"`      // "agent_start", "agent_response", "tool_start", "tool_result", "pause", "error"
	Agent     string      `json:"agent"`     // Agent ID/Name
	Content   string      `json:"content"`   // Main message
	Timestamp time.Time   `json:"timestamp"` // When this happened
	Metadata  interface{} `json:"metadata"`  // Extra data (tool results, etc.)
}

// ============================================================================
// CONFIGURATION TYPES
// ============================================================================

// RoutingSignal defines a signal that can be emitted by an agent
type RoutingSignal struct {
	Signal      string `yaml:"signal"`
	Target      string `yaml:"target"`
	Description string `yaml:"description"`
}

// AgentBehavior defines how an agent behaves in routing
type AgentBehavior struct {
	WaitForSignal bool   `yaml:"wait_for_signal"`
	AutoRoute     bool   `yaml:"auto_route"`
	IsTerminal    bool   `yaml:"is_terminal"`
	Description   string `yaml:"description"`
}

// ParallelGroupConfig defines a group of agents that should be executed in parallel
type ParallelGroupConfig struct {
	Agents         []string `yaml:"agents"`
	WaitForAll     bool     `yaml:"wait_for_all"`
	TimeoutSeconds int      `yaml:"timeout_seconds"`
	NextAgent      string   `yaml:"next_agent"`
	Description    string   `yaml:"description"`
}

// RoutingConfig defines routing rules for the crew
type RoutingConfig struct {
	Signals        map[string][]RoutingSignal     `yaml:"signals"`
	Defaults       map[string]string              `yaml:"defaults"`
	AgentBehaviors map[string]AgentBehavior       `yaml:"agent_behaviors"`
	ParallelGroups map[string]ParallelGroupConfig `yaml:"parallel_groups"`
}

// CrewConfig represents the crew configuration
type CrewConfig struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	EntryPoint  string `yaml:"entry_point"`

	Agents []string `yaml:"agents"`

	Settings struct {
		// Configuration Mode (Permissive vs Strict)
		ConfigMode string `yaml:"config_mode"` // "permissive" (default) or "strict"

		MaxHandoffs                int `yaml:"max_handoffs" required:"strict"`
		MaxRounds                  int `yaml:"max_rounds" required:"strict"`
		TimeoutSeconds             int `yaml:"timeout_seconds" required:"strict"`
		Language                   string `yaml:"language"`
		Organization               string `yaml:"organization"`

		// Configurable timeouts and output limits
		ParallelTimeoutSeconds     int `yaml:"parallel_timeout_seconds" required:"strict"`       // Timeout for parallel execution
		MaxToolOutputChars         int `yaml:"max_tool_output_chars" required:"strict"`          // Per-tool output limit
		MaxTotalToolOutputChars    int `yaml:"max_total_tool_output_chars" required:"strict"`    // Total limit for all tools

		// Extended configuration for all remaining hardcoded values
		ToolExecutionTimeoutSeconds int `yaml:"tool_execution_timeout_seconds" required:"strict"` // Timeout per tool execution
		ToolResultTimeoutSeconds    int `yaml:"tool_result_timeout_seconds" required:"strict"`    // Timeout for tool result processing
		MinToolTimeoutMillis        int `yaml:"min_tool_timeout_millis" required:"strict"`        // Min tool timeout
		StreamChunkTimeoutMillis    int `yaml:"stream_chunk_timeout_millis" required:"strict"`    // Stream chunk timeout
		SSEKeepAliveSeconds         int `yaml:"sse_keep_alive_seconds" required:"strict"`         // SSE keep-alive
		RequestStoreCleanupMinutes  int `yaml:"request_store_cleanup_minutes" required:"strict"`  // Cleanup interval

		// Retry and backoff
		RetryBackoffMinMillis int `yaml:"retry_backoff_min_millis" required:"strict"`       // Initial backoff
		RetryBackoffMaxSeconds int `yaml:"retry_backoff_max_seconds" required:"strict"`     // Max backoff

		// Input validation limits
		MaxInputSizeKB       int `yaml:"max_input_size_kb" required:"strict"`              // Max input size
		MinAgentIDLength     int `yaml:"min_agent_id_length" required:"strict"`            // Min agent ID length
		MaxAgentIDLength     int `yaml:"max_agent_id_length" required:"strict"`            // Max agent ID length
		MaxRequestBodySizeKB int `yaml:"max_request_body_size_kb" required:"strict"`       // Max request body

		// Output and storage
		StreamBufferSize    int `yaml:"stream_buffer_size" required:"strict"`             // Stream buffer size
		MaxStoredRequests   int `yaml:"max_stored_requests" required:"strict"`            // Max stored requests

		// Client cache
		ClientCacheTTLMinutes int `yaml:"client_cache_ttl_minutes" required:"strict"`       // Client cache TTL

		// Graceful shutdown
		GracefulShutdownCheckMillis int `yaml:"graceful_shutdown_check_millis" required:"strict"` // Shutdown check interval
		TimeoutWarningThresholdPct  int `yaml:"timeout_warning_threshold_pct" required:"strict"`  // Timeout warning %

		// Cost Control Parameters
		MaxTokensPerCall   int     `yaml:"max_tokens_per_call" required:"strict"`       // Max tokens per request
		MaxTokensPerDay    int     `yaml:"max_tokens_per_day" required:"strict"`        // Max tokens per 24h
		MaxCostPerDay      float64 `yaml:"max_cost_per_day" required:"strict"`          // Max cost per day in USD
		CostAlertThreshold float64 `yaml:"cost_alert_threshold" required:"strict"`      // Alert at % of budget

		// Memory Management Parameters
		MaxMemoryMB         int     `yaml:"max_memory_mb" required:"strict"`              // Max memory per request in MB
		MaxDailyMemoryGB    int     `yaml:"max_daily_memory_gb" required:"strict"`        // Max total memory per 24h in GB
		MemoryAlertPercent  float64 `yaml:"memory_alert_percent" required:"strict"`       // Alert when exceeds this %
		MaxContextWindow    int     `yaml:"max_context_window" required:"strict"`         // Max context window size in tokens
		ContextTrimPercent  float64 `yaml:"context_trim_percent" required:"strict"`       // Trim % when full
		SlowCallThresholdSec int    `yaml:"slow_call_threshold_seconds" required:"strict"`// Alert if exceeds this duration

		// Performance & Reliability Parameters
		MaxErrorsPerHour     int `yaml:"max_errors_per_hour" required:"strict"`          // Max errors per hour
		MaxErrorsPerDay      int `yaml:"max_errors_per_day" required:"strict"`           // Max errors per day
		MaxConsecutiveErrors int `yaml:"max_consecutive_errors" required:"strict"`       // Max consecutive errors

		// Rate Limiting & Quotas
		MaxCallsPerMinute  int  `yaml:"max_calls_per_minute" required:"strict"`           // Rate limit: calls per minute
		MaxCallsPerHour    int  `yaml:"max_calls_per_hour" required:"strict"`             // Rate limit: calls per hour
		MaxCallsPerDay     int  `yaml:"max_calls_per_day" required:"strict"`              // Rate limit: calls per 24 hours
		BlockOnQuotaExceed bool `yaml:"block_on_quota_exceed" required:"strict"`          // BLOCK request when quota exceeded
	} `yaml:"settings"`

	Routing *RoutingConfig `yaml:"routing"`
}

// ModelConfigYAML represents YAML configuration for a model (for parsing)
type ModelConfigYAML struct {
	Model       string `yaml:"model" required:"strict"`              // LLM model name
	Provider    string `yaml:"provider" required:"strict"`           // LLM provider name
	ProviderURL string `yaml:"provider_url"`                         // Optional provider URL
}

// CostLimitsConfig defines cost control quota limits
type CostLimitsConfig struct {
	MaxTokensPerCall               int     `yaml:"max_tokens_per_call"`             // Max tokens per API call
	MaxTokensPerDay                int     `yaml:"max_tokens_per_day"`              // Max tokens per 24 hours
	MaxCostPerDayUSD               float64 `yaml:"max_cost_per_day_usd"`            // Max USD cost per day
	AlertThreshold                 float64 `yaml:"alert_threshold"`                 // Warn at % of limit
	BlockOnCostExceed              bool    `yaml:"block_on_cost_exceed"`            // true=BLOCK request, false=WARN only
	InputTokenPricePerMillion      float64 `yaml:"input_token_price_per_million"`   // Cost per 1M input tokens
	OutputTokenPricePerMillion     float64 `yaml:"output_token_price_per_million"`  // Cost per 1M output tokens
}

// MemoryLimitsConfig defines memory quota limits
type MemoryLimitsConfig struct {
	MaxPerCallMB        int  `yaml:"max_per_call_mb"`        // Max MB per execution
	MaxPerDayMB         int  `yaml:"max_per_day_mb"`         // Max MB per 24 hours
	BlockOnMemoryExceed bool `yaml:"block_on_memory_exceed"` // true=BLOCK request, false=WARN only
}

// ErrorLimitsConfig defines error rate quota limits
type ErrorLimitsConfig struct {
	MaxConsecutive     int  `yaml:"max_consecutive"`        // Max consecutive failures
	MaxPerDay          int  `yaml:"max_per_day"`            // Max errors per 24 hours
	BlockOnErrorExceed bool `yaml:"block_on_error_exceed"`  // true=BLOCK request, false=WARN only
}

// LoggingConfig defines observability and monitoring settings
type LoggingConfig struct {
	EnableMemoryMetrics     bool   `yaml:"enable_memory_metrics"`      // Log memory usage per call
	EnablePerformanceMetrics bool  `yaml:"enable_performance_metrics"` // Log response metrics
	EnableQuotaWarnings     bool   `yaml:"enable_quota_warnings"`      // Log quota threshold alerts
	LogLevel                string `yaml:"log_level"`                  // debug/info/warn/error
}

// AgentConfig represents an agent configuration
type AgentConfig struct {
	ID             string           `yaml:"id" required:"strict"`
	Name           string           `yaml:"name" required:"strict"`
	Description    string           `yaml:"description"`
	Role           string           `yaml:"role" required:"strict"`
	Backstory      string           `yaml:"backstory"`
	Model          string           `yaml:"model"`         // Deprecated: Use Primary instead
	Temperature    float64          `yaml:"temperature"`
	IsTerminal     bool             `yaml:"is_terminal"`
	Tools          []string         `yaml:"tools"`
	HandoffTargets []string         `yaml:"handoff_targets"`
	SystemPrompt   string           `yaml:"system_prompt"`
	Provider       string           `yaml:"provider"`      // Deprecated: Use Primary.Provider instead
	ProviderURL    string           `yaml:"provider_url"`  // Deprecated: Use Primary.ProviderURL instead
	Primary        *ModelConfigYAML `yaml:"primary" required:"strict"`       // Primary LLM provider
	Backup         *ModelConfigYAML `yaml:"backup"`        // Fallback LLM provider

	// Nested quota and monitoring configurations
	CostLimits   *CostLimitsConfig   `yaml:"cost_limits"`      // Token usage and cost limits
	MemoryLimits *MemoryLimitsConfig `yaml:"memory_limits"`    // Memory usage limits
	ErrorLimits  *ErrorLimitsConfig  `yaml:"error_limits"`     // Error rate limits
	Logging      *LoggingConfig      `yaml:"logging"`          // Observability settings

	// Backward compatibility: Keep old flat fields for existing configurations
	MaxTokensPerCall   int     `yaml:"max_tokens_per_call"`   // DEPRECATED
	MaxTokensPerDay    int     `yaml:"max_tokens_per_day"`    // DEPRECATED
	MaxCostPerDay      float64 `yaml:"max_cost_per_day"`      // DEPRECATED
	CostAlertThreshold float64 `yaml:"cost_alert_threshold"`  // DEPRECATED
	EnforceCostLimits  bool    `yaml:"enforce_cost_limits"`   // DEPRECATED
}

// ============================================================================
// AGENT TYPES
// ============================================================================

// ModelConfig represents configuration for a specific LLM model and provider
type ModelConfig struct {
	Model       string // LLM model name (e.g., "gpt-4o", "deepseek-r1:32b")
	Provider    string // LLM provider (e.g., "openai", "ollama")
	ProviderURL string // Provider-specific URL
}

// AgentCostMetrics tracks cost and token usage for an agent
type AgentCostMetrics struct {
	CallCount     int          // Number of calls this period
	TotalTokens   int          // Total tokens used
	DailyCost     float64      // Total cost today
	LastResetTime time.Time    // When daily counter resets
	Mutex         sync.RWMutex // Protect concurrent access
}

// AgentMemoryMetrics tracks memory usage and quotas for an agent
type AgentMemoryMetrics struct {
	// Memory Usage (Runtime)
	CurrentMemoryMB    int     // Current memory usage in MB
	PeakMemoryMB       int     // Peak memory usage ever recorded
	AverageMemoryMB    int     // Average memory usage across calls
	MemoryTrendPercent float64 // Trend: positive = increasing, negative = decreasing

	// Memory Quota & Thresholds
	MaxMemoryMB        int     // Max allowed memory per call
	MaxDailyMemoryGB   int     // Max total memory per day
	MemoryAlertPercent float64 // Alert when exceeds this %

	// Context Window & History
	CurrentContextSize int     // Current conversation history in tokens
	MaxContextWindow   int     // Max context window
	ContextTrimPercent float64 // How much to trim when full

	// Call Metrics
	AverageCallDuration time.Duration // Average execution time per call
	SlowCallThreshold   time.Duration // Alert if exceeds this

	Mutex sync.RWMutex // Protect concurrent access
}

// AgentPerformanceMetrics tracks execution performance and reliability
type AgentPerformanceMetrics struct {
	// Response Quality
	SuccessfulCalls     int           // Number of successful executions
	FailedCalls         int           // Number of failed executions
	SuccessRate         float64       // Success rate percentage (0-100)
	AverageResponseTime time.Duration // Average time per call

	// Error Tracking
	LastError         string    // Last error message
	LastErrorTime     time.Time // When last error occurred
	ConsecutiveErrors int       // Number of consecutive errors
	ErrorCountToday   int       // Errors in last 24 hours

	// Performance Thresholds
	MaxErrorsPerHour     int // Alert if exceeds
	MaxErrorsPerDay      int // Block if exceeds
	MaxConsecutiveErrors int // Block after N consecutive

	Mutex sync.RWMutex // Protect concurrent access
}

// AgentQuotaLimits defines all quota constraints for an agent
type AgentQuotaLimits struct {
	// Cost Quotas
	MaxTokensPerCall int     // Max tokens per request
	MaxTokensPerDay  int     // Max tokens per 24h
	MaxCostPerDay    float64 // Max cost per 24h in USD
	CostAlertPercent float64 // Alert at % of budget

	// Memory Quotas
	MaxMemoryPerCall int // Max MB per request
	MaxMemoryPerDay  int // Max MB per 24h
	MaxContextWindow int // Max context tokens

	// Execution Quotas
	MaxCallsPerMinute int // Rate limiting: calls/minute
	MaxCallsPerHour   int // Rate limiting: calls/hour
	MaxCallsPerDay    int // Rate limiting: calls/day
	MaxErrorsPerHour  int // Error rate: errors/hour
	MaxErrorsPerDay   int // Error rate: errors/day

	// Enforcement
	BlockOnQuotaExceed bool // true=BLOCK request, false=WARN only
}

// AgentMetadata stores agent-specific metadata
type AgentMetadata struct {
	ID                  string
	Name                string
	Role                string
	CreatedAt           time.Time
	CostMetrics         *AgentCostMetrics
	MemoryMetrics       *AgentMemoryMetrics
	PerformanceMetrics  *AgentPerformanceMetrics
	QuotaLimits         *AgentQuotaLimits
	AllowedToolNames    []string
	HandoffTargetAgents []*Agent
}

// Agent represents an AI agent in the crew
type Agent struct {
	ID                 string
	Name               string
	Description        string
	Role               string
	Backstory          string
	Temperature        float64
	IsTerminal         bool
	Tools              []interface{}
	HandoffTargets     []*Agent
	SystemPrompt       string
	SystemPromptCache  string
	IsSystemPromptDirty bool
	PrimaryModel       *ModelConfig
	BackupModel        *ModelConfig
	Metadata           *AgentMetadata
	Quota              *AgentQuotaLimits
	CostMetrics        *AgentCostMetrics
	MemoryMetrics      *AgentMemoryMetrics
	PerformanceMetrics *AgentPerformanceMetrics
	ctx                context.Context
	Mutex              sync.RWMutex
}

// HardcodedDefaults stores hardcoded default configuration values
// These are fallback defaults when values are not specified in YAML
type HardcodedDefaults struct {
	// Timeouts (all in seconds)
	DefaultTimeoutSeconds           int
	DefaultToolExecutionTimeout     int
	DefaultParallelTimeout          int
	DefaultStreamChunkTimeout       int
	DefaultToolResultTimeout        int

	// Limits (characters, KB, MB, GB)
	DefaultMaxToolOutputChars    int
	DefaultMaxTotalToolOutputChars int
	DefaultMaxInputSizeKB        int
	DefaultMaxRequestBodySizeKB  int

	// ID constraints
	DefaultMinAgentIDLength int
	DefaultMaxAgentIDLength int

	// Token limits
	DefaultMaxTokensPerCall int
	DefaultMaxTokensPerDay  int

	// Cost limits
	DefaultMaxCostPerDay float64
	DefaultCostAlertPercent float64

	// Memory limits (MB)
	DefaultMaxMemoryPerCall int
	DefaultMaxMemoryPerDay  int

	// Memory limits (GB)
	DefaultMaxDailyMemoryGB int

	// Context window
	DefaultMaxContextWindow int
	DefaultContextTrimPercent float64

	// Rate limiting
	DefaultMaxCallsPerMinute int
	DefaultMaxCallsPerHour   int
	DefaultMaxCallsPerDay    int

	// Error handling
	DefaultMaxErrorsPerHour     int
	DefaultMaxErrorsPerDay      int
	DefaultMaxConsecutiveErrors int

	// Performance thresholds
	DefaultSlowCallThreshold int

	// Token pricing (per 1M tokens, in USD)
	DefaultInputTokenPrice  float64
	DefaultOutputTokenPrice float64

	// Other defaults
	DefaultParallelTimeoutSeconds int
	DefaultMaxHandoffs            int
	DefaultMaxRounds              int
	DefaultLanguage               string
	DefaultOrganization           string
	DefaultStreamBufferSize       int
	DefaultMaxStoredRequests      int
	DefaultClientCacheTTL         int
}
