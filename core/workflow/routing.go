// Package workflow provides workflow orchestration and execution functionality.
package workflow

import (
	"context"
	"fmt"

	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/signal"
)

// RoutingDecision represents the result of routing logic
type RoutingDecision struct {
	NextAgentID string
	Reason      string
	IsTerminal  bool
}

// DetermineNextAgent determines the next agent based on routing configuration
func DetermineNextAgent(currentAgent *common.Agent, response *common.AgentResponse, routing *common.RoutingConfig) (*RoutingDecision, error) {
	if currentAgent == nil {
		return nil, fmt.Errorf("current agent cannot be nil")
	}

	// Check if current agent is terminal
	if currentAgent.IsTerminal {
		return &RoutingDecision{
			IsTerminal: true,
			Reason:     "agent is marked as terminal",
		}, nil
	}

	// Check for handoff targets (placeholder for Phase 5)
	// In a full implementation, would check currentAgent.HandoffTargets
	// and route accordingly

	// No routing configured - execution ends
	return &RoutingDecision{
		IsTerminal: true,
		Reason:     "no handoff targets configured",
	}, nil
}

// DetermineNextAgentWithSignals determines next agent using signal priority
func DetermineNextAgentWithSignals(ctx context.Context, currentAgent *common.Agent, response *common.AgentResponse,
	routing *common.RoutingConfig, signalRegistry *signal.SignalRegistry) (*RoutingDecision, error) {

	if currentAgent == nil {
		return nil, fmt.Errorf("current agent cannot be nil")
	}

	// Priority 1: Check signals in response
	if signalRegistry != nil && response != nil && response.Signals != nil && len(response.Signals) > 0 {
		for _, sigName := range response.Signals {
			sig := &signal.Signal{
				Name:    sigName,
				AgentID: currentAgent.ID,
			}

			decision, err := signalRegistry.ProcessSignal(ctx, sig)
			if err == nil && decision != nil {
				if decision.IsTerminal || decision.NextAgentID != "" {
					return &RoutingDecision{
						NextAgentID: decision.NextAgentID,
						Reason:      decision.Reason,
						IsTerminal:  decision.IsTerminal,
					}, nil
				}
			}
		}
	}

	// Priority 2: Check terminal status
	if currentAgent.IsTerminal {
		return &RoutingDecision{
			IsTerminal: true,
			Reason:     "agent is marked as terminal",
		}, nil
	}

	// Priority 3: Check handoff targets
	if len(currentAgent.HandoffTargets) > 0 {
		return &RoutingDecision{
			NextAgentID: currentAgent.HandoffTargets[0].ID,
			Reason:      "default handoff target",
			IsTerminal:  false,
		}, nil
	}

	// No routing found
	return &RoutingDecision{
		IsTerminal: true,
		Reason:     "no handoff targets configured",
	}, nil
}

// RouteBySignal routes to an agent based on signal-based routing configuration
func RouteBySignal(signalName string, routing *common.RoutingConfig) (string, error) {
	if routing == nil {
		return "", fmt.Errorf("routing configuration is nil")
	}

	if signalName == "" {
		return "", fmt.Errorf("signal name is empty")
	}

	if routing.Signals == nil {
		return "", fmt.Errorf("no signals configured in routing")
	}

	// Search all agents for this signal
	for _, signalList := range routing.Signals {
		for _, sig := range signalList {
			if sig.Signal == signalName {
				return sig.Target, nil
			}
		}
	}

	return "", fmt.Errorf("signal '%s' not found in routing configuration", signalName)
}

// RouteByBehavior routes to an agent based on behavior-based routing configuration
func RouteByBehavior(behavior string, routing *common.RoutingConfig) (string, error) {
	if routing == nil {
		return "", fmt.Errorf("routing configuration is nil")
	}

	if behavior == "" {
		return "", fmt.Errorf("behavior is empty")
	}

	// Placeholder implementation for behavior-based routing
	// In a full implementation, would lookup behavior in routing.AgentBehaviors map
	// and return the target agent ID

	return "", fmt.Errorf("behavior '%s' not found in routing configuration", behavior)
}

// ValidateRouting validates the routing configuration
func ValidateRouting(routing *common.RoutingConfig, agents map[string]*common.Agent) error {
	if routing == nil {
		return nil // Routing is optional
	}

	// Validate signal targets exist
	if routing.Signals != nil {
		for agentID := range routing.Signals {
			if _, exists := agents[agentID]; !exists {
				return fmt.Errorf("signal source agent '%s' not found in agents map", agentID)
			}
		}
	}

	// Validate behavior targets exist
	if routing.AgentBehaviors != nil {
		for agentID := range routing.AgentBehaviors {
			if _, exists := agents[agentID]; !exists {
				return fmt.Errorf("behavior agent '%s' not found in agents map", agentID)
			}
		}
	}

	return nil
}
