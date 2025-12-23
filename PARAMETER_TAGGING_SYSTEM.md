# 📋 Parameter Tagging System for Agent Configuration

**Purpose:** Standardized tagging system to categorize and organize agent configuration parameters
**Status:** Design & Implementation Guide
**Date:** Dec 23, 2025

---

## 🎯 Core Principle

Parameters should be tagged by **semantic meaning**, not by development timeline.

❌ **BAD:** `# ✅ WEEK 1:` - Implementation timeline, not semantic
✅ **GOOD:** `# [QUOTA|COST]` - What it means and controls

---

## 📍 Tag Categories

### 1. Parameter Type Tags

| Tag | Meaning | Example |
|-----|---------|---------|
| `[QUOTA]` | Hard limit on resource usage | `max_tokens_per_call: 1000` |
| `[THRESHOLD]` | Alert/warning boundary | `cost_alert_threshold: 0.80` |
| `[FLAG]` | Boolean control/setting | `enforce_cost_limits: false` |
| `[METRIC]` | Measurement/tracking value | `current_memory_mb: 45` |
| `[CONFIG]` | Execution configuration | `temperature: 0.7` |

### 2. Resource Domain Tags

| Tag | Meaning |
|-----|---------|
| `[COST]` | Token/API cost control |
| `[MEMORY]` | Memory usage limits |
| `[ERROR]` | Error rate and limits |
| `[PERFORMANCE]` | Response time, throughput |
| `[BEHAVIOR]` | Agent behavior/personality |
| `[MODEL]` | LLM model selection |
| `[TOOL]` | Agent tool/capability |
| `[LOGGING]` | Observability settings |

### 3. Scope Tags

| Tag | Meaning |
|-----|---------|
| `[PER-CALL]` | Applied per API call |
| `[PER-DAY]` | Applied per 24-hour period |
| `[GLOBAL]` | Applied across all executions |
| `[RUNTIME]` | Updated during execution |

### 4. Data Type Tags (Optional)

| Tag | Meaning |
|-----|---------|
| `[INT]` | Integer value |
| `[FLOAT]` | Floating point value |
| `[BOOL]` | Boolean value |
| `[STRING]` | Text value |

---

## 📐 Tagging Format

### Standard Format
```
<parameter_name>: <value>  # [TAG1|TAG2] <description>
```

### Examples

```yaml
# Cost Control
max_tokens_per_call: 1000           # [QUOTA|COST|PER-CALL|INT] Max tokens per API call
max_tokens_per_day: 50000           # [QUOTA|COST|PER-DAY|INT] Max tokens per 24 hours
max_cost_per_day_usd: 10.0          # [QUOTA|COST|PER-DAY|FLOAT] Max USD cost per day
cost_alert_threshold: 0.80          # [THRESHOLD|COST|GLOBAL|FLOAT] Warn at 80% of limit
enforce_cost_limits: false          # [FLAG|COST|GLOBAL|BOOL] BLOCK vs WARN on limit

# Memory Control
max_memory_per_call_mb: 100         # [QUOTA|MEMORY|PER-CALL|INT] Max MB per call
max_memory_per_day_mb: 1000         # [QUOTA|MEMORY|PER-DAY|INT] Max MB per day
enforce_memory_limits: false        # [FLAG|MEMORY|GLOBAL|BOOL] BLOCK vs WARN on limit

# Error Control
max_consecutive_errors: 3           # [QUOTA|ERROR|PER-CALL|INT] Max consecutive failures
max_errors_per_day: 10              # [QUOTA|ERROR|PER-DAY|INT] Max errors per day
enforce_error_limits: false         # [FLAG|ERROR|GLOBAL|BOOL] BLOCK vs WARN on limit

# Behavior
temperature: 0.7                    # [CONFIG|BEHAVIOR|GLOBAL|FLOAT] LLM creativity (0.0-1.0)
is_terminal: true                   # [FLAG|BEHAVIOR|GLOBAL|BOOL] Is this a terminal agent?

# Logging
enable_memory_metrics: true         # [FLAG|LOGGING|GLOBAL|BOOL] Log memory after each call
enable_performance_metrics: true    # [FLAG|LOGGING|GLOBAL|BOOL] Log response metrics
log_level: "info"                   # [CONFIG|LOGGING|GLOBAL|STRING] debug/info/warn/error
```

