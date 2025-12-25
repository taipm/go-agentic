# Kế Hoạch Hành Động - Cải Thiện go-agentic

## Mục Đích (Dual Focus)

### Primary: Fix Critical Issues
Khắc phục 3 vấn đề nghiêm trọng gây ra vòng lặp vô hạn trong quiz example.

### Secondary: Improve Developer Experience (DX)
Cải thiện trải nghiệm phát triển từ **6.5/10 → 8.5+/10** bằng cách:
- Adopt best practices từ Anthropic SDK, LangChain, FastAPI
- Loại bỏ boilerplate code
- Auto-generate schemas từ Go structs
- Fail-fast validation
- Clear error messages

---

## GIAI ĐOẠN 1: CẶP ĐÔI (Tháng 1, Tuần 1-2)

### Tác vụ 1.1: Định Nghĩa State Persistence Contract

**Cái gì:**
- Định nghĩa interface cho state manager
- Xác định contract giữa workflow vs domain state
- Thiết kế atomic update semantics

**File cần tạo:**
```
core/state-management/
├─ contracts.go          (new)
│  ├─ StateManager interface
│  ├─ StateSnapshot struct
│  └─ StateUpdate interface
├─ execution_state.go    (modify)
│  └─ Add DomainState, StateHistory
└─ state_persistence.go  (new)
   └─ Persistence implementation
```

**Acceptance Criteria:**
- [ ] StateManager interface defined with 5 methods minimum
- [ ] Atomic update semantics guaranteed
- [ ] State snapshots recorded per round
- [ ] Tests pass for state CRUD operations

---

### Tác vụ 1.2: Integrate Tool Results into Workflow

**Cái gì:**
- Move tool execution from agent layer to workflow layer  
- Capture tool results as structured data
- Add tool results to ExecutionContext.History

**File cần sửa:**
```
core/workflow/
├─ execution.go         (modify)
│  ├─ Add ToolOrchestration() method
│  ├─ Execute tools after agent response
│  └─ Integrate results to history
└─ (no new files)

core/tools/
├─ executor.go          (modify)
│  └─ Create ExecuteToolCallsWithContext()
└─ orchestration.go     (new)
   ├─ ToolExecution struct
   └─ Tool result formatting
```

**Acceptance Criteria:**
- [ ] Tools executed from workflow layer
- [ ] Tool results appended to History
- [ ] Results accessible to next agent round
- [ ] Tool errors logged but don't stop workflow

---

### Tác vụ 1.3: Implement Signal-State Atomicity

**Cái gì:**
- Create atomic signal + state update operations
- Verify state was updated before considering signal success
- Rollback on verification failure

**File cần tạo:**
```
core/signal/
├─ atomic.go            (new)
│  ├─ AtomicSignal() function
│  └─ SignalWithStateVerification()
└─ verification.go      (new)
   └─ VerifySignalEffect()
```

**Acceptance Criteria:**
- [ ] Atomic operations cannot be split
- [ ] State verified before signal success
- [ ] Rollback on verification failure
- [ ] No race conditions in tests

---

## GIAI ĐOẠN 2: CÁC VẤN ĐỀ CHÍNH (Tháng 1, Tuần 3-4)

### Tác vụ 2.1: Implement Termination Logic

**Cái gì:**
- Domain-aware termination checking
- State-based vs signal-based termination
- Prevent infinite loops

**File cần tạo:**
```
core/termination/
├─ checker.go           (new)
│  ├─ TerminationChecker interface
│  └─ CheckTermination() function
└─ strategies.go        (new)
   ├─ StateBasedTermination
   └─ SignalBasedTermination
```

**Acceptance Criteria:**
- [ ] Quiz example detects completion (remaining=0)
- [ ] Terminal signal [END_EXAM] stops execution
- [ ] No infinite loops in test scenarios
- [ ] Max round limits still enforced as fallback

---

### Tác vụ 2.2: Fix Recursive Handoff Context

**Cái gì:**
- Format state for next agent
- Pass meaningful input instead of empty string
- Update history before handoff

**File cần sửa:**
```
core/workflow/
└─ execution.go         (modify)
   ├─ formatStateForNextAgent() (new)
   └─ Pass context in handoff (modify)
```

**Acceptance Criteria:**
- [ ] Next agent receives non-empty context
- [ ] State changes visible in next round
- [ ] Agent knows progress/history
- [ ] No information loss in handoff

---

### Tác vụ 2.3: Create Tool Orchestration Middleware

**Cái gì:**
- Middleware layer for tool execution
- Error handling and retry logic
- Tool result formatting and integration

**File cần tạo:**
```
core/workflow/
└─ tool_orchestration.go (new)
   ├─ ToolOrchestrator struct
   ├─ Execute() method
   └─ Error handling
```

**Acceptance Criteria:**
- [ ] All tools executed with error handling
- [ ] Partial success doesn't stop workflow
- [ ] Tool results properly formatted
- [ ] Tool timeouts respected

---

