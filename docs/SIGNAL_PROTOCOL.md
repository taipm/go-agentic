# 📡 Signal Protocol Specification

**Version**: 1.0
**Date**: 2025-12-24
**Status**: Production Ready

---

## 1. Overview

The Signal Protocol defines how agents communicate routing decisions and workflow state changes through signal emission. It provides a standardized, type-safe framework for agent-to-agent communication in the CrewAI system.

### Key Principles

1. **Explicit Over Implicit**: All signals must be explicitly defined and registered
2. **Type-Safe**: Signal behaviors are enforced through the registry
3. **Fail-Fast**: Invalid signals are detected at startup, not at runtime
4. **Traceable**: All signal events are logged for debugging and monitoring
5. **Extensible**: New signals can be added without changing core logic

---

## 2. Signal Format Standard

### Format Definition

```
[SIGNAL_NAME]
```

### Requirements

- **Case-Sensitive**: Signal names ARE case-sensitive
  - `[END]` ≠ `[end]` (treated as different signals)

- **Bracket Format**: Must be enclosed in square brackets
  - Valid: `[END_EXAM]`, `[QUESTION]`, `[KẾT_THÚC]`
  - Invalid: `END_EXAM`, `[END_EXAM`, `END_EXAM]`, `[]`

- **Content Required**: Must have at least one character inside brackets
  - Valid: `[A]`, `[END]`, `[QUESTION_READY]`
  - Invalid: `[]` (empty brackets)

- **Alphanumeric + Underscore**: Signal names should use:
  - A-Z, a-z, 0-9, underscore (_), hyphen (-)
  - Unicode supported for Vietnamese: `[KẾT_THÚC]`, `[CÂU_HỎI]`

### Signal Appearance Rules

Signals must appear in agent responses in one of these forms:

1. **Exact Match**
   ```
   Task complete. [END]
   ```

2. **Case-Insensitive Match**
   ```
   Task complete. [end]
   ```

3. **Normalized Match** (for Vietnamese, whitespace variations)
   ```
   Task complete. [ Kết Thúc ]
   Matches: [KẾT_THÚC]
   ```

---

## 3. Signal Behaviors

### SignalBehaviorRoute

**Purpose**: Route execution to another agent

**Characteristics**:
- Requires non-empty target agent/group
- Transfers control to target agent
- Common for multi-agent workflows

**Example**:
```yaml
routing:
  signals:
    teacher:
      - signal: "[QUESTION]"
        target: "question_handler"
        description: "Route to question handler"
```

**YAML Configuration**:
```go
Signal: "[QUESTION]"
Target: "question_handler"  // Must be non-empty
```

### SignalBehaviorTerminate

**Purpose**: End the workflow cleanly

**Characteristics**:
- Must have EMPTY target
- Exits ExecuteStream loop
- Highest priority

**Example**:
```yaml
routing:
  signals:
    teacher:
      - signal: "[END_EXAM]"
        target: ""              // EMPTY = terminate
        description: "End exam"
```

**YAML Configuration**:
```go
Signal: "[END_EXAM]"
Target: ""  // MUST be empty
```

### SignalBehaviorPause

**Purpose**: Pause execution and wait for external input

**Characteristics**:
- Pauses current agent
- Waits for user/external signal
- Can be resumed with ResumeAgent

**Example**:
```yaml
routing:
  signals:
    agent:
      - signal: "[WAIT]"
        target: ""
        description: "Wait for user input"
```

### SignalBehaviorParallel

**Purpose**: Trigger parallel group execution

**Characteristics**:
- Targets a parallel group (not individual agent)
- Executes group concurrently
- Rejoins after completion

**Example**:
```yaml
routing:
  signals:
    coordinator:
      - signal: "[START_PARALLEL]"
        target: "parallel_group_name"
```

### SignalBehaviorBroadcast

**Purpose**: Send signal to multiple agents

**Characteristics**:
- Message sent to multiple targets
- Each agent processes independently
- Used for notifications/updates

---

## 4. Signal Definitions

### Built-in Signals

The system provides these default signals:

#### Termination Signals
```
[END]           - Generic workflow termination
[END_EXAM]      - End exam workflow
[DONE]          - Task completion
[STOP]          - Immediate stop (high priority)
```

#### Routing Signals
```
[NEXT]          - Route to next agent
[QUESTION]      - Route to question handler
[ANSWER]        - Route to answer handler
```

#### Status Signals
```
[OK]            - Acknowledgment
[ERROR]         - Error occurred (high priority)
[RETRY]         - Retry operation
```

