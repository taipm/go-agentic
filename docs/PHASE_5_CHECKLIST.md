# Phase 5 Implementation Checklist

**Created:** 2025-12-25
**Updated:** 2025-12-25

---

## 🔴 CRITICAL - Agent Handoff Execution

### Task: Implement Agent Lookup and Recursive Execution
**File:** `core/workflow/execution.go` (Lines 193, 218)
**Effort:** 3-4 hours
**Priority:** CRITICAL

- [ ] **Analysis**
  - [ ] Review current `executeAgent()` function signature
  - [ ] Understand ExecutionContext structure
  - [ ] Map agent lookup requirements
  - [ ] Design recursive call pattern

- [ ] **Implementation**
  - [ ] Create agent lookup function in ExecutionContext
  - [ ] Modify routing decision handling to lookup agent
  - [ ] Implement recursive executeAgent call
  - [ ] Handle agent not found error
  - [ ] Update ExecutionContext state (CurrentAgent, HandoffCount)

- [ ] **Testing**
  - [ ] Unit test: Simple handoff (Agent A → Agent B)
  - [ ] Unit test: Multi-step handoff (A → B → C → Terminal)
  - [ ] Unit test: Invalid agent ID error
  - [ ] Unit test: Terminal agent stops execution
  - [ ] Integration test: Full workflow with 3 agents
  - [ ] Integration test: Signal routing with handoff

- [ ] **Documentation**
  - [ ] Update ExecutionContext documentation
  - [ ] Add code comment explaining handoff flow
  - [ ] Update routing module README

- [ ] **Review**
  - [ ] Code review checklist (see below)
  - [ ] Performance testing with long agent chains
  - [ ] Memory usage validation

---

## 🟠 HIGH - Behavior-Based Routing

### Task: Implement RouteByBehavior Function
**File:** `core/routing/signal.go` (Lines 36-51)
**Effort:** 1-2 hours
**Priority:** HIGH

- [ ] **Analysis**
  - [ ] Review AgentBehavior type definition
  - [ ] Understand behavior routing requirements
  - [ ] Check YAML parsing for behaviors

- [ ] **Implementation**
  - [ ] Implement behavior lookup logic
  - [ ] Add target agent ID retrieval
  - [ ] Implement error cases:
    - [ ] nil routing
    - [ ] empty behavior name
    - [ ] behavior not found
    - [ ] nil behavior config
    - [ ] empty target agent
  - [ ] Add comprehensive error messages

- [ ] **Testing**
  - [ ] Unit test: Valid behavior lookup
  - [ ] Unit test: Non-existent behavior
  - [ ] Unit test: Behavior with nil config
  - [ ] Unit test: Behavior with empty target
  - [ ] Unit test: nil routing parameter
  - [ ] Integration test: Behavior routing in workflow

- [ ] **Documentation**
  - [ ] Document behavior routing concept
  - [ ] Add examples to routing module README
  - [ ] Document expected YAML structure

---

## 🟠 HIGH - Parallel Group Execution

### Task: Implement Parallel Agent Execution
**Files:** `core/routing/parallel.go` (new), `core/common/types.go`
**Effort:** 4-6 hours
**Priority:** HIGH

- [ ] **Design**
  - [ ] Design parallel execution architecture
  - [ ] Plan concurrent goroutine pattern
  - [ ] Design result aggregation
  - [ ] Plan timeout handling
  - [ ] Design error handling strategy

- [ ] **Implementation**
  - [ ] Create `parallel.go` in routing module
  - [ ] Create `ParallelGroupExecutor` struct
  - [ ] Implement `Execute()` method:
    - [ ] Create result channels
    - [ ] Launch goroutines for each agent
    - [ ] Apply timeout
    - [ ] Collect results (WaitForAll vs FirstSuccess)
    - [ ] Handle partial failures
  - [ ] Add concurrent execution helper function

- [ ] **Testing**
  - [ ] Unit test: All agents succeed
  - [ ] Unit test: Some agents fail with WaitForAll=true
  - [ ] Unit test: FirstSuccess routing
  - [ ] Unit test: Timeout handling
  - [ ] Unit test: Concurrent execution correctness
  - [ ] Stress test: 10+ parallel agents
  - [ ] Integration test: Parallel group in workflow

