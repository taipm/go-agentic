# 📐 Hướng Dẫn Visual Kiến Trúc Core - go-agentic

## Phần 1: Request Lifecycle Flow

### Request đơn giản (Single Agent, No Routing)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        CLIENT BROWSER / API CLIENT                      │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │ SSE: GET /api/crew/stream?q="..."
                                 │
┌────────────────────────────────▼────────────────────────────────────────┐
│                    HTTP Server (http.go)                                │
│  ┌─────────────────────────────────────────────────────────────┐        │
│  │ StreamHandler():                                            │        │
│  │  1. Parse query + history from request                      │        │
│  │  2. Validate query (UTF-8, length, etc.) [Issue #10]       │        │
│  │  3. Validate history (roles, size, etc.)                    │        │
│  │  4. Snapshot executor state (RWMutex RLock) [Issue #1]     │        │
│  │  5. Create request-scoped executor (isolated copy)         │        │
│  │  6. Create streamChan with buffer=100                       │        │
│  │  7. Launch ExecuteStream in goroutine                       │        │
│  │  8. Send events to client via SSE format                    │        │
│  └─────────────────────────────────────────────────────────────┘        │
└────────────────────────────────┬────────────────────────────────────────┘
                                 │
┌────────────────────────────────▼────────────────────────────────────────┐
│              Crew Execution Engine (crew.go)                            │
│                                                                          │
│  ExecuteStream(ctx, input, streamChan):                                 │
│  ┌──────────────────────────────────────────────────────────┐           │
│  │ [STEP 1] Determine Starting Agent                        │           │
│  │  • If ResumeAgentID set → use that agent                 │           │
│  │  • Otherwise → use entry agent (first non-terminal)      │           │
│  └─────────────────────────────────┬──────────────────────┘           │
│                                     │                                   │
│  ┌─────────────────────────────────▼──────────────────────────┐        │
│  │ [STEP 2] MAIN EXECUTION LOOP (while handoffCount < max)   │        │
│  │                                                             │        │
│  │  ┌──────────────────────────────────────────────────────┐  │        │
│  │  │ [2a] Execute Agent                                   │  │        │
│  │  │  • Send "agent_start" event                          │  │        │
│  │  │  • Call ExecuteAgent(agent, input, history)          │  │        │
│  │  │  • Get AgentResponse {Content, ToolCalls[]}          │  │        │
│  │  │  • Record metrics: duration, success [Issue #14]     │  │        │
│  │  │  • Send "agent_response" event                       │  │        │
│  │  └──────────────────────────────────────────────────────┘  │        │
│  │           ▼                                                │        │
│  │  ┌──────────────────────────────────────────────────────┐  │        │
│  │  │ [2b] Tool Execution (if ToolCalls exist)             │  │        │
│  │  │                                                       │  │        │
│  │  │  For each ToolCall:                                   │  │        │
│  │  │  ├─ Send "tool_start" event                           │  │        │
│  │  │  ├─ executeCalls() [Issue #11: timeout mgmt]         │  │        │
│  │  │  │  ├─ Check sequence deadline                       │  │        │
│  │  │  │  ├─ Calculate per-tool timeout                    │  │        │
│  │  │  │  ├─ safeExecuteTool() [Issue #5: error recovery]  │  │        │
│  │  │  │  │  ├─ Validate arguments [Issue #25]             │  │        │
│  │  │  │  │  ├─ Execute tool.Handler()                     │  │        │
│  │  │  │  │  ├─ Panic recovery (defer-recover)             │  │        │
│  │  │  │  │  └─ Retry with backoff (max 2 retries)         │  │        │
│  │  │  │  └─ Record metrics (success, timeout, error)      │  │        │
│  │  │  ├─ Send "tool_result" event                          │  │        │
│  │  │  └─ Add result to history                             │  │        │
│  │  │                                                       │  │        │
│  │  │  After all tools complete:                            │  │        │
│  │  │  ├─ Format results (truncate if > 2000 chars)        │  │        │
│  │  │  ├─ Add to history                                    │  │        │
│  │  │  └─ Loop back to [2a] (agent analyzes results)       │  │        │
│  │  └──────────────────────────────────────────────────────┘  │        │
│  │           ▼                                                │        │
│  │  ┌──────────────────────────────────────────────────────┐  │        │
│  │  │ [2c] Routing Decision (Signal-Based)                 │  │        │
│  │  │                                                       │  │        │
│  │  │  Check: Does response contain routing signal?         │  │        │
│  │  │  • Look in crew.Routing.Signals[agentID]             │  │        │
│  │  │  • Signal match: signalMatchesContent() [Issue #4]   │  │        │
│  │  │  • If match found:                                    │  │        │
│  │  │    - Find target agent                                │  │        │
│  │  │    - Increment handoffCount                           │  │        │
│  │  │    - Set currentAgent = nextAgent                     │  │        │
│  │  │    - Loop back to [2a]                                │  │        │
│  │  │                                                       │  │        │
│  │  │  If no signal match → continue to [2d]               │  │        │
│  │  └──────────────────────────────────────────────────────┘  │        │
│  │           ▼                                                │        │
│  │  ┌──────────────────────────────────────────────────────┐  │        │
│  │  │ [2d] Wait-For-Signal Check (Pause Mechanism)         │  │        │
│  │  │                                                       │  │        │
│  │  │  Check: crew.Routing.AgentBehaviors[agentID]         │  │        │
│  │  │  • If WaitForSignal = true:                           │  │        │
│  │  │    - Send "pause" event [PAUSE:agentID]              │  │        │
│  │  │    - Return nil (execution pauses)                    │  │        │
│  │  │    - Client waits for user input                      │  │        │
│  │  │    - Next request sets ResumeAgentID = agentID       │  │        │
│  │  │    - Loop back to [STEP 1] (resume from this agent)  │  │        │
│  │  │                                                       │  │        │
│  │  │  If WaitForSignal = false → continue to [2e]          │  │        │
│  │  └──────────────────────────────────────────────────────┘  │        │
│  │           ▼                                                │        │
│  │  ┌──────────────────────────────────────────────────────┐  │        │
│  │  │ [2e] Terminal Check                                  │  │        │
│  │  │                                                       │  │        │
│  │  │  If agent.IsTerminal = true:                          │  │        │
│  │  │  └─ Return (execution ends)                           │  │        │
│  │  │                                                       │  │        │
│  │  │  If IsTerminal = false → continue to [2f]            │  │        │
│  │  └──────────────────────────────────────────────────────┘  │        │
│  │           ▼                                                │        │
│  │  ┌──────────────────────────────────────────────────────┐  │        │
│  │  │ [2f] Parallel Group Check                            │  │        │
│  │  │                                                       │  │        │
│  │  │  Check: crew.Routing.ParallelGroups[signal]          │  │        │
│  │  │  • If parallel group found:                           │  │        │
│  │  │    - Launch all agents in parallel (goroutines)      │  │        │
│  │  │    - Wait for all to complete                         │  │        │
│  │  │    - Aggregate results                                │  │        │
│  │  │    - Move to NextAgent in group                       │  │        │
│  │  │    - Loop back to [2a]                                │  │        │
│  │  │                                                       │  │        │
│  │  │  If no parallel group → continue to [2g]             │  │        │
│  │  └──────────────────────────────────────────────────────┘  │        │
│  │           ▼                                                │        │
│  │  ┌──────────────────────────────────────────────────────┐  │        │
│  │  │ [2g] Normal Handoff                                  │  │        │
│  │  │                                                       │  │        │
│  │  │  Increment handoffCount                               │  │        │
│  │  │  If handoffCount >= MaxHandoffs:                      │  │        │
│  │  │  └─ Return (max handoffs reached)                     │  │        │
│  │  │                                                       │  │        │
│  │  │  Find next agent (findNextAgent):                     │  │        │
│  │  │  • Check handoff_targets from config                  │  │        │
│  │  │  • Fallback: find any non-current agent               │  │        │
│  │  │                                                       │  │        │
│  │  │  If next agent found:                                 │  │        │
│  │  │  ├─ Set currentAgent = nextAgent                      │  │        │
│  │  │  ├─ Set input = response.Content                      │  │        │
│  │  │  └─ Loop back to [2a]                                 │  │        │
│  │  │                                                       │  │        │
│  │  │  If no next agent:                                    │  │        │
│  │  │  └─ Return (end of crew)                              │  │        │
│  │  └──────────────────────────────────────────────────────┘  │        │
│  └──────────────────────────────────────────────────────────┘        │
│                                                                        │
└────────────────────────────────┬───────────────────────────────────────┘
                                 │
                    ExecuteStream returns (or sends
                    final event before returning)
                                 │
┌────────────────────────────────▼───────────────────────────────────────┐
│                   HTTP Handler Event Loop (Lines 253-283)              │
│                                                                         │
│  while true:                                                            │
│    select:                                                              │
│    ├─ case event := <-streamChan:                                     │
│    │  └─ Send to client (SSE format: "data: {...}\n\n")              │
│    ├─ case <-time.After(30s):          (keep-alive)                  │
│    │  └─ Send ping event                                              │
│    └─ case <-ctx.Done():               (client disconnect)            │
│       └─ Close connection                                              │
│                                                                         │
│  On streamChan close:                                                   │
│  ├─ execErr is safely available (channel close synchronization)       │
│  ├─ Send "done" or "error" event                                      │
│  └─ Return (HTTP connection closes)                                    │
│                                                                         │
└────────────────────────────────┬───────────────────────────────────────┘
                                 │
┌────────────────────────────────▼───────────────────────────────────────┐
│                          CLIENT RECEIVES EVENTS                        │
│  • Rendered in browser as they arrive                                   │
│  • Real-time progress feedback                                          │
│  • On pause event: User provides input → new request                    │
└────────────────────────────────────────────────────────────────────────┘
```

---

## Phần 2: Tool Execution with Timeout Management (Issue #11)

```
executeCalls(ctx, calls, agent):
│
├─ [STEP 1] Setup Sequence Context
│  ├─ Default timeout: 30 seconds (sequence level)
│  ├─ Create TimeoutTracker
│  │  ├─ sequenceStartTime = now
│  │  ├─ sequenceDeadline = now + 30s
│  │  └─ overheadBudget = 500ms
│  └─ sequenceCtx, cancel = context.WithTimeout(ctx, 30s)
│
├─ [STEP 2] For each ToolCall
│  │
│  ├─ [A] Check Sequence Deadline (fail-fast)
│  │  └─ select { case <-sequenceCtx.Done(): return timeout }
│  │
│  ├─ [B] Calculate Per-Tool Timeout
│  │  │
│  │  ├─ Get default per-tool timeout: 5 seconds
│  │  ├─ Get remaining time until sequence deadline
│  │  ├─ Formula:
│  │  │  remaining = time.Until(sequenceDeadline)
│  │  │  available = remaining - overheadBudget
│  │  │  toolTimeout = min(perToolTimeout, available)
│  │  │
│  │  └─ Result: tool gets min(5s, available_time)
│  │     Prevents one tool from starving others
│  │
│  ├─ [C] Execute Tool with Timeout
│  │  │
│  │  ├─ toolCtx, toolCancel = context.WithTimeout(sequenceCtx, toolTimeout)
│  │  ├─ startTime = now
│  │  │
│  │  ├─ output, err = safeExecuteTool(toolCtx, tool, args)
│  │  │  ├─ WRAPPER: panic recovery
│  │  │  ├─ Argument validation
│  │  │  ├─ Retry logic (max 2 retries on transient errors)
│  │  │  │  ├─ Attempt 1: execute
│  │  │  │  ├─ If error → classify (transient? permanent?)
│  │  │  │  ├─ If transient → wait exponential backoff
│  │  │  │  │  ├─ Backoff = min(100ms * 2^attempt, 5s)
│  │  │  │  │  └─ Check context not cancelled
│  │  │  │  ├─ Attempt 2: execute
│  │  │  │  └─ If still error → return
│  │  │  └─ Return output or error
│  │  │
│  │  ├─ endTime = now
│  │  ├─ duration = endTime - startTime
│  │  ├─ toolCancel()
│  │  │
│  │  └─ Detect timeout:
│  │     timedOut = errors.Is(err, context.DeadlineExceeded)
│  │
│  ├─ [D] Update Timeout Tracker
│  │  └─ tracker.RecordToolExecution(duration)
│  │
│  ├─ [E] Record Metrics
│  │  ├─ Duration
│  │  ├─ Status: success | timeout | error
│  │  └─ Execution time
│  │
│  ├─ [F] Check Timeout Warning
│  │  └─ If remaining < 20% of total → log warning
│  │
│  └─ [G] Collect Result
│     └─ results.append(ToolResult{name, status, output})
│
└─ [STEP 3] Return Collected Results
   (if sequence timeout hit mid-loop, returns early with results collected so far)
```

### Timeout Calculation Example

```
Scenario:
  • SequenceTimeout: 30 seconds
  • DefaultPerToolTimeout: 5 seconds
  • OverheadBudget: 500ms (for LLM calls between tools)

Timeline:
  T=0s    Execute Tool1 (timeout: 5s)
  T=2s    Tool1 completes

  T=2s    Execute Tool2
          Remaining: 30s - 2s = 28s
          Available: 28s - 0.5s = 27.5s
          Tool2 timeout: min(5s, 27.5s) = 5s
  T=5s    Tool2 completes (after 3s)

  T=5s    Execute Tool3
          Remaining: 30s - 5s = 25s
          Available: 25s - 0.5s = 24.5s
          Tool3 timeout: min(5s, 24.5s) = 5s
  T=7s    Tool3 completes (after 2s)

  T=7s    Execute Tool4
          Remaining: 30s - 7s = 23s
          Available: 23s - 0.5s = 22.5s
          Tool4 timeout: min(5s, 22.5s) = 5s

  T=10s   Tool4 completes (after 3s)

  T=10s   All tools executed (10s used, 20s remaining)
          Agent can analyze and respond

Stress Case:
  T=0s    Tool1: timeout=5s (available: 29.5s)
  T=5s    Tool1 completes
  T=5s    Tool2: timeout=5s (available: 24.5s)
  T=10s   Tool2 completes
  T=10s   Tool3: timeout=5s (available: 19.5s)
  T=15s   Tool3 completes
  T=15s   Tool4: timeout=5s (available: 14.5s)
  T=20s   Tool4 completes
  T=20s   Tool5: timeout=5s (available: 9.5s)
  T=25s   Tool5 completes
  T=25s   Tool6: timeout=min(5s, 4.5s) = 4.5s (⚠️ Reduced!)
  T=29s   Tool6 completes
  T=30s   Sequence deadline reached
```

---

## Phần 3: Error Recovery Flow (Issue #5)

```
safeExecuteTool(ctx, tool, args):
│
└─ retryWithBackoff(ctx, tool, args, maxRetries=2):
   │
   ├─ [Attempt 0]
   │  ├─ safeExecuteToolOnce(ctx, tool, args)
   │  │  ├─ defer recover() { if panic → convert to error }
   │  │  ├─ validateToolArguments(tool, args)
   │  │  └─ tool.Handler(ctx, args)
   │  └─ Result: output or error
   │
   ├─ If success → return output
   │
   ├─ If error:
   │  ├─ classifyError(error):
   │  │  ├─ Is it context.DeadlineExceeded? → ErrorTypeTimeout (RETRYABLE)
   │  │  ├─ Is it "connection reset"? → ErrorTypeNetwork (RETRYABLE)
   │  │  ├─ Is it "panicked:"? → ErrorTypePanic (NON-RETRYABLE)
   │  │  ├─ Is it "required field"? → ErrorTypeValidation (NON-RETRYABLE)
   │  │  └─ Otherwise → ErrorTypeTemporary (RETRYABLE)
   │  │
   │  └─ if !isRetryable(errorType) → return error immediately
   │
   ├─ [Attempt 1]
   │  ├─ If attempt < maxRetries:
   │  │  ├─ Calculate backoff:
   │  │  │  baseDelay = 100ms * 2^0 = 100ms
   │  │  │  with jitter (effectively ~50-150ms)
   │  │  ├─ select {
   │  │  │  case <-ctx.Done(): return ctx.Err()
   │  │  │  case <-time.After(100ms): continue
   │  │  │ }
   │  │  ├─ safeExecuteToolOnce(ctx, tool, args)
   │  │  └─ Result: output or error
   │  │
   │  ├─ If success → return output
   │  └─ If error → classify again
   │
   ├─ [Attempt 2]
   │  ├─ If attempt < maxRetries:
   │  │  ├─ Calculate backoff:
   │  │  │  baseDelay = 100ms * 2^1 = 200ms
   │  │  ├─ select {
   │  │  │  case <-ctx.Done(): return ctx.Err()
   │  │  │  case <-time.After(200ms): continue
   │  │  │ }
   │  │  ├─ safeExecuteToolOnce(ctx, tool, args)
   │  │  └─ Result: output or error
   │  │
   │  ├─ If success → return output
   │  └─ If error → last attempt failed
   │
   └─ [Final]
      ├─ All retries exhausted
      └─ Return lastErr
```

### Example Error Scenarios

```
Scenario 1: Network Timeout (Transient)
  Attempt 1: Tool → "network unreachable" (transient)
  Wait 100ms
  Attempt 2: Tool → Success ✅
  Return result

Scenario 2: Invalid Argument (Permanent)
  Attempt 1: Tool → "required field missing" (validation)
  → isRetryable = false
  → Return error immediately ❌
  (No retry, fail fast)

Scenario 3: Tool Panic (Permanent)
  Attempt 1: Tool → panics
  defer-recover → "tool panicked: divide by zero"
  → isRetryable = false
  → Return error immediately ❌

Scenario 4: All Transient Errors
  Attempt 1: Tool → timeout
  Wait 100ms
  Attempt 2: Tool → timeout
  Wait 200ms
  Attempt 3: Tool → timeout
  Max retries exhausted
  Return error ❌ (3 total attempts)
```

---

## Phần 4: Thread Safety & Concurrency (Issue #1, #3)

### RWMutex Pattern in HTTPHandler

```
HTTPHandler {
  executor *CrewExecutor  (shared across requests)
  mu       sync.RWMutex   (protects writes to executor fields)
  validator *InputValidator
}

CONCURRENT ACCESS PATTERN:
┌──────────────────────────────────────────────────────────────┐
│                    Multiple Goroutines                       │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Goroutine 1: StreamHandler()                                │
│  ├─ h.mu.RLock()      [Acquire read lock]                  │
│  ├─ snapshot = snapshot{Verbose, ResumeAgentID}             │
│  ├─ h.mu.RUnlock()    [Release read lock]                  │
│  └─ Proceeds with snapshot (immutable, no lock needed)       │
│                                                               │
│  Goroutine 2: StreamHandler() (another request)              │
│  ├─ h.mu.RLock()      [Also acquires read lock]            │
│  ├─ snapshot = snapshot{Verbose, ResumeAgentID}             │
│  ├─ h.mu.RUnlock()    [Release read lock]                  │
│  └─ Proceeds independently                                   │
│     (Both goroutines run concurrently! RLock allows this)   │
│                                                               │
│  Goroutine 3: SetVerbose(true) (from CLI)                    │
│  ├─ h.mu.Lock()       [Acquire exclusive write lock]        │
│  ├─ h.executor.Verbose = true                               │
│  ├─ h.mu.Unlock()     [Release write lock]                  │
│  (Goroutines 1 & 2 must wait here if they call RLock!)      │
│
└──────────────────────────────────────────────────────────────┘

WHY RWMutex (not sync.Mutex)?
  • Pattern: Many reads (StreamHandlers), few writes (SetVerbose)
  • RLock: Multiple readers can proceed concurrently
  • Lock: Only one writer can proceed (exclusive)
  • Efficiency: Read-heavy workloads much faster than mutex
```

### Goroutine Leak Prevention (Issue #3)

```
ExecuteParallel() Pattern (using errgroup):
│
├─ [OLD] Using sync.WaitGroup:
│  ├─ Problem: If one goroutine panics → context not cancelled
│  ├─ Other goroutines continue running indefinitely
│  └─ Result: Goroutine leak! ⚠️
│
└─ [NEW] Using golang.org/x/sync/errgroup:
   │
   ├─ g, gctx := errgroup.WithContext(ctx)
   │  └─ gctx automatically propagates cancellation
   │
   ├─ For each agent:
   │  └─ g.Go(func() error { ... })
   │     └─ Launch goroutine
   │
   ├─ err := g.Wait()
   │  ├─ If any goroutine returns error:
   │  │  ├─ gctx is automatically cancelled
   │  │  ├─ All other goroutines receive cancellation
   │  │  └─ All goroutines exit cleanly
   │  └─ If all succeed:
   │     └─ Return nil, all results collected
   │
   └─ GUARANTEE: No goroutine left behind ✅

Execution Timeline:
  T=0   Agent1 launches
        Agent2 launches
        Agent3 launches

  T=1   Agent1 completes ✅

  T=2   Agent2 gets error → returns error
        └─ gctx.Done() is triggered
        └─ All other goroutines receive context cancellation

  T=2.1 Agent3 checks context → gctx.Done() received
        └─ Exits gracefully

  T=2.2 g.Wait() returns with error
        └─ All goroutines cleaned up ✅
```

---

## Phần 5: Configuration & Signal-Based Routing

### Routing Signal Matching (Lines 1063-1095)

```
Signal Matching Process:
│
├─ [STEP 1] Get routing signals for current agent
│  └─ signals = crew.Routing.Signals[currentAgent.ID]
│
├─ [STEP 2] For each signal definition
│  │
│  └─ signal = {Signal: "[KẾT THÚC]", Target: "executor", Description: "..."}
│
├─ [STEP 3] Check if signal appears in response
│  │
│  ├─ Method 1: Exact match
│  │  └─ strings.Contains(response, "[KẾT THÚC]")
│  │
│  ├─ Method 2: Normalized match
│  │  ├─ Trim whitespace from signal: "[ KẾT THÚC ]" → "[KẾT THÚC]"
│  │  └─ strings.Contains(response, "[KẾT THÚC]")
│  │
│  └─ Method 3: Bracket variation match
│     ├─ Extract inner signal: "[ KẾT THÚC ]" → "KẾT THÚC"
│     ├─ Try patterns:
│     │  ├─ "[KẾT THÚC]"
│     │  ├─ "[ KẾT THÚC ]"
│     │  └─ "[  KẾT THÚC  ]"
│     └─ Check each pattern in response
│
├─ [STEP 4] If match found
│  ├─ Find target agent: agent = findAgentByID(signal.Target)
│  ├─ Log: "[ROUTING] clarifier -> executor (signal: [KẾT THÚC])"
│  └─ Return nextAgent
│
└─ [STEP 5] If no signals match
   └─ Return nil (continue to normal handoff)

Example Routing Configuration:
```yaml
routing:
  signals:
    orchestrator:  # Agent ID
      - signal: "[CLARIFY]"
        target: clarifier
        description: "Route to clarifier for clarification"
      - signal: "[READY]"
        target: executor
        description: "Ready to execute tasks"
    clarifier:
      - signal: "[KẾT THÚC]"
        target: executor
        description: "Done asking questions"
```

Why This Design?
├─ ✅ Signals can be natural language (Vietnamese)
├─ ✅ Case-insensitive + whitespace-tolerant
├─ ✅ No code changes needed for new signals
├─ ✅ Agent output format is free (just include signal)
└─ ✅ Supports multiple signals per agent
```

---

## Phần 6: Parallel Execution Pattern

```
ExecuteParallel() for ParallelGroup:
│
├─ [Setup]
│  ├─ g, gctx := errgroup.WithContext(ctx)
│  ├─ resultMap := make(map[string]*AgentResponse)
│  └─ resultMutex := sync.Mutex{}
│
├─ [Launch All Agents Concurrently]
│  │
│  ├─ For agent1:
│  │  └─ g.Go(func() error {
│  │     agentCtx, cancel := context.WithTimeout(gctx, 60s)
│  │     response, err := ExecuteAgent(agentCtx, agent1, input, history, apiKey)
│  │     resultMutex.Lock()
│  │     resultMap[agent1.ID] = response
│  │     resultMutex.Unlock()
│  │     return err
│  │    })
│  │
│  ├─ For agent2:
│  │  └─ g.Go(func() error {
│  │     agentCtx, cancel := context.WithTimeout(gctx, 60s)
│  │     response, err := ExecuteAgent(agentCtx, agent2, input, history, apiKey)
│  │     resultMutex.Lock()
│  │     resultMap[agent2.ID] = response
│  │     resultMutex.Unlock()
│  │     return err
│  │    })
│  │
│  └─ ... (all agents launched simultaneously)
│
├─ [Wait For All To Complete]
│  │
│  ├─ Timeline:
│  │  T=0s    All 3 agents start
│  │  T=1s    Agent1 completes (fast)
│  │  T=2s    Agent2 completes (normal)
│  │  T=3s    Agent3 completes (slow)
│  │  T=3s    g.Wait() returns (when last agent finishes)
│  │
│  └─ Total time: ~3s (not 1+2+3=6s sequential)
│
├─ [Collect Results]
│  │
│  └─ resultMap = {
│       "agent1": AgentResponse{Content: "..."},
│       "agent2": AgentResponse{Content: "..."},
│       "agent3": AgentResponse{Content: "..."}
│     }
│
├─ [Aggregate Results]
│  │
│  └─ aggregateParallelResults():
│     ├─ Format: "[📊 PARALLEL EXECUTION RESULTS]"
│     ├─ For each result:
│     │  └─ "[agent_id]\n{result.Content}\n"
│     └─ Format: "[END PARALLEL RESULTS]"
│
└─ [Next Agent in Pipeline]
   └─ currentAgent = parallelGroup.NextAgent (e.g., "aggregator")
      └─ Feed aggregated results to aggregator
```

---

## Phần 7: State Management Per Request

```
Request Isolation Architecture:
│
├─ [SHARED STATE] (Read-only, immutable)
│  ├─ handler.executor.crew          (crew definition)
│  │  ├─ agents list
│  │  ├─ routing config
│  │  └─ other immutable definitions
│  ├─ handler.executor.apiKey        (API key)
│  └─ handler.executor.entryAgent    (entry agent definition)
│
├─ [PER-REQUEST ISOLATED STATE]
│  │
│  ├─ Request 1 Goroutine:
│  │  └─ executor1 = CrewExecutor{
│  │     crew: handler.executor.crew           (shared ref)
│  │     apiKey: handler.executor.apiKey       (shared ref)
│  │     history: copyHistory(req1.History)    (isolated copy!)
│  │     Verbose: snapshot.Verbose             (safe copy)
│  │     ResumeAgentID: snapshot.ResumeAgentID (safe copy)
│  │    }
│  │     └─ Executes: handler.executor.ExecuteStream(input1, streamChan1)
│  │
│  ├─ Request 2 Goroutine:
│  │  └─ executor2 = CrewExecutor{
│  │     crew: handler.executor.crew           (same crew ref)
│  │     apiKey: handler.executor.apiKey       (same key ref)
│  │     history: copyHistory(req2.History)    (isolated copy!)
│  │     Verbose: snapshot.Verbose             (safe copy)
│  │     ResumeAgentID: snapshot.ResumeAgentID (safe copy)
│  │    }
│  │     └─ Executes: handler.executor.ExecuteStream(input2, streamChan2)
│  │
│  └─ KEY POINT:
│     ├─ executor1.history is separate from executor2.history
│     ├─ Changes to executor1.history don't affect executor2
│     ├─ Each request has independent conversation thread
│     └─ Perfect for concurrent requests! ✅

History Copy Operation:
│
└─ copyHistory(original []Message):
   ├─ Create new slice: copied := make([]Message, len(original))
   ├─ Copy all messages: copy(copied, original)
   └─ Return isolated copy
      └─ Now executor can modify copied without affecting original
```

---

## Phần 8: Tool Output Size Management

```
Tool Execution Output Handling:
│
├─ Tool returns output (can be very large)
│  ├─ Example: Vector search returns 10MB embeddings
│  └─ Example: File content search returns entire file
│
├─ formatToolResults() (Lines 1414-1436)
│  │
│  ├─ For each result:
│  │  ├─ Check output size: len(result.Output)
│  │  ├─ If size > 2000 characters:
│  │  │  ├─ Truncate: output[:2000]
│  │  │  └─ Append: "[⚠️ OUTPUT TRUNCATED - Original: X characters]"
│  │  └─ Add to formatted string
│  │
│  └─ Return formatted results
│
├─ WHY 2000 CHARS?
│  ├─ LLM context token limit (~2000-4000 tokens for output)
│  ├─ Prevents context overflow in agent analysis
│  ├─ Forces sampling representative parts
│  └─ Agents can still work with truncated output
│
└─ AGENT CAN STILL EXTRACT VECTORS:
   └─ If result contains "[embedding_vector: ...]"
      └─ Agent can extract and use even if output truncated
```

---

## Phần 9: Request Lifecycle with Request ID Tracking

```
Request ID Correlation (Issue #17):
│
├─ [CLIENT] Makes request
│  └─ GET /api/crew/stream?q="..."
│
├─ [SERVER] StreamHandler receives
│  ├─ GenerateRequestID() → "550e8400-e29b-41d4-a716-446655440000"
│  ├─ ShortID → "req-550e8400e29b"
│  └─ Store in context: context.WithValue(ctx, RequestIDKey, fullID)
│
├─ [LOGS] All downstream operations include request ID
│  ├─ [550e8400] [AGENT START] orchestrator
│  ├─ [550e8400] [TOOL START] GetCPUUsage
│  ├─ [550e8400] [TOOL SUCCESS] GetCPUUsage → 3.5s
│  ├─ [550e8400] [ROUTING] orchestrator -> clarifier
│  ├─ [550e8400] [AGENT START] clarifier
│  └─ [550e8400] Done
│
├─ [REQUESTMETADATA] Accumulated
│  ├─ ID: "550e8400-e29b-41d4-a716-446655440000"
│  ├─ ShortID: "req-550e8400e29b"
│  ├─ StartTime: 2025-12-22 10:30:00
│  ├─ EndTime: 2025-12-22 10:30:15
│  ├─ Duration: 15s
│  ├─ AgentCalls: 3
│  ├─ ToolCalls: 5
│  └─ Events: [{type, agent, timestamp, data}, ...]
│
└─ [MONITORING] Query logs by request ID
   └─ All operations for this request grouped together
      └─ Debugging and performance analysis! ✅
```

---

**Complete Visual Architecture Reference**

This guide covers all critical flows and architectural decisions in the `./core` module.

Last updated: 2025-12-22
