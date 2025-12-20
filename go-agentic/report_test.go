package agentic

import (
	"strings"
	"testing"
	"time"
)

// Test Story 5.3: Generate HTML Test Reports

// TestGenerateHTMLReportBasic tests basic HTML report generation with single result
func TestGenerateHTMLReportBasic(t *testing.T) {
	testResult := &TestResult{
		Scenario: &TestScenario{
			ID:           "A",
			Name:         "Vague Issue",
			Description:  "Test description",
			UserInput:    "Test input",
			ExpectedFlow: []string{"clarifier"},
			Assertions:   []string{"Test assertion"},
		},
		Passed:     true,
		Duration:   100 * time.Millisecond,
		ActualFlow: []string{"clarifier"},
		Response: &TeamResponse{
			AgentID:   "clarifier",
			AgentName: "Clarifier",
			Content:   "Test response content",
		},
		Errors:   []string{},
		Warnings: []string{},
	}

	report := NewHTMLReport([]*TestResult{testResult})

	if report == nil {
		t.Fatal("NewHTMLReport returned nil")
	}

	htmlContent := report.ToHTML()

	if htmlContent == "" {
		t.Fatal("ToHTML returned empty string")
	}

	// Verify basic HTML structure
	if !strings.Contains(htmlContent, "<!DOCTYPE html>") {
		t.Error("HTML missing DOCTYPE")
	}

	if !strings.Contains(htmlContent, "<html") {
		t.Error("HTML missing html tag")
	}

	if !strings.Contains(htmlContent, "</html>") {
		t.Error("HTML missing closing html tag")
	}
}

// TestGenerateHTMLReportAll tests HTML report generation with multiple results
func TestGenerateHTMLReportAll(t *testing.T) {
	results := make([]*TestResult, 0)

	for i := 0; i < 5; i++ {
		result := &TestResult{
			Scenario: &TestScenario{
				ID:           "A",
				Name:         "Test Scenario",
				Description:  "Description",
				UserInput:    "Input",
				ExpectedFlow: []string{"clarifier"},
				Assertions:   []string{"Assertion"},
			},
			Passed:     i%2 == 0, // Alternate pass/fail
			Duration:   time.Duration(100*(i+1)) * time.Millisecond,
			ActualFlow: []string{"clarifier"},
			Response: &TeamResponse{
				AgentID:   "clarifier",
				AgentName: "Clarifier",
				Content:   "Response content",
			},
			Errors:   []string{},
			Warnings: []string{},
		}
		results = append(results, result)
	}

	report := NewHTMLReport(results)

	if report.TotalTests != 5 {
		t.Errorf("Expected 5 total tests, got %d", report.TotalTests)
	}

	if report.PassedTests != 3 {
		t.Errorf("Expected 3 passed tests, got %d", report.PassedTests)
	}

	if report.FailedTests != 2 {
		t.Errorf("Expected 2 failed tests, got %d", report.FailedTests)
	}

	htmlContent := report.ToHTML()

	if !strings.Contains(htmlContent, "Test Scenario") {
		t.Error("HTML missing test scenario name")
	}
}

// TestHTMLReportContainsSummary tests that report contains summary section
func TestHTMLReportContainsSummary(t *testing.T) {
	results := []*TestResult{
		{
			Scenario: &TestScenario{
				ID:           "A",
				Name:         "Test 1",
				Description:  "Desc",
				UserInput:    "Input",
				ExpectedFlow: []string{"executor"},
				Assertions:   []string{"Assert"},
			},
			Passed:     true,
			Duration:   50 * time.Millisecond,
			ActualFlow: []string{"executor"},
		},
		{
			Scenario: &TestScenario{
				ID:           "B",
				Name:         "Test 2",
				Description:  "Desc",
				UserInput:    "Input",
				ExpectedFlow: []string{"clarifier"},
				Assertions:   []string{"Assert"},
			},
			Passed:     false,
			Duration:   75 * time.Millisecond,
			ActualFlow: []string{"clarifier"},
		},
	}

	report := NewHTMLReport(results)
	htmlContent := report.ToHTML()

	// Verify summary contains expected metrics
	if !strings.Contains(htmlContent, "2") {
		t.Error("HTML missing total test count")
	}

	if !strings.Contains(htmlContent, "Total Tests") {
		t.Error("HTML missing 'Total Tests' label")
	}

	if !strings.Contains(htmlContent, "Passed") {
		t.Error("HTML missing 'Passed' label")
	}

	if !strings.Contains(htmlContent, "Failed") {
		t.Error("HTML missing 'Failed' label")
	}

	if !strings.Contains(htmlContent, "Pass Rate") {
		t.Error("HTML missing 'Pass Rate' label")
	}
}

