package crewai

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ✅ FIX for Issue #14 (Metrics & Observability)
// This file implements comprehensive metrics collection for production monitoring

// ExtendedExecutionMetrics extends the ExecutionMetrics from Issue #11 with more details
type ExtendedExecutionMetrics struct {
	ToolName   string        // Name of the tool executed
	Duration   time.Duration // Time taken to execute
	Status     string        // "success", "timeout", "error"
	TimedOut   bool          // True if tool execution exceeded timeout
	Success    bool          // True if execution succeeded
	Error      string        // Error message if execution failed
	StartTime  time.Time     // When tool execution started
	EndTime    time.Time     // When tool execution completed
}

// ToolMetrics tracks per-tool statistics
type ToolMetrics struct {
	ToolName        string        // Name of the tool
	ExecutionCount  int64         // Total executions
	SuccessCount    int64         // Successful executions
	ErrorCount      int64         // Failed executions
	TotalDuration   time.Duration // Total time spent
	AverageDuration time.Duration // Average per execution
	MinDuration     time.Duration // Fastest execution
	MaxDuration     time.Duration // Slowest execution
}

// AgentMetrics tracks per-agent statistics
type AgentMetrics struct {
	AgentID         string
	AgentName       string
	ExecutionCount  int64
	SuccessCount    int64
	ErrorCount      int64
	TimeoutCount    int64
	TotalDuration   time.Duration
	AverageDuration time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	ToolMetrics     map[string]*ToolMetrics // Per-tool stats within this agent
}

// SystemMetrics aggregates all metrics across the entire system
type SystemMetrics struct {
	StartTime           time.Time
	LastUpdated         time.Time
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	TotalExecutionTime  time.Duration
	AverageRequestTime  time.Duration
	MemoryUsage         uint64 // Current memory in bytes
	MaxMemoryUsage      uint64 // Peak memory in bytes
	CacheHits           int64
	CacheMisses         int64
	CacheHitRate        float64
	AgentMetrics        map[string]*AgentMetrics // Per-agent stats

	// ✅ Crew-level LLM Cost Tracking
	TotalTokens         int     // Total tokens used across all agents
	TotalCost           float64 // Total cost in USD across all agents
	SessionTokens       int     // Tokens in current session (reset on ClearHistory)
	SessionCost         float64 // Cost in current session
	LLMCallCount        int     // Total LLM API calls made
}

// MetricsCollector is the main component for collecting and aggregating metrics
// Thread-safe for concurrent access
type MetricsCollector struct {
	mu            sync.RWMutex
	systemMetrics *SystemMetrics
	enabled       bool

	// Tracking for ongoing execution
	currentExecution *executionTracker
}

// executionTracker tracks an ongoing execution
type executionTracker struct {
	agentID       string
	agentName     string
	startTime     time.Time
	success       bool
	error         string
	execMetrics   []ExtendedExecutionMetrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		systemMetrics: &SystemMetrics{
			StartTime:    time.Now(),
			AgentMetrics: make(map[string]*AgentMetrics),
		},
		enabled: true,
	}
}

// RecordToolExecution records execution of a single tool
func (mc *MetricsCollector) RecordToolExecution(toolName string, duration time.Duration, success bool) {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.systemMetrics.LastUpdated = time.Now()

	// Update current execution metrics
	if mc.currentExecution != nil {
		metric := ExtendedExecutionMetrics{
			ToolName:  toolName,
			Duration:  duration,
			Status:    statusString(success),
			Success:   success,
			StartTime: time.Now().Add(-duration),
			EndTime:   time.Now(),
		}
		mc.currentExecution.execMetrics = append(mc.currentExecution.execMetrics, metric)
	}

	// Update tool metrics within agent
	if mc.currentExecution != nil && mc.currentExecution.agentID != "" {
		agent, exists := mc.systemMetrics.AgentMetrics[mc.currentExecution.agentID]
		if !exists {
			agent = &AgentMetrics{
				AgentID:     mc.currentExecution.agentID,
				AgentName:   mc.currentExecution.agentName,
				ToolMetrics: make(map[string]*ToolMetrics),
			}
			mc.systemMetrics.AgentMetrics[mc.currentExecution.agentID] = agent
		}

		// Update tool metrics
		toolMetric, exists := agent.ToolMetrics[toolName]
		if !exists {
			toolMetric = &ToolMetrics{
				ToolName:    toolName,
				MinDuration: duration,
				MaxDuration: duration,
			}
			agent.ToolMetrics[toolName] = toolMetric
		}

		toolMetric.ExecutionCount++
		toolMetric.TotalDuration += duration

		// Update min/max
		if duration < toolMetric.MinDuration {
			toolMetric.MinDuration = duration
		}
		if duration > toolMetric.MaxDuration {
			toolMetric.MaxDuration = duration
		}

		// Update average
		if toolMetric.ExecutionCount > 0 {
			toolMetric.AverageDuration = toolMetric.TotalDuration / time.Duration(toolMetric.ExecutionCount)
		}

		// Update success/error
		if success {
			toolMetric.SuccessCount++
		} else {
			toolMetric.ErrorCount++
		}
	}
}

