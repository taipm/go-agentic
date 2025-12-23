# Quiz Exam - Multi-Agent Oral Exam Demo

Demo multi-agent workflow with 3 agents using signal-based routing and parallel execution.

## Description

Simulates an oral exam with:
- **Teacher**: Asks questions, grades answers using tools
- **Student**: Answers questions
- **Reporter**: Records exam progress in real-time

## Workflow Architecture

```
┌──────────────┐                      ┌──────────────┐
│   Teacher    │ ────[QUESTION]────▶  │   Student    │
│              │                      │              │
│  Tools:      │                      │  (No tools)  │
│  - GetStatus │                      │              │
│  - Record    │ ◀────[ANSWER]─────   │              │
│  - GetResult │                      │              │
└──────────────┘                      └──────────────┘
       │                                     │
       │ [QUESTION]                          │ [ANSWER]
       ▼                                     ▼
┌──────────────┐                      ┌──────────────┐
│   Reporter   │ ◀───────────────────│   Reporter   │
│              │                      │   (parallel) │
│  Tools:      │                      │              │
│  - GetStatus │                      │              │
│  - WriteRpt  │                      │              │
└──────────────┘                      └──────────────┘
       │
       │ [END] → Reporter → [DONE] → Terminate
       ▼
```

### Signal Flow

| Signal | Source | Target | Description |
|--------|--------|--------|-------------|
| `[QUESTION]` | Teacher | Student + Reporter (parallel) | Teacher asks a question |
| `[ANSWER]` | Student | Teacher + Reporter (parallel) | Student answers |
| `[END]` | Teacher | Reporter | Exam complete, final save |
| `[DONE]` | Reporter | (terminate) | Workflow ends |

## Rules

- 10 questions per exam
- 1 point per correct answer
- Score > 5 = **PASS**
- Score ≤ 5 = **FAIL**

## Tools

### Teacher Tools

| Tool | Description |
|------|-------------|
| `GetQuizStatus()` | Check progress: questions asked, current score |
| `RecordAnswer(question_number, question, student_answer, is_correct, teacher_comment)` | Record answer with grade |
| `GetFinalResult()` | Get final exam results |

### Reporter Tools

| Tool | Description |
|------|-------------|
| `GetQuizStatus()` | Check current exam state |
| `WriteReport()` | Save report to markdown file |

## Report Generation

Reports are auto-saved after each `RecordAnswer()` call. Each report includes:

- Exam info (student name, topic, timestamp)
- Question-by-question details (✅/❌)
- Teacher comments per question
- Final score and pass/fail status

**Example output:** `reports/exam_20251223_150416.md`

## Running the Demo

```bash
# Install dependencies
make deps

# Run demo
make run

# Run with verbose output
make run-verbose

# Run with custom output directory
go run ./cmd -output=./my-reports
```

## Requirements

- Go 1.21+
- Ollama running locally with model `qwen2.5-coder:7b`

## Project Structure

```
01-quiz-exam/
├── cmd/
│   └── main.go              # Entry point
├── config/
│   ├── crew.yaml            # Crew config with parallel routing
│   └── agents/
│       ├── teacher.yaml     # Teacher agent config
│       ├── student.yaml     # Student agent config
│       └── reporter.yaml    # Reporter agent config
├── internal/
│   └── tools.go             # Quiz tools implementation
├── reports/                 # Generated exam reports
├── go.mod
├── Makefile
└── README.md
```

## Configuration Highlights

### Parallel Groups (crew.yaml)

```yaml
routing:
  signals:
    teacher:
      - signal: "[QUESTION]"
        target: parallel_question  # -> Student + Reporter
      - signal: "[END]"
        target: reporter
    student:
      - signal: "[ANSWER]"
        target: parallel_answer    # -> Teacher + Reporter

  parallel_groups:
    parallel_question:
      agents: [student, reporter]
      wait_for_all: false
      timeout_seconds: 30
    parallel_answer:
      agents: [teacher, reporter]
      wait_for_all: false
      timeout_seconds: 30
```

### Agent Signals

| Agent | Signal | Action |
|-------|--------|--------|
| Teacher | `[QUESTION]` | Triggers parallel: Student answers, Reporter records |
| Teacher | `[END]` | Exam complete, triggers final report |
| Student | `[ANSWER]` | Triggers parallel: Teacher grades, Reporter records |
| Reporter | `[OK]` | Acknowledges (no routing) |
| Reporter | `[DONE]` | Terminates workflow |

## Features Demonstrated

- ✅ **Signal-based routing**: Bracket signals `[SIGNAL]` drive agent handoffs
- ✅ **Parallel execution**: Multiple agents process simultaneously
- ✅ **Tool memory**: QuizState maintains score across turns
- ✅ **Multi-agent workflow**: 3 agents with distinct roles
- ✅ **Streaming events**: Real-time output display
- ✅ **Auto-save reports**: Continuous markdown generation
- ✅ **STRICT mode config**: All parameters explicitly set
- ✅ **Termination signals**: Clean workflow exit with `target: ""`

## Core Framework Features Used

This example demonstrates key go-agentic features:

1. **Parallel Groups**: Route one signal to multiple agents simultaneously
2. **Signal Normalization**: Case-insensitive, whitespace-tolerant matching
3. **Termination Signals**: `target: ""` cleanly ends workflow
4. **Tool Calling**: Ollama text-based tool parsing with retry logic
5. **Quota Enforcement**: STRICT mode with all 30 parameters configured
