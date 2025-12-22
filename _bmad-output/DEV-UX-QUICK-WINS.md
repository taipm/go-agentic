# DEV UX Quick Wins - Immediate Action Items

**Created:** 2025-12-22
**Scope:** Highest impact improvements for users (this week)
**Effort:** ~3-5 hours total
**Expected Impact:** Unblock 70% of struggling users

---

## Quick Win 1: Complete research-assistant Example

### Current State
```
examples/research-assistant/
├── cmd/main.go          # Likely stub
├── config/             # Empty or incomplete
├── internal/           # Missing
└── README.md          # Missing
```

### What Users Expect
- Clone → Set env var → Run → Get research results
- 5 minute time to working
- Real web search + synthesis pattern
- Usable in production (not toy example)

### Required Work

**1. main.go (15 lines)**
```go
func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    executor, _ := crewai.NewCrewExecutorFromConfig(apiKey, "config", getAllTools())
    
    task := "Research: How does vector search work in production?"
    result, _ := executor.Execute(context.Background(), task)
    
    fmt.Println(result.Content)
}
```

**2. Config Files**
- crew.yaml: Define researcher → synthesizer → reviewer pipeline
- agents/researcher.yaml: Search and gather information
- agents/synthesizer.yaml: Combine findings with analysis
- agents/reviewer.yaml: Quality check and polish

**3. Tools (Use real APIs or mock)**
- WebSearch (use DuckDuckGo or mock)
- DocumentAnalysis (parse search results)
- CitationGenerator (format sources)

**4. README.md**
- What this example teaches
- How to run it
- How to modify for your domain
- 3 common use cases

### Time Estimate: 2-3 hours

---

## Quick Win 2: Write GUIDE_GETTING_STARTED.md

### Current Problem
```
User reads: 5 different README sections
Questions: "Which one is for me?"
Result: Confusion before running anything
```

### Solution: Single, Clear Getting Started Guide

**File:** `docs/GUIDE_GETTING_STARTED.md` (10 minutes to create, reusable forever)

```markdown
# Getting Started with go-agentic

## 🚀 Start Here (5 Minutes)

### Step 1: Set Up Your Environment
- Install Go 1.20+
- Get OpenAI API key from openai.com
- Set environment variable: `export OPENAI_API_KEY=sk-...`

### Step 2: Clone and Run
```bash
git clone <repo>
cd examples/it-support
go run ./cmd/main.go
```
Input: "My CPU is at 95%. What's happening?"
Output: Full system diagnosis

### Step 3: Try the Web UI (optional)
```bash
go run ./cmd/main.go --server --port 8081
# Open http://localhost:8081
```

## 📚 Learn More

- **I want to change the prompt** → GUIDE_MODIFYING_EXAMPLES.md
- **I want to add a new tool** → GUIDE_ADDING_TOOLS.md
- **I want to add a new agent** → GUIDE_ADDING_AGENTS.md
- **I want to deploy to production** → GUIDE_DEPLOYMENT.md

## ❓ Something Not Working?
- Check: docs/ERROR_CODES.md
- Common issue: OPENAI_API_KEY not set
- Common issue: config/ directory not found
```

### Time Estimate: 1 hour

---

## Quick Win 3: Write GUIDE_ADDING_TOOLS.md (with Templates)

### Current Problem
```
User wants: Add custom tool
Solution path: Read tools.go source code (15 min)
Better solution: Follow guide with copy-paste template (2 min)
```

### Solution: Template-Based Tool Addition Guide

**File:** `docs/GUIDE_ADDING_TOOLS.md` (2 hours to create well)

```markdown
# Adding Custom Tools to Your Crew

## Simple Tool Template (Copy-Paste)

```go
// In internal/tools.go

