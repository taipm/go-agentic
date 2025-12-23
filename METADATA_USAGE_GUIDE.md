# 📖 Agent Metadata Usage Guide

**Status:** ✅ READY FOR USE
**Date:** Dec 23, 2025
**Version:** WEEK 2

---

## 🎯 Quick Start: Accessing Agent Metadata

### Getting Agent Metadata

After agents are created from configuration, each agent has a unified metadata hub:

```go
// Get agent from crew
agent := crew.Agents[0]

// Access unified metadata
metadata := agent.Metadata
```

---

## 📊 Accessing Quotas

### Read All Quotas

```go
// Cost Quotas
agent.Metadata.Quotas.MaxTokensPerCall   // e.g., 1000
agent.Metadata.Quotas.MaxTokensPerDay    // e.g., 50000
agent.Metadata.Quotas.MaxCostPerDay      // e.g., 10.0
agent.Metadata.Quotas.CostAlertPercent   // e.g., 0.80

// Memory Quotas
agent.Metadata.Quotas.MaxMemoryPerCall   // e.g., 512 MB
agent.Metadata.Quotas.MaxMemoryPerDay    // e.g., 10240 MB (10 GB)
agent.Metadata.Quotas.MaxContextWindow   // e.g., 32000 tokens

// Execution Quotas
agent.Metadata.Quotas.MaxCallsPerMinute  // e.g., 60
agent.Metadata.Quotas.MaxCallsPerHour    // e.g., 1000
agent.Metadata.Quotas.MaxCallsPerDay     // e.g., 10000
agent.Metadata.Quotas.MaxErrorsPerHour   // e.g., 10
agent.Metadata.Quotas.MaxErrorsPerDay    // e.g., 50

// Enforcement
agent.Metadata.Quotas.EnforceQuotas      // true = block, false = warn
```

---

## 💰 Accessing Cost Metrics

### Thread-Safe Cost Access

```go
// Always use mutex for thread-safe access
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

// Read cost metrics
callCount := agent.Metadata.Cost.CallCount
totalTokens := agent.Metadata.Cost.TotalTokens
dailyCost := agent.Metadata.Cost.DailyCost
lastReset := agent.Metadata.Cost.LastResetTime

fmt.Printf("Agent '%s': %d calls, %d tokens, $%.4f cost\n",
    agent.Name, callCount, totalTokens, dailyCost)
```

### Backward Compatibility: Legacy Cost Access

```go
// Old WEEK 1 way still works for backward compatibility
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

legacyCost := agent.CostMetrics.DailyCost  // Same as agent.Metadata.Cost.DailyCost
```

---

## 🧠 Accessing Memory Metrics

### Memory Usage

```go
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

memory := agent.Metadata.Memory

// Current usage
currentMB := memory.CurrentMemoryMB      // e.g., 256 MB
peakMB := memory.PeakMemoryMB            // e.g., 512 MB
averageMB := memory.AverageMemoryMB      // e.g., 300 MB
trend := memory.MemoryTrendPercent       // e.g., 5.0 (trending up)

fmt.Printf("Memory: %d MB (peak: %d MB, avg: %d MB, trend: %+.1f%%)\n",
    currentMB, peakMB, averageMB, trend)
```

### Memory Quotas

```go
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

memory := agent.Metadata.Memory

// Quota limits
maxPerCall := memory.MaxMemoryMB         // e.g., 512 MB
maxPerDay := memory.MaxDailyMemoryGB     // e.g., 10 GB
alertPercent := memory.MemoryAlertPercent // e.g., 0.80 (80%)

// Check if usage is high
usagePercent := float64(memory.CurrentMemoryMB) / float64(maxPerCall) * 100
if usagePercent > alertPercent*100 {
    fmt.Printf("⚠️  Memory usage at %.1f%% (limit: %.0f%%)\n",
        usagePercent, alertPercent*100)
}
```

### Context Window

```go
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

memory := agent.Metadata.Memory

// Context size tracking
currentContext := memory.CurrentContextSize  // Current tokens in history
maxContext := memory.MaxContextWindow        // e.g., 32000 tokens
contextUsage := float64(currentContext) / float64(maxContext) * 100

if contextUsage > 80 {
    fmt.Printf("Context window: %.1f%% full (will trim %.0f%%)\n",
        contextUsage, memory.ContextTrimPercent*100)
}
```

---

## 🎯 Accessing Performance Metrics

### Success/Failure Rates

