# Parallel Execution Architecture - Quiz Exam

## 🎯 Overview

Cấu hình mới sử dụng **Parallel Execution** để tối ưu hóa quy trình thi:
- **Teacher** (orchestrator): Đặt câu hỏi, đánh giá, ghi nhận
- **Student** (responder): Trả lời câu hỏi
- **Reporter** (recorder): Ghi tài liệu

## 🔄 Execution Flow

### Phase 1: Teacher emits [QUESTION]
```
Teacher: "What is 2+2?" [QUESTION]
    ├─→ Student receives [QUESTION] (parallel)
    │   └─→ Student thinks & answers
    └─→ Reporter receives [QUESTION] (parallel)
        └─→ Reporter saves question to report
```

**Timing**: Student & Reporter process SIMULTANEOUSLY

### Phase 2: Student emits [ANSWER]
```
Student: "The answer is 4" [ANSWER]
    └─→ Teacher receives [ANSWER]
        └─→ Teacher evaluates correctness
```

**Timing**: Teacher waits for Student's answer

### Phase 3: Teacher evaluates & records
```
Teacher calls RecordAnswer(question="...", student_answer="...", is_correct=true)
    ├─→ Tool returns: {questions_remaining: 9, ...}
    ├─→ Teacher checks: 9 > 0? YES → Loop back to Phase 1
    └─→ Loop back: Call GetQuizStatus() and ask Q2
```

**Loop**: Repeats until questions_remaining = 0

### Phase 4: Final [END_EXAM]
```
RecordAnswer() returns: {questions_remaining: 0}
    └─→ Teacher emits [END_EXAM]
        └─→ Reporter receives [END_EXAM]
            └─→ Reporter finalizes report
```

**Termination**: Workflow ends cleanly

---

## 📊 Routing Configuration

### Signal Routing
```yaml
routing:
  signals:
    teacher:
      "[QUESTION]":
        parallel_targets:
          - student      # Answer the question
          - reporter     # Record the question
      "[ANSWER]": ""     # Processed internally
      "[END_EXAM]": reporter  # Notify reporter to finalize

    student:
      "[ANSWER]": teacher  # Send answer back

    reporter:
      "[QUESTION]": ""     # Process silently
      "[ANSWER]": ""       # Process silently
      "[END_EXAM]": ""     # Finalize and terminate
```

### Agent Behaviors
```yaml
agent_behaviors:
  teacher:
    wait_for_signal: true        # Wait for Student's [ANSWER]
    parallel_execution: true     # Allow parallel handoffs
  student:
    wait_for_signal: true        # Wait for Teacher's [QUESTION]
    parallel_execution: false    # Sequential processing
  reporter:
    wait_for_signal: true        # Listen for signals
    parallel_execution: false    # Sequential recording
```

---

## 🧠 Agent System Prompts

### Teacher's Role
- Orchestrator of the exam
- Acts as decision maker
- Waits for Student's answer before proceeding
- Dispatches both Student AND Reporter in parallel

**Key Workflow**:
1. GetQuizStatus() → check remaining
2. If > 0: Ask question → Emit [QUESTION] (parallel dispatch)
3. Wait for Student's [ANSWER]
4. Evaluate & RecordAnswer()
5. Check questions_remaining:
   - If > 0: Loop to step 1
   - If = 0: Emit [END_EXAM]

### Student's Role
- Pure responder
- ONLY answers questions
- No tools, no decision-making
- Fast, direct answers

**Key Rules**:
- Answer the question asked
- Provide specific answer (not the question)
- Always end with [ANSWER]
- Keep answers brief

### Reporter's Role
- Silent recorder
- NO thinking, NO chatting
- Triggered by [QUESTION], [ANSWER], [END_EXAM]
- Always calls WriteExamReport()

**Key Rules**:
- ALWAYS call WriteExamReport()
- Say only "Recorded." or "Finalized."
- End with [OK]
- Be FAST

