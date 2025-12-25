# 5W2H Analysis: RoutingDecision Type Duplication Issue

## Executive Summary

The `RoutingDecision` struct is defined in **two different locations** with **incompatible structures**, causing type mismatches when routing decisions are returned from signal handlers. This creates a critical incompatibility where the `Metadata` field exists in one definition but not the other.

---

## 1. WHAT - Vấn đề là gì?

### A. Definition 1: `signal.RoutingDecision` (core/signal/types.go:40-45)

```go
// Location: core/signal/types.go:40-45
type RoutingDecision struct {
    NextAgentID string
    Reason      string
    IsTerminal  bool
    Metadata    map[string]interface{}  // ← PRESENT
}
```

**Fields (4 fields):**
- `NextAgentID`: string - Agent to route to
- `Reason`: string - Why routing decision was made
- `IsTerminal`: bool - Whether execution should terminate
- `Metadata`: map[string]interface{} - Additional routing metadata

### B. Definition 2: `workflow.RoutingDecision` (core/workflow/routing.go:13-17)

```go
// Location: core/workflow/routing.go:13-17
type RoutingDecision struct {
    NextAgentID string
    Reason      string
    IsTerminal  bool
    // ← MISSING Metadata field
}
```

**Fields (3 fields):**
- `NextAgentID`: string
- `Reason`: string
- `IsTerminal`: bool

### C. The Problem

**Structural Incompatibility:**
- `signal.RoutingDecision` has 4 fields (includes Metadata)
- `workflow.RoutingDecision` has 3 fields (no Metadata)
- They are **different types** in Go (even though they have the same 3 base fields)
- **Cannot assign one type to the other** without explicit conversion
- **Cannot use them interchangeably** in function returns

---

## 2. WHY - Tại sao vấn đề này xảy ra?

### Root Cause Analysis

**1. Parallel Development Without Coordination**
- The `signal` package was developed as a signal-handling subsystem
- The `workflow` package was developed for workflow orchestration
- Both needed a "routing decision" type but didn't coordinate
- No central decision type was created in `common` package

**2. Different Requirements at Creation Time**
- **signal package needed**: Metadata for rich routing context (handler info, signal details, confidence, etc.)
- **workflow package needed**: Simple routing decision without metadata
- Different developers made independent decisions

**3. Lack of Package Structure**
- No clear "core types" package to centralize shared concepts
- `common` package exists but wasn't used for `RoutingDecision`

### Historical Timeline

```
Time 1: signal/types.go created
  └─ RoutingDecision with Metadata added

Time 2: workflow/routing.go created
  └─ RoutingDecision WITHOUT Metadata created
  └─ Developers unaware of signal/types.go definition

Time 3: workflow/execution.go integrates with signal
  └─ Type mismatch discovered when using both packages
  └─ Working around it with type conversion needed
```

---

## 3. WHERE - Ở đâu vấn đề này ảnh hưởng?

### A. File Locations

| Location | File | Lines | Status |
|----------|------|-------|--------|
| **Definition 1** | `core/signal/types.go` | 39-45 | PRIMARY definition |
| **Definition 2** | `core/workflow/routing.go` | 12-17 | DUPLICATE definition |

### B. Usage Points - Nơi nó được sử dụng/trả về

#### Signal Package Returns (`signal/types.go`)

**1. Handler.ProcessSignal()** - Line 122
```go
// Location: core/signal/handler.go:122-161
func (h *Handler) ProcessSignal(ctx context.Context, signal *Signal) (*RoutingDecision, error) {
    // ... handler logic ...

    // Returns signal.RoutingDecision WITH Metadata
    return &RoutingDecision{
        NextAgentID: handler.TargetAgent,
        Reason:      fmt.Sprintf("Signal '%s' matched handler '%s'", signal.Name, handler.Name),
        IsTerminal:  signal.Name == SignalTerminal,
        Metadata: map[string]interface{}{      // ← USES Metadata field
            "signal_name":  signal.Name,
            "handler_id":   handler.ID,
            "handler_name": handler.Name,
        },
    }, nil
}
```

**2. Handler.ProcessSignalWithPriority()** - Line 164-228
```go
// Location: core/signal/handler.go:164-228
func (h *Handler) ProcessSignalWithPriority(ctx context.Context, signal *Signal,
    priority []string) (*RoutingDecision, error) {
    // ...
    return &RoutingDecision{
        NextAgentID: selectedHandler.TargetAgent,
        Reason:      fmt.Sprintf("Signal '%s' matched handler '%s' (priority)", ...),
        IsTerminal:  signal.Name == SignalTerminal,
        Metadata: map[string]interface{}{      // ← USES Metadata field
            "signal_name":  signal.Name,
            "handler_id":   selectedHandler.ID,
            "priority":     minPriority,
        },
    }, nil
}
```

