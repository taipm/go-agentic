# 5W2H PHÂN TÍCH CHI TIẾT - QUIZ EXAM FIX STRATEGY

> **Tư duy hệ thống**: Xác định ĐÚNG vị trí cần sửa (CORE hay EXAMPLE) trước khi code

---

## 📌 TÓMBỘC EXECUTIVE SUMMARY

| Aspect | Finding | Fix Location |
|--------|---------|--------------|
| **Root Cause** | Teacher agent generates questions but doesn't pass to tools | EXAMPLE (Agent Prompt) |
| **Data Loss** | Questions/answers stored as `""` or `<nil>` | EXAMPLE (Tool Design) |
| **Auto-Save Issue** | Report written after EACH question (incomplete data) | EXAMPLE (Workflow) |
| **Code Complexity** | 5 separate tools with 50+ lines each | EXAMPLE (Simplification) |
| **Architecture Limitation** | No built-in data validation in core | CORE (Enhancement) |

---

# PHẦN 1: ROOT CAUSE ANALYSIS - 5W2H CHI TIẾT

## 1.1 WHAT? - CÁI GÌ BỊ SAI?

### Biểu Hiện (Symptoms)
```
Exam Report Output:
├── Question 1: [BLANK] → <nil>
├── Student Answer: <nil>
├── Score: ✅ CORRECT (+1 point)  ← Inconsistent!
└── File: exam_20251225_160432.md → All entries empty
```

### Dữ liệu Sai
```go
// Report shows this for ALL 10 questions:
RecordAnswer(
    question="",           // ❌ EMPTY!
    student_answer="<nil>", // ❌ MISSING!
    is_correct=true        // ✅ Paradox: correct despite empty answers
)
```

### Impact
- ✅ Exam completes successfully (10/10)
- ❌ Report useless (no questions/answers visible)
- ❌ Can't review student responses
- ❌ Can't improve assessment

---

## 1.2 WHERE? - Ở ĐÂU BỊ LỖI?

### Data Flow Diagram
```
LAYER 1: Agent (teacher.yaml)
├─ System Prompt: "Ask ONE new question"
├─ LLM Response: "Số nào là 2+2?"
└─ Issue: Question text exists ONLY in LLM output

                          ↓↓↓ DISCONNECT ↓↓↓

LAYER 2: Event Stream (main.go)
├─ Event Type: "agent_response"
├─ Content: "Số nào là 2+2?"
├─ Action: cleanResponse() → removes content, but doesn't forward to tool!
└─ Issue: Question received but NOT passed to RecordAnswer()

                          ↓↓↓ CRITICAL GAP ↓↓↓

LAYER 3: Tool Layer (internal/tools.go)
├─ RecordAnswer() called
├─ Args: { question: "", student_answer: "" }
└─ Issue: Tool receives EMPTY DATA

                          ↓↓↓ GARBAGE IN ↓↓↓

LAYER 4: Data Persistence (internal/tools.go:GenerateMarkdownReport)
├─ Writes: "**Câu hỏi:** "  ← Empty string
├─ Writes: "**Trả lời:** <nil>"  ← Null value
└─ Issue: Report has no data to display
```

### WHERE THE BREAKS HAPPEN

**Break 1: Agent → Tool Communication** (EXAMPLE layer)
```
Teacher generates question text ✅
     ↓
But doesn't include it in RecordAnswer tool call ❌
     ↓
Tool receives empty string ""
```

**Break 2: Stream Processing** (EXAMPLE layer - main.go)
```
Stream event contains question text ✅
     ↓
But main.go only logs/displays it, doesn't capture for tools ❌
     ↓
Never reaches RecordAnswer()
```

**Break 3: Tool Validation** (CORE + EXAMPLE)
```
RecordAnswer receives empty question ❌
     ↓
CORE: Accepts empty string without validation ⚠️
     ↓
EXAMPLE: Tool design allows null values to pass through ❌
```

---

## 1.3 WHY? - TẠI SAO LẠI BỊ LỖI?

### Root Cause Chain (5 levels deep)

**LEVEL 1: Agent Prompt Design (EXAMPLE - teacher.yaml)**
```yaml
# Current System Prompt (Line 40):
"Call RecordAnswer(question="...", student_answer="...", is_correct=true/false)"

# Problem:
- Instruction is there BUT
- LLM (Qwen 3.1.7B) is confused by nested quotes
- LLM sees: question="..." as string placeholder
- LLM generates question text SEPARATELY from tool call
```

