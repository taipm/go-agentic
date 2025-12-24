package crewai

import (
	"fmt"
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

// ValidateSignals validates all signals defined in the routing configuration
// It checks:
// 1. Signal format matches [NAME] pattern
// 2. Target agent/group exists (or is empty for termination)
// 3. No duplicate signal definitions
// Returns an error with detailed message if validation fails
func (ce *CrewExecutor) ValidateSignals() error {
	// If no routing configured, skip validation
	if ce.crew == nil || ce.crew.Routing == nil || len(ce.crew.Routing.Signals) == 0 {
		return nil
	}

	// Build a map of valid agent IDs and parallel group names for quick lookup
	validTargets := make(map[string]bool)

	// Add agent IDs as valid targets
	validAgents := make(map[string]bool)
	for _, agent := range ce.crew.Agents {
		validAgents[agent.ID] = true
		validTargets[agent.ID] = true
	}

	// Add parallel group names as valid targets (Phase 3.6 enhancement)
	if ce.crew.Routing.ParallelGroups != nil {
		for groupName := range ce.crew.Routing.ParallelGroups {
			validTargets[groupName] = true
		}
	}

	// Track all signal definitions to detect duplicates
	seenSignals := make(map[string]string) // signal -> agent that defines it

	// Validate each signal in the routing configuration
	for agentID, signals := range ce.crew.Routing.Signals {
		for _, signal := range signals {
			// 1. Validate signal format: must match [NAME] pattern
			if signal.Signal == "" {
				return fmt.Errorf("agent '%s' has signal with empty name - signal must be in [NAME] format", agentID)
			}

			// Check if signal is in brackets format
			if !isValidSignalFormat(signal.Signal) {
				return fmt.Errorf("agent '%s' has invalid signal format '%s' - must be in [NAME] format (e.g., [END_EXAM])", agentID, signal.Signal)
			}

			// 2. Validate target: either empty (termination), valid agent ID, or parallel group name
			if signal.Target != "" {
				if !validTargets[signal.Target] {
					return fmt.Errorf("agent '%s' emits signal '%s' targeting unknown target '%s' - target must be empty (terminate), valid agent ID, or parallel group name", agentID, signal.Signal, signal.Target)
				}
			}

			// 3. Check for duplicate signal definitions from same agent
			if existing, exists := seenSignals[signal.Signal]; exists && existing == agentID {
				return fmt.Errorf("agent '%s' has duplicate signal definition for '%s'", agentID, signal.Signal)
			}

			// Track this signal definition
			seenSignals[signal.Signal] = agentID
		}
	}

	// Phase 3.6: Validate parallel group contents
	if ce.crew.Routing.ParallelGroups != nil {
		for groupName, group := range ce.crew.Routing.ParallelGroups {
			// Check that group has agents defined
			if group.Agents == nil || len(group.Agents) == 0 {
				return fmt.Errorf("parallel group '%s' has no agents defined", groupName)
			}

			// Validate that all agents in the group exist
			for _, agentID := range group.Agents {
				if !validAgents[agentID] {
					return fmt.Errorf("parallel group '%s' references unknown agent '%s'", groupName, agentID)
				}
			}
		}
	}

	parallelGroupCount := 0
	if ce.crew.Routing.ParallelGroups != nil {
		parallelGroupCount = len(ce.crew.Routing.ParallelGroups)
	}
	log.Printf("Signal validation passed: %d signals defined across %d agents, %d parallel groups", countTotalSignals(ce.crew.Routing.Signals), len(ce.crew.Agents), parallelGroupCount)

	// Phase 3.5: Enhanced registry validation (optional)
	if ce.signalRegistry != nil {
		log.Printf("[PHASE-3.5] Validating signals against signal registry...")
		validator := NewSignalValidator(ce.signalRegistry)

		// Validate configuration against registry
		validationErrors := validator.ValidateConfiguration(ce.crew.Routing.Signals, validTargets)
		if len(validationErrors) > 0 {
			// Return first error, but log all of them
			for _, err := range validationErrors {
				log.Printf("[SIGNAL-REGISTRY-ERROR] %v", err)
			}
			return fmt.Errorf("signal registry validation failed: %v", validationErrors[0])
		}

		log.Printf("[PHASE-3.5] Signal registry validation passed ✅")
	}

	return nil
}

// isValidSignalFormat checks if a signal matches the [NAME] format
func isValidSignalFormat(signal string) bool {
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

// countTotalSignals returns the total number of signals across all agents
func countTotalSignals(signals map[string][]RoutingSignal) int {
	count := 0
	for _, signalList := range signals {
		count += len(signalList)
	}
	return count
}
