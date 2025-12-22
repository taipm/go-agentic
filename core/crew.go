package crewai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// copyHistory creates a deep copy of message history to ensure thread safety
// Each execution gets its own isolated history snapshot, preventing race conditions
// when concurrent requests execute and pause/resume
func copyHistory(original []Message) []Message {
	if len(original) == 0 {
		return []Message{}
	}
	// Create new slice with same capacity
	copied := make([]Message, len(original))
	// Copy all messages
	copy(copied, original)
	return copied
}

// extractRequiredFields extracts required field names from tool parameters
func extractRequiredFields(params map[string]interface{}) []string {
	var requiredFields []string
	if required, ok := params["required"].([]interface{}); ok {
		for _, r := range required {
			if rStr, ok := r.(string); ok {
				requiredFields = append(requiredFields, rStr)
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
		if _, ok := fieldValue.(string); !ok {
			return fmt.Errorf("tool '%s': parameter '%s' should be string, got %T", tool.Name, fieldName, fieldValue)
		}
	case "number", "integer":
		switch fieldValue.(type) {
		case float64, int, int64:
			// Valid number types
		default:
			return fmt.Errorf("tool '%s': parameter '%s' should be number, got %T", tool.Name, fieldName, fieldValue)
		}
	}
	return nil
}

// validateToolArguments validates tool arguments against tool definition
// ✅ FIX for Issue #25: Tool execution validation
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
// ✅ FIX for Issue #5 (Panic Risk): Catch any panic in tool execution and convert to error
// ✅ FIX for Issue #25: Validate arguments before execution
// This prevents one buggy tool from crashing the entire server
// Pattern: defer-recover catches panic and converts it to error (Go standard approach)
// ✅ FIX #5: Error classification for smart recovery decisions
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
// ✅ FIX #5: Helper function for error recovery strategy
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
// ✅ FIX #5: Helper function to determine retry strategy
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
// ✅ FIX #5: Helper function for exponential backoff calculation
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
// ✅ FIX #5: Main retry execution function
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
// ✅ FIX #5: Single execution attempt (used by retry wrapper)
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
// ✅ FIX #5: Enhanced with retry logic and error recovery
func safeExecuteTool(ctx context.Context, tool *Tool, args map[string]interface{}) (output string, err error) {
	// Default to 2 retries (3 total attempts: 1 initial + 2 retries)
	// This is reasonable for transient failures without significant latency impact
	maxRetries := 2

	return retryWithBackoff(ctx, tool, args, maxRetries)
}

// ✅ FIX for Issue #11 (Sequential Tool Timeout)
// ExecutionMetrics tracks execution time and status for tools
type ExecutionMetrics struct {
	ToolName     string        // Name of the tool executed
	Duration     time.Duration // Time taken to execute
	Status       string        // "success", "timeout", "error"
	TimedOut     bool          // True if tool execution exceeded timeout
	StartTime    time.Time     // When tool execution started
	EndTime      time.Time     // When tool execution completed
}

// ✅ FIX for Issue #11 (Enhanced Timeout Management)
// TimeoutTracker tracks sequence execution time and manages per-tool budgets
type TimeoutTracker struct {
	sequenceStartTime time.Time     // When sequence started
	sequenceDeadline  time.Time     // When sequence must complete
	overheadBudget    time.Duration // Estimated overhead per tool (e.g., 500ms for LLM calls)
	usedTime          time.Duration // Time already consumed in sequence
	mu                sync.Mutex    // Protect concurrent access
}

// NewTimeoutTracker creates a timeout tracker for a sequence
func NewTimeoutTracker(sequenceTimeout time.Duration, overheadBudget time.Duration) *TimeoutTracker {
	now := time.Now()
	return &TimeoutTracker{
		sequenceStartTime: now,
		sequenceDeadline:  now.Add(sequenceTimeout),
		overheadBudget:    overheadBudget,
		usedTime:          0,
	}
}

// GetRemainingTime returns how much time is left in the sequence
func (tt *TimeoutTracker) GetRemainingTime() time.Duration {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	remaining := time.Until(tt.sequenceDeadline)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CalculateToolTimeout calculates the appropriate timeout for the next tool
// accounting for: per-tool timeout, remaining sequence time, and overhead budget
func (tt *TimeoutTracker) CalculateToolTimeout(defaultTimeout, perToolTimeout time.Duration) time.Duration {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	// Start with per-tool timeout, fallback to default
	toolTimeout := perToolTimeout
	if toolTimeout <= 0 {
		toolTimeout = defaultTimeout
	}

	// Get remaining sequence time and subtract overhead budget
	remaining := time.Until(tt.sequenceDeadline)
	if remaining <= tt.overheadBudget {
		// Not enough time even for overhead
		return 100 * time.Millisecond // Minimal timeout to signal urgency
	}

	// Available time is remaining minus overhead
	availableTime := remaining - tt.overheadBudget

	// Use the minimum: per-tool timeout or available time
	if toolTimeout > availableTime {
		return availableTime
	}
	return toolTimeout
}

// RecordToolExecution records that a tool has finished and updates used time
func (tt *TimeoutTracker) RecordToolExecution(duration time.Duration) {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	tt.usedTime += duration
}

// IsTimeoutWarning returns true if we're within 20% of sequence deadline
func (tt *TimeoutTracker) IsTimeoutWarning() bool {
	tt.mu.Lock()
	defer tt.mu.Unlock()
	remaining := time.Until(tt.sequenceDeadline)
	totalDuration := tt.sequenceDeadline.Sub(tt.sequenceStartTime)
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
	Verbose        bool   // If true, print agent responses and tool results to stdout
	ResumeAgentID  string // If set, execution will start from this agent instead of entry agent
	ToolTimeouts   *ToolTimeoutConfig // ✅ FIX for Issue #11: Timeout configuration
	Metrics        *MetricsCollector  // ✅ FIX for Issue #14: Metrics collection for observability
}

// NewCrewExecutor creates a new crew executor
// Note: crew.Routing MUST be set for signal-based routing to work
func NewCrewExecutor(crew *Crew, apiKey string) *CrewExecutor {
	// Find entry agent (first agent that's not terminal)
	var entryAgent *Agent
	for _, agent := range crew.Agents {
		if !agent.IsTerminal {
			entryAgent = agent
			break
		}
	}

	return &CrewExecutor{
		crew:         crew,
		apiKey:       apiKey,
		entryAgent:   entryAgent,
		history:      []Message{},
		ToolTimeouts: NewToolTimeoutConfig(), // ✅ FIX for Issue #11: Initialize timeout config
		Metrics:      NewMetricsCollector(),  // ✅ FIX for Issue #14: Initialize metrics collector
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
	agentConfigs, err := LoadAgentConfigs(agentDir)
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

	// Create and return executor
	return NewCrewExecutor(crew, apiKey), nil
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

		// ✅ FIX for Issue #14: Start tracking agent execution time
		agentStartTime := time.Now()

		// Execute current agent
		response, err := ExecuteAgent(ctx, currentAgent, input, ce.history, ce.apiKey)
		agentEndTime := time.Now()
		agentDuration := agentEndTime.Sub(agentStartTime)

		if err != nil {
			streamChan <- NewStreamEvent("error", currentAgent.Name, fmt.Sprintf("Agent failed: %v", err))
			// ✅ FIX for Issue #14: Record failed agent execution
			if ce.Metrics != nil {
				ce.Metrics.RecordAgentExecution(currentAgent.ID, currentAgent.Name, agentDuration, false)
			}
			return fmt.Errorf("agent %s failed: %w", currentAgent.ID, err)
		}

		// ✅ FIX for Issue #14: Record successful agent execution
		if ce.Metrics != nil {
			ce.Metrics.RecordAgentExecution(currentAgent.ID, currentAgent.Name, agentDuration, true)
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
			resultText := formatToolResults(toolResults)

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

		// Check for routing signals from current agent (config-driven)
		// This is checked FIRST because if agent emits a routing signal,
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
		// Execute current agent
		log.Printf("[AGENT START] %s (%s)", currentAgent.Name, currentAgent.ID)
		response, err := ExecuteAgent(ctx, currentAgent, input, ce.history, ce.apiKey)
		if err != nil {
			log.Printf("[AGENT ERROR] %s (%s) - %v", currentAgent.Name, currentAgent.ID, err)
			return nil, fmt.Errorf("agent %s failed: %w", currentAgent.ID, err)
		}
		log.Printf("[AGENT END] %s (%s) - Success", currentAgent.Name, currentAgent.ID)

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
			resultText := formatToolResults(toolResults)
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

		// Check for routing signals from current agent (config-driven)
		// This is checked FIRST because if agent emits a routing signal,
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

// calculateToolTimeout determines the appropriate timeout for the next tool
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) calculateToolTimeout(tracker *TimeoutTracker, toolName string) time.Duration {
	if tracker != nil && ce.ToolTimeouts != nil {
		perToolTimeout := ce.ToolTimeouts.GetToolTimeout(toolName)
		return tracker.CalculateToolTimeout(ce.ToolTimeouts.DefaultToolTimeout, perToolTimeout)
	}
	if ce.ToolTimeouts != nil {
		return ce.ToolTimeouts.GetToolTimeout(toolName)
	}
	return 5 * time.Second
}

// logToolStart logs tool execution start with timeout details
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) logToolStart(tool *Tool, agent *Agent, timeout time.Duration, tracker *TimeoutTracker) {
	if tracker != nil {
		remaining := tracker.GetRemainingTime()
		log.Printf("[TOOL START] %s <- %s (timeout: %v, remaining: %v)", tool.Name, agent.ID, timeout, remaining)
	} else {
		log.Printf("[TOOL START] %s <- %s (timeout: %v)", tool.Name, agent.ID, timeout)
	}
}

// recordToolMetrics records execution metrics and detects timeouts
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) recordToolMetrics(tool *Tool, duration time.Duration, err error, startTime, endTime time.Time) {
	timedOut := err != nil && errors.Is(err, context.DeadlineExceeded)
	if ce.ToolTimeouts != nil && ce.ToolTimeouts.CollectMetrics {
		status := ce.getToolExecutionStatus(timedOut, err)
		ce.ToolTimeouts.ExecutionMetrics = append(ce.ToolTimeouts.ExecutionMetrics, ExecutionMetrics{
			ToolName:  tool.Name,
			Duration:  duration,
			Status:    status,
			TimedOut:  timedOut,
			StartTime: startTime,
			EndTime:   endTime,
		})
	}
	if ce.Metrics != nil {
		success := err == nil
		ce.Metrics.RecordToolExecution(tool.Name, duration, success)
	}
}

// getToolExecutionStatus determines the status based on error type
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) getToolExecutionStatus(timedOut bool, err error) string {
	if timedOut {
		return "timeout"
	}
	if err != nil {
		return "error"
	}
	return "success"
}

// handleToolNotFound logs and records tool not found error
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) handleToolNotFound(call ToolCall) ToolResult {
	log.Printf("[TOOL ERROR] %s - Tool not found", call.ToolName)
	return ToolResult{
		ToolName: call.ToolName,
		Status:   "error",
		Output:   fmt.Sprintf("Tool %s not found", call.ToolName),
	}
}

// handleSequenceTimeout logs and returns timeout result
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) handleSequenceTimeout(tool *Tool, agent *Agent) ToolResult {
	log.Printf("[TOOL TIMEOUT] %s <- %s - Sequence timeout exceeded", tool.Name, agent.ID)
	return ToolResult{
		ToolName: tool.Name,
		Status:   "error",
		Output:   "Tool execution timeout: sequence timeout exceeded",
	}
}

