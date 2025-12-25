# 🔍 Quiz-Exam Diagnostic Analysis
**Date**: 2025-12-25
**Status**: Configuration & Routing Analysis

---

## 📋 Configuration Summary

### ✅ What We Have
1. **Teacher Agent** (teacher.yaml)
   - Temperature: 0.7 (balanced)
   - Tools: GetQuizStatus, RecordAnswer ✅
   - Prompt: Updated with [CRITICAL] Step 8 ✅
   - Checks questions_remaining ✅

2. **Student Agent** (student.yaml)
   - Temperature: 0.8 (varied)
   - Tools: None (good - student just answers)
   - Prompt: Clear format with [ANSWER] marker ✅
   - Language: Vietnamese ✅

3. **Reporter Agent** (reporter.yaml)
   - Temperature: 0.1 (deterministic - GOOD)
   - Tools: GetQuizStatus, WriteExamReport ✅
   - Prompt: Simple - just record and emit [OK] ✅
   - Role: Silent recorder ✅

4. **Crew Configuration** (crew.yaml)
   - Entry point: teacher ✅
   - max_rounds: 21 (10×2+1) ✅
   - max_handoffs: 21 ✅
   - config_mode: strict ✅

---

## 🔴 Potential Issues Found

### Issue 1️⃣: Parallel Groups Routing Mismatch

**File**: crew.yaml (lines 36-45)

```yaml
parallel_groups:
  parallel_question:      # ❌ NOT TRIGGERED BY [QUESTION]
    agents: [student, reporter]
    wait_for_all: false
    timeout_seconds: 30
  parallel_answer:        # ❌ NOT TRIGGERED BY [ANSWER]
    agents: [teacher, reporter]
    wait_for_all: false
    timeout_seconds: 30
```

**Problem**:
- Parallel groups are defined but never explicitly invoked
- Teacher emits `[QUESTION]` → routes to student (lines 21-22)
- Student emits `[ANSWER]` → routes to teacher (lines 26-27)
- Parallel groups are NOT mentioned in signal routing
- Reporter never receives explicit signal to run

**Current Flow**:
```
Teacher asks [QUESTION]
  ↓ Routes to: student (not parallel)
Student answers [ANSWER]
  ↓ Routes to: teacher (not parallel)
Teacher calls RecordAnswer
  ↓ Tool runs, writes to file
[NO PARALLEL GROUP INVOKED FOR REPORTER]
```

**Expected vs Actual**:
- Expected: Teacher → [QUESTION] → Student & Reporter run in parallel
- Actual: Teacher → [QUESTION] → Student only

---

### Issue 2️⃣: Reporter Never Invoked in Signal Flow

**File**: crew.yaml signals (lines 19-34) + parallel_groups (lines 36-45)

**Problem**:
- Reporter routing is defined (lines 28-34):
  ```yaml
  reporter:
    - signal: "[OK]"
      target: ""
    - signal: "[DONE]"
      target: ""
    - signal: "[END_EXAM]"
      target: ""
  ```

- But Reporter is NEVER triggered because:
  1. Entry point is `teacher`
  2. Only Teacher → Student routing is active
  3. Reporter has no signal from Teacher or Student
  4. Parallel groups exist but aren't used

**Result**: Reporter sits idle, never writes updates, never emits [OK]

---

### Issue 3️⃣: Parallel Groups Not Connected to Signals

**File**: crew.yaml

**Problem**: Parallel groups are defined (lines 36-45) but:
- No signal mapping connects them
- No agent behavior triggers them
- They're essentially orphaned configuration

**Example**:
```yaml
signals:
  teacher:
    - signal: "[QUESTION]"
      target: student        # ❌ Should this use parallel_question group?

parallel_groups:
  parallel_question:         # ❌ Never referenced above!
    agents: [student, reporter]
```

---

### Issue 4️⃣: Missing Parallel Execution Configuration

**File**: crew.yaml

**Analysis**:
- `agent_behaviors` specifies `wait_for_signal: false` (lines 48-53)
- This means agents don't wait - they run immediately
- But there's no explicit "run these agents in parallel" instruction

**Current Architecture**:
```
Sequential Routing (Current):
Round 1: Teacher
Round 2: Student (after Teacher's [QUESTION])
Round 3: Teacher (after Student's [ANSWER])
...

Parallel Routing (Intended):
Round 1: Teacher
Round 2: Student & Reporter (parallel)
Round 3: Teacher & Reporter (parallel)
```

---

## 🔍 Root Cause Analysis

### Why Reporter Doesn't Work

1. **Entry point**: Teacher (line 10)
   - Only Teacher starts automatically

2. **Signal routing** (lines 19-27):
   - Teacher → [QUESTION] → Student
   - Student → [ANSWER] → Teacher
   - NO WAY to trigger Reporter!

