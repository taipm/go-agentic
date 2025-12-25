# Phân Tích Điểm Yếu Core Library - Go Agentic Framework
## Sử Dụng First Principles + 5W2H

---

## PHẦN 1: PHÂN TÍCH FIRST PRINCIPLES
### Nguyên Lý Cốt Lõi Của Hệ Thống

#### 🎯 MỤC TIÊU CỐT LÕI (Core Purpose)
- **Là gì**: Framework orchestration cho multi-agent workflow
- **Vấn đề được giải quyết**: Cho phép nhiều LLM agents phối hợp với nhau thực hiện tasks phức tạp
- **Yêu cầu thiết yếu**: State persistence, signal routing, tool execution, cost tracking

#### ⚙️ THÀNH PHẦN CỐT LÕI (Essential Components)

1. **State Management** - Trạng thái thực thi phải bền vững
   - Hiện tại: `ExecutionState` chỉ lưu metadata (timing)
   - Vấn đề: Không lưu lại task state (quiz, user data, etc.)

2. **Signal Routing** - Agents giao tiếp qua signals
   - Hiện tại: `SignalRegistry` phát hành signals nhưng handlers độc lập
   - Vấn đề: Không có synchronization giữa signal emission vs state update

3. **Tool Execution** - Thực thi tools phải có effects
   - Hiện tại: `ExecuteTool()` gọi handler nhưng không capture side effects
   - Vấn đề: Tool results không được persist, chỉ trả về string

4. **Agent Execution Loop** - Agent phải được chạy lặp lại
   - Hiện tại: `executeAgent()` trong workflow là recursive
   - Vấn đề: Vòng lặp không có termination conditions, infinite recursion có thể xảy ra

#### 🔄 DEPENDENCIES GỐC

```
Input (user_input)
    ↓
Workflow Executor (executeAgent)
    ├─→ Agent Execution (ExecuteAgent)
    │   ├─→ LLM Provider call
    │   ├─→ Signal Extraction (ExtractSignalsFromContent)
    │   └─→ Tool Call Parsing
    │
    ├─→ Tool Execution (ExecuteToolCalls)
    │   └─→ Individual Tool Execution (ExecuteTool)
    │
    ├─→ Signal Processing (SignalRegistry.ProcessSignal)
    │   └─→ Routing Decision (RoutingDecision)
    │
    └─→ Agent Handoff (Recursive executeAgent call)
        └─→ State Update? ❌ NOT FOUND
        └─→ Tool Result Persistence? ❌ NOT FOUND
```

---

## PHẦN 2: PHÂN TÍCH 5W2H
### Chi Tiết Vấn Đề Trong Quiz Example

### 🤔 WHAT - CÁI GÌ BỊ VỀ SAI?

**Vấn đề chính**: Infinite loop - `questions_remaining` luôn = 10
```
GetQuizStatus() → questions_remaining: 10
                  CorrectAnswers: 0
                  CurrentQuestion: 0
```

**Nguyên nhân**: `RecordAnswer()` tool không được gọi, hoặc không cập nhật state

---

### WHY - TẠI SAO?

**Cấp độ 1: Triệu chứng**
- ❌ Teacher lặp lại các bước (STEP 1-4) mà không thực thi RecordAnswer()
- ❌ State pointer `0x1400007ab40` không bao giờ thay đổi
- ❌ GetQuizStatus() luôn trả về state ban đầu

**Cấp độ 2: Nguyên nhân gốc rễ**

| Cơ chế | Vấn đề | Tác động |
|---------|--------|---------|
| **Agent thinking blocks dài** | Teacher lập kế hoạch nhưng không thực thi | Tool calls bị skip |
| **Pseudo-code trong thinking** | Nó viết `Call RecordAnswer()` nhưng không emit tool call signal | State không update |
| **Prompt không explicit** | Không có instruction "THỰC TỰC execute these tools" | Agent chỉ lên kế hoạch, không action |
| **State isolation** | GetQuizStatus() đọc từ pointer riêng không được sync | Mỗi call lại reset |
| **No termination logic** | Workflow không biết khi nào dừng lặp | Infinite loop |

