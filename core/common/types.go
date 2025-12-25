// Package common provides core types, constants, and utilities shared across the system.
// This package consolidates foundational types from types.go, config_types.go, and agent_types.go
// to establish a clear base layer with zero dependencies on other core packages.
package common

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// CORE DOMAIN TYPES
// ============================================================================

// Tool represents an executable tool that can be used by agents
type Tool struct {
	ID          string
	Name        string
	Description string
	Func        interface{} // The actual function to execute
	Input       interface{} // Input schema or parameters
	Parameters  interface{} // Parameters schema (can be JSON schema, map, etc.)
	Output      interface{} // Output schema or return type
}

// ToolTimeoutConfig manages timeout settings for tool execution
type ToolTimeoutConfig struct {
	DefaultToolTimeout time.Duration            // Default timeout for tool execution (default: 5s)
	SequenceTimeout    time.Duration            // Timeout for entire tool execution sequence (default: 30s)
	PerToolTimeout     map[string]time.Duration // Per-tool timeout overrides
	CollectMetrics     bool                     // Whether to collect timeout metrics
}

// NewToolTimeoutConfig creates a new tool timeout configuration with defaults
func NewToolTimeoutConfig() *ToolTimeoutConfig {
	return &ToolTimeoutConfig{
		DefaultToolTimeout: 5 * time.Second,
		SequenceTimeout:    30 * time.Second,
		PerToolTimeout:     make(map[string]time.Duration),
		CollectMetrics:     true,
	}
}

// GetToolTimeout returns the timeout for a specific tool
func (ttc *ToolTimeoutConfig) GetToolTimeout(toolName string) time.Duration {
	if timeout, exists := ttc.PerToolTimeout[toolName]; exists {
		return timeout
	}
	return ttc.DefaultToolTimeout
}

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
	Signals   []string // Signals emitted by agent (e.g., "[END_EXAM]")
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

