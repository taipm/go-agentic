# Tech-Spec: Crew-Level Cost Control with Hard Cap Enforcement

**Created:** 2025-12-22  
**Status:** Ready for Development  
**Priority:** HIGH (Budget Safety)  
**Estimated Impact:** System-wide cost protection + crew budget guarantee  

---

## Overview

### Problem Statement

While agent-level controls provide per-agent flexibility, there's no system-wide budget enforcement:

1. **No System-Wide Budget Limit:**
   - Crew has no maximum execution cost
   - Multiple agents can collectively exceed desired budget
   - No way to set hard limit on entire workflow

2. **Budget Hierarchy Confusion:**
   - What if agent limits total $40/day but we want crew max $2.50/execution?
   - No clear rule for which limit takes priority
   - Complex configuration with ambiguous semantics

3. **No Execution-Level Control:**
   - Can't limit cost PER EXECUTION (one workflow run)
   - Can't prevent expensive requests from happening
   - Only daily limits exist (too long for failure recovery)

**Current Behavior:**
```go
// No execution-level cost control
type Crew struct {
    Agents []*Agent
    // ... no budget fields
}

// Cost only checked per-agent
// No crew-wide enforcement
```

**Impact:**
- A single expensive request can burn significant budget
- Cannot enforce "max $2.50 per API call" limit
- Cannot guarantee daily budget won't be exceeded
- Unclear behavior when limits conflict

### Solution

Implement **crew-level hard cap** with clear budget hierarchy:

**Key Decision: CREW HARD CAP (not agents)**
- Crew MaxCostPerExecution = **HARD CAP** (absolute maximum per run)
- Crew MaxCostPerDay = **HARD CAP** (absolute maximum per day)
- Agent limits = **INFORMATIONAL** (for tracking/warnings only)

**Architecture:**
```
CrewExecutor.Execute(request)
    ↓
Estimate total cost for ALL agents
    ↓
🔴 CHECK CREW BUDGET FIRST:
    ├─ if cost > daily limit → BLOCK ❌
    ├─ if cost > per-execution limit → BLOCK ❌
    └─ else → Allow ✅
    ↓
For each agent:
    ├─ Check agent informational warning (log only)
    └─ Execute agent
    ↓
Update crew metrics
    ↓
Return results + total cost
```

**Why CREW HARD CAP?**
- Simplicity: One rule (crew always wins)
- Safety: System budget never exceeded
- Clarity: No ambiguity about hierarchy
- Debuggability: One decision point for cost blocks
- Production-Ready: Clear semantics, easy to operate

### Scope (In/Out)

**IN SCOPE:**
- ✅ Update `Crew` type with 3 cost control fields (MaxCostPerExecution, MaxCostPerDay, MaxTokensPerExecution)
- ✅ Implement `checkCrewBudget()` function
- ✅ Implement cost estimation for crews
- ✅ Integrate budget checks into `Execute()` method
- ✅ Track crew metrics (ExecutionCount, TotalTokens, DailyCost, AgentCosts breakdown)
- ✅ Load crew cost config from YAML
- ✅ Comprehensive test coverage (5 test cases)
- ✅ Debug logging for budget decisions

**OUT OF SCOPE:**
- ❌ Agent-level cost controls (separate tech-spec)
- ❌ Crew-level reporting/dashboard (separate feature)
- ❌ Historical analytics (separate feature)
- ❌ Tier-based pricing (future enhancement)

---

## Context for Development

### Codebase Patterns

**Crew Type Pattern (types.go):**
```go
// Current: No cost control
type Crew struct {
    Name              string
    Agents            []*Agent
    Config            *Config
    // ... other fields
}

// New: Add cost control fields
type Crew struct {
    // ... existing fields ...
    
    // Cost Control Configuration (HARD CAPS)
    MaxCostPerExecution float64 // Max cost per single execution (e.g., $2.50)
    MaxCostPerDay       float64 // Max cumulative cost per day (e.g., $100.00)
    MaxTokensPerExecution int   // Max tokens per single execution (e.g., 20000)
    
    // Runtime metrics (not serialized)
    CrewMetrics CrewCostMetrics `json:"-" yaml:"-"`
}

// New struct for crew metrics
type CrewCostMetrics struct {
    ExecutionCount  int                // Number of executions
    TotalTokens     int                // Total tokens used
    DailyCost       float64            // Total cost today
    LastResetTime   time.Time          // When daily counter resets
    AgentCosts      map[string]float64 // Per-agent cost breakdown
    mu              sync.RWMutex
}
```

