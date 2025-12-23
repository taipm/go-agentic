package crewai

import (
	"testing"
)

// ===== Issue #6: YAML Validation Tests =====

// TestValidateCrewConfigValidConfig verifies valid crew config passes validation
func TestValidateCrewConfigValidConfig(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator", "executor"},
	}
	config.Settings.MaxHandoffs = 5
	config.Settings.MaxRounds = 10
	config.Settings.TimeoutSeconds = 300
	config.Settings.Language = "en"

	err := ValidateCrewConfig(config)
	if err != nil {
		t.Errorf("Valid config should pass validation, got error: %v", err)
	}
}

// TestValidateCrewConfigMissingVersion validates version field is required
func TestValidateCrewConfigMissingVersion(t *testing.T) {
	config := &CrewConfig{
		Version:    "", // ← Missing!
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator"},
	}

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should require 'version' field")
	}
	if err.Error() != "required field 'version' is empty" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigMissingAgents validates agents field is required
func TestValidateCrewConfigMissingAgents(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "orchestrator",
		Agents:     []string{}, // ← Empty!
	}

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should require non-empty 'agents' list")
	}
	if err.Error() != "required field 'agents' is empty - at least one agent must be configured" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigMissingEntryPoint validates entry_point field is required
func TestValidateCrewConfigMissingEntryPoint(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "", // ← Missing!
		Agents:     []string{"orchestrator"},
	}

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should require 'entry_point' field")
	}
	if err.Error() != "required field 'entry_point' is empty" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigEntryPointNotInAgents validates entry_point must exist in agents
func TestValidateCrewConfigEntryPointNotInAgents(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "non_existent_agent", // ← Not in agents list!
		Agents:     []string{"orchestrator", "executor"},
	}

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should validate entry_point exists in agents list")
	}
	if err.Error() != "entry_point 'non_existent_agent' not found in agents list" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigNegativeMaxHandoffs validates max_handoffs >= 0
func TestValidateCrewConfigNegativeMaxHandoffs(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator"},
	}
	config.Settings.MaxHandoffs = -5 // ← Negative!
	config.Settings.MaxRounds = 10

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should validate max_handoffs >= 0")
	}
	if err.Error() != "settings.max_handoffs must be >= 0, got -5" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigInvalidMaxRounds validates max_rounds > 0
func TestValidateCrewConfigInvalidMaxRounds(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator"},
	}
	config.Settings.MaxHandoffs = 5
	config.Settings.MaxRounds = 0 // ← Invalid!

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should validate max_rounds > 0")
	}
	if err.Error() != "settings.max_rounds must be > 0, got 0" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigInvalidTimeout validates timeout_seconds > 0
func TestValidateCrewConfigInvalidTimeout(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator"},
	}
	config.Settings.MaxHandoffs = 5
	config.Settings.MaxRounds = 10
	config.Settings.TimeoutSeconds = -100 // ← Negative!

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should validate timeout_seconds > 0")
	}
	if err.Error() != "settings.timeout_seconds must be > 0, got -100" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigRoutingSignalInvalidAgent validates routing signals reference existing agents
func TestValidateCrewConfigRoutingSignalInvalidAgent(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator", "executor"},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"non_existent_agent": { // ← Not in agents list!
					{Signal: "[ROUTE]", Target: "executor"},
				},
			},
		},
	}
	config.Settings.MaxRounds = 10
	config.Settings.TimeoutSeconds = 300

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should validate routing signals reference existing agents")
	}
	if err.Error() != "routing.signals references non-existent agent 'non_existent_agent'" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigRoutingSignalTargetInvalid validates signal targets exist
func TestValidateCrewConfigRoutingSignalTargetInvalid(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator", "executor"},
		Routing: &RoutingConfig{
			Signals: map[string][]RoutingSignal{
				"orchestrator": {
					{Signal: "[ROUTE]", Target: "non_existent_target"}, // ← Invalid target!
				},
			},
		},
	}
	config.Settings.MaxRounds = 10
	config.Settings.TimeoutSeconds = 300

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should validate routing signal targets exist")
	}
	// Updated: error message now mentions "agent/group" since parallel groups are supported
	if err.Error() != "routing signal from agent 'orchestrator' targets non-existent agent/group 'non_existent_target'" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigBehaviorInvalidAgent validates agent_behaviors reference existing agents
