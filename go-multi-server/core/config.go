package crewai

import (
	"fmt"
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

// AgentConfig represents an agent configuration
type AgentConfig struct {
	ID             string   `yaml:"id"`
	Name           string   `yaml:"name"`
	Description    string   `yaml:"description"`
	Role           string   `yaml:"role"`
	Backstory      string   `yaml:"backstory"`
	Model          string   `yaml:"model"`
	Temperature    float64  `yaml:"temperature"`
	IsTerminal     bool     `yaml:"is_terminal"`
	Tools          []string `yaml:"tools"`
	HandoffTargets []string `yaml:"handoff_targets"`
	SystemPrompt   string   `yaml:"system_prompt"`
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
		return nil, fmt.Errorf("failed to parse crew config: %w", err)
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
		config.Model = "gpt-4o"
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
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
func CreateAgentFromConfig(config *AgentConfig, allTools map[string]*Tool) *Agent {
	agent := &Agent{
		ID:             config.ID,
		Name:           config.Name,
		Role:           config.Role,
		Backstory:      config.Backstory,
		Model:          config.Model,
		SystemPrompt:   config.SystemPrompt,
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
