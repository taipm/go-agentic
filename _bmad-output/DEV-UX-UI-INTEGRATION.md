# DEV UX + UI Integration Design

**Created:** 2025-12-22
**Focus:** Bridge the gap between documentation, code behavior, and visual feedback
**Goal:** Developers can see exactly what the docs describe

---

## The Problem: Gap Between Docs and Reality

### What Docs Say
```markdown
# Signal-Based Routing

"Agents emit explicit signals that determine routing"

Example:
- Orchestrator outputs "[ROUTE_EXECUTOR]"
- Framework detects signal and routes to executor agent
```

### What Developer Sees
```
UI Event Stream:
🔄 [orchestrator] Starting...
💬 [orchestrator] Here is my analysis...
🔄 [executor] Starting...  
```

### What's Missing
- **Where's the signal?** (`[ROUTE_EXECUTOR]` not visible in UI)
- **How was it matched?** (Signal matching logic hidden)
- **Why that agent?** (Decision process invisible)
- **What happened to history?** (Message accumulation silent)

---

## Current UI Architecture

### End-User View (Port 8081)
```
User Input
    ↓
Crew Execution
    ↓
Stream Events → SSE → HTML/CSS UI
    ↓
Color-Coded Event Display
```

**Shown**: Agent names, tool names, final responses
**Hidden**: Signals, history, routing logic, metrics

### Problem: One UI, Two User Types

| User Type | Needs | Current UI |
|-----------|-------|-----------|
| **End User** | "Is my problem solved?" | ✅ Perfect |
| **Developer** | "How did the system solve it?" | ❌ Missing |

---

## Proposed Solution: Layered UX

### Layer 1: End-User Mode (Current - Keep as-is)
```
Input: "My CPU is high"
Output: 
  ✅ [orchestrator] Starting...
  💬 [orchestrator] Let me diagnose this...
  🔧 [executor] Running diagnostics...
  ✅ [executor] Found: Chrome using 3GB memory
```

### Layer 2: Developer Mode (NEW - Show System Internals)
```
Input: "My CPU is high"

SIGNAL ROUTING ANALYSIS:
  ├─ Orchestrator response checked against signals:
  │  ├─ [ROUTE_CLARIFIER]: ❌ not found
  │  └─ [ROUTE_EXECUTOR]: ✅ MATCHED!
  └─ Decision: Route to executor

MESSAGE HISTORY:
  ├─ Message 1 (user): "My CPU is high"
  ├─ Message 2 (assistant): "Let me diagnose this..." [156 tokens]
  └─ History size: 2 messages, 156 tokens

TOOL EXECUTION:
  ├─ GetCPUUsage
  │  ├─ Parameters: {threshold: 80}
  │  ├─ Status: ✅ Success
  │  ├─ Duration: 234ms
  │  └─ Result: {usage: 95}
  └─ GetMemoryUsage
     ├─ Parameters: {}
     ├─ Status: ✅ Success  
     ├─ Duration: 189ms
     └─ Result: {total: 16gb, used: 12gb}

METRICS:
  ├─ Total tokens: 4,250
  ├─ Cost: $0.0064
  └─ Latency: 1.2s
```

---

## Implementation: Developer Mode UI

### 1. Toggle Button
```html
<div class="header">
  <h1>IT Support Crew</h1>
  <button id="dev-mode-toggle">Developer Mode: OFF</button>
</div>
```

### 2. Layered Display

#### End-User Stream (Always Visible)
```
🔄 orchestrator: Starting...
💬 orchestrator: Let me diagnose...
```

#### Developer Metadata (Toggle-able)
```
[Signal Analysis]
  Response text includes: "[ROUTE_EXECUTOR]"
  Pattern match against allowed_signals: ✅ MATCHED
  Target: executor

[Message History]
  Message 1 (user): "My CPU is high" [24 tokens]
  Message 2 (asst): "Let me diagnose..." [156 tokens]
  Total: 2 messages, 180 tokens
  
[Tool Details]
  GetCPUUsage(threshold=80) → 234ms → {usage: 95}
  GetMemoryUsage() → 189ms → {used: 12gb}
```

### 3. Interactive Elements

**Signal Matcher** (Click to debug):
```
Orchestrator output: "... our analysis shows CPU is high. [ROUTE_EXECUTOR]"
                                                         ↑ click here

Modal opens:
  Full response text ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ...
  Regex match check: /\[([A-Z_]+)\]/
  Found signals: ["ROUTE_EXECUTOR"]
  Allowed signals: ["ROUTE_EXECUTOR", "ROUTE_CLARIFIER", "ERROR"]
  Status: ✅ Valid signal
  Routes to: executor
```

