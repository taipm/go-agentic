# Phân Tích Chi Tiết - Điểm Yếu Core Library
## Kèm Code Examples và Visual Diagrams

---

## PHẦN 1: VISUAL DIAGRAMS

### 1.1 - Vòng Lặp Vô Hạn (Infinite Loop Visualization)

```
Timeline of Quiz Example Execution:

Time      Agent      LLM Call            State Update       Result
────────────────────────────────────────────────────────────────────
17:08:29  Teacher    3,112 tokens        ❌ questions=10    No progress
          └─ Calls:  $0.10923            (no update)        
             GetQuizStatus()

17:08:38  Teacher    3,387 tokens        ❌ questions=10    Still stuck
          └ Calls:   $0.1194             (no update)        
             [QUESTION]

17:08:49  Teacher    3,473 tokens        ❌ questions=10    Looping
          └─ Calls:  $0.11823            (no update)        
             [ANSWER]

17:08:58  Teacher    3,650 tokens        ❌ questions=10    Out of control
          └─ Calls:  $0.13038            (no update)        
             [END_EXAM]                 (ignored)

17:09:28  Teacher    4,161 tokens        ❌ questions=10    Still broken
          └─ Calls:  ongoing...          (no update)        
             GetQuizStatus() AGAIN

              ...infinite loop...
```

**Pattern Recognition:**
- ✓ LLM is working (tokens increase, cost tracked)
- ❌ State is frozen (questions_remaining always 10)
- ❌ Tool effects not persisted
- ❌ Workflow doesn't notice state is unchanged

---

### 1.2 - Architecture Decomposition

#### Current Architecture (Broken)

```
┌──────────────────────────────────────────────────────────────┐
│                     WORKFLOW EXECUTION                        │
│ (core/workflow/execution.go)                                 │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ executeAgent(input)                                    │  │
│  │ ├─ agent.ExecuteAgent()  ────────────────┐            │  │
│  │ │  └─ Calls LLM provider                 │            │  │
│  │ │  └─ Parses tool calls                  │            │  │
│  │ │  └─ Extracts signals                   │            │  │
│  │ │     └─ Returns ToolCall[] to response  │            │  │
│  │ │        🔴 Results NOT executed here     │            │  │
│  │ │        🔴 Side effects NOT captured     │            │  │
│  │ │                                         │            │  │
│  │ └─ Process signals                       │            │  │
│  │    ├─ [QUESTION] emitted                 │            │  │
│  │    ├─ [ANSWER] emitted                   │            │  │
│  │    └─ [END_EXAM] emitted                 │            │  │
│  │       🔴 But no side effects verified    │            │  │
│  │                                          │            │  │
│  │ └─ Handoff decision                      │            │  │
│  │    ├─ Routes to next agent               │            │  │
│  │    └─ Recursive call: executeAgent("")   │            │  │
│  │       🔴 With EMPTY input!               │            │  │
│  │       🔴 History unchanged               │            │  │
│  │       🔴 State not reset                 │            │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                    TOOL EXECUTION (ORPHANED)                  │
│ (core/tools/executor.go)                                     │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ ExecuteToolCalls(toolCalls)                            │  │
│  │ ├─ For each tool call:                                │  │
│  │ │  └─ ExecuteTool(handler, args)                      │  │
│  │ │     └─ Executes handler function                   │  │
│  │ │     └─ Returns result: string                       │  │
│  │ │                                                      │  │
│  │ ├─ Returns: map[string]string results                │  │
│  │ │  🔴 NOT integrated into workflow history           │  │
│  │ │  🔴 NOT persisted                                   │  │
│  │ │  🔴 Results discarded after function returns       │  │
│  │                                                        │  │
│  │ ☝️ THIS FUNCTION IS NEVER CALLED FROM WORKFLOW!       │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│            EXECUTION STATE (METRICS ONLY)                     │
│ (core/state-management/execution_state.go)                   │
│                                                               │
│  ExecutionState {                                             │
│    ✓ StartTime      time.Time                               │
│    ✓ EndTime        time.Time                               │
│    ✓ RoundCount     int                                      │
│    ✓ HandoffCount   int                                      │
│    ✓ TotalDuration  time.Duration                           │
│    ✓ RoundMetrics   map[int]*RoundMetric                    │
│                                                               │
│    🔴 Missing: Domain state                                  │
│    🔴 Missing: Tool results                                  │
│    🔴 Missing: State updates                                 │
│    🔴 Missing: Termination signals                          │
│  }                                                            │
└──────────────────────────────────────────────────────────────┘
```