**LEVEL 2: Signal-Based Routing Gap (CORE Architecture)**
```
Current Architecture:
- Teacher emits [QUESTION] signal ✅
- Student/Reporter receive parallel ✅
- BUT: No structured data attached to signals ❌

Expected for Data Capture:
- Signal should carry question metadata
- E.g., [QUESTION question="..." student_role="..."]
- CORE doesn't support signal payloads yet
```

**LEVEL 3: LLM Response Parsing (CORE + EXAMPLE)**
```
Ollama Response Structure:
[THINKING] <model reasoning>
[QUESTION] "Số nào là 2+2?"
RecordAnswer(question="", student_answer="", is_correct=true)

Why empty fields?
- Qwen model doesn't reliably fill nested JSON in tool calls
- Model puts data in free text, not in structured parameters
- Tool extraction doesn't cross-reference signal text
```

**LEVEL 4: Tool Parameter Defaults (EXAMPLE - tools.go)**
```go
// Line 390 in tools.go:
isCorrect := true  // ← DEFAULT: assume correct if not provided

// Consequence:
- Accepts empty questions as "valid"
- Marks every answer as correct regardless
- Creates illusion of success
```

**LEVEL 5: Report Generation (EXAMPLE - tools.go)**
```go
// Line 106-107:
sb.WriteString(fmt.Sprintf("**Câu hỏi:** %s\n\n", record.Question))
// If record.Question = "" → writes empty string
// Markdown shows: "**Câu hỏi:** "
```

---

## 1.4 WHEN? - KHI NÀO XẢY RA?

### Timeline of Failure

```
16:04:32 [CONFIG SUCCESS] System starts ✅
16:04:39 [TOOL PARSE] Ollama extracts 1 tool call (RecordAnswer) ✅
         ⚠️ BUT: question parameter = ""

16:04:39 [TOOL ENTRY] RecordAnswer() receives:
         { question: "", student_answer: "", is_correct: true }
         ❌ FIRST FAILURE POINT

16:04:39 [TOOL DEBUG] Before RecordAnswer: CurrentQuestion=0
16:04:39 [TOOL DEBUG] After RecordAnswer: CurrentQuestion=1
         ⚠️ Counter increments successfully (state works)
         ❌ But empty question is recorded

16:04:39 [AUTO-SAVE] WriteReportToFile called
         ❌ Report written with empty data

16:04:39 [TOOL EXIT] returns: is_complete=false, questions_remaining=9
         ✅ Logic works, but data is wrong

[LOOP] Questions 2-9: Same pattern repeats
[FINAL] 16:05:52 Report shows 10/10 with all empty questions
```

### Recurrence Pattern
- **First occurrence**: Question 1 (16:04:39)
- **Every question**: Same issue repeats 10 times
- **Every auto-save**: 10 reports written with incomplete data
- **Final report**: Combines all 10 empty entries

---

## 1.5 WHO? - AI NÓ? (System Components)

### Component Responsibility Matrix

```
┌─────────────────────────────────────────────────────────┐
│ WHO FAILS AT EACH STAGE                                 │
├──────────────────┬──────────┬─────────────────────────┤
│ Component        │ Location │ Failure Mode            │
├──────────────────┼──────────┼─────────────────────────┤
│ 1. LLM (Qwen)    │ EXTERNAL │ Doesn't fill question   │
│                  │          │ field in tool call      │
├──────────────────┼──────────┼─────────────────────────┤
│ 2. Teacher Agent │ EXAMPLE  │ Prompt doesn't guide    │
│ (teacher.yaml)   │          │ LLM to structure data   │
├──────────────────┼──────────┼─────────────────────────┤
│ 3. Tool Parser   │ CORE     │ Accepts empty values    │
│ (tools/extraction.go) │    │ without validation      │
├──────────────────┼──────────┼─────────────────────────┤
│ 4. RecordAnswer  │ EXAMPLE  │ No input validation     │
│ Tool (tools.go)  │          │ Defaults to "correct"   │
├──────────────────┼──────────┼─────────────────────────┤
│ 5. Report Gen    │ EXAMPLE  │ Writes empty strings    │
│ (tools.go)       │          │ to markdown file        │
├──────────────────┼──────────┼─────────────────────────┤
│ 6. Main Loop     │ EXAMPLE  │ Doesn't capture Q&A     │
│ (main.go)        │          │ from stream events      │
└──────────────────┴──────────┴─────────────────────────┘
```

