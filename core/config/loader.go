package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/validation"
)

// LoadCrewConfig loads the crew configuration from a YAML file
func LoadCrewConfig(path string) (*common.CrewConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read crew config: %w", err)
	}

	var config common.CrewConfig
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

	// Validate crew configuration
	if err := validation.ValidateCrewConfig(&config); err != nil {
		return nil, fmt.Errorf("crew config validation failed: %w", err)
	}

	log.Printf("[CONFIG SUCCESS] Crew config loaded: version=%s, agents=%d, entry=%s",
		config.Version, len(config.Agents), config.EntryPoint)
	return &config, nil
}

// LoadAgentConfig loads an agent configuration from a YAML file
func LoadAgentConfig(path string) (*common.AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent config: %w", err)
	}

	var config common.AgentConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent config YAML: %w", err)
	}

	// Handle backward compatibility
	if config.Primary == nil {
		config.Primary = &common.ModelConfigYAML{
			Model:       config.Model,
			Provider:    config.Provider,
			ProviderURL: config.ProviderURL,
		}
	}

	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	// Backward compatibility: Convert old flat format to new nested format
	if config.CostLimits == nil && config.MaxTokensPerCall > 0 {
		config.CostLimits = &common.CostLimitsConfig{
			MaxTokensPerCall:    config.MaxTokensPerCall,
			MaxTokensPerDay:     config.MaxTokensPerDay,
			MaxCostPerDayUSD:    config.MaxCostPerDay,
			AlertThreshold:      config.CostAlertThreshold,
			BlockOnCostExceed:   config.EnforceCostLimits,
		}
	}

	// Set defaults for cost control if still not configured
	if config.CostLimits == nil {
		config.CostLimits = &common.CostLimitsConfig{
			MaxTokensPerCall:           1000,
			MaxTokensPerDay:            50000,
			MaxCostPerDayUSD:           10.0,
			AlertThreshold:             0.80,
			BlockOnCostExceed:          true,
			InputTokenPricePerMillion:  0.15,
			OutputTokenPricePerMillion: 0.60,
		}
	}

	// Set defaults for memory control if not configured
	if config.MemoryLimits == nil {
		config.MemoryLimits = &common.MemoryLimitsConfig{
			MaxPerCallMB:        100,
			MaxPerDayMB:         1000,
			BlockOnMemoryExceed: true,
		}
	}

	// Set defaults for error control if not configured
	if config.ErrorLimits == nil {
		config.ErrorLimits = &common.ErrorLimitsConfig{
			MaxConsecutive:     3,
			MaxPerDay:          10,
			BlockOnErrorExceed: true,
		}
	}

	// Set defaults for logging if not configured
	if config.Logging == nil {
		config.Logging = &common.LoggingConfig{
			EnableMemoryMetrics:      true,
			EnablePerformanceMetrics: true,
			EnableQuotaWarnings:      true,
			LogLevel:                 "info",
		}
	}

	// Validate agent configuration
	if err := validation.ValidateAgentConfig(&config); err != nil {
		return nil, fmt.Errorf("agent config validation failed: %w", err)
	}

	return &config, nil
}

// LoadAgentConfigs loads multiple agent configurations from a map
func LoadAgentConfigs(agentPaths map[string]string) (map[string]*common.AgentConfig, error) {
	configs := make(map[string]*common.AgentConfig)

	for agentID, path := range agentPaths {
		cfg, err := LoadAgentConfig(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load agent config for %s: %w", agentID, err)
		}
		configs[agentID] = cfg
	}

	return configs, nil
}

// ExpandEnvVars expands environment variables in configuration strings
func ExpandEnvVars(value string) string {
	return os.ExpandEnv(value)
}
