package agentic

import (
	"strings"
	"testing"
	"time"
)

// TestNewHTMLReportBasic tests HTML report creation
func TestNewHTMLReportBasic(t *testing.T) {
	results := []*TestResult{
		{
			Scenario: &TestScenario{ID: "1", Name: "Test 1"},
			Passed:   true,
		},
		{
			Scenario: &TestScenario{ID: "2", Name: "Test 2"},
			Passed:   false,
		},
	}

	report := NewHTMLReport(results)

	if report == nil {
		t.Fatal("Expected non-nil report")
	}
	if report.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", report.TotalTests)
	}
	if report.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", report.PassedTests)
	}
	if report.FailedTests != 1 {
		t.Errorf("Expected 1 failed test, got %d", report.FailedTests)
	}
}

// TestNewHTMLReportEmpty tests report with no results
func TestNewHTMLReportEmpty(t *testing.T) {
	results := []*TestResult{}

	report := NewHTMLReport(results)

	if report.TotalTests != 0 {
		t.Error("Expected 0 total tests")
	}
	if report.PassedTests != 0 {
		t.Error("Expected 0 passed tests")
	}
	if report.FailedTests != 0 {
		t.Error("Expected 0 failed tests")
	}
}

// TestNewHTMLReportAllPassed tests report with all passed tests
func TestNewHTMLReportAllPassed(t *testing.T) {
	results := []*TestResult{
		{Scenario: &TestScenario{ID: "1", Name: "Test 1"}, Passed: true},
		{Scenario: &TestScenario{ID: "2", Name: "Test 2"}, Passed: true},
		{Scenario: &TestScenario{ID: "3", Name: "Test 3"}, Passed: true},
	}

	report := NewHTMLReport(results)

	if report.PassedTests != 3 {
		t.Errorf("Expected 3 passed tests, got %d", report.PassedTests)
	}
	if report.FailedTests != 0 {
		t.Errorf("Expected 0 failed tests, got %d", report.FailedTests)
	}
}

// TestNewHTMLReportAllFailed tests report with all failed tests
func TestNewHTMLReportAllFailed(t *testing.T) {
	results := []*TestResult{
		{Scenario: &TestScenario{ID: "1", Name: "Test 1"}, Passed: false},
		{Scenario: &TestScenario{ID: "2", Name: "Test 2"}, Passed: false},
	}

	report := NewHTMLReport(results)

	if report.PassedTests != 0 {
		t.Error("Expected 0 passed tests")
	}
	if report.FailedTests != 2 {
		t.Errorf("Expected 2 failed tests, got %d", report.FailedTests)
	}
}

// TestNewHTMLReportTimestamp tests that timestamp is set
func TestNewHTMLReportTimestamp(t *testing.T) {
	before := time.Now()
	report := NewHTMLReport([]*TestResult{})
	after := time.Now()

	if report.GeneratedAt.Before(before) || report.GeneratedAt.After(after) {
		t.Error("Timestamp should be between before and after")
	}
}

// TestToHTMLBasic tests HTML generation
func TestToHTMLBasic(t *testing.T) {
	results := []*TestResult{
		{
			Scenario: &TestScenario{
				ID:   "test1",
				Name: "Test Scenario 1",
			},
			Passed: true,
		},
	}

	report := NewHTMLReport(results)
	html := report.ToHTML()

	if html == "" {
		t.Fatal("Expected non-empty HTML")
	}

	// Check for HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should contain DOCTYPE")
	}
	if !strings.Contains(html, "</html>") {
		t.Error("HTML should be properly closed")
	}
}

// TestToHTMLContainsHeader tests HTML contains header
func TestToHTMLContainsHeader(t *testing.T) {
	report := NewHTMLReport([]*TestResult{})
	html := report.ToHTML()

	if !strings.Contains(html, "go-agentic Test Report") {
		t.Error("HTML should contain title")
	}
}

// TestToHTMLContainsSummary tests HTML contains summary section
func TestToHTMLContainsSummary(t *testing.T) {
	results := []*TestResult{
		{Scenario: &TestScenario{ID: "s1", Name: "Scenario 1"}, Passed: true},
		{Scenario: &TestScenario{ID: "s2", Name: "Scenario 2"}, Passed: false},
	}

	report := NewHTMLReport(results)
	html := report.ToHTML()

	if !strings.Contains(html, "2") {
		t.Error("HTML should contain total test count")
	}
}

// TestToHTMLValidStructure tests HTML has valid structure
func TestToHTMLValidStructure(t *testing.T) {
	results := []*TestResult{
		{
			Scenario: &TestScenario{
				ID:   "A",
				Name: "Scenario A",
			},
			Passed:   true,
			Response: &TeamResponse{AgentID: "agent1"},
		},
	}

	report := NewHTMLReport(results)
	html := report.ToHTML()

	// Check for key elements
	if !strings.Contains(html, "<head>") {
		t.Error("HTML should contain head")
	}
	if !strings.Contains(html, "<body>") {
		t.Error("HTML should contain body")
	}
	if !strings.Contains(html, "</body>") {
		t.Error("HTML should close body")
	}
	if !strings.Contains(html, "</head>") {
		t.Error("HTML should close head")
	}
}

// TestToHTMLMultipleResults tests HTML with multiple results
func TestToHTMLMultipleResults(t *testing.T) {
	results := []*TestResult{
		{
			Scenario: &TestScenario{ID: "A", Name: "Scenario A"},
			Passed:   true,
		},
		{
			Scenario: &TestScenario{ID: "B", Name: "Scenario B"},
			Passed:   false,
			Errors:   []string{"Error 1"},
		},
		{
			Scenario: &TestScenario{ID: "C", Name: "Scenario C"},
			Passed:   true,
		},
	}

	report := NewHTMLReport(results)
	html := report.ToHTML()

	// All scenarios should appear in HTML
	if !strings.Contains(html, "Scenario A") {
		t.Error("HTML should contain Scenario A")
	}
	if !strings.Contains(html, "Scenario B") {
		t.Error("HTML should contain Scenario B")
	}
	if !strings.Contains(html, "Scenario C") {
		t.Error("HTML should contain Scenario C")
	}
}