---

## 1.6 HOW? - LÀM THẾ NÀO ĐỂ SỬA?

### Fix Strategy by Component

#### **FIX #1: Agent Prompt (EXAMPLE - teacher.yaml) - CRITICAL**

**Problem**: LLM doesn't understand how to fill tool parameters

**Current**:
```yaml
system_prompt: |
  "Call RecordAnswer(question="...", student_answer="...", is_correct=true/false)"
```

**Fix**:
```yaml
system_prompt: |
  YOU MUST follow this EXACT format for each question:

  STEP 1: Ask the question
  "Here is my question: [WRITE THE EXACT QUESTION TEXT HERE]"

  STEP 2: Emit signal
  [QUESTION]

  STEP 3: Wait for student's [ANSWER] signal

  STEP 4: Extract and record
  Call RecordAnswer() with:
  - question: "[THE EXACT QUESTION FROM STEP 1]"
  - student_answer: "[STUDENT'S RESPONSE FROM THEIR [ANSWER]]"
  - is_correct: [true/false based on evaluation]

  EXAMPLE:
  "Here is my question: Số nào là 2+2?"
  [QUESTION]
  [WAIT FOR STUDENT ANSWER]
  Student responds: "Số 4"
  RecordAnswer(
    question="Số nào là 2+2?",
    student_answer="Số 4",
    is_correct=true
  )
```

**Impact**: ✅ LLM now explicitly extracts question text into tool call

---

#### **FIX #2: Tool Validation (EXAMPLE - tools.go) - CRITICAL**

**Problem**: Tool accepts empty question without validation

**Current Code** (Line 372-374):
```go
// Parse question
question, _ := args["question"].(string)
// ❌ No validation - accepts ""
```

**Fix**:
```go
// Parse and VALIDATE question
question, ok := args["question"].(string)
if !ok || strings.TrimSpace(question) == "" {
    return map[string]interface{}{
        "error": "VALIDATION FAILED: question cannot be empty",
        "status": "rejected",
        "hint": "Include the exact question text in RecordAnswer() call"
    }, nil
}
// ✅ Now rejects empty questions
```

**Additional Validations**:
```go
// Validate student_answer
studentAnswer, ok := args["student_answer"].(string)
if !ok || strings.TrimSpace(studentAnswer) == "" {
    return map[string]interface{}{
        "error": "VALIDATION FAILED: student_answer cannot be empty",
        "hint": "Extract the student's actual response text",
    }, nil
}

// Validate is_correct explicitly required
isCorrect, exists := args["is_correct"].(bool)
if !exists {
    return map[string]interface{}{
        "error": "VALIDATION FAILED: is_correct must be explicitly true or false",
        "hint": "Evaluate correctness and pass boolean value",
    }, nil
}
```

**Impact**: ❌ LLM now gets error feedback and must fix the data

---

#### **FIX #3: Remove Auto-Save (EXAMPLE - tools.go) - HIGH PRIORITY**

**Problem**: Report written 10x with incomplete data

**Current Code** (Line 424-428):
```go
// ❌ AUTO-SAVE: Write report after each answer
if err := state.WriteReportToFile(""); err != nil {
    fmt.Printf("  [Auto-save] Lỗi lưu biên bản: %v\n", err)
}
```

**Fix**:
```go
// ✅ REMOVED: Auto-save after EACH question
// Report only written at END of exam (eliminates intermediate files)
// - Single write operation at completion
// - No partial/incomplete reports
// - Cleaner workflow
```

**Implementation**:
```go
// In RecordAnswer, REMOVE the WriteReportToFile call entirely
// Instead, add to main.go event handler:

case "round_end":
    // Only write report at the very end
    if quizState.IsComplete {
        if err := quizState.WriteReportToFile(""); err != nil {
            fmt.Printf("Error writing final report: %v\n", err)
        }
    }
```

**Impact**: ✅ Single authoritative report file with complete data

---

#### **FIX #4: Consolidate Tools (EXAMPLE - tools.go) - MEDIUM PRIORITY**

**Problem**: 5 separate tools = high complexity + parameter redundancy

