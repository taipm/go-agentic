---
title: "Epics: go-agentic Library Quality & Robustness Improvements"
version: "1.0.0"
date: "2025-12-20"
project: "go-agentic"
status: "Validated & Ready for Development"
stepsCompleted: [1, 2, 3, 4]
inputDocuments:
  - PRD.md
  - Architecture.md
  - UX-Design.md
validationStatus: "All checks passed - 8/8 FRs, 8/8 NFRs, 28 stories, 0 blocking issues"
---

# Epics Document
## go-agentic Library Quality & Robustness Improvements

**Project:** go-agentic
**Version:** 1.0.0
**Date:** 2025-12-20
**Status:** Approved
**Epic Strategy:** User-Value Focused, Dependency-Clear, Comprehensive Testing

---

## EXECUTIVE SUMMARY

This document defines **7 epics** organized around user value with clear dependencies and comprehensive testing strategy. The epic structure enables:

- ✅ Sequential delivery with dependencies explicit
- ✅ Granular functionality per epic
- ✅ Parallel testing strategy
- ✅ End-to-end validation
- ✅ Regression testing coverage

**Total FRs Covered:** 8 (100%)
**Total NFRs Covered:** 8 (100%)
**Delivery Model:** Sequential with parallel testing

---

## EPIC DELIVERY STRATEGY

```
PHASE 1: Foundation
┌─────────────────────────────────────────┐
│ Epic 1: Configuration Integrity & Trust │
│ (All agents use configured models)      │
└──────────────┬──────────────────────────┘
               │
               ↓
PHASE 2: Core Functionality
┌──────────────────────────────────────────────────────────────┐
│ Epic 2a: Native Tool Call Parsing                            │
│ (Reliable tool call extraction via OpenAI native API)        │
│                                                               │
│ Epic 2b: Parameter Validation & Type Safety                  │
│ (JSON Schema validation before handler execution)            │
└──────────────┬────────────────────────────────┬──────────────┘
               │                                │
               ↓                                ↓
PHASE 3: Quality & Compatibility (PARALLEL)
┌─────────────────────────────┐  ┌──────────────────────────┐
│ Epic 3: Clear Error         │  │ Epic 4: Cross-Platform   │
│ Handling                    │  │ Compatibility            │
│ (Actionable error messages) │  │ (Windows, macOS, Linux)  │
└──────────────┬──────────────┘  └────────────┬─────────────┘
               │                              │
               └──────────────┬───────────────┘
                              │
                              ↓
PHASE 4: Testing & Validation
┌──────────────────────────────────────────────────────────────┐
│ Epic 5: Production-Ready Testing Framework                   │
│ (API alignment, >90% coverage, CI/CD setup)                  │
│                                                               │
│ Epic 6: Parallel Testing (runs with all phases)              │
│ (Unit, integration, platform-specific tests)                 │
│                                                               │
│ Epic 7: End-to-End Validation & Regression                   │
│ (Full workflow testing, backward compatibility verification) │
└──────────────────────────────────────────────────────────────┘
```

---

## EPIC LIST

### **EPIC 1: Configuration Integrity & Trust**

**Goal:** Users can configure agents with confidence knowing every setting will be honored exactly as specified.

**Priority:** CRITICAL (Foundation)
**Sequence:** Phase 1 - Start here
**Dependencies:** None (standalone foundation)
**Enables:** Epic 2a, 2b (tools need correct config)

**User Outcomes:**
- Agents use the configured Model (not hardcoded to gpt-4o-mini)
- Temperature settings are respected including 0.0 values
- All configuration settings apply consistently
- Users can predict agent behavior from configuration

**FRs Covered:**
- FR1: Agent Configuration Respect (gpt-4o-mini hardcoded bug)
- FR8: Temperature Configuration Validation (0.0 override bug)

**NFRs Covered:**
- NFR1: Correctness (100% config field respect)
- NFR6: Backward Compatibility (existing configs work)

**Implementation Scope:**
- Fix hardcoded `"gpt-4o-mini"` in agent.go:24 → use `agent.Model`
- Fix Temperature override logic in config.go → allow 0.0 values
- Add configuration validation on load
- Add logging to show actual model used per agent
- Update integration tests to verify model selection

**Technical Details:**
```go
// BEFORE (BROKEN):
params := openai.ChatCompletionNewParams{
    Model: "gpt-4o-mini",  // ❌ Hardcoded
}

// AFTER (FIXED):
params := openai.ChatCompletionNewParams{
    Model: agent.Model,  // ✅ Uses configuration
}
```

**Acceptance Criteria:**
- [ ] Agent.Model field is read from configuration
- [ ] Integration tests verify correct model used in API call
- [ ] Temperature config allows 0.0 (deterministic)
- [ ] Logs show which model each agent uses
- [ ] Existing examples work unchanged (backward compatible)
- [ ] Configuration validation prevents invalid values

**Testing Strategy:**
- Unit tests: Config loading and model assignment
- Integration tests: Verify OpenAI API call uses correct model
- Backward compatibility: Existing examples pass unchanged

**Estimated Effort:** Small (localized change)

---

### **EPIC 2a: Native Tool Call Parsing**

**Goal:** Users can execute tools with confidence they will be parsed reliably using OpenAI's native API mechanism.

**Priority:** CRITICAL (Core functionality)
**Sequence:** Phase 2 - After Epic 1
**Dependencies:** Epic 1 (needs config working)
**Enables:** Epic 2b, 3, 4, 6, 7

