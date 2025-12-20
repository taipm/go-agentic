package agentic

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

const (
	testModelGPT4 = "gpt-4o"
	testModelMini = "gpt-4o-mini"
)

// ============================================
// Helper Functions
// ============================================

func createTestYAMLFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "team.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test YAML file: %v", err)
	}
	return filePath
}

func getBasicToolHandlers() ToolHandlerRegistry {
	return ToolHandlerRegistry{
		"get_metrics": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "metrics_result", nil
		},
		"get_status": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "status_ok", nil
		},
	}
}

// ============================================
// LoadTeamFromYAML Tests
// ============================================

func TestLoadTeamFromYAMLBasic(t *testing.T) {
	yamlContent := `
team:
  name: "Test Team"
  config:
    maxRounds: 5
    maxHandoffs: 2

agents:
  orchestrator:
    id: "orchestrator"
    name: "Orchestrator"
    role: "Route requests"
    backstory: "You route requests to specialists"
    model: gpt-4o
    temperature: 0.7
    isTerminal: false
    tools: []

  executor:
    id: "executor"
    name: "Executor"
    role: "Execute tasks"
    backstory: "You execute tasks"
    model: gpt-4o-mini
    temperature: 0.5
    isTerminal: true
    tools:
      - get_metrics

tools:
  get_metrics:
    name: "GetMetrics"
    description: "Get system metrics"
`

	yamlPath := createTestYAMLFile(t, yamlContent)
	team, err := LoadTeamFromYAML(yamlPath, getBasicToolHandlers())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(team.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(team.Agents))
	}

	if team.MaxRounds != 5 {
		t.Errorf("expected maxRounds 5, got %d", team.MaxRounds)
	}

	if team.MaxHandoffs != 2 {
		t.Errorf("expected maxHandoffs 2, got %d", team.MaxHandoffs)
	}
}

func TestLoadTeamFromYAMLWithTools(t *testing.T) {
	yamlContent := `
team:
  name: "Test Team"
  config:
    maxRounds: 3
    maxHandoffs: 1

agents:
  worker:
    id: "worker"
    name: "Worker"
    role: "Execute"
    backstory: "You work"
    model: "gpt-4o-mini"
    temperature: 0.7
    isTerminal: true
    tools:
      - get_metrics
      - get_status

tools:
  get_metrics:
    name: "GetMetrics"
    description: "Get metrics"
  get_status:
    name: "GetStatus"
    description: "Get status"
`

	yamlPath := createTestYAMLFile(t, yamlContent)
	team, err := LoadTeamFromYAML(yamlPath, getBasicToolHandlers())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(team.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(team.Agents))
	}

	executor := team.Agents[0]
	if len(executor.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(executor.Tools))
	}

	if executor.Tools[0].Name != "GetMetrics" && executor.Tools[1].Name != "GetMetrics" {
		t.Error("expected GetMetrics tool to be present")
	}
}

