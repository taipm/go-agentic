# Chi Tiết Từng Hàm: core/metrics.go - Phân Tích Sâu

## 📋 Danh Sách Hàm

| # | Hàm | Dòng | Loại | Trạng Thái | Ghi Chú |
|---|---|---|---|---|---|
| 1 | `NewMetricsCollector()` | 97-106 | Constructor | ✅ OK | Khởi tạo collectors |
| 2 | `RecordToolExecution()` | 109-178 | Recording | ❌ DEAD | Không được gọi |
| 3 | `RecordAgentExecution()` | 181-240 | Recording | ✅ OK | Nhưng duplicate với Agent methods |
| 4 | `RecordLLMCall()` | 244-258 | Recording | ⚠️ DUPLICATE | Trùng với UpdateCostMetrics |
| 5 | `ResetSessionCost()` | 261-271 | Reset | ✅ OK | Session-level reset |
| 6 | `GetSessionCost()` | 274-279 | Query | ✅ OK | Thread-safe getter |
| 7 | `GetTotalCost()` | 282-287 | Query | ✅ OK | Thread-safe getter |
| 8 | `LogCrewCostSummary()` | 290-304 | Logging | ✅ OK | Diagnostic output |
| 9 | `RecordCacheHit()` | 307-317 | Recording | ⚠️ INEFFICIENT | Calls updateCacheHitRate() mỗi lần |
| 10 | `RecordCacheMiss()` | 320-330 | Recording | ⚠️ INEFFICIENT | Calls updateCacheHitRate() mỗi lần |
| 11 | `updateCacheHitRate()` | 333-338 | Internal | ⚠️ INEFFICIENT | Gọi quá nhiều lần |
| 12 | `UpdateMemoryUsage()` | 341-353 | Recording | ⚠️ DUPLICATE | Trùng với UpdateMemoryMetrics |
| 13 | `GetSystemMetrics()` | 356-363 | Query | ✅ OK | Returns copy (good) |
| 14 | `ExportMetrics()` | 366-378 | Export | ✅ OK | Supports JSON & Prometheus |
| 15 | `exportJSON()` | 381-392 | Internal | ✅ OK | JSON formatting |
| 16 | `exportPrometheus()` | 395-443 | Internal | ✅ OK | Prometheus formatting |
| 17 | `Reset()` | 446-454 | Reset | ✅ OK | Clear all metrics |
| 18 | `Enable()` | 457-461 | Control | ✅ OK | Enable collection |
| 19 | `Disable()` | 464-468 | Control | ✅ OK | Disable collection |
| 20 | `IsEnabled()` | 471-475 | Query | ✅ OK | Check status |
| 21 | `statusString()` | 478-483 | Helper | ✅ OK | Simple helper |

---

## 🔴 CHI TIẾT HÀNG HÀNG

---

### 1️⃣ `NewMetricsCollector()` - Constructor ✅ OK

**Vị trí**: Lines 97-106

**Mục đích**: Khởi tạo MetricsCollector mới

```go
func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        systemMetrics: &SystemMetrics{
            StartTime:    time.Now(),
            AgentMetrics: make(map[string]*AgentMetrics),
        },
        enabled: true,
    }
}
```

**Đánh giá**:
- ✅ Khởi tạo đúng tất cả fields
- ✅ Default enable metrics collection
- ⚠️ `currentExecution` không được khởi tạo (DEAD CODE)

**Khuyến nghị**:
- Remove `currentExecution` field nếu không dùng
- Hoặc khởi tạo nó nếu cần

---

### 2️⃣ `RecordToolExecution()` - Recording ❌ DEAD CODE

**Vị trí**: Lines 109-178

**Mục đích**: Record execution metrics cho từng tool

**Độ dài**: 70 lines

**Phức tạp**: HIGH (5 levels of nesting)

```go
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
            StartTime: time.Now().Add(-duration),  // ❌ BUG: Inaccurate calculation
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
```