**User Outcomes:**
- Tool calls are parsed reliably from agent responses
- Parsing works with native OpenAI API (robust, official)
- Fallback text parsing provides safety net
- Tools execute consistently regardless of response format

**FRs Covered:**
- FR2: Robust Tool Call Extraction (fragile text parsing)

**NFRs Covered:**
- NFR2: Robustness (99% success rate)
- NFR7: Performance (<100ms overhead)

**Implementation Scope:**
- Implement native OpenAI tool_calls parsing from API response
- Add fallback to text-based parsing for safety
- Update system prompts to use function calling syntax
- Convert Tool structs to OpenAI tool format
- Update agent execution flow to handle both methods

**Technical Details:**
```go
// BEFORE (FRAGILE TEXT PARSING):
toolCalls := extractToolCallsFromText(content, agent)

// AFTER (NATIVE API + FALLBACK):
var toolCalls []ToolCall

// Try native API first
if len(message.ToolCalls) > 0 {
    toolCalls = parseNativeToolCalls(message.ToolCalls)
}

// Fallback to text parsing if needed
if len(toolCalls) == 0 {
    toolCalls = extractToolCallsFromText(content, agent)
}
```

**Acceptance Criteria:**
- [ ] Native OpenAI tool_calls field is parsed
- [ ] JSON arguments properly deserialized
- [ ] Fallback text parsing works for safety
- [ ] 50+ format variations tested and pass
- [ ] No performance degradation (<100ms overhead)
- [ ] System prompts updated for function calling
- [ ] Tool definitions converted to OpenAI format

**Testing Strategy:**
- Unit tests: Tool call parsing (native and fallback)
- Integration tests: Agent execution with tool calls
- Format variation tests: 50+ different response formats
- Performance tests: Verify <100ms overhead

**Estimated Effort:** Medium (refactor existing code)

---

### **EPIC 2b: Parameter Validation & Type Safety**

**Goal:** Users get clear errors when invalid parameters are passed to tools, preventing runtime failures.

**Priority:** HIGH (Quality)
**Sequence:** Phase 2 - After Epic 1, can run parallel with 2a
**Dependencies:** Epic 1 (needs config)
**Enables:** Epic 6, 7 (validation enables comprehensive testing)

**User Outcomes:**
- Tool parameters are validated against JSON Schema
- Type mismatches caught early with clear errors
- Required parameters checked before execution
- Invalid inputs prevented from reaching tool handlers

**FRs Covered:**
- FR3: Tool Parameter Validation (not implemented)

**NFRs Covered:**
- NFR5: Type Safety (param validation)
- NFR1: Correctness (early error detection)

**Implementation Scope:**
- Create parameter validation function using JSON Schema
- Validate before tool handler execution
- Check required fields, data types
- Return clear error messages for validation failures
- Add validation tests for various parameter types

**Technical Details:**
```go
// BEFORE (NO VALIDATION):
tool.Handler(ctx, args)  // args could be wrong type!

// AFTER (VALIDATED):
if err := validateToolParameters(tool, args); err != nil {
    return "", fmt.Errorf("parameter validation failed: %w", err)
}
tool.Handler(ctx, args)  // args are now validated safe
```

**Acceptance Criteria:**
- [ ] Parameters validated against JSON Schema
- [ ] Type checking enforced (string, number, integer, boolean)
- [ ] Required fields verified before execution
- [ ] Clear error messages for validation failures
- [ ] Unknown parameters rejected
- [ ] Optional parameters handled correctly
- [ ] Tests cover type mismatches (int vs string, etc.)

**Testing Strategy:**
- Unit tests: JSON Schema validation logic
- Integration tests: Tool execution with invalid parameters
- Type mismatch tests: String instead of number, etc.
- Edge case tests: Empty strings, null values, etc.

**Estimated Effort:** Medium (new validation layer)

---

### **EPIC 3: Clear & Actionable Error Handling**

**Goal:** Users receive clear, actionable error messages that help them understand and fix issues quickly.

**Priority:** HIGH (User experience)
**Sequence:** Phase 3 - After Epic 1 & 2a
**Dependencies:** Epic 1 (needs working config), Epic 2a (tools need to work)
**Enables:** Epic 6, 7 (better error diagnostics)

**User Outcomes:**
- All errors are captured (no silent failures)
- Error messages distinguish between error types (PERMISSION_DENIED, TIMEOUT, NOT_FOUND)
- Errors include suggested actions (how to fix)
- Tool failures are clear and debuggable

**FRs Covered:**
- FR5: Semantic Correctness in Message History (tool results as "system" role)
- FR6: Comprehensive Error Handling (no ignored errors)

**NFRs Covered:**
- NFR1: Correctness (no silent failures)
- NFR4: Error Handling (meaningful messages)

**Implementation Scope:**
- Create ToolError type with Type, Message, Cause fields
- Replace all `_, _ = cmd.Output()` patterns with proper error handling
- Categorize errors: PERMISSION_DENIED, TIMEOUT, NOT_FOUND, COMMAND_FAILED, etc.
- Add suggested actions to error messages
- Fix tool result message role: use "system" instead of "user"
- Update all error handling paths with consistent pattern

**Technical Details:**
```go
// BEFORE (SILENT FAILURES):
output, _ := cmd.Output()  // Error ignored!
if len(output) == 0 {
    return "Service not running"  // Could be error or really not running
}

// AFTER (CLEAR ERRORS):
output, err := cmd.Output()
if err != nil {
    if exitErr, ok := err.(*exec.ExitError); ok {
        return fmt.Sprintf("[COMMAND_FAILED] Service check error (exit %d)\n"+
            "Suggestion: Run with elevated privileges or check command exists",
            exitErr.ExitCode())
    }
    return fmt.Sprintf("[PERMISSION_DENIED] Cannot check service\n"+
        "Suggestion: Run with sudo")
}
return parseServiceStatus(string(output))
```

