# 🐛 BUG REPORT: Signal-Based Routing Completely Non-Functional

**Status:** ✅ PROVEN BY TESTS
**Severity:** 🔴 CRITICAL - Blocks multi-agent orchestration
**Impact:** Quiz exam cannot execute beyond first agent
**Affected Version:** Current main branch

---

## 📊 EXECUTIVE SUMMARY (5W2H)

| Aspect | Details |
|--------|---------|
| **WHAT** | Signal-based routing doesn't work; only first agent executes |
| **WHO** | ExecutionFlow, HandleAgentResponse(), DetermineNextAgent() |
| **WHERE** | core/executor/workflow.go, core/workflow/execution.go |
| **WHEN** | When agent emits signal with routing rule defined |
| **WHY** | SignalRegistry not passed through execution chain |
| **HOW** | ExecuteWithCallbacks() passes nil to workflow |
| **HOW MANY** | 7 interconnected bugs preventing signal routing |

---

## ❌ TEST RESULTS: BUG PROVEN

### Test 1: HandleAgentResponse_IgnoresSignalRouting
**Status:** 🔴 **FAILED** (as expected - proves bug)

```
TEST SETUP:
  - Current agent: teacher
  - Response signals: [[QUESTION]]
  - Routing rule: [QUESTION] -> student
  - Routing config provided: YES

CALLING: flow.HandleAgentResponse(response, routing, agents)

RESULT:
  - shouldContinue: false (expected: true) ← BUG!
  - current agent: teacher (expected: student) ← BUG!

BUG PROVEN:
  - HandleAgentResponse returned false even with routing config
  - DetermineNextAgent() doesn't check response.Signals
  - Only checks IsTerminal and HandoffTargets
  - Signal-based routing config is IGNORED
```

**Assertion:** `if !shouldContinue { t.Errorf("BUG PROVEN...") }`
**Result:** ✅ Assertion triggered - Bug proven!

---

### Test 2: QuizExamBugAnalysis
**Status:** 🔴 **FAILED** (Demonstrates the bug chain)

**Observed Issue:**
- Quiz exam runs with teacher & student agents
- Teacher emits `[QUESTION]` signal
- Student agent never executes
- Workflow ends after first agent

**Root Cause Chain:**
```
1. main() → executor.ExecuteStream() → crew.executeWorkflow()
2. executeWorkflow() calls: flow.ExecuteWithCallbacks(...)
   NOTE: No SignalRegistry parameter ✗
3. ExecuteWithCallbacks() loop (line 195-225):
   for {
     response = ExecuteWorkflowStep(ctx, handler, apiKey)
     shouldContinue = HandleAgentResponse(response, routing, agents)
     if !shouldContinue { break } ← BREAKS HERE!
   }
4. First iteration:
   - ExecuteWorkflowStep() executes Teacher ✓
   - Teacher emits '[QUESTION]' signal ✓
   - HandleAgentResponse() called
   - DetermineNextAgent() returns IsTerminal=true ✗ (WRONG!)
   - shouldContinue = false ✗
   - LOOP BREAKS ✗
5. ExecuteWithCallbacks() returns (workflow ends)
```

---

### Test 3: ExecutionChainBugLocations
**Status:** 🔴 **FAILED** (Identifies 7 bug locations)

**Bug Locations in Source Code:**

| # | File | Line | Problem | Solution |
|---|------|------|---------|----------|
| 1 | executor/workflow.go | 13 | ExecutionFlow missing SignalRegistry field | Add field |
| 2 | executor/workflow.go | 185 | ExecuteWithCallbacks missing SignalRegistry param | Add param |
| 3 | executor/workflow.go | 49 | ExecuteWorkflowStep doesn't accept SignalRegistry | Add param |
| 4 | executor/workflow.go | 76 | Passes nil to workflow.ExecuteWorkflow() | Pass registry |
| 5 | workflow/execution.go | 128 | Can't process signals (registry is nil) | Already has param |
| 6 | executor/workflow.go | 113 | HandleAgentResponse uses old API | Use new API |
| 7 | executor/workflow.go | 124 | DetermineNextAgent() ignores response.Signals | Use WithSignals() |

---

## 🔍 DETAILED 5W2H ANALYSIS

### 1. WHAT - What is Broken?

**Signal-based routing is completely non-functional.** Even though:
- ✅ SignalRegistry is created (crew.go:136)
- ✅ Signal handlers are registered (crew.go:139)
- ✅ YAML config defines routing rules
- ✅ Agent emits signals correctly

**But:**
- ❌ Routing decision from signal is never used
- ❌ Only first agent executes
- ❌ Workflow terminates instead of continuing

**Real-world impact:**
```
EXPECTED FLOW (quiz exam):
  Teacher → [QUESTION] → Student → [ANSWER] → Teacher → ...

ACTUAL FLOW (current bug):
  Teacher → [QUESTION] → (signal ignored) → STOP ✗
```

