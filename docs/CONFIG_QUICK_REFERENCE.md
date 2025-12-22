# Configuration Quick Reference Guide

## 📋 File Structure

```
config/
├── crew.yaml              # Crew configuration (1 file)
└── agents/
    ├── agent-1.yaml       # Agent configuration (multiple files)
    ├── agent-2.yaml
    └── agent-3.yaml
```

---

## 🔧 crew.yaml Minimal Template

```yaml
version: "1.0"
name: my-crew
description: Brief description of what this crew does

entry_point: first-agent

agents:
  - first-agent
  - second-agent

settings:
  max_handoffs: 5
  max_rounds: 10
  timeout_seconds: 300
  language: en
```

---

## 🤖 agent.yaml Minimal Template

```yaml
id: unique-agent-id
name: Display Name
role: Agent Role
description: What this agent does

backstory: |
  Context and background information about the agent.
  Can be multiple lines.

model: gemma3:1b           # For Ollama
temperature: 0.7
is_terminal: true

provider: ollama
provider_url: http://localhost:11434

tools: []
```

---

## 📊 Quick Decision Trees

### Which Provider?

```
Do you have OPENAI_API_KEY set?
├─ YES → provider: openai, model: gpt-4-turbo
└─ NO → provider: ollama, model: gemma3:1b or deepseek-r1:1.5b
```

### How Many Agents?

```
One agent only?
├─ YES → is_terminal: true, handoff_targets: [], max_handoffs: 1
└─ NO  → Plan routing with signals and defaults
         - First agent: is_terminal: false
         - Middle agents: is_terminal: false
         - Last agent: is_terminal: true
```

### Temperature Setting?

```
Task type?
├─ Logic/Accuracy (coding, math, classification)
│  └─ temperature: 0.3-0.5
├─ Balanced (general Q&A, analysis)
│  └─ temperature: 0.5-0.7 ✓ (RECOMMENDED)
└─ Creative (brainstorming, writing)
   └─ temperature: 0.8-1.0
```

---

## 🎯 Common Patterns

### Single-Agent Crew (Learning/Simple Tasks)

**crew.yaml:**
```yaml
version: "1.0"
name: hello-crew
entry_point: hello-agent
agents:
  - hello-agent
```

**hello-agent.yaml:**
```yaml
id: hello-agent
name: Hello Agent
role: Helpful Assistant
description: Simple helpful assistant
backstory: I help users with their questions.
model: gemma3:1b
temperature: 0.7
is_terminal: true
provider: ollama
provider_url: http://localhost:11434
tools: []
```

### Linear Multi-Agent (Sequential Processing)

**crew.yaml:**
```yaml
version: "1.0"
name: processing-pipeline
entry_point: input-processor
agents:
  - input-processor
  - main-processor
  - output-generator

routing:
  defaults:
    input-processor: main-processor
    main-processor: output-generator
    output-generator: null
```

**Agents:**
- `input-processor`: is_terminal: false, handoff_targets: [main-processor]
- `main-processor`: is_terminal: false, handoff_targets: [output-generator]
- `output-generator`: is_terminal: true, handoff_targets: []

### Branching Multi-Agent (Conditional Routing)

**crew.yaml:**
```yaml
version: "1.0"
name: diagnostic-team
entry_point: router

agents:
  - router
  - simple-handler
  - complex-handler

routing:
  signals:
    router:
      - signal: "[SIMPLE]"
        target: simple-handler
      - signal: "[COMPLEX]"
        target: complex-handler
```

**Agents:**
- `router`: is_terminal: false
- `simple-handler`: is_terminal: true
- `complex-handler`: is_terminal: true

---

## ✅ Validation Checklist

Before running your crew:

- [ ] `version: "1.0"` exists in crew.yaml
- [ ] `name` is lowercase and hyphenated
- [ ] `entry_point` matches an agent id
- [ ] All agents in `agents:` list have corresponding .yaml files
- [ ] Each agent has required fields: id, name, role, description
- [ ] `id` in agent.yaml matches filename (without .yaml)
- [ ] `model` and `provider` are set correctly
- [ ] `is_terminal: true` for last agent in workflow
- [ ] `handoff_targets` is not empty for non-terminal agents
- [ ] Provider URL is accessible (for Ollama)
- [ ] All signal targets point to valid agent ids

---

## 🔌 Provider Setup

### Ollama Setup

```bash
# Install Ollama (if not done)
# Download from: https://ollama.ai

# Start Ollama server
ollama serve

# In another terminal, pull a model
ollama pull gemma3:1b
ollama pull deepseek-r1:1.5b
```

**In agent.yaml:**
```yaml
provider: ollama
provider_url: http://localhost:11434
model: gemma3:1b              # or other available model
```

### OpenAI Setup

```bash
# Set environment variable
export OPENAI_API_KEY=sk-...
```