---

## 🏗️ Hierarchical Organization in YAML

```yaml
# Group 1: Identity & Behavior
id: hello-agent                     # [CONFIG|BEHAVIOR|GLOBAL|STRING]
name: Hello Agent                   # [CONFIG|BEHAVIOR|GLOBAL|STRING]
role: Friendly Assistant            # [CONFIG|BEHAVIOR|GLOBAL|STRING]
temperature: 0.7                    # [CONFIG|BEHAVIOR|GLOBAL|FLOAT]
is_terminal: true                   # [FLAG|BEHAVIOR|GLOBAL|BOOL]

# Group 2: Model Configuration
primary:                            # [CONFIG|MODEL|GLOBAL]
  model: gemma3:1b
  provider: ollama
  provider_url: http://localhost:11434

backup:                             # [CONFIG|MODEL|GLOBAL]
  model: deepseek-r1:1.5b
  provider: ollama
  provider_url: http://localhost:11434

# Group 3: Resource Quotas
cost_limits:                        # [QUOTA|COST|GLOBAL]
  max_tokens_per_call: 1000         # [QUOTA|COST|PER-CALL|INT]
  max_tokens_per_day: 50000         # [QUOTA|COST|PER-DAY|INT]
  max_cost_per_day_usd: 10.0        # [QUOTA|COST|PER-DAY|FLOAT]
  alert_threshold: 0.80             # [THRESHOLD|COST|GLOBAL|FLOAT]
  enforce: false                    # [FLAG|COST|GLOBAL|BOOL]

memory_limits:                      # [QUOTA|MEMORY|GLOBAL]
  max_per_call_mb: 100              # [QUOTA|MEMORY|PER-CALL|INT]
  max_per_day_mb: 1000              # [QUOTA|MEMORY|PER-DAY|INT]
  enforce: false                    # [FLAG|MEMORY|GLOBAL|BOOL]

error_limits:                       # [QUOTA|ERROR|GLOBAL]
  max_consecutive: 3                # [QUOTA|ERROR|PER-CALL|INT]
  max_per_day: 10                   # [QUOTA|ERROR|PER-DAY|INT]
  enforce: false                    # [FLAG|ERROR|GLOBAL|BOOL]

# Group 4: Observability
logging:                            # [CONFIG|LOGGING|GLOBAL]
  enable_memory_metrics: true       # [FLAG|LOGGING|GLOBAL|BOOL]
  enable_performance_metrics: true  # [FLAG|LOGGING|GLOBAL|BOOL]
  enable_quota_warnings: true       # [FLAG|LOGGING|GLOBAL|BOOL]
  log_level: "info"                 # [CONFIG|LOGGING|GLOBAL|STRING]
```

---

## 💻 Code Tagging Convention

### Function Comments
```go
// UpdateMemoryMetrics updates memory usage metrics after a call
// [METRIC|MEMORY|RUNTIME] Tracks current, peak, average memory
func (a *Agent) UpdateMemoryMetrics(memoryUsedMB int, callDurationMs int64) {
    // ...
}

// CheckMemoryQuota checks if memory usage exceeds quota
// [QUOTA|MEMORY|ENFORCEMENT] Validates memory usage against limits
func (a *Agent) CheckMemoryQuota() error {
    // ...
}
```

### Variable Comments
```go
type AgentMemoryMetrics struct {
    CurrentMemoryMB   int     // [METRIC|MEMORY|RUNTIME|INT] Current memory in MB
    PeakMemoryMB      int     // [METRIC|MEMORY|RUNTIME|INT] Peak memory in MB
    MaxMemoryPerCall  int     // [QUOTA|MEMORY|PER-CALL|INT] Max allowed per call
    MaxMemoryPerDay   int     // [QUOTA|MEMORY|PER-DAY|INT] Max allowed per day
    EnforceQuotas     bool    // [FLAG|MEMORY|ENFORCEMENT|BOOL] BLOCK vs WARN
}
```

