package crewai

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// ValidationError represents a single validation error with context
type ValidationError struct {
	File     string // Which file (crew.yaml, agent.yaml, etc.)
	Section  string // Which section (agent, routing, etc.)
	Field    string // Which field (entry_point, temperature, etc.)
	Message  string // What's wrong
	Severity string // "error" or "warning"
	Fix      string // How to fix it
	Line     int    // Line number in file (if available)
}

// ConfigValidator orchestrates all validation checks
type ConfigValidator struct {
	crewConfig *CrewConfig
	agents     map[string]*AgentConfig
	errors     []ValidationError
	warnings   []ValidationError
	visited    map[string]bool    // For cycle detection
	mu         sync.RWMutex
}

// NewConfigValidator creates a new validator
func NewConfigValidator(crewConfig *CrewConfig, agents map[string]*AgentConfig) *ConfigValidator {
	return &ConfigValidator{
		crewConfig: crewConfig,
		agents:     agents,
		errors:     []ValidationError{},
		warnings:   []ValidationError{},
		visited:    make(map[string]bool),
	}
}

// ValidateAll runs all validation checks
func (cv *ConfigValidator) ValidateAll() error {
	if cv.crewConfig == nil {
		return fmt.Errorf("crew config is nil")
	}

	// Stage 1: Basic structure validation
	cv.validateCrewStructure()

	// Stage 2: Field validation
	cv.validateFields()

	// Stage 3: Agent validation
	cv.validateAgents()

	// Stage 4: Routing validation
	if cv.crewConfig.Routing != nil {
		cv.validateRouting()
	}

	// Stage 5: Graph validation
	cv.validateGraph()

	// Return error if there are any errors
	if len(cv.errors) > 0 {
		return cv.GenerateErrorReport()
	}

	return nil
}

// validateCrewStructure checks basic crew structure
func (cv *ConfigValidator) validateCrewStructure() {
	if cv.crewConfig.EntryPoint == "" {
		cv.addError("crew.yaml", "entry_point", "entry_point is required", "entry_point: orchestrator")
		return
	}

	if len(cv.crewConfig.Agents) == 0 {
		cv.addError("crew.yaml", "agents", "agents list cannot be empty", "agents:\n  - orchestrator\n  - executor")
		return
	}

	// Check entry_point exists in agents list
	found := false
	for _, agent := range cv.crewConfig.Agents {
		if agent == cv.crewConfig.EntryPoint {
			found = true
			break
		}
	}
	if !found {
		available := strings.Join(cv.crewConfig.Agents, ", ")
		cv.addError("crew.yaml", "entry_point",
			fmt.Sprintf("entry_point '%s' not found in agents list", cv.crewConfig.EntryPoint),
			fmt.Sprintf("entry_point must be one of: %s", available))
	}
}

// validateFields checks field values and ranges
func (cv *ConfigValidator) validateFields() {
	// Validate max_handoffs
	if cv.crewConfig.Settings.MaxHandoffs < 1 {
		cv.addError("crew.yaml", "settings.max_handoffs",
			"max_handoffs must be >= 1",
			"max_handoffs: 5")
	}

	// Validate max_rounds
	if cv.crewConfig.Settings.MaxRounds < 1 {
		cv.addError("crew.yaml", "settings.max_rounds",
			"max_rounds must be >= 1",
			"max_rounds: 10")
	}

	// Validate timeout
	if cv.crewConfig.Settings.TimeoutSeconds < 1 {
		cv.addError("crew.yaml", "settings.timeout_seconds",
			"timeout_seconds must be >= 1",
			"timeout_seconds: 300")
	}

	// Validate agent configurations
	for _, agentConfig := range cv.agents {
		cv.validateAgentConfig(agentConfig)
	}
}

