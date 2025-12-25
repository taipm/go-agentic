// Package executor provides the main execution orchestrator for crews.
package executor

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/workflow"
)

// Executor orchestrates the execution of a crew
type Executor struct {
	crew          *common.Crew
	apiKey        string
	entryAgent    *common.Agent
	verbose       bool
	resumeAgentID string
}

// NewExecutor creates a new executor for a crew
func NewExecutor(crew *common.Crew, apiKey string) (*Executor, error) {
	if crew == nil {
		return nil, fmt.Errorf("crew cannot be nil")
	}

	if len(crew.Agents) == 0 {
		return nil, fmt.Errorf("crew must have at least one agent")
	}

	// Find entry agent (default to first agent)
	// Note: crew.EntryPoint is in CrewConfig (parsed from YAML), not in Crew struct
	// For now, we always use the first agent. Full implementation would pass
	// the resolved entry point from the config layer.
	entryAgent := crew.Agents[0]

	return &Executor{
		crew:       crew,
		apiKey:     apiKey,
		entryAgent: entryAgent,
	}, nil
}

// SetVerbose enables or disables verbose output
func (e *Executor) SetVerbose(verbose bool) {
	if e != nil {
		e.verbose = verbose
	}
}

// SetResumeAgent sets the agent to resume execution from
func (e *Executor) SetResumeAgent(agentID string) {
	if e != nil {
		e.resumeAgentID = agentID
	}
}

// ClearResumeAgent clears the resume agent setting
func (e *Executor) ClearResumeAgent() {
	if e != nil {
		e.resumeAgentID = ""
	}
}

// GetResumeAgentID returns the resume agent ID
func (e *Executor) GetResumeAgentID() string {
	if e == nil {
		return ""
	}
	return e.resumeAgentID
}

// Execute runs the executor synchronously
func (e *Executor) Execute(ctx context.Context, input string) (*common.CrewResponse, error) {
	if e == nil {
		return nil, fmt.Errorf("executor is nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// Create a sync handler
	handler := workflow.NewSyncHandler()

	// Determine entry agent
	entryAgent := e.entryAgent
	if e.resumeAgentID != "" {
		for _, agent := range e.crew.Agents {
			if agent.ID == e.resumeAgentID {
				entryAgent = agent
				break
			}
		}
	}

	// Execute workflow
	history := []common.Message{}
	response, err := workflow.ExecuteWorkflow(ctx, entryAgent, input, history, handler, nil, e.apiKey)

	if err != nil {
		if e.verbose {
			log.Printf("[EXECUTOR ERROR] %v", err)
		}
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	if e.verbose {
		log.Printf("[EXECUTOR] Execution completed: %s", response.Content)
	}

	// Convert agent response to crew response
	crewResponse := &common.CrewResponse{
		AgentID:   response.AgentID,
		AgentName: response.AgentName,
		Content:   response.Content,
	}

	return crewResponse, nil
}

// ExecuteStream runs the executor with streaming output
func (e *Executor) ExecuteStream(ctx context.Context, input string, streamChan chan *common.StreamEvent) error {
	if e == nil {
		return fmt.Errorf("executor is nil")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if streamChan == nil {
		return fmt.Errorf("stream channel cannot be nil")
	}

	// Determine entry agent
	entryAgent := e.entryAgent
	if e.resumeAgentID != "" {
		for _, agent := range e.crew.Agents {
			if agent.ID == e.resumeAgentID {
				entryAgent = agent
				break
			}
		}
	}

	// Execute workflow with streaming
	history := []common.Message{}
	err := workflow.ExecuteWorkflowStream(ctx, entryAgent, input, history, streamChan, nil, e.apiKey)

	if err != nil {
		if e.verbose {
			log.Printf("[EXECUTOR ERROR] %v", err)
		}
		return fmt.Errorf("workflow streaming failed: %w", err)
	}

	return nil
}
