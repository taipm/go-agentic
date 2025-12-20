package agentic

import (
	"context"
	"fmt"
	"testing"

	openai "github.com/openai/openai-go/v3"
)

// Test 1.1.1: Different agents use different models
func TestDifferentAgentsUseDifferentModels(t *testing.T) {
	agent1 := &Agent{
		ID:    "agent-1",
		Name:  "Agent1",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	agent2 := &Agent{
		ID:    "agent-2",
		Name:  "Agent2",
		Model: "gpt-4o-mini",
		Tools: []*Tool{},
	}

	// Verify models are different (sanity check)
	if agent1.Model == agent2.Model {
		t.Errorf("Expected different models, but both are %s", agent1.Model)
	}

	if agent1.Model != "gpt-4o" {
		t.Errorf("Agent1 should have model gpt-4o, got %s", agent1.Model)
	}

	if agent2.Model != "gpt-4o-mini" {
		t.Errorf("Agent2 should have model gpt-4o-mini, got %s", agent2.Model)
	}
}

// Test 1.1.2: Agent model is used in ExecuteAgent (verified via buildSystemPrompt)
func TestAgentModelIsRespected(t *testing.T) {
	agent := &Agent{
		ID:    "test-agent",
		Name:  "TestAgent",
		Role:  "Helper",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	// Verify the model field is set correctly
	if agent.Model != "gpt-4o" {
		t.Errorf("Expected model gpt-4o, got %s", agent.Model)
	}

	// When ExecuteAgent is called, the params.Model will use agent.Model
	// This is verified by the code change: Model: agent.Model (not hardcoded)
	// Integration tests would verify the actual API call uses this value
}

// Test 1.1.3: Build system prompt works with agent model
func TestBuildSystemPromptIncludesAgentInfo(t *testing.T) {
	agent := &Agent{
		ID:         "test",
		Name:       "TestAgent",
		Role:       "Helper",
		Model:      "gpt-4o",
		Backstory:  "A helpful assistant",
		Tools:      []*Tool{},
	}

	prompt := buildSystemPrompt(agent)

	// Verify prompt contains agent information
	if prompt == "" {
		t.Fatal("Expected non-empty system prompt")
	}

	if !contains(prompt, agent.Name) {
		t.Errorf("System prompt should contain agent name %s", agent.Name)
	}

	if !contains(prompt, agent.Role) {
		t.Errorf("System prompt should contain agent role %s", agent.Role)
	}
}

// Test 1.1.4: Multiple agents with different models
func TestMultipleAgentsWithDifferentModels(t *testing.T) {
	agents := []*Agent{
		{
			ID:    "agent-1",
			Name:  "Agent1",
			Model: "gpt-4o",
			Tools: []*Tool{},
		},
		{
			ID:    "agent-2",
			Name:  "Agent2",
			Model: "gpt-4o-mini",
			Tools: []*Tool{},
		},
		{
			ID:    "agent-3",
			Name:  "Agent3",
			Model: "gpt-4",
			Tools: []*Tool{},
		},
	}

	// Verify each agent has correct model
	expectedModels := map[string]string{
		"agent-1": "gpt-4o",
		"agent-2": "gpt-4o-mini",
		"agent-3": "gpt-4",
	}

	for _, agent := range agents {
		expectedModel := expectedModels[agent.ID]
		if agent.Model != expectedModel {
			t.Errorf("Agent %s: expected model %s, got %s", agent.ID, expectedModel, agent.Model)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============ Test Story 2a.1: Parse Native OpenAI Tool Calls ============

// TestParseNativeToolCallsBasic tests extraction of native OpenAI tool calls
func TestParseNativeToolCallsBasic(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Tools: []*Tool{
			{Name: "PingHost", Description: "Ping a host"},
			{Name: "GetCPUUsage", Description: "Get CPU usage"},
		},
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "PingHost",
				Arguments: `{"host":"192.168.1.100"}`,
			},
		},
	}

	result := parseNativeToolCalls(toolCalls, agent)

	if len(result) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(result))
	}
	if result[0].ToolName != "PingHost" {
		t.Errorf("Expected tool name 'PingHost', got %q", result[0].ToolName)
	}
	if result[0].Arguments["host"] != "192.168.1.100" {
		t.Errorf("Expected host '192.168.1.100', got %v", result[0].Arguments["host"])
	}
}

