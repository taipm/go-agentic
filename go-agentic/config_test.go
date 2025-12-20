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
