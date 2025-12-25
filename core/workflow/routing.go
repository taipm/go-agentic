// Package workflow provides workflow orchestration and execution functionality.
package workflow

import (
	"context"
	"fmt"

	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/signal"
)

// DetermineNextAgent determines the next agent based on routing configuration
func DetermineNextAgent(currentAgent *common.Agent, response *common.AgentResponse, routing *common.RoutingConfig) (*common.RoutingDecision, error) {
	if currentAgent == nil {
		return nil, fmt.Errorf("current agent cannot be nil")
	}

	// Check if current agent is terminal
	if currentAgent.IsTerminal {
		return &common.RoutingDecision{
			IsTerminal: true,
			Reason:     "agent is marked as terminal",
		}, nil
	}

	// Check for handoff targets (placeholder for Phase 5)
	// In a full implementation, would check currentAgent.HandoffTargets
	// and route accordingly

	// No routing configured - execution ends
	return &common.RoutingDecision{
		IsTerminal: true,
		Reason:     "no handoff targets configured",
	}, nil
}

// DetermineNextAgentWithSignals determines next agent using signal priority
func DetermineNextAgentWithSignals(ctx context.Context, currentAgent *common.Agent, response *common.AgentResponse,
	routing *common.RoutingConfig, signalRegistry *signal.SignalRegistry) (*common.RoutingDecision, error) {

	if currentAgent == nil {
		return nil, fmt.Errorf("current agent cannot be nil")
	}

	// Priority 1: Check signals in response using routing configuration
	if response != nil && response.Signals != nil && len(response.Signals) > 0 {
		// Check if routing is configured
		if routing != nil && routing.Signals != nil {
			// Get signal routing rules for current agent
			if agentSignals, exists := routing.Signals[currentAgent.ID]; exists {
				// Look for matching signal in routing rules
				for _, sigName := range response.Signals {
					for _, routingSignal := range agentSignals {
						if routingSignal.Signal == sigName {
							// Check for terminal signal
							if routingSignal.Target == "" {
								return &common.RoutingDecision{
									IsTerminal: sigName == "[END_EXAM]",
									Reason:     fmt.Sprintf("signal '%s' marks terminal", sigName),
								}, nil
							}
							// Return the target agent
							return &common.RoutingDecision{
								NextAgentID: routingSignal.Target,
								Reason:      fmt.Sprintf("routed by signal '%s'", sigName),
								IsTerminal:  false,
							}, nil
						}
					}
				}
			}
		}

		// Fallback: Use signal registry if available
		if signalRegistry != nil {
			for _, sigName := range response.Signals {
				sig := &signal.Signal{
					Name:    sigName,
					AgentID: currentAgent.ID,
				}

				decision, err := signalRegistry.ProcessSignal(ctx, sig)
				if err == nil && decision != nil {
					if decision.IsTerminal || decision.NextAgentID != "" {
						return &common.RoutingDecision{
							NextAgentID: decision.NextAgentID,
							Reason:      decision.Reason,
							IsTerminal:  decision.IsTerminal,
							Metadata:    decision.Metadata,
						}, nil
					}
				}
			}
		}
	}

	// Priority 2: Check terminal status
	if currentAgent.IsTerminal {
		return &common.RoutingDecision{
			IsTerminal: true,
			Reason:     "agent is marked as terminal",
		}, nil
	}

	// Priority 3: Check handoff targets
	if len(currentAgent.HandoffTargets) > 0 {
		return &common.RoutingDecision{
			NextAgentID: currentAgent.HandoffTargets[0].ID,
			Reason:      "default handoff target",
			IsTerminal:  false,
		}, nil
	}

	// No routing found - but don't mark as terminal, allow continuation
	// Let other routing mechanisms handle it
	return &common.RoutingDecision{
		IsTerminal:  false,
		NextAgentID: "",
		Reason:      "no signal routing configured, will use default routing",
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