3. **Parallel groups** (lines 36-45):
   - Defined but NOT integrated with signal routing
   - No signal mapping says "use parallel_question here"
   - No agent behavior triggers them

4. **Result**:
   - Reporter never gets executed
   - WriteExamReport never called
   - Reporter never emits [OK] or [DONE]
   - Workflow only has Teacher ↔ Student loop

---

## ✅ 5 Solutions

### Solution 1: Add Reporter to Main Signal Flow (RECOMMENDED)

**File**: crew.yaml

```yaml
routing:
  signals:
    teacher:
      - signal: "[QUESTION]"
        target: [student, reporter]    # ← Add reporter as target
      - signal: "[END_EXAM]"
        target: ""
    student:
      - signal: "[ANSWER]"
        target: teacher
    reporter:
      - signal: "[OK]"
        target: ""
      - signal: "[END_EXAM]"
        target: ""
```

**Effort**: Very Low (add 1 target)
**Result**: Reporter runs after each question

---

### Solution 2: Use Parallel Groups in Signal Routing

**File**: crew.yaml

```yaml
routing:
  signals:
    teacher:
      - signal: "[QUESTION]"
        target: parallel_question    # ← Reference group

  parallel_groups:
    parallel_question:
      agents: [student, reporter]
      wait_for_all: false
```

**Effort**: Low (change target type)
**Result**: Student & Reporter run together

---

### Solution 3: Separate Reporter from Main Flow

**Alternative Architecture**:

```yaml
agents:
  - teacher
  - student
  - reporter

routing:
  signals:
    teacher:
      - signal: "[QUESTION]"
        target: student
      - signal: "[END_EXAM]"
        target: ""
      - signal: "[REPORT]"          # ← New signal
        target: reporter
    student:
      - signal: "[ANSWER]"
        target: teacher
    reporter:
      - signal: "[OK]"
        target: ""
```

Update teacher prompt:
```yaml
- After RecordAnswer: emit [REPORT] to trigger reporter
- Reporter receives [REPORT], calls WriteExamReport, emits [OK]
```

**Effort**: Medium (new signal + prompt update)
**Result**: Clean separation of concerns

---

### Solution 4: Make Reporter Always-On (Auto-Execute)

**File**: crew.yaml

```yaml
agent_behaviors:
  teacher:
    wait_for_signal: false
  student:
    wait_for_signal: false
  reporter:
    wait_for_signal: false
    auto_run: true              # ← New: Always runs
    run_after_agents: [teacher, student]  # ← Run after these
```

**Effort**: High (if auto_run not implemented)
**Result**: Reporter always participates

---

### Solution 5: Redesign as 3-Agent Coordinator

**Architecture**:
```
Teacher → [QUESTION] → Student
Student → [ANSWER] → Teacher
Teacher → [RECORD] → Coordinator
Coordinator → [REPORT] → Reporter
Reporter → [OK] → Coordinator
Coordinator → [CONTINUE/END] → Teacher
```

**Effort**: Highest
**Result**: Most robust

---

## 📊 Solution Comparison

| Solution | Difficulty | Impact | Reliability | Time |
|----------|-----------|--------|------------|------|
| 1. Add reporter target | Very Low | High | 80% | 5 min |
| 2. Use parallel groups | Low | High | 85% | 10 min |
| 3. Separate signal | Medium | Medium | 90% | 20 min |
| 4. Auto-run | High | Low | 60% | 30 min |
| 5. Coordinator | Highest | Highest | 95% | 60 min |

**Recommended**: **Solution 1** (Quick & Effective)

---

## 🧪 Diagnostic Test Plan

### Test 1: Verify Current Behavior
```bash
cd examples/01-quiz-exam
(sleep 15 && pkill -f "go run") & go run ./cmd/main.go 2>&1 | tee test-output.log

# Check:
# 1. How many times does Reporter appear in output?
# 2. Is WriteExamReport called?
# 3. Does report file exist and get updated?
```

### Test 2: Add Debug Logging
Add to crew.yaml:
```yaml
debug_mode: true
log_signals: true
log_routing: true
```

### Test 3: Verify Signal Emission
Watch for:
- `[QUESTION]` from teacher
- `[ANSWER]` from student
- `[OK]` from reporter (missing?)
- `[END_EXAM]` from teacher

---

## 📝 Summary

**Current State**:
- ✅ Teacher prompt: Correct
- ✅ Student agent: Correct
- ❌ Reporter: Never invoked
- ❌ Parallel groups: Orphaned
- ❌ WriteExamReport: Rarely called

**Root Cause**: Reporter is defined but never triggered in signal routing

**Recommended Fix**: Solution 1
- Add reporter as target in teacher's [QUESTION] signal
- 5 minutes, very safe, high confidence

**Next Steps**:
1. Apply Solution 1 to crew.yaml
2. Add debug logging to confirm signal routing
3. Run diagnostic test
4. Verify reporter WriteExamReport is called 10 times (once per question)

---
