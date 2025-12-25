# Kế Hoạch Refactoring: core/metrics.go - Priority-based Action Items

## 🎯 Tóm Tắt Thực Hành

| Priority | Item | Mức Độ | Effort | Impact |
|---|---|---|---|---|
| 🔴 P0 | Xóa RecordToolExecution (Dead Code) | CRITICAL | 1h | HIGH |
| 🔴 P0 | Fix memory average calculation | CRITICAL | 2h | HIGH |
| 🟠 P1 | Consolidate cost tracking | HIGH | 3h | HIGH |
| 🟠 P1 | Consolidate performance tracking | HIGH | 3h | MEDIUM |
| 🟡 P2 | Fix AverageCallDuration logic | MEDIUM | 1h | MEDIUM |
| 🟡 P2 | Optimize cache hit rate calculation | MEDIUM | 1h | MEDIUM |
| 🟡 P2 | Fix tool execution start time | MEDIUM | 30m | LOW |
| 🟢 P3 | Improve Prometheus export | LOW | 2h | LOW |

---

## 🔴 PHASE 1: CRITICAL FIXES (P0)

### Task 1.1: DELETE RecordToolExecution() - Dead Code Removal

**Priority**: CRITICAL
**Effort**: 1 hour
**Files Affected**: core/metrics.go
**Risk**: NONE (Dead code)

#### What to Remove
1. **Function**: `RecordToolExecution()` (lines 109-178)
2. **Helper Type**: `executionTracker` (lines 87-95)
3. **Field**: `MetricsCollector.currentExecution` (line 84)
4. **Struct**: `ExtendedExecutionMetrics` (lines 14-23)

#### Why It's Safe
```bash
$ grep -r "RecordToolExecution" /Users/taipm/GitHub/go-agentic/
# Returns empty - not called anywhere

$ grep -r "executionTracker" /Users/taipm/GitHub/go-agentic/
# Returns empty - not used anywhere

$ grep -r "ExtendedExecutionMetrics" /Users/taipm/GitHub/go-agentic/
# Returns empty - not used anywhere

$ grep -r "currentExecution" /Users/taipm/GitHub/go-agentic/
# Returns empty - not used anywhere
```

#### Steps
1. Remove lines 87-95 (executionTracker type)
2. Remove lines 14-23 (ExtendedExecutionMetrics type)
3. Remove line 84 (currentExecution field)
4. Remove lines 109-178 (RecordToolExecution method)
5. Run tests to confirm nothing breaks

#### Verification
```bash
go test ./core -v
```

---

### Task 1.2: Fix UpdateMemoryMetrics() Average Calculation Bug

**Priority**: CRITICAL
**Effort**: 2 hours
**Files Affected**: core/common/types.go
**Risk**: MEDIUM (Fixes calculation, may change behavior)

#### Current Bug
```go
// Location: common/types.go:705-713
if a.CostMetrics != nil {
    a.CostMetrics.Mutex.RLock()
    callCount := a.CostMetrics.CallCount
    a.CostMetrics.Mutex.RUnlock()

    if callCount > 0 {
        total := a.MemoryMetrics.PeakMemoryMB * callCount
        a.MemoryMetrics.AverageMemoryMB = total / callCount  // ❌ WRONG!
    }
}
```

**Problem**:
- Formula: `PeakMemoryMB * CallCount / CallCount = PeakMemoryMB`
- Should track: Sum of all memory usage / number of calls
- Result: AverageMemoryMB equals PeakMemoryMB (incorrect!)

#### Solution Approach

**Option A: Track Sum (Recommended)**
```go
// Add new field to AgentMemoryMetrics
type AgentMemoryMetrics struct {
    // ... existing fields ...
    TotalMemoryMB      int     // Sum of all memory samples
    SampleCount        int     // Number of memory samples
    AverageMemoryMB    int     // TotalMemoryMB / SampleCount
}

// Update method
func (a *Agent) UpdateMemoryMetrics(memoryMB int, durationMs int64) {
    // ...
    a.MemoryMetrics.TotalMemoryMB += memoryMB
    a.MemoryMetrics.SampleCount++

    if a.MemoryMetrics.SampleCount > 0 {
        a.MemoryMetrics.AverageMemoryMB =
            a.MemoryMetrics.TotalMemoryMB / a.MemoryMetrics.SampleCount
    }
}
```

**Option B: Simple Fix (Quick)**
```go
// Just use callCount as provided
// Assume callCount = number of memory samples
if callCount > 0 {
    // Need cumulative memory tracking from caller
    // Current approach can't work without sum
}
```