---

## ⏱️ Timing Optimization

### Traditional Sequential (OLD)
```
Q1 → S answers → Record → Q2 → S answers → Record → ... (total: 10*3=30 steps)
Teacher waits for Reporter ❌
```

### Parallel Execution (NEW)
```
Q1 + (S answers || Reporter saves) → Record → Q2 + (S answers || Reporter saves) → ...
Teacher sends both simultaneously ✅
```

**Benefit**: Reporter works while Student thinks = **faster overall**

---

## 🛡️ Safety Constraints

### Max Rounds & Handoffs
```yaml
settings:
  max_rounds: 21      # 10 questions × 2 (Q+A) + 1 ([END_EXAM])
  max_handoffs: 21    # Prevents infinite loops
```

### Timeouts
```yaml
parallel_timeout_seconds: 60      # Max for parallel execution
tool_execution_timeout_seconds: 5 # Per tool call
```

---

## ✅ Checklist: When to Use Parallel Execution

Use **Parallel** when:
- ✅ Multiple agents perform **independent** tasks
- ✅ Tasks can **start simultaneously**
- ✅ No dependencies between concurrent tasks
- ✅ Want to **reduce total execution time**

**Example**: Reporter saves while Student answers = PARALLEL ✅

Use **Sequential** when:
- ❌ Task B depends on Task A's result
- ❌ Specific ordering required
- ❌ One agent needs another's confirmation

**Example**: Teacher waits for Student's answer = SEQUENTIAL ❌

---

## 🔍 Implementation Notes

### Parallel Targets in YAML
```yaml
- signal: "[QUESTION]"
  target: ""  # Empty target
  parallel_targets:  # Multiple simultaneous targets
    - student
    - reporter
```

**Execution**: Framework sends [QUESTION] to BOTH at the same time

### Tool Integration
- **GetQuizStatus**: Checks remaining questions
- **RecordAnswer**: Saves answer to state
- **WriteExamReport**: Saves report to file (called by Reporter)

All tools are **thread-safe** (uses mutex locks)

---

## 📈 Performance Benefits

| Metric | Sequential | Parallel |
|--------|-----------|----------|
| Rounds | 21 | 21 |
| Teacher waits for | 1 agent (Student) | 0 agents (parallel) |
| Reporter blocks | Yes (after each Q+A) | No (simultaneous) |
| Idle time | Student thinks alone | Reporter saves while Student thinks |
| Total time | ~21 LLM calls | ~21 LLM calls (but faster) |

**Result**: Same number of steps, but **less waiting time** 🚀

---

## 🚀 Future Enhancements

1. **True Async Execution**: Use Go goroutines for actual parallelism
2. **Signal Aggregation**: Reporter waits for both [QUESTION] and [ANSWER] together
3. **Fallback Routing**: If Reporter slow, auto-route to backup handler
4. **Metrics Collection**: Track parallel vs sequential performance

---

## 🐛 Troubleshooting

### Reporter not recording?
- Check: Reporter receives [QUESTION] signal?
- Check: WriteExamReport() called successfully?
- Check: Report file permissions writable?

### Student not answering?
- Check: Student receives [QUESTION] signal?
- Check: Student prompt includes [ANSWER] marker?
- Check: LLM responded with valid format?

### Teacher not progressing?
- Check: RecordAnswer() returns valid JSON?
- Check: questions_remaining field parsed correctly?
- Check: [END_EXAM] emitted when remaining=0?

---

## 📝 Configuration Files

- `crew.yaml`: Routing & agent lists
- `agents/teacher.yaml`: Teacher system prompt & tools
- `agents/student.yaml`: Student system prompt & rules
- `agents/reporter.yaml`: Reporter system prompt & tools

---

**Status**: ✅ Parallel execution configured and ready for testing
**Test**: `go run ./examples/01-quiz-exam/cmd/main.go`
