# Signal-Based Routing System - Complete Documentation

## Overview

**Signal-based routing** is the core mechanism in Go-CrewAI that allows agents to communicate and coordinate through explicit signals emitted in their responses. This documentation provides comprehensive coverage from basics to advanced implementation.

---

## 📚 Documentation Hierarchy

```
SIGNAL_ROUTING_README.md (you are here)
├── SIGNAL_ROUTING_QUICK_REF.md .............. One-page quick reference
├── SIGNAL_ROUTING_FAQ.md ................... 30+ Q&A pairs
├── SIGNAL_ROUTING_GUIDE.md ................. Implementation guide
├── SIGNAL_ROUTING_DIAGRAM.md ............... Visual flowcharts
├── SIGNAL_BASED_ROUTING_ANALYSIS.md ........ Technical deep-dive
└── SIGNAL_ROUTING_DOCUMENTATION_INDEX.md ... Complete index
```

---

## 🚀 Quick Start

### 30-Second Version

```
Signal = [PATTERN] in agent response

Example config:
  routing:
    signals:
      agent1:
        - signal: "[NEXT]"
          target: agent2
      agent2:
        - signal: "[END]"
          target: ""  # Terminate

Agent response:
  "Work complete. [NEXT]"
         ↓
  Crew detects [NEXT]
         ↓
  Routes to agent2
```

### 5-Minute Overview

1. **Signal Definition**: `[PATTERN]` in agent response
2. **Matching**: 3-level (exact → case-insensitive → normalized)
3. **Routing**: If signal found, route to target agent or terminate
4. **History**: All agents see complete message history
5. **Limits**: Max handoffs prevents infinite loops

---

## 📖 Which Document Should I Read?

| Your Situation | Best Document |
|---|---|
| **Just implementing config** | [SIGNAL_ROUTING_QUICK_REF.md](SIGNAL_ROUTING_QUICK_REF.md) |
| **Need a quick answer** | [SIGNAL_ROUTING_FAQ.md](SIGNAL_ROUTING_FAQ.md) |
| **Learning to implement** | [SIGNAL_ROUTING_GUIDE.md](SIGNAL_ROUTING_GUIDE.md) |
| **Want visual explanation** | [SIGNAL_ROUTING_DIAGRAM.md](SIGNAL_ROUTING_DIAGRAM.md) |
| **Need complete details** | [SIGNAL_BASED_ROUTING_ANALYSIS.md](SIGNAL_BASED_ROUTING_ANALYSIS.md) |
| **Finding something specific** | [SIGNAL_ROUTING_DOCUMENTATION_INDEX.md](SIGNAL_ROUTING_DOCUMENTATION_INDEX.md) |

---

## ❓ Frequently Asked Questions

### What is a signal?
A signal is a string pattern in brackets `[PATTERN]` that an LLM agent emits in its response to trigger routing decisions.

### Does signal matching impact performance?
No - it takes microseconds while LLM calls take 100-500ms (bottleneck).

### Is information lost between agent handoffs?
No - every agent receives the complete message history of all previous exchanges.

### How do I prevent infinite loops?
Set `max_handoffs` to reasonable value (agents × 2-3).

### Can the same signal come from multiple agents?
Yes - each agent can independently emit the same signal with same or different targets.

---

## 🎯 Key Concepts

### Signal Matching (3-Level)

```
Level 1: Exact match
  "[SIGNAL]" in response? → YES: use it

Level 2: Case-insensitive
  "[signal]" matches "[SIGNAL]"? → YES: use it

Level 3: Normalized bracket
  "[ SIGNAL ]" equals "[signal]" (normalized)? → YES: use it

No match? → Fall back to traditional routing
```

### Signal Types

| Type | Target | Example |
|------|--------|---------|
| **Termination** | "" | `[DONE]` → "" |
| **Routing** | agent_id | `[ROUTE_X]` → agent_x |
| **Parallel** | group_id | `[ANALYZE]` → parallel_group |
| **Conditional** | agent_id | `[TYPE_A]` → handler_a |

### Execution Flow

```
Agent executes LLM call
    ↓
Tool calls? → Yes: Execute, re-run agent
    ↓
Termination signal? → Yes: End workflow
    ↓
Routing signal? → Yes: Hand off to next agent
    ↓
Wait-for-signal? → Yes: Pause and return
    ↓
Is-terminal? → Yes: End workflow
    ↓
Parallel group? → Yes: Execute in parallel
    ↓
Fallback routing
```

