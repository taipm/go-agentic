# Phase 3: Consolidate Duplicate Tracking - Implementation Plan

**Status**: IN PROGRESS
**Effort**: ~6 hours
**Complexity**: MEDIUM (Careful consolidation required)
**Risk**: MEDIUM (Breaking changes for Agent methods)

---

## 🎯 Consolidation Strategy

### Objective
Create **single source of truth** for metrics in `MetricsCollector` instead of duplicating logic between:
- `metrics.go` (System-level)
- `common/types.go` (Agent-level)

### Approach
1. ✅ Keep Agent metrics classes (for agent-specific tracking)
2. ✅ Keep Agent update methods (for backward compatibility initially)
3. 🔄 Extend SystemMetrics to track per-agent details
4. 🔄 Update RecordAgentExecution to capture all needed data
5. 🔄 Update crew.go to call metrics.RecordAgentExecution() instead of agent methods

---

## 📋 Task Breakdown

### Task 3.1: Extend SystemMetrics for Per-Agent Cost Tracking

**File**: `core/metrics.go`

**Current SystemMetrics**:
```go
type SystemMetrics struct {
    // ... existing fields ...
    TotalTokens   int     // System-level only
    TotalCost     float64 // System-level only
    SessionTokens int
    SessionCost   float64
    LLMCallCount  int
}
```

**Target**: Add per-agent cost tracking
```go
type SystemMetrics struct {
    // ... existing fields ...

    // ✅ PHASE 3: Per-agent cost tracking
    AgentCosts map[string]*AgentCostMetrics  // Agent-level cost metrics

    // ... existing fields ...
}
```

**Steps**:
1. Add `AgentCosts` map to SystemMetrics
2. Update `RecordLLMCall()` to track per-agent
3. Initialize map in `NewMetricsCollector()`

---

### Task 3.2: Consolidate Cost Tracking Logic

**File**: `core/metrics.go`

**Update RecordLLMCall()**:

```go
// Current (duplicated with Agent.UpdateCostMetrics):
func (mc *MetricsCollector) RecordLLMCall(agentID string, tokens int, cost float64) {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    mc.systemMetrics.LastUpdated = time.Now()
    mc.systemMetrics.TotalTokens += tokens
    mc.systemMetrics.TotalCost += cost
    mc.systemMetrics.SessionTokens += tokens
    mc.systemMetrics.SessionCost += cost
    mc.systemMetrics.LLMCallCount++
}

// ✅ PHASE 3: Also track per-agent
func (mc *MetricsCollector) RecordLLMCall(agentID string, tokens int, cost float64) {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    mc.systemMetrics.LastUpdated = time.Now()

    // System-level tracking
    mc.systemMetrics.TotalTokens += tokens
    mc.systemMetrics.TotalCost += cost
    mc.systemMetrics.SessionTokens += tokens
    mc.systemMetrics.SessionCost += cost
    mc.systemMetrics.LLMCallCount++

    // ✅ PHASE 3: Agent-level tracking
    if agentID != "" {
        agentCost, exists := mc.systemMetrics.AgentCosts[agentID]
        if !exists {
            agentCost = &AgentCostMetrics{
                LastResetTime: time.Now(),
            }
            mc.systemMetrics.AgentCosts[agentID] = agentCost
        }

        agentCost.Mutex.Lock()
        defer agentCost.Mutex.Unlock()

        agentCost.CallCount++
        agentCost.TotalTokens += tokens
        agentCost.DailyCost += cost

        // Reset daily if needed
        now := time.Now()
        if now.Sub(agentCost.LastResetTime) > 24*time.Hour {
            agentCost.DailyCost = cost
            agentCost.LastResetTime = now
        }
    }
}
```

---

### Task 3.3: Consolidate Performance Tracking

**File**: `core/metrics.go`

**Update RecordAgentExecution()**:

**Current** (missing some metrics):
```go
func (mc *MetricsCollector) RecordAgentExecution(agentID, agentName string, duration time.Duration, success bool) {
    // ... existing code ...

    if success {
        agent.SuccessCount++
    } else {
        agent.ErrorCount++
    }
}
```

**Target** (include all performance metrics):
```go
// ✅ PHASE 3: Track ConsecutiveErrors, ErrorCountToday, etc.
func (mc *MetricsCollector) RecordAgentExecution(agentID, agentName string, duration time.Duration, success bool, errorMsg string) {
    // ... existing code ...

    // ✅ PHASE 3: Track success/error/consecutive errors
    if success {
        agent.SuccessCount++
        agent.ConsecutiveErrors = 0  // ✅ NEW
    } else {
        agent.ErrorCount++
        agent.ConsecutiveErrors++    // ✅ NEW
        agent.LastError = errorMsg   // ✅ NEW
        agent.LastErrorTime = time.Now()  // ✅ NEW
        agent.ErrorCountToday++      // ✅ NEW
    }

    // ✅ PHASE 3: Calculate success rate
    total := agent.SuccessCount + agent.ErrorCount
    if total > 0 {
        agent.SuccessRate = (float64(agent.SuccessCount) / float64(total)) * 100
    }
}
```

