# CLEAN CODE PLAYBOOK for go-agentic

**Framework**: First Principles + Clean Code Principles + Speed of Execution  
**Date**: 2025-12-23  
**Methodology**: Real Code Analysis + Actionable Patterns

---

## I. TƯ DUY NÀO PHỤC VỤ CLEAN CODE TỐTNHẤT?

### 🎯 **3 TƯ DUY CHÍNH**

#### **1. FIRST PRINCIPLES THINKING (Elon Musk)**
**Câu hỏi cốt lõi**: "Tại sao code hiện tại viết như vậy?"

**Áp dụng cho Clean Code**:
- 🔥 **Essential vs. Accidental**: Phân biệt code cần thiết vs. complex không cần thiết
- 🔥 **Fundamental Building Blocks**: Build từ abstractions nhỏ nhất (functions, interfaces)
- 🔥 **Challenge Assumptions**: Tại sao biến này là global? Tại sao function này dài 500 dòng?
- 🔥 **Measure Everything**: Metrics: code coverage, cyclomatic complexity, cognitive load

**Trong go-agentic**:
```go
// ❌ BEFORE (Complex, Hard to Reason)
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input string) {
    ce.history = append(ce.history, Message{...})  // NO LOCK
    response, _ := agent.Execute(ctx, input, ce.history)  // RACE CONDITION
    ce.history = append(ce.history, Message{...})  // NO LOCK
}

// ✅ AFTER (First Principles - Essential Functions)
// Question: What is ESSENTIAL here?
// Answer: Add message safely + Execute agent + Emit signal

// Separate concerns:
func (ce *CrewExecutor) appendMessageSafe(msg Message) {
    ce.historyMutex.Lock()
    defer ce.historyMutex.Unlock()
    ce.history = append(ce.history, msg)
}

func (ce *CrewExecutor) executeAgent(ctx context.Context, agent *Agent) (*string, error) {
    // Single responsibility: execute, no mutation
    return agent.Execute(ctx, ce.getHistoryCopy())
}

func (ce *CrewExecutor) executeWithSignalRouting(ctx context.Context, input string) {
    ce.appendMessageSafe(Message{Role: "user", Content: input})
    response, signal := ce.executeAgent(ctx, ce.currentAgent)
    ce.appendMessageSafe(Message{Role: "assistant", Content: response})
    ce.routeBasedOnSignal(signal)
}
```

---

#### **2. CLEAN CODE PRINCIPLES (Robert C. Martin)**
**Nội dung**: Code được viết cho **người đọc**, không máy

**6 Nguyên Tắc Chính**:

##### **A. NAMES - Tên phải nói lên chủ ý**
```go
// ❌ BAD
ce.hist []Message           // What's "hist"? history? histology?
msg.r string                // "r" means what?
agent.md *AgentMetadata     // Too abbreviated

// ✅ GOOD
ce.conversationHistory []Message
msg.role string
agent.metadata *AgentMetadata
```

##### **B. FUNCTIONS - Function phải làm 1 thứ duy nhất, làm tốt**
```go
// ❌ BAD - Does too many things
func (ce *CrewExecutor) ProcessInput(ctx context.Context, input string) (string, error) {
    // 1. Add to history
    ce.history = append(ce.history, ...)
    // 2. Execute agent
    response, _ := ExecuteAgent(...)
    // 3. Process tools
    for toolCall := range toolCalls { ... }
    // 4. Update metrics
    ce.metrics.Update(...)
    // 5. Route signal
    nextAgent := ce.route(signal)
    // 6. Check quotas
    if ce.metrics.Cost > ce.crew.MaxCost { ... }
    return response, nil
}

// ✅ GOOD - Single responsibility
func (ce *CrewExecutor) addUserMessage(input string) {
    ce.historyMutex.Lock()
    defer ce.historyMutex.Unlock()
    ce.history = append(ce.history, Message{Role: "user", Content: input})
}

func (ce *CrewExecutor) callAgentModel(ctx context.Context, agent *Agent) (string, error) {
    return agent.Execute(ctx, ce.getHistoryCopy())
}

func (ce *CrewExecutor) processToolCalls(ctx context.Context, toolCalls []ToolCall) (map[string]string, error) {
    results := make(map[string]string)
    for _, call := range toolCalls {
        result, err := ce.executeTool(ctx, call)
        results[call.ID] = result
    }
    return results, nil
}

func (ce *CrewExecutor) enforceQuotas() error {
    if ce.metrics.DailyCost > ce.crew.Settings.MaxDailyBudget {
        return fmt.Errorf("daily cost limit exceeded")
    }
    return nil
}
```

