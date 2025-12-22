# Tech-Spec: Agent-Level Cost Control with Configurable Enforcement

**Created:** 2025-12-22  
**Status:** Ready for Development  
**Priority:** HIGH (Cost Management)  
**Estimated Impact:** Per-agent budget tracking + flexible enforcement  

---

## Overview

### Problem Statement

Agents currently have no per-agent cost control mechanisms. This leads to:

1. **Unpredictable Agent Costs:**
   - Some agents (like faq_searcher) may make expensive calls
   - No way to limit individual agent spending
   - One agent can consume entire crew budget

2. **Lack of Flexibility:**
   - Different agents have different cost profiles
   - Some agents should be strict (router, orchestrator)
   - Some agents should be flexible (searchers, explorers)
   - No way to configure this per-agent

3. **No Cost Visibility:**
   - Can't track which agents cost how much
   - Can't see cost trends per agent
   - Can't predict daily agent spending

**Current Behavior:**
```go
// No cost controls per agent
type Agent struct {
    ID    string
    Model string
    // ... no cost tracking
}

// All cost control happens at crew level (future)
// Agents are transparent - no budget enforcement
```

**Impact:**
- Cannot control runaway agents
- Cannot implement cost fairness across agents
- Cannot warn early about high-cost agents
- No alerting for agents approaching limits

### Solution

Implement **per-agent cost control** with configurable enforcement strategy:

**Key Decision: CONFIGURABLE Approach**
- Each agent independently decides: Block or Warn
- `EnforceCostLimits: true` → BLOCK if exceeded (default, safe)
- `EnforceCostLimits: false` → WARN only, continue (flexible)

**Architecture:**
```
Agent Receives Request
    ↓
Estimate tokens needed
    ↓
Check cost limits:
    ├─ if EnforceCostLimits=true:
    │  └─ Block if exceeds → Return Error ❌
    │
    └─ if EnforceCostLimits=false:
       └─ Warn if exceeds → Continue ⚠️
    ↓
Execute agent call
    ↓
Update metrics (tokens, cost, call count)
```

**Why CONFIGURABLE?**
- Flexibility: Each team chooses for their agent
- Safety: Default is to block (secure-by-default)
- Control: Critical agents can be strict, experimental agents can be flexible
- Production-Ready: No confusion, clear semantics

### Scope (In/Out)

**IN SCOPE:**
- ✅ Update `Agent` type with 5 cost control fields
- ✅ Implement `checkAgentCostLimits()` function
- ✅ Integrate cost checks into `Execute()` method
- ✅ Track per-agent metrics (CallCount, TotalTokens, DailyCost)
- ✅ Add token estimation for cost calculation
- ✅ Load cost config from agent YAML
- ✅ Comprehensive test coverage (5 test cases)
- ✅ Debug logging for cost decisions

**OUT OF SCOPE:**
- ❌ Crew-level cost controls (separate tech-spec)
- ❌ Cost reporting/dashboard (separate feature)
- ❌ Billing integration (handled elsewhere)
- ❌ Historical cost analysis (separate feature)

---

## Context for Development

### Codebase Patterns

**Agent Type Pattern (types.go):**
```go
// Current: No cost fields
type Agent struct {
    ID       string
    Name     string
    Role     string
    Model    string
    Tools    []Tool
    // ... other fields
}

// New: Add cost control fields
type Agent struct {
    // ... existing fields ...
    
    // Cost Control Configuration
    MaxTokensPerCall   int     // Limit per single call (e.g., 1000)
    MaxTokensPerDay    int     // Limit per 24 hours (e.g., 50000)
    MaxCostPerDay      float64 // Daily budget (e.g., 10.00)
    CostAlertThreshold float64 // Warning at % usage (e.g., 0.80 = 80%)
    EnforceCostLimits  bool    // true=block, false=warn (default: true)
    
    // Runtime Metrics
    CostMetrics AgentCostMetrics
}

// New: Metrics tracking structure
type AgentCostMetrics struct {
    CallCount      int
    TotalTokens    int
    DailyCost      float64
    LastResetTime  time.Time
    mu             sync.RWMutex
}
```