**The Problem**: Tool execution is isolated from workflow. State management doesn't know about tool results.

---

#### Required Architecture (Fixed)

```
┌──────────────────────────────────────────────────────────────┐
│                   WORKFLOW EXECUTION (ENHANCED)               │
│ (core/workflow/execution.go)                                 │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ executeAgent(input)                                    │  │
│  │                                                         │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ 1. AGENT EXECUTION                              │  │  │
│  │  │    agent.ExecuteAgent(input, history)           │  │  │
│  │  │    └─ Returns: response with ToolCall[]         │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │           │                                            │  │
│  │  ┌────────▼─────────────────────────────────────────┐  │  │
│  │  │ 2. TOOL ORCHESTRATION (NEW)                      │  │  │
│  │  │    toolResults = ExecuteToolCalls(response.Tools)│  │  │
│  │  │    ├─ Execute each tool ✓                        │  │  │
│  │  │    ├─ Capture results ✓                          │  │  │
│  │  │    └─ Add to History ✓                           │  │  │
│  │  │       execCtx.History.Append(                    │  │  │
│  │  │         Message{Role:Tool, Content:results}      │  │  │
│  │  │       )                                           │  │  │
│  │  └────────▼─────────────────────────────────────────┘  │  │
│  │           │                                            │  │
│  │  ┌────────▼─────────────────────────────────────────┐  │  │
│  │  │ 3. STATE PERSISTENCE (NEW)                       │  │  │
│  │  │    UpdateExecutionState(toolResults)             │  │  │
│  │  │    ├─ Atomic state update ✓                      │  │  │
│  │  │    ├─ Persist domain state ✓                     │  │  │
│  │  │    └─ Record state transitions ✓                 │  │  │
│  │  └────────▼─────────────────────────────────────────┘  │  │
│  │           │                                            │  │
│  │  ┌────────▼─────────────────────────────────────────┐  │  │
│  │  │ 4. SIGNAL & ROUTING                              │  │  │
│  │  │    ProcessSignals(response.Signals)              │  │  │
│  │  │    ├─ Emit signals ✓                             │  │  │
│  │  │    ├─ Process for routing ✓                      │  │  │
│  │  │    └─ Get routing decision ✓                     │  │  │
│  │  └────────▼─────────────────────────────────────────┘  │  │
│  │           │                                            │  │
│  │  ┌────────▼─────────────────────────────────────────┐  │  │
│  │  │ 5. TERMINATION CHECK (NEW)                       │  │  │
│  │  │    if IsTerminal(state, signals) {               │  │  │
│  │  │      return response  ✓ Exit                     │  │  │
│  │  │    }                                              │  │  │
│  │  │                                                  │  │  │
│  │  │    if HasHandoffTarget() {                       │  │  │
│  │  │      executeAgent(nextAgent, formattedState) ✓  │  │  │
│  │  │    }                                              │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│           EXECUTION STATE (COMPREHENSIVE)                     │
│ (core/state-management/execution_state.go - REDESIGNED)      │
│                                                               │
│  ExecutionState {                                             │
│    // Metrics (existing)                                      │
│    ✓ StartTime      time.Time                               │
│    ✓ RoundCount     int                                      │
│                                                               │
│    // NEW: Domain state                                       │
│    ✓ DomainState    map[string]interface{}                  │
│    ✓ StateHistory   []StateSnapshot                         │
│    ✓ ToolResults    map[string]interface{}                  │
│                                                               │
│    // NEW: Termination tracking                              │
│    ✓ IsTerminal     bool                                     │
│    ✓ TerminalReason string                                   │
│    ✓ TerminalSignal string                                   │
│  }                                                            │
└──────────────────────────────────────────────────────────────┘
```

