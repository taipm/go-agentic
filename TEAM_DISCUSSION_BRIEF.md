# 🎯 TEAM DISCUSSION BRIEF: Cost Control for Agents & Crew

**Meeting Focus:** Cost Limits & Control Mechanisms
**Duration:** 60 minutes
**Participants:** Dev Team, Architects, Product Leads

---

## 📋 AGENDA

### Part 1: Agent-Level Cost Control (20 min)
### Part 2: Crew-Level Cost Control (20 min)
### Part 3: Implementation Strategy (15 min)
### Part 4: Q&A & Decision (5 min)

---

## 🔴 PART 1: AGENT-LEVEL COST CONTROL

### Current Problem

**Location:** `core/agent.go:21`

```go
func ExecuteAgent(ctx context.Context, agent *Agent, input string,
                  history []Message, apiKey string) (*AgentResponse, error) {
    // ❌ NO COST LIMITS
    // ❌ NO TOKEN ESTIMATION
    // ❌ NO COST ALERTS

    messages := convertToProviderMessages(history)  // Could be huge!
    response, err := provider.Complete(ctx, request)  // Unlimited cost
}
```

### Financial Impact

**Scenario 1: Single Agent with Large History**
```
History: 1000 messages (accumulated over time)
Tokens: 25,000 per request
Cost: $0.63 per request
Daily (100 requests): $63
Monthly: $1,890 per user
Annual: $22,680 per user 💥
```

**Scenario 2: Agent Called in Loop**
```
for i := 0; i < 1000; i++ {
    response, _ := ExecuteAgent(ctx, agent, input, history, apiKey)
    // Each call: 25,000 tokens × $0.0000025 = $0.063
    // 1000 calls: $63 💥
    // Without limits, developer doesn't know!
}
```

### Proposed Solution: Agent Cost Config

```go
type Agent struct {
    // ... existing fields ...

    // 🆕 COST CONTROLS
    MaxTokensPerCall      int          // Max tokens per single call (e.g., 4000)
    MaxTokensPerDay       int          // Max tokens per day (e.g., 100,000)
    MaxCostPerDay         float64      // Max cost per day (e.g., $10)
    CostAlertThreshold    float64      // Alert when approaching limit (e.g., 80%)
    EnforceCostLimits     bool         // Reject calls if would exceed limit
}
```

### Implementation Questions for Team

1. **Should we BLOCK or WARN when agent exceeds limits?**
   - Option A: Block (return error) ← Safer
   - Option B: Warn (log + continue) ← More flexible
   - Option C: Both (configurable)

2. **Who should pay for exceeded costs?**
   - Option A: Customer account
   - Option B: Internal testing account (reject agent call)
   - Option C: Configurable per agent

3. **How granular should token estimation be?**
   - Option A: Rough estimate (current_tokens / 4) ← Fast
   - Option B: Detailed calculation ← Slower but accurate
   - Option C: Use provider's tokenizer ← Most accurate but external API call

4. **Should daily limits reset automatically?**
   - Option A: Yes, at UTC midnight ← Standard
   - Option B: Yes, at custom time ← Flexible
   - Option C: Manual reset only ← Controlled

### Key Decision: Cost Enforcement

```
We RECOMMEND:
┌─────────────────────────────────────────────────┐
│ Agent Cost Enforcement Strategy                 │
├─────────────────────────────────────────────────┤
│                                                 │
│ 1️⃣ PRE-CALL CHECK                              │
│    └─ Estimate tokens for this request          │
│    └─ Check against MaxTokensPerCall            │
│    └─ Check against daily remaining budget      │
│                                                 │
│ 2️⃣ DECISION POINT                              │
│    ├─ If OK: Execute call                       │
│    ├─ If 80%+ used: Warn + execute              │
│    └─ If exceeds: Block or warn (configurable) │
│                                                 │
│ 3️⃣ POST-CALL UPDATE                            │
│    └─ Record actual tokens used                 │
│    └─ Update daily counter                      │
│    └─ Check if approaching daily limit          │
│                                                 │
└─────────────────────────────────────────────────┘
```

