---
title: "Epic 1 Detailed Stories - Configuration Integrity & Trust"
epic: "Epic 1"
status: "Ready for Review"
date: "2025-12-20"
---

# Epic 1: Configuration Integrity & Trust - Detailed Stories

**Epic Goal:** Users can configure agents with confidence knowing every setting will be honored exactly as specified.

**Stories:** 3 stories
**Total Effort:** Small (4-7 hours)
**Priority:** CRITICAL - Foundation for all other epics
**Dependencies:** None

---

## Story 1.1: Agent Respects Configured Model

### Story Summary
As a developer, I want agents to use the model specified in their configuration, so that I can use different models for different agents without hardcoded overrides.

### Current Problem
- **Location:** [go-agentic/agent.go:24](go-agentic/agent.go#L24)
- **Issue:** Model is hardcoded to `"gpt-4o-mini"` regardless of `agent.Model` field
- **Impact:** Configuration is ignored, cannot use gpt-4o or other models per agent

### Acceptance Criteria

#### Given:
- An agent with `Model: "gpt-4o"` in its configuration
- Another agent with `Model: "gpt-4o-mini"` in its configuration

#### When:
- Story 1.1 is implemented (hardcoded value replaced with field)

#### Then:
- The first agent uses `"gpt-4o"` when calling the OpenAI API
- The second agent uses `"gpt-4o-mini"` when calling the OpenAI API
- IT Support example runs with correct models per agent
- Logs show which model each agent is using

### Implementation Details

#### File: `go-agentic/agent.go` (lines 23-26)

**BEFORE (Current - BROKEN):**
```go
// Create completion request
params := openai.ChatCompletionNewParams{
	Model:    "gpt-4o-mini",  // ❌ HARDCODED - ignores agent.Model
	Messages: messages,
}
```

**AFTER (Fixed - CORRECT):**
```go
// Create completion request
params := openai.ChatCompletionNewParams{
	Model:    agent.Model,     // ✅ Uses agent.Model from configuration
	Messages: messages,
}
```

#### Code Change Summary:
- Line 24: Replace string literal `"gpt-4o-mini"` with `agent.Model`
- No other changes needed
- This is a single-line fix

### Test Cases to Implement

#### Test 1.1.1: Model Assignment from Configuration
```go
func TestAgentUsesConfiguredModel_GPT4O(t *testing.T) {
	// Arrange
	agent := &Agent{
		ID:    "test-agent",
		Name:  "TestAgent",
		Model: "gpt-4o",  // Set to gpt-4o
	}

	// (Mock OpenAI client to capture params)

	// Act
	// Call ExecuteAgent with mocked client

	// Assert
	// Verify params.Model == "gpt-4o"
}
```

#### Test 1.1.2: Different Model per Agent
```go
func TestDifferentAgentsUseDifferentModels(t *testing.T) {
	// Arrange
	agent1 := &Agent{Model: "gpt-4o"}
	agent2 := &Agent{Model: "gpt-4o-mini"}

	// Act & Assert
	// Execute both agents
	// Verify each uses its configured model
	// agent1 params.Model == "gpt-4o"
	// agent2 params.Model == "gpt-4o-mini"
}
```

#### Test 1.1.3: IT Support Example with Mixed Models
```go
func TestITSupportExampleWithDifferentModels(t *testing.T) {
	// Load IT Support config with agents having different models
	// Execute team workflow
	// Verify all agents executed with correct models
}
```

#### Test 1.1.4: Log Verification
```go
func TestLogsShowCorrectModelUsed(t *testing.T) {
	// Run agent
	// Capture logs
	// Verify logs show "[INFO] Agent X using model Y"
}
```

### Acceptance Criteria Checklist
- [ ] Line 24 in agent.go uses `agent.Model` instead of `"gpt-4o-mini"`
- [ ] Test 1.1.1 passes: gpt-4o model is used when configured
- [ ] Test 1.1.2 passes: Different agents can use different models
- [ ] Test 1.1.3 passes: IT Support example works with different models
- [ ] Test 1.1.4 passes: Logs show correct model for each agent
- [ ] No other code changes needed
- [ ] Existing examples work unchanged (backward compatible)
- [ ] All tests pass: `make test`
- [ ] No linting errors: `make lint`

### Risk Assessment
**Risk Level:** ✅ LOW
- Single line change, isolated to one function
- No API contract changes
- No configuration format changes
- Backward compatible (uses agent.Model field already exists)
- Easy to verify (logs will show model used)

### Time Estimate
**1-2 hours** (including tests)

### Related Stories
- Depends on: None
- Blocks: None
- Story 1.3 will add validation to ensure Model is not empty

---

## Story 1.2: Temperature Configuration Respects All Valid Values

### Story Summary
As a developer, I want to set temperature to any valid OpenAI value (0.0-2.0), so that I can control agent creativity/determinism per use case.

### Current Problem
- **Location:** [go-agentic/config.go:58-60](go-agentic/config.go#L58-L60)
- **Issue:** Temperature=0.0 is overridden to 0.7 by this code:
  ```go
  if config.Temperature == 0 {
      config.Temperature = 0.7
  }
  ```
- **Impact:** Cannot use temperature=0.0 for deterministic responses (exact results every time)

### Root Cause Analysis
The code treats `0` as "not set" and applies default, but 0.0 is a valid value in OpenAI API (0.0-2.0 range).

### Acceptance Criteria

#### Given:
- Agent configuration with `Temperature: 0.0`
- Agent configuration with `Temperature: 1.0`
- Agent configuration with `Temperature: 2.0`

#### When:
- Configuration is loaded in LoadAgentConfig

#### Then:
- Temperature=0.0 remains 0.0 (not overridden to 0.7)
- Temperature=1.0 remains 1.0
- Temperature=2.0 remains 2.0
- Unspecified temperature (missing from config) defaults to 0.7
- Invalid temperatures (negative or >2.0) are rejected during story 1.3 validation

### Implementation Details

#### File: `go-agentic/config.go` (lines 58-60)

**BEFORE (Current - BROKEN):**
```go
// Set defaults
if config.Model == "" {
	config.Model = "gpt-4o-mini"
}
if config.Temperature == 0 {  // ❌ WRONG: treats 0 as "not set"
	config.Temperature = 0.7
}
```

**AFTER (Fixed - CORRECT):**

We need to distinguish between "not set" (should default) and "explicitly set to 0" (should keep).

There are two valid approaches:

**Option A: Use pointer type** (Best for backward compatibility)
```go
// In types.go, change AgentConfig.Temperature to pointer:
type AgentConfig struct {
	// ... other fields ...
	Temperature *float64 `yaml:"temperature"`
	// ...
}

// In config.go, handle the pointer:
if config.Temperature == nil {
	defaultTemp := 0.7
	config.Temperature = &defaultTemp
}
// Temperature is now used as *config.Temperature in agent initialization
```

**Option B: Use separate field** (Simpler, maintains backward compatibility)
```go
type AgentConfig struct {
	Temperature    float64 `yaml:"temperature"`
	TemperatureSet bool    // true if explicitly set in YAML
}

// In config.go:
// Parse YAML manually to check if temperature was present in file
// If present, keep value (even if 0.0), otherwise set default
```

**Option C: Validate non-zero only if explicitly checking** (Current approach fix)
```go
// Remove the default override, let it be 0 by default
// During story 1.3 validation, only apply default if value is not in valid range
```

**Recommendation:** Use Option A (pointer type) because:
- Clean Go idiom for "optional with default" pattern
- Explicit nil check is clearer than magic number
- Distinguishes "not provided" from "provided as 0"
- Used throughout Go standard library

#### Code Change:

**types.go - Change AgentConfig struct:**
```go
type AgentConfig struct {
	ID           string     `yaml:"id"`
	Name         string     `yaml:"name"`
	Role         string     `yaml:"role"`
	Backstory    string     `yaml:"backstory"`
	Model        string     `yaml:"model"`
	Temperature  *float64   `yaml:"temperature"` // ✅ Change from float64 to *float64
	SystemPrompt string     `yaml:"system_prompt"`
	IsTerminal   bool       `yaml:"is_terminal"`
	Tools        []string   `yaml:"tools"`
	HandoffTo    []string   `yaml:"handoff_to"`
}
```

**config.go - Update LoadAgentConfig:**
```go
// Set defaults
if config.Model == "" {
	config.Model = "gpt-4o-mini"
}
if config.Temperature == nil {  // ✅ Check for nil, not 0
	defaultTemp := 0.7
	config.Temperature = &defaultTemp
}
```

**config.go - Update CreateAgentFromConfig:**
```go
func CreateAgentFromConfig(config *AgentConfig, allTools map[string]*Tool) *Agent {
	agent := &Agent{
		ID:             config.ID,
		Name:           config.Name,
		Role:           config.Role,
		Backstory:      config.Backstory,
		Model:          config.Model,
		SystemPrompt:   config.SystemPrompt,
		Temperature:    *config.Temperature,  // ✅ Dereference pointer
		IsTerminal:     config.IsTerminal,
		HandoffTargets: config.HandoffTo,
		Tools:          []*Tool{},
	}
	// ... rest of function
}
```

### Test Cases to Implement

#### Test 1.2.1: Temperature 0.0 is Respected
```go
func TestTemperatureZeroIsRespected(t *testing.T) {
	configYAML := `
temperature: 0.0
model: gpt-4o
`
	// Parse config
	// Assert: config.Temperature == 0.0
}
```

#### Test 1.2.2: Temperature 1.0 is Respected
```go
func TestTemperature1Point0IsRespected(t *testing.T) {
	// Load config with temperature: 1.0
	// Assert: config.Temperature == 1.0
}
```

#### Test 1.2.3: Temperature 2.0 is Respected
```go
func TestTemperature2Point0IsRespected(t *testing.T) {
	// Load config with temperature: 2.0
	// Assert: config.Temperature == 2.0
}
```

#### Test 1.2.4: Missing Temperature Defaults to 0.7
```go
func TestMissingTemperatureDefaultsTo0Point7(t *testing.T) {
	configYAML := `
model: gpt-4o
# No temperature specified
`
	// Parse config
	// Assert: config.Temperature == 0.7
}
```

#### Test 1.2.5: Agent Uses Temperature from Config
```go
func TestAgentUsesTemperatureFromConfig(t *testing.T) {
	// Create config with temperature: 0.0
	agent := CreateAgentFromConfig(&config, tools)

	// Mock OpenAI call
	// Execute agent

	// Assert: API params.Temperature == 0.0
}
```

### Acceptance Criteria Checklist
- [ ] AgentConfig.Temperature changed to `*float64` in types.go
- [ ] LoadAgentConfig checks for nil, not 0
- [ ] CreateAgentFromConfig dereferences pointer correctly
- [ ] Test 1.2.1 passes: Temperature 0.0 is kept
- [ ] Test 1.2.2 passes: Temperature 1.0 is kept
- [ ] Test 1.2.3 passes: Temperature 2.0 is kept
- [ ] Test 1.2.4 passes: Missing temperature defaults to 0.7
- [ ] Test 1.2.5 passes: Agent API call uses correct temperature
- [ ] IT Support example still works with default/custom temperatures
- [ ] All tests pass: `make test`
- [ ] No linting errors: `make lint`

### Risk Assessment
**Risk Level:** ✅ LOW-MEDIUM
- Changes type of existing field (might affect external code)
- Requires updates in CreateAgentFromConfig and usage sites
- Backward compatible: external configs continue to work, just parsed differently
- All changes localized to config.go and types.go

### Time Estimate
**1-2 hours** (including tests and all reference updates)

### Related Stories
- Depends on: None
- Blocks: None
- Story 1.3 will add validation to ensure Temperature is in 0.0-2.0 range

---

## Story 1.3: Configuration Validation & Error Messages

### Story Summary
As a developer, I want clear validation errors when configuration is invalid, so that I catch configuration mistakes early with helpful messages.

### Current Problem
- **Issue:** No validation on agent configuration after loading
- **Impact:** Invalid configurations silently fail or produce cryptic API errors
- **Examples:**
  - Missing Model field might default to deprecated model
  - Invalid temperature (2.1, -5) accepted, causes API error
  - No logging of successful agent initialization

### Acceptance Criteria

#### Given:
- Invalid agent configuration (empty Model, invalid Temperature)
- Valid agent configuration

#### When:
- Configuration is loaded

#### Then:
- Invalid configurations return clear error messages
- Valid configurations load successfully and log confirmation
- Error messages explain what's wrong and how to fix it

### Implementation Details

#### File: `go-agentic/config.go` - Add validation function

**Add after LoadAgentConfig function:**
```go
// ValidateAgentConfig validates an agent configuration
func ValidateAgentConfig(config *AgentConfig) error {
	// Validate Model
	if config.Model == "" {
		return fmt.Errorf("agent config validation failed: Model must be specified (examples: gpt-4o, gpt-4o-mini)")
	}

	// Validate Temperature
	if config.Temperature != nil {
		temp := *config.Temperature
		if temp < 0.0 || temp > 2.0 {
			return fmt.Errorf("agent config validation failed: Temperature must be between 0.0 and 2.0, got %.1f", temp)
		}
	}

	return nil
}
```

#### File: `go-agentic/config.go` - Update LoadAgentConfig

**After unmarshaling, add validation:**
```go
func LoadAgentConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent config: %w", err)
	}

	var config AgentConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent config: %w", err)
	}

	// Set defaults
	if config.Model == "" {
		config.Model = "gpt-4o-mini"
	}
	if config.Temperature == nil {
		defaultTemp := 0.7
		config.Temperature = &defaultTemp
	}

	// ✅ NEW: Validate configuration
	if err := ValidateAgentConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
```

#### File: `go-agentic/agent.go` - Add logging on successful execution

**In ExecuteAgent function, add logging:**
```go
func ExecuteAgent(ctx context.Context, agent *Agent, input string, history []Message, apiKey string) (*AgentResponse, error) {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	// ✅ NEW: Log agent initialization
	fmt.Printf("[INFO] Agent '%s' (ID: %s) using model '%s' with temperature %.1f\n",
		agent.Name, agent.ID, agent.Model, agent.Temperature)

	// Build system prompt
	systemPrompt := buildSystemPrompt(agent)
	// ... rest of function
}
```

### Type Definitions

#### File: `go-agentic/types.go` - Add error types

**Add after existing type definitions:**
```go
// ConfigValidationError represents a configuration validation error
type ConfigValidationError struct {
	Field      string // Name of the invalid field
	Value      interface{}
	Reason     string // Why it's invalid
	Suggestion string // How to fix it
}

// Error implements the error interface
func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("config validation failed for '%s': %s (got %v). %s",
		e.Field, e.Reason, e.Value, e.Suggestion)
}
```

*Note: For Epic 1.3, we'll use simple fmt.Errorf. ConfigValidationError will be used in Epic 3.*

### Test Cases to Implement

#### Test 1.3.1: Empty Model Field Error
```go
func TestEmptyModelFieldReturnsError(t *testing.T) {
	config := &AgentConfig{
		ID:          "test",
		Name:        "TestAgent",
		Model:       "",  // ❌ Empty
		Temperature: &temp7,
	}

	err := ValidateAgentConfig(config)

	if err == nil {
		t.Fatal("expected error for empty Model, got nil")
	}
	if !strings.Contains(err.Error(), "Model must be specified") {
		t.Errorf("expected helpful error message, got: %v", err)
	}
}
```

#### Test 1.3.2: Invalid Temperature > 2.0 Error
```go
func TestInvalidTemperature2Point1ReturnsError(t *testing.T) {
	temp := 2.1
	config := &AgentConfig{
		ID:          "test",
		Name:        "TestAgent",
		Model:       "gpt-4o",
		Temperature: &temp,  // ❌ > 2.0
	}

	err := ValidateAgentConfig(config)

	if err == nil {
		t.Fatal("expected error for temperature 2.1")
	}
	if !strings.Contains(err.Error(), "between 0.0 and 2.0") {
		t.Errorf("expected range error message, got: %v", err)
	}
}
```

#### Test 1.3.3: Invalid Temperature < 0.0 Error
```go
func TestInvalidTemperatureNegativeReturnsError(t *testing.T) {
	temp := -1.0
	config := &AgentConfig{
		ID:          "test",
		Name:        "TestAgent",
		Model:       "gpt-4o",
		Temperature: &temp,  // ❌ < 0.0
	}

	err := ValidateAgentConfig(config)

	if err == nil {
		t.Fatal("expected error for negative temperature")
	}
}
```

#### Test 1.3.4: Valid Configuration Passes Validation
```go
func TestValidConfigurationPasses(t *testing.T) {
	temp := 0.7
	config := &AgentConfig{
		ID:          "test",
		Name:        "TestAgent",
		Model:       "gpt-4o",
		Temperature: &temp,  // ✅ Valid
	}

	err := ValidateAgentConfig(config)

	if err != nil {
		t.Errorf("expected nil error for valid config, got: %v", err)
	}
}
```

#### Test 1.3.5: Boundary Temperatures Pass
```go
func TestBoundaryTemperaturesPass(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
	}{
		{"zero", 0.0},
		{"low", 0.5},
		{"middle", 1.0},
		{"high", 1.5},
		{"max", 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temp := tt.temperature
			config := &AgentConfig{
				ID:          "test",
				Model:       "gpt-4o",
				Temperature: &temp,
			}

			err := ValidateAgentConfig(config)
			if err != nil {
				t.Errorf("expected valid temperature %.1f, got error: %v", tt.temperature, err)
			}
		})
	}
}
```

#### Test 1.3.6: LoadAgentConfig Validates
```go
func TestLoadAgentConfigValidatesConfig(t *testing.T) {
	// Create temp YAML file with invalid config
	invalidYAML := `
id: test
name: TestAgent
temperature: 2.5  # Invalid - too high
`
	tempFile := createTempFile(t, invalidYAML)
	defer os.Remove(tempFile)

	_, err := LoadAgentConfig(tempFile)

	if err == nil {
		t.Fatal("expected LoadAgentConfig to return error for invalid config")
	}
}
```

#### Test 1.3.7: Valid Config File Loads Successfully
```go
func TestValidConfigFileLoa

dsSuccessfully(t *testing.T) {
	validYAML := `
id: test-agent
name: TestAgent
role: Helper
backstory: A helpful agent
model: gpt-4o
temperature: 0.7
`
	tempFile := createTempFile(t, validYAML)
	defer os.Remove(tempFile)

	config, err := LoadAgentConfig(tempFile)

	if err != nil {
		t.Errorf("expected LoadAgentConfig to succeed, got error: %v", err)
	}
	if config == nil {
		t.Fatal("expected config to be loaded")
	}
}
```

#### Test 1.3.8: Backward Compatibility with v0.0.1 Configs
```go
func TestBackwardCompatibilityWithOldConfigs(t *testing.T) {
	// Config format from v0.0.1
	oldYAML := `
id: old-agent
name: OldAgent
role: Helper
model: gpt-4o-mini
`
	tempFile := createTempFile(t, oldYAML)
	defer os.Remove(tempFile)

	config, err := LoadAgentConfig(tempFile)

	if err != nil {
		t.Errorf("expected backward compatibility, got error: %v", err)
	}
	// Temperature should default to 0.7
	if config.Temperature == nil || *config.Temperature != 0.7 {
		t.Errorf("expected default temperature 0.7")
	}
}
```

### Acceptance Criteria Checklist
- [ ] ValidateAgentConfig function added to config.go
- [ ] LoadAgentConfig calls ValidateAgentConfig after loading
- [ ] LoadAgentConfig returns clear error for empty Model
- [ ] LoadAgentConfig returns clear error for invalid Temperature
- [ ] Validation error messages include: field name, expected values, suggestion
- [ ] ExecuteAgent logs agent initialization (model, temperature)
- [ ] Test 1.3.1 passes: Empty Model error
- [ ] Test 1.3.2 passes: Invalid Temperature > 2.0 error
- [ ] Test 1.3.3 passes: Invalid Temperature < 0.0 error
- [ ] Test 1.3.4 passes: Valid config passes validation
- [ ] Test 1.3.5 passes: Boundary values 0.0, 2.0 valid
- [ ] Test 1.3.6 passes: LoadAgentConfig validates
- [ ] Test 1.3.7 passes: Valid config loads successfully
- [ ] Test 1.3.8 passes: Backward compatible with v0.0.1
- [ ] All tests pass: `make test`
- [ ] No linting errors: `make lint`

### Risk Assessment
**Risk Level:** ✅ LOW
- Validation is additive (doesn't change working code)
- Error messages help debugging
- Backward compatible with existing configs
- Validation only rejects genuinely invalid configs

### Time Estimate
**2-3 hours** (including tests)

### Related Stories
- Depends on: Story 1.1, Story 1.2
- Blocks: None
- Epic 3 will build on error handling foundation with ToolError type

---

## Epic 1 Implementation Summary

### Total Effort Breakdown
| Story | Effort | Cumulative |
|-------|--------|-----------|
| 1.1   | 1-2h   | 1-2h      |
| 1.2   | 1-2h   | 2-4h      |
| 1.3   | 2-3h   | 4-7h      |

### Total Estimated Time: **4-7 hours**

### Key Files to Modify
1. **go-agentic/agent.go** - Line 24 (Story 1.1) + logging (Story 1.3)
2. **go-agentic/config.go** - Temperature handling (Story 1.2) + validation (Story 1.3)
3. **go-agentic/types.go** - AgentConfig.Temperature type change (Story 1.2)

### Test Files to Create/Modify
- Create: `go-agentic/*_test.go` for Story 1.1, 1.2, 1.3 tests
- Existing test structure to follow: Use table-driven tests for parametric testing

### Quality Gates

#### Before Implementation:
- [ ] All 3 stories thoroughly reviewed
- [ ] Team agrees on approach (especially pointer type for Story 1.2)
- [ ] Test cases are clear and comprehensive

#### During Implementation:
- [ ] Each story tested independently
- [ ] Run `make test` frequently
- [ ] No hardcoded values or ignored errors

#### Before PR:
- [ ] All tests pass: `make test`
- [ ] Coverage checked: `make coverage`
- [ ] Linting passes: `make lint`
- [ ] Code review checklist satisfied
- [ ] IT Support example runs without errors

### Success Criteria for Epic 1
✅ **Story 1.1 Complete:**
- Agent.Model is respected (no hardcoding)
- API calls use correct model per agent
- Logs show which model each agent uses
- Backward compatible (existing examples work)

✅ **Story 1.2 Complete:**
- Temperature 0.0 can be set and used
- All valid temperatures (0.0-2.0) work
- Unspecified temperature defaults to 0.7
- Backward compatible (existing configs work)

✅ **Story 1.3 Complete:**
- Invalid configurations rejected with clear errors
- Error messages explain what's wrong and how to fix
- Agent initialization logged successfully
- Backward compatible with v0.0.1 configs

✅ **Epic 1 Complete:**
- All 3 stories passing on Windows, macOS, Linux
- Code coverage >90% for modified code
- All tests passing: `make test`
- Linting passes: `make lint`
- Code review approved
- PR merged to main

---

## Next Steps After Epic 1

Once all 3 stories are complete and merged:

1. **Proceed to Epic 5:** Testing Framework (parallel work)
   - Set up test APIs and infrastructure
   - This supports testing for all subsequent epics

2. **Proceed to Epic 2a:** Native Tool Call Parsing
   - Use OpenAI's native tool_calls API
   - Add text parsing fallback

3. **Proceed to Epic 2b:** Parameter Validation
   - Validate tool parameters before handler execution
   - Clear error messages for mismatches

4. **Proceed to Epic 3 & 4 (parallel):**
   - Epic 3: Clear error handling with ToolError type
   - Epic 4: Cross-platform compatibility

5. **Conclude with Epic 7:** End-to-End Validation

---

## Questions & Clarifications Needed

Before starting implementation, please confirm:

1. **Story 1.2 - Pointer Type Approach:**
   - Do you agree with using `*float64` for Temperature in AgentConfig?
   - Alternative: use separate "TemperatureSet bool" field?

2. **Error Handling:**
   - Should validation errors be fatal (stop loading) or warnings (log but continue)?
   - Currently planned: fatal (return error from LoadAgentConfig)

3. **Logging:**
   - Should agent initialization logging go to stdout or structured logger?
   - Currently planned: fmt.Printf to stdout
   - Could be enhanced in future to use a logging library

4. **Test Coverage:**
   - Are the test cases comprehensive enough?
   - Should we add additional edge cases?

5. **Backward Compatibility:**
   - Do we need to support v0.0.0 configs, or just v0.0.1?
   - Currently planned: support v0.0.1 and forward

---

**Status:** Ready for team review and implementation start

