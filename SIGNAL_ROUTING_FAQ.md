# Signal-Based Routing: FAQ & Quick Reference

## Core Concepts

### Q: What exactly is a "signal"?
**A:** A signal is a string pattern within brackets that an LLM agent emits in its response to trigger routing decisions.

```
Agent response: "I have reviewed the code. [ROUTE_EXECUTOR]"
                                           ↑ SIGNAL ↑
```

**Key characteristics:**
- Format: `[PATTERN]` (must have brackets)
- Case-insensitive matching: `[ROUTE]` matches `[route]` or `[Route]`
- Whitespace-flexible: `[ ROUTE ]` matches `[ROUTE]` (normalized internally)
- User-defined (not built-in) - you define them in `crew.yaml`

---

### Q: How does signal matching work?
**A:** Three-level matching (failsafe design):

```
Level 1 (Fast): Exact string match
  "[ROUTE]" in response? → YES: use it

Level 2 (Medium): Case-insensitive match
  "[route]" matches "[ROUTE]"? → YES: use it

Level 3 (Flexible): Normalized bracket match
  "[ ROUTE ]" normalized to "[route]" matches "[route]"? → YES: use it

Result: NO MATCH → Fall back to traditional routing
```

**In practice:** All three levels run automatically, so signal matching is very robust.

---

### Q: What's the difference between termination and routing?
**A:** Both use signals, but with different targets:

```yaml
Termination (end workflow):
  signal: "[DONE]"
  target: ""              # ← Empty string = TERMINATE

Routing (hand off to agent):
  signal: "[ROUTE_X]"
  target: agent_x         # ← Agent ID = HANDOFF
```

**Checking order (in ExecuteStream):**
1. **First check:** Termination signals (target="")
2. **Then check:** Routing signals (target=agent_id)
3. **Prevents:** Agent from being routed if it emits termination

---

### Q: Do agents see the full conversation history?
**A:** YES! 100% - No information is lost.

```
Agent 1 (Teacher) executes
  History: [initial_context]
  Response: "[QUESTION] What is 2+2?"

Agent 2 (Student) executes
  History: [initial_context, teacher's question]
         ↑ Full history including Teacher's response ↑
  Response: "[ANSWER] 4"

Agent 3 (Reporter) executes
  History: [initial_context, teacher_q, student_a, marking_results]
         ↑ Can see everything ↑
```

**How:** Each agent receives `ce.history` (full message history) + current input.

---

## Configuration

### Q: What's the minimum crew.yaml config for signal routing?
**A:**
```yaml
version: "1.0"
name: my-crew
entry_point: agent_1         # ← Start here

agents:
  - agent_1
  - agent_2

routing:                       # ← Routing config
  signals:
    agent_1:
      - signal: "[NEXT]"
        target: agent_2
    agent_2:
      - signal: "[END]"
        target: ""             # ← Termination

settings:
  max_handoffs: 10            # ← Safety limit
  # ... other settings
```

---

### Q: Do I need routing config if I use handoff_targets?
**A:** No! You can use either:

**Option 1: Signal-based routing (explicit)**
```yaml
routing:
  signals:
    agent_a:
      - signal: "[CONTINUE]"
        target: agent_b
```

**Option 2: Handoff-based routing (implicit)**
```yaml
agents:
  - name: agent_a
    handoff_targets: [agent_b]  # ← Traditional routing
```

**Option 3: Mix both (recommended)**
```yaml
routing:
  signals:
    agent_a:
      - signal: "[SPECIAL]"      # ← Use signals for important decisions
        target: special_handler

agents:
  - name: agent_a
    handoff_targets: [agent_b]   # ← Use handoff_targets as fallback
```

**Best practice:** Use signals for intentional routing, handoff_targets as fallback.

---

### Q: How do I define multiple signals from one agent?
**A:** List them all under that agent:

```yaml
routing:
  signals:
    classifier:                    # ← One agent
      - signal: "[TYPE_A]"         # ← First signal
        target: handler_a
      - signal: "[TYPE_B]"         # ← Second signal
        target: handler_b
      - signal: "[TYPE_C]"         # ← Third signal
        target: handler_c
      - signal: "[UNKNOWN]"        # ← Fourth signal
        target: human_review
```

**Execution:** Agent emits ONE signal, Crew detects which one and routes accordingly.

---

## Execution & Flow

### Q: What determines when a signal is checked?
**A:** After agent executes and no tool calls remain:

```
Loop iteration:
  1. Execute agent
  2. Check for tool calls
     └─ If yes: execute tools, continue loop (agent re-executes)
  3. Check termination signal (target="")
  4. Check routing signal (target=agent_id)
  5. Check wait_for_signal
  6. Check is_terminal
  7. Check parallel_groups
  8. Fallback to traditional routing
```

**Key:** Tool results are fed back to the SAME agent first, allowing agent to analyze before routing.

---

### Q: How many times can an agent execute?
**A:** Unlimited within loop, but:

1. **Within single request:** Agent re-executes if tools return results
   ```
   Agent executes → Tool calls → Tool results → Agent re-executes → No tools → Routing
   ```

