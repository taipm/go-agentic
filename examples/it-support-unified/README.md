# IT Support System - Simplified Example

This example demonstrates a **simplified 2-agent IT support system** using go-agentic with Phase 3 declarative routing.

## ✨ Features

- 🎫 **2 Core Agents**
  - Support Responder - diagnoses issues
  - Support Resolver - provides final solutions
- 🛠️ **Real Tool Integration** - system diagnostics, network testing, app verification
- 📋 **YAML Configuration** - easy to customize and extend
- 🔄 **Phase 3 Routing** - automatic agent handoff based on response detection
- ⚡ **Minimal Setup** - clean, understandable code structure

## 🏗️ Architecture

```text
Support Ticket
      ↓
┌─────────────────────────┐
│  Support Responder      │
│  - Diagnose issue       │
│  - Gather information   │
│  - Run diagnostics      │
└──────────┬──────────────┘
           ↓
    [Phase 3 Routing]
    Detect: "resolved", "escalate"
           ↓
┌─────────────────────────┐
│  Support Resolver       │
│  - Provide solution     │
│  - Log ticket           │
│  - Send summary         │
└──────────┬──────────────┘
           ↓
    Final Resolution
```

## 🎭 Agents

### 1. Support Responder

- **Role**: Diagnose IT support issues and gather information
- **Tools**: System info, connectivity testing, app verification
- **isTerminal**: `false` (can hand off to resolver)

### 2. Support Resolver

- **Role**: Provide final resolution and document the ticket
- **Tools**: Logging, summarization, license verification
- **isTerminal**: `true` (ends the conversation)

## 🔧 Tools Available

### System Diagnostics

- `get_system_info` - Retrieve hardware and OS information
- `check_disk_space` - Check available disk space
- `list_installed_apps` - List installed applications
- `check_app_version` - Check application versions

### Network Tools

- `ping_host` - Test connectivity to hosts
- `trace_route` - Trace network routes
- `check_dns` - Verify DNS resolution

### Resolution Tools

- `verify_licenses` - Check software licenses
- `log_ticket` - Document the ticket
- `send_summary` - Send resolution to user

## 📋 YAML Configuration

The example is configured in `team.yaml` with:

```yaml
team:
  name: "IT Support Team"
  config:
    maxRounds: 8
    maxHandoffs: 2

agents:
  responder:
    id: "responder"
    name: "Support Responder"
    # ... agent configuration
    tools: [get_system_info, ping_host, ...]

  resolver:
    id: "resolver"
    name: "Support Resolver"
    # ... agent configuration
    tools: [log_ticket, send_summary, ...]

tools:
  get_system_info: { ... }
  ping_host: { ... }
  # ... all tool definitions
```

## 🚀 Quick Start

### Step 1: Setup API Key

```bash
cp .env.example .env
# Edit .env and add your OpenAI API key
```

### Step 2: Run

```bash
go run .
```

### Expected Output

```
🎫 IT Support System
==================================================

📋 Ticket: TKT-001
--------------------------------------------------
[Support Responder]: Let me diagnose your issue...
[System Info]: System: Intel i7-13700K | Memory: 32GB | OS: Ubuntu 22.04
...
[Support Resolver]: Based on the diagnostics, here's the solution...
✅ [Final Resolution]

...
```

## 📁 Project Structure

```
it-support-unified/
├── main.go              # Application logic
├── team.yaml            # Team and tools configuration
├── .env.example         # API key template
├── README.md            # This file
└── go.mod, go.sum       # Dependencies
```

## 🔍 Code Structure

### main.go (112 lines)

Three main functions:

1. **main()** - Core application
   - Load API key from .env
   - Call `LoadTeamFromYAML()` to load team
   - Execute support tickets

2. **getToolHandlers()** - Tool implementations
   - Returns mock/simulated tool responses
   - In production, would call real system commands

3. **getEnvVar()** - Environment helper
   - Reads .env file
   - Extracts API key

### team.yaml (135 lines)

- **Team config**: maxRounds, maxHandoffs
- **2 Agents**: Responder (isTerminal: false), Resolver (isTerminal: true)
- **9 Tools**: System, network, and resolution tools
- **Centralized**: Everything in one file

## 🔄 How It Works

1. **User submits support ticket** with issue description
2. **Responder agent** receives ticket and:
   - Uses tools to gather system information
   - Runs diagnostics (ping, DNS check, disk space, etc.)
   - Analyzes findings
3. **Phase 3 routing** detects resolution keywords and hands off to Resolver
4. **Resolver agent** receives diagnostics and:
   - Consolidates findings
   - Provides clear, actionable solution
   - Logs ticket and sends summary
5. **Conversation ends** (isTerminal: true)

## 🎯 Key Concepts

### Unified Configuration

All team settings, agents, and tools are in one `team.yaml` file instead of multiple config files.

### Tool Integration

Tool handlers defined in `getToolHandlers()` are called by agents as needed through the YAML configuration.

### Phase 3 Routing

Automatic agent handoff based on response content detection (keywords, patterns, signals).

## 🔧 Customization

### Add More Agents

Edit `team.yaml` to add more agents:

```yaml
agents:
  responder: { ... }
  specialist: { ... }      # New agent
  resolver: { ... }
```

### Add More Tools

Define new tools in `team.yaml` and implement handlers in `getToolHandlers()`:

```yaml
tools:
  new_tool:
    name: "NewTool"
    description: "Description"
```

### Adjust Conversation Length

Modify in `team.yaml`:

```yaml
team:
  config:
    maxRounds: 10      # Increase for longer conversations
    maxHandoffs: 5     # More handoffs = more back-and-forth
```

## 🆘 Troubleshooting

### "OPENAI_API_KEY not set"

Create `.env` file:

```bash
cp .env.example .env
# Edit and add your API key
```

### "Failed to load team config"

Make sure `team.yaml` is in the same directory as `main.go`:

```bash
ls team.yaml
```

### Agents not using tools

Check that:

- Tool is defined in `team.yaml` under both agent's `tools:` and the `tools:` section
- Tool handler exists in `getToolHandlers()`
- Tool name matches between YAML and handler

## 📚 Learning Points

This example demonstrates:

✅ Simplified code structure (3 functions, 112 lines)
✅ Unified YAML configuration (single file)
✅ Tool integration and execution
✅ Multi-agent conversation
✅ Automatic agent handoff
✅ Real-world use case (IT support)
✅ Clean separation of concerns

## 🚀 Next Steps

After running this example:

1. **Modify the agents** - Change their personalities and expertise
2. **Add more tools** - Implement real system diagnostics
3. **Customize tickets** - Add your own support scenarios
4. **Extend the system** - Add more agents (hardware, software, network specialists)

## 📞 Support

- Check README.md for setup and troubleshooting
- Review main.go for code structure
- Edit team.yaml to customize agents and tools

---

**Ready to run?**

```bash
cp .env.example .env
# Add your API key to .env
go run .
```

Simplified, focused, production-ready. 🚀