---

## 🔍 Tag Reference Quick Lookup

### By Type
- **[QUOTA]** - Resource limit boundaries
- **[THRESHOLD]** - Alert/warning levels
- **[FLAG]** - Boolean settings
- **[METRIC]** - Measured values
- **[CONFIG]** - Configuration values

### By Domain
- **[COST]** - Cost/token control
- **[MEMORY]** - Memory management
- **[ERROR]** - Error handling
- **[PERFORMANCE]** - Response/throughput
- **[BEHAVIOR]** - Agent personality
- **[MODEL]** - LLM selection
- **[LOGGING]** - Observability

### By Scope
- **[PER-CALL]** - Per API call
- **[PER-DAY]** - Per 24 hours
- **[GLOBAL]** - Always active
- **[RUNTIME]** - Live measurement

---

## ✅ Benefits of This System

1. **Self-Documenting** - Tags explain parameter purpose without needing external docs
2. **Searchable** - Easy to find all `[QUOTA]` or `[MEMORY]` parameters
3. **Timeline-Independent** - Not tied to development phases
4. **Scalable** - Works for 1 agent or 100 agents
5. **IDE-Friendly** - Can be parsed by IDE plugins for highlighting
6. **Consistency** - Same convention in YAML, Go code, and docs

---

## 🚀 Implementation Checklist

### Phase 1: YAML Config Files
- [ ] Remove all `WEEK 1`, `WEEK 2`, `WEEK 3` comments
- [ ] Add parameter tags to all fields in `hello-agent.yaml`
- [ ] Apply same tagging to all other agent configs
- [ ] Create `_template.yaml` with tagged structure

### Phase 2: Go Source Code
- [ ] Remove all `WEEK X:` comments from functions
- [ ] Add semantic tags to function comments
- [ ] Add tags to struct field comments
- [ ] Update metadata_logging.go with tagged output

### Phase 3: Documentation
- [ ] Create reference doc: "Parameter Tags Quick Reference"
- [ ] Add tagging examples to README
- [ ] Document tag combinations and their meanings
- [ ] Create troubleshooting guide by tag

---

## 📝 Example: Converting Old Comments to New Tags

### BEFORE (Timeline-based)
```yaml
# ✅ WEEK 1: Agent-level cost control configuration
# Set per-agent limits for token usage and cost
# Optional: All fields have sensible defaults if not specified

# Maximum tokens per API call (default: 1000 tokens)
max_tokens_per_call: 1000

# Maximum tokens per 24-hour period (default: 50,000 tokens/day)
max_tokens_per_day: 50000
```

### AFTER (Semantic-based with tags)
```yaml
cost_limits:
  max_tokens_per_call: 1000       # [QUOTA|COST|PER-CALL] Max tokens per API call
  max_tokens_per_day: 50000       # [QUOTA|COST|PER-DAY] Max tokens per 24 hours
```

**Improvements:**
- ❌ Removed timeline reference (`WEEK 1`)
- ✅ Clear semantic tags indicate purpose and scope
- ✅ Inline format reduces verbosity
- ✅ Information is at point of use, not separate comments

---

## 🎁 Benefits Over Timeline-Based Comments

| Aspect | Timeline-Based | Tag-Based |
|--------|---|---|
| **Readability** | Context-dependent | Self-contained |
| **Searchability** | Hard to find related params | Easy with tag search |
| **Maintenance** | Need to update phase docs | Tags never become stale |
| **Scalability** | Hard to organize 100+ params | Tag-based filtering works |
| **IDE Support** | Can't parse WEEK comments | Can parse and highlight tags |
| **Permanence** | Timeline becomes irrelevant | Tags always meaningful |

---

## ✨ Conclusion

**Remove WEEK labels → Use semantic parameter tags**

This creates a **timeless, self-documenting** configuration system that works for today and future versions.

---

**Status:** Ready for Implementation
**Next Steps:**
1. Apply tags to `hello-agent.yaml`
2. Apply tags to Go code
3. Remove all WEEK references
4. Test and verify functionality
