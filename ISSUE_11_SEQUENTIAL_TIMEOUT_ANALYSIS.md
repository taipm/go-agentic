# 📋 Issue #11: No Timeout for Sequential Tool Execution

**Project**: go-agentic Library
**Issue**: Sequential tool execution in ExecuteStream has no timeout protection
**File**: crew.go:484-530 (executeCalls function)
**Date**: 2025-12-22
**Status**: 🔍 **ANALYSIS IN PROGRESS**

---

## 🔴 Problem Statement

### Current Implementation

```go
// crew.go:485-530 (UNSAFE)
func (ce *CrewExecutor) executeCalls(ctx context.Context, calls []ToolCall, agent *Agent) []ToolResult {
    var results []ToolResult

    for _, call := range calls {
        // ... validation ...

        // ❌ PROBLEM: No timeout for individual tool
        // ❌ PROBLEM: No timeout for entire sequence
        output, err := safeExecuteTool(ctx, tool, call.Arguments)

        // If one tool hangs, entire agent hangs
        // If multiple tools hang sequentially, chain of hangs
    }

    return results
}

// safeExecuteTool wrapper (crew.go:443-475)
func safeExecuteTool(ctx context.Context, tool *Tool, args map[string]interface{}) (string, error) {
    // ✅ Has context but:
    // ❌ Context passed from caller
    // ❌ No internal timeout
    // ❌ If caller's context has long timeout, tool can run forever
}
```

### Contrast with Parallel Execution

```go
// crew.go:617-645 (SAFE - has timeout)
const ParallelAgentTimeout = 60 * time.Second

// Launch all agents in parallel using goroutines
for _, agent := range agents {
    go func(ag *Agent) {
        // ✅ Creates timeout context
        agentCtx, cancel := context.WithTimeout(ctx, ParallelAgentTimeout)
        defer cancel()

        // ✅ Agent execution has timeout
        response, err := ExecuteAgent(agentCtx, ag, input, ce.history, ce.apiKey)
    }
}
```

**The Problem**:
- ✅ **Parallel execution**: Has `ParallelAgentTimeout` (60 seconds)
- ❌ **Sequential execution**: NO timeout for tools
- ❌ **Inconsistent**: Parallel is protected, sequential is not

---

## 🎯 Failure Scenarios

### Scenario #1: Hanging Tool in Sequential Execution
```go
// Agent has tools: [GetStatus(), CheckHealth(), Restart()]
// Currently executing: GetStatus() → HANGS

ExecuteStream:
├─ Agent: Supervisor
│  ├─ GetStatus()        ← HANGS (no timeout!)
│  ├─ CheckHealth()      ← NEVER RUNS (blocked)
│  └─ Restart()          ← NEVER RUNS (blocked)
└─ Client: Waiting indefinitely...

// Result:
// ❌ Client connection stays open forever
// ❌ Memory leak (resources held)
// ❌ Thread/goroutine leak
// ❌ HTTP timeout (usually 30 seconds, but could be longer)
```

### Scenario #2: Chain of Slow Tools
```go
// Agent has tools: [Tool1, Tool2, Tool3]
// Each takes: 5 seconds (acceptable individually)
// Total: 15 seconds (acceptable)

// But if Tool2 hangs:
Tool1: 5 seconds ✅
Tool2: INFINITE ❌

// Result:
// ❌ Tool3 never runs
// ❌ User doesn't know where it's stuck
// ❌ Can't distinguish slow vs hanging
```

### Scenario #3: Sequential Tools with Network I/O
```go
// Parallel timeout: 60 seconds
// Scenario: 3 tools each make network calls

ExecuteStream → Agent Execution
├─ Tool1: Network call (may hang if server down)
│  ├─ Context timeout: 60s (from ExecuteAgent context)
│  └─ Tool timeout: NONE ❌
├─ Tool2: Network call
│  ├─ Context timeout: 60s (from ExecuteAgent context)
│  └─ Tool timeout: NONE ❌
└─ Tool3: Network call
   ├─ Context timeout: 60s (from ExecuteAgent context)
   └─ Tool timeout: NONE ❌

// If Tool1 hangs for 60s:
// ❌ All 60s timeout used
// ❌ Tool2, Tool3 can't run
// ❌ ExecuteAgent times out
// ❌ Agent fails
```