### Monitoring & Metrics

```go
type AgentCostMetrics struct {
    CallCount              int
    TotalTokensUsed        int
    TotalCostIncurred      float64
    DailyTokensRemaining   int
    DailyCostRemaining     float64
    CostPercentageUsed     float64  // 0-100%
    LastCallTimestamp      time.Time
    LastCallTokens         int
    LastCallCost           float64
}
```

---

## 🏢 PART 2: CREW-LEVEL COST CONTROL

### Current Problem

**Location:** `core/crew.go:406-475`

```go
type Crew struct {
    Agents                  []*Agent
    Tasks                   []*Task
    MaxRounds               int
    MaxHandoffs             int
    // ❌ NO CREW-LEVEL COST LIMITS
    // ❌ NO MULTI-AGENT COST TRACKING
    // ❌ NO ABORT-IF-OVER-BUDGET
}

func (ce *CrewExecutor) Execute(...) (*CrewResponse, error) {
    // ❌ No way to know total crew cost
    // ❌ Can't prevent runaway costs
    // ❌ Can't enforce budget per crew
}
```

### Financial Impact

**Scenario: Multi-Agent Workflow**
```
Crew: Router → FAQ-Searcher → Knowledge-Base → Aggregator

Agent 1 (Router):
  Cost: $0.10 per call
  Calls: 10 per request
  Subtotal: $1.00

Agent 2 (FAQ-Searcher):
  Cost: $0.25 per call
  Calls: 5 per request
  Subtotal: $1.25

Agent 3 (Knowledge-Base):
  Cost: $0.50 per call
  Calls: 3 per request
  Subtotal: $1.50

Agent 4 (Aggregator):
  Cost: $0.30 per call
  Calls: 2 per request
  Subtotal: $0.60

TOTAL PER REQUEST: $4.35
Per day (100 requests): $435
Per month: $13,050
Per year: $156,600 per user! 💥

WITHOUT CREW COST LIMITS: Developer has NO visibility!
```

### Proposed Solution: Crew Budget Config

```go
type Crew struct {
    // ... existing fields ...

    // 🆕 CREW-LEVEL COST CONTROLS
    MaxCostPerExecution   float64              // Max cost per single Execute() call
    MaxCostPerDay         float64              // Max cumulative cost per day
    MaxTokensPerExecution int                  // Max total tokens per execution
    BudgetExceededAction  string               // "block", "warn", "abort"
    CostTrackingEnabled   bool                 // Enable metrics collection

    // 🆕 MONITORING
    CostMetrics           *CrewCostMetrics     // Track cumulative costs
}

type CrewCostMetrics struct {
    ExecutionCount         int
    TotalTokensUsed        int
    TotalCostIncurred      float64
    DailyTokensRemaining   int
    DailyCostRemaining     float64
    CostPercentageUsed     float64  // 0-100%

    // Per-agent breakdown
    AgentCosts             map[string]AgentExecutionMetrics

    LastExecutionTime      time.Time
    LastExecutionTokens    int
    LastExecutionCost      float64
}

type AgentExecutionMetrics struct {
    CallCount        int
    TokensUsed       int
    CostIncurred     float64
    AverageCost      float64
}
```

### Implementation Strategy

```
CREW COST TRACKING FLOW:
─────────────────────────

1️⃣ CREW SETUP
   Initialize CrewCostMetrics
   Load budget from config
   Reset daily counter at startup

2️⃣ BEFORE AGENT EXECUTION
   └─ Sum remaining budgets from all agents
   └─ Check crew daily remaining
   └─ Estimate tokens for this agent call
   └─ Verify it fits in crew budget

3️⃣ AGENT EXECUTION
   └─ ExecuteAgent(agent, history, apiKey)
   └─ Collect actual tokens used
   └─ Calculate actual cost

4️⃣ AFTER AGENT EXECUTION
   └─ Update AgentCostMetrics
   └─ Update CrewCostMetrics
   └─ Check daily limit
   └─ Check execution limit

5️⃣ DECISION ON NEXT AGENT
   ├─ If crew budget OK: Continue
   ├─ If 80% used: Warn
   ├─ If exceeded: Block/Abort
   └─ Log metrics

6️⃣ CREW EXECUTION COMPLETE
   └─ Return summary with total cost
   └─ Log cost metrics to monitoring
```