**Configuration Pattern (YAML):**
```yaml
# crew.yaml - HARD CAPS
crew:
  name: "Multi-Agent Search"
  
  # HARD CAP: Cost per execution
  max_cost_per_execution: 2.50
  
  # HARD CAP: Cost per day
  max_cost_per_day: 100.00
  
  # HARD CAP: Tokens per execution
  max_tokens_per_execution: 20000

# agents/*.yaml - INFORMATIONAL ONLY
agents:
  - id: router
    max_cost_per_day: 10.00      # ℹ️ For tracking only
    
  - id: faq_searcher
    max_cost_per_day: 20.00      # ℹ️ For tracking only
```

### Cost Hierarchy Visualization

```
┌──────────────────────────────────────┐
│ CREW HARD CAP                        │ ← ENFORCED ✅
│ MaxCostPerExecution: $2.50           │
│ MaxCostPerDay: $100.00               │
└──────────────────────────────────────┘
           ▼ GATES ▼
┌──────────────────────────────────────────────┐
│ AGENT INFORMATIONAL LIMITS                   │ ← TRACKED ONLY ℹ️
│ Router: MaxCostPerDay $10                    │
│ FAQ Searcher: MaxCostPerDay $20              │
│ Knowledge Searcher: MaxCostPerDay $15        │
│ Aggregator: MaxCostPerDay $10                │
└──────────────────────────────────────────────┘
```

### Budget Check Integration Point

Current Execute pattern (crew.go):
```go
func (c *Crew) Execute(ctx context.Context, input string) (Result, error) {
    // Step 1: Prepare request
    req := &Request{...}
    
    // NEW: Step 2: Check crew budget FIRST
    // Before executing ANY agents
    
    // Step 3: Execute agents in parallel
    for _, agent := range c.Agents {
        result := agent.Execute(ctx, req)
    }
    
    // Step 4: Return results
}
```

---

## Detailed Design

### Step 1: Type Definitions (types.go)

**Location:** `internal/types.go`, around line 250-280 where Crew is defined

```go
type Crew struct {
    // ... existing fields (Name, Agents, Config, etc.) ...
    
    // Cost Control Configuration (HARD CAPS)
    MaxCostPerExecution float64 `yaml:"max_cost_per_execution"` // e.g., 2.50
    MaxCostPerDay       float64 `yaml:"max_cost_per_day"`       // e.g., 100.00
    MaxTokensPerExecution int   `yaml:"max_tokens_per_execution"` // e.g., 20000
    
    // Runtime metrics (not serialized)
    CrewMetrics CrewCostMetrics `json:"-" yaml:"-"`
}

// New struct for crew metrics
type CrewCostMetrics struct {
    ExecutionCount  int                // Total executions this period
    TotalTokens     int                // Total tokens consumed
    DailyCost       float64            // Total cost today
    LastResetTime   time.Time          // When daily counter resets
    AgentCosts      map[string]float64 // Cost breakdown per agent
    mu              sync.RWMutex
}
```

### Step 2: Cost Estimation Functions (crew.go)

**Location:** `internal/crew.go`, add new functions (~30 lines)

```go
// estimateTotalCostForRequest estimates cost for executing all agents on a request
func (c *Crew) estimateTotalCostForRequest(req *Request) float64 {
    var totalCost float64
    
    // Sum estimated costs from each agent
    for _, agent := range c.Agents {
        agentCost := c.estimateAgentCost(agent, req)
        totalCost += agentCost
    }
    
    return totalCost
}

// estimateAgentCost estimates cost for single agent on a request
func (c *Crew) estimateAgentCost(agent *Agent, req *Request) float64 {
    tokens := agent.estimateTokens(req.Content)
    return float64(tokens) * agent.CostPerToken
}

// estimateTotalTokensForRequest estimates total tokens for all agents
func (c *Crew) estimateTotalTokensForRequest(req *Request) int {
    var totalTokens int
    
    for _, agent := range c.Agents {
        agentTokens := agent.estimateTokens(req.Content)
        totalTokens += agentTokens
    }
    
    return totalTokens
}
```

### Step 3: Crew Budget Checker (crew.go)

**Location:** `internal/crew.go`, add new function (~50 lines)

