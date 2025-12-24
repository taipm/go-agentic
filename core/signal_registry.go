package crewai

import (
	"fmt"
	"log"
	"sync"
)

// SignalRegistry manages all signal definitions in the system
type SignalRegistry struct {
	definitions map[string]*SignalDefinition // signal name -> definition
	mu          sync.RWMutex                  // Protect concurrent access
}

// NewSignalRegistry creates a new signal registry
func NewSignalRegistry() *SignalRegistry {
	return &SignalRegistry{
		definitions: make(map[string]*SignalDefinition),
	}
}

// Register adds a signal definition to the registry
func (sr *SignalRegistry) Register(definition *SignalDefinition) error {
	if definition == nil {
		return fmt.Errorf("cannot register nil signal definition")
	}
	if definition.Name == "" {
		return fmt.Errorf("signal definition must have a name")
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	if _, exists := sr.definitions[definition.Name]; exists {
		return fmt.Errorf("signal '%s' already registered", definition.Name)
	}

	sr.definitions[definition.Name] = definition
	log.Printf("[SIGNAL-REGISTRY] Registered signal: %s (%s)", definition.Name, definition.Description)
	return nil
}

// RegisterBulk registers multiple signal definitions at once
func (sr *SignalRegistry) RegisterBulk(definitions []*SignalDefinition) error {
	for _, def := range definitions {
		if err := sr.Register(def); err != nil {
			return err
		}
	}
	return nil
}

// Get retrieves a signal definition by name
func (sr *SignalRegistry) Get(signalName string) *SignalDefinition {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.definitions[signalName]
}

// Exists checks if a signal is registered
func (sr *SignalRegistry) Exists(signalName string) bool {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	_, exists := sr.definitions[signalName]
	return exists
}

// GetAll returns all registered signal definitions
func (sr *SignalRegistry) GetAll() map[string]*SignalDefinition {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*SignalDefinition)
	for k, v := range sr.definitions {
		result[k] = v
	}
	return result
}

// Validate checks if a signal emission is valid according to the registry
func (sr *SignalRegistry) Validate(signalName string, agentID string, targetAgent string) error {
	definition := sr.Get(signalName)
	if definition == nil {
		return fmt.Errorf("signal '%s' not found in registry", signalName)
	}

	// Check if agent is allowed to emit this signal
	if !definition.AllowAllAgents && len(definition.AllowedAgents) > 0 {
		allowed := false
		for _, allowedAgent := range definition.AllowedAgents {
			if allowedAgent == agentID {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("agent '%s' is not allowed to emit signal '%s'", agentID, signalName)
		}
	}

	// Check if target is valid
	if definition.Behavior == SignalBehaviorRoute && targetAgent != "" {
		isValidTarget := false
		for _, valid := range definition.ValidTargets {
			if valid == targetAgent {
				isValidTarget = true
				break
			}
		}
		if !isValidTarget && len(definition.ValidTargets) > 0 {
			return fmt.Errorf("agent '%s' cannot be targeted by signal '%s' (valid: %v)", targetAgent, signalName, definition.ValidTargets)
		}
	}

	return nil
}

// Count returns the number of registered signals
func (sr *SignalRegistry) Count() int {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return len(sr.definitions)
}

// List returns a list of all registered signal names
func (sr *SignalRegistry) List() []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	names := make([]string, 0, len(sr.definitions))
	for name := range sr.definitions {
		names = append(names, name)
	}
	return names
}

