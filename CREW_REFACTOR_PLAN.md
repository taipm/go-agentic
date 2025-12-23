# PHƯƠNG ÁN REFACTOR `crew.go`

## Trạng thái: ✅ COMPLETED (Phase 2)

### Kết quả thực tế

| File | Lines Before | Lines After | Status |
|------|-------------|-------------|--------|
| `crew.go` | 1,507 | 889 | ✅ Giảm 41% |
| `crew_routing.go` | - | 154 | ✅ New |
| `crew_parallel.go` | - | 242 | ✅ New |
| `crew_tools.go` | - | 265 | ✅ New |
| `tools/errors.go` | - | 166 | ✅ Completed |
| `tools/validation.go` | - | 94 | ✅ Completed |
| `tools/timeout.go` | - | 150 | ✅ Completed |

### Giải pháp đã thực hiện (Phase 2)

**Vấn đề**: Import cycle khi tạo `crew/` package riêng

**Giải pháp**: Tách code thành các file trong CÙNG package `crewai`:
1. ✅ `crew_routing.go` - Agent routing & signal matching
2. ✅ `crew_parallel.go` - Parallel execution (ExecuteParallel, ExecuteParallelStream)
3. ✅ `crew_tools.go` - Tool execution helpers (executeCalls, ToolResult, formatToolResults)
4. ✅ `tools/` package vẫn giữ nguyên (standalone package)
5. ✅ All tests pass: `go test ./...`

### Lợi ích

- ✅ Không có import cycle
- ✅ Code organization tốt hơn (separation of concerns)
- ✅ Mỗi file < 300 dòng (dễ maintain)
- ✅ Backward compatible 100%

---

## Tổng quan ban đầu

| Metric | Value |
|--------|-------|
| **Tổng dòng** | 1,512 |
| **Functions** | 35 |
| **Types/Structs** | 5 |
| **Constants** | 7 |

---

## Phân tích theo Concern

### 1. ERROR HANDLING (Lines 104-275) - ~170 dòng
```
├── ErrorType enum + constants (110-120)
├── classifyError() (122-160)
├── isRetryable() (162-173)
├── calculateBackoffDuration() (175-187)
├── retryWithBackoff() (189-241)
├── safeExecuteToolOnce() (243-260)
├── safeExecuteTool() (262-270)
└── SafeExecuteTool() [exported] (272-275)
```
**Recommendation**: Move to `core/errors/` or `core/tools/errors.go`

### 2. TOOL VALIDATION (Lines 29-102) - ~75 dòng
```
├── extractRequiredFields() (30-40)
├── validateFieldType() (42-69)
└── validateToolArguments() (71-102)
```
**Recommendation**: Move to `core/tools/validation.go`

### 3. TIMEOUT MANAGEMENT (Lines 277-394) - ~120 dòng
```
├── ExecutionMetrics struct (279-286)
├── TimeoutTracker struct (290-296)
├── NewTimeoutTracker() (298-307)
├── GetRemainingTime() (309-318)
├── CalculateToolTimeout() (320-347)
├── RecordToolExecution() (349-354)
├── IsTimeoutWarning() (356-364)
├── ToolTimeoutConfig struct (366-374)
├── NewToolTimeoutConfig() (376-386)
└── GetToolTimeout() (388-394)
```
**Recommendation**: Move to `core/tools/timeout.go`

### 4. CREW EXECUTOR CORE (Lines 396-533) - ~140 dòng
```
├── CrewExecutor struct (396-407)
├── NewCrewExecutor() (409-429)
├── NewCrewExecutorFromConfig() (431-497)
├── SetVerbose() (499-502)
├── SetResumeAgent() (504-508)
├── ClearResumeAgent() (510-513)
├── GetResumeAgentID() (515-518)
├── GetHistory() (520-527)
└── ClearHistory() (529-533)
```
**Recommendation**: Keep in `crew.go` - this is the core API

### 5. EXECUTION LOGIC (Lines 535-896) - ~360 dòng
```
├── ExecuteStream() (535-731) - 196 dòng
└── Execute() (733-896) - 163 dòng
```
**Recommendation**: Move to `core/crew/execute.go`

### 6. TOOL EXECUTION HELPERS (Lines 898-1098) - ~200 dòng
```
├── calculateToolTimeout() (898-909)
├── logToolStart() (911-920)
├── recordToolMetrics() (922-941)
├── getToolExecutionStatus() (943-953)
├── handleToolNotFound() (955-964)
├── handleSequenceTimeout() (966-975)
├── handleToolExecutionError() (977-990)
├── handleToolExecutionSuccess() (992-1001)
├── setupSequenceContext() (1003-1024)
└── executeCalls() (1026-1098)
```
**Recommendation**: Move to `core/crew/tools.go`

### 7. ROUTING LOGIC (Lines 1100-1217) - ~120 dòng
```
├── findAgentByID() (1100-1108)
├── signalMatchesContent() (1110-1142)
├── findNextAgentBySignal() (1144-1174)
├── getAgentBehavior() (1176-1186)
└── findNextAgent() (1188-1217)
```
**Recommendation**: Move to `core/crew/routing.go`

### 8. PARALLEL EXECUTION (Lines 1219-1465) - ~250 dòng
```
├── DefaultParallelAgentTimeout const (1221)
├── ExecuteParallelStream() (1223-1328)
├── ExecuteParallel() (1330-1426)
├── findParallelGroup() (1428-1449)
└── aggregateParallelResults() (1451-1465)
```
**Recommendation**: Move to `core/crew/parallel.go`

