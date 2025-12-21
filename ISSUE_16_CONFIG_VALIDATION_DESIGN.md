# ✓ Issue #16: Configuration Validation - Design Document

**Date**: 2025-12-22
**Status**: DESIGN PHASE
**Priority**: MEDIUM (Score: 65/100)
**Effort**: LOW (1 day)

---

## 🎯 Objective

Implement comprehensive configuration validation system that:
- Detects invalid configurations at startup (fail-fast)
- Prevents circular reference loops
- Validates all required fields
- Checks for reachability (all configured agents must be accessible)
- Provides clear, actionable error messages
- Prevents runtime failures from bad config

---

## 📋 Current State Analysis

### What Exists

```go
// config.go has basic YAML loading
type CrewConfig struct {
    EntryPoint string
    Agents     []string
    Routing    RoutingConfig
    // ... other fields
}

// Loading is basic, no validation
config, err := LoadCrewConfig(path)
```

### Current Gaps

```
Missing Validation:
❌ No circular reference detection
❌ No reachability checks
❌ No required field validation
❌ No agent ID uniqueness check
❌ No signal target validation
❌ No tool availability validation
❌ No model name validation
❌ Poor error messages
```

### Validation Rules Needed

```
1. Crew Level:
   ✓ entry_point must exist in agents list
   ✓ agents list cannot be empty
   ✓ max_handoffs must be positive
   ✓ max_rounds must be positive

2. Agent Level:
   ✓ Agent ID must be unique
   ✓ Agent ID must match entry_point or routing targets
   ✓ Model must be valid (gpt-4o, gpt-4-turbo, etc.)
   ✓ Temperature must be 0-1
   ✓ Tools must exist and be registered

3. Routing Level:
   ✓ All signal targets must exist as agents
   ✓ No circular routing loops (A→B→C→A)
   ✓ Terminal agents cannot have outgoing signals
   ✓ Reachability from entry_point to all agents

4. Global:
   ✓ No duplicate agent IDs
   ✓ File syntax valid YAML
   ✓ All required files present
```

---

## 🏗️ Implementation Design

### 1. Validation Framework

```go
// ValidationError groups errors with context
type ValidationError struct {
    File      string // Which file
    Section   string // Which section
    Field     string // Which field
    Message   string // What's wrong
    Severity  string // "error" or "warning"
    Fix       string // How to fix it
}

// Validator orchestrates all checks
type ConfigValidator struct {
    config       *CrewConfig
    agents       map[string]*Agent
    errors       []ValidationError
    warnings     []ValidationError
}

// Validation methods
func (cv *ConfigValidator) ValidateAll() error
func (cv *ConfigValidator) ValidateCrewConfig() error
func (cv *ConfigValidator) ValidateAgents() error
func (cv *ConfigValidator) ValidateRouting() error
func (cv *ConfigValidator) DetectCircularReferences() error
func (cv *ConfigValidator) CheckReachability() error
func (cv *ConfigValidator) GenerateReport() string
```

### 2. Circular Reference Detection

**Algorithm**: Depth-First Search (DFS)

```
For each agent:
  1. Start DFS from agent
  2. Track visited nodes
  3. If we reach already-visited node: CIRCULAR!
  4. If we reach dead-end: OK (terminal agent)

Example:
  ✅ A→B→C→terminal: OK (linear)
  ❌ A→B→C→B: ERROR (circular)
  ❌ A→B→A: ERROR (direct loop)
```

### 3. Reachability Analysis

**Algorithm**: Graph Traversal

```
1. Build directed graph from routing config
2. Start from entry_point agent
3. DFS/BFS to all reachable agents
4. Any unreachable agent = ERROR

Example:
  Entry: orchestrator
  ✅ orchestrator→clarifier→executor: All reachable
  ❌ orchestrator→executor, but clarifier unreachable
```

### 4. Error Message Quality

**Bad**: `Error: invalid config`
**Good**:
```
❌ Config Validation Failed:

  File: config/crew.yaml
  Issue: entry_point agent not found

  Problem:
    entry_point: "dispatcher" (line 3)
    Available agents: ["orchestrator", "clarifier", "executor"]

  Solution:
    Change entry_point to one of: orchestrator, clarifier, executor

  Example:
    entry_point: orchestrator
```

### 5. Validation Stages

```
Stage 1: File Validation
  ├─ Files exist
  ├─ YAML syntax valid
  └─ Required fields present

Stage 2: Structure Validation
  ├─ Agent IDs unique
  ├─ Models valid
  └─ Temperature 0-1

Stage 3: Dependency Validation
  ├─ entry_point exists
  ├─ Signal targets exist
  └─ All agents accessible

Stage 4: Graph Validation
  ├─ No circular references
  ├─ No unreachable agents
  └─ Terminal agents correct

Stage 5: Tool Validation
  ├─ Tools exist and registered
  ├─ Tool parameters valid
  └─ No name conflicts
```

---

## 📝 Implementation Steps

### Step 1: Create validation.go (200+ lines)

```go
package crewai

type ValidationError struct {
    File     string
    Section  string
    Field    string
    Message  string
    Severity string
    Fix      string
}

type ConfigValidator struct {
    config   *CrewConfig
    agents   map[string]*Agent
    errors   []ValidationError
    warnings []ValidationError
}

func NewConfigValidator(config *CrewConfig, agents map[string]*Agent) *ConfigValidator {
    return &ConfigValidator{
        config:   config,
        agents:   agents,
        errors:   []ValidationError{},
        warnings: []ValidationError{},
    }
}

// ValidateAll runs all validation checks
func (cv *ConfigValidator) ValidateAll() error {
    // Stage 1: Basic structure
    // Stage 2: Field validation
    // Stage 3: Dependency validation
    // Stage 4: Graph validation

    if len(cv.errors) > 0 {
        return cv.GenerateErrorReport()
    }
    return nil
}

// Circular reference detection
func (cv *ConfigValidator) DetectCircularReferences() error {
    // DFS-based cycle detection
}

// Reachability analysis
func (cv *ConfigValidator) CheckReachability() error {
    // BFS from entry_point
}

// Error report generation
func (cv *ConfigValidator) GenerateReport() string {
    // Human-readable error output
}
```

