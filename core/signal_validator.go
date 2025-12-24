package crewai

import (
	"fmt"
	"log"
	"strings"
)

// SignalValidator provides comprehensive signal validation functionality
type SignalValidator struct {
	registry *SignalRegistry
}

// NewSignalValidator creates a new signal validator
func NewSignalValidator(registry *SignalRegistry) *SignalValidator {
	return &SignalValidator{
		registry: registry,
	}
}

// ValidateSignalEmission validates that a signal can be emitted by an agent
func (sv *SignalValidator) ValidateSignalEmission(signal string, agentID string) error {
	// Check format
	if !isSignalFormatValid(signal) {
		return fmt.Errorf("signal '%s' has invalid format - must match [NAME]", signal)
	}

	// Check if signal is in registry
	definition := sv.registry.Get(signal)
	if definition == nil {
		return fmt.Errorf("signal '%s' is not registered (unknown signal)", signal)
	}

	// Check if agent is allowed to emit
	if !definition.AllowAllAgents && len(definition.AllowedAgents) > 0 {
		allowed := false
		for _, a := range definition.AllowedAgents {
			if a == agentID {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("agent '%s' is not allowed to emit signal '%s'", agentID, signal)
		}
	}

	// Check deprecation
	if definition.DeprecatedMsg != "" {
		log.Printf("[SIGNAL-WARNING] Agent %s using deprecated signal %s: %s", agentID, signal, definition.DeprecatedMsg)
	}

	return nil
}

// ValidateSignalTarget validates that a signal can route to a target
func (sv *SignalValidator) ValidateSignalTarget(signal string, agentID string, targetAgent string, availableAgents map[string]bool) error {
	definition := sv.registry.Get(signal)
	if definition == nil {
		return fmt.Errorf("signal '%s' not registered", signal)
	}

	// Validate based on behavior
	switch definition.Behavior {
	case SignalBehaviorTerminate:
		// Termination signals must have empty target
		if targetAgent != "" {
			return fmt.Errorf("termination signal '%s' must have empty target, got '%s'", signal, targetAgent)
		}

	case SignalBehaviorRoute:
		// Routing signals must have valid target
		if targetAgent == "" {
			return fmt.Errorf("routing signal '%s' must have a target agent", signal)
		}

		// Check if target exists
		if !availableAgents[targetAgent] {
			return fmt.Errorf("signal '%s' targets unknown agent '%s'", signal, targetAgent)
		}

		// Check if target is in valid targets list (if specified)
		if len(definition.ValidTargets) > 0 {
			found := false
			for _, valid := range definition.ValidTargets {
				if valid == targetAgent {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("signal '%s' cannot target '%s' (valid: %v)", signal, targetAgent, definition.ValidTargets)
			}
		}

	case SignalBehaviorPause:
		// Pause signals should have empty target
		if targetAgent != "" {
			log.Printf("[SIGNAL-WARNING] Pause signal '%s' should have empty target, got '%s'", signal, targetAgent)
		}

	default:
		log.Printf("[SIGNAL-WARNING] Unknown signal behavior: %s", definition.Behavior)
	}

	return nil
}

// ValidateConfiguration validates a complete signal configuration
func (sv *SignalValidator) ValidateConfiguration(signals map[string][]RoutingSignal, agents map[string]bool) []error {
	var errors []error

	for agentID, signalList := range signals {
		if !agents[agentID] {
			errors = append(errors, fmt.Errorf("signals defined for unknown agent '%s'", agentID))
			continue
		}

		for _, sig := range signalList {
			// Validate emission
			if err := sv.ValidateSignalEmission(sig.Signal, agentID); err != nil {
				errors = append(errors, fmt.Errorf("agent '%s': %w", agentID, err))
				continue
			}

			// Validate target
			if err := sv.ValidateSignalTarget(sig.Signal, agentID, sig.Target, agents); err != nil {
				errors = append(errors, fmt.Errorf("agent '%s': %w", agentID, err))
			}
		}
	}

	return errors
}

// ValidateSignalInContent validates that a signal string appears correctly formatted in content
func (sv *SignalValidator) ValidateSignalInContent(signal string, content string) (bool, string) {
	// 1. Exact match
	if strings.Contains(content, signal) {
		return true, "exact"
	}

	// 2. Case-insensitive match
	if strings.Contains(strings.ToLower(content), strings.ToLower(signal)) {
		return true, "case_insensitive"
	}

	// 3. Normalized match (for Vietnamese with diacritics, etc)
	normalizedSignal := normalizeSignalText(signal)
	if len(signal) > 2 && signal[0] == '[' && signal[len(signal)-1] == ']' {
		// Check if normalized version exists in content
		for i := 0; i < len(content); i++ {
			if content[i] == '[' {
				for j := i + 1; j < len(content) && j < i+100; j++ {
					if content[j] == ']' {
						bracketContent := content[i : j+1]
						normalizedContent := normalizeSignalText(bracketContent)
						if normalizedContent == normalizedSignal {
							return true, "normalized"
						}
						break
					}
				}
			}
		}
	}

	return false, ""
}

// GetSignalInfo returns detailed information about a signal
func (sv *SignalValidator) GetSignalInfo(signal string) *SignalDefinition {
	return sv.registry.Get(signal)
}

// LogSignalEvent logs a signal event with full context
func (sv *SignalValidator) LogSignalEvent(event *SignalEvent) {
	definition := sv.registry.Get(event.Signal)
	if definition == nil {
		log.Printf("[SIGNAL-EVENT] Unknown signal '%s' from agent '%s' (not in registry)", event.Signal, event.AgentID)
		return
	}

	log.Printf("[SIGNAL-EVENT] Agent '%s' emitted '%s' (%s) - %s",
		event.AgentID, event.Signal, definition.Description, event.MatchMethod)
}

// GenerateSignalReport creates a comprehensive report of signal usage
func (sv *SignalValidator) GenerateSignalReport() string {
	var report strings.Builder

	report.WriteString("=== SIGNAL REGISTRY REPORT ===\n\n")

	allSignals := sv.registry.GetAll()
	report.WriteString(fmt.Sprintf("Total Signals Registered: %d\n\n", len(allSignals)))

	// Group by behavior
	behaviors := map[SignalBehavior][]*SignalDefinition{}
	for _, def := range allSignals {
		behaviors[def.Behavior] = append(behaviors[def.Behavior], def)
	}

	behaviorList := []SignalBehavior{SignalBehaviorTerminate, SignalBehaviorRoute, SignalBehaviorPause, SignalBehaviorParallel, SignalBehaviorBroadcast}
	for _, behavior := range behaviorList {
		if signals, ok := behaviors[behavior]; ok && len(signals) > 0 {
			report.WriteString(fmt.Sprintf("--- %s Signals (%d) ---\n", behavior, len(signals)))
			for _, sig := range signals {
				report.WriteString(fmt.Sprintf("  %s: %s\n", sig.Name, sig.Description))
				if sig.DeprecatedMsg != "" {
					report.WriteString(fmt.Sprintf("    [DEPRECATED] %s\n", sig.DeprecatedMsg))
				}
			}
			report.WriteString("\n")
		}
	}

	return report.String()
}