#### Steps
1. Add `TotalMemoryMB` and `SampleCount` fields to AgentMemoryMetrics
2. Update `UpdateMemoryMetrics()` to accumulate total
3. Calculate average correctly
4. Add unit tests for average calculation
5. Update any exports/dashboards using AverageMemoryMB

#### Verification
```go
func TestMemoryMetricsAverage(t *testing.T) {
    agent := createTestAgent()

    // Simulate 5 calls with varying memory
    agent.UpdateMemoryMetrics(100, 1000) // 100 MB
    agent.UpdateMemoryMetrics(150, 1200) // 150 MB
    agent.UpdateMemoryMetrics(120, 900)  // 120 MB
    agent.UpdateMemoryMetrics(200, 1500) // 200 MB
    agent.UpdateMemoryMetrics(80, 800)   // 80 MB

    // Average should be (100+150+120+200+80) / 5 = 130
    expected := 130
    actual := agent.MemoryMetrics.AverageMemoryMB

    if actual != expected {
        t.Errorf("Expected %d, got %d", expected, actual)
    }
}
```

---

## 🟠 PHASE 2: HIGH PRIORITY (P1)

### Task 2.1: Fix AverageCallDuration Logic Bug

**Priority**: HIGH
**Effort**: 1 hour
**Files Affected**: core/common/types.go
**Risk**: MEDIUM

#### Current Bug
```go
// Location: common/types.go:717-720
if durationMs > 0 {
    d := time.Duration(durationMs) * time.Millisecond
    a.MemoryMetrics.AverageCallDuration = d  // ❌ Overwrites each call!
}
```

**Problem**:
- Sets `AverageCallDuration` to **last duration**, not average
- Should be: `Sum of durations / Number of calls`

#### Solution
```go
// Add tracking fields
type AgentMemoryMetrics struct {
    // ...
    TotalDurationMs     int64           // Sum of all call durations
    CallDurationCount   int             // Number of calls tracked
    AverageCallDuration time.Duration   // Average execution time
}

// Fix update method
func (a *Agent) UpdateMemoryMetrics(memoryMB int, durationMs int64) {
    // ...
    if durationMs > 0 {
        a.MemoryMetrics.TotalDurationMs += durationMs
        a.MemoryMetrics.CallDurationCount++

        avg := a.MemoryMetrics.TotalDurationMs / int64(a.MemoryMetrics.CallDurationCount)
        a.MemoryMetrics.AverageCallDuration = time.Duration(avg) * time.Millisecond
    }
}
```

#### Steps
1. Add tracking fields to AgentMemoryMetrics
2. Update calculation logic
3. Add unit tests
4. Verify existing tests still pass

---

### Task 2.2: Consolidate Cost Tracking

**Priority**: HIGH
**Effort**: 3 hours
**Files Affected**: core/metrics.go, core/common/types.go, core/crew.go
**Risk**: MEDIUM (Consolidation)

#### Current State
```
metrics.go:RecordLLMCall()
  ↓
  SystemMetrics (Crew-level)
  - TotalTokens
  - TotalCost
  - SessionTokens
  - SessionCost
  - LLMCallCount

common/types.go:UpdateCostMetrics()
  ↓
  AgentCostMetrics (Agent-level)
  - TotalTokens (DUPLICATE!)
  - DailyCost
  - CallCount (DUPLICATE!)
```

#### Solution: Single Source of Truth

**Option A: Consolidate to metrics.go (Recommended)**
```go
// metrics.go - Add per-agent tracking
type SystemMetrics struct {
    // ... existing fields ...

    // Agent-level cost tracking (moved from Agent)
    AgentCosts map[string]*AgentCostMetrics
}

type AgentCostMetrics struct {
    TotalTokens   int     // Now at system level
    DailyCost     float64 // Now at system level
    CallCount     int     // Now at system level
    LastResetTime time.Time
}

// Remove from common/types.go:Agent
// Remove UpdateCostMetrics() from Agent
```

**Steps**:
1. Add `AgentCosts` map to SystemMetrics
2. Move `AgentCostMetrics` to metrics.go
3. Update `RecordLLMCall()` to track per-agent
4. Remove `Agent.UpdateCostMetrics()` method
5. Remove `Agent.CostMetrics` field
6. Update all call sites
7. Add migration guide for existing code

#### Call Sites to Update
```bash
$ grep -r "UpdateCostMetrics" /Users/taipm/GitHub/go-agentic/
$ grep -r "RecordLLMCall" /Users/taipm/GitHub/go-agentic/
```

#### Steps
1. Identify all update call sites
2. Update to call MetricsCollector.RecordLLMCall() instead
3. Run tests
4. Update documentation

---

### Task 2.3: Consolidate Performance Tracking

