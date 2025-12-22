# Configuration YAML Schema Reference

## crew.yaml JSON Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Go-Agentic Crew Configuration",
  "type": "object",
  "required": ["version", "name", "description", "entry_point", "agents"],
  "properties": {
    "version": {
      "type": "string",
      "description": "Schema version",
      "enum": ["1.0"],
      "example": "1.0"
    },
    "name": {
      "type": "string",
      "description": "Unique crew identifier (lowercase, hyphenated)",
      "pattern": "^[a-z0-9-]{3,50}$",
      "example": "hello-crew"
    },
    "description": {
      "type": "string",
      "description": "Human-readable crew description",
      "minLength": 10,
      "maxLength": 500,
      "example": "A minimal crew for greeting users"
    },
    "entry_point": {
      "type": "string",
      "description": "Agent ID to start execution",
      "pattern": "^[a-z0-9-_]{3,50}$",
      "example": "hello-agent"
    },
    "agents": {
      "type": "array",
      "description": "List of agent IDs in this crew",
      "minItems": 1,
      "maxItems": 100,
      "items": {
        "type": "string",
        "pattern": "^[a-z0-9-_]{3,50}$"
      },
      "example": ["orchestrator", "clarifier", "executor"]
    },
    "settings": {
      "type": "object",
      "description": "Global crew settings",
      "properties": {
        "max_handoffs": {
          "type": "integer",
          "description": "Maximum handoff count between agents",
          "minimum": 1,
          "maximum": 100,
          "default": 5,
          "example": 10
        },
        "max_rounds": {
          "type": "integer",
          "description": "Maximum processing rounds per agent",
          "minimum": 1,
          "maximum": 100,
          "default": 10,
          "example": 20
        },
        "timeout_seconds": {
          "type": "integer",
          "description": "Total execution timeout in seconds",
          "minimum": 10,
          "maximum": 3600,
          "default": 300,
          "example": 600
        },
        "language": {
          "type": "string",
          "description": "Crew primary language",
          "enum": ["en", "vi"],
          "default": "en",
          "example": "vi"
        },
        "organization": {
          "type": "string",
          "description": "Organization name",
          "minLength": 1,
          "maxLength": 100,
          "example": "IT-Support-Team"
        }
      },
      "additionalProperties": false
    },
    "routing": {
      "type": "object",
      "description": "Signal-based routing configuration",
      "properties": {
        "signals": {
          "type": "object",
          "description": "Signal definitions per agent",
          "patternProperties": {
            "^[a-z0-9-_]{3,50}$": {
              "type": "array",
              "items": {
                "type": "object",
                "required": ["signal", "target", "description"],
                "properties": {
                  "signal": {
                    "type": "string",
                    "description": "Signal name (uppercase in brackets)",
                    "pattern": "^\\[[A-Z_]+\\]$",
                    "example": "[ROUTE_EXECUTOR]"
                  },
                  "target": {
                    "oneOf": [
                      {
                        "type": "string",
                        "description": "Target agent ID",
                        "pattern": "^[a-z0-9-_]{3,50}$"
                      },
                      {
                        "type": "null",
                        "description": "Terminal signal (no target)"
                      }
                    ],
                    "example": "executor"
                  },
                  "description": {
                    "type": "string",
                    "description": "Signal purpose",
                    "minLength": 5,
                    "maxLength": 200
                  }
                }
              }
            }
          }
        },
        "defaults": {
          "type": "object",
          "description": "Default routing when no signal detected",
          "patternProperties": {
            "^[a-z0-9-_]{3,50}$": {
              "oneOf": [
                {
                  "type": "string",
                  "pattern": "^[a-z0-9-_]{3,50}$"
                },
                {
                  "type": "null"
                }
              ]
            }
          }
        },
        "agent_behaviors": {
          "type": "object",
          "description": "Agent-specific behavior configuration",
          "patternProperties": {
            "^[a-z0-9-_]{3,50}$": {
              "type": "object",
              "properties": {
                "wait_for_signal": {
                  "type": "boolean",
                  "default": true
                },
                "auto_route": {
                  "type": "boolean",
                  "default": false
                },
                "is_terminal": {
                  "type": "boolean",
                  "default": false
                },
                "description": {
                  "type": "string"
                }
              }
            }
          }
        }
      },
      "additionalProperties": false
    }
  },
  "additionalProperties": false
}
```

---

## agent.yaml JSON Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Go-Agentic Agent Configuration",
  "type": "object",
  "required": [
    "id", "name", "role", "description", "backstory",
    "model", "temperature", "provider"
  ],
  "properties": {
    "id": {
      "type": "string",
      "description": "Unique agent identifier (lowercase, hyphenated)",
      "pattern": "^[a-z0-9-_]{3,50}$",
      "example": "orchestrator"
    },
    "name": {
      "type": "string",
      "description": "Display name (can be any language)",
      "minLength": 1,
      "maxLength": 100,
      "example": "My"
    },
    "role": {
      "type": "string",
      "description": "Agent's primary role",
      "minLength": 5,
      "maxLength": 100,
      "example": "Điều phối viên hệ thống"
    },
    "description": {
      "type": "string",
      "description": "What agent does",
      "minLength": 10,
      "maxLength": 300,
      "example": "Coordinates IT support requests and routes to appropriate specialists"
    },
    "backstory": {
      "type": "string",
      "description": "Agent background and context (supports template variables)",
      "minLength": 20,
      "maxLength": 2000,
      "example": "I am {{name}}, a system coordinator..."
    },
    "model": {
      "type": "string",
      "description": "LLM model name",
      "minLength": 3,
      "maxLength": 50,
      "examples": [
        "gemma3:1b",
        "deepseek-r1:1.5b",
        "gpt-4-turbo",
        "gpt-4o"
      ]
    },
    "temperature": {
      "type": "number",
      "description": "Model creativity (0.0=deterministic, 1.0=creative)",
      "minimum": 0.0,
      "maximum": 1.0,
      "example": 0.7
    },
    "provider": {
      "type": "string",
      "description": "LLM provider",
      "enum": ["ollama", "openai"],
      "example": "ollama"
    },
    "provider_url": {
      "type": "string",
      "description": "Provider endpoint (required for Ollama)",
      "format": "uri",
      "example": "http://localhost:11434"
    },
    "is_terminal": {
      "type": "boolean",
      "description": "Is this agent an execution endpoint?",
      "default": false,
      "example": true
    },
    "handoff_targets": {
      "type": "array",
      "description": "Agents this agent can route to",
      "default": [],
      "items": {
        "type": "string",
        "pattern": "^[a-z0-9-_]{3,50}$"
      },
      "example": ["clarifier", "executor"]
    },
    "tools": {
      "type": "array",
      "description": "Available tools for this agent",
      "default": [],
      "items": {
        "type": "string",
        "pattern": "^[A-Z][a-zA-Z0-9]*$"
      },
      "example": ["GetCPUUsage", "GetMemoryUsage", "PingHost"]
    },
    "system_prompt": {
      "type": "string",
      "description": "Custom system prompt (overrides auto-generated)",
      "minLength": 50,
      "maxLength": 5000,
      "example": "You are {{name}}. Role: {{role}}. Backstory: {{backstory}}"
    }
  },
  "additionalProperties": false,
  "allOf": [
    {
      "if": {
        "properties": {
          "provider": { "const": "ollama" }
        }
      },
      "then": {
        "required": ["provider_url"]
      }
    }
  ]
}
```

