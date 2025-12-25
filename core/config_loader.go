package crewai

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/taipm/go-agentic/core/common"
	"gopkg.in/yaml.v3"
)

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
// ✅ FIX for Issue #5: Add configMode parameter for STRICT/PERMISSIVE mode validation
func LoadAgentConfig(path string, configMode ConfigMode) (*AgentConfig, error) {
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

		// ✅ FIX for Issue #5: Don't set defaults here - validation layer will handle
		// based on config mode (STRICT vs PERMISSIVE)
		// This prevents STRICT MODE from being bypassed
	}

	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	// Backward compatibility: Convert old flat format to new nested format
	// If new nested configs don't exist but old flat fields are set, populate nested configs
	if config.CostLimits == nil && config.MaxTokensPerCall > 0 {
		config.CostLimits = &CostLimitsConfig{
			MaxTokensPerCall:    config.MaxTokensPerCall,
			MaxTokensPerDay:     config.MaxTokensPerDay,
			MaxCostPerDayUSD:    config.MaxCostPerDay,
			AlertThreshold:      config.CostAlertThreshold,
			BlockOnCostExceed:   config.EnforceCostLimits,
		}
	}

	// Set defaults for cost control if still not configured
	if config.CostLimits == nil {
		config.CostLimits = &CostLimitsConfig{
			MaxTokensPerCall:  1000,   // Default: 1K tokens per call
			MaxTokensPerDay:   50000,  // Default: 50K tokens per day
			MaxCostPerDayUSD:  10.0,   // Default: $10 per day
			AlertThreshold:    0.80,   // Default: warn at 80% usage
			BlockOnCostExceed: true,   // ✅ Default: BLOCK mode (production-safe)
		}
	}

	// Set defaults for memory control if not configured
	if config.MemoryLimits == nil {
		config.MemoryLimits = &MemoryLimitsConfig{
			MaxPerCallMB:        100,   // Default: 100 MB per call
			MaxPerDayMB:         1000,  // Default: 1000 MB per day
			BlockOnMemoryExceed: true,  // ✅ Default: BLOCK mode (production-safe)
		}
	}

	// Set defaults for error control if not configured
	if config.ErrorLimits == nil {
		config.ErrorLimits = &ErrorLimitsConfig{
			MaxConsecutive:     3,    // Default: max 3 consecutive errors
			MaxPerDay:          10,   // Default: max 10 errors per day
			BlockOnErrorExceed: true, // ✅ Default: BLOCK mode (production-safe)
		}
	}

	// Set defaults for logging if not configured
	if config.Logging == nil {
		config.Logging = &LoggingConfig{
			EnableMemoryMetrics:      true,
			EnablePerformanceMetrics: true,
			EnableQuotaWarnings:      true,
			LogLevel:                 "info",
		}
	}

	// Validate agent configuration at load time
	// This catches invalid agent configs immediately with clear error messages
	// ✅ FIX for Issue #5: Pass configMode to validation for STRICT/PERMISSIVE logic
	if err := ValidateAgentConfig(&config, configMode); err != nil {
		return nil, fmt.Errorf("invalid agent configuration: %w", err)
	}

	return &config, nil
}

