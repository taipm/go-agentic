# Hello Crew: Minimal 1-Agent Example Design

**Created:** 2025-12-22
**Scope:** Simplest possible multi-agent crew implementation
**Goal:** User can learn complete system in 5 minutes
**Files:** 5 files total, ~200 lines code

---

## Vision

```
User opens README → Sees "Hello Crew" example
User thinks: "This is a crew? Just one agent?"
User runs it → "Wow, that's elegant. Now I understand."
User reads code → "3 files, ~50 lines. I can extend this."
User builds → Copies structure for their own crew
```

---

## Example Concept

### What It Does
```
User: "Tell me about yourself"
Agent: "I'm an AI assistant. I can help with coding questions..."
Done.
```

### What It Teaches
- How to create a simple agent
- How ExecuteStream works (streaming events)
- How to use web UI
- Pattern to build on for multi-agent crews

### Complexity Level: ⭐ (Simplest possible)

---

## File Structure

```
examples/00-hello-crew/
├── cmd/main.go                 # 40 lines: entry point
├── internal/hello.go           # 30 lines: hello agent definition
├── config/crew.yaml            # 10 lines: minimal config
├── config/agents/
│   └── hello-agent.yaml        # 15 lines: agent definition
├── .env.example                # 2 lines: just OPENAI_API_KEY
├── go.mod                       # Standard module
├── go.sum                       # Standard module
├── README.md                    # 40 lines: explanation + quick start
└── Makefile                     # 10 lines: make run, make build

Total: ~150 lines code, ~100 lines docs
```

---

## Implementation

### 1. cmd/main.go

```go
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/taipm/go-agentic/core"
)

func main() {
	// Parse flags
	serverMode := flag.Bool("server", false, "Run in server mode (Web UI)")
	port := flag.String("port", "8081", "Server port")
	flag.Parse()

	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY environment variable not set")
		os.Exit(1)
	}

	// Load crew from YAML config
	configDir := "config"
	helloAgent := core.CreateAgentFromConfig(
		loadAgentConfig(configDir+"/agents/hello-agent.yaml"),
		nil, // No tools needed
	)

	crew := &core.Crew{
		Agents:      []*core.Agent{helloAgent},
		MaxRounds:   5,
		MaxHandoffs: 0, // Single agent, no handoffs
	}

	executor := &core.CrewExecutor{
		Crew:           crew,
		APIKey:         apiKey,
		EntryAgent:     helloAgent,
		History:        []core.Message{},
		ToolTimeouts:   core.NewToolTimeoutConfig(),
		Metrics:        core.NewMetricsCollector(),
	}

	// Server mode: Web UI
	if *serverMode {
		runServer(executor, *port)
	} else {
		// CLI mode: Interactive chat
		runCLI(executor)
	}
}

func runCLI(executor *core.CrewExecutor) {
	fmt.Println("🤖 Hello Crew")
	fmt.Println("Type your message (or 'quit' to exit):")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "quit" {
			break
		}

		// Execute crew
		result, err := executor.Execute(context.Background(), input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\n%s\n\n", result.Content)
	}
}

func runServer(executor *core.CrewExecutor, port string) {
	fmt.Printf("🌐 Hello Crew Web UI: http://localhost:%s\n", port)
	
	handler := core.NewCrewHTTPHandler(executor)
	handler.Start(port)
}

func loadAgentConfig(path string) *core.AgentConfig {
	// TODO: Implement YAML loading
	// For now, return inline config
	return &core.AgentConfig{
		ID:          "hello-agent",
		Name:        "Assistant",
		Role:        "Helpful AI Assistant",
		Description: "Answers questions and has conversations",
		Backstory:   "You are a friendly AI assistant...",
		Model:       "gpt-4o-mini",
		Temperature: 0.7,
		Tools:       []string{},
		IsTerminal:  true,
	}
}
```

### 2. internal/hello.go