**Đánh giá**:
- ❌ **DEAD CODE** - Không được gọi từ bất kỳ đâu
- ❌ **BUG**: `StartTime: time.Now().Add(-duration)` là inaccurate
- ❌ **Incomplete**: `currentExecution` không bao giờ được initialized
- ❌ **Unused**: `TimedOut` flag không được set

**Call Sites**:
```bash
$ grep -r "RecordToolExecution" /Users/taipm/GitHub/go-agentic/
# (No results - confirmed not called)
```

**Khuyến nghị**:
- **DELETE** hàm này nếu không cần
- Hoặc implement properly nếu cần tool-level tracking

---

### 3️⃣ `RecordAgentExecution()` - Recording ✅ BUT DUPLICATE

**Vị trí**: Lines 181-240

**Mục đích**: Record execution metrics cho entire agent

**Độ dài**: 60 lines

**Phức tạp**: MEDIUM

```go
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
```

**Đánh giá**:
- ✅ Logic đúng (system-level aggregation)
- ✅ Handles min/max/average correctly
- ⚠️ **DUPLICATE** với `Agent.UpdatePerformanceMetrics()` (common/types.go)
- ❌ Missing `TimeoutCount` tracking (never set)

**Call Sites**:
```go
// crew.go:422
ce.Metrics.RecordAgentExecution(agent.ID, agent.Name, duration, success)
```

**Khuyến nghị**:
- **Consolidate**: Keep này ở metrics.go, remove Agent.UpdatePerformanceMetrics()
- Hoặc: Make Agent methods call MetricsCollector

---

### 4️⃣ `RecordLLMCall()` - Recording ⚠️ DUPLICATE COST TRACKING

**Vị trí**: Lines 244-258

**Mục đích**: Record LLM API call tokens & cost

**Độ dài**: 15 lines

```go
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
```

**Đánh giá**:
- ✅ Logic chính xác
- ⚠️ **DUPLICATE** với `Agent.UpdateCostMetrics()` (common/types.go:668)
- ❌ Không nhận `agentID` parameter nhưng không sử dụng

**Comparison**:
```
metrics.go:
  + TotalTokens (System-level)
  + TotalCost (System-level)
  + SessionTokens (Reset on ClearHistory)
  + SessionCost (Reset on ClearHistory)
  + LLMCallCount

common/types.go:
  + TotalTokens (Agent-level)
  + DailyCost (Auto-reset every 24h)
  + CallCount (Agent-specific)
```

**Khuyến nghị**:
- **Consolidate** vào metrics.go
- Có `agentID` là hint rằng nên track per-agent
- Thêm agentID tracking vào SystemMetrics

---

### 5️⃣ `ResetSessionCost()` - Reset ✅ OK

**Vị trí**: Lines 261-271

**Mục đích**: Reset session-level cost tracking

```go
func (mc *MetricsCollector) ResetSessionCost() {
    if !mc.enabled {
        return
    }

    mc.mu.Lock()
    defer mc.mu.Unlock()

    mc.systemMetrics.SessionTokens = 0
    mc.systemMetrics.SessionCost = 0
}
```

**Đánh giá**:
- ✅ Logic chính xác
- ✅ Thread-safe
- ✅ Simple & clear

**Call Sites**:
```bash
# Tìm xem có gọi từ đâu
grep -r "ResetSessionCost" /Users/taipm/GitHub/go-agentic/
```

**Khuyến nghị**:
- Keep as-is ✅

---

### 6️⃣ `GetSessionCost()` - Query ✅ OK

**Vị trí**: Lines 274-279

**Mục đích**: Get current session cost

```go
func (mc *MetricsCollector) GetSessionCost() (tokens int, cost float64) {
    mc.mu.RLock()
    defer mc.mu.RUnlock()

    return mc.systemMetrics.SessionTokens, mc.systemMetrics.SessionCost
}
```

**Đánh giá**:
- ✅ Simple & clean
- ✅ Thread-safe read
- ✅ Returns tuple (idiomatic Go)

