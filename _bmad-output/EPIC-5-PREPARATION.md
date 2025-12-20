---
title: "Epic 5 Preparation: Production-Ready Testing Framework"
date: "2025-12-20"
status: "Ready to Start"
priority: "MEDIUM"
sequence: "Phase 4"
dependencies: "Epic 1 ✅ Complete"
---

# Epic 5: Production-Ready Testing Framework
**Preparation Document for Next Sprint**

---

## 📋 Epic Overview

**Goal:** Users have comprehensive test framework with clear APIs, alignment with library implementation, and proper coverage.

**Priority:** MEDIUM (Quality infrastructure)
**Sequence:** Phase 4 - After Epic 1 ✅ (Complete) & Epic 2a
**Dependencies:** Epic 1 ✅ (SATISFIED - Complete)

**Total Effort:** Medium (test infrastructure)
**Stories:** 3 stories
**Estimated Implementation Time:** 5-7 hours

---

## ✅ Dependency Status

| Dependency | Status | Notes |
|------------|--------|-------|
| **Epic 1** | ✅ Complete | Configuration working, ready to test |
| **Epic 2a** | ⏳ Pending | Can run in parallel starting now |
| **Project Context** | ✅ Complete | Testing patterns defined |
| **Test Infrastructure** | ⏳ Ready to build | No blockers |

---

## 🎯 Epic Objectives

### Primary Goals
1. **Export Test APIs** - Make testing APIs available to users
2. **Comprehensive Scenarios** - Define test scenarios for each epic
3. **HTML Reporting** - Generate readable test reports
4. **Coverage Tracking** - Achieve >90% code coverage
5. **CI/CD Integration** - Automate test execution

### Success Criteria
- ✅ RunTestScenario() function exported and working
- ✅ GetTestScenarios() returns all scenarios
- ✅ GenerateHTMLReport() produces readable reports
- ✅ >90% code coverage achieved
- ✅ Example test.go uses exported APIs successfully
- ✅ CI/CD tracks coverage metrics

---

## 📖 User Stories Breakdown

### Story 5.1: Implement RunTestScenario API
**Status:** Ready for Implementation
**Effort:** Medium (4-5 hours)
**Files:** tests.go (new/updated)

**What Gets Built:**
```go
// EXPORTED FUNCTION - Available to users
func RunTestScenario(ctx context.Context, scenario *TestScenario,
    executor *TeamExecutor) *TestResult
```

**Inputs:**
- `TestScenario`: Predefined test case with user input and expected flow
- `TeamExecutor`: Team to execute against
- `context`: For timeouts and cancellation

**Outputs:**
- `TestResult`: Contains Pass/Fail status, execution time, errors, actual flow

**Example Usage (What Users Will Do):**
```go
// Get a test scenario
scenarios := GetTestScenarios()
scenario := scenarios[0]  // "Config Model Selection"

// Run it
result := RunTestScenario(ctx, scenario, executor)

// Check result
if result.Passed {
    fmt.Println("✅ Test passed")
} else {
    fmt.Println("❌ Test failed:", result.Errors)
}
```

**Acceptance Criteria:**
```gherkin
Given a TestScenario with defined user input and expected flow
When RunTestScenario is called with executor
Then actual agent flow compared to expected
And TestResult returned with Pass/Fail status

Given failed scenario
Then TestResult.Passed = false
And TestResult.Errors contains failure reason
And TestResult.ActualFlow shows what actually happened
```

**Key Implementation Details:**
- Execute user input through TeamExecutor
- Capture actual agent flow (which agents were used)
- Compare against expected flow
- Record execution time
- Collect any errors that occurred

---

### Story 5.2: Implement GetTestScenarios API
**Status:** Ready for Implementation
**Effort:** Small (2-3 hours)
**Files:** tests.go (new/updated)

**What Gets Built:**
```go
// EXPORTED FUNCTION - Available to users
func GetTestScenarios() []*TestScenario

// TEST SCENARIO DEFINITION
type TestScenario struct {
    ID           string
    Name         string
    Description  string
    UserInput    string          // What user asks
    ExpectedFlow []string        // Which agents should execute
    ExpectedTags []string        // What the result should contain
}
```

**Example Scenarios (At Least 10):**

| ID | Name | Description | Expected Flow |
|----|------|-------------|---------------|
| A | Config Model Selection | Verify agent uses configured model | orchestrator → executor |
| B | Tool Call Parsing | Parse tool calls from agent response | orchestrator → executor |
| C | Parameter Validation | Validate tool parameters | executor (with error) |
| D | Error Handling | Handle tool execution errors | executor (error path) |
| E | Cross-Platform (Windows) | Run on Windows with Windows commands | executor (Windows) |
| F | Cross-Platform (Linux) | Run on Linux with Linux commands | executor (Linux) |
| G | Temperature Control | Use configured temperature | executor (0.0 temp) |
| H | IT Support Workflow | Run full IT support scenario | orchestrator → clarifier → executor |
| I | Backward Compatibility | Load v0.0.1 configs | executor (v0.0.1) |
| J | Performance | Complete in <5 seconds | executor |