// handleToolExecutionError logs and returns tool execution error
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) handleToolExecutionError(tool *Tool, err error, duration time.Duration, timedOut bool) ToolResult {
	if timedOut {
		log.Printf("[TOOL TIMEOUT] %s - %v (%v)", tool.Name, err, duration)
	} else {
		log.Printf("[TOOL ERROR] %s - %v", tool.Name, err)
	}
	return ToolResult{
		ToolName: tool.Name,
		Status:   "error",
		Output:   err.Error(),
	}
}

// handleToolExecutionSuccess logs and returns tool execution success
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) handleToolExecutionSuccess(tool *Tool, duration time.Duration, output string) ToolResult {
	log.Printf("[TOOL SUCCESS] %s -> %d chars (%v)", tool.Name, len(output), duration)
	return ToolResult{
		ToolName: tool.Name,
		Status:   "success",
		Output:   output,
	}
}

// setupSequenceContext creates and configures the sequence execution context
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) setupSequenceContext(ctx context.Context) (context.Context, context.CancelFunc, *TimeoutTracker) {
	if ctx == nil {
		ctx = context.Background()
	}

	var sequenceCtx context.Context
	var sequenceCancel context.CancelFunc
	var timeoutTracker *TimeoutTracker

	if ce.ToolTimeouts != nil && ce.ToolTimeouts.SequenceTimeout > 0 {
		timeoutTracker = NewTimeoutTracker(ce.ToolTimeouts.SequenceTimeout, ce.ToolTimeouts.OverheadBudget)
		sequenceCtx, sequenceCancel = context.WithTimeout(ctx, ce.ToolTimeouts.SequenceTimeout)
	} else {
		// No sequence timeout configured, use context as-is
		sequenceCtx = ctx
		sequenceCancel = func() {} // no-op cancel for defer safety
	}

	return sequenceCtx, sequenceCancel, timeoutTracker
}