**Priority**: HIGH
**Effort**: 3 hours
**Files Affected**: core/metrics.go, core/common/types.go
**Risk**: MEDIUM

#### Current State
```
metrics.go:RecordAgentExecution()
  - ExecutionCount ✅
  - SuccessCount ✅
  - ErrorCount ✅
  - MinDuration ✅
  - MaxDuration ✅
  - AverageDuration ✅

common/types.go:UpdatePerformanceMetrics()
  - SuccessfulCalls (DUPLICATE!)
  - FailedCalls (DUPLICATE!)
  - SuccessRate (CALCULATED!)
  - ConsecutiveErrors (UNIQUE)
  - ErrorCountToday (UNIQUE)
  - LastError (UNIQUE)
  - LastErrorTime (UNIQUE)
```

#### Solution
```go
// metrics.go:AgentMetrics - Add missing fields
type AgentMetrics struct {
    // ... existing fields ...
    SuccessRate       float64  // Calculate like Agent does
    ConsecutiveErrors int      // Copy from Agent
    ErrorCountToday   int      // Copy from Agent
    LastError         string   // Copy from Agent
    LastErrorTime     time.Time // Copy from Agent
}

// Remove Agent.UpdatePerformanceMetrics()
// Have agent call metrics.RecordAgentExecution() instead
```

#### Implementation
```go
// Update RecordAgentExecution to track more metrics
func (mc *MetricsCollector) RecordAgentExecution(...) {
    // ... existing code ...

    // Calculate success rate
    total := agent.SuccessCount + agent.ErrorCount
    if total > 0 {
        agent.SuccessRate = (float64(agent.SuccessCount) / float64(total)) * 100
    }

    // Track consecutive errors
    if success {
        agent.ConsecutiveErrors = 0
    } else {
        agent.ConsecutiveErrors++
    }

    // Track today's errors
    agent.ErrorCountToday++
}
```

---

## 🟡 PHASE 3: MEDIUM PRIORITY (P2)

### Task 3.1: Optimize Cache Hit Rate Calculation

**Priority**: MEDIUM
**Effort**: 1 hour
**Files Affected**: core/metrics.go
**Risk**: NONE

#### Current Problem
```go
// Called every time!
func (mc *MetricsCollector) RecordCacheHit() {
    // ...
    mc.systemMetrics.CacheHits++
    mc.updateCacheHitRate()  // ⚠️ O(1) * millions = performance hit
}
```

#### Solution A: Lazy Evaluation (Recommended)
```go
// Remove updateCacheHitRate() calls
// Calculate on-demand instead

func (mc *MetricsCollector) RecordCacheHit() {
    if !mc.enabled {
        return
    }

    mc.mu.Lock()
    defer mc.mu.Unlock()

    mc.systemMetrics.CacheHits++
    // Don't calculate here
}

// New getter method
func (mc *MetricsCollector) GetCacheMetrics() (hits, misses int64, hitRate float64) {
    mc.mu.RLock()
    defer mc.mu.RUnlock()

    hits = mc.systemMetrics.CacheHits
    misses = mc.systemMetrics.CacheMisses

    total := hits + misses
    if total > 0 {
        hitRate = float64(hits) / float64(total)
    }
    return
}
```

#### Steps
1. Remove `updateCacheHitRate()` calls from RecordCacheHit/Miss
2. Remove `updateCacheHitRate()` method
3. Update GetSystemMetrics() to compute hitRate on demand
4. Or: Add GetCacheMetrics() getter as shown above
5. Test with high cache volume

---

### Task 3.2: Fix RecordToolExecution StartTime Bug

**Priority**: MEDIUM
**Effort**: 30 minutes
**Files Affected**: core/metrics.go (if keeping)
**Risk**: NONE

#### Current Bug
```go
StartTime: time.Now().Add(-duration),  // ❌ Inaccurate
EndTime:   time.Now(),
```

#### Fix (if tool execution tracking is kept)
```go
// Should pass actual startTime from caller
func (mc *MetricsCollector) RecordToolExecution(
    toolName string,
    startTime time.Time,  // ← Add this
    endTime time.Time,    // ← Add this
    success bool,
) {
    duration := endTime.Sub(startTime)

    metric := ExtendedExecutionMetrics{
        ToolName:  toolName,
        Duration:  duration,
        Status:    statusString(success),
        Success:   success,
        StartTime: startTime,  // ← Actual time
        EndTime:   endTime,    // ← Actual time
    }
}
```

**Note**: If Task 1.1 is done (delete method), this is not needed

---

## 🟢 PHASE 4: NICE-TO-HAVE (P3)

