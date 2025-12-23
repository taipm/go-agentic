package crewai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
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
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeTimeout          // Transient: exceeded deadline
	ErrorTypePanic            // Non-transient: panic in tool
	ErrorTypeValidation       // Non-transient: invalid arguments
	ErrorTypeNetwork          // Transient: connection issues
	ErrorTypeTemporary        // Transient: temporary failures
	ErrorTypePermanent        // Non-transient: permanent failure
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
// Starts at 100ms, doubles each attempt, capped at 5 seconds
func calculateBackoffDuration(attempt int) time.Duration {
	// Start with 100ms, double each attempt: 100ms, 200ms, 400ms, 800ms...
	baseDelay := time.Duration(100<<uint(attempt)) * time.Millisecond

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
	// Default to 2 retries (3 total attempts: 1 initial + 2 retries)
	// This is reasonable for transient failures without significant latency impact
	maxRetries := 2

	return retryWithBackoff(ctx, tool, args, maxRetries)
}

// ExecutionMetrics tracks execution time and status for tools
type ExecutionMetrics struct {
	ToolName     string        // Name of the tool executed
	Duration     time.Duration // Time taken to execute
	Status       string        // "success", "timeout", "error"
	TimedOut     bool          // True if tool execution exceeded timeout
	StartTime    time.Time     // When tool execution started
	EndTime      time.Time     // When tool execution completed
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
		return 100 * time.Millisecond // Minimal timeout to signal urgency
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
	warnThreshold := totalDuration / 5 // 20%
	return remaining < warnThreshold && remaining > 0
}

// ToolTimeoutConfig defines timeout behavior for tools
type ToolTimeoutConfig struct {
	DefaultToolTimeout  time.Duration            // Default timeout per tool (e.g., 5s)
	SequenceTimeout     time.Duration            // Max total time for all tools in sequence (e.g., 30s)
	PerToolTimeout      map[string]time.Duration // Per-tool overrides for specific tools
	OverheadBudget      time.Duration            // Estimated overhead per tool call (e.g., 500ms)
	CollectMetrics      bool                     // If true, collect execution metrics
	ExecutionMetrics    []ExecutionMetrics       // Collected metrics from last execution
}