**Khuyến nghị**:
- Keep as-is ✅

---

### 7️⃣ `GetTotalCost()` - Query ✅ OK

**Vị trí**: Lines 282-287

**Mục đích**: Get total cost across all sessions

```go
func (mc *MetricsCollector) GetTotalCost() (tokens int, cost float64, calls int) {
    mc.mu.RLock()
    defer mc.mu.RUnlock()

    return mc.systemMetrics.TotalTokens, mc.systemMetrics.TotalCost, mc.systemMetrics.LLMCallCount
}
```

**Đánh giá**:
- ✅ Simple & clean
- ✅ Thread-safe read
- ✅ Includes call count

**Khuyến nghị**:
- Keep as-is ✅

---

### 8️⃣ `LogCrewCostSummary()` - Logging ✅ OK

**Vị trí**: Lines 290-304

**Mục đích**: Log crew cost summary

```go
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
```

**Đánh giá**:
- ✅ Useful diagnostic output
- ✅ Thread-safe read
- ✅ Good formatting

**Khuyến nghị**:
- Keep as-is ✅
- Optional: Consider using log package instead of fmt.Printf

---

### 9️⃣ `RecordCacheHit()` - Recording ⚠️ INEFFICIENT

**Vị trí**: Lines 307-317

**Mục đích**: Record cache hit

```go
func (mc *MetricsCollector) RecordCacheHit() {
    if !mc.enabled {
        return
    }

    mc.mu.Lock()
    defer mc.mu.Unlock()

    mc.systemMetrics.CacheHits++
    mc.updateCacheHitRate()
}
```

**Vấn Đề**:
- ⚠️ Calls `updateCacheHitRate()` **MỖI LẦN** có cache hit
- Nếu có 1 triệu cache hits, tính toán hit rate 1 triệu lần!

**Khuyến nghị**:
- **Option 1**: Remove `updateCacheHitRate()` call, calculate on-demand ở getter
- **Option 2**: Batch updates (call every N hits)
- **Option 3**: Inline calculation

---

### 🔟 `RecordCacheMiss()` - Recording ⚠️ INEFFICIENT

**Vị trí**: Lines 320-330

**Mục đích**: Record cache miss

```go
func (mc *MetricsCollector) RecordCacheMiss() {
    if !mc.enabled {
        return
    }

    mc.mu.Lock()
    defer mc.mu.Unlock()

    mc.systemMetrics.CacheMisses++
    mc.updateCacheHitRate()
}
```

**Vấn Đề**: Same as RecordCacheHit()

**Khuyến nghị**: Same as RecordCacheHit()

---

### 1️⃣1️⃣ `updateCacheHitRate()` - Internal ⚠️ INEFFICIENT

**Vị trí**: Lines 333-338

**Mục đích**: Calculate cache hit rate

```go
func (mc *MetricsCollector) updateCacheHitRate() {
    total := mc.systemMetrics.CacheHits + mc.systemMetrics.CacheMisses
    if total > 0 {
        mc.systemMetrics.CacheHitRate = float64(mc.systemMetrics.CacheHits) / float64(total)
    }
}
```

**Vấn Đề**:
- ⚠️ Called **every time** RecordCacheHit/RecordCacheMiss is called
- **Performance Impact**: O(1) operation but repeated millions of times
- **Alternative**: Calculate on-demand (lazy evaluation)

**Recommended Solution**:
```go
// Option A: On-demand calculation
func (mc *MetricsCollector) GetCacheHitRate() float64 {
    mc.mu.RLock()
    defer mc.mu.RUnlock()

    total := mc.systemMetrics.CacheHits + mc.systemMetrics.CacheMisses
    if total > 0 {
        return float64(mc.systemMetrics.CacheHits) / float64(total)
    }
    return 0
}

// Then remove updateCacheHitRate() call
```

---

### 1️⃣2️⃣ `UpdateMemoryUsage()` - Recording ⚠️ DUPLICATE

**Vị trí**: Lines 341-353

**Mục đích**: Update system-level memory usage

