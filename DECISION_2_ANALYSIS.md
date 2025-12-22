# 🎯 Decision #2 Analysis: Budget Hierarchy

## The Question

**Khi cả Agent và Crew limits tồn tại, cái nào là PRIMARY?**

```
Scenario:
┌─────────────────────────────────────────┐
│ Crew Setup:                             │
│ • MaxCostPerExecution: $2.50            │
│ • MaxCostPerDay: $100/day               │
│                                         │
│ Agent Setup:                            │
│ • Router: MaxCostPerDay: $10            │
│ • FAQ: MaxCostPerDay: $10               │
│ • KB: MaxCostPerDay: $10                │
│ • Aggregator: MaxCostPerDay: $10        │
│                                         │
│ Problem: Các agent có thể request $40   │
│ nhưng crew chỉ allow $2.50!             │
│                                         │
│ Ai thắng?                               │
└─────────────────────────────────────────┘
```

---

## Option B: Crew Limit is Hard Cap

### Định nghĩa
```
Crew limit = ABSOLUTE MAXIMUM
           ↓
Agent limits = Informational only
           ↓
Nếu execution vượt crew limit → BLOCKED
Nếu agent limit vượt nhưng crew OK → ALLOWED (+ warning)
```

### Ví dụ Flow

**Scenario: Agent muốn gọi, tổng sẽ là $3 nhưng crew budget còn $2.50**

```go
ExecuteAgent(agent, history):
  1. Check agent limit ($10/day)
     └─ OK, agent used $2 today

  2. Check crew limit (HARD CAP!)
     ├─ Current execution would cost: $3
     ├─ Crew MaxCostPerExecution: $2.50
     ├─ Crew used so far: $2.40
     ├─ Remaining: $0.10
     └─ $3 > $0.10 → BLOCK! ❌

  3. Log: "Crew hard cap exceeded for this execution"

  4. Return error to user
```

### Triển khai

```go
type Crew struct {
    MaxCostPerExecution  float64  // 🔴 HARD CAP
    MaxCostPerDay        float64  // 🔴 HARD CAP
    MaxTokensPerExecution int     // 🔴 HARD CAP
}

func (ce *CrewExecutor) checkCrewBudget(estimatedCost float64) error {
    // Crew limits ALWAYS checked, ALWAYS enforced

    if estimatedCost > ce.crew.MaxCostPerExecution {
        return fmt.Errorf("crew budget exceeded: $%.2f > $%.2f",
            estimatedCost, ce.crew.MaxCostPerExecution)
    }

    if ce.crew.CostMetrics.DailyCostRemaining < estimatedCost {
        return fmt.Errorf("crew daily budget exceeded")
    }

    return nil  // Crew OK, proceed
}

// Agent limits are separate, don't block
func (agent *Agent) checkAgentBudget(estimatedCost float64) error {
    // Agent limits are SUGGESTIONS only

    if estimatedCost > agent.MaxCostPerDay {
        log.Printf("⚠️ Agent %s approaching daily limit", agent.ID)
        // But DON'T block - crew will handle it
    }

    return nil  // Never block here
}
```

### Ưu điểm
✅ **Simple hierarchy:** Crew rules everything
✅ **Easy to understand:** "Crew is the boss"
✅ **Easy to implement:** One check point
✅ **Easy to configure:** Set crew limit, done
✅ **Easy to debug:** Always know who blocked you
✅ **Prevents runaway:** Crew budget never exceeded
✅ **Clear accountability:** Crew = hard limit

### Nhược điểm
❌ Agent limits become "nice-to-have"
❌ Agent limits seem pointless
❌ Different control at two levels

---

## Option C: Both Independent

### Định nghĩa
```
Agent limits = Enforced per agent
Crew limits = Enforced for entire crew

Both checked independently:
- Agent limit can block
- Crew limit can block
- Request must pass BOTH checks
```

### Ví dụ Flow

**Scenario: Agent muốn gọi $3, crew còn $0.10**

```go
ExecuteAgent(agent, history):
  1. Check agent limit ($10/day)
     ├─ Used: $2
     ├─ Request: $3
     ├─ Remaining: $5
     └─ OK ✅

  2. Check crew limit ($2.50/execution)
     ├─ Current execution: $2.40
     ├─ Request: $3
     ├─ Would total: $5.40
     └─ EXCEEDS $2.50 → BLOCK! ❌

  3. Log: "Either agent or crew limit exceeded"

  4. Return error (unclear which limit)
```

### Triển khai

