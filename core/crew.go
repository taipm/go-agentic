package crewai

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taipm/go-agentic/core/signal"
)


// CrewExecutor handles the execution of a crew
type CrewExecutor struct {
	crew           *Crew
	apiKey         string
	entryAgent     *Agent
	history        *HistoryManager    // Manages conversation history with thread-safe operations
	Verbose        bool               // If true, print agent responses and tool results to stdout
	ResumeAgentID  string             // If set, execution will start from this agent instead of entry agent
	ToolTimeouts   *ToolTimeoutConfig // Timeout configuration
	Metrics        *MetricsCollector  // Metrics collection for observability
	defaults       *HardcodedDefaults // Runtime configuration defaults
	signalRegistry *signal.SignalRegistry    // Optional signal registry for enhanced validation (Phase 3.5)
}

// NewCrewExecutor creates a new crew executor
// Note: crew.Routing MUST be set for signal-based routing to work
// Best Practice: Use entry_point from crew.yaml instead of relying on IsTerminal
func NewCrewExecutor(crew *Crew, apiKey string) *CrewExecutor {
	// Validate input: crew cannot be nil
	if crew == nil {
		log.Println("WARNING: CrewExecutor created with nil crew - will need to set entry agent manually")
		return nil
	}

	// Find entry agent - first agent is default if no routing configured
	var entryAgent *Agent
	if len(crew.Agents) > 0 {
		entryAgent = crew.Agents[0] // First agent is default entry point
	}

	return &CrewExecutor{
		crew:         crew,
		apiKey:       apiKey,
		entryAgent:   entryAgent,
		history:      NewHistoryManager(),
		ToolTimeouts: NewToolTimeoutConfig(),
		Metrics:      NewMetricsCollector(),
		defaults:     DefaultHardcodedDefaults(),
	}
}

// SetSignalRegistry sets the signal registry for enhanced signal validation (Phase 3.5)
// When a registry is set, ValidateSignals() will perform enhanced validation using the registry
// This is optional - the executor works fine without a registry (Phase 2 validation only)
func (ce *CrewExecutor) SetSignalRegistry(registry *signal.SignalRegistry) {
	if ce != nil {
		ce.signalRegistry = registry
	}
}

// NewCrewExecutorFromConfig creates a crew executor by loading configuration from files
// This is the recommended way to initialize a crew with routing configuration
// tools: map of available tools that can be assigned to agents (can be empty map if no tools needed)
func NewCrewExecutorFromConfig(apiKey, configDir string, tools map[string]*Tool) (*CrewExecutor, error) {
	// Load crew configuration (includes routing config)
	crewConfig, err := LoadCrewConfig(fmt.Sprintf("%s/crew.yaml", configDir))
	if err != nil {
		return nil, fmt.Errorf("failed to load crew config: %w", err)
	}

	// Load agent configurations
	agentDir := fmt.Sprintf("%s/agents", configDir)
	// Extract configMode from crew config for agent loading
	configMode := PermissiveMode // Default to permissive for backward compatibility
	if crewConfig.Settings.ConfigMode != "" {
		configMode = ConfigMode(crewConfig.Settings.ConfigMode)
	}
	agentConfigs, err := LoadAgentConfigs(agentDir, configMode)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent configs: %w", err)
	}

	// Create agents from config with provided tools
	var agents []*Agent
	for _, agentID := range crewConfig.Agents {
		if config, exists := agentConfigs[agentID]; exists {
			agent := CreateAgentFromConfig(config, tools)
			agents = append(agents, agent)
		}
	}

	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents loaded from configuration")
	}

	// Create crew with routing configuration
	crew := &Crew{
		Agents:      agents,
		MaxRounds:   crewConfig.Settings.MaxRounds,
		MaxHandoffs: crewConfig.Settings.MaxHandoffs,
		Routing:     crewConfig.Routing, // ← Routing loaded from YAML
	}

	// Create executor
	executor := NewCrewExecutor(crew, apiKey)

	// Validate signals at startup (fail-fast approach)
	if err := executor.ValidateSignals(); err != nil {
		return nil, fmt.Errorf("signal validation failed: %w", err)
	}

	// Convert YAML config to runtime defaults
	executor.defaults = ConfigToHardcodedDefaults(crewConfig)

	// In STRICT MODE, validation errors are fatal
	if executor.defaults == nil {
		return nil, fmt.Errorf("STRICT MODE configuration validation failed - see errors above")
	}

	// Validate configuration and log mode warning if needed
	if err := executor.defaults.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	log.Println(executor.defaults.LogConfigurationMode())

	// Set entry agent based on entry_point from YAML (best practice)
	if crewConfig.EntryPoint != "" {
		for _, agent := range executor.crew.Agents {
			if agent.ID == crewConfig.EntryPoint {
				executor.entryAgent = agent
				break
			}
		}
	}

	return executor, nil
}