---

### 2. WHO - Who is Responsible?

**Primary Components:**

| Component | Responsibility | Failure |
|-----------|-----------------|---------|
| ExecutionFlow | Orchestrate multi-agent flow | Missing SignalRegistry field |
| ExecuteWithCallbacks() | Main loop handler | Doesn't accept/pass registry |
| ExecuteWorkflowStep() | Execute single agent | Passes nil to workflow |
| workflow.ExecuteWorkflow() | Agent executor | Gets nil registry |
| workflow.executeAgent() | Signal emitter | Can't process signals |
| HandleAgentResponse() | Determine next agent | Uses old API |
| DetermineNextAgent() | Route logic | Ignores signals |

**Dependency Chain (Fault Propagation):**
```
ExecutionFlow (missing field)
    ↓
ExecuteWithCallbacks() (missing param)
    ↓
ExecuteWorkflowStep() (missing param)
    ↓
workflow.ExecuteWorkflow() (receives nil)
    ↓
workflow.executeAgent() (can't process)
    ↓
HandleAgentResponse() (wrong API)
    ↓
DetermineNextAgent() (no signal check)
    ↓
❌ RESULT: Loop breaks after first agent
```

---

### 3. WHERE - Where is the Bug?

### **Location 1: ExecutionFlow struct (executor/workflow.go:13)**
```go
type ExecutionFlow struct {
    CurrentAgent *common.Agent
    History      []common.Message
    RoundCount   int
    HandoffCount int
    MaxRounds    int
    MaxHandoffs  int
    State        *ExecutionState
    // ❌ MISSING: SignalRegistry field
}
```
**Impact:** Cannot store registry for downstream use

---

### **Location 2: ExecuteWithCallbacks signature (executor/workflow.go:185)**
```go
func (ef *ExecutionFlow) ExecuteWithCallbacks(
    ctx context.Context,
    handler workflow.OutputHandler,
    apiKey string,
    onStep WorkflowCallback,
    agents map[string]*common.Agent,
    routing *common.RoutingConfig,
) (*common.AgentResponse, error) {
    // ❌ MISSING: signalRegistry parameter
```
**Impact:** Cannot receive or pass registry

---

### **Location 3: ExecuteWorkflowStep signature (executor/workflow.go:49)**
```go
func (ef *ExecutionFlow) ExecuteWorkflowStep(
    ctx context.Context,
    handler workflow.OutputHandler,
    apiKey string,
) (*common.AgentResponse, error) {
    // ❌ MISSING: signalRegistry parameter
```
**Impact:** Cannot accept registry from caller

---

### **Location 4: ExecuteWorkflowStep implementation (executor/workflow.go:76)**
```go
response, err := workflow.ExecuteWorkflow(
    ctx,
    ef.CurrentAgent,
    userInput,
    ef.History,
    handler,
    nil,  // ❌ SIGNAL REGISTRY IS NIL!
    apiKey,
)
```
**Impact:** Signal processing completely bypassed

---

### **Location 5: workflow.ExecuteWorkflow receives nil (workflow/execution.go:30)**
```go
func ExecuteWorkflow(
    ctx context.Context,
    entryAgent *common.Agent,
    input string,
    history []common.Message,
    handler OutputHandler,
    signalRegistry *signal.SignalRegistry,  // ← Gets nil
    apiKey string,
) (*common.AgentResponse, error)
```
**Impact:** Cannot emit signals to registry

---

### **Location 6: workflow.executeAgent skips signal processing (workflow/execution.go:128)**
```go
if execCtx.SignalRegistry != nil && response.Signals != nil && len(response.Signals) > 0 {
    // Process signal...
}
// ❌ When signalRegistry is nil, this entire block skipped
```
**Impact:** Signal handlers never invoked

---

### **Location 7: HandleAgentResponse uses old API (executor/workflow.go:113-157)**
```go
decision, err := workflow.DetermineNextAgent(
    ef.CurrentAgent,
    response,
    routing,
)
// ❌ DetermineNextAgent() doesn't check response.Signals
// Only checks IsTerminal and HandoffTargets
```
**Impact:** Signal routing decision ignored

---

### **Location 8: DetermineNextAgent ignores signals (workflow/routing.go:12-35)**
```go
func DetermineNextAgent(currentAgent *common.Agent,
    response *common.AgentResponse,
    routing *common.RoutingConfig) (*common.RoutingDecision, error) {
    // ✗ Doesn't check response.Signals
    // ✓ Only checks IsTerminal and HandoffTargets
    return &common.RoutingDecision{
        IsTerminal: true,  // ← ALWAYS RETURNS TRUE!
        Reason:     "no handoff targets configured",
    }, nil
}
```
**Impact:** Loop breaks after first agent