#### Control Signals
```
[WAIT]          - Pause and wait for input
```

### Custom Signal Definition

To add a custom signal:

```go
// 1. Define the signal
definition := &SignalDefinition{
    Name:          "[MY_SIGNAL]",
    Description:   "Description of what this signal does",
    AllowAllAgents: true,           // or specify list
    Behavior:      SignalBehaviorRoute,
    DefaultTarget: "target_agent",
    ValidTargets:  []string{"agent1", "agent2"},
    Example:       "Task ready. [MY_SIGNAL]",
    Priority:      50,
}

// 2. Register it
registry := LoadDefaultSignals()  // Get default registry
registry.Register(definition)     // Add your signal
```

---

## 5. Agent Emission Rules

### Who Can Emit?

**AllowAllAgents = true** (default)
```
Any agent can emit the signal
```

**AllowedAgents specified**
```
Only listed agents can emit
Example: AllowedAgents: []string{"teacher", "coordinator"}
```

### Validation

```go
// Validator checks:
1. Signal format is valid ([NAME] format)
2. Signal is registered in registry
3. Agent is allowed to emit this signal
4. Target agent/group exists (if routing)
5. Target is in valid targets list (if specified)
```

---

## 6. Configuration Examples

### Example 1: Simple Route

```yaml
version: "1.0"
entry_point: "teacher"
agents: ["teacher", "reporter"]

routing:
  signals:
    teacher:
      - signal: "[QUESTION_READY]"
        target: "reporter"
        description: "Questions ready for reporting"
      - signal: "[END_EXAM]"
        target: ""  # Empty = terminate
        description: "Exam complete"
```

### Example 2: Multiple Routes

```yaml
routing:
  signals:
    analyzer:
      - signal: "[DATA_READY]"
        target: "processor"
      - signal: "[ERROR]"
        target: "error_handler"
      - signal: "[SKIP]"
        target: "reporter"
      - signal: "[END]"
        target: ""
```

### Example 3: Vietnamese Signals

```yaml
routing:
  signals:
    giao_vien:
      - signal: "[CÂU_HỎI_SẴN_SÀNG]"
        target: "bao_cao"
      - signal: "[KẾT_THÚC_THI]"
        target: ""
```

---

## 7. Signal Matching Algorithm

The system uses a 3-level matching strategy:

### Level 1: Exact Match (Fastest)
```
If response contains: "Task done. [END]"
And signal is:        "[END]"
→ Match ✓
```

### Level 2: Case-Insensitive
```
If response contains: "Task done. [end]"
And signal is:        "[END]"
→ Match ✓
```

### Level 3: Normalized
```
If response contains: "Task done. [ Kết Thúc ]"
And signal is:        "[KẾT_THÚC]"
→ Match ✓
(Whitespace and case normalized)
```

---

## 8. Error Handling

### Unknown Signal

**Error**: Signal not in registry
```
signal '[UNKNOWN]' is not registered (unknown signal)
```

**Resolution**:
1. Register signal in registry
2. Or use existing signal from defaults

### Unauthorized Agent

**Error**: Agent not allowed to emit signal
```
agent 'reporter' is not allowed to emit signal '[QUESTION]'
```

**Resolution**:
1. Add agent to AllowedAgents list
2. Or create new signal for this agent

### Invalid Target

**Error**: Target agent doesn't exist
```
signal '[NEXT]' targets unknown agent 'unknown_agent'
```

**Resolution**:
1. Check target agent ID
2. Or use different signal

### Wrong Behavior

**Error**: Target doesn't match behavior
```
termination signal '[END]' must have empty target, got 'other_agent'
```

**Resolution**:
1. Use termination signal only with empty target
2. Use routing signal with valid target

---

## 9. Best Practices

### ✅ DO

1. **Register signals before use**
   ```go
   registry := LoadDefaultSignals()
   registry.Register(mySignal)
   ```

2. **Use descriptive names**
   ```
   [QUESTION_READY]   ✓
   [Q]                ✗ Too vague
   ```

3. **Validate early**
   ```go
   validator.ValidateConfiguration(signals, agents)
   ```

4. **Log signal events**
   ```
   [SIGNAL-EVENT] Agent 'teacher' emitted '[END_EXAM]'
   ```

5. **Test signal definitions**
   ```go
   TestValidateSignalEmission
   TestValidateSignalTarget
   ```

### ❌ DON'T