**Add to AgentMetrics struct**:
```go
type AgentMetrics struct {
    // ... existing fields ...

    // ✅ PHASE 3: Per-agent performance fields
    ConsecutiveErrors int
    ErrorCountToday   int
    LastError         string
    LastErrorTime     time.Time
    SuccessRate       float64
}
```

---

### Task 3.4: Consolidate Memory Tracking

**File**: `core/metrics.go`

**Update UpdateMemoryUsage()**:

```go
// ✅ PHASE 3: Store agent-level memory metrics too
func (mc *MetricsCollector) UpdateMemoryUsage(agentID string, memoryMB int) {
    if !mc.enabled {
        return
    }

    mc.mu.Lock()
    defer mc.mu.Unlock()

    // System-level memory
    currentBytes := uint64(memoryMB) * 1024 * 1024  // Convert MB to bytes
    mc.systemMetrics.MemoryUsage = currentBytes
    if currentBytes > mc.systemMetrics.MaxMemoryUsage {
        mc.systemMetrics.MaxMemoryUsage = currentBytes
    }

    // ✅ PHASE 3: Agent-level memory tracking
    if agentID != "" {
        agent, exists := mc.systemMetrics.AgentMetrics[agentID]
        if exists && agent != nil {
            // Track memory in agent metrics
            if agent.MemoryMetrics != nil {
                agent.MemoryMetrics.CurrentMemoryMB = memoryMB
                if memoryMB > agent.MemoryMetrics.PeakMemoryMB {
                    agent.MemoryMetrics.PeakMemoryMB = memoryMB
                }
                // ... rest of memory tracking ...
            }
        }
    }
}
```

---

### Task 3.5: Update crew.go Call Sites

**File**: `core/crew.go`

**Current**:
```go
func (ce *CrewExecutor) updateAgentMetrics(agent *Agent, success bool, duration time.Duration, memory int, errorMsg string) error {
    if agent == nil || agent.Metadata == nil {
        return nil
    }

    // Update performance metrics
    agent.UpdatePerformanceMetrics(success, errorMsg)

    // Update memory metrics (convert duration to milliseconds)
    durationMs := duration.Milliseconds()
    agent.UpdateMemoryMetrics(memory, durationMs)

    return nil
}
```

**Target**:
```go
// ✅ PHASE 3: Call MetricsCollector instead of agent methods
func (ce *CrewExecutor) updateAgentMetrics(agent *Agent, success bool, duration time.Duration, memory int, errorMsg string) error {
    if agent == nil {
        return nil
    }

    // ✅ PHASE 3: Use MetricsCollector for central tracking
    if ce.Metrics != nil {
        // Record execution (includes performance metrics)
        ce.Metrics.RecordAgentExecution(agent.ID, agent.Name, duration, success, errorMsg)

        // Record memory usage (agent-level)
        ce.Metrics.UpdateMemoryUsage(agent.ID, memory)
    }

    // ✅ Keep Agent methods for backward compatibility (initially)
    // But they'll sync with MetricsCollector
    agent.UpdatePerformanceMetrics(success, errorMsg)
    agent.UpdateMemoryMetrics(memory, duration.Milliseconds())

    return nil
}
```

---

### Task 3.6: Add Missing Fields to AgentMetrics

**File**: `core/metrics.go`

**Update AgentMetrics struct**:

```go
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
    ToolMetrics     map[string]*ToolMetrics

    // ✅ PHASE 3: Performance metrics (copied from Agent.PerformanceMetrics)
    ConsecutiveErrors int
    ErrorCountToday   int
    LastError         string
    LastErrorTime     time.Time
    SuccessRate       float64

    // ✅ PHASE 3: Memory metrics aggregation (optional)
    MemoryMetrics     *MemoryMetricsSnapshot  // Optional: peak memory, avg memory
}

// ✅ PHASE 3: Optional memory snapshot for system-level aggregation
type MemoryMetricsSnapshot struct {
    CurrentMemoryMB int
    PeakMemoryMB    int
}
```

---

## 🔄 Implementation Sequence

### Step 1: Extend SystemMetrics (30 min)
- Add `AgentCosts` map
- Add fields to `AgentMetrics`
- Update `NewMetricsCollector()`

### Step 2: Update RecordLLMCall (30 min)
- Add per-agent cost tracking logic
- Sync with AgentCostMetrics fields

### Step 3: Update RecordAgentExecution (1 hour)
- Add performance metric tracking
- Calculate success rate
- Track consecutive errors

### Step 4: Add UpdateMemoryUsage param (30 min)
- Add `agentID` parameter
- Track agent-level memory

### Step 5: Update crew.go call sites (1 hour)
- Find all calls to agent update methods
- Update to call MetricsCollector
- Add errorMsg parameter to RecordAgentExecution

