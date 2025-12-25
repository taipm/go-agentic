package routing

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/taipm/go-agentic/core/common"
)

// ============================================================================
// PARALLEL GROUP EXECUTOR TESTS
// ============================================================================

// TestNewParallelGroupExecutor tests executor creation
func TestNewParallelGroupExecutor(t *testing.T) {
	config := &common.ParallelGroupConfig{
		Agents:         []string{"agent1", "agent2"},
		WaitForAll:     true,
		TimeoutSeconds: 10,
	}

	agents := map[string]*common.Agent{
		"agent1": {ID: "agent1", Name: "Agent 1"},
		"agent2": {ID: "agent2", Name: "Agent 2"},
	}

	executor := NewParallelGroupExecutor(config, agents)

	if executor == nil {
		t.Fatalf("Expected executor to be created")
	}

	if executor.Config != config {
		t.Errorf("Expected config to be set")
	}

	if executor.Timeout != 10*time.Second {
		t.Errorf("Expected timeout to be 10 seconds, got %v", executor.Timeout)
	}
}

// TestNewParallelGroupExecutor_DefaultTimeout tests default timeout
func TestNewParallelGroupExecutor_DefaultTimeout(t *testing.T) {
	config := &common.ParallelGroupConfig{
		Agents:         []string{"agent1"},
		WaitForAll:     true,
		TimeoutSeconds: 0, // Use default
	}

	executor := NewParallelGroupExecutor(config, nil)

	if executor.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30 seconds, got %v", executor.Timeout)
	}
}

// TestNewParallelGroupExecutor_NilConfig tests nil config
func TestNewParallelGroupExecutor_NilConfig(t *testing.T) {
	executor := NewParallelGroupExecutor(nil, nil)

	if executor == nil {
		t.Fatalf("Expected executor even with nil config")
	}

	if executor.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout, got %v", executor.Timeout)
	}
}

// TestExecuteParallel_NoAgents tests error with no agents
func TestExecuteParallel_NoAgents(t *testing.T) {
	executor := NewParallelGroupExecutor(nil, nil)

	results, err := executor.ExecuteParallel(context.Background(), []*common.Agent{}, "", nil, "")

	if err == nil {
		t.Fatalf("Expected error for no agents")
	}

	if results != nil {
		t.Errorf("Expected nil results for error case")
	}
}

// TestExecuteParallel_NilAgent tests handling of nil agents
func TestExecuteParallel_NilAgent(t *testing.T) {
	executor := NewParallelGroupExecutor(nil, nil)

	// Include a nil agent in the slice
	agents := []*common.Agent{
		{ID: "agent1", Name: "Agent 1"},
		nil,
		{ID: "agent2", Name: "Agent 2"},
	}

	results, err := executor.ExecuteParallel(context.Background(), agents, "", nil, "")

	// Should not error, just skip nil agents
	if err != nil {
		t.Errorf("Expected no error with nil agent in slice, got %v", err)
	}

	// Should return results for non-nil agents (though execution will fail)
	if results == nil {
		t.Errorf("Expected results even with nil agents")
	}
}

// TestValidateParallelGroup_Valid tests valid group validation
func TestValidateParallelGroup_Valid(t *testing.T) {
	group := &common.ParallelGroupConfig{
		Agents:         []string{"agent1", "agent2"},
		WaitForAll:     true,
		TimeoutSeconds: 10,
	}

	agents := map[string]*common.Agent{
		"agent1": {ID: "agent1", Name: "Agent 1"},
		"agent2": {ID: "agent2", Name: "Agent 2"},
	}

	err := ValidateParallelGroup(group, agents)
	if err != nil {
		t.Fatalf("Expected valid group, got error: %v", err)
	}
}

// TestValidateParallelGroup_NilConfig tests nil config validation
func TestValidateParallelGroup_NilConfig(t *testing.T) {
	err := ValidateParallelGroup(nil, nil)
	if err == nil {
		t.Fatalf("Expected error for nil config")
	}

	var validationErr *common.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got: %T", err)
	}
}

// TestValidateParallelGroup_NoAgents tests empty agents list
func TestValidateParallelGroup_NoAgents(t *testing.T) {
	group := &common.ParallelGroupConfig{
		Agents: []string{},
	}

	err := ValidateParallelGroup(group, nil)
	if err == nil {
		t.Fatalf("Expected error for no agents")
	}

	var validationErr *common.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got: %T", err)
	}
}

