package agentic

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ValidateAgentConfig validates an agent configuration
func ValidateAgentConfig(config *AgentConfig) error {
	// Validate Model is not empty
	if config.Model == "" {
		return fmt.Errorf("agent config validation failed: Model must be specified (examples: gpt-4o, gpt-4o-mini)")
	}

	// Validate Temperature is in valid range if specified
	if config.Temperature != nil {
		temp := *config.Temperature
		if temp < 0.0 || temp > 2.0 {
			return fmt.Errorf("agent config validation failed: Temperature must be between 0.0 and 2.0, got %.1f", temp)
		}
	}

	return nil
}

// LoadTeamConfig loads the team configuration from a YAML file
func LoadTeamConfig(path string) (*TeamConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read team config: %w", err)
	}

	var config TeamConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse team config: %w", err)
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

	return &config, nil
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
		return nil, fmt.Errorf("failed to parse agent config: %w", err)
	}

	// Set defaults
	if config.Model == "" {
		config.Model = "gpt-4o-mini"
	}
	if config.Temperature == nil {
		defaultTemp := 0.7
		config.Temperature = &defaultTemp
	}

	// Validate configuration
	if err := ValidateAgentConfig(&config); err != nil {
		return nil, err
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

// CreateAgentFromConfig creates an Agent from an AgentConfig
// Supports tools map for agent creation
func CreateAgentFromConfig(config *AgentConfig, allTools map[string]*Tool) *Agent {
	temperature := 0.7 // default
	if config.Temperature != nil {
		temperature = *config.Temperature
	}

	agent := &Agent{
		ID:             config.ID,
		Name:           config.Name,
		Role:           config.Role,
		Backstory:      config.Backstory,
		Model:          config.Model,
		SystemPrompt:   config.SystemPrompt,
		Temperature:    temperature,
		IsTerminal:     config.IsTerminal,
		HandoffTargets: config.HandoffTo,
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

// Deprecated: Use LoadTeamConfig instead
func LoadCrewConfig(path string) (*TeamConfig, error) {
	return LoadTeamConfig(path)
}