**Acceptance Criteria:**
- [ ] No `_, _` error ignoring patterns in code
- [ ] All errors captured and categorized
- [ ] Error messages include Type, Reason, Action
- [ ] Tool results use "system" role (semantically correct)
- [ ] Service check distinguishes "not running" vs "error"
- [ ] All error types covered (permission, timeout, not found, etc.)
- [ ] Error messages helpful for user debugging

**Testing Strategy:**
- Unit tests: Error categorization logic
- Integration tests: Tool execution with errors
- Scenario tests: Permission denied, timeout, not found, etc.
- Message clarity tests: User can understand and fix

**Estimated Effort:** Medium (error handling across codebase)

---

### **EPIC 4: Cross-Platform Compatibility**

**Goal:** Users can deploy the library and examples seamlessly on Windows, macOS, and Linux without modifications.

**Priority:** HIGH (Usability)
**Sequence:** Phase 3 - Parallel with Epic 3
**Dependencies:** Epic 1 (config), Epic 2a (parsing)
**Enables:** Epic 6, 7 (cross-platform testing)

**User Outcomes:**
- Same code works on Windows, macOS, and Linux
- OS-specific commands handled automatically
- No platform-specific configuration needed
- Consistent tool output across all platforms

**FRs Covered:**
- FR4: Cross-Platform Tool Compatibility (Windows ping fails)

**NFRs Covered:**
- NFR3: Cross-Platform (3 OSes tested)
- NFR6: Backward Compatibility (no breaking changes)

**Implementation Scope:**
- Create OS-aware command wrappers for system tools
- Replace hardcoded flags with platform-specific variants
- Handle Windows (-n flag) vs Unix (-c flag) for ping
- Handle service check differences (systemctl vs launchctl vs net)
- Handle process listing differences (ps vs tasklist)
- Add runtime.GOOS checks throughout
- Test on Windows, macOS, Linux CI runners

**Technical Details:**
```go
// BEFORE (WINDOWS BREAKS):
cmd := exec.CommandContext(ctx, "ping", "-c", count, host)  // -c doesn't work on Windows!

// AFTER (CROSS-PLATFORM):
cmd, err := getPingCommand(ctx, host, count)

func getPingCommand(ctx context.Context, host string, count string) (*exec.Cmd, error) {
    switch runtime.GOOS {
    case "windows":
        return exec.CommandContext(ctx, "ping", "-n", count, host), nil
    case "darwin", "linux":
        return exec.CommandContext(ctx, "ping", "-c", count, host), nil
    default:
        return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
    }
}
```

**Acceptance Criteria:**
- [ ] Same test suite passes on Windows, macOS, Linux
- [ ] No hardcoded platform-specific flags
- [ ] CI/CD runs tests on all 3 platforms
- [ ] Ping command works with correct flags per OS
- [ ] Service checks work on all OSes
- [ ] Process listing works on all OSes
- [ ] Output consistent across platforms

**Testing Strategy:**
- Unit tests: Command selection logic per OS
- Integration tests: Tool execution on each OS
- CI/CD: Automated testing on Windows, macOS, Linux runners
- Cross-platform scenario tests: Same workflow on all OSes

**Estimated Effort:** Medium (platform detection in multiple places)

---

### **EPIC 5: Production-Ready Testing Framework**

**Goal:** Users have comprehensive test framework with clear APIs, alignment with library implementation, and proper coverage.

**Priority:** MEDIUM (Quality infrastructure)
**Sequence:** Phase 4 - After Epic 1 & 2a
**Dependencies:** Epic 1 (working config)
**Enables:** Epic 6, 7 (test infrastructure needed)

**User Outcomes:**
- Test framework APIs match library exported APIs
- Test scenarios are comprehensive and realistic
- HTML reports show clear test results
- >90% code coverage achieved
- Tests document expected behaviors

**FRs Covered:**
- FR7: Test Framework API Alignment (references non-existent APIs)

**NFRs Covered:**
- NFR8: Code Quality (>90% coverage)

**Implementation Scope:**
- Implement/export test scenario runner (RunTestScenario)
- Implement/export scenario getter (GetTestScenarios)
- Implement/export HTML report generator
- Create comprehensive test scenarios for each epic
- Generate readable HTML test reports
- Achieve >90% code coverage
- Set up CI/CD coverage tracking

**Technical Details:**
```go
// EXPORTED TEST APIs (new in library):
func RunTestScenario(ctx context.Context, scenario *TestScenario,
    executor *TeamExecutor) *TestResult

func GetTestScenarios() []*TestScenario

func GenerateHTMLReport(results []*TestResult) string

// Test scenarios for each epic:
var testScenarios = []*TestScenario{
    {ID: "A", Name: "Config Model Selection", ...},
    {ID: "B", Name: "Tool Call Parsing", ...},
    {ID: "C", Name: "Parameter Validation", ...},
    {ID: "D", Name: "Error Handling", ...},
    {ID: "E", Name: "Cross-Platform", ...},
    // etc.
}
```

**Acceptance Criteria:**
- [ ] RunTestScenario function exported and working
- [ ] GetTestScenarios returns all scenarios
- [ ] GenerateHTMLReport produces readable reports
- [ ] Comprehensive test scenarios cover all epics
- [ ] >90% code coverage achieved
- [ ] Example test.go uses exported APIs successfully
- [ ] CI/CD tracks coverage metrics