// TestHTMLReportValidHTML tests that generated HTML is valid/well-formed
func TestHTMLReportValidHTML(t *testing.T) {
	result := &TestResult{
		Scenario: &TestScenario{
			ID:           "test-id",
			Name:         "Test Name",
			Description:  "Test Description",
			UserInput:    "Test with special chars",
			ExpectedFlow: []string{"agent"},
			Assertions:   []string{"Assert 1", "Assert 2"},
		},
		Passed:     true,
		Duration:   100 * time.Millisecond,
		ActualFlow: []string{"agent"},
		Response: &TeamResponse{
			AgentID:   "agent",
			AgentName: "Agent",
			Content:   "Response content",
		},
		Errors:   []string{},
		Warnings: []string{},
	}

	report := NewHTMLReport([]*TestResult{result})
	htmlContent := report.ToHTML()

	// Verify basic HTML structure
	if !strings.Contains(htmlContent, "<!DOCTYPE html>") {
		t.Error("HTML missing DOCTYPE")
	}

	// Verify report contains expected sections
	if !strings.Contains(htmlContent, "User Input") {
		t.Error("HTML missing User Input section")
	}

	if !strings.Contains(htmlContent, "Agent Flow") {
		t.Error("HTML missing Agent Flow section")
	}

	if !strings.Contains(htmlContent, "Agent Response") {
		t.Error("HTML missing Agent Response section")
	}

	// Verify status badge exists
	if !strings.Contains(htmlContent, "PASSED") {
		t.Error("HTML missing PASSED status badge")
	}

	// Verify timestamp exists
	if !strings.Contains(htmlContent, "Generated") {
		t.Error("HTML missing generation timestamp")
	}

	// Verify closing tags
	if !strings.Contains(htmlContent, "</html>") {
		t.Error("HTML missing closing html tag")
	}
}

// TestHTMLReportEmptyResults tests report generation with no results
func TestHTMLReportEmptyResults(t *testing.T) {
	report := NewHTMLReport([]*TestResult{})

	if report.TotalTests != 0 {
		t.Errorf("Expected 0 total tests, got %d", report.TotalTests)
	}

	htmlContent := report.ToHTML()

	if htmlContent == "" {
		t.Fatal("ToHTML returned empty string for empty results")
	}

	// Should still have valid HTML structure
	if !strings.Contains(htmlContent, "<!DOCTYPE html>") {
		t.Error("Empty report missing DOCTYPE")
	}
}

// TestHTMLReportPassRate tests correct pass rate calculation
func TestHTMLReportPassRate(t *testing.T) {
	results := make([]*TestResult, 0)

	// Create 10 results with 7 passing
	for i := 0; i < 10; i++ {
		result := &TestResult{
			Scenario: &TestScenario{
				ID:           "test",
				Name:         "Test",
				Description:  "Desc",
				UserInput:    "Input",
				ExpectedFlow: []string{"agent"},
				Assertions:   []string{"Assert"},
			},
			Passed:     i < 7, // First 7 pass, last 3 fail
			Duration:   100 * time.Millisecond,
			ActualFlow: []string{"agent"},
		}
		results = append(results, result)
	}

	report := NewHTMLReport(results)

	if report.PassedTests != 7 {
		t.Errorf("Expected 7 passed, got %d", report.PassedTests)
	}

	if report.FailedTests != 3 {
		t.Errorf("Expected 3 failed, got %d", report.FailedTests)
	}

	htmlContent := report.ToHTML()

	// 70% pass rate should be in the report
	if !strings.Contains(htmlContent, "70") {
		t.Error("HTML missing 70% pass rate")
	}
}

// TestHTMLReportErrorHandling tests report generation with error results
func TestHTMLReportErrorHandling(t *testing.T) {
	result := &TestResult{
		Scenario: &TestScenario{
			ID:           "error-test",
			Name:         "Error Test",
			Description:  "Test with errors",
			UserInput:    "Input",
			ExpectedFlow: []string{"executor"},
			Assertions:   []string{"Should execute"},
		},
		Passed:     false,
		Duration:   200 * time.Millisecond,
		ActualFlow: []string{},
		Errors: []string{
			"Execution error: context cancelled",
			"Flow mismatch. Expected: [executor], Got: []",
		},
		Warnings: []string{
			"Warning: timeout approaching",
		},
	}

	report := NewHTMLReport([]*TestResult{result})
	htmlContent := report.ToHTML()

	if !strings.Contains(htmlContent, "FAILED") {
		t.Error("HTML missing FAILED status")
	}

	if !strings.Contains(htmlContent, "Errors") {
		t.Error("HTML missing Errors section")
	}

	if !strings.Contains(htmlContent, "Warnings") {
		t.Error("HTML missing Warnings section")
	}

	// Verify error messages are included (but escaped)
	if !strings.Contains(htmlContent, "Execution error") {
		t.Error("HTML missing error message content")
	}
}