```go
package internal

import (
	"context"

	"github.com/taipm/go-agentic/core"
)

// GetHelloAgent returns a simple conversational agent
func GetHelloAgent() *core.Agent {
	return &core.Agent{
		ID:   "hello-agent",
		Name: "Assistant",
		Role: "Helpful AI Assistant",

		Backstory: `You are a friendly and helpful AI assistant. 
Your purpose is to have helpful conversations and answer questions.
Be concise, friendly, and clear in your responses.
You don't need tools for this - just conversation.`,

		Model:       "gpt-4o-mini",
		Temperature: 0.7,
		Tools:       []*core.Tool{}, // No tools needed
		IsTerminal:  true,           // This is the final agent

		SystemPrompt: `You are {{name}}, a {{role}}.

{{backstory}}

Instructions:
- Keep responses concise (2-3 sentences)
- Be friendly and conversational
- Answer the user's question directly
- If asked about yourself, be honest: you're an AI assistant`,
	}
}
```

### 3. config/crew.yaml

```yaml
version: "1.0"
description: "Hello Crew - Minimal single-agent example"

entry_point: hello-agent

agents:
  - hello-agent

settings:
  max_handoffs: 0      # No handoffs - single agent
  max_rounds: 5        # Max conversation turns
  timeout_seconds: 30  # Quick responses

# Minimal routing (not used for single agent)
routing:
  signals: {}
  defaults: {}
```

### 4. config/agents/hello-agent.yaml

```yaml
id: hello-agent
name: Assistant
role: Helpful AI Assistant
description: Answers questions and has conversations

backstory: |
  You are a friendly and helpful AI assistant.
  Your purpose is to have helpful conversations and answer questions.
  Be concise, friendly, and clear in your responses.
  You don't need tools for this - just conversation.

model: gpt-4o-mini
temperature: 0.7

tools: []           # No tools needed
handoff_targets: [] # No handoffs - single agent
is_terminal: true   # This agent returns final response

system_prompt: |
  You are {{name}}, a {{role}}.
  
  {{backstory}}
  
  Instructions:
  - Keep responses concise (2-3 sentences)
  - Be friendly and conversational
  - Answer the user's question directly
  - If asked about yourself, be honest: you're an AI assistant
```

### 5. .env.example

```bash
# Get your API key from https://platform.openai.com/api-keys
OPENAI_API_KEY=sk-...
```

### 6. README.md

```markdown
# Hello Crew: Your First Multi-Agent Crew

The simplest possible crew - just one agent having conversations.

**Learn**: How crews work, ExecuteStream, basic agent patterns
**Time**: 5 minutes to "hello world", 10 minutes to understand code
**Next**: Copy this and add more agents for your use case

## Quick Start (2 Minutes)

### Prerequisites
- Go 1.20+
- OpenAI API key from https://platform.openai.com/api-keys

### Setup

```bash
# 1. Clone repo
git clone <repo>
cd examples/00-hello-crew

# 2. Set environment
cp .env.example .env
# Edit .env: add your OPENAI_API_KEY

# 3. Run
go run ./cmd/main.go
```

### Use It

#### CLI Mode (Interactive)
```bash
go run ./cmd/main.go

> Tell me about yourself
Assistant: I'm an AI assistant here to help with questions and conversation...

> What can you help with?
Assistant: I can help with coding questions, explanations, creative writing...

> quit
```

#### Web UI Mode
```bash
go run ./cmd/main.go --server --port 8081
# Opens http://localhost:8081
```

## How It Works (Code Walkthrough - 10 Minutes)

### The Concept
```
Input (user question)
    ↓
Hello Agent (processes question)
    ↓
Output (agent response)
    ↓
Done (single agent is terminal)
```

### Key Code Sections

**main.go (40 lines)**
1. Read OPENAI_API_KEY from environment
2. Load agent from YAML config
3. Create Crew with 1 agent
4. Create CrewExecutor
5. Run in CLI or server mode

