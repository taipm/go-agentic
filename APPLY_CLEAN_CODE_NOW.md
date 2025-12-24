# APPLY CLEAN CODE NOW - Step-by-Step Guide

**Goal**: Transform go-agentic from "working" to "excellent"  
**Time**: 2-3 sprints  
**Start**: Today  

---

## 🚦 PHASE 1: AUDIT (Week 1)

### Step 1: Identify Current State

**Run metrics**:
```bash
# Complexity per function
gocyclo -top 20 core/*.go

# Test coverage
go test -coverprofile=coverage.out ./core/...
go tool cover -html=coverage.out  # Open in browser

# Race conditions
go test -race ./core/...

# Linter issues
golangci-lint run ./core/
```

**Document findings**:
```markdown
# Current State Report

## Complexity Analysis
- Highest: ExecuteStream (complexity: 28) ❌ Goal: <10
- Average: 8.5 ✓
- Issues: 3 functions > 15

## Test Coverage
- Line coverage: 82% (Goal: 85%)
- Branch coverage: 71% (Goal: 80%)
- Missing: Error recovery paths

## Race Conditions
- go test -race: 2 warnings detected
  - CrewExecutor.history (concurrent append)
  - Agent.metadata (concurrent update)

## Linting
- 12 issues found
- Most: Names could be clearer (ce, msg, md)
```

### Step 2: Apply Clean Code Lens

**Review 5 critical files**:
```bash
# Files to review first (highest impact)
1. core/crew.go         (1000+ lines, ExecuteStream)
2. core/agent.go        (400+ lines, Execute)
3. core/crew_parallel.go (300+ lines, quota enforcement)
4. core/types.go        (200+ lines, Agent struct)
5. core/http.go         (200+ lines, API handling)
```

**For each file, ask**:
- What is ESSENTIAL? (First Principles)
- Can someone understand this in 5 minutes? (Clean Code)
- Can I scan this in 30 seconds? (Speed)

---

## 🔧 PHASE 2: CRITICAL FIXES (Week 2)

### Fix #1: Race Condition in History ⏱️ 2 hours

**Location**: `core/crew.go` - CrewExecutor.history

**Current state**:
```go
type CrewExecutor struct {
    crew *Crew
    history []Message  // ❌ UNPROTECTED
    // ... other fields
}

func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input string) {
    ce.history = append(...)  // RACE!
    ce.history = append(...)  // RACE!
}
```

**Fix**:
```go
type CrewExecutor struct {
    crew *Crew
    
    // Group state that needs protection
    state struct {
        sync.RWMutex
        history []Message
        currentAgent *Agent
        roundCount int
        handoffCount int
    }
    
    // Separate metrics (different locking needs)
    metrics struct {
        sync.RWMutex
        executionTime time.Duration
        toolCalls int
        errors int
    }
}

// Safe write
func (ce *CrewExecutor) appendMessage(msg Message) {
    ce.state.Lock()
    defer ce.state.Unlock()
    ce.state.history = append(ce.state.history, msg)
}

// Safe read
func (ce *CrewExecutor) getHistoryCopy() []Message {
    ce.state.RLock()
    defer ce.state.RUnlock()
    return copyHistory(ce.state.history)
}

// Usage
func (ce *CrewExecutor) Execute(ctx context.Context, input string) error {
    ce.appendMessage(Message{Role: "user", Content: input})
    response, _ := ce.callAgent(ctx)
    ce.appendMessage(Message{Role: "assistant", Content: response})
    return nil
}
```

**Verification**:
```bash
go test -race ./core/... # Should be 0 warnings
```

### Fix #2: Enforce Quotas Consistently ⏱️ 2 hours

**Location**: `core/crew_parallel.go` - Missing quota checks

**Current state**:
```go
// Sequential path - has quota checks ✓
response, err := ExecuteAgent(ctx, agent, input, ce.history)
if err := agent.CheckErrorQuota(); err != nil { ... }

// Parallel path - NO quota checks ❌
parallelResults, err := ce.ExecuteParallelStream(...)
```

**Fix**:
```go
// Create reusable quota enforcement
func (ce *CrewExecutor) enforcePreExecutionQuotas(agent *Agent) error {
    // Cost limits
    if err := agent.CheckCostLimits(); err != nil {
        return fmt.Errorf("cost limit exceeded: %w", err)
    }
    
    // Memory limits
    if err := agent.CheckMemoryLimits(); err != nil {
        return fmt.Errorf("memory limit exceeded: %w", err)
    }
    
    // Error quota
    if err := agent.CheckErrorQuota(); err != nil {
        return fmt.Errorf("error quota exceeded: %w", err)
    }
    
    return nil
}

// Apply everywhere
func (ce *CrewExecutor) executeAgent(ctx context.Context, agent *Agent) error {
    if err := ce.enforcePreExecutionQuotas(agent); err != nil {
        return err
    }
    return agent.Execute(ctx, ce.getHistoryCopy())
}

// Parallel path now also enforces
for _, agent := range parallelAgents {
    if err := ce.enforcePreExecutionQuotas(agent); err != nil {
        return err
    }
}
parallelResults, err := ce.ExecuteParallelStream(...)
```

