package crewai

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RoutingSignal defines a signal that can be emitted by an agent
type RoutingSignal struct {
	Signal      string `yaml:"signal"`
	Target      string `yaml:"target"`
	Description string `yaml:"description"`
}

// AgentBehavior defines how an agent behaves in routing
type AgentBehavior struct {
	WaitForSignal bool   `yaml:"wait_for_signal"`
	AutoRoute     bool   `yaml:"auto_route"`
	IsTerminal    bool   `yaml:"is_terminal"`
	Description   string `yaml:"description"`
}

// ParallelGroupConfig defines a group of agents that should be executed in parallel
type ParallelGroupConfig struct {
	Agents         []string `yaml:"agents"`
	WaitForAll     bool     `yaml:"wait_for_all"`
	TimeoutSeconds int      `yaml:"timeout_seconds"`
	NextAgent      string   `yaml:"next_agent"`
	Description    string   `yaml:"description"`
}

// RoutingConfig defines routing rules for the crew
type RoutingConfig struct {
	Signals        map[string][]RoutingSignal     `yaml:"signals"`
	Defaults       map[string]string              `yaml:"defaults"`
	AgentBehaviors map[string]AgentBehavior       `yaml:"agent_behaviors"`
	ParallelGroups map[string]ParallelGroupConfig `yaml:"parallel_groups"`
}

// CrewConfig represents the crew configuration
type CrewConfig struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	EntryPoint  string `yaml:"entry_point"`

	Agents []string `yaml:"agents"`

	Settings struct {
		MaxHandoffs    int    `yaml:"max_handoffs"`
		MaxRounds      int    `yaml:"max_rounds"`
		TimeoutSeconds int    `yaml:"timeout_seconds"`
		Language       string `yaml:"language"`
		Organization   string `yaml:"organization"`
	} `yaml:"settings"`

	Routing *RoutingConfig `yaml:"routing"`
}

// ModelConfigYAML represents YAML configuration for a model (for parsing)
type ModelConfigYAML struct {
	Model       string `yaml:"model"`
	Provider    string `yaml:"provider"`
	ProviderURL string `yaml:"provider_url"`
}

// AgentConfig represents an agent configuration
type AgentConfig struct {
	ID             string           `yaml:"id"`
	Name           string           `yaml:"name"`
	Description    string           `yaml:"description"`
	Role           string           `yaml:"role"`
	Backstory      string           `yaml:"backstory"`
	Model          string           `yaml:"model"`         // Deprecated: Use Primary instead
	Temperature    float64          `yaml:"temperature"`
	IsTerminal     bool             `yaml:"is_terminal"`
	Tools          []string         `yaml:"tools"`
	HandoffTargets []string         `yaml:"handoff_targets"`
	SystemPrompt   string           `yaml:"system_prompt"`
	Provider       string           `yaml:"provider"`      // Deprecated: Use Primary.Provider instead
	ProviderURL    string           `yaml:"provider_url"`  // Deprecated: Use Primary.ProviderURL instead
	Primary        *ModelConfigYAML `yaml:"primary"`       // Primary LLM model configuration
	Backup         *ModelConfigYAML `yaml:"backup"`        // Backup LLM model configuration
}

// LoadCrewConfig loads the crew configuration from a YAML file
func LoadCrewConfig(path string) (*CrewConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read crew config: %w", err)
	}

	var config CrewConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse crew config YAML: %w", err)
	}

	// Set defaults
	if config.Settings.MaxHandoffs == 0 {
		config.Settings.MaxHandoffs = 5
	}
	if config.Settings.MaxRounds == 0 {
		config.Settings.MaxRounds = 10
	}
	if config.Settings.TimeoutSeconds == 0 {
		config.Settings.TimeoutSeconds = 300
	}
	if config.Settings.Language == "" {
		config.Settings.Language = "en"
	}

	// ✅ FIX for Issue #6: Validate configuration at load time
	// This catches invalid configs immediately with clear error messages
	if err := ValidateCrewConfig(&config); err != nil {
		log.Printf("[CONFIG ERROR] Failed to validate crew config: %v", err)
		return nil, fmt.Errorf("invalid crew configuration: %w", err)
	}

	log.Printf("[CONFIG SUCCESS] Crew config loaded: version=%s, agents=%d, entry=%s",
		config.Version, len(config.Agents), config.EntryPoint)
	return &config, nil
}

