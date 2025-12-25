// Package execution provides workflow execution coordination for multi-agent systems.
package execution

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/routing"
	"github.com/taipm/go-agentic/core/signal"
	"github.com/taipm/go-agentic/core/state-management"
	"github.com/taipm/go-agentic/core/tools"
	"github.com/taipm/go-agentic/core/workflow"
)

// ExecutionFlow represents the state of workflow execution.
type ExecutionFlow struct {
	CurrentAgent   *common.Agent
	History        []common.Message
	RoundCount     int
	HandoffCount   int
	MaxRounds      int
	MaxHandoffs    int
	State          *statemanagement.ExecutionState
	SignalRegistry *signal.SignalRegistry
}

// NewExecutionFlow creates a new ExecutionFlow with initial state.
func NewExecutionFlow(entryAgent *common.Agent, maxRounds, maxHandoffs int) *ExecutionFlow {
	return &ExecutionFlow{
		CurrentAgent: entryAgent,
		History:      make([]common.Message, 0),
		RoundCount:   0,
		HandoffCount: 0,
		MaxRounds:    maxRounds,
		MaxHandoffs:  maxHandoffs,
		State:        statemanagement.NewExecutionState(),
	}
}

// CanContinue checks if execution can continue based on round and handoff limits.
func (ef *ExecutionFlow) CanContinue() error {
	if ef.RoundCount >= ef.MaxRounds {
		return fmt.Errorf("maximum rounds (%d) reached", ef.MaxRounds)
	}
	if ef.HandoffCount >= ef.MaxHandoffs {
		return fmt.Errorf("maximum handoffs (%d) reached", ef.MaxHandoffs)
	}
	return nil
}

// ExecuteWorkflowStep executes a single agent in the workflow.
// Returns the agent response and any error that occurred.
func (ef *ExecutionFlow) ExecuteWorkflowStep(
	ctx context.Context,
	handler workflow.OutputHandler,
	apiKey string,
) (*common.AgentResponse, error) {
	if ef.CurrentAgent == nil {
		return nil, fmt.Errorf("no current agent set for workflow execution")
	}

	// Check execution limits
	if err := ef.CanContinue(); err != nil {
		return nil, err
	}

	ef.RoundCount++

	// Record start time
	startTime := time.Now()

	// Create input for agent
	userInput := ""
	if len(ef.History) > 0 {
		// Use conversation history as context
		userInput = ef.History[len(ef.History)-1].Content
	}

	// Execute agent using workflow package
	response, err := workflow.ExecuteWorkflow(
		ctx,
		ef.CurrentAgent,
		userInput,
		ef.History,
		handler,
		ef.SignalRegistry,
		apiKey,
	)

	// Record metrics
	duration := time.Since(startTime)
	ef.State.RecordRound(ef.CurrentAgent.ID, duration, err == nil)

	if err != nil {
		// Emit error event
		handler.HandleError(err)
		return nil, err
	}

	// Add response to history
	responseMsg := common.Message{
		Role:    "assistant",
		Content: response.Content,
	}
	ef.History = append(ef.History, responseMsg)

	// Execute tool calls if present in response
	if response != nil && len(response.ToolCalls) > 0 {
		toolResults, toolErr := tools.ExecuteToolCalls(ctx, response.ToolCalls, ef.CurrentAgent.Tools)

		// Add tool results to history if any tools executed successfully
		if len(toolResults) > 0 {
			resultMsg := common.Message{
				Role:    "system",
				Content: tools.FormatToolResults(toolResults),
			}
			ef.History = append(ef.History, resultMsg)
		}

		// Log any tool execution errors (but don't fail the workflow)
		if toolErr != nil {
			log.Printf("[WORKFLOW] Tool execution had errors: %v", toolErr)
		}
	}

	// Emit response event
	if err := handler.HandleAgentResponse(response); err != nil {
		return nil, err
	}

	return response, nil
}

