package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

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

// LoadAgentConfigs loads all agent configurations from a directory
// Returns a map of agent ID to agent config
func LoadAgentConfigs(dir string) (map[string]*common.AgentConfig, error) {
	configs := make(map[string]*common.AgentConfig)

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

// ExpandEnvVars expands environment variables in configuration strings
func ExpandEnvVars(value string) string {
	return os.ExpandEnv(value)
}

// CreateAgentFromConfig creates an Agent from an AgentConfig
// Initializes all metadata, quotas, and metrics for the agent
// Note: allTools is map of tool name to tool object (generic interface{})
func CreateAgentFromConfig(agentConfig *common.AgentConfig, allTools map[string]interface{}) *common.Agent {
	// Step 1: Convert model configs
	primary, backup := convertToModelConfig(agentConfig.Primary, agentConfig.Backup)

	// Step 2: Create metadata with quotas and metrics
	metadata := buildAgentMetadata(agentConfig)

	// Step 3: Create agent with all fields
	agent := &common.Agent{
		ID:             agentConfig.ID,
		Name:           agentConfig.Name,
		Role:           agentConfig.Role,
		Backstory:      agentConfig.Backstory,
		SystemPrompt:   agentConfig.SystemPrompt,
		PrimaryModel:   primary,
		BackupModel:    backup,
		Temperature:    agentConfig.Temperature,
		IsTerminal:     agentConfig.IsTerminal,
		HandoffTargets: []*common.Agent{},
		Tools:          []interface{}{},

		Metadata:           metadata,
		Quota:              metadata.QuotaLimits,
		CostMetrics:        metadata.CostMetrics,
		MemoryMetrics:      metadata.MemoryMetrics,
		PerformanceMetrics: metadata.PerformanceMetrics,
	}

	// Step 4: Add tools from config
	addAgentTools(agent, agentConfig, allTools)

	return agent
}

// convertToModelConfig converts YAML model config to runtime ModelConfig
func convertToModelConfig(primary *common.ModelConfigYAML, backup *common.ModelConfigYAML) (*common.ModelConfig, *common.ModelConfig) {
	primaryConfig := &common.ModelConfig{
		Model:       primary.Model,
		Provider:    primary.Provider,
		ProviderURL: primary.ProviderURL,
	}

	var backupConfig *common.ModelConfig
	if backup != nil {
		backupConfig = &common.ModelConfig{
			Model:       backup.Model,
			Provider:    backup.Provider,
			ProviderURL: backup.ProviderURL,
		}
	}

	return primaryConfig, backupConfig
}

// buildAgentMetadata creates AgentMetadata with all metrics initialized
func buildAgentMetadata(agentConfig *common.AgentConfig) *common.AgentMetadata {
	now := time.Now()

	return &common.AgentMetadata{
		ID:        agentConfig.ID,
		Name:      agentConfig.Name,
		Role:      agentConfig.Role,
		CreatedAt: now,
		QuotaLimits: buildAgentQuotas(agentConfig),

		CostMetrics: &common.AgentCostMetrics{
			CallCount:     0,
			TotalTokens:   0,
			DailyCost:     0,
			LastResetTime: time.Time{},
			Mutex:         sync.RWMutex{},
		},

		MemoryMetrics: &common.AgentMemoryMetrics{
			CurrentMemoryMB:     0,
			PeakMemoryMB:        0,
			AverageMemoryMB:     0,
			MemoryTrendPercent:  0,
			MaxMemoryMB:         512,
			MaxDailyMemoryGB:    10,
			MemoryAlertPercent:  0.80,
			CurrentContextSize:  0,
			MaxContextWindow:    32000,
			ContextTrimPercent:  0.20,
			AverageCallDuration: 0,
			SlowCallThreshold:   30 * time.Second,
			Mutex:               sync.RWMutex{},
		},

		PerformanceMetrics: &common.AgentPerformanceMetrics{
			SuccessfulCalls:      0,
			FailedCalls:          0,
			SuccessRate:          100.0,
			AverageResponseTime:  0,
			LastError:            "",
			LastErrorTime:        time.Time{},
			ConsecutiveErrors:    0,
			ErrorCountToday:      0,
			MaxErrorsPerHour:     10,
			MaxErrorsPerDay:      50,
			MaxConsecutiveErrors: 5,
			Mutex:                sync.RWMutex{},
		},

		AllowedToolNames: agentConfig.Tools,
	}
}

// buildAgentQuotas creates quota limits from agent config
func buildAgentQuotas(agentConfig *common.AgentConfig) *common.AgentQuotaLimits {
	return &common.AgentQuotaLimits{
		MaxTokensPerCall:   agentConfig.MaxTokensPerCall,
		MaxTokensPerDay:    agentConfig.MaxTokensPerDay,
		MaxCostPerDay:      agentConfig.MaxCostPerDay,
		CostAlertPercent:   agentConfig.CostAlertThreshold,
		MaxMemoryPerCall:   512,
		MaxMemoryPerDay:    10240,
		MaxContextWindow:   32000,
		MaxCallsPerMinute:  60,
		MaxCallsPerHour:    1000,
		MaxCallsPerDay:     10000,
		MaxErrorsPerHour:   10,
		MaxErrorsPerDay:    50,
		BlockOnQuotaExceed: agentConfig.EnforceCostLimits,
	}
}

// addAgentTools adds configured tools to the agent from the tools map
func addAgentTools(agent *common.Agent, agentConfig *common.AgentConfig, allTools map[string]interface{}) {
	for _, toolName := range agentConfig.Tools {
		if tool, exists := allTools[toolName]; exists {
			agent.Tools = append(agent.Tools, tool)
		}
	}
}
