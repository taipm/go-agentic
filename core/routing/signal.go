// Package routing provides agent routing decision logic for workflow orchestration.
package routing

import (
	"fmt"

	"github.com/taipm/go-agentic/core/common"
)

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
// Validates that a behavior exists in the routing configuration and returns the behavior config
// The behavior acts as a routing key that can be used to determine agent flow characteristics
func RouteByBehavior(behavior string, routing *common.RoutingConfig) (string, error) {
	if routing == nil {
		return "", fmt.Errorf("routing configuration is nil")
	}

	if behavior == "" {
		return "", fmt.Errorf("behavior name is empty")
	}

	if routing.AgentBehaviors == nil || len(routing.AgentBehaviors) == 0 {
		return "", fmt.Errorf("no agent behaviors configured in routing")
	}

	// Lookup behavior in routing configuration
	behaviorConfig, exists := routing.AgentBehaviors[behavior]
	if !exists {
		return "", fmt.Errorf("behavior '%s' not found in routing configuration", behavior)
	}

	// Validate behavior configuration exists (AgentBehavior is a value type, not pointer)
	// Return the behavior name as the routing key
	// The behavior name can be used by higher-level routing to determine next steps
	// Based on behavior properties like: AutoRoute, IsTerminal, WaitForSignal
	_ = behaviorConfig // Reference behavior config to prevent unused variable

	return behavior, nil
}