**Testing Strategy:**
- Test the test framework itself (meta-testing)
- Verify all exported APIs work correctly
- Check HTML report generation
- Validate test scenario execution
- Coverage analysis and gap filling

**Estimated Effort:** Medium (test infrastructure)

---

### **EPIC 6: Parallel Testing Strategy (Runs with all phases)**

**Goal:** Implement comprehensive testing across all epics simultaneously to catch issues early and prevent regression.

**Priority:** HIGH (Quality assurance)
**Sequence:** Phase 4 - Runs parallel with Epic 1-5
**Dependencies:** None (testing framework runs alongside development)
**Enables:** Epic 7 (end-to-end validation)

**User Outcomes:**
- Unit tests validate individual components
- Integration tests validate components working together
- Platform-specific tests validate on Windows, macOS, Linux
- Tests run automatically on every change
- Early detection of regressions

**Implementation Scope:**
- Unit tests for each epic's components
- Integration tests for epic workflows
- Platform-specific test suites for each OS
- CI/CD pipeline with automated testing
- Test coverage tracking and reporting

**Test Coverage by Epic:**

**Epic 1 Tests:**
- Config loading validation
- Model assignment verification
- Temperature field handling
- Backward compatibility

**Epic 2a Tests:**
- Native tool_calls parsing
- Text fallback parsing
- Format variation robustness
- Performance benchmarks

**Epic 2b Tests:**
- JSON Schema validation
- Type checking (string, number, integer, boolean)
- Required field validation
- Error messages clarity

**Epic 3 Tests:**
- Error categorization (permission, timeout, etc.)
- Error message quality
- Message role semantics
- Tool result handling

**Epic 4 Tests:**
- Windows command variants
- macOS command variants
- Linux command variants
- Cross-platform consistency

**Epic 5 Tests:**
- Test API functionality
- HTML report generation
- Coverage calculation
- Scenario execution

**Acceptance Criteria:**
- [ ] Unit tests: >85% component coverage
- [ ] Integration tests: All workflows tested
- [ ] Platform tests: Pass on Windows, macOS, Linux
- [ ] CI/CD: Automated on every commit
- [ ] Coverage tracking: >90% overall
- [ ] No flaky tests (reliable results)

**Testing Strategy:**
- Test-driven development (tests first)
- Continuous integration (every commit)
- Multi-platform testing
- Coverage-driven (aim for >90%)

**Estimated Effort:** Ongoing (throughout all phases)

---

### **EPIC 7: End-to-End Validation & Regression Testing**

**Goal:** Ensure all epics work together correctly and verify no regressions were introduced.

**Priority:** HIGH (System integrity)
**Sequence:** Phase 4 - Final validation after all epics
**Dependencies:** Epic 1-6 complete
**Enables:** Production release

**User Outcomes:**
- Full user workflows tested end-to-end
- All platforms work together correctly
- Backward compatibility verified
- No regressions introduced
- System is production-ready

**Implementation Scope:**
- End-to-end user workflow tests (full agent conversations)
- Multi-platform integration tests
- Backward compatibility verification (v0.0.1 examples still work)
- Regression test suite
- Performance baseline and validation
- Release validation checklist

**E2E Test Scenarios:**

**Scenario 1: Library User - Configure & Execute**
```
Setup: Create 3 agents with different models
Action: Execute with various user inputs
Verify: Correct models used, tools executed, errors clear
Platforms: Windows, macOS, Linux
```

**Scenario 2: IT Support Example - Real Workflow**
```
Setup: Load IT Support crew
Action: Run diagnostic on slow computer (real-world scenario)
Verify: Correct flow (orchestrator → clarifier → executor)
         Tools executed reliably
         Clear results returned
Platforms: Windows, macOS, Linux
```

**Scenario 3: Cross-Platform Deployment**
```
Setup: Same code on all 3 platforms
Action: Run identical tests
Verify: Identical results across platforms
        No OS-specific failures
```

**Scenario 4: Backward Compatibility**
```
Setup: Load v0.0.1-alpha.1 examples
Action: Execute with new library v0.0.2-alpha.2
Verify: All examples work unchanged
        No breaking changes
```

**Regression Test Suite:**
- Previously fixed issues don't resurface
- Performance doesn't degrade
- Coverage doesn't decrease
- All platform tests still pass

**Acceptance Criteria:**
- [ ] 20+ end-to-end test scenarios pass
- [ ] All workflows work on Windows, macOS, Linux
- [ ] Backward compatibility verified
- [ ] No regressions detected
- [ ] Performance baseline met
- [ ] Release validation passed
- [ ] Documentation complete

**Testing Strategy:**
- Real-world workflow simulation
- Multi-platform validation
- Regression prevention
- Release readiness verification

**Estimated Effort:** Medium (comprehensive validation)

---

## REQUIREMENTS COVERAGE MAP

### FR Coverage (8/8 = 100%)

| FR | Epic | Mapping |
|----|------|---------|
| **FR1** | Epic 1 | Agent model configuration respected in execution |
| **FR2** | Epic 2a | Native tool call parsing with text fallback |
| **FR3** | Epic 2b | JSON Schema parameter validation |
| **FR4** | Epic 4 | Cross-platform OS-aware tool implementations |
| **FR5** | Epic 3 | Tool results use correct semantic role |
| **FR6** | Epic 3 | Error handling with categorization |
| **FR7** | Epic 5 | Test framework APIs aligned with library |
| **FR8** | Epic 1 | Temperature config not overridden |