##### **C. COMMENTS - Code tự nói, comments giải thích WHY**
```go
// ❌ BAD
// Add message to history
ce.history = append(ce.history, msg)

// ✅ GOOD
// Add user message to conversation history for context preservation
// Requires lock to prevent race conditions in concurrent execution
ce.historyMutex.Lock()
ce.history = append(ce.history, msg)
ce.historyMutex.Unlock()

// ❌ BAD
return errors.New("invalid signal")

// ✅ GOOD
// Signal must match [ROUTE_XXXXX] pattern for routing to work
// Return error if pattern doesn't match to prevent misconfiguration
return fmt.Errorf("invalid signal format: expected [ROUTE_XXXXX], got %s", signal)
```

##### **D. ERRORS - Error handling phải explicit, not implicit**
```go
// ❌ BAD
response, _ := agent.Execute(ctx, history)  // Ignore error!
cost, _ := estimateTokens(input)             // What if fails?

// ✅ GOOD
response, err := agent.Execute(ctx, history)
if err != nil {
    log.Printf("agent execution failed: %v", err)
    return "", fmt.Errorf("executing agent %s: %w", agent.ID, err)
}

cost, err := estimateTokens(input)
if err != nil {
    return 0, fmt.Errorf("estimating tokens: %w", err)
}
```

##### **E. STRUCTURE - Organize code by concern, not by location**
```go
// ❌ BAD - Mixed concerns
type CrewExecutor struct {
    crew *Crew
    history []Message
    currentAgent *Agent
    roundCount int
    handoffCount int
    metrics *AgentMetrics
    costMetrics *CostMetrics
    memoryMetrics *MemoryMetrics
    mutex sync.RWMutex
    apiKey string
    providers map[string]LLMProvider
    tools map[string]Tool
    // ... 20 more fields
}

// ✅ GOOD - Grouped by responsibility
type ExecutionState struct {
    conversationHistory []Message
    currentAgent        *Agent
    roundCount          int
    handoffCount        int
    stateMutex          sync.RWMutex
}

type ExecutionMetrics struct {
    Cost              *CostMetrics
    Memory            *MemoryMetrics
    Performance       *PerformanceMetrics
}

type ExecutionContext struct {
    crew              *Crew
    state             *ExecutionState
    metrics           *ExecutionMetrics
    providers         *ProviderFactory
    tools             *ToolRegistry
}

type CrewExecutor struct {
    context *ExecutionContext
}
```

##### **F. ABSTRACTION - Hide complexity, expose interface**
```go
// ❌ BAD - Implementation details exposed
type Agent struct {
    Name        string
    Role        string
    // ... 40 fields
    metadata    *AgentMetadata
    costMetrics *AgentCostMetrics
    memoryMetrics *AgentMemoryMetrics
}

// User must know about internal fields:
agent.metadata.Update(...)
agent.costMetrics.Track(cost)

// ✅ GOOD - Clean interface
type Agent interface {
    ID() string
    Role() string
    Execute(ctx context.Context, input string) (response string, signal string, err error)
    UpdateMetrics(metrics ExecutionMetrics)
}

type agentImpl struct {
    id       string
    role     string
    internal *agentInternal  // Hide all complexity here
}

func (a *agentImpl) Execute(ctx context.Context, input string) (string, string, error) {
    // Caller doesn't know about metadata, cost tracking, memory limits
    // All hidden inside Execute()
    return a.internal.execute(ctx, input)
}
```

