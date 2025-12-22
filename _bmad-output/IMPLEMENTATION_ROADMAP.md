# 🚀 Implementation Roadmap: Agent & Crew Cost Controls

**Status:** ✅ Tech-Specs Complete, Ready for Implementation  
**Date:** 2025-12-22  
**Decision:** CONFIGURABLE agents + CREW HARD CAP  
**Timeline:** 3 weeks (Week 1-3)  
**Team:** 2-3 developers

---

## 📋 Overview

This roadmap guides the implementation of cost controls for agents and crews based on two key decisions:

1. **Decision #1: Agent Cost Blocking = CONFIGURABLE** ✅
   - Each agent chooses: `enforce_cost_limits: true` (block) or `false` (warn)
   - Tech-Spec: `tech-spec-agent-cost-control.md`

2. **Decision #2: Budget Hierarchy = CREW HARD CAP** ✅
   - Crew limits = absolute maximum (HARD CAP)
   - Agent limits = informational only (warnings, no blocking)
   - Tech-Spec: `tech-spec-crew-cost-control.md`

---

## 📚 Reference Documents

All technical details in:

| Document | Purpose | Size | Owner |
|----------|---------|------|-------|
| **tech-spec-agent-cost-control.md** | Agent-level implementation | 8 KB | Agent Team |
| **tech-spec-crew-cost-control.md** | Crew-level implementation | 10 KB | Crew Team |
| **TEAM_DISCUSSION_BRIEF.md** | Team discussion context | 18 KB | All |
| **FINAL_DECISION_SUMMARY.md** | Decision documentation | 12 KB | All |

---

## 🎯 WEEK 1: Agent-Level Cost Control

### Objective
Implement per-agent cost tracking with configurable enforcement (block vs warn)

### Tasks

#### Monday: Update Types (30 minutes)

**File:** `internal/types.go`  
**Task:** Add 5 cost control fields to Agent struct

```go
type Agent struct {
    // ... existing fields ...
    
    // NEW: Cost Control Fields
    MaxTokensPerCall   int     `yaml:"max_tokens_per_call"`   // e.g., 1000
    MaxTokensPerDay    int     `yaml:"max_tokens_per_day"`    // e.g., 50000
    MaxCostPerDay      float64 `yaml:"max_cost_per_day"`      // e.g., 10.00
    CostAlertThreshold float64 `yaml:"cost_alert_threshold"`  // e.g., 0.80
    EnforceCostLimits  bool    `yaml:"enforce_cost_limits"`   // true=block, false=warn
    
    // NEW: Runtime Metrics
    CostMetrics AgentCostMetrics `json:"-" yaml:"-"`
}

type AgentCostMetrics struct {
    CallCount      int
    TotalTokens    int
    DailyCost      float64
    LastResetTime  time.Time
    mu             sync.RWMutex
}
```

**Verification:**
- [ ] Agent struct compiles
- [ ] YAML marshaling works
- [ ] No type conflicts

---

#### Tuesday: Token Estimation (30 minutes)

**File:** `internal/agent.go`  
**Task:** Implement token estimation and cost calculation

```go
// Verify this exists or add if missing
func (a *Agent) estimateTokens(content string) int {
    return len(content) / 4  // 1 token ≈ 4 characters
}

// Add cost calculation
func (a *Agent) calculateCost(tokens int) float64 {
    const CostPerToken = 0.00000015  // $0.15 per 1M tokens
    return float64(tokens) * CostPerToken
}
```

**Verification:**
- [ ] estimateTokens() works for various inputs
- [ ] calculateCost() returns reasonable values
- [ ] Token-to-cost conversion is accurate

---

#### Wednesday: Cost Check Function (1.5 hours)

**File:** `internal/agent.go`  
**Task:** Implement cost limit checking with block/warn modes

Key function: `checkAgentCostLimits(estimatedTokens int) error`

Features:
- ✅ Block mode (EnforceCostLimits=true): Return error if exceeded
- ✅ Warn mode (EnforceCostLimits=false): Log warning, allow execution
- ✅ Per-call token limits
- ✅ Daily budget limits
- ✅ Daily reset (24-hour window)
- ✅ Alert threshold warnings (80% usage)

See tech-spec for full implementation.

**Verification:**
- [ ] Block mode blocks correctly
- [ ] Warn mode warns but allows
- [ ] Daily reset happens properly
- [ ] Thread-safe with mutex

---

#### Thursday: Integration (1.5 hours)

**Files:** `internal/agent.go`, `internal/config.go`  
**Tasks:**
1. Integrate cost checks into `Execute()` method
2. Update metrics after successful execution
3. Update `LoadAgentConfig()` to parse YAML fields
4. Set sensible defaults

**Code Flow:**
```
Execute(input)
    ├─ Estimate tokens
    ├─ Check cost limits → if error, return early
    ├─ Execute agent call
    ├─ Update metrics
    └─ Return result
```

