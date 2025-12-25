package crewai

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
)

// ============================================================================
// MIGRATED FROM config_validator.go: Helper Functions & Public APIs
// ============================================================================

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

// ValidateRequiredFields performs generic field-level validation based on struct tags
// ✅ FIX for Issue #5 Phase 2: Generic required field validation using reflection
// Supports tags like: `required:"strict"` to enforce strict mode validation
// Returns a list of missing required fields and their descriptions
func ValidateRequiredFields(config interface{}, configMode ConfigMode, entityID string) ([]string, error) {
	var missingFields []string

	// Only check required fields in STRICT MODE
	if configMode != StrictMode {
		return missingFields, nil
	}

	// Use reflection to inspect struct fields and their tags
	t := reflect.TypeOf(config)
	v := reflect.ValueOf(config)

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return []string{fmt.Sprintf("config object is nil")}, nil
		}
		t = t.Elem()
		v = v.Elem()
	}

	// Iterate through all fields in the struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Check if field has required:"strict" tag
		requiredTag := field.Tag.Get("required")
		if requiredTag != "strict" {
			continue
		}

		// Check if field is empty/nil
		isEmpty := false
		switch fieldVal.Kind() {
		case reflect.String:
			isEmpty = fieldVal.String() == ""
		case reflect.Ptr, reflect.Interface:
			isEmpty = fieldVal.IsNil()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			isEmpty = fieldVal.Int() == 0
		case reflect.Float32, reflect.Float64:
			isEmpty = fieldVal.Float() == 0
		case reflect.Bool:
			isEmpty = !fieldVal.Bool()
		case reflect.Slice, reflect.Array:
			isEmpty = fieldVal.Len() == 0
		case reflect.Map:
			isEmpty = fieldVal.Len() == 0
		}

		// If field is required but empty, add to missing fields
		if isEmpty {
			yamlTag := field.Tag.Get("yaml")
			missingFields = append(missingFields, yamlTag)
		}
	}

	return missingFields, nil
}

// validateCrewRequiredFields checks for essential crew configuration fields
func validateCrewRequiredFields(config *CrewConfig) error {
	if config.Version == "" {
		return fmt.Errorf("required field 'version' is empty")
	}
	if len(config.Agents) == 0 {
		return fmt.Errorf("required field 'agents' is empty - at least one agent must be configured")
	}
	if config.EntryPoint == "" {
		return fmt.Errorf("required field 'entry_point' is empty")
	}
	return nil
}

// validateEntryPointAndBuildMap validates that entry_point exists and builds agent map
func validateEntryPointAndBuildMap(config *CrewConfig) (map[string]bool, error) {
	agentMap := make(map[string]bool)
	entryExists := false

	for _, agent := range config.Agents {
		agentMap[agent] = true
		if agent == config.EntryPoint {
			entryExists = true
		}
	}

	if !entryExists {
		return nil, fmt.Errorf("entry_point '%s' not found in agents list", config.EntryPoint)
	}

	return agentMap, nil
}

// validateCrewSettings validates settings constraints
func validateCrewSettings(config *CrewConfig) error {
	if config.Settings.MaxHandoffs < 0 {
		return fmt.Errorf("settings.max_handoffs must be >= 0, got %d", config.Settings.MaxHandoffs)
	}
	if config.Settings.MaxRounds <= 0 {
		return fmt.Errorf("settings.max_rounds must be > 0, got %d", config.Settings.MaxRounds)
	}
	if config.Settings.TimeoutSeconds <= 0 {
		return fmt.Errorf("settings.timeout_seconds must be > 0, got %d", config.Settings.TimeoutSeconds)
	}
	return nil
}