**Cấp độ 3: Sâu hơn - Architecture flaws**

1. **State Management Architecture**
   ```
   Problem: ExecutionState ⟂ Tool Results
   
   ExecutionState (metric-only):
   ├─ RoundCount ✓
   ├─ HandoffCount ✓
   └─ Timing ✓
   
   Tool Results: Ephemeral in memory
   ├─ Not persisted
   ├─ Not accessible across rounds
   └─ Lost after handler finishes
   ```

2. **Signal Flow Architecture**
   ```
   Problem: Signal emission ≠ State update
   
   Current:
   signal.Emit([ANSWER])
   └─ No guarantee handler updates state
   
   Expected:
   signal.Emit([ANSWER])
   └─ Triggers state atomic update
   └─ Confirms update before continuing
   ```

3. **Tool Execution Architecture**
   ```
   Problem: Tool results not captured in execution context
   
   Current:
   ExecuteToolCalls() → results map
   └─ Results printed/logged
   └─ But not stored in ExecutionContext.History
   └─ So next agent doesn't see them
   
   Expected:
   ExecuteToolCalls() → results
   └─ Append to History
   └─ Available in next executeAgent() call
   ```

---

### WHERE - Ở ĐẦU?

#### **In Code - Các File Liên Quan**

| Vị trí | File | Vấn đề |
|--------|------|--------|
| **State Storage** | `core/state-management/execution_state.go` | Chỉ lưu metrics, không lưu quiz state |
| **Tool Execution** | `core/tools/executor.go` | Trả về results nhưng không persist |
| **Workflow** | `core/workflow/execution.go` | `executeAgent()` không capture tool results |
| **Agent Execution** | `core/agent/execution.go` | Signal extraction nhưng không state update |
| **Signal Processing** | `core/signal/registry.go` | Phát hành signal, không validate side effects |

#### **In Architecture**
```
Layer 1: Agent Execution (agent/execution.go)
└─ Calls: ExecuteAgent()
└─ Returns: AgentResponse
└─ Ví dụ: [QUESTION] "Số nào là 2+2?" được tạo

↓ (Missing: Tool execution context capture)

Layer 2: Tool Execution (tools/executor.go)
└─ Calls: ExecuteToolCalls() 
└─ Returns: map[string]string results
└─ Ví dụ: RecordAnswer() trả về "success"

↓ (Missing: Result persistence)

Layer 3: Workflow State (state-management/execution_state.go)
└─ Stores: Timing, round count
└─ Missing: Quiz state, correct answers count

↓ (Missing: State sync back to tool)

Layer 4: Next Round (workflow/execution.go)
└─ Calls: executeAgent() lại
└─ Reads: ExecutionState.RoundCount
└─ Result: Vòng lặp không thoát, questions_remaining = 10 mãi
```

---

### WHO - AI CHỊU TRÁCH NHIỆM?

**Về Design:**
- Core library architects - Chưa define state persistence model
- Workflow execution designer - Chưa implement tool-result ↔ state mapping

**Về Example Implementation:**
- Example creator - Chưa check state was updated
- Teacher agent prompt - Không rõ `RecordAnswer()` phải được thực thi

**Về Integration:**
- Quiz tool state management - Không expose state updates
- Signal registry - Không enforce "no-side-effects" detection

---

### WHEN - KHI NÀO PHÁT SINH?

**Timeline của vấn đề:**

```
17:08:29 - ExecuteAgent() called, teacher starts thinking
           ↓
17:08:38 - Agent emits [QUESTION], [ANSWER], [END_EXAM] signals
           ↓
17:08:38 - GetQuizStatus() called → still 10 questions
           ↓
17:08:49 - Teacher emits [ANSWER], [END_EXAM] signals again
           ↓
17:08:49 - GetQuizStatus() still returns 10 questions
           ↓
17:08:58 - Loop continues...
           ↓
17:09:28 - Still looping, no progress
           ↓
Workflow never terminates ❌
```