```go
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

perf := agent.Metadata.Performance

successfulCalls := perf.SuccessfulCalls
failedCalls := perf.FailedCalls
successRate := perf.SuccessRate        // 0-100

fmt.Printf("Performance: %d successful, %d failed (%.1f%% success rate)\n",
    successfulCalls, failedCalls, successRate)
```

### Response Time

```go
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

perf := agent.Metadata.Performance

avgTime := perf.AverageResponseTime
slowThreshold := perf.SlowCallThreshold

if avgTime > slowThreshold {
    fmt.Printf("⚠️  Average response time %.1fs exceeds threshold %.1fs\n",
        avgTime.Seconds(), slowThreshold.Seconds())
}
```

### Error Tracking

```go
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

perf := agent.Metadata.Performance

lastError := perf.LastError              // Last error message
lastErrorTime := perf.LastErrorTime      // When it occurred
consecutiveErrors := perf.ConsecutiveErrors
errorCountToday := perf.ErrorCountToday

if consecutiveErrors > perf.MaxConsecutiveErrors {
    fmt.Printf("🚨 Too many consecutive errors: %d (max: %d)\n",
        consecutiveErrors, perf.MaxConsecutiveErrors)
}

if errorCountToday > perf.MaxErrorsPerDay {
    fmt.Printf("🚨 Daily error limit exceeded: %d / %d\n",
        errorCountToday, perf.MaxErrorsPerDay)
}
```

---

## 🔐 Thread Safety Best Practices

### Pattern 1: Quick Read

```go
agent.Metadata.Mutex.RLock()
callCount := agent.Metadata.Cost.CallCount
dailyCost := agent.Metadata.Cost.DailyCost
agent.Metadata.Mutex.RUnlock()

// Safe to use values outside lock
fmt.Printf("Calls: %d, Cost: $%.4f\n", callCount, dailyCost)
```

### Pattern 2: Read Multiple Metrics

```go
agent.Metadata.Mutex.RLock()
defer agent.Metadata.Mutex.RUnlock()

// Read all related metrics in one critical section
callCount := agent.Metadata.Cost.CallCount
totalTokens := agent.Metadata.Cost.TotalTokens
dailyCost := agent.Metadata.Cost.DailyCost
memoryUsage := agent.Metadata.Memory.CurrentMemoryMB
successRate := agent.Metadata.Performance.SuccessRate

// Use metrics outside lock
fmt.Printf("Agent %s: %d calls, %d tokens, $%.4f, %d MB memory, %.1f%% success\n",
    agent.Name, callCount, totalTokens, dailyCost, memoryUsage, successRate)
```

### Pattern 3: Conditional Update (Must Use Lock)

```go
agent.Metadata.Mutex.Lock()
defer agent.Metadata.Mutex.Unlock()

// Safe to update in critical section
if agent.Metadata.Cost.DailyCost > agent.Metadata.Quotas.MaxCostPerDay {
    fmt.Println("Daily cost limit exceeded!")
}
agent.Metadata.Cost.DailyCost += 0.50
agent.Metadata.Cost.CallCount++
```

---

## 📝 Real-World Examples

### Example 1: Monitor Agent Health

```go
func monitorAgentHealth(agent *agenticcore.Agent) {
    agent.Metadata.Mutex.RLock()
    defer agent.Metadata.Mutex.RUnlock()

    cost := agent.Metadata.Cost
    perf := agent.Metadata.Performance
    memory := agent.Metadata.Memory

    // Check cost
    costUsage := cost.DailyCost / agent.Metadata.Quotas.MaxCostPerDay
    if costUsage > 0.80 {
        fmt.Printf("⚠️  %s: Cost at %.0f%%\n", agent.Name, costUsage*100)
    }

    // Check performance
    if perf.SuccessRate < 95.0 {
        fmt.Printf("⚠️  %s: Success rate only %.1f%%\n", agent.Name, perf.SuccessRate)
    }

    // Check memory
    if memory.CurrentMemoryMB > int(float64(memory.MaxMemoryMB)*0.90) {
        fmt.Printf("⚠️  %s: Memory at %.0f%%\n", agent.Name,
            float64(memory.CurrentMemoryMB)/float64(memory.MaxMemoryMB)*100)
    }
}
```

### Example 2: Generate Agent Report

```go
func generateAgentReport(agent *agenticcore.Agent) string {
    agent.Metadata.Mutex.RLock()
    defer agent.Metadata.Mutex.RUnlock()

    return fmt.Sprintf(`
