---
title: "UX Design: go-agentic Library Integration Experience"
version: "1.0.0"
date: "2025-12-20"
project: "go-agentic"
type: "UX Design"
status: "In Review"
---

# UX Design Document
## go-agentic Library Integration Experience

**Project:** go-agentic
**Version:** 1.0.0
**Date:** 2025-12-20
**Audience:** Developers using go-agentic library

---

## 1. UX VISION

### Target User Experience

**For Library Users (Go Developers):**
> "I can confidently configure multi-agent systems, knowing each setting will be honored, tools will execute reliably, and errors will be clear and actionable."

**For Example Users (DevOps/Operators):**
> "The IT Support example works seamlessly on any platform I deploy it to, with clear error messages when something goes wrong."

**For Contributors:**
> "The codebase is well-organized, testable, and I can understand failure modes immediately."

---

## 2. USER JOURNEYS

### Journey 1: Library User - Configure and Deploy Multi-Agent System

```
┌─────────────────────────────────────────────────────────────┐
│ USER GOALS                                                   │
├─────────────────────────────────────────────────────────────┤
│ 1. Create agent with specific model (e.g., gpt-4o)          │
│ 2. Define tools for agent to use                            │
│ 3. Execute and verify model was used                        │
│ 4. Handle errors gracefully                                  │
└─────────────────────────────────────────────────────────────┘

CURRENT STATE (Problematic):
┌──────────────┬──────────────┬──────────────┬──────────────┐
│ Configure    │ Define Tools │ Execute      │ Debug        │
│ Agent with   │              │              │              │
│ Model: gpt-4o│              │ Model Used:  │ Error: ?     │
│              │              │ gpt-4o-mini  │ Unclear!     │
│              │              │ ❌ WRONG!    │              │
└──────────────┴──────────────┴──────────────┴──────────────┘

DESIRED STATE (After Fix):
┌──────────────┬──────────────┬──────────────┬──────────────┐
│ Configure    │ Define Tools │ Execute      │ Debug        │
│ Agent with   │ + Validation │              │              │
│ Model: gpt-4 │              │ Model Used:  │ Error: Clear │
│              │              │ gpt-4o ✅    │ & Actionable │
│              │              │              │              │
└──────────────┴──────────────┴──────────────┴──────────────┘
```

### Journey 2: Example User - Deploy IT Support on Windows

```
CURRENT STATE (Fails):
┌──────────────┬──────────────┬──────────────┬──────────────┐
│ Load Example │ Configure    │ Test Ping    │ Debug        │
│              │ on Windows   │              │              │
│              │              │ Error: "ping │ What's wrong?│
│              │              │ -c unknown"  │ (confusing)  │
│              │              │ ❌ FAILS     │              │
└──────────────┴──────────────┴──────────────┴──────────────┘

DESIRED STATE (After Fix):
┌──────────────┬──────────────┬──────────────┬──────────────┐
│ Load Example │ Auto-detect  │ Test Ping    │ Works!       │
│              │ Windows      │              │              │
│              │ Use -n flag  │ Command:     │ Clear Success│
│              │              │ ping -n 4    │ Message      │
│              │              │ host ✅      │              │
└──────────────┴──────────────┴──────────────┴──────────────┘
```

### Journey 3: Error Handling - Service Status Check Fails

```
CURRENT STATE (Confusing):
User checks service status:
  → Tool runs: systemctl is-active nginx
  → Error occurs (permission denied)
  → Tool returns: "Service nginx is not running"
  → User thinks: "Service crashed!" 😕
  → Truth: "Command failed due to permissions" 😞

DESIRED STATE (Clear):
User checks service status:
  → Tool runs: systemctl is-active nginx
  → Error occurs (permission denied)
  → Tool returns: "[ERROR] Permission denied: cannot check service status"
  → User knows: "Run with sudo" ✅
  → Error is actionable
```

---

## 3. INTERACTION PATTERNS

### Pattern 1: Configuration Transparency

