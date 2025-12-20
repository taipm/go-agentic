---
title: "Epic 1 Story Map & Visual Overview"
date: "2025-12-20"
purpose: "Visual representation of Epic 1 stories, changes, and dependencies"
---

# Epic 1 Story Map & Visual Overview

## 📊 Story Dependency Flow

```
┌─────────────────────────────────────────────────────────┐
│              EPIC 1: Configuration Integrity & Trust     │
│                                                          │
│  Goal: Users can configure agents with confidence       │
│  that every setting will be honored exactly             │
└─────────────────────────────────────────────────────────┘

                           │
                           ▼

        ┌──────────────────┴──────────────────┐
        │                                      │
        ▼                                      ▼
   ┌─────────────┐                    ┌──────────────┐
   │  Story 1.1  │                    │  Story 1.2   │
   │   (1-2h)    │                    │   (1-2h)     │
   │   LOW RISK  │                    │ LOW-MED RISK │
   └─────────────┘                    └──────────────┘
        │                                      │
        │  Agent Model Config                  │  Temperature Range
        │  No dependencies                     │  No dependencies
        │  Blocks: Nothing                     │  Blocks: Nothing
        │                                      │
        └──────────────────┬──────────────────┘
                           │
                    Logical dependency
                    (implementation order)
                           │
                           ▼
                    ┌──────────────┐
                    │  Story 1.3   │
                    │   (2-3h)     │
                    │   LOW RISK   │
                    └──────────────┘
                         │
                Story 1.3 Validation
              Depends on 1.1 & 1.2
                (logical only)
```

---

## 📝 Story Details at a Glance

### Story 1.1: Agent Model Config ⚡

```
┌─ PROBLEM ─────────────────────────────────────────┐
│ Line 24 of agent.go:                              │
│   Model: "gpt-4o-mini"  ← HARDCODED               │
│                                                   │
│ Result: Ignores agent.Model field from config    │
└───────────────────────────────────────────────────┘

┌─ FIX ─────────────────────────────────────────────┐
│ Replace line 24:                                  │
│   FROM: Model: "gpt-4o-mini"                     │
│   TO:   Model: agent.Model                       │
│                                                   │
│ Lines changed: 1 (just line 24)                  │
└───────────────────────────────────────────────────┘

┌─ IMPACT ──────────────────────────────────────────┐
│ ✅ Agent1 uses "gpt-4o"                          │
│ ✅ Agent2 uses "gpt-4o-mini"                     │
│ ✅ Per-agent model selection works               │
│ ✅ Backward compatible                            │
└───────────────────────────────────────────────────┘

┌─ TESTS (4 total) ─────────────────────────────────┐
│ ✓ Different agents use different models           │
│ ✓ API calls use correct model value              │
│ ✓ Logs show which model each agent uses          │
│ ✓ IT Support example still works                 │
└───────────────────────────────────────────────────┘
```

### Story 1.2: Temperature Range Config 🌡️

```
┌─ PROBLEM ─────────────────────────────────────────┐
│ Lines 58-60 of config.go:                         │
│   if config.Temperature == 0 {                   │
│       config.Temperature = 0.7  ← OVERRIDE        │
│   }                                               │
│                                                   │
│ Result: Temperature 0.0 forced to 0.7            │
│ Cannot use deterministic responses               │
└───────────────────────────────────────────────────┘

┌─ ROOT CAUSE ──────────────────────────────────────┐
│ Code treats 0 as "not set" and applies default   │
│ But 0.0 is valid OpenAI value (0.0-2.0 range)   │
│                                                   │
│ Need to distinguish:                             │
│ - "not provided" → default to 0.7                │
│ - "provided as 0.0" → keep 0.0                   │
└───────────────────────────────────────────────────┘

┌─ FIX (Pointer Type) ──────────────────────────────┐
│ Change 1: types.go                               │
│   FROM: Temperature float64                      │
│   TO:   Temperature *float64                     │
│                                                   │
│ Change 2: config.go - Check nil, not 0           │
│   FROM: if config.Temperature == 0               │
│   TO:   if config.Temperature == nil             │
│                                                   │
│ Change 3: config.go - Dereference in struct     │
│   FROM: Temperature: config.Temperature           │
│   TO:   Temperature: *config.Temperature         │
│                                                   │
│ Files changed: 2 (types.go, config.go)          │
│ Lines changed: ~5 lines total                     │
└───────────────────────────────────────────────────┘

┌─ IMPACT ──────────────────────────────────────────┐
│ ✅ Temperature 0.0 → used as 0.0 (deterministic)│
│ ✅ Temperature 1.0 → used as 1.0                 │
│ ✅ Temperature 2.0 → used as 2.0                 │
│ ✅ Missing temp → defaults to 0.7               │
│ ✅ Backward compatible                            │
└───────────────────────────────────────────────────┘

┌─ TESTS (5 total) ─────────────────────────────────┐
│ ✓ 0.0 temperature respected (not overridden)    │
│ ✓ 1.0 temperature respected                      │
│ ✓ 2.0 temperature respected                      │
│ ✓ Missing temperature defaults to 0.7           │
│ ✓ Agent API calls use correct temperature       │
└───────────────────────────────────────────────────┘
```

