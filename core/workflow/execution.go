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

// ExecuteWorkflow executes the workflow starting from an entry agent
func ExecuteWorkflow(ctx context.Context, entryAgent *common.Agent, input string, history []common.Message, handler OutputHandler, signalRegistry *signal.SignalRegistry, apiKey string) (*common.AgentResponse, error) {
	if entryAgent == nil {
		return nil, fmt.Errorf("entry agent cannot be nil")
	}

	if handler == nil {
		handler = &NoOpHandler{}
	}

	execCtx := &ExecutionContext{
		CurrentAgent:   entryAgent,
		History:        history,
		MaxHandoffs:    5,  // Default max handoffs
		MaxRounds:      10, // Default max rounds
		StartTime:      time.Now(),
		handler:        handler,
		SignalRegistry: signalRegistry,
	}

	return executeAgent(ctx, execCtx, input, apiKey)
}

// executeAgent executes the current agent and handles routing
func executeAgent(ctx context.Context, execCtx *ExecutionContext, input string, apiKey string) (*common.AgentResponse, error) {
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
	if execCtx.SignalRegistry != nil {
		_ = execCtx.SignalRegistry.Emit(&signal.Signal{
			Name:    signal.SignalAgentStart,
			AgentID: execCtx.CurrentAgent.ID,
			Metadata: map[string]interface{}{
				"round": execCtx.RoundCount,
				"input": input,
			},
		})
	}

	// Execute current agent
	startTime := time.Now()
	response, err := agent.ExecuteAgent(ctx, execCtx.CurrentAgent, input, execCtx.History, apiKey)
	execCtx.LastAgentTime = time.Since(startTime)

	if err != nil {
		// SIGNAL 2: Emit agent:error
		if execCtx.SignalRegistry != nil {
			_ = execCtx.SignalRegistry.Emit(&signal.Signal{
				Name:    signal.SignalAgentError,
				AgentID: execCtx.CurrentAgent.ID,
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			})
		}

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
	if execCtx.SignalRegistry != nil {
		_ = execCtx.SignalRegistry.Emit(&signal.Signal{
			Name:    signal.SignalAgentEnd,
			AgentID: execCtx.CurrentAgent.ID,
			Metadata: map[string]interface{}{
				"duration_ms": execCtx.LastAgentTime.Milliseconds(),
			},
		})
	}

	// SIGNAL 4: Process custom signals from agent response
	var routingDecision *signal.RoutingDecision
	if execCtx.SignalRegistry != nil && response.Signals != nil && len(response.Signals) > 0 {
		for _, sigName := range response.Signals {
			sig := &signal.Signal{
				Name:    sigName,
				AgentID: execCtx.CurrentAgent.ID,
			}

			// Emit signal
			_ = execCtx.SignalRegistry.Emit(sig)

			// Process signal for routing decision
			decision, err := execCtx.SignalRegistry.ProcessSignal(ctx, sig)
			if err == nil && decision != nil {
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
		if execCtx.SignalRegistry != nil {
			_ = execCtx.SignalRegistry.Emit(&signal.Signal{
				Name:    signal.SignalHandoff,
				AgentID: execCtx.CurrentAgent.ID,
				Metadata: map[string]interface{}{
					"from_agent":     execCtx.CurrentAgent.ID,
					"to_agent":       routingDecision.NextAgentID,
					"routing_type":   "signal",
					"routing_reason": routingDecision.Reason,
				},
			})
		}

		// TODO: Look up next agent by ID and continue execution
		// For now, return (will be implemented in crew.go integration)
		return response, nil
	}

	// Priority 2: Check if agent is terminal
	if execCtx.CurrentAgent.IsTerminal {
		// SIGNAL 6: Emit route:terminal
		if execCtx.SignalRegistry != nil {
			_ = execCtx.SignalRegistry.Emit(&signal.Signal{
				Name:    signal.SignalTerminal,
				AgentID: execCtx.CurrentAgent.ID,
				Metadata: map[string]interface{}{
					"reason": "agent marked as terminal",
				},
			})
		}
		return response, nil
	}

	// Priority 3: Check for handoffs
	if len(execCtx.CurrentAgent.HandoffTargets) > 0 && execCtx.HandoffCount < execCtx.MaxHandoffs {
		nextAgentID := execCtx.CurrentAgent.HandoffTargets[0].ID

		// SIGNAL 7: Emit route:handoff
		if execCtx.SignalRegistry != nil {
			_ = execCtx.SignalRegistry.Emit(&signal.Signal{
				Name:    signal.SignalHandoff,
				AgentID: execCtx.CurrentAgent.ID,
				Metadata: map[string]interface{}{
					"from_agent":   execCtx.CurrentAgent.ID,
					"to_agent":     nextAgentID,
					"routing_type": "handoff_target",
				},
			})
		}

		// TODO: Look up next agent by ID and continue execution
		// For now, return response
		return response, nil
	}

	// No routing found - terminal
	if execCtx.SignalRegistry != nil {
		_ = execCtx.SignalRegistry.Emit(&signal.Signal{
			Name:    signal.SignalTerminal,
			AgentID: execCtx.CurrentAgent.ID,
			Metadata: map[string]interface{}{
				"reason": "no handoff targets configured",
			},
		})
	}

	return response, nil
}

// ExecuteWorkflowStream executes the workflow with streaming output
func ExecuteWorkflowStream(ctx context.Context, entryAgent *common.Agent, input string, history []common.Message, streamChan chan *StreamEvent, signalRegistry *signal.SignalRegistry, apiKey string) error {
	if entryAgent == nil {
		return fmt.Errorf("entry agent cannot be nil")
	}

	handler := NewStreamHandler(streamChan)

	execCtx := &ExecutionContext{
		CurrentAgent:   entryAgent,
		History:        history,
		MaxHandoffs:    5,
		MaxRounds:      10,
		StartTime:      time.Now(),
		handler:        handler,
		SignalRegistry: signalRegistry,
	}

	_, err := executeAgent(ctx, execCtx, input, apiKey)
	if err != nil {
		_ = handler.HandleError(err)
		return err
	}

	return nil
}

// ExecuteAgentWithMetrics executes an agent and updates metrics
func ExecuteAgentWithMetrics(ctx context.Context, agentObj *common.Agent, input string, history []common.Message, apiKey string) (*common.AgentResponse, error) {
	if agentObj == nil {
		return nil, fmt.Errorf("agent cannot be nil")
	}

	startTime := time.Now()
	response, err := agent.ExecuteAgent(ctx, agentObj, input, history, apiKey)
	duration := time.Since(startTime)

	// Update metrics (placeholder for full implementation)
	_ = duration

	return response, err
}
