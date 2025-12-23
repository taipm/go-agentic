package crewai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// =============================================================================
// TOOL EXECUTION LOGIC
// Tool execution, timeout management, and result formatting
// Extracted from crew.go to reduce file size
// =============================================================================

// ToolResult represents the result of executing a tool
type ToolResult struct {
	ToolName string
	Status   string
	Output   string
}

// calculateToolTimeout determines the appropriate timeout for the next tool
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) calculateToolTimeout(tracker *TimeoutTracker, toolName string) time.Duration {
	if tracker != nil && ce.ToolTimeouts != nil {
		perToolTimeout := ce.ToolTimeouts.GetToolTimeout(toolName)
		return tracker.CalculateToolTimeout(ce.ToolTimeouts.DefaultToolTimeout, perToolTimeout)
	}
	if ce.ToolTimeouts != nil {
		return ce.ToolTimeouts.GetToolTimeout(toolName)
	}
	return 5 * time.Second
}

// logToolStart logs tool execution start with timeout details
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) logToolStart(tool *Tool, agent *Agent, timeout time.Duration, tracker *TimeoutTracker) {
	if tracker != nil {
		remaining := tracker.GetRemainingTime()
		log.Printf("[TOOL START] %s <- %s (timeout: %v, remaining: %v)", tool.Name, agent.ID, timeout, remaining)
	} else {
		log.Printf("[TOOL START] %s <- %s (timeout: %v)", tool.Name, agent.ID, timeout)
	}
}

// recordToolMetrics records execution metrics and detects timeouts
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) recordToolMetrics(tool *Tool, duration time.Duration, err error, startTime, endTime time.Time) {
	timedOut := err != nil && errors.Is(err, context.DeadlineExceeded)
	if ce.ToolTimeouts != nil && ce.ToolTimeouts.CollectMetrics {
		status := ce.getToolExecutionStatus(timedOut, err)
		ce.ToolTimeouts.ExecutionMetrics = append(ce.ToolTimeouts.ExecutionMetrics, ExecutionMetrics{
			ToolName:  tool.Name,
			Duration:  duration,
			Status:    status,
			TimedOut:  timedOut,
			StartTime: startTime,
			EndTime:   endTime,
		})
	}
	if ce.Metrics != nil {
		success := err == nil
		ce.Metrics.RecordToolExecution(tool.Name, duration, success)
	}
}

// getToolExecutionStatus determines the status based on error type
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) getToolExecutionStatus(timedOut bool, err error) string {
	if timedOut {
		return "timeout"
	}
	if err != nil {
		return "error"
	}
	return "success"
}

// handleToolNotFound logs and records tool not found error
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) handleToolNotFound(call ToolCall) ToolResult {
	log.Printf("[TOOL ERROR] %s - Tool not found", call.ToolName)
	return ToolResult{
		ToolName: call.ToolName,
		Status:   "error",
		Output:   fmt.Sprintf("Tool %s not found", call.ToolName),
	}
}

// handleSequenceTimeout logs and returns timeout result
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) handleSequenceTimeout(tool *Tool, agent *Agent) ToolResult {
	log.Printf("[TOOL TIMEOUT] %s <- %s - Sequence timeout exceeded", tool.Name, agent.ID)
	return ToolResult{
		ToolName: tool.Name,
		Status:   "error",
		Output:   "Tool execution timeout: sequence timeout exceeded",
	}
}

// handleToolExecutionError logs and returns tool execution error
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) handleToolExecutionError(tool *Tool, err error, duration time.Duration, timedOut bool) ToolResult {
	if timedOut {
		log.Printf("[TOOL TIMEOUT] %s - %v (%v)", tool.Name, err, duration)
	} else {
		log.Printf("[TOOL ERROR] %s - %v", tool.Name, err)
	}
	return ToolResult{
		ToolName: tool.Name,
		Status:   "error",
		Output:   err.Error(),
	}
}

// handleToolExecutionSuccess logs and returns tool execution success
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) handleToolExecutionSuccess(tool *Tool, duration time.Duration, output string) ToolResult {
	log.Printf("[TOOL SUCCESS] %s -> %d chars (%v)", tool.Name, len(output), duration)
	return ToolResult{
		ToolName: tool.Name,
		Status:   "success",
		Output:   output,
	}
}

// setupSequenceContext creates and configures the sequence execution context
// ✅ FIX #4: Helper function to reduce cognitive complexity
func (ce *CrewExecutor) setupSequenceContext(ctx context.Context) (context.Context, context.CancelFunc, *TimeoutTracker) {
	if ctx == nil {
		ctx = context.Background()
	}

	var sequenceCtx context.Context
	var sequenceCancel context.CancelFunc
	var timeoutTracker *TimeoutTracker

	if ce.ToolTimeouts != nil && ce.ToolTimeouts.SequenceTimeout > 0 {
		timeoutTracker = NewTimeoutTracker(ce.ToolTimeouts.SequenceTimeout, ce.ToolTimeouts.OverheadBudget)
		sequenceCtx, sequenceCancel = context.WithTimeout(ctx, ce.ToolTimeouts.SequenceTimeout)
	} else {
		// No sequence timeout configured, use context as-is
		sequenceCtx = ctx
		sequenceCancel = func() {} // no-op cancel for defer safety
	}

	return sequenceCtx, sequenceCancel, timeoutTracker
}