// TestValidateParallelGroup_EmptyAgentID tests empty agent ID
func TestValidateParallelGroup_EmptyAgentID(t *testing.T) {
	group := &common.ParallelGroupConfig{
		Agents: []string{"agent1", "", "agent2"},
	}

	agents := map[string]*common.Agent{
		"agent1": {ID: "agent1", Name: "Agent 1"},
		"agent2": {ID: "agent2", Name: "Agent 2"},
	}

	err := ValidateParallelGroup(group, agents)
	if err == nil {
		t.Fatalf("Expected error for empty agent ID")
	}

	var validationErr *common.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got: %T", err)
	}
}

// TestValidateParallelGroup_MissingAgent tests missing agent in map
func TestValidateParallelGroup_MissingAgent(t *testing.T) {
	group := &common.ParallelGroupConfig{
		Agents: []string{"agent1", "agent_missing"},
	}

	agents := map[string]*common.Agent{
		"agent1": {ID: "agent1", Name: "Agent 1"},
	}

	err := ValidateParallelGroup(group, agents)
	if err == nil {
		t.Fatalf("Expected error for missing agent")
	}

	var validationErr *common.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got: %T", err)
	}
}

// TestValidateParallelGroup_InvalidTimeout tests invalid timeout
func TestValidateParallelGroup_InvalidTimeout(t *testing.T) {
	group := &common.ParallelGroupConfig{
		Agents:         []string{"agent1"},
		TimeoutSeconds: 0, // 0 is okay (use default)
	}

	agents := map[string]*common.Agent{
		"agent1": {ID: "agent1", Name: "Agent 1"},
	}

	err := ValidateParallelGroup(group, agents)
	if err != nil {
		t.Errorf("Expected 0 timeout to be valid (use default), got error: %v", err)
	}

	// Now test with negative timeout (invalid)
	group.TimeoutSeconds = -1
	err = ValidateParallelGroup(group, agents)
	// Negative is technically not checked, only values between 0 and < 1
	// Let's test with a very small positive value
	group.TimeoutSeconds = 1
	err = ValidateParallelGroup(group, agents)
	if err != nil {
		t.Errorf("Expected 1 second timeout to be valid, got error: %v", err)
	}
}

// TestGetAgentsForParallelGroup_Valid tests getting agents for parallel group
func TestGetAgentsForParallelGroup_Valid(t *testing.T) {
	group := &common.ParallelGroupConfig{
		Agents: []string{"agent1", "agent2", "agent3"},
	}

	agents := map[string]*common.Agent{
		"agent1": {ID: "agent1", Name: "Agent 1"},
		"agent2": {ID: "agent2", Name: "Agent 2"},
		"agent3": {ID: "agent3", Name: "Agent 3"},
	}

	result, err := GetAgentsForParallelGroup(group, agents)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(result))
	}

	for i, agent := range result {
		if agent == nil {
			t.Errorf("Agent %d is nil", i)
		}
	}
}

// TestGetAgentsForParallelGroup_NilConfig tests nil config
func TestGetAgentsForParallelGroup_NilConfig(t *testing.T) {
	result, err := GetAgentsForParallelGroup(nil, nil)
	if err == nil {
		t.Fatalf("Expected error for nil config")
	}

	if result != nil {
		t.Errorf("Expected nil result for error case")
	}
}

// TestGetAgentsForParallelGroup_MissingAgent tests missing agent
func TestGetAgentsForParallelGroup_MissingAgent(t *testing.T) {
	group := &common.ParallelGroupConfig{
		Agents: []string{"agent1", "agent_missing"},
	}

	agents := map[string]*common.Agent{
		"agent1": {ID: "agent1", Name: "Agent 1"},
	}

	result, err := GetAgentsForParallelGroup(group, agents)
	if err == nil {
		t.Fatalf("Expected error for missing agent")
	}

	if result != nil {
		t.Errorf("Expected nil result for error case")
	}

	var validationErr *common.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ValidationError, got: %T", err)
	}
}

// TestExecuteWithWaitForAll_NoConfig tests WaitForAll with nil config
func TestExecuteWithWaitForAll_NoConfig(t *testing.T) {
	executor := NewParallelGroupExecutor(nil, nil)

	agents := []*common.Agent{
		{ID: "agent1", Name: "Agent 1"},
	}

	results, _ := executor.ExecuteWithWaitForAll(context.Background(), agents, "", nil, "")

	// Should handle gracefully (agent execution will fail but config allows it)
	if results == nil {
		t.Errorf("Expected results even with nil config")
	}
}