**3. SignalRegistry.ProcessSignal()** - Line 147
```go
// Location: core/signal/registry.go:147-156
func (sr *SignalRegistry) ProcessSignal(ctx context.Context, signal *Signal) (*RoutingDecision, error) {
    // ... delegates to handler ...
    return sr.handler.ProcessSignal(ctx, signal)  // Returns signal.RoutingDecision
}
```

**4. SignalRegistry.ProcessSignalWithPriority()** - Line 159
```go
// Location: core/signal/registry.go:159-168
func (sr *SignalRegistry) ProcessSignalWithPriority(ctx context.Context, signal *Signal,
    priority []string) (*RoutingDecision, error) {
    return sr.handler.ProcessSignalWithPriority(ctx, signal, priority)  // Returns signal.RoutingDecision
}
```

#### Workflow Package Returns (`workflow/routing.go`)

**1. DetermineNextAgent()** - Line 20
```go
// Location: core/workflow/routing.go:20-42
func DetermineNextAgent(currentAgent *common.Agent, response *common.AgentResponse,
    routing *common.RoutingConfig) (*RoutingDecision, error) {
    // ... logic ...

    // Returns workflow.RoutingDecision (no Metadata)
    return &RoutingDecision{
        IsTerminal: true,
        Reason:     "agent is marked as terminal",
    }, nil
}
```

**2. DetermineNextAgentWithSignals()** - Line 45
```go
// Location: core/workflow/routing.go:45-95
func DetermineNextAgentWithSignals(ctx context.Context, currentAgent *common.Agent,
    response *common.AgentResponse, routing *common.RoutingConfig,
    signalRegistry *signal.SignalRegistry) (*RoutingDecision, error) {

    // Line 60-62: Receives signal.RoutingDecision from registry
    decision, err := signalRegistry.ProcessSignal(ctx, sig)
    if err == nil && decision != nil {
        // But then creates workflow.RoutingDecision to return!
        return &RoutingDecision{
            NextAgentID: decision.NextAgentID,
            Reason:      decision.Reason,
            IsTerminal:  decision.IsTerminal,
            // ← Cannot include decision.Metadata (field doesn't exist in this struct)
        }, nil
    }

    // Returns workflow.RoutingDecision (no Metadata)
    return &RoutingDecision{
        IsTerminal: true,
        Reason:     "no handoff targets configured",
    }, nil
}
```

#### Critical Integration Point

**workflow/execution.go:127-154** - The point where types conflict:

```go
// Location: core/workflow/execution.go:127-154
// SIGNAL 4: Process custom signals from agent response
var routingDecision *signal.RoutingDecision      // ← Uses signal.RoutingDecision
if execCtx.SignalRegistry != nil && response.Signals != nil && len(response.Signals) > 0 {
    for _, sigName := range response.Signals {
        sig := &signal.Signal{
            Name:    sigName,
            AgentID: execCtx.CurrentAgent.ID,
        }

        // Emit signal
        _ = execCtx.SignalRegistry.Emit(sig)

        // Process signal for routing decision
        decision, err := execCtx.SignalRegistry.ProcessSignal(ctx, sig)  // Returns signal.RoutingDecision
        if err == nil && decision != nil {
            routingDecision = decision     // ← Assigns signal.RoutingDecision

            // If terminal signal, stop execution
            if decision.IsTerminal {
                return response, nil
            }

            // Found routing decision, stop processing signals
            if decision.NextAgentID != "" {
                break
            }
        }
    }
}

// Line 162: Uses routingDecision.Metadata (if accessed)
if routingDecision != nil && routingDecision.NextAgentID != "" {
    // Use routingDecision - includes Metadata from signal package
    // This works ONLY because we're using signal.RoutingDecision here
}
```

### C. Impact Summary

| Component | Returns | Problem |
|-----------|---------|---------|
| `signal.Handler` | `signal.RoutingDecision` | Has Metadata ✓ |
| `signal.SignalRegistry` | `signal.RoutingDecision` | Has Metadata ✓ |
| `workflow.DetermineNextAgent()` | `workflow.RoutingDecision` | NO Metadata ✗ |
| `workflow.DetermineNextAgentWithSignals()` | `workflow.RoutingDecision` | Loses Metadata from signal decision |
| `workflow.execution.go` | Uses `signal.RoutingDecision` | Works with signal package |

---

## 4. WHO - Ai chịu ảnh hưởng?

### A. Developers/Components Affected

| Component | Impact | Severity |
|-----------|--------|----------|
| **signal/handler.go** | Creates signal.RoutingDecision with Metadata | - |
| **signal/registry.go** | Returns signal.RoutingDecision with Metadata | - |
| **workflow/routing.go** | Creates workflow.RoutingDecision without Metadata | HIGH |
| **workflow/execution.go** | Receives signal.RoutingDecision, works correctly | - |
| **Future code** | Cannot use both types interchangeably | HIGH |
| **Type conversions** | Need explicit conversion functions | MODERATE |

