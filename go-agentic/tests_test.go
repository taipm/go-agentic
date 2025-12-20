package agentic

import (
	"context"
	"testing"
)

// TestGetTestScenariosBasic tests that test scenarios are returned
func TestGetTestScenariosBasic(t *testing.T) {
	scenarios := GetTestScenarios()

	if scenarios == nil {
		t.Fatal("Expected non-nil scenarios")
	}
	if len(scenarios) == 0 {
		t.Fatal("Expected non-empty scenarios")
	}
}

// TestGetTestScenariosCount tests the number of test scenarios
func TestGetTestScenariosCount(t *testing.T) {
	scenarios := GetTestScenarios()

	// Should have at least 4 scenarios (A, B, C, D, E)
	if len(scenarios) < 4 {
		t.Errorf("Expected at least 4 scenarios, got %d", len(scenarios))
	}
}

// TestGetTestScenariosContainsRequired tests each scenario has required fields
func TestGetTestScenariosContainsRequired(t *testing.T) {
	scenarios := GetTestScenarios()

	for i, scenario := range scenarios {
		if scenario.ID == "" {
			t.Errorf("Scenario %d missing ID", i)
		}
		if scenario.Name == "" {
			t.Errorf("Scenario %d missing Name", i)
		}
		if scenario.Description == "" {
			t.Errorf("Scenario %d missing Description", i)
		}
		if scenario.UserInput == "" {
			t.Errorf("Scenario %d missing UserInput", i)
		}
		if len(scenario.ExpectedFlow) == 0 {
			t.Errorf("Scenario %d missing ExpectedFlow", i)
		}
	}
}

// TestGetTestScenariosUniqueFIDs tests scenario IDs are unique
func TestGetTestScenariosUniqueIDs(t *testing.T) {
	scenarios := GetTestScenarios()

	seen := make(map[string]bool)
	for _, scenario := range scenarios {
		if seen[scenario.ID] {
			t.Errorf("Duplicate scenario ID: %s", scenario.ID)
		}
		seen[scenario.ID] = true
	}
}

// TestRunTestScenarioBasic tests executing a test scenario
func TestRunTestScenarioBasic(t *testing.T) {
	agent1 := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      "gpt-4o-mini",
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent1},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, "test-api-key")

	scenario := &TestScenario{
		ID:           "test1",
		Name:         "Test Scenario",
		UserInput:    "test input",
		ExpectedFlow: []string{"agent1"},
		Assertions:   []string{},
	}

	ctx := context.Background()
	result := RunTestScenario(ctx, scenario, executor)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Scenario != scenario {
		t.Error("Scenario not set correctly in result")
	}
	if result.Duration == 0 {
		t.Error("Duration should be set")
	}
}

// TestRecordAgentFlowBasic tests recording agent flow
func TestRecordAgentFlowBasic(t *testing.T) {
	response := &TeamResponse{
		AgentID:   "agent1",
		AgentName: "Agent1",
	}

	flow := recordAgentFlow(response)

	if len(flow) != 1 {
		t.Errorf("Expected 1 agent in flow, got %d", len(flow))
	}
	if flow[0] != "agent1" {
		t.Error("Expected agent1 in flow")
	}
}

// TestRecordAgentFlowNilResponse tests recording with nil response
func TestRecordAgentFlowNilResponse(t *testing.T) {
	flow := recordAgentFlow(nil)

	if len(flow) != 0 {
		t.Errorf("Expected empty flow for nil response, got %d", len(flow))
	}
}

// TestValidateFlowBasic tests flow validation
func TestValidateFlowBasic(t *testing.T) {
	expected := []string{"agent1", "agent2"}
	actual := []string{"agent1", "agent2"}

	valid := validateFlow(expected, actual)

	if !valid {
		t.Error("Expected flows to match")
	}
}

// TestValidateFlowDifference tests flow validation with different flows
func TestValidateFlowDifference(t *testing.T) {
	expected := []string{"agent1", "agent2"}
	actual := []string{"agent1", "agent3"}

	valid := validateFlow(expected, actual)

	if valid {
		t.Error("Expected flows to differ")
	}
}

// TestValidateFlowDifferentLength tests flow validation with different lengths
func TestValidateFlowDifferentLength(t *testing.T) {
	expected := []string{"agent1"}
	actual := []string{"agent1", "agent2"}

	valid := validateFlow(expected, actual)

	if valid {
		t.Error("Expected flows to differ by length")
	}
}

// TestValidateFlowEmpty tests flow validation with empty flows
func TestValidateFlowEmpty(t *testing.T) {
	expected := []string{}
	actual := []string{}

	valid := validateFlow(expected, actual)

	if !valid {
		t.Error("Expected empty flows to match")
	}
}

// TestContainsQuestionsBasic tests question detection
func TestContainsQuestionsBasic(t *testing.T) {
	tests := []struct {
		content    string
		hasQuests  bool
	}{
		{"What is your issue?", true},
		{"Can you help?", true},
		{"This is a statement", false},
		{"?", true},
		{"", false},
	}

	for _, test := range tests {
		result := containsQuestions(test.content)
		if result != test.hasQuests {
			t.Errorf("For content '%s', expected %v but got %v", test.content, test.hasQuests, result)
		}
	}
}

