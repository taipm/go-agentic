# Parallel Execution Architecture - Visual Diagrams

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        QUIZ EXAM SYSTEM                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Entry Point: teacher                                             │
│                                                                   │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐   │
│  │   TEACHER    │      │   STUDENT    │      │   REPORTER   │   │
│  │ Orchestrator │      │  Responder   │      │   Recorder   │   │
│  ├──────────────┤      ├──────────────┤      ├──────────────┤   │
│  │ Tools:       │      │ Tools:       │      │ Tools:       │   │
│  │ - GetQuiz    │      │ (none)       │      │ - WriteReport│   │
│  │ - RecordAns  │      │              │      │ - GetQuiz    │   │
│  └──────────────┘      └──────────────┘      └──────────────┘   │
│       ▲                      ▲                      ▲              │
│       └──────────┬───────────┴──────────┬──────────┘              │
│                  │                      │                         │
│             Signals & Routing           │                         │
│                                    File System                    │
│                                  (report.md)                      │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔄 Execution Flow - Full Cycle

```
╔════════════════════════════════════════════════════════════════════════════╗
║                          QUESTION CYCLE (N=1..10)                          ║
╚════════════════════════════════════════════════════════════════════════════╝

STEP 1: Teacher Checks Status
┌──────────────────────────────────────────────────────────────────────────┐
│ Teacher: GetQuizStatus()                                                  │
│ Returns: {remaining: 10, current_score: 0}                              │
└──────────────────────────────────────────────────────────────────────────┘

STEP 2: Teacher Asks Question & PARALLEL DISPATCH
┌────────────────────────────────────────────────────────────────────────┐
│ Teacher emits: [QUESTION] "What is 2+2?"                              │
│                                                                         │
│    ┌─────────────────────────────────────────────────────────┐        │
│    │ PARALLEL ROUTING TARGETS                                │        │
│    ├─────────────────────────────────────────────────────────┤        │
│    │                                                          │        │
│    │  Target 1: Student                                      │        │
│    │  ├─ Receives: [QUESTION]                               │        │
│    │  ├─ Reads question: "What is 2+2?"                     │        │
│    │  └─ Status: THINKING...                                │        │
│    │                                                          │        │
│    │  Target 2: Reporter                                     │        │
│    │  ├─ Receives: [QUESTION]                               │        │
│    │  ├─ Calls: WriteExamReport()                           │        │
│    │  └─ Emits: [OK]                                        │        │
│    │                                                          │        │
│    └─────────────────────────────────────────────────────────┘        │
│                                                                         │
│  ⏱️  TIMING: Both triggered SIMULTANEOUSLY                            │
│      Teacher doesn't wait for Reporter                                 │
│                                                                         │
└────────────────────────────────────────────────────────────────────────┘

STEP 3: Student Processes
┌──────────────────────────────────────────────────────────────────────────┐
│ Student: Thinking...                                                      │
│ (This can take 1-2 seconds with LLM)                                     │
│                                                                           │
│ Time available: T1 (question sent) → T3 (answer received)              │
│ Time used: Student LLM inference                                         │
│ Time gained: While student thinks, reporter saves (parallel!)           │
│                                                                           │
│ Student emits: [ANSWER] "The answer is 4"                              │
│ Routes to: teacher                                                       │
└──────────────────────────────────────────────────────────────────────────┘

STEP 4: Teacher Evaluates & Records
┌──────────────────────────────────────────────────────────────────────────┐
│ Teacher receives: [ANSWER] "The answer is 4"                            │
│ Teacher evaluates: Is "4" correct for "2+2"? YES ✅                    │
│                                                                           │
│ Teacher calls: RecordAnswer(                                             │
│   question: "What is 2+2?",                                             │
│   student_answer: "The answer is 4",                                    │
│   is_correct: true                                                       │
│ )                                                                         │
│                                                                           │
│ Tool returns: {                                                           │
│   total_score: 1,                                                        │
│   questions_remaining: 9,                                                │
│   is_complete: false                                                     │
│ }                                                                         │
└──────────────────────────────────────────────────────────────────────────┘

STEP 5: Teacher Checks Loop Condition
┌──────────────────────────────────────────────────────────────────────────┐
│ Teacher: Check questions_remaining?                                      │
│                                                                           │
│ ┌─ IF questions_remaining > 0: LOOP back to STEP 1                     │
│ │  Ask Question 2, 3, ..., 10                                          │
│ │                                                                        │
│ └─ IF questions_remaining = 0: TERMINATE                               │
│    Emit [END_EXAM]                                                      │
│    Reporter finalizes report                                            │
│                                                                          │
│ Current: questions_remaining = 9 → Continue loop                       │
└──────────────────────────────────────────────────────────────────────────┘

REPEAT STEPS 1-5 for remaining 9 questions...

╔════════════════════════════════════════════════════════════════════════════╗
║                    AFTER 10 QUESTIONS: FINAL STATE                         ║
╚════════════════════════════════════════════════════════════════════════════╝

RecordAnswer() returns: questions_remaining: 0
   ↓
Teacher emits: [END_EXAM]
   ├─→ Reporter receives [END_EXAM]
   │   ├─ Calls WriteExamReport() (final)
   │   └─ Emits [OK]
   │
   └─→ Workflow terminates
       Report saved to: examples/01-quiz-exam/reports/exam_TIMESTAMP.md
```

