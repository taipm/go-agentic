package agentic

import (
	"context"
	"testing"
)

// ============================================
// AgentBuilder Tests
// ============================================

func TestNewAgentBuilderSetsDefaults(t *testing.T) {
	agent := NewAgent("test-id", "Test Agent").
		WithRole("Test Role").
		WithBackstory("Test Backstory").
		Build()

	if agent.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got '%s'", agent.ID)
	}
	if agent.Name != "Test Agent" {
		t.Errorf("expected Name 'Test Agent', got '%s'", agent.Name)
	}
	if agent.Model != "gpt-4o-mini" {
		t.Errorf("expected default model 'gpt-4o-mini', got '%s'", agent.Model)
	}
	if agent.Temperature != 0.7 {
		t.Errorf("expected default temperature 0.7, got %f", agent.Temperature)
	}
}

func TestAgentBuilderWithRole(t *testing.T) {
	agent := NewAgent("id", "name").
		WithRole("Test Role").
		WithBackstory("backstory").
		Build()

	if agent.Role != "Test Role" {
		t.Errorf("expected role 'Test Role', got '%s'", agent.Role)
	}
}

func TestAgentBuilderWithBackstory(t *testing.T) {
	backstory := "This is a test backstory"
	agent := NewAgent("id", "name").
		WithRole("role").
		WithBackstory(backstory).
		Build()

	if agent.Backstory != backstory {
		t.Errorf("expected backstory '%s', got '%s'", backstory, agent.Backstory)
	}
}

func TestAgentBuilderWithModel(t *testing.T) {
	agent := NewAgent("id", "name").
		WithRole("role").
		WithBackstory("backstory").
		WithModel("gpt-4o").
		Build()

	if agent.Model != "gpt-4o" {
		t.Errorf("expected model 'gpt-4o', got '%s'", agent.Model)
	}
}

func TestAgentBuilderWithTemperature(t *testing.T) {
	agent := NewAgent("id", "name").
		WithRole("role").
		WithBackstory("backstory").
		WithTemperature(0.5).
		Build()

	if agent.Temperature != 0.5 {
		t.Errorf("expected temperature 0.5, got %f", agent.Temperature)
	}
}

func TestAgentBuilderTemperatureValidation(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid temperature")
		}
	}()

	NewAgent("id", "name").WithRole("role").WithBackstory("backstory").WithTemperature(2.5).Build()
}

func TestAgentBuilderAddTool(t *testing.T) {
	tool := NewTool("test", "test tool").
		NoParameters().
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result", nil
		}).
		Build()

	agent := NewAgent("id", "name").
		WithRole("role").
		WithBackstory("backstory").
		AddTool(tool).
		Build()

	if len(agent.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(agent.Tools))
	}
	if agent.Tools[0].Name != "test" {
		t.Errorf("expected tool name 'test', got '%s'", agent.Tools[0].Name)
	}
}

func TestAgentBuilderAddMultipleTools(t *testing.T) {
	tool1 := NewTool("tool1", "first tool").
		NoParameters().
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result1", nil
		}).
		Build()

	tool2 := NewTool("tool2", "second tool").
		NoParameters().
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result2", nil
		}).
		Build()

	agent := NewAgent("id", "name").
		WithRole("role").
		WithBackstory("backstory").
		AddTools(tool1, tool2).
		Build()

	if len(agent.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(agent.Tools))
	}
}

func TestAgentBuilderSetTerminal(t *testing.T) {
	agent := NewAgent("id", "name").
		WithRole("role").
		WithBackstory("backstory").
		SetTerminal(true).
		Build()

	if !agent.IsTerminal {
		t.Error("expected agent to be terminal")
	}
}

func TestAgentBuilderWithHandoff(t *testing.T) {
	agent := NewAgent("id", "name").
		WithRole("role").
		WithBackstory("backstory").
		WithHandoff("other-agent").
		Build()

	if len(agent.HandoffTargets) != 1 {
		t.Errorf("expected 1 handoff target, got %d", len(agent.HandoffTargets))
	}
	if agent.HandoffTargets[0] != "other-agent" {
		t.Errorf("expected handoff target 'other-agent', got '%s'", agent.HandoffTargets[0])
	}
}

func TestAgentBuilderWithHandoffs(t *testing.T) {
	agent := NewAgent("id", "name").
		WithRole("role").
		WithBackstory("backstory").
		WithHandoffs("agent1", "agent2", "agent3").
		Build()

	if len(agent.HandoffTargets) != 3 {
		t.Errorf("expected 3 handoff targets, got %d", len(agent.HandoffTargets))
	}
}