**Khi nào vấn đề trở nên critical:**
- Khi quiz state không được save → không có single source of truth
- Khi tool results không persistent → next iteration không biết gì xảy ra
- Khi state update không atomic → race conditions có thể xảy ra

---

### HOW - LÀM CÁCH NÀO NÓ XẢY RA?

#### **Mechanism 1: State Isolation**
```go
// Problem: State is local to tool handler
func GetQuizStatus() string {
    // state là local variable trong handler
    // Không được shared với ExecutionContext
    return fmt.Sprintf("remaining: %d", state.Remaining)
}

// Solution: State phải persist globally
type QuizState struct {
    mu sync.RWMutex
    CorrectAnswers int
    Questions []Question
}

var globalQuizState QuizState  // Accessible across all tools
```

**Impact**: Mỗi lần gọi GetQuizStatus(), nó đọc từ initial state, không phải updated state

---

#### **Mechanism 2: Tool Results Not Propagated**
```go
// Current flow:
ExecuteToolCalls() → results: map[string]string
└─ Results logged
└─ NOT added to history
└─ NOT accessible to next agent round

// Expected flow:
ExecuteToolCalls() → results: map[string]string
└─ Format as Message
└─ Add to ExecutionContext.History
└─ Next executeAgent() call sees history
```

**Problem in code** (`workflow/execution.go`):
```go
// executeAgent() doesn't call ExecuteToolCalls at all!
// Tool execution happens in agent layer, results stay local
response, err := agent.ExecuteAgent(ctx, execCtx.CurrentAgent, input, execCtx.History, apiKey)
// execCtx.History is NOT updated with tool results
```

---

#### **Mechanism 3: Signal ≠ State Atomic Operation**
```go
// Problem: Signal emission and state update are separate
execCtx.emitSignal("record:answer", metadata)  // ← Emitted
RecordAnswer(args)  // ← State update
// No guarantee order, no atomicity

// Solution: Atomic signal + state update
func (ctx *ExecutionContext) RecordAnswerAtomic(answer string) error {
    mu.Lock()
    defer mu.Unlock()
    
    // 1. Update state
    err := state.RecordAnswer(answer)
    
    // 2. Emit signal (guaranteed after state update)
    ctx.emitSignal("record:answer", map{...})
    
    // 3. Return
    return err
}
```

---

#### **Mechanism 4: Infinite Loop - No Termination Check**
```go
// Current in workflow/execution.go:
func executeAgent(...) (*common.AgentResponse, error) {
    if execCtx.RoundCount >= execCtx.MaxRounds {
        return nil, NewQuotaExceededError()  // ✓ Has max rounds check
    }
    
    // Execute agent
    response, _ := agent.ExecuteAgent(...)
    
    // Process signals and route to next agent
    if routingDecision != nil && routingDecision.NextAgentID != "" {
        execCtx.CurrentAgent = nextAgent
        execCtx.HandoffCount++
        return executeAgent(ctx, execCtx, "", apiKey, agents)  // ← Recursive
    }
    
    // BUT PROBLEM: GetQuizStatus() doesn't indicate quiz completion
    // So routingDecision is never terminal
    // → Infinite loop within single round due to repeated GetQuizStatus() calls
}
```

**Root cause**: Teacher agent's prompt causes it to repeatedly call GetQuizStatus() without calling RecordAnswer() first

---

### HOW MUCH - BAO NHIÊU?

#### **Cost Impact**
```
Current logs show:
- Round 1: 3,112 tokens → $0.10923
- Round 2: 3,387 tokens → $0.1194
- Round 3: 3,473 tokens → $0.11823
- Round 4: 3,650 tokens → $0.13038  ← Token growth
- Round 5: 4,161 tokens → ongoing

Pattern: +300-500 tokens per round (context window growth)
Loop continues indefinitely → Cost approaches ∞
```

#### **Time Impact**
```
17:08:29 → 17:09:28 = 59 seconds for ~4 iterations
Average: ~15 seconds per iteration
If loop runs for 1 minute: Cost ≈ $1+
If loop runs for 10 minutes: Cost ≈ $10+
```