**Acceptance Criteria:**
```gherkin
Given GetTestScenarios() is called
When invoked
Then returns []*TestScenario with all scenarios

Given returned scenarios
Then each has: ID, Name, Description, UserInput, ExpectedFlow
And scenarios cover all major features
And at least 10 scenarios available
```

**Key Implementation Details:**
- Define test scenarios that cover all epics
- Each scenario should test real user workflow
- Include both success and error cases
- Document expected agent flow
- Make scenarios repeatable and isolated

---

### Story 5.3: Generate HTML Test Reports
**Status:** Ready for Implementation
**Effort:** Small (2-3 hours)
**Files:** report.go (new/updated)

**What Gets Built:**
```go
// EXPORTED FUNCTION - Available to users
func GenerateHTMLReport(results []*TestResult) string

// Returns HTML string that can be:
// 1. Written to file
// 2. Served via HTTP
// 3. Displayed in CI/CD
```

**Example Output (HTML Format):**
```html
<!DOCTYPE html>
<html>
<head>
    <title>Test Report - go-agentic</title>
    <style>
        /* Professional styling */
        .pass { background: #d4edda; color: #155724; }
        .fail { background: #f8d7da; color: #721c24; }
    </style>
</head>
<body>
    <h1>Test Report - go-agentic</h1>

    <div class="summary">
        <h2>Summary</h2>
        <p>10/10 tests PASSED ✅</p>
        <p>Coverage: 92.5%</p>
        <p>Duration: 12.34s</p>
    </div>

    <table>
        <tr><th>Test ID</th><th>Name</th><th>Status</th><th>Duration</th></tr>
        <tr class="pass">
            <td>A</td>
            <td>Config Model Selection</td>
            <td>PASS ✅</td>
            <td>1.2s</td>
        </tr>
        <!-- More rows... -->
    </table>

    <div class="coverage">
        <h2>Coverage</h2>
        <p>agent.go: 85%</p>
        <p>config.go: 92%</p>
        <p>types.go: 100%</p>
    </div>
</body>
</html>
```

**When Used:**
- CI/CD generates after every test run
- Uploaded as artifact to GitHub Actions
- Linked in PR for easy review
- Used for coverage tracking

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

**Key Implementation Details:**
- Professional HTML/CSS styling
- Clear pass/fail indicators (colors, icons)
- Summary statistics at top
- Detailed per-test results
- Coverage breakdown by file
- Compatible with CI/CD artifact storage

---

## 📊 Implementation Checklist

### Pre-Implementation
- [ ] Review Epic 1 completion (✅ already complete)
- [ ] Understand TestScenario structure
- [ ] Plan HTML report template
- [ ] Identify coverage gaps

### Story 5.1: RunTestScenario
- [ ] Create TestScenario struct
- [ ] Create TestResult struct
- [ ] Implement RunTestScenario function
- [ ] Test with sample scenario
- [ ] Add to exports in library
- [ ] Write 5+ unit tests

### Story 5.2: GetTestScenarios
- [ ] Define all test scenarios (10+)
- [ ] Create scenario data structures
- [ ] Implement GetTestScenarios function
- [ ] Add to exports in library
- [ ] Document each scenario
- [ ] Write tests for scenario availability

### Story 5.3: HTML Reports
- [ ] Design HTML template
- [ ] Create CSS styling
- [ ] Implement GenerateHTMLReport function
- [ ] Add coverage calculation
- [ ] Test report generation
- [ ] Verify browser rendering
- [ ] Write 3+ unit tests

### Post-Implementation
- [ ] Run all tests
- [ ] Verify >90% coverage
- [ ] Generate sample report
- [ ] Update documentation
- [ ] Test with example.go
- [ ] Code review preparation

---

## 📁 File Structure (Expected)

### New/Modified Files
```
go-agentic/
├── tests.go                 (NEW - test infrastructure)
├── report.go                (NEW - HTML reporting)
├── types.go                 (MODIFY - add TestScenario, TestResult types)
├── tests_test.go            (NEW - test infrastructure tests)
└── report_test.go           (NEW - report generation tests)

_bmad-output/
└── EPIC-5-*.md              (documentation)
```