### Task 4.1: Improve Prometheus Export

**Priority**: LOW
**Effort**: 2 hours
**Files Affected**: core/metrics.go
**Risk**: NONE

#### Issues to Fix
1. String concatenation inefficiency (use strings.Builder)
2. Tool metrics not exported
3. Missing agent-specific metrics (success rate)

#### Current
```go
func (mc *MetricsCollector) exportPrometheus() (string, error) {
    var result string

    // String concatenation (inefficient)
    result += fmt.Sprintf("...")
    result += fmt.Sprintf("...")
    // This creates N temporary strings!
}
```

#### Better
```go
func (mc *MetricsCollector) exportPrometheus() (string, error) {
    var builder strings.Builder

    // Write directly to builder
    builder.WriteString("# HELP crew_requests_total...\n")
    builder.WriteString("# TYPE crew_requests_total counter\n")

    // ... more metrics ...

    // Add tool metrics
    for agentID, agent := range mc.systemMetrics.AgentMetrics {
        for toolName, tool := range agent.ToolMetrics {
            fmt.Fprintf(&builder,
                "crew_tool_executions{agent=\"%s\",tool=\"%s\"} %d\n",
                agentID, toolName, tool.ExecutionCount)
        }
    }

    return builder.String(), nil
}
```

#### Steps
1. Refactor string building to use strings.Builder
2. Add tool metrics export
3. Add agent success rate export
4. Test Prometheus scrape compatibility
5. Document metrics schema

---

## 📋 IMPLEMENTATION CHECKLIST

### Phase 1: Critical Fixes
- [ ] Task 1.1: Remove RecordToolExecution dead code
- [ ] Task 1.2: Fix UpdateMemoryMetrics average calculation
- [ ] Run: `go test ./core -v`
- [ ] Run: `go vet ./...`

### Phase 2: High Priority
- [ ] Task 2.1: Fix AverageCallDuration logic
- [ ] Task 2.2: Consolidate cost tracking
- [ ] Task 2.3: Consolidate performance tracking
- [ ] Update call sites in crew.go
- [ ] Run all tests
- [ ] Update documentation

### Phase 3: Medium Priority
- [ ] Task 3.1: Optimize cache hit rate
- [ ] Task 3.2: Fix tool execution start time (if needed)
- [ ] Performance testing with high volumes

### Phase 4: Nice-to-Have
- [ ] Task 4.1: Improve Prometheus export
- [ ] Add metrics schema documentation
- [ ] Performance benchmarks

---

## 🧪 TESTING STRATEGY

### Unit Tests to Add/Update
```go
// Test memory metrics calculation
TestMemoryMetricsAverage()
TestAverageCallDuration()

// Test cost consolidation
TestCostTrackingConsolidation()
TestAgentLevelCostTracking()

// Test performance metrics
TestSuccessRateCalculation()
TestConsecutiveErrorTracking()

// Test cache optimization
TestCacheHitRateCalculation()
TestCacheLazyEvaluation()
```

### Integration Tests
- `crew_test.go`: Update metrics recording tests
- `agent_test.go`: Update agent metrics tests

### Performance Tests
```go
// Before/after comparison
BenchmarkCacheHitRecording()     // Should be faster after optimization
BenchmarkMetricsCollection()     // Overall performance
```

---

## 🔄 MIGRATION NOTES

### Breaking Changes
- `Agent.UpdateCostMetrics()` → Removed (use MetricsCollector)
- `Agent.UpdatePerformanceMetrics()` → Removed (use MetricsCollector)
- `Agent.UpdateMemoryMetrics()` → Signature change (add TotalMemoryMB tracking)

### Deprecation Path
1. Phase 1: Keep old methods, mark as deprecated
2. Phase 2: Log warnings when old methods called
3. Phase 3: Remove deprecated methods

### Update Required In
- [ ] `core/crew.go:updateAgentMetrics()`
- [ ] `core/agent/execution.go`
- [ ] All tests referencing old methods
- [ ] Any external consumers of these APIs

---

## 🎯 Success Criteria

| Task | Success Criteria |
|---|---|
| 1.1 | All tests pass, no dead code references |
| 1.2 | Average = (100+150+120) / 3, not max value |
| 2.1 | Average duration = sum / count, not last value |
| 2.2 | Single source of truth for cost metrics |
| 2.3 | Single source of truth for performance metrics |
| 3.1 | Hit rate computed on-demand, not cached |
| 4.1 | Tool metrics visible in Prometheus export |

---

**Created**: 2025-12-25
**Estimated Total Effort**: 13 hours (P0-P3)
**Quick Wins**: Phase 1 (3 hours) fixes most critical issues