**Configuration Pattern (YAML):**
```yaml
# agents/router.yaml
agent:
  id: router
  name: "Query Router"
  role: "orchestrator"
  
  # NEW: Cost Control Configuration
  max_tokens_per_call: 1000        # Block calls > 1000 tokens
  max_tokens_per_day: 50000        # Reset daily at midnight
  max_cost_per_day: 10.00          # Stop at $10/day
  cost_alert_threshold: 0.80       # Warn at 80% usage
  enforce_cost_limits: true        # BLOCK if exceeded (safe default)

# agents/faq_searcher.yaml
agent:
  id: faq_searcher
  name: "FAQ Knowledge Base"
  
  # NEW: Cost Control Configuration
  max_tokens_per_call: 2000        # Higher limit (search needs more)
  max_tokens_per_day: 100000       # Higher daily budget
  max_cost_per_day: 20.00          # Higher daily limit
  cost_alert_threshold: 0.80
  enforce_cost_limits: false       # WARN only (flexible)
```

### Token Estimation Method

Uses character-to-token estimation (built-in pattern):
```go
// Existing pattern in agent.go
func (a *Agent) estimateTokens(content string) int {
    // Rough estimation: 1 token ≈ 4 characters
    // Actual: OpenAI tokenizer, but this is good enough
    return len(content) / 4
}
```

**Token-to-Cost Calculation:**
```go
// Standard OpenAI pricing (example)
const CostPerToken float64 = 0.0000015  // $0.15 per 1M tokens

estimatedCost := float64(estimatedTokens) * CostPerToken
// 1000 tokens = $0.0015
// 10000 tokens = $0.015
```

### Cost Check Implementation Pattern

```go
// New function: Check cost limits
func (a *Agent) checkAgentCostLimits(estimatedTokens int) error {
    // Step 1: Calculate estimated cost
    estimatedCost := float64(estimatedTokens) * a.CostPerToken
    
    // Step 2: Check if in warn-only mode
    if !a.EnforceCostLimits {
        // Warn if approaching limit, but don't block
        if a.CostMetrics.DailyCost + estimatedCost > a.MaxCostPerDay * 0.8 {
            log.Printf("⚠️ [WARN] Agent %s approaching cost limit: $%.2f/day",
                a.ID, a.MaxCostPerDay)
        }
        return nil  // Always allow execution in warn mode
    }
    
    // Step 3: Check strict mode (EnforceCostLimits=true)
    // Check per-call limit
    if estimatedTokens > a.MaxTokensPerCall {
        return fmt.Errorf("❌ [BLOCK] Agent %s: request %d tokens > limit %d tokens",
            a.ID, estimatedTokens, a.MaxTokensPerCall)
    }
    
    // Check daily budget
    newDailyCost := a.CostMetrics.DailyCost + estimatedCost
    if newDailyCost > a.MaxCostPerDay {
        return fmt.Errorf("❌ [BLOCK] Agent %s: daily cost $%.2f + $%.2f > limit $%.2f",
            a.ID, a.CostMetrics.DailyCost, estimatedCost, a.MaxCostPerDay)
    }
    
    return nil  // Cost checks passed
}
```

### Integration Point: Execute() Method

Current pattern (agent.go):
```go
func (a *Agent) Execute(ctx context.Context, input string) (string, error) {
    // Step 1: Prepare request
    req := &Request{...}
    
    // Step 2: Call LLM
    resp, err := a.client.Call(ctx, req)
    if err != nil {
        return "", err
    }
    
    // NEW: Add cost control here
    // After preparing request but BEFORE calling LLM
    tokens := a.estimateTokens(input)
    if err := a.checkAgentCostLimits(tokens); err != nil {
        return "", err  // Return error if cost exceeded
    }
    
    // Step 3: Update metrics AFTER successful execution
    a.CostMetrics.mu.Lock()
    a.CostMetrics.CallCount++
    a.CostMetrics.TotalTokens += tokens
    a.CostMetrics.DailyCost += cost
    a.CostMetrics.mu.Unlock()
    
    return resp.Content, nil
}
```

---

## Detailed Design

### Step 1: Type Definitions (agent.go)

Add to Agent struct (~25 lines):

**Location:** `internal/types.go`, around line 150-200 where Agent is defined

```go
type Agent struct {
    // ... existing fields (ID, Name, Role, Model, Tools, etc.) ...
    
    // Cost Control Configuration
    MaxTokensPerCall   int     `yaml:"max_tokens_per_call"`   // Max tokens per single call
    MaxTokensPerDay    int     `yaml:"max_tokens_per_day"`    // Max cumulative tokens per day
    MaxCostPerDay      float64 `yaml:"max_cost_per_day"`      // Max daily budget in USD
    CostAlertThreshold float64 `yaml:"cost_alert_threshold"`  // Warn when % of budget used (0.0-1.0)
    EnforceCostLimits  bool    `yaml:"enforce_cost_limits"`   // true=block, false=warn
    
    // Runtime metrics (not serialized)
    CostMetrics AgentCostMetrics `json:"-" yaml:"-"`
}

// New struct for metrics
type AgentCostMetrics struct {
    CallCount      int       // Number of calls this period
    TotalTokens    int       // Total tokens used
    DailyCost      float64   // Total cost today
    LastResetTime  time.Time // When daily counter resets
    mu             sync.RWMutex
}
```