---

## Type Definitions

### String Formats

**agent_id / crew_name Pattern**: `^[a-z0-9-_]{3,50}$`
```
Examples:
  ✓ hello-agent
  ✓ data_analyzer
  ✓ orchestrator-v2
  ✗ HelloAgent       (uppercase not allowed)
  ✗ hello agent      (space not allowed)
  ✗ ab               (too short, min 3)
```

**signal Pattern**: `^\\[[A-Z_]+\\]$`
```
Examples:
  ✓ [ROUTE_EXECUTOR]
  ✓ [KẾT_THÚC]
  ✓ [COMPLETE]
  ✗ [route_executor]  (lowercase not allowed)
  ✗ ROUTE_EXECUTOR    (missing brackets)
  ✗ [Route_Executor]  (mixed case not allowed)
```

**tool_name Pattern**: `^[A-Z][a-zA-Z0-9]*$`
```
Examples:
  ✓ GetCPUUsage
  ✓ CheckDiskSpace
  ✓ PingHost
  ✗ getCPUUsage       (lowercase start not allowed)
  ✗ GetCPU_Usage      (underscores not allowed)
```

### Enumerations

**provider Enum**:
```yaml
- ollama    # Local LLM via Ollama
- openai    # OpenAI API
```

**language Enum**:
```yaml
- en        # English
- vi        # Vietnamese
```