**Verification:**
- [ ] Cost checks happen before execution
- [ ] Metrics update after execution
- [ ] YAML config loads correctly
- [ ] Defaults work if not specified

---

#### Friday: Testing (2 hours)

**File:** `internal/agent_test.go`  
**Write 5 tests:**

1. **TestAgent_BlockMode_PerCallTokenLimit** - Block mode blocks correctly
2. **TestAgent_WarnMode_AllowsExceedingLimit** - Warn mode allows execution
3. **TestAgent_ConfigurablePerAgent** - Different agents have different enforcement
4. **TestAgent_MetricsTracking** - Metrics updated correctly
5. **TestAgent_DailyMetricsReset** - Daily reset after 24 hours

**Run:**
```bash
go test ./internal -v -race -cover
```

**Success Criteria:**
- [ ] All 5 tests pass
- [ ] Coverage > 80%
- [ ] No race conditions (-race flag)
- [ ] Metrics logged for debug

**Total Week 1:** 7 hours (1 developer)

---

## 🎯 WEEK 2: Crew-Level Cost Control

### Objective
Implement crew-wide hard cap enforcement that gates all agent executions

### Tasks

#### Monday: Update Crew Types (30 minutes)

**File:** `internal/types.go`  
**Task:** Add 3 cost control fields to Crew struct

```go
type Crew struct {
    // ... existing fields ...
    
    // NEW: Cost Control Fields (HARD CAPS)
    MaxCostPerExecution float64 `yaml:"max_cost_per_execution"`   // e.g., 2.50
    MaxCostPerDay       float64 `yaml:"max_cost_per_day"`         // e.g., 100.00
    MaxTokensPerExecution int   `yaml:"max_tokens_per_execution"` // e.g., 20000
    
    // NEW: Runtime Metrics
    CrewMetrics CrewCostMetrics `json:"-" yaml:"-"`
}

type CrewCostMetrics struct {
    ExecutionCount  int
    TotalTokens     int
    DailyCost       float64
    LastResetTime   time.Time
    AgentCosts      map[string]float64
    mu              sync.RWMutex
}
```

**Verification:**
- [ ] Crew struct compiles
- [ ] YAML marshaling works
- [ ] No conflicts with agent fields

---

#### Tuesday: Cost Estimation (1 hour)

**File:** `internal/crew.go`  
**Task:** Implement crew-level cost estimation

Functions needed:
1. `estimateTotalCostForRequest()` - Sum all agent costs
2. `estimateAgentCost()` - Single agent cost
3. `estimateTotalTokensForRequest()` - Sum all tokens

Key insight: Estimate for ALL agents together (parallel execution)

**Verification:**
- [ ] Total cost = sum of agent costs
- [ ] Estimates are reasonable
- [ ] No double-counting

---

#### Wednesday: Budget Checker (1.5 hours)

**File:** `internal/crew.go`  
**Task:** Implement `checkCrewBudget()` with hard cap enforcement

Function: `checkCrewBudget(estimatedCost, estimatedTokens) error`

Checks (in order):
1. Per-execution cost limit (HARD CAP)
2. Per-execution token limit (HARD CAP)
3. Daily cost limit (HARD CAP)

Returns error if ANY check fails → **blocks execution**

Important: Agent warnings are logged AFTER budget is approved (in Execute)

**Verification:**
- [ ] All 3 limits enforced
- [ ] Correct error messages
- [ ] Daily reset works
- [ ] Thread-safe

---

#### Thursday: Integration (1.5 hours)

**Files:** `internal/crew.go`, `internal/config.go`  
**Tasks:**
1. Integrate crew budget check into `Execute()` method
2. Call `checkCrewBudget()` BEFORE executing any agents
3. Track per-agent costs in AgentCosts map
4. Update `LoadCrewConfig()` for YAML fields

**Code Flow:**
```
Execute(request)
    ├─ Estimate total cost & tokens
    ├─ 🔴 Check crew budget → if error, BLOCK (return early)
    ├─ For each agent:
    │  ├─ Check agent warnings (ℹ️ informational only)
    │  ├─ Execute agent
    │  └─ Track cost
    └─ Update crew metrics & return
```

**Verification:**
- [ ] Crew check happens FIRST
- [ ] Agents execute only if crew OK
- [ ] Per-agent costs tracked correctly
- [ ] Metrics updated

---

#### Friday: Testing (2 hours)

**File:** `internal/crew_test.go`  
**Write 5 tests:**

1. **TestCrew_PerExecutionCostLimit** - Blocks if per-execution exceeded
2. **TestCrew_DailyLimit** - Blocks if daily limit exceeded
3. **TestCrew_MultipleAgentsUnderHardCap** - Multiple agents OK if crew cap OK
4. **TestCrew_AgentLimitsAreInformational** - Agent limits don't block
5. **TestCrew_MetricsTrackingAndReset** - Metrics and daily reset work