### Step 2: Token Estimation (agent.go)

Verify existing method works:

**Location:** `internal/agent.go`, look for `estimateTokens()` method

```go
// Already exists - just verify
func (a *Agent) estimateTokens(content string) int {
    // Rough estimation: 1 token ≈ 4 characters
    return len(content) / 4
}
```

If doesn't exist, add it:
```go
func (a *Agent) estimateTokens(content string) int {
    // Implementation: Estimate tokens
    // Using: 1 token ≈ 4 characters (OpenAI approximation)
    return len(content) / 4
}

// Helper: Calculate cost
func (a *Agent) calculateCost(tokens int) float64 {
    // OpenAI pricing: ~$0.15 per 1M input tokens
    const CostPerToken = 0.00000015
    return float64(tokens) * CostPerToken
}
```

### Step 3: Cost Limit Checker (agent.go)

**Location:** `internal/agent.go`, add new function ~40 lines

```go
// checkAgentCostLimits verifies the agent hasn't exceeded its cost limits
func (a *Agent) checkAgentCostLimits(estimatedTokens int) error {
    // Initialize metrics if needed
    if a.CostMetrics.LastResetTime.IsZero() {
        a.CostMetrics.LastResetTime = time.Now()
    }
    
    // Check if daily reset needed
    a.checkAndResetDailyMetrics()
    
    // Calculate estimated cost
    estimatedCost := a.calculateCost(estimatedTokens)
    
    // Mode 1: Warn only (do not block)
    if !a.EnforceCostLimits {
        a.CostMetrics.mu.RLock()
        currentCost := a.CostMetrics.DailyCost
        a.CostMetrics.mu.RUnlock()
        
        // Warn at alert threshold
        if currentCost > a.MaxCostPerDay * a.CostAlertThreshold {
            log.Printf("⚠️  [WARN] Agent %s approaching cost limit: $%.4f / $%.2f (%.0f%%)",
                a.ID, currentCost, a.MaxCostPerDay,
                (currentCost/a.MaxCostPerDay)*100)
        }
        return nil  // Never block in warn mode
    }
    
    // Mode 2: Enforce limits (block on exceeded)
    
    // Check 1: Per-call token limit
    if estimatedTokens > a.MaxTokensPerCall {
        return fmt.Errorf(
            "❌ [COST BLOCK] Agent '%s': request exceeds per-call limit "+
            "(estimated %d tokens > limit %d tokens)",
            a.ID, estimatedTokens, a.MaxTokensPerCall)
    }
    
    // Check 2: Daily budget
    a.CostMetrics.mu.RLock()
    currentDailyCost := a.CostMetrics.DailyCost
    a.CostMetrics.mu.RUnlock()
    
    newDailyCost := currentDailyCost + estimatedCost
    if newDailyCost > a.MaxCostPerDay {
        return fmt.Errorf(
            "❌ [COST BLOCK] Agent '%s': would exceed daily budget "+
            "(current $%.4f + estimated $%.4f > limit $%.2f)",
            a.ID, currentDailyCost, estimatedCost, a.MaxCostPerDay)
    }
    
    return nil
}

// checkAndResetDailyMetrics resets daily counters if a day has passed
func (a *Agent) checkAndResetDailyMetrics() {
    a.CostMetrics.mu.Lock()
    defer a.CostMetrics.mu.Unlock()
    
    now := time.Now()
    // If last reset was > 24 hours ago, reset
    if now.Sub(a.CostMetrics.LastResetTime) > 24*time.Hour {
        a.CostMetrics.DailyCost = 0
        a.CostMetrics.LastResetTime = now
    }
}
```

### Step 4: Integration with Execute() (agent.go)

**Location:** `internal/agent.go`, modify `Execute()` method

Current code:
```go
func (a *Agent) Execute(ctx context.Context, input string) (string, error) {
    // ... prepare request ...
    
    resp, err := a.client.Call(ctx, req)
    
    // ... handle response ...
}
```