```go
// checkCrewBudget verifies the crew budget can accommodate this execution
func (c *Crew) checkCrewBudget(estimatedCost float64, estimatedTokens int) error {
    // Initialize metrics if needed
    if c.CrewMetrics.LastResetTime.IsZero() {
        c.CrewMetrics.LastResetTime = time.Now()
    }
    
    // Check if daily reset needed
    c.checkAndResetDailyMetrics()
    
    // 🔴 CHECK 1: Per-execution cost limit (HARD CAP)
    if estimatedCost > c.MaxCostPerExecution {
        return fmt.Errorf(
            "❌ [BUDGET BLOCK] Crew execution cost $%.4f > limit $%.2f (per-execution cap)",
            estimatedCost, c.MaxCostPerExecution)
    }
    
    // 🔴 CHECK 2: Per-execution token limit (HARD CAP)
    if estimatedTokens > c.MaxTokensPerExecution {
        return fmt.Errorf(
            "❌ [BUDGET BLOCK] Crew execution tokens %d > limit %d (per-execution cap)",
            estimatedTokens, c.MaxTokensPerExecution)
    }
    
    // 🔴 CHECK 3: Daily cost limit (HARD CAP)
    c.CrewMetrics.mu.RLock()
    currentDailyCost := c.CrewMetrics.DailyCost
    c.CrewMetrics.mu.RUnlock()
    
    newDailyCost := currentDailyCost + estimatedCost
    if newDailyCost > c.MaxCostPerDay {
        return fmt.Errorf(
            "❌ [BUDGET BLOCK] Crew daily cost $%.4f + $%.4f > limit $%.2f",
            currentDailyCost, estimatedCost, c.MaxCostPerDay)
    }
    
    return nil
}

// checkAndResetDailyMetrics resets daily counters if a day has passed
func (c *Crew) checkAndResetDailyMetrics() {
    c.CrewMetrics.mu.Lock()
    defer c.CrewMetrics.mu.Unlock()
    
    now := time.Now()
    // If last reset was > 24 hours ago, reset
    if now.Sub(c.CrewMetrics.LastResetTime) > 24*time.Hour {
        c.CrewMetrics.DailyCost = 0
        c.CrewMetrics.ExecutionCount = 0
        c.CrewMetrics.TotalTokens = 0
        c.CrewMetrics.LastResetTime = now
    }
}
```

### Step 4: Integration with Execute() (crew.go)

**Location:** `internal/crew.go`, modify `Execute()` method

Current code:
```go
func (c *Crew) Execute(ctx context.Context, input string) (Result, error) {
    // Prepare request
    req := &Request{...}
    
    // Execute agents in parallel
    results := make([]Result, len(c.Agents))
    for i, agent := range c.Agents {
        results[i] = agent.Execute(ctx, req)
    }
    
    return results, nil
}
```

Modified code (add crew budget checks):
```go
func (c *Crew) Execute(ctx context.Context, input string) (Result, error) {
    // Step 1: Prepare request
    req := &Request{Content: input}
    
    // Step 2: Estimate total cost and tokens
    estimatedCost := c.estimateTotalCostForRequest(req)
    estimatedTokens := c.estimateTotalTokensForRequest(req)
    
    // Step 3: 🔴 CHECK CREW BUDGET FIRST (HARD CAP)
    if err := c.checkCrewBudget(estimatedCost, estimatedTokens); err != nil {
        return nil, err  // BLOCK immediately if budget exceeded
    }
    
    // Step 4: Execute agents (now safe, budget is OK)
    results := make([]Result, len(c.Agents))
    var totalActualCost float64
    
    for i, agent := range c.Agents {
        // ℹ️ Check agent informational warnings (don't block)
        agentCost := c.estimateAgentCost(agent, req)
        if agentCost > agent.MaxCostPerDay * agent.CostAlertThreshold {
            log.Printf("⚠️  [WARN] Agent '%s' approaching cost limit in crew execution",
                agent.ID)
        }
        
        // Execute agent
        result, err := agent.Execute(ctx, req)
        if err != nil {
            return nil, err
        }
        
        results[i] = result
        totalActualCost += result.Cost
        
        // Track per-agent cost
        c.CrewMetrics.mu.Lock()
        c.CrewMetrics.AgentCosts[agent.ID] = result.Cost
        c.CrewMetrics.mu.Unlock()
    }
    
    // Step 5: Update crew metrics
    c.CrewMetrics.mu.Lock()
    c.CrewMetrics.ExecutionCount++
    c.CrewMetrics.TotalTokens += estimatedTokens
    c.CrewMetrics.DailyCost += totalActualCost
    c.CrewMetrics.mu.Unlock()
    
    return &Result{
        Outputs: results,
        Cost:    totalActualCost,
    }, nil
}
```

### Step 5: Configuration Loading (config.go)

**Location:** `internal/config.go`, update `LoadCrewConfig()`