### Step 6: Testing & Verification (1.5 hours)
- Run existing tests
- Verify metrics accuracy
- Check thread safety
- Validate consolidation

---

## ⚠️ Breaking Changes

### Signature Changes

**RecordAgentExecution()**:
```go
// BEFORE:
func (mc *MetricsCollector) RecordAgentExecution(agentID, agentName string, duration time.Duration, success bool)

// AFTER:
func (mc *MetricsCollector) RecordAgentExecution(agentID, agentName string, duration time.Duration, success bool, errorMsg string)
```

**UpdateMemoryUsage()**:
```go
// BEFORE:
func (mc *MetricsCollector) UpdateMemoryUsage(current uint64)

// AFTER:
func (mc *MetricsCollector) UpdateMemoryUsage(agentID string, memoryMB int)
```

### Call Site Updates Required
- Find all calls to these methods
- Update signatures with new parameters
- Update calling code

---

## 📊 Expected Impact

### Before Consolidation
```
metrics.go:RecordAgentExecution()
  ├─ Tracks: ExecutionCount, SuccessCount, ErrorCount
  ├─ Tracks: Min/Max/Average duration
  └─ Missing: ConsecutiveErrors, ErrorCountToday, LastError

Agent.UpdatePerformanceMetrics()
  ├─ Tracks: SuccessfulCalls, FailedCalls
  ├─ Tracks: SuccessRate
  ├─ Tracks: ConsecutiveErrors, ErrorCountToday
  └─ Missing: Min/Max/Average duration
```

### After Consolidation
```
metrics.go:RecordAgentExecution() - SINGLE SOURCE OF TRUTH
  ├─ Tracks: ExecutionCount, SuccessCount, ErrorCount
  ├─ Tracks: Min/Max/Average duration
  ├─ Tracks: SuccessRate
  ├─ Tracks: ConsecutiveErrors, ErrorCountToday, LastError
  ├─ Tracks: Per-agent cost (via RecordLLMCall)
  └─ Tracks: Per-agent memory (via UpdateMemoryUsage)
```

### Code Quality Improvements
- **Elimination of duplicate logic**: ~100+ lines
- **Single source of truth**: All metrics in one place
- **Better maintainability**: Update once, everywhere
- **Easier to add new metrics**: One place to add

---

## 🧪 Testing Strategy

### Unit Tests Needed
1. Cost consolidation: RecordLLMCall tracks per-agent
2. Performance consolidation: RecordAgentExecution tracks all metrics
3. Memory consolidation: UpdateMemoryUsage tracks per-agent
4. Accuracy: Metrics match what Agent methods were tracking

### Test Cases
```go
// Test cost consolidation
func TestRecordLLMCallPerAgent(t *testing.T) {
    mc := NewMetricsCollector()

    // Record for agent A
    mc.RecordLLMCall("agent-a", 100, 0.01)
    mc.RecordLLMCall("agent-a", 200, 0.02)

    // Verify system total
    assert.Equal(t, 300, mc.systemMetrics.TotalTokens)
    assert.Equal(t, 0.03, mc.systemMetrics.TotalCost)

    // Verify agent-level
    agentCost := mc.systemMetrics.AgentCosts["agent-a"]
    assert.Equal(t, 300, agentCost.TotalTokens)
    assert.Equal(t, 2, agentCost.CallCount)
}

// Test performance consolidation
func TestRecordAgentExecutionPerformance(t *testing.T) {
    mc := NewMetricsCollector()

    // Record success
    mc.RecordAgentExecution("agent-a", "Agent A", 100*time.Millisecond, true, "")

    // Record failure
    mc.RecordAgentExecution("agent-a", "Agent A", 150*time.Millisecond, false, "timeout")

    // Verify metrics
    agent := mc.systemMetrics.AgentMetrics["agent-a"]
    assert.Equal(t, int64(2), agent.ExecutionCount)
    assert.Equal(t, int64(1), agent.SuccessCount)
    assert.Equal(t, int64(1), agent.ErrorCount)
    assert.Equal(t, "timeout", agent.LastError)
    assert.Equal(t, 50.0, agent.SuccessRate)  // 1/2
}
```

---

## 🎯 Success Criteria

- [x] Extend SystemMetrics for per-agent tracking
- [ ] RecordLLMCall tracks both system and agent costs
- [ ] RecordAgentExecution captures all performance metrics
- [ ] UpdateMemoryUsage tracks agent-level memory
- [ ] crew.go updated to use MetricsCollector
- [ ] All tests pass
- [ ] No undefined references
- [ ] Code compiles successfully

---

## 📝 Rollback Plan

If consolidation causes issues:

1. **Keep Agent methods**: Don't delete, just deprecate
2. **Keep dual tracking**: Both metrics.go and Agent methods
3. **Add deprecation warnings**: Alert users
4. **Timeline**: Phase out over 2-3 releases

---

**Estimated Completion**: ~6 hours
**Next Phase**: Optimization (cache rate, Prometheus export)