### Key Questions for Team

1. **Should crew-level limits override agent-level limits?**
   - Option A: Crew limit = hard cap (overrides agent limit)
   - Option B: Sum of agent limits = crew limit
   - Option C: Both enforced independently

2. **What happens when crew budget exceeded?**
   - Option A: Block immediately (safest)
   - Option B: Pause execution, wait for approval
   - Option C: Warn, continue (risky)
   - Option D: Configurable per crew

3. **How should we handle agent handoffs under budget pressure?**
   - Option A: Pause if next agent would exceed budget
   - Option B: Skip agent, go to next
   - Option C: Summarize and continue with reduced context

4. **Should we support cost budgets by customer/account?**
   - Option A: Yes, crew-level budgets per customer
   - Option B: Yes, global budget limit per customer
   - Option C: Manual configuration only

### Integration Points

```
Current Execute Flow:           With Cost Control:
──────────────────────────────  ─────────────────────────────

1. ExecuteStream()              1. ExecuteStream()
   └─ Agent iteration              └─ Check crew budget ✅
      └─ ExecuteAgent()            └─ Agent iteration
         └─ LLM call                  └─ Check agent budget ✅
                                       └─ ExecuteAgent()
                                          └─ Estimate tokens ✅
                                          └─ LLM call
                                          └─ Record actual cost ✅
                                       └─ Update crew metrics ✅
                                    └─ Check crew limits ✅
```

---

## ⚙️ PART 3: IMPLEMENTATION STRATEGY

### Phase 1: Agent-Level Cost Control (Week 1-2)

**Step 1a: Update Agent Type**
```go
// core/types.go
type Agent struct {
    // ... existing 39 lines ...

    // Cost controls (10 new lines)
    MaxTokensPerCall      int      `yaml:"max_tokens_per_call"`
    MaxTokensPerDay       int      `yaml:"max_tokens_per_day"`
    MaxCostPerDay         float64  `yaml:"max_cost_per_day"`
    CostAlertThreshold    float64  `yaml:"cost_alert_threshold"`
    EnforceCostLimits     bool     `yaml:"enforce_cost_limits"`
}
```

**Step 1b: Add Token Estimator**
```go
// core/crew.go - Already exists! estimateTokens()
func estimateTokens(messages []Message) int
```

**Step 1c: Add Pre-Call Cost Check**
```go
// core/agent.go - New function
func (agent *Agent) checkCostLimits(history []Message, maxContextTokens int) error {
    // Estimate tokens for this request
    tokens := estimateTokens(history)

    // Check against MaxTokensPerCall
    if agent.MaxTokensPerCall > 0 && tokens > agent.MaxTokensPerCall {
        if agent.EnforceCostLimits {
            return fmt.Errorf("request exceeds agent max tokens: %d > %d",
                tokens, agent.MaxTokensPerCall)
        }
        log.Printf("⚠️ [COST WARNING] Agent %s: %d tokens > limit %d",
            agent.ID, tokens, agent.MaxTokensPerCall)
    }

    return nil
}
```

**Step 1d: Update ExecuteAgent**
```go
func ExecuteAgent(ctx context.Context, agent *Agent, input string,
                  history []Message, apiKey string) (*AgentResponse, error) {

    // ✅ NEW: Check agent cost limits
    if err := agent.checkCostLimits(history, agent.MaxContextTokens); err != nil {
        return nil, err
    }

    // ... existing execution ...
}
```