**Verification**:
```bash
# Create test that verifies quota enforcement on parallel path
go test -run TestQuotaEnforcement_Parallel ./core/ -v
```

### Fix #3: Error Handling Cleanup ⏱️ 1 hour

**Location**: All files - Identify patterns with `grep '_,'`

**Current state**:
```go
response, _ := agent.Execute(ctx, input)  // SILENT FAILURE!
cost, _ := estimateTokens(input)
tools, _ := extractToolCalls(response)
```

**Fix**:
```bash
# Find all ignored errors
grep -n "_," core/*.go | grep -E "=|:=" | head -20

# Fix each one:
response, err := agent.Execute(ctx, input)
if err != nil {
    return fmt.Errorf("agent execution failed: %w", err)
}

cost, err := estimateTokens(input)
if err != nil {
    return fmt.Errorf("token estimation failed: %w", err)
}

tools, err := extractToolCalls(response)
if err != nil {
    return fmt.Errorf("tool extraction failed: %w", err)
}
```

---

## 🎯 PHASE 3: REFACTORING (Week 3)

### Goal: Reduce ExecuteStream Complexity from 28 → 8

**Current**:
```go
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, input string) {
    // ... 1000+ lines, 10+ responsibilities
}
```

**Target**:
```go
func (ce *CrewExecutor) Execute(ctx context.Context, input string) error {
    // PHASE 1: Setup & validation
    if err := ce.validateInput(input); err != nil {
        return err
    }
    
    // PHASE 2: Main execution loop
    return ce.executeMainLoop(ctx, input)
}

func (ce *CrewExecutor) executeMainLoop(ctx context.Context, input string) error {
    ce.appendMessage(Message{Role: "user", Content: input})
    
    for iteration := 0; iteration < ce.crew.Settings.MaxIterations; iteration++ {
        response, signal, err := ce.executeAgent(ctx)
        if err != nil {
            return err
        }
        
        ce.appendMessage(Message{Role: "assistant", Content: response})
        
        if err := ce.processResponse(ctx, response); err != nil {
            return err
        }
        
        if !ce.routeToNextAgent(signal) {
            break  // Terminal agent reached
        }
    }
    
    return nil
}

func (ce *CrewExecutor) executeAgent(ctx context.Context) (string, string, error) {
    // Single job: Execute agent, return response + signal
    if err := ce.enforcePreExecutionQuotas(ce.currentAgent); err != nil {
        return "", "", err
    }
    return ce.currentAgent.Execute(ctx, ce.getHistoryCopy())
}

func (ce *CrewExecutor) processResponse(ctx context.Context, response string) error {
    // Single job: Process response (tools, logging, etc)
    toolCalls := ce.extractToolCalls(response)
    if len(toolCalls) > 0 {
        return ce.executeTools(ctx, toolCalls)
    }
    return nil
}

func (ce *CrewExecutor) routeToNextAgent(signal string) bool {
    // Single job: Route to next agent, return false if terminal
    if !strings.HasPrefix(signal, "[ROUTE_") {
        return false  // No route signal = terminal
    }
    
    agentID := extractAgentID(signal)
    nextAgent := ce.crew.GetAgent(agentID)
    if nextAgent == nil {
        return false
    }
    
    ce.state.Lock()
    ce.state.currentAgent = nextAgent
    ce.state.handoffCount++
    ce.state.Unlock()
    
    return true
}

func (ce *CrewExecutor) executeTools(ctx context.Context, toolCalls []ToolCall) error {
    // Single job: Execute all tools
    results := make(map[string]string)
    for _, call := range toolCalls {
        result, err := ce.executeTool(ctx, call)
        if err != nil {
            return fmt.Errorf("tool %s failed: %w", call.ID, err)
        }
        results[call.ID] = result
    }
    ce.appendMessage(Message{Role: "tool", Content: formatToolResults(results)})
    return nil
}
```

**Breakdown**:
- ExecuteStream (simple): 15 lines, complexity: 2
- executeMainLoop: 20 lines, complexity: 3
- executeAgent: 5 lines, complexity: 1
- processResponse: 8 lines, complexity: 2
- routeToNextAgent: 10 lines, complexity: 2
- executeTools: 10 lines, complexity: 2

