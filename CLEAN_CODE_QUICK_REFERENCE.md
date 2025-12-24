# CLEAN CODE - QUICK REFERENCE CARD

**Use when**: Starting any coding task, reviewing code, refactoring

---

## 🎯 THREE THINKING PATTERNS

### **1️⃣ FIRST PRINCIPLES (Ask: "What is ESSENTIAL?")**

| Question | Apply to Code | Example |
|----------|---------------|---------|
| "Tại sao tồn tại?" | Function purpose | Why does `updateMetrics()` exist? Is it essential? |
| "Cái gì cốt lõi?" | Remove accidental complexity | Remove logging from core logic → move to separate function |
| "Tại sao người viết như vậy?" | Challenge assumptions | Why is history shared? Can we use channels instead? |
| "Có thể bỏ được không?" | Minimize code | Do we need 5 quota checks or 1 placed strategically? |

**Action**: Before refactoring → Ask 3 questions above

---

### **2️⃣ CLEAN CODE (Ask: "Will someone understand this?")**

| Principle | Pattern | Code Example |
|-----------|---------|--------------|
| **Names** | Use intention-revealing names | `currentAgent` not `ca` |
| **Functions** | Do one thing well | 1 function = 1 responsibility |
| **Comments** | Explain WHY, not WHAT | Why lock here, not what lock does |
| **Errors** | Handle explicitly | `if err != nil { return err }` |
| **Structure** | Group related code | All history access together |

**Action**: Apply to every function/variable

---

### **3️⃣ SPEED OF EXECUTION (Ask: "Can I scan & understand in 30 seconds?")**

| Aspect | Pattern | Bad | Good |
|--------|---------|-----|------|
| **Locality** | Related code close | History logic scattered across 500 lines | All history in 1 struct |
| **Consistency** | Same pattern everywhere | Some locks, some not | All shared state locked |
| **Obviousness** | Clear intent | `ce.h = append(...)` | `ce.addMessage(msg)` |

**Action**: Make code **scannable** in under 1 minute

---

## 📋 CLEAN CODE CHECKLIST (30 seconds)

```
Function: [name]
Lines: [count]
Responsibilities: [how many things does it do?]

CHECK:
☐ Does ONE thing (if >1, split it)
☐ <20 lines (if >20, extract helpers)
☐ Clear name (read without documentation)
☐ Error handling explicit (no ignored errors)
☐ Shared state protected (mutex if concurrent)
☐ Comments explain WHY (not WHAT)
☐ No dead code
☐ No magic numbers

Result: ✅ READY or 🔧 REFACTOR
```

---

## 🚀 QUICK PATTERNS

### **Pattern 1: Safe Shared State**
```go
// ❌ NOT SAFE
type Executor struct {
    history []Message
}
ce.history = append(...)  // Race!

// ✅ SAFE
type Executor struct {
    state struct {
        sync.RWMutex
        history []Message
    }
}
func (e *Executor) appendMessage(msg Message) {
    e.state.Lock()
    defer e.state.Unlock()
    e.state.history = append(e.state.history, msg)
}
```

### **Pattern 2: Single Responsibility**
```go
// ❌ TOO MANY JOBS
func Execute(input string) error {
    validate(input)              // Job 1
    response := callLLM()        // Job 2
    processTools(response)       // Job 3
    updateMetrics()              // Job 4
    route(signal)                // Job 5
}

// ✅ ONE JOB
func Execute(input string) error {
    return executeMainLoop(input)
}
func executeMainLoop(input string) error {
    if err := validate(input); err != nil { return err }
    if err := callAndProcess(); err != nil { return err }
    return route()
}
func validate(input string) error { ... }      // 1 job
func callAndProcess() error { ... }            // 1 job
func route() error { ... }                     // 1 job
```

### **Pattern 3: Explicit Error Handling**
```go
// ❌ SILENT FAILURE
response, _ := agent.Execute()
cost, _ := estimate()

// ✅ EXPLICIT
response, err := agent.Execute()
if err != nil {
    return fmt.Errorf("execute failed: %w", err)
}
cost, err := estimate()
if err != nil {
    return fmt.Errorf("estimate failed: %w", err)
}
```

