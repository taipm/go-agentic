// Package routing provides agent routing decision logic for workflow orchestration.
package routing

import (
	"fmt"

	"github.com/taipm/go-agentic/core/common"
)

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