**Run:**
```bash
go test ./internal -v -race -cover
```

**Success Criteria:**
- [ ] All 5 tests pass
- [ ] Coverage > 80%
- [ ] No race conditions
- [ ] Budget checks logged

**Total Week 2:** 7 hours (1-2 developers)

---

## 🎯 WEEK 3: Monitoring & Production

### Objective
Add visibility into costs and prepare for production deployment

### Tasks

#### Monday: Metrics Endpoint (2 hours)

**File:** `internal/http.go` or new `internal/metrics.go`  
**Task:** Create GET `/metrics/crew-costs` endpoint

Response format:
```json
{
  "crew": {
    "execution_count": 42,
    "total_tokens": 105000,
    "daily_cost": 35.50,
    "daily_limit": 100.00,
    "remaining_budget": 64.50
  },
  "agents": {
    "router": {
      "call_count": 42,
      "total_tokens": 20000,
      "daily_cost": 5.00,
      "daily_limit": 10.00
    },
    "faq_searcher": {
      "call_count": 35,
      "total_tokens": 60000,
      "daily_cost": 15.00,
      "daily_limit": 20.00
    }
  }
}
```

**Verification:**
- [ ] Endpoint returns correct data
- [ ] Authentication/authorization (if needed)
- [ ] Proper HTTP status codes

---

#### Tuesday: Dashboard (2 hours)

**Files:** Create `web/dashboard.html`  
**Task:** Create simple HTML dashboard for cost visualization

Features:
- Real-time crew budget status (pie chart)
- Daily cost trend chart
- Per-agent cost breakdown (bar chart)
- Budget remaining display

Uses simple JavaScript + Chart.js (no framework needed)

**Verification:**
- [ ] Dashboard loads
- [ ] Charts display correctly
- [ ] Refreshes properly

---

#### Wednesday: Documentation (2 hours)

**Files:** Create multiple docs in `docs/`

1. **docs/COST_CONTROL_ARCHITECTURE.md** - Overall architecture
2. **docs/CONFIGURATION_GUIDE.md** - How to configure
3. **docs/TROUBLESHOOTING.md** - Common issues
4. **docs/EXAMPLES.md** - Example configurations

Create example configs:
- `config/crew.yaml.example`
- `config/agents/*.yaml.example`

**Verification:**
- [ ] Clear and complete
- [ ] Examples work
- [ ] Team understands

---

#### Thursday: Load Testing (2 hours)

**File:** `internal/load_test.go`  
**Task:** Verify system under load

Test scenarios:
1. 100 parallel executions (stress test)
2. Rapid cost updates (race condition test)
3. Large message histories (integration)
4. Memory usage with high execution count

**Run:**
```bash
go test ./internal -race -bench=. -benchtime=10s
```

**Verification:**
- [ ] No race conditions with `-race`
- [ ] Metrics accurate under load
- [ ] Memory usage acceptable
- [ ] No deadlocks

---

#### Friday: Production Deployment (2 hours)

**Tasks:**
1. Staging deployment
2. Final verification
3. Production rollout
4. Monitoring first 24 hours

**Verification:**
- [ ] Staging deployment works
- [ ] Metrics endpoint working
- [ ] Cost tracking accurate
- [ ] No issues in first 24h production

**Total Week 3:** 10 hours (1-2 developers + DevOps)

---

## 📊 Implementation Summary

| Phase | Duration | Effort | Owner | Status |
|-------|----------|--------|-------|--------|
| **Week 1: Agent Controls** | 5 days | 7 hours | Agent Team | Not Started |
| **Week 2: Crew Controls** | 5 days | 7 hours | Crew Team | Not Started |
| **Week 3: Monitoring** | 5 days | 10 hours | DevOps + QA | Not Started |
| **TOTAL** | **3 weeks** | **24 hours** | **2-3 people** | **Ready** |

### Code Changes Summary

| Component | Files | Lines | Tests |
|-----------|-------|-------|-------|
| Agent Cost Control | 3 | ~60 | 5 |
| Crew Cost Control | 3 | ~70 | 5 |
| Metrics & Monitoring | 3 | ~50 | - |
| **TOTAL** | **9 files** | **~180** | **10** |

---

## ✅ Week-by-Week Checklist

### Week 1 Completion Criteria
- [ ] Agent struct updated with 5 cost fields
- [ ] Token estimation implemented
- [ ] Cost checking implemented (block & warn modes)
- [ ] Integrated into Execute()
- [ ] Config loading updated
- [ ] 5 unit tests passing
- [ ] No race conditions
- [ ] Code reviewed & approved

