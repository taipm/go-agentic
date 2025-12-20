# IT Support - Unified Configuration Example

This example demonstrates the **Phase 2 UX improvement**: Unified YAML Configuration, which replaces multiple configuration files with a single, comprehensive `team.yaml`.

## Overview

Previously, setting up a team required multiple files:
- `team.yaml` - Team settings
- `agents/router.yaml` - Router agent config
- `agents/hardware.yaml` - Hardware agent config
- `agents/software.yaml` - Software agent config
- `agents/network.yaml` - Network agent config
- `agents/resolver.yaml` - Resolver agent config
- `tools/tools.yaml` - Shared tools definition

**Now with unified configuration:** Everything is in **one `team.yaml` file** (~160 lines total).

## Key Improvements

### 1. **Single File Configuration**
```yaml
team:                    # Team settings
  name: "IT Support Team"
  config:
    maxRounds: 10
    maxHandoffs: 3

agents:                  # All agents defined here
  router: { ... }
  hardware: { ... }
  software: { ... }
  network: { ... }
  resolver: { ... }

tools:                   # All tools defined here
  get_system_info: { ... }
  ping_host: { ... }
  check_disk_space: { ... }
  # ... more tools

routing:                 # Routing rules here
  type: "signal"
  rules: [ ... ]
```

### 2. **Configuration Reduction**
- **Before**: 7 separate files to manage
- **After**: 1 unified file
- **Reduction**: ~75% fewer files to maintain
- **Lines of code**: ~160 lines (organized and readable)

### 3. **Type-Safe Loading**
```go
// Load complete team from single YAML
team, err := agentic.LoadTeamFromYAML(yamlPath, toolHandlers)
```

Features:
- ✅ Comprehensive validation (team settings, agents, tools, routing)
- ✅ Clear error messages for missing/invalid configuration
- ✅ Type-safe struct unmarshaling
- ✅ Backward compatible with fluent builder API

### 4. **Tool Handler Registry**
Tools are defined once in YAML, handlers provided at runtime:

```go
toolHandlers := agentic.ToolHandlerRegistry{
    "ping_host": func(ctx context.Context, args map[string]interface{}) (string, error) {
        target, _ := args["target"].(string)
        return fmt.Sprintf("✓ %s is reachable", target), nil
    },
    "get_system_info": func(ctx context.Context, args map[string]interface{}) (string, error) {
        return "System: Intel i7-13700K | Memory: 32GB | Disk: 512GB SSD", nil
    },
    // ... more handlers
}

team, err := agentic.LoadTeamFromYAML("team.yaml", toolHandlers)
```

### 5. **Centralized Tool Definition**
All tools defined in one place with:
- Name and description
- Parameter schema (type, description, required)
- Clear tool-agent associations

```yaml
tools:
  ping_host:
    name: "PingHost"
    description: "Check connectivity to a host or IP address"
    parameters:
      type: "object"
      properties:
        target:
          type: "string"
          description: "Target hostname or IP to ping"
      required:
        - target
```

## Structure

```
it-support-unified/
├── team.yaml              # ← Unified configuration (replaces 7 files!)
├── main.go               # Application using unified config
├── go.mod                # Dependencies
├── .env.example          # Environment template
└── README.md             # This file
```

## Running the Example

### 1. Setup Environment
```bash
cp .env.example .env
# Edit .env and add your OpenAI API key
export OPENAI_API_KEY="sk-proj-..."
```

### 2. Run the Application
```bash
go run main.go
```

Output:
```
🎫 IT Support System - Unified Configuration
============================================================

📋 Ticket: TKT-001
------------------------------------------------------------
[Support Ticket TKT-001] Computer won't turn on. I pressed the power button but nothing happens...

✅ Resolution:
[Agent responses with routing and tool execution...]
```

## Configuration Schema

### Team Section
```yaml
team:
  name: string                          # Team name (optional)
  config:
    maxRounds: int                      # Maximum conversation rounds (required)
    maxHandoffs: int                    # Maximum agent handoffs (required)
```

