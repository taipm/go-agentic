# Hello Crew: Implementation Action Plan

**Created:** 2025-12-22
**Timeline:** 3-4 hours
**Effort:** 1 developer
**Priority:** CRITICAL - Foundation for all other examples

---

## Phase 0: Preparation (15 minutes)

### 0.1 Create Directory Structure
```bash
mkdir -p examples/00-hello-crew/{cmd,internal,config/agents}
cd examples/00-hello-crew
```

### 0.2 Initialize Module
```bash
# From examples/00-hello-crew/
cat > go.mod << 'EOF'
module github.com/taipm/go-agentic/examples/hello-crew

go 1.20

require github.com/taipm/go-agentic/core v0.0.0

replace github.com/taipm/go-agentic/core => ../../core
EOF

touch go.sum
```

### 0.3 Verify core Library Accessible
```bash
# From go-agentic root
go mod tidy
# Ensure core/ compiles
go build ./core
```

---

## Phase 1: Core Code (90 minutes)

### 1.1 Implement cmd/main.go (30 minutes)

**File:** `examples/00-hello-crew/cmd/main.go`

**Content:** (See HELLO-CREW-DESIGN.md - cmd/main.go section)

**Key Sections:**
- [ ] Package declaration + imports
- [ ] main() function
- [ ] Flag parsing (--server, --port)
- [ ] OPENAI_API_KEY check
- [ ] Config loading logic
- [ ] Crew/Executor creation
- [ ] runCLI() function
- [ ] runServer() function
- [ ] loadAgentConfig() placeholder

**Testing:**
```bash
go build ./cmd/main.go  # Should compile
```

### 1.2 Implement internal/hello.go (30 minutes)

**File:** `examples/00-hello-crew/internal/hello.go`

**Content:** (See HELLO-CREW-DESIGN.md - internal/hello.go section)

**Key Sections:**
- [ ] Package declaration
- [ ] GetHelloAgent() function
- [ ] Agent struct fields (all required)
- [ ] SystemPrompt with template variables

**Testing:**
```go
// Quick test in main_test.go
agent := GetHelloAgent()
assert agent.ID == "hello-agent"
assert agent.IsTerminal == true
assert agent.Tools == []  // No tools
```

### 1.3 Create Config Files (30 minutes)

#### 1.3.1 config/crew.yaml
```yaml
version: "1.0"
description: "Hello Crew - Minimal single-agent example"

entry_point: hello-agent

agents:
  - hello-agent

settings:
  max_handoffs: 0
  max_rounds: 5
  timeout_seconds: 30

routing:
  signals: {}
  defaults: {}
```

**Validation:**
```bash
# Manually verify YAML is valid
cat config/crew.yaml  # Should print without errors
```

#### 1.3.2 config/agents/hello-agent.yaml
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

tools: []
handoff_targets: []
is_terminal: true

system_prompt: |
  You are {{name}}, a {{role}}.
  
  {{backstory}}
  
  Instructions:
  - Keep responses concise (2-3 sentences)
  - Be friendly and conversational
  - Answer the user's question directly
  - If asked about yourself, be honest: you're an AI assistant
```

#### 1.3.3 .env.example
```bash
# Get your API key from https://platform.openai.com/api-keys
OPENAI_API_KEY=sk-...
```

---

## Phase 2: Documentation (60 minutes)

### 2.1 Write README.md (40 minutes)

**File:** `examples/00-hello-crew/README.md`

**Content:** (See HELLO-CREW-DESIGN.md - README.md section)

**Sections:**
- [ ] Title + description
- [ ] Quick Start (2 min walkthrough)
- [ ] How It Works (10 min code walkthrough)
- [ ] Modify It (3 examples)
- [ ] Scale It (path to multi-agent)
- [ ] Architecture Diagram
- [ ] Streaming Events Explanation
- [ ] FAQ (5-6 common questions)
- [ ] Next Steps

**Testing:**
- [ ] README renders properly in markdown viewer
- [ ] All code examples are syntactically correct
- [ ] Links are valid

### 2.2 Create Makefile (10 minutes)

**File:** `examples/00-hello-crew/Makefile`

**Content:**
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

**Testing:**
```bash
make help      # Should show commands
make build     # Should compile
make clean     # Should remove binary
```

### 2.3 Create .gitignore

**File:** `examples/00-hello-crew/.gitignore`

```
hello-crew
.env
.vscode
.idea
*.swp
```

---

## Phase 3: Testing (45 minutes)

### 3.1 Compilation Test
```bash
cd examples/00-hello-crew
go build ./cmd/main.go
# Should complete with no errors
```

### 3.2 CLI Mode Test
```bash
# Setup
cp .env.example .env
# Edit .env with OPENAI_API_KEY