---

### 1.3 - Data Flow Comparison

#### Current (Broken)

```
Quiz Tool State (Global/Local)
         │
         │ (isolated, not synced)
         ▼
    Local State Ptr: 0x1400007ab40
    ├─ CorrectAnswers: 0 (NOT UPDATED)
    ├─ CurrentQuestion: 0 (NOT UPDATED)
    └─ Questions: 10 (NOT DECREMENTED)
         │
         │ GetQuizStatus() called
         │ (Always returns initial state)
         ▼
    ExecutionContext
    ├─ History: [user_input, agent_response, ...]
    │  🔴 NO tool results appended
    │  🔴 NO state snapshot
    └─ State metadata only
         │
         │ (Agent doesn't see tool effects)
         ▼
    Next executeAgent() Round
    └─ Same input, same history
       → Same agent response
       → Same tool calls
       → Infinite loop!
```

#### Fixed

```
Quiz Tool State (Persistent)
         │
         │ (sync with execution state)
         ▼
    Persistent State Store
    ├─ CorrectAnswers: 0 → 1 → 2 → 3 (UPDATED)
    ├─ CurrentQuestion: 0 → 1 → 2 → 3 (UPDATED)
    └─ Questions: 10 → 9 → 8 → 7 (DECREMENTED)
         │
         │ RecordAnswer() called
         │ (Updates state atomically)
         ▼
    ExecutionContext
    ├─ History: [user_input, agent_response, tool_results, ...]
    │  ✓ Tool results appended
    │  ✓ State snapshot recorded
    └─ Full state with domain data
         │
         │ (Next agent sees context)
         ▼
    Next executeAgent() Round
    └─ Input: "Continue with Q3"
       History: [..., RecordAnswer result, GetQuizStatus: 7 remaining]
       → Agent knows progress
       → Doesn't repeat Q2
       → Moves to Q3
       → Loop terminates when questions_remaining == 0
```

---

## PHẦN 2: CODE EXAMPLES

### 2.1 - Problem Code (Current - Broken)

#### Location: `core/workflow/execution.go` lines 70-150

```go
// 🔴 CURRENT IMPLEMENTATION (BROKEN)
func executeAgent(ctx context.Context, execCtx *ExecutionContext, 
                  input string, apiKey string, 
                  agents map[string]*common.Agent) (*common.AgentResponse, error) {
    
    // Execute agent
    response, err := agent.ExecuteAgent(ctx, execCtx.CurrentAgent, 
                                       input, execCtx.History, apiKey)
    // 🔴 PROBLEM 1: response.ToolCalls are extracted but NOT EXECUTED
    // ToolCalls are just returned as data, no side effects captured
    
    // Add response to history
    execCtx.History = append(execCtx.History, common.Message{
        Role:    common.RoleAssistant,
        Content: response.Content,
    })
    // 🔴 PROBLEM 2: Tool results not added to history
    // Agent doesn't see what tools did
    
    // Process signals
    if execCtx.SignalRegistry != nil && response.Signals != nil {
        for _, sigName := range response.Signals {
            sig := &signal.Signal{
                Name:    sigName,
                AgentID: execCtx.CurrentAgent.ID,
            }
            execCtx.SignalRegistry.Emit(sig)
            // 🔴 PROBLEM 3: Signal emitted but no state update verification
            // [ANSWER] signal doesn't guarantee RecordAnswer() was called
        }
    }
    
    // Handle handoff
    if routingDecision != nil && routingDecision.NextAgentID != "" {
        nextAgent, err := lookupNextAgent(agents, routingDecision.NextAgentID, ...)
        execCtx.CurrentAgent = nextAgent
        execCtx.HandoffCount++
        
        return executeAgent(ctx, execCtx, "", apiKey, agents)
        // 🔴 PROBLEM 4: Empty string as input!
        // Agent doesn't know what happened or what to do next
        // 🔴 PROBLEM 5: History unchanged
        // New agent sees same history as before, makes same decisions
    }
    
    return response, nil
}
```