// RoutingDecision represents the result of routing logic
type RoutingDecision struct {
	NextAgentID string                 // Agent ID to route to
	Reason      string                 // Why this routing decision was made
	IsTerminal  bool                   // Whether execution should terminate
	Metadata    map[string]interface{} // Additional routing context and metadata
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

		MaxHandoffs    int    `yaml:"max_handoffs" required:"strict"`
		MaxRounds      int    `yaml:"max_rounds" required:"strict"`
		TimeoutSeconds int    `yaml:"timeout_seconds" required:"strict"`
		Language       string `yaml:"language"`
		Organization   string `yaml:"organization"`

		// Configurable timeouts and output limits
		ParallelTimeoutSeconds  int `yaml:"parallel_timeout_seconds" required:"strict"`    // Timeout for parallel execution
		MaxToolOutputChars      int `yaml:"max_tool_output_chars" required:"strict"`       // Per-tool output limit
		MaxTotalToolOutputChars int `yaml:"max_total_tool_output_chars" required:"strict"` // Total limit for all tools

		// Extended configuration for all remaining hardcoded values
		ToolExecutionTimeoutSeconds int `yaml:"tool_execution_timeout_seconds" required:"strict"` // Timeout per tool execution
		ToolResultTimeoutSeconds    int `yaml:"tool_result_timeout_seconds" required:"strict"`    // Timeout for tool result processing
		MinToolTimeoutMillis        int `yaml:"min_tool_timeout_millis" required:"strict"`        // Min tool timeout
		StreamChunkTimeoutMillis    int `yaml:"stream_chunk_timeout_millis" required:"strict"`    // Stream chunk timeout
		SSEKeepAliveSeconds         int `yaml:"sse_keep_alive_seconds" required:"strict"`         // SSE keep-alive
		RequestStoreCleanupMinutes  int `yaml:"request_store_cleanup_minutes" required:"strict"`  // Cleanup interval

		// Retry and backoff
		RetryBackoffMinMillis  int `yaml:"retry_backoff_min_millis" required:"strict"`  // Initial backoff
		RetryBackoffMaxSeconds int `yaml:"retry_backoff_max_seconds" required:"strict"` // Max backoff

		// Input validation limits
		MaxInputSizeKB       int `yaml:"max_input_size_kb" required:"strict"`        // Max input size
		MinAgentIDLength     int `yaml:"min_agent_id_length" required:"strict"`      // Min agent ID length
		MaxAgentIDLength     int `yaml:"max_agent_id_length" required:"strict"`      // Max agent ID length
		MaxRequestBodySizeKB int `yaml:"max_request_body_size_kb" required:"strict"` // Max request body

		// Output and storage
		StreamBufferSize  int `yaml:"stream_buffer_size" required:"strict"`  // Stream buffer size
		MaxStoredRequests int `yaml:"max_stored_requests" required:"strict"` // Max stored requests

		// Client cache
		ClientCacheTTLMinutes int `yaml:"client_cache_ttl_minutes" required:"strict"` // Client cache TTL

		// Graceful shutdown
		GracefulShutdownCheckMillis int `yaml:"graceful_shutdown_check_millis" required:"strict"` // Shutdown check interval
		TimeoutWarningThresholdPct  int `yaml:"timeout_warning_threshold_pct" required:"strict"`  // Timeout warning %

		// Cost Control Parameters
		MaxTokensPerCall   int     `yaml:"max_tokens_per_call" required:"strict"`  // Max tokens per request
		MaxTokensPerDay    int     `yaml:"max_tokens_per_day" required:"strict"`   // Max tokens per 24h
		MaxCostPerDay      float64 `yaml:"max_cost_per_day" required:"strict"`     // Max cost per day in USD
		CostAlertThreshold float64 `yaml:"cost_alert_threshold" required:"strict"` // Alert at % of budget

		// Memory Management Parameters
		MaxMemoryMB          int     `yaml:"max_memory_mb" required:"strict"`               // Max memory per request in MB
		MaxDailyMemoryGB     int     `yaml:"max_daily_memory_gb" required:"strict"`         // Max total memory per 24h in GB
		MemoryAlertPercent   float64 `yaml:"memory_alert_percent" required:"strict"`        // Alert when exceeds this %
		MaxContextWindow     int     `yaml:"max_context_window" required:"strict"`          // Max context window size in tokens
		ContextTrimPercent   float64 `yaml:"context_trim_percent" required:"strict"`        // Trim % when full
		SlowCallThresholdSec int     `yaml:"slow_call_threshold_seconds" required:"strict"` // Alert if exceeds this duration

		// Performance & Reliability Parameters
		MaxErrorsPerHour     int `yaml:"max_errors_per_hour" required:"strict"`    // Max errors per hour
		MaxErrorsPerDay      int `yaml:"max_errors_per_day" required:"strict"`     // Max errors per day
		MaxConsecutiveErrors int `yaml:"max_consecutive_errors" required:"strict"` // Max consecutive errors

		// Rate Limiting & Quotas
		MaxCallsPerMinute  int  `yaml:"max_calls_per_minute" required:"strict"`  // Rate limit: calls per minute
		MaxCallsPerHour    int  `yaml:"max_calls_per_hour" required:"strict"`    // Rate limit: calls per hour
		MaxCallsPerDay     int  `yaml:"max_calls_per_day" required:"strict"`     // Rate limit: calls per 24 hours
		BlockOnQuotaExceed bool `yaml:"block_on_quota_exceed" required:"strict"` // BLOCK request when quota exceeded
	} `yaml:"settings"`

	Routing *RoutingConfig `yaml:"routing"`
}

// ModelConfigYAML represents YAML configuration for a model (for parsing)
type ModelConfigYAML struct {
	Model       string `yaml:"model" required:"strict"`    // LLM model name
	Provider    string `yaml:"provider" required:"strict"` // LLM provider name
	ProviderURL string `yaml:"provider_url"`               // Optional provider URL
}

// CostLimitsConfig defines cost control quota limits
type CostLimitsConfig struct {
	MaxTokensPerCall           int     `yaml:"max_tokens_per_call"`            // Max tokens per API call
	MaxTokensPerDay            int     `yaml:"max_tokens_per_day"`             // Max tokens per 24 hours
	MaxCostPerDayUSD           float64 `yaml:"max_cost_per_day_usd"`           // Max USD cost per day
	AlertThreshold             float64 `yaml:"alert_threshold"`                // Warn at % of limit
	BlockOnCostExceed          bool    `yaml:"block_on_cost_exceed"`           // true=BLOCK request, false=WARN only
	InputTokenPricePerMillion  float64 `yaml:"input_token_price_per_million"`  // Cost per 1M input tokens
	OutputTokenPricePerMillion float64 `yaml:"output_token_price_per_million"` // Cost per 1M output tokens
}