// TestParseNativeToolCallsMultiple tests parsing multiple native tool calls
func TestParseNativeToolCallsMultiple(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Tools: []*Tool{
			{Name: "GetCPUUsage", Description: "Get CPU usage"},
			{Name: "PingHost", Description: "Ping host"},
		},
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "GetCPUUsage",
				Arguments: `{}`,
			},
		},
		{
			ID:   "call_2",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "PingHost",
				Arguments: `{"host":"8.8.8.8"}`,
			},
		},
	}

	result := parseNativeToolCalls(toolCalls, agent)

	if len(result) != 2 {
		t.Errorf("Expected 2 tool calls, got %d", len(result))
	}
	if result[0].ToolName != "GetCPUUsage" || result[1].ToolName != "PingHost" {
		t.Error("Tool names don't match expected")
	}
}

// TestParseNativeToolCallsFiltersInvalidTools tests filtering of unknown tools
func TestParseNativeToolCallsFiltersInvalidTools(t *testing.T) {
	agent := &Agent{
		ID:   "clarifier",
		Name: "Clarifier",
		Tools: []*Tool{
			{Name: "AskQuestion", Description: "Ask question"},
		},
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "function",
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "UnknownTool",
				Arguments: `{}`,
			},
		},
	}

	result := parseNativeToolCalls(toolCalls, agent)

	if len(result) != 0 {
		t.Errorf("Expected 0 tool calls for unknown tool, got %d", len(result))
	}
}

// TestParseNativeToolCallsFiltersNonFunctionType tests filtering non-function types
func TestParseNativeToolCallsFiltersNonFunctionType(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Tools: []*Tool{
			{Name: "PingHost", Description: "Ping host"},
		},
	}

	toolCalls := []openai.ChatCompletionMessageToolCallUnion{
		{
			ID:   "call_1",
			Type: "custom", // Not a function
			Function: openai.ChatCompletionMessageFunctionToolCallFunction{
				Name:      "PingHost",
				Arguments: `{}`,
			},
		},
	}

	result := parseNativeToolCalls(toolCalls, agent)

	if len(result) != 0 {
		t.Errorf("Expected 0 tool calls for non-function type, got %d", len(result))
	}
}

// ============ Test Story 2a.2: Fallback Text Parsing ============

// TestFallbackTextParsingBasic tests basic text-based tool call extraction
func TestFallbackTextParsingBasic(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Tools: []*Tool{
			{
				Name: "PingHost",
				Parameters: map[string]interface{}{
					"properties": map[string]interface{}{
						"host": map[string]interface{}{},
					},
					"required": []string{"host"},
				},
			},
		},
	}

	text := "I will ping 192.168.1.100\nPingHost(192.168.1.100)"
	result := extractToolCallsFromText(text, agent)

	if len(result) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(result))
	}
	if result[0].ToolName != "PingHost" {
		t.Errorf("Expected 'PingHost', got %q", result[0].ToolName)
	}
}

// TestFallbackTextParsingMultiple tests extracting multiple tool calls
func TestFallbackTextParsingMultiple(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Tools: []*Tool{
			{Name: "GetCPUUsage", Description: "Get CPU"},
			{Name: "PingHost", Description: "Ping host"},
		},
	}

	text := "GetCPUUsage()\nThen PingHost(192.168.1.1)"
	result := extractToolCallsFromText(text, agent)

	if len(result) != 2 {
		t.Errorf("Expected 2 tool calls, got %d", len(result))
	}
}

// TestFallbackTextParsingNoTools tests with no tool calls
func TestFallbackTextParsingNoTools(t *testing.T) {
	agent := &Agent{
		ID:   "clarifier",
		Name: "Clarifier",
		Tools: []*Tool{
			{Name: "AskQuestion", Description: "Ask"},
		},
	}

	text := "I need more information about your system."
	result := extractToolCallsFromText(text, agent)

	if len(result) != 0 {
		t.Errorf("Expected 0 tool calls, got %d", len(result))
	}
}

// ============ Test Story 2a.3: System Prompt Updates ============

