# Signal-Based Routing: Visual Diagrams

## 1. Signal Matching Algorithm

```
Agent Response Content
         │
         ▼
    ┌─────────────────────────────┐
    │ Level 1: Exact Match        │
    │ strings.Contains(content)   │
    │ "[ROUTE]" in response?      │
    └─────────────────────────────┘
         │
      YES│ ──────→ ✅ MATCH (return true)
         │
      NO │
         ▼
    ┌─────────────────────────────┐
    │ Level 2: Case-Insensitive   │
    │ strings.ToLower()           │
    │ "[route]" in response?      │
    └─────────────────────────────┘
         │
      YES│ ──────→ ✅ MATCH (return true)
         │
      NO │
         ▼
    ┌──────────────────────────────┐
    │ Level 3: Normalized Bracket  │
    │ Extract [...] patterns       │
    │ Collapse spaces              │
    │ "[  ROUTE  ]" == "[route]"  │
    └──────────────────────────────┘
         │
      YES│ ──────→ ✅ MATCH (return true)
         │
      NO │
         ▼
         ❌ NO MATCH (return false)
```

## 2. ExecuteStream Main Loop

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ START: ExecuteStream(ctx, userInput, streamChan)                            │
│        • Add user input to history                                          │
│        • Set currentAgent = entry agent (or resume agent)                   │
│        • handoffCount = 0                                                   │
└──────────────────────────────────────────────────────────────┬──────────────┘
                                                               │
                                                               ▼
                    ┌──────────────────────────────────────────────────────┐
                    │ Agent Loop: for { ... }                              │
                    │                                                      │
                    │ 1. Log agent_start event                            │
                    │ 2. Trim history if needed (Issue #1)                │
                    │ 3. Execute agent LLM call                           │
                    │    response := ExecuteAgent(agent, input, history)  │
                    │ 4. Record metrics (Issue #14)                       │
                    │ 5. Add response to history                          │
                    │ 6. Send agent_response event                        │
                    └────────────────────┬───────────────────────────────┘
                                         │
                                         ▼
                    ┌──────────────────────────────────────────────────────┐
                    │ DO WE HAVE TOOL CALLS?                               │
                    │                                                      │
                    │ if len(response.ToolCalls) > 0:                     │
                    │   • Execute each tool                               │
                    │   • Add tool results to history                     │
                    │   • Set input = toolResults                         │
                    │   • Continue loop (agent re-executes)               │
                    └────────────────────┬───────────────────────────────┘
                                         │
                        NO Tool Calls    │    Has Tool Calls
                    ────────────────────┴──────────────────
                    │                                      │
                    ▼                                      ▼
            (Skip to next step)              (Add to history, continue loop)
                    │
                    ▼
    ┌──────────────────────────────────────────────────────────┐
    │ CHECK TERMINATION SIGNAL (target = "")                   │
    │                                                          │
    │ sig.Target == "" && signalMatches(sig.Signal, response)? │
    │                                                          │
    │ Example: "[DONE]" target=""                             │
    └──────────────────────┬────────────────────────────────┘
                           │
            YES (terminate) │ NO (continue checking)
                   ┌────────┴─────────┐
                   │                  │
                   ▼                  ▼
            ┌────────────────┐  ┌──────────────────────────────┐
            │ RETURN nil     │  │ CHECK ROUTING SIGNAL         │
            │ End workflow   │  │ (target = agent_id)          │
            │ Log: TERMINATE │  │                              │
            └────────────────┘  │ sig.Target != "" &&          │
                                │ signalMatches(...)?          │
                                │                              │
                                │ Example: "[ROUTE_X]" →X      │
                                └──────────────┬───────────────┘
                                               │
                          YES (route found)    │    NO route
                                   ┌──────────┴──────────┐
                                   │                     │
                                   ▼                     ▼
                            ┌────────────────┐  ┌──────────────────────────────┐
                            │ currentAgent   │  │ CHECK wait_for_signal        │
                            │   = nextAgent  │  │                              │
                            │ handoffCount++ │  │ behavior.WaitForSignal == T? │
                            │ continue loop  │  │                              │
                            └────────────────┘  │ Example: Pause & await input │
                                                └──────────────┬───────────────┘
                                                               │
                                                YES │   NO
                                            ┌─────┴────────┐
                                            │              │
                                            ▼              ▼
                                       ┌─────────┐  ┌──────────────────────┐
                                       │ RETURN  │  │ CHECK is_terminal    │
                                       │ (PAUSE) │  │                      │
                                       │ with    │  │ agent.IsTerminal?    │
                                       │ agent_id│  │                      │
                                       └─────────┘  └──────────────┬───────┘
                                                                   │
                                                    YES │   NO
                                                ┌─────┴────────┐
                                                │              │
                                                ▼              ▼
                                           ┌─────────┐  ┌──────────────────────┐
                                           │ RETURN  │  │ CHECK parallel_groups│
                                           │ (END)   │  │                      │
                                           │Terminal │  │ targetGroup = sig.T? │
                                           └─────────┘  │                      │
                                                        │ Execute agents in    │
                                                        │ parallel, aggregate  │
                                                        │ results              │
                                                        │                      │
                                                        │ handoffCount++       │
                                                        │ currentAgent = next  │
                                                        │ continue loop        │
                                                        └──────────────┬───────┘
                                                                       │
                                                           Parallel found
                                                                       │
                                                          ┌────────────┘
                                                          │
                                                          ▼
                                    ┌───────────────────────────────┐
                                    │ No parallel, use fallback     │
                                    │ routing:                      │
                                    │                               │
                                    │ 1. handoff_targets from agent │
                                    │ 2. Any other agent            │
                                    │ 3. None → END                 │
                                    │                               │
                                    │ handoffCount++                │
                                    │ if >= MaxHandoffs: RETURN     │
                                    │ else: continue loop           │
                                    └───────────────────────────────┘
```

## 3. Signal Matching Example

```
Agent Response:
"I think we need to call the executor. [  ROUTE_EXECUTOR  ]"

Signal Definition:
[ROUTE_EXECUTOR] → executor

Matching Process:

Step 1: Exact Match?
  strings.Contains("I think...ROUTE_EXECUTOR...", "[ROUTE_EXECUTOR]")
  ✅ YES → MATCH FOUND

  Return nextAgent = executor
```

```
Agent Response:
"Let me route to [ Route_Executor ] please"

Signal Definition:
[ROUTE_EXECUTOR] → executor

Matching Process:

Step 1: Exact Match?
  strings.Contains("...[ Route_Executor ]...", "[ROUTE_EXECUTOR]")
  ❌ NO (different case, spaces)

Step 2: Case-Insensitive?
  strings.ToLower("...[ Route_Executor ]...").Contains(
    strings.ToLower("[ROUTE_EXECUTOR]")
  )
  ✅ YES → MATCH FOUND

  Return nextAgent = executor
```

```
Agent Response:
"The decision is [ ROUTE  EXECUTOR ]"

Signal Definition:
[ROUTE_EXECUTOR] → executor

Matching Process:

Step 1: Exact Match?
  ❌ NO

Step 2: Case-Insensitive?
  ❌ NO (internal spaces differ)

Step 3: Normalized Bracket Match?
  Extract: "ROUTE  EXECUTOR" from [ ROUTE  EXECUTOR ]
  Normalize: "ROUTE  EXECUTOR" → "route executor" (multiple spaces)
  Collapse: "route executor" → "route executor" (single space)
  Result: "[route executor]"

  Extract: "ROUTE_EXECUTOR" from [ROUTE_EXECUTOR]
  Normalize: "route_executor"

  Compare: "[route executor]" != "[route_executor]"
  ❌ NO MATCH

  Falls back to traditional routing
```

```
Agent Response:
"[ KẾT  THÚC  THI ]"

Signal Definition:
[KẾT THÚC THI] → reporter

Matching Process:

Step 1: Exact Match?
  ❌ NO (spaces differ)

Step 2: Case-Insensitive?
  ❌ NO (accented characters + spaces)

Step 3: Normalized Bracket Match?
  Extract from response: "KẾT  THÚC  THI"
  Normalize: "kết thúc thi" (lowercase, collapse spaces)

  Extract from signal: "KẾT THÚC THI"
  Normalize: "kết thúc thi" (lowercase)

  Compare: "[kết thúc thi]" == "[kết thúc thi]"
  ✅ YES → MATCH FOUND

  Return nextAgent = reporter
```

## 4. History Preservation Through Handoffs

```
START: User input "Start exam"
┌────────────────────────────────────┐
│ History:                           │
│  [0] user: "Start exam"            │
└────────────────────────────────────┘
           │
           ▼
    ┌────────────────────────────┐
    │ Agent 1: Teacher           │
    │ LLM Call Input: "Start..." │
    │ Sees History: [0]          │
    │ Response: [QUESTION]       │
    └────────────────────────────┘
           │
           ▼
┌────────────────────────────────────┐
│ History:                           │
│  [0] user: "Start exam"            │
│  [1] assistant: "[QUESTION]..."    │ ← Teacher response added
└────────────────────────────────────┘
           │
           ▼
    Parallel Execution: [Student, Reporter]
           │
    ┌──────┴──────┐
    │             │
    ▼             ▼
  Student       Reporter
  Sees:         Sees:
  Hist[0:2]     Hist[0:2]
  (Complete)    (Complete)
    │             │
    ▼             ▼
 Response:     Response:
 [ANSWER]      [OK]
           │
           ▼
┌────────────────────────────────────┐
│ History:                           │
│  [0] user: "Start exam"            │
│  [1] assistant: "[QUESTION]..."    │
│  [2] user: "Student: [ANSWER]..."  │ ← Both responses aggregated
│      "Reporter: [OK]..."           │
└────────────────────────────────────┘
           │
           ▼
    ┌────────────────────────────┐
    │ Agent 1: Teacher (resume)  │
    │ LLM Call Input: Aggregated │
    │ Sees History: [0:3]        │ ← Can see what everyone did!
    │ Response: [QUESTION]       │
    └────────────────────────────┘
           │
           ▼
     (Cycle repeats...)

KEY: Each agent sees complete history of all previous decisions
```

## 5. Handoff Limit Enforcement

```
MaxHandoffs = 5 Configuration

Agent Chain:
  Teacher → Student → Teacher → Reporter → Executor

Handoff Counting:

  Teacher executes
  (no handoff yet)
        │
        ▼ [SIGNAL found: route to Student]

  Student executes
  handoffCount = 1 (1 < 5) ✅ Continue
        │
        ▼ [SIGNAL found: route to Teacher]

  Teacher executes
  handoffCount = 2 (2 < 5) ✅ Continue
        │
        ▼ [SIGNAL found: route to Reporter]

  Reporter executes
  handoffCount = 3 (3 < 5) ✅ Continue
        │
        ▼ [SIGNAL found: route to Executor]

  Executor executes
  handoffCount = 4 (4 < 5) ✅ Continue
        │
        ▼ [SIGNAL: route to Student again?]

  Student would execute
  handoffCount = 5 (5 >= 5) ❌ STOP

  Return CrewResponse (end workflow)
  Log: "Max handoffs exceeded"
```

## 6. Signal Definition Flow

```
┌─────────────────────────────────────┐
│ crew.yaml                           │
│                                     │
│ routing:                            │
│   signals:                          │
│     agent_id:                       │
│       - signal: "[PATTERN]"         │
│         target: "next_agent"        │ target = "" for terminate
│         description: "..."          │
│       - signal: "[DONE]"            │
│         target: ""                  │
└──────────────┬──────────────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │ LoadCrewConfig()         │
    │ Parse YAML               │
    └──────────────┬───────────┘
               │
               ▼
    ┌──────────────────────────────────────────┐
    │ RoutingConfig populated:                 │
    │ {                                        │
    │   Signals: map[agent_id][]RoutingSignal  │
    │   AgentBehaviors: map[agent_id]Behavior  │
    │   ParallelGroups: map[id]GroupConfig     │
    │ }                                        │
    └──────────────┬───────────────────────────┘
               │
               ▼
    ┌──────────────────────────────────────────┐
    │ Crew.Routing = RoutingConfig             │
    │ CrewExecutor.crew.Routing set            │
    └──────────────┬───────────────────────────┘
               │
               ▼
    ┌──────────────────────────────────────────┐
    │ At Runtime:                              │
    │ - checkTerminationSignal()               │
    │ - findNextAgentBySignal()                │
    │ - findParallelGroup()                    │
    │ Access: ce.crew.Routing.Signals[agent_id]│
    └──────────────────────────────────────────┘
```

## 7. Performance Bottleneck Analysis

```
Agent Execution Timeline:

┌─────────────┬─────────────────┬──────────┬──────────┐
│ Prepare     │ LLM Call        │ Process  │ Routing  │
│ Input       │ (50-500ms)      │ Response │ (μs)     │
│ (µs)        │                 │ (µs)     │          │
└─────────────┴─────────────────┴──────────┴──────────┘
        ▲              ▲             ▲          ▲
        │              │             │          │
      <1%            95%+           <1%      <0.1%

   BOTTLENECK: LLM Call (Network latency + model inference)
   NOT: Signal routing (negligible microseconds)
```

## 8. Error Recovery Flow

```
Signal Target = "" (terminate) but target agent not found

Agent: executor
Signal: "[COMPLETE]"
Target: "" (termination signal)

Routing Behavior:
  1. checkTerminationSignal()
     - sig.Target == ""? YES
     - signalMatches("[COMPLETE]")? YES
     - Return TerminationResult{ShouldTerminate: true}

  2. ExecuteStream sees ShouldTerminate=true
     - Send terminate event
     - Return nil (workflow ends)

Result: ✅ Workflow terminates correctly
        Even though target is empty!


Signal Target = "unknown_agent" but agent doesn't exist

Agent: router
Signal: "[ROUTE_X]"
Target: "unknown_agent" (agent_id doesn't exist)

Routing Behavior:
  1. findNextAgentBySignal()
     - signalMatches("[ROUTE_X]")? YES
     - findAgentByID("unknown_agent")? NO
     - Return nil (next agent not found)

  2. Falls back to traditional routing
     - Check handoff_targets
     - Check any other agent
     - Or end execution

Result: ⚠️  Graceful fallback instead of crash
```

## 9. Parallel Group Execution Model

```
┌─────────────────────────────────────┐
│ Agent emits signal: "[QUESTION]"    │
│ Targets: parallel_question group    │
└──────────────┬──────────────────────┘
               │
               ▼
    ┌──────────────────────────────┐
    │ findParallelGroup()          │
    │ Check if target is a group   │
    │ Parallel Groups: {           │
    │   parallel_question: {...}   │
    │ }                            │
    └──────────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────────────────────┐
    │ Parallel Execution:                      │
    │ agents: [student, reporter]              │
    │ timeout: 30 seconds                      │
    │ wait_for_all: false                      │
    │                                          │
    │ ┌──────────────────┐  ┌────────────────┐│
    │ │ Student          │  │ Reporter       ││
    │ │ Executes in      │  │ Executes in    ││
    │ │ parallel         │  │ parallel       ││
    │ └──────────────────┘  └────────────────┘│
    │        │                    │            │
    │        └────────┬───────────┘            │
    │                 │                        │
    │                 ▼                        │
    │         Aggregate results                │
    │                 │                        │
    │                 ▼                        │
    │         Add to history                   │
    └──────────────┬──────────────────────────┘
               │
               ▼
    ┌──────────────────────────────┐
    │ Move to next_agent (if set)  │
    │ nextAgent specified in config│
    │ handoffCount++               │
    │ continue loop                │
    └──────────────────────────────┘
```