1. **Don't use signals not in registry**
   ```go
   // ✗ Wrong - signal not registered
   response = "[CUSTOM_UNKNOWN]"
   ```

2. **Don't mix termination and routing**
   ```yaml
   # ✗ Wrong - termination signal with target
   - signal: "[END]"
     target: "agent"
   ```

3. **Don't emit from unauthorized agents**
   ```go
   // ✗ Wrong - agent not in AllowedAgents
   agent.Emit("[RESTRICTED_SIGNAL]")
   ```

4. **Don't create unregistered signals**
   ```go
   // ✗ Wrong - signal invented on the fly
   if response.Contains("[MAYBE]") { ... }
   ```

5. **Don't assume signal existence**
   ```go
   // ✗ Wrong - should validate first
   definition := registry.Get(signal)
   // Use definition without checking nil
   ```

---

## 10. Testing

### Test Signal Registration

```go
func TestSignalRegistry(t *testing.T) {
    registry := NewSignalRegistry()
    def := &SignalDefinition{
        Name: "[TEST]",
        Behavior: SignalBehaviorRoute,
    }
    err := registry.Register(def)
    assert.NoError(t, err)
    assert.True(t, registry.Exists("[TEST]"))
}
```

### Test Signal Validation

```go
func TestSignalValidation(t *testing.T) {
    validator := NewSignalValidator(registry)
    err := validator.ValidateSignalEmission("[END]", "agent1")
    assert.NoError(t, err)
}
```

### Test Signal Matching

```go
func TestSignalMatching(t *testing.T) {
    matches, method := validator.ValidateSignalInContent(
        "[END]",
        "Task done. [END]",
    )
    assert.True(t, matches)
    assert.Equal(t, "exact", method)
}
```

---

## 11. Deprecation

### Marking Signals Deprecated

```go
definition := &SignalDefinition{
    Name: "[OLD_SIGNAL]",
    Description: "...",
    DeprecatedMsg: "Use [NEW_SIGNAL] instead",
}
```

### Migration Path

```
[OLD_SIGNAL] → [NEW_SIGNAL]
1. Mark [OLD_SIGNAL] as deprecated
2. Add warnings in logs
3. Support both during transition
4. Remove old signal after migration
```

---

## 12. Monitoring

### Signal Event Logging

```
[SIGNAL-EVENT] Agent 'teacher' emitted '[END_EXAM]' (exact) - Terminates exam workflow
[SIGNAL-SUCCESS] Agent 'teacher' routed to 'reporter' via signal [QUESTION]
[SIGNAL-ERROR] Agent 'teacher' emitted signal [UNKNOWN] targeting unknown agent 'agent_x'
```

### Signal Statistics

```go
validator.GenerateSignalReport()
// Output:
// === SIGNAL REGISTRY REPORT ===
// Total Signals Registered: 12
// --- route Signals (5) ---
// --- terminate Signals (3) ---
```

---

## 13. Migration Guide

### From Phase 2 to Phase 3

**Phase 2** (Validation + Logging):
- ValidateSignals() checks format and targets
- Comprehensive logging of signal decisions

**Phase 3** (Registry + Protocol):
- All signals must be registered
- SignalRegistry manages definitions
- SignalValidator enforces rules
- SIGNAL_PROTOCOL.md documents standard

**Migration**:
```go
// Old code (still works)
executor := NewCrewExecutorFromConfig(...)
executor.ValidateSignals()

// New code (with registry)
registry := LoadDefaultSignals()
registry.Register(customSignal)
validator := NewSignalValidator(registry)
validator.ValidateConfiguration(signals, agents)
```

---

## Appendix A: Signal Checklist

Before deploying a signal:

- [ ] Signal registered in registry
- [ ] Format matches [NAME] standard
- [ ] Behavior type defined (route/terminate/pause/parallel)
- [ ] Agent permissions configured
- [ ] Target agents exist
- [ ] Example response documented
- [ ] Tests written
- [ ] Log messages checked
- [ ] Documentation updated
- [ ] No deprecated signals used

---

## Appendix B: Quick Reference

| Aspect | Details |
|--------|---------|
| Format | `[SIGNAL_NAME]` |
| Case | Case-sensitive |
| Termination Target | Empty string `""` |
| Routing Target | Agent ID or group name |
| Max Signals | Unlimited |
| Registration | Required |
| Validation | At startup + runtime |
| Logging | All events logged |
| Testing | Comprehensive tests required |

---

**End of Signal Protocol Specification v1.0**