---

#### **3. EXECUTION SPEED THINKING (NVIDIA - Parallel Mindset)**
**Nguyên tắc**: Code nhanh để hiểu = Code nhanh để chỉnh sửa

**3 Quy Tắc**:
1. **Locality** - Code liên quan phải gần nhau
2. **Consistency** - Pattern giống nhau = Easy to scan
3. **Obviousness** - Điều gì sẽ xảy ra phải hiển nhiên

**Ví dụ**:
```go
// ❌ SLOW TO UNDERSTAND
func (ce *CrewExecutor) Execute(ctx context.Context, input string) error {
    // Add to history (line 50)
    ce.history = append(...)
    
    // ... 200 lines later ...
    
    // Check quota (line 250) - wait, where's the lock?
    if ce.metrics.Cost > ce.crew.MaxCost {
        return errors.New("over budget")
    }
    
    // ... 100 lines later ...
    
    // Update history again (line 350) - is this thread-safe?
    ce.history = append(...)
}

// ✅ FAST TO UNDERSTAND - Group by responsibility
func (ce *CrewExecutor) Execute(ctx context.Context, input string) error {
    // PHASE 1: Input Validation & Safety
    if err := ce.validateInput(input); err != nil {
        return err
    }
    
    // PHASE 2: Quota Checks (all in one place)
    if err := ce.enforceQuotaLimits(); err != nil {
        return err
    }
    
    // PHASE 3: Message Management (thread-safe operations)
    ce.addUserMessage(input)
    defer func() { ce.recordExecution() }()
    
    // PHASE 4: Agent Execution
    response, signal, err := ce.executeCurrentAgent(ctx)
    if err != nil {
        return err
    }
    ce.addAssistantMessage(response)
    
    // PHASE 5: Post-Processing
    if err := ce.processToolCalls(ctx, signal); err != nil {
        return err
    }
    
    // PHASE 6: Routing
    ce.routeToNextAgent(signal)
    
    return nil
}
```

---

## II. PROMPT SPECIFIC CHO CLEAN CODE

### **PROMPT #1: Code Review với Clean Code Lens**

```prompt
Thực hiện code review cho [FILE_NAME] theo CLEAN CODE PRINCIPLES:

STRUCTURE (5 min):
1. Names
   - Biến/function names có nói lên chủ ý không?
   - Có abbreviation không hiểu được không?
   - Tên có phù hợp với responsibility không?

2. Functions
   - Mỗi function làm mấy thứ? (ideal = 1 thứ)
   - Function dài bao nhiêu dòng? (ideal = <20 dòng)
   - Cyclomatic complexity là gì? (ideal = <5)

3. Comments
   - Có comment giải thích code không? (BAD - code tự nói)
   - Comments giải thích WHY chứ không phải WHAT? (GOOD)
   - Có stale/outdated comments không?

4. Error Handling
   - Có ignored errors ("_" assignment)?
   - Error messages descriptive không?
   - Error handling logic rõ ràng không?

5. Structure
   - Các concern có được group không?
   - Có god objects không? (>30 fields)
   - Dependencies có clear không?

SEVERITY LEVELS:
- 🔴 RED: Race condition, logic error, memory leak
- 🟡 YELLOW: Readability, naming, structure
- 🟢 GREEN: Nice-to-have improvements

OUTPUT:
For each issue, provide:
- Location (file:line)
- Severity level
- Problem explanation
- Before/After code example
- Why this matters for maintenance
```

### **PROMPT #2: Refactor Function từ Complex → Simple**