**History Inspector** (Click to expand):
```
History (2 messages, 180 tokens)
  ├─ [0] user: "My CPU is high"
  │    └─ Tokens: 24 | Role: user
  └─ [1] assistant: "Let me diagnose this..."
       └─ Tokens: 156 | Role: assistant
       
View as: JSON | Raw | Formatted

JSON View:
[
  {role: "user", content: "My CPU is high", tokens: 24},
  {role: "assistant", content: "...", tokens: 156}
]
```

**Tool Debugger** (Click tool result):
```
GetCPUUsage result clicked:

Status: ✅ Success
Duration: 234ms
Timeout: 10s
Parameters sent: {threshold: 80}

Raw output:
  CPU: 95%
  Cores: 8
  Per-core: [92%, 87%, 93%, 95%, 88%, 90%, 91%, 94%]

Validation: ✅ Passed
  Type check: number ✅
  Range check: 0-100 ✅
```

---

## Code Changes Needed

### 1. Enrich StreamEvents with Metadata

**Before:**
```go
streamChan <- NewStreamEvent("tool_result", "executor",
    fmt.Sprintf("GetCPUUsage → CPU: 95%"))
```

**After:**
```go
event := NewStreamEvent("tool_result", "executor", 
    fmt.Sprintf("GetCPUUsage → CPU: 95%"))

if devMode {
    event.Metadata = map[string]interface{}{
        "tool_name": "GetCPUUsage",
        "tool_params": call.Arguments,
        "execution_time_ms": 234,
        "execution_status": "success",
        "tool_output": rawOutput,
        "validation_passed": true,
    }
}

streamChan <- event
```

### 2. Add Signal Metadata

```go
// When detecting signal routing
nextAgent := ce.findNextAgentBySignal(currentAgent, response.Content)

if devMode {
    // Record which signal was matched
    event := NewStreamEvent("routing_decision", currentAgent.Name, "")
    event.Metadata = map[string]interface{}{
        "response_text": response.Content,
        "signal_searched": signals,  // All signals checked
        "signal_matched": matchedSignal,
        "target_agent": nextAgent.ID,
        "match_confidence": 1.0,
    }
    streamChan <- event
}
```

### 3. Add History Metadata

```go
// After each history append
if devMode {
    event := NewStreamEvent("history_update", "system", "")
    event.Metadata = map[string]interface{}{
        "action": "append",  // append, prune, etc
        "message_role": msg.Role,
        "message_length": len(msg.Content),
        "estimated_tokens": estimateTokens(msg),
        "total_messages": len(ce.history),
        "total_tokens": estimateHistoryTokens(ce.history),
    }
    streamChan <- event
}
```

### 4. Expose `/debug` Endpoints

```go
// In http.go
router.HandleFunc("/debug/signals", func(w http.ResponseWriter, r *http.Request) {
    signals := map[string]interface{}{
        "crew_signals": h.executor.crew.SignalSchema,
        "agent_signals": map[string][]string{...},
    }
    json.NewEncoder(w).Encode(signals)
})

router.HandleFunc("/debug/history", func(w http.ResponseWriter, r *http.Request) {
    history := map[string]interface{}{
        "messages": h.executor.history,
        "message_count": len(h.executor.history),
        "estimated_tokens": estimateTokens(h.executor.history),
        "max_messages": h.executor.MaxMessagesPerRequest,
    }
    json.NewEncoder(w).Encode(history)
})

router.HandleFunc("/debug/tools", func(w http.ResponseWriter, r *http.Request) {
    tools := map[string]interface{}{}
    for agentID, agent := range h.executor.crew.Agents {
        tools[agentID] = agent.Tools  // List of tools
    }
    json.NewEncoder(w).Encode(tools)
})

router.HandleFunc("/debug/config", func(w http.ResponseWriter, r *http.Request) {
    config := map[string]interface{}{
        "crew": h.executor.crew,
        "routing": h.executor.crew.Routing,
        "settings": h.executor.crew.Settings,
    }
    json.NewEncoder(w).Encode(config)
})
```

### 5. UI Component: Developer Panel

```html
<div id="dev-panel" class="hidden">
  <div class="tabs">
    <button class="tab-button active" data-tab="signals">Signals</button>
    <button class="tab-button" data-tab="history">History</button>
    <button class="tab-button" data-tab="tools">Tools</button>
    <button class="tab-button" data-tab="metrics">Metrics</button>
  </div>

  <div id="signals-tab" class="tab-content">
    <h3>Signal Routing Analysis</h3>
    <div id="signal-details"></div>
  </div>

  <div id="history-tab" class="tab-content hidden">
    <h3>Message History</h3>
    <div id="history-details"></div>
  </div>

  <div id="tools-tab" class="tab-content hidden">
    <h3>Tool Execution Details</h3>
    <div id="tools-details"></div>
  </div>

  <div id="metrics-tab" class="tab-content hidden">
    <h3>Performance Metrics</h3>
    <div id="metrics-details"></div>
  </div>
</div>

<style>
#dev-panel {
  border-top: 2px solid #333;
  padding: 20px;
  background: #f5f5f5;
  font-family: monospace;
  font-size: 12px;
}

#dev-panel.hidden {
  display: none;
}

.tabs {
  display: flex;
  gap: 10px;
  border-bottom: 1px solid #ccc;
  margin-bottom: 15px;
}

.tab-button {
  padding: 5px 10px;
  border: none;
  background: #ddd;
  cursor: pointer;
}

.tab-button.active {
  background: #333;
  color: white;
}

.tab-content {
  display: none;
}

.tab-content.active {
  display: block;
}
</style>
```