```go
func ExecuteAgent(agent *Agent, ...) error {
    estimatedCost := estimateCost(history)

    // Check BOTH limits
    if err := agent.checkLimit(estimatedCost); err != nil {
        return err  // Agent limit exceeded
    }

    if err := crew.checkLimit(estimatedCost); err != nil {
        return err  // Crew limit exceeded
    }

    // Both OK, proceed
    return executeCall()
}
```

### Ưu điểm
✅ Agent limits are actually useful
✅ Crew limits are actually useful
✅ Distributed control (agents responsible too)
✅ Per-agent tracking is meaningful

### Nhược điểm
❌ **More complex:** Two checks, two errors
❌ **Confusing:** Which limit blocked? Agent or crew?
❌ **Harder to debug:** Multiple failure points
❌ **Harder to configure:** Must tune both levels
❌ **More logic:** More code to maintain
❌ **Edge cases:** What if agent says OK but crew says NO?

---

## Detailed Comparison

### Scenario 1: Agent uses $8/day, tries to spend $5 more

```
Agent MaxCostPerDay: $10
Crew MaxCostPerDay: $100 (so far: $80)

Request cost: $5

┌─────────────────────────────────────┐
│ Option B: Crew Hard Cap             │
├─────────────────────────────────────┤
│ 1. Check crew daily: $80 + $5 = $85  │
│    vs $100 limit → OK ✅             │
│ 2. Check crew exec: $5 vs $2.50      │
│    → BLOCK if > $2.50 ❌             │
│ 3. Final: Execute if OK              │
│                                     │
│ Clarity: "Crew execution limit hit"  │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Option C: Both Independent          │
├─────────────────────────────────────┤
│ 1. Check agent: $8 + $5 = $13        │
│    vs $10 limit → BLOCK ❌           │
│ 2. Check crew: passes                │
│ 3. Final: BLOCKED by agent           │
│                                     │
│ Clarity: Confused - which limit?     │
└─────────────────────────────────────┘
```

### Scenario 2: Router agent already at $8, tries $5

```
Agent MaxCostPerDay: $10 (router used $8)
Crew MaxCostPerDay: $100 (crew used $95)
Crew MaxCostPerExecution: $2.50

Request cost: $3

OPTION B: Crew Hard Cap
─────────────────────────
1. Check crew daily: $95 + $3 = $98 vs $100 → OK ✅
2. Check crew exec: $3 vs $2.50 → BLOCK ❌
Result: BLOCKED (clear reason: execution limit)

OPTION C: Both Independent
──────────────────────────
1. Check agent: $8 + $3 = $11 vs $10 → BLOCK ❌
2. Check crew: would also be blocked
Result: BLOCKED (unclear which limit triggered first)
```

---

## Implementation Complexity

### Option B: Crew Hard Cap

**Implementation effort: MINIMAL**

```
Files to modify: 2
  └─ core/types.go (add fields)
  └─ core/crew.go (add 1 check function)

New functions: 1
  └─ CrewExecutor.checkCrewBudget()

New logic: ~20 lines
  └─ Compare estimated cost vs crew limits
  └─ Return error if exceeded

Test cases: 5
  └─ Crew daily limit
  └─ Crew execution limit
  └─ Both OK
  └─ Combination scenarios
```

### Option C: Both Independent

**Implementation effort: MODERATE**

```
Files to modify: 3
  └─ core/types.go (add fields)
  └─ core/agent.go (add check)
  └─ core/crew.go (add check)

New functions: 2
  └─ Agent.checkLimit()
  └─ CrewExecutor.checkLimit()

New logic: ~50 lines
  └─ Two separate check paths
  └─ Two different error messages
  └─ Complex failure scenarios

Test cases: 12+
  └─ Agent limits alone
  └─ Crew limits alone
  └─ Both triggered simultaneously
  └─ Confusing edge cases
```

---

## Configuration Complexity

### Option B: Crew Hard Cap

**Configuration: SIMPLE**

```yaml
crew:
  name: "Multi-Agent"
  max_cost_per_execution: 2.50    # This is the law
  max_cost_per_day: 100.00        # This is the law

agents:
  - router:
      max_cost_per_day: 10.00     # Informational only
      # Team sees this as guidance: "prefer to use $10"
      # But crew can override if needed
```

**Setup time: 5 minutes**
- Just set crew limits
- Agent limits are optional

### Option C: Both Independent

**Configuration: COMPLEX**