**Current Interaction:**
```go
agent := &agentic.Agent{
    Model: "gpt-4o",           // User sets this
    // ...
}
executor.Execute(ctx, input)
// ❓ Which model was used? Unclear from logs
```

**Desired Interaction:**
```go
agent := &agentic.Agent{
    Model: "gpt-4o",
}
executor.Execute(ctx, input)
// Logs should show: [INFO] Agent "orchestrator" using model "gpt-4o"
```

**Implementation Pattern:**
```go
// Add logging in agent execution
log.Printf("[INFO] Executing agent %s with model %s", agent.Name, agent.Model)
```

### Pattern 2: Tool Execution Feedback

**Current Interaction (Fragile):**
```
Agent: "I will call GetCPUUsage ( ) to check"
Library parses: Doesn't detect (has space)
Result: No tool call, misleading response
```

**Desired Interaction (Robust):**
```
Agent: "I will use the GetCPUUsage function"
Library: Uses native OpenAI API tool_calls
Result: Reliably executed

Agent: "Please call GetCPUUsage()"
Library: Fallback text parser works
Result: Still executed
```

**User-Facing Benefit:**
> "Tool calls are reliable regardless of how the agent phrases it"

### Pattern 3: Error Messages

**Current Error (Unclear):**
```
Error: failed to check service: exit status 1
```

**Desired Error (Actionable):**
```
Error: [PERMISSION_DENIED] Cannot check service status
- Reason: systemctl command requires elevated privileges
- Suggestion: Run with 'sudo' or add user to 'systemd-journal' group
- Raw error: Permission denied
```

### Pattern 4: Parameter Validation

**Current Interaction (Late Failure):**
```go
tool.Handler(ctx, map[string]interface{}{
    "path": 123,  // Wrong type!
})
// Fails inside handler with type assertion error
```

**Desired Interaction (Early Failure):**
```go
// Validation happens before handler call
err := validateToolParameters(tool, map[string]interface{}{
    "path": 123,
})
// Returns: "parameter 'path': expected string, got int"
// Clear error before handler is called
```

---

## 4. ERROR HANDLING EXPERIENCE

### Error Severity Levels

**LEVEL 1: Configuration Error** (happens during setup)
```
❌ Error: Configuration validation failed
   Field: Agent "executor" Model
   Issue: Empty model string
   Action: Set Model to valid value (e.g., "gpt-4o-mini")
```

**LEVEL 2: Execution Error** (happens during agent run)
```
❌ Error: Tool execution failed
   Tool: GetCPUUsage
   Reason: Command timed out
   Context: Executing on system with 10,000+ processes
   Action: Increase timeout or optimize system
```

**LEVEL 3: Tool Error** (happens in tool handler)
```
❌ Error: [PERMISSION_DENIED] Service check failed
   Service: apache2
   Reason: systemctl requires elevated privileges
   Action: Run with 'sudo' or configure service access
```

### Error Message Components

All errors should include:
1. **Error Type** - What kind of error (PERMISSION_DENIED, TIMEOUT, INVALID_PARAMETER)
2. **Context** - Where it happened (which agent, which tool)
3. **Reason** - Why it happened (clear cause)
4. **Action** - What to do about it (suggested fix)
5. **Raw Error** - Technical details if needed

### Error Message Templates

```
[ERROR_TYPE] High-level problem
   Context: Where it happened
   Reason: Why it happened
   Action: How to fix it
   Details: Technical info if applicable
```

---

## 5. CONFIGURATION EXPERIENCE

### Configuration Clarity

**User Mental Model (SHOULD BE):**
```
Agent Configuration
├── Model Selection
│   ├── Fast but less capable: gpt-4o-mini
│   ├── Balanced: gpt-4o
│   └── Powerful but expensive: (future models)
├── Tool Access
│   ├── Which tools agent can use
│   └── Parameter constraints
└── Behavior
    ├── Temperature: creativity level (0=deterministic, 1=creative)
    └── Is Terminal: stops workflow after?

✅ EACH SETTING IS RESPECTED IN EXECUTION
```

