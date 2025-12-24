# Signal-Based Routing Analysis: core/crew.go

Phân tích chi tiết cơ chế signal-based routing trong Go-CrewAI framework.

---

## 1. SIGNAL DEFINITION & ARCHITECTURE

### 1.1 Định nghĩa Signal

**Signal** là một string pattern bao quanh bởi dấu ngoặc vuông `[...]`, được LLM emit trong response content.

```go
// Type: RoutingSignal (core/config.go:15)
type RoutingSignal struct {
	Signal      string `yaml:"signal"`        // Pattern to match, e.g. "[ROUTE_EXECUTOR]"
	Target      string `yaml:"target"`        // Agent ID to route to (empty = TERMINATE)
	Description string `yaml:"description"`   // Documentation
}
```

**Examples từ config:**
```yaml
routing:
  signals:
    teacher:
      - signal: "[QUESTION]"
        target: parallel_question      # Route tới parallel group
      - signal: "[END]"
        target: reporter               # Route tới agent cụ thể
    executor:
      - signal: "[COMPLETE]"
        target: ""                     # Empty = TERMINATE workflow
```

### 1.2 Signal Storage Architecture

Signals được lưu trong `RoutingConfig` (core/config.go:39):

```go
type RoutingConfig struct {
	Signals        map[string][]RoutingSignal     // Key: agent_id, Value: signals emitted by that agent
	Defaults       map[string]string              // Default routing per agent
	AgentBehaviors map[string]AgentBehavior       // Behavior config (wait_for_signal, is_terminal, etc)
	ParallelGroups map[string]ParallelGroupConfig // Parallel execution groups
}
```

**Lookup Pattern:**
```
Agent emits signal → Crew has entry in crew.Routing.Signals[agent_id] → Array of RoutingSignal
```

---

## 2. SIGNAL MATCHING & DETECTION

### 2.1 Signal Matching Logic (core/crew_routing.go:49-90)

**Function:** `signalMatchesContent(signal, content string) bool`

Matching strategy có 3 levels (prioritized):

```go
// Level 1: Exact string match (FASTEST)
if strings.Contains(content, signal) {
    return true  // "[COMPLETE]" matches if found anywhere in response
}

// Level 2: Case-insensitive match
if strings.Contains(strings.ToLower(content), strings.ToLower(signal)) {
    return true  // "[ Hoàn thành ]" matches "[hoàn thành]"
}

// Level 3: Normalized bracketed match (MOST FLEXIBLE)
// Handles whitespace variations: "[ KẾT THÚC  THI ]" == "[kết thúc thi]"
// - Extract pattern from [...] brackets
// - Normalize internal spaces to single space
// - Collapse multiple spaces
// - Compare normalized versions
```

**Example Matching:**
```
Signal:        "[ROUTE_EXECUTOR]"
Response 1:    "..."[ROUTE_EXECUTOR]..."      ✅ Match (Level 1: exact)
Response 2:    "..."[route_executor]..."      ✅ Match (Level 2: case-insensitive)
Response 3:    "...[ Route Executor ]..."     ✅ Match (Level 3: normalized)
Response 4:    "...no signal in response..."  ❌ No match
```

### 2.2 Whitespace Normalization (core/crew_routing.go:24-47)

```go
func normalizeSignalText(text string) string {
    // "[  KẾT THÚC  ]" → "[kết thúc]"
    text = strings.ToLower(text)
    if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
        inner := strings.TrimPrefix(strings.TrimSuffix(text, "]"), "[")
        inner = strings.TrimSpace(inner)
        // Collapse multiple spaces into single: "a  b  c" → "a b c"
        parts := strings.Fields(inner)
        inner = strings.Join(parts, " ")
        text = "[" + inner + "]"
    }
    return text
}
```

---

## 3. ROUTING FLOW EXECUTION

### 3.1 Signal-Based Routing Flow (core/crew.go:784-793)

