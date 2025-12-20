package agentic

import (
	"context"
	"time"
)

// Tool represents a function that an agent can call
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Handler     func(ctx context.Context, args map[string]interface{}) (string, error)
}

// Agent represents an AI agent in the team
type Agent struct {
	ID             string
	Name           string
	Role           string
	Backstory      string
	Model          string
	SystemPrompt   string // Custom system prompt from config
	Tools          []*Tool
	Temperature    float64
	IsTerminal     bool
	HandoffTargets []string // List of agent IDs that this agent can hand off to
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
	ID       string
	ToolName string
	Arguments map[string]interface{}
}

// ErrorType categorizes different types of errors for better diagnostics
type ErrorType string

const (
	// Command execution errors
	ErrorTypePermissionDenied ErrorType = "PERMISSION_DENIED"
	ErrorTypeNotFound         ErrorType = "NOT_FOUND"
	ErrorTypeTimeout          ErrorType = "TIMEOUT"
	ErrorTypeCommandFailed    ErrorType = "COMMAND_FAILED"
	ErrorTypeParseFailed      ErrorType = "PARSE_FAILED"
	ErrorTypeParameterError   ErrorType = "PARAMETER_ERROR"
	ErrorTypeSystemError      ErrorType = "SYSTEM_ERROR"
	ErrorTypeUnknown          ErrorType = "UNKNOWN"
)

// ToolError represents a structured error from tool execution
type ToolError struct {
	Type           ErrorType // Error classification
	Message        string    // Human-readable error message
	Cause          string    // Underlying error or reason
	SuggestedAction string    // How to fix/resolve the error
}

// AgentResponse represents a response from an agent
type AgentResponse struct {
	AgentID   string
	AgentName string
	Content   string
	ToolCalls []ToolCall
}

// TeamResponse represents the final response from the team
type TeamResponse struct {
	AgentID    string
	AgentName  string
	Content    string
	ToolCalls  []ToolCall
	IsTerminal bool
}

// Team represents a group of agents working together
type Team struct {
	Agents      []*Agent
	Tasks       []*Task
	MaxRounds   int
	MaxHandoffs int
	Routing     *RoutingConfig                // Routing configuration from team.yaml
	RoutingRules []UnifiedRoutingRule        // Routing rules from unified config
}

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

// RoutingConfig defines routing rules for the team
type RoutingConfig struct {
	Signals        map[string][]RoutingSignal `yaml:"signals"`
	Defaults       map[string]string          `yaml:"defaults"`
	AgentBehaviors map[string]AgentBehavior   `yaml:"agent_behaviors"`
}

// TeamConfig represents the team configuration
type TeamConfig struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	EntryPoint  string `yaml:"entry_point"`

	Agents []string `yaml:"agents"`

	Settings struct {
		MaxHandoffs     int    `yaml:"max_handoffs"`
		MaxRounds       int    `yaml:"max_rounds"`
		TimeoutSeconds  int    `yaml:"timeout_seconds"`
		Language        string `yaml:"language"`
		Organization    string `yaml:"organization"`
	} `yaml:"settings"`

	Routing *RoutingConfig `yaml:"routing"`
}

// AgentConfig represents agent configuration
type AgentConfig struct {
	ID           string                 `yaml:"id"`
	Name         string                 `yaml:"name"`
	Role         string                 `yaml:"role"`
	Backstory    string                 `yaml:"backstory"`
	Model        string                 `yaml:"model"`
	SystemPrompt string                 `yaml:"system_prompt"`
	Tools        []string               `yaml:"tools"`
	Temperature  *float64               `yaml:"temperature"`
	IsTerminal   bool                   `yaml:"is_terminal"`
	HandoffTo    []string               `yaml:"handoff_to"`
	Config       map[string]interface{} `yaml:"config"`
}

// StreamEvent represents a streaming event sent to the client
type StreamEvent struct {
	Type      string      `json:"type"`      // "agent_start", "agent_response", "tool_start", "tool_result", "pause", "error"
	Agent     string      `json:"agent"`     // Agent ID/Name
	Content   string      `json:"content"`   // Main message
	Timestamp time.Time   `json:"timestamp"` // When this happened
	Metadata  interface{} `json:"metadata"`  // Extra data (tool results, etc.)
}

// Deprecated: Use Team instead
type Crew = Team

// Deprecated: Use TeamConfig instead
type CrewConfig = TeamConfig

// Deprecated: Use TeamResponse instead
type CrewResponse = TeamResponse