**Current Experience (BROKEN):**
```
Agent Configuration
├── Model: "gpt-4o"
│   └── Actually Uses: "gpt-4o-mini" ❌ (IGNORED!)
├── Tool Access: [Tool1, Tool2]
│   └── Validation: None ❌ (MISSING!)
└── Behavior: Temperature=0.7
    └── Override Bug: 0.0 → 0.7 ❌ (WRONG!)
```

### Configuration Validation

**When loading configuration:**

```
✅ Valid Configuration Example:
- Model: "gpt-4o-mini" (exists in OpenAI API)
- Temperature: 0.7 (0.0-2.0 range)
- Tools: [GetCPUUsage, PingHost] (defined in crew)

❌ Invalid Configuration Would Show:
- Model: "gpt-5" → Error: "Unknown model"
- Temperature: 3.0 → Error: "Temperature must be 0.0-2.0"
- Tools: [NonExistent] → Error: "Tool not found"
```

---

## 6. PLATFORM-SPECIFIC GUIDANCE

### Windows Users

**Discovery:**
User runs IT Support example on Windows.

**Experience (Current - BROKEN):**
```
Error: ping command failed with -c flag
You think: "This library doesn't work on Windows"
You leave: "Go to Python instead"
```

**Experience (Desired):**
```
Ping test automatically uses -n flag on Windows
Result: Same behavior as macOS/Linux
You think: "This library is professional grade"
You stay: "Use it for production"
```

**Design Pattern:**
```
No user action needed - library auto-detects platform and adapts
(No configuration, no warnings, just works)
```

### macOS/Linux Users

**Discovery:**
User runs IT Support example on their system.

**Experience (Current - WORKS):**
```
Everything works as expected
```

**Experience (Desired - BETTER):**
```
Same functionality
+ Clear error messages if something fails
+ Better diagnostics
```

---

## 7. FEEDBACK & DEBUGGING

### What Users Need to Know

**When Configuration Works:**
```
[INFO] Agent "orchestrator" initialized
       - Model: gpt-4o-mini
       - Tools: 0 available
       - Status: Ready

[INFO] Agent "executor" initialized
       - Model: gpt-4o-mini
       - Tools: 13 available (GetCPUUsage, PingHost, ...)
       - Status: Ready
```

**When Execution Succeeds:**
```
[DEBUG] Executing agent "orchestrator"
        Input: "My computer is slow"
        Model: gpt-4o-mini

[DEBUG] Agent response received
        Content: "I need more information. What OS are you using?"
        Tool calls: 0

[DEBUG] No tool calls, routing to next agent...

[INFO] Workflow complete
       Final agent: "orchestrator"
       Status: Success
```

**When Tool Executes:**
```
[DEBUG] Tool call detected: GetCPUUsage()

[DEBUG] Executing tool: GetCPUUsage
        Arguments: {}

[INFO] Tool executed successfully
       Result: "45.2%"
       Duration: 123ms
```

**When Error Occurs:**
```
[ERROR] Tool execution failed
        Tool: CheckServiceStatus
        Service: nginx
        Error: [PERMISSION_DENIED] systemctl requires elevated privileges

        Suggestion: Run with sudo or configure service access
        Raw error: exit status 1: "Failed to get properties:
                                  Access denied"
```

---

## 8. ACCESSIBILITY & CLARITY

### Language Clarity

**AVOID (Technical Jargon):**
```
"Tool call parsing via regex pattern ToolName\\(.*\\) failed"
"Exit code 126: Cannot execute binary"
```

**USE (User-Friendly):**
```
"Could not execute tool: Command not found or not executable"
"Suggestion: Check that the command is installed and in PATH"
```

### Consistent Terminology

- "Agent" not "bot" or "actor"
- "Tool call" not "function invocation"
- "Workflow" not "execution flow"
- "Handoff" not "delegation"
- "Terminal agent" not "end agent"

### Information Hierarchy

