# Signal-Based Routing: Quick Reference Card

## At a Glance

```
Signal = [PATTERN] in agent response
  ↓
Crew detects signal
  ↓
Route to target agent or terminate
  ↓
Continue execution
```

---

## Signal Matching (3-Level)

```
[ROUTE_EXECUTOR]
      ↓
Try Level 1: Exact match?        → [ROUTE_EXECUTOR] ✅
Try Level 2: Case-insensitive?   → [route_executor] ✅
Try Level 3: Normalized bracket? → [ ROUTE EXECUTOR ] ✅
None match?                       → Fallback routing
```

---

## YAML Config Template

```yaml
# Minimal config
routing:
  signals:
    agent_id:
      - signal: "[SIGNAL]"
        target: next_agent      # or "" to terminate

agent_behaviors:
  agent_id:
    wait_for_signal: false      # Pause for user?
    is_terminal: false          # End workflow?

parallel_groups:
  group_id:
    agents: [a, b, c]
    wait_for_all: true
    timeout_seconds: 60
    next_agent: handler
```

---

## Signal Types

| Type | Target | Example | Meaning |
|------|--------|---------|---------|
| **Route** | agent_id | `[ROUTE_X]` → agent_x | Route to specific agent |
| **Terminate** | "" | `[DONE]` → "" | End workflow |
| **Decision** | agent_id | `[TYPE_A]` → handler_a | Conditional routing |
| **Parallel** | group_id | `[ANALYZE]` → parallel_group | Parallel execution |

---

## Execution Order

```
1. Agent executes
   ↓
2. Tool calls? → Yes: Execute tools, re-run agent
   ↓
3. Termination signal? → Yes: End workflow
   ↓
4. Routing signal? → Yes: Handoff to next agent
   ↓
5. Wait-for-signal? → Yes: Pause and return
   ↓
6. Is-terminal? → Yes: End workflow
   ↓
7. Parallel group? → Yes: Parallel execution
   ↓
8. Fallback routing → Traditional agent handoff
```

---

## History Preservation

```
Agent 1 → Agent 2 → Agent 3
 [M0]      [M0,M1]   [M0,M1,M2]
  ↑         ↑         ↑
  User      M1        Both
  input     added     added

All agents see complete history ✅
No information loss ✅
```

---

## Handoff Limit

```
max_handoffs = (agents × 2-3)

Example: 5 agents → max_handoffs = 10-15

Reach limit?
  ├─ handoffCount >= max_handoffs
  └─ Workflow ends (return)
```

---

## Common Patterns

### Sequential
```yaml
a:
  - signal: "[NEXT]"
    target: b
b:
  - signal: "[NEXT]"
    target: c
c:
  - signal: "[END]"
    target: ""
```

### Decision Tree
```yaml
classifier:
  - signal: "[TYPE_A]"
    target: handler_a
  - signal: "[TYPE_B]"
    target: handler_b
```

### Parallel
```yaml
coordinator:
  - signal: "[ANALYZE]"
    target: parallel_review

parallel_groups:
  parallel_review:
    agents: [expert1, expert2]
    next_agent: synthesizer
```

### Conditional Retry
```yaml
executor:
  - signal: "[RETRY]"
    target: executor      # Back to self
  - signal: "[SUCCESS]"
    target: next_agent
  - signal: "[FAIL]"
    target: ""           # Terminate
```

---

## Agent Instructions Template

```python
"""
You are a [role].

Your job is [responsibility].

When done, emit ONE signal:
- [SIGNAL_1]: When [condition_1]
- [SIGNAL_2]: When [condition_2]
- [SIGNAL_END]: When work is complete

Example: "Work complete. [SIGNAL_1]"
"""
```

---

## Debugging Checklist

- [ ] Signal has brackets: `[SIGNAL]` not `SIGNAL`
- [ ] Signal appears in response: Check logs
- [ ] Target agent exists: Check agents list
- [ ] Target ID matches exactly: Check case-sensitivity
- [ ] max_handoffs is reasonable: (agents × 2-3)
- [ ] Routing config loaded: Check for nil routing
- [ ] Verbose mode enabled: `executor.SetVerbose(true)`

---

## Performance Notes

```
LLM call:       100-500 ms (bottleneck)
Signal matching: ~10 µs    (negligible)
Handoff:        ~1 µs      (negligible)
History size:   Grows with messages, trim automatically
```

---

## Common Mistakes

```
❌ signal: "DONE"              → No brackets
❌ target: "Agent"             → Case mismatch
❌ target: "unknown_agent"     → Doesn't exist
❌ "[ROUTE ]" in config        → Works but inefficient

✅ signal: "[DONE]"
✅ target: "agent"
✅ target: ""                  (for terminate)
✅ signal: "[ROUTE]"           (efficient)
```

---

## File Locations

| File | Purpose |
|------|---------|
| `core/config.go` | Type definitions |
| `core/crew_routing.go` | Routing logic |
| `core/crew.go` | Main execution |
| `examples/*/config/crew.yaml` | Example configs |

---

## Key Functions

```go
// Main detection functions
checkTerminationSignal(agent, response)    // target=""
findNextAgentBySignal(agent, response)     // target=agent_id
findParallelGroup(agent, signal)           // target=group_id
signalMatchesContent(signal, response)     // 3-level match
```

---

## Decision Tree

```
Signal in response?
├─ NO → Fallback routing
└─ YES
   ├─ target = ""?
   │  └─ YES → Terminate ✅
   └─ target = agent_id?
      └─ YES
         ├─ Agent exists?
         │  ├─ NO → Fallback routing ⚠️
         │  └─ YES → Route ✅
         └─ Agent doesn't exist → Fallback ⚠️
```

---

## LLM Prompt for Signals

```
Based on your analysis, emit exactly one of:
- [SIGNAL_1]: Use when ...
- [SIGNAL_2]: Use when ...
- [SIGNAL_END]: When done

Always end your response with the signal in brackets.
Example: "Analysis complete. [SIGNAL_1]"
```

---

## Testing Template

```bash
# 1. Check config
go run main.go -validate-config

# 2. Run with verbose
executor.SetVerbose(true)

# 3. Check logs for [ROUTING] entries
# Expected: [ROUTING] agent_a -> agent_b (signal: [SIGNAL])

# 4. Verify handoff count
# Should see: handoffCount increments at each routing

# 5. Check history growth
history := executor.GetHistory()
log.Printf("History size: %d messages", len(history))
```

---

## Production Checklist

- [ ] All signals documented in agent instructions
- [ ] All target agents exist in crew
- [ ] max_handoffs set appropriately
- [ ] Termination signals defined (target="")
- [ ] Error handling for missing signals
- [ ] Logging enabled for debugging
- [ ] Timeout configured for complex workflows
- [ ] Test cases cover all routing paths

---

## One-Liners

```
Add signal:    Add entry to routing.signals[agent_id]
Terminate:     Use target: "" (empty string)
Route:         Use target: "agent_id" (agent exists)
Parallel:      Use target: "group_id" + define parallel_groups
Debug:         Enable verbose, check [ROUTING] logs
Max handoffs:  Set to agents × 2-3
History:       Fully preserved, accessible to all agents
Performance:   Negligible, LLM is the bottleneck
```

---

## Links

- [Full Analysis](./SIGNAL_BASED_ROUTING_ANALYSIS.md) - Deep dive
- [Visual Diagrams](./SIGNAL_ROUTING_DIAGRAM.md) - Flowcharts
- [Implementation Guide](./SIGNAL_ROUTING_GUIDE.md) - How-to
- [FAQ](./SIGNAL_ROUTING_FAQ.md) - Q&A