Modified code:
```go
func (a *Agent) Execute(ctx context.Context, input string) (string, error) {
    // Step 1: Estimate cost BEFORE execution
    estimatedTokens := a.estimateTokens(input)
    
    // Step 2: Check cost limits
    if err := a.checkAgentCostLimits(estimatedTokens); err != nil {
        return "", err  // Return error, don't execute
    }
    
    // Step 3: Original execution code
    req := &Request{
        Content: input,
        // ... other fields ...
    }
    
    resp, err := a.client.Call(ctx, req)
    if err != nil {
        return "", err
    }
    
    // Step 4: Update metrics AFTER successful execution
    actualTokens := a.estimateTokens(resp.Content)  // Or track from API response
    actualCost := a.calculateCost(actualTokens)
    
    a.CostMetrics.mu.Lock()
    a.CostMetrics.CallCount++
    a.CostMetrics.TotalTokens += estimatedTokens
    a.CostMetrics.DailyCost += actualCost
    a.CostMetrics.mu.Unlock()
    
    return resp.Content, nil
}
```

### Step 5: Configuration Loading (config.go)

**Location:** `internal/config.go`, update `LoadAgentConfig()`

```go
func LoadAgentConfig(path string) (*Agent, error) {
    // Load YAML file
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    agent := &Agent{
        CostMetrics: AgentCostMetrics{
            LastResetTime: time.Now(),
        },
    }
    
    err = yaml.Unmarshal(data, &agent)
    if err != nil {
        return nil, err
    }
    
    // Set defaults if not specified
    if agent.MaxTokensPerCall == 0 {
        agent.MaxTokensPerCall = 5000  // Default
    }
    if agent.MaxTokensPerDay == 0 {
        agent.MaxTokensPerDay = 1000000  // Default
    }
    if agent.MaxCostPerDay == 0 {
        agent.MaxCostPerDay = 100.00  // Default $100/day
    }
    if agent.CostAlertThreshold == 0 {
        agent.CostAlertThreshold = 0.80  // Default 80%
    }
    if !agent.EnforceCostLimits {
        agent.EnforceCostLimits = true  // Default safe
    }
    
    return agent, nil
}
```

---

## Testing Strategy

### Test 1: Agent Respects Per-Call Token Limit (Block Mode)

**File:** `internal/agent_test.go`

```go
func TestAgent_BlockMode_PerCallTokenLimit(t *testing.T) {
    agent := &Agent{
        ID:                "test_agent",
        MaxTokensPerCall:  100,    // Very restrictive for testing
        MaxCostPerDay:     1000.0,
        EnforceCostLimits: true,
        CostMetrics: AgentCostMetrics{
            LastResetTime: time.Now(),
        },
    }
    
    // Create request with > 100 tokens (400 characters = 100 tokens)
    largeInput := strings.Repeat("a", 500)  // 125 tokens
    
    err := agent.checkAgentCostLimits(agent.estimateTokens(largeInput))
    
    // Should error in block mode
    if err == nil {
        t.Fatal("expected error for exceeding per-call token limit")
    }
    
    if !strings.Contains(err.Error(), "COST BLOCK") {
        t.Fatalf("expected COST BLOCK error, got: %v", err)
    }
}
```

### Test 2: Agent Warns But Allows in Warn Mode

**File:** `internal/agent_test.go`

```go
func TestAgent_WarnMode_AllowsExceedingLimit(t *testing.T) {
    agent := &Agent{
        ID:                 "test_agent",
        MaxTokensPerCall:   100,
        MaxCostPerDay:      1000.0,
        EnforceCostLimits:  false,  // WARN MODE
        CostMetrics: AgentCostMetrics{
            LastResetTime: time.Now(),
        },
    }
    
    // Create request with > 100 tokens
    largeInput := strings.Repeat("a", 500)  // 125 tokens
    
    err := agent.checkAgentCostLimits(agent.estimateTokens(largeInput))
    
    // Should NOT error in warn mode
    if err != nil {
        t.Fatalf("expected no error in warn mode, got: %v", err)
    }
}
```

### Test 3: Configurable Per Agent (Mixed Enforcement)

**File:** `internal/agent_test.go`