**Current**:
```
Tool 1: GetQuizStatus()       (46 lines)
Tool 2: RecordAnswer()        (87 lines)
Tool 3: GetFinalResult()      (16 lines)
Tool 4: WriteExamReport()     (27 lines)
Tool 5: SetExamInfo()         (24 lines)
────────────────────────────────
TOTAL: ~200 lines of boilerplate
```

**Fix - Consolidate to 2 Tools**:

```go
// TOOL 1: QuizExam(action, params)
tools["QuizExam"] = &agenticcore.Tool{
    Name: "QuizExam",
    Description: "Unified exam control - orchestrates all exam operations",
    Parameters: {
        "action": {
            "enum": ["record_answer", "get_status", "get_result"]
        },
        "question": "...",
        "student_answer": "...",
        "is_correct": true/false,
    },
    Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
        action, _ := args["action"].(string)
        switch action {
        case "record_answer":
            // ✅ Validation + recording logic
        case "get_status":
            // ✅ Return current status
        case "get_result":
            // ✅ Return final score
        }
    },
}

// TOOL 2: QuizReport(teacher_comment)
tools["QuizReport"] = &agenticcore.Tool{
    Name: "QuizReport",
    Description: "Finalize exam report with teacher's summary",
    Parameters: {
        "teacher_final_comment": "..."
    },
    Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
        // Only called at END of exam
        // Writes final report to file
    },
}
```

**Benefits**:
- 💾 Reduces code from 200+ → 100 lines
- 🎯 Single entry point for all exam actions
- ✅ Easier to validate parameters
- 📊 Clearer data flow

---

#### **FIX #5: Main Loop Capture (EXAMPLE - main.go) - OPTIONAL ENHANCEMENT**

**Problem**: Question text generated but not captured for archival

**Current**: Questions only in stream, not stored

**Optional Fix**:
```go
// In main.go event handler, capture [QUESTION] signal context:
case "signal":
    if event.Content == "[QUESTION]" {
        // Store question text for audit trail
        lastQuestion = currentAgent + ": " + event.Content
    }
    if event.Content == "[ANSWER]" {
        // Store answer text for audit trail
        lastAnswer = currentAgent + ": " + event.Content
    }

// Later: pass to report generation for transparency
```

**Impact**: 📋 Optional - improves transparency but not critical

---

## 1.7 HOW MUCH? - PHẠM VI SỬA

### Cost-Benefit Analysis

| Fix | Lines Changed | Effort | Priority | Benefit |
|-----|----------------|--------|----------|---------|
| Fix #1: Agent Prompt | 20-30 lines | 10 mins | ⭐⭐⭐ CRITICAL | Data flows correctly |
| Fix #2: Tool Validation | 15-20 lines | 15 mins | ⭐⭐⭐ CRITICAL | Rejects bad data |
| Fix #3: Remove Auto-Save | 10 lines | 5 mins | ⭐⭐⭐ CRITICAL | Single clean report |
| Fix #4: Consolidate Tools | 80-100 lines | 30 mins | ⭐⭐ MEDIUM | Code simplification |
| Fix #5: Main Loop Capture | 15 lines | 10 mins | ⭐ LOW | Audit trail (optional) |

**TOTAL EFFORT**: ~70 minutes
**TOTAL BENEFIT**: All data captured, code simplified, system validated

---

# PART 2: CORE vs EXAMPLE FIX DECISION MATRIX

## 2.1 Which Issues Are CORE Problems?

### CORE Issues (Architecture/Framework Level)

**Issue**: No structured validation in tool parameter extraction

**Current (core/tools/extraction.go)**:
```go
// ❌ Accepts anything without validation
value, _ := args[param].(string)
```

**Should Be (CORE Enhancement)**:
```go
// ✅ CORE should provide validation framework
// tools/validator.go (new file)
type ParamValidator interface {
    Validate(value interface{}) error
}

type StringValidator struct {
    Required bool
    MinLen   int
    MaxLen   int
}

func (v StringValidator) Validate(value interface{}) error {
    s, ok := value.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", value)
    }
    if v.Required && strings.TrimSpace(s) == "" {
        return fmt.Errorf("required string cannot be empty")
    }
    // ... more validation
}
```

**Integration**:
```go
// Tool definition becomes:
tools["RecordAnswer"] = &agenticcore.Tool{
    Name: "RecordAnswer",
    Parameters: map[string]interface{}{
        "question": map[string]interface{}{
            "type": "string",
            "validation": StringValidator{Required: true, MinLen: 5}
        }
    }
}
```

