# CORE vs EXAMPLE - Fix Decision Matrix

**Question**: Should we fix this in CORE or EXAMPLE?

**Answer**: Primarily EXAMPLE (80%), optionally CORE (20%)

---

## Executive Summary

| Component | Issue | Fix Location | Scope | Breaking Change |
|-----------|-------|--------------|-------|-----------------|
| LLM Prompting | Question data not in tool call | EXAMPLE | Local | No |
| Tool Validation | No input validation | EXAMPLE | Local | No |
| Workflow | Auto-save writes incomplete data | EXAMPLE | Local | No |
| Tool Design | 5 tools too complex | EXAMPLE | Local | No |
| **CORE Framework** | No validation framework | **CORE (Optional)** | **Framework** | **No** |

**Recommendation**: Fix EXAMPLE first (60 min), CORE later if needed (optional)

---

## 1. LLM PROMPTING - EXAMPLE ONLY ✅

### The Issue
Teacher agent generates questions, but doesn't include them in RecordAnswer tool call

### Why It's EXAMPLE
- **Root cause**: Specific to how teacher.yaml instructs the LLM
- **Not a framework problem**: Core executor works fine
- **Solution**: Rewrite teacher.yaml system_prompt
- **Impact**: Only affects this example

### EXAMPLE Fix
```yaml
# teacher.yaml system_prompt (rewrite lines 28-74)
```

### CORE Does NOT Need to Change
- ✅ Core executor handles signals correctly
- ✅ Core tool extraction works fine
- ✅ Core has no opinion on prompt quality

---

## 2. TOOL VALIDATION - EXAMPLE ONLY ✅

### The Issue
RecordAnswer accepts empty question="" without error

### Why It's EXAMPLE
- **Root cause**: Quiz-specific tool design choice
- **Not a framework problem**: Core tools work, just missing validation
- **Solution**: Add validation checks in RecordAnswer function
- **Impact**: Only affects this example's tools

### EXAMPLE Fix
```go
// internal/tools.go - RecordAnswer function
if strings.TrimSpace(question) == "" {
    return error: "question cannot be empty"
}
```

### CORE Could Enhance (Optional)
Core COULD provide validation framework, but it's not REQUIRED

```go
// OPTIONAL: core/tools/validator.go (new)
type StringValidator struct {
    Required bool
    MinLen   int
}
```

---

## 3. AUTO-SAVE REDUNDANCY - EXAMPLE ONLY ✅

### The Issue
Report written after each of 10 questions = 10 incomplete files

### Why It's EXAMPLE
- **Root cause**: Quiz-specific workflow design
- **Not a framework problem**: Core has no auto-save mechanism
- **Solution**: Remove WriteReportToFile call from RecordAnswer
- **Impact**: Only affects this example's reporting

### EXAMPLE Fix
```go
// internal/tools.go - DELETE lines 424-428
// if err := state.WriteReportToFile(""); err != nil { ... }
```

### CORE Does NOT Need to Change
- ✅ Core has no opinion on when to save
- ✅ Core doesn't auto-save anything
- ✅ Tool results returned correctly

---

## 4. TOOL COMPLEXITY - EXAMPLE ONLY ✅

### The Issue
5 separate tools (GetQuizStatus, RecordAnswer, etc.) = 200+ boilerplate lines

### Why It's EXAMPLE
- **Root cause**: Design choice to create 5 separate tools
- **Not a framework problem**: Core allows any number of tools
- **Solution**: Consolidate into 2 unified tools
- **Impact**: Only affects this example's tool design

### EXAMPLE Fix
```go
// Consolidate:
// GetQuizStatus + RecordAnswer + GetFinalResult → QuizExam(action=...)
// SetExamInfo + WriteExamReport → Handled by agents
```

### CORE Does NOT Need to Change
- ✅ Core supports both 5 tools and 2 tools equally
- ✅ Core has no design guidance on tool granularity
- ✅ This is an example-level optimization

---

## 5. VALIDATION FRAMEWORK - CORE (Optional) ⭐

### The Issue
Core tools don't have built-in validation (every tool must implement its own)

### Why It Could Be CORE
- **Benefit**: All future examples can reuse validation framework
- **Not breaking**: Validation is opt-in, backward compatible
- **Scope**: Framework enhancement, not fix
- **Priority**: Low (examples can validate themselves)

### CORE Enhancement (If Implemented)
```go
// NEW: core/tools/validator.go
type ParamValidator interface {
    Validate(value interface{}) error
}

type StringValidator struct {
    Required bool
    MinLen   int
}

// Tool definition includes validators
type Tool struct {
    Name       string
    Validators map[string]ParamValidator
    Func       interface{}
}
```

### Example Integration
```go
// In quiz-exam example:
tools["RecordAnswer"] = &Tool{
    Validators: map[string]ParamValidator{
        "question":       StringValidator{Required: true, MinLen: 5},
        "student_answer": StringValidator{Required: true, MinLen: 1},
    }
}
```

