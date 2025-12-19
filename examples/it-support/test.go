package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/taipm/go-agentic"
)

// runTests executes all test scenarios
func runTests(crewConfig *agentic.CrewConfig, agentConfigs map[string]*agentic.AgentConfig, allTools map[string]*agentic.Tool) error {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🧪 GO-CREWAI TEST SUITE")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	// Create crew
	agents := createAgentsFromConfig(crewConfig, agentConfigs, allTools)
	crew := &agentic.Crew{
		Agents:      agents,
		MaxRounds:   crewConfig.Settings.MaxRounds,
		MaxHandoffs: crewConfig.Settings.MaxHandoffs,
		Routing:     crewConfig.Routing,
	}
	executor := agentic.NewTeamExecutor(crew, os.Getenv("OPENAI_API_KEY"))

	// Get test scenarios
	scenarios := agentic.GetTestScenarios()

	fmt.Printf("Running %d test scenarios...\n\n", len(scenarios))

	// Run each scenario
	var results []*agentic.TestResult
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for i, scenario := range scenarios {
		fmt.Printf("[%d/%d] Scenario %s: %s ... ",
			i+1, len(scenarios), scenario.ID, scenario.Name)

		result := agentic.RunTestScenario(ctx, scenario, executor)
		results = append(results, result)

		if result.Passed {
			fmt.Printf("✓ PASSED (%dms)\n", result.Duration.Milliseconds())
		} else {
			fmt.Printf("✗ FAILED (%dms)\n", result.Duration.Milliseconds())
			for _, errMsg := range result.Errors {
				fmt.Printf("  └─ %s\n", errMsg)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))

	// Generate HTML report
	report := agentic.NewHTMLReport(results)
	htmlContent := report.ToHTML()

	// Save report
	reportPath := filepath.Join(".", "test-report.html")
	err := os.WriteFile(reportPath, []byte(htmlContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	fmt.Printf("\n✅ Test Report Generated: %s\n", reportPath)
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  Total Tests: %d\n", report.TotalTests)
	fmt.Printf("  Passed:      %d\n", report.PassedTests)
	fmt.Printf("  Failed:      %d\n", report.FailedTests)

	if report.TotalTests > 0 {
		passRate := (report.PassedTests * 100) / report.TotalTests
		fmt.Printf("  Pass Rate:   %d%%\n", passRate)
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")

	if report.FailedTests > 0 {
		return fmt.Errorf("test suite failed: %d tests failed", report.FailedTests)
	}

	return nil
}