**Impact on CORE**:
- Adds optional validation framework (not breaking)
- Tools can opt-in to validation
- Improves data quality across all examples
- Small (~50 lines), high value

---

## 2.2 Which Issues Are EXAMPLE Problems?

### EXAMPLE Issues (Specific to Quiz-Exam)

| Issue | Location | Scope | Fix Type |
|-------|----------|-------|----------|
| Agent prompt unclear | teacher.yaml | Local | Rewrite |
| Empty question fields | tools.go | Local | Add validation |
| Auto-save redundancy | tools.go | Local | Remove |
| Tool consolidation | tools.go | Local | Refactor |
| Report generation | tools.go | Local | Already OK |

**Key Point**: 80% of fixes are EXAMPLE-specific, don't need CORE changes

---

# PART 3: IMPLEMENTATION ROADMAP

## 3.1 Phase 1: CRITICAL FIXES (30 mins) - EXAMPLE ONLY

### Step 1: Fix Teacher Agent Prompt
**File**: `examples/01-quiz-exam/config/agents/teacher.yaml`

**Change**: Rewrite system_prompt to explicitly guide LLM on data capture

```yaml
# OLD (Lines 28-74)
system_prompt: |
  You are an exam teacher orchestrating a 10-question oral exam.
  ...
  Call RecordAnswer(question="...", student_answer="...", is_correct=true/false)

# NEW (Lines 28-XX)
system_prompt: |
  You are an exam teacher orchestrating a 10-question oral exam.

  WORKFLOW - 4 STRICT STEPS:

  STEP 1: Check remaining questions
  → Call QuizExam(action="get_status")
  → Check response: questions_remaining

  STEP 2: Ask the question (if remaining > 0)
  → Generate question: "Nội dung câu hỏi"
  → Emit signal: [QUESTION]
  → Wait for student's [ANSWER] signal

  STEP 3: Record the answer
  → Extract from student response
  → Call QuizExam(action="record_answer", question="...", student_answer="...", is_correct=true/false)
  → Check response: questions_remaining

  STEP 4: Continue or end
  → If questions_remaining > 0: Loop to STEP 1
  → If questions_remaining = 0: Call QuizReport(...), emit [END_EXAM]
```

---

### Step 2: Add Validation to RecordAnswer
**File**: `examples/01-quiz-exam/internal/tools.go`

**Changes at Line 347-410**:

```go
Func: agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
    fmt.Fprintf(os.Stderr, "[TOOL ENTRY] RecordAnswer() called with args: %v\n", args)

    // ✅ NEW: Validate question (CRITICAL)
    question, ok := args["question"].(string)
    if !ok || strings.TrimSpace(question) == "" {
        fmt.Printf("\n[VALIDATION ERROR] RecordAnswer requires non-empty 'question'\n")
        return json.Marshal(map[string]interface{}{
            "error": "Validation failed: question cannot be empty or missing",
            "received_question": question,
            "hint": "Extract the exact question text and pass it to the tool",
        })
    }

    // ✅ NEW: Validate student_answer (CRITICAL)
    studentAnswer := ""
    switch v := args["student_answer"].(type) {
    case string:
        studentAnswer = v
    case float64, int64, int:
        studentAnswer = fmt.Sprintf("%v", v)
    default:
        studentAnswer = fmt.Sprintf("%v", v)
    }

    if strings.TrimSpace(studentAnswer) == "" {
        fmt.Printf("\n[VALIDATION ERROR] RecordAnswer requires non-empty 'student_answer'\n")
        return json.Marshal(map[string]interface{}{
            "error": "Validation failed: student_answer cannot be empty or missing",
            "received_answer": studentAnswer,
            "hint": "Extract the student's actual response text",
        })
    }

    // ✅ NEW: Validate is_correct explicitly (CRITICAL)
    isCorrect := false
    if ic, exists := args["is_correct"]; exists && ic != nil {
        if b, ok := ic.(bool); ok {
            isCorrect = b
        }
    } else {
        fmt.Printf("\n[VALIDATION ERROR] RecordAnswer requires explicit 'is_correct' (true/false)\n")
        return json.Marshal(map[string]interface{}{
            "error": "Validation failed: is_correct must be explicitly provided",
            "hint": "Evaluate the answer and pass either true or false",
        })
    }

    // Parse teacher_comment (optional)
    teacherComment, _ := args["teacher_comment"].(string)

    // ✅ Proceed with recording (validation passed)
    fmt.Fprintf(os.Stderr, "[TOOL DEBUG] Validation passed: Q='%s', A='%s', Correct=%v\n",
        question, studentAnswer, isCorrect)

    // ... rest of code (unchanged)
}),
```

