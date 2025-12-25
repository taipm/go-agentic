// Package workflow provides workflow orchestration and execution functionality.
// Routing functions have been moved to core/routing package.
// These re-exports maintain backward compatibility.
package workflow

import (
	"github.com/taipm/go-agentic/core/routing"
)

// DetermineNextAgent - re-export from routing package
var DetermineNextAgent = routing.DetermineNextAgent

// DetermineNextAgentWithSignals - re-export from routing package
var DetermineNextAgentWithSignals = routing.DetermineNextAgentWithSignals

// RouteBySignal - re-export from routing package
var RouteBySignal = routing.RouteBySignal

// RouteByBehavior - re-export from routing package
var RouteByBehavior = routing.RouteByBehavior

// ValidateRouting - re-export from routing package
var ValidateRouting = routing.ValidateRouting
