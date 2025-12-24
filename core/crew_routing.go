package crewai

import (
	"log"
	"strings"
)

// =============================================================================
// ROUTING LOGIC
// Agent routing and signal-based handoff logic
// Extracted from crew.go to reduce file size
// =============================================================================

// findAgentByID finds an agent by its ID
func (ce *CrewExecutor) findAgentByID(id string) *Agent {
	for _, agent := range ce.crew.Agents {
		if agent.ID == id {
			return agent
		}
	}
	return nil
}

// normalizeSignalText normalizes text for signal matching
// - Converts to lowercase
// - Normalizes whitespace inside brackets
// - Trims outer whitespace
func normalizeSignalText(text string) string {
	// Convert to lowercase for case-insensitive matching
	text = strings.ToLower(text)

	// Normalize whitespace around and inside brackets
	// "[  KẾT THÚC  ]" -> "[kết thúc]"
	text = strings.TrimSpace(text)

	// If it's a bracketed signal, normalize internal spaces
	if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
		inner := strings.TrimPrefix(strings.TrimSuffix(text, "]"), "[")
		inner = strings.TrimSpace(inner)
		// Collapse multiple spaces into single space
		parts := strings.Fields(inner)
		inner = strings.Join(parts, " ")
		text = "[" + inner + "]"
	}

	return text
}

// signalMatchesContent checks if a signal appears in response content (handles variations)
// ✅ ENHANCED: Case-insensitive + whitespace normalization for robust matching
// Handles: "[KẾT THÚC THI]" == "[ Kết thúc thi ]" == "[kết thúc thi]"
func signalMatchesContent(signal, content string) bool {
	// 1. Exact match first (fastest)
	if strings.Contains(content, signal) {
		return true
	}

	// 2. Case-insensitive match
	signalLower := strings.ToLower(signal)
	contentLower := strings.ToLower(content)
	if strings.Contains(contentLower, signalLower) {
		return true
	}

	// 3. Normalized match for bracketed signals
	// "[KẾT THÚC THI]" should match "[ Kết thúc thi ]"
	if strings.HasPrefix(signal, "[") && strings.HasSuffix(signal, "]") {
		normalizedSignal := normalizeSignalText(signal)

		// Find all bracketed patterns in content and compare normalized
		// Use a sliding window to find "[...]" patterns
		for i := 0; i < len(content); i++ {
			if content[i] == '[' {
				// Find matching "]"
				for j := i + 1; j < len(content) && j < i+100; j++ {
					if content[j] == ']' {
						bracketContent := content[i : j+1]
						normalizedContent := normalizeSignalText(bracketContent)
						if normalizedContent == normalizedSignal {
							return true
						}
						break
					}
				}
			}
		}
	}

	return false
}

// TerminationResult indicates whether workflow should terminate
type TerminationResult struct {
	ShouldTerminate bool
	Signal          string
}

// checkTerminationSignal checks if current agent's response contains a termination signal
// Termination signals have Target == "" in config
func (ce *CrewExecutor) checkTerminationSignal(current *Agent, responseContent string) *TerminationResult {
	if ce.crew.Routing == nil {
		log.Printf("[SIGNAL-DEBUG] Agent %s: No routing configured", current.ID)
		return nil
	}

	signals, exists := ce.crew.Routing.Signals[current.ID]
	if !exists || len(signals) == 0 {
		log.Printf("[SIGNAL-DEBUG] Agent %s: No signals configured", current.ID)
		return nil
	}

	log.Printf("[SIGNAL-CHECK] Agent %s: Checking %d termination signals in response", current.ID, len(signals))
	for _, sig := range signals {
		// Termination signal: Target is empty string
		if sig.Target == "" {
			log.Printf("[SIGNAL-MATCH] Agent %s: Testing termination signal %s", current.ID, sig.Signal)
			if signalMatchesContent(sig.Signal, responseContent) {
				log.Printf("[SIGNAL-FOUND] Agent %s emitted termination signal: %s", current.ID, sig.Signal)
				return &TerminationResult{
					ShouldTerminate: true,
					Signal:          sig.Signal,
				}
			}
		}
	}

	log.Printf("[SIGNAL-NO-TERMINATION] Agent %s: No termination signal detected", current.ID)
	return nil
}