---

### Step 3: Remove Auto-Save from RecordAnswer
**File**: `examples/01-quiz-exam/internal/tools.go`

**Changes at Line 424-428** - DELETE THESE LINES:

```go
// ❌ REMOVE THIS ENTIRE BLOCK:
// ✅ AUTO-SAVE: Write report after each answer to ensure file is updated
if err := state.WriteReportToFile(""); err != nil {
    fmt.Printf("  [Auto-save] Lỗi lưu biên bản: %v\n", err)
    fmt.Fprintf(os.Stderr, "[TOOL ERROR] WriteReportToFile failed: %v\n", err)
}
```

**Explanation**: Report will be written ONCE at end via main.go event handler

---

## 3.2 Phase 2: CODE SIMPLIFICATION (30 mins) - EXAMPLE

### Consolidate 5 Tools → 2 Tools
**File**: `examples/01-quiz-exam/internal/tools.go`

**New Structure**:
```go
func CreateQuizTools(state *QuizState) map[string]*agenticcore.Tool {
    tools := make(map[string]*agenticcore.Tool)

    // TOOL 1: Unified QuizExam control (combines GetQuizStatus + RecordAnswer + GetFinalResult)
    tools["QuizExam"] = &agenticcore.Tool{
        Name: "QuizExam",
        Description: "Unified exam control: get_status, record_answer, get_result",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "action": map[string]interface{}{
                    "type": "string",
                    "enum": []string{"get_status", "record_answer", "get_result"},
                    "description": "What to do: get_status | record_answer | get_result",
                },
                "question": map[string]interface{}{
                    "type": "string",
                    "description": "Question text (for record_answer only)",
                },
                "student_answer": map[string]interface{}{
                    "type": "string",
                    "description": "Student's response (for record_answer only)",
                },
                "is_correct": map[string]interface{}{
                    "type": "boolean",
                    "description": "Is answer correct? (for record_answer only)",
                },
            },
            "required": []string{"action"},
        },
        Func: agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
            action, _ := args["action"].(string)

            switch action {
            case "get_status":
                result := state.GetStatus()
                fmt.Printf("\n[TOOL] QuizExam(action=get_status)\n")
                fmt.Printf("  Câu đã hỏi: %d/%d\n", result["questions_asked"], result["total_questions"])
                fmt.Printf("  Điểm hiện tại: %d\n", result["current_score"])
                fmt.Printf("  Còn lại: %d câu\n\n", result["questions_remaining"])
                jsonBytes, _ := json.Marshal(result)
                return string(jsonBytes), nil

            case "record_answer":
                // [Validation code from Phase 1, Step 2]
                // [Recording logic]

            case "get_result":
                result := state.GetFinalResult()
                fmt.Printf("\n[TOOL] QuizExam(action=get_result)\n")
                fmt.Printf("  ========== KẾT QUẢ THI ==========\n")
                fmt.Printf("  Điểm: %d/%d (%s)\n", result["final_score"], result["max_score"], result["percentage"])
                fmt.Printf("  Kết quả: %s\n", result["grade"])
                fmt.Printf("  ==================================\n\n")
                jsonBytes, _ := json.Marshal(result)
                return string(jsonBytes), nil
            }

            return `{"error": "unknown action"}`, nil
        }),
    }

    // TOOL 2: Write final report (only at exam end)
    tools["QuizReport"] = &agenticcore.Tool{
        Name: "QuizReport",
        Description: "Finalize exam report with teacher's summary (call at exam end)",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "teacher_final_comment": map[string]interface{}{
                    "type": "string",
                    "description": "Teacher's summary comment",
                },
            },
        },
        Func: agentictools.ToolHandler(func(ctx context.Context, args map[string]interface{}) (string, error) {
            teacherComment, _ := args["teacher_final_comment"].(string)
            err := state.WriteReportToFile(teacherComment)
            if err != nil {
                return fmt.Sprintf(`{"error": "%v"}`, err), nil
            }
            return `{"success": true, "message": "Report written"}`, nil
        }),
    }

    return tools
}
```

**Benefits**:
- Code: 200 lines → 100 lines (50% reduction)
- Clarity: Single entry point per function
- Validation: Consistent error handling