**hello.go (30 lines)**
- Define a conversational agent
- No tools (pure LLM)
- System prompt with instructions

**crew.yaml (10 lines)**
- entry_point: hello-agent (start here)
- agents: [hello-agent] (just one)
- settings: Basic configuration

## Modify It (10 Minutes)

### Make Agent More Specific

**In config/agents/hello-agent.yaml**:
```yaml
role: Python Coding Expert    # Change this
backstory: |
  You are an expert Python developer...
```

Restart and try:
```
> How do I sort a list in Python?
```

### Add Personality

**In system_prompt**:
```yaml
system_prompt: |
  You are {{name}}.
  Use 🤖 emoji when greeting!
  Add friendly jokes when appropriate.
```

### Change Model

**In config/agents/hello-agent.yaml**:
```yaml
model: gpt-4-turbo  # Faster/cheaper/different
```

## Scale It (Build Multi-Agent)

When ready, follow the pattern:

1. **Add second agent** (copy hello-agent.yaml)
2. **Define signals** (routing between agents)
3. **Add handoff_targets** (which agents can route to)
4. **Update crew.yaml** (add new agent to list)
5. **Create routing logic** (in system prompts)

See `examples/01-it-support/` for multi-agent pattern.

## Architecture Diagram

```
CrewExecutor
├── crew: Crew
│   └── agents: [hello-agent]
└── apiKey: "sk-..."

Execution:
User Input
    ↓ ExecuteStream()
    ↓ (streaming events)
Agent Processing
    ↓
LLM API Call
    ↓
Response
    ↓ Terminal
Done
```

## Streaming Events (What You See)

When you run the web UI, you see these events in real-time:

```
🔄 [hello-agent] Starting...
  (Agent is processing)

💬 [hello-agent] I'm a helpful AI assistant...
  (Agent's response arrives)

✅ Completed
  (All done)
```

## Common Questions

**Q: Can I add tools?**
A: Yes! Copy `examples/01-it-support/internal/tools.go` pattern.

**Q: How do I add another agent?**
A: See `examples/01-it-support/` for signal-based routing example.

**Q: Why is it just one agent?**
A: Start simple, learn the concept, then scale. Multi-agent adds complexity.

**Q: Can I use different LLM?**
A: Only OpenAI currently. Other models coming soon.

**Q: How do I deploy?**
A: See `docs/GUIDE_DEPLOYMENT.md`

## Next Steps

1. ✅ Run this example
2. ✅ Modify agent backstory/role
3. ✅ Test different prompts
4. 📖 Read: `docs/GUIDE_SIGNAL_ROUTING.md`
5. 🔨 Build: Multi-agent crew with routing

## File Reference

| File | Purpose | Lines |
|------|---------|-------|
| cmd/main.go | Entry point | 40 |
| internal/hello.go | Agent definition | 30 |
| config/crew.yaml | Crew config | 10 |
| config/agents/hello-agent.yaml | Agent config | 15 |
| .env.example | Environment template | 2 |

**Total: ~100 lines code**

## Learning Path

```
Hello Crew (this example)
    ↓ Understand: ExecuteStream, agents, system prompts
    ↓
IT Support (01-it-support)
    ↓ Learn: Multi-agent routing, signal-based handoffs
    ↓
Research Assistant (02-research-assistant)
    ↓ Discover: Complex workflows, tool integration
```

---

**Ready to learn how crews work? Run it now:** `go run ./cmd/main.go`
```

### 7. Makefile

```makefile
.PHONY: run run-server build clean help

help:
	@echo "Hello Crew - Minimal 1-Agent Example"
	@echo ""
	@echo "Commands:"
	@echo "  make run         Run in CLI mode (interactive chat)"
	@echo "  make run-server  Run in Web UI mode (http://localhost:8081)"
	@echo "  make build       Build binary"
	@echo "  make clean       Remove binary"

run:
	go run ./cmd/main.go

run-server:
	go run ./cmd/main.go --server --port 8081