### NFR Coverage (8/8 = 100%)

| NFR | Epic | Mapping |
|-----|------|---------|
| **NFR1** | Epic 1, 3 | Configuration and error correctness |
| **NFR2** | Epic 2a | Robust tool call parsing (99% success) |
| **NFR3** | Epic 4 | Cross-platform test validation |
| **NFR4** | Epic 3 | Comprehensive error handling |
| **NFR5** | Epic 2b | Type safety via parameter validation |
| **NFR6** | Epic 1, 4 | Backward compatibility maintained |
| **NFR7** | Epic 2a | Performance <100ms overhead |
| **NFR8** | Epic 5, 6 | >90% code coverage, quality standards |

---

## EPIC DEPENDENCIES MATRIX

```
Epic 1 (Config)
    ├─ Enables: Epic 2a, 2b, 3, 4
    └─ Status: Foundation (must be first)

Epic 2a (Tool Parsing)
    ├─ Depends on: Epic 1
    ├─ Enables: Epic 2b, 3, 4, 6, 7
    └─ Can parallel: None (must be sequential after 1)

Epic 2b (Param Validation)
    ├─ Depends on: Epic 1
    ├─ Can parallel: With Epic 3, 4 (after 2a starts)
    └─ Enables: Epic 6, 7

Epic 3 (Error Handling)
    ├─ Depends on: Epic 1, 2a (ideally)
    ├─ Can parallel: With Epic 4
    └─ Enables: Epic 6, 7

Epic 4 (Cross-Platform)
    ├─ Depends on: Epic 1, 2a
    ├─ Can parallel: With Epic 3
    └─ Enables: Epic 6, 7

Epic 5 (Test Framework)
    ├─ Depends on: Epic 1, 2a
    ├─ Can parallel: With Epic 3, 4
    └─ Enables: Epic 6, 7

Epic 6 (Parallel Testing)
    ├─ Depends on: Test infrastructure ready
    ├─ Runs parallel: With Epic 1-5
    └─ Enables: Epic 7

Epic 7 (E2E Validation)
    ├─ Depends on: Epic 1-6 complete
    ├─ Validates: All epics work together
    └─ Enables: Production release
```

---

## DELIVERY TIMELINE (Recommended Sequence)

```
WEEK 1-2:
  Epic 1: Configuration Integrity
  + Epic 5: Test Framework setup
  + Epic 6: Unit tests (parallel)

WEEK 3-4:
  Epic 2a: Native Tool Parsing
  Epic 2b: Parameter Validation (can start after 2a begins)
  + Epic 6: Integration tests (parallel)

WEEK 5:
  Epic 3: Error Handling (parallel with 4)
  Epic 4: Cross-Platform (parallel with 3)
  + Epic 6: Platform-specific tests (parallel)

WEEK 6:
  Epic 7: End-to-End Validation & Regression
  Final review and release preparation

Release: v0.0.2-alpha.2
```

---

## SUCCESS CRITERIA (Epic Level)

| Epic | Done When |
|------|-----------|
| **1** | Model used matches config, temperature allows 0.0 |
| **2a** | Tool calls parsed reliably, fallback works |
| **2b** | Parameters validated, clear errors on type mismatch |
| **3** | All errors captured, messages actionable, role semantics correct |
| **4** | Same tests pass on Windows, macOS, Linux |
| **5** | >90% coverage, test APIs exported and working |
| **6** | All tests pass on all platforms, no flaky tests |
| **7** | E2E workflows succeed, backward compatibility verified |

---

## RELEASE CRITERIA (v0.0.2-alpha.2)

- ✅ All 7 epics complete and tested
- ✅ All 8 FRs implemented and verified
- ✅ All 8 NFRs achieved and validated
- ✅ >90% code coverage
- ✅ All tests passing (Windows, macOS, Linux)
- ✅ E2E workflows validated
- ✅ Backward compatibility confirmed
- ✅ Documentation updated
- ✅ Code review approved
- ✅ Security scan passed

---

---

# DETAILED USER STORIES

## EPIC 1: Configuration Integrity & Trust

### Story 1.1: Agent Respects Configured Model

**As a** Go developer building multi-agent systems,
**I want** agents to use the model specified in their configuration,
**So that** I can optimize cost vs quality by selecting different models per agent.

**Acceptance Criteria:**

```gherkin
Given an Agent configured with Model = "gpt-4o"
When the agent is executed with ExecuteAgent()
Then the OpenAI API call SHALL use Model = "gpt-4o" (not hardcoded "gpt-4o-mini")
And the logs SHALL show "[INFO] Executing agent X with model gpt-4o"

Given an Agent configured with Model = "gpt-4o-mini"
When the agent is executed
Then the OpenAI API call SHALL use Model = "gpt-4o-mini"

Given an IT Support example with 3 agents (different models)
When all agents are executed
Then each agent SHALL use their configured model
And the API calls SHALL reflect the correct models
```

**Implementation Files:** agent.go (line 24)
**Effort:** Small

---

### Story 1.2: Temperature Configuration Respects All Valid Values

**As a** Go developer fine-tuning agent behavior,
**I want** temperature configuration to be respected for all valid values including 0.0,
**So that** I can control agent creativity (0.0=deterministic, 1.0=creative).

**Acceptance Criteria:**