// executeCalls executes tool calls from an agent with per-tool and sequence timeouts
// ✅ FIX for Issue #11 (Sequential Tool Timeout): Add timeout protection for hanging tools
// ✅ FIX #4: Enhanced timeout management with remaining time calculation and overhead tracking
func (ce *CrewExecutor) executeCalls(ctx context.Context, calls []ToolCall, agent *Agent) []ToolResult {
	var results []ToolResult

	toolMap := make(map[string]*Tool)
	for _, tool := range agent.Tools {
		toolMap[tool.Name] = tool
	}

	sequenceCtx, sequenceCancel, timeoutTracker := ce.setupSequenceContext(ctx)
	defer sequenceCancel()

	// Reset metrics for this execution
	if ce.ToolTimeouts != nil && ce.ToolTimeouts.CollectMetrics {
		ce.ToolTimeouts.ExecutionMetrics = []ExecutionMetrics{}
	}

	for _, call := range calls {
		tool, ok := toolMap[call.ToolName]
		if !ok {
			results = append(results, ce.handleToolNotFound(call))
			continue
		}

		// ✅ FIX for Issue #11: Check sequence deadline before executing tool
		select {
		case <-sequenceCtx.Done():
			results = append(results, ce.handleSequenceTimeout(tool, agent))
			return results
		default:
		}

		// ✅ FIX #4: Calculate timeout with remaining sequence time
		toolTimeout := ce.calculateToolTimeout(timeoutTracker, tool.Name)
		ce.logToolStart(tool, agent, toolTimeout, timeoutTracker)

		toolCtx, toolCancel := context.WithTimeout(sequenceCtx, toolTimeout)
		startTime := time.Now()
		output, err := safeExecuteTool(toolCtx, tool, call.Arguments)
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		toolCancel()

		// ✅ FIX #4: Record tool execution in timeout tracker
		if timeoutTracker != nil {
			timeoutTracker.RecordToolExecution(duration)
		}

		// ✅ FIX #4: Record metrics
		ce.recordToolMetrics(tool, duration, err, startTime, endTime)

		// ✅ FIX #4: Check if approaching timeout
		if timeoutTracker != nil && timeoutTracker.IsTimeoutWarning() {
			remaining := timeoutTracker.GetRemainingTime()
			log.Printf("[TIMEOUT WARNING] Sequence timeout approaching - only %v remaining", remaining)
		}

		// Handle tool execution result
		timedOut := err != nil && errors.Is(err, context.DeadlineExceeded)
		if err != nil {
			results = append(results, ce.handleToolExecutionError(tool, err, duration, timedOut))
		} else {
			results = append(results, ce.handleToolExecutionSuccess(tool, duration, output))
		}

		// Note: Not hiding embedding vectors anymore - agents need to see vectors to extract and use them
		// Verbose output is handled by the caller, not here
	}

	return results
}

