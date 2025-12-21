# ⚙️ Configuration Guide - go-agentic

**Status**: Production Ready
**Version**: 1.0
**Last Updated**: 2025-12-22

---

## 📋 Overview

go-agentic uses **YAML configuration files** to define system behavior. This approach provides:
- **No code changes** for different workflows
- **Easy versioning** and rollback
- **Non-technical** configuration management
- **Clear separation** between config and code

Configuration files are organized in:
```
config/
├── crew.yaml              # Crew definition and routing
└── agents/
    ├── orchestrator.yaml  # Orchestrator agent config
    ├── clarifier.yaml     # Clarifier agent config
    └── executor.yaml      # Executor agent config
```

---

## 🏗️ Core Configuration Files

### 1. crew.yaml (Crew Definition)

Defines the complete crew structure, agent sequence, and routing rules.

#### Full Example with Comments

```yaml
# ===== Crew Configuration =====
# Defines agents, their sequence, and routing rules

# Entry point: Which agent handles new requests?
entry_point: orchestrator

# All agents in this crew (loaded from agents/ directory)
agents:
  - orchestrator
  - clarifier
  - executor

# Crew-level settings
max_handoffs: 5          # Maximum agent handoffs before stopping
                         # (prevents infinite loops)

max_rounds: 10           # Maximum tool execution rounds per agent
                         # (prevents agent from getting stuck)

collect_metrics: true    # Enable metrics collection
                         # (for monitoring and observability)

# ===== Routing Configuration =====
# Defines how agents communicate and hand off work

routing:
  # Signal-based routing: Map signals to target agents
  # Agents emit signals (e.g., "[ROUTE_EXECUTOR]") in their responses
  # Framework routes to agent defined in routing config

  signals:
    # Orchestrator agent routing
    orchestrator:
      - signal: "[ROUTE_CLARIFIER]"
        target: clarifier
        description: "Route to clarifier for more information"

      - signal: "[ROUTE_EXECUTOR]"
        target: executor
        description: "Route to executor for task completion"

      - signal: "[DONE]"
        target: null
        description: "Orchestrator completed, no further routing"

    # Clarifier agent routing
    clarifier:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
        description: "Information gathered, route to executor"

      - signal: "[REQUEST_CLARIFICATION]"
        target: clarifier
        description: "Continue gathering info from user"

      - signal: "[DONE]"
        target: null
        description: "Clarification complete"

    # Executor agent routing (typically terminal)
    executor:
      - signal: "[DONE]"
        target: null
        description: "Task complete, no further routing"

  # Default behavior if no signal matched
  # Options: "same" (retry agent), "next" (next agent), "done" (finish)
  default_behavior: "done"

# ===== Agent Behavior Configuration =====
# Agent-specific behavior and preferences

agent_behaviors:
  orchestrator:
    # Temperature: 0 = deterministic, 1 = creative
    # 0.7 is good for routing (balance of structure and creativity)
    temperature: 0.7

    # Model to use for this agent
    model: "gpt-4o"

    # Max tokens for this agent's response
    max_tokens: 2000

  clarifier:
    temperature: 0.5  # More focused on getting info
    model: "gpt-4o"
    max_tokens: 1500

  executor:
    temperature: 0.3  # More deterministic for tasks
    model: "gpt-4o"
    max_tokens: 2000

# ===== Tool Timeout Configuration =====
# When tools take too long, we stop them

timeouts:
  # Default timeout for all tools (seconds)
  default_tool_timeout: 5

  # Timeout for entire tool sequence in one round (seconds)
  sequence_timeout: 30

  # Override timeout for specific tools
  per_tool_timeout:
    "GetCPUUsage": 3         # Quick system call
    "PingHost": 10           # Network might be slow
    "ResolveDNS": 5          # DNS timeouts
    "CheckServiceStatus": 5

# ===== Logging Configuration =====
logging:
  # Log level: debug, info, warn, error
  level: "info"

  # Include timestamps in logs?
  include_timestamp: true

  # Log agent thinking process?
  log_agent_thinking: true

  # Log tool execution details?
  log_tool_execution: true

# ===== Advanced Configuration =====

# Request validation
validation:
  # Maximum input length (characters)
  max_input_length: 10000

  # Allowed models
  allowed_models:
    - "gpt-4o"
    - "gpt-4-turbo"

# Performance settings
performance:
  # Maximum concurrent requests
  max_concurrent_requests: 100

  # Request queue size
  queue_size: 200

  # Graceful shutdown timeout (seconds)
  graceful_shutdown_timeout: 30

# Streaming settings
streaming:
  # Buffer size for SSE streaming (bytes)
  buffer_size: 8192

  # Flush interval (milliseconds)
  flush_interval: 100
```