var GetMyCustomTool = &crewai.Tool{
    Name:        "GetMyCustom",           // ← Change this
    Description: "Does something useful", // ← Change this
    
    Parameters: map[string]interface{}{
        "type":       "object",
        "properties": map[string]interface{}{
            "input_param": map[string]interface{}{
                "type":        "string",
                "description": "What this param does",
            },
        },
        "required": []string{"input_param"},
    },
    
    Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
        param := args["input_param"].(string)
        
        // ← Your logic here
        result := doSomething(param)
        
        return result, nil
    },
}
```

## Example 1: Simple String Return

```go
var GetUserEmail = &crewai.Tool{
    Name:        "GetUserEmail",
    Description: "Look up user's email by name",
    Parameters: map[string]interface{}{...},
    Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
        name := args["name"].(string)
        email := lookupEmail(name)  // Your function
        return fmt.Sprintf("Email: %s", email), nil
    },
}
```

## Example 2: With Multiple Parameters

```go
var CalculateTax = &crewai.Tool{
    Name:        "CalculateTax",
    Description: "Calculate tax based on income and state",
    Parameters: map[string]interface{}{
        "properties": map[string]interface{}{
            "income": {...},
            "state":  {...},
        },
        "required": []string{"income", "state"},
    },
    Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
        income := args["income"].(float64)
        state := args["state"].(string)
        tax := calculateTax(income, state)
        return fmt.Sprintf("Tax owed: $%.2f", tax), nil
    },
}
```

## Example 3: With Error Handling

```go
var VerifyEmail = &crewai.Tool{
    Name: "VerifyEmail",
    Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
        email := args["email"].(string)
        
        if !isValidEmail(email) {
            return "", fmt.Errorf("invalid email format: %s", email)
        }
        
        verified := checkEmailExists(email)
        if !verified {
            return "", fmt.Errorf("email not found: %s", email)
        }
        
        return fmt.Sprintf("Email verified: %s", email), nil
    },
}
```

## Now Register Your Tool

### In internal/crew.go:

```go
func getAllMyTools() map[string]*crewai.Tool {
    return map[string]*crewai.Tool{
        "GetMyCustom":  GetMyCustomTool,  // ← Add here
        // ... existing tools ...
    }
}
```

### In config/agents/executor.yaml:

```yaml
tools:
  - GetMyCustom          # ← Add here (must match tool name)
  - GetCPUUsage
  # ... other tools ...
```

## Test It

```bash
# Edit config/agents/executor.yaml to reference new tool
# Restart: go run ./cmd/main.go
# Try a task that uses your tool

# In CLI: "Use GetMyCustom to find the user's email"
# Agent should now use your custom tool
```

## Common Patterns

### Pattern 1: API Call Tool
```go
Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(...)
    // Parse JSON response
    return formatted, nil
}
```

### Pattern 2: Database Query
```go
Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
    db := getConnection()  // Your DB connection
    rows, err := db.Query(...)
    // Format results
    return formatted, nil
}
```

### Pattern 3: Shell Command Execution
```go
Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
    cmd := exec.CommandContext(ctx, "command", "args...")
    output, err := cmd.Output()
    return string(output), err
}
```

## Troubleshooting

**Error: "tool not found"**
- Check: Tool name in agents/executor.yaml matches tool Name in Go code
- Tool names are case-sensitive

**Error: "type assertion failed"**
- Check: Parameter types match your assertions
- String: `args["param"].(string)`
- Number: `args["param"].(float64)`
- Boolean: `args["param"].(bool)`

**Tool not being called**
- Ensure tool is in agents/executor.yaml tools list
- Ensure agent role/instructions mention when to use it
```

### Time Estimate: 2 hours (with examples)

---

## Quick Win 4: Write GUIDE_SIGNAL_ROUTING.md

### Current Problem
```
User sees: orchestrator.yaml (160+ lines of signals)
Questions: "What are these [ROUTE_*] things?"
Result: Confused about core routing concept
```

### Solution: Clear guide to signal-based routing

**File:** `docs/GUIDE_SIGNAL_ROUTING.md` (1.5 hours)

```markdown
# Signal-Based Routing Explained

## What Are Signals?

A signal is a **explicit message from agent saying "route to X"**.

Instead of: "Agent1 → Agent2 (automagic)"
We have: "Agent1 → outputs [ROUTE_AGENT2] → framework reads signal → routes to Agent2"

**Why this is better:**
- Agent explicitly decides who gets the work
- Clear in system prompt what agent should do
- Errors are obvious (agent forgot signal or wrong signal)
- Easy to debug (check output for signals)

## Example: IT Support Routing

### Agent Roles

**Orchestrator** (Entry point)
- Reads customer issue
- Decides: "Do I have enough info or need more?"
- Outputs: [ROUTE_CLARIFIER] or [ROUTE_EXECUTOR]

**Clarifier** (Info gatherer)
- Asks probing questions
- Gathers context
- Outputs: [KẾT_THÚC] (Vietnamese for COMPLETE)

**Executor** (Action taker)
- Gets all context
- Runs diagnostics
- Outputs: (terminal, no signal needed)

### How Signals Connect Them

#### In crew.yaml:

```yaml
routing:
  signals:
    orchestrator:
      - signal: "[ROUTE_CLARIFIER]"
        target: clarifier
      - signal: "[ROUTE_EXECUTOR]"
        target: executor
    
    clarifier:
      - signal: "[KẾT_THÚC]"
        target: executor
    
    executor:
      # Terminal agent, no signals defined
