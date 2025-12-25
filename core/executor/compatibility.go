// Package executor provides the main execution orchestrator for crews.
// This file provides backward compatibility re-exports from new modules.
package executor

import (
	"github.com/taipm/go-agentic/core/execution"
	"github.com/taipm/go-agentic/core/routing"
	statemanagement "github.com/taipm/go-agentic/core/state-management"
)

// ExecutionFlow - re-export from execution module
type ExecutionFlow = execution.ExecutionFlow

// NewExecutionFlow - re-export from execution module
var NewExecutionFlow = execution.NewExecutionFlow

// ExecutionState - re-export from state-management module
type ExecutionState = statemanagement.ExecutionState

// NewExecutionState - re-export from state-management module
var NewExecutionState = statemanagement.NewExecutionState

// RoundMetric - re-export from state-management module
type RoundMetric = statemanagement.RoundMetric

// ExecutionMetrics - re-export from state-management module
type ExecutionMetrics = statemanagement.ExecutionMetrics

// HistoryManager - re-export from state-management module
type HistoryManager = statemanagement.HistoryManager

// NewHistoryManager - re-export from state-management module
var NewHistoryManager = statemanagement.NewHistoryManager

// NewHistoryManagerWithConfig - re-export from state-management module
var NewHistoryManagerWithConfig = statemanagement.NewHistoryManagerWithConfig

// Routing exports
// DetermineNextAgent - re-export from routing module
var DetermineNextAgent = routing.DetermineNextAgent

// DetermineNextAgentWithSignals - re-export from routing module
var DetermineNextAgentWithSignals = routing.DetermineNextAgentWithSignals

// RouteBySignal - re-export from routing module
var RouteBySignal = routing.RouteBySignal

// RouteByBehavior - re-export from routing module
var RouteByBehavior = routing.RouteByBehavior

// ValidateRouting - re-export from routing module
var ValidateRouting = routing.ValidateRouting