// RecordAgentExecution records execution of an entire agent
func (mc *MetricsCollector) RecordAgentExecution(agentID, agentName string, duration time.Duration, success bool) {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.systemMetrics.LastUpdated = time.Now()
	mc.systemMetrics.TotalRequests++

	if success {
		mc.systemMetrics.SuccessfulRequests++
	} else {
		mc.systemMetrics.FailedRequests++
	}

	mc.systemMetrics.TotalExecutionTime += duration

	// Update average
	if mc.systemMetrics.TotalRequests > 0 {
		mc.systemMetrics.AverageRequestTime = mc.systemMetrics.TotalExecutionTime / time.Duration(mc.systemMetrics.TotalRequests)
	}

	// Update agent metrics
	agent, exists := mc.systemMetrics.AgentMetrics[agentID]
	if !exists {
		agent = &AgentMetrics{
			AgentID:     agentID,
			AgentName:   agentName,
			MinDuration: duration,
			MaxDuration: duration,
			ToolMetrics: make(map[string]*ToolMetrics),
		}
		mc.systemMetrics.AgentMetrics[agentID] = agent
	}

	agent.ExecutionCount++
	agent.TotalDuration += duration

	// Update min/max
	if duration < agent.MinDuration {
		agent.MinDuration = duration
	}
	if duration > agent.MaxDuration {
		agent.MaxDuration = duration
	}

	// Update average
	if agent.ExecutionCount > 0 {
		agent.AverageDuration = agent.TotalDuration / time.Duration(agent.ExecutionCount)
	}

	// Update success/error/timeout
	if success {
		agent.SuccessCount++
	} else {
		agent.ErrorCount++
	}
}

// RecordLLMCall records an LLM API call with token usage and cost
// ✅ Crew-level cost tracking - called after each ExecuteAgent
func (mc *MetricsCollector) RecordLLMCall(agentID string, tokens int, cost float64) {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.systemMetrics.LastUpdated = time.Now()
	mc.systemMetrics.TotalTokens += tokens
	mc.systemMetrics.TotalCost += cost
	mc.systemMetrics.SessionTokens += tokens
	mc.systemMetrics.SessionCost += cost
	mc.systemMetrics.LLMCallCount++
}

// ResetSessionCost resets session-level cost tracking (called on ClearHistory)
func (mc *MetricsCollector) ResetSessionCost() {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.systemMetrics.SessionTokens = 0
	mc.systemMetrics.SessionCost = 0
}

// GetSessionCost returns current session cost metrics
func (mc *MetricsCollector) GetSessionCost() (tokens int, cost float64) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.systemMetrics.SessionTokens, mc.systemMetrics.SessionCost
}

// GetTotalCost returns total cost metrics across all sessions
func (mc *MetricsCollector) GetTotalCost() (tokens int, cost float64, calls int) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.systemMetrics.TotalTokens, mc.systemMetrics.TotalCost, mc.systemMetrics.LLMCallCount
}

// LogCrewCostSummary logs the current crew cost summary
func (mc *MetricsCollector) LogCrewCostSummary() {
	if !mc.enabled {
		return
	}

	mc.mu.RLock()
	defer mc.mu.RUnlock()

	fmt.Printf("[CREW COST] Session: %d tokens ($%.6f) | Total: %d tokens ($%.6f) | LLM Calls: %d\n",
		mc.systemMetrics.SessionTokens,
		mc.systemMetrics.SessionCost,
		mc.systemMetrics.TotalTokens,
		mc.systemMetrics.TotalCost,
		mc.systemMetrics.LLMCallCount)
}

// RecordCacheHit records a cache hit
func (mc *MetricsCollector) RecordCacheHit() {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.systemMetrics.CacheHits++
	mc.updateCacheHitRate()
}

// RecordCacheMiss records a cache miss
func (mc *MetricsCollector) RecordCacheMiss() {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.systemMetrics.CacheMisses++
	mc.updateCacheHitRate()
}

// updateCacheHitRate calculates cache hit rate (must be called with lock held)
func (mc *MetricsCollector) updateCacheHitRate() {
	total := mc.systemMetrics.CacheHits + mc.systemMetrics.CacheMisses
	if total > 0 {
		mc.systemMetrics.CacheHitRate = float64(mc.systemMetrics.CacheHits) / float64(total)
	}
}

// UpdateMemoryUsage updates current memory usage
func (mc *MetricsCollector) UpdateMemoryUsage(current uint64) {
	if !mc.enabled {
		return
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.systemMetrics.MemoryUsage = current
	if current > mc.systemMetrics.MaxMemoryUsage {
		mc.systemMetrics.MaxMemoryUsage = current
	}
}

// GetSystemMetrics returns a copy of current system metrics
func (mc *MetricsCollector) GetSystemMetrics() *SystemMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Return a copy to prevent external modifications
	metrics := *mc.systemMetrics
	return &metrics
}