```prompt
Refactor [FUNCTION_NAME] để đạt Clean Code standards:

CURRENT STATE:
- Lines: [N]
- Responsibilities: [list]
- Complexity: [cyclomatic complexity]
- Issues: [what's wrong]

TARGET STATE:
- Lines: <20 each
- Responsibilities: 1 per function
- Complexity: <5
- Pattern: Single Responsibility Principle

APPROACH:
1. Identify all responsibilities in function
2. Extract each into separate function
3. Create coordinator function that calls them
4. Remove duplication
5. Add clear error handling

REFACTORING CHECKLIST:
- ☐ Each extracted function does ONE thing
- ☐ Function names clearly describe purpose
- ☐ No hidden side effects
- ☐ Error handling explicit
- ☐ Original behavior preserved
- ☐ Tests still pass

EXAMPLE PATTERN:
// Before: 1 function, 50 lines, 5 responsibilities
func ProcessRequest(input string) error {
    // 1. Validate
    // 2. Transform
    // 3. Execute
    // 4. Store result
    // 5. Return response
}

// After: 5 functions, 5-10 lines each
func ProcessRequest(input string) error {
    if err := validate(input); err != nil { return err }
    transformed := transform(input)
    result, err := execute(transformed)
    if err != nil { return err }
    if err := store(result); err != nil { return err }
    return respond(result)
}

func validate(input string) error { ... }      // 1 thing
func transform(input string) string { ... }    // 1 thing
func execute(data string) (string, error) { ...} // 1 thing
func store(result string) error { ... }        // 1 thing
func respond(result string) error { ... }      // 1 thing
```

### **PROMPT #3: Extract Interface untuk Hide Complexity**

```prompt
Create interface untuk [STRUCT_NAME] để hide internal complexity:

CURRENT STRUCT:
```go
type [Struct] struct {
    // [20+ fields]
}

// Public methods that expose internals:
func (s *[Struct]) GetInternal() { ... }
func (s *[Struct]) SetInternal() { ... }
```

GOAL:
- Expose only essential behavior
- Hide all implementation details
- Make it easy to mock for testing

INTERFACE EXTRACTION STEPS:

1. Identify ALL public methods
2. Classify into:
   - ESSENTIAL: Core functionality (keep in interface)
   - INTERNAL: Implementation detail (hide behind interface)

3. Create interface with ESSENTIAL methods only
4. Keep INTERNAL methods on struct
5. Users depend on interface, not concrete struct

TEMPLATE:
```go
// Interface - what users depend on
type [Actor] interface {
    [Essential Method 1](ctx context.Context, args ...) (result, error)
    [Essential Method 2](ctx context.Context, args ...) (result, error)
    // ... only 3-5 methods max
}

// Implementation - implementation details hidden
type [struct]Impl struct {
    // Private fields - users don't know/care
    internalState *state
    internalTools *toolSet
}

func (impl *[struct]Impl) EssentialMethod1(ctx context.Context, args) (result, error) {
    // Can use all private fields and methods
    impl.internalHelper()
    return impl.computeResult(args)
}

// Internal helper - NOT in interface, NOT visible to users
func (impl *[struct]Impl) internalHelper() { ... }
func (impl *[struct]Impl) computeResult(args) result { ... }
```

WHY THIS MATTERS:
- Mocking becomes trivial (mock interface, not 50-field struct)
- Maintainability improves (can refactor internals freely)
- API surface stays stable (interface changes rarely)
- Cognitive load reduces (users see only what they need)
```

### **PROMPT #4: Add Mutex Correctly để Prevent Concurrency Bugs**

```prompt
Secure [COMPONENT] against race conditions:

ANALYSIS:
- What shared state exists?
- What concurrent access patterns?
- What can go wrong?

FIXING STRATEGY:

Option 1: Fine-Grained Locks (Preferred)
```go
type Component struct {
    // Separate data by access pattern
    state struct {
        sync.RWMutex
        data []Item
    }
    
    metrics struct {
        sync.RWMutex
        count int
        cost  float64
    }
}

func (c *Component) ReadState() []Item {
    c.state.RLock()
    defer c.state.RUnlock()
    return copyOf(c.state.data)  // Return copy, not reference
}

