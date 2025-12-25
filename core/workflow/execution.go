// Package workflow provides workflow orchestration and execution functionality.
package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/taipm/go-agentic/core/agent"
	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/signal"
)

// Workflow execution constants
const (
	DefaultMaxHandoffs      = 5
	DefaultMaxRounds        = 10
	ErrTargetAgentNotFound  = "target agent not found: %s"
)

// ExecutionContext holds the state during workflow execution
type ExecutionContext struct {
	CurrentAgent    *common.Agent
	History         []common.Message
	HandoffCount    int
	RoundCount      int
	MaxHandoffs     int
	MaxRounds       int
	StartTime       time.Time
	LastAgentTime   time.Duration
	TotalTime       time.Duration
	handler         OutputHandler
	SignalRegistry  *signal.SignalRegistry
}

// emitSignal emits a signal with the given name and metadata
func (execCtx *ExecutionContext) emitSignal(signalName string, metadata map[string]interface{}) {
	if execCtx.SignalRegistry == nil {
		return
	}

	if err := execCtx.SignalRegistry.Emit(&signal.Signal{
		Name:     signalName,
		AgentID:  execCtx.CurrentAgent.ID,
		Metadata: metadata,
	}); err != nil {
		fmt.Printf("[WARN] Failed to emit signal '%s': %v\n", signalName, err)
	}
}

// lookupNextAgent looks up an agent in the agents map and returns it
// Returns ExecutionError if agent not found
func lookupNextAgent(agents map[string]*common.Agent, agentID string, currentAgentID string) (*common.Agent, error) {
	if agents == nil {
		return nil, nil
	}

	nextAgent, exists := agents[agentID]
	if !exists {
		return nil, &common.ExecutionError{
			AgentID:   currentAgentID,
			ErrorType: common.ErrorTypeValidation,
			Err:       fmt.Errorf(ErrTargetAgentNotFound, agentID),
		}
	}
	return nextAgent, nil
}

// ExecuteWorkflow executes the workflow starting from an entry agent
// agents: map of all available agents for handoff lookup (can be nil if no handoffs needed)
func ExecuteWorkflow(ctx context.Context, entryAgent *common.Agent, input string, history []common.Message, handler OutputHandler, signalRegistry *signal.SignalRegistry, apiKey string, agents map[string]*common.Agent) (*common.AgentResponse, error) {
	if entryAgent == nil {
		return nil, fmt.Errorf("entry agent cannot be nil")
	}

	if handler == nil {
		handler = NewNoOpHandler()
	}

	execCtx := &ExecutionContext{
		CurrentAgent:   entryAgent,
		History:        history,
		MaxHandoffs:    DefaultMaxHandoffs,
		MaxRounds:      DefaultMaxRounds,
		StartTime:      time.Now(),
		handler:        handler,
		SignalRegistry: signalRegistry,
	}

	return executeAgent(ctx, execCtx, input, apiKey, agents)
}

