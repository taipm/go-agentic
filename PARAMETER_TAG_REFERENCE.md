# 📖 Parameter Tag Reference Guide

**Quick lookup for all semantic parameter tags used in agent configuration**

---

## 🔍 Quick Tag Lookup

### By Tag Type

#### [QUOTA] - Resource Limits
Hard boundaries on resource usage. Execution blocked/warned when exceeded.
- `[QUOTA|COST|PER-CALL]` - Max tokens per API call
- `[QUOTA|COST|PER-DAY]` - Max tokens per 24 hours
- `[QUOTA|MEMORY|PER-CALL]` - Max memory per execution
- `[QUOTA|MEMORY|PER-DAY]` - Max memory per 24 hours
- `[QUOTA|ERROR|PER-CALL]` - Max consecutive failures
- `[QUOTA|ERROR|PER-DAY]` - Max errors per 24 hours

#### [THRESHOLD] - Alert Levels
Warning boundaries. Trigger alerts when crossed.
- `[THRESHOLD|COST|GLOBAL]` - Warn at % of daily limit
- `[THRESHOLD|PERFORMANCE]` - Alert on slow execution

#### [FLAG] - Boolean Controls
Enable/disable features or enforcement modes.
- `[FLAG|BEHAVIOR|BOOL]` - Agent behavior control
- `[FLAG|COST|ENFORCEMENT|BOOL]` - BLOCK vs WARN mode
- `[FLAG|MEMORY|ENFORCEMENT|BOOL]` - BLOCK vs WARN mode
- `[FLAG|ERROR|ENFORCEMENT|BOOL]` - BLOCK vs WARN mode
- `[FLAG|LOGGING|BOOL]` - Enable/disable logging

#### [METRIC] - Measurements
Track and report values.
- `[METRIC|MEMORY|RUNTIME]` - Current memory tracking
- `[METRIC|PERFORMANCE|RUNTIME]` - Success rate tracking
- `[METRIC|MEMORY]` - Query memory status
- `[METRIC|PERFORMANCE]` - Query performance status

#### [CONFIG] - Settings
Configuration values for behavior and features.
- `[CONFIG|BEHAVIOR|FLOAT]` - LLM temperature (0.0-1.0)
- `[CONFIG|MODEL]` - LLM provider selection
- `[CONFIG|LOGGING|STRING]` - Log level setting

---

## 🎯 By Domain

### Cost Control [COST]
Token usage and API cost management.

```yaml
cost_limits:
  max_tokens_per_call: 1000     # [QUOTA|COST|PER-CALL|INT]
  max_tokens_per_day: 50000     # [QUOTA|COST|PER-DAY|INT]
  max_cost_per_day_usd: 10.0    # [QUOTA|COST|PER-DAY|FLOAT]
  alert_threshold: 0.80          # [THRESHOLD|COST|GLOBAL|FLOAT]
  enforce: false                 # [FLAG|COST|ENFORCEMENT|BOOL]
```

### Memory Management [MEMORY]
Memory usage limits and tracking.

```yaml
memory_limits:
  max_per_call_mb: 100          # [QUOTA|MEMORY|PER-CALL|INT]
  max_per_day_mb: 1000          # [QUOTA|MEMORY|PER-DAY|INT]
  enforce: false                # [FLAG|MEMORY|ENFORCEMENT|BOOL]
```

### Error Handling [ERROR]
Error rate and failure limits.

```yaml
error_limits:
  max_consecutive: 3            # [QUOTA|ERROR|PER-CALL|INT]
  max_per_day: 10               # [QUOTA|ERROR|PER-DAY|INT]
  enforce: false                # [FLAG|ERROR|ENFORCEMENT|BOOL]
```

### Observability [LOGGING]
Logging and metrics collection.

```yaml
logging:
  enable_memory_metrics: true    # [FLAG|LOGGING|BOOL]
  enable_performance_metrics: true # [FLAG|LOGGING|BOOL]
  enable_quota_warnings: true    # [FLAG|LOGGING|BOOL]
  log_level: "info"             # [CONFIG|LOGGING|STRING]
```

### Agent Behavior [BEHAVIOR]
Personality and execution characteristics.

```yaml
temperature: 0.7               # [CONFIG|BEHAVIOR|FLOAT] Creativity
is_terminal: true              # [FLAG|BEHAVIOR|BOOL] Terminal agent
system_prompt: |               # [CONFIG|BEHAVIOR] Instructions
  ...
```

### Model Configuration [MODEL]
LLM provider and model selection.

```yaml
primary:                       # [CONFIG|MODEL] Primary provider
  model: gemma3:1b
  provider: ollama
  provider_url: http://localhost:11434

backup:                        # [CONFIG|MODEL] Fallback provider
  model: deepseek-r1:1.5b
  provider: ollama
  provider_url: http://localhost:11434
```