---

## 📊 Timing Comparison

### BEFORE: Sequential Execution
```
Time  │ Teacher              │ Student              │ Reporter
──────┼──────────────────────┼──────────────────────┼──────────────────────
T0    │ GetQuizStatus()      │                      │
T1    │ Ask Q1 [QUESTION]    │ Receives [QUESTION]  │
T2    │                      │ Thinking...          │ (waiting)
T3    │                      │ Thinking...          │ (waiting)
T4    │ (waiting)            │ [ANSWER] ───→        │ (waiting)
T5    │ ← receives [ANSWER]  │                      │ (waiting)
T6    │ Evaluate & Record    │                      │ (waiting)
T7    │ WriteReport ────────────────────────────────→ Record it
T8    │                      │                      │ Saving...
T9    │                      │                      │ [OK] back
T10   │ ← receives [OK]      │                      │
      │ (blocked on Reporter)                       │
      │                                              │
T11   │ GetQuizStatus()      │                      │
...   │ Loop back Q2         │                      │
```

**Total Time**: ~10 * (T_question + T_answer + T_record + T_report)
**Bottleneck**: Reporter blocks Teacher after each answer

---

### AFTER: Parallel Execution
```
Time  │ Teacher              │ Student              │ Reporter
──────┼──────────────────────┼──────────────────────┼──────────────────────
T0    │ GetQuizStatus()      │                      │
T1    │ Ask Q1 [QUESTION]    │ Receives [QUESTION]  │ Receives [QUESTION]
T2    │ (waiting for         │ Thinking...          │ → WriteReport()
      │  Student's answer)   │                      │
T3    │                      │ Thinking...          │ [OK] (done!)
T4    │                      │ [ANSWER] ───→        │
T5    │ ← receives [ANSWER]  │                      │
T6    │ Evaluate & Record    │                      │
      │ (no wait for Reporter!)                     │
T7    │ GetQuizStatus()      │                      │
T8    │ Ask Q2 [QUESTION]    │ Receives [QUESTION]  │ Receives [QUESTION]
T9    │ (waiting for         │ Thinking...          │ → WriteReport()
      │  Student's answer)   │                      │
...   │ Loop back            │                      │
```

**Total Time**: ~10 * (T_question + T_answer + T_record)
**Bottleneck**: Only Teacher waiting for Student (correct dependency)
**Gain**: Reporter finishes DURING Student thinking (parallel!)

**Time Saved**: ~10 * T_report per cycle = significant for multiple cycles

---

## 🎯 Signal Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                      SIGNAL ROUTING                              │
└─────────────────────────────────────────────────────────────────┘