```go
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
```

**Đánh giá**:
- ✅ Logic chính xác (peak tracking)
- ⚠️ **DUPLICATE** với `Agent.UpdateMemoryMetrics()` (common/types.go:690)
- ⚠️ Unit mismatch: System uses bytes, Agent uses MB
- ❌ No average calculation (unlike Agent-level)

**Comparison**:
```
metrics.go:UpdateMemoryUsage():
  - Current: uint64 (bytes)
  - Peak: uint64 (bytes)
  - Average: ❌ NOT tracked

common/types.go:UpdateMemoryMetrics():
  - Current: int (MB)
  - Peak: int (MB)
  - Average: int (MB) ❌ Calculated WRONG!
  - Call Duration: tracked
```

**Khuyến nghị**:
- **Consolidate** vào metrics.go
- Use consistent units (preferably MB)
- Fix average calculation ở agent level

---

### 1️⃣3️⃣ `GetSystemMetrics()` - Query ✅ OK

**Vị trí**: Lines 356-363

**Mục đích**: Get copy of system metrics

```go
func (mc *MetricsCollector) GetSystemMetrics() *SystemMetrics {
    mc.mu.RLock()
    defer mc.mu.RUnlock()

    // Return a copy to prevent external modifications
    metrics := *mc.systemMetrics
    return &metrics
}
```

**Đánh giá**:
- ✅ Returns copy (prevents external modification)
- ✅ Thread-safe
- ⚠️ Shallow copy (AgentMetrics map pointers are shared)

**Potential Issue**:
```go
// User could modify agent metrics through returned copy
metrics := collector.GetSystemMetrics()
metrics.AgentMetrics[agentID].ExecutionCount = 999  // Modifies original!
```

**Khuyến nghị**:
- Either document that AgentMetrics are shared
- Or implement deep copy
- Current implementation is acceptable for most use cases

---

### 1️⃣4️⃣ `ExportMetrics()` - Export ✅ OK

**Vị trí**: Lines 366-378

**Mục đích**: Export metrics in specified format

```go
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
```

**Đánh giá**:
- ✅ Clean dispatcher
- ✅ Thread-safe
- ✅ Good error message

**Khuyến nghị**:
- Keep as-is ✅
- Could add more formats (CSV, YAML) in future

---

### 1️⃣5️⃣ `exportJSON()` - Internal ✅ OK

**Vị trí**: Lines 381-392

**Mục đích**: Export metrics as JSON

```go
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
```

**Đánh giá**:
- ✅ Simple & clean
- ✅ Good error handling
- ✅ Pretty-printed JSON (indent)

**Khuyến nghị**:
- Keep as-is ✅

---

### 1️⃣6️⃣ `exportPrometheus()` - Internal ✅ OK

**Vị trí**: Lines 395-443

**Mục đích**: Export metrics as Prometheus format

**Độ dài**: 50 lines

```go
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
```

**Đánh giá**:
- ✅ Complete Prometheus format
- ✅ Good metric naming (crew_* prefix)
- ✅ Includes HELP & TYPE comments
- ⚠️ String concatenation (could use strings.Builder for efficiency)
- ⚠️ No tool metrics export (defined but not exported)

**Optimization**:
```go
// Current: String concatenation (inefficient)
result += fmt.Sprintf("...")
result += fmt.Sprintf("...")

// Better: Use strings.Builder
var builder strings.Builder
builder.WriteString("# HELP crew_requests_total...\n")
// ...
return builder.String(), nil
```

**Khuyến nghị**:
- Refactor string building to use strings.Builder
- Add tool metrics export (currently missing)

---

### 1️⃣7️⃣ `Reset()` - Reset ✅ OK

**Vị trí**: Lines 446-454

**Mục đích**: Reset all metrics (useful for testing)

```go
func (mc *MetricsCollector) Reset() {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    mc.systemMetrics = &SystemMetrics{
        StartTime:    time.Now(),
        AgentMetrics: make(map[string]*AgentMetrics),
    }
}
```