**version Enum**:
```yaml
- "1.0"     # Current schema version
```

---

## Validation Rules

### crew.yaml Validation Rules

1. **entry_point Must Exist**
   ```
   entry_point value must be in agents array
   ```

2. **No Circular Agent Lists**
   ```
   agents array must not be empty
   All agent IDs must be unique within crew
   ```

3. **Settings Constraints**
   ```
   max_handoffs: 1 ≤ value ≤ 100
   max_rounds: 1 ≤ value ≤ 100
   timeout_seconds: 10 ≤ value ≤ 3600
   ```

4. **Routing Validation**
   ```
   All signal targets must be in agents array (or null)
   All default targets must be in agents array (or null)
   agent_behaviors keys must be in agents array
   ```

### agent.yaml Validation Rules

1. **ID Must Be Valid**
   ```
   id must match pattern: ^[a-z0-9-_]{3,50}$
   id must be unique within crew
   ```

2. **Provider Constraints**
   ```
   IF provider == "ollama" THEN provider_url is required
   IF provider == "openai" THEN provider_url is optional (uses API key)
   ```

3. **Temperature Range**
   ```
   temperature must be: 0.0 ≤ value ≤ 1.0
   ```

4. **Tool Names**
   ```
   All tools must follow pattern: ^[A-Z][a-zA-Z0-9]*$
   Tools must be registered in system
   ```

5. **Terminal Agent Rules**
   ```
   IF is_terminal == true THEN handoff_targets should be empty []
   IF is_terminal == false THEN handoff_targets can be non-empty
   ```

---

## Complete Valid Examples

### Minimal Valid crew.yaml

```yaml
version: "1.0"
name: minimal-crew
description: Minimal valid crew configuration
entry_point: agent1
agents:
  - agent1
```

### Minimal Valid agent.yaml

```yaml
id: agent1
name: Agent One
role: Basic Agent
description: A basic agent with minimal configuration
backstory: I am a simple agent.
model: gemma3:1b
temperature: 0.7
provider: ollama
provider_url: http://localhost:11434
```

### Complete Valid crew.yaml

```yaml
version: "1.0"
name: complete-crew
description: Complete crew configuration with all optional fields

entry_point: coordinator

agents:
  - coordinator
  - processor
  - finalizer

settings:
  max_handoffs: 10
  max_rounds: 15
  timeout_seconds: 600
  language: en
  organization: Example-Org

routing:
  signals:
    coordinator:
      - signal: "[PROCESS]"
        target: processor
        description: "Start processing"
    processor:
      - signal: "[FINALIZE]"
        target: finalizer
        description: "Ready for final step"
    finalizer:
      - signal: "[DONE]"
        target: null
        description: "Processing complete"

  defaults:
    coordinator: processor
    processor: finalizer
    finalizer: null

  agent_behaviors:
    coordinator:
      wait_for_signal: true
      auto_route: false
      description: "Coordinator waits for explicit signals"
    processor:
      wait_for_signal: true
      auto_route: false
    finalizer:
      is_terminal: true
      description: "Finalizer is the endpoint"
```

### Complete Valid agent.yaml

