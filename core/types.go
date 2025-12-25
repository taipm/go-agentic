package crewai

import (
	"time"
)

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