### Week 2 Completion Criteria
- [ ] Crew struct updated with 3 cost fields
- [ ] Cost estimation for crews implemented
- [ ] Budget checker implemented
- [ ] Integrated into Execute()
- [ ] Config loading updated
- [ ] 5 unit tests passing
- [ ] Integration with agents verified
- [ ] Code reviewed & approved

### Week 3 Completion Criteria
- [ ] Metrics endpoint working
- [ ] Dashboard functioning
- [ ] Documentation complete
- [ ] Load testing passed
- [ ] Staging deployment successful
- [ ] Production deployment successful
- [ ] Monitoring active

---

## 🚀 Getting Started

### Before Week 1

1. **Setup:**
   ```bash
   cd go-agentic
   git checkout -b feat/cost-controls
   ```

2. **Review Tech-Specs:**
   - Read `_bmad-output/tech-spec-agent-cost-control.md` (Agent Team)
   - Read `_bmad-output/tech-spec-crew-cost-control.md` (Crew Team)

3. **Setup Development:**
   ```bash
   go test ./internal -v  # Verify tests work
   ```

### Week 1 Day 1 (Monday)

1. Create feature branch (if not done)
2. Open `tech-spec-agent-cost-control.md`
3. Update `internal/types.go` with Agent fields
4. Run: `go build ./...`
5. Create test file: `internal/agent_test.go`

### Week 1 Day 5 (Friday)

1. All tests passing: `go test ./internal -v -race`
2. Code review
3. Merge to staging branch
4. Week 2 team starts

---

## 📞 Communication

### Daily Standups
- **When:** 10 AM daily
- **Duration:** 15 min
- **Topics:** Blockers, progress, help needed

### Weekly Sync
- **When:** Friday 3 PM
- **Duration:** 30 min
- **Topics:** Week recap, next week preview

### Issues/Blockers
- **Escalation:** Immediately to team lead
- **Resolution:** Debug together, pair programming if needed

---

## 🎯 Success Metrics

After 3 weeks, you should have:

1. **Functionality:**
   - ✅ Agent cost tracking per-agent
   - ✅ Configurable enforcement (block/warn)
   - ✅ Crew-level hard cap enforcement
   - ✅ Per-agent cost breakdown

2. **Reliability:**
   - ✅ 10 comprehensive tests (all passing)
   - ✅ No race conditions (-race flag)
   - ✅ Proper daily reset

3. **Observability:**
   - ✅ Metrics endpoint (`GET /metrics/crew-costs`)
   - ✅ Dashboard for cost visualization
   - ✅ Debug logging for decisions

4. **Documentation:**
   - ✅ Architecture guide
   - ✅ Configuration guide
   - ✅ Example configurations
   - ✅ Troubleshooting guide

5. **Production Ready:**
   - ✅ Staging deployment successful
   - ✅ 24h production monitoring passed
   - ✅ Team trained on system

---

## ⚠️ Common Issues & Solutions

### Issue: Tests Fail with Race Conditions
**Solution:** Use `-race` flag to identify, add mutex locks to affected code

### Issue: Metrics Don't Update
**Solution:** Check that mutex is unlocked after updates, verify timing

### Issue: YAML Config Not Loading
**Solution:** Verify struct tags are lowercase with `yaml:` prefix

### Issue: Cost Estimates Wrong
**Solution:** Verify token estimation: 1 token ≈ 4 characters

---

## 📝 Configuration Examples

### Agent Configuration
```yaml
# config/agents/router.yaml
agent:
  id: router
  name: Query Router
  max_tokens_per_call: 1000
  max_tokens_per_day: 50000
  max_cost_per_day: 10.00
  cost_alert_threshold: 0.80
  enforce_cost_limits: true  # Block if exceeded
```

### Crew Configuration
```yaml
# crew.yaml
crew:
  name: Multi-Agent Search
  max_cost_per_execution: 2.50
  max_cost_per_day: 100.00
  max_tokens_per_execution: 20000
```

---

## 🔗 Related Documents

- **tech-spec-agent-cost-control.md** - Detailed agent implementation
- **tech-spec-crew-cost-control.md** - Detailed crew implementation
- **FINAL_DECISION_SUMMARY.md** - Decision documentation
- **TEAM_DISCUSSION_BRIEF.md** - Team discussion context

---

## 📅 Timeline at a Glance

```
Week 1 (Agent):          ████████░░ Agent cost controls
Week 2 (Crew):           ░░████████ Crew hard cap
Week 3 (Monitoring):     ░░░░████████ Monitoring & production
                         ├─────────────────────────────────┤
                         Mon     Wed     Fri     Mon     Fri
```

---

**Status:** ✅ **Ready to Begin Implementation**

**Next Steps:**
1. Assign team members
2. Schedule Week 1 kickoff
3. Begin Monday with types.go updates

Good luck! 🚀