Agent Report: %s
===============
Cost Control:
  - Daily Cost: $%.4f / $%.2f (%.1f%%)
  - Tokens: %d / %d (%.1f%%)
  - Calls: %d

Performance:
  - Success Rate: %.1f%%
  - Avg Response: %.2fs
  - Errors Today: %d / %d

Memory Usage:
  - Current: %d MB / %d MB (%.1f%%)
  - Peak: %d MB
  - Context: %d / %d tokens (%.1f%%)
`,
        agent.Name,
        agent.Metadata.Cost.DailyCost,
        agent.Metadata.Quotas.MaxCostPerDay,
        (agent.Metadata.Cost.DailyCost/agent.Metadata.Quotas.MaxCostPerDay)*100,
        agent.Metadata.Cost.TotalTokens,
        agent.Metadata.Quotas.MaxTokensPerDay,
        float64(agent.Metadata.Cost.TotalTokens)/float64(agent.Metadata.Quotas.MaxTokensPerDay)*100,
        agent.Metadata.Cost.CallCount,
        agent.Metadata.Performance.SuccessRate,
        agent.Metadata.Performance.AverageResponseTime.Seconds(),
        agent.Metadata.Performance.ErrorCountToday,
        agent.Metadata.Performance.MaxErrorsPerDay,
        agent.Metadata.Memory.CurrentMemoryMB,
        agent.Metadata.Memory.MaxMemoryMB,
        float64(agent.Metadata.Memory.CurrentMemoryMB)/float64(agent.Metadata.Memory.MaxMemoryMB)*100,
        agent.Metadata.Memory.PeakMemoryMB,
        agent.Metadata.Memory.CurrentContextSize,
        agent.Metadata.Memory.MaxContextWindow,
        float64(agent.Metadata.Memory.CurrentContextSize)/float64(agent.Metadata.Memory.MaxContextWindow)*100,
    )
}
```

### Example 3: Check Quota Compliance

```go
func checkQuotaCompliance(agent *agenticcore.Agent) []string {
    agent.Metadata.Mutex.RLock()
    defer agent.Metadata.Mutex.RUnlock()

    var violations []string

    // Cost quotas
    if agent.Metadata.Cost.DailyCost > agent.Metadata.Quotas.MaxCostPerDay {
        violations = append(violations,
            fmt.Sprintf("Daily cost exceeded: $%.2f > $%.2f",
                agent.Metadata.Cost.DailyCost,
                agent.Metadata.Quotas.MaxCostPerDay))
    }

    if agent.Metadata.Cost.TotalTokens > agent.Metadata.Quotas.MaxTokensPerDay {
        violations = append(violations,
            fmt.Sprintf("Daily tokens exceeded: %d > %d",
                agent.Metadata.Cost.TotalTokens,
                agent.Metadata.Quotas.MaxTokensPerDay))
    }

    // Memory quotas
    if agent.Metadata.Memory.CurrentMemoryMB > agent.Metadata.Quotas.MaxMemoryPerCall {
        violations = append(violations,
            fmt.Sprintf("Memory exceeded: %d MB > %d MB",
                agent.Metadata.Memory.CurrentMemoryMB,
                agent.Metadata.Quotas.MaxMemoryPerCall))
    }

    // Performance quotas
    if agent.Metadata.Performance.ErrorCountToday > agent.Metadata.Performance.MaxErrorsPerDay {
        violations = append(violations,
            fmt.Sprintf("Daily errors exceeded: %d > %d",
                agent.Metadata.Performance.ErrorCountToday,
                agent.Metadata.Performance.MaxErrorsPerDay))
    }

    return violations
}
```

---

## 🔍 Debugging Tips

### View All Metadata at Once