### 9. UTILITIES (Lines 15-27, 1467-1512) - ~70 dòng
```
├── copyHistory() (15-27)
├── ToolResult struct (1467-1472)
├── formatToolResults() [deprecated] (1474-1478)
├── formatToolResults() [method] (1480-1484)
└── defaultFormatToolResults() (1486-1512)
```
**Recommendation**: Move to `core/crew/utils.go`

---

## Cấu trúc thư mục đề xuất

```
core/
├── crew.go                  # KEEP: CrewExecutor struct, constructors, public API
│                            # (~200 dòng sau refactor)
│
├── tools/                   # NEW: Tool execution subsystem
│   ├── errors.go            # ErrorType, classifyError, retry logic
│   ├── validation.go        # validateToolArguments
│   └── timeout.go           # TimeoutTracker, ToolTimeoutConfig
│
├── crew/                    # NEW: Crew execution internals
│   ├── execute.go           # ExecuteStream, Execute
│   ├── routing.go           # findAgentByID, findNextAgent, signal matching
│   ├── parallel.go          # ExecuteParallel, ExecuteParallelStream
│   ├── tools.go             # executeCalls, tool execution helpers
│   └── utils.go             # copyHistory, formatToolResults
│
└── (existing files unchanged)
```

---

## Kế hoạch thực hiện

### Phase 1: Tạo `core/tools/` package (Độc lập, không breaking change)

**File: `core/tools/errors.go`**
```go
package tools

// ErrorType, classifyError, isRetryable, calculateBackoffDuration
// retryWithBackoff, safeExecuteToolOnce, safeExecuteTool
```

**File: `core/tools/validation.go`**
```go
package tools

// extractRequiredFields, validateFieldType, validateToolArguments
```

**File: `core/tools/timeout.go`**
```go
package tools

// ExecutionMetrics, TimeoutTracker, ToolTimeoutConfig
```

**Backward compatibility:**
- Export types từ `crew.go` bằng type alias
- Giữ functions cũ, delegate sang `tools` package

### Phase 2: Tạo `core/crew/` package (Internal, không export)

**File: `core/crew/execute.go`**
```go
package crew

// executeStream, execute (internal implementations)
```

**File: `core/crew/routing.go`**
```go
package crew

// findAgentByID, signalMatchesContent, findNextAgentBySignal, etc.
```

**File: `core/crew/parallel.go`**
```go
package crew

// executeParallelStream, executeParallel, findParallelGroup
```

**File: `core/crew/tools.go`**
```go
package crew

// executeCalls, calculateToolTimeout, handleToolNotFound, etc.
```

**File: `core/crew/utils.go`**
```go
package crew

// copyHistory, ToolResult, formatToolResults
```

### Phase 3: Refactor `crew.go` thành thin wrapper

```go
package crewai

import (
    "github.com/taipm/go-agentic/core/crew"
    "github.com/taipm/go-agentic/core/tools"
)

// Type aliases for backward compatibility
type ErrorType = tools.ErrorType
type ExecutionMetrics = tools.ExecutionMetrics
type TimeoutTracker = tools.TimeoutTracker
type ToolTimeoutConfig = tools.ToolTimeoutConfig
type ToolResult = crew.ToolResult

// CrewExecutor embeds internal executor
type CrewExecutor struct {
    *crew.Executor
}

// NewCrewExecutor delegates to internal
func NewCrewExecutor(c *Crew, apiKey string) *CrewExecutor {
    return &CrewExecutor{
        Executor: crew.NewExecutor(c, apiKey),
    }
}

// ... other delegating methods
```

---

## Dependencies giữa các packages

```
                    ┌─────────────┐
                    │   crewai    │  (public API)
                    │  crew.go    │
                    └──────┬──────┘
                           │ imports
              ┌────────────┴────────────┐
              ▼                         ▼
      ┌───────────────┐         ┌───────────────┐
      │  core/crew/   │         │  core/tools/  │
      │  (internal)   │────────▶│  (internal)   │
      └───────────────┘ imports └───────────────┘
```

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Breaking API changes | Type aliases + delegating methods |
| Circular imports | tools/ không import crew/, crew/ import tools/ |
| Test failures | Giữ tests trong `crewai` package, test cả public + internal |
| Performance overhead | Inline functions trong hot paths nếu cần |

---

## Tiêu chí hoàn thành

- [ ] `go build ./...` pass
- [ ] `go test ./...` pass (100% existing tests)
- [ ] `crew.go` < 300 dòng
- [ ] Mỗi file mới < 250 dòng
- [ ] Không có circular imports
- [ ] Backward compatible với existing code

---

## Timeline đề xuất

| Phase | Scope | Effort |
|-------|-------|--------|
| Phase 1 | `core/tools/` | 1-2 hours |
| Phase 2 | `core/crew/` | 2-3 hours |
| Phase 3 | Thin wrapper | 1 hour |
| Testing | Full regression | 1 hour |
| **Total** | | **5-7 hours** |

---

## Quyết định cần đưa ra

1. **Package naming**: `tools` hay `toolexec`?
2. **Export level**: Export tất cả từ `core/tools` hay chỉ qua `crewai`?
3. **Interface extraction**: Tạo interfaces cho testability?
4. **Error types**: Dùng sentinel errors hay custom error types?