// validateAgentConfig checks individual agent configuration
func (cv *ConfigValidator) validateAgentConfig(agent *AgentConfig) {
	if agent.ID == "" {
		cv.addError(fmt.Sprintf("agents/%s.yaml", agent.Name), "id",
			"Agent ID is required",
			"id: orchestrator")
		return
	}

	if agent.Model == "" {
		cv.addError(fmt.Sprintf("agents/%s.yaml", agent.ID), "model",
			"Model is required",
			"model: gpt-4o")
		return
	}

	// Validate model name
	validModels := map[string]bool{
		"gpt-4o":       true,
		"gpt-4-turbo":  true,
		"gpt-4":        true,
		"gpt-3.5-turbo": true,
	}
	if !validModels[agent.Model] {
		available := strings.Join([]string{"gpt-4o", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo"}, ", ")
		cv.addWarning(fmt.Sprintf("agents/%s.yaml", agent.ID), "model",
			fmt.Sprintf("Model '%s' may not be valid", agent.Model),
			fmt.Sprintf("Use one of: %s", available))
	}

	// Validate temperature
	if agent.Temperature < 0 || agent.Temperature > 1 {
		cv.addError(fmt.Sprintf("agents/%s.yaml", agent.ID), "temperature",
			fmt.Sprintf("Temperature must be 0-1, got %.2f", agent.Temperature),
			"temperature: 0.7")
	}

	// Check for duplicate agent IDs
	count := 0
	for _, a := range cv.agents {
		if a.ID == agent.ID {
			count++
		}
	}
	if count > 1 {
		cv.addError(fmt.Sprintf("agents/%s.yaml", agent.ID), "id",
			fmt.Sprintf("Duplicate agent ID: '%s'", agent.ID),
			fmt.Sprintf("Ensure all agent IDs are unique"))
	}
}

// validateAgents checks agent-specific validations
func (cv *ConfigValidator) validateAgents() {
	// Collect all agent IDs for lookup
	agentIDs := make(map[string]bool)
	for _, agent := range cv.agents {
		agentIDs[agent.ID] = true
	}

	// Validate each agent
	for _, agentID := range cv.crewConfig.Agents {
		if _, exists := cv.agents[agentID]; !exists {
			cv.addError("crew.yaml", "agents",
				fmt.Sprintf("Agent '%s' not found in configuration", agentID),
				fmt.Sprintf("Load configuration for agent '%s'", agentID))
		}
	}
}

// validateRouting checks routing configuration
func (cv *ConfigValidator) validateRouting() {
	if cv.crewConfig.Routing == nil {
		return
	}

	// Collect all agent IDs
	agentIDs := make(map[string]bool)
	for _, agent := range cv.agents {
		agentIDs[agent.ID] = true
	}

	// Validate signals
	for agentID, signals := range cv.crewConfig.Routing.Signals {
		// Verify agent exists
		if _, exists := agentIDs[agentID]; !exists && agentID != "" {
			cv.addWarning("crew.yaml", "routing.signals",
				fmt.Sprintf("Signal defined for non-existent agent '%s'", agentID),
				fmt.Sprintf("Remove signal or add agent '%s'", agentID))
			continue
		}

		// Validate each signal's target
		for _, signal := range signals {
			if signal.Target == "" || signal.Target == "null" {
				// Terminal signal (OK)
				continue
			}

			if _, exists := agentIDs[signal.Target]; !exists {
				cv.addError("crew.yaml", "routing.signals",
					fmt.Sprintf("Signal '%s' routes to non-existent agent '%s'", signal.Signal, signal.Target),
					fmt.Sprintf("Target must be one of: %v", agentIDs))
			}
		}
	}

	// Check for circular references
	cv.DetectCircularReferences()
}

// DetectCircularReferences detects circular routing patterns
func (cv *ConfigValidator) DetectCircularReferences() error {
	if cv.crewConfig.Routing == nil || cv.crewConfig.Routing.Signals == nil {
		return nil
	}

	agentIDs := make(map[string]bool)
	for _, agent := range cv.agents {
		agentIDs[agent.ID] = true
	}

	// For each agent, check for cycles
	for agentID := range agentIDs {
		cv.visited = make(map[string]bool) // Reset visited for each starting point
		if cv.hasCycle(agentID) {
			cv.addError("crew.yaml", "routing",
				fmt.Sprintf("Circular routing detected starting from agent '%s'", agentID),
				"Modify routing signals to break the cycle")
		}
	}

	return nil
}

// hasCycle detects if there's a cycle starting from agentID
func (cv *ConfigValidator) hasCycle(agentID string) bool {
	if cv.visited[agentID] {
		return true // Cycle detected
	}

	cv.visited[agentID] = true

	// Get all targets this agent can route to
	signals, ok := cv.crewConfig.Routing.Signals[agentID]
	if !ok {
		return false // No signals = no cycle
	}

	for _, signal := range signals {
		if signal.Target == "" || signal.Target == "null" {
			continue // Terminal signal
		}

		if cv.hasCycle(signal.Target) {
			return true
		}
	}

	cv.visited[agentID] = false // Backtrack
	return false
}

// CheckReachability verifies all agents can be reached from entry point
func (cv *ConfigValidator) CheckReachability() error {
	if cv.crewConfig.Routing == nil || len(cv.crewConfig.Agents) == 0 {
		return nil
	}

	// BFS from entry point
	reachable := make(map[string]bool)
	queue := []string{cv.crewConfig.EntryPoint}
	reachable[cv.crewConfig.EntryPoint] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		signals, ok := cv.crewConfig.Routing.Signals[current]
		if !ok {
			continue
		}

		for _, signal := range signals {
			if signal.Target == "" || signal.Target == "null" {
				continue
			}

			if !reachable[signal.Target] {
				reachable[signal.Target] = true
				queue = append(queue, signal.Target)
			}
		}
	}

	// Check if all agents are reachable
	for _, agentID := range cv.crewConfig.Agents {
		if !reachable[agentID] {
			cv.addWarning("crew.yaml", "routing",
				fmt.Sprintf("Agent '%s' is not reachable from entry point '%s'", agentID, cv.crewConfig.EntryPoint),
				fmt.Sprintf("Add routing signals to make '%s' reachable", agentID))
		}
	}

	return nil
}

// validateGraph performs graph validation
func (cv *ConfigValidator) validateGraph() {
	// Check reachability
	cv.CheckReachability()
}

// addError adds a validation error
func (cv *ConfigValidator) addError(file string, field string, message string, fix string) {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	cv.errors = append(cv.errors, ValidationError{
		File:     file,
		Field:    field,
		Message:  message,
		Severity: "error",
		Fix:      fix,
	})
}

// addWarning adds a validation warning
func (cv *ConfigValidator) addWarning(file string, field string, message string, fix string) {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	cv.warnings = append(cv.warnings, ValidationError{
		File:     file,
		Field:    field,
		Message:  message,
		Severity: "warning",
		Fix:      fix,
	})
}

// GenerateErrorReport generates a human-readable error report
func (cv *ConfigValidator) GenerateErrorReport() error {
	if len(cv.errors) == 0 {
		return nil
	}

	var report strings.Builder
	report.WriteString("\n❌ Configuration Validation Failed:\n\n")

	for i, err := range cv.errors {
		report.WriteString(fmt.Sprintf("  Error %d:\n", i+1))
		report.WriteString(fmt.Sprintf("    File: %s\n", err.File))
		report.WriteString(fmt.Sprintf("    Field: %s\n", err.Field))
		report.WriteString(fmt.Sprintf("    Problem: %s\n", err.Message))
		report.WriteString(fmt.Sprintf("    Solution: %s\n", err.Fix))
		report.WriteString("\n")
	}

	if len(cv.warnings) > 0 {
		report.WriteString(fmt.Sprintf("⚠️  %d warning(s):\n\n", len(cv.warnings)))
		for _, warn := range cv.warnings {
			report.WriteString(fmt.Sprintf("  ⚠️  %s (%s): %s\n", warn.File, warn.Field, warn.Message))
		}
		report.WriteString("\n")
	}

	return fmt.Errorf("%s", report.String())
}

// PrintReport prints a detailed validation report
func (cv *ConfigValidator) PrintReport() {
	if len(cv.errors) == 0 && len(cv.warnings) == 0 {
		fmt.Println("✓ Configuration is valid")
		return
	}

	if len(cv.errors) > 0 {
		fmt.Printf("❌ %d error(s) found:\n\n", len(cv.errors))
		for i, err := range cv.errors {
			fmt.Printf("  %d. %s (%s)\n", i+1, err.Message, err.File)
			fmt.Printf("     Fix: %s\n\n", err.Fix)
		}
	}

	if len(cv.warnings) > 0 {
		fmt.Printf("⚠️  %d warning(s):\n\n", len(cv.warnings))
		for i, warn := range cv.warnings {
			fmt.Printf("  %d. %s (%s)\n", i+1, warn.Message, warn.File)
			fmt.Printf("     Fix: %s\n\n", warn.Fix)
		}
	}
}

// GetErrors returns all errors
func (cv *ConfigValidator) GetErrors() []ValidationError {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	return cv.errors
}

// GetWarnings returns all warnings
func (cv *ConfigValidator) GetWarnings() []ValidationError {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	return cv.warnings
}

// IsValid returns true if no errors (warnings are OK)
func (cv *ConfigValidator) IsValid() bool {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	return len(cv.errors) == 0
}

// ✅ PHASE 2: ValidationErrorFormatter - JSON output support
// ErrorDetail represents a single validation error in JSON format
type ErrorDetail struct {
	File     string `json:"file"`
	Field    string `json:"field"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Fix      string `json:"fix"`
	Line     int    `json:"line,omitempty"`
}

// ErrorSummary provides count and validity status
type ErrorSummary struct {
	TotalErrors   int  `json:"total_errors"`
	TotalWarnings int  `json:"total_warnings"`
	IsValid       bool `json:"is_valid"`
}

// ErrorResponse is the complete JSON structure for validation results
type ErrorResponse struct {
	Success  bool           `json:"success"`
	Errors   []ErrorDetail  `json:"errors"`
	Warnings []ErrorDetail  `json:"warnings"`
	Summary  ErrorSummary   `json:"summary"`
}

// ToJSON converts validation results to JSON format
// Returns pretty-printed JSON that can be:
// - Consumed by APIs and clients (parse JSON directly)
// - Sent to logging systems (structured logging)
// - Inspected by developers (readable format)
//
// Example output:
// {
//   "success": false,
//   "errors": [
//     {
//       "file": "crew.yaml",
//       "field": "entry_point",
//       "message": "entry_point is required",
//       "severity": "error",
//       "fix": "entry_point: orchestrator"
//     }
//   ],
//   "warnings": [],
//   "summary": {
//     "total_errors": 1,
//     "total_warnings": 0,
//     "is_valid": false
//   }
// }
func (cv *ConfigValidator) ToJSON() ([]byte, error) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()

	// Convert validation errors to JSON format
	errors := make([]ErrorDetail, len(cv.errors))
	for i, err := range cv.errors {
		errors[i] = ErrorDetail{
			File:     err.File,
			Field:    err.Field,
			Message:  err.Message,
			Severity: err.Severity,
			Fix:      err.Fix,
			Line:     err.Line,
		}
	}

	// Convert validation warnings to JSON format
	warnings := make([]ErrorDetail, len(cv.warnings))
	for i, warn := range cv.warnings {
		warnings[i] = ErrorDetail{
			File:     warn.File,
			Field:    warn.Field,
			Message:  warn.Message,
			Severity: warn.Severity,
			Fix:      warn.Fix,
			Line:     warn.Line,
		}
	}

	// Build response with summary
	resp := ErrorResponse{
		Success:  len(cv.errors) == 0,
		Errors:   errors,
		Warnings: warnings,
		Summary: ErrorSummary{
			TotalErrors:   len(cv.errors),
			TotalWarnings: len(cv.warnings),
			IsValid:       len(cv.errors) == 0,
		},
	}

	// Return pretty-printed JSON
	return json.MarshalIndent(resp, "", "  ")
}