```yaml
crew:
  name: "Multi-Agent"
  max_cost_per_execution: 2.50    # This one enforced
  max_cost_per_day: 100.00        # This one enforced

agents:
  - router:
      max_cost_per_day: 10.00     # This one enforced too
      # Now what? Which one wins?
      # Do I set them the same?
      # What's the relationship?
  - faq:
      max_cost_per_day: 10.00
  - kb:
      max_cost_per_day: 10.00
  - aggregator:
      max_cost_per_day: 10.00

  # Total: $40/day for agents
  # But crew only allows $100/day
  # So agents could spend all $40 and still have $60 for... what?
```

**Setup time: 30+ minutes**
- Must decide per-agent limits
- Must decide crew limits
- Coordinate between them
- Document the strategy
- Team gets confused

---

## Maintenance & Debugging

### Option B: Crew Hard Cap

**Debugging: EASY**

```
Error in logs: "Crew execution limit exceeded: $3.00 > $2.50"
→ Clear: crew limit was the issue
→ Fix: increase crew limit OR reduce request size
→ Simple decision tree
```

**Support: EASY**

Q: "Why did my execution get blocked?"
A: "Crew hard cap reached. Check `MaxCostPerExecution`"

Q: "How do I increase budget?"
A: "Increase crew `MaxCostPerExecution` or `MaxCostPerDay`"

### Option C: Both Independent

**Debugging: HARD**

```
Error in logs: "Cost limit exceeded: $3.00"
→ Unclear: agent limit or crew limit?
→ Must check both configurations
→ Complex decision tree
```

**Support: HARD**

Q: "Why did my execution get blocked?"
A: "Could be agent limit or crew limit. Check both."

Q: "How do I increase budget?"
A: "Depends on which limit blocked you. Check the error logs carefully."

---

## Real-World Usage Patterns

### Pattern 1: Single-Tenant Setup

```
Company A uses go-agentic
  └─ 1 crew for all workflows
  └─ 4 agents in the crew

OPTION B (Crew Hard Cap):
  └─ Set crew budget once: $100/day
  └─ Each agent uses as needed
  └─ Simple ✅

OPTION C (Both Independent):
  └─ Set crew budget: $100/day
  └─ Set each agent limit: $25 each
  └─ Question: Why the per-agent limits if crew controls total?
  └─ Complicated ❌
```

### Pattern 2: Multi-Tenant Setup

```
Customer 1: Gets crew with $50/day
Customer 2: Gets crew with $100/day
Customer 3: Gets crew with $200/day

OPTION B (Crew Hard Cap):
  └─ Crew limit = customer's budget
  └─ Agents share the budget
  └─ Simple ✅

OPTION C (Both Independent):
  └─ Set crew limit = customer's budget
  └─ Set agent limits = ???
  └─ Do we split $100 equally among 4 agents?
  └─ What if one agent is used more often?
  └─ Complicated ❌
```

### Pattern 3: Billing & Analytics

```
OPTION B (Crew Hard Cap):
  └─ Track: crew total cost per execution
  └─ Track: crew total cost per day
  └─ Report: "Crew X used $95.50 of $100 budget today"
  └─ Simple ✅

OPTION C (Both Independent):
  └─ Track: crew cost
  └─ Track: agent cost
  └─ Report: "Crew used $95.50, but agents show $105.20"
  └─ Which one is right? ❌
```

---

## Code Example Comparison

### Option B: Crew Hard Cap

```go
// core/crew.go
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input string,
                                      streamChan chan *StreamEvent) error {
    for {
        // 🔑 ONE check: crew budget
        estimatedTokens := estimateTokens(ce.history)
        estimatedCost := float64(estimatedTokens) * costPerToken

        // Crew hard cap check
        if estimatedCost > ce.crew.MaxCostPerExecution {
            streamChan <- NewStreamEvent("error", "system",
                fmt.Sprintf("Crew execution limit exceeded: $%.2f > $%.2f",
                    estimatedCost, ce.crew.MaxCostPerExecution))
            return fmt.Errorf("crew execution limit exceeded")
        }

        if ce.crew.CostMetrics.DailyCostRemaining < estimatedCost {
            streamChan <- NewStreamEvent("error", "system",
                "Crew daily limit exceeded")
            return fmt.Errorf("crew daily limit exceeded")
        }

        // Agent warning (informational only)
        if agent.MaxCostPerDay > 0 {
            if agent.CostMetrics.DailyCostUsed + estimatedCost > agent.MaxCostPerDay {
                log.Printf("⚠️ Agent %s: approaching daily limit", agent.ID)
            }
        }

        // Execute agent
        response, err := ExecuteAgent(ctx, agent, input, ce.history, ce.apiKey)

        // Update crew metrics (these track actual cost)
        ce.updateCrewCostMetrics(response, actualCost)

        // Continue to next agent or finish
    }
}
```

