package agentic

import (
	"os"
	"testing"
)

// Test 1.2.1: Temperature 0.0 is respected (not overridden)
func TestTemperatureZeroIsRespected(t *testing.T) {
	temp := 0.0
	config := &AgentConfig{
		ID:          "test",
		Name:        "TestAgent",
		Model:       "gpt-4o",
		Temperature: &temp,
	}

	if config.Temperature == nil {
		t.Fatal("expected Temperature to be set")
	}

	if *config.Temperature != 0.0 {
		t.Errorf("expected Temperature 0.0, got %.1f", *config.Temperature)
	}
}

// Test 1.2.2: Temperature 1.0 is respected
func TestTemperature1Point0IsRespected(t *testing.T) {
	temp := 1.0
	config := &AgentConfig{
		ID:          "test",
		Name:        "TestAgent",
		Model:       "gpt-4o",
		Temperature: &temp,
	}

	if config.Temperature == nil {
		t.Fatal("expected Temperature to be set")
	}

	if *config.Temperature != 1.0 {
		t.Errorf("expected Temperature 1.0, got %.1f", *config.Temperature)
	}
}

// Test 1.2.3: Temperature 2.0 is respected
func TestTemperature2Point0IsRespected(t *testing.T) {
	temp := 2.0
	config := &AgentConfig{
		ID:          "test",
		Name:        "TestAgent",
		Model:       "gpt-4o",
		Temperature: &temp,
	}

	if config.Temperature == nil {
		t.Fatal("expected Temperature to be set")
	}

	if *config.Temperature != 2.0 {
		t.Errorf("expected Temperature 2.0, got %.1f", *config.Temperature)
	}
}

// Test 1.2.4: Missing Temperature defaults to 0.7
func TestMissingTemperatureDefaultsTo0Point7(t *testing.T) {
	// Create a temporary YAML file without temperature
	yamlContent := `
id: test-agent
name: TestAgent
role: Helper
model: gpt-4o
backstory: A test agent
`

	tempFile, err := os.CreateTemp("", "agent-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(yamlContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tempFile.Close()

	config, err := LoadAgentConfig(tempFile.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if config.Temperature == nil {
		t.Fatal("expected Temperature to be set to default")
	}

	if *config.Temperature != 0.7 {
		t.Errorf("expected default Temperature 0.7, got %.1f", *config.Temperature)
	}
}

// Test 1.2.5: Boundary temperatures (0.0, 0.5, 1.0, 1.5, 2.0) work correctly
func TestBoundaryTemperaturesWork(t *testing.T) {
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

			if config.Temperature == nil {
				t.Fatal("expected Temperature to be set")
			}

			if *config.Temperature != tt.temperature {
				t.Errorf("expected Temperature %.1f, got %.1f", tt.temperature, *config.Temperature)
			}
		})
	}
}

// Test 1.2.6: CreateAgentFromConfig dereferences temperature correctly
func TestCreateAgentFromConfigDereferencesTemperature(t *testing.T) {
	tests := []struct {
		name        string
		temp        *float64
		expectedTemp float64
	}{
		{
			name: "explicit temperature",
			temp: func() *float64 { t := 0.0; return &t }(),
			expectedTemp: 0.0,
		},
		{
			name: "explicit 2.0",
			temp: func() *float64 { t := 2.0; return &t }(),
			expectedTemp: 2.0,
		},
		{
			name: "nil uses default",
			temp: nil,
			expectedTemp: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &AgentConfig{
				ID:          "test",
				Name:        "TestAgent",
				Model:       "gpt-4o",
				Temperature: tt.temp,
			}

			agent := CreateAgentFromConfig(config, map[string]*Tool{})

			if agent.Temperature != tt.expectedTemp {
				t.Errorf("expected agent Temperature %.1f, got %.1f", tt.expectedTemp, agent.Temperature)
			}
		})
	}
}

// Test 1.3.1: Empty Model field returns error
func TestValidateAgentConfigEmptyModel(t *testing.T) {
	config := &AgentConfig{
		ID:   "test",
		Name: "TestAgent",
		Model: "",
	}

	err := ValidateAgentConfig(config)

	if err == nil {
		t.Fatal("expected error for empty Model, got nil")
	}

	if !contains(err.Error(), "Model must be specified") {
		t.Errorf("expected helpful error message, got: %v", err)
	}
}