**Why it's broken:**
1. ToolCalls extracted but never executed
2. Tool results never added to history  
3. Signals emitted without verifying side effects
4. Handoff passes empty input and unchanged history
5. GetQuizStatus() always returns initial state because RecordAnswer() was never called

---

#### Location: `core/tools/executor.go` - The Orphaned Function

```go
// 🔴 THIS FUNCTION EXISTS BUT IS NEVER CALLED FROM WORKFLOW!
func ExecuteToolCalls(ctx context.Context, toolCalls []common.ToolCall, 
                      agentTools []interface{}) (map[string]string, error) {
    results := make(map[string]string)
    
    for _, call := range toolCalls {
        tool, exists := toolMap[call.ToolName]
        if !exists {
            // Tool not found - continue
            continue
        }
        
        // Execute the tool
        result, err := ExecuteTool(ctx, call.ToolName, tool, call.Arguments)
        if err != nil {
            // Log error but continue
            continue
        }
        
        // Store result
        results[call.ToolName] = result
    }
    
    return results, nil
    // 🔴 PROBLEM: Results returned but not integrated
    // - Not added to ExecutionContext.History
    // - Not persisted to ExecutionState
    // - Not available to next agent round
}
```

**Why it's broken:**
- Function exists but workflow never calls it
- Tool execution happens in LLM response parsing (agent layer)
- Tool results are lost immediately
- Next agent round starts with no knowledge of what tools did

---

### 2.2 - Fixed Code (Solution)

#### Enhanced Workflow Execution