```
ExecuteStream execution loop:
│
├─ Agent executes LLM call → Response with content
│
├─ [IF] Tool calls exist
│  │   └─ Execute tools & feed results back to same agent
│  │       └─ Continue loop (agent analyzes results & emits signal)
│  │
│  └─ [THEN] Check termination signal
│      │     (Target = "" → TERMINATE)
│      │
│      ├─ YES: Return nil, workflow ends
│      │
│      └─ NO: Check routing signal
│            (Target = agent_id → HANDOFF)
│            │
│            ├─ YES: currentAgent = nextAgent, continue loop
│            │
│            └─ NO: Check wait_for_signal
│                  │
│                  ├─ YES: Pause execution, return (await user input)
│                  │
│                  └─ NO: Check is_terminal
│                        │
│                        ├─ YES: Return (terminal agent)
│                        │
│                        └─ NO: Check parallel_groups
│                              │
│                              └─ Continue to next agent
```

**Code Location:** [crew.go:775-812](core/crew.go#L775-L812)

```go
// Check TERMINATION signals FIRST (target="")
terminationResult := ce.checkTerminationSignal(currentAgent, response.Content)
if terminationResult != nil && terminationResult.ShouldTerminate {
    streamChan <- NewStreamEvent("terminate", currentAgent.Name,
        fmt.Sprintf("[TERMINATE] Workflow ended by signal: %s", terminationResult.Signal))
    return nil
}

// Check ROUTING signals SECOND (target=agent_id)
nextAgent := ce.findNextAgentBySignal(currentAgent, response.Content)
if nextAgent != nil {
    currentAgent = nextAgent
    input = response.Content  // ← Pass response to next agent
    handoffCount++
    continue
}

// Check BEHAVIOR signals THIRD (wait_for_signal)
behavior := ce.getAgentBehavior(currentAgent.ID)
if behavior != nil && behavior.WaitForSignal {
    streamChan <- NewStreamEvent("pause", currentAgent.Name,
        fmt.Sprintf("[PAUSE:%s] Waiting for user input", currentAgent.ID))
    return nil
}
```

### 3.2 Termination Signal Detection (core/crew_routing.go:98-124)

```go
func (ce *CrewExecutor) checkTerminationSignal(current *Agent, responseContent string) *TerminationResult {
    // 1. Check if Routing exists
    if ce.crew.Routing == nil {
        return nil  // No routing config → can't terminate via signal
    }

    // 2. Get signals for current agent
    signals, exists := ce.crew.Routing.Signals[current.ID]
    if !exists || len(signals) == 0 {
        return nil  // No signals defined for this agent
    }

    // 3. Find termination signal (Target == "")
    for _, sig := range signals {
        if sig.Target == "" {  // ← Key: TERMINATION = empty target
            if signalMatchesContent(sig.Signal, responseContent) {
                log.Printf("[ROUTING] %s -> TERMINATE (signal: %s)", current.ID, sig.Signal)
                return &TerminationResult{
                    ShouldTerminate: true,
                    Signal:          sig.Signal,
                }
            }
        }
    }
    return nil
}
```

### 3.3 Agent Routing Detection (core/crew_routing.go:126-157)

```go
func (ce *CrewExecutor) findNextAgentBySignal(current *Agent, responseContent string) *Agent {
    // 1. Check if Routing exists
    if ce.crew.Routing == nil {
        return nil
    }

    // 2. Get signals for current agent
    signals, exists := ce.crew.Routing.Signals[current.ID]
    if !exists || len(signals) == 0 {
        return nil  // Agent doesn't emit any routing signals
    }

    // 3. Check which signal is present in response
    for _, sig := range signals {
        if sig.Target == "" {
            continue  // Skip termination signals
        }

        // Signal match in response?
        if signalMatchesContent(sig.Signal, responseContent) {
            // Find target agent
            nextAgent := ce.findAgentByID(sig.Target)
            if nextAgent != nil {
                log.Printf("[ROUTING] %s -> %s (signal: %s)",
                    current.ID, nextAgent.ID, sig.Signal)
            }
            return nextAgent  // Can be nil if target not found!
        }
    }
    return nil  // No signal found in response
}
```

---

## 4. CONTEXT PRESERVATION THROUGH HANDOFFS

### 4.1 Message History Architecture

**Global history maintained in CrewExecutor:**
```go
type CrewExecutor struct {
    history []Message  // ← Shared conversation history
    // ...
}
```

**Message structure:**
```go
type Message struct {
    Role    string  // "user", "assistant", "system"
    Content string  // Text content (includes tool results, embeddings, etc)
}
```

### 4.2 Context Flow Through Handoffs (core/crew.go)

**Execution sequence:**

```
1. User input → ce.history.append(Message{role: "user", content: input})
                           ↓
2. Agent executes → response := ExecuteAgent(ctx, agent, input, ce.history, apiKey)
                    • Agent receives FULL ce.history
                    • Agent knows all previous agent responses
                           ↓
3. Agent response → ce.history.append(Message{role: "assistant", content: response.Content})
                           ↓
4. [IF] Tool calls exist
    └─ Execute tools → ce.history.append(Message{role: "user", content: toolResults})
       Continue loop → Agent re-executes with updated history
                           ↓
5. [IF] Routing signal found
    └─ Handoff → input = response.Content
                 currentAgent = nextAgent
                 continue loop
                      ↓
6. Next agent executes → response := ExecuteAgent(ctx, nextAgent, input, ce.history, apiKey)
                         • Receives SAME ce.history + new input
                         • Can see all previous agent decisions
```

**Code reference:** [crew.go:673-794](core/crew.go#L673-L794)

### 4.3 No Information Loss?

✅ **YES - Information is fully preserved:**

1. **Message History Passed Completely:** Each agent receives `ce.history` which contains ALL previous messages

2. **Previous Agent Response Available:** Agent N+1 can see Agent N's response in history

3. **Tool Results Included:** Tool execution results are added to history before next agent

4. **Input Parameter:** Fresh response from Agent N passed as `input` to Agent N+1

**Example with 3 agents:**
```
History after Teacher executes:
  [0] system: Initial context
  [1] user: "Start exam"
  [2] assistant: "[QUESTION] What is 2+2?"  ← Teacher response

History after Student executes:
  [0] system: Initial context
  [1] user: "Start exam"
  [2] assistant: "[QUESTION] What is 2+2?"  ← Teacher response (visible to Student)
  [3] user: "[ANSWER] 4"                     ← Student response
  [4] user: Tool results from marking tool

History after Reporter executes:
  [0] system: Initial context
  [1] user: "Start exam"
  [2] assistant: "[QUESTION] What is 2+2?"  ← Can see what Teacher asked
  [3] user: "[ANSWER] 4"                     ← Can see what Student answered
  [4] user: Tool results                     ← Can see marking results
  [5] user: "Report saved"                   ← Reporter sees everything
```

### 4.4 Deep Copy Safety (core/crew.go:13-25)

```go
// copyHistory creates a deep copy to ensure thread safety
func copyHistory(original []Message) []Message {
    if len(original) == 0 {
        return []Message{}
    }
    copied := make([]Message, len(original))
    copy(copied, original)  // Shallow copy is safe since Message is value type
    return copied
}
```

---

## 5. HANDOFF LIMITS & ENFORCEMENT

### 5.1 Max Handoffs Configuration

**Configured in crew.yaml:**
```yaml
settings:
  max_handoffs: 30  # Maximum number of agent-to-agent transitions
```

**Stored in Crew struct:**
```go
type Crew struct {
    MaxHandoffs int  // ← Limit on handoffs
    // ...
}
```

### 5.2 Handoff Counting & Enforcement (core/crew.go)

**Execution pattern:**

```go
handoffCount := 0  // Line 654 (ExecuteStream), Line 890 (Execute)

for {
    // ... Execute agent ...

    // [1] Routing signal found → next agent
    if nextAgent != nil {
        currentAgent = nextAgent
        handoffCount++  // ← Count at line 791, 964
        continue
    }

    // [2] Parallel execution → next agent
    if parallelGroup.NextAgent != "" {
        // ...
        handoffCount++  // ← Count at line 852, 1030
        continue
    }

    // [3] Regular fallback → next agent
    handoffCount++  // ← Count at line 860, 1038
    if handoffCount >= ce.crew.MaxHandoffs {  // ← ENFORCEMENT
        return nil  // Exit loop, workflow ends
    }

    nextAgent = ce.findNextAgent(currentAgent)
    // ...
}
```

**Enforcement location:** [crew.go:861-862](core/crew.go#L861-L862) and [1039-1040](core/crew.go#L1039-L1040)

### 5.3 Handoff Limit Behavior

| Condition | Behavior |
|-----------|----------|
| `handoffCount < MaxHandoffs` | Continue routing normally |
| `handoffCount >= MaxHandoffs` | Return CrewResponse (end workflow) |
| Timeout occurs first | ExecuteStream returns error |

**Example with MaxHandoffs=5:**
```
Handoff 1: teacher → student      (count=1)
Handoff 2: student → teacher      (count=2)
Handoff 3: teacher → reporter     (count=3)
Handoff 4: reporter → teacher     (count=4)
Handoff 5: teacher → executor     (count=5)
           ↓
Handoff 6: executor → ...         (count=6 >= 5) ❌ STOP
           Workflow ends, return response
```

---

## 6. PARALLEL GROUP EXECUTION

### 6.1 Parallel Group Configuration

**Config structure (core/config.go:29-36):**

```go
type ParallelGroupConfig struct {
    Agents         []string `yaml:"agents"`       // Which agents to run in parallel
    WaitForAll     bool     `yaml:"wait_for_all"` // Wait for all to complete?
    TimeoutSeconds int      `yaml:"timeout_seconds"` // Max time for parallel execution
    NextAgent      string   `yaml:"next_agent"`   // Who to handoff to after parallel
    Description    string   `yaml:"description"`
}
```

**YAML example:**
```yaml
routing:
  signals:
    teacher:
      - signal: "[QUESTION]"
        target: parallel_question  # ← Routes to parallel GROUP not agent

  parallel_groups:
    parallel_question:
      agents: [student, reporter]     # Both run in parallel
      wait_for_all: false             # Don't wait for all to finish
      timeout_seconds: 30             # Max 30s for parallel execution
      next_agent: teacher             # After parallel, goto teacher
```

### 6.2 Parallel Execution Detection (core/crew_routing.go:202-223)

```go
func (ce *CrewExecutor) findParallelGroup(agentID string, signalContent string) *ParallelGroupConfig {
    if ce.crew.Routing == nil || ce.crew.Routing.ParallelGroups == nil {
        return nil
    }

    // Check if agent's signal targets a parallel group
    if signals, exists := ce.crew.Routing.Signals[agentID]; exists {
        for _, signal := range signals {
            if signalMatchesContent(signal.Signal, signalContent) {
                // Check if this signal targets a parallel group (not a single agent)
                if parallelGroup, exists := ce.crew.Routing.ParallelGroups[signal.Target]; exists {
                    return &parallelGroup  // Found parallel group!
                }
            }
        }
    }
    return nil
}
```

### 6.3 Parallel Execution Flow (core/crew.go:813-857)

```go
// Check for parallel group execution AFTER routing signals
parallelGroup := ce.findParallelGroup(currentAgent.ID, response.Content)
if parallelGroup != nil {
    // Get agents for this parallel group
    var parallelAgents []*Agent
    for _, agentID := range parallelGroup.Agents {
        if agent, exists := agentMap[agentID]; exists {
            parallelAgents = append(parallelAgents, agent)
        }
    }

    if len(parallelAgents) > 0 {
        // Execute all parallel agents with input (contains tool results & vectors)
        parallelResults, err := ce.ExecuteParallelStream(ctx, input, parallelAgents, streamChan)
        if err != nil {
            return err
        }

        // Aggregate results from all parallel agents
        aggregatedInput := ce.aggregateParallelResults(parallelResults)

        // Add aggregated to history
        ce.history = append(ce.history, Message{
            Role:    "user",
            Content: aggregatedInput,
        })

        // Move to next agent in pipeline
        if parallelGroup.NextAgent != "" {
            currentAgent = nextAgent  // NextAgent from config
            input = aggregatedInput
            handoffCount++
            continue
        }
    }
}
```

---

## 7. TIMEOUT & BEHAVIOR SIGNALS

### 7.1 Agent Behavior Configuration

```go
type AgentBehavior struct {
    WaitForSignal bool   `yaml:"wait_for_signal"` // Pause and wait for user
    AutoRoute     bool   `yaml:"auto_route"`      // Auto-route when signal found
    IsTerminal    bool   `yaml:"is_terminal"`     // Terminal agent (no handoff)
    Description   string
}
```

**YAML example:**
```yaml
routing:
  agent_behaviors:
    orchestrator:
      wait_for_signal: true   # Pause execution, await user input
      auto_route: false       # Don't auto-route
      description: "Orchestrator waits for explicit signal"
    executor:
      is_terminal: true       # Terminal, no handoff
      description: "Executor is the final agent"
```

### 7.2 Wait-For-Signal Behavior (core/crew.go:795-806)

```go
behavior := ce.getAgentBehavior(currentAgent.ID)
if behavior != nil && behavior.WaitForSignal {
    // Pause execution, return with agent ID for resume
    streamChan <- NewStreamEvent("pause", currentAgent.Name,
        fmt.Sprintf("[PAUSE:%s] Waiting for user input", currentAgent.ID))
    return nil  // ← Stop execution here
}
```

**Flow when WaitForSignal=true:**
```
Agent executes → Response contains signal → NOT routed automatically
                                          ↓
                                      Signal found but paused
                                          ↓
                                   Return to client with [PAUSE:agent_id]
                                          ↓
                             Client receives PausedAgentID = agent_id
                                          ↓
                      On next request, SetResumeAgent(agent_id)
                                          ↓
                                 Continue from that agent
```

### 7.3 No Timeout on Signal Routing

**Key finding:** Signal routing itself has **NO TIMEOUT**

- Signal matching is synchronous string comparison (fast)
- Timeout only applies to:
  1. **Agent execution** (LLM call) - via `context.WithTimeout`
  2. **Tool execution** - via `ToolTimeoutConfig`
  3. **Parallel execution** - via `TimeoutSeconds` in parallel group config

```go
// No timeout on:
- signalMatchesContent() ← Direct string matching
- checkTerminationSignal() ← O(n) loop over signals
- findNextAgentBySignal() ← O(n) loop over signals
```

---

## 8. SIGNAL REGISTRY & DISCOVERY

### 8.1 Built-In Signals (Examples from codebase)

| Signal Pattern | Target | Meaning | Examples |
|---|---|---|---|
| `[ROUTE_*]` | agent_id | Route to specific agent | `[ROUTE_EXECUTOR]`, `[ROUTE_CLARIFIER]` |
| `[QUESTION]` | group/agent | Question asked | `[QUESTION]` |
| `[ANSWER]` | group/agent | Answer provided | `[ANSWER]` |
| `[COMPLETE]` | empty | Workflow complete | `[COMPLETE]`, `[DONE]` |
| `[END]` | agent_id | End current phase | `[END]` |
| `[KẾT THÚC]` | agent_id | End (Vietnamese) | `[KẾT THÚC THI]`, `[KẾT THÚC]` |
| `[OK]` | empty | Acknowledge | `[OK]` |
| `[PAUSE]` | N/A | Wait for signal | Used by framework |

**Note:** Signals are USER-DEFINED in crew.yaml, not built-in!

### 8.2 Signal Discovery

```go
// Get all signals for an agent
signals := ce.crew.Routing.Signals[agent_id]

// Check what's available
for _, sig := range signals {
    fmt.Printf("Agent %s emits: %s → %s\n",
        agent_id, sig.Signal, sig.Target)
}
```

---

## 9. PERFORMANCE ANALYSIS

### 9.1 Signal Matching Performance

| Operation | Complexity | Time |
|---|---|---|
| Exact match | O(n) | ~microseconds |
| Case-insensitive | O(n) | ~microseconds |
| Normalized bracket | O(n×m) | ~microseconds (m=bracket search window) |
| **Total per signal** | O(n) | ~microseconds |
| **Signals per agent** | O(k×n) | k signals, n content length |

**Optimization:** Uses fast `strings.Contains` first (Level 1 match), only falls back to normalization if needed.

### 9.2 Routing Logic Performance

```
Agent execution: ~50-500ms (LLM call)
                    ↓
Signal matching: ~microseconds ← NEGLIGIBLE
                    ↓
Handoff: ~microseconds ← NEGLIGIBLE
                    ↓
Next agent execution: ~50-500ms
```

**Bottleneck:** LLM execution, NOT signal routing.

### 9.3 Memory Usage

- `RoutingConfig` stored in Crew struct (loaded once at startup)
- No per-request allocations for routing
- Signal matching uses stack-allocated strings (no heap)

---

## 10. ERROR HANDLING & EDGE CASES

### 10.1 Invalid Routing Configuration

**Problem:** Signal target agent doesn't exist

```go
// findNextAgentBySignal at line 148-152
nextAgent := ce.findAgentByID(sig.Target)
if nextAgent != nil {
    log.Printf("[ROUTING] %s -> %s (signal: %s)", current.ID, nextAgent.ID, sig.Signal)
}
return nextAgent  // ← Can be nil!
```

**Behavior:** Returns nil, which triggers fallback routing:
```go
// If signal routing fails, use fallback routing
if nextAgent == nil {
    // Check handoff_targets
    // Check any other agent
    // Check parallel groups
}
```

### 10.2 No Signals Defined

**Routing config missing or empty:**

```go
if ce.crew.Routing == nil {
    return nil  // checkTerminationSignal
    return nil  // findNextAgentBySignal
}

signals, exists := ce.crew.Routing.Signals[current.ID]
if !exists || len(signals) == 0 {
    return nil  // No signals for this agent
}
```

**Behavior:** Agents without signals use fallback routing

### 10.3 Signal Not Found in Response

```go
// No matching signal in agent response
→ checkTerminationSignal returns nil
→ findNextAgentBySignal returns nil
→ Fallback to traditional routing (handoff_targets, next agent)
```

### 10.4 Circular Routing

**Config:** A → B → A (circle!)

```yaml
routing:
  signals:
    a:
      - signal: "[ROUTE_B]"
        target: b
    b:
      - signal: "[ROUTE_A]"
        target: a  # Circle!
```

**Behavior:**
- Handoff count tracking prevents infinite loops
- When `handoffCount >= MaxHandoffs`, stop
- No runtime detection of circles (design trade-off: flexibility vs safety)

---

## 11. BEST PRACTICES & DESIGN PATTERNS

### 11.1 When to Use Signal Routing

✅ **GOOD:**
- Sequential workflows with clear handoff points
- Expert-to-expert handoffs (teacher → student)
- Decision points (choose between routes based on content)
- Workflows with variable routing logic

❌ **NOT GOOD:**
- Simple linear pipelines (use handoff_targets instead)
- Unknown agent decisions (random routing)
- Feedback loops (implement differently)

### 11.2 Signal Design Patterns

**Pattern 1: Explicit Routes**
```yaml
signals:
  orchestrator:
    - signal: "[COMPLEX]"
      target: executor
    - signal: "[SIMPLE]"
      target: clarifier
```

**Pattern 2: Multi-Signal Agent**
```yaml
signals:
  teacher:
    - signal: "[QUESTION]"
      target: parallel_question
    - signal: "[END]"
      target: reporter
    - signal: "[DONE]"
      target: ""  # Terminate
```

**Pattern 3: Conditional Handoff**
```yaml
signals:
  analyzer:
    - signal: "[NEEDS_HUMAN]"
      target: human_reviewer
    - signal: "[AUTO_APPROVE]"
      target: approval_agent
```

### 11.3 Debugging Signal Issues

**Check 1: Is signal in response?**
```go
log.Printf("Agent response: %s", response.Content)
// Check if [SIGNAL] pattern is actually in response
```

**Check 2: Is routing config loaded?**
```go
if ce.crew.Routing == nil {
    log.Println("ERROR: Routing config not loaded!")
}
```

**Check 3: Trace routing decisions**
```
[ROUTING] teacher -> parallel_question (signal: [QUESTION])
[ROUTING] teacher -> TERMINATE (signal: [DONE])
[ROUTING] No next agent found for executor
```

---

## 12. CONTEXT PRESERVATION EXAMPLES

### Example: 3-Agent Quiz System

```yaml
agents: [teacher, student, reporter]
routing:
  signals:
    teacher:
      - signal: "[QUESTION]"
        target: parallel_question
    student:
      - signal: "[ANSWER]"
        target: parallel_answer
    reporter:
      - signal: "[DONE]"
        target: ""
```

**Execution trace with history:**

```
ROUND 1:
  History: []

1. User input: "Start quiz"
   History: [user: "Start quiz"]

2. Teacher executes
   Input: "Start quiz"
   History: [user: "Start quiz"]
   Response: "Question 1: What is 2+2? [QUESTION]"
   History: [user: "Start quiz", assistant: "Question 1: What is 2+2? [QUESTION]"]

3. Parallel execution [student, reporter]
   Student sees History: [user: "Start quiz", assistant: "Question 1..."]
   Reporter sees History: [user: "Start quiz", assistant: "Question 1..."]

   Student Response: "The answer is 4 [ANSWER]"
   Reporter Response: "Question 1 registered [OK]"

   History: [user: "Start quiz", assistant: "Question 1...",
             user: "[Student] Answer: 4\n[Reporter] Question registered"]

ROUND 2:
4. Teacher executes (resume)
   Input: "[Student] Answer: 4..."
   History: [user: "Start quiz", assistant: "Q1...", user: "Results...", ...]
   Response: "Correct! Next: What is 3+3? [QUESTION]"

   → Parallel execution again
   → History grows with each round
   → All agents see full conversation context
```

**Key insight:** Each agent decision influenced by complete history of previous decisions.

---

## 13. SUMMARY TABLE

| Aspect | Details |
|--------|---------|
| **Signal Definition** | String in brackets: `[PATTERN]` |
| **Storage** | `Routing.Signals[agent_id][]RoutingSignal` |
| **Matching** | Exact → Case-insensitive → Normalized (3-level) |
| **Termination** | Target = "" (empty string) |
| **Routing** | Target = agent_id (non-empty) |
| **Parallel** | Target = parallel_group_id |
| **Handoff Limit** | `MaxHandoffs` configuration |
| **Context Loss** | ✅ ZERO - Full history preserved |
| **Timeout** | Only on agent/tool execution, not routing |
| **Performance** | Negligible (~microseconds) |
| **Debug Signal** | Check logs for `[ROUTING]` entries |

---

## 14. FILES & CODE LOCATIONS

| Concept | File | Line |
|---------|------|------|
| Signal matching | `core/crew_routing.go` | 49-90 |
| Termination check | `core/crew_routing.go` | 98-124 |
| Routing check | `core/crew_routing.go` | 126-157 |
| Parallel detection | `core/crew_routing.go` | 202-223 |
| Execution loop | `core/crew.go` | 656-873 |
| Handoff counting | `core/crew.go` | 654, 791, 861 |
| History management | `core/crew.go` | 525-615 |
| Type definitions | `core/config.go` | 14-44 |

---

## 15. RELATED DOCUMENTATION

- **Hardcoded Defaults Fixes:** `phase_5_implementation`
- **Performance Metrics:** Issue #14
- **Error Handling:** Issue #5
- **Context Trimming:** Issue #1
- **Configuration Modes:** Phase 5.1 (Strict/Permissive)