#### Key Concepts Explained

**entry_point**: Which agent receives the initial user request?
```yaml
entry_point: orchestrator  # User query goes to Orchestrator first
```

**agents**: List of all agents in the crew
```yaml
agents:
  - orchestrator  # Loads from config/agents/orchestrator.yaml
  - clarifier
  - executor
```

**max_handoffs**: Safety limit on agent-to-agent routing
```yaml
max_handoffs: 5  # Prevents infinite loops
# If agents route to each other > 5 times, stop and return error
```

**Signals**: How agents communicate routing decisions
```yaml
signals:
  orchestrator:
    - signal: "[ROUTE_EXECUTOR]"  # Agent emits this
      target: executor             # Framework routes here
```

---

### 2. Agent Configuration Files

Located in `config/agents/` directory. Each agent has its own YAML file.

#### agents/orchestrator.yaml

```yaml
# ===== Agent Definition =====
id: orchestrator
name: "Intelligent Request Router"

# What does this agent do?
role: |
  You are an intelligent request router that analyzes incoming requests
  and directs them to the appropriate specialist agents.

  Your responsibilities:
  1. Analyze the user's request
  2. Determine if clarification is needed
  3. Route to the right expert agent

  If you need more information to route correctly, emit: [ROUTE_CLARIFIER]
  If you can route to an executor, emit: [ROUTE_EXECUTOR]
  When done, emit: [DONE]

# Background and expertise context
backstory: |
  You are the entry point for the support system. You've handled thousands
  of requests and can quickly identify the nature of any problem.
  You have deep knowledge of all available specialist agents.

# LLM Model configuration
model: "gpt-4o"           # Which model to use
temperature: 0.7          # Balance between consistent and creative
max_tokens: 2000          # Maximum response length

# Which tools this agent can use
tools:
  - "GetSystemInfo"       # Tool names from config/tools/ or built-in

# Is this a terminal agent? (Can finish the workflow)
is_terminal: false        # Orchestrator typically routes, not executes

# System prompt customization
system_prompt_suffix: |
  Always be concise. Route decisively.
  The user expects a quick initial response.
```

#### agents/clarifier.yaml

```yaml
id: clarifier
name: "Information Gatherer"

role: |
  You are a clarification specialist. You ask focused questions to gather
  the information needed for the executor to complete the task.

  If you have enough information, emit: [ROUTE_EXECUTOR]
  If you need more information, emit: [REQUEST_CLARIFICATION]
  When done, emit: [DONE]

backstory: |
  You are skilled at asking targeted questions. You understand the domain
  and know exactly what information is needed to solve problems efficiently.

model: "gpt-4o"
temperature: 0.5          # More focused on information gathering
max_tokens: 1500
tools: []                 # Clarifier typically doesn't execute tools
is_terminal: false
```

#### agents/executor.yaml