### Story 1.3: Configuration Validation ✓

```
┌─ PROBLEM ─────────────────────────────────────────┐
│ No validation after config loading                │
│ - Invalid configs accepted silently               │
│ - Errors only appear in OpenAI API calls         │
│ - Hard to debug, confusing error messages        │
└───────────────────────────────────────────────────┘

┌─ SOLUTION ────────────────────────────────────────┐
│ 1. Add ValidateAgentConfig() function             │
│    - Check: Model is not empty                    │
│    - Check: Temperature is 0.0-2.0                │
│    - Return: Clear error message with fix hint    │
│                                                   │
│ 2. Call validation in LoadAgentConfig()          │
│    - Validate after unmarshaling YAML             │
│    - Return error if invalid                      │
│    - Never load invalid config                    │
│                                                   │
│ 3. Add logging in ExecuteAgent()                 │
│    - Show which model each agent uses            │
│    - Show which temperature each agent uses      │
└───────────────────────────────────────────────────┘

┌─ ERROR MESSAGES ──────────────────────────────────┐
│ Empty Model:                                      │
│   "Model must be specified (examples:            │
│    gpt-4o, gpt-4o-mini)"                         │
│                                                   │
│ Invalid Temperature > 2.0:                       │
│   "Temperature must be between 0.0 and 2.0,     │
│    got 2.5"                                      │
│                                                   │
│ Invalid Temperature < 0.0:                       │
│   "Temperature must be between 0.0 and 2.0,     │
│    got -1.0"                                     │
└───────────────────────────────────────────────────┘

┌─ FILES CHANGED ───────────────────────────────────┐
│ config.go: Add ValidateAgentConfig() function     │
│ config.go: Update LoadAgentConfig() to validate   │
│ agent.go:  Add logging in ExecuteAgent()          │
│                                                   │
│ Files changed: 2 (config.go, agent.go)          │
│ Lines changed: ~20 lines total                    │
└───────────────────────────────────────────────────┘

┌─ TESTS (8 total) ─────────────────────────────────┐
│ ✓ Empty model returns clear error                │
│ ✓ Temperature 2.1 returns clear error            │
│ ✓ Temperature -1.0 returns clear error           │
│ ✓ Valid config passes validation                 │
│ ✓ Boundary values (0.0, 2.0) pass                │
│ ✓ LoadAgentConfig validates configs              │
│ ✓ Valid file loads successfully                  │
│ ✓ v0.0.1 configs still work (backward compat)   │
└───────────────────────────────────────────────────┘
```

---

## 🗂️ File Change Summary

```
┌─ Story 1.1 Changes ───────────────────────────────┐
│ File: go-agentic/agent.go                         │
│ Line 24:  "gpt-4o-mini" → agent.Model             │
│ Changes:  1 line                                  │
│ Impact:   Single-line fix                         │
└───────────────────────────────────────────────────┘

┌─ Story 1.2 Changes ───────────────────────────────┐
│ File: go-agentic/types.go                         │
│ Change: AgentConfig.Temperature float64 → *float64│
│ Impact:  Allows nil distinction                   │
│                                                   │
│ File: go-agentic/config.go                        │
│ Line 58: if Temperature == 0 → if Temperature == nil
│ Line 99: dereference *config.Temperature          │
│ Changes: ~5 lines                                 │
│ Impact:  Pointer type handling                    │
└───────────────────────────────────────────────────┘

┌─ Story 1.3 Changes ───────────────────────────────┐
│ File: go-agentic/config.go                        │
│ Add: ValidateAgentConfig(cfg) error function     │
│ Update: LoadAgentConfig to call validation       │
│ Changes: ~20 lines                                │
│ Impact:  Validation logic                         │
│                                                   │
│ File: go-agentic/agent.go                         │
│ Add: Logging in ExecuteAgent()                    │
│ Changes: ~3 lines                                 │
│ Impact:  Execution logging                        │
└───────────────────────────────────────────────────┘

TOTAL CHANGES: ~4 files, ~30 lines of code
```

