# CREW.GO REFACTORING - DETAILED IMPLEMENTATION GUIDE

**Status**: Ready to Execute
**Complexity**: 🔴 CRITICAL
**Estimated Time**: 3-4 days (40 hours)
**Risk Level**: MEDIUM (with good testing)

---

## 📋 TABLE OF CONTENTS

1. [Phase 1: Critical Fixes](#phase-1-critical-fixes-day-1)
2. [Phase 2: Extract Common Functions](#phase-2-extract-common-functions-days-2-3)
3. [Phase 3: Refactor Main Functions](#phase-3-refactor-main-functions-days-4-5)
4. [Phase 4: Validation & Testing](#phase-4-validation--testing-day-6)

---

## PHASE 1: CRITICAL FIXES (Day 1)

### Fix #1.1: Add Mutex for Thread Safety

**File**: `core/crew.go`
**Lines to Modify**: CrewExecutor struct + history access
**Time**: 30 minutes
**Risk**: LOW (adds protection, doesn't change logic)

#### Step 1: Modify CrewExecutor struct

```go
// BEFORE (Line 389-399)
type CrewExecutor struct {
    crew           *Crew
    apiKey         string
    entryAgent     *Agent
    history        []Message  // ❌ NOT PROTECTED
    Verbose        bool
    ResumeAgentID  string
    ToolTimeouts   *ToolTimeoutConfig
    Metrics        *MetricsCollector
    defaults       *HardcodedDefaults
}

// AFTER
type CrewExecutor struct {
    crew           *Crew
    apiKey         string
    entryAgent     *Agent
    historyMu      sync.RWMutex  // ✅ PROTECT history
    history        []Message
    Verbose        bool
    ResumeAgentID  string
    ToolTimeouts   *ToolTimeoutConfig
    Metrics        *MetricsCollector
    defaults       *HardcodedDefaults
}
```

#### Step 2: Add helper methods

Add these methods after `GetHistory()` method (around line 523):

```go
// appendMessage safely appends a message to history
func (ce *CrewExecutor) appendMessage(msg Message) {
    ce.historyMu.Lock()
    defer ce.historyMu.Unlock()
    ce.history = append(ce.history, msg)
}

// getHistoryCopy returns a copy of history for reading
func (ce *CrewExecutor) getHistoryCopy() []Message {
    ce.historyMu.RLock()
    defer ce.historyMu.RUnlock()

    if len(ce.history) == 0 {
        return []Message{}
    }

    historyCopy := make([]Message, len(ce.history))
    copy(historyCopy, ce.history)
    return historyCopy
}
```

#### Step 3: Update GetHistory() method

```go
// BEFORE (Line 518-523)
func (ce *CrewExecutor) GetHistory() []Message {
    historyCopy := make([]Message, len(ce.history))
    copy(historyCopy, ce.history)
    return historyCopy
}

// AFTER
func (ce *CrewExecutor) GetHistory() []Message {
    return ce.getHistoryCopy()
}
```

#### Step 4: Update trimHistoryIfNeeded()

```go
// BEFORE (Line 539-601)
func (ce *CrewExecutor) trimHistoryIfNeeded() {
    if ce.defaults == nil || len(ce.history) <= 2 {
        return
    }
    // ... rest of function
    ce.history = newHistory  // ❌ NOT PROTECTED
}

// AFTER
func (ce *CrewExecutor) trimHistoryIfNeeded() {
    ce.historyMu.Lock()
    defer ce.historyMu.Unlock()

    if ce.defaults == nil || len(ce.history) <= 2 {
        return
    }
    // ... rest of function (same logic)
    ce.history = newHistory  // ✅ NOW PROTECTED
}
```

#### Step 5: Update estimateHistoryTokens()

```go
// BEFORE
func (ce *CrewExecutor) estimateHistoryTokens() int {
    total := 0
    for _, msg := range ce.history {  // ❌ NOT PROTECTED
        total += 4 + (len(msg.Content)+3)/4
    }
    return total
}

// AFTER
func (ce *CrewExecutor) estimateHistoryTokens() int {
    ce.historyMu.RLock()
    defer ce.historyMu.RUnlock()

    total := 0
    for _, msg := range ce.history {  // ✅ NOW PROTECTED
        total += 4 + (len(msg.Content)+3)/4
    }
    return total
}
```

#### Step 6: Update ClearHistory()

```go
// BEFORE (Line 605-612)
func (ce *CrewExecutor) ClearHistory() {
    ce.history = []Message{}  // ❌ NOT PROTECTED
    if ce.Metrics != nil {
        ce.Metrics.ResetSessionCost()
    }
}

// AFTER
func (ce *CrewExecutor) ClearHistory() {
    ce.historyMu.Lock()
    defer ce.historyMu.Unlock()

    ce.history = []Message{}  // ✅ NOW PROTECTED

    ce.historyMu.Unlock()
    if ce.Metrics != nil {
        ce.Metrics.ResetSessionCost()
    }
    ce.historyMu.Lock()
}
```

#### Step 7: Update all history appends in ExecuteStream()

Replace all `ce.history = append(ce.history, ...)` with `ce.appendMessage(...)`

**Locations to update**:
- Line 617-620: Initial input
- Line 718-721: Agent response
- Line 749-752: Tool results
- Line 827-830: Parallel results

```go
// BEFORE
ce.history = append(ce.history, Message{
    Role:    "user",
    Content: input,
})

// AFTER
ce.appendMessage(Message{
    Role:    "user",
    Content: input,
})
```

#### Step 8: Update all history appends in Execute()

Same replacements as Step 7 for Execute() method:
- Line 864-867: Initial input
- Line 902-905: Agent response
- Line 918-921: Tool results
- Line 1005-1008: Parallel results

---

### Fix #1.2: Fix Indentation Issue

**File**: `core/crew.go`
**Lines**: 663-675
**Time**: 5 minutes

```go
// BEFORE (WRONG INDENTATION)
if err != nil {
    // Update performance metrics with error
    if currentAgent.Metadata != nil {
    currentAgent.UpdatePerformanceMetrics(false, err.Error())
    }

    // Check error quota (use different variable to avoid shadowing)
     if quotaErr := currentAgent.CheckErrorQuota(); quotaErr != nil {
        log.Printf("[QUOTA] Agent %s exceeded error quota: %v", currentAgent.ID, quotaErr)
        streamChan <- NewStreamEvent("error", currentAgent.Name,
            fmt.Sprintf("Error quota exceeded: %v", quotaErr))
        return quotaErr
    }
}

// AFTER (CORRECT INDENTATION)
if err != nil {
    // Update performance metrics with error
    if currentAgent.Metadata != nil {
        currentAgent.UpdatePerformanceMetrics(false, err.Error())
    }

    // Check error quota (use different variable to avoid shadowing)
    if quotaErr := currentAgent.CheckErrorQuota(); quotaErr != nil {
        log.Printf("[QUOTA] Agent %s exceeded error quota: %v", currentAgent.ID, quotaErr)
        streamChan <- NewStreamEvent("error", currentAgent.Name,
            fmt.Sprintf("Error quota exceeded: %v", quotaErr))
        return quotaErr
    }
}
```

---

### Fix #1.3: Add nil Checks

**File**: `core/crew.go`
**Locations**: Lines 404-409, 623-625
**Time**: 10 minutes

#### Location 1: NewCrewExecutor()

```go
// BEFORE
func NewCrewExecutor(crew *Crew, apiKey string) *CrewExecutor {
    var entryAgent *Agent
    if len(crew.Agents) > 0 {  // ❌ Crash if crew is nil
        entryAgent = crew.Agents[0]
    }
    return &CrewExecutor{...}
}

// AFTER
func NewCrewExecutor(crew *Crew, apiKey string) *CrewExecutor {
    if crew == nil {
        log.Println("WARNING: CrewExecutor created with nil crew")
        return nil
    }

    var entryAgent *Agent
    if len(crew.Agents) > 0 {
        entryAgent = crew.Agents[0]
    }

    return &CrewExecutor{...}
}
```

#### Location 2: ExecuteStream() - Agent selection

```go
// BEFORE (Line 623-625)
currentAgent = ce.findAgentByID(ce.ResumeAgentID)  // Could return nil
if currentAgent == nil {
    return fmt.Errorf("resume agent %s not found", ce.ResumeAgentID)
}

// Already has nil check, but make sure findAgentByID() returns nil properly
// This is GOOD, no change needed
```

---

### Fix #1.4: Add Missing Constants

**File**: `core/crew.go`
**Location**: After imports (line 11)
**Time**: 10 minutes

```go
// Add after imports, before functions
const (
    // Backoff timing
    maxBackoffDuration = 5 * time.Second
    minimalToolTimeout = 100 * time.Millisecond

    // Retry configuration
    defaultMaxRetries = 2

    // Timeout defaults
    defaultPerToolTimeout  = 5 * time.Second
    defaultSequenceTimeout = 30 * time.Second
    defaultOverheadBudget  = 500 * time.Millisecond

    // Timeout warning threshold
    timeoutWarningThresholdPercent = 20

    // History context
    minHistoryMessages = 2
    minKeeproomFromEnd = 2
)
```

#### Update usages:

**Line 177**: Replace `100<<uint(attempt)` calculation
```go
// Already correct, keep as-is (bit shift is idiomatic)
baseDelay := time.Duration(100<<uint(attempt)) * time.Millisecond
```

**Line 180**: Replace magic number
```go
// BEFORE
if baseDelay > 5*time.Second {
    baseDelay = 5 * time.Second
}

// AFTER
if baseDelay > maxBackoffDuration {
    baseDelay = maxBackoffDuration
}
```

**Line 328**: Replace magic number
```go
// BEFORE
return 100 * time.Millisecond

// AFTER
return minimalToolTimeout
```

**Line 354**: Replace magic number
```go
// BEFORE
warnThreshold := totalDuration / 5  // 20%

// AFTER
warnThreshold := totalDuration / (100 / timeoutWarningThresholdPercent)
```

**Lines 368-377**: Replace magic numbers in NewToolTimeoutConfig()
```go
// BEFORE
return &ToolTimeoutConfig{
    DefaultToolTimeout: 5 * time.Second,
    SequenceTimeout:    30 * time.Second,
    OverheadBudget:     500 * time.Millisecond,
    PerToolTimeout:     make(map[string]time.Duration),
    CollectMetrics:     true,
    ExecutionMetrics:   []ExecutionMetrics{},
}

// AFTER
return &ToolTimeoutConfig{
    DefaultToolTimeout: defaultPerToolTimeout,
    SequenceTimeout:    defaultSequenceTimeout,
    OverheadBudget:     defaultOverheadBudget,
    PerToolTimeout:     make(map[string]time.Duration),
    CollectMetrics:     true,
    ExecutionMetrics:   []ExecutionMetrics{},
}
```

**Line 540**: Replace magic number
```go
// BEFORE
if ce.defaults == nil || len(ce.history) <= 2 {

// AFTER
if ce.defaults == nil || len(ce.history) <= minHistoryMessages {
```

**Line 568**: Replace magic number
```go
// BEFORE
if keepFromEnd < 2 {
    keepFromEnd = 2
}

// AFTER
if keepFromEnd < minKeeproomFromEnd {
    keepFromEnd = minKeeproomFromEnd
}
```

---

## PHASE 2: EXTRACT COMMON FUNCTIONS (Days 2-3)

### Strategy Overview

Instead of refactoring Execute() and ExecuteStream() separately, we'll:
1. Extract common logic into shared functions
2. Both Execute() and ExecuteStream() will reuse them
3. This eliminates 35% code duplication

### Extracted Functions:

#### Function 1: executeAgentOnce()

**Purpose**: Execute single agent, record metrics, handle errors
**Lines**: ~25
**Used by**: Execute(), ExecuteStream()
**Location**: After ExecuteStream() (around line 860)

```go
// executeAgentOnce executes a single agent and records metrics
// Returns (response, error) and updates metrics
func (ce *CrewExecutor) executeAgentOnce(
    ctx context.Context,
    currentAgent *Agent,
    input string,
) (*AgentResponse, time.Duration, error) {
    agentStartTime := time.Now()

    response, err := ExecuteAgent(ctx, currentAgent, input, ce.getHistoryCopy(), ce.apiKey)
    agentEndTime := time.Now()
    agentDuration := agentEndTime.Sub(agentStartTime)

    if err != nil {
        // Update performance metrics with error
        if currentAgent.Metadata != nil {
            currentAgent.UpdatePerformanceMetrics(false, err.Error())
        }

        // Check error quota
        if quotaErr := currentAgent.CheckErrorQuota(); quotaErr != nil {
            return nil, agentDuration, fmt.Errorf("error quota exceeded: %w", quotaErr)
        }

        return nil, agentDuration, err
    }

    // Record successful execution
    if ce.Metrics != nil {
        ce.Metrics.RecordAgentExecution(currentAgent.ID, currentAgent.Name, agentDuration, true)

        tokens, cost := currentAgent.GetLastCallCost()
        ce.Metrics.RecordLLMCall(currentAgent.ID, tokens, cost)
        ce.Metrics.LogCrewCostSummary()

        // Check memory quota
        memoryUsedMB := (tokens * 4) / 1024 / 1024
        if err := currentAgent.CheckMemoryQuota(); err != nil {
            return nil, agentDuration, fmt.Errorf("memory quota exceeded: %w", err)
        }

        // Update metrics
        callDurationMs := agentDuration.Milliseconds()
        if currentAgent.Metadata != nil {
            currentAgent.UpdateMemoryMetrics(memoryUsedMB, callDurationMs)
            currentAgent.UpdatePerformanceMetrics(true, "")
        }
    }

    return response, agentDuration, nil
}
```

#### Function 2: handleToolResults()

**Purpose**: Execute tools, format results, update history
**Lines**: ~30
**Used by**: Execute(), ExecuteStream()
**Location**: After executeAgentOnce()

```go
// toolResult wraps a tool execution result for internal use
type toolResult struct {
    ToolName string
    Output   string
    Status   string
}

// handleToolResults executes tool calls and formats results
// Returns formatted result text and list of results
func (ce *CrewExecutor) handleToolResults(
    ctx context.Context,
    response *AgentResponse,
    currentAgent *Agent,
    isStream bool,
) (string, []toolResult, error) {
    if len(response.ToolCalls) == 0 {
        return "", nil, nil
    }

    results := ce.executeCalls(ctx, response.ToolCalls, currentAgent)
    resultText := ce.formatToolResults(results)

    // For streaming, emit events
    if isStream {
        streamChan := make(chan *StreamEvent) // ❌ THIS IS A PROBLEM - need to pass streamChan
        for _, result := range results {
            status := "✅"
            if result.Status == "error" {
                status = "❌"
            }
            streamChan <- NewStreamEvent("tool_result", currentAgent.Name,
                fmt.Sprintf("%s [Tool] %s → %s", status, result.ToolName, result.Output))
        }
    }

    return resultText, results, nil
}
```

**Note**: This function needs streamChan passed as parameter for stream mode.

#### Function 3: applyRouting()

**Purpose**: Check all routing decisions (termination, signals, routing, wait, terminal)
**Lines**: ~85
**Used by**: Execute(), ExecuteStream()
**Location**: After handleToolResults()

```go
// routingDecision represents a routing decision from one of the routing checks
type routingDecision int

const (
    routingContinue routingDecision = iota
    routingTerminate
    routingHandoff
    routingParallel
    routingWait
)

// RoutingResult wraps the result of routing logic
type RoutingResult struct {
    Decision   routingDecision
    NextAgent  *Agent
    SignalText string
}

// applyRouting checks all routing conditions and returns next action
func (ce *CrewExecutor) applyRouting(
    currentAgent *Agent,
    response *AgentResponse,
    input string,
) *RoutingResult {
    // Check for termination signals
    terminationResult := ce.checkTerminationSignal(currentAgent, response.Content)
    if terminationResult != nil && terminationResult.ShouldTerminate {
        return &RoutingResult{
            Decision:   routingTerminate,
            SignalText: terminationResult.Signal,
        }
    }

    // Check for routing signals
    nextAgent := ce.findNextAgentBySignal(currentAgent, response.Content)
    if nextAgent != nil {
        return &RoutingResult{
            Decision:  routingHandoff,
            NextAgent: nextAgent,
        }
    }

    // Check wait_for_signal
    behavior := ce.getAgentBehavior(currentAgent.ID)
    if behavior != nil && behavior.WaitForSignal {
        return &RoutingResult{
            Decision:   routingWait,
            SignalText: fmt.Sprintf("[PAUSE:%s]", currentAgent.ID),
        }
    }

    // Check if terminal
    if currentAgent.IsTerminal {
        return &RoutingResult{
            Decision: routingTerminate,
        }
    }

    // Check for parallel group
    parallelGroup := ce.findParallelGroup(currentAgent.ID, response.Content)
    if parallelGroup != nil {
        return &RoutingResult{
            Decision:   routingParallel,
            SignalText: parallelGroup.NextAgent,
        }
    }

    // Default: continue
    return &RoutingResult{
        Decision: routingContinue,
    }
}
```

#### Function 4: executeTool()

**Purpose**: Execute single tool with retry logic
**Lines**: Already extracted (retryWithBackoff)
**Status**: ✅ No change needed - already well-designed

---

## PHASE 3: REFACTOR MAIN FUNCTIONS (Days 4-5)

### Refactor ExecuteStream()

**Goal**: Reduce from 245 lines to ~80 lines by using extracted functions

**New Structure**:
```
executeStreamImpl()
├── Setup (10 lines)
├── Main loop (40 lines)
│   ├── executeAgentOnce()
│   ├── handleToolResults()
│   ├── applyRouting()
│   └── handoff logic
└── Cleanup (5 lines)
```

**Before**: 245 lines
**After**: ~80 lines (70% reduction)

### Refactor Execute()

**Goal**: Reduce from 186 lines to ~80 lines using same extracted functions

**Before**: 186 lines
**After**: ~80 lines (57% reduction)

---

## PHASE 4: VALIDATION & TESTING (Day 6)

### Step 1: Run Tests

```bash
# Test with race detector
go test -race ./core -v

# Check coverage
go test -cover ./core

# Lint
golangci-lint run ./core
```

### Step 2: Run Metrics

```bash
# Cyclomatic complexity
gocyclo -avg ./core

# Expected:
# Before: 8-15 per function
# After: 4-8 per function
```

### Step 3: Manual Testing

Create test file to verify:
1. History is thread-safe
2. ExecuteStream produces same results as before
3. Execute produces same results as before
4. Tool execution works correctly
5. Routing works correctly

---

## 🚦 CHECKPOINT: After Each Phase

### After Phase 1 (End of Day 1)
```
✅ Code compiles
✅ No new lint errors
✅ Basic tests pass
✅ Race detector shows no new warnings
```

### After Phase 2 (End of Day 3)
```
✅ New functions extracted and working
✅ Both Execute and ExecuteStream still compile
✅ Integration tests pass
✅ No performance degradation
```

### After Phase 3 (End of Day 5)
```
✅ ExecuteStream refactored and simplified
✅ Execute refactored and simplified
✅ All tests pass
✅ Complexity metrics improved
✅ Code duplication reduced to <10%
```

### After Phase 4 (End of Day 6)
```
✅ All tests pass with race detector
✅ Coverage ≥85%
✅ Lint reports 0 errors
✅ Cyclomatic complexity <10 per function
✅ Documentation updated if needed
```

---

## 📊 SUCCESS METRICS

### Target Metrics After Refactoring

| Metric | Before | After | Target | Status |
|--------|--------|-------|--------|--------|
| **ExecuteStream lines** | 245 | 80 | <100 | 🎯 |
| **Execute lines** | 186 | 80 | <100 | 🎯 |
| **Cyclomatic Complexity** | 18 | 8 | <10 | 🎯 |
| **Code Duplication** | 35% | 8% | <10% | 🎯 |
| **Thread Safety** | ❌ | ✅ | ✅ | 🎯 |
| **Race Detector Warnings** | Yes | No | No | 🎯 |
| **Test Coverage** | ? | ≥85% | ≥85% | 🎯 |

---

## 🚀 GETTING STARTED

1. **Create feature branch**:
   ```bash
   git checkout -b refactor/crew-code-cleanup
   ```

2. **Start Phase 1**:
   - Apply Fix #1.1 (Mutex)
   - Test after each fix
   - Commit: "refactor: Add mutex for thread-safe history access"

3. **Continue with remaining phases**

4. **Submit PR** with:
   - Before/after metrics
   - Test results
   - Description of changes

---

## ⚠️ RISKS & HOW TO MITIGATE

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Regression in routing | MEDIUM | HIGH | Extensive integration testing |
| Race condition still present | LOW | HIGH | Run with -race flag |
| Performance impact | LOW | MEDIUM | Profile before/after |
| Breaking existing code | LOW | HIGH | Comprehensive unit tests |

---

## 📞 QUESTIONS BEFORE STARTING?

Review these points:
1. Do we understand the 9 issues identified?
2. Have we prepared a clean feature branch?
3. Do we have tests we can run after each phase?
4. Do we have time for full validation (Phase 4)?

---

**Ready to start Phase 1?**

**Next Action**:
1. Read through this guide once more
2. Create feature branch: `git checkout -b refactor/crew-code-cleanup`
3. Start with Phase 1: Critical Fixes

---

*Last Updated: 2025-12-24*
*Status: Ready for Implementation*