```

#### In agents/orchestrator.yaml:

```yaml
allowed_signals:
  - "[ROUTE_CLARIFIER]"
  - "[ROUTE_EXECUTOR]"

system_prompt: |
  You are the IT support orchestrator.
  
  Always end your response with ONE signal:
  - [ROUTE_CLARIFIER] if you need more info about the issue
  - [ROUTE_EXECUTOR] if you have enough info to diagnose
  
  Example:
  Input: "My internet is down"
  Output: "I need more info. Is it WiFi or Ethernet? Is your router on? 
           [ROUTE_CLARIFIER]"
  
  Input: "My internet is down. Ethernet. Router is on. It worked 5 mins ago."
  Output: "I have enough info to diagnose. [ROUTE_EXECUTOR]"
```

## Step-by-Step: Create Custom Routing

### Scenario: Support → Expert → Manager (escalation)

#### Step 1: Define Agents in crew.yaml

```yaml
agents:
  - support
  - expert
  - manager
```

#### Step 2: Define Signals in crew.yaml

```yaml
routing:
  signals:
    support:
      - signal: "[RESOLVED]"
        target: null              # Terminal - done
      - signal: "[ESCALATE_EXPERT]"
        target: expert            # Send to expert
    
    expert:
      - signal: "[ESCALATE_MANAGER]"
        target: manager           # Send to manager
      - signal: "[RESOLVED]"
        target: null              # Done
    
    manager:
      - signal: "[RESOLVED]"
        target: null              # Always terminal
```

#### Step 3: Create agent configs

```yaml
# agents/support.yaml
allowed_signals:
  - "[RESOLVED]"
  - "[ESCALATE_EXPERT]"

system_prompt: |
  You are a support agent.
  
  If you can resolve the issue, end with [RESOLVED].
  If issue is complex, end with [ESCALATE_EXPERT].
  
  Guidance:
  - Simple: billing, password reset, basic setup → [RESOLVED]
  - Complex: bugs, system design, integration → [ESCALATE_EXPERT]
  - Very complex: architecture decisions, policy exceptions → [ESCALATE_MANAGER]
```

#### Step 4: Test

```bash
# Start app
go run ./cmd/main.go

# Try simple issue
Input: "I forgot my password"
Support processes → [RESOLVED]
Done

# Try complex issue
Input: "How do I integrate with your API?"
Support → [ESCALATE_EXPERT]
Expert processes → [ESCALATE_MANAGER] (if policy decision needed)
Manager finalizes
```

## Common Patterns

### Pattern 1: Binary Decision (Yes/No)

```yaml
signals:
  checker:
    - signal: "[APPROVED]"
      target: executor
    - signal: "[REJECTED]"
      target: requester
```

### Pattern 2: Multi-Way (3+ branches)

```yaml
signals:
  router:
    - signal: "[ROUTE_SALES]"
      target: sales
    - signal: "[ROUTE_SUPPORT]"
      target: support
    - signal: "[ROUTE_BILLING]"
      target: billing
```

### Pattern 3: Sequential Pipeline

```yaml
signals:
  collector:
    - signal: "[COLLECTED]"
      target: analyzer
  
  analyzer:
    - signal: "[ANALYZED]"
      target: synthesizer
  
  synthesizer:
    - signal: "[READY]"
      target: null  # Terminal
