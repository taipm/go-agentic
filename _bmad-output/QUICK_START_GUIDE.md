# ⚡ Quick Start Guide: Cost Control Implementation

**Duration:** 3 weeks | **Effort:** 24 hours | **Team:** 2-3 developers

---

## 📚 Read These (in order)

1. **This file** (5 min) - Overview
2. **FINAL_DECISION_SUMMARY.md** (10 min) - Decisions made
3. **tech-spec-agent-cost-control.md** (20 min) - Agent implementation
4. **tech-spec-crew-cost-control.md** (20 min) - Crew implementation
5. **IMPLEMENTATION_ROADMAP.md** (30 min) - Week-by-week plan

---

## 🎯 The 2 Decisions

### Decision #1: Agent Cost Blocking
**CHOSEN: CONFIGURABLE** ✅

- Each agent picks: `enforce_cost_limits: true` (block) or `false` (warn)
- Safe by default (block)
- Flexible when needed (warn)

**Example:**
```yaml
agents:
  - id: router
    enforce_cost_limits: true   # Strict
  - id: searcher  
    enforce_cost_limits: false  # Flexible
```

### Decision #2: Budget Hierarchy
**CHOSEN: CREW HARD CAP** ✅

- Crew limits = absolute maximum (enforced)
- Agent limits = informational only (warnings)
- Crew always wins

**Example:**
```yaml
crew:
  max_cost_per_execution: 2.50  # Hard cap, never exceeded
  max_cost_per_day: 100.00      # Hard cap, never exceeded

agents:
  - id: router
    max_cost_per_day: 10.00     # Just tracking, not enforced
```

---

## 📋 What Gets Changed

### Week 1: Agent-Level Controls

**3 Files Modified:**
- `internal/types.go` - Add 5 fields to Agent struct
- `internal/agent.go` - Add cost checking logic
- `internal/config.go` - Load cost config from YAML

**5 Fields Added to Agent:**
```go
MaxTokensPerCall   int     // Per-call limit
MaxTokensPerDay    int     // Daily limit
MaxCostPerDay      float64 // Daily budget
CostAlertThreshold float64 // Warning threshold
EnforceCostLimits  bool    // Block or warn
```

**Code Lines:** ~60 lines  
**Tests:** 5 new tests  
**Effort:** 7 hours  

### Week 2: Crew-Level Controls

**3 Files Modified:**
- `internal/types.go` - Add 3 fields to Crew struct
- `internal/crew.go` - Add crew budget checking
- `internal/config.go` - Load crew config from YAML

**3 Fields Added to Crew:**
```go
MaxCostPerExecution float64 // Per-execution cap
MaxCostPerDay       float64 // Daily cap
MaxTokensPerExecution int   // Token cap
```

**Code Lines:** ~70 lines  
**Tests:** 5 new tests  
**Effort:** 7 hours  

### Week 3: Monitoring & Production

**3 New Files:**
- Metrics endpoint (`GET /metrics/crew-costs`)
- HTML dashboard
- Documentation files

**Effort:** 10 hours  

---

## 🚀 Week-by-Week Overview

### Week 1: Monday-Friday (Agent Controls)

| Day | Task | Time |
|-----|------|------|
| Mon | Update types.go | 30m |
| Tue | Add token estimation | 30m |
| Wed | Implement cost checks | 1.5h |
| Thu | Integrate with Execute() | 1.5h |
| Fri | Write & run tests | 2h |

**Outcome:** Agent cost tracking working ✅

### Week 2: Monday-Friday (Crew Controls)

| Day | Task | Time |
|-----|------|------|
| Mon | Update crew types | 30m |
| Tue | Cost estimation | 1h |
| Wed | Budget checker | 1.5h |
| Thu | Integrate with Execute() | 1.5h |
| Fri | Write & run tests | 2h |

**Outcome:** Crew hard cap enforced ✅

### Week 3: Monday-Friday (Monitoring)

| Day | Task | Time |
|-----|------|------|
| Mon | Metrics endpoint | 2h |
| Tue | Dashboard HTML | 2h |
| Wed | Documentation | 2h |
| Thu | Load testing | 2h |
| Fri | Production deployment | 2h |

**Outcome:** System in production ✅

---

## ✅ Success Criteria

### Week 1
- [ ] Agent struct compiles
- [ ] 5 tests passing
- [ ] Cost tracking works
- [ ] Block & warn modes work

### Week 2
- [ ] Crew struct compiles
- [ ] 5 tests passing
- [ ] Hard cap enforced
- [ ] Integration with agents works

### Week 3
- [ ] Metrics endpoint responding
- [ ] Dashboard loading
- [ ] No race conditions
- [ ] Production deployment successful

---

## 📖 Key Files to Know

