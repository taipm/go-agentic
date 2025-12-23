package crewai

import (
	"fmt"
	"time"
)

// LogMetadataInfo logs comprehensive metadata information for an agent
// ✅ WEEK 2: Metadata logging for visibility into quotas and metrics
func LogMetadataInfo(agent *Agent, label string) {
	if agent == nil || agent.Metadata == nil {
		return
	}

	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	fmt.Printf("\n[METADATA] Agent '%s' %s\n", agent.Name, label)

	// Cost Quotas
	fmt.Printf("  💰 Cost Quotas: %d tokens/call, %d tokens/day, $%.2f/day (%.0f%% alert)\n",
		agent.Metadata.Quotas.MaxTokensPerCall,
		agent.Metadata.Quotas.MaxTokensPerDay,
		agent.Metadata.Quotas.MaxCostPerDay,
		agent.Metadata.Quotas.CostAlertPercent*100)

	// Memory Quotas
	fmt.Printf("  🧠 Memory Quotas: %d MB/call, %d GB/day, %d token context (%.0f%% alert)\n",
		agent.Metadata.Quotas.MaxMemoryPerCall,
		agent.Metadata.Quotas.MaxMemoryPerDay/1024,
		agent.Metadata.Quotas.MaxContextWindow,
		agent.Metadata.Memory.MemoryAlertPercent*100)

	// Execution Quotas
	fmt.Printf("  ⚙️  Execution: %d/min, %d/hour, %d/day calls | %d errors/hour, %d errors/day\n",
		agent.Metadata.Quotas.MaxCallsPerMinute,
		agent.Metadata.Quotas.MaxCallsPerHour,
		agent.Metadata.Quotas.MaxCallsPerDay,
		agent.Metadata.Quotas.MaxErrorsPerHour,
		agent.Metadata.Quotas.MaxErrorsPerDay)
}

// LogMetadataMetrics logs current metric values
// ✅ WEEK 2: Real-time metric display
func LogMetadataMetrics(agent *Agent) {
	if agent == nil || agent.Metadata == nil {
		return
	}

	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	// Cost metrics
	costPercent := 0.0
	if agent.Metadata.Quotas.MaxCostPerDay > 0 {
		costPercent = (agent.Metadata.Cost.DailyCost / agent.Metadata.Quotas.MaxCostPerDay) * 100
	}

	tokenPercent := 0.0
	if agent.Metadata.Quotas.MaxTokensPerDay > 0 {
		tokenPercent = (float64(agent.Metadata.Cost.TotalTokens) / float64(agent.Metadata.Quotas.MaxTokensPerDay)) * 100
	}

	fmt.Printf("[METRICS] Agent '%s': Calls=%d | Cost=$%.4f/%.2f (%.1f%%) | Tokens=%d/%d (%.1f%%)\n",
		agent.Name,
		agent.Metadata.Cost.CallCount,
		agent.Metadata.Cost.DailyCost,
		agent.Metadata.Quotas.MaxCostPerDay,
		costPercent,
		agent.Metadata.Cost.TotalTokens,
		agent.Metadata.Quotas.MaxTokensPerDay,
		tokenPercent)

	// Memory metrics
	if agent.Metadata.Memory.CurrentMemoryMB > 0 {
		memPercent := (float64(agent.Metadata.Memory.CurrentMemoryMB) / float64(agent.Metadata.Quotas.MaxMemoryPerCall)) * 100
		fmt.Printf("[MEMORY] Agent '%s': %d MB (peak: %d MB, avg: %d MB, %.1f%% of limit)\n",
			agent.Name,
			agent.Metadata.Memory.CurrentMemoryMB,
			agent.Metadata.Memory.PeakMemoryMB,
			agent.Metadata.Memory.AverageMemoryMB,
			memPercent)
	}

	// Performance metrics
	if agent.Metadata.Performance.SuccessfulCalls > 0 || agent.Metadata.Performance.FailedCalls > 0 {
		fmt.Printf("[PERFORMANCE] Agent '%s': Success=%.1f%% (%d ok, %d failed) | Errors today=%d/%d\n",
			agent.Name,
			agent.Metadata.Performance.SuccessRate,
			agent.Metadata.Performance.SuccessfulCalls,
			agent.Metadata.Performance.FailedCalls,
			agent.Metadata.Performance.ErrorCountToday,
			agent.Metadata.Performance.MaxErrorsPerDay)
	}
}

