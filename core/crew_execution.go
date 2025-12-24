package crewai

import (
	"context"
	"fmt"
	"log"
	"time"
)

// OutputHandler defines the interface for handling execution output
// Allows Execute() and ExecuteStream() to use same core logic with different outputs
type OutputHandler interface {
	// onAgentStart is called when an agent starts execution
	onAgentStart(ctx context.Context, agent *Agent)

	// onAgentResponse is called after agent produces a response
	onAgentResponse(ctx context.Context, agent *Agent, response *AgentResponse)

	// onToolStart is called when a tool is about to execute
	onToolStart(ctx context.Context, agent *Agent, toolName string)

	// onToolResult is called after tool execution completes
	onToolResult(ctx context.Context, agent *Agent, result ToolResult)

	// onTerminationSignal is called when workflow should terminate
	onTerminationSignal(ctx context.Context, agent *Agent, signal string)

	// onPauseSignal is called when workflow should pause
	onPauseSignal(ctx context.Context, agent *Agent, agentID string)

	// onRoutingSignal is called when routing to next agent
	onRoutingSignal(ctx context.Context, fromAgent *Agent, toAgent *Agent)

	// onParallelExecution is called before parallel agent execution
	onParallelExecution(ctx context.Context, agents []*Agent)

	// onError is called when an error occurs
	onError(ctx context.Context, agent *Agent, err error)

	// onMaxHandoffsExceeded is called when max handoffs reached
	onMaxHandoffsExceeded(ctx context.Context, agent *Agent)

	// onMaxRoundsExceeded is called when max rounds reached (for future use)
	onMaxRoundsExceeded(ctx context.Context, agent *Agent)

	// finalizeExecution returns the final result
	// For sync: CrewResponse
	// For stream: error
	finalizeExecution(ctx context.Context, agent *Agent, response *AgentResponse, reason string) interface{}
}

// executionContext holds state during workflow execution
// Shared between executeWorkflow and handlers
type executionContext struct {
	handoffCount   int           // Number of agent handoffs so far
	toolResultText string        // Formatted tool results to feed back
	lastResponse   *AgentResponse // Last agent response
	lastAgent      *Agent        // Current/last executing agent
}