```go
func TestAgent_ConfigurableEnforcement(t *testing.T) {
    strictAgent := &Agent{
        ID: "router",
        EnforceCostLimits: true,  // Strict
        MaxTokensPerCall: 100,
        CostMetrics: AgentCostMetrics{LastResetTime: time.Now()},
    }
    
    flexibleAgent := &Agent{
        ID: "searcher",
        EnforceCostLimits: false,  // Flexible
        MaxTokensPerCall: 100,
        CostMetrics: AgentCostMetrics{LastResetTime: time.Now()},
    }
    
    largeInput := strings.Repeat("a", 500)  // Exceeds limit
    
    // Strict agent should block
    err := strictAgent.checkAgentCostLimits(strictAgent.estimateTokens(largeInput))
    if err == nil {
        t.Fatal("expected error for strict agent")
    }
    
    // Flexible agent should allow
    err = flexibleAgent.checkAgentCostLimits(flexibleAgent.estimateTokens(largeInput))
    if err != nil {
        t.Fatalf("unexpected error for flexible agent: %v", err)
    }
}
```

### Test 4: Metrics Tracking

**File:** `internal/agent_test.go`

```go
func TestAgent_MetricsTracking(t *testing.T) {
    agent := &Agent{
        ID: "test",
        EnforceCostLimits: false,  // Allow execution
        MaxCostPerDay: 1000.0,
        CostMetrics: AgentCostMetrics{
            LastResetTime: time.Now(),
        },
    }
    
    // Mock execution to update metrics
    input := "test input"
    tokens := agent.estimateTokens(input)
    cost := agent.calculateCost(tokens)
    
    agent.CostMetrics.mu.Lock()
    agent.CostMetrics.CallCount++
    agent.CostMetrics.TotalTokens += tokens
    agent.CostMetrics.DailyCost += cost
    agent.CostMetrics.mu.Unlock()
    
    // Verify metrics updated
    agent.CostMetrics.mu.RLock()
    if agent.CostMetrics.CallCount != 1 {
        t.Fatalf("expected CallCount=1, got %d", agent.CostMetrics.CallCount)
    }
    if agent.CostMetrics.TotalTokens == 0 {
        t.Fatal("expected TotalTokens > 0")
    }
    if agent.CostMetrics.DailyCost == 0 {
        t.Fatal("expected DailyCost > 0")
    }
    agent.CostMetrics.mu.RUnlock()
}
```

### Test 5: Daily Reset

**File:** `internal/agent_test.go`

```go
func TestAgent_DailyMetricsReset(t *testing.T) {
    agent := &Agent{
        ID: "test",
        CostMetrics: AgentCostMetrics{
            LastResetTime: time.Now().Add(-25 * time.Hour),  // Over 24 hours ago
            DailyCost:     50.0,  // Has accumulated cost
        },
    }
    
    // Call reset
    agent.checkAndResetDailyMetrics()
    
    // Should be reset
    agent.CostMetrics.mu.RLock()
    if agent.CostMetrics.DailyCost != 0 {
        t.Fatalf("expected DailyCost=0 after reset, got %f", agent.CostMetrics.DailyCost)
    }
    agent.CostMetrics.mu.RUnlock()
}
```

---

## Implementation Timeline

**Week 1: Agent-Level Cost Control**

| Day | Task | Hours | Files |
|-----|------|-------|-------|
| Mon | Update types.go with Agent fields | 0.5 | types.go |
| Tue | Implement estimateTokens(), calculateCost() | 0.5 | agent.go |
| Wed | Add checkAgentCostLimits() function | 1.5 | agent.go |
| Thu | Integrate into Execute(), update config.go | 1.5 | agent.go, config.go |
| Fri | Write 5 tests, verify all pass | 2.0 | agent_test.go |
| **TOTAL** | **Agent-Level Implementation** | **7 hours** | **3 files** |

---

## Success Criteria

- [ ] Agent struct has 5 new cost control fields
- [ ] Agent YAML config loads cost fields correctly
- [ ] `checkAgentCostLimits()` blocks in strict mode
- [ ] `checkAgentCostLimits()` warns but allows in flexible mode
- [ ] Metrics are tracked correctly (CallCount, TotalTokens, DailyCost)
- [ ] Daily reset happens after 24 hours
- [ ] All 5 tests pass with > 80% code coverage
- [ ] No race conditions (verified with `-race` flag)
- [ ] Debug logging shows when blocks/warns occur
- [ ] Configuration examples work correctly

---

## Deployment Checklist

- [ ] Code reviewed (2+ approvals)
- [ ] Tests passing with -race flag
- [ ] No performance regression
- [ ] Configuration examples provided
- [ ] Documentation updated
- [ ] Team trained on configuration
- [ ] Staging deployment successful

---

## Related Tech-Specs

- **tech-spec-crew-cost-control.md** - Crew-level hard cap
- **tech-spec-message-history-limit.md** - Message history pruning

---

**Status:** ✅ Ready for Development  
**Owner:** [Agent Team Lead]  
**Next Step:** Begin Week 1 implementation