// findNextAgentBySignal finds the next agent based on routing signals (config-driven)
// Returns nil if no routing signal found (not termination - use checkTerminationSignal for that)
func (ce *CrewExecutor) findNextAgentBySignal(current *Agent, responseContent string) *Agent {
	if ce.crew.Routing == nil {
		log.Printf("[SIGNAL-DEBUG] Agent %s: No routing configured for signal-based routing", current.ID)
		return nil
	}

	// Get signals defined for current agent in config
	signals, exists := ce.crew.Routing.Signals[current.ID]
	if !exists || len(signals) == 0 {
		log.Printf("[SIGNAL-DEBUG] Agent %s: No routing signals configured", current.ID)
		return nil
	}

	log.Printf("[SIGNAL-ROUTING] Agent %s: Attempting to match %d routing signals", current.ID, len(signals))

	// Check which signal is present in the response
	for _, sig := range signals {
		if sig.Target == "" {
			continue // Skip termination signals (handled by checkTerminationSignal)
		}

		// Check if signal matches response content
		log.Printf("[SIGNAL-TEST] Agent %s: Testing routing signal %s -> %s", current.ID, sig.Signal, sig.Target)
		if signalMatchesContent(sig.Signal, responseContent) {
			// Found matching signal, find the target agent
			nextAgent := ce.findAgentByID(sig.Target)
			if nextAgent != nil {
				log.Printf("[SIGNAL-SUCCESS] Agent %s routed to %s via signal %s", current.ID, nextAgent.ID, sig.Signal)
			} else {
				log.Printf("[SIGNAL-ERROR] Agent %s emitted signal %s targeting unknown agent %s", current.ID, sig.Signal, sig.Target)
			}
			return nextAgent
		}
	}

	log.Printf("[SIGNAL-NO-MATCH] Agent %s: No routing signals matched response content", current.ID)
	return nil
}

// getAgentBehavior retrieves behavior config for an agent
func (ce *CrewExecutor) getAgentBehavior(agentID string) *AgentBehavior {
	if ce.crew.Routing == nil || ce.crew.Routing.AgentBehaviors == nil {
		return nil
	}
	behavior, exists := ce.crew.Routing.AgentBehaviors[agentID]
	if !exists {
		return nil
	}
	return &behavior
}

// findNextAgent finds the next appropriate agent for handoff
func (ce *CrewExecutor) findNextAgent(current *Agent) *Agent {
	log.Printf("[HANDOFF] Agent %s: Finding next agent for handoff", current.ID)

	// First, try to use handoff_targets from current agent config
	if len(current.HandoffTargets) > 0 {
		log.Printf("[HANDOFF-TARGET] Agent %s: Has %d configured handoff targets", current.ID, len(current.HandoffTargets))
		// Create a map of agents by ID for quick lookup
		agentMap := make(map[string]*Agent)
		for _, agent := range ce.crew.Agents {
			agentMap[agent.ID] = agent
		}

		// Try to find the first available handoff target
		for _, targetID := range current.HandoffTargets {
			if agent, exists := agentMap[targetID]; exists && agent.ID != current.ID {
				log.Printf("[HANDOFF-SUCCESS] Agent %s handoff to %s (configured target)", current.ID, agent.ID)
				return agent
			}
		}
		log.Printf("[HANDOFF-NO-TARGET] Agent %s: No configured handoff targets available", current.ID)
	}

	// Fallback: Find any other agent (not terminal-only strategy)
	log.Printf("[HANDOFF-FALLBACK] Agent %s: Using fallback - routing to any available agent", current.ID)
	for _, agent := range ce.crew.Agents {
		if agent.ID != current.ID {
			log.Printf("[HANDOFF-FALLBACK-SUCCESS] Agent %s fallback handoff to %s", current.ID, agent.ID)
			return agent
		}
	}

	log.Printf("[HANDOFF-ERROR] No next agent found for %s - is this the only agent?", current.ID)
	return nil
}

// findParallelGroup finds a parallel group configuration for the given agent
// Returns the parallel group if the agent's signal matches a parallel group target
func (ce *CrewExecutor) findParallelGroup(agentID string, signalContent string) *ParallelGroupConfig {
	if ce.crew.Routing == nil || ce.crew.Routing.ParallelGroups == nil {
		log.Printf("[PARALLEL-DEBUG] Agent %s: No parallel groups configured", agentID)
		return nil
	}

	log.Printf("[PARALLEL-CHECK] Agent %s: Checking for parallel group signals", agentID)

	// Check if this agent emits a signal that targets a parallel group
	if signals, exists := ce.crew.Routing.Signals[agentID]; exists {
		log.Printf("[PARALLEL-SIGNALS] Agent %s: Has %d signals to check for parallel groups", agentID, len(signals))
		for _, signal := range signals {
			// Check if the agent's response contains the signal
			log.Printf("[PARALLEL-TEST] Agent %s: Testing signal %s for parallel group target", agentID, signal.Signal)
			if signalMatchesContent(signal.Signal, signalContent) {
				// Check if this signal targets a parallel group
				if parallelGroup, exists := ce.crew.Routing.ParallelGroups[signal.Target]; exists {
					log.Printf("[PARALLEL-FOUND] Agent %s triggers parallel group %s via signal %s", agentID, signal.Target, signal.Signal)
					return &parallelGroup
				}
				log.Printf("[PARALLEL-TARGET-NOT-FOUND] Agent %s signal %s targets unknown parallel group %s", agentID, signal.Signal, signal.Target)
			}
		}
	}

	log.Printf("[PARALLEL-NO-MATCH] Agent %s: No signals matched for parallel group execution", agentID)
	return nil
}