```yaml
id: executor
name: "Task Executor"

role: |
  You are the executor agent. You use available tools to investigate
  and solve problems.

  When complete, emit: [DONE]

backstory: |
  You are a technical expert with access to system diagnostic tools.
  You solve problems methodically, gathering data and analyzing it.

model: "gpt-4o"
temperature: 0.3          # Deterministic, task-focused
max_tokens: 2000
tools:
  - "GetCPUUsage"
  - "GetMemoryUsage"
  - "GetDiskSpace"
  - "GetSystemInfo"
  - "GetRunningProcesses"
  - "PingHost"
  - "CheckServiceStatus"
  - "ResolveDNS"
is_terminal: true         # Executor is the last agent
```

---

## 🛠️ Advanced Configuration Topics

### Tool-Specific Timeout

Different tools may need different timeouts:

```yaml
timeouts:
  per_tool_timeout:
    "GetCPUUsage": 3         # Fast local call
    "PingHost": 10           # Might be slow network
    "CheckServiceStatus": 5  # Depends on service
    "ResolveDNS": 8          # DNS can be variable
```

### Custom System Prompts

Enhance agent behavior with suffix prompts:

```yaml
# In agent configuration:
system_prompt_suffix: |
  Instructions specific to this deployment:
  - Always validate inputs before using them
  - Log all tool executions for audit purposes
  - Prefer safe operations over fast ones
```

### Routing Strategies

#### Strategy 1: Explicit Signal Routing (Recommended)

Agents explicitly emit signals for routing:

```yaml
orchestrator:
  - signal: "[ROUTE_EXECUTOR]"
    target: executor
```

**Pros**: Deterministic, auditable, agent-agnostic
**Cons**: Requires agent coordination

#### Strategy 2: Default Behavior

What happens if agent doesn't emit a signal?

```yaml
routing:
  default_behavior: "done"   # Options: "done", "same", "next"

# "done" = finish workflow
# "same" = retry same agent
# "next" = go to next agent in sequence
```

---

## 📊 Configuration Validation

### Required Fields

Every configuration file must have:

```yaml
# crew.yaml
entry_point: (required)  # Starting agent
agents: (required)       # Agent list

# agents/*.yaml
id: (required)           # Unique identifier
name: (required)         # Display name
role: (required)         # Agent's role
model: (required)        # LLM model
```

### Validation Rules

1. **Agent IDs must be unique** across all agents
2. **Signal targets must exist** in agent list
3. **entry_point must exist** in agents list
4. **Model names must be valid** (gpt-4o, gpt-4-turbo, etc.)
5. **Tool names must be available** (registered in system)
6. **Timeout values must be positive**

### Pre-flight Validation

The system validates configuration at startup:

```bash
$ go-agentic --validate config/
✅ Validating configuration...
✅ crew.yaml: Valid
✅ agents/orchestrator.yaml: Valid
✅ agents/clarifier.yaml: Valid
✅ agents/executor.yaml: Valid
✅ All routing signals valid
✅ All tools available
✅ Configuration loaded successfully
```

---

## 🔄 Common Configuration Patterns

### Pattern 1: Simple Linear Workflow

```
User → Orchestrator → Executor → Response
```

**crew.yaml**:
```yaml
entry_point: orchestrator
agents:
  - orchestrator
  - executor

routing:
  signals:
    orchestrator:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
      - signal: "[DONE]"
        target: null
```

### Pattern 2: With Clarification

```
User → Orchestrator → Clarifier → Executor → Response
```

**crew.yaml**:
```yaml
entry_point: orchestrator
agents:
  - orchestrator
  - clarifier
  - executor

routing:
  signals:
    orchestrator:
      - signal: "[ROUTE_CLARIFIER]"
        target: clarifier
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
    clarifier:
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
```

### Pattern 3: Multi-Stage Analysis

```
User → Router → Analyzer → Researcher → Reporter → Response
```

**crew.yaml**:
```yaml
entry_point: router
agents:
  - router
  - analyzer
  - researcher
  - reporter

routing:
  signals:
    router:
      - signal: "[ANALYZE]"
        target: analyzer
    analyzer:
      - signal: "[RESEARCH]"
        target: researcher
    researcher:
      - signal: "[REPORT]"
        target: reporter
```