// validateSignals validates routing signals reference valid agents and formats
func validateSignals(signals map[string][]RoutingSignal, agentMap, parallelGroupMap map[string]bool) error {
	for agentID, signalList := range signals {
		if !agentMap[agentID] {
			return fmt.Errorf("routing.signals references non-existent agent '%s'", agentID)
		}

		for _, signal := range signalList {
			if signal.Signal == "" {
				return fmt.Errorf("agent '%s' has signal with empty name - must be in [NAME] format", agentID)
			}
			if !isSignalFormatValid(signal.Signal) {
				return fmt.Errorf("agent '%s' has invalid signal format '%s' - must be in [NAME] format (e.g., [END_EXAM])", agentID, signal.Signal)
			}

			if signal.Target != "" && !agentMap[signal.Target] && !parallelGroupMap[signal.Target] {
				return fmt.Errorf("routing signal from agent '%s' targets non-existent agent/group '%s'", agentID, signal.Target)
			}
		}
	}
	return nil
}

// validateAgentBehaviors validates agent behaviors reference existing agents
func validateAgentBehaviors(behaviors map[string]AgentBehavior, agentMap map[string]bool) error {
	for agentID := range behaviors {
		if !agentMap[agentID] {
			return fmt.Errorf("routing.agent_behaviors references non-existent agent '%s'", agentID)
		}
	}
	return nil
}

// validateParallelGroups validates parallel groups structure and references
func validateParallelGroups(groups map[string]ParallelGroupConfig, agentMap map[string]bool) error {
	for groupName, group := range groups {
		if len(group.Agents) == 0 {
			return fmt.Errorf("parallel_group '%s' has no agents", groupName)
		}

		for _, agentID := range group.Agents {
			if !agentMap[agentID] {
				return fmt.Errorf("parallel_group '%s' references non-existent agent '%s'", groupName, agentID)
			}
		}

		if group.NextAgent != "" && !agentMap[group.NextAgent] {
			return fmt.Errorf("parallel_group '%s' next_agent '%s' does not exist", groupName, group.NextAgent)
		}

		if group.TimeoutSeconds <= 0 {
			return fmt.Errorf("parallel_group '%s' timeout_seconds must be > 0, got %d", groupName, group.TimeoutSeconds)
		}
	}
	return nil
}

// validateRoutingReferences validates routing signals and parallel groups
func validateRoutingReferences(config *CrewConfig, agentMap map[string]bool) error {
	if config.Routing == nil {
		return nil
	}

	// Build parallel groups map
	parallelGroupMap := make(map[string]bool)
	for groupName := range config.Routing.ParallelGroups {
		parallelGroupMap[groupName] = true
	}

	// Validate signals
	if err := validateSignals(config.Routing.Signals, agentMap, parallelGroupMap); err != nil {
		return err
	}

	// Validate agent behaviors
	if err := validateAgentBehaviors(config.Routing.AgentBehaviors, agentMap); err != nil {
		return err
	}

	// Validate parallel groups
	if err := validateParallelGroups(config.Routing.ParallelGroups, agentMap); err != nil {
		return err
	}

	return nil
}

// ValidateCrewConfig validates crew configuration structure and constraints
// ✅ FIX for Issue #6: Validate YAML config at load time instead of runtime
// This prevents invalid configs from causing runtime crashes
func ValidateCrewConfig(config *CrewConfig) error {
	// Step 1: Validate required fields
	if err := validateCrewRequiredFields(config); err != nil {
		return err
	}

	// Step 2: Validate entry_point and build agent map
	agentMap, err := validateEntryPointAndBuildMap(config)
	if err != nil {
		return err
	}

	// Step 3: Validate settings constraints
	if err := validateCrewSettings(config); err != nil {
		return err
	}

	// Step 4: Validate routing references
	if err := validateRoutingReferences(config, agentMap); err != nil {
		return err
	}

	return nil
}

// validateAgentBasicConstraints validates temperature and quota constraints
func validateAgentBasicConstraints(config *AgentConfig) error {
	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("agent '%s': temperature must be between 0 and 2, got %f", config.ID, config.Temperature)
	}

	if config.MaxTokensPerCall < 0 {
		return fmt.Errorf("agent '%s': max_tokens_per_call must be >= 0, got %d", config.ID, config.MaxTokensPerCall)
	}
	if config.MaxTokensPerDay < 0 {
		return fmt.Errorf("agent '%s': max_tokens_per_day must be >= 0, got %d", config.ID, config.MaxTokensPerDay)
	}
	if config.MaxCostPerDay < 0 {
		return fmt.Errorf("agent '%s': max_cost_per_day must be >= 0, got %f", config.ID, config.MaxCostPerDay)
	}
	if config.CostAlertThreshold < 0 || config.CostAlertThreshold > 1 {
		return fmt.Errorf("agent '%s': cost_alert_threshold must be between 0 and 1, got %f", config.ID, config.CostAlertThreshold)
	}

	return nil
}