// executeCalls executes tool calls from an agent with per-tool and sequence timeouts
// ✅ FIX for Issue #11 (Sequential Tool Timeout): Add timeout protection for hanging tools
// ✅ FIX #4: Enhanced timeout management with remaining time calculation and overhead tracking
func (ce *CrewExecutor) executeCalls(ctx context.Context, calls []ToolCall, agent *Agent) []ToolResult {
	var results []ToolResult

	toolMap := make(map[string]*Tool)
	for _, tool := range agent.Tools {
		toolMap[tool.Name] = tool
	}

	sequenceCtx, sequenceCancel, timeoutTracker := ce.setupSequenceContext(ctx)
	defer sequenceCancel()

	// Reset metrics for this execution
	if ce.ToolTimeouts != nil && ce.ToolTimeouts.CollectMetrics {
		ce.ToolTimeouts.ExecutionMetrics = []ExecutionMetrics{}
	}

	for _, call := range calls {
		tool, ok := toolMap[call.ToolName]
		if !ok {
			results = append(results, ce.handleToolNotFound(call))
			continue
		}

		// ✅ FIX for Issue #11: Check sequence deadline before executing tool
		select {
		case <-sequenceCtx.Done():
			results = append(results, ce.handleSequenceTimeout(tool, agent))
			return results
		default:
		}

		// ✅ FIX #4: Calculate timeout with remaining sequence time
		toolTimeout := ce.calculateToolTimeout(timeoutTracker, tool.Name)
		ce.logToolStart(tool, agent, toolTimeout, timeoutTracker)

		toolCtx, toolCancel := context.WithTimeout(sequenceCtx, toolTimeout)
		startTime := time.Now()
		output, err := safeExecuteTool(toolCtx, tool, call.Arguments)
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		toolCancel()

		// ✅ FIX #4: Record tool execution in timeout tracker
		if timeoutTracker != nil {
			timeoutTracker.RecordToolExecution(duration)
		}

		// ✅ FIX #4: Record metrics
		ce.recordToolMetrics(tool, duration, err, startTime, endTime)

		// ✅ FIX #4: Check if approaching timeout
		if timeoutTracker != nil && timeoutTracker.IsTimeoutWarning() {
			remaining := timeoutTracker.GetRemainingTime()
			log.Printf("[TIMEOUT WARNING] Sequence timeout approaching - only %v remaining", remaining)
		}

		// Handle tool execution result
		timedOut := err != nil && errors.Is(err, context.DeadlineExceeded)
		if err != nil {
			results = append(results, ce.handleToolExecutionError(tool, err, duration, timedOut))
		} else {
			results = append(results, ce.handleToolExecutionSuccess(tool, duration, output))
		}

		// Note: Not hiding embedding vectors anymore - agents need to see vectors to extract and use them
		// Verbose output is handled by the caller, not here
	}

	return results
}

// formatToolResults formats tool results for agent feedback
// ✅ DEPRECATED: Use ce.formatToolResults() method instead (for configurability)
func formatToolResults(results []ToolResult) string {
	return defaultFormatToolResults(results, 2000, 4000) // Default: 2000 per-tool, 4000 total
}

// formatToolResults is a method on CrewExecutor to use configurable MaxToolOutputChars
// ✅ FIX #5: Uses crew.MaxToolOutputChars field instead of hardcoded constant
// ✅ FIX for MEDIUM Issue: Now also uses MaxTotalToolOutputChars to prevent unbounded token usage
func (ce *CrewExecutor) formatToolResults(results []ToolResult) string {
	return defaultFormatToolResults(results, ce.crew.MaxToolOutputChars, ce.crew.MaxTotalToolOutputChars)
}

// defaultFormatToolResults is the actual implementation shared by both functions
// ✅ FIX for MEDIUM Issue: Added maxTotalChars parameter to limit total output across all tools
// This prevents unbounded token usage when multiple tools return large outputs
func defaultFormatToolResults(results []ToolResult, maxOutputChars int, maxTotalChars int) string {
	// Default to 2000 per-tool if not configured
	if maxOutputChars <= 0 {
		maxOutputChars = 2000
	}
	// Default to 4000 total if not configured (~1K tokens)
	if maxTotalChars <= 0 {
		maxTotalChars = 4000
	}

	var sb strings.Builder
	totalChars := 0
	summarizedCount := 0

	sb.WriteString("\n[📊 TOOL EXECUTION RESULTS]\n\n")

	for i, result := range results {
		// ✅ Check if we've exceeded total limit - summarize remaining tools
		if totalChars >= maxTotalChars {
			summarizedCount = len(results) - i
			break
		}

		sb.WriteString(fmt.Sprintf("%s:\n", result.ToolName))
		sb.WriteString(fmt.Sprintf("  Status: %s\n", result.Status))

		output := result.Output

		// ✅ Calculate how much space we have left for this tool
		remainingTotal := maxTotalChars - totalChars
		effectiveMax := maxOutputChars
		if remainingTotal < effectiveMax {
			effectiveMax = remainingTotal
		}

		// ✅ Truncate per-tool output if needed
		if len(output) > effectiveMax {
			output = output[:effectiveMax] + fmt.Sprintf("\n\n[⚠️ OUTPUT TRUNCATED - Original size: %d characters]", len(result.Output))
		}

		outputLine := fmt.Sprintf("  Output: %s\n\n", output)
		sb.WriteString(outputLine)
		totalChars += len(outputLine)
	}

	// ✅ Add summary for tools that were skipped due to total limit
	if summarizedCount > 0 {
		sb.WriteString(fmt.Sprintf("[⚠️ %d additional tool(s) summarized to save context - total output limit: %d chars]\n",
			summarizedCount, maxTotalChars))

		// Add brief status-only summary for skipped tools
		skippedStart := len(results) - summarizedCount
		for _, result := range results[skippedStart:] {
			sb.WriteString(fmt.Sprintf("  • %s: %s (output omitted)\n", result.ToolName, result.Status))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("[END RESULTS]\n")

	return sb.String()
}