// SetVerbose enables or disables verbose output
func (ce *CrewExecutor) SetVerbose(verbose bool) {
	ce.Verbose = verbose
}

// SetResumeAgent sets the agent to resume from in the next execution
// This is used when continuing a conversation after a PAUSE (wait_for_signal)
func (ce *CrewExecutor) SetResumeAgent(agentID string) {
	ce.ResumeAgentID = agentID
}

// ClearResumeAgent clears the resume agent, so next execution starts from entry agent
func (ce *CrewExecutor) ClearResumeAgent() {
	ce.ResumeAgentID = ""
}

// GetResumeAgentID returns the current resume agent ID
func (ce *CrewExecutor) GetResumeAgentID() string {
	return ce.ResumeAgentID
}

// appendMessage safely appends a message to history with mutex protection
// Delegates to HistoryManager for thread-safe operations
func (ce *CrewExecutor) appendMessage(msg Message) {
	if ce.history != nil {
		ce.history.Append(msg)
	}
}

// getHistoryCopy returns a copy of history for safe reading
// Caller can safely read the returned copy without affecting concurrent writers
// Delegates to HistoryManager
func (ce *CrewExecutor) getHistoryCopy() []Message {
	if ce.history == nil {
		return []Message{}
	}
	return ce.history.Copy()
}

// GetHistory returns a copy of the conversation history
// Allows inspection for debugging memory issues
func (ce *CrewExecutor) GetHistory() []Message {
	return ce.getHistoryCopy()
}

// estimateHistoryTokens estimates total tokens in conversation history
// Uses approximation: 1 token ≈ TokenDivisor characters (OpenAI convention)
// Delegates to HistoryManager
func (ce *CrewExecutor) estimateHistoryTokens() int {
	if ce.history == nil {
		return 0
	}
	return ce.history.EstimateTokens()
}

// trimHistoryIfNeeded trims conversation history to fit within context window
// Uses ce.defaults.MaxContextWindow and ce.defaults.ContextTrimPercent
// Strategy: Keep first + recent messages, remove oldest in middle when over limit
// Delegates to HistoryManager
func (ce *CrewExecutor) trimHistoryIfNeeded() {
	if ce.history == nil || ce.defaults == nil {
		return
	}
	ce.history.TrimIfNeeded(ce.defaults.MaxContextWindow, ce.defaults.ContextTrimPercent)
}

// ClearHistory clears the conversation history
// Useful for starting fresh conversations
func (ce *CrewExecutor) ClearHistory() {
	if ce.history != nil {
		ce.history.Clear()
	}

	// ✅ Reset session cost tracking when starting fresh
	if ce.Metrics != nil {
		ce.Metrics.ResetSessionCost()
	}
}

// ===== PHASE 2: EXTRACTED HELPER FUNCTIONS =====