```yaml
id: data-processor
name: Data Processor
role: Data Processing Specialist
description: Processes and transforms data using available tools

backstory: |
  I am {{name}}, a data processing specialist with deep expertise
  in data transformation and quality assurance.
  My role is to {{role}}.

  I work methodically, ensuring data integrity at every step.
  I can handle various data formats and provide detailed
  quality reports.

model: gpt-4-turbo
temperature: 0.5
is_terminal: false

provider: openai

handoff_targets:
  - validator
  - analyzer

tools:
  - LoadData
  - TransformData
  - ValidateSchema
  - ComputeStatistics
  - GenerateReport

system_prompt: |
  You are {{name}}.
  Role: {{role}}
  Description: {{description}}

  Background: {{backstory}}

  Available tools: LoadData, TransformData, ValidateSchema,
  ComputeStatistics, GenerateReport

  When processing data:
  1. Load data using LoadData
  2. Transform if needed
  3. Validate schema
  4. Generate statistics
  5. Report status

  Always ensure data quality before handoff.
```

---

## Common Mistakes & Fixes

### Mistake 1: entry_point Not in agents List

❌ **Wrong**:
```yaml
entry_point: orchestrator
agents:
  - router        # NOT orchestrator!
  - executor
```

✓ **Correct**:
```yaml
entry_point: orchestrator
agents:
  - orchestrator
  - executor
```

### Mistake 2: Invalid Agent ID Pattern

❌ **Wrong**:
```yaml
id: HelloAgent          # Uppercase not allowed
id: hello agent         # Space not allowed
id: hello_agent_v2      # Too long, exceeds 50 chars
```

✓ **Correct**:
```yaml
id: hello-agent
id: hello_agent_v1
id: data-processor
```

### Mistake 3: Missing Required Provider URL

❌ **Wrong**:
```yaml
provider: ollama
# Missing provider_url!
```

✓ **Correct**:
```yaml
provider: ollama
provider_url: http://localhost:11434
```

### Mistake 4: Invalid Signal Format

❌ **Wrong**:
```yaml
signal: route_executor              # No brackets
signal: [route_executor]            # Lowercase not allowed
signal: [ROUTE_EXECUTOR             # Missing closing bracket
```

✓ **Correct**:
```yaml
signal: "[ROUTE_EXECUTOR]"
signal: "[DATA_READY]"
signal: "[COMPLETE]"
```

### Mistake 5: Terminal Agent with handoff_targets

❌ **Wrong**:
```yaml
is_terminal: true
handoff_targets:
  - other_agent       # Terminal agents shouldn't have targets
```

✓ **Correct**:
```yaml
is_terminal: true
handoff_targets: []   # Empty for terminal agents
```

---

## YAML Syntax Quick Tips

### Multiline Strings

```yaml
# Option 1: Keep newlines (|)
backstory: |
  Line 1
  Line 2
  Line 3

# Option 2: Preserve trailing newlines (|+)
system_prompt: |+
  Line 1
  Line 2

# Option 3: Remove trailing newlines (|-)
backstory: |-
  Line 1
  Line 2
```

### Arrays

```yaml
# Inline style
agents: [agent1, agent2, agent3]

# Block style (preferred)
agents:
  - agent1
  - agent2
  - agent3
```

### Objects

```yaml
# Short nested objects
settings: {max_handoffs: 5, timeout_seconds: 300}

# Long nested objects (preferred)
settings:
  max_handoffs: 5
  timeout_seconds: 300
```

### Comments

```yaml
# This is a comment
name: my-crew  # Inline comment

# Multi-line comment
# Line 1
# Line 2
```

---

## Tools Reference

All tool names follow PascalCase pattern:

**Basic System Tools**:
- GetCPUUsage
- GetMemoryUsage
- GetDiskSpace
- GetSystemInfo
- GetRunningProcesses

**Network Tools**:
- PingHost
- CheckServiceStatus
- ResolveDNS
- CheckNetworkStatus

**Advanced Tools**:
- CheckMemoryStatus
- CheckDiskStatus
- ExecuteCommand
- GetSystemDiagnostics
- TerminateProcess
- UninstallApplication

**Custom Tools**: Follow same PascalCase pattern

---

## References

- [Full Configuration Specification](CONFIG_SPECIFICATION.md)
- [Quick Reference Guide](CONFIG_QUICK_REFERENCE.md)
- [Core Library Documentation](./LIBRARY_USAGE.md)