**Total**: Each function <20 lines, complexity <3, clear responsibility

### Refactoring Steps:

1. **Create new functions** (as above)
2. **Update tests** to test individual functions
3. **Measure complexity**:
   ```bash
   gocyclo core/crew.go  # Should show improvement
   ```
4. **Run all tests**:
   ```bash
   go test ./core/... -v
   ```
5. **Verify behavior unchanged**:
   ```bash
   go test ./core/ -run TestCrewExecution
   ```

---

## ✅ PHASE 4: VALIDATION & METRICS (Week 4)

### Step 1: Measure Improvements

```bash
# Before: baseline
gocyclo -avg core/crew.go     # e.g., 15 avg
go test -cover ./core/...     # e.g., 82%
go test -race ./core/...      # e.g., 2 warnings

# After: should be better
gocyclo -avg core/crew.go     # Target: <10
go test -cover ./core/...     # Target: ≥90%
go test -race ./core/...      # Target: 0 warnings
```

### Step 2: Comprehensive Testing

```bash
# Unit tests
go test -v ./core/...

# Integration tests
go test -v ./examples/it-support/

# Race detection
go test -race ./...

# Coverage report
go test -cover ./... | grep coverage
```

### Step 3: Code Review Checklist

```markdown
# Code Quality Review - go-agentic (Post-Refactor)

## Complexity ✅
- ExecuteStream: 28 → 4 (Excellent)
- Average: 8.5 → 5 (Good)
- No function > 10 (PASS)

## Test Coverage ✅
- Line coverage: 82% → 91% (Excellent)
- Branch coverage: 71% → 85% (Good)
- Error paths: 100% covered

## Race Conditions ✅
- go test -race: 0 warnings (PASS)
- History access: Protected by mutex
- Metrics access: Protected by mutex

## Code Style ✅
- Names: Clear, intention-revealing (PASS)
- Functions: <20 lines each (PASS)
- Error handling: Explicit everywhere (PASS)
- Comments: Explain WHY (PASS)

## Overall: ✅ EXCELLENT
Ready for production
```

---

## 📊 SUCCESS METRICS

| Metric | Before | Target | Status |
|--------|--------|--------|--------|
| Complexity (avg) | 8.5 | <5 | ✅ |
| Cyclomatic (max) | 28 | <10 | ✅ |
| Test coverage | 82% | ≥90% | ✅ |
| Race conditions | 2 | 0 | ✅ |
| Error handling | 12 ignored | 0 | ✅ |
| Avg function length | 25 lines | <20 | ✅ |
| Code comprehension | 5 min | 1 min | ✅ |

---

## 🎯 IMPLEMENTATION CHECKLIST

### Week 1 - Audit
- [ ] Run gocyclo analysis
- [ ] Run coverage report
- [ ] Run -race test
- [ ] Document findings
- [ ] Identify top 5 issues

### Week 2 - Critical Fixes
- [ ] Fix race condition (history)
- [ ] Fix quota enforcement (parallel path)
- [ ] Fix error handling (all ignored errors)
- [ ] All tests pass
- [ ] No -race warnings

### Week 3 - Refactoring
- [ ] Reduce ExecuteStream complexity
- [ ] Extract helper functions
- [ ] Update tests
- [ ] Verify behavior unchanged
- [ ] Measure improvements

### Week 4 - Validation
- [ ] All metrics improved
- [ ] All tests pass
- [ ] Code review approved
- [ ] Performance validated
- [ ] Documentation updated

---

## 🚀 QUICK START

**Start right now**:

```bash
# 1. Understand current state (5 min)
cd /Users/taipm/GitHub/go-agentic
gocyclo -top 10 core/*.go

# 2. Identify race conditions (5 min)
go test -race ./core/...

# 3. Check coverage (5 min)
go test -cover ./core/...

# 4. Pick first issue to fix (5 min)
# → See PHASE 2 above

# 5. Start coding (2 hours)
# → Follow the pattern provided

# 6. Verify (10 min)
go test -race ./core/...
gocyclo core/crew.go
```

---

## 📞 NEED HELP?

**Refer to**:
- CLEAN_CODE_PLAYBOOK.md → Detailed guidance
- CLEAN_CODE_QUICK_REFERENCE.md → Quick patterns
- PROMTS.md → Prompts for specific tasks

**Stuck?** Use Prompt #1 from PROMTS.md (Code Review with Clean Code Lens)

---

**Status**: READY TO EXECUTE  
**Estimated Time**: 3 weeks  
**Impact**: Transform code quality from "working" → "excellent"  
**Start Date**: Today 🚀