```go
// ✅ FIXED IMPLEMENTATION
func executeAgent(ctx context.Context, execCtx *ExecutionContext, 
                  input string, apiKey string, 
                  agents map[string]*common.Agent) (*common.AgentResponse, error) {
    
    // ───────────────────────────────────────────────────────────────────
    // STEP 1: AGENT EXECUTION (unchanged)
    // ───────────────────────────────────────────────────────────────────
    response, err := agent.ExecuteAgent(ctx, execCtx.CurrentAgent, 
                                       input, execCtx.History, apiKey)
    if err != nil {
        return nil, err
    }
    
    // ───────────────────────────────────────────────────────────────────
    // STEP 2: TOOL ORCHESTRATION (NEW)
    // ───────────────────────────────────────────────────────────────────
    var toolResults map[string]interface{}
    
    if len(response.ToolCalls) > 0 {
        // Execute all tool calls from the agent response
        toolResults, err := tools.ExecuteToolCallsWithContext(
            ctx, 
            response.ToolCalls, 
            execCtx.CurrentAgent.Tools,
            execCtx,  // Pass execution context for state updates
        )
        if err != nil {
            // Log error but continue (partial success)
            execCtx.LogToolExecutionError(err)
        }
        
        // Add tool results to history so next agent sees them
        if len(toolResults) > 0 {
            toolResultMsg := common.Message{
                Role:    common.RoleAssistant,  // Or new RoleTool?
                Content: formatToolResults(toolResults),
                Metadata: map[string]interface{}{
                    "type":   "tool_results",
                    "tools":  toolResults,
                    "timestamp": time.Now(),
                },
            }
            execCtx.History = append(execCtx.History, toolResultMsg)
            
            // Update execution state with tool results
            execCtx.StateManager.RecordToolResults(
                response.ToolCalls,
                toolResults,
            )
        }
    }
    
    // ───────────────────────────────────────────────────────────────────
    // STEP 3: STATE PERSISTENCE (NEW)
    // ───────────────────────────────────────────────────────────────────
    currentState := execCtx.StateManager.GetCurrentState()
    execCtx.StateManager.PersistSnapshot(common.StateSnapshot{
        RoundNumber:   execCtx.RoundCount,
        AgentID:       execCtx.CurrentAgent.ID,
        DomainState:   currentState,
        ToolResults:   toolResults,
        History:       execCtx.History,
        Timestamp:     time.Now(),
    })
    
    // ───────────────────────────────────────────────────────────────────
    // STEP 4: SIGNAL & ROUTING (enhanced)
    // ───────────────────────────────────────────────────────────────────
    var routingDecision *common.RoutingDecision
    
    if execCtx.SignalRegistry != nil && response.Signals != nil {
        for _, sigName := range response.Signals {
            sig := &signal.Signal{
                Name:    sigName,
                AgentID: execCtx.CurrentAgent.ID,
            }
            
            // Verify state was updated before routing
            if shouldVerifySignal(sigName) {
                verified := execCtx.StateManager.VerifySignalEffect(sigName)
                if !verified {
                    // Log warning but continue
                    fmt.Printf("[WARN] Signal %s not verified\n", sigName)
                }
            }
            
            execCtx.SignalRegistry.Emit(sig)
            
            // Get routing decision
            decision, err := execCtx.SignalRegistry.ProcessSignal(ctx, sig)
            if err == nil && decision != nil {
                routingDecision = decision
                if decision.IsTerminal {
                    return response, nil
                }
                if decision.NextAgentID != "" {
                    break  // Found routing, stop processing signals
                }
            }
        }
    }
    
    // ───────────────────────────────────────────────────────────────────
    // STEP 5: TERMINATION CHECK (NEW)
    // ───────────────────────────────────────────────────────────────────
    isTerminal, reason := execCtx.CheckTermination()
    if isTerminal {
        execCtx.emitSignal(signal.SignalTerminal, map[string]interface{}{
            "reason": reason,
        })
        return response, nil
    }
    
    // ───────────────────────────────────────────────────────────────────
    // STEP 6: HANDOFF WITH CONTEXT
    // ───────────────────────────────────────────────────────────────────
    nextAgentID := ""
    
    if routingDecision != nil && routingDecision.NextAgentID != "" {
        nextAgentID = routingDecision.NextAgentID
    } else if len(execCtx.CurrentAgent.HandoffTargets) > 0 {
        nextAgentID = execCtx.CurrentAgent.HandoffTargets[0].ID
    }
    
    if nextAgentID != "" {
        nextAgent, err := lookupNextAgent(agents, nextAgentID, ...)
        if err != nil {
            return nil, err
        }
        
        execCtx.CurrentAgent = nextAgent
        execCtx.HandoffCount++
        
        // Format state for next agent
        nextInput := formatStateForNextAgent(
            execCtx.StateManager.GetCurrentState(),
            execCtx.History,
        )
        
        // ✅ Recursive call with proper context
        return executeAgent(ctx, execCtx, nextInput, apiKey, agents)
    }
    
    // No handoff
    return response, nil
}
```

---

#### Enhanced Tool Execution with State Tracking

```go
// ✅ NEW FUNCTION: Tool execution with state context
func ExecuteToolCallsWithContext(
    ctx context.Context,
    toolCalls []common.ToolCall,
    agentTools []interface{},
    execCtx *ExecutionContext,  // For state updates
) (map[string]interface{}, error) {
    
    results := make(map[string]interface{})
    toolMap := buildToolMap(agentTools)
    
    for _, call := range toolCalls {
        tool, exists := toolMap[call.ToolName]
        if !exists {
            continue
        }
        
        // Execute with state capture
        result, err := executeToolWithCapture(
            ctx,
            call.ToolName,
            tool,
            call.Arguments,
            execCtx.StateManager,  // Pass state manager
        )
        
        if err != nil {
            continue
        }
        
        results[call.ToolName] = result
        
        // ✅ NEW: Notify state manager of tool effect
        execCtx.StateManager.RecordToolEffect(
            call.ToolName,
            result,
        )
    }
    
    return results, nil
}

// ✅ NEW FUNCTION: Execute tool and capture side effects
func executeToolWithCapture(
    ctx context.Context,
    toolName string,
    tool interface{},
    args map[string]interface{},
    stateManager *StateManager,
) (interface{}, error) {
    
    // Take state snapshot before
    stateBefore := stateManager.GetCurrentState()
    
    // Execute tool
    result, err := ExecuteTool(ctx, toolName, tool, args)
    if err != nil {
        return nil, err
    }
    
    // Take state snapshot after
    stateAfter := stateManager.GetCurrentState()
    
    // Detect side effects
    sideEffects := detectSideEffects(stateBefore, stateAfter)
    
    // Validate tool did what it claims
    if toolName == "RecordAnswer" {
        if !sideEffects.ContainsKey("CorrectAnswers") {
            fmt.Printf("[WARN] RecordAnswer didn't update state\n")
        }
    }
    
    // Record the tool execution with its effects
    stateManager.RecordToolExecution(common.ToolExecution{
        ToolName:    toolName,
        Arguments:   args,
        Result:      result,
        StateBefore: stateBefore,
        StateAfter:  stateAfter,
        SideEffects: sideEffects,
        Timestamp:   time.Now(),
    })
    
    return result, nil
}
```