// ExportMetrics exports metrics in specified format (json or prometheus)
func (mc *MetricsCollector) ExportMetrics(format string) (string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	switch format {
	case "json":
		return mc.exportJSON()
	case "prometheus":
		return mc.exportPrometheus()
	default:
		return "", fmt.Errorf("unsupported export format: %s (supported: json, prometheus)", format)
	}
}

// exportJSON exports metrics in JSON format
func (mc *MetricsCollector) exportJSON() (string, error) {
	data := map[string]interface{}{
		"system_metrics": mc.systemMetrics,
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metrics: %w", err)
	}

	return string(jsonBytes), nil
}

// exportPrometheus exports metrics in Prometheus format
func (mc *MetricsCollector) exportPrometheus() (string, error) {
	var result string

	// System metrics
	result += fmt.Sprintf("# HELP crew_requests_total Total requests processed\n")
	result += fmt.Sprintf("# TYPE crew_requests_total counter\n")
	result += fmt.Sprintf("crew_requests_total{status=\"success\"} %d\n", mc.systemMetrics.SuccessfulRequests)
	result += fmt.Sprintf("crew_requests_total{status=\"error\"} %d\n", mc.systemMetrics.FailedRequests)
	result += fmt.Sprintf("\n")

	// Average request time
	result += fmt.Sprintf("# HELP crew_average_request_duration_seconds Average request duration\n")
	result += fmt.Sprintf("# TYPE crew_average_request_duration_seconds gauge\n")
	result += fmt.Sprintf("crew_average_request_duration_seconds %f\n", mc.systemMetrics.AverageRequestTime.Seconds())
	result += fmt.Sprintf("\n")

	// Cache metrics
	result += fmt.Sprintf("# HELP crew_cache_hits_total Total cache hits\n")
	result += fmt.Sprintf("# TYPE crew_cache_hits_total counter\n")
	result += fmt.Sprintf("crew_cache_hits_total %d\n", mc.systemMetrics.CacheHits)
	result += fmt.Sprintf("# HELP crew_cache_misses_total Total cache misses\n")
	result += fmt.Sprintf("# TYPE crew_cache_misses_total counter\n")
	result += fmt.Sprintf("crew_cache_misses_total %d\n", mc.systemMetrics.CacheMisses)
	result += fmt.Sprintf("# HELP crew_cache_hit_rate Cache hit rate\n")
	result += fmt.Sprintf("# TYPE crew_cache_hit_rate gauge\n")
	result += fmt.Sprintf("crew_cache_hit_rate %f\n", mc.systemMetrics.CacheHitRate)
	result += fmt.Sprintf("\n")

	// Memory metrics
	result += fmt.Sprintf("# HELP crew_memory_usage_bytes Current memory usage\n")
	result += fmt.Sprintf("# TYPE crew_memory_usage_bytes gauge\n")
	result += fmt.Sprintf("crew_memory_usage_bytes %d\n", mc.systemMetrics.MemoryUsage)
	result += fmt.Sprintf("# HELP crew_max_memory_usage_bytes Maximum memory usage\n")
	result += fmt.Sprintf("# TYPE crew_max_memory_usage_bytes gauge\n")
	result += fmt.Sprintf("crew_max_memory_usage_bytes %d\n", mc.systemMetrics.MaxMemoryUsage)
	result += fmt.Sprintf("\n")

	// Agent metrics
	for agentID, agent := range mc.systemMetrics.AgentMetrics {
		result += fmt.Sprintf("# Agent %s (%s)\n", agentID, agent.AgentName)
		result += fmt.Sprintf("crew_agent_executions{agent=\"%s\"} %d\n", agentID, agent.ExecutionCount)
		result += fmt.Sprintf("crew_agent_successes{agent=\"%s\"} %d\n", agentID, agent.SuccessCount)
		result += fmt.Sprintf("crew_agent_errors{agent=\"%s\"} %d\n", agentID, agent.ErrorCount)
		result += fmt.Sprintf("crew_agent_average_duration{agent=\"%s\"} %f\n", agentID, agent.AverageDuration.Seconds())
		result += fmt.Sprintf("\n")
	}

	return result, nil
}

// Reset resets all metrics (useful for testing)
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.systemMetrics = &SystemMetrics{
		StartTime:    time.Now(),
		AgentMetrics: make(map[string]*AgentMetrics),
	}
}

// Enable enables metrics collection
func (mc *MetricsCollector) Enable() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.enabled = true
}

// Disable disables metrics collection
func (mc *MetricsCollector) Disable() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.enabled = false
}

// IsEnabled returns whether metrics collection is enabled
func (mc *MetricsCollector) IsEnabled() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.enabled
}

// Helper function to convert boolean to status string
func statusString(success bool) string {
	if success {
		return "success"
	}
	return "error"
}