---

## ⏱️ Implementation Timeline

```
Day 1 - Morning (1-2 hours)
┌──────────────────────────────┐
│ Story 1.1: Model Config       │  Branch: feat/epic-1-story-1-1-*
│ ✓ Edit agent.go:24           │
│ ✓ Add 4 tests                 │  Tests: 4
│ ✓ make test (pass)            │  Time:  1-2h
│ ✓ make lint (pass)            │  Risk:  LOW
│ ✓ Create PR                   │
└──────────────────────────────┘
           │
           ▼ (merge PR)

Day 1 - Afternoon (1-2 hours)
┌──────────────────────────────┐
│ Story 1.2: Temperature Range  │  Branch: feat/epic-1-story-1-2-*
│ ✓ Edit types.go               │
│ ✓ Edit config.go (~5 lines)   │  Tests: 5
│ ✓ Add 5 tests                 │  Time:  1-2h
│ ✓ make test (pass)            │  Risk:  LOW-MED
│ ✓ make lint (pass)            │
│ ✓ Create PR                   │
└──────────────────────────────┘
           │
           ▼ (merge PR)

Day 2 - Morning (2-3 hours)
┌──────────────────────────────┐
│ Story 1.3: Validation         │  Branch: feat/epic-1-story-1-3-*
│ ✓ Add ValidateAgentConfig()   │
│ ✓ Update LoadAgentConfig()    │  Tests: 8
│ ✓ Add logging in agent.go     │  Time:  2-3h
│ ✓ Add 8 tests                 │  Risk:  LOW
│ ✓ make test (pass)            │
│ ✓ make lint (pass)            │
│ ✓ Create PR                   │
└──────────────────────────────┘
           │
           ▼ (merge PR)

Day 2 - Afternoon
┌──────────────────────────────┐
│ Epic 1 Complete ✅            │
│ ✓ All 3 stories merged        │
│ ✓ All tests passing           │  Total Time: 4-7h
│ ✓ Ready for next epics        │
└──────────────────────────────┘

TOTAL: 4-7 hours across 1-2 days
```

---

## 🧪 Testing Strategy

```
┌─ Story 1.1 Tests ─────────────────────────────────┐
│ Unit Tests:                                       │
│   ✓ Model field from config is used              │
│   ✓ Different agents use different models        │
│                                                   │
│ Integration Tests:                                │
│   ✓ API call includes correct model              │
│   ✓ Mock OpenAI call verifies params             │
│                                                   │
│ System Tests:                                     │
│   ✓ IT Support example runs with correct models  │
│                                                   │
│ Logging Tests:                                    │
│   ✓ Agent initialization shows correct model     │
└───────────────────────────────────────────────────┘

┌─ Story 1.2 Tests ─────────────────────────────────┐
│ Unit Tests:                                       │
│   ✓ Temperature 0.0 respected (not override)     │
│   ✓ Temperature 1.0 respected                     │
│   ✓ Temperature 2.0 respected                     │
│   ✓ Missing temperature defaults to 0.7          │
│                                                   │
│ Integration Tests:                                │
│   ✓ API call includes correct temperature       │
│                                                   │
│ Backward Compatibility Tests:                     │
│   ✓ v0.0.1 configs without temperature work     │
└───────────────────────────────────────────────────┘

┌─ Story 1.3 Tests ─────────────────────────────────┐
│ Unit Tests:                                       │
│   ✓ ValidateAgentConfig rejects empty model      │
│   ✓ ValidateAgentConfig rejects temp > 2.0       │
│   ✓ ValidateAgentConfig rejects temp < 0.0       │
│   ✓ ValidateAgentConfig accepts valid configs    │
│   ✓ ValidateAgentConfig accepts boundaries       │
│                                                   │
│ Integration Tests:                                │
│   ✓ LoadAgentConfig validates after loading      │
│   ✓ LoadAgentConfig returns error for invalid    │
│                                                   │
│ File I/O Tests:                                   │
│   ✓ Valid YAML file loads and validates          │
│   ✓ Invalid YAML file rejected with error        │
│                                                   │
│ Backward Compatibility Tests:                     │
│   ✓ v0.0.1 config files still load and validate │
└───────────────────────────────────────────────────┘

TOTAL TESTS: 17 tests
COVERAGE TARGET: >90%
```