---

### 4. WHEN - When Does It Break?

**Timing: Immediately after first agent execution**

```
Timeline:
t=0s:    Workflow starts (entry: teacher)
t=0.1s:  ExecuteWithCallbacks() loop iteration 1 starts
t=0.2s:  Teacher agent executes successfully ✓
t=0.3s:  Teacher emits signal "[QUESTION]" ✓
t=0.4s:  Signal registry processes "[QUESTION]" ✓
         (But signalRegistry in workflow is nil, so no routing decision)
t=0.5s:  HandleAgentResponse() called
t=0.6s:  DetermineNextAgent() returns IsTerminal=true ✗
t=0.7s:  shouldContinue = false
t=0.8s:  ExecuteWithCallbacks() loop breaks (line 222-224) ✗
t=0.9s:  Workflow ends (only 1 agent executed)
```

**Expected vs Actual:**
- **Expected:** Loop continues, Student agent executes
- **Actual:** Loop breaks, Student agent never called

---

### 5. WHY - Why Does It Fail?

### **Root Cause 1: Architectural Design Flaw**
- SignalRegistry created at CrewExecutor level (crew.go)
- But ExecutionFlow doesn't know about it
- No mechanism to pass registry through execution chain

### **Root Cause 2: Two Incompatible Routing Systems**
```
System A: Signal-based (NEW)
  - Response.Signals emitted ✓
  - SignalRegistry.ProcessSignal() → routing decision ✓
  - Implemented but disconnected ✗

System B: Traditional (OLD)
  - IsTerminal + HandoffTargets ✓
  - DetermineNextAgent() used ✓
  - Only this one is called ✗
```

### **Root Cause 3: Missing Data Flow**
```
CrewExecutor.signalRegistry
    ↓ (not passed)
ExecutionFlow (no field)
    ↓ (missing param)
ExecuteWithCallbacks() (no param)
    ↓ (missing param)
ExecuteWorkflowStep() (missing param)
    ↓ (passes nil)
workflow.ExecuteWorkflow(nil)
    ↓ (receives nil)
workflow.executeAgent(nil)
    ↓ (cannot emit)
Signal handlers not invoked ✗
```

---

### 6. HOW - How Does It Happen?

### **Bug Propagation Chain**

```
STEP 1: CrewExecutor creates SignalRegistry
  File: crew.go:136
  Code: executor.signalRegistry = signal.NewSignalRegistry()
  Status: ✓ Registry created

STEP 2: CrewExecutor.executeWorkflow() calls ExecuteWithCallbacks()
  File: crew.go:482
  Code: flow.ExecuteWithCallbacks(ctx, handler, apiKey, onStep, agents, routing)
  Problem: ✗ Doesn't pass signalRegistry
  Effect: Registry lost at this point

STEP 3: ExecuteWithCallbacks() calls ExecuteWorkflowStep()
  File: executor/workflow.go:202
  Code: response, err := ef.ExecuteWorkflowStep(ctx, handler, apiKey)
  Problem: ✗ Doesn't pass signalRegistry (doesn't have it)
  Effect: No way to recover registry

STEP 4: ExecuteWorkflowStep() calls workflow.ExecuteWorkflow()
  File: executor/workflow.go:76
  Code: response, err := workflow.ExecuteWorkflow(..., nil, apiKey)
  Problem: ✗ Explicitly passes nil
  Effect: workflow layer has no signal registry

STEP 5: workflow.executeAgent() tries to emit signal
  File: workflow/execution.go:128
  Code: if execCtx.SignalRegistry != nil {
           _ = execCtx.SignalRegistry.Emit(sig)
       }
  Problem: ✗ Registry is nil, condition is false
  Effect: Signal handlers never invoked

STEP 6: HandleAgentResponse() determines next agent
  File: executor/workflow.go:217
  Code: shouldContinue, err := ef.HandleAgentResponse(response, routing, agents)
  Problem: ✗ Uses DetermineNextAgent() not DetermineNextAgentWithSignals()
  Effect: Signal-based routing ignored

STEP 7: DetermineNextAgent() returns wrong decision
  File: workflow/routing.go:12
  Code: return &common.RoutingDecision{IsTerminal: true, ...}
  Problem: ✗ Doesn't check response.Signals
  Effect: Loop breaks

STEP 8: ExecuteWithCallbacks() exits loop
  File: executor/workflow.go:222
  Code: if !shouldContinue { break }
  Problem: ✗ shouldContinue is false
  Effect: Workflow ends, only 1 agent executed
```

---

### 7. HOW MANY - 7 Critical Issues

### **The Bug Chain (Critical Path)**