func (c *Component) UpdateState(item Item) {
    c.state.Lock()
    defer c.state.Unlock()
    c.state.data = append(c.state.data, item)
}

// Metrics update separate from state update
func (c *Component) RecordMetric(cost float64) {
    c.metrics.Lock()
    defer c.metrics.Unlock()
    c.metrics.cost += cost
    c.metrics.count++
}
```

Option 2: Coarse-Grained Lock (Simpler, Slower)
```go
type Component struct {
    mu    sync.RWMutex  // Protects everything below
    state []Item
    count int
    cost  float64
}

func (c *Component) Execute() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    // All operations atomic
    c.state = append(c.state, item)
    c.cost += newCost
    return nil
}
```

CHECKLIST:
- ☐ Identify what needs protection (state, metrics)
- ☐ Choose lock granularity (fine vs coarse)
- ☐ Use defer unlock() to prevent deadlock
- ☐ Never hold lock across I/O operations
- ☐ Copy data when returning (don't expose references)
- ☐ Test with -race flag: go test -race

TESTING:
```go
func TestRaceCondition(t *testing.T) {
    c := NewComponent()
    
    // Concurrent reads should not panic
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _ = c.ReadState()
        }()
    }
    
    // Concurrent writes should not panic
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            c.UpdateState(Item{Index: idx})
        }(i)
    }
    
    wg.Wait()
}

// Run: go test -race ./...
```
```

### **PROMPT #5: Metric untuk Measure Code Quality**

```prompt
Measure code quality untuk [COMPONENT]:

METRICS TO TRACK:

1. COMPLEXITY (Cyclomatic)
   ```
   go get github.com/fzipp/gocyclo
   gocyclo -top 10 [file].go
   ```
   - Goal: All functions < 10 complexity
   - Alert: Any function > 15

2. CODE COVERAGE
   ```
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```
   - Goal: ≥ 85% coverage
   - Must have: All error paths tested

3. NAMING QUALITY
   ```
   // Length analysis
   grep -oE '\b[a-z]{1,2}\b' [file].go | sort | uniq -c | sort -rn
   // Find 1-2 letter variables (potential problems)
   ```
   - Goal: No single-letter variables (except loop indices)
   - Exception: i, j, k (loop indices only)

4. FUNCTION LENGTH
   ```
   wc -l [file].go | awk '{print $1}'
   // Divide by number of functions
   ```
   - Goal: Average < 20 lines per function
   - Max: 50 lines for any single function

5. DEPENDENCY ANALYSIS
   ```
   go mod graph | head -20
   ```
   - Goal: No circular dependencies
   - Alert: Deep dependency chains (>5 levels)

REPORTING TEMPLATE:
```
Code Quality Report: [Component]
================================

Complexity:
  ✓ 5 functions > 10 complexity (target: 0)
  ✓ Highest: execute() at 18

Coverage:
  ✓ Line coverage: 82%
  ✓ Missing: error recovery paths (3 cases)

Naming:
  ✗ Found 12 single-letter variables
    - ce (should be CrewExecutor?)
    - msg (should be Message?)

Function Length:
  ✓ Average: 18 lines (target: 20)
  ✓ Longest: 45 lines (acceptable)

Recommendations:
  1. Reduce execute() complexity from 18 → <10 (extract 2 functions)
  2. Add tests for error recovery paths (increase coverage to 90%)
  3. Rename abbreviated variables for clarity
```
```

---

## III. CLEAN CODE PATTERNS TRONG GO-AGENTIC

### **Pattern #1: Safe History Management**

