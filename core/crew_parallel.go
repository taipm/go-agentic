package crewai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// =============================================================================
// PARALLEL EXECUTION LOGIC
// Parallel agent execution and result aggregation
// Extracted from crew.go to reduce file size
// =============================================================================

// DefaultParallelAgentTimeout is the default timeout for each parallel agent execution
// ✅ FIX #4: Now uses Crew.ParallelAgentTimeout field for configurability
const DefaultParallelAgentTimeout = 60 * time.Second

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

	// ✅ Phase 5: Determine timeout - use defaults from HardcodedDefaults
	parallelTimeout := ce.defaults.ParallelAgentTimeout
	if parallelTimeout <= 0 {
		parallelTimeout = 60 * time.Second // Fallback
	}

	// Launch all agents in parallel using goroutines
	for _, agent := range agents {
		wg.Add(1)
		go func(ag *Agent) {
			defer wg.Done()

			// Create timeout context for this agent (✅ FIX #4: Now configurable)
			agentCtx, cancel := context.WithTimeout(ctx, parallelTimeout)
			defer cancel()

			// Send agent start event
			streamChan <- NewStreamEvent("agent_start", ag.Name,
				fmt.Sprintf("🔄 [Parallel] %s starting...", ag.Name))

			// ✅ Track execution time for metrics
			agentStartTime := time.Now()

			// Execute the agent with timeout
			response, err := ExecuteAgent(agentCtx, ag, input, ce.getHistoryCopy(), ce.apiKey)
			agentDuration := time.Since(agentStartTime)

			if err != nil {
			// ✅ ISSUE #2: Update performance metrics FIRST with error
			if ag.Metadata != nil {
			ag.UpdatePerformanceMetrics(false, err.Error())
			}

			// ✅ ISSUE #2: Check error quota (use different variable to avoid shadowing)
			if quotaErr := ag.CheckErrorQuota(); quotaErr != nil {
			 log.Printf("[QUOTA] Agent %s exceeded error quota: %v", ag.ID, quotaErr)
			  streamChan <- NewStreamEvent("error", ag.Name,
					fmt.Sprintf("Error quota exceeded: %v", quotaErr))
				errorChan <- quotaErr
				return
			}

			// Original error handling
			if ce.Metrics != nil {
				ce.Metrics.RecordAgentExecution(ag.ID, ag.Name, agentDuration, false)
			}
			streamChan <- NewStreamEvent("error", ag.Name,
				fmt.Sprintf("❌ Agent failed: %v", err))
			errorChan <- fmt.Errorf("agent %s failed: %w", ag.ID, err)
			return
		}

			// ✅ Record successful parallel agent execution and cost
			if ce.Metrics != nil {
				ce.Metrics.RecordAgentExecution(ag.ID, ag.Name, agentDuration, true)

				// ✅ Crew-level cost tracking for parallel agents
				tokens, cost := ag.GetLastCallCost()
				ce.Metrics.RecordLLMCall(ag.ID, tokens, cost)

				// ✅ ISSUE #1: Check memory quota AFTER execution, BEFORE metrics update
				// Note: Memory is estimated based on token count. Actual usage may differ.
				// Formula: 1 token ≈ 4 bytes (conservative estimate)
				memoryUsedMB := (tokens * 4) / 1024 / 1024

				if err := ag.CheckMemoryQuota(); err != nil {
					log.Printf("[QUOTA] Agent %s exceeded memory quota: %v", ag.ID, err)
					streamChan <- NewStreamEvent("error", ag.Name,
						fmt.Sprintf("Memory quota exceeded: %v", err))
					errorChan <- err
					return
				}

				// ✅ ISSUE #3: Update memory & performance metrics
				if ag.Metadata != nil {
					ag.UpdateMemoryMetrics(memoryUsedMB, agentDuration.Milliseconds())
					ag.UpdatePerformanceMetrics(true, "")
				}
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

	// ✅ Phase 5: Determine timeout - use defaults from HardcodedDefaults
	// ✅ FIX #4: Now uses configurable Crew.ParallelAgentTimeout
	parallelTimeout := ce.defaults.ParallelAgentTimeout
	if parallelTimeout <= 0 {
		parallelTimeout = 60 * time.Second // Fallback
	}

	// Thread-safe result map
	resultMap := make(map[string]*AgentResponse)
	resultMutex := sync.Mutex{}

	// Launch all agents in parallel
	for _, agent := range agents {
		ag := agent // Capture for closure (important!)

		g.Go(func() error {
			if ce.Verbose {
				fmt.Printf("\n🔄 [Parallel] %s starting...\n", ag.Name)
			}

			// Create timeout context for this agent
			// gctx automatically propagates cancellation from parent or if another goroutine errors
			agentCtx, cancel := context.WithTimeout(gctx, parallelTimeout)
			defer cancel()

			// ✅ Track execution time for metrics
			agentStartTime := time.Now()

			// Execute the agent with timeout
			// If agentCtx is cancelled, ExecuteAgent should return immediately
			response, err := ExecuteAgent(agentCtx, ag, input, ce.getHistoryCopy(), ce.apiKey)
			agentDuration := time.Since(agentStartTime)

			if err != nil {
				// ✅ Record failed parallel agent execution
				if ce.Metrics != nil {
					ce.Metrics.RecordAgentExecution(ag.ID, ag.Name, agentDuration, false)
				}
				if ce.Verbose {
					fmt.Printf("❌ [Parallel] %s failed: %v\n", ag.Name, err)
				}
				// Return error - this will cancel all other goroutines automatically
				return fmt.Errorf("agent %s failed: %w", ag.ID, err)
			}

			// ✅ Record successful parallel agent execution and cost
			if ce.Metrics != nil {
				ce.Metrics.RecordAgentExecution(ag.ID, ag.Name, agentDuration, true)

				// ✅ Crew-level cost tracking for parallel agents
				tokens, cost := ag.GetLastCallCost()
				ce.Metrics.RecordLLMCall(ag.ID, tokens, cost)
			}

			if ce.Verbose {
				fmt.Printf("\n[%s]: %s\n", ag.Name, response.Content)
			}

			// Execute tool calls if any
			if len(response.ToolCalls) > 0 {
				// Pass agentCtx to executeCalls for proper cancellation support
				toolResults := ce.executeCalls(agentCtx, response.ToolCalls, ag)

				if ce.Verbose {
					resultText := ce.formatToolResults(toolResults)
					fmt.Println(resultText)
				}
			}

			// Store result thread-safely
			resultMutex.Lock()
			resultMap[response.AgentID] = response
			resultMutex.Unlock()

			return nil // ✅ Goroutine completes, cleaned up automatically
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