2. **Across handoffs:** Limited by `max_handoffs`
   ```
   Agent A → Agent B → Agent C → ...
   └─ Count: 1    2     3

   When count >= max_handoffs: stop
   ```

---

### Q: What happens if a signal targets a non-existent agent?
**A:** Graceful fallback:

```
Signal found: "[ROUTE_X]" → target: "nonexistent_agent"
         │
         ├─ findAgentByID("nonexistent_agent") returns nil
         │
         └─ Falls back to traditional routing:
            1. Check handoff_targets
            2. Find any other agent
            3. Or end execution
```

**Result:** Workflow doesn't crash, uses fallback routing instead.

---

### Q: Can two agents emit the same signal?
**A:** Yes, but independently:

```yaml
routing:
  signals:
    agent_a:
      - signal: "[CONTINUE]"
        target: agent_c
    agent_b:
      - signal: "[CONTINUE]"
        target: agent_c
```

**Behavior:**
- Agent A emits "[CONTINUE]" → routes to C
- Agent B emits "[CONTINUE]" → routes to C
- Different agents, same signal pattern, same target: OK

---

## Handoffs & Limits

### Q: How is handoff count tracked?
**A:** Count increments each time agent routes to next agent:

```go
handoffCount := 0  // Start at 0

for {
    // ... execute current agent ...

    if nextAgent != nil {  // Signal routing found
        currentAgent = nextAgent
        handoffCount++     // ← Increment here
        continue
    }

    // ... other conditions ...

    handoffCount++         // ← Also increment for fallback routing
    if handoffCount >= ce.crew.MaxHandoffs {
        return nil  // ← Stop if limit reached
    }
}
```

---

### Q: How do I choose max_handoffs?
**A:** Rule of thumb:

```
max_handoffs = (number of agents) × 2-3

Examples:
  3 agents:   max_handoffs = 6-9
  5 agents:   max_handoffs = 10-15
  10 agents:  max_handoffs = 20-30
```

**Why 2-3x?** Allows for:
- Main path through all agents
- Retry loops or backtracking
- Parallel groups counting as handoffs

**Too low:** Workflow ends prematurely
**Too high:** Allows runaway loops (use timeout instead)

---

## Parallel Execution

### Q: How do parallel groups work with signals?
**A:** Signal targets a group instead of an agent:

```yaml
routing:
  signals:
    coordinator:
      - signal: "[ANALYZE]"
        target: parallel_review    # ← Group ID, not agent ID

  parallel_groups:
    parallel_review:               # ← Group definition
      agents: [expert1, expert2, expert3]
      wait_for_all: true
      timeout_seconds: 60
      next_agent: synthesizer
```

**Execution:**
```
Coordinator emits "[ANALYZE]"
    ↓
Crew detects "parallel_review" group
    ↓
Expert1, Expert2, Expert3 ALL execute in parallel
    ↓
Results aggregated into single input
    ↓
Synthesizer receives aggregated input
    ↓
Continue routing from synthesizer
```

---

### Q: What does wait_for_all mean in parallel_groups?
**A:** Controls whether to wait for all agents to finish:

```yaml
wait_for_all: true   # ← Wait for slowest agent
              # Aggregate when all complete
              # If timeout: aggregate what finished

wait_for_all: false  # ← Don't wait for all
              # Aggregate as agents complete
              # Faster execution, possibly incomplete
```

**Recommendation:** Use `true` for accuracy, `false` for speed.

---

## Behavior & Signals

### Q: What does wait_for_signal do?
**A:** Pauses execution and waits for user input:

```yaml
routing:
  agent_behaviors:
    creator:
      wait_for_signal: true   # ← Pause before routing
```

**Execution:**
```
Creator executes → Response ready
    ↓
Check: wait_for_signal = true?
    ↓
YES: Send pause event [PAUSE:creator]
    │  Return nil
    │  CLI shows: "Waiting for input..."
    │
Client receives pause event with creator ID
    │
User provides new input
    │
Client calls SetResumeAgent("creator")
    │
Next request: Creator resumes from where it left off
```

**Use case:** Workflow needs user approval before continuing.

---

### Q: What does is_terminal mean?
**A:** Marks agent as final (no further routing):

```yaml
routing:
  agent_behaviors:
    executor:
      is_terminal: true   # ← Last agent, no handoff
```

**Behavior:**
```
Executor executes → Response ready
    ↓
Check: is_terminal = true?
    ↓
YES: Return response
     │  Workflow ends
     │  No further routing
```

**Use case:** Executor is the final step, results are complete.

---

## Debugging

### Q: How do I debug signal routing issues?
**A:** Enable verbose logging:

```go
executor.SetVerbose(true)
response, err := executor.Execute(ctx, input)
```

**Look for log patterns:**
```
[AGENT START] agent_a (agent_a_id)
[AGENT END] agent_a (agent_a_id) - Success
[ROUTING] agent_a -> agent_b (signal: [NEXT])    # ← Routing detected!
[AGENT START] agent_b (agent_b_id)
```