---

## 🔐 Security Configuration

### API Key Management

Store API keys in environment variables, not configuration:

```bash
# .env file (not checked into git)
export OPENAI_API_KEY="sk-..."
```

### Configuration File Permissions

```bash
# Restrict config file access
chmod 600 config/crew.yaml
chmod 600 config/agents/*.yaml
```

### Sensitive Information

Never store in configuration files:
- ❌ API keys
- ❌ Database passwords
- ❌ Authentication tokens
- ❌ Private URLs

Use environment variables instead:
```bash
export DB_PASSWORD="secret"
```

---

## 📈 Performance Tuning Configuration

### For High-Concurrency Scenarios

```yaml
performance:
  max_concurrent_requests: 200  # Increase from default 100
  queue_size: 500               # Larger queue

timeouts:
  sequence_timeout: 60          # More time per round
  default_tool_timeout: 10      # More generous per-tool
```

### For Low-Latency Scenarios

```yaml
performance:
  max_concurrent_requests: 50   # Less overhead
  queue_size: 100               # Smaller queue

timeouts:
  sequence_timeout: 15          # Stricter limits
  default_tool_timeout: 3       # Fast fail

agent_behaviors:
  # Every agent:
  temperature: 0.1              # More deterministic
  max_tokens: 500               # Shorter responses
```

### For Cost Optimization

```yaml
agent_behaviors:
  orchestrator:
    max_tokens: 500             # Use fewer tokens
    model: "gpt-4-turbo"        # Cheaper model

  executor:
    max_tokens: 1000            # Only executor needs detail
    model: "gpt-4o"             # Better for tasks
```

---

## 🚀 Deployment Configuration Profiles

### Development Profile

```yaml
# config/crew.dev.yaml
logging:
  level: "debug"
  log_agent_thinking: true

performance:
  max_concurrent_requests: 10

timeouts:
  default_tool_timeout: 30
```

### Staging Profile

```yaml
# config/crew.staging.yaml
logging:
  level: "info"

performance:
  max_concurrent_requests: 50

timeouts:
  default_tool_timeout: 10
```

### Production Profile

```yaml
# config/crew.prod.yaml
logging:
  level: "warn"  # Less verbose

performance:
  max_concurrent_requests: 200  # Handle more load

timeouts:
  default_tool_timeout: 5       # Stricter
```

Load with environment variable:
```bash
export GO_AGENTIC_CONFIG="prod"
go-agentic --config-profile $GO_AGENTIC_CONFIG
```

---

## ✅ Configuration Checklist

Before deploying, verify:

- [ ] All agent IDs are unique
- [ ] entry_point agent exists
- [ ] All routing targets exist
- [ ] Tool names are correct
- [ ] Timeouts are reasonable
- [ ] API keys set via environment
- [ ] No hardcoded secrets in files
- [ ] File permissions are restrictive (600)
- [ ] Configuration validates without errors
- [ ] Performance settings match deployment

---

## 🎓 Best Practices

1. **Version control configuration** (without secrets)
2. **Use profiles** for different environments
3. **Document custom signals** in code comments
4. **Keep agent backstories detailed** (helps LLM)
5. **Test routing logic** before production
6. **Monitor metrics** to tune timeouts
7. **Use explicit signals** instead of parsing text
8. **Keep tool list minimal** (confuses agents)
9. **Validate configuration** at startup
10. **Reload configuration** without restart (optional)

---

## 🔗 Related Documentation

- [Architecture Overview](ARCHITECTURE_OVERVIEW.md) - System design
- [API Reference](API_REFERENCE.md) - HTTP endpoints
- [Troubleshooting Guide](TROUBLESHOOTING_GUIDE.md) - Common issues
- [Deployment Guide](DEPLOYMENT_GUIDE.md) - Production setup

---

**Version**: 1.0
**Last Updated**: 2025-12-22
**Status**: Production Ready ✅
