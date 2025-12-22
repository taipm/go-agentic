package crewai

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

// ModelConfig represents configuration for a specific LLM model and provider
type ModelConfig struct {
	Model       string  // LLM model name (e.g., "gpt-4o", "deepseek-r1:32b")
	Provider    string  // LLM provider (e.g., "openai", "ollama")
	ProviderURL string  // Provider-specific URL (e.g., "https://api.openai.com", "http://localhost:11434")
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