# Run
go run ./cmd/main.go

# Test input
> Tell me about yourself
# Should get agent response

# Test quit
> quit
# Should exit cleanly
```

### 3.3 Web Server Test (if HTTP handler exists)
```bash
go run ./cmd/main.go --server --port 8081
# Visit http://localhost:8081
# Should see web UI
# Test a query
# Should see streaming events
```

### 3.4 Modification Test
```bash
# Edit config/agents/hello-agent.yaml
# Change backstory to "Python expert"
# Restart app
# Test: "How do I sort a list in Python?"
# Should respond as Python expert
```

### 3.5 Code Clarity Test
```bash
# Read cmd/main.go (2 minutes)
# Could you understand it? Yes/No

# Read internal/hello.go (1 minute)
# Could you understand it? Yes/No

# Read config files (30 seconds)
# Could you understand it? Yes/No
```

**Success Criteria:**
- All tests pass
- Code is understandable
- README examples are accurate

---

## Phase 4: Integration (30 minutes)

### 4.1 Update Main examples/README.md

**Current:**
```markdown
## Examples

- [IT Support](./it-support/) - Multi-agent routing with signal-based handoffs
- [Research Assistant](./research-assistant/) - Document analysis and synthesis
- [Vector Search](./vector-search/) - Semantic search with Qdrant
- [Customer Service](./customer-service/) - Multi-language support
- [Data Extraction](./data-extraction/) - PDF processing
```

**Updated:**
```markdown
## Examples

### 🎯 Start Here

- **[00-hello-crew](./00-hello-crew/)** ⭐ START HERE
  - Minimal 1-agent example (5 min to run)
  - Learn: ExecuteStream, agents, system prompts
  - ~100 lines code, perfect for learning

### 🚀 Intermediate

- [01-it-support](./it-support/) - Multi-agent routing
  - Learn: Signal-based handoffs, tool integration
  - 3 agents, 13 tools, production-ready
  
- [02-research-assistant](./research-assistant/) - (Coming soon)
  - Learn: Complex workflows, document analysis

### 🔬 Advanced

- [03-vector-search](./vector-search/) - Semantic search
- [04-customer-service](./customer-service/) - Multi-language
- [05-data-extraction](./data-extraction/) - PDF processing
```

### 4.2 Add Link to Getting Started Docs

**In docs/GUIDE_GETTING_STARTED.md:**

Add section:
```markdown
## Start With Hello Crew

The fastest way to see go-agentic in action:

```bash
cd examples/00-hello-crew
cp .env.example .env
# Add OPENAI_API_KEY to .env

go run ./cmd/main.go
```

That's it! You now have a working 1-agent crew.

