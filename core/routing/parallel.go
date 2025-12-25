// Package routing provides agent routing decision logic for workflow orchestration.
package routing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/taipm/go-agentic/core/common"
)

// ParallelResult represents the result of executing an agent in parallel
type ParallelResult struct {
	AgentID  string
	Response *common.AgentResponse
	Error    error
	Duration time.Duration
}

// ParallelGroupExecutor manages concurrent execution of agent groups
type ParallelGroupExecutor struct {
	Config  *common.ParallelGroupConfig
	Agents  map[string]*common.Agent
	Timeout time.Duration
}

// NewParallelGroupExecutor creates a new executor for parallel agent groups
func NewParallelGroupExecutor(config *common.ParallelGroupConfig, agents map[string]*common.Agent) *ParallelGroupExecutor {
	timeout := 30 * time.Second // Default timeout
	if config != nil && config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}

	return &ParallelGroupExecutor{
		Config:  config,
		Agents:  agents,
		Timeout: timeout,
	}
}

// ExecuteParallel executes multiple agents concurrently
// Returns results in the order they complete (not in agent order)
func (pge *ParallelGroupExecutor) ExecuteParallel(
	ctx context.Context,
	agents []*common.Agent,
	input string,
	history []common.Message,
	apiKey string,
) ([]ParallelResult, error) {

	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents provided for parallel execution")
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, pge.Timeout)
	defer cancel()

	// Channel to collect results
	resultsChan := make(chan ParallelResult, len(agents))
	var wg sync.WaitGroup

	// Launch goroutine for each agent
	for _, agent := range agents {
		if agent == nil {
			continue
		}

		wg.Add(1)
		go func(ag *common.Agent) {
			defer wg.Done()

			startTime := time.Now()
			response, err := executeAgentConcurrent(execCtx, ag, input, history, apiKey)
			duration := time.Since(startTime)

			resultsChan <- ParallelResult{
				AgentID:  ag.ID,
				Response: response,
				Error:    err,
				Duration: duration,
			}
		}(agent)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	results := make([]ParallelResult, 0, len(agents))
	for result := range resultsChan {
		results = append(results, result)
	}

	return results, nil
}

// ExecuteWithWaitForAll waits for all agents to complete or timeout
// Returns error if any agent fails and WaitForAll is true
func (pge *ParallelGroupExecutor) ExecuteWithWaitForAll(
	ctx context.Context,
	agents []*common.Agent,
	input string,
	history []common.Message,
	apiKey string,
) ([]ParallelResult, error) {

	results, err := pge.ExecuteParallel(ctx, agents, input, history, apiKey)
	if err != nil {
		return nil, err
	}

	// Check for failures if WaitForAll is true
	if pge.Config != nil && pge.Config.WaitForAll {
		for _, result := range results {
			if result.Error != nil {
				return results, fmt.Errorf("agent '%s' failed: %w", result.AgentID, result.Error)
			}
		}
	}

	return results, nil
}

// ExecuteWithFirstSuccess returns the first successful result
// Returns error if all agents fail
func (pge *ParallelGroupExecutor) ExecuteWithFirstSuccess(
	ctx context.Context,
	agents []*common.Agent,
	input string,
	history []common.Message,
	apiKey string,
) (*ParallelResult, error) {

	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents provided for parallel execution")
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, pge.Timeout)
	defer cancel()

	// Channel to collect first success
	successChan := make(chan ParallelResult, 1)
	var wg sync.WaitGroup

	// Launch goroutine for each agent
	for _, agent := range agents {
		if agent == nil {
			continue
		}

		wg.Add(1)
		go func(ag *common.Agent) {
			defer wg.Done()

			startTime := time.Now()
			response, err := executeAgentConcurrent(execCtx, ag, input, history, apiKey)
			duration := time.Since(startTime)

			// Send successful result to channel (non-blocking)
			if err == nil {
				select {
				case successChan <- ParallelResult{
					AgentID:  ag.ID,
					Response: response,
					Error:    nil,
					Duration: duration,
				}:
				default:
					// Another agent already succeeded
				}
			}
		}(agent)
	}

	// Wait for all goroutines to complete in background
	go func() {
		wg.Wait()
		close(successChan)
	}()

	// Wait for first success or timeout
	select {
	case result, ok := <-successChan:
		if ok && result.Response != nil {
			return &result, nil
		}
	case <-execCtx.Done():
		return nil, fmt.Errorf("parallel execution timeout after %v", pge.Timeout)
	}

	return nil, fmt.Errorf("all agents failed in parallel execution")
}

// executeAgentConcurrent executes a single agent (concurrent-safe wrapper)
func executeAgentConcurrent(
	ctx context.Context,
	agent *common.Agent,
	input string,
	history []common.Message,
	apiKey string,
) (*common.AgentResponse, error) {

	if agent == nil {
		return nil, fmt.Errorf("agent cannot be nil")
	}

	// Import from agent package to avoid circular imports
	// This would be called from the agent execution module
	// For now, return a marker that this would execute the agent
	_ = ctx
	_ = input
	_ = history
	_ = apiKey

	return nil, fmt.Errorf("agent execution not implemented in this context")
}

// ExecuteGroupWithStrategy executes agents based on the parallel group strategy
// WaitForAll=true: waits for all agents, returns error if any fail
// WaitForAll=false: returns first successful result
func (pge *ParallelGroupExecutor) ExecuteGroupWithStrategy(
	ctx context.Context,
	agents []*common.Agent,
	input string,
	history []common.Message,
	apiKey string,
) (interface{}, error) {

	if pge.Config == nil {
		return nil, fmt.Errorf("parallel group config is nil")
	}

	if pge.Config.WaitForAll {
		// Wait for all agents to complete
		return pge.ExecuteWithWaitForAll(ctx, agents, input, history, apiKey)
	}

	// Return first successful result
	return pge.ExecuteWithFirstSuccess(ctx, agents, input, history, apiKey)
}

// ValidateParallelGroup validates parallel group configuration
func ValidateParallelGroup(group *common.ParallelGroupConfig, agents map[string]*common.Agent) error {
	if group == nil {
		return fmt.Errorf("parallel group config is nil")
	}

	if len(group.Agents) == 0 {
		return fmt.Errorf("parallel group must have at least one agent")
	}

	// Validate all agents exist
	for _, agentID := range group.Agents {
		if agentID == "" {
			return fmt.Errorf("parallel group contains empty agent ID")
		}

		if _, exists := agents[agentID]; !exists {
			return fmt.Errorf("agent '%s' in parallel group not found", agentID)
		}
	}

	// Validate timeout is reasonable (at least 1 second)
	if group.TimeoutSeconds > 0 && group.TimeoutSeconds < 1 {
		return fmt.Errorf("parallel group timeout must be at least 1 second, got %d", group.TimeoutSeconds)
	}

	return nil
}

// GetAgentsForParallelGroup returns agent pointers for a parallel group
func GetAgentsForParallelGroup(group *common.ParallelGroupConfig, agents map[string]*common.Agent) ([]*common.Agent, error) {
	if group == nil {
		return nil, fmt.Errorf("parallel group config is nil")
	}

	result := make([]*common.Agent, 0, len(group.Agents))

	for _, agentID := range group.Agents {
		agent, exists := agents[agentID]
		if !exists {
			return nil, fmt.Errorf("agent '%s' not found in agents map", agentID)
		}
		result = append(result, agent)
	}

	return result, nil
}
