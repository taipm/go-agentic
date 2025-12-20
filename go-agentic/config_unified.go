package agentic

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ============================================
// Unified Configuration Types
// ============================================

// UnifiedTeamConfig represents a complete team configuration in a single YAML file
type UnifiedTeamConfig struct {
	Team    UnifiedTeamMetadata              `yaml:"team"`
	Agents  map[string]*UnifiedAgentConfig   `yaml:"agents"`
	Tools   map[string]*UnifiedToolConfig    `yaml:"tools"`
	Routing *UnifiedRoutingConfig            `yaml:"routing,omitempty"`
}

// UnifiedTeamMetadata contains team-level settings
type UnifiedTeamMetadata struct {
	Name   string `yaml:"name,omitempty"`
	Config struct {
		MaxRounds   int `yaml:"maxRounds"`
		MaxHandoffs int `yaml:"maxHandoffs"`
	} `yaml:"config"`
}

// UnifiedAgentConfig represents an agent in unified configuration
type UnifiedAgentConfig struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Role        string   `yaml:"role"`
	Backstory   string   `yaml:"backstory"`
	Model       string   `yaml:"model"`
	Temperature float64  `yaml:"temperature"`
	IsTerminal  bool     `yaml:"isTerminal"`
	Tools       []string `yaml:"tools,omitempty"`
}

// UnifiedToolConfig represents a tool in unified configuration
type UnifiedToolConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Parameters  map[string]interface{} `yaml:"parameters,omitempty"`
}

// UnifiedRoutingConfig defines routing rules
type UnifiedRoutingConfig struct {
	Type  string                     `yaml:"type"`
	Rules []UnifiedRoutingRule       `yaml:"rules,omitempty"`
}

// UnifiedRoutingRule defines a single routing rule
type UnifiedRoutingRule struct {
	FromAgent   string  `yaml:"from_agent"`
	Trigger     string  `yaml:"trigger"`
	TargetAgent *string `yaml:"target_agent,omitempty"`
	Description string  `yaml:"description,omitempty"`
}

// ============================================
// Configuration Loader
// ============================================

// ToolHandlerRegistry maps tool names to their handler functions
type ToolHandlerRegistry map[string]ToolHandler

// LoadTeamFromYAML loads a complete team configuration from a single YAML file
// toolHandlers is a map of tool IDs/names to their handler functions
func LoadTeamFromYAML(yamlPath string, toolHandlers ToolHandlerRegistry) (*Team, error) {
	// Read YAML file
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Parse YAML
	var config UnifiedTeamConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate team config
	if err := validateUnifiedTeamConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid team configuration: %w", err)
	}

	// Create agents from config
	agents := make([]*Agent, 0, len(config.Agents))
	agentMap := make(map[string]*Agent)

	for agentID, agentCfg := range config.Agents {
		if agentCfg.ID == "" {
			agentCfg.ID = agentID
		}

		// Create agent using fluent builder
		builder := NewAgent(agentCfg.ID, agentCfg.Name).
			WithRole(agentCfg.Role).
			WithBackstory(agentCfg.Backstory).
			WithModel(agentCfg.Model).
			WithTemperature(agentCfg.Temperature).
			SetTerminal(agentCfg.IsTerminal)

		// Add tools from config
		for _, toolID := range agentCfg.Tools {
			if handler, ok := toolHandlers[toolID]; ok {
				toolCfg, toolExists := config.Tools[toolID]
				if !toolExists {
					return nil, fmt.Errorf("tool '%s' referenced but not defined in tools section", toolID)
				}

				// Create tool using fluent builder
				tool := NewTool(toolCfg.Name, toolCfg.Description)

				// Add parameters if defined
				if toolCfg.Parameters != nil {
					if props, ok := toolCfg.Parameters["properties"].(map[string]interface{}); ok && len(props) > 0 {
						// Convert to ParameterDef map
						params := make(map[string]ParameterDef)
						for paramName, paramDef := range props {
							if paramMap, ok := paramDef.(map[string]interface{}); ok {
								paramType := "string"
								if t, ok := paramMap["type"].(string); ok {
									paramType = t
								}
								description := ""
								if desc, ok := paramMap["description"].(string); ok {
									description = desc
								}
								params[paramName] = ParameterDef{
									Type:        paramType,
									Description: description,
									Required:    false, // TODO: parse from required list if needed
								}
							}
						}
						tool.WithParameters(params)
					} else {
						tool.NoParameters()
					}
				} else {
					tool.NoParameters()
				}

				tool.Handler(handler)
				builder.AddTool(tool.Build())
			} else {
				return nil, fmt.Errorf("tool '%s' referenced but no handler provided", toolID)
			}
		}

		agent := builder.Build()
		agents = append(agents, agent)
		agentMap[agent.ID] = agent
	}

	// Create team using fluent builder
	team := NewTeam().
		AddAgents(agents...).
		WithMaxRounds(config.Team.Config.MaxRounds).
		WithMaxHandoffs(config.Team.Config.MaxHandoffs).
		Build()

	// Set routing if provided
	if config.Routing != nil {
		team.RoutingRules = config.Routing.Rules
	}

	return team, nil
}