// findAgentByID finds an agent by its ID
func (ce *CrewExecutor) findAgentByID(id string) *Agent {
	for _, agent := range ce.crew.Agents {
		if agent.ID == id {
			return agent
		}
	}
	return nil
}

// signalMatchesContent checks if a signal appears in response content (handles variations)
// ✅ FIX for signal matching: Handles "[ KẾT THÚC ]" matching "[KẾT THÚC]"
func signalMatchesContent(signal, content string) bool {
	// Exact match
	if strings.Contains(content, signal) {
		return true
	}

	// Normalized match (trim whitespace)
	normalizedSignal := strings.TrimSpace(signal)
	if strings.Contains(content, normalizedSignal) {
		return true
	}

	// Handle bracket variations like "[ SIGNAL ]" vs "[SIGNAL]"
	if strings.HasPrefix(signal, "[") && strings.HasSuffix(signal, "]") {
		innerSignal := strings.TrimPrefix(strings.TrimSuffix(signal, "]"), "[")
		innerSignal = strings.TrimSpace(innerSignal)
		// Check if inner signal appears in brackets (with any spacing)
		patterns := []string{
			"[" + innerSignal + "]",
			"[ " + innerSignal + " ]",
			"[  " + innerSignal + "  ]",
		}
		for _, pattern := range patterns {
			if strings.Contains(content, pattern) {
				return true
			}
		}
	}

	return false
}