// LoadAndValidateCrewConfig loads crew config and performs comprehensive validation
// including circular routing detection and reachability analysis
// ✅ Issue #16: Configuration Validation - Advanced validation with circular reference detection
func LoadAndValidateCrewConfig(crewConfigPath string, agentConfigs map[string]*AgentConfig) (*CrewConfig, error) {
	// Load crew configuration
	config, err := LoadCrewConfig(crewConfigPath)
	if err != nil {
		return nil, err
	}

	// Perform comprehensive validation with circular routing detection
	validator := NewConfigValidator(config, agentConfigs)
	if err := validator.ValidateAll(); err != nil {
		log.Printf("[CONFIG VALIDATION ERROR] %v", err)
		return nil, fmt.Errorf("comprehensive configuration validation failed: %w", err)
	}

	// Check for warnings
	warnings := validator.GetWarnings()
	if len(warnings) > 0 {
		log.Printf("[CONFIG WARNINGS] %d warning(s) found during validation:", len(warnings))
		for _, w := range warnings {
			log.Printf("  - %s: %s", w.Field, w.Message)
		}
	}

	return config, nil
}

// LoadAgentConfig loads an agent configuration from a YAML file
func LoadAgentConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent config: %w", err)
	}

	var config AgentConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent config YAML: %w", err)
	}

	// Handle backward compatibility: convert old format to new format
	if config.Primary == nil {
		// Old format: model, provider, provider_url at top level
		config.Primary = &ModelConfigYAML{
			Model:       config.Model,
			Provider:    config.Provider,
			ProviderURL: config.ProviderURL,
		}

		// Set defaults for backward compatibility
		if config.Primary.Model == "" {
			config.Primary.Model = "gpt-4o"
		}
		if config.Primary.Provider == "" {
			config.Primary.Provider = "openai"
		}
	}

	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	// ✅ FIX for Issue #6: Validate agent configuration at load time
	// This catches invalid agent configs immediately with clear error messages
	if err := ValidateAgentConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid agent configuration: %w", err)
	}

	return &config, nil
}

// LoadAgentConfigs loads all agent configurations from a directory
func LoadAgentConfigs(dir string) (map[string]*AgentConfig, error) {
	configs := make(map[string]*AgentConfig)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			filePath := filepath.Join(dir, entry.Name())
			config, err := LoadAgentConfig(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to load agent config %s: %w", entry.Name(), err)
			}
			if config.ID != "" {
				configs[config.ID] = config
			}
		}
	}

	return configs, nil
}