// LoadDefaultSignals loads the default set of signals for the system
// These are standard signals used in common workflows
func LoadDefaultSignals() *SignalRegistry {
	registry := NewSignalRegistry()

	defaultSignals := []*SignalDefinition{
		// Termination signals
		{
			Name:          "[END]",
			Description:   "Generic workflow termination signal",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorTerminate,
			Example:       "Task complete. [END]",
			Priority:      100,
		},
		{
			Name:          "[END_EXAM]",
			Description:   "Terminates exam workflow",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorTerminate,
			Example:       "Exam finished. Score: 100%. [END_EXAM]",
			Priority:      100,
		},
		{
			Name:          "[DONE]",
			Description:   "Task completion signal",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorTerminate,
			Example:       "Work is done. [DONE]",
			Priority:      100,
		},
		{
			Name:          "[STOP]",
			Description:   "Immediate stop signal",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorTerminate,
			Example:       "Critical error encountered. [STOP]",
			Priority:      110, // Higher priority than [END]
		},

		// Routing signals
		{
			Name:          "[NEXT]",
			Description:   "Route to next agent",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorRoute,
			Example:       "Ready for next step. [NEXT]",
			Priority:      50,
		},
		{
			Name:          "[QUESTION]",
			Description:   "Route to question handler",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorRoute,
			Example:       "Question ready. [QUESTION]",
			Priority:      50,
		},
		{
			Name:          "[ANSWER]",
			Description:   "Route to answer handler",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorRoute,
			Example:       "Answer provided. [ANSWER]",
			Priority:      50,
		},

		// Status signals
		{
			Name:          "[OK]",
			Description:   "Acknowledgment signal",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorRoute,
			Example:       "OK, proceeding. [OK]",
			Priority:      30,
		},
		{
			Name:          "[ERROR]",
			Description:   "Error occurred signal",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorRoute,
			Example:       "Error: insufficient data. [ERROR]",
			Priority:      90, // Higher priority
		},
		{
			Name:          "[RETRY]",
			Description:   "Retry operation signal",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorRoute,
			Example:       "Retrying operation. [RETRY]",
			Priority:      70,
		},

		// Pause signal
		{
			Name:          "[WAIT]",
			Description:   "Pause and wait for input signal",
			AllowAllAgents: true,
			Behavior:      SignalBehaviorPause,
			Example:       "Waiting for user input. [WAIT]",
			Priority:      80,
		},
	}

	if err := registry.RegisterBulk(defaultSignals); err != nil {
		log.Printf("[SIGNAL-REGISTRY] Error loading default signals: %v", err)
	}

	log.Printf("[SIGNAL-REGISTRY] Loaded %d default signals", len(defaultSignals))
	return registry
}

// ValidateAgainstRegistry validates a signal configuration against the registry
func (sr *SignalRegistry) ValidateAgainstRegistry(signals map[string][]RoutingSignal) []SignalValidationError {
	var errors []SignalValidationError

	for agentID, signalList := range signals {
		for _, signal := range signalList {
			// Check if signal is registered
			if !sr.Exists(signal.Signal) {
				errors = append(errors, SignalValidationError{
					Signal:   signal.Signal,
					AgentID:  agentID,
					ErrorMsg: fmt.Sprintf("signal '%s' is not registered in the signal registry", signal.Signal),
					Severity: "warning",
				})
			} else {
				// Validate against definition
				definition := sr.Get(signal.Signal)
				if err := sr.Validate(signal.Signal, agentID, signal.Target); err != nil {
					errors = append(errors, SignalValidationError{
						Signal:   signal.Signal,
						AgentID:  agentID,
						ErrorMsg: err.Error(),
						Severity: "error",
					})
				}

				// Warn if deprecated
				if definition.DeprecatedMsg != "" {
					errors = append(errors, SignalValidationError{
						Signal:   signal.Signal,
						AgentID:  agentID,
						ErrorMsg: fmt.Sprintf("DEPRECATED: %s", definition.DeprecatedMsg),
						Severity: "warning",
					})
				}
			}
		}
	}

	return errors
}

// GetSignalsByBehavior returns all signals of a specific behavior type
func (sr *SignalRegistry) GetSignalsByBehavior(behavior SignalBehavior) []*SignalDefinition {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var result []*SignalDefinition
	for _, def := range sr.definitions {
		if def.Behavior == behavior {
			result = append(result, def)
		}
	}
	return result
}

// GetTerminationSignals returns all termination signals
func (sr *SignalRegistry) GetTerminationSignals() []*SignalDefinition {
	return sr.GetSignalsByBehavior(SignalBehaviorTerminate)
}

// GetRoutingSignals returns all routing signals
func (sr *SignalRegistry) GetRoutingSignals() []*SignalDefinition {
	return sr.GetSignalsByBehavior(SignalBehaviorRoute)
}