### Phase 2: Crew-Level Cost Control (Week 2-3)

**Step 2a: Update Crew Type**
```go
// core/types.go
type Crew struct {
    // ... existing 10 fields ...

    // Cost controls (5 new fields)
    MaxCostPerExecution  float64
    MaxCostPerDay        float64
    MaxTokensPerExecution int
    BudgetExceededAction string
    CostMetrics          *CrewCostMetrics
}

// New type
type CrewCostMetrics struct {
    ExecutionCount      int
    TotalCostIncurred   float64
    DailyCostRemaining  float64
    // ... fields ...
}
```

**Step 2b: Add Crew Cost Checker**
```go
// core/crew.go - New function
func (ce *CrewExecutor) checkCrewBudget(estimatedTokens int) error {
    if ce.crew.CostMetrics == nil {
        return nil  // No metrics tracking
    }

    estimatedCost := float64(estimatedTokens) * 0.0000025

    if ce.crew.MaxCostPerExecution > 0 {
        if estimatedCost > ce.crew.MaxCostPerExecution {
            return fmt.Errorf("execution would exceed crew budget: $%.2f > $%.2f",
                estimatedCost, ce.crew.MaxCostPerExecution)
        }
    }

    return nil
}
```

**Step 2c: Update Execute Methods**
```go
// Update ExecuteStream() at line 499
func (ce *CrewExecutor) ExecuteStream(...) error {
    // ... existing code ...

    for {
        // ✅ NEW: Check crew budget before agent execution
        estimatedTokens := estimateTokens(ce.history)
        if err := ce.checkCrewBudget(estimatedTokens); err != nil {
            streamChan <- NewStreamEvent("error", "system",
                fmt.Sprintf("Crew budget exceeded: %v", err))
            return err
        }

        // ... existing agent execution ...

        // ✅ NEW: Update metrics after execution
        ce.updateCrewCostMetrics(response)
    }
}
```

### Phase 3: Monitoring & Reporting (Week 3)

**Step 3a: Add Cost Reporting**
```go
// core/crew.go - New method
func (ce *CrewExecutor) GetCostReport() map[string]interface{} {
    return map[string]interface{}{
        "execution_count": ce.crew.CostMetrics.ExecutionCount,
        "total_cost": ce.crew.CostMetrics.TotalCostIncurred,
        "daily_remaining": ce.crew.CostMetrics.DailyCostRemaining,
        "percent_used": ce.crew.CostMetrics.CostPercentageUsed,
        "by_agent": ce.crew.CostMetrics.AgentCosts,
    }
}
```

**Step 3b: Add Monitoring Endpoint**
```go
// core/http.go - New endpoint
// GET /metrics/crew-costs
func (h *Handler) CrewCostsMetrics(w http.ResponseWriter, r *http.Request) {
    report := h.executor.GetCostReport()
    json.NewEncoder(w).Encode(report)
}
```

---

## 📊 PART 4: COMPARISON TABLE

### Agent-Level vs Crew-Level

| Aspect | Agent-Level | Crew-Level |
|--------|------------|-----------|
| **Scope** | Single agent execution | Entire workflow |
| **Budget Type** | Per-call or per-day | Per-execution or per-day |
| **Enforcement** | Block/warn on call | Block/warn on agent loop |
| **Granularity** | Tight, specific | Broad, aggregate |
| **Use Case** | Prevent agent runaway | Control workflow costs |
| **Config** | In agent YAML | In crew YAML |

### Configuration Example

```yaml
# crew.yaml
crew_name: "Multi-Agent System"

# 🆕 Crew-level budget
max_cost_per_execution: 2.50    # Max $2.50 per Execute() call
max_cost_per_day: 50.00          # Max $50 per day
max_tokens_per_execution: 10000  # Max 10k tokens per execution
budget_exceeded_action: "block"   # Block if would exceed

# Individual agents
agents:
  - id: "router"
    max_tokens_per_call: 1000    # 🆕 Agent-level limit
    max_tokens_per_day: 20000    # 🆕 Agent daily limit
    max_cost_per_day: 5.00        # 🆕 Agent daily budget
    enforce_cost_limits: true     # 🆕 Strict enforcement
```

