package crewai

import (
	"context"
	"sync"
	"time"
)

// Tool represents a function that an agent can call
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Handler     func(ctx context.Context, args map[string]interface{}) (string, error)
}

// ModelConfig represents configuration for a specific LLM model and provider
type ModelConfig struct {
	Model       string  // LLM model name (e.g., "gpt-4o", "deepseek-r1:32b")
	Provider    string  // LLM provider (e.g., "openai", "ollama")
	ProviderURL string  // Provider-specific URL (e.g., "https://api.openai.com", "http://localhost:11434")
}

// ✅ WEEK 1+: Unified Agent Metrics & Monitoring
// Consolidated metadata for comprehensive agent tracking

// AgentCostMetrics tracks cost and token usage for an agent
type AgentCostMetrics struct {
	CallCount      int           // Number of calls this period
	TotalTokens    int           // Total tokens used
	DailyCost      float64       // Total cost today
	LastResetTime  time.Time     // When daily counter resets
	Mutex          sync.RWMutex  // Protect concurrent access
}

// AgentMemoryMetrics tracks memory usage and quotas for an agent
type AgentMemoryMetrics struct {
	// Memory Usage (Runtime)
	CurrentMemoryMB    int       // Current memory usage in MB
	PeakMemoryMB       int       // Peak memory usage ever recorded
	AverageMemoryMB    int       // Average memory usage across calls
	MemoryTrendPercent float64   // Trend: positive = increasing, negative = decreasing

	// Memory Quota & Thresholds
	MaxMemoryMB        int       // Max allowed memory per call (default: 512 MB)
	MaxDailyMemoryGB   int       // Max total memory per day (default: 10 GB)
	MemoryAlertPercent float64   // Alert when exceeds this % (default: 80%)

	// Context Window & History
	CurrentContextSize int       // Current conversation history in tokens
	MaxContextWindow   int       // Max context window (default: 32000 for gpt-4)
	ContextTrimPercent float64   // How much to trim when full (default: 20%)

	// Call Metrics
	AverageCallDuration time.Duration // Average execution time per call
	SlowCallThreshold   time.Duration // Alert if exceeds this (default: 30s)

	Mutex              sync.RWMutex  // Protect concurrent access
}

// AgentPerformanceMetrics tracks execution performance and reliability
type AgentPerformanceMetrics struct {
	// Response Quality
	SuccessfulCalls      int       // Number of successful executions
	FailedCalls          int       // Number of failed executions
	SuccessRate          float64   // Success rate percentage (0-100)
	AverageResponseTime  time.Duration // Average time per call

	// Error Tracking
	LastError            string    // Last error message
	LastErrorTime        time.Time // When last error occurred
	ConsecutiveErrors    int       // Number of consecutive errors
	ErrorCountToday      int       // Errors in last 24 hours

	// Performance Thresholds
	MaxErrorsPerHour     int       // Alert if exceeds (default: 10)
	MaxErrorsPerDay      int       // Block if exceeds (default: 50)
	MaxConsecutiveErrors int       // Block after N consecutive (default: 5)

	Mutex                sync.RWMutex  // Protect concurrent access
}

// AgentQuotaLimits defines all quota constraints for an agent
type AgentQuotaLimits struct {
	// Cost Quotas
	MaxTokensPerCall   int     // Max tokens per request
	MaxTokensPerDay    int     // Max tokens per 24h
	MaxCostPerDay      float64 // Max cost per 24h in USD
	CostAlertPercent   float64 // Alert at % of budget

	// Memory Quotas
	MaxMemoryPerCall   int     // Max MB per request
	MaxMemoryPerDay    int     // Max MB per 24h
	MaxContextWindow   int     // Max context tokens

	// Execution Quotas
	MaxCallsPerMinute  int     // Rate limiting: calls/minute
	MaxCallsPerHour    int     // Rate limiting: calls/hour
	MaxCallsPerDay     int     // Rate limiting: calls/day
	MaxErrorsPerHour   int     // Error rate: errors/hour
	MaxErrorsPerDay    int     // Error rate: errors/day

	// Enforcement
	EnforceQuotas      bool    // true=block, false=warn (default: true for critical agents)
}

// AgentMetadata is the unified metadata hub for agent monitoring
// Consolidates cost, memory, performance, and quota tracking
type AgentMetadata struct {
	// Core Identifiers
	AgentID        string
	AgentName      string
	CreatedTime    time.Time
	LastAccessTime time.Time

	// Configuration & Quotas
	Quotas           AgentQuotaLimits
	EnforceCostLimits bool       // Legacy: kept for backward compatibility

	// Runtime Metrics (Updated in real-time)
	Cost             AgentCostMetrics
	Memory           AgentMemoryMetrics
	Performance      AgentPerformanceMetrics

	// Global Mutex for all metrics
	Mutex            sync.RWMutex
}

// Agent represents an AI agent in the crew
type Agent struct {
	ID             string
	Name           string
	Role           string
	Backstory      string
	Model          string       // Deprecated: Use Primary.Model instead
	Provider       string       // Deprecated: Use Primary.Provider instead
	ProviderURL    string       // Deprecated: Use Primary.ProviderURL instead
	Primary        *ModelConfig // Primary LLM model configuration (required)
	Backup         *ModelConfig // Backup LLM model configuration (optional)
	SystemPrompt   string       // Custom system prompt from config
	Tools          []*Tool
	Temperature    float64
	IsTerminal     bool
	HandoffTargets []string // List of agent IDs that this agent can hand off to

	// ✅ WEEK 2: Unified Agent Metadata (Consolidates cost, memory, performance, and quotas)
	Metadata *AgentMetadata `json:"-" yaml:"-"` // Unified metadata hub for all agent monitoring

	// ✅ LEGACY: Cost Control Configuration (Week 1) - For backward compatibility
	// These fields are now stored in Metadata.Quotas and Metadata.Cost
	MaxTokensPerCall   int     `yaml:"max_tokens_per_call"`   // Max tokens per single call (e.g., 1000)
	MaxTokensPerDay    int     `yaml:"max_tokens_per_day"`    // Max cumulative tokens per day (e.g., 50000)
	MaxCostPerDay      float64 `yaml:"max_cost_per_day"`      // Max daily budget in USD (e.g., 10.00)
	CostAlertThreshold float64 `yaml:"cost_alert_threshold"`  // Warn when % of budget used (e.g., 0.80 = 80%)
	EnforceCostLimits  bool    `yaml:"enforce_cost_limits"`   // true=block, false=warn (default: true)
	CostMetrics        AgentCostMetrics `json:"-" yaml:"-"` // Runtime metrics (deprecated, use Metadata.Cost)
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
// ✅ FIX #4 & #5: Made ParallelAgentTimeout and MaxToolOutputChars configurable (were hardcoded constants)
type Crew struct {
	Agents                  []*Agent
	Tasks                   []*Task
	MaxRounds               int
	MaxHandoffs             int
	ParallelAgentTimeout    time.Duration // ✅ FIX #4: Timeout for parallel agent execution (default: 60s)
	MaxToolOutputChars      int           // ✅ FIX #5: Max characters in tool output (default: 2000)
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
