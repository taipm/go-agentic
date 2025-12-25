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