func TestValidateCrewConfigBehaviorInvalidAgent(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator"},
		Routing: &RoutingConfig{
			AgentBehaviors: map[string]AgentBehavior{
				"non_existent_agent": { // ← Invalid agent!
					WaitForSignal: true,
				},
			},
		},
	}
	config.Settings.MaxRounds = 10
	config.Settings.TimeoutSeconds = 300

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should validate agent_behaviors reference existing agents")
	}
	if err.Error() != "routing.agent_behaviors references non-existent agent 'non_existent_agent'" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigParallelGroupInvalidAgent validates parallel groups reference existing agents
func TestValidateCrewConfigParallelGroupInvalidAgent(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator", "executor"},
		Routing: &RoutingConfig{
			ParallelGroups: map[string]ParallelGroupConfig{
				"group1": {
					Agents:         []string{"non_existent_agent"}, // ← Invalid!
					TimeoutSeconds: 60,
				},
			},
		},
	}
	config.Settings.MaxRounds = 10
	config.Settings.TimeoutSeconds = 300

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should validate parallel groups reference existing agents")
	}
	if err.Error() != "parallel_group 'group1' references non-existent agent 'non_existent_agent'" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateCrewConfigParallelGroupNoAgents validates parallel groups have agents
func TestValidateCrewConfigParallelGroupNoAgents(t *testing.T) {
	config := &CrewConfig{
		Version:    "1.0",
		EntryPoint: "orchestrator",
		Agents:     []string{"orchestrator"},
		Routing: &RoutingConfig{
			ParallelGroups: map[string]ParallelGroupConfig{
				"group1": {
					Agents:         []string{}, // ← Empty!
					TimeoutSeconds: 60,
				},
			},
		},
	}
	config.Settings.MaxRounds = 10
	config.Settings.TimeoutSeconds = 300

	err := ValidateCrewConfig(config)
	if err == nil {
		t.Error("Should validate parallel groups have agents")
	}
	if err.Error() != "parallel_group 'group1' has no agents" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateAgentConfigValidConfig verifies valid agent config passes validation
func TestValidateAgentConfigValidConfig(t *testing.T) {
	config := &AgentConfig{
		ID:          "agent1",
		Name:        "Test Agent",
		Role:        "test_role",
		Temperature: 0.7,
		Primary: &ModelConfigYAML{
			Model:    "gpt-4o",
			Provider: "openai",
		},
	}

	err := ValidateAgentConfig(config, PermissiveMode)
	if err != nil {
		t.Errorf("Valid agent config should pass validation, got error: %v", err)
	}
}

// TestValidateAgentConfigMissingID validates ID is required in STRICT mode
func TestValidateAgentConfigMissingID(t *testing.T) {
	config := &AgentConfig{
		ID:   "", // ← Missing!
		Name: "Test Agent",
		Role: "test_role",
		Primary: &ModelConfigYAML{
			Model:    "gpt-4o",
			Provider: "openai",
		},
	}

	err := ValidateAgentConfig(config, StrictMode)
	if err == nil {
		t.Error("Should require 'id' field in STRICT mode")
	}
	if !contains(err.Error(), "missing required fields in STRICT MODE") {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateAgentConfigMissingName validates Name is required in STRICT mode
func TestValidateAgentConfigMissingName(t *testing.T) {
	config := &AgentConfig{
		ID:   "agent1",
		Name: "", // ← Missing!
		Role: "test_role",
		Primary: &ModelConfigYAML{
			Model:    "gpt-4o",
			Provider: "openai",
		},
	}

	err := ValidateAgentConfig(config, StrictMode)
	if err == nil {
		t.Error("Should require 'name' field in STRICT mode")
	}
	if !contains(err.Error(), "missing required fields in STRICT MODE") {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateAgentConfigMissingRole validates Role is required in STRICT mode
func TestValidateAgentConfigMissingRole(t *testing.T) {
	config := &AgentConfig{
		ID:   "agent1",
		Name: "Test Agent",
		Role: "", // ← Missing!
		Primary: &ModelConfigYAML{
			Model:    "gpt-4o",
			Provider: "openai",
		},
	}

	err := ValidateAgentConfig(config, StrictMode)
	if err == nil {
		t.Error("Should require 'role' field in STRICT mode")
	}
	if !contains(err.Error(), "missing required fields in STRICT MODE") {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateAgentConfigInvalidTemperature validates temperature range [0, 2]
func TestValidateAgentConfigInvalidTemperature(t *testing.T) {
	testCases := []struct {
		temp float64
		desc string
	}{
		{-0.5, "negative"},
		{2.5, "too high"},
		{3.0, "way too high"},
	}

	for _, tc := range testCases {
		config := &AgentConfig{
			ID:          "agent1",
			Name:        "Test Agent",
			Role:        "test_role",
			Temperature: tc.temp,
		}

		err := ValidateAgentConfig(config, PermissiveMode)
		if err == nil {
			t.Errorf("Should validate temperature range for %s value %.1f", tc.desc, tc.temp)
		}
		if err.Error() != "agent 'agent1': temperature must be between 0 and 2, got "+string(rune(int(tc.temp))) {
			// Check that error message mentions the temperature constraint
			if !contains(err.Error(), "temperature must be between 0 and 2") {
				t.Errorf("Error should mention temperature range, got: %v", err)
			}
		}
	}
}

// TestValidateAgentConfigTemperatureBoundaries validates temperature boundaries
func TestValidateAgentConfigTemperatureBoundaries(t *testing.T) {
	testCases := []struct {
		temp  float64
		valid bool
	}{
		{0.0, true},   // Min valid
		{1.0, true},   // Middle
		{2.0, true},   // Max valid
		{-0.1, false}, // Just below min
		{2.1, false},  // Just above max
	}

	for _, tc := range testCases {
		config := &AgentConfig{
			ID:          "agent1",
			Name:        "Test Agent",
			Role:        "test_role",
			Temperature: tc.temp,
			Primary: &ModelConfigYAML{
				Model:    "gpt-4o",
				Provider: "openai",
			},
		}

		err := ValidateAgentConfig(config, PermissiveMode)
		if tc.valid && err != nil {
			t.Errorf("Temperature %.1f should be valid but got error: %v", tc.temp, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("Temperature %.1f should be invalid", tc.temp)
		}
	}
}

// Helper function for string containment check
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ===== Backup LLM Model Tests (NEW) =====

// TestValidateAgentConfigWithPrimaryModel validates agent with primary model
func TestValidateAgentConfigWithPrimaryModel(t *testing.T) {
	config := &AgentConfig{
		ID:   "agent1",
		Name: "Test Agent",
		Role: "test_role",
		Primary: &ModelConfigYAML{
			Model:       "gpt-4o",
			Provider:    "openai",
			ProviderURL: "https://api.openai.com",
		},
		Temperature: 0.7,
	}

	err := ValidateAgentConfig(config, PermissiveMode)
	if err != nil {
		t.Errorf("Valid agent with primary model should pass validation, got error: %v", err)
	}
}

// TestValidateAgentConfigWithPrimaryAndBackup validates agent with primary and backup
func TestValidateAgentConfigWithPrimaryAndBackup(t *testing.T) {
	config := &AgentConfig{
		ID:   "agent1",
		Name: "Test Agent",
		Role: "test_role",
		Primary: &ModelConfigYAML{
			Model:       "gpt-4o",
			Provider:    "openai",
			ProviderURL: "https://api.openai.com",
		},
		Backup: &ModelConfigYAML{
			Model:       "deepseek-r1:32b",
			Provider:    "ollama",
			ProviderURL: "http://localhost:11434",
		},
		Temperature: 0.7,
	}

	err := ValidateAgentConfig(config, PermissiveMode)
	if err != nil {
		t.Errorf("Valid agent with primary and backup should pass validation, got error: %v", err)
	}
}

// TestValidateAgentConfigMissingPrimaryModel validates primary model is required
func TestValidateAgentConfigMissingPrimaryModel(t *testing.T) {
	config := &AgentConfig{
		ID:   "agent1",
		Name: "Test Agent",
		Role: "test_role",
		// Primary is nil!
		Temperature: 0.7,
	}

	err := ValidateAgentConfig(config, PermissiveMode)
	if err == nil {
		t.Error("Should require primary model configuration")
	}
	if err.Error() != "agent 'agent1': primary model configuration is missing" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateAgentConfigEmptyPrimaryModel validates primary.model is required in STRICT mode
func TestValidateAgentConfigEmptyPrimaryModel(t *testing.T) {
	config := &AgentConfig{
		ID:   "agent1",
		Name: "Test Agent",
		Role: "test_role",
		Primary: &ModelConfigYAML{
			Model:       "", // ← Empty!
			Provider:    "openai",
			ProviderURL: "https://api.openai.com",
		},
		Temperature: 0.7,
	}

	err := ValidateAgentConfig(config, StrictMode)
	if err == nil {
		t.Error("Should require primary.model in STRICT mode")
	}
	if !contains(err.Error(), "primary model configuration incomplete in STRICT MODE") {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateAgentConfigEmptyPrimaryProvider validates primary.provider is required in STRICT mode
func TestValidateAgentConfigEmptyPrimaryProvider(t *testing.T) {
	config := &AgentConfig{
		ID:   "agent1",
		Name: "Test Agent",
		Role: "test_role",
		Primary: &ModelConfigYAML{
			Model:       "gpt-4o",
			Provider:    "", // ← Empty!
			ProviderURL: "https://api.openai.com",
		},
		Temperature: 0.7,
	}

	err := ValidateAgentConfig(config, StrictMode)
	if err == nil {
		t.Error("Should require primary.provider in STRICT mode")
	}
	if !contains(err.Error(), "primary model configuration incomplete in STRICT MODE") {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateAgentConfigEmptyBackupModel validates backup.model is required if backup is specified
func TestValidateAgentConfigEmptyBackupModel(t *testing.T) {
	config := &AgentConfig{
		ID:   "agent1",
		Name: "Test Agent",
		Role: "test_role",
		Primary: &ModelConfigYAML{
			Model:       "gpt-4o",
			Provider:    "openai",
			ProviderURL: "https://api.openai.com",
		},
		Backup: &ModelConfigYAML{
			Model:       "", // ← Empty!
			Provider:    "ollama",
			ProviderURL: "http://localhost:11434",
		},
		Temperature: 0.7,
	}

	err := ValidateAgentConfig(config, PermissiveMode)
	if err == nil {
		t.Error("Should require backup.model if backup is specified")
	}
	if err.Error() != "agent 'agent1': backup.model must not be empty if backup is specified" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestValidateAgentConfigEmptyBackupProvider validates backup.provider is required if backup is specified
func TestValidateAgentConfigEmptyBackupProvider(t *testing.T) {
	config := &AgentConfig{
		ID:   "agent1",
		Name: "Test Agent",
		Role: "test_role",
		Primary: &ModelConfigYAML{
			Model:       "gpt-4o",
			Provider:    "openai",
			ProviderURL: "https://api.openai.com",
		},
		Backup: &ModelConfigYAML{
			Model:       "deepseek-r1:32b",
			Provider:    "", // ← Empty!
			ProviderURL: "http://localhost:11434",
		},
		Temperature: 0.7,
	}

	err := ValidateAgentConfig(config, PermissiveMode)
	if err == nil {
		t.Error("Should require backup.provider if backup is specified")
	}
	if err.Error() != "agent 'agent1': backup.provider must not be empty if backup is specified" {
		t.Errorf("Wrong error message: %v", err)
	}
}