// validatePrimaryModelConfig validates primary model configuration
func validatePrimaryModelConfig(config *AgentConfig, configMode ConfigMode) error {
	if config.Primary == nil {
		return fmt.Errorf("agent '%s': primary model configuration is missing", config.ID)
	}

	// Check required fields in STRICT mode
	primaryMissingFields, _ := ValidateRequiredFields(config.Primary, configMode, config.ID)
	if len(primaryMissingFields) > 0 {
		return fmt.Errorf(
			"agent '%s': primary model configuration incomplete in STRICT MODE: missing %v\n"+
				"    Fix: Add all required fields to your agent config file, e.g.:\n"+
				"         primary:\n"+
				"           model: gpt-4o\n"+
				"           provider: openai",
			config.ID, primaryMissingFields)
	}

	// Apply defaults in PERMISSIVE mode
	if configMode != StrictMode {
		if config.Primary.Model == "" {
			config.Primary.Model = "gpt-4o"
			log.Printf("[CONFIG] agent '%s': primary.model not configured, using default 'gpt-4o' in PERMISSIVE mode", config.ID)
		}
		if config.Primary.Provider == "" {
			config.Primary.Provider = "openai"
			log.Printf("[CONFIG] agent '%s': primary.provider not configured, using default 'openai' in PERMISSIVE mode", config.ID)
		}
	}

	return nil
}

// validateBackupModelConfig validates backup model configuration if present
func validateBackupModelConfig(config *AgentConfig) error {
	if config.Backup == nil {
		return nil
	}

	if config.Backup.Model == "" {
		return fmt.Errorf("agent '%s': backup.model must not be empty if backup is specified", config.ID)
	}
	if config.Backup.Provider == "" {
		return fmt.Errorf("agent '%s': backup.provider must not be empty if backup is specified", config.ID)
	}

	log.Printf("[CONFIG INFO] agent '%s': backup model '%s' (%s) configured", config.ID, config.Backup.Model, config.Backup.Provider)
	return nil
}

// warnAgentContextMissing warns if agent has no system prompt or backstory
func warnAgentContextMissing(config *AgentConfig) {
	if config.SystemPrompt == "" && config.Backstory == "" {
		log.Printf("[CONFIG WARNING] agent '%s': both 'system_prompt' and 'backstory' are empty - agent may not have proper context", config.ID)
	}
}

// ValidateAgentConfig validates agent configuration structure and constraints
// ✅ FIX for Issue #6: Validate agent config at load time
// ✅ Support for primary/backup LLM model configuration
// ✅ FIX for Issue #5: Add configMode parameter for STRICT/PERMISSIVE mode validation
// ✅ FIX for Issue #5 Phase 2: Use tag-based validation for required fields
func ValidateAgentConfig(config *AgentConfig, configMode ConfigMode) error {
	// Step 1: Check required fields
	missingFields, _ := ValidateRequiredFields(config, configMode, config.ID)
	if len(missingFields) > 0 {
		return fmt.Errorf(
			"agent '%s': missing required fields in STRICT MODE: %v\n"+
				"    Explicit configuration is mandatory.\n"+
				"    Check your agent config file for: %v",
			config.ID, missingFields, missingFields)
	}

	// Step 2: Validate basic constraints
	if err := validateAgentBasicConstraints(config); err != nil {
		return err
	}

	// Step 3: Validate primary model config
	if err := validatePrimaryModelConfig(config, configMode); err != nil {
		return err
	}

	// Step 4: Validate backup model config if present
	if err := validateBackupModelConfig(config); err != nil {
		return err
	}

	// Step 5: Warn about missing context
	warnAgentContextMissing(config)

	return nil
}

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
