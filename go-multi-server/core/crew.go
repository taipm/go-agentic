package crewai

import (
	"context"
	"fmt"
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

// CrewExecutor handles the execution of a crew
type CrewExecutor struct {
	crew          *Crew
	apiKey        string
	entryAgent    *Agent
	history       []Message
	Verbose       bool   // If true, print agent responses and tool results to stdout
	ResumeAgentID string // If set, execution will start from this agent instead of entry agent
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
		crew:       crew,
		apiKey:     apiKey,
		entryAgent: entryAgent,
		history:    []Message{},
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

		// Execute current agent
		response, err := ExecuteAgent(ctx, currentAgent, input, ce.history, ce.apiKey)
		if err != nil {
			streamChan <- NewStreamEvent("error", currentAgent.Name, fmt.Sprintf("Agent failed: %v", err))
			return fmt.Errorf("agent %s failed: %w", currentAgent.ID, err)
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
		response, err := ExecuteAgent(ctx, currentAgent, input, ce.history, ce.apiKey)
		if err != nil {
			return nil, fmt.Errorf("agent %s failed: %w", currentAgent.ID, err)
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

// executeCalls executes tool calls from an agent
func (ce *CrewExecutor) executeCalls(ctx context.Context, calls []ToolCall, agent *Agent) []ToolResult {
	var results []ToolResult

	toolMap := make(map[string]*Tool)
	for _, tool := range agent.Tools {
		toolMap[tool.Name] = tool
	}

	for _, call := range calls {
		tool, ok := toolMap[call.ToolName]
		if !ok {
			results = append(results, ToolResult{
				ToolName: call.ToolName,
				Status:   "error",
				Output:   fmt.Sprintf("Tool %s not found", call.ToolName),
			})
			continue
		}

		output, err := tool.Handler(ctx, call.Arguments)
		if err != nil {
			results = append(results, ToolResult{
				ToolName: call.ToolName,
				Status:   "error",
				Output:   err.Error(),
			})
		} else {
			results = append(results, ToolResult{
				ToolName: call.ToolName,
				Status:   "success",
				Output:   output,
			})
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
		if strings.Contains(responseContent, sig.Signal) && sig.Target != "" {
			// Found matching signal, find the target agent
			return ce.findAgentByID(sig.Target)
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
				return agent
			}
		}
	}

	// Fallback: Find any other agent (not terminal-only strategy)
	for _, agent := range ce.crew.Agents {
		if agent.ID != current.ID {
			return agent
		}
	}

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
	var sb strings.Builder

	sb.WriteString("\n[📊 TOOL EXECUTION RESULTS]\n\n")

	for _, result := range results {
		sb.WriteString(fmt.Sprintf("%s:\n", result.ToolName))
		sb.WriteString(fmt.Sprintf("  Status: %s\n", result.Status))
		sb.WriteString(fmt.Sprintf("  Output: %s\n\n", result.Output))
	}

	sb.WriteString("[END RESULTS]\n")

	return sb.String()
}
