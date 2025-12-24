package crewai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/taipm/go-agentic/core/signal"
)

// ===== Token Calculation Constants =====
const (
	// TokenBaseValue: Base tokens allocated per message
	// Used in token estimation for context window management
	TokenBaseValue = 4

	// TokenPaddingValue: Padding added to content length for token calculation
	// Used in formula: baseTokens + (contentLength + padding) / divisor
	TokenPaddingValue = 3

	// TokenDivisor: Divisor for token calculation
	// Normalizes content length to approximate token count
	TokenDivisor = 4

	// MinHistoryLength: Minimum messages to keep before trimming context
	// Preserves at least this many recent messages even during aggressive trimming
	MinHistoryLength = 2

	// PercentDivisor: Divisor to convert percentage values (e.g., 20 -> 0.20)
	// Used when reading percentage configuration like ContextTrimPercent
	PercentDivisor = 100.0
)

// ===== Message & Event Constants =====
const (
	// Message Role Constants - define the source of a message
	RoleUser      = "user"      // Messages from the user/human
	RoleAssistant = "assistant" // Messages from the AI agent
	RoleSystem    = "system"    // System-level messages (e.g., errors, events)

	// Event Type Constants - define the type of stream event
	EventTypeError      = "error"       // Event indicates an error occurred
	EventTypeToolResult = "tool_result" // Event contains results from tool execution
)

// ===== Timing & Retry Constants =====
const (
	// BaseRetryDelay: Initial delay for exponential backoff retry strategy
	// Subsequent attempts double this duration: 100ms, 200ms, 400ms, etc.
	BaseRetryDelay = 100 * time.Millisecond

	// MinTimeoutValue: Minimum timeout for urgent operations
	// Used when remaining time in sequence is critically low
	MinTimeoutValue = 100 * time.Millisecond

	// WarnThresholdRatio: Ratio for timeout warning threshold (20% = 1/5)
	// Used to warn when approaching sequence deadline
	WarnThresholdRatio = 5
)

// copyHistory creates a deep copy of message history to ensure thread safety
// Each execution gets its own isolated history snapshot, preventing race conditions
// when concurrent requests execute and pause/resume
func copyHistory(original []Message) []Message {
	if len(original) == 0 {
		return []Message{}
	}
	copied := make([]Message, len(original))
	copy(copied, original)
	return copied
}

// extractRequiredFields extracts required field names from tool parameters
func extractRequiredFields(params map[string]interface{}) []string {
	var requiredFields []string
	if required, ok := params["required"].([]interface{}); ok {
		for _, fieldName := range required {
			if fieldStr, ok := fieldName.(string); ok {
				requiredFields = append(requiredFields, fieldStr)
			}
		}
	}
	return requiredFields
}

// validateFieldType validates a single field's type against schema
func validateFieldType(tool *Tool, fieldName string, fieldValue interface{}, propSchema interface{}) error {
	propMap, ok := propSchema.(map[string]interface{})
	if !ok {
		return nil // Skip if schema is not a map
	}

	expectedType, ok := propMap["type"].(string)
	if !ok {
		return nil // Skip if type not specified
	}

	// Validate based on expected type
	switch expectedType {
	case "string":
		// Allow numeric types to be coerced to string (common with text-parsed tool calls from Ollama)
		switch fieldValue.(type) {
		case string:
		case float64, int, int64, int32:
			// Numeric types can be coerced to string - validation passes
			// Handler should do the actual conversion
		default:
			return fmt.Errorf("tool '%s': parameter '%s' should be string, got %T", tool.Name, fieldName, fieldValue)
		}
	case "number", "integer":
		switch fieldValue.(type) {
		case float64, int, int64:
		default:
			return fmt.Errorf("tool '%s': parameter '%s' should be number, got %T", tool.Name, fieldName, fieldValue)
		}
	}
	return nil
}

// validateToolArguments validates tool arguments against tool definition
// Required parameters must be present and types must match schema
func validateToolArguments(tool *Tool, args map[string]interface{}) error {
	if tool.Parameters == nil {
		return nil // No parameters defined, so any args are acceptable
	}

	// Get parameter schema
	properties, ok := tool.Parameters["properties"].(map[string]interface{})
	if !ok {
		return nil // No properties defined, skip validation
	}

	// Check required fields are present
	requiredFields := extractRequiredFields(tool.Parameters)
	for _, fieldName := range requiredFields {
		if _, exists := args[fieldName]; !exists {
			return fmt.Errorf("tool '%s': required parameter '%s' is missing", tool.Name, fieldName)
		}
	}

	// Validate parameter types
	for argName, argValue := range args {
		if propSchema, exists := properties[argName]; exists {
			if err := validateFieldType(tool, argName, argValue, propSchema); err != nil {
				return err
			}
		}
	}

	return nil
}

