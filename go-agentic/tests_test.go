package agentic

import (
	"context"
	"testing"
)

// Test Story 5.1: RunTestScenario API

// TestRunTestScenarioSuccess tests successful scenario execution
func TestRunTestScenarioSuccess(t *testing.T) {
	ctx := context.Background()
	scenario := &TestScenario{
		ID:           "A",
		Name:         "Vague Issue - Slow Computer",
		Description:  "User reports vague problem",
		UserInput:    "Máy tính của tôi chậm lắm",
		ExpectedFlow: []string{"clarifier"},
		Assertions:   []string{"Response should contain questions"},
	}

	// Create a minimal team with clarifier agent
	team := &Team{
		Agents: []*Agent{
			{
				ID:   "clarifier",
				Name: "Clarifier",
				Role: "Clarifier",
			},
		},
	}

	executor := NewTeamExecutor(team, "test-api-key")

	result := RunTestScenario(ctx, scenario, executor)

	if result == nil {
		t.Fatal("RunTestScenario returned nil result")
	}

	if result.Scenario.ID != "A" {
		t.Errorf("Expected scenario ID 'A', got %q", result.Scenario.ID)
	}

	if result.Duration == 0 {
		t.Error("Expected duration > 0")
	}
}

// TestRunTestScenarioFlowMismatch tests detection of flow mismatch
func TestRunTestScenarioFlowMismatch(t *testing.T) {
	ctx := context.Background()
	scenario := &TestScenario{
		ID:           "B",
		Name:         "Clear Issue with Specific IP",
		Description:  "User provides specific server IP",
		UserInput:    "Server 192.168.1.50 không ping được",
		ExpectedFlow: []string{"executor"},
		Assertions:   []string{"Should execute diagnostic tools"},
	}

	team := &Team{
		Agents: []*Agent{
			{
				ID:   "clarifier",
				Name: "Clarifier",
				Role: "Clarifier",
			},
		},
	}

	executor := NewTeamExecutor(team, "test-api-key")

	result := RunTestScenario(ctx, scenario, executor)

	if result == nil {
		t.Fatal("RunTestScenario returned nil result")
	}

	// If flow doesn't match expected, test should detect it
	if len(result.Errors) == 0 && len(result.ActualFlow) > 0 && result.ActualFlow[0] != "executor" {
		t.Error("Expected flow mismatch error or actual flow to be 'executor'")
	}
}

// TestRunTestScenarioWithError tests error handling with empty team
func TestRunTestScenarioWithError(t *testing.T) {
	ctx := context.Background()
	scenario := &TestScenario{
		ID:           "test-error",
		Name:         "Error Test Scenario",
		Description:  "Test scenario with error",
		UserInput:    "Test input",
		ExpectedFlow: []string{"executor"},
		Assertions:   []string{},
	}

	// Create executor with no agents to trigger execution error
	team := &Team{
		Agents: []*Agent{},
	}
	executor := NewTeamExecutor(team, "test-api-key")

	result := RunTestScenario(ctx, scenario, executor)

	if result == nil {
		t.Fatal("RunTestScenario returned nil result")
	}

	// When there's no matching agent or execution fails, we should see errors
	// The test should still complete and not pass
	if result.Passed && len(result.Errors) == 0 {
		t.Error("Expected either errors or Passed to be false")
	}
}

// TestRunTestScenarioContextCancellation tests context cancellation handling
func TestRunTestScenarioContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	scenario := &TestScenario{
		ID:           "test-cancel",
		Name:         "Context Cancellation Test",
		Description:  "Test context cancellation",
		UserInput:    "Test input",
		ExpectedFlow: []string{"executor"},
		Assertions:   []string{},
	}

	team := &Team{
		Agents: []*Agent{},
	}

	executor := NewTeamExecutor(team, "test-api-key")

	result := RunTestScenario(ctx, scenario, executor)

	if result == nil {
		t.Fatal("RunTestScenario returned nil result")
	}

	// Result should complete even with cancelled context
	if result.Duration == 0 {
		t.Error("Expected duration > 0 even with cancelled context")
	}
}

