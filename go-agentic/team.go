package agentic

import (
	"context"
	"fmt"
	"strings"
)

// TeamExecutor handles the execution of a team
type TeamExecutor struct {
	team       *Team
	apiKey     string
	entryAgent *Agent
	history    []Message
}

// NewTeamExecutor creates a new team executor
func NewTeamExecutor(team *Team, apiKey string) *TeamExecutor {
	// Find entry agent (first agent that's not terminal)
	var entryAgent *Agent
	for _, agent := range team.Agents {
		if !agent.IsTerminal {
			entryAgent = agent
			break
		}
	}

	return &TeamExecutor{
		team:       team,
		apiKey:     apiKey,
		entryAgent: entryAgent,
		history:    []Message{},
	}
}

// ExecuteStream runs the team with streaming events
func (te *TeamExecutor) ExecuteStream(ctx context.Context, input string, streamChan chan *StreamEvent) error {
	// Add user input to history
	te.history = append(te.history, Message{
		Role:    "user",
		Content: input,
	})

	// Start with entry agent
	currentAgent := te.entryAgent
	if currentAgent == nil {
		return fmt.Errorf("no entry agent found")
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
		response, err := ExecuteAgent(ctx, currentAgent, input, te.history, te.apiKey)
		if err != nil {
			streamChan <- NewStreamEvent("error", currentAgent.Name, fmt.Sprintf("Agent failed: %v", err))
			return fmt.Errorf("agent %s failed: %w", currentAgent.ID, err)
		}

		// Send agent response event
		streamChan <- NewStreamEvent("agent_response", currentAgent.Name, response.Content)

		// Add agent response to history
		te.history = append(te.history, Message{
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

			toolResults := te.executeCalls(ctx, response.ToolCalls, currentAgent)

			// Send tool results
			for _, result := range toolResults {
				status := "✅"
				if result.Status == "error" {
					status = "❌"
				}
				streamChan <- NewStreamEvent("tool_result", currentAgent.Name,
					fmt.Sprintf("%s [Tool] %s → %s", status, result.ToolName, result.Output))
			}

			// Format results for feedback
			resultText := formatToolResults(toolResults)

			// Add results to history
			te.history = append(te.history, Message{
				Role:    "user",
				Content: resultText,
			})

			// Feed results back to current agent for analysis
			input = resultText
			// Continue loop to let agent process results
			continue
		}

		// Check if current agent is terminal (only after tool execution)
		if currentAgent.IsTerminal {
			return nil
		}

		// Check for routing signals from current agent (config-driven)
		nextAgent := te.findNextAgentBySignal(currentAgent, response.Content)
		if nextAgent != nil {
			currentAgent = nextAgent
			input = response.Content
			handoffCount++
			continue
		}

		// Check if agent waits for signal (from config)
		behavior := te.getAgentBehavior(currentAgent.ID)
		if behavior != nil && behavior.WaitForSignal {
			// Agent waits for explicit signal, send pause event and stop streaming
			streamChan <- NewStreamEvent("pause", currentAgent.Name, "[PAUSE] Waiting for user input")
			return nil
		}

		// For other agents, handoff normally
		handoffCount++
		if handoffCount >= te.team.MaxHandoffs {
			return nil
		}

		// Find next agent based on handoff targets
		nextAgent = te.findNextAgent(currentAgent)
		if nextAgent == nil {
			return nil
		}

		currentAgent = nextAgent
		input = response.Content
	}
}

// Execute runs the team with the given input
func (te *TeamExecutor) Execute(ctx context.Context, input string) (*TeamResponse, error) {
	// Add user input to history
	te.history = append(te.history, Message{
		Role:    "user",
		Content: input,
	})

	// Start with entry agent
	currentAgent := te.entryAgent
	if currentAgent == nil {
		return nil, fmt.Errorf("no entry agent found")
	}

	handoffCount := 0

	for {
		// Execute current agent
		response, err := ExecuteAgent(ctx, currentAgent, input, te.history, te.apiKey)
		if err != nil {
			return nil, fmt.Errorf("agent %s failed: %w", currentAgent.ID, err)
		}

		fmt.Printf("\n[%s]: %s\n", currentAgent.Name, response.Content)

		// Add agent response to history
		te.history = append(te.history, Message{
			Role:    "assistant",
			Content: response.Content,
		})

		// Execute any tool calls BEFORE checking if terminal
		if len(response.ToolCalls) > 0 {
			toolResults := te.executeCalls(ctx, response.ToolCalls, currentAgent)

			// Format results for feedback
			resultText := formatToolResults(toolResults)
			fmt.Println(resultText)

			// Add results to history
			te.history = append(te.history, Message{
				Role:    "user",
				Content: resultText,
			})

			// Feed results back to current agent for analysis
			input = resultText
			// Continue loop to let agent process results
			continue
		}

		// Check if current agent is terminal (only after tool execution)
		if currentAgent.IsTerminal {
			return &TeamResponse{
				AgentID:    currentAgent.ID,
				AgentName:  currentAgent.Name,
				Content:    response.Content,
				ToolCalls:  response.ToolCalls,
				IsTerminal: true,
			}, nil
		}

		// Check for routing signals from current agent (config-driven)
		nextAgent := te.findNextAgentBySignal(currentAgent, response.Content)
		if nextAgent != nil {
			currentAgent = nextAgent
			input = response.Content
			handoffCount++
			continue
		}

		// Check if agent waits for signal (from config)
		behavior := te.getAgentBehavior(currentAgent.ID)
		if behavior != nil && behavior.WaitForSignal {
			// Agent waits for explicit signal, return and wait for next input
			return &TeamResponse{
				AgentID:   currentAgent.ID,
				AgentName: currentAgent.Name,
				Content:   response.Content,
				ToolCalls: response.ToolCalls,
			}, nil
		}

		// For other agents, handoff normally
		handoffCount++
		if handoffCount >= te.team.MaxHandoffs {
			return &TeamResponse{
				AgentID:   currentAgent.ID,
				AgentName: currentAgent.Name,
				Content:   response.Content,
				ToolCalls: response.ToolCalls,
			}, nil
		}

		// Find next agent based on handoff targets
		nextAgent = te.findNextAgent(currentAgent)
		if nextAgent == nil {
			return &TeamResponse{
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
func (te *TeamExecutor) executeCalls(ctx context.Context, calls []ToolCall, agent *Agent) []ToolResult {
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

		fmt.Printf("[TOOL RESULT] %s: %s\n", call.ToolName, output)
	}

	return results
}

// findAgentByID finds an agent by its ID
func (te *TeamExecutor) findAgentByID(id string) *Agent {
	for _, agent := range te.team.Agents {
		if agent.ID == id {
			return agent
		}
	}
	return nil
}

// findNextAgentBySignal finds the next agent based on routing signals (config-driven)
func (te *TeamExecutor) findNextAgentBySignal(current *Agent, responseContent string) *Agent {
	if te.team.Routing == nil {
		return nil
	}

	// Get signals defined for current agent in config
	signals, exists := te.team.Routing.Signals[current.ID]
	if !exists || len(signals) == 0 {
		return nil
	}

	// Check which signal is present in the response
	for _, sig := range signals {
		if strings.Contains(responseContent, sig.Signal) && sig.Target != "" {
			// Found matching signal, find the target agent
			return te.findAgentByID(sig.Target)
		}
	}

	return nil
}

// getAgentBehavior retrieves behavior config for an agent
func (te *TeamExecutor) getAgentBehavior(agentID string) *AgentBehavior {
	if te.team.Routing == nil || te.team.Routing.AgentBehaviors == nil {
		return nil
	}
	behavior, exists := te.team.Routing.AgentBehaviors[agentID]
	if !exists {
		return nil
	}
	return &behavior
}

// findNextAgent finds the next appropriate agent for handoff
func (te *TeamExecutor) findNextAgent(current *Agent) *Agent {
	// First, try to use handoff_targets from current agent config
	if len(current.HandoffTargets) > 0 {
		// Create a map of agents by ID for quick lookup
		agentMap := make(map[string]*Agent)
		for _, agent := range te.team.Agents {
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
	for _, agent := range te.team.Agents {
		if agent.ID != current.ID {
			return agent
		}
	}

	return nil
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

// Deprecated: Use NewTeamExecutor instead
func NewCrewExecutor(team *Team, apiKey string) *TeamExecutor {
	return NewTeamExecutor(team, apiKey)
}