---

## 3.3 Phase 3: CORE ENHANCEMENT (30 mins) - OPTIONAL

### Add Validation Framework to CORE
**File**: `core/tools/validator.go` (NEW)

```go
package tools

import (
    "fmt"
    "strings"
)

// ParamValidator provides validation capability for tool parameters
type ParamValidator interface {
    Validate(value interface{}) error
    Description() string
}

// StringValidator validates string parameters
type StringValidator struct {
    Required bool
    MinLen   int
    MaxLen   int
    Pattern  string // Regex pattern
}

func (v StringValidator) Validate(value interface{}) error {
    s, ok := value.(string)
    if !ok {
        return fmt.Errorf("expected string, got %T", value)
    }

    if v.Required && strings.TrimSpace(s) == "" {
        return fmt.Errorf("required parameter cannot be empty")
    }

    if len(s) < v.MinLen {
        return fmt.Errorf("string too short (min %d chars)", v.MinLen)
    }

    if v.MaxLen > 0 && len(s) > v.MaxLen {
        return fmt.Errorf("string too long (max %d chars)", v.MaxLen)
    }

    return nil
}

// BoolValidator validates boolean parameters
type BoolValidator struct {
    Required bool
}

func (v BoolValidator) Validate(value interface{}) error {
    _, ok := value.(bool)
    if !ok && v.Required {
        return fmt.Errorf("expected boolean, got %T", value)
    }
    return nil
}
```

**Update Tool Definition**:
```go
// In common/types.go, add optional Validators field:
type Tool struct {
    Name        string
    Description string
    Parameters  map[string]interface{}
    Validators  map[string]ParamValidator  // ✅ NEW
    Func        interface{}
}
```

**Update Tool Extraction**:
```go
// In tools/extraction.go, use validators:
func extractToolCall(toolDef *Tool, args map[string]interface{}) error {
    if toolDef.Validators != nil {
        for paramName, validator := range toolDef.Validators {
            if value, exists := args[paramName]; exists {
                if err := validator.Validate(value); err != nil {
                    return fmt.Errorf("validation failed for %s: %w", paramName, err)
                }
            }
        }
    }
    return nil
}
```

---

## 3.4 Implementation Order

### Week 1: CRITICAL FIXES (Fix Example)
```
Day 1:
  ✅ Update teacher.yaml system_prompt (30 mins)
  ✅ Add validation to RecordAnswer (45 mins)
  ✅ Remove auto-save from tools.go (10 mins)
  ✅ Test with quiz-exam example (30 mins)

Day 2:
  ✅ Consolidate tools: 5 → 2 (90 mins)
  ✅ Update student.yaml prompts for new tool (30 mins)
  ✅ Update main.go event handlers (20 mins)
  ✅ Full end-to-end test (30 mins)
```

### Week 2: CORE ENHANCEMENT (Optional)
```
Day 3:
  ✅ Design validation framework (30 mins)
  ✅ Implement in core/tools/validator.go (60 mins)
  ✅ Update Tool definition (15 mins)
  ✅ Update tool extraction logic (30 mins)
  ✅ Add tests (45 mins)
```

---

# SUMMARY TABLE: CORE vs EXAMPLE

| Issue | Root Cause | Fix Location | Effort | Impact |
|-------|-----------|--------------|--------|--------|
| **Q1: Missing question data** | LLM doesn't fill tool params | EXAMPLE (teacher.yaml) | 30 min | ⭐⭐⭐ Data now flows |
| **Q2: Empty fields accepted** | No validation | EXAMPLE (tools.go) | 45 min | ⭐⭐⭐ Bad data rejected |
| **Q3: 10 partial reports** | Auto-save after each Q | EXAMPLE (tools.go) | 10 min | ⭐⭐⭐ Single clean report |
| **Q4: Code too complex** | 5 tools + boilerplate | EXAMPLE (tools.go) | 90 min | ⭐⭐ Maintainability |
| **Q5: No data validation** | CORE lacks validators | CORE (new) | 120 min | ⭐⭐ Optional enhancement |

---

## KEY DECISION

### Do NOT Modify CORE for Quiz-Exam Fixes

✅ **Reason**: Issues are EXAMPLE-specific design, not CORE architecture

✅ **Benefit**: Example stays independent, isolated, testable

✅ **Optional**: Add validation framework to CORE LATER for all examples

