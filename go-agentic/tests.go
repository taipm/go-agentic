package agentic

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TestScenario represents a test case for the crew
type TestScenario struct {
	ID           string
	Name         string
	Description  string
	UserInput    string
	ExpectedFlow []string // Expected agent sequence
	Assertions   []string // Assertions to validate
}

// TestResult tracks the result of a test
type TestResult struct {
	Scenario      *TestScenario
	Passed        bool
	Duration      time.Duration
	ActualFlow    []string
	Response      *TeamResponse
	Errors        []string
	Warnings      []string
	ExecutionLog  string
}

// GetTestScenarios returns all test scenarios
func GetTestScenarios() []*TestScenario {
	return []*TestScenario{
		// Scenario A: Vague Issue
		{
			ID:   "A",
			Name: "Vague Issue - Slow Computer",
			Description: "User reports vague problem without specific server or hostname. " +
				"System should route to Clarifier for information gathering.",
			UserInput: "Máy tính của tôi chậm lắm",
			ExpectedFlow: []string{
				"clarifier",    // Final agent in flow is Clarifier (Ngân)
			},
			Assertions: []string{
				"Clarifier asks clarifying questions (2-3 questions)",
				"System returns after Clarifier response (no auto-handoff)",
				"Response contains question marks indicating need for more info",
				"Clarifier asks about OS, hardware, recent changes",
			},
		},

		// Scenario B: Clear Issue with IP
		{
			ID:   "B",
			Name: "Clear Issue with Specific IP",
			Description: "User provides specific server IP and clear problem description. " +
				"System should route to Executor with explicit routing signal.",
			UserInput: "Server 192.168.1.50 không ping được, check cho tôi",
			ExpectedFlow: []string{
				"executor",     // Routes to Executor with IP + problem
			},
			Assertions: []string{
				"Executor runs diagnostic tools on the IP",
				"Executor provides network diagnostics",
				"Response contains tool execution results",
			},
		},

		// Scenario C: Partial Info (Port Issue)
		{
			ID:   "C",
			Name: "Specific Problem But Missing Server Info",
			Description: "User describes specific problem (port 3306) but doesn't specify which server. " +
				"System should route to Clarifier to get server hostname/IP.",
			UserInput: "Cổng 3306 không open",
			ExpectedFlow: []string{
				"clarifier",    // Routes to Clarifier for server identification
			},
			Assertions: []string{
				"Clarifier asks for server hostname/IP",
				"Clarifier asks about service type on that port",
				"System returns after Clarifier response (no auto-handoff)",
			},
		},

		// Scenario D: Network Problem with Location
		{
			ID:   "D",
			Name: "Network Problem with Location But No Server",
			Description: "User reports network issue with location but no specific server. " +
				"System should ask for clarification.",
			UserInput: "Không vào được internet từ phòng A5",
			ExpectedFlow: []string{
				"clarifier",    // Routes to Clarifier for details
			},
			Assertions: []string{
				"Clarifier asks which machine or IP",
				"Clarifier asks about connection type (wired/wireless)",
				"No auto-handoff to Executor",
			},
		},

		// Scenario E: Service Status Check with Hostname
		{
			ID:   "E",
			Name: "Service Check with Clear Hostname",
			Description: "User asks to check service status on a specific server by hostname. " +
				"System routes to Executor with explicit routing signal.",
			UserInput: "Check xem service MySQL trên server-app-01 còn chạy không",
			ExpectedFlow: []string{
				"executor",     // Routes to Executor with hostname + service check request
			},
			Assertions: []string{
				"Executor runs diagnostic tools on the hostname",
				"Executor provides service status information",
				"Response contains tool execution results",
			},
		},

		// Scenario F: Generic Help Request
		{
			ID:   "F",
			Name: "Generic Help Request",
			Description: "User provides extremely vague request with no context. " +
				"System should ask for detailed clarification.",
			UserInput: "Giúp tôi check hệ thống",
			ExpectedFlow: []string{
				"clarifier",    // Routes to Clarifier
			},
			Assertions: []string{
				"Clarifier provides structured questions (specific problem, OS, etc.)",
				"System returns waiting for detailed user response",
				"Response asks about what system and what issues",
			},
		},

		// Scenario G: CPU Usage Issue with IP
		{
			ID:   "G",
			Name: "Performance Issue with IP Address",
			Description: "User reports high CPU usage with IP address. " +
				"System routes to Executor with explicit routing signal.",
			UserInput: "CPU cao trên 192.168.1.100, cần kiểm tra",
			ExpectedFlow: []string{
				"executor",     // Routes to Executor with IP + performance issue
			},
			Assertions: []string{
				"Executor runs CPU diagnostic tools",
				"Executor provides performance analysis",
				"Response contains tool execution results",
			},
		},

		// Scenario H: Disk Space with Hostname
		{
			ID:   "H",
			Name: "Storage Issue with Hostname",
			Description: "User reports disk space issue with clear hostname. " +
				"System routes to Executor with explicit routing signal.",
			UserInput: "Ổ đĩa server-backup không còn chỗ, check ngay",
			ExpectedFlow: []string{
				"executor",     // Routes to Executor with hostname + urgency keyword
			},
			Assertions: []string{
				"Executor runs disk diagnostic tools",
				"Executor provides storage analysis",
				"Response contains tool execution results",
			},
		},

		// Scenario I: Multiple Affected Systems (Clarification Needed)
		{
			ID:   "I",
			Name: "Multiple Systems Issue - Need Clarification",
			Description: "User reports issue affecting multiple systems but vague. " +
				"System should ask for clarification about which systems.",
			UserInput: "Hệ thống bị mất kết nối",
			ExpectedFlow: []string{
				"clarifier",    // Routes to Clarifier
			},
			Assertions: []string{
				"Clarifier asks which systems are affected",
				"Clarifier asks about specific machine IPs or names",
				"Waits for detailed user response",
			},
		},

		// Scenario J: Already Specific - Multi-Step Diagnosis
		{
			ID:   "J",
			Name: "Complete Information - Full Diagnosis",
			Description: "User provides comprehensive information with IP and multiple issues. " +
				"System routes to Executor with explicit routing signal.",
			UserInput: "Server 10.0.0.25 chạy chậm, CPU cao, cần check toàn bộ",
			ExpectedFlow: []string{
				"executor",     // Routes to Executor with IP + comprehensive diagnostics request
			},
			Assertions: []string{
				"Executor runs comprehensive diagnostic tools",
				"Executor analyzes multiple system metrics",
				"Response contains tool execution results",
			},
		},
	}
}