---

#### State Management Enhancement

```go
// ✅ NEW: Enhanced ExecutionState
type ExecutionState struct {
    // Existing fields
    StartTime      time.Time
    RoundCount     int
    HandoffCount   int
    
    // NEW: Domain state
    DomainState    map[string]interface{}  // Quiz state, conversation state, etc.
    StateHistory   []StateSnapshot         // Track state changes
    ToolResults    map[string]interface{}  // Latest tool results
    
    // NEW: Termination tracking
    IsTerminal     bool
    TerminalReason string
    TerminalSignal string
    
    mu sync.RWMutex
}

// ✅ NEW: State snapshot for history
type StateSnapshot struct {
    RoundNumber   int
    AgentID       string
    DomainState   map[string]interface{}
    ToolResults   map[string]interface{}
    History       []Message
    Timestamp     time.Time
}

// ✅ NEW: Check if workflow should terminate
func (es *ExecutionState) CheckTermination() (bool, string) {
    es.mu.RLock()
    defer es.mu.RUnlock()
    
    // Domain-specific termination check
    if val, ok := es.DomainState["quiz_complete"]; ok {
        if complete, ok := val.(bool); ok && complete {
            return true, "Quiz completed"
        }
    }
    
    // Signal-based termination
    if es.IsTerminal {
        return true, es.TerminalReason
    }
    
    return false, ""
}

// ✅ NEW: Record tool results with state update
func (es *ExecutionState) RecordToolEffect(toolName string, result interface{}) {
    es.mu.Lock()
    defer es.mu.Unlock()
    
    es.ToolResults[toolName] = result
    
    // Update domain state based on tool result
    // Example for quiz:
    if toolName == "RecordAnswer" {
        if answerResult, ok := result.(map[string]interface{}); ok {
            if correct, ok := answerResult["is_correct"].(bool); ok {
                if correct {
                    es.DomainState["correct_count"] = 
                        es.DomainState["correct_count"].(int) + 1
                }
            }
            if remaining, ok := answerResult["remaining"].(int); ok {
                es.DomainState["questions_remaining"] = remaining
                if remaining == 0 {
                    es.DomainState["quiz_complete"] = true
                }
            }
        }
    }
}
```

---

### 2.3 - Quiz Example with Proper State Flow

#### Before (Broken)

```go
// Quiz tool implementation (broken because state isolated)
var quizState = &QuizState{
    CorrectAnswers: 0,
    CurrentQuestion: 0,
    Questions: []string{"2+2?", "3+3?", ...},
}

func GetQuizStatus() string {
    return fmt.Sprintf("remaining: %d", len(quizState.Questions) - quizState.CurrentQuestion)
    // Always returns 10 because CurrentQuestion never updates!
}

func RecordAnswer(answer string) string {
    // Never called! Teacher's prompt doesn't trigger this
    // Even if called, state isolated from workflow
    return "recorded"
}

// Teacher execution
// Round 1: GetQuizStatus() → "remaining: 10"
// Round 2: GetQuizStatus() → "remaining: 10" (no progress!)
// Round 3: GetQuizStatus() → "remaining: 10" (infinite loop!)
```

