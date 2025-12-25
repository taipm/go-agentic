# 🚀 QUICK FIX GUIDE - Quiz Exam Data Issues

**Document**: See [5W2H_ANALYSIS.md](5W2H_ANALYSIS.md) for full details

---

## THE PROBLEM IN 1 MINUTE

```
✅ Exam completes: 10/10 perfect score
❌ BUT all questions are blank in report
❌ All answers show as <nil>
❌ Why? Question text never reaches the RecordAnswer tool
```

---

## THE FIXES NEEDED (Priority Order)

### PHASE 1: CRITICAL (Fix in 30 mins)

#### FIX #1: Teacher Agent Prompt - Update `config/agents/teacher.yaml`

**What's wrong**: LLM generates questions but doesn't include them in RecordAnswer() call

**How to fix**: Rewrite system_prompt to explicitly guide the LLM

```yaml
# Lines 28-74: REPLACE with clearer workflow

system_prompt: |
  WORKFLOW - 4 STRICT STEPS:

  STEP 1: Check remaining
  → Call QuizExam(action="get_status")

  STEP 2: Ask question (if remaining > 0)
  → Generate: "Nội dung câu hỏi"
  → Emit: [QUESTION]
  → Wait: [ANSWER] signal

  STEP 3: Record answer
  → Call QuizExam(action="record_answer",
      question="[EXACT QUESTION TEXT]",
      student_answer="[STUDENT RESPONSE]",
      is_correct=true/false)

  STEP 4: Continue or end
  → If remaining > 0: Loop to STEP 1
  → If remaining = 0: Call QuizReport(...), emit [END_EXAM]
```

**Impact**: ✅ LLM now explicitly extracts question text into tool call

---

#### FIX #2: Add Validation - Update `internal/tools.go` (RecordAnswer, Line 347-410)

**What's wrong**: Tool accepts empty questions without error

**How to fix**: Add validation that REJECTS empty data

```go
// Add these checks at start of RecordAnswer Func:

// ✅ Validate question (CRITICAL)
question, ok := args["question"].(string)
if !ok || strings.TrimSpace(question) == "" {
    return json.Marshal(map[string]interface{}{
        "error": "Validation failed: question cannot be empty",
        "hint": "Extract the exact question text",
    })
}

// ✅ Validate student_answer (CRITICAL)
studentAnswer, ok := args["student_answer"].(string)
if !ok || strings.TrimSpace(studentAnswer) == "" {
    return json.Marshal(map[string]interface{}{
        "error": "Validation failed: student_answer cannot be empty",
        "hint": "Extract the student's actual response",
    })
}

// ✅ Validate is_correct explicitly (CRITICAL)
isCorrect, exists := args["is_correct"].(bool)
if !exists {
    return json.Marshal(map[string]interface{}{
        "error": "Validation failed: is_correct must be explicitly true/false",
    })
}
```

**Impact**: ❌ LLM gets error feedback and must fix the data

---

#### FIX #3: Remove Auto-Save - Update `internal/tools.go` (Line 424-428)

**What's wrong**: Report written after EACH question = 10 incomplete reports

**How to fix**: DELETE this entire block:

```go
// ❌ REMOVE THIS:
if err := state.WriteReportToFile(""); err != nil {
    fmt.Printf("  [Auto-save] Lỗi lưu biên bản: %v\n", err)
}
```

**Why**: Report will be written ONCE at end (cleaner, faster, single authoritative file)

**Impact**: ✅ Single clean report with complete data

---

### PHASE 2: CODE SIMPLIFICATION (Optional, 30 mins)

#### FIX #4: Consolidate Tools (Reduce 5 tools → 2 tools)

**Current**: 5 separate tools = 200+ lines of repetitive code
```
GetQuizStatus
RecordAnswer
GetFinalResult
WriteExamReport
SetExamInfo
```

**Simplified**:
```
QuizExam(action="get_status"|"record_answer"|"get_result")
QuizReport(teacher_final_comment)
```

**Benefits**:
- 50% less code
- Single entry point per function
- Easier to validate
- Clearer data flow