Teacher Agent:
───────────────

   ┌─────────────────────┐
   │  [QUESTION] signal  │
   └──────────┬──────────┘
              │
         ┌────┴─────┐
         ▼          ▼
    Student     Reporter
    ────────    ────────
    [QUESTION]  [QUESTION]
         │          │
         │          └────→ WriteExamReport()
         │                      │
         │                      └─→ [OK]
         │
         └──→ (Student thinks & answers)
                      │
                      ▼
                   [ANSWER]
                      │
                      └────→ Teacher
                                │
                                └─→ Evaluate & RecordAnswer()
                                      │
                                      ├─→ questions_remaining: 9?
                                      │   YES: Loop back
                                      │
                                      └─→ questions_remaining: 0?
                                          YES: [END_EXAM]
                                                  │
                                                  └──→ Reporter
                                                         │
                                                         └─→ Finalize

Teacher Agent:
───────────────
   ┌──────────────────┐
   │  [END_EXAM] sig  │
   └────────┬─────────┘
            │
            ▼
        Reporter
        ────────
        [END_EXAM]
             │
             └─→ WriteExamReport() (final)
                     │
                     └─→ [OK]
                           │
                           └─→ Workflow terminates
```

---

## 🔀 Parallel vs Sequential Comparison

### Sequential Model (OLD)
```
Teacher ──[Q]──→ Student
         ◀──[A]──
         ──[Record]──→ (internal)
         ──[Report]──→ Reporter
         ◀──[OK]──
         (blocked here!) ❌

         ──[Q]──→ Student  (next iteration)
```

### Parallel Model (NEW)
```
Teacher ──[Q]──┬──→ Student (receives)
               │
               └──→ Reporter (receives)

         ◀──[A]── Student       (waits only here!)

         ──[Record]──→ (internal)

Reporter [saves in parallel]   (no blocking!)
         ──[Q]──→ Student  (next iteration) ✅
```

---

## 📈 Resource Utilization

```
Sequential Model (OLD):
──────────────────────

Teacher:  ████░░░░░░░░░░░░░░  (thinking)
Student:  ░░░░████░░░░░░░░░░  (answering)
Reporter: ░░░░░░░░████░░░░░░  (saving) - BLOCKS Teacher!
          └─ idle while Reporter saves

Total: 20 units (3 busy, 1 blocked)


Parallel Model (NEW):
───────────────────

Teacher:  ████░░░░░░░░░░░░░░  (thinking)
Student:  ░░░░████░░░░░░░░░░  (answering)
Reporter: ░░░░████░░░░░░░░░░  (saving) - PARALLEL!
          └─ no idle time!

Total: 20 units (3 busy, 0 blocked) ✅
```

---

## 🎓 Key Concepts

### 1. Parallel Dispatch
- **What**: Send multiple [QUESTION] signals to different agents
- **When**: Teacher asks question
- **To**: Student (answer) + Reporter (save)
- **Benefit**: Both start working immediately

### 2. Sequential Wait
- **What**: Wait for Student's [ANSWER] before continuing
- **Why**: Answer is required to evaluate and record
- **Not waiting for**: Reporter (that's why parallel!)
- **Benefit**: Correct dependency

### 3. Signal-based Routing
- **What**: Use [QUESTION], [ANSWER], [END_EXAM] markers
- **How**: Framework routes based on signal content
- **Why**: Decouples agents, enables flexible orchestration

### 4. Orchestrator Pattern
- **What**: Teacher coordinates the entire workflow
- **Not a participant**: Just dispatches and waits for answers
- **Delegates**: Recording, reporting to specialized agents
- **Benefit**: Clear separation of concerns

---

## ✅ Verification Checklist

When testing parallel execution, verify:

- [ ] Teacher emits [QUESTION]
- [ ] Student receives [QUESTION] (check logs)
- [ ] Reporter receives [QUESTION] (check logs)
- [ ] Both process simultaneously (check timestamps)
- [ ] Student emits [ANSWER]
- [ ] Teacher receives [ANSWER]
- [ ] Teacher evaluates & records
- [ ] Loop back to question 2
- [ ] After 10 questions: Teacher emits [END_EXAM]
- [ ] Reporter receives [END_EXAM]
- [ ] Report file generated correctly

---

**Last Updated**: 2025-12-25
**Status**: ✅ Architecture designed for parallel execution
**Implementation**: Phase 3.1+