#### After (Fixed)

```go
// Quiz tool with shared state manager
type QuizStateManager struct {
    state map[string]interface{}
    mu    sync.Mutex
}

var globalQuizState = &QuizStateManager{
    state: map[string]interface{}{
        "correct_count": 0,
        "current_question": 0,
        "total_questions": 10,
        "quiz_complete": false,
    },
}

func GetQuizStatus(ctx context.Context) string {
    remaining := globalQuizState.state["total_questions"].(int) - 
                 globalQuizState.state["current_question"].(int)
    return fmt.Sprintf("remaining: %d, correct: %d", 
                      remaining, 
                      globalQuizState.state["correct_count"])
}

func RecordAnswer(ctx context.Context, answer string) map[string]interface{} {
    globalQuizState.mu.Lock()
    defer globalQuizState.mu.Unlock()
    
    currentQ := globalQuizState.state["current_question"].(int)
    question := questions[currentQ]
    
    isCorrect := evaluateAnswer(question, answer)
    
    // Update state
    if isCorrect {
        globalQuizState.state["correct_count"] = 
            globalQuizState.state["correct_count"].(int) + 1
    }
    globalQuizState.state["current_question"] = currentQ + 1
    
    // Check if complete
    if currentQ + 1 >= 10 {
        globalQuizState.state["quiz_complete"] = true
    }
    
    return map[string]interface{}{
        "is_correct": isCorrect,
        "remaining": 10 - (currentQ + 1),
        "total_correct": globalQuizState.state["correct_count"],
    }
}

// Teacher execution with fixed workflow
// Round 1:
//   - executeAgent("Start quiz")
//   - Agent: "I'll ask Q1"
//   - Signal: [QUESTION]
//   - Calls: GetQuizStatus() → "remaining: 10"
//   - Tool results: appended to history ✓
//   - State update: recorded ✓
//   → Handoff to student/reporter
//
// Round 2:
//   - executeAgent("Student answered: 4")
//   - Agent sees history with GetQuizStatus result ✓
//   - Calls: RecordAnswer("4")
//   - Tool results: {"is_correct": true, "remaining": 9} ✓
//   - State update: current_question = 1, correct_count = 1 ✓
//   - History now shows: GetQuizStatus → RecordAnswer → Q2
//   → Loop continues with NEW input
//
// Round 3-10: Same pattern, questions_remaining decreases
//
// Round 11:
//   - executeAgent("Student answered: 20")
//   - Tool: RecordAnswer("20")
//   - State: current_question = 10, quiz_complete = true ✓
//   - CheckTermination() → true, "Quiz completed" ✓
//   - Signal: [END_EXAM]
//   → Workflow terminates ✓
```

---

## PHẦN 3: COMPARISON TABLE

### Functionality Matrix

| Feature | Current | Fixed | Impact |
|---------|---------|-------|--------|
| **Tool Execution** | Agent layer only | Workflow layer + Agent | Tools actually execute |
| **Tool Results** | Not integrated | Appended to History | Next agent sees results |
| **State Persistence** | Metrics only | Full domain state | State survives rounds |
| **Signal Verification** | Emitted only | Verified + Emitted | Signals reliable |
| **Termination** | Max rounds only | Domain-aware check | Proper exit conditions |
| **Handoff Context** | Empty input | Formatted state summary | Agent knows progress |
| **Cost Control** | Uncontrolled | Checkpoint-based | Bounded cost growth |
| **Debugging** | Logs only | State snapshots | Full audit trail |

---

## CONCLUSION

The core library's infinite loop issue is not a bug in the quiz example—it's a **predictable consequence of incomplete architecture**. The three critical missing pieces are:

1. **State Persistence Layer** - No mechanism to persist domain state between rounds
2. **Tool Orchestration Layer** - Tool execution isolated from workflow
3. **Termination Logic Layer** - No domain-aware termination checking

Without these three layers, any stateful multi-agent workflow will fail exactly like the quiz example.

---

