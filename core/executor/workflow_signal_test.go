package executor

import (
	"strings"
	"testing"

	"github.com/taipm/go-agentic/core/common"
)

// ============================================================================
// TEST: Prove Bug in Signal Routing
// ============================================================================

// TestHandleAgentResponse_IgnoresSignalRouting proves the bug
func TestHandleAgentResponse_IgnoresSignalRouting(t *testing.T) {
	t.Logf("\nTEST: HandleAgentResponse_IgnoresSignalRouting")
	t.Logf("PURPOSE: Prove that HandleAgentResponse doesn't use signal-based routing")
	t.Logf("EXPECTED: shouldContinue=true, nextAgent='student'")
	t.Logf("ACTUAL: shouldContinue=false, no routing\n")

	// Create agents
	teacher := &common.Agent{
		ID:         "teacher",
		Name:       "Teacher",
		Role:       "Teacher",
		IsTerminal: false,
	}
	student := &common.Agent{
		ID:         "student",
		Name:       "Student",
		Role:       "Student",
		IsTerminal: false,
	}

	agents := map[string]*common.Agent{
		"teacher": teacher,
		"student": student,
	}

	// Routing with signal rule
	routing := &common.RoutingConfig{
		Signals: map[string][]common.RoutingSignal{
			"teacher": {
				common.RoutingSignal{
					Signal: "[QUESTION]",
					Target: "student",
				},
			},
		},
	}

	// Create flow
	flow := NewExecutionFlow(teacher, 10, 5)

	// Response with signal
	response := &common.AgentResponse{
		AgentID:   "teacher",
		AgentName: "Teacher",
		Content:   "What is 2+2? [QUESTION]",
		Signals:   []string{"[QUESTION]"},
		ToolCalls: []common.ToolCall{},
	}

	t.Logf("TEST SETUP:")
	t.Logf("  - Current agent: %s", flow.CurrentAgent.ID)
	t.Logf("  - Response signals: %v", response.Signals)
	t.Logf("  - Routing rule: [QUESTION] -> student")
	t.Logf("  - Routing config provided: YES\n")

	// Call HandleAgentResponse
	t.Logf("CALLING: flow.HandleAgentResponse(response, routing, agents)")
	shouldContinue, err := flow.HandleAgentResponse(response, routing, agents)

	if err != nil {
		t.Logf("Error returned: %v", err)
	}

	// RESULT
	t.Logf("\nRESULT:")
	t.Logf("  - shouldContinue: %v (expected: true)", shouldContinue)
	t.Logf("  - current agent: %s (expected: student)", flow.CurrentAgent.ID)

	if !shouldContinue {
		t.Errorf("\nBUG PROVEN: HandleAgentResponse returned false")
		t.Errorf("  Expected: shouldContinue=true (route to student via [QUESTION] signal)")
		t.Errorf("  Actual: shouldContinue=false (workflow stops)")
		t.Errorf("")
		t.Errorf("ROOT CAUSE: workflow.go line 113-157")
		t.Errorf("  - DetermineNextAgent() doesn't check response.Signals")
		t.Errorf("  - Only checks IsTerminal and HandoffTargets")
		t.Errorf("  - Signal-based routing config is IGNORED")
		return
	}

	t.Logf("\nStatus: If you see this, routing worked (unexpected in buggy code)")
}