```gherkin
Given an Agent configured with Temperature = 0.0
When configuration is loaded
Then Temperature SHALL remain 0.0 (not overridden to 0.7)
And the agent SHALL execute with Temperature = 0.0

Given an Agent configured with Temperature = 1.0
When configuration is loaded
Then Temperature SHALL remain 1.0

Given an Agent configured with Temperature = 0.5
When configuration is loaded
Then Temperature SHALL remain 0.5

Given invalid Temperature (e.g., 3.0, -1.0)
When configuration is loaded
Then validation SHALL reject with clear error message
```

**Implementation Files:** config.go (temperature override logic)
**Effort:** Small

---

### Story 1.3: Configuration Validation & Error Messages

**As a** Go developer setting up agents,
**I want** clear error messages when configuration is invalid,
**So that** I can fix configuration issues quickly without guessing.

**Acceptance Criteria:**

```gherkin
Given invalid Agent configuration (empty Model)
When configuration is loaded
Then error SHALL be: "[ERROR] Agent 'X' Model field is required"

Given invalid Temperature value (e.g., 5.0)
When configuration is loaded
Then error SHALL be: "[ERROR] Agent 'X' Temperature must be 0.0-2.0, got 5.0"

Given valid configuration
When loaded
Then no errors SHALL occur
And logs SHALL show "[INFO] Agent 'X' initialized successfully"

Given existing v0.0.1 configuration files
When loaded with v0.0.2 library
Then all settings SHALL be respected (backward compatible)
```

**Implementation Files:** config.go, types.go
**Effort:** Small

---

## EPIC 2a: Native Tool Call Parsing

### Story 2a.1: Parse Native OpenAI Tool Calls

**As a** Go developer using go-agentic tools,
**I want** tool calls to be extracted from OpenAI's native `tool_calls` field,
**So that** I can rely on official API behavior instead of fragile text parsing.

**Acceptance Criteria:**

```gherkin
Given an agent response with native tool_calls from OpenAI
When ExecuteAgent processes the response
Then tool_calls SHALL be extracted from message.ToolCalls
And each tool call SHALL have ID, function name, and arguments
And arguments SHALL be properly deserialized from JSON

Given a response with 2 tool calls: [GetCPUUsage(), PingHost("8.8.8.8")]
When parsed
Then 2 ToolCall objects SHALL be created
And ToolCall[0].ToolName = "GetCPUUsage"
And ToolCall[1].ToolName = "PingHost"
And ToolCall[1].Arguments["host"] = "8.8.8.8"

Given an agent with 0 native tool_calls
When parsed
Then no errors SHALL occur
And tool execution SHALL proceed (or skip if no calls)
```

**Implementation Files:** agent.go (ExecuteAgent function)
**Effort:** Medium

---

### Story 2a.2: Fallback Text Parsing for Compatibility

**As a** Go developer maintaining backward compatibility,
**I want** a fallback text-based parser if native tool_calls aren't available,
**So that** the library continues working with older agent formats.

**Acceptance Criteria:**

```gherkin
Given an agent response with NO native tool_calls
But with text containing "GetCPUUsage()"
When parsed
Then fallback text parser SHALL detect the tool call
And ToolCall SHALL be created with ToolName and Arguments

Given formatted agent response: "I will call GetCPUUsage() to check"
When parsed
Then text parser SHALL extract GetCPUUsage()
And argument parser SHALL handle empty args

Given response with multiple calls: "CheckServiceStatus(nginx) then PingHost(8.8.8.8)"
When parsed with fallback
Then both tool calls SHALL be extracted

Given response without any tool calls
When parsed
Then empty tool calls list SHALL be returned
And no errors SHALL occur
```

**Implementation Files:** agent.go (extractToolCallsFromText function)
**Effort:** Small

---

### Story 2a.3: Update System Prompts for Function Calling

**As a** library maintainer optimizing agent communication,
**I want** system prompts updated to request function calling format,
**So that** agents naturally emit structured tool calls instead of text.

**Acceptance Criteria:**

```gherkin
Given an Agent with tools defined
When buildSystemPrompt is called
Then system prompt SHALL instruct agent to use function_calls
And prompt SHALL describe available tools in OpenAI format
And prompt SHALL NOT instruct text-based tool call format

Given system prompt for agent with 3 tools
When displayed
Then tools SHALL be listed with descriptions
And agent SHALL understand format expected

Given agent executing with new prompts
When tool is called
Then native tool_calls SHALL be used (if model supports)
```

**Implementation Files:** agent.go (buildSystemPrompt function)
**Effort:** Medium

---

## EPIC 2b: Parameter Validation & Type Safety

### Story 2b.1: Validate Tool Parameters Against JSON Schema

**As a** Go developer using tools,
**I want** tool parameters validated against their JSON Schema before execution,
**So that** I get clear error messages about invalid types instead of cryptic runtime errors.

**Acceptance Criteria:**

```gherkin
Given a Tool with Parameters: { "type": "object", "properties": { "host": {"type": "string"} } }
When validateToolParameters is called with args: { "host": "192.168.1.1" }
Then validation SHALL pass
And tool execution SHALL proceed

Given the same tool with args: { "host": 123 }  (integer instead of string)
Then validation SHALL fail with: "[ERROR] Parameter 'host': expected string, got int"
And tool execution SHALL be prevented

Given a Tool with required parameter "host"
When called without "host" parameter
Then validation SHALL fail with: "[ERROR] Required parameter missing: 'host'"

Given a Tool with optional parameter "count"
When called without "count" (but has other required params)
Then validation SHALL pass (optional params can be missing)
```

**Implementation Files:** agent.go, team.go (new validation functions)
**Effort:** Medium

