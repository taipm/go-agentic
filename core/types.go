// Package core provides type aliases and backward compatibility re-exports.
// All core types are defined in the common package to maintain a single source of truth.
// This file serves as a backward compatibility layer for existing code that imports from crewai.
package crewai

import (
	"github.com/taipm/go-agentic/core/common"
	"github.com/taipm/go-agentic/core/executor"
	"github.com/taipm/go-agentic/core/validation"
)

// ============================================================================
// TYPE ALIASES FOR BACKWARD COMPATIBILITY
// These types are defined in the common package but re-exported here
// to maintain backward compatibility with existing code in the root crewai package
// ============================================================================

// Agent - re-export from common package
type Agent = common.Agent

// ModelConfig - re-export from common package
type ModelConfig = common.ModelConfig

// CrewConfig - re-export from common package
type CrewConfig = common.CrewConfig

// AgentConfig - re-export from common package
type AgentConfig = common.AgentConfig

// RoutingConfig - re-export from common package
type RoutingConfig = common.RoutingConfig

// RoutingSignal - re-export from common package
type RoutingSignal = common.RoutingSignal

// AgentBehavior - re-export from common package
type AgentBehavior = common.AgentBehavior

// ParallelGroupConfig - re-export from common package
type ParallelGroupConfig = common.ParallelGroupConfig

// ModelConfigYAML - re-export from common package
type ModelConfigYAML = common.ModelConfigYAML

// CostLimitsConfig - re-export from common package
type CostLimitsConfig = common.CostLimitsConfig

// AgentQuotaLimits - re-export from common package
type AgentQuotaLimits = common.AgentQuotaLimits

// AgentMetadata - re-export from common package
type AgentMetadata = common.AgentMetadata

// MemoryLimitsConfig - re-export from common package
type MemoryLimitsConfig = common.MemoryLimitsConfig

// ErrorLimitsConfig - re-export from common package
type ErrorLimitsConfig = common.ErrorLimitsConfig

// LoggingConfig - re-export from common package
type LoggingConfig = common.LoggingConfig

// AgentCostMetrics - re-export from common package
type AgentCostMetrics = common.AgentCostMetrics

// AgentMemoryMetrics - re-export from common package
type AgentMemoryMetrics = common.AgentMemoryMetrics

// AgentPerformanceMetrics - re-export from common package
type AgentPerformanceMetrics = common.AgentPerformanceMetrics

// StreamEvent - re-export from common package
type StreamEvent = common.StreamEvent

// Tool - re-export from common package
type Tool = common.Tool

// ToolTimeoutConfig - re-export from common package
type ToolTimeoutConfig = common.ToolTimeoutConfig

// Task - re-export from common package
type Task = common.Task

// Message - re-export from common package
type Message = common.Message

// ToolCall - re-export from common package
type ToolCall = common.ToolCall

// AgentResponse - re-export from common package
type AgentResponse = common.AgentResponse

// CrewResponse - re-export from common package
type CrewResponse = common.CrewResponse

// Crew - re-export from common package
type Crew = common.Crew

// HistoryManager - re-export from executor package
type HistoryManager = executor.HistoryManager

// NewHistoryManager creates a new HistoryManager with default settings
func NewHistoryManager() *HistoryManager {
	return executor.NewHistoryManager()
}

// NewToolTimeoutConfig creates a new tool timeout configuration with defaults
// Delegates to common.NewToolTimeoutConfig()
func NewToolTimeoutConfig() *ToolTimeoutConfig {
	return common.NewToolTimeoutConfig()
}

// ============================================================================
// VALIDATION FUNCTION RE-EXPORTS
// ============================================================================

// ValidateCrewConfig validates a crew configuration
// Re-exports from validation package for backward compatibility
func ValidateCrewConfig(cfg *CrewConfig) error {
	return validation.ValidateCrewConfig(cfg)
}

// ValidateAgentConfig validates an agent configuration
// Re-exports from validation package for backward compatibility
func ValidateAgentConfig(cfg *AgentConfig) error {
	return validation.ValidateAgentConfig(cfg)
}

// ============================================================================
// CONSTANTS RE-EXPORTS FROM COMMON PACKAGE
// ============================================================================

// Token Calculation Constants (re-exported from common)
const (
	TokenBaseValue    = common.TokenBaseValue
	TokenPaddingValue = common.TokenPaddingValue
	TokenDivisor      = common.TokenDivisor
)

// Role Constants (re-exported from common)
const (
	RoleUser      = common.RoleUser
	RoleAssistant = common.RoleAssistant
)

// Event Type Constants (re-exported from common)
const (
	EventTypeError      = common.EventTypeError
	EventTypeToolResult = common.EventTypeToolResult
)