---

## 📋 Configuration Template

```yaml
version: "1.0"
name: my-crew
entry_point: agent_1

agents:
  - agent_1
  - agent_2
  - agent_3

routing:
  signals:
    agent_1:
      - signal: "[CONTINUE]"
        target: agent_2
      - signal: "[END]"
        target: ""
    agent_2:
      - signal: "[PROCESS]"
        target: agent_3
    agent_3:
      - signal: "[DONE]"
        target: ""

agent_behaviors:
  agent_3:
    is_terminal: true

settings:
  max_handoffs: 10
  max_rounds: 20
  # ... other settings
```

---

## ✅ Implementation Checklist

- [ ] Define all agents in `agents` list
- [ ] Set `entry_point` to first agent
- [ ] Create `routing.signals` with signal definitions
- [ ] Set `target` to agent ID or "" (empty for terminate)
- [ ] Set `max_handoffs` appropriately
- [ ] Add signal instructions to agent prompts
- [ ] Test signal matching (check exact format)
- [ ] Verify all target agents exist

---

## 🔍 Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| Signal not detected | Check signal format: `[SIGNAL]` with brackets |
| Target agent not found | Verify agent ID exists and matches exactly |
| Infinite loop | Check handoff targets, set reasonable max_handoffs |
| Wrong routing | Check signal matching: exact, case, spaces |
| Lost information | History is always preserved (no loss) |

---

## 📊 Key Statistics

- **Signal matching time**: ~10 microseconds (negligible)
- **LLM call time**: 100-500 ms (bottleneck)
- **Memory per signal**: Negligible (KB)
- **History preservation**: 100% (complete history visible)
- **Handoff limit**: agents × 2-3 (recommended)

---

## 📁 File Locations

| Concept | File | Lines |
|---------|------|-------|
| **Signal matching** | `core/crew_routing.go` | 49-90 |
| **Termination check** | `core/crew_routing.go` | 98-124 |
| **Routing check** | `core/crew_routing.go` | 126-157 |
| **Main execution** | `core/crew.go` | 656-873 |
| **Type definitions** | `core/config.go` | 14-44 |

---

## 🧪 Example Configurations