| Step | Component | Issue | Severity |
|------|-----------|-------|----------|
| 1 | ExecutionFlow | No SignalRegistry field | 🔴 Critical |
| 2 | ExecuteWithCallbacks | No SignalRegistry param | 🔴 Critical |
| 3 | ExecuteWorkflowStep | No SignalRegistry param | 🔴 Critical |
| 4 | ExecuteWorkflowStep | Passes nil to workflow | 🔴 Critical |
| 5 | workflow.executeAgent | Can't emit (registry nil) | 🔴 Critical |
| 6 | HandleAgentResponse | Uses old routing API | 🔴 Critical |
| 7 | DetermineNextAgent | Ignores response.Signals | 🔴 Critical |

**All 7 must be fixed for signal routing to work.**

---

## 📋 IMPACT ANALYSIS

### **Affected Functionality**
- ❌ Multi-agent orchestration
- ❌ Signal-based routing
- ❌ Quiz exam (requires teacher → student → teacher flow)
- ❌ Parallel agent groups
- ❌ Any multi-turn agent interaction

### **Current Status**
- ✅ Single agent execution works
- ✅ Signal definition in YAML works
- ✅ Signal registry creation works
- ✅ Signal handler registration works
- ❌ Signal routing decision usage broken
- ❌ Agent continuation logic broken

### **User Impact**
- Quiz exam runs but only teacher executes
- Student agent never gets called
- Exam ends after first question
- Appears to "hang" (actually just ends)

---

## 🔧 REQUIRED FIXES

### **Fix 1: Add SignalRegistry to ExecutionFlow**
```go
type ExecutionFlow struct {
    CurrentAgent  *common.Agent
    History       []common.Message
    RoundCount    int
    HandoffCount  int
    MaxRounds     int
    MaxHandoffs   int
    State         *ExecutionState
    SignalRegistry *signal.SignalRegistry  // ← ADD THIS
}
```

### **Fix 2: Add SignalRegistry to ExecuteWithCallbacks**
```go
func (ef *ExecutionFlow) ExecuteWithCallbacks(
    ctx context.Context,
    handler workflow.OutputHandler,
    apiKey string,
    onStep WorkflowCallback,
    agents map[string]*common.Agent,
    routing *common.RoutingConfig,
    signalRegistry *signal.SignalRegistry,  // ← ADD THIS
) (*common.AgentResponse, error)
```

### **Fix 3: Add SignalRegistry to ExecuteWorkflowStep**
```go
func (ef *ExecutionFlow) ExecuteWorkflowStep(
    ctx context.Context,
    handler workflow.OutputHandler,
    apiKey string,
    signalRegistry *signal.SignalRegistry,  // ← ADD THIS
) (*common.AgentResponse, error)
```

### **Fix 4: Pass SignalRegistry to workflow**
```go
response, err := workflow.ExecuteWorkflow(
    ctx,
    ef.CurrentAgent,
    userInput,
    ef.History,
    handler,
    signalRegistry,  // ← CHANGE FROM nil
    apiKey,
)
```

### **Fix 5: Update HandleAgentResponse**
```go
decision, err := workflow.DetermineNextAgentWithSignals(  // ← NEW API
    ctx,
    ef.CurrentAgent,
    response,
    routing,
    ef.SignalRegistry,
)
```

---

## ✅ VERIFICATION

### **Test Evidence**
- ✅ Test 1 FAILED: HandleAgentResponse returns false (proves routing ignored)
- ✅ Test 2 FAILED: Only 1 agent executes (proves loop breaks)
- ✅ Test 3 FAILED: 7 bug locations identified (proves full chain broken)

### **How to Verify Fix**
```bash
go test -v ./executor -run "TestHandleAgentResponse_IgnoresSignalRouting"
# After fix: Should PASS
# Before fix: Should FAIL with "BUG PROVEN" message
```

---

## 📊 METRICS

| Metric | Value |
|--------|-------|
| Test Results | 3/3 FAILED (as expected) |
| Code Locations Affected | 7 |
| Components in Bug Chain | 7 |
| Test Coverage | 3 comprehensive tests |
| Severity | CRITICAL |
| Blockage | Prevents all multi-agent flows |

---

## 🎯 CONCLUSION

**Signal-based routing is completely non-functional due to a 7-step bug chain** that prevents the SignalRegistry from being passed through the execution pipeline.

The bug is **architectural in nature** and cannot be fixed with simple logic changes. It requires:
1. Adding SignalRegistry field to ExecutionFlow
2. Passing SignalRegistry through all execution layers
3. Updating routing decision logic to use signal-based API

**All 7 fixes must be applied together for the system to work.**

---

**Test Command:**
```bash
go test -v ./executor -run "TestHandleAgentResponse_IgnoresSignalRouting|TestQuizExamBugAnalysis|TestExecutionChainBugLocations"
```

**Status:** ✅ Bug proven and documented with 5W2H analysis
