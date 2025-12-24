package crewai

// SignalBehavior defines how a signal should be handled
type SignalBehavior string

const (
	SignalBehaviorRoute      SignalBehavior = "route"      // Signal routes to target agent
	SignalBehaviorTerminate  SignalBehavior = "terminate"  // Signal terminates workflow
	SignalBehaviorParallel   SignalBehavior = "parallel"   // Signal triggers parallel group
	SignalBehaviorPause      SignalBehavior = "pause"      // Signal pauses execution (wait for resume)
	SignalBehaviorBroadcast  SignalBehavior = "broadcast"  // Signal sent to multiple agents
)

// SignalDefinition defines a single signal in the system
type SignalDefinition struct {
	// Identity
	Name        string // e.g., "[END_EXAM]"
	Description string // e.g., "Terminates exam workflow"

	// Permissions - which agents can emit this signal
	AllowedAgents []string // Agent IDs that can emit; empty = all agents
	AllowAllAgents bool    // If true, any agent can emit

	// Routing
	Behavior      SignalBehavior // How to handle the signal (route, terminate, etc)
	DefaultTarget string         // Default target if not specified in config
	ValidTargets  []string       // List of valid target agents/groups

	// Validation
	IsRequired  bool   // Must be emitted by agent?
	IsInternal  bool   // Internal signal (hidden from logs)?
	MaxOccurs   int    // Max times signal can be emitted (-1 = unlimited)
	Priority    int    // Priority for signal matching (higher = matched first)

	// Documentation
	Example    string // Example of how signal appears in output
	RelatedTo  []string // Related signals
	DeprecatedMsg string // If non-empty, signal is deprecated with this message
}

// SignalEvent represents an occurrence of a signal in a workflow
type SignalEvent struct {
	Signal       string // Signal name (e.g., "[END_EXAM]")
	AgentID      string // Agent that emitted the signal
	Timestamp    int64  // Unix timestamp
	ResponseText string // Full response containing the signal
	MatchMethod  string // How signal was matched: "exact", "case_insensitive", "normalized"
}

// SignalValidationError represents a validation error for a signal
type SignalValidationError struct {
	Signal   string
	AgentID  string
	ErrorMsg string
	Severity string // "error" or "warning"
}

// SignalUsageStats tracks signal usage across workflow execution
type SignalUsageStats struct {
	SignalName    string
	EmitCount     int
	SuccessCount  int
	FailureCount  int
	AverageTime   float64 // milliseconds
	LastUsed      int64   // Unix timestamp
}