---

### Story 2b.2: Integrate Parameter Validation into Tool Execution

**As a** library user,
**I want** parameter validation to happen before tool handlers are called,
**So that** errors are caught early with clear messages.

**Acceptance Criteria:**

```gherkin
Given tool execution in TeamExecutor
When a tool call is about to be executed
Then validateToolParameters SHALL be called first
And if validation fails, error SHALL be returned immediately
And tool handler SHALL never be called with invalid parameters

Given agent response: ToolCall{ToolName: "PingHost", Arguments: {"host": 123}}
When tool execution attempted
Then validation error returned (type mismatch)
And handler NOT called
And user sees: "[ERROR] Parameter 'host': expected string, got int"

Given valid parameters: {"host": "8.8.8.8"}
When tool execution attempted
Then validation passes
And handler IS called
And tool executes successfully
```

**Implementation Files:** team.go (tool execution loop)
**Effort:** Small

---

## EPIC 3: Clear & Actionable Error Handling

### Story 3.1: Create ToolError Type with Categorization

**As a** library maintainer,
**I want** a structured ToolError type that categorizes errors,
**So that** errors can be handled programmatically and messages can be consistent.

**Acceptance Criteria:**

```gherkin
Given a ToolError with Type = "PERMISSION_DENIED"
When converted to string
Then message format: "[PERMISSION_DENIED] Description: suggestion"

Given ToolError constructor
When creating error with Type="NOT_FOUND", Message="Command not found"
Then ToolError.Error() returns formatted message

Given various error types to categorize:
- PERMISSION_DENIED (access denied)
- TIMEOUT (command timed out)
- NOT_FOUND (command not found)
- COMMAND_FAILED (non-zero exit)
- INVALID_INPUT (parameter error)
When appropriate type selected
Then error categorization is clear
```

**Implementation Files:** types.go (new ToolError struct)
**Effort:** Small

---

### Story 3.2: Replace Silent Errors with Proper Error Handling

**As a** Go developer troubleshooting issues,
**I want** all command errors captured and reported,
**So that** I know whether a service is "not running" or "command failed".

**Acceptance Criteria:**

```gherkin
Given a tool that runs a system command
When the command fails (non-zero exit code)
Then error SHALL be captured (not silently ignored)
And error type SHALL be categorized (PERMISSION_DENIED, TIMEOUT, etc.)
And error message SHALL include:
  - What failed (which command)
  - Why it failed (error details)
  - How to fix it (suggestion)

Given CheckServiceStatus(nginx) when service doesn't exist
Then error: "[PERMISSION_DENIED] Cannot check service status\nSuggestion: Run with sudo"
NOT: "Service not running" (misleading!)

Given PingHost(invalid.ip) with network timeout
Then error: "[TIMEOUT] Ping to invalid.ip timed out\nSuggestion: Check network connectivity"

Given all system tool handlers (ping, service check, etc.)
When errors occur
Then all SHALL follow same error format
And no `_, _` error ignoring patterns SHALL exist in code
```

**Implementation Files:** example_it_support.go (all tool handlers)
**Effort:** Medium

---

### Story 3.3: Fix Tool Result Message Role (Semantic Correctness)

**As a** LLM context manager,
**I want** tool results added to message history with correct semantic role,
**So that** the LLM understands message origin correctly.

**Acceptance Criteria:**

```gherkin
Given tool execution returns results
When results are added to conversation history
Then message.Role SHALL be "system" or "tool" (not "user")
And message.Content SHALL clearly indicate tool results

Given history with sequence:
  1. User message (role: "user")
  2. Agent response (role: "assistant")
  3. Tool result (role: "system")  ← Correct semantic
  4. Agent response after tool (role: "assistant")
When LLM processes history
Then context SHALL be semantically correct

Given message: "Tool GetCPUUsage returned: 45.2%"
When added to history
Then Role SHALL be "system"
And content prefix like "[Tool Result]" for clarity
```

**Implementation Files:** team.go (history management)
**Effort:** Small

---

## EPIC 4: Cross-Platform Compatibility

### Story 4.1: Implement OS-Aware Command Wrappers

**As a** DevOps engineer deploying IT Support on multiple OSes,
**I want** system commands to use OS-appropriate flags automatically,
**So that** the same code works on Windows, macOS, and Linux.

**Acceptance Criteria:**

```gherkin
Given getPingCommand helper function
When called on Windows
Then returns: exec.Command("ping", "-n", count, host)

When called on macOS or Linux
Then returns: exec.Command("ping", "-c", count, host)

When called on unknown OS
Then returns error: "[ERROR] unsupported OS: [osname]"

Given IT Support example with PingHost tool
When executed on Windows
Then ping command uses -n flag
And output returned successfully

When executed on macOS
Then ping command uses -c flag
And output returned successfully

When executed on Linux
Then ping command uses -c flag
And output returned successfully
```

**Implementation Files:** example_it_support.go (tool handlers)
**Effort:** Medium

---

### Story 4.2: Handle Cross-Platform Service Checking

**As a** system administrator,
**I want** service status checks to work on all platforms,
**So that** diagnostics work across my mixed OS infrastructure.

**Acceptance Criteria:**

```gherkin
Given CheckServiceStatus(nginx) on Windows
When executed
Then uses: "net start" or "Get-Service" command
And returns: "Service is running" or "Service not found"

When executed on macOS
Then uses: "launchctl list | grep service"
And returns consistent result

When executed on Linux
Then uses: "systemctl is-active service"
And returns consistent result

Given same service name across all platforms
When checked
Then results are semantically equivalent
And error handling is consistent
```