// executeAgent executes the current agent and handles routing
// agents: map of all available agents for handoff lookup (can be nil)
func executeAgent(ctx context.Context, execCtx *ExecutionContext, input string, apiKey string, agents map[string]*common.Agent) (*common.AgentResponse, error) {
	if execCtx.RoundCount >= execCtx.MaxRounds {
		return nil, fmt.Errorf("max rounds (%d) exceeded", execCtx.MaxRounds)
	}

	execCtx.RoundCount++

	// Add user input to history on first iteration
	if execCtx.RoundCount == 1 {
		execCtx.History = append(execCtx.History, common.Message{
			Role:    common.RoleUser,
			Content: input,
		})
	}

	// SIGNAL 1: Emit agent:start
	execCtx.emitSignal(signal.SignalAgentStart, map[string]interface{}{
		"round": execCtx.RoundCount,
		"input": input,
	})

	// Execute current agent
	startTime := time.Now()
	response, err := agent.ExecuteAgent(ctx, execCtx.CurrentAgent, input, execCtx.History, apiKey)
	execCtx.LastAgentTime = time.Since(startTime)

	if err != nil {
		// SIGNAL 2: Emit agent:error
		execCtx.emitSignal(signal.SignalAgentError, map[string]interface{}{
			"error": err.Error(),
		})

		handler := execCtx.handler
		if handler != nil {
			_ = handler.HandleError(err)
		}
		return nil, err
	}

	// Add agent response to history
	execCtx.History = append(execCtx.History, common.Message{
		Role:    common.RoleAssistant,
		Content: response.Content,
	})

	// Notify handler
	if execCtx.handler != nil {
		_ = execCtx.handler.HandleAgentResponse(response)
	}

	// SIGNAL 3: Emit agent:end
	execCtx.emitSignal(signal.SignalAgentEnd, map[string]interface{}{
		"duration_ms": execCtx.LastAgentTime.Milliseconds(),
	})

	// SIGNAL 4: Process custom signals from agent response
	var routingDecision *common.RoutingDecision
	if execCtx.SignalRegistry != nil && response.Signals != nil && len(response.Signals) > 0 {
		for _, sigName := range response.Signals {
			sig := &signal.Signal{
				Name:    sigName,
				AgentID: execCtx.CurrentAgent.ID,
			}

			// Emit signal
			if err := execCtx.SignalRegistry.Emit(sig); err != nil {
				fmt.Printf("[WARN] Failed to emit signal '%s': %v\n", sigName, err)
				continue
			}

			// Process signal for routing decision
			decision, err := execCtx.SignalRegistry.ProcessSignal(ctx, sig)
			if err != nil {
				fmt.Printf("[WARN] Failed to process signal '%s': %v\n", sigName, err)
				continue
			}

			if decision != nil {
				routingDecision = decision

				// If terminal signal, stop execution
				if decision.IsTerminal {
					return response, nil
				}

				// Found routing decision, stop processing signals
				if decision.NextAgentID != "" {
					break
				}
			}
		}
	}

	// ROUTING PRIORITY:
	// 1. Signal-based routing (if decision found)
	// 2. Terminal agent check
	// 3. Default handoff targets

	// Priority 1: Signal routing takes precedence
	if routingDecision != nil && routingDecision.NextAgentID != "" {
		// SIGNAL 5: Emit route:handoff
		execCtx.emitSignal(signal.SignalHandoff, map[string]interface{}{
			"from_agent":     execCtx.CurrentAgent.ID,
			"to_agent":       routingDecision.NextAgentID,
			"routing_type":   "signal",
			"routing_reason": routingDecision.Reason,
		})

		// Phase 5 Implementation: Look up next agent and continue execution
		nextAgent, err := lookupNextAgent(agents, routingDecision.NextAgentID, execCtx.CurrentAgent.ID)
		if err != nil {
			return nil, err
		}

		if nextAgent != nil {
			// Update execution context for next agent
			execCtx.CurrentAgent = nextAgent
			execCtx.HandoffCount++

			// Recursively execute next agent with remaining input from response
			// Use agent's response content as input for next agent
			return executeAgent(ctx, execCtx, "", apiKey, agents)
		}

		// If agents map not provided, cannot continue (backward compatibility)
		return response, nil
	}

	// Priority 2: Check if agent is terminal
	if execCtx.CurrentAgent.IsTerminal {
		// SIGNAL 6: Emit route:terminal
		execCtx.emitSignal(signal.SignalTerminal, map[string]interface{}{
			"reason": "agent marked as terminal",
		})
		return response, nil
	}

	// Priority 3: Check for handoffs
	if len(execCtx.CurrentAgent.HandoffTargets) > 0 && execCtx.HandoffCount < execCtx.MaxHandoffs {
		nextAgentID := execCtx.CurrentAgent.HandoffTargets[0].ID

		// SIGNAL 7: Emit route:handoff
		execCtx.emitSignal(signal.SignalHandoff, map[string]interface{}{
			"from_agent":   execCtx.CurrentAgent.ID,
			"to_agent":     nextAgentID,
			"routing_type": "handoff_target",
		})

		// Phase 5 Implementation: Look up next agent and continue execution
		nextAgent, err := lookupNextAgent(agents, nextAgentID, execCtx.CurrentAgent.ID)
		if err != nil {
			return nil, err
		}

		if nextAgent != nil {
			// Update execution context for next agent
			execCtx.CurrentAgent = nextAgent
			execCtx.HandoffCount++

			// Recursively execute next agent
			return executeAgent(ctx, execCtx, "", apiKey, agents)
		}

		// If agents map not provided, cannot continue (backward compatibility)
		return response, nil
	}

	// No routing found - terminal
	execCtx.emitSignal(signal.SignalTerminal, map[string]interface{}{
		"reason": "no handoff targets configured",
	})

	return response, nil
}

// ExecuteWorkflowStream executes the workflow with streaming output
func ExecuteWorkflowStream(ctx context.Context, entryAgent *common.Agent, input string, history []common.Message, streamChan chan *common.StreamEvent, signalRegistry *signal.SignalRegistry, apiKey string) error {
	if entryAgent == nil {
		return fmt.Errorf("entry agent cannot be nil")
	}

	handler := NewStreamHandler(streamChan)

	execCtx := &ExecutionContext{
		CurrentAgent:   entryAgent,
		History:        history,
		MaxHandoffs:    DefaultMaxHandoffs,
		MaxRounds:      DefaultMaxRounds,
		StartTime:      time.Now(),
		handler:        handler,
		SignalRegistry: signalRegistry,
	}

	_, err := executeAgent(ctx, execCtx, input, apiKey, nil)
	if err != nil {
		_ = handler.HandleError(err)
		return err
	}

	return nil
}