---

## 📏 By Scope

### [PER-CALL] - Per API Execution
Limits applied to each individual API call.
- `[QUOTA|COST|PER-CALL|INT]` - Max tokens per call
- `[QUOTA|MEMORY|PER-CALL|INT]` - Max memory per call
- `[QUOTA|ERROR|PER-CALL|INT]` - Max consecutive errors

### [PER-DAY] - Per 24-Hour Period
Limits applied within each day (resets at midnight).
- `[QUOTA|COST|PER-DAY|INT]` - Max tokens per day
- `[QUOTA|MEMORY|PER-DAY|INT]` - Max memory per day
- `[QUOTA|ERROR|PER-DAY|INT]` - Max errors per day

### [GLOBAL] - Always Active
Settings that apply across all calls and days.
- `[THRESHOLD|COST|GLOBAL]` - Alert threshold
- `[CONFIG|BEHAVIOR|FLOAT]` - Temperature
- `[CONFIG|LOGGING]` - Logging settings

### [RUNTIME] - Live Measurement
Values tracked during execution.
- `[METRIC|MEMORY|RUNTIME]` - Current memory
- `[METRIC|PERFORMANCE|RUNTIME]` - Success rate
- `[THRESHOLD|PERFORMANCE]` - Slow call alert

---

## 📝 By Data Type

### [INT] - Integer Values
Whole numbers, typically counts or sizes.
- `max_tokens_per_call: 1000`
- `max_per_call_mb: 100`
- `max_consecutive: 3`

### [FLOAT] - Floating Point Values
Decimal numbers, typically percentages or costs.
- `max_cost_per_day_usd: 10.0`
- `alert_threshold: 0.80`
- `temperature: 0.7`

### [BOOL] - Boolean Values
True/false flags.
- `enforce: false`
- `is_terminal: true`
- `enable_memory_metrics: true`

### [STRING] - Text Values
String configuration values.
- `log_level: "info"`
- `model: "gemma3:1b"`
- `provider: "ollama"`

---

## 🔗 Tag Combinations

### Common Patterns

```
[QUOTA|DOMAIN|SCOPE|TYPE]
├─ [QUOTA|COST|PER-CALL|INT]
├─ [QUOTA|MEMORY|PER-DAY|INT]
└─ [QUOTA|ERROR|PER-CALL|INT]

[THRESHOLD|DOMAIN|SCOPE]
├─ [THRESHOLD|COST|GLOBAL|FLOAT]
└─ [THRESHOLD|PERFORMANCE]

[FLAG|DOMAIN|ENFORCEMENT|TYPE]
├─ [FLAG|COST|ENFORCEMENT|BOOL]
├─ [FLAG|MEMORY|ENFORCEMENT|BOOL]
└─ [FLAG|ERROR|ENFORCEMENT|BOOL]

[METRIC|DOMAIN|SCOPE]
├─ [METRIC|MEMORY|RUNTIME]
├─ [METRIC|PERFORMANCE|RUNTIME]
└─ [METRIC|MEMORY]

[CONFIG|DOMAIN|TYPE]
├─ [CONFIG|BEHAVIOR|FLOAT]
├─ [CONFIG|LOGGING|STRING]
└─ [CONFIG|MODEL]
```

---

## 📊 Complete Parameter Matrix

| Parameter | Tag | Domain | Scope | Type |
|-----------|-----|--------|-------|------|
| `max_tokens_per_call` | `[QUOTA\|COST\|PER-CALL\|INT]` | COST | PER-CALL | INT |
| `max_tokens_per_day` | `[QUOTA\|COST\|PER-DAY\|INT]` | COST | PER-DAY | INT |
| `max_cost_per_day_usd` | `[QUOTA\|COST\|PER-DAY\|FLOAT]` | COST | PER-DAY | FLOAT |
| `alert_threshold` | `[THRESHOLD\|COST\|GLOBAL\|FLOAT]` | COST | GLOBAL | FLOAT |
| `cost.enforce` | `[FLAG\|COST\|ENFORCEMENT\|BOOL]` | COST | GLOBAL | BOOL |
| `max_per_call_mb` | `[QUOTA\|MEMORY\|PER-CALL\|INT]` | MEMORY | PER-CALL | INT |
| `max_per_day_mb` | `[QUOTA\|MEMORY\|PER-DAY\|INT]` | MEMORY | PER-DAY | INT |
| `memory.enforce` | `[FLAG\|MEMORY\|ENFORCEMENT\|BOOL]` | MEMORY | GLOBAL | BOOL |
| `max_consecutive` | `[QUOTA\|ERROR\|PER-CALL\|INT]` | ERROR | PER-CALL | INT |
| `max_errors_per_day` | `[QUOTA\|ERROR\|PER-DAY\|INT]` | ERROR | PER-DAY | INT |
| `error.enforce` | `[FLAG\|ERROR\|ENFORCEMENT\|BOOL]` | ERROR | GLOBAL | BOOL |
| `enable_memory_metrics` | `[FLAG\|LOGGING\|BOOL]` | LOGGING | GLOBAL | BOOL |
| `enable_performance_metrics` | `[FLAG\|LOGGING\|BOOL]` | LOGGING | GLOBAL | BOOL |
| `enable_quota_warnings` | `[FLAG\|LOGGING\|BOOL]` | LOGGING | GLOBAL | BOOL |
| `log_level` | `[CONFIG\|LOGGING\|STRING]` | LOGGING | GLOBAL | STRING |
| `temperature` | `[CONFIG\|BEHAVIOR\|FLOAT]` | BEHAVIOR | GLOBAL | FLOAT |
| `is_terminal` | `[FLAG\|BEHAVIOR\|BOOL]` | BEHAVIOR | GLOBAL | BOOL |
| `model` | `[CONFIG\|MODEL]` | MODEL | GLOBAL | STRING |
| `provider` | `[CONFIG\|MODEL]` | MODEL | GLOBAL | STRING |