// MemoryLimitsConfig defines memory quota limits
type MemoryLimitsConfig struct {
	MaxPerCallMB        int  `yaml:"max_per_call_mb"`        // Max MB per execution
	MaxPerDayMB         int  `yaml:"max_per_day_mb"`         // Max MB per 24 hours
	BlockOnMemoryExceed bool `yaml:"block_on_memory_exceed"` // true=BLOCK request, false=WARN only
}

// ErrorLimitsConfig defines error rate quota limits
type ErrorLimitsConfig struct {
	MaxConsecutive     int  `yaml:"max_consecutive"`       // Max consecutive failures
	MaxPerDay          int  `yaml:"max_per_day"`           // Max errors per 24 hours
	BlockOnErrorExceed bool `yaml:"block_on_error_exceed"` // true=BLOCK request, false=WARN only
}

// LoggingConfig defines observability and monitoring settings
type LoggingConfig struct {
	EnableMemoryMetrics      bool   `yaml:"enable_memory_metrics"`      // Log memory usage per call
	EnablePerformanceMetrics bool   `yaml:"enable_performance_metrics"` // Log response metrics
	EnableQuotaWarnings      bool   `yaml:"enable_quota_warnings"`      // Log quota threshold alerts
	LogLevel                 string `yaml:"log_level"`                  // debug/info/warn/error
}

// AgentConfig represents an agent configuration
type AgentConfig struct {
	ID             string           `yaml:"id" required:"strict"`
	Name           string           `yaml:"name" required:"strict"`
	Description    string           `yaml:"description"`
	Role           string           `yaml:"role" required:"strict"`
	Backstory      string           `yaml:"backstory"`
	Model          string           `yaml:"model"` // Deprecated: Use Primary instead
	Temperature    float64          `yaml:"temperature"`
	IsTerminal     bool             `yaml:"is_terminal"`
	Tools          []string         `yaml:"tools"`
	HandoffTargets []string         `yaml:"handoff_targets"`
	SystemPrompt   string           `yaml:"system_prompt"`
	Provider       string           `yaml:"provider"`                  // Deprecated: Use Primary.Provider instead
	ProviderURL    string           `yaml:"provider_url"`              // Deprecated: Use Primary.ProviderURL instead
	Primary        *ModelConfigYAML `yaml:"primary" required:"strict"` // Primary LLM provider
	Backup         *ModelConfigYAML `yaml:"backup"`                    // Fallback LLM provider

	// Nested quota and monitoring configurations
	CostLimits   *CostLimitsConfig   `yaml:"cost_limits"`   // Token usage and cost limits
	MemoryLimits *MemoryLimitsConfig `yaml:"memory_limits"` // Memory usage limits
	ErrorLimits  *ErrorLimitsConfig  `yaml:"error_limits"`  // Error rate limits
	Logging      *LoggingConfig      `yaml:"logging"`       // Observability settings

	// Backward compatibility: Keep old flat fields for existing configurations
	MaxTokensPerCall   int     `yaml:"max_tokens_per_call"`  // DEPRECATED
	MaxTokensPerDay    int     `yaml:"max_tokens_per_day"`   // DEPRECATED
	MaxCostPerDay      float64 `yaml:"max_cost_per_day"`     // DEPRECATED
	CostAlertThreshold float64 `yaml:"cost_alert_threshold"` // DEPRECATED
	EnforceCostLimits  bool    `yaml:"enforce_cost_limits"`  // DEPRECATED
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
	TotalMemoryMB      int     // ✅ PHASE 2: Sum of all memory samples (for accurate average)
	MemorySampleCount  int     // ✅ PHASE 2: Number of memory samples collected
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
	TotalDurationMs     int64         // ✅ PHASE 2: Sum of all call durations in milliseconds
	CallDurationCount   int           // ✅ PHASE 2: Number of calls with duration tracking
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
	ID                  string
	Name                string
	Description         string
	Role                string
	Backstory           string
	Temperature         float64
	IsTerminal          bool
	Tools               []interface{}
	HandoffTargets      []*Agent
	SystemPrompt        string
	SystemPromptCache   string
	IsSystemPromptDirty bool
	SystemPromptMutex   sync.RWMutex // ✅ PHASE 10b: Protects SystemPromptCache and IsSystemPromptDirty
	PrimaryModel        *ModelConfig
	BackupModel         *ModelConfig
	Metadata            *AgentMetadata
	Quota               *AgentQuotaLimits
	CostMetrics         *AgentCostMetrics
	MemoryMetrics       *AgentMemoryMetrics
	PerformanceMetrics  *AgentPerformanceMetrics
	ctx                 context.Context
	Mutex               sync.RWMutex

	// ✅ PHASE 1: Backward compatibility fields for old agent_execution.go
	// These are populated from PrimaryModel if not explicitly set
	Provider    string // Deprecated: Use PrimaryModel.Provider instead
	Model       string // Deprecated: Use PrimaryModel.Model instead
	ProviderURL string // Deprecated: Use PrimaryModel.ProviderURL instead
}

