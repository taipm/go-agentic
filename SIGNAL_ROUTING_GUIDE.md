# Signal-Based Routing: Implementation & Usage Guide

## Quick Start

### Minimum Configuration

```yaml
# crew.yaml
version: "1.0"
name: my-crew
entry_point: agent_1

agents:
  - agent_1
  - agent_2

routing:
  signals:
    agent_1:
      - signal: "[NEXT]"
        target: agent_2
    agent_2:
      - signal: "[END]"
        target: ""  # Terminate

settings:
  max_handoffs: 10
  # ... other settings
```

---

## 1. SIGNAL DESIGN GUIDE

### 1.1 Naming Conventions

**Good Signal Names:**
```yaml
signals:
  router:
    - signal: "[ROUTE_EXECUTOR]"      # Explicit routing
      target: executor
    - signal: "[ROUTE_CLARIFIER]"     # Explicit routing
      target: clarifier
    - signal: "[NEEDS_HUMAN]"         # Conditional decision
      target: human_reviewer
    - signal: "[COMPLETE]"             # Status signal
      target: ""
```

**Avoid:**
```yaml
signals:
  router:
    - signal: "[X]"              # ❌ Too vague
      target: executor
    - signal: "next"             # ❌ No brackets
      target: executor
    - signal: "[route executor]" # ⚠️  Works but prefer [ROUTE_EXECUTOR]
      target: executor
```

### 1.2 Signal Scope Design

**Per-Agent Signals (Recommended):**
```yaml
routing:
  signals:
    teacher:           # Only teacher can emit these
      - signal: "[QUESTION]"
        target: parallel_question
    student:           # Only student can emit these
      - signal: "[ANSWER]"
        target: parallel_answer
    reporter:          # Only reporter can emit these
      - signal: "[DONE]"
        target: ""
```

**Benefits:**
- Clear ownership
- Easy to debug (agent X → signal Y expected)
- Prevents accidental routing from wrong agent
- Supports multiple signals per agent

### 1.3 Conditional Routing Patterns

**Pattern: Multiple Signals from Same Agent**
```yaml
routing:
  signals:
    analyzer:
      - signal: "[CRITICAL]"         # High priority
        target: human_team
      - signal: "[HIGH]"             # Medium priority
        target: escalation_team
      - signal: "[RESOLVED]"         # Resolved automatically
        target: ""
```

**LLM Instruction:**
```python
agent_instructions = """
Based on analysis, emit ONE of:
- [CRITICAL] if issue is critical (safety, security, major feature broken)
- [HIGH] if issue is high priority (performance, major bug)
- [RESOLVED] if issue is resolved automatically
"""
```

**Pattern: Sequential Decision Points**
```yaml
routing:
  signals:
    classifier:
      - signal: "[HUMAN_REQUIRED]"
        target: human_review
      - signal: "[NEEDS_CONTEXT]"
        target: context_gatherer
      - signal: "[READY_TO_PROCESS]"
        target: processor
```

### 1.4 Termination Signal Pattern

```yaml
routing:
  signals:
    executor:
      - signal: "[ERROR_CRITICAL]"    # Fail the workflow
        target: ""
      - signal: "[SUCCESS]"           # Complete successfully
        target: ""
      - signal: "[PARTIAL]"           # Partial completion
        target: ""
```

**Key:** Target = "" terminates immediately, don't use for handoffs

---

## 2. IMPLEMENTATION CHECKLIST

### 2.1 Configuration Checklist

- [ ] Define all agents in `agents` list
- [ ] Set `entry_point` to first agent
- [ ] Create `routing.signals` section
- [ ] For each agent that routes: add entries in `signals[agent_id]`
- [ ] For each routing signal: set `target` to next agent ID or empty ""
- [ ] Set `max_handoffs` (typically max agents × 2)
- [ ] Set `max_rounds` (typically max agents × 3)
- [ ] Use `parallel_groups` for agents that execute simultaneously
- [ ] Define `agent_behaviors` for wait_for_signal or is_terminal agents

### 2.2 Agent Instruction Checklist

- [ ] Instructions tell agent WHEN to emit signals
- [ ] Instructions tell agent WHICH signals are available
- [ ] Instructions provide examples of signal usage
- [ ] Instructions explain what each signal means

**Template:**
```python
agent_instructions = """
You are a {role}. Your job is to {responsibility}.

When you are done, emit ONE of these signals:
- [SIGNAL_1]: Use when {condition_1}
- [SIGNAL_2]: Use when {condition_2}
- [SIGNAL_END]: Use when work is complete

Example responses:
- "I have reviewed the code. [SIGNAL_1]"
- "The feature is implemented. [SIGNAL_END]"

IMPORTANT: Always emit exactly one signal at the end of your response.
"""
```