**If you don't see `[ROUTING]` entry:**
1. Signal not found (check format)
2. Target not found (check agent IDs)
3. Falls back to traditional routing (check logs further)

---

### Q: How do I check if my signal is being detected?
**A:** Add test logging:

```go
// In your agent response, print signal explicitly
response := "I am done. [SIGNAL]"
log.Printf("Agent response: %q", response)  // ← Check signal is in response

// Trace matching
if signalMatchesContent("[SIGNAL]", response) {
    log.Println("✅ Signal matched!")
} else {
    log.Println("❌ Signal NOT matched")
}
```

---

### Q: What if handoff limit is reached?
**A:** Workflow ends silently:

```
Agent 5 routes to Agent 6
    │
    └─ handoffCount = 10, max_handoffs = 10
       │
       ├─ handoffCount >= MaxHandoffs? YES
       │
       └─ Return CrewResponse (workflow ends)
          Log: (nothing special, just returns)
```

**How to detect:**
```go
response, err := executor.Execute(ctx, input)
if response != nil && !response.IsTerminal {
    log.Println("Workflow ended due to handoff limit")
}
```

---

## Performance

### Q: Does signal matching impact performance?
**A:** Negligible - microseconds:

```
LLM call:           100-500 ms ███████████████ (99%+ of time)
Signal matching:    1-10 µs     (0.001% of time)
                                ↑ Invisible
```

**Optimization tip:** Use exact signal formats (no extra spaces) to match on Level 1 instead of Level 3.

---

### Q: Does routing affect memory usage?
**A:** No significant impact:

```
RoutingConfig loaded once at startup
  ├─ Signals: ~KB (one entry per signal)
  ├─ Behaviors: ~KB (one entry per agent)
  └─ ParallelGroups: ~KB (one entry per group)

Per-request memory: ~0 (routing is stateless)
```

**Real memory usage:** Message history (MB+) grows over time, not routing.

---

## Common Mistakes

### Q: What's the most common signal routing mistake?
**A:** Forgetting brackets:

```yaml
# ❌ WRONG
routing:
  signals:
    agent:
      - signal: "DONE"           # ← No brackets!
        target: ""

# ✅ CORRECT
routing:
  signals:
    agent:
      - signal: "[DONE]"         # ← With brackets
        target: ""
```

**How to avoid:** Check config has `[` and `]` around signal.

---

### Q: Second most common mistake?
**A:** Wrong target agent ID:

```yaml
# ❌ WRONG
agents:
  - name: executor             # ← ID is "executor"

routing:
  signals:
    router:
      - signal: "[ROUTE]"
        target: "Executor"     # ← Case mismatch (capital E)

# ✅ CORRECT
routing:
  signals:
    router:
      - signal: "[ROUTE]"
        target: "executor"     # ← Matches exactly
```

**How to avoid:** Copy-paste agent ID from agents list.

---

### Q: Third mistake?
**A:** Not defining what agent should emit in instructions:

```python
# ❌ WRONG
agent_instructions = "Process the request"
# ← Agent doesn't know it should emit [SIGNAL]

# ✅ CORRECT
agent_instructions = """
Process the request.
When done, emit [SIGNAL] to continue.
Example: "Processing complete. [SIGNAL]"
"""
```

**How to avoid:** Always explain signals in agent instructions.

---

## Quick Reference Table

| Question | Answer |
|----------|--------|
| **Signal format?** | `[PATTERN]` with brackets |
| **Matching?** | 3-level: exact → case-insensitive → normalized |
| **Termination?** | target = "" (empty string) |
| **Routing?** | target = agent_id |
| **Information loss?** | Zero - full history preserved |
| **Timeout on routing?** | No - only on agent execution |
| **Handoff count?** | Increments per routing, stops at max_handoffs |
| **Max handoffs?** | agents × 2-3 (recommendation) |
| **History visible?** | Yes to all agents |
| **Performance impact?** | Negligible (microseconds) |
| **Memory impact?** | Negligible (KB) |
| **Fallback routing?** | Yes, if signal not found |
| **Invalid target?** | Graceful fallback, no crash |

---

## Code Locations Quick Reference

| Topic | File | Line |
|-------|------|------|
| Signal matching | `crew_routing.go` | 49-90 |
| Termination check | `crew_routing.go` | 98-124 |
| Routing check | `crew_routing.go` | 126-157 |
| Main loop | `crew.go` | 656-873 |
| Handoff counting | `crew.go` | 654, 791, 861 |
| Type definitions | `config.go` | 14-44 |

---

## Related Resources

- [Detailed Analysis](./SIGNAL_BASED_ROUTING_ANALYSIS.md)
- [Visual Diagrams](./SIGNAL_ROUTING_DIAGRAM.md)
- [Implementation Guide](./SIGNAL_ROUTING_GUIDE.md)
- [Example: Quiz System](./examples/01-quiz-exam/config/crew.yaml)
- [Example: IT Support](./examples/it-support/config/crew.yaml)