```

## Debugging Signals

### Problem: Agent not routing

```
Expected: Agent outputs [ROUTE_EXPERT]
Reality: Agent doesn't output signal
Check:
1. Agent system_prompt tells it when to emit signal?
2. Is signal in allowed_signals?
3. Is signal exactly spelled right? ([ROUTE_EXPERT] not [Route_Expert])
```

### Problem: Wrong routing destination

```
Expected: [ROUTE_EXPERT] → expert
Reality: [ROUTE_EXPERT] → something else
Check:
1. routing.signals.agent_name: [signal matches target?
2. target agent ID exists in agents list?
3. Signal pattern matched correctly?
```

## Important Rules

1. **Signals must be in [BRACKETS]**
   - ✅ [ROUTE_EXPERT]
   - ❌ ROUTE_EXPERT
   - ❌ Route Expert

2. **Signals must be in allowed_signals**
   - If agent outputs [ROUTE_X] but X not in allowed_signals → error

3. **Signals must be defined in crew.yaml**
   - If crew.yaml doesn't know about signal → agent not routed

4. **Use clear, descriptive signal names**
   - ✅ [ROUTE_EXPERT]
   - ✅ [ESCALATE]
   - ❌ [NEXT]
```

### Time Estimate: 1.5 hours

---

## Quick Win 5: Improve Error Messages

### Current State
```
Error: "tool execution error"
What to do: ??? (requires debugging logs)
```

### Better Errors

```
Error: Signal Validation Failed
  Agent 'orchestrator' outputs signal [ROUTE_INVALID]
  But [ROUTE_INVALID] not found in allowed_signals
  
  Fix: Add '[ROUTE_INVALID]' to orchestrator.yaml:allowed_signals
  
  Or if this is a typo:
  - Did you mean '[ROUTE_EXECUTOR]'?
  - Did you mean '[ROUTE_CLARIFIER]'?
  
  See: docs/GUIDE_SIGNAL_ROUTING.md

---

Error: Configuration File Not Found
  Expected: config/crew.yaml
  Actual: File does not exist
  
  Fix: Create directory structure:
    config/
    ├── crew.yaml
    └── agents/
        ├── agent1.yaml
        └── agent2.yaml
  
  Or pass configDir to NewCrewExecutorFromConfig()
  
  See: docs/GUIDE_GETTING_STARTED.md
```

### Implementation
Update error handling in `core/config.go` and `core/validation.go` to provide helpful messages.

### Time Estimate: 1.5 hours

---

## Total Quick Wins Effort

| Task | Hours | Priority |
|------|-------|----------|
| Complete research-assistant example | 2-3 | ⭐⭐⭐ |
| Write GUIDE_GETTING_STARTED.md | 1 | ⭐⭐⭐ |
| Write GUIDE_ADDING_TOOLS.md | 2 | ⭐⭐⭐ |
| Write GUIDE_SIGNAL_ROUTING.md | 1.5 | ⭐⭐⭐ |
| Improve error messages | 1.5 | ⭐⭐ |
| **TOTAL** | **8-9 hours** | |

---

## Expected Impact

### Before Quick Wins
- User struggles: 50% of new users hit friction
- Needs: 15+ minutes to debug simple issues
- Reads source code: 70% of custom tool additions
- Examples: Only 1/5 complete

### After Quick Wins
- User struggles: 15% of new users (unblocked 70%)
- Needs: < 2 minutes to debug (if error messages are good)
- Reads source code: 10% (now have guides)
- Examples: 2/5 complete

---

## Implementation Checklist

### Week 1 Actions

- [ ] Complete research-assistant example
  - [ ] Implement main.go
  - [ ] Create crew.yaml with researcher → synthesizer pipeline
  - [ ] Create agent configs
  - [ ] Implement web search tool
  - [ ] Write README.md

- [ ] Create 4 documentation guides
  - [ ] docs/GUIDE_GETTING_STARTED.md
  - [ ] docs/GUIDE_ADDING_TOOLS.md (with 5 examples)
  - [ ] docs/GUIDE_SIGNAL_ROUTING.md
  - [ ] Update main README.md with links to guides

- [ ] Improve error messages
  - [ ] Configuration not found → suggest directory structure
  - [ ] Signal not found → suggest possible signals
  - [ ] Tool not found → suggest available tools
  - [ ] Parameter validation → show parameter schema

---

## Success Criteria

After implementing these quick wins:

✅ New user can run IT Support example in 2 minutes
✅ New user can add custom tool following guide (5 minutes)
✅ New user can create new agent following guide (10 minutes)
✅ Error messages guide user to fix + documentation
✅ Research-assistant example is production-ready
✅ Zero questions about "how do I...?" in issues (they're in guides)

---

## Recommended Execution Order

1. **Start:** Write GUIDE_GETTING_STARTED.md (foundation)
2. **Then:** Write GUIDE_ADDING_TOOLS.md (unblock tool additions)
3. **Then:** Write GUIDE_SIGNAL_ROUTING.md (unblock customization)
4. **Parallel:** Complete research-assistant example
5. **Finally:** Improve error messages (polish)

This order unblocks users progressively while building supporting content.