// Test 1.3.2: Invalid Temperature > 2.0 returns error
func TestValidateAgentConfigTemperature2Point1Error(t *testing.T) {
	temp := 2.1
	config := &AgentConfig{
		ID:          "test",
		Name:        "TestAgent",
		Model:       "gpt-4o",
		Temperature: &temp,
	}

	err := ValidateAgentConfig(config)

	if err == nil {
		t.Fatal("expected error for temperature 2.1")
	}

	if !contains(err.Error(), "between 0.0 and 2.0") {
		t.Errorf("expected range error message, got: %v", err)
	}
}

// Test 1.3.3: Invalid Temperature < 0.0 returns error
func TestValidateAgentConfigTemperatureNegativeError(t *testing.T) {
	temp := -1.0
	config := &AgentConfig{
		ID:          "test",
		Name:        "TestAgent",
		Model:       "gpt-4o",
		Temperature: &temp,
	}

	err := ValidateAgentConfig(config)

	if err == nil {
		t.Fatal("expected error for negative temperature")
	}

	if !contains(err.Error(), "between 0.0 and 2.0") {
		t.Errorf("expected range error message, got: %v", err)
	}
}

// Test 1.3.4: Valid configuration passes validation
func TestValidateAgentConfigValid(t *testing.T) {
	temp := 0.7
	config := &AgentConfig{
		ID:          "test",
		Name:        "TestAgent",
		Model:       "gpt-4o",
		Temperature: &temp,
	}

	err := ValidateAgentConfig(config)

	if err != nil {
		t.Errorf("expected nil error for valid config, got: %v", err)
	}
}