// TestToHTMLPassFail tests HTML indicates pass/fail status
func TestToHTMLPassFail(t *testing.T) {
	results := []*TestResult{
		{
			Scenario: &TestScenario{ID: "pass", Name: "Passing Test"},
			Passed:   true,
		},
		{
			Scenario: &TestScenario{ID: "fail", Name: "Failing Test"},
			Passed:   false,
			Errors:   []string{"Test failed"},
		},
	}

	report := NewHTMLReport(results)
	html := report.ToHTML()

	// Should distinguish between pass and fail
	if !strings.Contains(html, "Passing Test") {
		t.Error("HTML should contain passing test")
	}
	if !strings.Contains(html, "Failing Test") {
		t.Error("HTML should contain failing test")
	}
}

// TestToHTMLWithErrors tests HTML includes error information
func TestToHTMLWithErrors(t *testing.T) {
	results := []*TestResult{
		{
			Scenario: &TestScenario{ID: "error", Name: "Error Test"},
			Passed:   false,
			Errors: []string{
				"Assertion failed",
				"Expected X, got Y",
			},
		},
	}

	report := NewHTMLReport(results)
	html := report.ToHTML()

	// Errors should be visible
	if !strings.Contains(html, "Assertion failed") {
		t.Error("HTML should contain error message")
	}
}

// TestToHTMLWithWarnings tests HTML includes warning information
func TestToHTMLWithWarnings(t *testing.T) {
	results := []*TestResult{
		{
			Scenario: &TestScenario{ID: "warn", Name: "Warning Test"},
			Passed:   true,
			Warnings: []string{"Slow execution"},
		},
	}

	report := NewHTMLReport(results)
	html := report.ToHTML()

	// Should include warnings
	if !strings.Contains(html, "Slow execution") {
		t.Error("HTML should contain warning message")
	}
}

// TestToHTMLEmptyResults tests HTML with empty results
func TestToHTMLEmptyResults(t *testing.T) {
	report := NewHTMLReport([]*TestResult{})
	html := report.ToHTML()

	// Should still produce valid HTML
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("Empty report should still be valid HTML")
	}
	if !strings.Contains(html, "</html>") {
		t.Error("Empty report should be properly closed")
	}
}

// TestToHTMLContainsMetadata tests HTML includes metadata
func TestToHTMLContainsMetadata(t *testing.T) {
	results := []*TestResult{
		{
			Scenario: &TestScenario{
				ID:   "test",
				Name: "Test",
			},
			Duration: 100 * time.Millisecond,
			Passed:   true,
		},
	}

	report := NewHTMLReport(results)
	html := report.ToHTML()

	// Should contain timing info
	if len(html) < 100 {
		t.Error("HTML should contain substantial content")
	}
}

// TestHTMLReportPassRate tests pass rate calculation
func TestHTMLReportPassRate(t *testing.T) {
	tests := []struct {
		passed     int
		failed     int
		name       string
	}{
		{3, 0, "All passed"},
		{0, 3, "All failed"},
		{2, 2, "Half passed"},
		{1, 0, "Single pass"},
		{0, 1, "Single fail"},
	}

	for _, test := range tests {
		results := make([]*TestResult, test.passed+test.failed)
		for i := 0; i < test.passed; i++ {
			results[i] = &TestResult{Passed: true}
		}
		for i := test.passed; i < test.passed+test.failed; i++ {
			results[i] = &TestResult{Passed: false}
		}

		report := NewHTMLReport(results)

		if report.PassedTests != test.passed {
			t.Errorf("Test %s: expected %d passed, got %d", test.name, test.passed, report.PassedTests)
		}
		if report.FailedTests != test.failed {
			t.Errorf("Test %s: expected %d failed, got %d", test.name, test.failed, report.FailedTests)
		}
	}
}

// TestHTMLReportWithTestContent tests report with full test details
func TestHTMLReportWithTestContent(t *testing.T) {
	results := []*TestResult{
		{
			Scenario: &TestScenario{
				ID:           "scenario1",
				Name:         "Full Scenario",
				Description:  "This is a test scenario",
				UserInput:    "test input",
				ExpectedFlow: []string{"agent1", "agent2"},
				Assertions:   []string{"assertion 1"},
			},
			Passed:     true,
			Duration:   50 * time.Millisecond,
			ActualFlow: []string{"agent1", "agent2"},
			Response: &TeamResponse{
				AgentID:   "agent2",
				AgentName: "Agent 2",
				Content:   "Test response",
			},
			Errors:   []string{},
			Warnings: []string{"Test warning"},
		},
	}

	report := NewHTMLReport(results)
	html := report.ToHTML()

	// Should contain all test details
	if !strings.Contains(html, "scenario1") {
		t.Error("HTML should contain scenario ID")
	}
	if !strings.Contains(html, "Full Scenario") {
		t.Error("HTML should contain scenario name")
	}
	if !strings.Contains(html, "test input") {
		t.Error("HTML should contain user input")
	}
}

// TestNewHTMLReportNilResults tests handling nil results
func TestNewHTMLReportNilResults(t *testing.T) {
	// Should handle nil results gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NewHTMLReport should not panic on nil results, but got: %v", r)
		}
	}()

	report := NewHTMLReport(nil)
	if report != nil {
		// If not nil, it should at least have valid structure
		_ = report.ToHTML()
	}
}