// safeExecuteTool wraps tool execution with panic recovery for graceful error handling
// This prevents one buggy tool from crashing the entire server
// Pattern: defer-recover catches panic and converts it to error (Go standard approach)
type ErrorType int

const (
	ErrorTypeUnknown    ErrorType = iota
	ErrorTypeTimeout              // Transient: exceeded deadline
	ErrorTypePanic                // Non-transient: panic in tool
	ErrorTypeValidation           // Non-transient: invalid arguments
	ErrorTypeNetwork              // Transient: connection issues
	ErrorTypeTemporary            // Transient: temporary failures
	ErrorTypePermanent            // Non-transient: permanent failure
)

// classifyError determines if an error is transient (retryable) or permanent
// Transient errors (timeout, network) trigger retry logic
// Permanent errors (validation, panic) fail immediately
func classifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errStr := err.Error()

	// Timeout errors are transient
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorTypeTimeout
	}

	// Check for panic signature
	if strings.Contains(errStr, "panicked:") {
		return ErrorTypePanic
	}

	// Check for validation errors (non-transient)
	if strings.Contains(errStr, "required field") || strings.Contains(errStr, "parameter") {
		return ErrorTypeValidation
	}

	// Check for network-like errors (transient)
	networkPatterns := []string{
		"connection refused", "connection reset", "broken pipe",
		"network unreachable", "host unreachable", "no such host",
		"temporary failure", "i/o timeout",
	}
	for _, pattern := range networkPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return ErrorTypeNetwork
		}
	}

	// Default to temporary (transient) for unknown errors
	return ErrorTypeTemporary
}

// isRetryable determines if an error type should trigger a retry
func isRetryable(errType ErrorType) bool {
	switch errType {
	case ErrorTypeTimeout, ErrorTypeNetwork, ErrorTypeTemporary:
		return true
	case ErrorTypePanic, ErrorTypeValidation, ErrorTypePermanent, ErrorTypeUnknown:
		return false
	default:
		return false
	}
}

// calculateBackoffDuration returns exponential backoff with jitter
// Starts at BaseRetryDelay, doubles each attempt, capped at 5 seconds
func calculateBackoffDuration(attempt int) time.Duration {
	// Start with BaseRetryDelay, double each attempt
	baseDelay := BaseRetryDelay * (1 << uint(attempt))

	// Cap at 5 seconds
	if baseDelay > 5*time.Second {
		baseDelay = 5 * time.Second
	}

	return baseDelay
}