### Scenario #4: No Differentiation Between Slow and Hanging
```go
// Developer doesn't know if tool:
// - Is slow (working, but takes time)
// - Is hanging (stuck, waiting for external resource)
// - Is in infinite loop
// - Is deadlocked

// Without timeout:
// ❌ No way to kill hanging tool
// ❌ No way to measure tool performance
// ❌ Can't set SLA for individual tools
```

### Scenario #5: Exponential Backoff Without Timeout
```go
// Tool uses exponential backoff:
Tool.ExecuteWithRetry():
├─ Attempt 1: fails, wait 1s
├─ Attempt 2: fails, wait 2s
├─ Attempt 3: fails, wait 4s
├─ Attempt 4: fails, wait 8s
├─ ...exponential growth...
└─ Attempt N: still waiting...

// Without timeout on tool:
// ❌ Tool retry loop runs indefinitely
// ❌ Can exhaust all retries waiting forever
// ❌ No upper bound on execution time
```

---

## 📊 Timeout Strategy

### Current Timeout Model

```
HTTP Request Timeout (30-300s, depends on server)
├─ ExecuteStream (no explicit timeout)
│  └─ ExecuteAgent (from HTTP context timeout)
│     └─ executeCalls (from ExecuteAgent context)
│        └─ safeExecuteTool (from executeCalls context)
│           └─ Tool.Handler (from safeExecuteTool context)
│
PROBLEM: All layers share same timeout!
❌ One slow tool affects everything
❌ No per-tool timeout control
❌ No per-sequence timeout
```

### Recommended Timeout Model

```
HTTP Request Timeout (60s recommended)
├─ ExecuteStream Timeout (55s, leave 5s for cleanup)
│  └─ ExecuteAgent Timeout (50s)
│     ├─ Agent LLM Call (30-40s typical)
│     └─ executeCalls Timeout (10-15s for all tools)
│        └─ Per-Tool Timeout (5s default, configurable)
│           └─ Tool.Handler with explicit context timeout

BENEFITS:
✅ Each layer has own timeout
✅ Predictable behavior
✅ Can fail fast at appropriate layer
✅ Tools don't affect agent execution time significantly
```

---

## 🎯 Solutions Comparison

### Solution 1: Per-Tool Timeout Only

**Approach**: Add timeout for each individual tool execution

```go
const ToolTimeout = 5 * time.Second

func (ce *CrewExecutor) executeCalls(ctx context.Context, calls []ToolCall, agent *Agent) []ToolResult {
    var results []ToolResult

    for _, call := range calls {
        tool, ok := toolMap[call.ToolName]
        if !ok {
            // ... error handling ...
            continue
        }

        // ✅ Create per-tool timeout
        toolCtx, cancel := context.WithTimeout(ctx, ToolTimeout)
        defer cancel()

        output, err := safeExecuteTool(toolCtx, tool, call.Arguments)
        if err == context.DeadlineExceeded {
            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "timeout",
                Output:   "Tool execution timed out after 5 seconds",
            })
        } else if err != nil {
            // ... other errors ...
        }
    }

    return results
}
```

**Advantages**:
- ✅ Prevents individual tool hangs
- ✅ Simple to implement
- ✅ Clear per-tool timeout
- ✅ Easy to configure

**Disadvantages**:
- ❌ Doesn't limit total sequence time
- ❌ Multiple tools can each use full timeout
- ❌ No overall protection
- ❌ If 10 tools timeout, takes 50s total

**Breaking Changes**: NONE

---

### Solution 2: Per-Tool + Sequence Timeout

**Approach**: Add both per-tool and total sequence timeout