---

## 🎯 RECOMMENDATIONS

### For Agent-Level:
✅ **DO IMPLEMENT:**
- MaxTokensPerCall (prevents single call runaway)
- MaxTokensPerDay (prevents daily runaway)
- Pre-call token estimation
- Warnings at 80% threshold

❌ **DON'T IMPLEMENT YET:**
- Per-provider rate limiting (future)
- Automatic cost optimization (future)
- A/B testing with cost constraints (future)

### For Crew-Level:
✅ **DO IMPLEMENT:**
- MaxCostPerExecution (per workflow run)
- MaxCostPerDay (daily aggregate)
- Crew cost metrics tracking
- Integration with monitoring

❌ **DON'T IMPLEMENT YET:**
- Customer-level budgets (future)
- Cost prediction/forecasting (future)
- Automatic agent skip on budget (future)

---

## ⏱️ TIMELINE

```
Week 1: Agent-Level Implementation
  Mon: Add Agent type fields
  Tue: Implement cost checker
  Wed: Add token estimation
  Thu: Update ExecuteAgent
  Fri: Test & staging

Week 2: Crew-Level Implementation
  Mon: Add Crew cost metrics
  Tue: Implement crew budget checker
  Wed: Update Execute/ExecuteStream
  Thu: Integration testing
  Fri: Staging deployment

Week 3: Monitoring & Production
  Mon: Add reporting endpoint
  Tue: Dashboarding
  Wed: Documentation
  Thu: Team training
  Fri: Production deployment

Total: 3 weeks
```

---

## 💰 EXPECTED IMPACT

**Before:**
- Unbounded agent costs
- No visibility into spending
- Runaway costs possible
- No daily budgeting

**After:**
- Controlled agent costs ✅
- Real-time cost visibility ✅
- Budget enforcement ✅
- Daily tracking & reporting ✅

**Estimated Savings:** 20-30% of LLM costs through better visibility and control

---

## 🚀 DECISION POINTS FOR TEAM

### Decision 1: Block vs Warn
**Question:** When agent/crew exceeds budget, should we:
- A) BLOCK: Return error, reject call ← Safest
- B) WARN: Log warning, continue call ← Risky
- C) BOTH: Configurable per agent/crew ← Most flexible

**Recommendation:** Option C (configurable)

---

### Decision 2: Budget Hierarchy
**Question:** If both agent and crew limits exist, which wins?
- A) Agent limit is primary
- B) Crew limit is primary (hard cap)
- C) Both enforced independently
- D) Most restrictive wins

**Recommendation:** Option B (Crew is hard cap)

---

### Decision 3: Release Timeline
**Question:** Ship when ready or by specific date?
- A) Ship when complete (3 weeks, quality-first)
- B) Ship by end of quarter (deadline-driven)
- C) Ship MVP in 1 week, full in 2 weeks (phased)

**Recommendation:** Option C (phased approach)

---

## 📚 SUPPORTING DOCUMENTS

Detailed implementations available in:
- [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md) - Code examples
- [MEMORY_ANALYSIS.md](./MEMORY_ANALYSIS.md) - Complete analysis
- [MEMORY_ISSUES_SUMMARY.txt](./MEMORY_ISSUES_SUMMARY.txt) - Quick reference

---

## ✅ NEXT STEPS

1. **Review** this brief (30 min)
2. **Discuss** decision points (30 min)
3. **Decide** on approach (10 min)
4. **Assign** tasks to team members (10 min)

**Expected outcome:** Clear decision on implementation strategy

---

**Prepared by:** Memory & Cost Analysis Team
**Date:** 2025-12-22
**Status:** Ready for Team Discussion 🎯
