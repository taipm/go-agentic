package crewai

import (
	"time"

	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/executor"
)

// ============================================================================
// TYPE ALIASES FOR BACKWARD COMPATIBILITY
// These types are defined in the common package but re-exported here
// to maintain backward compatibility with existing code in the root crewai package
// ============================================================================

// Agent - re-export from common package for backward compatibility
type Agent = common.Agent

// ModelConfig - re-export from common package for backward compatibility
type ModelConfig = common.ModelConfig

// CrewConfig - re-export from common package for backward compatibility
type CrewConfig = common.CrewConfig

// AgentConfig - re-export from common package for backward compatibility
type AgentConfig = common.AgentConfig

// RoutingConfig - re-export from common package for backward compatibility
type RoutingConfig = common.RoutingConfig

// RoutingSignal - re-export from common package for backward compatibility
type RoutingSignal = common.RoutingSignal

// AgentBehavior - re-export from common package for backward compatibility
type AgentBehavior = common.AgentBehavior

// ParallelGroupConfig - re-export from common package for backward compatibility
type ParallelGroupConfig = common.ParallelGroupConfig

// ModelConfigYAML - re-export from common package for backward compatibility
type ModelConfigYAML = common.ModelConfigYAML

// CostLimitsConfig - re-export from common package for backward compatibility
type CostLimitsConfig = common.CostLimitsConfig

// AgentQuotaLimits - re-export from common package for backward compatibility
type AgentQuotaLimits = common.AgentQuotaLimits

// AgentMetadata - re-export from common package for backward compatibility
type AgentMetadata = common.AgentMetadata

// HistoryManager - re-export from executor package for backward compatibility
type HistoryManager = executor.HistoryManager

// ============================================================================
// CORE DOMAIN TYPES (DEFINED LOCALLY)
// ============================================================================

// ToolTimeoutConfig manages timeout settings for tool execution
type ToolTimeoutConfig struct {
	DefaultToolTimeout time.Duration         // Default timeout for tool execution (default: 5s)
	SequenceTimeout    time.Duration         // Timeout for entire tool execution sequence (default: 30s)
	PerToolTimeout     map[string]time.Duration // Per-tool timeout overrides
	CollectMetrics     bool                  // Whether to collect timeout metrics
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

// Tool represents an executable tool that can be used by agents
type Tool struct {
	ID          string
	Name        string
	Description string
	Func        interface{} // The actual function to execute
	Input       interface{} // Input schema or parameters
	Output      interface{} // Output schema or return type
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
	ParallelAgentTimeout    time.Duration  // ✅ FIX #4: Timeout for parallel agent execution (default: 60s)
	MaxToolOutputChars      int            // ✅ FIX #5: Max characters per tool output (default: 2000)
	MaxTotalToolOutputChars int            // ✅ FIX: Max TOTAL characters for all tools combined (default: 4000)
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