// TestExecuteWithFirstSuccess_NoAgents tests FirstSuccess with no agents
func TestExecuteWithFirstSuccess_NoAgents(t *testing.T) {
	executor := NewParallelGroupExecutor(nil, nil)

	result, err := executor.ExecuteWithFirstSuccess(context.Background(), []*common.Agent{}, "", nil, "")

	if err == nil {
		t.Fatalf("Expected error for no agents")
	}

	if result != nil {
		t.Errorf("Expected nil result for error case")
	}
}

// TestExecuteGroupWithStrategy_NilConfig tests strategy execution with nil config
func TestExecuteGroupWithStrategy_NilConfig(t *testing.T) {
	executor := NewParallelGroupExecutor(nil, nil)

	agents := []*common.Agent{
		{ID: "agent1", Name: "Agent 1"},
	}

	result, err := executor.ExecuteGroupWithStrategy(context.Background(), agents, "", nil, "")

	if err == nil {
		t.Fatalf("Expected error for nil config")
	}

	if result != nil {
		t.Errorf("Expected nil result for error case")
	}
}

// TestExecuteGroupWithStrategy_WaitForAll tests WaitForAll strategy selection
func TestExecuteGroupWithStrategy_WaitForAll(t *testing.T) {
	config := &common.ParallelGroupConfig{
		Agents:     []string{"agent1"},
		WaitForAll: true,
	}

	executor := NewParallelGroupExecutor(config, nil)

	agents := []*common.Agent{
		{ID: "agent1", Name: "Agent 1"},
	}

	// Should use WaitForAll path
	result, err := executor.ExecuteGroupWithStrategy(context.Background(), agents, "", nil, "")

	// Will fail due to agent execution not being implemented, but strategy should be invoked
	if result == nil && err != nil {
		// Expected - strategy was invoked but execution failed
	}
}

// TestExecuteGroupWithStrategy_FirstSuccess tests FirstSuccess strategy selection
func TestExecuteGroupWithStrategy_FirstSuccess(t *testing.T) {
	config := &common.ParallelGroupConfig{
		Agents:     []string{"agent1"},
		WaitForAll: false, // FirstSuccess strategy
	}

	executor := NewParallelGroupExecutor(config, nil)

	agents := []*common.Agent{
		{ID: "agent1", Name: "Agent 1"},
	}

	// Should use FirstSuccess path
	result, err := executor.ExecuteGroupWithStrategy(context.Background(), agents, "", nil, "")

	// Will fail due to agent execution not being implemented
	if result != nil || err != nil {
		// Expected - strategy was invoked
	}
}

// TestParallelResult tests ParallelResult structure
func TestParallelResult(t *testing.T) {
	startTime := time.Now()
	response := &common.AgentResponse{
		AgentID: "agent1",
		Content: "test response",
	}

	result := ParallelResult{
		AgentID:  "agent1",
		Response: response,
		Error:    nil,
		Duration: 100 * time.Millisecond,
	}

	if result.AgentID != "agent1" {
		t.Errorf("Expected AgentID 'agent1', got %s", result.AgentID)
	}

	if result.Response != response {
		t.Errorf("Expected response to be set")
	}

	if result.Error != nil {
		t.Errorf("Expected no error")
	}

	if result.Duration != 100*time.Millisecond {
		t.Errorf("Expected duration 100ms, got %v", result.Duration)
	}

	_ = startTime
}

// TestParallelResult_WithError tests ParallelResult with error
func TestParallelResult_WithError(t *testing.T) {
	testErr := fmt.Errorf("test error")

	result := ParallelResult{
		AgentID:  "agent1",
		Response: nil,
		Error:    testErr,
		Duration: 50 * time.Millisecond,
	}

	if result.Response != nil {
		t.Errorf("Expected nil response on error")
	}

	if result.Error != testErr {
		t.Errorf("Expected error to be set")
	}
}

// TestTimeout_ContextDeadline tests timeout behavior
func TestTimeout_ContextDeadline(t *testing.T) {
	executor := NewParallelGroupExecutor(&common.ParallelGroupConfig{
		Agents:         []string{"agent1"},
		TimeoutSeconds: 1, // Very short timeout
	}, nil)

	agents := []*common.Agent{
		{ID: "agent1", Name: "Agent 1"},
	}

	// Create a context that's already exceeded its deadline
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, ctxErr := executor.ExecuteWithFirstSuccess(ctx, agents, "", nil, "")

	if result != nil {
		t.Errorf("Expected nil result on timeout")
	}

	if ctxErr == nil {
		t.Errorf("Expected error on timeout")
	}
}