**Đánh giá**:
- ✅ Simple & clean
- ✅ Updates StartTime (fresh start)
- ✅ Useful for testing

**Khuyến nghị**:
- Keep as-is ✅

---

### 1️⃣8️⃣ `Enable()` - Control ✅ OK

**Vị trí**: Lines 457-461

**Mục đích**: Enable metrics collection

```go
func (mc *MetricsCollector) Enable() {
    mc.mu.Lock()
    defer mc.mu.Unlock()
    mc.enabled = true
}
```

**Đánh giá**:
- ✅ Simple & thread-safe

**Khuyến nghị**:
- Keep as-is ✅

---

### 1️⃣9️⃣ `Disable()` - Control ✅ OK

**Vị trí**: Lines 464-468

**Mục đích**: Disable metrics collection

```go
func (mc *MetricsCollector) Disable() {
    mc.mu.Lock()
    defer mc.mu.Unlock()
    mc.enabled = false
}
```

**Đánh giá**:
- ✅ Simple & thread-safe

**Khuyến nghị**:
- Keep as-is ✅

---

### 2️⃣0️⃣ `IsEnabled()` - Query ✅ OK

**Vị trí**: Lines 471-475

**Mục đích**: Check if metrics collection is enabled

```go
func (mc *MetricsCollector) IsEnabled() bool {
    mc.mu.RLock()
    defer mc.mu.RUnlock()
    return mc.enabled
}
```

**Đánh giá**:
- ✅ Simple & thread-safe

**Khuyến nghị**:
- Keep as-is ✅

---

### 2️⃣1️⃣ `statusString()` - Helper ✅ OK

**Vị trí**: Lines 478-483

**Mục đích**: Convert boolean to status string

```go
func statusString(success bool) string {
    if success {
        return "success"
    }
    return "error"
}
```

**Đánh giá**:
- ✅ Simple & clear
- ✅ Used by RecordToolExecution()

**Khuyến nghị**:
- Keep as-is ✅
- Could use ternary-like pattern but Go doesn't support it

---

## 📊 SUMMARY TABLE

| # | Hàm | Status | Mức Độ | Action |
|---|---|---|---|---|
| 1 | NewMetricsCollector | ✅ OK | - | Keep, remove currentExecution |
| 2 | RecordToolExecution | ❌ DEAD | CRITICAL | DELETE |
| 3 | RecordAgentExecution | ⚠️ DUPLICATE | HIGH | Consolidate |
| 4 | RecordLLMCall | ⚠️ DUPLICATE | HIGH | Consolidate |
| 5 | ResetSessionCost | ✅ OK | - | Keep |
| 6 | GetSessionCost | ✅ OK | - | Keep |
| 7 | GetTotalCost | ✅ OK | - | Keep |
| 8 | LogCrewCostSummary | ✅ OK | - | Keep |
| 9 | RecordCacheHit | ⚠️ INEFFICIENT | MEDIUM | Optimize |
| 10 | RecordCacheMiss | ⚠️ INEFFICIENT | MEDIUM | Optimize |
| 11 | updateCacheHitRate | ⚠️ INEFFICIENT | MEDIUM | Optimize |
| 12 | UpdateMemoryUsage | ⚠️ DUPLICATE | HIGH | Consolidate |
| 13 | GetSystemMetrics | ✅ OK | - | Keep |
| 14 | ExportMetrics | ✅ OK | - | Keep |
| 15 | exportJSON | ✅ OK | - | Keep |
| 16 | exportPrometheus | ⚠️ INCOMPLETE | MEDIUM | Improve |
| 17 | Reset | ✅ OK | - | Keep |
| 18 | Enable | ✅ OK | - | Keep |
| 19 | Disable | ✅ OK | - | Keep |
| 20 | IsEnabled | ✅ OK | - | Keep |
| 21 | statusString | ✅ OK | - | Keep |

---

**Generated**: 2025-12-25
**Analysis Depth**: COMPREHENSIVE