// Test 1.3.5: Boundary temperatures pass validation
func TestValidateAgentConfigBoundaryTemperatures(t *testing.T) {
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

// Test 1.3.6: LoadAgentConfig validates after loading
func TestLoadAgentConfigValidatesConfig(t *testing.T) {
	// Create temp YAML with invalid temperature
	invalidYAML := `
id: test
name: TestAgent
role: Helper
model: gpt-4o
temperature: 2.5
`

	tempFile, err := os.CreateTemp("", "agent-invalid-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(invalidYAML); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tempFile.Close()

	_, err = LoadAgentConfig(tempFile.Name())

	if err == nil {
		t.Fatal("expected LoadAgentConfig to return error for invalid config")
	}

	if !contains(err.Error(), "between 0.0 and 2.0") {
		t.Errorf("expected validation error, got: %v", err)
	}
}

// Test 1.3.7: Valid config file loads successfully
func TestLoadAgentConfigValid(t *testing.T) {
	validYAML := `
id: test-agent
name: TestAgent
role: Helper
backstory: A helpful agent
model: gpt-4o
temperature: 0.7
`

	tempFile, err := os.CreateTemp("", "agent-valid-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(validYAML); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tempFile.Close()

	config, err := LoadAgentConfig(tempFile.Name())

	if err != nil {
		t.Errorf("expected LoadAgentConfig to succeed, got error: %v", err)
	}

	if config == nil {
		t.Fatal("expected config to be loaded")
	}

	if config.Model != "gpt-4o" {
		t.Errorf("expected model gpt-4o, got %s", config.Model)
	}
}

// Test 1.3.8: Backward compatibility with old configs (no temperature specified)
func TestLoadAgentConfigBackwardCompatibility(t *testing.T) {
	oldYAML := `
id: old-agent
name: OldAgent
role: Helper
model: gpt-4o-mini
`

	tempFile, err := os.CreateTemp("", "agent-old-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(oldYAML); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tempFile.Close()

	config, err := LoadAgentConfig(tempFile.Name())

	if err != nil {
		t.Errorf("expected backward compatibility, got error: %v", err)
	}

	// Temperature should default to 0.7
	if config.Temperature == nil {
		t.Fatal("expected Temperature to have default value")
	}

	if *config.Temperature != 0.7 {
		t.Errorf("expected default temperature 0.7, got %.1f", *config.Temperature)
	}
}

// Test 2.1: LoadAgentConfig with file not found
func TestLoadAgentConfigFileNotFound(t *testing.T) {
	_, err := LoadAgentConfig("/nonexistent/path/to/agent.yaml")

	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	if !contains(err.Error(), "no such file") && !contains(err.Error(), "not found") {
		t.Errorf("expected file not found error, got: %v", err)
	}
}

// Test 2.2: LoadAgentConfig with invalid YAML
func TestLoadAgentConfigInvalidYAML(t *testing.T) {
	invalidYAML := `
id: test
name: TestAgent
invalid: [ this is not valid yaml
`

	tempFile, err := os.CreateTemp("", "agent-invalid-yaml-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(invalidYAML); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tempFile.Close()

	_, err = LoadAgentConfig(tempFile.Name())

	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

// Test 3.1: LoadAgentConfigs from directory with valid files
func TestLoadAgentConfigsBasic(t *testing.T) {
	dir := t.TempDir()

	// Create agent1.yaml
	agent1YAML := `
id: agent1
name: Agent1
model: gpt-4o
temperature: 0.7
`
	if err := os.WriteFile(dir+"/agent1.yaml", []byte(agent1YAML), 0644); err != nil {
		t.Fatalf("failed to write agent1.yaml: %v", err)
	}

	// Create agent2.yaml
	agent2YAML := `
id: agent2
name: Agent2
model: gpt-4o-mini
temperature: 0.5
`
	if err := os.WriteFile(dir+"/agent2.yaml", []byte(agent2YAML), 0644); err != nil {
		t.Fatalf("failed to write agent2.yaml: %v", err)
	}

	configs, err := LoadAgentConfigs(dir)

	if err != nil {
		t.Errorf("expected LoadAgentConfigs to succeed, got error: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}

	// Verify both agents are in the map
	if _, exists := configs["agent1"]; !exists {
		t.Error("expected agent1 in configs")
	}
	if _, exists := configs["agent2"]; !exists {
		t.Error("expected agent2 in configs")
	}
}

// Test 3.2: LoadAgentConfigs ignores non-YAML files
func TestLoadAgentConfigsIgnoresNonYAML(t *testing.T) {
	dir := t.TempDir()

	// Create valid agent config
	agentYAML := `
id: agent1
name: Agent1
model: gpt-4o
`
	if err := os.WriteFile(dir+"/agent.yaml", []byte(agentYAML), 0644); err != nil {
		t.Fatalf("failed to write agent.yaml: %v", err)
	}

	// Create non-YAML files that should be ignored
	if err := os.WriteFile(dir+"/readme.txt", []byte("This is a readme"), 0644); err != nil {
		t.Fatalf("failed to write readme.txt: %v", err)
	}

	configs, err := LoadAgentConfigs(dir)

	if err != nil {
		t.Errorf("expected LoadAgentConfigs to succeed, got error: %v", err)
	}

	if len(configs) != 1 {
		t.Errorf("expected 1 config (non-YAML ignored), got %d", len(configs))
	}
}

// Test 3.3: LoadAgentConfigs with empty directory
func TestLoadAgentConfigsEmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	configs, err := LoadAgentConfigs(dir)

	if err != nil {
		t.Errorf("expected LoadAgentConfigs to handle empty directory, got error: %v", err)
	}

	if len(configs) != 0 {
		t.Errorf("expected 0 configs for empty directory, got %d", len(configs))
	}
}

// Test 3.4: LoadAgentConfigs with directory not found
func TestLoadAgentConfigsDirectoryNotFound(t *testing.T) {
	_, err := LoadAgentConfigs("/nonexistent/directory/path")

	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

// Test 3.5: LoadAgentConfigs skips configs without ID
func TestLoadAgentConfigsSkipsNoID(t *testing.T) {
	dir := t.TempDir()

	// Create config with ID (should be included)
	validYAML := `
id: agent-with-id
name: ValidAgent
model: gpt-4o
`
	if err := os.WriteFile(dir+"/valid.yaml", []byte(validYAML), 0644); err != nil {
		t.Fatalf("failed to write valid.yaml: %v", err)
	}

	// Create config without ID (should be skipped)
	invalidYAML := `
name: NoIDAgent
model: gpt-4o
`
	if err := os.WriteFile(dir+"/no-id.yaml", []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write no-id.yaml: %v", err)
	}

	configs, err := LoadAgentConfigs(dir)

	if err != nil {
		t.Errorf("expected LoadAgentConfigs to skip invalid configs, got error: %v", err)
	}

	if len(configs) != 1 {
		t.Errorf("expected 1 config (no-ID skipped), got %d", len(configs))
	}

	// Verify the valid config exists in map
	if _, exists := configs["agent-with-id"]; !exists {
		t.Error("expected agent-with-id in configs")
	}
}

// Test 4.1: LoadTeamConfig basic functionality
func TestLoadTeamConfigBasic(t *testing.T) {
	dir := t.TempDir()

	teamYAML := `
version: "1.0"
description: "Test Team"
entry_point: "agent1"
agents:
  - agent1
  - agent2
settings:
  max_handoffs: 10
  max_rounds: 100
  timeout_seconds: 300
`
	teamFile := dir + "/team.yaml"
	if err := os.WriteFile(teamFile, []byte(teamYAML), 0644); err != nil {
		t.Fatalf("failed to write team.yaml: %v", err)
	}

	config, err := LoadTeamConfig(teamFile)

	if err != nil {
		t.Errorf("expected LoadTeamConfig to succeed, got error: %v", err)
	}

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	if config.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", config.Version)
	}

	if config.Settings.MaxHandoffs != 10 {
		t.Errorf("expected max_handoffs 10, got %d", config.Settings.MaxHandoffs)
	}

	if len(config.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(config.Agents))
	}
}

// Test 4.2: LoadTeamConfig with file not found
func TestLoadTeamConfigFileNotFound(t *testing.T) {
	_, err := LoadTeamConfig("/nonexistent/path/to/team.yaml")

	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

	if !contains(err.Error(), "no such file") && !contains(err.Error(), "not found") {
		t.Errorf("expected file not found error, got: %v", err)
	}
}

// Test 4.3: LoadTeamConfig with invalid YAML
func TestLoadTeamConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()

	invalidYAML := `
version: "1.0"
invalid: [ this is not valid yaml
`
	teamFile := dir + "/team.yaml"
	if err := os.WriteFile(teamFile, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write team.yaml: %v", err)
	}

	_, err := LoadTeamConfig(teamFile)

	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

// Test 4.4: LoadTeamConfig with minimal config
func TestLoadTeamConfigMinimal(t *testing.T) {
	dir := t.TempDir()

	teamYAML := `
version: "1.0"
entry_point: "agent1"
agents:
  - agent1
`
	teamFile := dir + "/team.yaml"
	if err := os.WriteFile(teamFile, []byte(teamYAML), 0644); err != nil {
		t.Fatalf("failed to write team.yaml: %v", err)
	}

	config, err := LoadTeamConfig(teamFile)

	if err != nil {
		t.Errorf("expected LoadTeamConfig to succeed, got error: %v", err)
	}

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	if len(config.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(config.Agents))
	}
}

// Test 5.1: CreateAgentFromConfig with basic config
func TestCreateAgentFromConfigBasic(t *testing.T) {
	config := &AgentConfig{
		ID:   "agent1",
		Name: "Agent1",
		Model: "gpt-4o",
	}

	agent := CreateAgentFromConfig(config, map[string]*Tool{})

	if agent == nil {
		t.Fatal("expected non-nil agent")
	}

	if agent.ID != "agent1" {
		t.Errorf("expected ID agent1, got %s", agent.ID)
	}

	if agent.Name != "Agent1" {
		t.Errorf("expected Name Agent1, got %s", agent.Name)
	}

	if agent.Model != "gpt-4o" {
		t.Errorf("expected Model gpt-4o, got %s", agent.Model)
	}
}

// Test 5.2: CreateAgentFromConfig with custom temperature
func TestCreateAgentFromConfigWithTemperature(t *testing.T) {
	temp := 0.5
	config := &AgentConfig{
		ID:          "agent1",
		Name:        "Agent1",
		Model:       "gpt-4o",
		Temperature: &temp,
	}

	agent := CreateAgentFromConfig(config, map[string]*Tool{})

	if agent.Temperature != 0.5 {
		t.Errorf("expected Temperature 0.5, got %.1f", agent.Temperature)
	}
}

// Test 5.3: CreateAgentFromConfig with nil temperature uses default
func TestCreateAgentFromConfigDefaultTemperature(t *testing.T) {
	config := &AgentConfig{
		ID:    "agent1",
		Name:  "Agent1",
		Model: "gpt-4o",
	}

	agent := CreateAgentFromConfig(config, map[string]*Tool{})

	if agent.Temperature != 0.7 {
		t.Errorf("expected default Temperature 0.7, got %.1f", agent.Temperature)
	}
}

// Test 6.1: LoadCrewConfig with valid team config (wrapper around LoadTeamConfig)
func TestLoadCrewConfigBasic(t *testing.T) {
	dir := t.TempDir()

	teamYAML := `
version: "1.0"
description: "Test Team"
entry_point: "orchestrator"
agents:
  - orchestrator
  - executor
settings:
  max_handoffs: 10
  max_rounds: 100
`
	teamFile := dir + "/team.yaml"
	if err := os.WriteFile(teamFile, []byte(teamYAML), 0644); err != nil {
		t.Fatalf("failed to write team.yaml: %v", err)
	}

	config, err := LoadCrewConfig(teamFile)

	if err != nil {
		t.Errorf("expected LoadCrewConfig to succeed, got error: %v", err)
	}

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	if config.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", config.Version)
	}

	if len(config.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(config.Agents))
	}
}

// Test 6.2: LoadCrewConfig with file not found
func TestLoadCrewConfigFileNotFound(t *testing.T) {
	_, err := LoadCrewConfig("/nonexistent/team.yaml")

	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}
