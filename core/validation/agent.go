package validation

import (
	"fmt"

	"github.com/taipm/go-agentic/core/common"
)

// ValidateAgentConfig validates an agent configuration
func ValidateAgentConfig(cfg *common.AgentConfig) error {
	if cfg == nil {
		return &common.ValidationError{
			Field:   "agent",
			Message: "agent config is nil",
		}
	}

	// Basic validation
	if cfg.ID == "" {
		return &common.ValidationError{
			Field:   "id",
			Message: "agent id is required",
		}
	}

	if cfg.Name == "" {
		return &common.ValidationError{
			Field:   "name",
			Message: "agent name is required",
		}
	}

	if cfg.Role == "" {
		return &common.ValidationError{
			Field:   "role",
			Message: "agent role is required",
		}
	}

	if cfg.Primary == nil {
		return &common.ValidationError{
			Field:   "primary",
			Message: "primary model configuration is required",
		}
	}

	if cfg.Primary.Model == "" {
		return &common.ValidationError{
			Field:   "primary.model",
			Message: "primary model name is required",
		}
	}

	if cfg.Primary.Provider == "" {
		return &common.ValidationError{
			Field:   "primary.provider",
			Message: "primary provider is required",
		}
	}

	return nil
}

// ValidateAgentBasicConstraints validates basic agent constraints
func ValidateAgentBasicConstraints(cfg *common.AgentConfig) error {
	return nil
}

// ValidateAgentTemperature validates agent temperature setting
func ValidateAgentTemperature(temp float64) error {
	if temp < 0 || temp > 2 {
		return &common.ValidationError{
			Field:   "temperature",
			Message: fmt.Sprintf("temperature must be between 0 and 2, got %f", temp),
		}
	}
	return nil
}

// ValidateAgentModel validates agent model configuration
func ValidateAgentModel(model string) error {
	if model == "" {
		return &common.ValidationError{
			Field:   "model",
			Message: "model name cannot be empty",
		}
	}
	return nil
}

// ValidateAgentProvider validates agent provider
func ValidateAgentProvider(provider string) error {
	if provider == "" {
		return &common.ValidationError{
			Field:   "provider",
			Message: "provider name cannot be empty",
		}
	}
	return nil
}