```go
const (
    ToolTimeout        = 5 * time.Second
    SequenceTimeout    = 30 * time.Second
)

func (ce *CrewExecutor) executeCalls(ctx context.Context, calls []ToolCall, agent *Agent) []ToolResult {
    var results []ToolResult

    // ✅ Create sequence-level timeout
    seqCtx, cancel := context.WithTimeout(ctx, SequenceTimeout)
    defer cancel()

    startTime := time.Now()

    for _, call := range calls {
        // Check if sequence timeout exceeded
        remaining := time.Until(seqCtx.Deadline())
        if remaining <= 0 {
            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "timeout",
                Output:   "Tool sequence execution time exceeded",
            })
            continue
        }

        // Use min of per-tool timeout and remaining sequence time
        toolTimeout := ToolTimeout
        if remaining < ToolTimeout {
            toolTimeout = remaining - 1*time.Second // Leave buffer
        }

        toolCtx, cancel := context.WithTimeout(seqCtx, toolTimeout)
        output, err := safeExecuteTool(toolCtx, tool, call.Arguments)
        cancel()

        // Handle timeout
        if err == context.DeadlineExceeded {
            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "timeout",
                Output:   "Tool execution timed out",
            })
        }
    }

    elapsed := time.Since(startTime)
    log.Printf("[TOOL SEQUENCE] Completed %d tools in %v", len(calls), elapsed)

    return results
}
```

**Advantages**:
- ✅ Per-tool timeout (individual protection)
- ✅ Sequence timeout (overall protection)
- ✅ Smart timeout allocation
- ✅ Prevents resource exhaustion

**Disadvantages**:
- ❌ More complex logic
- ❌ Need to calculate remaining time
- ❌ Slightly more overhead

**Breaking Changes**: NONE

---

### Solution 3: Configurable Timeouts with Metrics (Best)

**Approach**: Add configurable timeout with monitoring and per-tool settings

```go
type ToolTimeoutConfig struct {
    DefaultToolTimeout  time.Duration
    SequenceTimeout     time.Duration
    PerToolTimeout      map[string]time.Duration  // Tool-specific overrides
}

type ExecutionMetrics struct {
    ToolName       string
    Duration       time.Duration
    Status         string  // "success", "timeout", "error"
    TimedOut       bool
}

func (ce *CrewExecutor) executeCalls(ctx context.Context, calls []ToolCall, agent *Agent) []ToolResult {
    var results []ToolResult
    var metrics []ExecutionMetrics

    // Create sequence context
    seqCtx, cancel := context.WithTimeout(ctx, ce.config.SequenceTimeout)
    defer cancel()

    sequenceStart := time.Now()

    for _, call := range calls {
        // Determine timeout for this tool
        toolTimeout := ce.config.DefaultToolTimeout
        if override, exists := ce.config.PerToolTimeout[call.ToolName]; exists {
            toolTimeout = override
        }

        // Check remaining sequence time
        remaining := time.Until(seqCtx.Deadline())
        if remaining <= 0 {
            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "timeout",
                Output:   "Sequence timeout exceeded",
            })
            continue
        }

        // Use minimum of tool timeout and remaining sequence time
        if toolTimeout > remaining-500*time.Millisecond {
            toolTimeout = remaining - 500*time.Millisecond
        }

        // Execute with timeout
        toolCtx, cancel := context.WithTimeout(seqCtx, toolTimeout)
        toolStart := time.Now()

        output, err := safeExecuteTool(toolCtx, tool, call.Arguments)

        toolDuration := time.Since(toolStart)
        cancel()

        // Record metrics
        metric := ExecutionMetrics{
            ToolName: call.ToolName,
            Duration: toolDuration,
        }

        // Handle results
        if err == context.DeadlineExceeded {
            metric.Status = "timeout"
            metric.TimedOut = true
            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "timeout",
                Output:   fmt.Sprintf("Tool timed out after %v", toolDuration),
            })
            log.Printf("[TOOL TIMEOUT] %s - exceeded timeout of %v",
                call.ToolName, toolTimeout)
        } else if err != nil {
            metric.Status = "error"
            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "error",
                Output:   err.Error(),
            })
        } else {
            metric.Status = "success"
            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "success",
                Output:   output,
            })
        }

        metrics = append(metrics, metric)

        // Log slow tools (>80% of timeout)
        if toolDuration > time.Duration(float64(toolTimeout)*0.8) {
            log.Printf("[TOOL SLOW] %s took %v (near timeout %v)",
                call.ToolName, toolDuration, toolTimeout)
        }
    }

    totalDuration := time.Since(sequenceStart)
    log.Printf("[TOOL SEQUENCE] Executed %d tools in %v (timeout: %v)",
        len(calls), totalDuration, ce.config.SequenceTimeout)

    return results
}
```