// LogMetadataQuotaStatus logs quota usage status with warnings
// ✅ WEEK 2: Alert when approaching limits
func LogMetadataQuotaStatus(agent *Agent) {
	if agent == nil || agent.Metadata == nil {
		return
	}

	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	alerts := []string{}

	// Check cost quota
	if agent.Metadata.Quotas.MaxCostPerDay > 0 {
		costPercent := agent.Metadata.Cost.DailyCost / agent.Metadata.Quotas.MaxCostPerDay
		if costPercent >= agent.Metadata.Quotas.CostAlertPercent {
			alerts = append(alerts, fmt.Sprintf("COST: %.0f%% of daily budget used ($%.4f/$%.2f)",
				costPercent*100, agent.Metadata.Cost.DailyCost, agent.Metadata.Quotas.MaxCostPerDay))
		}
	}

	// Check token quota
	if agent.Metadata.Quotas.MaxTokensPerDay > 0 {
		tokenPercent := float64(agent.Metadata.Cost.TotalTokens) / float64(agent.Metadata.Quotas.MaxTokensPerDay)
		if tokenPercent >= agent.Metadata.Quotas.CostAlertPercent {
			alerts = append(alerts, fmt.Sprintf("TOKENS: %.0f%% of daily limit (%d/%d)",
				tokenPercent*100, agent.Metadata.Cost.TotalTokens, agent.Metadata.Quotas.MaxTokensPerDay))
		}
	}

	// Check memory quota
	if agent.Metadata.Memory.CurrentMemoryMB > 0 && agent.Metadata.Quotas.MaxMemoryPerCall > 0 {
		memPercent := float64(agent.Metadata.Memory.CurrentMemoryMB) / float64(agent.Metadata.Quotas.MaxMemoryPerCall)
		if memPercent >= agent.Metadata.Memory.MemoryAlertPercent {
			alerts = append(alerts, fmt.Sprintf("MEMORY: %.0f%% of per-call limit (%d/%d MB)",
				memPercent*100, agent.Metadata.Memory.CurrentMemoryMB, agent.Metadata.Quotas.MaxMemoryPerCall))
		}
	}

	// Check error quota
	if agent.Metadata.Performance.ErrorCountToday >= agent.Metadata.Performance.MaxErrorsPerDay {
		alerts = append(alerts, fmt.Sprintf("ERRORS: Daily limit reached (%d/%d)",
			agent.Metadata.Performance.ErrorCountToday, agent.Metadata.Performance.MaxErrorsPerDay))
	}

	// Check consecutive errors
	if agent.Metadata.Performance.ConsecutiveErrors >= agent.Metadata.Performance.MaxConsecutiveErrors {
		alerts = append(alerts, fmt.Sprintf("CONSECUTIVE: %d consecutive errors (max: %d)",
			agent.Metadata.Performance.ConsecutiveErrors, agent.Metadata.Performance.MaxConsecutiveErrors))
	}

	// Log alerts if any
	if len(alerts) > 0 {
		fmt.Printf("\n⚠️  [QUOTA ALERT] Agent '%s':\n", agent.Name)
		for _, alert := range alerts {
			fmt.Printf("     • %s\n", alert)
		}
	}
}