### Sequential Pipeline (A → B → C)
See: [SIGNAL_ROUTING_GUIDE.md §3.1](SIGNAL_ROUTING_GUIDE.md#31-sequential-pipeline)

### Decision Tree (Multiple Routes)
See: [SIGNAL_ROUTING_GUIDE.md §3.2](SIGNAL_ROUTING_GUIDE.md#32-decision-tree)

### Parallel Processing
See: [SIGNAL_ROUTING_GUIDE.md §3.3](SIGNAL_ROUTING_GUIDE.md#33-parallel-processing-with-aggregation)

### Error Handling
See: [SIGNAL_ROUTING_GUIDE.md §3.4](SIGNAL_ROUTING_GUIDE.md#34-error-handling--recovery)

### Real Examples
- [examples/01-quiz-exam/config/crew.yaml](examples/01-quiz-exam/config/crew.yaml) - Quiz system
- [examples/it-support/config/crew.yaml](examples/it-support/config/crew.yaml) - IT support

---

## 🎓 Learning Paths

### Path 1: "Just Tell Me What I Need" (20 min)
1. [SIGNAL_ROUTING_QUICK_REF.md](SIGNAL_ROUTING_QUICK_REF.md)
2. Example config from real examples
3. Implement based on checklist

### Path 2: "I Want to Understand Properly" (60 min)
1. [SIGNAL_ROUTING_QUICK_REF.md](SIGNAL_ROUTING_QUICK_REF.md)
2. [SIGNAL_ROUTING_DIAGRAM.md](SIGNAL_ROUTING_DIAGRAM.md)
3. [SIGNAL_BASED_ROUTING_ANALYSIS.md](SIGNAL_BASED_ROUTING_ANALYSIS.md) §1-5

### Path 3: "Complete Mastery" (3 hours)
Read all 5 documentation files in order:
1. [SIGNAL_ROUTING_QUICK_REF.md](SIGNAL_ROUTING_QUICK_REF.md)
2. [SIGNAL_ROUTING_FAQ.md](SIGNAL_ROUTING_FAQ.md)
3. [SIGNAL_ROUTING_GUIDE.md](SIGNAL_ROUTING_GUIDE.md)
4. [SIGNAL_ROUTING_DIAGRAM.md](SIGNAL_ROUTING_DIAGRAM.md)
5. [SIGNAL_BASED_ROUTING_ANALYSIS.md](SIGNAL_BASED_ROUTING_ANALYSIS.md)

---

## 🐛 Debugging

### Enable Verbose Mode
```go
executor.SetVerbose(true)
response, err := executor.Execute(ctx, input)
```

### Look for Routing Logs
```
[ROUTING] agent_a -> agent_b (signal: [NEXT])    ✅ Routing detected
[ROUTING] No next agent found for executor       ⚠️  No routing signal
```

### Check Signal Format
```yaml
✅ signal: "[SIGNAL]"              # Correct
❌ signal: "SIGNAL"                # Missing brackets
❌ signal: "[signal with space]"  # OK but use [SIGNAL] style
```

---

## 🚀 Production Checklist

- [ ] All signals documented in agent instructions
- [ ] All target agents exist in crew config
- [ ] `max_handoffs` set appropriately
- [ ] Termination signals defined (target="")
- [ ] Timeout configured for workflow complexity
- [ ] Test cases cover all routing paths
- [ ] Logging enabled for debugging
- [ ] Monitoring configured for signal issues

---

## 📚 Complete Documentation Index

See [SIGNAL_ROUTING_DOCUMENTATION_INDEX.md](SIGNAL_ROUTING_DOCUMENTATION_INDEX.md) for:
- Detailed section-by-section navigation
- How to find specific topics
- Reading paths by experience level
- Complete file reference

---

## 💡 Pro Tips

1. **Use explicit signal formats** - `[ROUTE_TARGET]` is clearer than `[x]`
2. **Document signals in instructions** - Tell agent WHEN to use each signal
3. **Set reasonable max_handoffs** - agents × 2-3 prevents issues
4. **Enable verbose logging** - Makes debugging much easier
5. **Use handoff_targets as fallback** - Signal routing + handoff_targets combo
6. **Test all paths** - Ensure each routing decision works
7. **Monitor handoff logs** - Detect routing issues in production

---

## 🤔 When to Use Signal Routing

✅ **GOOD:**
- Sequential workflows with decision points
- Expert handoffs requiring routing logic
- Conditional branching
- Variable routing based on content
- Multi-stage approval workflows

❌ **NOT GOOD:**
- Simple linear pipelines (use handoff_targets)
- Random routing (use explicit logic)
- Feedback loops (handle differently)

---

## 📞 Quick Links

| Need | Link |
|------|------|
| Quick reference | [SIGNAL_ROUTING_QUICK_REF.md](SIGNAL_ROUTING_QUICK_REF.md) |
| Questions | [SIGNAL_ROUTING_FAQ.md](SIGNAL_ROUTING_FAQ.md) |
| Implementation help | [SIGNAL_ROUTING_GUIDE.md](SIGNAL_ROUTING_GUIDE.md) |
| Visual explanation | [SIGNAL_ROUTING_DIAGRAM.md](SIGNAL_ROUTING_DIAGRAM.md) |
| Technical details | [SIGNAL_BASED_ROUTING_ANALYSIS.md](SIGNAL_BASED_ROUTING_ANALYSIS.md) |
| Find anything | [SIGNAL_ROUTING_DOCUMENTATION_INDEX.md](SIGNAL_ROUTING_DOCUMENTATION_INDEX.md) |

---

## ✨ Documentation Status

| Document | Status | Completeness |
|----------|--------|--------------|
| SIGNAL_ROUTING_QUICK_REF.md | ✅ Complete | 100% |
| SIGNAL_ROUTING_FAQ.md | ✅ Complete | 100% |
| SIGNAL_ROUTING_GUIDE.md | ✅ Complete | 100% |
| SIGNAL_ROUTING_DIAGRAM.md | ✅ Complete | 100% |
| SIGNAL_BASED_ROUTING_ANALYSIS.md | ✅ Complete | 100% |
| SIGNAL_ROUTING_DOCUMENTATION_INDEX.md | ✅ Complete | 100% |

---

**This documentation was created 2025-12-23 as a comprehensive guide to signal-based routing in the Go-CrewAI framework.**