**Advantages**:
- ✅ Configurable per-tool timeout
- ✅ Configurable sequence timeout
- ✅ Execution metrics
- ✅ Performance monitoring
- ✅ Slow tool detection
- ✅ Detailed logging

**Disadvantages**:
- ❌ More complex implementation
- ❌ More code to maintain
- ❌ Requires configuration management

**Breaking Changes**: NONE (all optional)

---

## 🏆 RECOMMENDATION: **Solution 3 - Configurable with Metrics**

### Why This Is Best

1. **Reliability**: Per-tool + sequence timeout = complete protection
2. **Observability**: Metrics help identify slow/hanging tools
3. **Flexibility**: Configure per tool or globally
4. **Performance**: Can optimize timeout values based on metrics
5. **Maintainability**: Clear logging for debugging

---

## 📈 Implementation Plan: Solution 3

### Step 1: Add Timeout Configuration

**File**: types.go (new section)

```go
// ToolTimeoutConfig configures timeout behavior for tool execution
type ToolTimeoutConfig struct {
    // DefaultToolTimeout is the default timeout for each tool
    DefaultToolTimeout time.Duration

    // SequenceTimeout is the total timeout for executing all tools in sequence
    SequenceTimeout time.Duration

    // PerToolTimeout allows per-tool timeout overrides
    // If a tool is in this map, its specific timeout is used instead of default
    PerToolTimeout map[string]time.Duration
}

// DefaultToolTimeoutConfig returns recommended defaults
func DefaultToolTimeoutConfig() *ToolTimeoutConfig {
    return &ToolTimeoutConfig{
        DefaultToolTimeout: 5 * time.Second,
        SequenceTimeout:    30 * time.Second,
        PerToolTimeout:     make(map[string]time.Duration),
    }
}
```

### Step 2: Add Metrics Type

**File**: types.go (add)

```go
// ExecutionMetrics tracks performance of tool execution
type ExecutionMetrics struct {
    ToolName   string
    Duration   time.Duration
    Status     string  // "success", "timeout", "error"
    TimedOut   bool
}
```

### Step 3: Add Config to CrewExecutor

**File**: crew.go (modify CrewExecutor struct)

```go
type CrewExecutor struct {
    // ... existing fields ...
    config          *ToolTimeoutConfig  // ← ADD THIS
}

// Add to NewCrewExecutor or create factory function
executor.config = DefaultToolTimeoutConfig()
```

### Step 4: Implement Timeout Logic

**File**: crew.go (modify executeCalls)