// FormatMetadataReport generates a complete metadata report
// ✅ WEEK 2: Comprehensive agent status report
func FormatMetadataReport(agent *Agent) string {
	if agent == nil || agent.Metadata == nil {
		return "Agent metadata not available\n"
	}

	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	const timeFormat = "2006-01-02 15:04:05"

	report := fmt.Sprintf(
		`
╔════════════════════════════════════════════════════════════╗
║              AGENT METADATA REPORT                         ║
║              Agent: %s                        ║
╚════════════════════════════════════════════════════════════╝

📋 IDENTIFICATION
  ID: %s
  Name: %s
  Role: %s
  Created: %s

💰 COST STATUS
  Quota: $%.2f/day | %d tokens/call, %d tokens/day
  Usage: $%.4f spent | %d tokens used | %d calls
  Status: %.1f%% of daily budget

🧠 MEMORY STATUS
  Quota: %d MB/call | %d GB/day | %d token context
  Usage: %d MB current | %d MB peak | %d MB average
  Trend: %+.1f%%
  Slow call threshold: %.1fs

⚙️  EXECUTION STATUS
  Rate Limits: %d calls/min, %d calls/hour, %d calls/day
  Error Limits: %d/hour, %d/day, max %d consecutive
  Enforcement: %v

📊 PERFORMANCE
  Success Rate: %.1f%% (%d successful, %d failed)
  Avg Response: %.2fs
  Errors Today: %d / %d
  Last Error: %s (at %s)

⏱️  TIMING
  Last Access: %s (%.0f seconds ago)
`,
		agent.Name,
		agent.Metadata.AgentID,
		agent.Metadata.AgentName,
		agent.Role,
		agent.Metadata.CreatedTime.Format(timeFormat),
		agent.Metadata.Quotas.MaxCostPerDay,
		agent.Metadata.Quotas.MaxTokensPerCall,
		agent.Metadata.Quotas.MaxTokensPerDay,
		agent.Metadata.Cost.DailyCost,
		agent.Metadata.Cost.TotalTokens,
		agent.Metadata.Cost.CallCount,
		(agent.Metadata.Cost.DailyCost/agent.Metadata.Quotas.MaxCostPerDay)*100,
		agent.Metadata.Quotas.MaxMemoryPerCall,
		agent.Metadata.Quotas.MaxMemoryPerDay/1024,
		agent.Metadata.Quotas.MaxContextWindow,
		agent.Metadata.Memory.CurrentMemoryMB,
		agent.Metadata.Memory.PeakMemoryMB,
		agent.Metadata.Memory.AverageMemoryMB,
		agent.Metadata.Memory.MemoryTrendPercent,
		agent.Metadata.Memory.SlowCallThreshold.Seconds(),
		agent.Metadata.Quotas.MaxCallsPerMinute,
		agent.Metadata.Quotas.MaxCallsPerHour,
		agent.Metadata.Quotas.MaxCallsPerDay,
		agent.Metadata.Quotas.MaxErrorsPerHour,
		agent.Metadata.Quotas.MaxErrorsPerDay,
		agent.Metadata.Performance.MaxConsecutiveErrors,
		agent.Metadata.Quotas.EnforceQuotas,
		agent.Metadata.Performance.SuccessRate,
		agent.Metadata.Performance.SuccessfulCalls,
		agent.Metadata.Performance.FailedCalls,
		agent.Metadata.Performance.AverageResponseTime.Seconds(),
		agent.Metadata.Performance.ErrorCountToday,
		agent.Metadata.Performance.MaxErrorsPerDay,
		agent.Metadata.Performance.LastError,
		agent.Metadata.Performance.LastErrorTime.Format(timeFormat),
		agent.Metadata.LastAccessTime.Format(timeFormat),
		time.Since(agent.Metadata.LastAccessTime).Seconds(),
	)

	return report
}

// crewMetricsAggregator holds aggregated crew metrics
type crewMetricsAggregator struct {
	totalCallCount int64
	totalTokens    int64
	totalCost      float64
	totalMemory    int64
	totalErrors    int64
	successCount   int64
	failureCount   int64
}

// aggregateCrewMetrics collects metrics from all agents in crew
func aggregateCrewMetrics(crew *Crew) *crewMetricsAggregator {
	agg := &crewMetricsAggregator{}

	for _, agent := range crew.Agents {
		if agent == nil || agent.Metadata == nil {
			continue
		}

		agent.Metadata.Mutex.RLock()
		agg.totalCallCount += int64(agent.Metadata.Cost.CallCount)
		agg.totalTokens += int64(agent.Metadata.Cost.TotalTokens)
		agg.totalCost += agent.Metadata.Cost.DailyCost
		agg.totalMemory += int64(agent.Metadata.Memory.CurrentMemoryMB)
		agg.totalErrors += int64(agent.Metadata.Performance.ErrorCountToday)
		agg.successCount += int64(agent.Metadata.Performance.SuccessfulCalls)
		agg.failureCount += int64(agent.Metadata.Performance.FailedCalls)
		agent.Metadata.Mutex.RUnlock()
	}

	return agg
}

// logAgentMetrics logs metrics for a single agent
func logAgentMetrics(agent *Agent) {
	if agent == nil || agent.Metadata == nil {
		return
	}

	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	costPercent := 0.0
	if agent.Metadata.Quotas.MaxCostPerDay > 0 {
		costPercent = (agent.Metadata.Cost.DailyCost / agent.Metadata.Quotas.MaxCostPerDay) * 100
	}

	tokenPercent := 0.0
	if agent.Metadata.Quotas.MaxTokensPerDay > 0 {
		tokenPercent = (float64(agent.Metadata.Cost.TotalTokens) / float64(agent.Metadata.Quotas.MaxTokensPerDay)) * 100
	}

	fmt.Printf("  Agent: %s (%s)\n", agent.Name, agent.ID)
	fmt.Printf("    💰 Cost: $%.4f/%.2f (%.1f%%) | Tokens: %d/%d (%.1f%%)\n",
		agent.Metadata.Cost.DailyCost,
		agent.Metadata.Quotas.MaxCostPerDay,
		costPercent,
		agent.Metadata.Cost.TotalTokens,
		agent.Metadata.Quotas.MaxTokensPerDay,
		tokenPercent)

	if agent.Metadata.Memory.CurrentMemoryMB > 0 {
		fmt.Printf("    🧠 Memory: %d MB (peak: %d MB)\n",
			agent.Metadata.Memory.CurrentMemoryMB,
			agent.Metadata.Memory.PeakMemoryMB)
	}

	if agent.Metadata.Performance.SuccessfulCalls > 0 || agent.Metadata.Performance.FailedCalls > 0 {
		fmt.Printf("    📈 Performance: %.1f%% success (%d ok, %d failed) | Errors: %d\n",
			agent.Metadata.Performance.SuccessRate,
			agent.Metadata.Performance.SuccessfulCalls,
			agent.Metadata.Performance.FailedCalls,
			agent.Metadata.Performance.ErrorCountToday)
	}

	if agent.Metadata.Cost.CallCount == 0 {
		fmt.Printf("    ⏱️  Not executed yet\n")
	} else {
		fmt.Printf("    ⏱️  Calls: %d\n", agent.Metadata.Cost.CallCount)
	}

	fmt.Printf("\n")
}