- [ ] **Documentation**
  - [ ] Document parallel execution semantics
  - [ ] Add WaitForAll vs FirstSuccess explanation
  - [ ] Document timeout behavior

- [ ] **Review**
  - [ ] Race condition testing
  - [ ] Goroutine leak detection
  - [ ] Performance benchmarks

---

## 🟡 MEDIUM - Cost Tracking

### Task 1: Fix Cost Calculation
**File:** `core/common/types.go` (Lines 663-677)
**Effort:** 2-3 hours

- [ ] **Analysis**
  - [ ] Identify all calls to `CalculateCost()`
  - [ ] Determine input vs output token tracking needs
  - [ ] Review configured pricing fields
  - [ ] Check provider pricing references

- [ ] **Implementation**
  - [ ] Update `CalculateCost()` signature (add outputTokens param)
  - [ ] Use configured pricing from CostLimitsConfig
  - [ ] Add fallback to defaults
  - [ ] Separate input/output cost calculation
  - [ ] Update all callers with proper token counts

- [ ] **Testing**
  - [ ] Unit test: Input vs output cost calculation
  - [ ] Unit test: Uses configured pricing
  - [ ] Unit test: Fallback to defaults
  - [ ] Unit test: Zero tokens = zero cost
  - [ ] Integration test: Cost calculation in workflow

### Task 2: Add Per-Tool Cost Metrics
**Files:** `core/metrics.go`, `core/tools/executor.go`
**Effort:** 3-4 hours

- [ ] **Implementation**
  - [ ] Add ToolMetrics struct to metrics.go
  - [ ] Update AgentMetrics to include tool breakdown
  - [ ] Update MetricsCollector.RecordToolExecution()
  - [ ] Wire tool cost tracking into ExecuteTool()
  - [ ] Aggregate tool costs to agent metrics

- [ ] **Testing**
  - [ ] Unit test: Tool metric recording
  - [ ] Unit test: Cost aggregation
  - [ ] Integration test: Full workflow metrics

---

## 🟡 MEDIUM - Error Handling

### Task: Replace Generic Errors with Typed Errors
**Files:** Multiple (workflow, routing, execution)
**Effort:** 4-5 hours