### 2.3 Testing Checklist

- [ ] Test happy path (all signals found as expected)
- [ ] Test missing signal (agent doesn't emit expected signal)
- [ ] Test invalid target (signal targets non-existent agent)
- [ ] Test handoff limit (reach max_handoffs and stop)
- [ ] Test termination signal (workflow ends at correct agent)
- [ ] Test parallel execution (multiple agents run together)
- [ ] Test wait_for_signal (pause and resume)

---

## 3. SIGNAL DESIGN FOR SPECIFIC USE CASES

### 3.1 Sequential Pipeline

**Use Case:** Agent A → Agent B → Agent C (linear workflow)

```yaml
routing:
  signals:
    agent_a:
      - signal: "[CONTINUE]"
        target: agent_b
    agent_b:
      - signal: "[CONTINUE]"
        target: agent_c
    agent_c:
      - signal: "[DONE]"
        target: ""

agent_behaviors:
  agent_a:
    description: "First step in pipeline"
  agent_b:
    description: "Second step"
  agent_c:
    is_terminal: true
    description: "Final step"
```

### 3.2 Decision Tree

**Use Case:** Route to different agents based on input classification

```yaml
routing:
  signals:
    classifier:
      - signal: "[TYPE_A]"
        target: handler_a
        description: "Route to handler for type A"
      - signal: "[TYPE_B]"
        target: handler_b
        description: "Route to handler for type B"
      - signal: "[TYPE_C]"
        target: handler_c
        description: "Route to handler for type C"
      - signal: "[UNKNOWN]"
        target: human_review
        description: "Human review for unknown types"

handler_a:
  - signal: "[COMPLETE_A]"
    target: ""
handler_b:
  - signal: "[COMPLETE_B]"
    target: ""
handler_c:
  - signal: "[COMPLETE_C]"
    target: ""
```

**LLM Instruction for Classifier:**
```python
classifier_instructions = """
Classify the input into ONE category and emit the matching signal:
- [TYPE_A]: Input has characteristics {...}
- [TYPE_B]: Input has characteristics {...}
- [TYPE_C]: Input has characteristics {...}
- [UNKNOWN]: Input doesn't fit any category

Example: "This is definitely type B because {...}. [TYPE_B]"
"""
```

### 3.3 Parallel Processing with Aggregation

**Use Case:** Multiple experts analyze in parallel, then results aggregated

```yaml
routing:
  signals:
    coordinator:
      - signal: "[PARALLEL_ANALYZE]"
        target: parallel_review

  parallel_groups:
    parallel_review:
      agents: [expert_a, expert_b, expert_c]
      wait_for_all: true
      timeout_seconds: 60
      next_agent: synthesizer

  synthesizer:
    - signal: "[SYNTHESIS_COMPLETE]"
      target: ""
```

**Flow:**
```
Coordinator emits "[PARALLEL_ANALYZE]"
    ↓
Parallel execution: Expert A, B, C all run with same input
    ↓
Results aggregated: "[Expert A]: ...\n[Expert B]: ...\n[Expert C]: ..."
    ↓
Synthesizer processes aggregated results
    ↓
Synthesizer emits "[SYNTHESIS_COMPLETE]"
    ↓
Workflow ends
```

### 3.4 Error Handling & Recovery

**Use Case:** Attempt action, retry on failure, escalate if persistent

```yaml
routing:
  signals:
    executor:
      - signal: "[SUCCESS]"
        target: ""
      - signal: "[RETRY]"
        target: executor        # ← Retry same agent
      - signal: "[ESCALATE]"
        target: human_escalation

  human_escalation:
    - signal: "[MANUAL_RESOLUTION]"
      target: ""

agent_behaviors:
  executor:
    wait_for_signal: false
```

**LLM Instruction for Executor:**
```python
executor_instructions = """
Attempt to execute the task.

Emit ONE of:
- [SUCCESS]: If task completed without errors
- [RETRY]: If task failed due to transient error (network, timeout)
         Only emit 2 times max. Then escalate.
- [ESCALATE]: If task failed due to permanent error (missing data, invalid)
             Or if retries exhausted.

Example: "Network timeout occurred. Retrying... [RETRY]"
"""
```

### 3.5 Multi-Stage Approval Workflow

**Use Case:** Create → Review → Approve → Publish

```yaml
routing:
  signals:
    creator:
      - signal: "[READY_FOR_REVIEW]"
        target: reviewer
      - signal: "[DRAFT]"
        target: ""              # Wait for external signal

    reviewer:
      - signal: "[APPROVED]"
        target: publisher
      - signal: "[REJECTED]"
        target: creator        # Back to creator for revision
      - signal: "[NEEDS_CLARIFICATION]"
        target: clarifier

    clarifier:
      - signal: "[CLARIFIED]"
        target: reviewer

    publisher:
      - signal: "[PUBLISHED]"
        target: ""

agent_behaviors:
  creator:
    wait_for_signal: true      # Waits if in DRAFT state
  reviewer:
    wait_for_signal: false
  publisher:
    is_terminal: true          # Final step
```

---

## 4. DEBUGGING SIGNAL ROUTING

### 4.1 Tracing Execution

**Enable logging to see signal routing:**

```go
// In your code where you load the crew
executor, err := NewCrewExecutorFromConfig(apiKey, configDir, tools)
executor.SetVerbose(true)  // Enable verbose output

// Run the crew
response, err := executor.Execute(ctx, userInput)
```

**Output with signal tracing:**
```
[AGENT START] teacher (teacher_id)
[AGENT END] teacher (teacher_id) - Success
[ROUTING] teacher -> parallel_question (signal: [QUESTION])
[AGENT START] student (student_id)
[AGENT START] reporter (reporter_id)
[AGENT END] student (student_id) - Success
[AGENT END] reporter (reporter_id) - Success
[ROUTING] No next agent found for student
```

### 4.2 Common Issues & Solutions

**Issue 1: Signal not detected**
```
Symptom: Agent response doesn't trigger routing
         Falls back to default routing instead

Debugging:
  1. Check if signal appears in response content
     log.Printf("Response: %s", response.Content)

  2. Check signal format matches exactly
     Expected: "[SIGNAL]"
     Response: "Some text [SIGNAL] more text" ✅
     Response: "[SIGNAL ]" (extra space) - needs level 3 matching

  3. Check if signal is defined in config
     routing.signals[agent_id] should contain the signal

  4. Check for case sensitivity
     "[SIGNAL]" vs "[signal]" - both work (case-insensitive)
```

**Issue 2: Target agent not found**
```
Symptom: Signal emitted but routing fails silently
         Falls back to traditional routing

Debugging:
  1. Check agent ID exists in crew
     agents: [agent_a, agent_b] ✅
     target: agent_c (non-existent) ❌

  2. Check agent ID spelling matches exactly
     Signal: target: "executor"
     Agents: ["executor"] ✅
     Agents: ["Executor"] ❌ (case-sensitive)

  3. Check Routing.Signals structure loaded correctly
     if ce.crew.Routing == nil { log.Println("No routing!") }
```

**Issue 3: Infinite routing loop**
```
Symptom: Agent keeps handing off to same agent
         or cycles: A → B → A → B...

Debugging:
  1. Check signal definitions don't create cycles
  2. Check max_handoffs is reasonable (not too high)
  3. Add logging to track routing path
     log.Printf("[ROUTING] %s -> %s", current.ID, next.ID)
```

**Issue 4: Handoff limit reached unexpectedly**
```
Symptom: Workflow ends before expected
         "handoffCount >= MaxHandoffs"

Debugging:
  1. Count expected handoffs in design
  2. Check max_handoffs is >= expected count
  3. Check for unnecessary fallback routing
     Every hop without explicit signal counts as handoff
```

### 4.3 Validation Commands

```bash
# Check YAML syntax
go run ./examples/00-hello-crew -validate-config

# Check routing configuration
go run ./examples/01-quiz-exam -verbose

# Test signal matching
echo "[ROUTE_EXECUTOR]" | grep "\[ROUTE"  # Check signal format
```

---

## 5. PERFORMANCE OPTIMIZATION

### 5.1 Reducing Handoffs

**Problem:** Too many handoffs = slow workflow

**Solution 1: Combine agents**
```yaml
# Before: 5 agents, 4 handoffs
before:
  agents: [step1, step2, step3, step4, step5]
  # Each step hands off to next

# After: 2-3 agents, 1-2 handoffs
after:
  agents: [planner, executor, reporter]
  # Planner does step1+step2, Executor does step3+step4, Reporter does step5
```

**Solution 2: Use parallel groups**
```yaml
# Before: Sequential execution (3 handoffs)
before:
  teacher -> student -> reporter

# After: Parallel execution (1 handoff)
after:
  teacher -> [student, reporter] in parallel -> output
```

### 5.2 Signal Matching Performance

**Current:** 3-level matching (exact → case-insensitive → normalized)

**Optimization:** Use exact signals
```yaml
# Good (likely matches on level 1)
signals:
  agent:
    - signal: "[ROUTE]"       # Clear, no spaces
      target: next_agent

# Avoid (falls through to level 2-3)
signals:
  agent:
    - signal: "[ ROUTE ]"     # Spaces require normalization
      target: next_agent
```

**Benchmark:**
- Level 1 match: ~0.1 microseconds
- Level 2 match: ~0.2 microseconds
- Level 3 match: ~1-2 microseconds
- **Total per agent:** <10 microseconds

Negligible compared to LLM call (~100-500ms)

### 5.3 History Size Optimization

**Problem:** History grows unbounded, slowing LLM calls

**Solution:** Context trimming (automatic)
```yaml
settings:
  max_context_window: 32000      # Max tokens in history
  context_trim_percent: 20.0     # Trim 20% when full

# Automatic: Keeps first message + recent messages
# After trim: [Initial context] [trimmed N messages] [Recent messages]
```

---

## 6. PRODUCTION CHECKLIST

### 6.1 Pre-Deployment

- [ ] All signal formats verified (bracket notation)
- [ ] All target agents exist in crew
- [ ] max_handoffs set to reasonable value (2-3x agent count)
- [ ] Agent instructions include signal documentation
- [ ] Error signals defined (termination points)
- [ ] Test cases cover all routing paths
- [ ] Logging enabled for debugging
- [ ] Timeout values reasonable for workflow complexity

### 6.2 Monitoring

```yaml
# Add to logging
[ROUTING] agent_a -> agent_b (signal: [SIGNAL])

# Monitor for:
- Frequent "[ROUTING] No next agent found"  ← Missing routing
- Frequent handoffs near max_handoffs       ← Inefficient routing
- Signals not found (fallback routing)      ← Signal format issues
```

### 6.3 Safety Guards

```go
// In your execution code
if ce.crew.Routing == nil {
    log.Println("WARNING: No routing configuration loaded")
    // Fallback to simple sequential routing
}

// Check max_handoffs reasonable
if ce.crew.MaxHandoffs < len(ce.crew.Agents) {
    log.Printf("WARNING: max_handoffs (%d) < agent count (%d)",
        ce.crew.MaxHandoffs, len(ce.crew.Agents))
}
```

---

## 7. EXAMPLES

### Example 1: Simple 2-Agent Workflow

**Config:**
```yaml
version: "1.0"
name: simple-workflow
entry_point: analyzer

agents:
  - analyzer
  - executor

routing:
  signals:
    analyzer:
      - signal: "[READY]"
        target: executor
    executor:
      - signal: "[DONE]"
        target: ""
```

**Agent Instructions (Analyzer):**
```python
You are an analyzer. Read the user input and decide what to do.
Output your analysis, then emit [READY] when ready to execute.

Example: "The request is valid and straightforward. [READY]"
```

**Agent Instructions (Executor):**
```python
You are an executor. Implement the analysis from the analyzer.
When complete, emit [DONE].

Example: "Implementation complete and tested. [DONE]"
```

### Example 2: Decision Tree with 3 Branches

**Config:**
```yaml
routing:
  signals:
    router:
      - signal: "[SIMPLE]"
        target: simple_handler
      - signal: "[COMPLEX]"
        target: complex_handler
      - signal: "[REQUIRES_DATA]"
        target: data_gatherer

  simple_handler:
    - signal: "[COMPLETE]"
      target: ""

  complex_handler:
    - signal: "[COMPLETE]"
      target: ""

  data_gatherer:
    - signal: "[DATA_READY]"
      target: complex_handler
```

### Example 3: Parallel + Aggregation

**Config:**
```yaml
routing:
  signals:
    reviewer:
      - signal: "[ANALYZE]"
        target: parallel_team

  parallel_groups:
    parallel_team:
      agents: [backend_expert, frontend_expert, security_expert]
      wait_for_all: true
      timeout_seconds: 60
      next_agent: synthesizer

  synthesizer:
    - signal: "[REPORT]"
      target: ""
```

---

## 8. REFERENCES

- [Signal-Based Routing Analysis](./SIGNAL_BASED_ROUTING_ANALYSIS.md)
- [Visual Diagrams](./SIGNAL_ROUTING_DIAGRAM.md)
- [crew_routing.go](./core/crew_routing.go) - Implementation
- [crew.go ExecuteStream](./core/crew.go#L629) - Main execution loop
- [Examples: 01-quiz-exam](./examples/01-quiz-exam/config/crew.yaml)
- [Examples: it-support](./examples/it-support/config/crew.yaml)

