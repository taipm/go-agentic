// Package workflow provides workflow orchestration and execution functionality.
package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/taipm/go-agentic/core/agent"
	"github.com/taipm/go-agentic/core/common"
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
}

// ExecuteWorkflow executes the workflow starting from an entry agent
func ExecuteWorkflow(ctx context.Context, entryAgent *common.Agent, input string, history []common.Message, handler OutputHandler, apiKey string) (*common.AgentResponse, error) {
	if entryAgent == nil {
		return nil, fmt.Errorf("entry agent cannot be nil")
	}

	if handler == nil {
		handler = &NoOpHandler{}
	}

	execCtx := &ExecutionContext{
		CurrentAgent: entryAgent,
		History:      history,
		MaxHandoffs:  5,  // Default max handoffs
		MaxRounds:    10, // Default max rounds
		StartTime:    time.Now(),
		handler:      handler,
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

	// Execute current agent
	startTime := time.Now()
	response, err := agent.ExecuteAgent(ctx, execCtx.CurrentAgent, input, execCtx.History, apiKey)
	execCtx.LastAgentTime = time.Since(startTime)

	if err != nil {
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

	// Check if agent is terminal (execution ends)
	if execCtx.CurrentAgent.IsTerminal {
		return response, nil
	}

	// Check for handoffs
	if len(execCtx.CurrentAgent.HandoffTargets) > 0 && execCtx.HandoffCount < execCtx.MaxHandoffs {
		// Route to next agent
		_ = execCtx.CurrentAgent.HandoffTargets[0] // Next agent ID for future implementation
		// In a full implementation, would look up agent from crew by ID
		// For now, return response
		return response, nil
	}

	return response, nil
}

// ExecuteWorkflowStream executes the workflow with streaming output
func ExecuteWorkflowStream(ctx context.Context, entryAgent *common.Agent, input string, history []common.Message, streamChan chan *StreamEvent, apiKey string) error {
	if entryAgent == nil {
		return fmt.Errorf("entry agent cannot be nil")
	}

	handler := NewStreamHandler(streamChan)

	execCtx := &ExecutionContext{
		CurrentAgent: entryAgent,
		History:      history,
		MaxHandoffs:  5,
		MaxRounds:    10,
		StartTime:    time.Now(),
		handler:      handler,
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