---

## 🎓 Tag Grammar

### Format Structure
```
[TYPE|DOMAIN|SCOPE|DATATYPE]
 │      │      │     └─ Optional: INT, FLOAT, BOOL, STRING
 │      │      └─ Optional: PER-CALL, PER-DAY, GLOBAL, RUNTIME
 │      └─ Domain: COST, MEMORY, ERROR, PERFORMANCE, BEHAVIOR, MODEL, LOGGING
 └─ Type: QUOTA, THRESHOLD, FLAG, METRIC, CONFIG
```

### Examples

**Minimal (Type + Domain):**
- `[QUOTA|COST]` - Cost quota (unspecified scope/type)
- `[CONFIG|LOGGING]` - Logging config (unspecified type)

**Standard (Type + Domain + Scope):**
- `[QUOTA|COST|PER-CALL]` - Cost quota per call
- `[FLAG|MEMORY|ENFORCEMENT]` - Memory enforcement flag

**Complete (Type + Domain + Scope + DataType):**
- `[QUOTA|COST|PER-CALL|INT]` - Cost quota per call (integer)
- `[THRESHOLD|COST|GLOBAL|FLOAT]` - Cost threshold (percentage)

---

## 📚 Usage Examples

### In YAML Config
```yaml
# Tagged parameter
max_tokens_per_call: 1000  # [QUOTA|COST|PER-CALL|INT] Max tokens per call

# Tagged section
cost_limits:               # [QUOTA|COST] Cost control configuration
  max_tokens_per_call: 1000
  enforce: false
```

### In Go Code
```go
// Tagged function
// [QUOTA|MEMORY|ENFORCEMENT] Validates memory usage against limits
func (a *Agent) CheckMemoryQuota() error {
    // ...
}

// Tagged struct field
type MemoryLimitsConfig struct {
    MaxPerCallMB int  // [QUOTA|MEMORY|PER-CALL|INT] Max MB per execution
    Enforce      bool // [FLAG|MEMORY|ENFORCEMENT|BOOL] BLOCK vs WARN
}
```

---

## 🔍 Finding Parameters by Purpose

**I want to limit...**
- Tokens per API call → `[QUOTA|COST|PER-CALL]`
- Total cost per day → `[QUOTA|COST|PER-DAY]`
- Memory per execution → `[QUOTA|MEMORY|PER-CALL]`
- Consecutive errors → `[QUOTA|ERROR|PER-CALL]`

**I want to track...**
- Memory usage → `[METRIC|MEMORY|RUNTIME]`
- Success rate → `[METRIC|PERFORMANCE|RUNTIME]`

**I want to enable/disable...**
- Memory logging → `[FLAG|LOGGING|BOOL]`
- Quota enforcement → `[FLAG|*|ENFORCEMENT|BOOL]`

**I want to set threshold...**
- Cost alert level → `[THRESHOLD|COST|GLOBAL]`
- Slow call duration → `[THRESHOLD|PERFORMANCE]`

**I want to configure...**
- Agent creativity → `[CONFIG|BEHAVIOR|FLOAT]`
- LLM model → `[CONFIG|MODEL]`
- Log level → `[CONFIG|LOGGING|STRING]`

---

## 📝 Notes

- Tags are **optional** in code (for readability) but recommended
- Tags in YAML comments are **best practice**
- Tags enable IDE plugins, linting, and doc generation
- Tag system is **extensible** - can add new categories as needed
- Tags are **semantic** - independent of development timeline

---

**Last Updated:** Dec 23, 2025
**Status:** Complete Reference Guide