// TestQuizExamBugAnalysis describes the bug chain
func TestQuizExamBugAnalysis(t *testing.T) {
	sep := strings.Repeat("=", 80)
	t.Logf("\n%s", sep)
	t.Logf("ANALYSIS: Why Quiz Exam Only Executes First Agent")
	t.Logf("%s", sep)

	t.Logf("\nOBSERVED ISSUE:")
	t.Logf("  - Quiz exam runs: make run")
	t.Logf("  - Teacher emits '[QUESTION]' signal")
	t.Logf("  - Then: Program stops (only Teacher executed, Student never called)")

	t.Logf("\nWHAT ACTUALLY HAPPENS:")
	t.Logf("")
	t.Logf("  1. main() -> executor.ExecuteStream() -> crew.executeWorkflow()")
	t.Logf("  2. executeWorkflow() calls: flow.ExecuteWithCallbacks(ctx, handler, apiKey, onStep, agents, routing)")
	t.Logf("     NOTE: No SignalRegistry parameter passed")
	t.Logf("")
	t.Logf("  3. ExecuteWithCallbacks() line 195-225 does:")
	t.Logf("     for {")
	t.Logf("       response = ExecuteWorkflowStep(ctx, handler, apiKey)")
	t.Logf("       shouldContinue = HandleAgentResponse(response, routing, agents)")
	t.Logf("       if !shouldContinue { break }")
	t.Logf("     }")
	t.Logf("")
	t.Logf("  4. First iteration:")
	t.Logf("     - ExecuteWorkflowStep() executes Teacher")
	t.Logf("     - Teacher emits '[QUESTION]' signal")
	t.Logf("     - HandleAgentResponse() called")
	t.Logf("     - DetermineNextAgent() returns IsTerminal=true (WRONG!)")
	t.Logf("     - shouldContinue = false")
	t.Logf("     - LOOP BREAKS")
	t.Logf("")
	t.Logf("  5. ExecuteWithCallbacks() returns (workflow ends)")

	t.Logf("\nWHY DetermineNextAgent() FAILS:")
	t.Logf("  - Located: workflow.go line 124")
	t.Logf("  - Doesn't check response.Signals")
	t.Logf("  - Only checks IsTerminal and HandoffTargets")
	t.Logf("  - Signal routing rules are IGNORED")

	t.Logf("\nWHY Signal REGISTRY NOT USED:")
	t.Logf("  - ExecuteWorkflowStep() line 82: passes nil as signalRegistry")
	t.Logf("  - workflow.ExecuteWorkflow() has no signal registry")
	t.Logf("  - Signal processing bypassed completely")

	t.Logf("\nBUG CHAIN (7 steps):")

	bugs := []string{
		"1. ExecutionFlow missing SignalRegistry field",
		"2. ExecuteWithCallbacks missing SignalRegistry param",
		"3. ExecuteWorkflowStep doesn't accept SignalRegistry",
		"4. ExecuteWorkflowStep passes nil to workflow.ExecuteWorkflow()",
		"5. workflow.executeAgent() can't process signals (registry is nil)",
		"6. HandleAgentResponse() uses old DetermineNextAgent() API",
		"7. DetermineNextAgent() ignores response.Signals",
	}

	for _, bug := range bugs {
		t.Logf("  %s", bug)
	}

	t.Errorf("\nCONCLUSION: Signal-based routing completely non-functional")
	t.Errorf("  Cannot be fixed with simple logic changes")
	t.Errorf("  Requires architectural changes to pass SignalRegistry through execution chain")
}

// TestExecutionChainBugLocations shows where bugs are in code
func TestExecutionChainBugLocations(t *testing.T) {
	t.Logf("\nBUG LOCATIONS IN SOURCE CODE:")

	locations := []struct {
		file string
		line int
		code string
		fix  string
	}{
		{"executor/workflow.go", 13, "type ExecutionFlow struct {...}", "Add SignalRegistry field"},
		{"executor/workflow.go", 185, "func (ef *ExecutionFlow) ExecuteWithCallbacks(...)", "Add SignalRegistry parameter"},
		{"executor/workflow.go", 49, "func (ef *ExecutionFlow) ExecuteWorkflowStep(...)", "Add SignalRegistry parameter"},
		{"executor/workflow.go", 76, "response, err := workflow.ExecuteWorkflow(..., nil, ...)", "Pass SignalRegistry instead of nil"},
		{"workflow/execution.go", 30, "func ExecuteWorkflow(ctx context.Context, ...)", "Already has signalRegistry param"},
		{"executor/workflow.go", 113, "func (ef *ExecutionFlow) HandleAgentResponse(...)", "Use DetermineNextAgentWithSignals()"},
		{"executor/workflow.go", 124, "decision, err := workflow.DetermineNextAgent(...)", "Change to DetermineNextAgentWithSignals()"},
	}

	for _, loc := range locations {
		t.Logf("\n  [%s:%d]", loc.file, loc.line)
		t.Logf("    Code: %s", loc.code)
		t.Logf("    Fix: %s", loc.fix)
	}

	t.Errorf("\nAll 7 locations must be fixed for signal routing to work")
}