### Step 2: Modify config.go (30+ lines)

Add validation call to LoadCrewConfig:

```go
func LoadCrewConfig(path string) (*CrewConfig, error) {
    // Existing loading code...
    config := &CrewConfig{}

    // Load YAML...

    // NEW: Validate configuration
    validator := NewConfigValidator(config, agents)
    if err := validator.ValidateAll(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }

    return config, nil
}
```

### Step 3: Add validation tests (200+ lines)

```go
func TestValidConfigAccepted(t *testing.T)
func TestMissingEntryPoint(t *testing.T)
func TestEntryPointNotFound(t *testing.T)
func TestCircularRouting(t *testing.T)
func TestUnreachableAgent(t *testing.T)
func TestDuplicateAgentID(t *testing.T)
func TestInvalidModel(t *testing.T)
func TestInvalidTemperature(t *testing.T)
func TestMissingTool(t *testing.T)
func TestErrorMessageQuality(t *testing.T)
```

### Step 4: Update CLI (20+ lines)

Add validation command:

```bash
$ go-agentic --validate config/
✓ Validating config/crew.yaml...
✓ Validating config/agents/orchestrator.yaml...
✓ Validating config/agents/clarifier.yaml...
✓ Validating config/agents/executor.yaml...
✓ Checking reachability...
✓ Checking circular references...
✓ All validations passed ✓

$ go-agentic --validate config/ (with errors)
✗ Configuration Validation Failed

  ❌ entry_point "dispatcher" not found
     Available: orchestrator, clarifier, executor
     Fix: Change entry_point to "orchestrator"
```

---

## 🧪 Test Coverage

### Test Cases (10+ tests)

1. **Valid Configuration**: Accepts valid config
2. **Missing entry_point**: Rejects missing entry_point
3. **entry_point Not Found**: Rejects non-existent entry_point
4. **Circular Routing**: Detects A→B→A loop
5. **Unreachable Agent**: Detects unreachable agents
6. **Duplicate Agent IDs**: Rejects duplicates
7. **Invalid Model**: Rejects unknown model names
8. **Invalid Temperature**: Rejects temperature > 1
9. **Missing Tool**: Rejects non-existent tool
10. **Error Message Quality**: Verifies helpful messages

### Test Examples

```go
// Valid configuration passes
func TestValidConfigAccepted(t *testing.T) {
    config := &CrewConfig{
        EntryPoint: "orchestrator",
        Agents: ["orchestrator", "clarifier", "executor"],
    }
    validator := NewConfigValidator(config, agents)
    err := validator.ValidateAll()
    if err != nil {
        t.Fatalf("Valid config rejected: %v", err)
    }
}

// Circular reference detected
func TestCircularRouting(t *testing.T) {
    config := &CrewConfig{
        // ... orchestrator→clarifier→orchestrator
    }
    validator := NewConfigValidator(config, agents)
    err := validator.ValidateAll()
    if !strings.Contains(err.Error(), "circular") {
        t.Fatal("Circular reference not detected")
    }
}

// Unreachable agent detected
func TestUnreachableAgent(t *testing.T) {
    config := &CrewConfig{
        EntryPoint: "orchestrator",
        Agents: ["orchestrator", "clarifier", "executor"],
        // But routing only connects orchestrator→clarifier
    }
    validator := NewConfigValidator(config, agents)
    err := validator.ValidateAll()
    if !strings.Contains(err.Error(), "unreachable") {
        t.Fatal("Unreachable agent not detected")
    }
}
```

---

## ✅ Acceptance Criteria

### Functional Requirements
- ✅ Detects missing entry_point
- ✅ Detects invalid entry_point (not in agents list)
- ✅ Detects circular routing loops (any depth)
- ✅ Detects unreachable agents
- ✅ Detects duplicate agent IDs
- ✅ Validates model names (gpt-4o, gpt-4-turbo)
- ✅ Validates temperature (0-1 range)
- ✅ Detects missing or invalid tools
- ✅ Provides clear error messages with fixes
- ✅ Fails at startup (not runtime)

### Quality Requirements
- ✅ All error messages < 200 chars and actionable
- ✅ Test coverage > 95%
- ✅ No race conditions
- ✅ < 50ms validation time
- ✅ Zero breaking changes

---

## 📊 Success Metrics

- ✅ Invalid config detected before deployment
- ✅ Clear error messages reduce debugging time
- ✅ All validation scenarios covered by tests
- ✅ Validation completes in < 50ms
- ✅ 100% of issues from bad config prevented

---

## 🎯 Implementation Checklist

- [ ] Create validation.go (200+ lines)
- [ ] Implement structure validation
- [ ] Implement circular reference detection
- [ ] Implement reachability analysis
- [ ] Create error messages
- [ ] Integrate into config loading
- [ ] Add validation command
- [ ] Write 10+ tests
- [ ] Test error message quality
- [ ] Documentation

---

**Status**: DESIGN COMPLETE
**Next**: Implementation of validation.go

---

*Design Date: 2025-12-22*
*Phase 3 Issue #16*