// LogCrewMetadataReport logs comprehensive metadata for all agents in a crew
// ✅ WEEK 2: Crew-level metadata aggregation and reporting
func LogCrewMetadataReport(crew *Crew) {
	if crew == nil || len(crew.Agents) == 0 {
		return
	}

	fmt.Printf("\n╔════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              CREW METADATA AGGREGATION REPORT              ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("📊 AGENTS METRICS SUMMARY:\n\n")

	// Display per-agent metrics
	for _, agent := range crew.Agents {
		logAgentMetrics(agent)
	}

	// Aggregate and display totals
	agg := aggregateCrewMetrics(crew)
	fmt.Printf("📈 CREW AGGREGATED TOTALS:\n")
	fmt.Printf("  Total Calls: %d\n", agg.totalCallCount)
	fmt.Printf("  Total Tokens: %d\n", agg.totalTokens)
	fmt.Printf("  Total Cost: $%.4f\n", agg.totalCost)
	if agg.totalMemory > 0 {
		fmt.Printf("  Total Memory: %d MB\n", agg.totalMemory)
	}
	fmt.Printf("  Success Rate: %.1f%% (%d succeeded, %d failed)\n",
		calculateSuccessRate(agg.successCount, agg.failureCount), agg.successCount, agg.failureCount)
	if agg.totalErrors > 0 {
		fmt.Printf("  Total Errors: %d\n", agg.totalErrors)
	}
}

// checkAgentQuotaAlerts collects quota alerts for a single agent
func checkAgentQuotaAlerts(agent *Agent) []string {
	var alerts []string
	if agent == nil || agent.Metadata == nil {
		return alerts
	}

	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	// Check cost quota
	if agent.Metadata.Quotas.MaxCostPerDay > 0 {
		costPercent := agent.Metadata.Cost.DailyCost / agent.Metadata.Quotas.MaxCostPerDay
		if costPercent >= agent.Metadata.Quotas.CostAlertPercent {
			alerts = append(alerts, fmt.Sprintf("%s: COST %.0f%% ($%.4f/$%.2f)",
				agent.Name, costPercent*100,
				agent.Metadata.Cost.DailyCost, agent.Metadata.Quotas.MaxCostPerDay))
		}
	}

	// Check token quota
	if agent.Metadata.Quotas.MaxTokensPerDay > 0 {
		tokenPercent := float64(agent.Metadata.Cost.TotalTokens) / float64(agent.Metadata.Quotas.MaxTokensPerDay)
		if tokenPercent >= agent.Metadata.Quotas.CostAlertPercent {
			alerts = append(alerts, fmt.Sprintf("%s: TOKENS %.0f%% (%d/%d)",
				agent.Name, tokenPercent*100,
				agent.Metadata.Cost.TotalTokens, agent.Metadata.Quotas.MaxTokensPerDay))
		}
	}

	// Check memory quota
	if agent.Metadata.Memory.CurrentMemoryMB > 0 && agent.Metadata.Quotas.MaxMemoryPerCall > 0 {
		memPercent := float64(agent.Metadata.Memory.CurrentMemoryMB) / float64(agent.Metadata.Quotas.MaxMemoryPerCall)
		if memPercent >= agent.Metadata.Memory.MemoryAlertPercent {
			alerts = append(alerts, fmt.Sprintf("%s: MEMORY %.0f%% (%d/%d MB)",
				agent.Name, memPercent*100,
				agent.Metadata.Memory.CurrentMemoryMB, agent.Metadata.Quotas.MaxMemoryPerCall))
		}
	}

	// Check error quota
	if agent.Metadata.Performance.ErrorCountToday >= agent.Metadata.Performance.MaxErrorsPerDay {
		alerts = append(alerts, fmt.Sprintf("%s: ERROR LIMIT reached (%d/%d)",
			agent.Name, agent.Metadata.Performance.ErrorCountToday, agent.Metadata.Performance.MaxErrorsPerDay))
	}

	return alerts
}

// LogCrewQuotaStatus logs quota status for all agents in crew
// ✅ WEEK 2: Crew-level quota alerts
func LogCrewQuotaStatus(crew *Crew) {
	if crew == nil || len(crew.Agents) == 0 {
		return
	}

	var allAlerts []string

	for _, agent := range crew.Agents {
		alerts := checkAgentQuotaAlerts(agent)
		allAlerts = append(allAlerts, alerts...)
	}

	// Log alerts if any
	if len(allAlerts) > 0 {
		fmt.Printf("\n⚠️  [CREW QUOTA ALERTS]:\n")
		for _, alert := range allAlerts {
			fmt.Printf("     • %s\n", alert)
		}
	}
}

// calculateSuccessRate calculates success rate from success and failure counts
func calculateSuccessRate(successCount, failureCount int64) float64 {
	total := successCount + failureCount
	if total == 0 {
		return 100.0
	}
	return (float64(successCount) / float64(total)) * 100
}

// LogMemoryMetrics logs detailed memory usage information
// ✅ WEEK 3: Display memory tracking metrics
func LogMemoryMetrics(agent *Agent) {
	if agent == nil || agent.Metadata == nil {
		return
	}

	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	memPercent := 0.0
	if agent.Metadata.Quotas.MaxMemoryPerCall > 0 {
		memPercent = (float64(agent.Metadata.Memory.CurrentMemoryMB) / float64(agent.Metadata.Quotas.MaxMemoryPerCall)) * 100
	}

	fmt.Printf("[MEMORY] Agent '%s': Current=%d MB (Peak=%d MB, Avg=%d MB) | Usage=%.1f%% | Trend=%+.1f%%\n",
		agent.Name,
		agent.Metadata.Memory.CurrentMemoryMB,
		agent.Metadata.Memory.PeakMemoryMB,
		agent.Metadata.Memory.AverageMemoryMB,
		memPercent,
		agent.Metadata.Memory.MemoryTrendPercent)

	// Check context window usage
	if agent.Metadata.Memory.CurrentContextSize > 0 {
		contextPercent := (float64(agent.Metadata.Memory.CurrentContextSize) / float64(agent.Metadata.Memory.MaxContextWindow)) * 100
		fmt.Printf("[CONTEXT] Agent '%s': %d tokens used / %d max (%.1f%%)\n",
			agent.Name,
			agent.Metadata.Memory.CurrentContextSize,
			agent.Metadata.Memory.MaxContextWindow,
			contextPercent)
	}
}

// LogPerformanceMetrics logs performance and reliability metrics
// ✅ WEEK 3: Display performance tracking metrics
func LogPerformanceMetrics(agent *Agent) {
	if agent == nil || agent.Metadata == nil {
		return
	}

	agent.Metadata.Mutex.RLock()
	defer agent.Metadata.Mutex.RUnlock()

	total := agent.Metadata.Performance.SuccessfulCalls + agent.Metadata.Performance.FailedCalls
	if total == 0 {
		return // No calls yet
	}

	fmt.Printf("[PERFORMANCE] Agent '%s': Success=%.1f%% (%d ok, %d failed) | Avg Response=%.2fs | Errors=%d (consecutive=%d)\n",
		agent.Name,
		agent.Metadata.Performance.SuccessRate,
		agent.Metadata.Performance.SuccessfulCalls,
		agent.Metadata.Performance.FailedCalls,
		agent.Metadata.Performance.AverageResponseTime.Seconds(),
		agent.Metadata.Performance.ErrorCountToday,
		agent.Metadata.Performance.ConsecutiveErrors)

	// Show last error if exists
	if agent.Metadata.Performance.LastError != "" {
		fmt.Printf("[LAST ERROR] Agent '%s': %s (at %s)\n",
			agent.Name,
			agent.Metadata.Performance.LastError,
			agent.Metadata.Performance.LastErrorTime.Format("2006-01-02 15:04:05"))
	}
}

// LogMemoryAndPerformanceStatus logs combined memory + performance status
// ✅ WEEK 3: Comprehensive execution metrics
func LogMemoryAndPerformanceStatus(agent *Agent) {
	LogMemoryMetrics(agent)
	LogPerformanceMetrics(agent)
}