### **Pattern 4: Hide Complexity**
```go
// ❌ EXPOSED
func (a *Agent) Execute() {
    a.metadata.Lock()
    a.metadata.tokens += estimate()
    a.metadata.Unlock()
    a.costMetrics.Lock()
    a.costMetrics.cost += calculate()
    a.costMetrics.Unlock()
}

// ✅ HIDDEN
func (a *Agent) Execute() error {
    return a.internal.execute()  // Complexity hidden
}
```

---

## 🔍 METRICS TO MEASURE

**Run before → After refactoring**:

```bash
# Complexity
gocyclo -avg .

# Coverage
go test -cover ./...

# Race conditions
go test -race ./...

# Lint
golangci-lint run ./...
```

**Goals**:
- Complexity: 5-10 average
- Coverage: ≥85%
- -race: 0 warnings
- Lint: 0 errors

---

## 📝 REVIEW TEMPLATE (5 min per function)

```
Function: [name]

✅ Good:
- [What's well done]

🔴 Issues:
1. [Issue 1] → Fix: [action]
2. [Issue 2] → Fix: [action]

🟡 Improvements:
1. [Nice-to-have]
2. [Future enhancement]

Decision: APPROVE / REQUEST CHANGES
```

---

## 🎯 WHEN TO USE WHICH PROMPT

| Situation | Prompt | Goal |
|-----------|--------|------|
| Starting refactor | First Principles | Understand essential |
| Reviewing code | Clean Code Lens | Find issues |
| Simplifying function | SRP Pattern | Split into 1-job functions |
| Fixing race condition | Mutex Pattern | Make thread-safe |
| Hiding complexity | Interface Pattern | Clean API |
| Measuring quality | Metrics | Track improvement |

---

## 💡 DECISION TREE

```
Start here: This code feels messy

Q1: Is it hard to understand?
  ├─ YES → Apply "Clean Code" (better names, clear structure)
  └─ NO → Go to Q2

Q2: Is it doing too much?
  ├─ YES → Apply "SRP Pattern" (split into single jobs)
  └─ NO → Go to Q3

Q3: Is it concurrent/shared state?
  ├─ YES → Apply "Mutex Pattern" (protect with locks)
  └─ NO → Go to Q4

Q4: Is it exposing too much?
  ├─ YES → Apply "Interface Pattern" (hide complexity)
  └─ NO → Code is likely fine

Result: Apply matching refactoring
```

---

## 📌 TOP 3 PRIORITIES FOR GO-AGENTIC

### 🔴 **Priority 1: Fix Race Condition (Crew History)**
- **What**: History modified without lock
- **Impact**: Lost data, panic, corruption
- **Fix**: Add `sync.RWMutex` to state access
- **Prompt**: Pattern 1 (Safe Shared State)
- **Time**: 1 hour

### 🔴 **Priority 2: Enforce Quotas Consistently**
- **What**: Quota checks missing on parallel path
- **Impact**: Cost overrun, budget bypass
- **Fix**: Apply same quota check everywhere
- **Prompt**: First Principles + SRP
- **Time**: 2 hours

### 🟡 **Priority 3: Simplify ExecuteStream**
- **What**: 1000+ lines, 10+ responsibilities
- **Impact**: Hard to understand, maintain, test
- **Fix**: Break into 5-10 focused functions
- **Prompt**: SRP Pattern
- **Time**: 4 hours

---

## 🎓 MINDSET SHIFT

| Before (Hard to Maintain) | After (Clean Code) |
|---------------------------|-------------------|
| "How do I add this feature?" | "How do I make this understandable?" |
| Code for compiler | Code for human |
| Minimize lines | Minimize cognitive load |
| One big function | Many small functions |
| Lock everything | Lock only what matters |
| Hide errors with "_" | Handle errors explicitly |

---

## ✅ READY TO START?

1. **Pick a component** (crew.go, agent.go, etc.)
2. **Read with checklist** (above)
3. **Identify top 3 issues**
4. **Select matching prompt** (from section above)
5. **Refactor**
6. **Measure** (run metrics)
7. **Done!**

---

**Print this card & keep nearby** 📌  
**Reference when stuck** 💪  
**Code quality improves** 📈
