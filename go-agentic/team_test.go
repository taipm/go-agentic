package agentic

import (
	"context"
	"fmt"
	"testing"
)

const (
	testModel  = "gpt-4o-mini"
	testAPIKey = "test-api-key"
	testInput  = "test input"
)

// TestNewTeamExecutorCreation tests team executor initialization
func TestNewTeamExecutorCreation(t *testing.T) {
	agents := []*Agent{
		{
			ID:          "agent1",
			Name:        "Agent1",
			Model:       testModel,
			Temperature: 0.7,
			IsTerminal:  false,
		},
		{
			ID:          "agent2",
			Name:        "Agent2",
			Model:       testModel,
			Temperature: 0.7,
			IsTerminal:  true,
		},
	}

	team := &Team{
		Agents:      agents,
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	if executor == nil {
		t.Fatal("Expected non-nil executor")
	}
	if executor.team != team {
		t.Error("Team not set correctly")
	}
	if executor.apiKey != testAPIKey {
		t.Error("API key not set correctly")
	}
	if executor.entryAgent != agents[0] {
		t.Error("Entry agent should be first non-terminal agent")
	}
	if len(executor.history) != 0 {
		t.Error("History should be empty initially")
	}
}

// TestNewTeamExecutorNoEntryAgent tests executor creation with only terminal agents
func TestNewTeamExecutorNoEntryAgent(t *testing.T) {
	agents := []*Agent{
		{
			ID:          "agent1",
			Name:        "Agent1",
			Model:       testModel,
			Temperature: 0.7,
			IsTerminal:  true,
		},
	}

	team := &Team{
		Agents:      agents,
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	if executor.entryAgent != nil {
		t.Error("Entry agent should be nil when no non-terminal agents")
	}
}

// TestExecuteNoEntryAgent tests execute fails when no entry agent found
func TestExecuteNoEntryAgent(t *testing.T) {
	agents := []*Agent{
		{
			ID:         "agent1",
			Name:       "Agent1",
			Model:      testModel,
			IsTerminal: true,
		},
	}

	team := &Team{
		Agents:      agents,
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)
	ctx := context.Background()

	result, err := executor.Execute(ctx, testInput)

	if err == nil {
		t.Error("Expected error when no entry agent")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
}

// TestExecuteStreamNoEntryAgent tests ExecuteStream fails when no entry agent found
func TestExecuteStreamNoEntryAgent(t *testing.T) {
	agents := []*Agent{
		{
			ID:         "agent1",
			Name:       "Agent1",
			Model:      testModel,
			IsTerminal: true,
		},
	}

	team := &Team{
		Agents:      agents,
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)
	ctx := context.Background()
	streamChan := make(chan *StreamEvent, 100)

	err := executor.ExecuteStream(ctx, testInput, streamChan)

	if err == nil {
		t.Error("Expected error when no entry agent")
	}
}

// TestExecuteStreamContextCancellation tests ExecuteStream respects context cancellation
func TestExecuteStreamContextCancellation(t *testing.T) {
	agents := []*Agent{
		{
			ID:          "agent1",
			Name:        "Agent1",
			Model:       testModel,
			Temperature: 0.7,
			IsTerminal:  false,
		},
	}

	team := &Team{
		Agents:      agents,
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	streamChan := make(chan *StreamEvent, 100)

	err := executor.ExecuteStream(ctx, testInput, streamChan)

	if err == nil {
		t.Error("Expected context cancellation error")
	}
}

// TestFindAgentByIDBasic tests finding agents by ID
func TestFindAgentByIDBasic(t *testing.T) {
	agent1 := &Agent{
		ID:   "agent1",
		Name: "Agent1",
	}
	agent2 := &Agent{
		ID:   "agent2",
		Name: "Agent2",
	}

	team := &Team{
		Agents:      []*Agent{agent1, agent2},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	found := executor.findAgentByID("agent2")
	if found != agent2 {
		t.Error("Expected to find agent2")
	}

	notFound := executor.findAgentByID("agent3")
	if notFound != nil {
		t.Error("Expected nil for non-existent agent")
	}
}

// TestGetAgentBehaviorExists tests retrieving agent behavior
func TestGetAgentBehaviorExists(t *testing.T) {
	agent1 := &Agent{
		ID:   "orchestrator",
		Name: "Orchestrator",
	}

	routing := &RoutingConfig{
		AgentBehaviors: map[string]AgentBehavior{
			"orchestrator": {
				WaitForSignal: true,
			},
		},
	}

	team := &Team{
		Agents:      []*Agent{agent1},
		MaxHandoffs: 10,
		Routing:     routing,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	found := executor.getAgentBehavior("orchestrator")
	if found == nil {
		t.Error("Expected to find agent behavior")
	}
	if !found.WaitForSignal {
		t.Error("Expected WaitForSignal to be true")
	}
}

// TestGetAgentBehaviorNotExists tests behavior lookup for non-existent agent
func TestGetAgentBehaviorNotExists(t *testing.T) {
	agent1 := &Agent{
		ID:   "orchestrator",
		Name: "Orchestrator",
	}

	team := &Team{
		Agents:      []*Agent{agent1},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	found := executor.getAgentBehavior("nonexistent")
	if found != nil {
		t.Error("Expected nil for non-existent behavior")
	}
}

// TestFindNextAgentBySignalBasic tests routing via signal detection
func TestFindNextAgentBySignalBasic(t *testing.T) {
	orchestrator := &Agent{
		ID:   "orchestrator",
		Name: "Orchestrator",
	}
	executor := &Agent{
		ID:   "executor",
		Name: "Executor",
	}

	routing := &RoutingConfig{
		Signals: map[string][]RoutingSignal{
			"orchestrator": {
				{
					Signal: "[ROUTE TO EXECUTOR]",
					Target: "executor",
				},
			},
		},
	}

	team := &Team{
		Agents:      []*Agent{orchestrator, executor},
		MaxHandoffs: 10,
		Routing:     routing,
	}

	teamExecutor := NewTeamExecutor(team, testAPIKey)

	agentResponse := "[ROUTE TO EXECUTOR] I need the executor to handle this"
	nextAgent := teamExecutor.findNextAgentBySignal(orchestrator, agentResponse)

	if nextAgent != executor {
		t.Error("Expected to find executor via signal")
	}
}

// TestFindNextAgentBySignalNoMatch tests signal routing with no matching signal
func TestFindNextAgentBySignalNoMatch(t *testing.T) {
	orchestrator := &Agent{
		ID:   "orchestrator",
		Name: "Orchestrator",
	}
	executor := &Agent{
		ID:   "executor",
		Name: "Executor",
	}

	routing := &RoutingConfig{
		Signals: map[string][]RoutingSignal{
			"orchestrator": {
				{
					Signal: "[ROUTE TO EXECUTOR]",
					Target: "executor",
				},
			},
		},
	}

	team := &Team{
		Agents:      []*Agent{orchestrator, executor},
		MaxHandoffs: 10,
		Routing:     routing,
	}

	teamExecutor := NewTeamExecutor(team, testAPIKey)

	agentResponse := "This is a normal response without signal"
	nextAgent := teamExecutor.findNextAgentBySignal(orchestrator, agentResponse)

	if nextAgent != nil {
		t.Error("Expected nil when no matching signal")
	}
}

// TestFindNextAgentBasic tests default agent handoff
func TestFindNextAgentBasic(t *testing.T) {
	agent1 := &Agent{
		ID:              "agent1",
		Name:            "Agent1",
		IsTerminal:      false,
		HandoffTargets:  []string{"agent2"},
	}
	agent2 := &Agent{
		ID:         "agent2",
		Name:       "Agent2",
		IsTerminal: false,
	}

	team := &Team{
		Agents:      []*Agent{agent1, agent2},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	nextAgent := executor.findNextAgent(agent1)
	if nextAgent != agent2 {
		t.Error("Expected to find agent2 as next agent")
	}
}

// TestFindNextAgentNoHandoffTarget tests no next agent when no handoff configured
func TestFindNextAgentNoHandoffTarget(t *testing.T) {
	agent1 := &Agent{
		ID:              "agent1",
		Name:            "Agent1",
		IsTerminal:      false,
		HandoffTargets:  []string{},
	}

	team := &Team{
		Agents:      []*Agent{agent1},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	nextAgent := executor.findNextAgent(agent1)
	if nextAgent != nil {
		t.Error("Expected nil when no handoff target")
	}
}

// TestFormatToolResultsBasic tests tool result formatting
func TestFormatToolResultsBasic(t *testing.T) {
	results := []ToolResult{
		{
			ToolName: "calculator",
			Status:   "success",
			Output:   "42",
		},
		{
			ToolName: "search",
			Status:   "error",
			Output:   "Connection failed",
		},
	}

	formatted := formatToolResults(results)

	if formatted == "" {
		t.Error("Expected non-empty formatted results")
	}
	if !contains(formatted, "calculator") {
		t.Error("Expected calculator tool in results")
	}
	if !contains(formatted, "search") {
		t.Error("Expected search tool in results")
	}
}

// TestFormatToolResultsEmpty tests formatting empty tool results
func TestFormatToolResultsEmpty(t *testing.T) {
	results := []ToolResult{}

	formatted := formatToolResults(results)

	if formatted == "" {
		t.Error("Expected non-empty formatted results even when empty")
	}
}

// TestExecuteCallsToolNotFound tests tool execution with missing tool
func TestExecuteCallsToolNotFound(t *testing.T) {
	agent := &Agent{
		ID:    "agent1",
		Name:  "Agent1",
		Tools: []*Tool{},
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	calls := []ToolCall{
		{
			ToolName:  "missing_tool",
			Arguments: map[string]interface{}{},
		},
	}

	results := executor.executeCalls(context.Background(), calls, agent)

	if len(results) != 1 {
		t.Fatal("Expected 1 result")
	}
	if results[0].Status != "error" {
		t.Error("Expected error status for missing tool")
	}
	if !contains(results[0].Output, "not found") {
		t.Error("Expected 'not found' in error message")
	}
}

// TestExecuteCallsToolExecution tests successful tool execution
func TestExecuteCallsToolExecution(t *testing.T) {
	tool := &Tool{
		Name:        "test_tool",
		Description: "Test tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "Tool executed", nil
		},
	}

	agent := &Agent{
		ID:    "agent1",
		Name:  "Agent1",
		Tools: []*Tool{tool},
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	calls := []ToolCall{
		{
			ToolName:  "test_tool",
			Arguments: map[string]interface{}{},
		},
	}

	results := executor.executeCalls(context.Background(), calls, agent)

	if len(results) != 1 {
		t.Fatal("Expected 1 result")
	}
	if results[0].Status != "success" {
		t.Error("Expected success status")
	}
	if results[0].Output != "Tool executed" {
		t.Error("Expected correct tool output")
	}
}

// TestExecuteCallsToolError tests tool execution with error
func TestExecuteCallsToolError(t *testing.T) {
	tool := &Tool{
		Name:        "failing_tool",
		Description: "Failing tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "", fmt.Errorf("tool execution failed")
		},
	}

	agent := &Agent{
		ID:    "agent1",
		Name:  "Agent1",
		Tools: []*Tool{tool},
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	calls := []ToolCall{
		{
			ToolName:  "failing_tool",
			Arguments: map[string]interface{}{},
		},
	}

	results := executor.executeCalls(context.Background(), calls, agent)

	if len(results) != 1 {
		t.Fatal("Expected 1 result")
	}
	if results[0].Status != "error" {
		t.Error("Expected error status")
	}
	if !contains(results[0].Output, "failed") {
		t.Error("Expected error message in output")
	}
}

// TestMaxHandoffLimitReached tests handoff configuration
func TestMaxHandoffLimitReached(t *testing.T) {
	agent1 := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		IsTerminal: false,
	}
	agent2 := &Agent{
		ID:         "agent2",
		Name:       "Agent2",
		IsTerminal: false,
	}

	team := &Team{
		Agents:      []*Agent{agent1, agent2},
		MaxHandoffs: 0,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	if executor.team.MaxHandoffs != 0 {
		t.Error("MaxHandoffs should be 0")
	}
}

// TestExecuteAddsUserInputToHistory tests Execute adds user input to history
func TestExecuteAddsUserInputToHistory(t *testing.T) {
	agent1 := &Agent{
		ID:         "agent1",
		Name:       "Agent1",
		Model:      testModel,
		IsTerminal: true,
	}

	team := &Team{
		Agents:      []*Agent{agent1},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	if len(executor.history) != 0 {
		t.Error("History should be empty initially")
	}

	// We can't test full Execute without OpenAI API, but we can verify history initialization
	// Just verify the structure is ready
	if executor.history == nil {
		t.Error("History should be initialized (non-nil)")
	}
}

// TestFindNextAgentBySignalNoConfig tests routing without signal config
func TestFindNextAgentBySignalNoConfig(t *testing.T) {
	agent := &Agent{
		ID:   "agent1",
		Name: "Agent1",
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
		// No routing configured
	}

	executor := NewTeamExecutor(team, testAPIKey)

	// With no routing config, should return nil
	result := executor.findNextAgentBySignal(agent, "some response")

	if result != nil {
		t.Error("Expected nil when no routing configured")
	}
}

// TestFindNextAgentBySignalEmptySignals tests with empty signals list
func TestFindNextAgentBySignalEmptySignals(t *testing.T) {
	agent := &Agent{
		ID:   "agent1",
		Name: "Agent1",
	}

	routing := &RoutingConfig{
		Signals: map[string][]RoutingSignal{
			"agent1": {}, // Empty signals list
		},
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
		Routing:     routing,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	// With empty signals, should return nil
	result := executor.findNextAgentBySignal(agent, "response without signal")

	if result != nil {
		t.Error("Expected nil when signals list is empty")
	}
}

// TestGetAgentBehaviorNoRouting tests getting behavior when no routing config
func TestGetAgentBehaviorNoRouting(t *testing.T) {
	agent := &Agent{
		ID:   "agent1",
		Name: "Agent1",
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
		// No routing configured
	}

	executor := NewTeamExecutor(team, testAPIKey)

	behavior := executor.getAgentBehavior("agent1")

	if behavior != nil {
		t.Error("Expected nil when no routing configured")
	}
}

// TestFindNextAgentWithMultipleTargets tests agent with multiple handoff targets
func TestFindNextAgentWithMultipleTargets(t *testing.T) {
	agent1 := &Agent{
		ID:              "agent1",
		Name:            "Agent1",
		IsTerminal:      false,
		HandoffTargets:  []string{"agent2", "agent3"},
	}
	agent2 := &Agent{
		ID:         "agent2",
		Name:       "Agent2",
		IsTerminal: false,
	}
	agent3 := &Agent{
		ID:         "agent3",
		Name:       "Agent3",
		IsTerminal: false,
	}

	team := &Team{
		Agents:      []*Agent{agent1, agent2, agent3},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	// Should find first target
	nextAgent := executor.findNextAgent(agent1)

	if nextAgent != agent2 {
		t.Error("Expected agent2 as first handoff target")
	}
}

// TestExecuteCallsMultiple tests executing multiple tool calls
func TestExecuteCallsMultiple(t *testing.T) {
	tool1 := &Tool{
		Name:        "tool1",
		Description: "Tool 1",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "Tool 1 result", nil
		},
	}

	tool2 := &Tool{
		Name:        "tool2",
		Description: "Tool 2",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "Tool 2 result", nil
		},
	}

	agent := &Agent{
		ID:    "agent1",
		Name:  "Agent1",
		Tools: []*Tool{tool1, tool2},
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	calls := []ToolCall{
		{ToolName: "tool1", Arguments: map[string]interface{}{}},
		{ToolName: "tool2", Arguments: map[string]interface{}{}},
	}

	results := executor.executeCalls(context.Background(), calls, agent)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0].Status != "success" {
		t.Error("Expected first call to succeed")
	}

	if results[1].Status != "success" {
		t.Error("Expected second call to succeed")
	}
}

// TestFormatToolResultsWithMultipleTools tests formatting multiple tool results
func TestFormatToolResultsWithMultipleTools(t *testing.T) {
	results := []ToolResult{
		{
			ToolName: "tool1",
			Status:   "success",
			Output:   "Result 1",
		},
		{
			ToolName: "tool2",
			Status:   "success",
			Output:   "Result 2",
		},
		{
			ToolName: "tool3",
			Status:   "error",
			Output:   "Error message",
		},
	}

	formatted := formatToolResults(results)

	if formatted == "" {
		t.Error("Expected non-empty formatted results")
	}

	if !contains(formatted, "tool1") {
		t.Error("Expected tool1 in formatted results")
	}

	if !contains(formatted, "tool2") {
		t.Error("Expected tool2 in formatted results")
	}

	if !contains(formatted, "tool3") {
		t.Error("Expected tool3 in formatted results")
	}
}

// TestNewTeamExecutorInitializesHistory tests history is initialized
func TestNewTeamExecutorInitializesHistory(t *testing.T) {
	agent := &Agent{
		ID:    "agent1",
		Name:  "Agent1",
		Model: testModel,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewTeamExecutor(team, testAPIKey)

	if executor.history == nil {
		t.Error("Expected history to be initialized")
	}

	if len(executor.history) != 0 {
		t.Error("Expected empty history initially")
	}
}

// TestNewCrewExecutor tests the deprecated wrapper function
func TestNewCrewExecutor(t *testing.T) {
	agent := &Agent{
		ID:    "agent1",
		Name:  "Agent1",
		Model: testModel,
	}

	team := &Team{
		Agents:      []*Agent{agent},
		MaxHandoffs: 10,
	}

	executor := NewCrewExecutor(team, testAPIKey)

	if executor == nil {
		t.Fatal("Expected non-nil executor")
	}

	if executor.team != team {
		t.Error("Expected executor to have same team")
	}

	if executor.apiKey != testAPIKey {
		t.Error("Expected executor to have same API key")
	}
}

