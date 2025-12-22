# ⚡ Core Module - Quick Reference Card

**One-page reference for navigating and understanding the ./core module architecture**

---

## 📂 File Structure

```
core/
├── types.go                  [86 lines]  - Data structures (Agent, Tool, Crew, etc.)
├── agent.go                  [469 lines] - LLM execution, tool call extraction
├── crew.go                   [1437 lines]- Orchestration, routing, timeouts
├── config.go                 [200 lines] - YAML loading, configuration
├── validation.go             [400 lines] - Config validation, circular routing
├── http.go                   [414 lines] - HTTP handler, input validation, SSE
├── streaming.go              [55 lines]  - SSE event formatting
├── metrics.go                [300 lines] - Metrics collection, observability
├── request_tracking.go       [200 lines] - Request ID tracking
├── shutdown.go               [150 lines] - Graceful shutdown
├── html_client.go            [?]         - Example HTML client
├── report.go                 [?]         - Reporting utilities
├── tests.go                  [?]         - Test helpers
└── *_test.go                 [1500 lines]- 7 test files
```

---

## 🎯 Five Critical Design Decisions

| Decision | What | Why | Where |
|----------|------|-----|-------|
| **3-Layer Timeouts** | Sequence(30s) → PerTool(5s) → Context | Prevent starvation, fail safely | crew.go:285-359 |
| **Signal Routing** | Config-driven, not hard-coded | Deploy changes without code | config.go:12-42 |
| **RWMutex Pattern** | Read-heavy for concurrent requests | No blocking for readers | http.go:126-138 |
| **Hybrid Tool Extract** | OpenAI native + fallback text parse | Support all model types | agent.go:275-356 |
| **Error Classification** | Transient vs permanent retry decision | Smart recovery without loops | crew.go:122-160 |

---

## 🔄 Execution Flow Summary

```
Request → Validate → EntryAgent → [Execute → Routing → Pause? → Terminal?] → Response
                                       ↓
                              [ToolExecution]
                                       ↓
                           [Timeout Tracking, Retry]
```

**Key Decision Points**:
1. **After agent response**: Check for tool calls
2. **Tool execution fails**: Classify error, retry if transient
3. **After tools complete**: Check for routing signal
4. **No signal**: Check for wait_for_signal (pause)
5. **No pause**: Check if terminal
6. **Not terminal**: Look for parallel group
7. **No parallel**: Normal handoff

---

## 🛡️ Thread Safety Mechanisms

| Layer | Mechanism | Protects |
|-------|-----------|----------|
| **HTTPHandler** | RWMutex | executor field access (Verbose, ResumeAgentID) |
| **Per-Request** | Isolated copy | Each request has own history (no sharing) |
| **MetricsCollector** | RWMutex | systemMetrics (read-heavy) |
| **Tool Execution** | Context-based | Goroutine cancellation via deadlines |
| **Parallel Exec** | sync.Mutex + errgroup | Result map + automatic cancellation |

---

## ⏱️ Timeout Strategy (3 Layers)

```
REQUEST CONTEXT (from HTTP)
└─ Dies if client disconnects

SEQUENCE TIMEOUT (config: 30s default)
├─ Total time for all tools in one request
└─ Lines: crew.go:294-302, 958-977

PER-TOOL TIMEOUT (config: 5s default)
├─ Calculated: min(perToolTimeout, remainingSequenceTime - overhead)
└─ Lines: crew.go:317-342, 1013-1015

Example: 30s sequence, 5s per-tool overhead 500ms
  Tool 1: 5s (remaining: 24.5s) ✓
  Tool 2: 5s (remaining: 19.5s) ✓
  Tool 3: min(5s, 14.5s) = 4.5s (⚠️ Reduced!)
```

---

## 🔧 Tool Execution Wrapper (safeExecuteTool)