**Implementation Files:** example_it_support.go (CheckServiceStatus handler)
**Effort:** Medium

---

### Story 4.3: Setup Cross-Platform CI/CD Testing

**As a** maintainer ensuring quality,
**I want** automated tests running on Windows, macOS, and Linux,
**So that** I catch platform-specific issues before release.

**Acceptance Criteria:**

```gherkin
Given GitHub Actions CI/CD configuration
When push to main branch
Then tests SHALL run on: ubuntu-latest, macos-latest, windows-latest

Given same test suite on all 3 platforms
Then ALL tests SHALL pass on all platforms
And no platform-specific skips/xfails

Given test output
Then results show: "PASS [windows] PASS [macos] PASS [linux]"
And cross-platform compatibility verified
```

**Implementation Files:** .github/workflows/*.yml
**Effort:** Small

---

## EPIC 5: Production-Ready Testing Framework

### Story 5.1: Implement RunTestScenario API

**As a** developer testing go-agentic,
**I want** to run predefined test scenarios programmatically,
**So that** I can verify system functionality automatically.

**Acceptance Criteria:**

```gherkin
Given exported function: RunTestScenario(ctx, scenario, executor)
When called with a TestScenario
Then returns: *TestResult with Pass/Fail status

Given TestScenario with:
  - ID: "A"
  - Name: "Config Model Selection"
  - UserInput: "Check CPU"
  - ExpectedFlow: ["orchestrator", "executor"]
When RunTestScenario called
Then executor executes the input
And actual agent flow compared to expected
And TestResult shows: Pass/Fail

Given failed scenario
Then TestResult.Passed = false
And TestResult.Errors contains reason
And TestResult.ActualFlow shows what happened
```

**Implementation Files:** tests.go (new exported API)
**Effort:** Medium

---

### Story 5.2: Implement GetTestScenarios API

**As a** developer getting available tests,
**I want** to query all predefined test scenarios,
**So that** I can see what scenarios are available and run them.

**Acceptance Criteria:**

```gherkin
Given GetTestScenarios() function call
When invoked
Then returns: []*TestScenario with all scenarios

Given returned scenarios
Then each has: ID, Name, Description, UserInput, ExpectedFlow
And scenarios cover all major features
And at least 10 scenarios available (1 per feature area)

Given scenario list
When displayed
Then user can see: "Scenario A: Config Model Selection"
And understand what each tests
```

**Implementation Files:** tests.go (scenario definitions)
**Effort:** Small

---

### Story 5.3: Generate HTML Test Reports

**As a** project manager viewing test results,
**I want** HTML reports showing test results clearly,
**So that** I can see which tests passed/failed at a glance.

**Acceptance Criteria:**

```gherkin
Given test results from multiple scenarios
When GenerateHTMLReport(results) called
Then returns HTML string with:
  - Summary: X tests passed, Y tests failed
  - Per-test: Name, Status (PASS/FAIL), Duration
  - Coverage metrics: Overall coverage %
  - Pass/Fail indicators (green/red)

Given report in HTML
When opened in browser
Then shows: "10/10 tests PASSED ✅"
And each test result clearly visible
And helpful for CI/CD integration
```

**Implementation Files:** report.go
**Effort:** Small

---

## EPIC 6: Parallel Testing (Unit + Integration + Platform-Specific)

### Story 6.1: Unit Tests for Epic 1 (Configuration)
- Test model configuration assignment
- Test temperature field handling
- Test configuration validation
- Test backward compatibility

### Story 6.2: Unit Tests for Epic 2a (Tool Parsing)
- Test native tool_calls extraction
- Test text fallback parsing
- Test argument deserialization
- Test 50+ format variations

### Story 6.3: Unit Tests for Epic 2b (Parameter Validation)
- Test JSON Schema validation
- Test type checking (string, number, integer, boolean)
- Test required field validation
- Test optional field handling

### Story 6.4: Unit Tests for Epic 3 (Error Handling)
- Test ToolError categorization
- Test error message formatting
- Test error suggestions
- Test message role semantics

### Story 6.5: Unit Tests for Epic 4 (Cross-Platform)
- Test OS detection (Windows, macOS, Linux)
- Test command variant selection
- Test fallback commands
- Test error handling per OS

### Story 6.6: Integration Tests (All Components)
- Test agent execution end-to-end
- Test tool parsing + validation + execution
- Test error propagation
- Test configuration + model usage

### Story 6.7: Platform-Specific Tests
- Windows test suite (ping, service checks)
- macOS test suite (launchctl, ps)
- Linux test suite (systemctl, free)

---

## EPIC 7: End-to-End Validation & Regression Testing

### Story 7.1: E2E Workflow Tests
- Test full agent conversation flows
- Test orchestrator → clarifier → executor sequence
- Test IT Support example real-world scenarios

### Story 7.2: Multi-Platform Integration Tests
- Run same tests on Windows, macOS, Linux
- Verify consistent results across platforms
- Test cross-platform tool compatibility

### Story 7.3: Backward Compatibility Tests
- Load v0.0.1-alpha.1 examples
- Execute with v0.0.2-alpha.2 library
- Verify all features work unchanged

### Story 7.4: Regression Test Suite
- Verify previously fixed issues don't resurface
- Test performance baseline
- Test coverage doesn't decrease
- Test all platform tests still pass

---

**Document Status:** Stories Created - Ready for Development
**Next Steps:** STEP 04 - Final validation and release preparation