### B. Developer Friction Points

1. **When developing in signal package**: Easy to use RoutingDecision with full Metadata
2. **When developing in workflow package**: Creates incomplete RoutingDecision
3. **When integrating packages**: Type mismatch errors at compile time
4. **When accessing Metadata**: Only available from signal.RoutingDecision
5. **When maintaining code**: Two definitions to keep in sync (but incompatible!)

---

## 5. HOW - Làm thế nào vấn đề này phát sinh và ảnh hưởng?

### A. Technical Type System Impact

**Go's Type Strictness:**
```go
// In Go, these are DIFFERENT types, not the same
type A struct {
    X string
    Y bool
    Z map[string]interface{}
}

type B struct {
    X string
    Y bool
}

var a *A = &A{X: "test", Y: true, Z: map[string]interface{}{}}
var b *B = a  // ✗ COMPILE ERROR - cannot assign *A to *B (different types!)

// Even though they share the same first 2 fields:
// Cannot assign one to the other
```

### B. Function Signature Incompatibility

**Cannot Create Unified Interfaces:**
```go
// Cannot create a single interface that works for both:
interface {
    GetRoutingDecision() *RoutingDecision  // WHICH RoutingDecision?
}

// Callers don't know if they get:
// - signal.RoutingDecision (with Metadata)
// - workflow.RoutingDecision (without Metadata)
```

### C. Data Loss Scenarios

**Scenario 1: Signal routing in DetermineNextAgentWithSignals()**
```go
// Receives signal.RoutingDecision WITH Metadata
decision, err := signalRegistry.ProcessSignal(ctx, sig)

// But must return workflow.RoutingDecision WITHOUT Metadata
return &RoutingDecision{
    NextAgentID: decision.NextAgentID,
    Reason:      decision.Reason,
    IsTerminal:  decision.IsTerminal,
    // ✗ Cannot access decision.Metadata - field doesn't exist in this struct!
}, nil
```

**Result**: Metadata is **discarded** → Lost routing context information

### D. Current Workaround

**In workflow/execution.go:127**, the code handles this by:
1. Using `signal.RoutingDecision` type directly (not workflow version)
2. Avoiding calls to `DetermineNextAgentWithSignals()` which has the incompatibility
3. Works but creates inconsistency - some functions use signal types, others use workflow types

---

## 6. WHEN - Khi nào vấn đề này xảy ra?

### A. Timeline of Issue Creation

| Phase | Event | File | Status |
|-------|-------|------|--------|
| **Refactoring Phase 1** | `signal` package created | signal/types.go | Added RoutingDecision with Metadata |
| **Refactoring Phase 2** | `workflow` package created | workflow/routing.go | Added RoutingDecision WITHOUT Metadata |
| **Refactoring Phase 3** | Integration attempted | workflow/execution.go | Type mismatch discovered |
| **Current** | Coexisting incompatibly | Both files | Duplicate definitions |

### B. When It Causes Problems (Runtime/Compile)

**Compile-Time Issues:**
- Cannot assign `signal.RoutingDecision` to `workflow.RoutingDecision` variable
- Cannot use in function calls expecting different type
- Type checkers flag the mismatch

**Runtime Issues:**
- Loss of metadata when converting between types
- Inconsistent structure expectations
- Confusion in code reviews

### C. Downstream Impact Timing

```
User calls ExecuteWorkflow()
    ↓
ExecuteWorkflow calls executeAgent()
    ↓
executeAgent() calls signalRegistry.ProcessSignal()
    ↓
ProcessSignal() returns signal.RoutingDecision WITH Metadata  ← Type A
    ↓
executeAgent() stores in signal.RoutingDecision variable      ← Type A (works fine)
    ↓
But if code later tries to call DetermineNextAgentWithSignals()
    ↓
Function expects workflow.RoutingDecision              ← Type B
    ↓
    ✗ TYPE MISMATCH ERROR
```

---

## 7. HOW (Implementation) - Làm thế nào để sửa?

### Current Flawed Implementation

**Problem Pattern:**
```go
// Two packages define the same semantic concept differently

// signal/types.go (Version 1 - WITH Metadata)
type RoutingDecision struct {
    NextAgentID string
    Reason      string
    IsTerminal  bool
    Metadata    map[string]interface{}
}

// workflow/routing.go (Version 2 - WITHOUT Metadata)
type RoutingDecision struct {
    NextAgentID string
    Reason      string
    IsTerminal  bool
}
```

### Solution 1: Unify in `common` Package (RECOMMENDED)