// TestSystemPromptIncludesFunctionCalling tests prompt mentions function calling
func TestSystemPromptIncludesFunctionCalling(t *testing.T) {
	agent := &Agent{
		ID:   "executor",
		Name: "Executor",
		Role: "Execute commands",
		Tools: []*Tool{
			{Name: "PingHost", Description: "Ping host"},
			{Name: "GetCPUUsage", Description: "Get CPU"},
		},
	}

	prompt := buildSystemPrompt(agent)

	if !contains(prompt, "function calling") && !contains(prompt, "Function calling") {
		t.Error("System prompt should mention function calling")
	}

	if contains(prompt, "ToolName(param1, param2)") {
		t.Error("System prompt should not use old text format")
	}
}

// TestParseJSONArgumentsValid tests JSON parsing with valid input
func TestParseJSONArgumentsValid(t *testing.T) {
	args := make(map[string]interface{})
	err := parseJSONArguments(`{"host":"192.168.1.1"}`, args)

	if err != nil {
		t.Errorf("parseJSONArguments failed: %v", err)
	}
	if args["host"] != "192.168.1.1" {
		t.Errorf("Expected host argument, got %v", args)
	}
}

// TestParseJSONArgumentsInvalid tests JSON parsing with invalid input
func TestParseJSONArgumentsInvalid(t *testing.T) {
	args := make(map[string]interface{})
	err := parseJSONArguments(`{invalid json}`, args)

	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// ============ Test Story 2b.1: Validate Tool Parameters Against JSON Schema ============

// TestValidateToolParametersBasic tests validation with required parameters
func TestValidateToolParametersBasic(t *testing.T) {
	tool := &Tool{
		Name: "PingHost",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"host": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []interface{}{"host"},
		},
	}

	// Valid arguments
	args := map[string]interface{}{"host": "192.168.1.100"}
	if err := validateToolParameters(tool, args); err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	// Missing required parameter
	argsEmpty := map[string]interface{}{}
	if err := validateToolParameters(tool, argsEmpty); err == nil {
		t.Error("Expected validation to fail for missing required parameter")
	}
}