**Next:** Read the [code walkthrough](../../examples/00-hello-crew/README.md#how-it-works)
```

### 4.3 Update examples/00-hello-crew/README.md Cross-References

Ensure links to other docs:
- [ ] Link to docs/GUIDE_SIGNAL_ROUTING.md
- [ ] Link to docs/GUIDE_ADDING_AGENTS.md
- [ ] Link to examples/01-it-support for scaling

---

## Phase 5: Verification (30 minutes)

### 5.1 Complete Feature Checklist

- [ ] Directory structure created
- [ ] cmd/main.go implemented and compiles
- [ ] internal/hello.go implemented and compiles
- [ ] config/crew.yaml created
- [ ] config/agents/hello-agent.yaml created
- [ ] .env.example created
- [ ] README.md written with all sections
- [ ] Makefile created
- [ ] .gitignore created
- [ ] go.mod/go.sum configured

### 5.2 Functional Testing Checklist

- [ ] `go run ./cmd/main.go` runs
- [ ] CLI mode: accepts input and returns response
- [ ] CLI mode: agent provides sensible response
- [ ] CLI mode: quit command works
- [ ] Web UI mode: `--server` flag works
- [ ] Web UI mode: http://localhost:8081 loads
- [ ] Web UI mode: streaming events visible
- [ ] Modification: editing config changes behavior
- [ ] No compiler warnings
- [ ] No runtime errors with valid API key

### 5.3 User Experience Testing

- [ ] New user reads README
- [ ] New user runs example in < 3 minutes
- [ ] New user understands code in < 10 minutes
- [ ] New user can modify it in < 5 minutes
- [ ] New user feels "I can extend this"

### 5.4 Documentation Quality

- [ ] README sections complete
- [ ] Code examples syntactically correct
- [ ] Links valid and helpful
- [ ] Walkthrough clear and step-by-step
- [ ] Next steps obvious

---

## Success Criteria: Definition of Done

### Code Quality
✅ Compiles without errors
✅ Runs without crashes (with valid API key)
✅ Code is clean and readable (< 100 lines + 50 lines config)
✅ No hardcoded values (all configurable)

### Functionality
✅ CLI mode works (input → response → quit)
✅ Web UI mode works (http://localhost:8081)
✅ Streaming events visible
✅ Configuration changes reflected in behavior

### Documentation
✅ README complete with 7 sections
✅ Code walkthrough takes ~10 minutes
✅ User can modify agent in < 5 minutes
✅ Clear "next steps" to scale

### User Experience
✅ "Hello World" in 2-3 minutes
✅ Understand code in 10 minutes
✅ Extend it in 15 minutes
✅ User feels ready to build own crew

---

## Timeline Summary

| Phase | Task | Time | Status |
|-------|------|------|--------|
| 0 | Setup directories + modules | 15m | ⏳ |
| 1 | Implement 3 code files + configs | 90m | ⏳ |
| 2 | Write README + Makefile | 60m | ⏳ |
| 3 | Run all tests | 45m | ⏳ |
| 4 | Update cross-references | 30m | ⏳ |
| 5 | Final verification | 30m | ⏳ |
| **Total** | **Complete Hello Crew** | **270m (4.5h)** | ⏳ |

---

## Immediate Next Steps

### TODAY (This Implementation)

1. [ ] Create directory structure
2. [ ] Implement code files (cmd/main.go, internal/hello.go)
3. [ ] Create config files
4. [ ] Write README
5. [ ] Create Makefile
6. [ ] Test everything

### TOMORROW (Integration)

7. [ ] Update examples/README.md
8. [ ] Update docs/GUIDE_GETTING_STARTED.md
9. [ ] Verify cross-references
10. [ ] Get team feedback

### NEXT WEEK (Documentation)

11. [ ] Complete other examples based on Hello Crew pattern
12. [ ] Write guides (GUIDE_ADDING_AGENTS.md, etc.)
13. [ ] Update main README.md

---

## File Checklist

Create these files:

```
examples/00-hello-crew/
├── cmd/main.go                 # 40 lines
├── internal/hello.go           # 30 lines
├── config/crew.yaml            # 10 lines
├── config/agents/
│   └── hello-agent.yaml        # 15 lines
├── .env.example                # 2 lines
├── .gitignore                  # 5 lines
├── go.mod                       # 5 lines
├── go.sum                       # (empty or minimal)
├── Makefile                     # 15 lines
└── README.md                    # 150 lines

Total files: 10
Total lines: ~280 (code: ~100, config: ~30, docs: ~150)
```

---

## Risk & Mitigation

| Risk | Mitigation |
|------|-----------|
| HTTP handler may not be implemented | Use CLI mode only if needed |
| Config loading may be stubbed | Inline config in code (fallback) |
| YAML parsing may fail | Use JSON or inline Go config |
| API key missing | Clear error message + fix guide |

---

## Success Definition

After implementing Hello Crew:

✅ **Learning Path Clear**: Users see 00-hello-crew → 01-it-support → advanced examples
✅ **Fastest Path to Value**: 2-3 minutes to working system
✅ **Confidence Building**: Users understand code and can modify it
✅ **Foundation Solid**: Pattern for all future examples
✅ **Documentation Gap Closed**: Hello Crew fills "what is a crew?" gap

---

## Go Build It! 🚀

Ready to create the most elegant, simple, perfect-for-learning example of go-agentic?

This is the foundation. Everything else builds on this.

**Let's make go-agentic feel complete.**