**See**: [5W2H_ANALYSIS.md - Phase 2](5W2H_ANALYSIS.md#32-phase-2-code-simplification-30-mins---example) for full code

---

### PHASE 3: CORE ENHANCEMENT (Optional, 30 mins)

#### FIX #5: Add Validation Framework to CORE

**Adds**: `core/tools/validator.go` with reusable validation

**Benefits**: All future examples can use built-in validation

**See**: [5W2H_ANALYSIS.md - Phase 3](5W2H_ANALYSIS.md#33-phase-3-core-enhancement-30-mins---optional) for details

---

## WHICH LAYER TO FIX?

| Issue | Layer | Reason |
|-------|-------|--------|
| Missing question data | **EXAMPLE** (agent prompt) | LLM instruction issue |
| Empty fields accepted | **EXAMPLE** (tool design) | Specific to quiz example |
| Auto-save redundancy | **EXAMPLE** (workflow) | Quiz-specific pattern |
| Code too complex | **EXAMPLE** (refactoring) | Design choice |
| No built-in validation | CORE (optional) | Would benefit all examples |

**Conclusion**: 80% fixes are EXAMPLE-only, 20% optional CORE enhancement

---

## TESTING THE FIX

After applying all 3 critical fixes, run:

```bash
cd examples/01-quiz-exam
go run ./cmd/main.go
```

Expected output:
```
[TOOL] RecordAnswer(question=1, is_correct=true)
  Câu hỏi: Số nào là 2+2?                    ← ✅ NOW SHOWS!
  Trả lời: Số 4                              ← ✅ NOW SHOWS!
  Kết quả: ĐÚNG (+1 điểm)
  Tổng điểm: 1
  Còn lại: 9 câu
```

Report file (`reports/exam_*.md`):
```markdown
### Câu 1 ✅
**Câu hỏi:** Số nào là 2+2?                 ← ✅ NOT EMPTY!
**Trả lời:** Số 4                           ← ✅ NOT NIL!
**Kết quả:** Đúng (+1 điểm)
```

---

## EFFORT ESTIMATE

| Fix | Time | Priority |
|-----|------|----------|
| Fix #1: Agent Prompt | 20 min | ⭐⭐⭐ CRITICAL |
| Fix #2: Validation | 30 min | ⭐⭐⭐ CRITICAL |
| Fix #3: Remove Auto-Save | 10 min | ⭐⭐⭐ CRITICAL |
| Fix #4: Tool Consolidation | 90 min | ⭐⭐ NICE-TO-HAVE |
| Fix #5: CORE Validation | 120 min | ⭐ OPTIONAL |
| **TOTAL FOR CRITICAL** | **60 min** | **All working** |

---

## NEXT STEPS

1. ✅ Read [5W2H_ANALYSIS.md](5W2H_ANALYSIS.md) for full context
2. ✅ Apply Fixes #1-3 (60 minutes)
3. ✅ Run tests to verify
4. (Optional) Apply Fixes #4-5 for code simplification

---

## FILES TO MODIFY

### CRITICAL CHANGES
- `config/agents/teacher.yaml` - Rewrite system_prompt (lines 28-74)
- `internal/tools.go` - Add validation (lines 347-410), remove auto-save (lines 424-428)

### OPTIONAL CHANGES
- `internal/tools.go` - Consolidate tools (refactor entire CreateQuizTools function)
- `core/tools/validator.go` - New file for validation framework
- `config/agents/student.yaml` - Update prompts for new tool names

---

## KEY INSIGHT FROM 5W2H ANALYSIS

**WHY are questions empty?**
→ LLM generates question text, but doesn't include it in tool parameters

**WHERE does it break?**
→ Between agent output and tool input (no data forwarding)

**WHO fails?**
→ Both LLM (confusion) and tool (no validation)

**HOW to fix?**
→ (1) Clearer agent prompt, (2) Validation errors, (3) Remove redundant saves

**WHEN will it work?**
→ After LLM gets validation errors, it learns to include question text

**HOW MUCH effort?**
→ 60 minutes for critical fixes, 180 minutes for complete refactoring