func TestLoadTeamFromYAMLMissingFile(t *testing.T) {
	handlers := getBasicToolHandlers()
	_, err := LoadTeamFromYAML("/nonexistent/path/team.yaml", handlers)

	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadTeamFromYAMLInvalidYAML(t *testing.T) {
	yamlContent := `
this: is: not: valid: yaml: [
`

	yamlPath := createTestYAMLFile(t, yamlContent)
	handlers := getBasicToolHandlers()
	_, err := LoadTeamFromYAML(yamlPath, handlers)

	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadTeamFromYAMLMissingAgent(t *testing.T) {
	yamlContent := `
team:
  config:
    maxRounds: 5
    maxHandoffs: 2

agents: {}

tools: {}
`

	yamlPath := createTestYAMLFile(t, yamlContent)
	handlers := getBasicToolHandlers()
	_, err := LoadTeamFromYAML(yamlPath, handlers)

	if err == nil {
		t.Error("expected error for empty agents")
	}
}

func TestLoadTeamFromYAMLNoTerminalAgent(t *testing.T) {
	yamlContent := `
team:
  config:
    maxRounds: 5
    maxHandoffs: 2

agents:
  agent1:
    id: "agent1"
    name: "Agent 1"
    role: "Role"
    backstory: "Backstory"
    model: "gpt-4o"
    temperature: 0.7
    isTerminal: false
    tools: []

tools: {}
`

	yamlPath := createTestYAMLFile(t, yamlContent)
	handlers := getBasicToolHandlers()
	_, err := LoadTeamFromYAML(yamlPath, handlers)

	if err == nil {
		t.Error("expected error for no terminal agent")
	}
}

func TestLoadTeamFromYAMLMissingToolHandler(t *testing.T) {
	yamlContent := `
team:
  config:
    maxRounds: 5
    maxHandoffs: 2

agents:
  worker:
    id: "worker"
    name: "Worker"
    role: "Work"
    backstory: "Works"
    model: "gpt-4o"
    temperature: 0.7
    isTerminal: true
    tools:
      - missing_tool

tools:
  missing_tool:
    name: "MissingTool"
    description: "Missing"
`

	yamlPath := createTestYAMLFile(t, yamlContent)
	handlers := ToolHandlerRegistry{} // No handlers
	_, err := LoadTeamFromYAML(yamlPath, handlers)

	if err == nil {
		t.Error("expected error for missing tool handler")
	}
}

func TestLoadTeamFromYAMLToolReferencedButNotDefined(t *testing.T) {
	yamlContent := `
team:
  config:
    maxRounds: 5
    maxHandoffs: 2

agents:
  worker:
    id: "worker"
    name: "Worker"
    role: "Work"
    backstory: "Works"
    model: "gpt-4o"
    temperature: 0.7
    isTerminal: true
    tools:
      - undefined_tool

tools: {}
`

	yamlPath := createTestYAMLFile(t, yamlContent)
	handlers := ToolHandlerRegistry{
		"undefined_tool": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result", nil
		},
	}
	_, err := LoadTeamFromYAML(yamlPath, handlers)

	if err == nil {
		t.Error("expected error for tool not defined in tools section")
	}
}

// ============================================
// Validation Tests
// ============================================

func TestValidateUnifiedTeamConfigMaxRoundsZero(t *testing.T) {
	config := &UnifiedTeamConfig{
		Team: UnifiedTeamMetadata{
			Config: struct {
				MaxRounds   int `yaml:"maxRounds"`
				MaxHandoffs int `yaml:"maxHandoffs"`
			}{
				MaxRounds:   0,
				MaxHandoffs: 2,
			},
		},
		Agents: map[string]*UnifiedAgentConfig{
			"agent": {
				Name:       "Agent",
				Role:       "Role",
				Backstory:  "Story",
				Model:      testModelGPT4,
				IsTerminal: true,
			},
		},
	}

	err := validateUnifiedTeamConfig(config)
	if err == nil {
		t.Error("expected error for maxRounds <= 0")
	}
}

func TestValidateUnifiedTeamConfigMaxHandoffsNegative(t *testing.T) {
	config := &UnifiedTeamConfig{
		Team: UnifiedTeamMetadata{
			Config: struct {
				MaxRounds   int `yaml:"maxRounds"`
				MaxHandoffs int `yaml:"maxHandoffs"`
			}{
				MaxRounds:   5,
				MaxHandoffs: -1,
			},
		},
		Agents: map[string]*UnifiedAgentConfig{
			"agent": {
				Name:       "Agent",
				Role:       "Role",
				Backstory:  "Story",
				Model:      testModelGPT4,
				IsTerminal: true,
			},
		},
	}

	err := validateUnifiedTeamConfig(config)
	if err == nil {
		t.Error("expected error for maxHandoffs < 0")
	}
}

func TestValidateUnifiedTeamConfigMissingAgentName(t *testing.T) {
	config := &UnifiedTeamConfig{
		Team: UnifiedTeamMetadata{
			Config: struct {
				MaxRounds   int `yaml:"maxRounds"`
				MaxHandoffs int `yaml:"maxHandoffs"`
			}{
				MaxRounds:   5,
				MaxHandoffs: 2,
			},
		},
		Agents: map[string]*UnifiedAgentConfig{
			"agent": {
				Name:       "",
				Role:       "Role",
				Backstory:  "Story",
				Model:      testModelGPT4,
				IsTerminal: true,
			},
		},
	}

	err := validateUnifiedTeamConfig(config)
	if err == nil {
		t.Error("expected error for missing agent name")
	}
}

func TestValidateUnifiedTeamConfigMissingAgentRole(t *testing.T) {
	config := &UnifiedTeamConfig{
		Team: UnifiedTeamMetadata{
			Config: struct {
				MaxRounds   int `yaml:"maxRounds"`
				MaxHandoffs int `yaml:"maxHandoffs"`
			}{
				MaxRounds:   5,
				MaxHandoffs: 2,
			},
		},
		Agents: map[string]*UnifiedAgentConfig{
			"agent": {
				Name:       "Agent",
				Role:       "",
				Backstory:  "Story",
				Model:      testModelGPT4,
				IsTerminal: true,
			},
		},
	}

	err := validateUnifiedTeamConfig(config)
	if err == nil {
		t.Error("expected error for missing agent role")
	}
}

func TestValidateUnifiedTeamConfigMissingAgentBackstory(t *testing.T) {
	config := &UnifiedTeamConfig{
		Team: UnifiedTeamMetadata{
			Config: struct {
				MaxRounds   int `yaml:"maxRounds"`
				MaxHandoffs int `yaml:"maxHandoffs"`
			}{
				MaxRounds:   5,
				MaxHandoffs: 2,
			},
		},
		Agents: map[string]*UnifiedAgentConfig{
			"agent": {
				Name:       "Agent",
				Role:       "Role",
				Backstory:  "",
				Model:      testModelGPT4,
				IsTerminal: true,
			},
		},
	}

	err := validateUnifiedTeamConfig(config)
	if err == nil {
		t.Error("expected error for missing agent backstory")
	}
}

func TestValidateUnifiedTeamConfigMissingAgentModel(t *testing.T) {
	config := &UnifiedTeamConfig{
		Team: UnifiedTeamMetadata{
			Config: struct {
				MaxRounds   int `yaml:"maxRounds"`
				MaxHandoffs int `yaml:"maxHandoffs"`
			}{
				MaxRounds:   5,
				MaxHandoffs: 2,
			},
		},
		Agents: map[string]*UnifiedAgentConfig{
			"agent": {
				Name:       "Agent",
				Role:       "Role",
				Backstory:  "Story",
				Model:      "",
				IsTerminal: true,
			},
		},
	}

	err := validateUnifiedTeamConfig(config)
	if err == nil {
		t.Error("expected error for missing agent model")
	}
}

func TestValidateUnifiedTeamConfigAgentReferencesUndefinedTool(t *testing.T) {
	config := &UnifiedTeamConfig{
		Team: UnifiedTeamMetadata{
			Config: struct {
				MaxRounds   int `yaml:"maxRounds"`
				MaxHandoffs int `yaml:"maxHandoffs"`
			}{
				MaxRounds:   5,
				MaxHandoffs: 2,
			},
		},
		Agents: map[string]*UnifiedAgentConfig{
			"agent": {
				Name:       "Agent",
				Role:       "Role",
				Backstory:  "Story",
				Model:      testModelGPT4,
				IsTerminal: true,
				Tools:      []string{"undefined_tool"},
			},
		},
		Tools: map[string]*UnifiedToolConfig{},
	}

	err := validateUnifiedTeamConfig(config)
	if err == nil {
		t.Error("expected error for agent referencing undefined tool")
	}
}

func TestValidateUnifiedTeamConfigMissingToolName(t *testing.T) {
	config := &UnifiedTeamConfig{
		Team: UnifiedTeamMetadata{
			Config: struct {
				MaxRounds   int `yaml:"maxRounds"`
				MaxHandoffs int `yaml:"maxHandoffs"`
			}{
				MaxRounds:   5,
				MaxHandoffs: 2,
			},
		},
		Agents: map[string]*UnifiedAgentConfig{
			"agent": {
				Name:       "Agent",
				Role:       "Role",
				Backstory:  "Story",
				Model:      testModelGPT4,
				IsTerminal: true,
			},
		},
		Tools: map[string]*UnifiedToolConfig{
			"tool": {
				Name:        "",
				Description: "Description",
			},
		},
	}

	err := validateUnifiedTeamConfig(config)
	if err == nil {
		t.Error("expected error for missing tool name")
	}
}

func TestValidateUnifiedTeamConfigMissingToolDescription(t *testing.T) {
	config := &UnifiedTeamConfig{
		Team: UnifiedTeamMetadata{
			Config: struct {
				MaxRounds   int `yaml:"maxRounds"`
				MaxHandoffs int `yaml:"maxHandoffs"`
			}{
				MaxRounds:   5,
				MaxHandoffs: 2,
			},
		},
		Agents: map[string]*UnifiedAgentConfig{
			"agent": {
				Name:       "Agent",
				Role:       "Role",
				Backstory:  "Story",
				Model:      testModelGPT4,
				IsTerminal: true,
			},
		},
		Tools: map[string]*UnifiedToolConfig{
			"tool": {
				Name:        "Tool",
				Description: "",
			},
		},
	}

	err := validateUnifiedTeamConfig(config)
	if err == nil {
		t.Error("expected error for missing tool description")
	}
}

// ============================================
// LoadTeamFromYAMLWithDefaults Tests
// ============================================

func TestLoadTeamFromYAMLWithDefaults(t *testing.T) {
	yamlContent := `
team:
  config:
    maxRounds: 0
    maxHandoffs: -1

agents:
  agent:
    id: "agent"
    name: "Agent"
    role: "Role"
    backstory: "Story"
    model: gpt-4o
    temperature: 0.7
    isTerminal: true
    tools: []

tools: {}
`

	yamlPath := createTestYAMLFile(t, yamlContent)
	_, err := LoadTeamFromYAMLWithDefaults(yamlPath, ToolHandlerRegistry{})

	// Validation should still fail for invalid config
	if err == nil {
		t.Error("expected validation error even with defaults")
	}
}

// ============================================
// ExportTeamToYAML Tests
// ============================================

func TestExportTeamToYAML(t *testing.T) {
	agent := NewAgent("test", "Test Agent").
		WithRole("Tester").
		WithBackstory("Test backstory").
		WithModel("gpt-4o").
		WithTemperature(0.7).
		SetTerminal(true).
		Build()

	team := NewTeam().
		AddAgent(agent).
		WithMaxRounds(5).
		WithMaxHandoffs(2).
		Build()

	yamlBytes, err := ExportTeamToYAML(team)
	if err != nil {
		t.Fatalf("expected no error exporting team, got %v", err)
	}

	if len(yamlBytes) == 0 {
		t.Error("expected non-empty YAML output")
	}
}

func TestExportTeamToYAMLRoundTrip(t *testing.T) {
	// Create original team
	tool := NewTool("metric", "Get metric").
		WithParameter("name", "string", "metric name", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result", nil
		}).
		Build()

	agent := NewAgent("worker", "Worker").
		WithRole("Execute").
		WithBackstory("Executes tasks").
		WithModel("gpt-4o-mini").
		WithTemperature(0.5).
		SetTerminal(true).
		AddTool(tool).
		Build()

	originalTeam := NewTeam().
		AddAgent(agent).
		WithMaxRounds(5).
		WithMaxHandoffs(2).
		Build()

	// Export to YAML
	yamlBytes, err := ExportTeamToYAML(originalTeam)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Write to temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "team.yaml")
	if err := os.WriteFile(filePath, yamlBytes, 0644); err != nil {
		t.Fatalf("failed to write YAML: %v", err)
	}

	// Load from YAML
	handlers := ToolHandlerRegistry{
		"metric": func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result", nil
		},
	}
	loadedTeam, err := LoadTeamFromYAML(filePath, handlers)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Verify round-trip integrity
	if len(loadedTeam.Agents) != len(originalTeam.Agents) {
		t.Errorf("expected %d agents, got %d", len(originalTeam.Agents), len(loadedTeam.Agents))
	}

	if loadedTeam.MaxRounds != originalTeam.MaxRounds {
		t.Errorf("expected maxRounds %d, got %d", originalTeam.MaxRounds, loadedTeam.MaxRounds)
	}

	if loadedTeam.MaxHandoffs != originalTeam.MaxHandoffs {
		t.Errorf("expected maxHandoffs %d, got %d", originalTeam.MaxHandoffs, loadedTeam.MaxHandoffs)
	}
}