```go
func debugAgent(agent *agenticcore.Agent) {
    agent.Metadata.Mutex.RLock()
    defer agent.Metadata.Mutex.RUnlock()

    fmt.Printf("=== Agent Metadata: %s ===\n", agent.Name)
    fmt.Printf("Created: %s, Last Access: %s\n",
        agent.Metadata.CreatedTime,
        agent.Metadata.LastAccessTime)

    fmt.Println("\nCost Metrics:")
    fmt.Printf("  CallCount: %d\n", agent.Metadata.Cost.CallCount)
    fmt.Printf("  TotalTokens: %d\n", agent.Metadata.Cost.TotalTokens)
    fmt.Printf("  DailyCost: $%.6f\n", agent.Metadata.Cost.DailyCost)

    fmt.Println("\nQuotas:")
    fmt.Printf("  MaxTokensPerDay: %d\n", agent.Metadata.Quotas.MaxTokensPerDay)
    fmt.Printf("  MaxCostPerDay: $%.2f\n", agent.Metadata.Quotas.MaxCostPerDay)
    fmt.Printf("  MaxMemoryPerCall: %d MB\n", agent.Metadata.Quotas.MaxMemoryPerCall)

    fmt.Println("\nPerformance:")
    fmt.Printf("  Success Rate: %.1f%%\n", agent.Metadata.Performance.SuccessRate)
    fmt.Printf("  Calls: %d / %d\n",
        agent.Metadata.Performance.SuccessfulCalls,
        agent.Metadata.Performance.SuccessfulCalls+agent.Metadata.Performance.FailedCalls)
}
```

---

## ⚠️ Important Notes

### Thread Safety
- **Always use mutex** when accessing metadata from multiple goroutines
- Use `RLock()` for read-only access (multiple readers allowed)
- Use `Lock()` for updates (exclusive access)
- Always `defer Unlock()` to prevent deadlocks

### Backward Compatibility
- Old code using `agent.CostMetrics` still works
- New code should prefer `agent.Metadata.Cost`
- Both access the same underlying data (partially synced)

### Default Values
- All quotas have sensible defaults (see WEEK_2_METADATA_INTEGRATION.md)
- Metrics are zero-initialized and updated during execution
- Timestamps use `time.Time{}` for initialization

---

## 📚 Complete API Reference

### Metadata Top-Level
```go
agent.Metadata.AgentID          // string
agent.Metadata.AgentName        // string
agent.Metadata.CreatedTime      // time.Time
agent.Metadata.LastAccessTime   // time.Time
agent.Metadata.Quotas           // AgentQuotaLimits
agent.Metadata.Cost             // AgentCostMetrics
agent.Metadata.Memory           // AgentMemoryMetrics
agent.Metadata.Performance      // AgentPerformanceMetrics
agent.Metadata.Mutex            // sync.RWMutex
```

### AgentQuotaLimits
```go
Quotas.MaxTokensPerCall         // int
Quotas.MaxTokensPerDay          // int
Quotas.MaxCostPerDay            // float64
Quotas.CostAlertPercent         // float64
Quotas.MaxMemoryPerCall         // int
Quotas.MaxMemoryPerDay          // int
Quotas.MaxContextWindow         // int
Quotas.MaxCallsPerMinute        // int
Quotas.MaxCallsPerHour          // int
Quotas.MaxCallsPerDay           // int
Quotas.MaxErrorsPerHour         // int
Quotas.MaxErrorsPerDay          // int
Quotas.EnforceQuotas            // bool
```

### AgentCostMetrics
```go
Cost.CallCount                  // int
Cost.TotalTokens                // int
Cost.DailyCost                  // float64
Cost.LastResetTime              // time.Time
Cost.Mutex                      // sync.RWMutex
```

### AgentMemoryMetrics
```go
Memory.CurrentMemoryMB          // int
Memory.PeakMemoryMB             // int
Memory.AverageMemoryMB          // int
Memory.MemoryTrendPercent       // float64
Memory.MaxMemoryMB              // int
Memory.MaxDailyMemoryGB         // int
Memory.MemoryAlertPercent       // float64
Memory.CurrentContextSize       // int
Memory.MaxContextWindow         // int
Memory.ContextTrimPercent       // float64
Memory.AverageCallDuration      // time.Duration
Memory.SlowCallThreshold        // time.Duration
Memory.Mutex                    // sync.RWMutex
```

### AgentPerformanceMetrics
```go
Performance.SuccessfulCalls     // int
Performance.FailedCalls         // int
Performance.SuccessRate         // float64
Performance.AverageResponseTime // time.Duration
Performance.LastError           // string
Performance.LastErrorTime       // time.Time
Performance.ConsecutiveErrors   // int
Performance.ErrorCountToday     // int
Performance.MaxErrorsPerHour    // int
Performance.MaxErrorsPerDay     // int
Performance.MaxConsecutiveErrors // int
Performance.Mutex               // sync.RWMutex
```

---

**Status:** ✅ GUIDE COMPLETE
**Version:** WEEK 2 - Agent Metadata Usage Guide