// executeWorkflow is the core execution logic extracted from Execute/ExecuteStream
// This is the single source of truth for execution logic
// The handler parameter determines how results are output
func (ce *CrewExecutor) executeWorkflow(ctx context.Context, input string, handler OutputHandler) interface{} {
	// Add user input to history
	ce.appendMessage(Message{
		Role:    RoleUser,
		Content: input,
	})

	// Determine starting agent: resume agent or entry agent
	var currentAgent *Agent
	if ce.ResumeAgentID != "" {
		// Resume from paused agent
		currentAgent = ce.findAgentByID(ce.ResumeAgentID)
		if currentAgent == nil {
			return handler.finalizeExecution(ctx, nil, nil, fmt.Sprintf("resume agent %s not found", ce.ResumeAgentID))
		}
		// Clear resume agent after using it
		ce.ResumeAgentID = ""
	} else {
		// Start with entry agent
		currentAgent = ce.entryAgent
		if currentAgent == nil {
			return handler.finalizeExecution(ctx, nil, nil, "no entry agent found")
		}
	}

	execCtx := &executionContext{handoffCount: 0}

	for {
		select {
		case <-ctx.Done():
			return handler.finalizeExecution(ctx, currentAgent, nil, fmt.Sprintf("context cancelled: %v", ctx.Err()))
		default:
		}

		// Call handler for agent start
		handler.onAgentStart(ctx, currentAgent)

		// Trim history before LLM call to prevent cost leakage
		ce.trimHistoryIfNeeded()

		// Track agent execution time for metrics
		agentStartTime := time.Now()

		// Execute current agent
		response, err := ExecuteAgent(ctx, currentAgent, input, ce.getHistoryCopy(), ce.apiKey)
		agentEndTime := time.Now()
		agentDuration := agentEndTime.Sub(agentStartTime)

		if err != nil {
			// Update performance metrics with error
			if currentAgent.Metadata != nil {
				currentAgent.UpdatePerformanceMetrics(false, err.Error())
			}

			// Check error quota
			if quotaErr := currentAgent.CheckErrorQuota(); quotaErr != nil {
				log.Printf("[QUOTA] Agent %s exceeded error quota: %v", currentAgent.ID, quotaErr)
				handler.onError(ctx, currentAgent, quotaErr)
				return handler.finalizeExecution(ctx, currentAgent, nil, fmt.Sprintf("error quota exceeded: %v", quotaErr))
			}

			// Report agent error
			handler.onError(ctx, currentAgent, err)

			// Record failed agent execution
			if ce.Metrics != nil {
				ce.Metrics.RecordAgentExecution(currentAgent.ID, currentAgent.Name, agentDuration, false)
			}
			return handler.finalizeExecution(ctx, currentAgent, nil, fmt.Sprintf("agent %s failed: %v", currentAgent.ID, err))
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
				handler.onError(ctx, currentAgent, err)
				return handler.finalizeExecution(ctx, currentAgent, nil, fmt.Sprintf("memory quota exceeded: %v", err))
			}

			// Update memory & performance metrics (only after quota check passes)
			callDurationMs := agentDuration.Milliseconds()
			if currentAgent.Metadata != nil {
				currentAgent.UpdateMemoryMetrics(memoryUsedMB, callDurationMs)
				currentAgent.UpdatePerformanceMetrics(true, "")
			}
		}

		// Call handler for agent response
		handler.onAgentResponse(ctx, currentAgent, response)

		// Add agent response to history
		ce.appendMessage(Message{
			Role:    RoleAssistant,
			Content: response.Content,
		})

		// Execute any tool calls BEFORE checking if terminal
		if len(response.ToolCalls) > 0 {
			for _, toolCall := range response.ToolCalls {
				handler.onToolStart(ctx, currentAgent, toolCall.ToolName)
			}

			toolResults := ce.executeCalls(ctx, response.ToolCalls, currentAgent)

			// Send tool results
			for _, result := range toolResults {
				handler.onToolResult(ctx, currentAgent, result)
			}

			// Format results for feedback
			resultText := ce.formatToolResults(toolResults)
			execCtx.toolResultText = resultText

			// Add results to history
			ce.appendMessage(Message{
				Role:    RoleUser,
				Content: resultText,
			})

			// Feed results back to current agent for analysis
			input = resultText
			continue
		}

		// Check for termination signals first (before routing)
		// If agent emits a termination signal like [KẾT THÚC THI], workflow should end
		terminationResult := ce.checkTerminationSignal(currentAgent, response.Content)
		if terminationResult != nil && terminationResult.ShouldTerminate {
			handler.onTerminationSignal(ctx, currentAgent, terminationResult.Signal)
			return handler.finalizeExecution(ctx, currentAgent, response, fmt.Sprintf("terminated by signal: %s", terminationResult.Signal))
		}

		// Check for routing signals from current agent (config-driven)
		nextAgent := ce.findNextAgentBySignal(currentAgent, response.Content)
		if nextAgent != nil {
			handler.onRoutingSignal(ctx, currentAgent, nextAgent)
			currentAgent = nextAgent
			input = response.Content
			execCtx.handoffCount++
			continue
		}

		// Check wait_for_signal BEFORE terminal check
		behavior := ce.getAgentBehavior(currentAgent.ID)
		if behavior != nil && behavior.WaitForSignal {
			handler.onPauseSignal(ctx, currentAgent, currentAgent.ID)
			return handler.finalizeExecution(ctx, currentAgent, response, fmt.Sprintf("paused: waiting for signal"))
		}

		// Check if current agent is terminal (only after tool execution and wait_for_signal)
		if currentAgent.IsTerminal {
			return handler.finalizeExecution(ctx, currentAgent, response, "terminal agent")
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
				handler.onParallelExecution(ctx, parallelAgents)

				// Execute all parallel agents
				parallelResults, err := ce.ExecuteParallel(ctx, input, parallelAgents)
				if err != nil {
					handler.onError(ctx, currentAgent, err)
					return handler.finalizeExecution(ctx, currentAgent, nil, fmt.Sprintf("parallel execution failed: %v", err))
				}

				// Aggregate results
				aggregatedInput := ce.aggregateParallelResults(parallelResults)

				// Add aggregated results to history
				ce.appendMessage(Message{
					Role:    RoleUser,
					Content: aggregatedInput,
				})

				// Move to next agent in the pipeline
				if parallelGroup.NextAgent != "" {
					if nextAgent, exists := agentMap[parallelGroup.NextAgent]; exists {
						currentAgent = nextAgent
						input = aggregatedInput
						execCtx.handoffCount++
						continue
					}
				}
			}
		}

		// For other agents, handoff normally
		execCtx.handoffCount++
		if execCtx.handoffCount >= ce.crew.MaxHandoffs {
			handler.onMaxHandoffsExceeded(ctx, currentAgent)
			return handler.finalizeExecution(ctx, currentAgent, response, "max handoffs exceeded")
		}

		// Find next agent based on handoff targets
		nextAgent = ce.findNextAgent(currentAgent)
		if nextAgent == nil {
			return handler.finalizeExecution(ctx, currentAgent, response, "no next agent found")
		}

		currentAgent = nextAgent
		input = response.Content
	}
}