Most important info first:
1. **Error summary** (one line, clear)
2. **Context** (which agent/tool, what input)
3. **Root cause** (why it happened)
4. **Suggested action** (how to fix)
5. **Technical details** (for debugging)

---

## 9. DOCUMENTATION EXPERIENCE

### What Documentation Should Cover

**For Each Issue Fixed:**

1. **What was wrong** (problem description)
2. **Why it was wrong** (impact explanation)
3. **How it's fixed** (solution explanation)
4. **Code examples** (before/after)
5. **Migration guide** (if any changes needed)

### Example Documentation

```markdown
## Model Configuration is Now Honored

### The Issue
Previously, agent.Model was ignored and all agents used "gpt-4o-mini".

### Example
You configured an agent with Model: "gpt-4o" for intelligence, but it
actually used "gpt-4o-mini" for cost. This was:
- ❌ Unexpected (configuration ignored)
- ❌ Wasteful (couldn't optimize cost vs quality)
- ❌ Confusing (unclear why agent wasn't as smart)

### The Fix
Now agent.Model is properly respected:
- ✅ Configuration honored
- ✅ Can optimize cost vs quality per agent
- ✅ Predictable behavior

### No Action Needed
Existing code works unchanged. Configuration will now be properly used.

### Example
```go
orchestrator := &agentic.Agent{
    Model: "gpt-4o",  // ← Now actually used!
}
```
```

---

## 10. TESTING & VALIDATION EXPERIENCE

### User-Facing Test Feedback

**Current Experience (Unclear):**
```
go test ./...
ok      github.com/taipm/go-agentic  2.345s
```

**Desired Experience (Clear):**
```
Running IT Support Example Tests

✅ Test: Vague Issue Handling
   Input: "My computer is slow"
   Expected flow: orchestrator → clarifier → executor
   Result: PASS

✅ Test: Network Diagnostics
   Input: "Can't reach server at 192.168.1.100"
   Expected: PingHost tool called
   Result: PASS

✅ Test: Cross-Platform Compatibility
   Platform: windows
   Test: PingHost command
   Command used: ping -n 4 host
   Result: PASS

===========================================
Summary: 10/10 tests passed
Coverage: 92%
Platform: darwin (macOS)
===========================================
```

### Regression Prevention

Users should be confident that:
- ✅ Configuration changes don't break existing code
- ✅ Tool updates are backward compatible
- ✅ Error handling improvements make debugging easier
- ✅ Cross-platform fixes work everywhere

---

## 11. INTEGRATION CHECKLIST

### For Library Users (Go Developers)

**Before:**
- [ ] Test on your target platform (Windows/macOS/Linux)
- [ ] Verify your agent model configuration
- [ ] Check for any custom error handling

**After Update:**
- [ ] Verify agent uses configured model (check logs)
- [ ] Tool calls work reliably
- [ ] Error messages are clearer
- [ ] No code changes needed (backward compatible)

### For Example Users (DevOps/Operators)

**Before:**
- [ ] Run on current platform
- [ ] Note any error messages

**After Update:**
- [ ] Run on multiple platforms (Windows, macOS, Linux)
- [ ] Errors are clearer and actionable
- [ ] Same code works everywhere

---

## 12. SUCCESS METRICS (from User Perspective)

| Metric | Before | After | Success Criteria |
|--------|--------|-------|------------------|
| Agent Model Respected | ❌ 0% | ✅ 100% | All agents use config |
| Tool Reliability | 🟡 80% | ✅ 99% | Parsing robust |
| Cross-Platform Works | ❌ 1/3 | ✅ 3/3 | All platforms pass |
| Error Clarity | 🟡 50% | ✅ 95% | Errors actionable |
| Setup Time | 🟡 30 min | ✅ 5 min | Minimal friction |
| Confidence Level | 🟡 Medium | ✅ High | Know what's happening |

---

**Document Status:** Ready for Epic & Story Creation
**Next Steps:** Use all 3 documents (PRD, Architecture, UX) to create epics and user stories