// LoadAgentConfigs loads all agent configurations from a directory
// ✅ FIX for Issue #5: Add configMode parameter for STRICT/PERMISSIVE validation
func LoadAgentConfigs(dir string, configMode ConfigMode) (map[string]*AgentConfig, error) {
	configs := make(map[string]*AgentConfig)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yaml" {
			filePath := filepath.Join(dir, entry.Name())
			// ✅ FIX for Issue #5: Pass configMode to LoadAgentConfig
			config, err := LoadAgentConfig(filePath, configMode)
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
// convertToModelConfig converts YAML model config to runtime ModelConfig
func convertToModelConfig(primary *ModelConfigYAML, backup *ModelConfigYAML) (*ModelConfig, *ModelConfig) {
	primaryConfig := &ModelConfig{
		Model:       primary.Model,
		Provider:    primary.Provider,
		ProviderURL: primary.ProviderURL,
	}

	var backupConfig *ModelConfig
	if backup != nil {
		backupConfig = &ModelConfig{
			Model:       backup.Model,
			Provider:    backup.Provider,
			ProviderURL: backup.ProviderURL,
		}
	}

	return primaryConfig, backupConfig
}

// buildAgentQuotas creates quota limits from agent config
func buildAgentQuotas(config *AgentConfig) *common.AgentQuotaLimits {
	return &common.AgentQuotaLimits{
		MaxTokensPerCall:   config.MaxTokensPerCall,
		MaxTokensPerDay:    config.MaxTokensPerDay,
		MaxCostPerDay:      config.MaxCostPerDay,
		CostAlertPercent:   config.CostAlertThreshold,
		MaxMemoryPerCall:   512,
		MaxMemoryPerDay:    10240,
		MaxContextWindow:   32000,
		MaxCallsPerMinute:  60,
		MaxCallsPerHour:    1000,
		MaxCallsPerDay:     10000,
		MaxErrorsPerHour:   10,
		MaxErrorsPerDay:    50,
		BlockOnQuotaExceed: config.EnforceCostLimits,
	}
}

// buildAgentMetadata creates AgentMetadata with all metrics initialized
func buildAgentMetadata(config *AgentConfig) *AgentMetadata {
	now := time.Now()

	return &AgentMetadata{
		ID:        config.ID,
		Name:      config.Name,
		Role:      config.Role,
		CreatedAt: now,
		QuotaLimits: buildAgentQuotas(config),

		CostMetrics: &AgentCostMetrics{
			CallCount:     0,
			TotalTokens:   0,
			DailyCost:     0,
			LastResetTime: time.Time{},
			Mutex:         sync.RWMutex{},
		},

		MemoryMetrics: &AgentMemoryMetrics{
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

		PerformanceMetrics: &AgentPerformanceMetrics{
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

		AllowedToolNames: config.Tools,
	}
}

// addAgentTools adds configured tools to the agent from the tools map
func addAgentTools(agent *Agent, config *AgentConfig, allTools map[string]*Tool) {
	for _, toolName := range config.Tools {
		if tool, exists := allTools[toolName]; exists {
			agent.Tools = append(agent.Tools, tool)
		}
	}
}

// ✅ WEEK 2: Initialize unified AgentMetadata with quotas and metrics
func CreateAgentFromConfig(config *AgentConfig, allTools map[string]*Tool) *Agent {
	// Step 1: Convert model configs
	primary, backup := convertToModelConfig(config.Primary, config.Backup)

	// Step 2: Create metadata with quotas and metrics
	metadata := buildAgentMetadata(config)

	// Step 3: Create agent with all fields
	agent := &Agent{
		ID:             config.ID,
		Name:           config.Name,
		Role:           config.Role,
		Backstory:      config.Backstory,
		SystemPrompt:   config.SystemPrompt,
		PrimaryModel:   primary,
		BackupModel:    backup,
		Temperature:    config.Temperature,
		IsTerminal:     config.IsTerminal,
		HandoffTargets: []*Agent{}, // TODO: Resolve string IDs to Agent pointers after all agents are created
		Tools:          []interface{}{},

		Metadata: metadata,
		Quota:    metadata.QuotaLimits,
		CostMetrics: metadata.CostMetrics,
		MemoryMetrics: metadata.MemoryMetrics,
		PerformanceMetrics: metadata.PerformanceMetrics,
	}

	// Step 4: Add tools from config
	addAgentTools(agent, config, allTools)

	return agent
}

// getInputTokenPrice extracts input token price from config, returns default if not set
func getInputTokenPrice(costLimits *CostLimitsConfig) float64 {
	if costLimits != nil && costLimits.InputTokenPricePerMillion > 0 {
		return costLimits.InputTokenPricePerMillion
	}
	return 0.15 // Default: gpt-4o-mini pricing
}

// getOutputTokenPrice extracts output token price from config, returns default if not set
func getOutputTokenPrice(costLimits *CostLimitsConfig) float64 {
	if costLimits != nil && costLimits.OutputTokenPricePerMillion > 0 {
		return costLimits.OutputTokenPricePerMillion
	}
	return 0.60 // Default: gpt-4o-mini pricing
}

// ConfigToHardcodedDefaults converts CrewConfig settings to HardcodedDefaults struct
// ✅ Phase 4: Extended Configuration - Maps YAML values to runtime defaults
// Returns defaults with YAML overrides applied; validation is performed after conversion
// ✅ Phase 5.1: In STRICT MODE, missing values are NOT defaulted and remain 0, causing validation to fail
func ConfigToHardcodedDefaults(config *CrewConfig) *HardcodedDefaults {
	// In PERMISSIVE MODE: Start with all defaults
	// In STRICT MODE: Start with all 0 values (except mode), require explicit YAML config
	var defaults *HardcodedDefaults

	// ✅ Phase 5.1: Check config mode FIRST
	configMode := PermissiveMode
	if config.Settings.ConfigMode != "" {
		configMode = ConfigMode(config.Settings.ConfigMode)
	}

	// In STRICT MODE, don't use defaults - start with empty values
	if configMode == StrictMode {
		defaults = &HardcodedDefaults{
			Mode: StrictMode,
			// All timeout/int fields default to 0
			// All duration fields default to 0 (0 seconds)
			// All float fields default to 0
			// Validation will catch these as errors
		}
	} else {
		// In PERMISSIVE MODE, start with defaults
		defaults = DefaultHardcodedDefaults()
		defaults.Mode = PermissiveMode
	}

	// Phase 1 configurations
	if config.Settings.ParallelTimeoutSeconds > 0 {
		defaults.ParallelAgentTimeout = time.Duration(config.Settings.ParallelTimeoutSeconds) * time.Second
	}
	if config.Settings.MaxToolOutputChars > 0 {
		defaults.MaxToolOutputChars = config.Settings.MaxToolOutputChars
	}
	if config.Settings.MaxTotalToolOutputChars > 0 {
		defaults.MaxTotalToolOutputChars = config.Settings.MaxTotalToolOutputChars
	}

	// Phase 4 timeout configurations
	if config.Settings.ToolExecutionTimeoutSeconds > 0 {
		defaults.ToolExecutionTimeout = time.Duration(config.Settings.ToolExecutionTimeoutSeconds) * time.Second
	}
	if config.Settings.ToolResultTimeoutSeconds > 0 {
		defaults.ToolResultTimeout = time.Duration(config.Settings.ToolResultTimeoutSeconds) * time.Second
	}
	if config.Settings.MinToolTimeoutMillis > 0 {
		defaults.MinToolTimeout = time.Duration(config.Settings.MinToolTimeoutMillis) * time.Millisecond
	}
	if config.Settings.StreamChunkTimeoutMillis > 0 {
		defaults.StreamChunkTimeout = time.Duration(config.Settings.StreamChunkTimeoutMillis) * time.Millisecond
	}
	if config.Settings.SSEKeepAliveSeconds > 0 {
		defaults.SSEKeepAliveInterval = time.Duration(config.Settings.SSEKeepAliveSeconds) * time.Second
	}
	if config.Settings.RequestStoreCleanupMinutes > 0 {
		defaults.RequestStoreCleanupInterval = time.Duration(config.Settings.RequestStoreCleanupMinutes) * time.Minute
	}

	// Phase 4 retry and backoff configurations
	if config.Settings.RetryBackoffMinMillis > 0 {
		defaults.RetryBackoffMinDuration = time.Duration(config.Settings.RetryBackoffMinMillis) * time.Millisecond
	}
	if config.Settings.RetryBackoffMaxSeconds > 0 {
		defaults.RetryBackoffMaxDuration = time.Duration(config.Settings.RetryBackoffMaxSeconds) * time.Second
	}

	// Phase 4 input validation limits
	if config.Settings.MaxInputSizeKB > 0 {
		defaults.MaxInputSize = config.Settings.MaxInputSizeKB * 1024
	}
	if config.Settings.MinAgentIDLength > 0 {
		defaults.MinAgentIDLength = config.Settings.MinAgentIDLength
	}
	if config.Settings.MaxAgentIDLength > 0 {
		defaults.MaxAgentIDLength = config.Settings.MaxAgentIDLength
	}
	if config.Settings.MaxRequestBodySizeKB > 0 {
		defaults.MaxRequestBodySize = config.Settings.MaxRequestBodySizeKB * 1024
	}

	// Phase 4 output and storage
	if config.Settings.StreamBufferSize > 0 {
		defaults.StreamBufferSize = config.Settings.StreamBufferSize
	}
	if config.Settings.MaxStoredRequests > 0 {
		defaults.MaxStoredRequests = config.Settings.MaxStoredRequests
	}

	// Phase 4 client cache
	if config.Settings.ClientCacheTTLMinutes > 0 {
		defaults.ClientCacheTTL = time.Duration(config.Settings.ClientCacheTTLMinutes) * time.Minute
	}

	// Phase 4 graceful shutdown
	if config.Settings.GracefulShutdownCheckMillis > 0 {
		defaults.GracefulShutdownCheckInterval = time.Duration(config.Settings.GracefulShutdownCheckMillis) * time.Millisecond
	}
	if config.Settings.TimeoutWarningThresholdPct > 0 && config.Settings.TimeoutWarningThresholdPct <= 100 {
		defaults.TimeoutWarningThreshold = float64(config.Settings.TimeoutWarningThresholdPct) / 100.0
	}

	// ✅ WEEK 1: Cost Control configurations
	if config.Settings.MaxTokensPerCall > 0 {
		defaults.MaxTokensPerCall = config.Settings.MaxTokensPerCall
	}
	if config.Settings.MaxTokensPerDay > 0 {
		defaults.MaxTokensPerDay = config.Settings.MaxTokensPerDay
	}
	if config.Settings.MaxCostPerDay > 0 {
		defaults.MaxCostPerDay = config.Settings.MaxCostPerDay
	}
	if config.Settings.CostAlertThreshold > 0 {
		defaults.CostAlertThreshold = config.Settings.CostAlertThreshold
	}

	// ✅ WEEK 2: Memory Management configurations
	if config.Settings.MaxMemoryMB > 0 {
		defaults.MaxMemoryMB = config.Settings.MaxMemoryMB
	}
	if config.Settings.MaxDailyMemoryGB > 0 {
		defaults.MaxDailyMemoryGB = config.Settings.MaxDailyMemoryGB
	}
	if config.Settings.MemoryAlertPercent > 0 {
		defaults.MemoryAlertPercent = config.Settings.MemoryAlertPercent
	}
	if config.Settings.MaxContextWindow > 0 {
		defaults.MaxContextWindow = config.Settings.MaxContextWindow
	}
	if config.Settings.ContextTrimPercent > 0 {
		defaults.ContextTrimPercent = config.Settings.ContextTrimPercent
	}
	if config.Settings.SlowCallThresholdSec > 0 {
		defaults.SlowCallThreshold = time.Duration(config.Settings.SlowCallThresholdSec) * time.Second
	}

	// ✅ WEEK 2: Performance & Reliability configurations
	if config.Settings.MaxErrorsPerHour > 0 {
		defaults.MaxErrorsPerHour = config.Settings.MaxErrorsPerHour
	}
	if config.Settings.MaxErrorsPerDay > 0 {
		defaults.MaxErrorsPerDay = config.Settings.MaxErrorsPerDay
	}
	if config.Settings.MaxConsecutiveErrors > 0 {
		defaults.MaxConsecutiveErrors = config.Settings.MaxConsecutiveErrors
	}

	// ✅ WEEK 2: Rate Limiting & Quotas configurations
	if config.Settings.MaxCallsPerMinute > 0 {
		defaults.MaxCallsPerMinute = config.Settings.MaxCallsPerMinute
	}
	if config.Settings.MaxCallsPerHour > 0 {
		defaults.MaxCallsPerHour = config.Settings.MaxCallsPerHour
	}
	if config.Settings.MaxCallsPerDay > 0 {
		defaults.MaxCallsPerDay = config.Settings.MaxCallsPerDay
	}
	// BlockOnQuotaExceed: true=BLOCK requests when quota exceeded, false=WARN only
	// Default is true (production-safe), only override if explicitly set in YAML
	if config.Settings.BlockOnQuotaExceed {
		defaults.BlockOnQuotaExceed = true
	}

	// Validate all converted values
	if err := defaults.Validate(); err != nil {
		// ✅ Phase 5.1: In STRICT MODE, validation errors are FATAL - no fallback
		if defaults.Mode == StrictMode {
			log.Printf("[CONFIG ERROR] STRICT MODE validation failed: %v", err)
			// Return nil will be caught by caller
			return nil
		}
		// In PERMISSIVE MODE, fallback to defaults
		log.Printf("[CONFIG WARNING] Failed to validate defaults after conversion: %v - using fallback defaults", err)
		return DefaultHardcodedDefaults()
	}

	return defaults
}
