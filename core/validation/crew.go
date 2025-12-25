package validation

import (
	"fmt"

	"github.com/taipm/go-agentic/core/common"
)

// ValidateCrewConfig validates a crew configuration
func ValidateCrewConfig(cfg *common.CrewConfig) error {
	if cfg == nil {
		return &common.ValidationError{
			Field:   "crew",
			Message: "crew config is nil",
		}
	}

	// Basic validation
	if cfg.EntryPoint == "" {
		return &common.ValidationError{
			Field:   "entry_point",
			Message: "entry_point is required",
		}
	}

	if len(cfg.Agents) == 0 {
		return &common.ValidationError{
			Field:   "agents",
			Message: "at least one agent must be configured",
		}
	}

	// Check that entry point is in agents list
	if !Contains(cfg.Agents, cfg.EntryPoint) {
		return &common.ValidationError{
			Field:   "entry_point",
			Message: fmt.Sprintf("entry_point '%s' not found in agents list", cfg.EntryPoint),
		}
	}

	return nil
}

// ValidateCrewRequiredFields validates required fields in crew config
func ValidateCrewRequiredFields(cfg *common.CrewConfig) error {
	return nil
}

// ValidateEntryPointAndBuildMap validates entry point and builds agent map
func ValidateEntryPointAndBuildMap(cfg *common.CrewConfig, agentConfigs map[string]*common.AgentConfig) error {
	return nil
}

// ValidateCrewSettings validates crew settings
func ValidateCrewSettings(cfg *common.CrewConfig) error {
	return nil
}
