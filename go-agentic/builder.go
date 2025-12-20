package agentic

import (
	"context"
	"fmt"
)

// ============================================
// AgentBuilder for fluent agent configuration
// ============================================

// AgentBuilder provides a fluent API for creating agents
type AgentBuilder struct {
	agent *Agent
}

// NewAgent starts building an agent with ID and Name
func NewAgent(id, name string) *AgentBuilder {
	return &AgentBuilder{
		agent: &Agent{
			ID:            id,
			Name:          name,
			Model:         "gpt-4o-mini", // Default, can be overridden
			Temperature:   0.7,            // Default
			Tools:         []*Tool{},
			HandoffTargets: []string{},
		},
	}
}

// WithRole sets the agent's role
func (ab *AgentBuilder) WithRole(role string) *AgentBuilder {
	ab.agent.Role = role
	return ab
}

// WithBackstory sets the agent's backstory
func (ab *AgentBuilder) WithBackstory(backstory string) *AgentBuilder {
	ab.agent.Backstory = backstory
	return ab
}

// WithModel sets the model (overrides default "gpt-4o-mini")
func (ab *AgentBuilder) WithModel(model string) *AgentBuilder {
	ab.agent.Model = model
	return ab
}

// WithTemperature sets the temperature (0.0 to 2.0)
func (ab *AgentBuilder) WithTemperature(temp float64) *AgentBuilder {
	if temp < 0 || temp > 2.0 {
		panic(fmt.Sprintf("temperature must be between 0 and 2.0, got %.2f", temp))
	}
	ab.agent.Temperature = temp
	return ab
}

// WithSystemPrompt sets a custom system prompt
func (ab *AgentBuilder) WithSystemPrompt(prompt string) *AgentBuilder {
	ab.agent.SystemPrompt = prompt
	return ab
}

// AddTool adds a single tool to the agent
func (ab *AgentBuilder) AddTool(tool *Tool) *AgentBuilder {
	if tool != nil {
		ab.agent.Tools = append(ab.agent.Tools, tool)
	}
	return ab
}

// AddTools adds multiple tools to the agent
func (ab *AgentBuilder) AddTools(tools ...*Tool) *AgentBuilder {
	for _, tool := range tools {
		if tool != nil {
			ab.agent.Tools = append(ab.agent.Tools, tool)
		}
	}
	return ab
}

// SetTerminal marks this agent as a terminal agent (final step in workflow)
func (ab *AgentBuilder) SetTerminal(terminal bool) *AgentBuilder {
	ab.agent.IsTerminal = terminal
	return ab
}

// WithHandoff adds a single handoff target
func (ab *AgentBuilder) WithHandoff(target string) *AgentBuilder {
	if target != "" {
		ab.agent.HandoffTargets = append(ab.agent.HandoffTargets, target)
	}
	return ab
}

// WithHandoffs sets all handoff targets (replaces existing)
func (ab *AgentBuilder) WithHandoffs(targets ...string) *AgentBuilder {
	ab.agent.HandoffTargets = targets
	return ab
}

// Build returns the configured agent and validates it
func (ab *AgentBuilder) Build() *Agent {
	// Validate required fields
	if ab.agent.ID == "" {
		panic("Agent ID is required")
	}
	if ab.agent.Name == "" {
		panic("Agent Name is required")
	}
	if ab.agent.Role == "" {
		panic("Agent Role is required")
	}
	if ab.agent.Backstory == "" {
		panic("Agent Backstory is required")
	}

	return ab.agent
}

// ============================================
// ToolBuilder for simpler tool definition
// ============================================

// ToolBuilder provides a fluent API for creating tools
type ToolBuilder struct {
	tool *Tool
}