```go
// ❌ BEFORE (Race condition)
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input string) {
    ce.history = append(ce.history, Message{...})  // NO LOCK!
    ce.history = append(ce.history, Message{...})
}

// ✅ AFTER (Clean)
type CrewExecutor struct {
    state struct {
        sync.RWMutex
        history []Message
    }
}

// Safe read
func (ce *CrewExecutor) getHistoryCopy() []Message {
    ce.state.RLock()
    defer ce.state.RUnlock()
    return copyHistory(ce.state.history)  // Return copy, not reference
}

// Safe write
func (ce *CrewExecutor) appendMessage(msg Message) {
    ce.state.Lock()
    defer ce.state.Unlock()
    ce.state.history = append(ce.state.history, msg)
}

// Usage
func (ce *CrewExecutor) Execute(ctx context.Context, input string) error {
    ce.appendMessage(Message{Role: "user", Content: input})
    response, _ := ce.callAgent(ctx)
    ce.appendMessage(Message{Role: "assistant", Content: response})
    return nil
}
```

### **Pattern #2: Clear Error Handling**

```go
// ❌ BEFORE (Silent failures)
response, _ := agent.Execute(ctx, input)
cost, _ := estimateTokens(input)

// ✅ AFTER (Explicit)
response, err := agent.Execute(ctx, input)
if err != nil {
    return fmt.Errorf("executing agent %s: %w", agent.ID, err)
}

cost, err := estimateTokens(input)
if err != nil {
    return fmt.Errorf("estimating cost: %w", err)
}
```

### **Pattern #3: Quota Enforcement Consistency**

```go
// ✅ PATTERN - Apply everywhere
func (ce *CrewExecutor) executeAgent(ctx context.Context, agent *Agent) error {
    // STEP 1: Pre-flight checks (always first)
    if err := agent.CheckCostLimits(ctx); err != nil {
        return err
    }
    if err := agent.CheckMemoryLimits(); err != nil {
        return err
    }
    
    // STEP 2: Execute
    response, _ := agent.Execute(ctx, ce.getHistoryCopy())
    
    // STEP 3: Post-execution checks
    if err := agent.UpdateMetrics(response); err != nil {
        return err
    }
    
    return nil
}

// Apply SAME pattern everywhere:
// - Sequential execution path
// - Parallel execution path
// - Tool execution
// - Any external LLM call
```

### **Pattern #4: Single Responsibility Principle**

```go
// ❌ BEFORE - One God Function
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input string) {
    // Add to history
    ce.history = append(...)
    
    // Execute agent
    response, _ := ce.callAgent()
    
    // Process tools
    for tc := range toolCalls { ... }
    
    // Update metrics
    ce.updateMetrics()
    
    // Route
    nextAgent := ce.route(signal)
    
    // Check quotas
    if ce.exceedsQuota() { ... }
    
    // Stream output
    ce.stream(response)
}

// ✅ AFTER - Each function one job
func (ce *CrewExecutor) Execute(ctx context.Context, input string) error {
    if err := ce.validateInput(input); err != nil {
        return err
    }
    
    if err := ce.enforceQuotas(); err != nil {
        return err
    }
    
    if err := ce.executeMainLoop(ctx, input); err != nil {
        return err
    }
    
    return nil
}

func (ce *CrewExecutor) executeMainLoop(ctx context.Context, input string) error {
    ce.addUserMessage(input)
    
    for !ce.shouldTerminate() {
        response, signal, err := ce.executeAgent(ctx)
        if err != nil {
            return err
        }
        
        ce.addAssistantMessage(response)
        
        if toolCalls := ce.extractToolCalls(response); len(toolCalls) > 0 {
            if err := ce.executeTools(ctx, toolCalls); err != nil {
                return err
            }
        }
        
        nextAgent := ce.routeBySignal(signal)
        if nextAgent == nil {
            break
        }
        
        ce.moveToAgent(nextAgent)
    }
    
    return nil
}

func (ce *CrewExecutor) executeAgent(ctx context.Context) (string, string, error) {
    // Single job: call agent, return response + signal
    return ce.currentAgent.Execute(ctx, ce.getHistoryCopy())
}

func (ce *CrewExecutor) executeTools(ctx context.Context, toolCalls []ToolCall) error {
    // Single job: execute tools, return results
    for _, call := range toolCalls {
        _, err := ce.executeTool(ctx, call)
        if err != nil {
            return fmt.Errorf("tool %s failed: %w", call.ID, err)
        }
    }
    return nil
}
```