### Types to Define
```go
// In types.go or tests.go
type TestScenario struct {
    ID           string
    Name         string
    Description  string
    UserInput    string
    ExpectedFlow []string
    ExpectedTags []string
}

type TestResult struct {
    ScenarioID  string
    Passed      bool
    Duration    time.Duration
    Errors      []string
    ActualFlow  []string
    StartTime   time.Time
    EndTime     time.Time
}
```

---

## 🧪 Testing Strategy for Epic 5

### Unit Tests
```
tests_test.go:
- TestRunTestScenarioSuccess
- TestRunTestScenarioFailure
- TestGetTestScenariosCount
- TestGetTestScenariosContent
- TestTestResultPassing
- TestTestResultFailing

report_test.go:
- TestGenerateHTMLReportBasic
- TestGenerateHTMLReportMultiple
- TestHTMLReportContainsCorrectStats
- TestHTMLReportFormatting
```

### Integration Tests
- Run actual TestScenario against example executor
- Verify TestResult fields are populated
- Test edge cases (no scenarios, empty results)
- Test HTML report with real results

### Coverage Goals
- `tests.go` implementations: >90%
- `report.go` implementations: >90%
- Overall project: >85%

---

## 🔗 Dependencies & Parallel Work

### Can Start Immediately (Epic 1 ✅ Complete)
- ✅ Epic 5 implementation (no blockers)
- ✅ Epic 6 unit tests for Epic 1 (parallel testing)

### Can Start After 2a Begins
- Epic 5 integration tests (once parsing works)
- Epic 6 integration tests (once tools work)

### Must Precede
- Epic 7 (end-to-end testing depends on framework)
- Release (coverage requirements)

---

## 📈 Success Metrics

| Metric | Target | Verification |
|--------|--------|--------------|
| Stories Complete | 3/3 | All implemented |
| Tests Passing | 10+ | go test passes |
| Code Coverage | >90% | go tool cover |
| API Exports | 3 functions | goacode shows exported |
| Test Scenarios | 10+ | GetTestScenarios returns |
| HTML Report | Functional | Renders in browser |

---

## 🚀 Getting Started

### Recommended Implementation Order
1. **First:** Define types (TestScenario, TestResult)
2. **Second:** Implement GetTestScenarios() + scenarios
3. **Third:** Implement RunTestScenario()
4. **Fourth:** Implement GenerateHTMLReport()
5. **Finally:** Integration tests and examples

### Dev Environment Setup
```bash
# Ensure Epic 1 is merged
git pull origin main

# Create epic-5 branch
git checkout -b feature/epic-5-testing-framework

# Run existing tests
go test ./go-agentic/...

# Run coverage
go test -cover ./go-agentic/...

# Watch for changes
go test -v ./go-agentic/... -run TestScenario
```

---

## 📚 Related Documentation

**Already Available:**
- ✅ `epics.md` - Full Epic 5 specification
- ✅ `project-context.md` - Testing patterns
- ✅ `Architecture.md` - System design
- ✅ `EPIC-1-IMPLEMENTATION-COMPLETE.md` - Foundation complete

**To Create During Implementation:**
- Epic 5 detailed stories
- Epic 5 implementation guide
- Test scenario catalog
- HTML report examples

---

## ⚠️ Known Constraints

### From project-context.md
- ✅ Test scenarios must be isolated and repeatable
- ✅ No external dependencies for test execution
- ✅ Test results must be serializable (JSON, HTML)
- ✅ Coverage must account for all code paths
- ✅ Performance tests should complete in <5 seconds

### From Architecture.md
- ✅ TestScenario is immutable (no mutations)
- ✅ HTML report generated server-side (no JS dependencies)
- ✅ Export only stable APIs (no alpha/beta)
- ✅ Backward compatible with v0.0.1 test expectations

---

## 📞 Questions to Consider

Before starting implementation:

1. **Test Scenario Coverage:**
   - Should we include chaos/error scenarios?
   - How many scenarios per epic is ideal?
   - Should tests be data-driven?

2. **Report Generation:**
   - Should reports include execution logs?
   - What's the target browser compatibility?
   - Should we support CI/CD JSON export?

3. **Performance:**
   - What's acceptable test execution time?
   - Should tests run in parallel?
   - How do we handle flaky tests?

4. **Integration:**
   - How does this integrate with GitHub Actions?
   - Should reports be stored/archived?
   - How do we track coverage over time?

---

## ✅ Readiness Checklist

Before implementation:
- [x] Epic 1 complete ✅
- [x] Dependencies clear ✅
- [x] Stories understood ✅
- [x] File structure planned ✅
- [x] Types defined ✅
- [x] Testing strategy clear ✅
- [x] Success metrics defined ✅

**Status: READY TO IMPLEMENT** ✅

---

**Next Step:** User authorization to begin Epic 5 implementation