// ToolHandler is the function signature for tool handlers
type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// NewTool starts building a tool with name and description
func NewTool(name, description string) *ToolBuilder {
	return &ToolBuilder{
		tool: &Tool{
			Name:        name,
			Description: description,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
	}
}

// NoParameters indicates this tool takes no parameters
func (tb *ToolBuilder) NoParameters() *ToolBuilder {
	tb.tool.Parameters = map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
	return tb
}

// WithParameter adds a parameter to the tool's schema
func (tb *ToolBuilder) WithParameter(name, paramType, description string, required bool) *ToolBuilder {
	props := tb.tool.Parameters["properties"].(map[string]interface{})
	props[name] = map[string]interface{}{
		"type":        paramType,
		"description": description,
	}

	// Add to required list if needed
	if required {
		requiredList := tb.tool.Parameters["required"].([]string)
		tb.tool.Parameters["required"] = append(requiredList, name)
	}

	return tb
}

// WithParameters sets all parameters at once
func (tb *ToolBuilder) WithParameters(params map[string]ParameterDef) *ToolBuilder {
	props := make(map[string]interface{})
	required := make([]string, 0)

	for name, def := range params {
		props[name] = map[string]interface{}{
			"type":        def.Type,
			"description": def.Description,
		}
		if def.Required {
			required = append(required, name)
		}
	}

	tb.tool.Parameters["properties"] = props
	tb.tool.Parameters["required"] = required

	return tb
}

// ParameterDef describes a tool parameter
type ParameterDef struct {
	Type        string // "string", "number", "integer", "boolean", etc.
	Description string
	Required    bool
}

// Handler sets the handler function for this tool
func (tb *ToolBuilder) Handler(handler ToolHandler) *ToolBuilder {
	tb.tool.Handler = handler
	return tb
}

// Build returns the configured tool and validates it
func (tb *ToolBuilder) Build() *Tool {
	// Validate required fields
	if tb.tool.Name == "" {
		panic("Tool Name is required")
	}
	if tb.tool.Description == "" {
		panic("Tool Description is required")
	}
	if tb.tool.Handler == nil {
		panic("Tool Handler is required")
	}

	return tb.tool
}

// ============================================
// TeamBuilder for fluent team configuration
// ============================================

// TeamBuilder provides a fluent API for creating teams
type TeamBuilder struct {
	team *Team
}

// NewTeam starts building a team
func NewTeam() *TeamBuilder {
	return &TeamBuilder{
		team: &Team{
			Agents:      []*Agent{},
			MaxRounds:   10,
			MaxHandoffs: 3,
		},
	}
}

// WithName sets the team name (for identification)
func (tb *TeamBuilder) WithName(name string) *TeamBuilder {
	// Store in first agent's Name for now (teams don't have name field)
	// This is a placeholder; in Phase 2 we'll add proper team metadata
	return tb
}

// AddAgent adds an agent to the team
func (tb *TeamBuilder) AddAgent(agent *Agent) *TeamBuilder {
	if agent != nil {
		tb.team.Agents = append(tb.team.Agents, agent)
	}
	return tb
}

// AddAgents adds multiple agents to the team
func (tb *TeamBuilder) AddAgents(agents ...*Agent) *TeamBuilder {
	for _, agent := range agents {
		if agent != nil {
			tb.team.Agents = append(tb.team.Agents, agent)
		}
	}
	return tb
}

// WithMaxRounds sets maximum rounds for team execution
func (tb *TeamBuilder) WithMaxRounds(rounds int) *TeamBuilder {
	if rounds > 0 {
		tb.team.MaxRounds = rounds
	}
	return tb
}

// WithMaxHandoffs sets maximum handoffs for team execution
func (tb *TeamBuilder) WithMaxHandoffs(handoffs int) *TeamBuilder {
	if handoffs >= 0 {
		tb.team.MaxHandoffs = handoffs
	}
	return tb
}

// Build returns the configured team and validates it
func (tb *TeamBuilder) Build() *Team {
	// Validate required fields
	if len(tb.team.Agents) == 0 {
		panic("Team must have at least one agent")
	}

	// Verify at least one agent is terminal
	hasTerminal := false
	for _, agent := range tb.team.Agents {
		if agent.IsTerminal {
			hasTerminal = true
			break
		}
	}
	if !hasTerminal {
		panic("Team must have at least one terminal agent")
	}

	return tb.team
}