func TestAgentBuilderMissingRole(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing role")
		}
	}()

	NewAgent("id", "name").WithBackstory("test").Build()
}

func TestAgentBuilderMissingBackstory(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing backstory")
		}
	}()

	NewAgent("id", "name").WithRole("test").Build()
}

func TestAgentBuilderChaining(t *testing.T) {
	agent := NewAgent("orchestrator", "Orchestrator").
		WithRole("Route tasks").
		WithBackstory("You are an orchestrator").
		WithModel("gpt-4o").
		WithTemperature(0.8).
		SetTerminal(false).
		WithHandoffs("executor", "clarifier").
		Build()

	if agent.ID != "orchestrator" ||
		agent.Name != "Orchestrator" ||
		agent.Role != "Route tasks" ||
		agent.Model != "gpt-4o" ||
		agent.Temperature != 0.8 ||
		agent.IsTerminal ||
		len(agent.HandoffTargets) != 2 {
		t.Error("chaining builder calls failed")
	}
}

// ============================================
// ToolBuilder Tests
// ============================================

func TestNewToolBuilderSetsDefaults(t *testing.T) {
	tool := NewTool("test", "test description").
		NoParameters().
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result", nil
		}).
		Build()

	if tool.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", tool.Name)
	}
	if tool.Description != "test description" {
		t.Errorf("expected description 'test description', got '%s'", tool.Description)
	}
}

func TestToolBuilderNoParameters(t *testing.T) {
	tool := NewTool("test", "test").
		NoParameters().
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result", nil
		}).
		Build()

	if tool.Handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestToolBuilderWithParameter(t *testing.T) {
	tool := NewTool("test", "test").
		WithParameter("metric", "string", "metric name", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result", nil
		}).
		Build()

	props := tool.Parameters["properties"].(map[string]interface{})
	if _, ok := props["metric"]; !ok {
		t.Error("expected 'metric' parameter to be set")
	}

	required := tool.Parameters["required"].([]string)
	if len(required) != 1 || required[0] != "metric" {
		t.Error("expected 'metric' to be in required list")
	}
}

func TestToolBuilderWithMultipleParameters(t *testing.T) {
	params := map[string]ParameterDef{
		"metric": {
			Type:        "string",
			Description: "metric name",
			Required:    true,
		},
		"threshold": {
			Type:        "number",
			Description: "threshold value",
			Required:    false,
		},
	}

	tool := NewTool("test", "test").
		WithParameters(params).
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result", nil
		}).
		Build()

	props := tool.Parameters["properties"].(map[string]interface{})
	if len(props) != 2 {
		t.Errorf("expected 2 properties, got %d", len(props))
	}

	required := tool.Parameters["required"].([]string)
	if len(required) != 1 {
		t.Errorf("expected 1 required parameter, got %d", len(required))
	}
}

func TestToolBuilderMissingHandler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing handler")
		}
	}()

	NewTool("test", "test").Build()
}

func TestToolBuilderMissingName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for missing name")
		}
	}()

	NewTool("", "description").
		NoParameters().
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "result", nil
		}).
		Build()
}

func TestToolBuilderChaining(t *testing.T) {
	tool := NewTool("GetMetrics", "Get system metrics").
		WithParameter("metric", "string", "metric name", true).
		WithParameter("period", "integer", "period in seconds", false).
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "metrics", nil
		}).
		Build()

	if tool.Name != "GetMetrics" ||
		tool.Description != "Get system metrics" ||
		tool.Handler == nil {
		t.Error("chaining builder calls failed")
	}

	props := tool.Parameters["properties"].(map[string]interface{})
	if len(props) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(props))
	}
}

// ============================================
// TeamBuilder Tests
// ============================================

func TestNewTeamBuilderSetsDefaults(t *testing.T) {
	agent1 := NewAgent("a1", "Agent 1").WithRole("role1").WithBackstory("b1").SetTerminal(true).Build()

	team := NewTeam().
		AddAgent(agent1).
		Build()

	if len(team.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(team.Agents))
	}
	if team.MaxRounds != 10 {
		t.Errorf("expected default MaxRounds 10, got %d", team.MaxRounds)
	}
	if team.MaxHandoffs != 3 {
		t.Errorf("expected default MaxHandoffs 3, got %d", team.MaxHandoffs)
	}
}