// RunTestScenario executes a single test scenario
func RunTestScenario(ctx context.Context, scenario *TestScenario, executor *TeamExecutor) *TestResult {
	startTime := time.Now()

	result := &TestResult{
		Scenario:   scenario,
		ActualFlow: []string{},
		Errors:     []string{},
		Warnings:   []string{},
	}

	// Execute the scenario
	response, err := executor.Execute(ctx, scenario.UserInput)
	result.Duration = time.Since(startTime)
	result.Response = response

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Execution error: %v", err))
		result.Passed = false
		return result
	}

	// Record actual flow
	result.ActualFlow = recordAgentFlow(response)

	// Validate expected flow
	if !validateFlow(scenario.ExpectedFlow, result.ActualFlow) {
		result.Errors = append(result.Errors,
			fmt.Sprintf("Flow mismatch. Expected: %v, Got: %v",
				scenario.ExpectedFlow, result.ActualFlow))
	}

	// Validate assertions
	validationErrors := validateAssertions(scenario, response, result.ActualFlow)
	result.Errors = append(result.Errors, validationErrors...)

	result.Passed = len(result.Errors) == 0

	return result
}

// recordAgentFlow records which agents responded
func recordAgentFlow(response *TeamResponse) []string {
	var flow []string
	if response != nil {
		flow = append(flow, response.AgentID)
	}
	return flow
}

// validateFlow checks if actual flow matches expected
func validateFlow(expected, actual []string) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i := range expected {
		if expected[i] != actual[i] {
			return false
		}
	}
	return true
}

// validateAssertions validates scenario-specific assertions
func validateAssertions(scenario *TestScenario, response *TeamResponse, flow []string) []string {
	var errors []string

	if response == nil {
		errors = append(errors, "Response is nil")
		return errors
	}

	// Validate based on expected flow
	if len(scenario.ExpectedFlow) > 0 {
		expectedAgent := scenario.ExpectedFlow[len(scenario.ExpectedFlow)-1]
		if response.AgentID != expectedAgent {
			errors = append(errors, fmt.Sprintf("Scenario %s: Expected %s, got %s", scenario.ID, expectedAgent, response.AgentID))
		}

		// For Clarifier, expect clarifying questions
		if expectedAgent == "clarifier" && !containsQuestions(response.Content) {
			errors = append(errors, fmt.Sprintf("Scenario %s: Response should contain clarifying questions", scenario.ID))
		}

		// For Executor, expect tool execution or diagnostics
		if expectedAgent == "executor" {
			if !containsAny(response.Content, "TOOL", "GetCPU", "GetMemory", "CheckNetwork", "PingHost", "Chẩn đoán") {
				errors = append(errors, fmt.Sprintf("Scenario %s: Executor should show diagnostic output", scenario.ID))
			}
		}
	}

	return errors
}

// containsQuestions checks if response contains question marks
func containsQuestions(content string) bool {
	return len(content) > 0 && strings.Contains(content, "?")
}

// containsAny checks if content contains any of the given strings
func containsAny(content string, strs ...string) bool {
	for _, s := range strs {
		if strings.Contains(content, s) {
			return true
		}
	}
	return false
}