### Decision: Implement in CORE?
- ✅ **YES if**: Planning to make validation a standard pattern
- ❌ **NO if**: Examples can just add their own validation

**Recommendation**: Do EXAMPLE validation first, add CORE framework later

---

## FIX STRATEGY COMPARISON

### STRATEGY A: EXAMPLE ONLY (Recommended for now)

```
Timeline:
Week 1:
  ✅ Fix teacher.yaml (20 min)
  ✅ Add validation to tools.go (30 min)
  ✅ Remove auto-save (10 min)
  ✅ Test (30 min)
  TOTAL: 60 min, all working

Week 2:
  (Optional) Consolidate tools (90 min)
  (Optional) Refactor codebase
```

**Pros**:
- Fast (60 min to working state)
- Example stays independent
- No breaking changes
- Examples is testable in isolation

**Cons**:
- Each example must implement validation
- Code duplication across examples
- No reusable patterns

---

### STRATEGY B: EXAMPLE + CORE (Complete Solution)

```
Timeline:
Week 1:
  ✅ Fix teacher.yaml (20 min)
  ✅ Add validation to tools.go (30 min)
  ✅ Remove auto-save (10 min)
  ✅ Test (30 min)
  SUBTOTAL: 60 min

Week 2:
  ✅ Implement core/tools/validator.go (60 min)
  ✅ Update Tool type definition (15 min)
  ✅ Update tool extraction logic (30 min)
  ✅ Add tests (45 min)
  ✅ Update quiz-exam to use validators (20 min)
  SUBTOTAL: 170 min

TOTAL: 230 min for complete refactoring
```

**Pros**:
- Framework improvement benefits all future examples
- Validation becomes standard pattern
- Less code duplication
- Professional, production-ready

**Cons**:
- Longer timeline (170 min additional)
- More testing required
- Requires CORE changes (broader scope)

---

## DECISION MATRIX

### Should We Fix in CORE?

```
Question                          Answer    Evidence
─────────────────────────────────────────────────────────────
Is this a CORE bug?              NO         Core executor works fine
Does CORE lack critical feature? NO         Tools work, just no validation
Would CORE benefit?              MAYBE      Validation framework useful
Is it required to fix example?   NO         Example can validate itself
Can CORE stay backward compat?   YES        Validators optional
Will other examples need it?     MAYBE      Depends on future examples
```

---

## FINAL RECOMMENDATION

### IMMEDIATE ACTION (Week 1)
**Fix EXAMPLE only** - Apply Fixes #1-3 in 60 minutes

✅ Teacher.yaml system_prompt rewrite
✅ Add validation to RecordAnswer tool
✅ Remove auto-save from workflow

**Why?**
- Fixes the immediate problem
- All changes stay in `/examples/01-quiz-exam/`
- No CORE changes needed
- Minimal risk of side effects

---

### OPTIONAL ACTION (Week 2+)
**Enhance CORE** if validation framework becomes needed

```
Condition: IF (other examples need validation)
    THEN: Implement core/tools/validator.go
    ELSE: Leave CORE unchanged
```

**Why?**
- Wait until there's proven need
- Avoid premature optimization
- YAGNI principle (You Aren't Gonna Need It)
- Can always add later without breaking changes

---

## FILES AFFECTED BY DECISION

### If We Fix EXAMPLE ONLY (Recommended)
```
Modified:
  examples/01-quiz-exam/config/agents/teacher.yaml
  examples/01-quiz-exam/internal/tools.go

New:
  examples/01-quiz-exam/5W2H_ANALYSIS.md
  examples/01-quiz-exam/QUICK_FIX_GUIDE.md

Unchanged:
  core/** (no changes)
```

### If We Also Enhance CORE (Optional)
```
Modified:
  examples/01-quiz-exam/config/agents/teacher.yaml
  examples/01-quiz-exam/internal/tools.go
  core/common/types.go (add Validators field)
  core/tools/extraction.go (use validators)

New:
  core/tools/validator.go
  core/tools/validator_test.go

Unchanged:
  core/crew.go
  core/executor/** (no changes to execution logic)
```

---

## SUMMARY TABLE

| Phase | What | Where | Time | Breaking? | Required? |
|-------|------|-------|------|-----------|-----------|
| **1** | Fix LLM prompt | EXAMPLE | 20m | No | ✅ YES |
| **1** | Add validation | EXAMPLE | 30m | No | ✅ YES |
| **1** | Remove auto-save | EXAMPLE | 10m | No | ✅ YES |
| **2** | Consolidate tools | EXAMPLE | 90m | No | ⭐ Nice |
| **3** | Validation framework | CORE | 170m | No | ❌ No |

**Recommendation**: Do Phase 1 (60 min), Phase 2 optional, Phase 3 only if needed