func TestTeamBuilderAddAgent(t *testing.T) {
	agent1 := NewAgent("a1", "Agent 1").WithRole("role1").WithBackstory("b1").SetTerminal(false).Build()
	agent2 := NewAgent("a2", "Agent 2").WithRole("role2").WithBackstory("b2").SetTerminal(true).Build()

	team := NewTeam().
		AddAgent(agent1).
		AddAgent(agent2).
		Build()

	if len(team.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(team.Agents))
	}
}

func TestTeamBuilderAddMultipleAgents(t *testing.T) {
	agent1 := NewAgent("a1", "Agent 1").WithRole("role1").WithBackstory("b1").SetTerminal(false).Build()
	agent2 := NewAgent("a2", "Agent 2").WithRole("role2").WithBackstory("b2").SetTerminal(true).Build()

	team := NewTeam().
		AddAgents(agent1, agent2).
		Build()

	if len(team.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(team.Agents))
	}
}

func TestTeamBuilderWithMaxRounds(t *testing.T) {
	agent := NewAgent("a1", "Agent 1").WithRole("role").WithBackstory("backstory").SetTerminal(true).Build()

	team := NewTeam().
		AddAgent(agent).
		WithMaxRounds(20).
		Build()

	if team.MaxRounds != 20 {
		t.Errorf("expected MaxRounds 20, got %d", team.MaxRounds)
	}
}

func TestTeamBuilderWithMaxHandoffs(t *testing.T) {
	agent := NewAgent("a1", "Agent 1").WithRole("role").WithBackstory("backstory").SetTerminal(true).Build()

	team := NewTeam().
		AddAgent(agent).
		WithMaxHandoffs(5).
		Build()

	if team.MaxHandoffs != 5 {
		t.Errorf("expected MaxHandoffs 5, got %d", team.MaxHandoffs)
	}
}

func TestTeamBuilderMissingAgents(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for team with no agents")
		}
	}()

	NewTeam().Build()
}

func TestTeamBuilderMissingTerminalAgent(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for team with no terminal agent")
		}
	}()

	agent := NewAgent("a1", "Agent 1").WithRole("role").WithBackstory("backstory").SetTerminal(false).Build()
	NewTeam().AddAgent(agent).Build()
}

func TestTeamBuilderChaining(t *testing.T) {
	agent1 := NewAgent("a1", "Agent 1").WithRole("role1").WithBackstory("b1").SetTerminal(false).Build()
	agent2 := NewAgent("a2", "Agent 2").WithRole("role2").WithBackstory("b2").SetTerminal(true).Build()

	team := NewTeam().
		AddAgents(agent1, agent2).
		WithMaxRounds(15).
		WithMaxHandoffs(4).
		Build()

	if len(team.Agents) != 2 ||
		team.MaxRounds != 15 ||
		team.MaxHandoffs != 4 {
		t.Error("chaining builder calls failed")
	}
}

// ============================================
// Integration Tests
// ============================================

func TestFluentAPICompleteWorkflow(t *testing.T) {
	// Create tools
	cpuTool := NewTool("GetCPU", "Get CPU usage").
		NoParameters().
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "50%", nil
		}).
		Build()

	memTool := NewTool("GetMemory", "Get memory usage").
		NoParameters().
		Handler(func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "70%", nil
		}).
		Build()

	// Create agents
	orchestrator := NewAgent("orchestrator", "Orchestrator").
		WithRole("Route requests").
		WithBackstory("You route IT requests").
		WithModel("gpt-4o").
		SetTerminal(false).
		WithHandoff("executor").
		Build()

	executor := NewAgent("executor", "Executor").
		WithRole("Execute diagnostics").
		WithBackstory("You execute diagnostics").
		WithModel("gpt-4o-mini").
		SetTerminal(true).
		AddTools(cpuTool, memTool).
		Build()

	// Create team
	team := NewTeam().
		AddAgents(orchestrator, executor).
		WithMaxRounds(10).
		WithMaxHandoffs(3).
		Build()

	// Verify team structure
	if len(team.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(team.Agents))
	}

	if len(executor.Tools) != 2 {
		t.Errorf("expected executor to have 2 tools, got %d", len(executor.Tools))
	}

	if team.MaxRounds != 10 || team.MaxHandoffs != 3 {
		t.Error("team configuration incorrect")
	}
}