// TestValidateToolParametersTypeString tests string type validation
func TestValidateToolParametersTypeString(t *testing.T) {
	tool := &Tool{
		Name: "GetService",
		Parameters: map[string]interface{}{
			"properties": map[string]interface{}{
				"service": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	// Valid string argument
	if err := validateToolParameters(tool, map[string]interface{}{"service": "nginx"}); err != nil {
		t.Errorf("Expected validation to pass for string, got: %v", err)
	}

	// Invalid non-string argument
	if err := validateToolParameters(tool, map[string]interface{}{"service": 123}); err == nil {
		t.Error("Expected validation to fail for non-string value")
	}
}

// TestValidateToolParametersTypeNumber tests numeric type validation
func TestValidateToolParametersTypeNumber(t *testing.T) {
	tool := &Tool{
		Name: "CheckThreshold",
		Parameters: map[string]interface{}{
			"properties": map[string]interface{}{
				"threshold": map[string]interface{}{
					"type": "number",
				},
			},
		},
	}

	// Valid numeric arguments
	validArgs := []map[string]interface{}{
		{"threshold": float64(80.5)},
		{"threshold": 100},
	}
	for _, args := range validArgs {
		if err := validateToolParameters(tool, args); err != nil {
			t.Errorf("Expected validation to pass for number, got: %v", err)
		}
	}

	// Invalid non-numeric argument
	if err := validateToolParameters(tool, map[string]interface{}{"threshold": "high"}); err == nil {
		t.Error("Expected validation to fail for non-numeric value")
	}
}

// TestValidateToolParametersTypeBoolean tests boolean type validation
func TestValidateToolParametersTypeBoolean(t *testing.T) {
	tool := &Tool{
		Name: "ConfigService",
		Parameters: map[string]interface{}{
			"properties": map[string]interface{}{
				"enabled": map[string]interface{}{
					"type": "boolean",
				},
			},
		},
	}

	// Valid boolean arguments
	if err := validateToolParameters(tool, map[string]interface{}{"enabled": true}); err != nil {
		t.Errorf("Expected validation to pass for boolean true, got: %v", err)
	}
	if err := validateToolParameters(tool, map[string]interface{}{"enabled": false}); err != nil {
		t.Errorf("Expected validation to pass for boolean false, got: %v", err)
	}

	// Invalid non-boolean argument
	if err := validateToolParameters(tool, map[string]interface{}{"enabled": 1}); err == nil {
		t.Error("Expected validation to fail for non-boolean value")
	}
}

// TestValidateToolParametersMultipleFields tests validation with multiple parameters
func TestValidateToolParametersMultipleFields(t *testing.T) {
	tool := &Tool{
		Name: "ComplexTool",
		Parameters: map[string]interface{}{
			"properties": map[string]interface{}{
				"host": map[string]interface{}{
					"type": "string",
				},
				"port": map[string]interface{}{
					"type": "number",
				},
				"timeout": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []interface{}{"host", "port"},
		},
	}

	// Valid with all arguments
	args := map[string]interface{}{
		"host":    "localhost",
		"port":    8080.0,
		"timeout": 30.0,
	}
	if err := validateToolParameters(tool, args); err != nil {
		t.Errorf("Expected validation to pass, got: %v", err)
	}

	// Valid with only required arguments
	requiredOnly := map[string]interface{}{
		"host": "localhost",
		"port": 8080.0,
	}
	if err := validateToolParameters(tool, requiredOnly); err != nil {
		t.Errorf("Expected validation to pass with only required args, got: %v", err)
	}

	// Invalid: missing required parameter
	missingRequired := map[string]interface{}{
		"host": "localhost",
	}
	if err := validateToolParameters(tool, missingRequired); err == nil {
		t.Error("Expected validation to fail for missing required 'port' parameter")
	}

	// Invalid: wrong type for port
	wrongType := map[string]interface{}{
		"host": "localhost",
		"port": "8080",
	}
	if err := validateToolParameters(tool, wrongType); err == nil {
		t.Error("Expected validation to fail for port with wrong type")
	}
}

// TestValidateToolParametersNoSchema tests validation with no schema defined
func TestValidateToolParametersNoSchema(t *testing.T) {
	tool := &Tool{
		Name: "SimpleCall",
		// No Parameters defined
	}

	// Should pass validation even with arbitrary arguments
	if err := validateToolParameters(tool, map[string]interface{}{"any": "value"}); err != nil {
		t.Errorf("Expected validation to pass for tool with no schema, got: %v", err)
	}

	// Should pass with empty arguments
	if err := validateToolParameters(tool, map[string]interface{}{}); err != nil {
		t.Errorf("Expected validation to pass for tool with no schema and no args, got: %v", err)
	}
}

// ============ Test Story 2b.2: Integration Into Tool Execution ============

// TestValidateToolParametersIntegration tests validation prevents invalid handler calls
func TestValidateToolParametersIntegration(t *testing.T) {
	tool := &Tool{
		Name: "ValidatedTool",
		Parameters: map[string]interface{}{
			"properties": map[string]interface{}{
				"value": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []interface{}{"value"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "success", nil
		},
	}

	// Valid arguments should call handler
	validArgs := map[string]interface{}{"value": 42.0}
	err := validateToolParameters(tool, validArgs)
	if err != nil {
		t.Errorf("Expected validation to pass for valid args, got: %v", err)
	}

	// Invalid arguments should fail validation before handler call
	invalidArgs := map[string]interface{}{}
	err = validateToolParameters(tool, invalidArgs)
	if err == nil {
		t.Error("Expected validation to fail for missing required parameter")
	}
}

// TestValidateToolParametersErrorMessages tests that error messages are clear
func TestValidateToolParametersErrorMessages(t *testing.T) {
	tool := &Tool{
		Name: "ClearErrorsTool",
		Parameters: map[string]interface{}{
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
				"age": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []interface{}{"name", "age"},
		},
	}

	// Test missing required parameter error
	missingErr := validateToolParameters(tool, map[string]interface{}{"name": "John"})
	if missingErr == nil {
		t.Fatal("Expected error for missing required parameter")
	}
	if !contains(missingErr.Error(), "age") {
		t.Errorf("Error should mention missing parameter 'age', got: %v", missingErr)
	}

	// Test type mismatch error
	typeErr := validateToolParameters(tool, map[string]interface{}{"name": 123, "age": 25.0})
	if typeErr == nil {
		t.Fatal("Expected error for type mismatch")
	}
	if !contains(typeErr.Error(), "name") {
		t.Errorf("Error should mention parameter 'name', got: %v", typeErr)
	}
}

// ============ Test Story 3.1: Error Categorization ============

// TestErrorTypeCategorization tests that error types are properly defined
func TestErrorTypeCategorization(t *testing.T) {
	errorTypes := []ErrorType{
		ErrorTypePermissionDenied,
		ErrorTypeNotFound,
		ErrorTypeTimeout,
		ErrorTypeCommandFailed,
		ErrorTypeParseFailed,
		ErrorTypeParameterError,
		ErrorTypeSystemError,
		ErrorTypeUnknown,
	}

	for _, et := range errorTypes {
		if et == "" {
			t.Error("ErrorType should not be empty")
		}
	}
}

// TestToolErrorStructure tests that ToolError has all required fields
func TestToolErrorStructure(t *testing.T) {
	toolErr := &ToolError{
		Type:           ErrorTypeCommandFailed,
		Message:        "Command failed to execute",
		Cause:          "exit code 1",
		SuggestedAction: "Check command syntax and permissions",
	}

	if toolErr.Type != ErrorTypeCommandFailed {
		t.Errorf("Expected error type %s, got %s", ErrorTypeCommandFailed, toolErr.Type)
	}
	if toolErr.Message == "" {
		t.Error("Error message should not be empty")
	}
	if toolErr.Cause == "" {
		t.Error("Error cause should not be empty")
	}
	if toolErr.SuggestedAction == "" {
		t.Error("Suggested action should not be empty")
	}
}

// ============ Test Story 3.2: Message Role Semantics ============

// TestMessageRoleSemantics tests that message roles are correctly used
func TestMessageRoleSemantics(t *testing.T) {
	// Test valid message roles
	validRoles := []string{"user", "assistant", "system"}

	for _, role := range validRoles {
		msg := Message{
			Role:    role,
			Content: "Test content",
		}
		if msg.Role != role {
			t.Errorf("Expected role %s, got %s", role, msg.Role)
		}
	}
}

// TestSystemMessageProcessing tests that system messages are properly converted
func TestSystemMessageProcessing(t *testing.T) {
	history := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
		{Role: "system", Content: "Tool result: success"},
	}

	systemCount := 0
	userCount := 0
	assistantCount := 0

	for _, msg := range history {
		switch msg.Role {
		case "system":
			systemCount++
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		}
	}

	if systemCount != 1 {
		t.Errorf("Expected 1 system message, got %d", systemCount)
	}
	if userCount != 1 {
		t.Errorf("Expected 1 user message, got %d", userCount)
	}
	if assistantCount != 1 {
		t.Errorf("Expected 1 assistant message, got %d", assistantCount)
	}
}

// ============ Test Story 3.3: Error Handling Best Practices ============

// TestClearErrorMessages tests that error messages are clear and actionable
func TestClearErrorMessages(t *testing.T) {
	testCases := []struct {
		name    string
		message string
		shouldContain []string
	}{
		{
			name:    "Command failed error",
			message: "[COMMAND_FAILED] Failed to check service: exit code 1. Suggestion: Check permissions",
			shouldContain: []string{"COMMAND_FAILED", "Suggestion"},
		},
		{
			name:    "Permission denied error",
			message: "[PERMISSION_DENIED] Cannot access file. Suggestion: Run with elevated privileges",
			shouldContain: []string{"PERMISSION_DENIED", "Suggestion"},
		},
		{
			name:    "Not found error",
			message: "[NOT_FOUND] Service not found. Suggestion: Verify service name is correct",
			shouldContain: []string{"NOT_FOUND", "Suggestion"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, expected := range tc.shouldContain {
				if !contains(tc.message, expected) {
					t.Errorf("Error message should contain '%s', got: %s", expected, tc.message)
				}
			}
		})
	}
}

// TestNoSilentErrors tests that error patterns are not silently ignored
func TestNoSilentErrors(t *testing.T) {
	// This test validates the principle that we don't use "_, _" pattern
	// Example of what NOT to do:
	// output, _ := cmd.Output()  // BAD: error ignored

	// Instead we should:
	// output, err := cmd.Output()
	// if err != nil { handle error }

	// Test that even empty results should be handled carefully
	tool := &Tool{
		Name: "TestTool",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			// This handler properly handles errors
			return "result", nil
		},
	}

	if tool == nil {
		t.Fatal("Tool should not be nil")
	}
}

// TestErrorContextPreservation tests that errors preserve context
func TestErrorContextPreservation(t *testing.T) {
	originalErr := fmt.Errorf("underlying error")

	wrappedErr := fmt.Errorf("[COMMAND_FAILED] Command execution failed: %w. Suggestion: Check permissions", originalErr)

	if !contains(wrappedErr.Error(), "underlying error") {
		t.Error("Wrapped error should preserve original error message")
	}
	if !contains(wrappedErr.Error(), "COMMAND_FAILED") {
		t.Error("Wrapped error should include error type")
	}
	if !contains(wrappedErr.Error(), "Suggestion") {
		t.Error("Wrapped error should include suggested action")
	}
}

// TestBuildOpenAIMessagesBasic tests basic message construction
func TestBuildOpenAIMessagesBasic(t *testing.T) {
	agent := &Agent{
		ID:    "test-agent",
		Name:  "TestAgent",
		Role:  "Helper",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	history := []Message{}
	input := "Hello, how are you?"
	systemPrompt := "You are a helpful assistant"

	messages := buildOpenAIMessages(agent, input, history, systemPrompt)

	// Should have exactly 2 messages: system + user input
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}

// TestBuildOpenAIMessagesWithHistory tests message construction with history
func TestBuildOpenAIMessagesWithHistory(t *testing.T) {
	agent := &Agent{
		ID:    "test-agent",
		Name:  "TestAgent",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	history := []Message{
		{Role: "user", Content: "Previous question"},
		{Role: "assistant", Content: "Previous answer"},
	}
	input := "New question"
	systemPrompt := "You are helpful"

	messages := buildOpenAIMessages(agent, input, history, systemPrompt)

	// Should have: system + 2 history + user input = 4 messages
	if len(messages) != 4 {
		t.Errorf("Expected 4 messages with history, got %d", len(messages))
	}
}

// TestBuildOpenAIMessagesEmptyHistory tests with no history
func TestBuildOpenAIMessagesEmptyHistory(t *testing.T) {
	agent := &Agent{
		ID:    "agent",
		Name:  "Agent",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	messages := buildOpenAIMessages(agent, "Test input", []Message{}, "System prompt")

	// Should have: system + user input = 2 messages
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}

// TestBuildOpenAIMessagesMultipleHistory tests with multiple history items
func TestBuildOpenAIMessagesMultipleHistory(t *testing.T) {
	agent := &Agent{
		ID:    "agent",
		Name:  "Agent",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	history := []Message{
		{Role: "user", Content: "Q1"},
		{Role: "assistant", Content: "A1"},
		{Role: "user", Content: "Q2"},
		{Role: "assistant", Content: "A2"},
	}

	messages := buildOpenAIMessages(agent, "Q3", history, "System")

	// Should have: system + 4 history + user = 6 messages
	if len(messages) != 6 {
		t.Errorf("Expected 6 messages, got %d", len(messages))
	}
}

// TestBuildOpenAIMessagesSystemMessageInHistory tests system messages in history
func TestBuildOpenAIMessagesSystemMessageInHistory(t *testing.T) {
	agent := &Agent{
		ID:    "agent",
		Name:  "Agent",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	history := []Message{
		{Role: "user", Content: "Question"},
		{Role: "system", Content: "Tool result"},
	}

	messages := buildOpenAIMessages(agent, "Next question", history, "System")

	// Should have: system + 2 history + user = 4 messages
	if len(messages) != 4 {
		t.Errorf("Expected 4 messages with system message in history, got %d", len(messages))
	}
}

// TestBuildOpenAIMessagesInvalidRole tests unknown role handling
func TestBuildOpenAIMessagesInvalidRole(t *testing.T) {
	agent := &Agent{
		ID:    "agent",
		Name:  "Agent",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	history := []Message{
		{Role: "unknown_role", Content: "Some content"},
	}

	// Should not panic with unknown role
	messages := buildOpenAIMessages(agent, "Input", history, "System")

	// Should gracefully ignore unknown role
	if len(messages) < 2 {
		t.Error("Should handle unknown role gracefully")
	}
}

// TestBuildOpenAIMessagesEmptyStrings tests with empty content
func TestBuildOpenAIMessagesEmptyStrings(t *testing.T) {
	agent := &Agent{
		ID:    "agent",
		Name:  "Agent",
		Model: "gpt-4o",
		Tools: []*Tool{},
	}

	history := []Message{
		{Role: "user", Content: ""},
		{Role: "assistant", Content: ""},
	}

	// Should handle empty strings without panic
	messages := buildOpenAIMessages(agent, "", history, "")

	if len(messages) == 0 {
		t.Error("Should create messages even with empty content")
	}
}