---

## 📋 Implementation Checklist

```
BEFORE IMPLEMENTATION
├─ [ ] Review epic-1-detailed-stories.md
├─ [ ] Review epic-1-review-checklist.md
├─ [ ] Team discussion: pointer type approach?
├─ [ ] Team approval: all 3 stories ready?
├─ [ ] Environment: Go 1.25.5 available?
├─ [ ] Environment: OpenAI API key configured?
├─ [ ] Environment: make test works?
└─ [ ] Branch: ready to create branches?

STORY 1.1 IMPLEMENTATION
├─ [ ] Create branch: feat/epic-1-story-1-1-*
├─ [ ] Edit agent.go line 24
├─ [ ] Add 4 tests
├─ [ ] Run: make test (all pass?)
├─ [ ] Run: make lint (pass?)
├─ [ ] Run: make coverage (>90%?)
├─ [ ] Push to remote
├─ [ ] Create PR with story details
└─ [ ] Wait for review + merge

STORY 1.2 IMPLEMENTATION
├─ [ ] Create branch: feat/epic-1-story-1-2-*
├─ [ ] Edit types.go (Temperature type)
├─ [ ] Edit config.go (~5 lines)
├─ [ ] Add 5 tests
├─ [ ] Run: make test (all pass?)
├─ [ ] Run: make lint (pass?)
├─ [ ] Run: make coverage (>90%?)
├─ [ ] Push to remote
├─ [ ] Create PR with story details
└─ [ ] Wait for review + merge

STORY 1.3 IMPLEMENTATION
├─ [ ] Create branch: feat/epic-1-story-1-3-*
├─ [ ] Add ValidateAgentConfig() to config.go
├─ [ ] Update LoadAgentConfig() to call validate
├─ [ ] Add logging to ExecuteAgent() in agent.go
├─ [ ] Add 8 tests
├─ [ ] Run: make test (all pass?)
├─ [ ] Run: make lint (pass?)
├─ [ ] Run: make coverage (>90%?)
├─ [ ] Push to remote
├─ [ ] Create PR with story details
└─ [ ] Wait for review + merge

EPIC 1 COMPLETION
├─ [ ] All 3 story PRs merged to main
├─ [ ] Full test suite passes: make test
├─ [ ] Overall coverage >90%: make coverage
├─ [ ] No linting issues: make lint
├─ [ ] IT Support example runs: go run ./examples/it-support
├─ [ ] Ready for Epic 5 (Testing Framework)
└─ [ ] ✅ EPIC 1 COMPLETE
```

---

## 🎯 Success Criteria

```
Story 1.1 SUCCESS
├─ Agent.Model is used (not "gpt-4o-mini")
├─ Different agents use different models
├─ API calls include correct model
├─ Logs show model per agent
├─ Backward compatible
└─ All 4 tests passing

Story 1.2 SUCCESS
├─ Temperature 0.0 works (not overridden)
├─ All values 0.0-2.0 respected
├─ Default 0.7 for missing temp
├─ API calls use correct temperature
├─ Backward compatible
└─ All 5 tests passing

Story 1.3 SUCCESS
├─ Empty model error is clear
├─ Invalid temp error is clear
├─ Valid config passes
├─ Initialization logged
├─ Backward compatible
└─ All 8 tests passing

EPIC 1 SUCCESS
├─ All 3 stories merged ✅
├─ All 17 tests passing ✅
├─ Coverage >90% ✅
├─ Linting clean ✅
├─ Backward compatible ✅
└─ Ready for next epics ✅
```

---

## 🔄 Story Dependencies & Sequencing

```
LOGICAL DEPENDENCIES (order matters):

1.1 (Model Config)
  ↓
  └─→ Must do first (simplest, isolated)
       Result: agent.Model is used ✅

1.2 (Temperature)
  ↓
  └─→ Can do in parallel OR after 1.1
       (no technical dependency, but logical: foundation first)
       Result: Temperature range works ✅

1.3 (Validation)
  ↓
  └─→ Do last (builds on 1.1 & 1.2)
       (logical dependency: validates what 1.1 & 1.2 create)
       Result: Config validated with clear errors ✅

RECOMMENDED SEQUENCE:
Story 1.1 → (merge) → Story 1.2 → (merge) → Story 1.3

PARALLEL POSSIBLE:
Could do 1.1 and 1.2 in parallel if teams available,
but must complete before 1.3
```

---

This story map provides a visual overview of Epic 1 structure, changes, and implementation plan.

