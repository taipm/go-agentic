package common

import "time"

// Token Calculation Constants
const TokenBaseValue = 4
const TokenPaddingValue = 3
const TokenDivisor = 4
const MinHistoryLength = 2
const PercentDivisor = 100.0

// Message Role Constants
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

// Event Type Constants
const (
	EventTypeError      = "error"
	EventTypeToolResult = "tool_result"
)

// Timing & Retry Constants
const BaseRetryDelay = 100 * time.Millisecond
const MinTimeoutValue = 100 * time.Millisecond
const WarnThresholdRatio = 5