// TestRunTestScenarioMissingTag tests detection of missing expected tags
func TestRunTestScenarioMissingTag(t *testing.T) {
	ctx := context.Background()
	scenario := &TestScenario{
		ID:           "test-tag",
		Name:         "Missing Tag Test",
		Description:  "Test missing expected tag",
		UserInput:    "Test input",
		ExpectedFlow: []string{"missing-agent"},
		Assertions:   []string{},
	}

	team := &Team{
		Agents: []*Agent{
			{
				ID:   "executor",
				Name: "Executor",
				Role: "Executor",
			},
		},
	}

	executor := NewTeamExecutor(team, "test-api-key")

	result := RunTestScenario(ctx, scenario, executor)

	if result == nil {
		t.Fatal("RunTestScenario returned nil result")
	}

	// Should detect flow mismatch when expected tag differs
	if len(result.ActualFlow) > 0 && result.ActualFlow[0] != "missing-agent" && len(result.Errors) == 0 {
		t.Error("Expected error when flow doesn't match")
	}
}

// Test Story 5.2: GetTestScenarios API

// TestGetTestScenariosCount tests that minimum 10 scenarios are returned
func TestGetTestScenariosCount(t *testing.T) {
	scenarios := GetTestScenarios()

	if len(scenarios) < 10 {
		t.Errorf("Expected at least 10 scenarios, got %d", len(scenarios))
	}

	if len(scenarios) > 20 {
		t.Errorf("Expected at most 20 scenarios, got %d", len(scenarios))
	}
}

// TestGetTestScenariosUnique tests that all scenario IDs are unique
func TestGetTestScenariosUnique(t *testing.T) {
	scenarios := GetTestScenarios()
	seen := make(map[string]bool)

	for _, scenario := range scenarios {
		if seen[scenario.ID] {
			t.Errorf("Duplicate scenario ID: %s", scenario.ID)
		}
		seen[scenario.ID] = true
	}

	if len(seen) != len(scenarios) {
		t.Errorf("Expected %d unique IDs, got %d", len(scenarios), len(seen))
	}
}

// TestGetTestScenariosContent tests that scenarios have required fields
func TestGetTestScenariosContent(t *testing.T) {
	scenarios := GetTestScenarios()

	for _, scenario := range scenarios {
		if scenario.ID == "" {
			t.Error("Scenario missing ID")
		}

		if scenario.Name == "" {
			t.Error("Scenario missing Name")
		}

		if scenario.Description == "" {
			t.Error("Scenario missing Description")
		}

		if scenario.UserInput == "" {
			t.Error("Scenario missing UserInput")
		}

		if len(scenario.ExpectedFlow) == 0 {
			t.Errorf("Scenario %s missing ExpectedFlow", scenario.ID)
		}

		if len(scenario.Assertions) == 0 {
			t.Errorf("Scenario %s missing Assertions", scenario.ID)
		}
	}
}

// TestGetTestScenariosNonEmpty tests that no scenario has empty values
func TestGetTestScenariosNonEmpty(t *testing.T) {
	scenarios := GetTestScenarios()

	if len(scenarios) == 0 {
		t.Fatal("No scenarios returned")
	}

	for _, scenario := range scenarios {
		// Check all required fields are non-empty strings
		if scenario.ID == "" || scenario.Name == "" || scenario.Description == "" || scenario.UserInput == "" {
			t.Errorf("Scenario %v has empty fields", scenario)
		}

		// Check flow and assertions are not empty
		if len(scenario.ExpectedFlow) == 0 || len(scenario.Assertions) == 0 {
			t.Errorf("Scenario %s missing flow or assertions", scenario.ID)
		}

		// Verify flow contains valid agent names
		for _, agent := range scenario.ExpectedFlow {
			if agent != "clarifier" && agent != "executor" {
				t.Errorf("Scenario %s has invalid agent in flow: %s", scenario.ID, agent)
			}
		}
	}
}

// TestGetTestScenariosFeatureAreas tests that scenarios cover all feature areas
func TestGetTestScenariosFeatureAreas(t *testing.T) {
	scenarios := GetTestScenarios()

	// Expected scenario IDs (A-J) covering different feature areas
	expectedIDs := map[string]bool{
		"A": false, // Vague Issue
		"B": false, // Clear with IP
		"C": false, // Partial Info
		"D": false, // Network Problem
		"E": false, // Service Check
		"F": false, // Generic Help
		"G": false, // CPU Issue
		"H": false, // Disk Space
		"I": false, // Multiple Systems
		"J": false, // Complete Info
	}

	for _, scenario := range scenarios {
		expectedIDs[scenario.ID] = true
	}

	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Expected scenario %s not found", id)
		}
	}
}