// ValidateCrewConfig validates crew configuration structure and constraints
// ✅ FIX for Issue #6: Validate YAML config at load time instead of runtime
// This prevents invalid configs from causing runtime crashes
func ValidateCrewConfig(config *CrewConfig) error {
	// Validate required fields
	if config.Version == "" {
		return fmt.Errorf("required field 'version' is empty")
	}
	if len(config.Agents) == 0 {
		return fmt.Errorf("required field 'agents' is empty - at least one agent must be configured")
	}
	if config.EntryPoint == "" {
		return fmt.Errorf("required field 'entry_point' is empty")
	}

	// Validate entry_point exists in agents
	entryExists := false
	agentMap := make(map[string]bool)
	for _, agent := range config.Agents {
		agentMap[agent] = true
		if agent == config.EntryPoint {
			entryExists = true
		}
	}
	if !entryExists {
		return fmt.Errorf("entry_point '%s' not found in agents list", config.EntryPoint)
	}

	// Validate field constraints
	if config.Settings.MaxHandoffs < 0 {
		return fmt.Errorf("settings.max_handoffs must be >= 0, got %d", config.Settings.MaxHandoffs)
	}
	if config.Settings.MaxRounds <= 0 {
		return fmt.Errorf("settings.max_rounds must be > 0, got %d", config.Settings.MaxRounds)
	}
	if config.Settings.TimeoutSeconds <= 0 {
		return fmt.Errorf("settings.timeout_seconds must be > 0, got %d", config.Settings.TimeoutSeconds)
	}

	// Validate routing references
	if config.Routing != nil {
		// Validate signals reference existing agents
		for agentID, signals := range config.Routing.Signals {
			if !agentMap[agentID] {
				return fmt.Errorf("routing.signals references non-existent agent '%s'", agentID)
			}
			for _, signal := range signals {
				// Allow empty target for terminal signals
				if signal.Target != "" && !agentMap[signal.Target] {
					return fmt.Errorf("routing signal from agent '%s' targets non-existent agent '%s'", agentID, signal.Target)
				}
			}
		}

		// Validate agent behaviors reference existing agents
		for agentID := range config.Routing.AgentBehaviors {
			if !agentMap[agentID] {
				return fmt.Errorf("routing.agent_behaviors references non-existent agent '%s'", agentID)
			}
		}

		// Validate parallel groups reference existing agents
		for groupName, group := range config.Routing.ParallelGroups {
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
	}

	return nil
}

// ValidateAgentConfig validates agent configuration structure and constraints
// ✅ FIX for Issue #6: Validate agent config at load time
// ✅ Support for primary/backup LLM model configuration
func ValidateAgentConfig(config *AgentConfig) error {
	// ✅ FIX for Issue #23: Enhanced required field validation
	// Validate required fields strictly
	if config.ID == "" {
		return fmt.Errorf("agent: required field 'id' is empty")
	}
	if config.Name == "" {
		return fmt.Errorf("agent '%s': required field 'name' is empty", config.ID)
	}
	if config.Role == "" {
		return fmt.Errorf("agent '%s': required field 'role' is empty", config.ID)
	}

	// Validate field constraints
	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("agent '%s': temperature must be between 0 and 2, got %f", config.ID, config.Temperature)
	}

	// Validate primary LLM model configuration
	if config.Primary == nil {
		return fmt.Errorf("agent '%s': primary model configuration is missing", config.ID)
	}
	if config.Primary.Model == "" {
		return fmt.Errorf("agent '%s': primary.model is required", config.ID)
	}
	if config.Primary.Provider == "" {
		return fmt.Errorf("agent '%s': primary.provider is required", config.ID)
	}

	// Validate backup model configuration if present
	if config.Backup != nil {
		if config.Backup.Model == "" {
			return fmt.Errorf("agent '%s': backup.model must not be empty if backup is specified", config.ID)
		}
		if config.Backup.Provider == "" {
			return fmt.Errorf("agent '%s': backup.provider must not be empty if backup is specified", config.ID)
		}
		// Log info about backup configuration
		log.Printf("[CONFIG INFO] agent '%s': backup model '%s' (%s) configured", config.ID, config.Backup.Model, config.Backup.Provider)
	}

	// Warn about suspicious configurations
	if config.SystemPrompt == "" && config.Backstory == "" {
		log.Printf("[CONFIG WARNING] agent '%s': both 'system_prompt' and 'backstory' are empty - agent may not have proper context", config.ID)
	}

	return nil
}

// CreateAgentFromConfig creates an Agent from an AgentConfig
func CreateAgentFromConfig(config *AgentConfig, allTools map[string]*Tool) *Agent {
	// Convert YAML model config to runtime ModelConfig
	primary := &ModelConfig{
		Model:       config.Primary.Model,
		Provider:    config.Primary.Provider,
		ProviderURL: config.Primary.ProviderURL,
	}

	var backup *ModelConfig
	if config.Backup != nil {
		backup = &ModelConfig{
			Model:       config.Backup.Model,
			Provider:    config.Backup.Provider,
			ProviderURL: config.Backup.ProviderURL,
		}
	}

	agent := &Agent{
		ID:             config.ID,
		Name:           config.Name,
		Role:           config.Role,
		Backstory:      config.Backstory,
		Model:          config.Model,             // For backward compatibility
		SystemPrompt:   config.SystemPrompt,
		Provider:       config.Provider,          // For backward compatibility
		ProviderURL:    config.ProviderURL,       // For backward compatibility
		Primary:        primary,                  // New: primary model config
		Backup:         backup,                   // New: backup model config (optional)
		Temperature:    config.Temperature,
		IsTerminal:     config.IsTerminal,
		HandoffTargets: config.HandoffTargets,
		Tools:          []*Tool{},
	}

	// Add tools from config
	for _, toolName := range config.Tools {
		if tool, exists := allTools[toolName]; exists {
			agent.Tools = append(agent.Tools, tool)
		}
	}

	return agent
}
