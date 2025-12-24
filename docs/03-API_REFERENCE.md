# API Reference

## Types

### Agent

```go
type Agent struct {
	ID          string       // Unique agent identifier
	Name        string       // Display name
	Role        string       // Agent's role/expertise
	Backstory   string       // Background and context
	Model       string       // LLM model to use (e.g., "gpt-4o-mini")
	Provider    string       // LLM provider: "openai" or "ollama"
	ProviderURL string       // Provider URL (for Ollama: "http://localhost:11434")
	Tools       []*Tool      // Available tools
	Temperature float64      // LLM temperature (0.0-1.0, default: 0.7)
	IsTerminal  bool         // True if this is a final agent (no handoffs)
}
```

### Crew

```go
type Crew struct {
	Agents          []*Agent        // List of agents
	Tasks           []*Task         // Tasks to execute
	MaxRounds       int             // Max tool execution iterations (default: 10)
	MaxHandoffs     int             // Max agent handoffs (default: 5)
	Config          *CrewConfig     // Crew configuration
}
```

### Tool

```go
type Tool struct {
	Name        string                                                              // Tool name
	Description string                                                              // What it does
	Parameters  map[string]interface{}                                              // JSON schema
	Handler     func(ctx context.Context, args map[string]interface{}) (string, error) // Implementation
}
```

### Response

```go
type Response struct {
	ID        string      // Response ID
	AgentName string      // Which agent produced this
	AgentID   string      // Agent ID
	Content   string      // Response content
	Signal    string      // Routing signal (if any)
	Metadata  Metadata    // Execution metadata
}
```

## CrewExecutor

### NewCrewExecutor

```go
func NewCrewExecutor(crew *Crew, apiKey string) *CrewExecutor
```

Creates a new executor with an API key for the configured provider.

### Execute

```go
func (ce *CrewExecutor) Execute(ctx context.Context, task string) (*Response, error)
```

Executes the crew with a given task. Returns final response from terminal agent.

### ExecuteStream

```go
func (ce *CrewExecutor) ExecuteStream(ctx context.Context, task string, streamChan chan<- *Response) error
```

Executes the crew and streams responses through the channel. Call `close(streamChan)` when done.

## Configuration

### YAML Format

Agent configuration (agents/my-agent.yaml):
```yaml
id: my-agent
name: My Agent
role: Specialist in area X
backstory: Background and expertise
model: gpt-4o-mini
provider: openai  # or "ollama"
provider_url: http://localhost:11434  # Optional, for Ollama
temperature: 0.7
tools:
  - check_status
```

Crew configuration (crew.yaml):
```yaml
agents:
  - my-agent
  - other-agent

tasks:
  - id: main-task
    description: "Task description"

max_rounds: 10
max_handoffs: 5

routing:
  signals:
    my-agent:
      - signal: "[ROUTE_OTHER]"
        target: other-agent
```

## Common Patterns

### Creating Agents Programmatically

```go
agent := &crewai.Agent{
	ID:        "specialist",
	Name:      "Specialist",
	Role:      "Domain Expert",
	Backstory: "Expert in the field",
	Model:     "gpt-4o-mini",
	Tools:     tools,
}
```

### Creating Tools

```go
tool := &crewai.Tool{
	Name:        "CheckStatus",
	Description: "Check the status of something",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target": map[string]interface{}{
				"type":        "string",
				"description": "What to check",
			},
		},
		"required": []string{"target"},
	},
	Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
		target := args["target"].(string)
		// Implementation
		return "Status OK", nil
	},
}
```

### Loading Configuration

```go
crew, err := crewai.LoadCrewConfig("path/to/crew.yaml")
if err != nil {
	panic(err)
}

executor := crewai.NewCrewExecutor(crew, apiKey)
result, err := executor.Execute(ctx, "user task")
```

### Streaming Results

```go
responseChan := make(chan *crewai.Response)
go func() {
	if err := executor.ExecuteStream(ctx, "task", responseChan); err != nil {
		log.Fatal(err)
	}
}()

for response := range responseChan {
	fmt.Printf("[%s]: %s\n", response.AgentName, response.Content)
}
```

## Error Handling

Common errors:

- `ErrMissingAPIKey`: API key not provided
- `ErrInvalidConfig`: Configuration is invalid
- `ErrAgentNotFound`: Referenced agent doesn't exist
- `ErrToolExecution`: Tool execution failed
- `ErrMaxHandoffs`: Exceeded maximum handoffs
- `ErrContextCancelled`: Context was cancelled

Always check errors:

```go
result, err := executor.Execute(ctx, "task")
if err != nil {
	// Handle specific error
	if errors.Is(err, crewai.ErrMaxHandoffs) {
		log.Println("Too many handoffs, stopping")
	}
}
```