// sendStreamEvent safely sends a stream event to the channel
// It handles nil channels gracefully and won't block indefinitely
func (ce *CrewExecutor) sendStreamEvent(streamChan chan *StreamEvent, eventType string, agentName string, message string) {
	if streamChan == nil {
		return
	}

	select {
	case streamChan <- NewStreamEvent(eventType, agentName, message):
		// Event sent successfully
	case <-time.After(100 * time.Millisecond):
		// Timeout - channel might be full or blocked
		log.Printf("WARNING: stream event send timeout for event: %s", eventType)
	}
}

// handleAgentError handles errors from agent execution
// It logs the error, sends a stream event, and returns the error
// This centralizes error handling logic that was previously duplicated
func (ce *CrewExecutor) handleAgentError(ctx context.Context, agent *Agent, err error, streamChan chan *StreamEvent) error {
	if err == nil {
		return nil
	}

	// Log the error
	log.Printf("[ERROR] Agent %s: %v", agent.ID, err)

	// Send stream event
	ce.sendStreamEvent(streamChan, EventTypeError, agent.Name,
		fmt.Sprintf("Agent failed: %v", err))

	// Update performance metrics
	if agent.Metadata != nil {
		agent.UpdatePerformanceMetrics(false, err.Error())
	}

	return err
}

// updateAgentMetrics updates agent performance and memory metrics after execution
// It handles nil agent and nil metadata gracefully
// memory: estimated memory usage in MB (int)
// duration: execution duration as time.Duration
func (ce *CrewExecutor) updateAgentMetrics(agent *Agent, success bool, duration time.Duration, memory int, errorMsg string) error {
	if agent == nil || agent.Metadata == nil {
		return nil
	}

	// Update performance metrics
	agent.UpdatePerformanceMetrics(success, errorMsg)

	// Update memory metrics (convert duration to milliseconds)
	durationMs := duration.Milliseconds()
	agent.UpdateMemoryMetrics(memory, durationMs)

	return nil
}

// calculateMessageTokens calculates the token count for a message
// Uses the formula: base_tokens + (content_length + padding) / divisor
func calculateMessageTokens(msg Message) int {
	return TokenBaseValue + (len(msg.Content)+TokenPaddingValue)/TokenDivisor
}

// addUserMessageToHistory adds a user message to the conversation history
// This is a convenience wrapper around appendMessage
func (ce *CrewExecutor) addUserMessageToHistory(content string) {
	ce.appendMessage(Message{
		Role:    RoleUser,
		Content: content,
	})
}

// addAssistantMessageToHistory adds an assistant message to the conversation history
// This is a convenience wrapper around appendMessage
func (ce *CrewExecutor) addAssistantMessageToHistory(content string) {
	ce.appendMessage(Message{
		Role:    RoleAssistant,
		Content: content,
	})
}

// recordAgentExecution records metrics about agent execution
// Updates both internal metrics and global metrics collector
func (ce *CrewExecutor) recordAgentExecution(agent *Agent, duration time.Duration, success bool) {
	if agent == nil || ce.Metrics == nil {
		return
	}
	ce.Metrics.RecordAgentExecution(agent.ID, agent.Name, duration, success)
}

// ExecuteStream runs the crew with streaming events
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input string, streamChan chan *StreamEvent) error {
	handler := NewStreamHandler(streamChan)
	result := ce.executeWorkflow(ctx, input, handler)
	if err, ok := result.(error); ok {
		return err
	}
	return nil
}

// Execute runs the crew with the given input
func (ce *CrewExecutor) Execute(ctx context.Context, input string) (*CrewResponse, error) {
	handler := NewSyncHandler(ce, ce.Verbose)
	result := ce.executeWorkflow(ctx, input, handler)

	// Handle the result based on its type
	if err, ok := result.(error); ok {
		return nil, err
	}

	if response, ok := result.(*CrewResponse); ok {
		return response, nil
	}

	// Should not reach here
	return nil, fmt.Errorf("unexpected result type from executeWorkflow")
}