// HardcodedDefaults stores hardcoded default configuration values
// These are fallback defaults when values are not specified in YAML
type HardcodedDefaults struct {
	// Timeouts (all in seconds)
	DefaultTimeoutSeconds       int
	DefaultToolExecutionTimeout int
	DefaultParallelTimeout      int
	DefaultStreamChunkTimeout   int
	DefaultToolResultTimeout    int

	// Limits (characters, KB, MB, GB)
	DefaultMaxToolOutputChars      int
	DefaultMaxTotalToolOutputChars int
	DefaultMaxInputSizeKB          int
	DefaultMaxRequestBodySizeKB    int

	// ID constraints
	DefaultMinAgentIDLength int
	DefaultMaxAgentIDLength int

	// Token limits
	DefaultMaxTokensPerCall int
	DefaultMaxTokensPerDay  int

	// Cost limits
	DefaultMaxCostPerDay    float64
	DefaultCostAlertPercent float64

	// Memory limits (MB)
	DefaultMaxMemoryPerCall int
	DefaultMaxMemoryPerDay  int

	// Memory limits (GB)
	DefaultMaxDailyMemoryGB int

	// Context window
	DefaultMaxContextWindow   int
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

// ============================================================================
// AGENT METHODS - PHASE 2: Cost, Memory, Performance Quota Management
// These methods support the agent_execution.go cost/quota tracking system
// ============================================================================

// estimateTokens calculates token count (assumes non-nil agent)
func (a *Agent) estimateTokens(content string) int {
	charCount := len(content)
	estimatedTokens := (charCount + 3) / 4 // Round up division
	return max(estimatedTokens, 1)
}

// EstimateTokens estimates token count for content
// Uses rough heuristic: ~4 characters per token (varies by model)
func (a *Agent) EstimateTokens(content string) int {
	if a == nil {
		return 0
	}
	return a.estimateTokens(content)
}

// CheckCostLimits validates cost limits before execution
// Returns error if execution would exceed configured quotas
func (a *Agent) CheckCostLimits(estimatedTokens int) error {
	if a == nil {
		return nil
	}

	if a.CostMetrics == nil {
		return nil
	}

	a.CostMetrics.Mutex.RLock()
	defer a.CostMetrics.Mutex.RUnlock()

	if a.Quota == nil {
		return nil // No quota configured
	}

	// Check max tokens per call
	if a.Quota.MaxTokensPerCall > 0 && estimatedTokens > a.Quota.MaxTokensPerCall {
		return fmt.Errorf(
			"agent '%s': estimated tokens (%d) exceeds max per call (%d)",
			a.ID, estimatedTokens, a.Quota.MaxTokensPerCall,
		)
	}

	// Check max tokens per day
	if a.Quota.MaxTokensPerDay > 0 && a.CostMetrics.TotalTokens+estimatedTokens > a.Quota.MaxTokensPerDay {
		return fmt.Errorf(
			"agent '%s': total daily tokens (%d) would exceed limit (%d)",
			a.ID, a.CostMetrics.TotalTokens+estimatedTokens, a.Quota.MaxTokensPerDay,
		)
	}

	// Check max cost per day (estimated)
	if a.Quota.MaxCostPerDay > 0 {
		estimatedCost := a.CalculateCost(estimatedTokens)
		if a.CostMetrics.DailyCost+estimatedCost > a.Quota.MaxCostPerDay {
			return fmt.Errorf(
				"agent '%s': daily cost ($%.4f) would exceed limit ($%.4f)",
				a.ID, a.CostMetrics.DailyCost+estimatedCost, a.Quota.MaxCostPerDay,
			)
		}
	}

	return nil
}

// CheckErrorQuota validates error rate quotas
// Returns error if error limits have been exceeded
func (a *Agent) CheckErrorQuota() error {
	if a == nil {
		return nil
	}

	if a.PerformanceMetrics == nil {
		return nil
	}

	a.PerformanceMetrics.Mutex.RLock()
	defer a.PerformanceMetrics.Mutex.RUnlock()

	if a.Quota == nil {
		return nil // No quota configured
	}

	// Check consecutive errors
	if a.Quota.MaxErrorsPerHour > 0 && a.PerformanceMetrics.ConsecutiveErrors > a.Quota.MaxErrorsPerHour {
		return fmt.Errorf(
			"agent '%s': consecutive errors (%d) exceeded limit (%d)",
			a.ID, a.PerformanceMetrics.ConsecutiveErrors, a.Quota.MaxErrorsPerHour,
		)
	}

	// Check daily error count
	if a.Quota.MaxErrorsPerDay > 0 && a.PerformanceMetrics.ErrorCountToday >= a.Quota.MaxErrorsPerDay {
		return fmt.Errorf(
			"agent '%s': daily errors (%d) reached limit (%d)",
			a.ID, a.PerformanceMetrics.ErrorCountToday, a.Quota.MaxErrorsPerDay,
		)
	}

	return nil
}

// CheckMemoryQuota validates memory usage quotas
// Returns error if memory limits have been exceeded
func (a *Agent) CheckMemoryQuota() error {
	if a == nil {
		return nil
	}

	if a.MemoryMetrics == nil {
		return nil
	}

	a.MemoryMetrics.Mutex.RLock()
	defer a.MemoryMetrics.Mutex.RUnlock()

	if a.Quota == nil {
		return nil // No quota configured
	}

	// Check memory per call
	if a.Quota.MaxMemoryPerCall > 0 && a.MemoryMetrics.CurrentMemoryMB > a.Quota.MaxMemoryPerCall {
		return fmt.Errorf(
			"agent '%s': current memory (%dMB) exceeds per-call limit (%dMB)",
			a.ID, a.MemoryMetrics.CurrentMemoryMB, a.Quota.MaxMemoryPerCall,
		)
	}

	// Check peak memory
	if a.Quota.MaxMemoryPerDay > 0 && a.MemoryMetrics.PeakMemoryMB > a.Quota.MaxMemoryPerDay {
		return fmt.Errorf(
			"agent '%s': peak memory (%dMB) exceeds daily limit (%dMB)",
			a.ID, a.MemoryMetrics.PeakMemoryMB, a.Quota.MaxMemoryPerDay,
		)
	}

	return nil
}

// CalculateCost estimates cost from token count
// Uses approximate pricing per token
func (a *Agent) CalculateCost(tokenCount int) float64 {
	if a == nil || a.Quota == nil {
		return 0
	}

	// Default pricing approximation (varies by model)
	// GPT-4: ~$30/1M input tokens, ~$60/1M output tokens
	// Assume 60% input, 40% output = average $42/1M
	averagePricePerMillion := 0.042

	costPerToken := averagePricePerMillion / 1_000_000.0
	return float64(tokenCount) * costPerToken
}

// UpdateCostMetrics updates cost tracking after execution
// Increments token count and cost, resets daily counter if 24h passed
func (a *Agent) UpdateCostMetrics(tokenCount int, cost float64) {
	if a == nil || a.CostMetrics == nil {
		return
	}

	a.CostMetrics.Mutex.Lock()
	defer a.CostMetrics.Mutex.Unlock()

	a.CostMetrics.CallCount++
	a.CostMetrics.TotalTokens += tokenCount
	a.CostMetrics.DailyCost += cost

	// Check if we need to reset daily counter (24 hours have passed)
	now := time.Now()
	if now.Sub(a.CostMetrics.LastResetTime) > 24*time.Hour {
		a.CostMetrics.DailyCost = cost
		a.CostMetrics.LastResetTime = now
	}
}

// UpdateMemoryMetrics updates memory tracking after execution
// Tracks current, peak, and average memory usage
// ✅ PHASE 2: Fixed calculation bugs - now tracks sum for accurate averages
func (a *Agent) UpdateMemoryMetrics(memoryMB int, durationMs int64) {
	if a == nil || a.MemoryMetrics == nil {
		return
	}

	a.MemoryMetrics.Mutex.Lock()
	defer a.MemoryMetrics.Mutex.Unlock()

	// Update current and peak memory
	a.MemoryMetrics.CurrentMemoryMB = memoryMB

	if memoryMB > a.MemoryMetrics.PeakMemoryMB {
		a.MemoryMetrics.PeakMemoryMB = memoryMB
	}

	// ✅ PHASE 2: Track sum of memory for accurate average calculation
	a.MemoryMetrics.TotalMemoryMB += memoryMB
	a.MemoryMetrics.MemorySampleCount++

	// Calculate average memory usage = Sum / Count (not Peak * Count / Count!)
	if a.MemoryMetrics.MemorySampleCount > 0 {
		a.MemoryMetrics.AverageMemoryMB = a.MemoryMetrics.TotalMemoryMB / a.MemoryMetrics.MemorySampleCount
	}

	// ✅ PHASE 2: Track sum of durations for accurate call duration average
	if durationMs > 0 {
		a.MemoryMetrics.TotalDurationMs += durationMs
		a.MemoryMetrics.CallDurationCount++

		// Calculate average duration = Total / Count
		if a.MemoryMetrics.CallDurationCount > 0 {
			avgMs := a.MemoryMetrics.TotalDurationMs / int64(a.MemoryMetrics.CallDurationCount)
			a.MemoryMetrics.AverageCallDuration = time.Duration(avgMs) * time.Millisecond
		}
	}
}

// UpdatePerformanceMetrics tracks success/failure metrics
// Updates success rate, consecutive errors, and error count
func (a *Agent) UpdatePerformanceMetrics(success bool, errorMsg string) {
	if a == nil || a.PerformanceMetrics == nil {
		return
	}

	a.PerformanceMetrics.Mutex.Lock()
	defer a.PerformanceMetrics.Mutex.Unlock()

	if success {
		a.PerformanceMetrics.SuccessfulCalls++
		a.PerformanceMetrics.ConsecutiveErrors = 0
	} else {
		a.PerformanceMetrics.FailedCalls++
		a.PerformanceMetrics.ConsecutiveErrors++
		a.PerformanceMetrics.LastError = errorMsg
		a.PerformanceMetrics.LastErrorTime = time.Now()
		a.PerformanceMetrics.ErrorCountToday++
	}

	// Update success rate
	total := a.PerformanceMetrics.SuccessfulCalls + a.PerformanceMetrics.FailedCalls
	if total > 0 {
		a.PerformanceMetrics.SuccessRate = (float64(a.PerformanceMetrics.SuccessfulCalls) / float64(total)) * 100
	}
}

// CheckSlowCall checks if execution time exceeded threshold
// Logs warning if call duration exceeds configured slow call threshold
func (a *Agent) CheckSlowCall(duration time.Duration) {
	if a == nil || a.MemoryMetrics == nil {
		return
	}

	a.MemoryMetrics.Mutex.RLock()
	defer a.MemoryMetrics.Mutex.RUnlock()

	if a.MemoryMetrics.SlowCallThreshold > 0 && duration > a.MemoryMetrics.SlowCallThreshold {
		fmt.Printf("[SLOW_CALL] Agent '%s' execution took %v (threshold: %v)\n",
			a.ID, duration, a.MemoryMetrics.SlowCallThreshold)
	}
}