```go
// ✅ FIX for Issue #11: Add timeout protection for sequential tool execution
func (ce *CrewExecutor) executeCalls(ctx context.Context, calls []ToolCall, agent *Agent) []ToolResult {
    var results []ToolResult

    // Create sequence-level timeout context
    seqCtx, cancel := context.WithTimeout(ctx, ce.config.SequenceTimeout)
    defer cancel()

    sequenceStart := time.Now()

    for _, call := range calls {
        // Get tool-specific timeout or use default
        toolTimeout := ce.config.DefaultToolTimeout
        if override, exists := ce.config.PerToolTimeout[call.ToolName]; exists {
            toolTimeout = override
        }

        // Check remaining sequence time
        remaining := time.Until(seqCtx.Deadline())
        if remaining <= 0 {
            log.Printf("[TOOL TIMEOUT] Sequence timeout exceeded, skipping %s", call.ToolName)
            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "timeout",
                Output:   "Sequence timeout exceeded",
            })
            continue
        }

        // Use minimum to prevent individual tool from exceeding sequence
        if toolTimeout > remaining-500*time.Millisecond {
            toolTimeout = remaining - 500*time.Millisecond
        }

        // Execute tool with timeout
        toolCtx, cancel := context.WithTimeout(seqCtx, toolTimeout)
        toolStart := time.Now()

        log.Printf("[TOOL START] %s <- %s (timeout: %v)", tool.Name, agent.ID, toolTimeout)
        output, err := safeExecuteTool(toolCtx, tool, call.Arguments)
        cancel()

        toolDuration := time.Since(toolStart)

        // Handle timeout
        if err == context.DeadlineExceeded {
            log.Printf("[TOOL TIMEOUT] %s exceeded %v timeout after %v",
                call.ToolName, toolTimeout, toolDuration)
            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "timeout",
                Output:   fmt.Sprintf("Timeout after %v", toolDuration),
            })
        } else if err != nil {
            log.Printf("[TOOL ERROR] %s - %v", tool.Name, err)
            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "error",
                Output:   err.Error(),
            })
        } else {
            log.Printf("[TOOL SUCCESS] %s -> %d chars (duration: %v)",
                tool.Name, len(output), toolDuration)

            // Warn if tool is slow (>80% of timeout)
            if toolDuration > time.Duration(float64(toolTimeout)*0.8) {
                log.Printf("[TOOL SLOW] %s took %v (near timeout %v)",
                    call.ToolName, toolDuration, toolTimeout)
            }

            results = append(results, ToolResult{
                ToolName: call.ToolName,
                Status:   "success",
                Output:   output,
            })
        }
    }

    totalDuration := time.Since(sequenceStart)
    log.Printf("[TOOL SEQUENCE] Executed %d tools in %v (timeout: %v)",
        len(calls), totalDuration, ce.config.SequenceTimeout)

    return results
}
```

### Step 5: Add Tests

**Tests to add**:
- `TestExecuteCallsWithTimeout` - Tool timeout triggers correctly
- `TestExecuteCallsSequenceTimeout` - Sequence timeout prevents running all tools
- `TestExecuteCallsSlowTool` - Slow tool detection logging
- `TestToolTimeoutConfigPerTool` - Per-tool timeout override
- `TestExecuteCallsMetricsCollection` - Metrics properly recorded

---

## ✅ Benefits

### Reliability
- ✅ No hanging tools (per-tool timeout)
- ✅ No exhausted sequence time (sequence timeout)
- ✅ Fail-fast on timeout
- ✅ Clear timeout messages

### Operations
- ✅ Monitor slow tools
- ✅ Collect execution metrics
- ✅ Tune timeout values
- ✅ Detailed logging

### Performance
- ✅ Predictable timeout behavior
- ✅ Resource cleanup on timeout
- ✅ No context leaks

---

## 📊 Break Changes Analysis

### Changes Required

| Component | Change | Type | Impact |
|-----------|--------|------|--------|
| CrewExecutor | Add config field | Internal | None |
| executeCalls | Add timeout logic | Internal | None |
| Error handling | Add timeout status | Output change | None (new status) |
| Logging | Add timeout logs | Internal | None |

### Breaking Changes Count: **ZERO** ✅

- No API signature changes
- New status field "timeout" doesn't break existing code
- All configuration optional with sensible defaults

---

*Generated: 2025-12-22*
*Status*: 🔍 **ANALYSIS COMPLETE - AWAITING APPROVAL**
*Recommendation*: ✅ **Proceed with Solution 3 (Configurable + Metrics)**