// findNextAgentBySignal finds the next agent based on routing signals (config-driven)
func (ce *CrewExecutor) findNextAgentBySignal(current *Agent, responseContent string) *Agent {
	if ce.crew.Routing == nil {
		return nil
	}

	// Get signals defined for current agent in config
	signals, exists := ce.crew.Routing.Signals[current.ID]
	if !exists || len(signals) == 0 {
		return nil
	}

	// Check which signal is present in the response
	for _, sig := range signals {
		if sig.Target == "" {
			continue // Skip signals without target
		}

		// Check if signal matches response content
		if signalMatchesContent(sig.Signal, responseContent) {
			// Found matching signal, find the target agent
			nextAgent := ce.findAgentByID(sig.Target)
			if nextAgent != nil {
				log.Printf("[ROUTING] %s -> %s (signal: %s)", current.ID, nextAgent.ID, sig.Signal)
			}
			return nextAgent
		}
	}

	return nil
}

// getAgentBehavior retrieves behavior config for an agent
func (ce *CrewExecutor) getAgentBehavior(agentID string) *AgentBehavior {
	if ce.crew.Routing == nil || ce.crew.Routing.AgentBehaviors == nil {
		return nil
	}
	behavior, exists := ce.crew.Routing.AgentBehaviors[agentID]
	if !exists {
		return nil
	}
	return &behavior
}

// findNextAgent finds the next appropriate agent for handoff
func (ce *CrewExecutor) findNextAgent(current *Agent) *Agent {
	// First, try to use handoff_targets from current agent config
	if len(current.HandoffTargets) > 0 {
		// Create a map of agents by ID for quick lookup
		agentMap := make(map[string]*Agent)
		for _, agent := range ce.crew.Agents {
			agentMap[agent.ID] = agent
		}

		// Try to find the first available handoff target
		for _, targetID := range current.HandoffTargets {
			if agent, exists := agentMap[targetID]; exists && agent.ID != current.ID {
				log.Printf("[ROUTING] %s -> %s (handoff_targets)", current.ID, agent.ID)
				return agent
			}
		}
	}

	// Fallback: Find any other agent (not terminal-only strategy)
	for _, agent := range ce.crew.Agents {
		if agent.ID != current.ID {
			log.Printf("[ROUTING] %s -> %s (fallback)", current.ID, agent.ID)
			return agent
		}
	}

	log.Printf("[ROUTING] No next agent found for %s", current.ID)
	return nil
}

// ParallelAgentTimeout is the default timeout for each parallel agent execution
const ParallelAgentTimeout = 60 * time.Second