// retryWithBackoff executes a tool with exponential backoff retry logic
// Returns immediately on non-retryable errors
func retryWithBackoff(ctx context.Context, tool *Tool, args map[string]interface{}, maxRetries int) (string, error) {
	// Handle nil context
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Try to execute the tool
		output, err := safeExecuteToolOnce(ctx, tool, args)

		// If successful, return
		if err == nil {
			if attempt > 0 {
				log.Printf("[TOOL RETRY] %s succeeded on attempt %d", tool.Name, attempt+1)
			}
			return output, nil
		}

		lastErr = err
		errType := classifyError(err)

		// If non-retryable, return immediately
		if !isRetryable(errType) {
			log.Printf("[TOOL ERROR] %s failed with non-retryable error: %v", tool.Name, err)
			return "", err
		}

		// If this was the last attempt, return the error
		if attempt == maxRetries {
			log.Printf("[TOOL ERROR] %s failed after %d retries: %v", tool.Name, maxRetries+1, err)
			return "", err
		}

		// Calculate backoff and wait before retry
		backoff := calculateBackoffDuration(attempt)
		log.Printf("[TOOL RETRY] %s failed on attempt %d (type: %v), retrying in %v: %v",
			tool.Name, attempt+1, errType, backoff, err)

		select {
		case <-ctx.Done():
			log.Printf("[TOOL RETRY] %s context cancelled during backoff", tool.Name)
			return "", ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	return "", lastErr
}

// safeExecuteToolOnce executes a tool once without retry
// Panic recovery prevents one buggy tool from crashing the entire system
func safeExecuteToolOnce(ctx context.Context, tool *Tool, args map[string]interface{}) (output string, err error) {
	defer func() {
		// Catch panic and convert to error
		if r := recover(); r != nil {
			err = fmt.Errorf("tool %s panicked: %v", tool.Name, r)
		}
	}()

	// Validate arguments before execution
	if validationErr := validateToolArguments(tool, args); validationErr != nil {
		return "", validationErr
	}

	// Execute tool - if it panics, defer above will catch it
	return tool.Handler(ctx, args)
}

// safeExecuteTool is the main entry point for tool execution with error recovery
// Attempts retry on transient errors with exponential backoff
func safeExecuteTool(ctx context.Context, tool *Tool, args map[string]interface{}) (output string, err error) {
	// Default to 0 retries (1 total attempt: no retry)
	// COST HINT: Each retry multiplies LLM API costs. Set higher values only if:
	// - You have unreliable network/tools that frequently timeout
	// - Cost is not a concern (e.g., using local Ollama)
	// - You need high reliability over cost efficiency
	// Values: 0 = no retry (cheapest), 1 = 2 attempts, 2 = 3 attempts (most expensive)
	maxRetries := 0

	return retryWithBackoff(ctx, tool, args, maxRetries)
}

// ExecutionMetrics tracks execution time and status for tools
type ExecutionMetrics struct {
	ToolName  string        // Name of the tool executed
	Duration  time.Duration // Time taken to execute
	Status    string        // "success", "timeout", "error"
	TimedOut  bool          // True if tool execution exceeded timeout
	StartTime time.Time     // When tool execution started
	EndTime   time.Time     // When tool execution completed
}

// TimeoutTracker tracks sequence execution time and manages per-tool budgets
// Prevents tools from exceeding allocated time in a sequence
type TimeoutTracker struct {
	startTime      time.Time     // When sequence started
	deadline       time.Time     // When sequence must complete
	overheadBudget time.Duration // Estimated overhead per tool (e.g., 500ms for LLM calls)
	usedTime       time.Duration // Time already consumed in sequence
	mu             sync.Mutex    // Protect concurrent access
}

// NewTimeoutTracker creates a timeout tracker for a sequence
func NewTimeoutTracker(sequenceTimeout time.Duration, overheadBudget time.Duration) *TimeoutTracker {
	now := time.Now()
	return &TimeoutTracker{
		startTime:      now,
		deadline:       now.Add(sequenceTimeout),
		overheadBudget: overheadBudget,
		usedTime:       0,
	}
}

// GetRemainingTime returns how much time is left in the sequence
func (t *TimeoutTracker) GetRemainingTime() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	remaining := time.Until(t.deadline)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CalculateToolTimeout calculates the appropriate timeout for the next tool
// accounting for: per-tool timeout, remaining sequence time, and overhead budget
func (t *TimeoutTracker) CalculateToolTimeout(defaultTimeout, perToolTimeout time.Duration) time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Start with per-tool timeout, fallback to default
	toolTimeout := perToolTimeout
	if toolTimeout <= 0 {
		toolTimeout = defaultTimeout
	}

	// Get remaining sequence time and subtract overhead budget
	remaining := time.Until(t.deadline)
	if remaining <= t.overheadBudget {
		// Not enough time even for overhead
		return MinTimeoutValue // Minimal timeout to signal urgency
	}

	// Available time is remaining minus overhead
	availableTime := remaining - t.overheadBudget

	// Use the minimum: per-tool timeout or available time
	if toolTimeout > availableTime {
		return availableTime
	}
	return toolTimeout
}

// RecordToolExecution records that a tool has finished and updates used time
func (t *TimeoutTracker) RecordToolExecution(duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.usedTime += duration
}

// IsTimeoutWarning returns true if we're within 20% of sequence deadline
func (t *TimeoutTracker) IsTimeoutWarning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	remaining := time.Until(t.deadline)
	totalDuration := t.deadline.Sub(t.startTime)
	warnThreshold := totalDuration / time.Duration(WarnThresholdRatio) // 20%
	return remaining < warnThreshold && remaining > 0
}

// ToolTimeoutConfig defines timeout behavior for tools
type ToolTimeoutConfig struct {
	DefaultToolTimeout time.Duration            // Default timeout per tool (e.g., 5s)
	SequenceTimeout    time.Duration            // Max total time for all tools in sequence (e.g., 30s)
	PerToolTimeout     map[string]time.Duration // Per-tool overrides for specific tools
	OverheadBudget     time.Duration            // Estimated overhead per tool call (e.g., 500ms)
	CollectMetrics     bool                     // If true, collect execution metrics
	ExecutionMetrics   []ExecutionMetrics       // Collected metrics from last execution
}

// NewToolTimeoutConfig creates a timeout config with recommended defaults
func NewToolTimeoutConfig() *ToolTimeoutConfig {
	return &ToolTimeoutConfig{
		DefaultToolTimeout: 5 * time.Second,        // 5s per tool
		SequenceTimeout:    30 * time.Second,       // 30s total for all sequential tools
		OverheadBudget:     500 * time.Millisecond, // 500ms overhead for LLM calls and context switches
		PerToolTimeout:     make(map[string]time.Duration),
		CollectMetrics:     true,
		ExecutionMetrics:   []ExecutionMetrics{},
	}
}

// GetToolTimeout gets the timeout for a specific tool (checks per-tool overrides first)
func (tc *ToolTimeoutConfig) GetToolTimeout(toolName string) time.Duration {
	if timeout, exists := tc.PerToolTimeout[toolName]; exists {
		return timeout
	}
	return tc.DefaultToolTimeout
}

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