```go
func LoadCrewConfig(path string) (*Crew, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    crew := &Crew{
        CrewMetrics: CrewCostMetrics{
            LastResetTime: time.Now(),
            AgentCosts:    make(map[string]float64),
        },
    }
    
    err = yaml.Unmarshal(data, &crew)
    if err != nil {
        return nil, err
    }
    
    // Set defaults if not specified
    if crew.MaxCostPerExecution == 0 {
        crew.MaxCostPerExecution = 10.0  // Default $10/execution
    }
    if crew.MaxCostPerDay == 0 {
        crew.MaxCostPerDay = 100.0  // Default $100/day
    }
    if crew.MaxTokensPerExecution == 0 {
        crew.MaxTokensPerExecution = 50000  // Default 50K tokens
    }
    
    return crew, nil
}
```

---

## Testing Strategy

### Test 1: Crew Respects Per-Execution Cost Limit

**File:** `internal/crew_test.go`

```go
func TestCrew_PerExecutionCostLimit(t *testing.T) {
    crew := &Crew{
        Name: "test_crew",
        MaxCostPerExecution: 2.50,  // Max $2.50 per execution
        MaxCostPerDay:       100.0,
        MaxTokensPerExecution: 20000,
        Agents: []*Agent{
            {ID: "agent1", MaxTokensPerCall: 5000},
            {ID: "agent2", MaxTokensPerCall: 5000},
        },
        CrewMetrics: CrewCostMetrics{
            LastResetTime: time.Now(),
            AgentCosts:    make(map[string]float64),
        },
    }
    
    // Create request that costs $3.00 (exceeds $2.50 limit)
    req := &Request{Content: strings.Repeat("a", 50000)}  // High cost
    
    estimatedCost := crew.estimateTotalCostForRequest(req)
    estimatedTokens := crew.estimateTotalTokensForRequest(req)
    
    err := crew.checkCrewBudget(estimatedCost, estimatedTokens)
    
    // Should error - exceeds per-execution limit
    if err == nil {
        t.Fatal("expected error for exceeding per-execution limit")
    }
    
    if !strings.Contains(err.Error(), "BUDGET BLOCK") {
        t.Fatalf("expected BUDGET BLOCK error, got: %v", err)
    }
}
```

### Test 2: Crew Respects Daily Limit

**File:** `internal/crew_test.go`

```go
func TestCrew_DailyLimit(t *testing.T) {
    crew := &Crew{
        Name: "test_crew",
        MaxCostPerExecution: 10.0,
        MaxCostPerDay: 20.0,  // Max $20/day
        MaxTokensPerExecution: 50000,
        Agents: []*Agent{
            {ID: "agent1"},
        },
        CrewMetrics: CrewCostMetrics{
            LastResetTime: time.Now(),
            DailyCost:     18.0,  // Already used $18
            AgentCosts:    make(map[string]float64),
        },
    }
    
    // Try to spend $5 more (would total $23 > $20 limit)
    req := &Request{Content: "test"}
    
    // Manually set cost to test
    err := crew.checkCrewBudget(5.0, 10000)
    
    // Should error - would exceed daily limit
    if err == nil {
        t.Fatal("expected error for exceeding daily limit")
    }
}
```

### Test 3: Multiple Agents Respect Crew Hard Cap

**File:** `internal/crew_test.go`