build:
	go build -o hello-crew ./cmd/main.go

clean:
	rm -f hello-crew
```

---

## User Journey

### 5-Minute "Hello World"

```
0. User reads README
   ↓ "This looks simple"

1. Clone repo (10 sec)
   git clone <repo>
   cd examples/00-hello-crew

2. Set environment (20 sec)
   cp .env.example .env
   # Edit .env with API key

3. Run (10 sec)
   go run ./cmd/main.go

4. Interact (2 min)
   > Tell me about yourself
   Assistant: I'm an AI assistant...

5. Understand (2 min)
   Read code, see: It's just 3 files, ~80 lines
   "I could extend this"
```

### 10-Minute "I Understand"

```
6. Open README code walkthrough
   Sees diagram:
   Input → Agent → Output
   
7. Read main.go (2 min)
   Understands: Load config, create crew, execute
   
8. Read hello.go (1 min)
   Understands: Agent has role, backstory, prompt
   
9. Read crew.yaml (30 sec)
   Understands: Config structure
   
10. Try modification (2 min)
    Edit hello-agent.yaml backstory
    Restart, see different behavior
    "I can customize this!"
```

### 15-Minute "I Can Build"

```
11. User copies example
    cp -r 00-hello-crew my-crew
    
12. User modifies:
    - Agent role (Python expert)
    - Backstory (specific instructions)
    - System prompt (custom behavior)
    
13. User runs:
    go run ./my-crew/cmd/main.go
    
14. User verifies:
    "Works perfectly. Now what if I add another agent?"
    
15. User reads IT Support example (01-it-support)
    "Ah, I see signal routing..."
    "I can build this for my domain"
```

---

## Testing Checklist

- [ ] Clone example
- [ ] Set OPENAI_API_KEY
- [ ] `go run ./cmd/main.go` works
- [ ] Asks for input
- [ ] Agent responds
- [ ] `make run-server` launches web UI
- [ ] Web UI shows streaming events
- [ ] Edit config, restart, behavior changes
- [ ] User feels "I could build on this"

---

## Documentation Integration

This example supports these docs:

**docs/GUIDE_GETTING_STARTED.md** links to:
```
See examples/00-hello-crew for minimal example
(3 files, ~80 lines, 5 minutes to run)
```

**docs/GUIDE_ADDING_AGENTS.md** starts with:
```
Build on Hello Crew by copying its structure:
1. Create new agent config
2. Add to crew.yaml agents list
3. Define routing signals
```

**docs/GUIDE_SIGNAL_ROUTING.md** says:
```
Hello Crew uses no routing (single agent).
For multi-agent routing, see IT Support example (01-it-support).
```

---

## Success Criteria

✅ User runs example in < 3 minutes
✅ User understands code in < 10 minutes
✅ User can modify it in < 5 minutes
✅ User feels ready to build own crew
✅ Code is copyable pattern for their domain
✅ Serves as reference for all other examples

---

## Implementation Checklist

- [ ] Create directory structure
- [ ] Write cmd/main.go
- [ ] Write internal/hello.go
- [ ] Create config files (crew.yaml, agent.yaml)
- [ ] Create .env.example
- [ ] Create README.md with walkthrough
- [ ] Create Makefile
- [ ] Test locally (run, modify, scale)
- [ ] Integrate with main examples/README.md
- [ ] Link from docs/GUIDE_GETTING_STARTED.md

---

## Estimated Effort

| Task | Time |
|------|------|
| Code implementation | 1-2 hours |
| Testing | 30 minutes |
| Documentation | 1 hour |
| Integration | 30 minutes |
| **Total** | **3-4 hours** |

---

## Why This Matters

**Hello Crew** is the foundation for all learning:
- Fastest path to "working system"
- Simplest code to understand
- Easiest to modify and extend
- Pattern for building own crews
- Builds confidence before multi-agent complexity

Starting here transforms user perception from:
> "This looks complicated..." 

to:
> "I can build this! It's elegant!"