// HandleAgentResponse processes agent response and determines next step.
// Returns true if workflow should continue, false if terminal.
func (ef *ExecutionFlow) HandleAgentResponse(
	ctx context.Context,
	response *common.AgentResponse,
	routingConfig *common.RoutingConfig,
	agentsMap map[string]*common.Agent,
) (bool, error) {
	if response == nil {
		return false, fmt.Errorf("agent response is nil")
	}

	// Check if this is a terminal agent
	if routingConfig != nil {
		// Priority 1: Use signal-based routing if SignalRegistry available
		var decision *common.RoutingDecision
		var err error

		if ef.SignalRegistry != nil {
			// Use signal-aware routing function
			decision, err = routing.DetermineNextAgentWithSignals(
				ctx,
				ef.CurrentAgent,
				response,
				routingConfig,
				ef.SignalRegistry,
			)
		} else {
			// Fallback to simple routing (no signal support)
			decision, err = routing.DetermineNextAgent(
				ef.CurrentAgent,
				response,
				routingConfig,
			)
		}

		if err != nil {
			return false, fmt.Errorf("failed to determine next agent: %w", err)
		}

		if decision.IsTerminal {
			return false, nil // Workflow complete
		}

		// Get next agent
		nextAgentID := decision.NextAgentID
		if nextAgentID != "" {
			if nextAgent, exists := agentsMap[nextAgentID]; exists {
				ef.CurrentAgent = nextAgent
				ef.HandoffCount++
				ef.State.RecordHandoff()
				return true, nil // Continue with next agent
			}
			return false, fmt.Errorf("agent not found: %s", nextAgentID)
		}

		// No next agent determined - continue anyway (allow fallthrough)
		return true, nil
	}

	// No routing config - use default behavior (first agent terminates)
	if ef.CurrentAgent == agentsMap[ef.CurrentAgent.ID] {
		// Check if it's the only agent or we should continue
		return false, nil
	}

	return true, nil
}

// GetWorkflowStatus returns current execution status.
func (ef *ExecutionFlow) GetWorkflowStatus() map[string]interface{} {
	return map[string]interface{}{
		"current_agent":   ef.CurrentAgent.ID,
		"round_count":     ef.RoundCount,
		"handoff_count":   ef.HandoffCount,
		"max_rounds":      ef.MaxRounds,
		"max_handoffs":    ef.MaxHandoffs,
		"history_length":  len(ef.History),
		"can_continue":    ef.CanContinue() == nil,
	}
}

// Reset clears the workflow state for a new execution.
func (ef *ExecutionFlow) Reset(entryAgent *common.Agent) {
	ef.CurrentAgent = entryAgent
	ef.History = make([]common.Message, 0)
	ef.RoundCount = 0
	ef.HandoffCount = 0
	ef.State.Reset()
}

// ExecuteWorkflowWithCallback executes workflow with callback for each step.
type WorkflowCallback func(step int, agent *common.Agent, response *common.AgentResponse) error

// ExecuteWithCallbacks runs workflow execution with custom callbacks.
func (ef *ExecutionFlow) ExecuteWithCallbacks(
	ctx context.Context,
	handler workflow.OutputHandler,
	apiKey string,
	onStep WorkflowCallback,
	agents map[string]*common.Agent,
	routingConfig *common.RoutingConfig,
	signalRegistry *signal.SignalRegistry,
) (*common.AgentResponse, error) {
	// Assign signal registry for routing decisions
	ef.SignalRegistry = signalRegistry

	var lastResponse *common.AgentResponse

	for {
		// Check if we can continue
		if err := ef.CanContinue(); err != nil {
			return lastResponse, err
		}

		// Execute workflow step
		response, err := ef.ExecuteWorkflowStep(ctx, handler, apiKey)
		if err != nil {
			return lastResponse, err
		}

		lastResponse = response

		// Call step callback if provided
		if onStep != nil {
			if err := onStep(ef.RoundCount, ef.CurrentAgent, response); err != nil {
				return lastResponse, err
			}
		}

		// Check if workflow should continue
		shouldContinue, err := ef.HandleAgentResponse(ctx, response, routingConfig, agents)
		if err != nil {
			return lastResponse, err
		}

		if !shouldContinue {
			break
		}
	}

	// Finish execution
	ef.State.Finish()
	return lastResponse, nil
}

// Copy creates a copy of the ExecutionFlow for branching/parallel execution.
func (ef *ExecutionFlow) Copy() *ExecutionFlow {
	return &ExecutionFlow{
		CurrentAgent:   ef.CurrentAgent,
		History:        append([]common.Message{}, ef.History...),
		RoundCount:     ef.RoundCount,
		HandoffCount:   ef.HandoffCount,
		MaxRounds:      ef.MaxRounds,
		MaxHandoffs:    ef.MaxHandoffs,
		State:          ef.State.Copy(),
		SignalRegistry: ef.SignalRegistry,
	}
}

// ValidateFlow checks if the execution flow is valid for continued execution.
func (ef *ExecutionFlow) ValidateFlow() error {
	if ef.CurrentAgent == nil {
		return fmt.Errorf("current agent is nil")
	}
	if ef.MaxRounds <= 0 {
		return fmt.Errorf("max rounds must be positive")
	}
	if ef.MaxHandoffs < 0 {
		return fmt.Errorf("max handoffs cannot be negative")
	}
	if ef.State == nil {
		return fmt.Errorf("execution state is nil")
	}
	return nil
}