// ============================================
// Validation
// ============================================

// validateUnifiedTeamConfig validates the unified team configuration
func validateUnifiedTeamConfig(config *UnifiedTeamConfig) error {
	// Validate team config
	if config.Team.Config.MaxRounds <= 0 {
		return fmt.Errorf("maxRounds must be > 0, got %d", config.Team.Config.MaxRounds)
	}
	if config.Team.Config.MaxHandoffs < 0 {
		return fmt.Errorf("maxHandoffs must be >= 0, got %d", config.Team.Config.MaxHandoffs)
	}

	// Validate agents section exists
	if len(config.Agents) == 0 {
		return fmt.Errorf("at least one agent must be defined")
	}

	// Validate each agent
	hasTerminal := false
	for agentID, agent := range config.Agents {
		if agent.Name == "" {
			return fmt.Errorf("agent '%s' has empty name", agentID)
		}
		if agent.Role == "" {
			return fmt.Errorf("agent '%s' has empty role", agentID)
		}
		if agent.Backstory == "" {
			return fmt.Errorf("agent '%s' has empty backstory", agentID)
		}
		if agent.Model == "" {
			return fmt.Errorf("agent '%s' has empty model", agentID)
		}

		if agent.IsTerminal {
			hasTerminal = true
		}

		// Validate referenced tools exist
		for _, toolID := range agent.Tools {
			if _, ok := config.Tools[toolID]; !ok {
				return fmt.Errorf("agent '%s' references tool '%s' which is not defined", agentID, toolID)
			}
		}
	}

	if !hasTerminal {
		return fmt.Errorf("at least one agent must have isTerminal=true")
	}

	// Validate tools section
	for toolID, tool := range config.Tools {
		if tool.Name == "" {
			return fmt.Errorf("tool '%s' has empty name", toolID)
		}
		if tool.Description == "" {
			return fmt.Errorf("tool '%s' has empty description", toolID)
		}
	}

	return nil
}

// ============================================
// Backward Compatibility Helpers
// ============================================

// LoadTeamFromYAMLWithDefaults loads team config with sensible defaults
func LoadTeamFromYAMLWithDefaults(yamlPath string, toolHandlers ToolHandlerRegistry) (*Team, error) {
	team, err := LoadTeamFromYAML(yamlPath, toolHandlers)
	if err != nil {
		return nil, err
	}

	// Set defaults if not provided
	if team.MaxRounds <= 0 {
		team.MaxRounds = 10
	}
	if team.MaxHandoffs < 0 {
		team.MaxHandoffs = 3
	}

	return team, nil
}

// ============================================
// Configuration Export (for reference/migration)
// ============================================

// ExportTeamToYAML exports a team configuration to YAML format
func ExportTeamToYAML(team *Team) ([]byte, error) {
	config := UnifiedTeamConfig{
		Team: UnifiedTeamMetadata{
			Config: struct {
				MaxRounds   int `yaml:"maxRounds"`
				MaxHandoffs int `yaml:"maxHandoffs"`
			}{
				MaxRounds:   team.MaxRounds,
				MaxHandoffs: team.MaxHandoffs,
			},
		},
		Agents: make(map[string]*UnifiedAgentConfig),
		Tools:  make(map[string]*UnifiedToolConfig),
	}

	// Export agents
	for _, agent := range team.Agents {
		agentCfg := &UnifiedAgentConfig{
			ID:          agent.ID,
			Name:        agent.Name,
			Role:        agent.Role,
			Backstory:   agent.Backstory,
			Model:       agent.Model,
			Temperature: agent.Temperature,
			IsTerminal:  agent.IsTerminal,
			Tools:       make([]string, 0),
		}

		// Export tool references
		for _, tool := range agent.Tools {
			agentCfg.Tools = append(agentCfg.Tools, tool.Name)

			// Export tool definition if not already exported
			if _, exists := config.Tools[tool.Name]; !exists {
				config.Tools[tool.Name] = &UnifiedToolConfig{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				}
			}
		}

		config.Agents[agent.ID] = agentCfg
	}

	// Export routing if present
	if len(team.RoutingRules) > 0 {
		config.Routing = &UnifiedRoutingConfig{
			Type:  "signal",
			Rules: team.RoutingRules,
		}
	}

	return yaml.Marshal(config)
}