// TestContainsAnyBasic tests string search
func TestContainsAnyBasic(t *testing.T) {
	tests := []struct {
		content  string
		strings  []string
		expected bool
	}{
		{"This contains TOOL output", []string{"TOOL"}, true},
		{"CPU usage is high", []string{"CPU", "Memory"}, true},
		{"Network diagnostics", []string{"Speed", "Network"}, true},
		{"No match here", []string{"TOOL", "CPU"}, false},
		{"", []string{"something"}, false},
	}

	for _, test := range tests {
		result := containsAny(test.content, test.strings...)
		if result != test.expected {
			t.Errorf("For content '%s' with strings %v, expected %v but got %v",
				test.content, test.strings, test.expected, result)
		}
	}
}

// TestContainsAnyEmpty tests with empty strings list
func TestContainsAnyEmpty(t *testing.T) {
	result := containsAny("content", []string{}...)

	if result {
		t.Error("Expected false when searching for no strings")
	}
}

// TestValidateAssertionsBasic tests assertion validation
func TestValidateAssertionsBasic(t *testing.T) {
	scenario := &TestScenario{
		ID:           "test1",
		ExpectedFlow: []string{"agent1"},
	}

	response := &TeamResponse{
		AgentID: "agent1",
		Content: "Response content",
	}

	errors := validateAssertions(scenario, response, []string{"agent1"})

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %v", errors)
	}
}

// TestValidateAssertionsNilResponse tests with nil response
func TestValidateAssertionsNilResponse(t *testing.T) {
	scenario := &TestScenario{
		ID:           "test1",
		ExpectedFlow: []string{"agent1"},
	}

	errors := validateAssertions(scenario, nil, []string{})

	if len(errors) == 0 {
		t.Error("Expected error for nil response")
	}
}

// TestValidateAssertionsMismatchedAgent tests with wrong agent
func TestValidateAssertionsMismatchedAgent(t *testing.T) {
	scenario := &TestScenario{
		ID:           "test1",
		ExpectedFlow: []string{"agent1"},
	}

	response := &TeamResponse{
		AgentID: "agent2",
		Content: "Response",
	}

	errors := validateAssertions(scenario, response, []string{"agent2"})

	if len(errors) == 0 {
		t.Error("Expected error for mismatched agent")
	}
}

// TestValidateAssertionsClarifierWithQuestions tests clarifier assertions
func TestValidateAssertionsClarifierWithQuestions(t *testing.T) {
	scenario := &TestScenario{
		ID:           "test1",
		ExpectedFlow: []string{"clarifier"},
	}

	response := &TeamResponse{
		AgentID: "clarifier",
		Content: "Can you provide more details?",
	}

	errors := validateAssertions(scenario, response, []string{"clarifier"})

	// Should pass because response contains question mark
	if len(errors) > 0 {
		t.Errorf("Expected no errors for clarifier with questions, got %v", errors)
	}
}

// TestValidateAssertionsClarifierNoQuestions tests clarifier without questions
func TestValidateAssertionsClarifierNoQuestions(t *testing.T) {
	scenario := &TestScenario{
		ID:           "test1",
		ExpectedFlow: []string{"clarifier"},
	}

	response := &TeamResponse{
		AgentID: "clarifier",
		Content: "Okay let me help you",
	}

	errors := validateAssertions(scenario, response, []string{"clarifier"})

	// Should fail because response doesn't contain question mark
	if len(errors) == 0 {
		t.Error("Expected error for clarifier without questions")
	}
}

// TestValidateAssertionsExecutorWithDiagnostics tests executor assertions
func TestValidateAssertionsExecutorWithDiagnostics(t *testing.T) {
	scenario := &TestScenario{
		ID:           "test1",
		ExpectedFlow: []string{"executor"},
	}

	response := &TeamResponse{
		AgentID: "executor",
		Content: "Running GetCPU diagnostic tool",
	}

	errors := validateAssertions(scenario, response, []string{"executor"})

	// Should pass because response contains diagnostic keyword
	if len(errors) > 0 {
		t.Errorf("Expected no errors for executor with diagnostics, got %v", errors)
	}
}

// TestValidateAssertionsEmptyFlow tests with empty expected flow
func TestValidateAssertionsEmptyFlow(t *testing.T) {
	scenario := &TestScenario{
		ID:           "test1",
		ExpectedFlow: []string{},
	}

	response := &TeamResponse{
		AgentID: "agent1",
		Content: "Response",
	}

	errors := validateAssertions(scenario, response, []string{})

	if len(errors) > 0 {
		t.Errorf("Expected no errors for empty flow, got %v", errors)
	}
}

// TestTestResultStructure tests TestResult fields are properly set
func TestTestResultStructure(t *testing.T) {
	scenario := &TestScenario{
		ID:           "test1",
		Name:         "Test",
		UserInput:    "input",
		ExpectedFlow: []string{"agent1"},
	}

	result := &TestResult{
		Scenario:   scenario,
		Passed:     true,
		ActualFlow: []string{"agent1"},
		Errors:     []string{},
		Warnings:   []string{},
	}

	if result.Scenario != scenario {
		t.Error("Scenario not set")
	}
	if !result.Passed {
		t.Error("Passed should be true")
	}
	if len(result.ActualFlow) != 1 {
		t.Error("ActualFlow should have 1 element")
	}
}
