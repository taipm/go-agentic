// Package routing provides agent routing decision logic for workflow orchestration.
package routing

import (
	"context"
	"fmt"

	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/signal"
)

// DetermineNextAgent determines the next agent based on routing configuration
func DetermineNextAgent(currentAgent *common.Agent, response *common.AgentResponse, routing *common.RoutingConfig) (*common.RoutingDecision, error) {
	if currentAgent == nil {
		return nil, common.NewValidationError("agent", "current agent cannot be nil")
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
		return nil, common.NewValidationError("agent", "current agent cannot be nil")
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
	}

	// Fallback: Use signal registry if available
	if response != nil && response.Signals != nil && len(response.Signals) > 0 && signalRegistry != nil {
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