```go
func TestCrew_MultipleAgentsUnderHardCap(t *testing.T) {
    crew := &Crew{
        Name: "test_crew",
        MaxCostPerExecution: 2.50,  // Crew limit is $2.50
        MaxCostPerDay: 100.0,
        MaxTokensPerExecution: 20000,
        Agents: []*Agent{
            {ID: "router", MaxCostPerDay: 10.0},       // Agent could spend $10
            {ID: "faq_searcher", MaxCostPerDay: 20.0}, // Agent could spend $20
            {ID: "searcher", MaxCostPerDay: 15.0},     // Agent could spend $15
            {ID: "aggregator", MaxCostPerDay: 10.0},   // Agent could spend $10
            // Total agent limits: $55/day
        },
        CrewMetrics: CrewCostMetrics{
            LastResetTime: time.Now(),
            AgentCosts:    make(map[string]float64),
        },
    }
    
    // Request that would cost $1.50 (within crew cap)
    req := &Request{Content: strings.Repeat("a", 5000)}
    estimatedCost := 1.50
    
    err := crew.checkCrewBudget(estimatedCost, 5000)
    
    // Should allow - within crew hard cap
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### Test 4: Agent Limits are Informational (Not Blocking)

**File:** `internal/crew_test.go`

```go
func TestCrew_AgentLimitsAreInformational(t *testing.T) {
    agent := &Agent{
        ID:            "expensive_agent",
        MaxCostPerDay: 10.0,  // Agent limit
        EnforceCostLimits: true,
    }
    
    crew := &Crew{
        Name: "test_crew",
        MaxCostPerExecution: 50.0,  // Crew limit much higher
        MaxCostPerDay: 1000.0,
        Agents: []*Agent{agent},
        CrewMetrics: CrewCostMetrics{
            LastResetTime: time.Now(),
            AgentCosts:    make(map[string]float64),
        },
    }
    
    // Agent cost is $15 (exceeds agent's $10 limit)
    // But crew can afford it
    estimatedCost := 15.0
    
    err := crew.checkCrewBudget(estimatedCost, 5000)
    
    // Should allow - crew budget is OK
    // Agent warnings are logged but don't block
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### Test 5: Crew Metrics Tracking and Daily Reset

**File:** `internal/crew_test.go`

```go
func TestCrew_MetricsTrackingAndReset(t *testing.T) {
    crew := &Crew{
        Name: "test_crew",
        CrewMetrics: CrewCostMetrics{
            LastResetTime: time.Now().Add(-25 * time.Hour),  // 25 hours ago
            ExecutionCount: 10,
            DailyCost: 75.0,  // Already spent $75
            TotalTokens: 100000,
            AgentCosts: make(map[string]float64),
        },
    }
    
    // Check if reset happens
    crew.checkAndResetDailyMetrics()
    
    crew.CrewMetrics.mu.RLock()
    
    // Should be reset
    if crew.CrewMetrics.DailyCost != 0 {
        t.Fatalf("expected DailyCost=0 after reset, got %.2f", crew.CrewMetrics.DailyCost)
    }
    
    if crew.CrewMetrics.ExecutionCount != 0 {
        t.Fatalf("expected ExecutionCount=0 after reset, got %d", crew.CrewMetrics.ExecutionCount)
    }
    
    crew.CrewMetrics.mu.RUnlock()
}
```

---

## Implementation Timeline

**Week 2: Crew-Level Cost Control**

| Day | Task | Hours | Files |
|-----|------|-------|-------|
| Mon | Update types.go with Crew fields | 0.5 | types.go |
| Tue | Implement cost estimation functions | 1.0 | crew.go |
| Wed | Add checkCrewBudget() function | 1.5 | crew.go |
| Thu | Integrate into Execute(), update config.go | 1.5 | crew.go, config.go |
| Fri | Write 5 tests, verify all pass | 2.0 | crew_test.go |
| **TOTAL** | **Crew-Level Implementation** | **7 hours** | **3 files** |

---

## Success Criteria

- [ ] Crew struct has 3 new cost control fields
- [ ] Crew YAML config loads cost fields correctly
- [ ] `checkCrewBudget()` enforces per-execution limit
- [ ] `checkCrewBudget()` enforces per-day limit
- [ ] Crew budget takes priority over agent limits
- [ ] Agent informational warnings are logged (not blocking)
- [ ] Metrics are tracked correctly (ExecutionCount, TotalTokens, DailyCost, AgentCosts)
- [ ] Daily reset happens after 24 hours
- [ ] All 5 tests pass with > 80% code coverage
- [ ] No race conditions (verified with `-race` flag)
- [ ] Debug logging shows budget decisions
- [ ] Configuration examples work correctly

---

## Deployment Checklist

- [ ] Code reviewed (2+ approvals)
- [ ] Tests passing with -race flag
- [ ] No performance regression
- [ ] Configuration examples provided
- [ ] Documentation updated
- [ ] Integration with agent controls verified
- [ ] Staging deployment successful

---

## Budget Hierarchy Reference

### HARD CAP (Crew) - ALWAYS ENFORCED:
```yaml
crew:
  max_cost_per_execution: 2.50  # 🔴 NO OVERRIDE
  max_cost_per_day: 100.00      # 🔴 NO OVERRIDE
```

### INFORMATIONAL (Agents) - NEVER BLOCKING:
```yaml
agents:
  - id: router
    max_cost_per_day: 10.00     # ℹ️ ADVISORY ONLY
  - id: searcher
    max_cost_per_day: 20.00     # ℹ️ ADVISORY ONLY
```

---

## Related Tech-Specs

- **tech-spec-agent-cost-control.md** - Per-agent configurable enforcement
- **tech-spec-message-history-limit.md** - Message history pruning

---

**Status:** ✅ Ready for Development  
**Owner:** [Crew Team Lead]  
**Next Step:** Begin Week 2 implementation