---

## IV. CHECKLIST - APPLY CLEAN CODE PRINCIPLES

### **Before Submit, Check:**

- [ ] **Names**
  - [ ] No single-letter variables (except i, j, k)
  - [ ] Function names describe what they do
  - [ ] Struct field names are clear

- [ ] **Functions**
  - [ ] Each function does ONE thing
  - [ ] Average line count < 20
  - [ ] Cyclomatic complexity < 10
  - [ ] Error handling explicit

- [ ] **Structure**
  - [ ] Related code grouped together
  - [ ] No god objects (>20 fields)
  - [ ] Dependencies clear

- [ ] **Concurrency**
  - [ ] All shared state protected by mutex
  - [ ] No race condition warnings: `go test -race`
  - [ ] Locks held as short as possible

- [ ] **Testing**
  - [ ] ≥85% code coverage
  - [ ] All error paths tested
  - [ ] Concurrency tested

- [ ] **Comments**
  - [ ] No obvious comments ("add item to list")
  - [ ] Complex logic explained (WHY, not WHAT)
  - [ ] No stale/outdated comments

---

## V. FIRST PRINCIPLES + CLEAN CODE = GO-AGENTIC EXCELLENCE

### **Synthesis - Tổng Hợp**

```
First Principles (Essential) → Clean Code (Expression) → Speed (Execution)

Example: Fix Race Condition in History

FIRST PRINCIPLES:
  "What is ESSENTIAL?"
  → Managing mutable shared state (history)
  "What can go wrong?"
  → Concurrent access, lost writes, panic
  "What's minimum fix?"
  → Protect with mutex

CLEAN CODE:
  "How to express it clearly?"
  → Separate concerns (state access from business logic)
  → Clear function names: addMessage, getHistoryCopy, appendSafe
  → Obvious pattern: Lock → Modify → Unlock

SPEED:
  "How to scan & understand?"
  → All history access in 1 place
  → Clear lock pattern (always: Lock defer Unlock)
  → No hidden side effects

Result: Clear + Safe + Fast to understand code
```

---

## VI. EXECUTION ROADMAP

### **Phase 1: Code Review (Week 1)**
Apply Prompt #1 to identify all issues
- [ ] Run cyclomatic complexity analysis
- [ ] Run test coverage report
- [ ] Identify race conditions: `go test -race`
- [ ] List all issues by severity

### **Phase 2: Critical Fixes (Week 2)**
Fix 🔴 RED issues (race conditions, logic errors)
- [ ] Add mutex to CrewExecutor.history
- [ ] Add quota enforcement to parallel path
- [ ] Fix error handling (no more ignored errors)

### **Phase 3: Refactoring (Week 3)**
Apply Prompt #2 to simplify functions
- [ ] Reduce ExecuteStream complexity
- [ ] Break into single-responsibility functions
- [ ] Extract interfaces to hide complexity

### **Phase 4: Testing & Validation (Week 4)**
Verify quality improvements
- [ ] Increase coverage to ≥90%
- [ ] All -race tests pass
- [ ] All functions < 20 lines
- [ ] Complexity < 10

---

## VII. TƯ DUY SUMMARY

| Tư Duy | Câu Hỏi | Áp Dụng Clean Code | Kết Quả |
|--------|--------|-------------------|---------|
| **First Principles** | "Cái gì essential?" | Identify what to refactor | Clear, minimal code |
| **Clean Code** | "Người sẽ hiểu không?" | Express intent clearly | Maintainable code |
| **Speed of Execution** | "Hiểu nhanh được không?" | Group by concern | Fast to modify |

**Kết luận**: Combine cả 3 tư duy → **Code Excellence**

---

**Status**: READY FOR IMPLEMENTATION  
**Last Updated**: 2025-12-23  
**Next Step**: Apply Prompt #1 (Code Review) to identify issues