// SyncHandler handles synchronous execution (Execute mode)
// Returns CrewResponse
type SyncHandler struct {
	verbose bool
	ce      *CrewExecutor
}

// NewSyncHandler creates a handler for synchronous execution
func NewSyncHandler(ce *CrewExecutor, verbose bool) *SyncHandler {
	return &SyncHandler{
		verbose: verbose,
		ce:      ce,
	}
}

func (h *SyncHandler) onAgentStart(ctx context.Context, agent *Agent) {
	if agent != nil {
		log.Printf("[AGENT START] %s (%s)", agent.Name, agent.ID)
	}
}

func (h *SyncHandler) onAgentResponse(ctx context.Context, agent *Agent, response *AgentResponse) {
	if agent != nil {
		log.Printf("[AGENT END] %s (%s) - Success", agent.Name, agent.ID)
	}
	if h.verbose && response != nil {
		fmt.Printf("\n[%s]: %s\n", agent.Name, response.Content)
	}
}

func (h *SyncHandler) onToolStart(ctx context.Context, agent *Agent, toolName string) {
	// Silent in sync mode
}

func (h *SyncHandler) onToolResult(ctx context.Context, agent *Agent, result ToolResult) {
	// Silent in sync mode
}

func (h *SyncHandler) onTerminationSignal(ctx context.Context, agent *Agent, signal string) {
	if agent != nil {
		log.Printf("[TERMINATE] Workflow ended by signal: %s", signal)
	}
}

func (h *SyncHandler) onPauseSignal(ctx context.Context, agent *Agent, agentID string) {
	// Silent - caller will get CrewResponse with PausedAgentID
}

func (h *SyncHandler) onRoutingSignal(ctx context.Context, fromAgent *Agent, toAgent *Agent) {
	// Silent in sync mode
}

func (h *SyncHandler) onParallelExecution(ctx context.Context, agents []*Agent) {
	// Silent in sync mode
}

func (h *SyncHandler) onError(ctx context.Context, agent *Agent, err error) {
	if agent != nil {
		log.Printf("[AGENT ERROR] %s (%s) - %v", agent.Name, agent.ID, err)
	}
}

func (h *SyncHandler) onMaxHandoffsExceeded(ctx context.Context, agent *Agent) {
	log.Printf("[MAX HANDOFFS] Workflow ended - max handoffs exceeded")
}

func (h *SyncHandler) onMaxRoundsExceeded(ctx context.Context, agent *Agent) {
	// Not used in current implementation
}

func (h *SyncHandler) finalizeExecution(ctx context.Context, agent *Agent, response *AgentResponse, reason string) interface{} {
	// Check if it's an error case (reason starts with keywords indicating error)
	if reason == "terminal agent" || reason == "no next agent found" || reason == "paused: waiting for signal" || reason == "terminated by signal" || reason == "max handoffs exceeded" {
		// Success cases - return CrewResponse
		if agent == nil {
			return &CrewResponse{}
		}

		result := &CrewResponse{
			AgentID:   agent.ID,
			AgentName: agent.Name,
		}

		if response != nil {
			result.Content = response.Content
			result.ToolCalls = response.ToolCalls
		}

		// Set flags based on reason
		if reason == "terminal agent" {
			result.IsTerminal = true
		} else if reason == "paused: waiting for signal" {
			result.PausedAgentID = agent.ID
		} else if reason == "terminated by signal" {
			result.IsTerminal = true
		}

		return result
	}

	// Error case - return error
	return fmt.Errorf("%s", reason)
}