## GIAI ĐOẠN 3: CẢI THIỆN (Tháng 2, Tuần 1-2)

### Tác vụ 3.1: Cost Enforcement

**Cái gì:**
- Enforce cost budgets per agent/tool
- Check before execution, not after
- Prevent budget overruns

**Acceptance Criteria:**
- [ ] Quiz example cost bounded
- [ ] Tool execution cost checked
- [ ] No execution if over budget

---

### Tác vụ 3.2: Configuration Validation

**Cái gì:**
- Detect circular dependencies in handoffs
- Validate all tool parameters
- Pre-flight configuration checks

**Acceptance Criteria:**
- [ ] Invalid configs rejected at init
- [ ] Circular dependencies detected
- [ ] Helpful error messages

---

### Tác vụ 3.3: Signal Registry Interface

**Cái gì:**
- Convert to interface instead of concrete type
- Enable multiple implementations
- Improve testability

**Acceptance Criteria:**
- [ ] Interface defined
- [ ] Current implementation works with interface
- [ ] Can be mocked in tests

---

## TESTING STRATEGY

### Unit Tests
```
core/state-management/
├─ execution_state_test.go     (enhance)
├─ state_persistence_test.go   (new)
└─ atomic_updates_test.go      (new)

core/workflow/
└─ execution_tool_integration_test.go (new)

core/termination/
├─ checker_test.go             (new)
└─ strategies_test.go          (new)
```

### Integration Tests
```
examples/01-quiz-exam/
├─ test_quiz_completion.go     (new)
│  └─ Should complete quiz, not infinite loop
├─ test_state_persistence.go   (new)
│  └─ Should track correct answers
└─ test_tool_execution.go      (new)
   └─ Should execute RecordAnswer properly
```

### Acceptance Criteria for All Tests
- [ ] Quiz example completes successfully
- [ ] No infinite loops detected
- [ ] State persists across rounds
- [ ] Tool results visible to agents
- [ ] Cost bounded and logged

---

## SUCCESS METRICS

| Metric | Before | After | Target |
|--------|--------|-------|--------|
| Quiz Example Completion | ❌ Infinite loop | ✓ Terminates | 10 rounds |
| Rounds to Complete | ∞ | 10 | < 15 |
| Cost Control | $∞ | Bounded | < $0.50 |
| State Persistence | 0% | 100% | 100% |
| Tool Execution | 0% | 100% | 100% |
| Termination Accuracy | 0% | 100% | 100% |

---

## RISK ASSESSMENT

### High Risk Areas
1. **Breaking Changes** - State management changes to ExecutionContext
   - Mitigation: Backward compatibility layer
2. **Integration Testing** - Tool + State + Workflow integration
   - Mitigation: Comprehensive integration tests
3. **Performance** - State snapshots could impact performance
   - Mitigation: Implement async snapshots

### Monitoring Plan
- Log all state transitions
- Track termination reasons
- Monitor token growth patterns
- Alert on infinite loop detection (> 20 rounds)

---

## ROLLOUT PLAN

### Phase 1: Internal Testing (Week 1-2)
- Develop and test with quiz example
- Run full test suite
- Document any breaking changes

### Phase 2: Beta Release (Week 3)
- Release as beta branch
- Collect feedback
- Fix critical issues

### Phase 3: Production Release (Week 4)
- Merge to main
- Update documentation
- Release as new version

---

## DOCUMENTATION UPDATES NEEDED

- [ ] Update API documentation for new StateManager interface
- [ ] Add examples for stateful workflows
- [ ] Document termination logic
- [ ] Add tool orchestration guide
- [ ] Update migration guide for breaking changes

---

## EFFORT ESTIMATE

| Phase | Tasks | Effort | Risk |
|-------|-------|--------|------|
| 1 | State + Tools + Signals | 40 hours | Medium |
| 2 | Termination + Context + Middleware | 30 hours | Medium |
| 3 | Enhancements | 20 hours | Low |
| Testing | All phases | 30 hours | High |
| **Total** | **15 tasks** | **120 hours** | **Medium** |

**Timeline:** 4 weeks with 1 developer = 1-2 developers, 2-3 weeks

---

## DEPENDENCIES

- Go 1.20+
- No new external dependencies
- Existing test frameworks (testing, testify)

---

## SUCCESS CRITERIA FOR ENTIRE PROJECT

The project is complete when:

1. ✓ Quiz example completes without infinite loop
2. ✓ All critical weaknesses (#1-3) addressed
3. ✓ All major weaknesses (#4-6) addressed  
4. ✓ Test coverage > 85%
5. ✓ Zero regressions in existing functionality
6. ✓ Documentation updated
7. ✓ Breaking changes documented with migration guide

---

## APPROVAL & SIGN-OFF

**Current Status:** ⏳ Planning Phase

**Next Step:** Review this plan and approve before starting implementation

**Decision Points:**
- [ ] Approve overall approach?
- [ ] Approve timeline (4 weeks)?
- [ ] Accept breaking changes?
- [ ] Approve resource allocation?

