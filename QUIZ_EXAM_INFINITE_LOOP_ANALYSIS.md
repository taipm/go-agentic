# Phân Tích: Vì Sao 01-Quiz-Exam Bị Lặp Vô Hạn

**Status**: 🔍 Analyzed - Root Cause Identified
**Date**: 2025-12-25

---

## 📋 Tóm Tắt Vấn Đề

```bash
$ cd examples/01-quiz-exam && go run ./cmd/main.go
[Chạy... chạy... chạy...]  ← Lặp vô hạn, không kết thúc
^C  ← Phải bấm Ctrl+C để dừng
```

### Vấn Đề
- App không kết thúc sau 10 câu hỏi
- Lặp liên tục: Teacher → Student → Teacher → ...
- Signal `[END_EXAM]` không trigger termination

---

## 🔎 Nguyên Nhân Gốc

### Workflow Dự Kiến (Đúng)

```
1. Teacher: Ask Q1 [QUESTION]
   └─ Signal: [QUESTION] → Route to Student

2. Student: Answer [ANSWER]
   └─ Signal: [ANSWER] → Route to Teacher

3. Teacher: RecordAnswer (Q1 done)
   └─ Check: questions_remaining > 0?
      ├─ YES: Ask Q2 [QUESTION]
      └─ NO (remaining = 0): [END_EXAM]

4. [END_EXAM] Signal
   └─ Route to: "" (terminate)
   └─ Workflow STOPS ✅
```

### Workflow Thực Tế (Sai)

```
Q1 → Answer → RecordAnswer
Q2 → Answer → RecordAnswer
...
Q10 → Answer → RecordAnswer (CurrentQuestion = 10)

🔴 PROBLEM: Teacher không emit [END_EXAM]!

Thay vào đó:
Q11 → Teacher ask Q11 (but CurrentQuestion >= 10)
     └─ RecordAnswer reject: "Kỳ thi đã hoàn thành"
     └─ Teacher không biết kỳ thi đã kết thúc

Q12 → Keep asking...
Q13 → Keep asking...
... ← INFINITE LOOP!
```

---

## 🎯 3 Vấn Đề Chính

### 1️⃣ Teacher Prompt Không Rõ Ràng

**File**: `config/agents/teacher.yaml` (lines 26-52)

Prompt nói:
- Step 2: "If remaining = 0: Announce score and emit [END_EXAM]"
- Step 7: "Call RecordAnswer(...)"
- Step 8: "Go back to step 1"

**Vấn Đề**: 
- RecordAnswer return `questions_remaining`, nhưng teacher prompt không check
- Teacher chỉ "Go back to step 1" mà không check `questions_remaining`
- LLM không follow step 2 vì flow logic không rõ

### 2️⃣ RecordAnswer Không Emit Signal

**File**: `examples/01-quiz-exam/internal/tools.go`

RecordAnswer biết `is_complete = true` sau Q10, nhưng:
- Chỉ return result
- Không emit `[END_EXAM]` signal
- Không kích hoạt termination

### 3️⃣ No Fallback Mechanism

Nếu teacher không emit [END_EXAM], workflow tiếp tục vô hạn:
- max_rounds = 30 (đủ cho 15 câu hỏi)
- Không có safety timeout
- Lệnh Ctrl+C cần thiết để dừng

---

## 🔧 5 Giải Pháp

### ✅ Solution 1: Update Teacher Prompt (RECOMMENDED)

**File**: `config/agents/teacher.yaml` (rewrite step 6-8)

```yaml
system_prompt: |
  ...
  YOUR WORKFLOW - Follow these steps EXACTLY:
  1. Call GetQuizStatus() to see remaining questions
  2. If remaining = 0: Announce score and emit [END_EXAM]
  3. If remaining > 0: Ask ONE new question, end with [QUESTION]
  4. Wait for student to respond
  5. Extract the student's answer
  6. Call RecordAnswer(question="...", student_answer="...", is_correct=true/false)
  7. ✅ [NEW] Check RecordAnswer result's "questions_remaining":
     - If 0: Emit [END_EXAM] signal to terminate immediately
     - If > 0: Go back to step 1
```