// StreamHandler handles asynchronous execution (ExecuteStream mode)
// Sends events via streamChan
type StreamHandler struct {
	streamChan chan *StreamEvent
}

// NewStreamHandler creates a handler for streaming execution
func NewStreamHandler(streamChan chan *StreamEvent) *StreamHandler {
	return &StreamHandler{streamChan: streamChan}
}

func (h *StreamHandler) sendEvent(eventType, agentName, message string) {
	if h.streamChan == nil {
		return
	}

	select {
	case h.streamChan <- NewStreamEvent(eventType, agentName, message):
		// Event sent successfully
	case <-time.After(100 * time.Millisecond):
		log.Printf("WARNING: stream event send timeout for event: %s", eventType)
	}
}

func (h *StreamHandler) onAgentStart(ctx context.Context, agent *Agent) {
	if agent != nil {
		h.sendEvent("agent_start", agent.Name, fmt.Sprintf("🔄 Starting %s...", agent.Name))
	}
}

func (h *StreamHandler) onAgentResponse(ctx context.Context, agent *Agent, response *AgentResponse) {
	if agent != nil && response != nil {
		h.sendEvent("agent_response", agent.Name, response.Content)
	}
}

func (h *StreamHandler) onToolStart(ctx context.Context, agent *Agent, toolName string) {
	if agent != nil {
		h.sendEvent("tool_start", agent.Name, fmt.Sprintf("🔧 [Tool] %s → Executing...", toolName))
	}
}

func (h *StreamHandler) onToolResult(ctx context.Context, agent *Agent, result ToolResult) {
	if agent != nil {
		status := "✅"
		if result.Status == "error" {
			status = "❌"
		}
		h.sendEvent(EventTypeToolResult, agent.Name, fmt.Sprintf("%s [Tool] %s → %s", status, result.ToolName, result.Output))
	}
}

func (h *StreamHandler) onTerminationSignal(ctx context.Context, agent *Agent, signal string) {
	if agent != nil {
		h.sendEvent("terminate", agent.Name, fmt.Sprintf("[TERMINATE] Workflow ended by signal: %s", signal))
	}
}

func (h *StreamHandler) onPauseSignal(ctx context.Context, agent *Agent, agentID string) {
	if agent != nil {
		h.sendEvent("pause", agent.Name, fmt.Sprintf("[PAUSE:%s] Waiting for user input", agentID))
	}
}

func (h *StreamHandler) onRoutingSignal(ctx context.Context, fromAgent *Agent, toAgent *Agent) {
	// Can be extended to send routing events
}

func (h *StreamHandler) onParallelExecution(ctx context.Context, agents []*Agent) {
	// Can be extended to send parallel execution events
}

func (h *StreamHandler) onError(ctx context.Context, agent *Agent, err error) {
	if agent != nil {
		h.sendEvent(EventTypeError, agent.Name, fmt.Sprintf("Agent failed: %v", err))
	}
}

func (h *StreamHandler) onMaxHandoffsExceeded(ctx context.Context, agent *Agent) {
	h.sendEvent(EventTypeError, "system", "Max handoffs exceeded")
}

func (h *StreamHandler) onMaxRoundsExceeded(ctx context.Context, agent *Agent) {
	h.sendEvent(EventTypeError, "system", "Max rounds exceeded")
}

func (h *StreamHandler) finalizeExecution(ctx context.Context, agent *Agent, response *AgentResponse, reason string) interface{} {
	// Stream mode always returns an error (or nil on success)
	// Check if it's a success case
	if reason == "terminal agent" || reason == "no next agent found" || reason == "paused: waiting for signal" || reason == "terminated by signal" || reason == "max handoffs exceeded" {
		return nil
	}

	// Error case
	return fmt.Errorf("%s", reason)
}