- [ ] **Audit Phase**
  - [ ] List all `fmt.Errorf()` in workflow/*.go
  - [ ] List all `fmt.Errorf()` in routing/*.go
  - [ ] List all `fmt.Errorf()` in execution/*.go
  - [ ] Categorize errors by type

- [ ] **Migration**
  - [ ] Replace validation errors → ValidationError
  - [ ] Replace execution errors → ExecutionError
  - [ ] Replace timeout errors → TimeoutError
  - [ ] Replace configuration errors → ConfigurationError
  - [ ] Replace provider errors → ProviderError
  - [ ] Replace tool errors → ToolExecutionError

- [ ] **Testing**
  - [ ] Unit test: Error type creation
  - [ ] Unit test: Error message formatting
  - [ ] Unit test: Error wrapping with `fmt.Errorf("%w", err)`
  - [ ] Integration test: Error propagation

- [ ] **Documentation**
  - [ ] Document error types and when to use
  - [ ] Add error handling patterns to contributing guide

---

## 🟡 MEDIUM - Structured Logging

### Task: Add Request Correlation and Structured Logs
**Files:** `core/http.go`, `core/workflow/execution.go`, others
**Effort:** 5-6 hours

- [ ] **Setup**
  - [ ] Add slog to imports (or add zerolog dependency)
  - [ ] Create logger initialization in main
  - [ ] Design request ID generation
  - [ ] Plan request ID propagation through context

- [ ] **HTTP Handler Integration**
  - [ ] Generate request ID in HTTP handler
  - [ ] Add request ID to context
  - [ ] Pass logger with request ID to workflow
  - [ ] Log request start/completion

- [ ] **Workflow Logging**
  - [ ] Add structured agent execution logs
  - [ ] Add tool execution logs with timing
  - [ ] Add error logs with context
  - [ ] Add signal emission logs
  - [ ] Add handoff logs

- [ ] **Configuration**
  - [ ] Add log level configuration (DEBUG/INFO/WARN/ERROR)
  - [ ] Add JSON output format option
  - [ ] Add file output option (optional)

- [ ] **Testing**
  - [ ] Unit test: Logger creation
  - [ ] Integration test: Request correlation
  - [ ] Integration test: Log output format

---

## 🟡 MEDIUM - Validation Error Context

### Task: Improve Error Messages with File/Line Context
**Files:** `core/validation/*.go`, `core/common/errors.go`
**Effort:** 3-4 hours

- [ ] **Update ValidationError Structure**
  - [ ] Add FilePath field
  - [ ] Add LineNumber field
  - [ ] Add Suggestion field
  - [ ] Update Error() method

- [ ] **Integrate with Config Loader**
  - [ ] Track file path during YAML parsing
  - [ ] Track line numbers during parsing
  - [ ] Pass context to validation functions
  - [ ] Include in all ValidationError returns

- [ ] **Testing**
  - [ ] Unit test: Error message with file context
  - [ ] Unit test: Error message with suggestions
  - [ ] Integration test: Config validation errors

---

## 📋 Code Review Checklist

### For All Pull Requests

- [ ] **Correctness**
  - [ ] Logic is correct and handles edge cases
  - [ ] Error cases are properly handled
  - [ ] No race conditions or deadlocks
  - [ ] No goroutine leaks
  - [ ] No memory leaks

- [ ] **Testing**
  - [ ] Unit tests written and passing
  - [ ] Integration tests written and passing
  - [ ] Test coverage >= 80%
  - [ ] Edge cases tested
  - [ ] Error cases tested

- [ ] **Code Quality**
  - [ ] Follows Go conventions
  - [ ] No unused variables or imports
  - [ ] Comments explain why, not what
  - [ ] Function documentation present
  - [ ] Names are clear and descriptive

- [ ] **Performance**
  - [ ] No unnecessary allocations
  - [ ] No unnecessary goroutines
  - [ ] Reasonable algorithmic complexity
  - [ ] No performance regressions

- [ ] **Documentation**
  - [ ] README updated if needed
  - [ ] Code comments added
  - [ ] Function documentation present
  - [ ] Examples added to docs

- [ ] **Backward Compatibility**
  - [ ] No breaking API changes
  - [ ] Deprecation notices added if needed
  - [ ] Migration path documented

---

## 📊 Progress Tracking

### Sprint 1: Critical Routing
| Task | Status | Owner | Due | Notes |
|------|--------|-------|-----|-------|
| Agent handoff execution | ⬜ | - | 2025-12-29 | CRITICAL |
| Behavior-based routing | ⬜ | - | 2025-12-29 | HIGH |
| Tests for routing | ⬜ | - | 2025-12-29 | Must have |

### Sprint 2: Cost Tracking
| Task | Status | Owner | Due | Notes |
|------|--------|-------|-----|-------|
| Fix cost calculation | ⬜ | - | 2026-01-05 | Use configured pricing |
| Per-tool metrics | ⬜ | - | 2026-01-05 | Cost breakdown |
| Tests for metrics | ⬜ | - | 2026-01-05 | Must have |

### Sprint 3: Developer Experience
| Task | Status | Owner | Due | Notes |
|------|--------|-------|-----|-------|
| Typed error migration | ⬜ | - | 2026-01-12 | Systematic replacement |
| Structured logging | ⬜ | - | 2026-01-12 | With correlation IDs |
| Validation context | ⬜ | - | 2026-01-12 | File/line numbers |

---

## Questions for Planning

### Before Starting Sprint 1:
1. Who is the development owner for each task?
2. Should we use semver for API changes?
3. Do we want to add e2e tests?
4. What's the testing infrastructure like (CI/CD)?

### Before Starting Sprint 2:
1. How are input/output tokens currently tracked?
2. Should cost calculations be async?
3. Do we need historical cost data?

### Before Starting Sprint 3:
1. Should we use Go's slog or add zerolog?
2. What's the preferred log output (JSON/text)?
3. Should logging be configurable via env vars?

---

## Resources

- **Go Error Handling:** https://go.dev/blog/error-handling-and-go
- **Go Structured Logging:** https://go.dev/blog/slog
- **Go Concurrency:** https://go.dev/blog/pipelines
- **Go Context:** https://go.dev/blog/context

---

**Legend:**
- ⬜ Not Started
- 🔵 In Progress
- ✅ Complete
- ❌ Blocked

**Last Updated:** 2025-12-25