---

## Bridging Docs and UI

### Before: Docs → Code Gap

```
Documentation says:
  "Signal-based routing determines agent handoffs"
  
Developer reads: ✅
  "Yes, [ROUTE_EXECUTOR] pattern"
  
Developer runs: ❌
  "I see agent changed but not HOW"
  
Developer fixes: 
  1. Read source code
  2. Add console logs
  3. Re-run
```

### After: Docs → UI Alignment

```
Documentation says:
  "Signal-based routing determines agent handoffs"
  
Developer reads: ✅
  "Yes, [ROUTE_EXECUTOR] pattern"
  
Developer runs: ✅
  "I SEE signal matched: [ROUTE_EXECUTOR] → executor"
  
Developer fixes: 
  1. Open Developer Mode
  2. See signal matching live
  3. No code changes needed
```

### Documentation Sections Enabled by UI

| Docs Section | Before | After |
|--------------|--------|-------|
| Signal Routing | Must read source code | Live debug in UI |
| Message History | Must read logs | Inspect in history tab |
| Tool Execution | Trust it worked | See params, duration, output |
| Performance | No visibility | Metrics dashboard |
| Configuration | Edit files, restart | See full config loaded |
| Error Debugging | Read exception stack | Dev panel shows context |

---

## Phase Implementation

### Phase 0: Quick Win (1-2 hours)
- [ ] Add `devMode` parameter to StreamEvent
- [ ] Enrichrich tool_result events with tool_params + execution time
- [ ] Add `/debug/history` endpoint
- [ ] Add toggle button for verbose logging

### Phase 1: Foundation (4-5 hours)
- [ ] Add signal matching metadata to events
- [ ] Add history update events
- [ ] Expose `/debug/signals`, `/debug/tools`, `/debug/config`
- [ ] Create basic Developer Panel UI with tabs

### Phase 2: Interactive (6-8 hours)
- [ ] Signal matcher modal (click to debug)
- [ ] History inspector (click to expand)
- [ ] Tool details modal
- [ ] Metrics dashboard

### Phase 3: Polish (3-4 hours)
- [ ] Keyboard shortcuts (Ctrl+D for dev mode)
- [ ] Export debug data (JSON download)
- [ ] Time-travel debugging (replay execution)
- [ ] Documentation linking (docs link from each panel)

---

## Expected Impact

### Before Developer Mode
```
Developer adds new routing:
  1. Edit orchestrator.yaml system_prompt (30 mins)
  2. Rebuild app (1 min)
  3. Run test (1 min)
  4. Routing wrong → examine source (30 mins)
  5. Fix and repeat

Total: 2+ hours for small change
```

### After Developer Mode
```
Developer adds new routing:
  1. Edit orchestrator.yaml system_prompt (30 mins)
  2. Restart app (5 secs)
  3. Toggle Developer Mode
  4. See signal matching in real-time
  5. Verify or fix immediately

Total: 30-45 minutes, zero source code reading
```

---

## Documentation Integration Points

Docs should link to UI features:

```markdown
# Guide: Signal-Based Routing

See docs/GUIDE_SIGNAL_ROUTING.md

To understand signal matching in your crew:
1. Enable Developer Mode (Ctrl+D)
2. Click "Signals" tab
3. See which signals were checked
4. See which signal matched
5. See target agent selected
```

---

## Success Criteria

✅ Developer can see signals being matched in UI
✅ Developer can inspect message history without reading code
✅ Developer can verify tool execution parameters
✅ Developer can understand routing decisions from UI alone
✅ Developer can debug issues without source code access
✅ Zero "but how did it know to do that?" questions

---

## Conclusion

**Current State**: Documentation explains complex concepts that developers can't see in action.

**Proposed State**: Documentation concepts visualized and inspectable in Developer Mode.

**Impact**: Developers understand system behavior faster, modify with confidence, debug without source code.

**Effort**: 14-20 hours to build full developer mode with interactive debugger.

**Recommendation**: Implement Phase 0 (quick win) immediately, then Phase 1 before scaling production deployments.