**In agent.yaml:**
```yaml
provider: openai
model: gpt-4-turbo            # or gpt-4o, gpt-4o-mini
```

---

## 📝 Field Reference Table

### crew.yaml Fields

| Field | Type | Required | Example |
|-------|------|----------|---------|
| version | string | ✓ | "1.0" |
| name | string | ✓ | "data-team" |
| description | string | ✓ | "Multi-agent data pipeline" |
| entry_point | string | ✓ | "coordinator" |
| agents | array | ✓ | ["agent1", "agent2"] |
| settings.max_handoffs | int | ✗ | 5 |
| settings.max_rounds | int | ✗ | 10 |
| settings.timeout_seconds | int | ✗ | 300 |
| settings.language | string | ✗ | "en" |
| settings.organization | string | ✗ | "TeamName" |

### agent.yaml Fields

| Field | Type | Required | Default | Example |
|-------|------|----------|---------|---------|
| id | string | ✓ | - | "analyzer" |
| name | string | ✓ | - | "Data Analyzer" |
| role | string | ✓ | - | "Statistical Analyst" |
| description | string | ✓ | - | "Analyzes data patterns" |
| backstory | string | ✓ | - | "I am skilled..." |
| model | string | ✓ | - | "gpt-4-turbo" |
| temperature | number | ✓ | - | 0.7 |
| provider | string | ✓ | - | "openai" |
| provider_url | string | ✓ (Ollama) | - | "http://localhost:11434" |
| is_terminal | boolean | ✗ | false | true |
| handoff_targets | array | ✗ | [] | ["agent2"] |
| tools | array | ✗ | [] | ["GetData", "Analyze"] |
| system_prompt | string | ✗ | auto | "You are..." |

---

## 🚀 Quick Start Examples

### Example 1: Hello World (30 seconds)

**crew.yaml:**
```yaml
version: "1.0"
name: hello-crew
description: Simple hello world
entry_point: greeter
agents:
  - greeter
```

**agents/greeter.yaml:**
```yaml
id: greeter
name: Greeter
role: Greeter
description: Says hello
backstory: I greet people warmly.
model: gemma3:1b
temperature: 0.7
is_terminal: true
provider: ollama
provider_url: http://localhost:11434
tools: []
```

**Run:** `go run cmd/main.go` → Type "Hello"

### Example 2: Three-Agent Team (IT Support Style)

**crew.yaml:**
```yaml
version: "1.0"
name: support-team
entry_point: router
agents:
  - router
  - info-gatherer
  - expert
settings:
  max_handoffs: 5
  timeout_seconds: 300

routing:
  defaults:
    router: info-gatherer
    info-gatherer: expert
    expert: null
```

**agents/router.yaml:**
```yaml
id: router
name: Support Router
role: Router
description: Routes support tickets
backstory: I route tickets efficiently.
model: gemma3:1b
temperature: 0.5
is_terminal: false
provider: ollama
provider_url: http://localhost:11434
tools: []
handoff_targets:
  - info-gatherer
  - expert
```

**agents/info-gatherer.yaml:** (similar, is_terminal: false)

**agents/expert.yaml:** (similar, is_terminal: true)

---

## 🎓 Learning Path

1. **Start**: Single-agent crew (`00-hello-crew`)
   - Understand basic structure
   - Learn agent fields
   - Test with Ollama

2. **Next**: Multi-agent linear workflow
   - Add second agent
   - Use handoff_targets
   - Test signal routing

3. **Advanced**: Branching workflow
   - Implement conditional logic
   - Use complex routing rules
   - Add multiple tools

4. **Expert**: Full IT Support system
   - Multiple signal types
   - Advanced routing logic
   - Tool integration
   - Real diagnostic tools

---

## 💡 Tips & Tricks

**Tip 1**: Use template variables in backstory and system_prompt
```yaml
backstory: |
  I am {{name}}, a {{role}}.
  {{description}}
```

**Tip 2**: Multiline strings should use `|` or `|-`
```yaml
backstory: |
  First line
  Second line
  Third line
```

**Tip 3**: Test agents individually before creating crew
```bash
# Create simple single-agent crew first
# Then add more agents one by one
```

**Tip 4**: Use comments in YAML to document decisions
```yaml
# Using deepseek-r1 because it's smaller but accurate
model: deepseek-r1:1.5b

# Lower temperature for deterministic classification
temperature: 0.3
```

**Tip 5**: Keep agent responsibilities focused
- One main job per agent
- Clear, descriptive names
- Detailed backstory for context

---

## 🔗 Related Documentation

- [Full Configuration Specification](CONFIG_SPECIFICATION.md)
- [IT Support Example](../examples/it-support/README.md)
- [Hello Crew Example](../examples/00-hello-crew/README.md)
- [Core Library Documentation](./LIBRARY_USAGE.md)