**Lines of code: ~30**
**Complexity: Low**

### Option C: Both Independent

```go
// core/crew.go
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input string,
                                      streamChan chan *StreamEvent) error {
    for {
        estimatedTokens := estimateTokens(ce.history)
        estimatedCost := float64(estimatedTokens) * costPerToken

        // 🔑 TWO checks: agent AND crew

        // Check 1: Agent limit
        if agent.MaxCostPerDay > 0 {
            if agent.CostMetrics.DailyCostUsed + estimatedCost > agent.MaxCostPerDay {
                streamChan <- NewStreamEvent("error", "system",
                    fmt.Sprintf("Agent %s daily limit exceeded", agent.ID))
                return fmt.Errorf("agent daily limit exceeded")
            }
        }

        // Check 2: Crew execution limit
        if estimatedCost > ce.crew.MaxCostPerExecution {
            streamChan <- NewStreamEvent("error", "system",
                fmt.Sprintf("Crew execution limit exceeded: $%.2f > $%.2f",
                    estimatedCost, ce.crew.MaxCostPerExecution))
            return fmt.Errorf("crew execution limit exceeded")
        }

        // Check 3: Crew daily limit
        if ce.crew.CostMetrics.DailyCostRemaining < estimatedCost {
            streamChan <- NewStreamEvent("error", "system",
                "Crew daily limit exceeded")
            return fmt.Errorf("crew daily limit exceeded")
        }

        // Problem: What if checks contradict each other?
        // Agent says OK but crew says NO
        // User gets confused error message

        // Execute agent
        response, err := ExecuteAgent(ctx, agent, input, ce.history, ce.apiKey)

        // Update both agent and crew metrics
        ce.updateAgentCostMetrics(agent, response, actualCost)
        ce.updateCrewCostMetrics(response, actualCost)

        // Continue to next agent or finish
    }
}
```

**Lines of code: ~50**
**Complexity: High**
**Edge cases: Many**

---

## Decision Recommendation

### Best Choice: **Option B - Crew Limit is Hard Cap**

#### Why?

| Criteria | Option B | Option C |
|----------|----------|----------|
| **Simplicity** | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| **Easy Deploy** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Easy Config** | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| **Easy Debug** | ⭐⭐⭐⭐⭐ | ⭐⭐ |
| **Clear Errors** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Maintenance** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Code Complexity** | Low | High |
| **Test Cases** | 5 | 15+ |
| **User Confusion** | Minimal | High |

### Implementation Plan (Option B)

```
Week 2: Crew Cost Control
├─ Add to Crew type:
│  └─ MaxCostPerExecution: float64
│  └─ MaxCostPerDay: float64
│  └─ MaxTokensPerExecution: int
│
├─ Add to CrewCostMetrics:
│  └─ DailyCostRemaining: float64
│  └─ ExecutionCostUsed: float64
│
├─ Implement checks:
│  └─ CrewExecutor.checkCrewBudget()
│
├─ Integrate into Execute():
│  └─ Call checkCrewBudget() before ExecuteAgent()
│  └─ Update metrics after ExecuteAgent()
│
├─ Keep Agent limits informational:
│  └─ Log warnings only
│  └─ Don't block
│
└─ Test & deploy
   └─ 5 test cases
   └─ 2 days implementation
   └─ Ready for Week 3 monitoring
```

---

## Summary

### The Winning Decision: **B) Crew Limit is Hard Cap**

```
✅ SIMPLE        - One clear hierarchy
✅ EASY DEPLOY   - ~20 lines of code
✅ EASY CONFIG   - Set crew limit, done
✅ EASY DEBUG    - Clear error messages
✅ EASY MAINT    - Few edge cases
✅ PRACTICAL     - Works for real use cases

❌ NOT complex   - No dual control confusion
❌ NOT confusing - Users know exactly what limit they hit
❌ NOT hard      - No multi-point failure paths
```

---

## Next Steps

Once team agrees on **Decision #2: Crew Hard Cap**:

1. **Document it:**
   - Add to TEAM_DISCUSSION_BRIEF.md
   - Add to IMPLEMENTATION_GUIDE.md

2. **Communicate it:**
   - Team knows crew limit is absolute
   - Agent limits are advisory only

3. **Implement it:**
   - Week 2 implementation
   - ~2 days of work
   - Ready for production

4. **Monitor it:**
   - Track crew budget usage
   - Alert when approaching limits
   - Week 3 complete

---

**Recommendation: Go with Option B - Crew Limit is Hard Cap** ✅

Simple, practical, easy to deploy, easy to maintain.

Perfect for production use. 🚀
