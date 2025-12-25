// Package validation provides configuration validation rules.
package validation

import (
	"fmt"

	"github.com/taipm/go-agentic/core/common"
)

// ValidateSignals validates routing signals reference valid agents and formats
func ValidateSignals(signals map[string]interface{}, agentMap map[string]bool) error {
	if signals == nil {
		return nil
	}

	for agentID, signalList := range signals {
		if !agentMap[agentID] {
			return &common.ValidationError{
				Field:   fmt.Sprintf("signals[%s]", agentID),
				Message: fmt.Sprintf("references non-existent agent '%s'", agentID),
			}
		}

		// Type assertion to handle the signal list
		signalArray, ok := signalList.([]interface{})
		if !ok {
			continue
		}

		for _, sig := range signalArray {
			signalStr, ok := sig.(string)
			if !ok {
				continue
			}

			if signalStr == "" {
				return &common.ValidationError{
					Field:   fmt.Sprintf("signals[%s]", agentID),
					Message: "signal name cannot be empty - must be in [NAME] format",
				}
			}

			if !isSignalFormatValid(signalStr) {
				return &common.ValidationError{
					Field:   fmt.Sprintf("signals[%s]", agentID),
					Message: fmt.Sprintf("invalid signal format '%s' - must be in [NAME] format (e.g., [END_EXAM])", signalStr),
				}
			}
		}
	}
	return nil
}

// ValidateParallelGroups validates parallel groups structure and references
func ValidateParallelGroups(groups map[string]interface{}, agentMap map[string]bool) error {
	if groups == nil {
		return nil
	}

	for groupName := range groups {
		if _, exists := agentMap[groupName]; !exists {
			// Group names don't have to match agent names
			continue
		}
	}
	return nil
}

// DetectCircularReferences detects cycles in the agent routing graph
func DetectCircularReferences(config *common.CrewConfig, agents map[string]*common.AgentConfig) error {
	if config == nil || agents == nil {
		return nil
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for agentID := range agents {
		if !visited[agentID] {
			if hasCycle(agentID, visited, recStack, agents) {
				return &common.ValidationError{
					Field:   "routing",
					Message: fmt.Sprintf("circular reference detected involving agent '%s'", agentID),
				}
			}
		}
	}

	return nil
}

// hasCycle is a helper function for detecting cycles using DFS
func hasCycle(agentID string, visited, recStack map[string]bool, agents map[string]*common.AgentConfig) bool {
	visited[agentID] = true
	recStack[agentID] = true

	agent, exists := agents[agentID]
	if !exists || agent == nil {
		recStack[agentID] = false
		return false
	}

	// Check handoff targets for cycles (stored as agent ID strings in AgentConfig)
	if len(agent.HandoffTargets) > 0 {
		for _, targetID := range agent.HandoffTargets {
			if targetID == "" {
				continue
			}
			if !visited[targetID] {
				if hasCycle(targetID, visited, recStack, agents) {
					return true
				}
			} else if recStack[targetID] {
				return true
			}
		}
	}

	recStack[agentID] = false
	return false
}

// CheckReachability checks if all agents are reachable from the entry point
func CheckReachability(config *common.CrewConfig, agents map[string]*common.AgentConfig) error {
	if config == nil || agents == nil {
		return nil
	}

	if config.EntryPoint == "" {
		return nil
	}

	visited := make(map[string]bool)
	queue := []string{config.EntryPoint}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		agent, exists := agents[current]
		if !exists || agent == nil {
			continue
		}

		// Add handoff targets to queue (stored as agent ID strings in AgentConfig)
		if len(agent.HandoffTargets) > 0 {
			for _, targetID := range agent.HandoffTargets {
				if targetID != "" && !visited[targetID] {
					queue = append(queue, targetID)
				}
			}
		}
	}

	// Check if all agents were visited
	for agentID := range agents {
		if !visited[agentID] {
			return &common.ValidationError{
				Field:   "entry_point",
				Message: fmt.Sprintf("agent '%s' is not reachable from entry point '%s'", agentID, config.EntryPoint),
			}
		}
	}

	return nil
}

// isSignalFormatValid checks if a signal matches the [NAME] format
func isSignalFormatValid(signal string) bool {
	if len(signal) < 3 {
		return false // Minimum: [X]
	}
	// Must start with [ and end with ]
	if signal[0] != '[' || signal[len(signal)-1] != ']' {
		return false
	}
	// Must have content inside brackets
	inner := signal[1 : len(signal)-1]
	return len(inner) > 0
}