**Step 1: Define in core/common/types.go**
```go
// Location: core/common/types.go
type RoutingDecision struct {
    NextAgentID string                 // Agent to route to
    Reason      string                 // Why routing was decided
    IsTerminal  bool                   // Whether execution ends
    Metadata    map[string]interface{} // Additional context
}
```

**Step 2: Remove duplicate from signal/types.go**
```go
// DELETE from core/signal/types.go:39-45
// Was:
// type RoutingDecision struct {
//     NextAgentID string
//     Reason      string
//     IsTerminal  bool
//     Metadata    map[string]interface{}
// }

// Import instead:
// import "github.com/taipm/go-agentic/core/common"
// Then use: common.RoutingDecision
```

**Step 3: Remove duplicate from workflow/routing.go**
```go
// DELETE from core/workflow/routing.go:12-17
// Was:
// type RoutingDecision struct {
//     NextAgentID string
//     Reason      string
//     IsTerminal  bool
// }

// Import instead:
// import "github.com/taipm/go-agentic/core/common"
// Then use: common.RoutingDecision
```

**Step 4: Update all files to use common.RoutingDecision**
- signal/handler.go → use common.RoutingDecision
- signal/registry.go → use common.RoutingDecision
- workflow/routing.go → use common.RoutingDecision
- workflow/execution.go → use common.RoutingDecision

### Changes Required by File

| File | Changes | Lines |
|------|---------|-------|
| core/common/types.go | Add RoutingDecision definition | ~10 lines |
| core/signal/types.go | Remove RoutingDecision, add import | -10 lines |
| core/signal/handler.go | Update return type in 2 functions | 2 functions |
| core/signal/registry.go | Update import/usage | 2 functions |
| core/workflow/routing.go | Remove RoutingDecision, add import, update all functions | All 4 functions |
| core/workflow/execution.go | Update type reference | 1 variable declaration |

---

## 8. Summary Table: Difference Between Two Definitions

| Aspect | signal.RoutingDecision | workflow.RoutingDecision | Common Solution |
|--------|------------------------|--------------------------|-----------------|
| **Location** | core/signal/types.go:39 | core/workflow/routing.go:12 | core/common/types.go |
| **NextAgentID** | ✓ string | ✓ string | ✓ string |
| **Reason** | ✓ string | ✓ string | ✓ string |
| **IsTerminal** | ✓ bool | ✓ bool | ✓ bool |
| **Metadata** | ✓ map[string]interface{} | ✗ MISSING | ✓ map[string]interface{} |
| **Type Compatibility** | Incompatible | Incompatible | Single unified type |
| **Package** | signal package | workflow package | common package (shared) |

---

## 9. Impact Assessment

### Breaking Changes (Minor)
- Code importing `signal.RoutingDecision` needs to change to `common.RoutingDecision`
- Code importing `workflow.RoutingDecision` needs to change to `common.RoutingDecision`

### Benefits of Consolidation
✅ Single source of truth for routing decisions
✅ All packages see consistent structure with Metadata
✅ No data loss when routing decisions flow between packages
✅ Easier to extend in future (only one place to modify)
✅ Type-safe - compiler enforces single definition
✅ Clearer domain model - routing is a core concept

### Code Examples After Fix

**Before (Broken):**
```go
// signal package returns one type
decision := signalRegistry.ProcessSignal(ctx, sig)  // signal.RoutingDecision

// workflow package returns different type
routingDecision := DetermineNextAgent(...)  // workflow.RoutingDecision

// Cannot use together - type mismatch!
```

**After (Fixed):**
```go
// Both use same type from common
decision := signalRegistry.ProcessSignal(ctx, sig)  // common.RoutingDecision
routingDecision := DetermineNextAgent(...)  // common.RoutingDecision

// Can use interchangeably - type safe!
if decision.Metadata != nil {
    // Access metadata from any routing decision
}
```

---

## 10. Verification Checklist

After consolidation, verify:

- [ ] `common/types.go` has single `RoutingDecision` definition with all 4 fields
- [ ] `signal/types.go` no longer defines `RoutingDecision`
- [ ] `workflow/routing.go` no longer defines `RoutingDecision`
- [ ] All signal package functions return `common.RoutingDecision`
- [ ] All workflow package functions return `common.RoutingDecision`
- [ ] All imports updated to `import "github.com/taipm/go-agentic/core/common"`
- [ ] No compile errors when importing both signal and workflow packages
- [ ] Tests pass with unified type
- [ ] Type compatibility verified across packages

---

## Conclusion

The `RoutingDecision` type duplication is a **critical issue** caused by parallel package development without coordination. The solution is straightforward: **consolidate both definitions into a single unified definition in the `common` package**, which will immediately resolve type incompatibilities and enable proper data flow across the signal and workflow packages.