```
INPUT: tool, args, timeout
  ↓
[VALIDATION] validateToolArguments()
  ↓
[PANIC RECOVERY] defer-recover()
  ↓
[RETRY LOGIC] retryWithBackoff(maxRetries=2)
  │
  ├─ Attempt 1: Execute tool
  │  If error: Classify (transient vs permanent)
  │  If permanent: Return error immediately
  │  If transient: Continue to next attempt
  │
  ├─ Wait 100ms (exponential backoff)
  │
  ├─ Attempt 2: Execute tool
  │  If error: Classify
  │  If permanent: Return error
  │  If transient: Continue
  │
  ├─ Wait 200ms
  │
  └─ Attempt 3: Execute tool
     Return result (final attempt)
  ↓
OUTPUT: result, error
```

---

## 🎯 Routing Configuration (YAML)

```yaml
routing:
  # Signal-based handoffs
  signals:
    orchestrator:     # From agent ID
      - signal: "[CLARIFY]"
        target: clarifier
      - signal: "[READY]"
        target: executor

  # Per-agent behavior
  agent_behaviors:
    clarifier:
      wait_for_signal: true    # Pause here
      is_terminal: false

  # Parallel execution groups
  parallel_groups:
    search_team:
      agents: ["faq_searcher", "knowledge_searcher"]
      next_agent: aggregator
```

---

## 📊 Metrics Collected

```
Per-Agent:
  • ExecutionCount, SuccessCount, ErrorCount, TimeoutCount
  • AverageDuration, MinDuration, MaxDuration

Per-Tool (within agent):
  • ExecutionCount, SuccessCount, ErrorCount
  • AverageDuration, TotalDuration

System-Wide:
  • TotalRequests, SuccessRate
  • AverageRequestTime
  • MemoryUsage, CacheHitRate

Export Formats:
  • JSON: /metrics?format=json
  • Prometheus: /metrics?format=prometheus
```

---

## 🔐 Input Validation

```
Query Validation:
  ✓ Length: 1-10,000 chars
  ✓ UTF-8 valid
  ✓ No null bytes
  ✓ No control chars (except \n, \t)

History Validation:
  ✓ Max 1,000 messages
  ✓ Roles: {user, assistant, system} only
  ✓ Per-message max 100KB
  ✓ UTF-8 valid

AgentID Validation:
  ✓ Not empty
  ✓ Pattern: [a-zA-Z0-9_-]{1-128}

Lines: http.go:24-114
```

---

## 🚨 Error Types & Recovery

```
Transient Errors (RETRY UP TO 2 TIMES):
  • ErrorTypeTimeout     → context.DeadlineExceeded
  • ErrorTypeNetwork     → "connection reset", "host unreachable"
  • ErrorTypeTemporary   → Unknown errors (assume transient)

Non-Transient Errors (FAIL IMMEDIATELY):
  • ErrorTypePanic       → Tool panicked
  • ErrorTypeValidation  → "required field missing"
  • ErrorTypePermanent   → Marked explicitly

Retry Strategy:
  Backoff: 100ms, 200ms (capped at 5s)
  Max attempts: 3 (initial + 2 retries)
  Lines: crew.go:189-270
```

---

## 🎯 Key Functions by Concern

| Concern | Function | Lines | Purpose |
|---------|----------|-------|---------|
| **Agent Execution** | ExecuteAgent() | agent.go:87 | Call LLM, get response |
| **Tool Extraction** | extractFromOpenAIToolCalls() | agent.go:275 | Parse OpenAI native format |
| **Tool Execution** | executeCalls() | crew.go:982 | Execute with timeouts/retries |
| **Routing** | findNextAgentBySignal() | crew.go:1098 | Signal-based handoff |
| **Main Loop** | ExecuteStream() | crew.go:489 | Orchestration logic |
| **Validation** | validateToolArguments() | crew.go:73 | Argument type checking |
| **Config Load** | LoadCrewConfig() | config.go:79 | Load YAML + validate |
| **Metrics** | RecordAgentExecution() | metrics.go:? | Track performance |

---

## 🔍 Debugging Checklist

**Issue: Request hangs**
- [ ] Check sequence timeout (30s default): ToolTimeoutConfig.SequenceTimeout
- [ ] Check per-tool timeout (5s default): ToolTimeoutConfig.PerToolTimeout
- [ ] Review tool logs for timeout classification
- [ ] Increase timeout in config if expected to be slow