### Tech Specs (Read First)
- `_bmad-output/tech-spec-agent-cost-control.md` - Agent details
- `_bmad-output/tech-spec-crew-cost-control.md` - Crew details

### Implementation
- `internal/types.go` - Type definitions
- `internal/agent.go` - Agent logic
- `internal/crew.go` - Crew logic
- `internal/config.go` - Configuration loading

### Testing
- `internal/agent_test.go` - Agent tests
- `internal/crew_test.go` - Crew tests

### Documentation
- `_bmad-output/IMPLEMENTATION_ROADMAP.md` - Detailed plan
- `_bmad-output/FINAL_DECISION_SUMMARY.md` - Decisions

---

## 🎯 Quick Configuration

### Agent Configuration (agents/router.yaml)
```yaml
agent:
  id: router
  name: "Query Router"
  
  # New fields
  max_tokens_per_call: 1000
  max_tokens_per_day: 50000
  max_cost_per_day: 10.00
  cost_alert_threshold: 0.80
  enforce_cost_limits: true
```

### Crew Configuration (crew.yaml)
```yaml
crew:
  name: "Multi-Agent Search"
  
  # New fields
  max_cost_per_execution: 2.50
  max_cost_per_day: 100.00
  max_tokens_per_execution: 20000
```

---

## 🔧 Testing Commands

### Run All Tests
```bash
go test ./internal -v
```

### Run with Race Detection
```bash
go test ./internal -v -race
```

### Run Specific Test
```bash
go test ./internal -v -run TestAgent_BlockMode_PerCallTokenLimit
```

### Check Coverage
```bash
go test ./internal -cover
```

---

## 🐛 Debugging Tips

### Check Metrics
```bash
curl http://localhost:8080/metrics/crew-costs
```

### Enable Debug Logging
Look for log lines:
- `[BLOCK]` - Cost limit exceeded
- `[WARN]` - Approaching limit
- `[RESET]` - Daily reset happened

### Common Issues

**Tests Failing with Race Conditions?**
- Add `sync.RWMutex` to protect shared data
- Use `mu.Lock()` / `mu.Unlock()`
- Run with `-race` flag to find issues

**Cost Estimates Wrong?**
- Check token calculation: 1 token = 4 characters
- Verify cost per token constant

**Config Not Loading?**
- Check YAML struct tags: `yaml:"field_name"`
- Use lowercase field names
- Verify defaults are set

---

## 📞 Getting Help

### Tech Spec Questions
- Read the relevant tech-spec thoroughly
- Check implementation examples in tech-spec

### Code Questions
- Look at similar existing code patterns
- Check test cases for examples

### Design Questions
- Review FINAL_DECISION_SUMMARY.md
- Review TEAM_DISCUSSION_BRIEF.md

---

## 🚀 Starting Monday

### Morning Prep (30 min)
1. Clone fresh copy: `git checkout -b feat/cost-controls`
2. Read `tech-spec-agent-cost-control.md`
3. Setup your IDE with `internal/` folder

### Day 1 Task (30 min)
1. Open `internal/types.go`
2. Add 5 Agent cost fields (see tech-spec)
3. Run `go build ./...` to verify

### End of Week Check
1. Run `go test ./internal -v -race`
2. All 5 agent tests should pass
3. No race condition warnings

---

## 📊 Team Roles

### Agent Team Lead (Week 1)
- Update types.go
- Implement cost checks
- Write agent tests
- Code review

### Crew Team Lead (Week 2)
- Update crew types
- Implement budget checks
- Write crew tests
- Integration testing

### DevOps/QA Lead (Week 3)
- Setup metrics endpoint
- Create dashboard
- Load testing
- Production deployment

---

## 🎓 Learning Resources

Inside the code:
- Check existing token estimation patterns
- Review Go concurrency patterns (sync.RWMutex)
- Check YAML unmarshaling examples

In the tech-specs:
- Full code examples for each function
- Test case templates
- Configuration examples

---

## ⏱️ Time Investment

**Per Developer:**
- Week 1: 7 hours (agent-level, 1 dev)
- Week 2: 7 hours (crew-level, 1 dev)
- Week 3: 5 hours (monitoring, 0.5 dev)

**Total:** ~24 hours of engineering effort
**Team Size:** 2-3 developers
**Duration:** 3 calendar weeks

---

## ✨ Final Checklist

Before implementation starts:

- [ ] Read this Quick Start
- [ ] Read both tech-specs
- [ ] Understand the 2 decisions
- [ ] Know the 3 files to modify (Week 1)
- [ ] Know the testing strategy
- [ ] Have go-agentic project open
- [ ] Have created feature branch

**Ready to start?** Begin with Monday tasks in IMPLEMENTATION_ROADMAP.md

---

**Questions?** Check the relevant tech-spec document first.

Good luck! 🚀