// ExecuteParallelStream executes multiple agents in parallel and collects their results
// Used for parallel execution of agents within a parallel group
func (ce *CrewExecutor) ExecuteParallelStream(
	ctx context.Context,
	input string,
	agents []*Agent,
	streamChan chan *StreamEvent,
) (map[string]*AgentResponse, error) {

	// Create a WaitGroup for synchronization
	var wg sync.WaitGroup
	resultMap := make(map[string]*AgentResponse)
	resultChan := make(chan *AgentResponse, len(agents))
	errorChan := make(chan error, len(agents))
	mu := sync.Mutex{}

	// Launch all agents in parallel using goroutines
	for _, agent := range agents {
		wg.Add(1)
		go func(ag *Agent) {
			defer wg.Done()

			// Create timeout context for this agent
			agentCtx, cancel := context.WithTimeout(ctx, ParallelAgentTimeout)
			defer cancel()

			// Send agent start event
			streamChan <- NewStreamEvent("agent_start", ag.Name,
				fmt.Sprintf("🔄 [Parallel] %s starting...", ag.Name))

			// Execute the agent with timeout
			response, err := ExecuteAgent(agentCtx, ag, input, ce.history, ce.apiKey)
			if err != nil {
				streamChan <- NewStreamEvent("error", ag.Name,
					fmt.Sprintf("❌ Agent failed: %v", err))
				errorChan <- fmt.Errorf("agent %s failed: %w", ag.ID, err)
				return
			}

			// Send agent response event
			streamChan <- NewStreamEvent("agent_response", ag.Name, response.Content)

			// Execute tool calls if any
			if len(response.ToolCalls) > 0 {
				for _, toolCall := range response.ToolCalls {
					streamChan <- NewStreamEvent("tool_start", ag.Name,
						fmt.Sprintf("🔧 [Tool] %s → Executing...", toolCall.ToolName))
				}

				toolResults := ce.executeCalls(ctx, response.ToolCalls, ag)

				for _, result := range toolResults {
					status := "✅"
					if result.Status == "error" {
						status = "❌"
					}
					streamChan <- NewStreamEvent("tool_result", ag.Name,
						fmt.Sprintf("%s [Tool] %s → %s", status, result.ToolName, result.Output))
				}
			}

			resultChan <- response
		}(agent)
	}

	// Wait for all agents to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	for result := range resultChan {
		mu.Lock()
		resultMap[result.AgentID] = result
		mu.Unlock()
	}

	// Check for errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	// Return partial results if some agents succeeded
	if len(resultMap) > 0 {
		if len(errors) > 0 {
			streamChan <- NewStreamEvent("warning", "system",
				fmt.Sprintf("⚠️ %d agents failed, continuing with %d results",
					len(errors), len(resultMap)))
		}
		return resultMap, nil
	}

	// All agents failed
	if len(errors) > 0 {
		return nil, fmt.Errorf("parallel execution failed: %v", errors[0])
	}

	return resultMap, nil
}