**Issue: Tool fails, doesn't retry**
- [ ] Check error type: Is it being classified correctly?
- [ ] Verify error message: Should match transient patterns
- [ ] Check if validation error: Those don't retry (fail fast)
- [ ] Review retry logic: classifyError() lines crew.go:124

**Issue: Concurrent requests interfere**
- [ ] Verify RWMutex is protecting shared state
- [ ] Check executor snapshot is created per request
- [ ] Verify history is deep-copied (not shared reference)
- [ ] Lines: http.go:205-225

**Issue: Tool panics crash system**
- [ ] safeExecuteToolOnce() should catch panic
- [ ] Check defer-recover is present (crew.go:244-251)
- [ ] Panic should convert to error, not crash

**Issue: Signal routing not working**
- [ ] Verify signal in crew.yaml routing section
- [ ] Check signalMatchesContent() handles variations
- [ ] Test with log output: "[ROUTING] ..." messages
- [ ] Lines: crew.go:1065-1095

---

## 📋 Configuration Checklist

Before deployment:
- [ ] crew.yaml defined with all agents
- [ ] agents/*.yaml files created (one per agent)
- [ ] routing.signals section configured
- [ ] All signal targets exist in agents list
- [ ] No circular routing detected (validator runs)
- [ ] Timeout values reasonable for tools
- [ ] Tools implemented and registered

Monitoring setup:
- [ ] Request ID logging enabled
- [ ] Metrics exported to Prometheus
- [ ] Alerts on timeout warnings (>80% of deadline)
- [ ] Alerts on tool execution errors
- [ ] Dashboards for agent/tool metrics

---

## 🚀 Common Operations

**Start Server**:
```go
executor, _ := NewCrewExecutorFromConfig(apiKey, "config", tools)
StartHTTPServer(executor, 8080)
```

**Handle Pause/Resume**:
```go
// Client receives: event.Type="pause", content="[PAUSE:agentID]"
// Extract agentID from content
// Next request sets: handler.SetResumeAgent(agentID)
```

**Monitor Performance**:
```go
// GET /metrics?format=prometheus
// GET /health → {"status": "ok"}
// All logs include request ID for correlation
```

**Configure Tool Timeouts**:
```go
executor.ToolTimeouts.PerToolTimeout["slow_tool"] = 15 * time.Second
```

---

## 🧪 Testing Patterns

**Simple Flow Test**:
```go
// 1. Create crew and executor
// 2. Call Execute() with input
// 3. Verify CrewResponse (Content, IsTerminal)
// 4. Check tool calls were made
```

**Routing Test**:
```go
// 1. Agent response contains signal
// 2. Verify findNextAgentBySignal() returns correct agent
// 3. Execute should route to target
```

**Timeout Test**:
```go
// 1. Create slow tool (sleeps >5s)
// 2. Execute with default timeout
// 3. Verify error is timeout (DeadlineExceeded)
// 4. Verify metrics show timeout
```

---

## 📞 Quick Lookups

**How long does execution take?**
- Typically 10-60s
- 1-2s per agent (LLM)
- 1-5s per tool (varies)

**How many concurrent requests?**
- Unbounded in theory (one goroutine per request)
- Practical limit: ~100 concurrent (depends on tool I/O)

**What's the max message history?**
- 1,000 messages max (validation limit)
- Each message max 100KB (validation limit)

**Can I modify routing at runtime?**
- No, configuration is loaded at startup
- Requires server restart to change routing

**How do I add a new agent?**
- Create agents/{id}.yaml
- Add to crew.yaml agents list
- Optionally add routing signals
- Restart server

---

## 🎓 Key Insight

The architecture is built around **safety** and **observability**:

- **Safety**: Panic recovery, error classification, timeout boundaries
- **Observability**: Request IDs, metrics collection, structured logging
- **Flexibility**: Configuration-driven routing, per-agent customization
- **Concurrency**: RWMutex, context propagation, isolated per-request state

---

**Print this page for quick reference! 🖨️**

Last Updated: 2025-12-22