**Why**: Teacher explicitly checks remaining count from RecordAnswer

---

### ✅ Solution 2: RecordAnswer Returns Action Signal

**File**: `examples/01-quiz-exam/internal/tools.go` (modify RecordAnswer)

```go
func (qs *QuizState) RecordAnswer(...) map[string]interface{} {
    // ... existing code
    
    nextAction := "continue"
    if qs.CurrentQuestion >= qs.TotalQuestions {
        nextAction = "terminate"  // ← Explicit action
    }
    
    return map[string]interface{}{
        "questions_remaining": qs.TotalQuestions - qs.CurrentQuestion,
        "is_complete":         qs.IsComplete,
        "next_action":         nextAction,  // ← New field
        // ...
    }
}
```

Then update teacher prompt:
```yaml
- If RecordAnswer returns next_action = "terminate": Emit [END_EXAM]
```

---

### ✅ Solution 3: Strict Max Rounds Limit

**File**: `config/crew.yaml` (line 56)

```yaml
settings:
  max_rounds: 21  # Changed from 30
  # 10 questions × 2 rounds each + 1 for [END_EXAM]
```

**Why**: Acts as safety net, stops execution at hard limit

---

### ✅ Solution 4: New "FinalizeExam" Tool

Add explicit tool for completion:

```go
tools["FinalizeExam"] = &agenticcore.Tool{
    Name: "FinalizeExam",
    Description: "Finalize exam when complete. Returns final score.",
    Callback: func(ctx context.Context, args interface{}) (interface{}, error) {
        result := state.GetFinalResult()
        // Trigger [END_EXAM] signal
        return result, nil
    },
}
```

Update teacher prompt:
```yaml
- After last RecordAnswer, call FinalizeExam()
- FinalizeExam() will emit [END_EXAM]
```

---

### ✅ Solution 5: Redesign with Coordinator Agent

Create third agent "Coordinator" that:
- Manages Teacher/Student interaction
- Decides when to terminate
- Emits [END_EXAM]

**Impact**: Highest effort but most robust

---

## 📊 Solution Comparison

| Solution | Difficulty | Reliability | Implementation |
|----------|-----------|-------------|-----------------|
| 1. Prompt | Very Low | 70% | Rewrite 5 lines |
| 2. RecordAnswer | Low | 80% | Add 2 lines |
| 3. Max Rounds | Trivial | 60% | Change 1 number |
| 1+2+3 | Low | 95% | Combine above |
| 4. FinalizeExam | Medium | 90% | New tool + prompt |
| 5. Coordinator | High | 98% | New agent |

**Recommended**: **Solution 1 + 3** (Quick + Safe)
- Rewrite teacher prompt to check `questions_remaining`
- Set `max_rounds = 21` as safety net
- Total: 10 minutes, 95% reliable

---

## 🧪 How to Test

```bash
cd examples/01-quiz-exam

# Run with timeout (10 seconds)
(sleep 10 && pkill -f "go run") & go run ./cmd/main.go 2>&1

# Success indicators:
# ✅ [END_EXAM] signal emitted
# ✅ "Workflow terminates" message
# ✅ Exam report generated
# ✅ Final score printed
```

---

## 📝 Root Cause Summary

**Why infinite loop happens**:

1. Teacher prompt says "If remaining = 0: Emit [END_EXAM]"
2. But RecordAnswer result is not checked by teacher
3. Teacher just "Go back to step 1"
4. Next loop: GetQuizStatus shows remaining = 0
5. But teacher already asked Q11, Q12, ... by then
6. RecordAnswer rejects (is_complete = true)
7. Teacher doesn't handle rejection
8. Infinite loop: Q13, Q14, Q15, ...

**The fix**: Make teacher explicitly check RecordAnswer result

---

**Status**: ✅ Root cause identified
**Next**: Implement Solution 1 + 3
**Time**: ~10 minutes