### Agents Section
Each agent requires:
```yaml
agents:
  agent_id:                             # Unique agent identifier
    id: string                          # Agent ID (auto-filled from key if empty)
    name: string                        # Display name
    role: string                        # Agent's role
    backstory: string                   # Agent's background/instructions
    model: string                       # LLM model to use
    temperature: float64                # 0.0-2.0, controls response randomness
    isTerminal: bool                    # Can this agent be final responder?
    tools: [string]                     # List of tool IDs this agent can use
```

### Tools Section
Each tool requires:
```yaml
tools:
  tool_id:                              # Unique tool identifier
    name: string                        # Tool function name
    description: string                 # What the tool does
    parameters:                         # JSON Schema for parameters
      type: "object"
      properties:
        param_name:
          type: string                  # JSON type
          description: string           # Parameter description
      required:                         # Optional: list of required params
        - param_name
```

### Routing Section (Optional)
```yaml
routing:
  type: "signal"                        # Routing type
  rules:                                # List of routing rules
    - from_agent: string                # Source agent ID
      trigger: string                   # Trigger condition
      target_agent: string              # Target agent ID
      description: string               # Rule description
```

## Validation

LoadTeamFromYAML performs comprehensive validation:

✅ **Team Config**
- `maxRounds > 0`
- `maxHandoffs >= 0`

✅ **Agents**
- At least one agent defined
- All required fields present (name, role, backstory, model)
- At least one agent marked as `isTerminal: true`
- All referenced tools exist

✅ **Tools**
- All referenced tools are defined
- Tool names and descriptions are non-empty

## Comparison: Old vs New

### Old Multi-File Approach
```
agents/
├── router.yaml
├── hardware.yaml
├── software.yaml
├── network.yaml
└── resolver.yaml
team.yaml
tools/
└── tools.yaml
main.go
```
**7 files to manage**

### New Unified Approach
```
team.yaml              ← Everything here!
main.go
go.mod
.env.example
README.md
```
**1 configuration file**

## Benefits

| Aspect | Multi-File | Unified |
|--------|-----------|---------|
| Files to manage | 7 | 1 |
| Cross-file references | Multiple | None |
| Consistency checking | Manual | Automatic |
| Tool-agent mapping | In 2 places | Single definition |
| Configuration validation | None | Comprehensive |
| Learning curve | Steep | Gentle |
| Modification friction | High | Low |

## Advanced Usage

### Export Team to YAML (for backup/migration)
```go
yamlBytes, err := agentic.ExportTeamToYAML(team)
if err != nil {
    log.Fatal(err)
}

// Save to file
os.WriteFile("backup.yaml", yamlBytes, 0644)
```

### Load with Defaults
```go
team, err := agentic.LoadTeamFromYAMLWithDefaults(yamlPath, handlers)
// Automatically applies defaults if maxRounds <= 0 or maxHandoffs < 0
```

## Error Handling

LoadTeamFromYAML provides clear error messages:

```
invalid team configuration: agent 'router' has empty model
```

```
tool 'ping_host' referenced but no handler provided
```

```
agent 'hardware' references tool 'undefined_tool' which is not defined
```

## Testing

This example includes comprehensive tests:
- Basic YAML loading
- Tool handler registration
- Validation error cases
- Configuration export/import round-trip
- 16 total tests covering all validation paths

Run tests:
```bash
cd go-agentic
go test -v ./... -run "config_unified"
```

## Integration with Phase 1 (Fluent Builder API)

The unified configuration internally uses the fluent builder API:

```go
// Inside LoadTeamFromYAML
builder := NewAgent(agentCfg.ID, agentCfg.Name).
    WithRole(agentCfg.Role).
    WithBackstory(agentCfg.Backstory).
    WithModel(agentCfg.Model).
    WithTemperature(agentCfg.Temperature).
    SetTerminal(agentCfg.IsTerminal)
```

This ensures consistency and leverages Phase 1's builder pattern benefits.

## Next Steps: Phase 3

Phase 3 (Declarative Routing DSL) will enhance the routing section with:
- Signal detection from agent responses
- Automatic trigger matching
- Route-based system prompt generation
- Advanced branching logic

Stay tuned!