#### **State Space Explosion**
```
Without tool result persistence:
- ExecutionContext.History grows but tool results are lost
- GetQuizStatus() keeps returning initial state
- State becomes increasingly stale
- Divergence between "what system thinks" vs "what really happened"
```

---

## PHẦN 3: ĐIỂM YẾU CHÍNH CỦA CORE LIBRARY

### 📋 RANKED WEAKNESSES (By Impact & Severity)

#### **TIER 1: CRITICAL (Breaks Core Functionality)**

##### 1️⃣ **State Persistence Architecture**
**Problem**: Core library không định nghĩa how to persist domain state

```
What's missing:
├─ Execution state chỉ lưu metrics, không lưu:
│  ├─ Tool results
│  ├─ Domain data (quiz answers, conversation state)
│  ├─ Intermediate computation results
│  └─ State transitions
│
└─ Result: 
   └─ Quiz state không được lưu
   └─ RecordAnswer() effects không persist
   └─ GetQuizStatus() always returns initial state
```

**Where in code**:
- `core/state-management/execution_state.go` - Chỉ định metrics
- `core/workflow/execution.go` - Không capture tool results
- `core/tools/executor.go` - Results trả về string, không structured

**Impact**: 
- ❌ Multi-round workflows fail (can't remember previous actions)
- ❌ State divergence (agent's mental model ≠ actual state)
- ❌ Infinite loops (GetQuizStatus returns old state)

**Severity**: 🔴 **CRITICAL** - Framework unworkable for stateful applications

---

##### 2️⃣ **Tool Result Integration Gap**
**Problem**: Tool execution results không được integrated vào execution context

```
Current flow:
ExecuteToolCalls() [tools/executor.go]
├─ Executes tools
├─ Returns results map[string]string
└─ BUT: Results not added to ExecutionContext.History ❌

Expected flow:
ExecuteToolCalls() [must be in workflow/execution.go]
├─ Executes tools
├─ Formats results as Message
├─ Appends to ExecutionContext.History ✓
└─ Next agent sees full context
```

**Where in code**:
- `core/tools/executor.go:ExecuteToolCalls()` - Missing History append
- `core/workflow/execution.go:executeAgent()` - Tool execution logic absent
- `core/agent/execution.go` - Only handles LLM response, not tool results

**Impact**:
- ❌ Next agent doesn't see tool results
- ❌ Tool side effects lost
- ❌ Loop can't progress (no feedback from tools)

**Severity**: 🔴 **CRITICAL** - Tool execution becomes noop

---

##### 3️⃣ **Signal-State Synchronization**
**Problem**: Signals emitted but no guarantee of state update

```
Current:
├─ signal.Emit([QUESTION]) 
├─ signal.Emit([ANSWER])
│  └─ No guarantee RecordAnswer() actually executed
├─ signal.Emit([END_EXAM])
└─ Result: Multiple signals without side effects ❌

Expected:
├─ signal.Emit([QUESTION])
├─ signal.Emit([ANSWER])
│  └─ Atomic with state.RecordAnswer()
├─ signal.Emit([END_EXAM])
└─ Result: State consistent with signals ✓
```

**Where in code**:
- `core/signal/registry.go` - Emit vs ProcessSignal separated
- `core/workflow/execution.go` - Signals processed but no state capture
- `core/signal/handler.go` - Handlers don't update execution state

**Impact**:
- ❌ Race conditions (signal processed before state updated)
- ❌ Observer pattern without side effects
- ❌ Signals become decorative (no real effects)

**Severity**: 🔴 **CRITICAL** - Signals unreliable for routing

---

#### **TIER 2: MAJOR (Causes Degradation)**

##### 4️⃣ **Infinite Loop Conditions**
**Problem**: Workflow doesn't have proper termination logic

```
Current implementation:
├─ Max rounds check: ✓ Exists
├─ Max handoffs check: ✓ Exists
└─ Domain termination check: ❌ MISSING
   └─ How to detect "quiz is complete"?
   └─ How to detect "goal reached"?

Symptom:
GetQuizStatus() returns {remaining: 10}
├─ Not a terminal signal
├─ Not captured by workflow termination logic
└─ Agent loops indefinitely
```

**Where in code**:
- `core/workflow/execution.go:executeAgent()` - No domain termination check
- `core/agent/execution.go` - Teacher prompt doesn't emit terminal signal
- Logic is missing, not incorrect

**Impact**:
- ❌ Infinite loops on stateful workflows
- ❌ Cost explosion (uncontrolled token usage)
- ❌ Time explosion (job hangs)

**Severity**: 🟠 **MAJOR** - Framework unusable for long-running tasks

---

##### 5️⃣ **Recursive Execution Without Context Reset**
**Problem**: Each recursive call doesn't clear intermediate state

```
Current:
Round 1: executeAgent() 
├─ input = "Start quiz"
├─ History = [user_msg]
└─ Calls: agent.ExecuteAgent()
   
Round 2: executeAgent() 
├─ input = "" (empty!)
├─ History still has old messages
└─ Agent tries to answer empty input ❌

Expected:
├─ Clear or reset context between rounds
├─ Or pass clear signals about what changed
└─ Agent can track state progression
```

**Where in code**:
- `core/workflow/execution.go:executeAgent()` L142-150
  ```go
  // After handoff:
  execCtx.CurrentAgent = nextAgent
  execCtx.HandoffCount++
  return executeAgent(ctx, execCtx, "", apiKey, agents)  // ← Empty input!
  ```

**Impact**:
- ❌ Agent doesn't know state changed
- ❌ Repeated tool calls (no progress)
- ❌ Context pollution (old messages affect new agent)

**Severity**: 🟠 **MAJOR** - Handoff semantics unclear

---

##### 6️⃣ **No Tool Execution Orchestration Layer**
**Problem**: Tools executed at agent layer, not workflow layer

```
Current architecture:
Agent (agent/execution.go)
└─ Extracts tool calls from LLM response ✓
└─ NOT implemented here - just returns ToolCall array
   
Workflow (workflow/execution.go)
└─ Should execute tools here ❌
└─ But doesn't - delegates back to agent
   
Tool (tools/executor.go)
├─ ExecuteToolCalls() exists
└─ But not called from workflow
```

**Where in code**:
- `core/agent/execution.go:ExecuteAgent()` - Returns response with tool calls
- `core/workflow/execution.go:executeAgent()` - Line 90-120: No tool execution
- `core/tools/executor.go` - Tool execution functions exist but unused

**Impact**:
- ❌ Tool results not integrated into workflow
- ❌ No tool error handling at workflow level
- ❌ No tool retry mechanism
- ❌ Impossible to enforce tool execution order

**Severity**: 🟠 **MAJOR** - Tool execution bypassed in workflow

---

#### **TIER 3: MODERATE (Limits Flexibility)**

##### 7️⃣ **Message Type Flexibility**
**Problem**: Tool results are strings, not structured messages

```
Current:
map[string]string → "tool_name": "result string"
├─ Lost: Tool metadata (execution time, status code)
├─ Lost: Structured data (objects, arrays)
└─ Agent must parse strings

Expected:
interface{} → Can be any structured type
├─ Preserve: Tool metadata
├─ Preserve: Data structure
└─ Type-safe agent processing
```

**Where in code**:
- `core/tools/executor.go:ExecuteToolCalls()` - Returns `map[string]string`
- `core/common/types.go:ToolCall` - Arguments are `map[string]interface{}`
- Result types are asymmetric

**Impact**:
- ⚠️ Tool metadata lost
- ⚠️ Complex tool results require parsing
- ⚠️ Type safety reduced

**Severity**: 🟡 **MODERATE** - Works but inflexible

---

##### 8️⃣ **Cost Tracking Incomplete**
**Problem**: Cost tracking exists but not enforced at workflow level

```
Current:
├─ Cost calculated per LLM call ✓
├─ Logged in metrics ✓
└─ But NOT:
   ├─ Enforced by quota
   ├─ Checked before each tool call
   └─ Prevented from exceeding budget

Example: Tool can cost $5 but framework doesn't check
if it would exceed budget before executing.
```

**Where in code**:
- `core/cost/tracker.go` - Tracks cost
- `core/cost/budget.go` - Budget config exists
- `core/executor/executor.go` - Not checked before execution
- `core/workflow/execution.go` - No cost guards

**Impact**:
- ⚠️ Cost overruns possible
- ⚠️ No per-tool cost limits
- ⚠️ No budget enforcement

**Severity**: 🟡 **MODERATE** - Tracking works, enforcement missing

---

##### 9️⃣ **Signal Registry Coupling**
**Problem**: Core workflow tightly coupled to signal registry

```
Current:
ExecuteWorkflow(..., signalRegistry *signal.SignalRegistry)
├─ signalRegistry is optional parameter
├─ If nil, signals not processed
└─ But workflow hardcoded to call SignalRegistry methods

Expected:
├─ Signal registry injected as interface
├─ Workflow works with or without registry
├─ Multiple registry implementations possible
```

**Where in code**:
- `core/workflow/execution.go:ExecuteWorkflow()` - signalRegistry parameter
- Direct calls to `execCtx.SignalRegistry.Emit()` without interface
- `core/signal/registry.go` - Concrete type, not interface

**Impact**:
- ⚠️ Hard to test without real signal registry
- ⚠️ Can't substitute different routing mechanisms
- ⚠️ Tight coupling

**Severity**: 🟡 **MODERATE** - Works for current needs, not extensible

---

##### 🔟 **Agent Configuration Validation Gaps**
**Problem**: Not all agent configuration options validated

```
Current:
├─ Agent.Name validated
├─ Agent.Role validated
├─ Agent.Tools validated (partially)
└─ BUT NOT:
   ├─ Tool.Func is actually callable
   ├─ Tool parameters match expected schema
   ├─ Handoff targets exist
   └─ Circular handoff dependencies

Example: Agent A handoff to Agent B, Agent B handoff to Agent A
→ Infinite loop at runtime, not caught at config time
```

**Where in code**:
- `core/validation/agent.go` - Validation exists but incomplete
- `core/tools/executor.go:ExecuteTool()` - Validates at runtime, not config time
- No pre-flight check for circular dependencies

**Impact**:
- ⚠️ Runtime errors caught late
- ⚠️ Debugging harder
- ⚠️ Configuration mistakes not prevented

**Severity**: 🟡 **MODERATE** - Error handling exists but detection later

---

### 📊 SUMMARY TABLE

| # | Weakness | Type | Severity | Impact | Fix Effort |
|---|----------|------|----------|--------|-----------|
| 1 | State Persistence | Architecture | 🔴 CRITICAL | State loss | High |
| 2 | Tool Result Integration | Architecture | 🔴 CRITICAL | Tool noop | High |
| 3 | Signal-State Sync | Architecture | 🔴 CRITICAL | Signals unreliable | High |
| 4 | Infinite Loop Conditions | Logic | 🟠 MAJOR | Cost/time explosion | Medium |
| 5 | Recursive Context Reset | Logic | 🟠 MAJOR | Handoff broken | Medium |
| 6 | Tool Orchestration Layer | Architecture | 🟠 MAJOR | Tools bypassed | High |
| 7 | Message Type Flexibility | Design | 🟡 MODERATE | Inflexible | Low |
| 8 | Cost Tracking Enforcement | Feature | 🟡 MODERATE | Overrun risk | Medium |
| 9 | Signal Registry Coupling | Coupling | 🟡 MODERATE | Testability | Low |
| 10 | Agent Config Validation | Validation | 🟡 MODERATE | Late errors | Low |

---

## PHẦN 4: ROOT CAUSE ANALYSIS (5W2H Summary)

### **The Core Problem Statement**

**5W2H Analysis Result:**

| Question | Answer |
|----------|--------|
| **WHAT** | Infinite loop: `questions_remaining` stays 10, state never updates |
| **WHY** | Tool results not persisted; state isolation; signals without side effects |
| **WHERE** | State-management, workflow, tools, signal layers all fragmented |
| **WHO** | Architecture: Missing state contract; Example: No RecordAnswer call |
| **WHEN** | Happens immediately when stateful workflow starts |
| **HOW** | 3 mechanisms: (1) State isolation (2) Results not propagated (3) Signal ≠ State |
| **HOW MUCH** | Cost grows $0.1-0.13 per iteration, infinite loop = unbounded cost |

### **Connecting the Dots**

```
Root Cause Chain:
┌─────────────────────────────────────────────────────────────┐
│ Core Library Design Flaws                                   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ❌ State Persistence Not Defined                           │
│     ↓                                                        │
│  ❌ Tool Results Not Integrated into ExecutionContext       │
│     ↓                                                        │
│  ❌ Signals Emitted Without Enforcing State Updates         │
│     ↓                                                        │
│  ❌ Workflow Doesn't Check Domain Termination              │
│     ↓                                                        │
│  ❌ Quiz Example: RecordAnswer() Not Called                │
│     ↓                                                        │
│  ❌ GetQuizStatus() Always Returns Initial State            │
│     ↓                                                        │
│  ⚡ RESULT: Infinite Loop                                   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## PHẦN 5: ARCHITECTURAL IMPLICATIONS

### **Missing Abstraction Layers**

The core library is missing 3 critical layers:

```
Current (Broken):
┌─────────────────┐
│ Agent Layer     │  ← Executes LLM
└────────┬────────┘
         │
┌────────▼────────┐
│ Workflow Layer  │  ← Routes agents (but no tool orchestration)
└────────┬────────┘
         │
┌────────▼──────────────┐
│ Execution State       │  ← Only tracks metrics
└───────────────────────┘

Missing:
┌──────────────────────────┐
│ State Management Layer    │  ← Persist domain state
└──────┬───────────────────┘
       │
┌──────▼───────────────────┐
│ Tool Orchestration Layer │  ← Execute & integrate tool results
└──────┬───────────────────┘
       │
┌──────▼──────────────────┐
│ Termination Check Layer  │  ← Domain-specific termination logic
└─────────────────────────┘
```

### **Contract That's Missing**

**What the framework needs to define:**

```go
// 1. State Update Contract
type StateUpdate struct {
    AgentID    string
    ToolName   string
    Results    interface{}
    Timestamp  time.Time
}

// 2. Tool Execution Contract
type ToolExecution struct {
    ToolCall    ToolCall
    Result      interface{}
    Error       error
    SideEffects StateUpdate
}

// 3. Domain Termination Contract
type TerminationChecker interface {
    IsTerminal(ctx context.Context, state interface{}) bool
    Reason() string
}
```

Without these contracts, the framework can't guarantee:
- ❌ State consistency
- ❌ Tool effect causality
- ❌ Workflow termination

---

## PHẦN 6: RECOMMENDATIONS (By Priority)

### **CRITICAL FIXES (Do First)**

1. **Implement State Persistence Layer**
   - Define `ExecutionState` to include domain state, not just metrics
   - Implement atomic state updates
   - Persist state between rounds

2. **Integrate Tool Results into Workflow**
   - Move tool execution from agent layer to workflow layer
   - Add tool results to ExecutionContext.History
   - Ensure next agent sees all tool outputs

3. **Enforce Signal-State Atomicity**
   - Create atomic signal + state update operations
   - Validate state updates before emitting signals
   - Rollback on signal failures

### **MAJOR FIXES (Do Second)**

4. Add domain termination logic to workflow
5. Implement tool orchestration layer with error handling
6. Fix recursive context passing

### **CLEANUP (Do Third)**

7. Convert signal registry to interface
8. Add comprehensive config validation
9. Enforce cost budgets
10. Improve message type system

---

## CONCLUSION

**The core library has a fundamental architectural flaw:**

> The framework orchestrates agent execution but **does NOT orchestrate state management**. It treats state as external to the framework, causing the infinite loop in the quiz example.

**The quiz infinite loop is not a bug—it's the predictable result of an incomplete architecture.**

---

*Analysis generated using First Principles + 5W2H methodology*
*Date: 2025-12-25*