// ExecuteParallel executes multiple agents in parallel for Non-Stream mode
// Uses errgroup for automatic context propagation and goroutine cleanup
// If any goroutine errors, all others are cancelled automatically
func (ce *CrewExecutor) ExecuteParallel(
	ctx context.Context,
	input string,
	agents []*Agent,
) (map[string]*AgentResponse, error) {

	// ✅ FIX for Issue #3 (Goroutine Leak): Use errgroup for automatic context propagation
	// Create errgroup with context cancellation support
	// If any goroutine errors, all others are cancelled automatically
	g, gctx := errgroup.WithContext(ctx)

	// Thread-safe result map
	resultMap := make(map[string]*AgentResponse)
	resultMutex := sync.Mutex{}

	// Launch all agents in parallel
	for _, agent := range agents {
		ag := agent  // Capture for closure (important!)

		g.Go(func() error {
			if ce.Verbose {
				fmt.Printf("\n🔄 [Parallel] %s starting...\n", ag.Name)
			}

			// Create timeout context for this agent
			// gctx automatically propagates cancellation from parent or if another goroutine errors
			agentCtx, cancel := context.WithTimeout(gctx, ParallelAgentTimeout)
			defer cancel()

			// Execute the agent with timeout
			// If agentCtx is cancelled, ExecuteAgent should return immediately
			response, err := ExecuteAgent(agentCtx, ag, input, ce.history, ce.apiKey)
			if err != nil {
				if ce.Verbose {
					fmt.Printf("❌ [Parallel] %s failed: %v\n", ag.Name, err)
				}
				// Return error - this will cancel all other goroutines automatically
				return fmt.Errorf("agent %s failed: %w", ag.ID, err)
			}

			if ce.Verbose {
				fmt.Printf("\n[%s]: %s\n", ag.Name, response.Content)
			}

			// Execute tool calls if any
			if len(response.ToolCalls) > 0 {
				// Pass agentCtx to executeCalls for proper cancellation support
				toolResults := ce.executeCalls(agentCtx, response.ToolCalls, ag)

				if ce.Verbose {
					resultText := formatToolResults(toolResults)
					fmt.Println(resultText)
				}
			}

			// Store result thread-safely
			resultMutex.Lock()
			resultMap[response.AgentID] = response
			resultMutex.Unlock()

			return nil  // ✅ Goroutine completes, cleaned up automatically
		})
	}

	// Wait for all goroutines to complete
	// Automatically cancels remaining goroutines if any error occurs
	// Guaranteed cleanup: no goroutines left behind
	err := g.Wait()

	// Return results even if some agents failed (graceful degradation)
	if len(resultMap) > 0 {
		if err != nil && ce.Verbose {
			// Some agents failed, but we have partial results
			fmt.Printf("⚠️ Parallel execution had errors, but returning %d results\n",
				len(resultMap))
		}
		return resultMap, nil
	}

	// All agents failed
	if err != nil {
		return nil, fmt.Errorf("parallel execution failed: %w", err)
	}

	// Should not reach here (if all agents fail, err != nil from g.Wait())
	return nil, fmt.Errorf("parallel execution produced no results")
}

// findParallelGroup finds a parallel group configuration for the given agent
// Returns the parallel group if the agent's signal matches a parallel group target
func (ce *CrewExecutor) findParallelGroup(agentID string, signalContent string) *ParallelGroupConfig {
	if ce.crew.Routing == nil || ce.crew.Routing.ParallelGroups == nil {
		return nil
	}

	// Check if this agent emits a signal that targets a parallel group
	if signals, exists := ce.crew.Routing.Signals[agentID]; exists {
		for _, signal := range signals {
			// Check if the agent's response contains the signal
			if strings.Contains(signalContent, signal.Signal) {
				// Check if this signal targets a parallel group
				if parallelGroup, exists := ce.crew.Routing.ParallelGroups[signal.Target]; exists {
					return &parallelGroup
				}
			}
		}
	}

	return nil
}

// aggregateParallelResults combines results from multiple parallel agents into a single input
func (ce *CrewExecutor) aggregateParallelResults(results map[string]*AgentResponse) string {
	var sb strings.Builder

	sb.WriteString("\n[📊 PARALLEL EXECUTION RESULTS]\n\n")

	for agentID, result := range results {
		sb.WriteString(fmt.Sprintf("[%s]\n", agentID))
		sb.WriteString(fmt.Sprintf("%s\n\n", result.Content))
	}

	sb.WriteString("[END PARALLEL RESULTS]\n")

	return sb.String()
}

// ToolResult represents the result of executing a tool
type ToolResult struct {
	ToolName string
	Status   string
	Output   string
}

// formatToolResults formats tool results for agent feedback
func formatToolResults(results []ToolResult) string {
	const maxOutputChars = 2000 // Maximum characters per tool output to prevent context overflow

	var sb strings.Builder

	sb.WriteString("\n[📊 TOOL EXECUTION RESULTS]\n\n")

	for _, result := range results {
		sb.WriteString(fmt.Sprintf("%s:\n", result.ToolName))
		sb.WriteString(fmt.Sprintf("  Status: %s\n", result.Status))

		output := result.Output
		if len(output) > maxOutputChars {
			// Truncate output and indicate it was truncated
			output = output[:maxOutputChars] + fmt.Sprintf("\n\n[⚠️ OUTPUT TRUNCATED - Original size: %d characters]", len(result.Output))
		}
		sb.WriteString(fmt.Sprintf("  Output: %s\n\n", output))
	}

	sb.WriteString("[END RESULTS]\n")

	return sb.String()
}