// NewToolTimeoutConfig creates a timeout config with recommended defaults
func NewToolTimeoutConfig() *ToolTimeoutConfig {
	return &ToolTimeoutConfig{
		DefaultToolTimeout: 5 * time.Second,    // 5s per tool
		SequenceTimeout:    30 * time.Second,   // 30s total for all sequential tools
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
	history        []Message
	Verbose        bool               // If true, print agent responses and tool results to stdout
	ResumeAgentID  string             // If set, execution will start from this agent instead of entry agent
	ToolTimeouts   *ToolTimeoutConfig // Timeout configuration
	Metrics        *MetricsCollector  // Metrics collection for observability
	defaults       *HardcodedDefaults // Runtime configuration defaults
}

// NewCrewExecutor creates a new crew executor
// Note: crew.Routing MUST be set for signal-based routing to work
// Best Practice: Use entry_point from crew.yaml instead of relying on IsTerminal
func NewCrewExecutor(crew *Crew, apiKey string) *CrewExecutor {
	// Find entry agent - first agent is default if no routing configured
	var entryAgent *Agent
	if len(crew.Agents) > 0 {
		entryAgent = crew.Agents[0] // First agent is default entry point
	}

	return &CrewExecutor{
		crew:         crew,
		apiKey:       apiKey,
		entryAgent:   entryAgent,
		history:      []Message{},
		ToolTimeouts: NewToolTimeoutConfig(),
		Metrics:      NewMetricsCollector(),
		defaults:     DefaultHardcodedDefaults(),
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

// GetHistory returns a copy of the conversation history
// Allows inspection for debugging memory issues
func (ce *CrewExecutor) GetHistory() []Message {
	// Return a copy to prevent external modification
	historyCopy := make([]Message, len(ce.history))
	copy(historyCopy, ce.history)
	return historyCopy
}

// estimateHistoryTokens estimates total tokens in conversation history
// Uses approximation: 1 token ≈ 4 characters (OpenAI convention)
func (ce *CrewExecutor) estimateHistoryTokens() int {
	total := 0
	for _, msg := range ce.history {
		// Role overhead (~4 tokens) + content tokens
		total += 4 + (len(msg.Content)+3)/4
	}
	return total
}

// trimHistoryIfNeeded trims conversation history to fit within context window
// Uses ce.defaults.MaxContextWindow and ce.defaults.ContextTrimPercent
// Strategy: Keep first + recent messages, remove oldest in middle when over limit
func (ce *CrewExecutor) trimHistoryIfNeeded() {
	if ce.defaults == nil || len(ce.history) <= 2 {
		return
	}

	currentTokens := ce.estimateHistoryTokens()
	maxTokens := ce.defaults.MaxContextWindow

	// Check if within limit
	if currentTokens <= maxTokens {
		return
	}

	// Calculate target after trimming (remove ContextTrimPercent of max)
	trimPercent := ce.defaults.ContextTrimPercent / 100.0 // Convert from 20.0 to 0.20
	targetTokens := int(float64(maxTokens) * (1.0 - trimPercent))

	// Calculate how many messages to keep from end
	// Keep removing from middle until under target
	keepFromEnd := len(ce.history) - 1
	tokensFromEnd := 0

	for i := len(ce.history) - 1; i > 0 && tokensFromEnd < targetTokens; i-- {
		msgTokens := 4 + (len(ce.history[i].Content)+3)/4
		tokensFromEnd += msgTokens
		keepFromEnd = len(ce.history) - i
	}

	// Ensure we keep at least 2 messages from end
	if keepFromEnd < 2 {
		keepFromEnd = 2
	}

	// Build trimmed history: first message + summary + last N messages
	if len(ce.history) > keepFromEnd+1 {
		trimmedCount := len(ce.history) - keepFromEnd - 1

		newHistory := make([]Message, 0, keepFromEnd+2)

		// Keep first message
		newHistory = append(newHistory, ce.history[0])

		// Add summary for trimmed content
		newHistory = append(newHistory, Message{
			Role:    "system",
			Content: fmt.Sprintf("[%d earlier messages trimmed to fit context window]", trimmedCount),
		})

		// Keep last N messages
		startIdx := len(ce.history) - keepFromEnd
		newHistory = append(newHistory, ce.history[startIdx:]...)

		newTokens := 0
		for _, msg := range newHistory {
			newTokens += 4 + (len(msg.Content)+3)/4
		}

		log.Printf("[CONTEXT TRIM] %d→%d messages, ~%d→%d tokens (saved ~%d tokens)",
			len(ce.history), len(newHistory), currentTokens, newTokens, currentTokens-newTokens)

		ce.history = newHistory
	}
}

// ClearHistory clears the conversation history
// Useful for starting fresh conversations
func (ce *CrewExecutor) ClearHistory() {
	ce.history = []Message{}

	// ✅ Reset session cost tracking when starting fresh
	if ce.Metrics != nil {
		ce.Metrics.ResetSessionCost()
	}
}

// ExecuteStream runs the crew with streaming events
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input string, streamChan chan *StreamEvent) error {
	// Add user input to history
	ce.history = append(ce.history, Message{
		Role:    "user",
		Content: input,
	})

	// Determine starting agent: resume agent or entry agent
	var currentAgent *Agent
	if ce.ResumeAgentID != "" {
		// Resume from paused agent
		currentAgent = ce.findAgentByID(ce.ResumeAgentID)
		if currentAgent == nil {
			return fmt.Errorf("resume agent %s not found", ce.ResumeAgentID)
		}
		// Clear resume agent after using it
		ce.ResumeAgentID = ""
	} else {
		// Start with entry agent
		currentAgent = ce.entryAgent
		if currentAgent == nil {
			return fmt.Errorf("no entry agent found")
		}
	}

	handoffCount := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Send agent start event
		streamChan <- NewStreamEvent("agent_start", currentAgent.Name, fmt.Sprintf("🔄 Starting %s...", currentAgent.Name))

		// Trim history before LLM call to prevent cost leakage
		ce.trimHistoryIfNeeded()

		// Track agent execution time for metrics
		agentStartTime := time.Now()

		// Execute current agent
		response, err := ExecuteAgent(ctx, currentAgent, input, ce.history, ce.apiKey)
		agentEndTime := time.Now()
		agentDuration := agentEndTime.Sub(agentStartTime)

		if err != nil {
		// Update performance metrics with error
		if currentAgent.Metadata != nil {
		 currentAgent.UpdatePerformanceMetrics(false, err.Error())
		}

		// Check error quota (use different variable to avoid shadowing)
		 if quotaErr := currentAgent.CheckErrorQuota(); quotaErr != nil {
			log.Printf("[QUOTA] Agent %s exceeded error quota: %v", currentAgent.ID, quotaErr)
			streamChan <- NewStreamEvent("error", currentAgent.Name,
				fmt.Sprintf("Error quota exceeded: %v", quotaErr))
			return quotaErr
		}

		// Report agent error
		streamChan <- NewStreamEvent("error", currentAgent.Name, fmt.Sprintf("Agent failed: %v", err))
		// Record failed agent execution
		if ce.Metrics != nil {
			ce.Metrics.RecordAgentExecution(currentAgent.ID, currentAgent.Name, agentDuration, false)
		}
		return fmt.Errorf("agent %s failed: %w", currentAgent.ID, err)
	}

		// Record successful agent execution
		if ce.Metrics != nil {
			ce.Metrics.RecordAgentExecution(currentAgent.ID, currentAgent.Name, agentDuration, true)

			// Aggregate agent's last LLM call cost for crew-level tracking
			tokens, cost := currentAgent.GetLastCallCost()
			ce.Metrics.RecordLLMCall(currentAgent.ID, tokens, cost)
			ce.Metrics.LogCrewCostSummary()

			// Check memory quota after execution
			// Memory estimated based on token count: 1 token ≈ 4 bytes
			memoryUsedMB := (tokens * 4) / 1024 / 1024

			if err := currentAgent.CheckMemoryQuota(); err != nil {
				log.Printf("[QUOTA] Agent %s exceeded memory quota: %v", currentAgent.ID, err)
				streamChan <- NewStreamEvent("error", currentAgent.Name,
					fmt.Sprintf("Memory quota exceeded: %v", err))
				return err
			}

			// Update memory & performance metrics (only after quota check passes)
			callDurationMs := agentDuration.Milliseconds()
			if currentAgent.Metadata != nil {
				currentAgent.UpdateMemoryMetrics(memoryUsedMB, callDurationMs)
				currentAgent.UpdatePerformanceMetrics(true, "")
			}
		}

		// Send agent response event
		streamChan <- NewStreamEvent("agent_response", currentAgent.Name, response.Content)

		// Add agent response to history
		ce.history = append(ce.history, Message{
			Role:    "assistant",
			Content: response.Content,
		})

		// Execute any tool calls BEFORE checking if terminal
		if len(response.ToolCalls) > 0 {
			for _, toolCall := range response.ToolCalls {
				// Send tool start event
				streamChan <- NewStreamEvent("tool_start", currentAgent.Name,
					fmt.Sprintf("🔧 [Tool] %s → Executing...", toolCall.ToolName))
			}

			toolResults := ce.executeCalls(ctx, response.ToolCalls, currentAgent)

			// Send tool results
			for _, result := range toolResults {
				status := "✅"
				if result.Status == "error" {
					status = "❌"
				}

				// Note: Show full output including embedding vectors - agents need to extract vectors
				streamChan <- NewStreamEvent("tool_result", currentAgent.Name,
					fmt.Sprintf("%s [Tool] %s → %s", status, result.ToolName, result.Output))
			}

			// Format results for feedback
			resultText := ce.formatToolResults(toolResults)

			// Add results to history
			ce.history = append(ce.history, Message{
				Role:    "user",
				Content: resultText,
			})

			// Feed results back to current agent for analysis
			// This matches Non-Stream behavior - agent can analyze tool results before routing
			input = resultText
			continue
		}

		// Check for termination signals first (before routing)
		// If agent emits a termination signal like [KẾT THÚC THI], workflow should end
		terminationResult := ce.checkTerminationSignal(currentAgent, response.Content)
		if terminationResult != nil && terminationResult.ShouldTerminate {
			streamChan <- NewStreamEvent("terminate", currentAgent.Name,
				fmt.Sprintf("[TERMINATE] Workflow ended by signal: %s", terminationResult.Signal))
			return nil
		}

		// Check for routing signals from current agent (config-driven)
		// This is checked AFTER termination because if agent emits a routing signal,
		// it means it wants to hand off to another agent, not pause
		nextAgent := ce.findNextAgentBySignal(currentAgent, response.Content)
		if nextAgent != nil {
			currentAgent = nextAgent
			input = response.Content
			handoffCount++
			continue
		}

		// Check wait_for_signal BEFORE terminal check
		// This allows terminal agents (like Creator) to pause for user confirmation
		// before actually finishing. If no routing signal was found AND agent has
		// wait_for_signal enabled, pause and wait for user input.
		behavior := ce.getAgentBehavior(currentAgent.ID)
		if behavior != nil && behavior.WaitForSignal {
			// Agent waits for explicit signal, send pause event with agent ID for resume
			// Format: [PAUSE:agent_id] allows CLI to know which agent to resume from
			streamChan <- NewStreamEvent("pause", currentAgent.Name,
				fmt.Sprintf("[PAUSE:%s] Waiting for user input", currentAgent.ID))
			return nil
		}

		// Check if current agent is terminal (only after tool execution and wait_for_signal)
		if currentAgent.IsTerminal {
			return nil
		}

		// Check for parallel group execution
		parallelGroup := ce.findParallelGroup(currentAgent.ID, response.Content)
		if parallelGroup != nil {
			// Get the agents for this parallel group
			var parallelAgents []*Agent
			agentMap := make(map[string]*Agent)
			for _, agent := range ce.crew.Agents {
				agentMap[agent.ID] = agent
			}

			for _, agentID := range parallelGroup.Agents {
				if agent, exists := agentMap[agentID]; exists {
					parallelAgents = append(parallelAgents, agent)
				}
			}

			if len(parallelAgents) > 0 {
				// Execute all parallel agents
				// IMPORTANT: Pass 'input' (which has tool results + embedding vector), not 'response.Content' (which only has text)
				parallelResults, err := ce.ExecuteParallelStream(ctx, input, parallelAgents, streamChan)
				if err != nil {
					streamChan <- NewStreamEvent("error", "system", fmt.Sprintf("Parallel execution failed: %v", err))
					return err
				}

				// Aggregate results
				aggregatedInput := ce.aggregateParallelResults(parallelResults)

				// Add aggregated results to history
				ce.history = append(ce.history, Message{
					Role:    "user",
					Content: aggregatedInput,
				})

				// Move to next agent in the pipeline
				if parallelGroup.NextAgent != "" {
					if nextAgent, exists := agentMap[parallelGroup.NextAgent]; exists {
						currentAgent = nextAgent
						input = aggregatedInput
						handoffCount++
						continue
					}
				}
			}
		}

		// For other agents, handoff normally
		handoffCount++
		if handoffCount >= ce.crew.MaxHandoffs {
			return nil
		}

		// Find next agent based on handoff targets
		nextAgent = ce.findNextAgent(currentAgent)
		if nextAgent == nil {
			return nil
		}

		currentAgent = nextAgent
		input = response.Content
	}
}

// Execute runs the crew with the given input
func (ce *CrewExecutor) Execute(ctx context.Context, input string) (*CrewResponse, error) {
	// Add user input to history
	ce.history = append(ce.history, Message{
		Role:    "user",
		Content: input,
	})

	// Start with entry agent
	currentAgent := ce.entryAgent
	if currentAgent == nil {
		return nil, fmt.Errorf("no entry agent found")
	}

	handoffCount := 0

	for {
		// Trim history before LLM call to prevent cost leakage
		ce.trimHistoryIfNeeded()

		// Execute current agent
		log.Printf("[AGENT START] %s (%s)", currentAgent.Name, currentAgent.ID)
		response, err := ExecuteAgent(ctx, currentAgent, input, ce.history, ce.apiKey)
		if err != nil {
			log.Printf("[AGENT ERROR] %s (%s) - %v", currentAgent.Name, currentAgent.ID, err)
			return nil, fmt.Errorf("agent %s failed: %w", currentAgent.ID, err)
		}
		log.Printf("[AGENT END] %s (%s) - Success", currentAgent.Name, currentAgent.ID)

		// Aggregate agent's last LLM call cost for crew-level tracking
		if ce.Metrics != nil {
			tokens, cost := currentAgent.GetLastCallCost()
			ce.Metrics.RecordLLMCall(currentAgent.ID, tokens, cost)
			ce.Metrics.LogCrewCostSummary()
		}

		if ce.Verbose {
			fmt.Printf("\n[%s]: %s\n", currentAgent.Name, response.Content)
		}

		// Add agent response to history
		ce.history = append(ce.history, Message{
			Role:    "assistant",
			Content: response.Content,
		})

		// Execute any tool calls BEFORE checking if terminal
		if len(response.ToolCalls) > 0 {
			toolResults := ce.executeCalls(ctx, response.ToolCalls, currentAgent)

			// Format results for feedback
			resultText := ce.formatToolResults(toolResults)
			if ce.Verbose {
				fmt.Println(resultText)
			}

			// Add results to history
			ce.history = append(ce.history, Message{
				Role:    "user",
				Content: resultText,
			})

			// Feed results back to current agent for analysis
			input = resultText
			// Continue loop to let agent process results
			continue
		}

		// Check for termination signals first (before routing)
		// If agent emits a termination signal like [KẾT THÚC THI], workflow should end
		terminationResult := ce.checkTerminationSignal(currentAgent, response.Content)
		if terminationResult != nil && terminationResult.ShouldTerminate {
			log.Printf("[TERMINATE] Workflow ended by signal: %s", terminationResult.Signal)
			return &CrewResponse{
				AgentID:    currentAgent.ID,
				AgentName:  currentAgent.Name,
				Content:    response.Content,
				IsTerminal: true,
			}, nil
		}

		// Check for routing signals from current agent (config-driven)
		// This is checked AFTER termination because if agent emits a routing signal,
		// it means it wants to hand off to another agent, not pause
		nextAgent := ce.findNextAgentBySignal(currentAgent, response.Content)
		if nextAgent != nil {
			currentAgent = nextAgent
			input = response.Content
			handoffCount++
			continue
		}

		// Check wait_for_signal BEFORE terminal check
		// This allows terminal agents (like Creator) to pause for user confirmation
		behavior := ce.getAgentBehavior(currentAgent.ID)
		if behavior != nil && behavior.WaitForSignal {
			// Agent waits for explicit signal, return and wait for next input
			return &CrewResponse{
				AgentID:       currentAgent.ID,
				AgentName:     currentAgent.Name,
				Content:       response.Content,
				ToolCalls:     response.ToolCalls,
				PausedAgentID: currentAgent.ID, // Include paused agent ID for resume
			}, nil
		}

		// Check if current agent is terminal (only after tool execution and wait_for_signal)
		if currentAgent.IsTerminal {
			return &CrewResponse{
				AgentID:    currentAgent.ID,
				AgentName:  currentAgent.Name,
				Content:    response.Content,
				ToolCalls:  response.ToolCalls,
				IsTerminal: true,
			}, nil
		}

		// Check for parallel group execution
		parallelGroup := ce.findParallelGroup(currentAgent.ID, response.Content)
		if parallelGroup != nil {
			// Get the agents for this parallel group
			var parallelAgents []*Agent
			agentMap := make(map[string]*Agent)
			for _, agent := range ce.crew.Agents {
				agentMap[agent.ID] = agent
			}

			for _, agentID := range parallelGroup.Agents {
				if agent, exists := agentMap[agentID]; exists {
					parallelAgents = append(parallelAgents, agent)
				}
			}

			if len(parallelAgents) > 0 {
				// Execute all parallel agents
				parallelResults, err := ce.ExecuteParallel(ctx, input, parallelAgents)
				if err != nil {
					return nil, fmt.Errorf("parallel execution failed: %w", err)
				}

				// Aggregate results
				aggregatedInput := ce.aggregateParallelResults(parallelResults)

				// Add aggregated results to history
				ce.history = append(ce.history, Message{
					Role:    "user",
					Content: aggregatedInput,
				})

				// Move to next agent in the pipeline
				if parallelGroup.NextAgent != "" {
					if nextAgent, exists := agentMap[parallelGroup.NextAgent]; exists {
						currentAgent = nextAgent
						input = aggregatedInput
						handoffCount++
						continue
					}
				}
			}
		}

		// For other agents, handoff normally
		handoffCount++
		if handoffCount >= ce.crew.MaxHandoffs {
			return &CrewResponse{
				AgentID:   currentAgent.ID,
				AgentName: currentAgent.Name,
				Content:   response.Content,
				ToolCalls: response.ToolCalls,
			}, nil
		}

		// Find next agent based on handoff targets
		nextAgent = ce.findNextAgent(currentAgent)
		if nextAgent == nil {
			return &CrewResponse{
				AgentID:   currentAgent.ID,
				AgentName: currentAgent.Name,
				Content:   response.Content,
				ToolCalls: response.ToolCalls,
			}, nil
		}

		currentAgent = nextAgent
		input = response.Content
	}
}
