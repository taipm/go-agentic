# 📋 Agent YAML Configuration Structure Analysis

**File:** `examples/00-hello-crew/config/agents/hello-agent.yaml`
**Date:** Dec 23, 2025
**Status:** Analysis & Recommendations Complete

---

## 🔍 Current Structure Overview

### Current Organization (Lines 1-91)

```
1. Identity Fields (id, name, role, description, backstory)
2. Behavior Fields (temperature, is_terminal)
3. Model Configuration (primary, backup, deprecated comments)
4. Tools (empty array)
5. Cost Control (WEEK 1 - 6 fields)
6. System Prompt (WEEK 3 - detailed instructions)
```

**Total Sections:** 6
**Total Lines:** 91
**Readability Score:** ⚠️ Medium (mixed grouping logic)

---

## ⚠️ Identified Structural Issues

### Issue 1: Identity vs. Behavior Separation
**Problem:** Agent identity (id, name, role, description, backstory) is mixed with behavior controls (temperature, is_terminal) on lines 1-11.
- Lines 1-4: Identity metadata
- Lines 6-8: Narrative backstory
- Lines 10-11: Behavior flags (temperature, is_terminal)

**Impact:** Readers must jump between different logical sections to understand agent setup.

---

### Issue 2: Configuration Scattered Across File
**Problem:** Cost control fields (lines 35-56) are separated from model configuration (lines 13-31) by tools array and comments.

**Current Flow:**
```
Model Config (lines 15-25)
↓
Tools (line 33)
↓
Cost Control (lines 35-56)
```

**Impact:** Related configuration scattered; hard to find all "runtime" settings in one place.

---

### Issue 3: Inconsistent Comment Styles
**Problem:** Multiple comment patterns used:
- Line 13-14: `# NEW:` comment blocks with detailed explanation
- Line 20-21: `# Optional` with reasoning
- Line 27-28: `# DEPRECATED:` with multi-line explanation
- Line 35-37: `# ✅ WEEK 1:` with explicit week marking
- Line 39, 42, 45, etc.: `#` with parameter descriptions

**Impact:** No consistent documentation pattern; hard to know what style to follow for new fields.

---

### Issue 4: Week Markers In Comments
**Problem:** Comments reference "WEEK 1" (line 35) and "WEEK 3" (not visible in cost section) but no clear phase organization.

**Impact:** Developers need external knowledge to understand which settings belong to which development phase.

---

### Issue 5: System Prompt Placement
**Problem:** System prompt (lines 58-91) is last in file, yet it's critical for agent behavior.

**Current Priority Order:**
1. Identity (high importance)
2. Model Config (high importance) ← Should appear earlier
3. Cost Control (medium importance)
4. System Prompt (highest importance!) ← Should appear earlier

**Impact:** Most critical section appears last; visual hierarchy inverted.

---

### Issue 6: Missing Sections for WEEK 2/3 Features
**Problem:** File has cost control (WEEK 1) and system prompt (WEEK 3), but no explicit section for:
- Memory metrics configuration (WEEK 3)
- Performance metrics configuration (WEEK 3)
- Quota enforcement settings (WEEK 3)

**Impact:** Unclear how to configure newly added features; no designated section.

---

## ✅ Recommended Structure

### Proposed Organization (6 Logical Sections)

```yaml
# SECTION 1: AGENT IDENTITY (Lines 1-9)
# What is this agent? Basic metadata for identification
id:
name:
role:
description:
backstory:

# SECTION 2: MODEL & EXECUTION CONFIGURATION (Lines 10-20)
# How should the agent behave? Core execution settings
temperature:
is_terminal:
primary:
  model:
  provider:
  provider_url:
backup:
  model:
  provider:
  provider_url:

# SECTION 3: TOOLS & CAPABILITIES (Lines 21-22)
# What can this agent do? Available functions/tools
tools:

# SECTION 4: SYSTEM PROMPT & INSTRUCTIONS (Lines 23-60)
# How should the agent think? Detailed behavior guidance
system_prompt:

# SECTION 5: RESOURCE & QUOTA CONTROLS (Lines 61-75)
# Limits and guardrails. Cost, memory, performance thresholds
max_tokens_per_call:
max_tokens_per_day:
max_cost_per_day:
cost_alert_threshold:
enforce_cost_limits:
max_memory_per_call_mb:        # NEW - WEEK 3
max_memory_per_day_mb:         # NEW - WEEK 3
enforce_memory_limits:         # NEW - WEEK 3
max_consecutive_errors:        # NEW - WEEK 3
max_errors_per_day:            # NEW - WEEK 3
enforce_error_limits:          # NEW - WEEK 3

# SECTION 6: LOGGING & MONITORING (Lines 76-85)
# What should be logged? Observability settings
enable_memory_logging:         # NEW
enable_performance_logging:    # NEW
enable_quota_logging:          # NEW
log_level:                     # NEW
```

**Benefits:**
- ✅ Logical flow: Identity → Execution → Capabilities → Behavior → Controls → Monitoring
- ✅ Related fields grouped together
- ✅ Most important section (System Prompt) moved higher
- ✅ Space reserved for WEEK 2/3 features
- ✅ Clear section separators with comments

---

## 📊 Structure Comparison: Current vs. Recommended

### Current Structure (91 lines)
```
IDENTITY        Lines 1-8    (8 lines)  ✅ Clear
BEHAVIOR        Lines 10-11  (2 lines)  ✅ Clear
MODELS          Lines 13-31  (19 lines) ⚠️ With deprecated comments
TOOLS           Line 33      (1 line)   ✅ Clear
COST CONTROL    Lines 35-56  (22 lines) ⚠️ Verbose comments
SYSTEM PROMPT   Lines 58-91  (34 lines) ✅ Clear but placed last
```

**Issues:** Mixed ordering, cost control before system prompt, scattered WEEK labels

---

### Recommended Structure (Reorganized)
```
IDENTITY            (9 lines)   ✅ First - agent identification
MODEL & EXECUTION   (12 lines)  ✅ Early - how agent runs
TOOLS              (2 lines)   ✅ What agent can do
SYSTEM PROMPT      (35 lines)  ✅ MOVED HIGHER - most critical
RESOURCE CONTROLS  (20 lines)  ✅ Grouped together
MONITORING         (10 lines)  ✅ Observability settings
```

**Benefits:** Logical flow, critical section higher, WEEK features grouped, cleaner comments

---

## 🎯 Specific Recommendations

### Recommendation 1: Create Clear Section Separators

**Current:**
```yaml
temperature: 0.7
is_terminal: true

# NEW: Primary/Backup LLM model configuration
primary:
```

**Recommended:**
```yaml
# ─────────────────────────────────────────────────────────
# SECTION 2: MODEL & EXECUTION CONFIGURATION
# How should the agent behave? Core execution settings
# ─────────────────────────────────────────────────────────

temperature: 0.7
is_terminal: true

# Model Configuration: Primary with automatic fallback to backup
primary:
```

**Why:** Clear section boundaries help readers understand document structure at a glance.

---

### Recommendation 2: Consolidate Cost Control Comments

**Current (Lines 35-56):**
```yaml
# ✅ WEEK 1: Agent-level cost control configuration
# Set per-agent limits for token usage and cost
# Optional: All fields have sensible defaults if not specified

# Maximum tokens per API call (default: 1000 tokens)
max_tokens_per_call: 1000

# Maximum tokens per 24-hour period (default: 50,000 tokens/day)
max_tokens_per_day: 50000

# Maximum cost per 24-hour period in USD (default: $10/day)
max_cost_per_day: 10.0

# Alert threshold: warn when usage exceeds this % of daily limit (default: 0.80 = 80%)
# Range: 0.0 to 1.0
# E.g., 0.80 = warn when $8 spent out of $10 daily limit
cost_alert_threshold: 0.80

# Enforcement mode (default: false = warn only)
#   true  = BLOCK execution if limit exceeded (strict budget control)
#   false = WARN only (log warning but execute anyway)
enforce_cost_limits: false
```

**Recommended:**
```yaml
# ─────────────────────────────────────────────────────────
# SECTION 5: RESOURCE & QUOTA CONTROLS
# Limits and guardrails for token usage, cost, errors, memory
# All fields have sensible defaults if not specified
# ─────────────────────────────────────────────────────────

# Cost Control (WEEK 1 Feature)
cost_limits:
  max_tokens_per_call: 1000          # default: 1000 tokens
  max_tokens_per_day: 50000          # default: 50,000 tokens/day
  max_cost_per_day_usd: 10.0         # default: $10/day
  alert_threshold: 0.80              # warn at 80% of daily limit
  enforce: false                     # true=BLOCK, false=WARN ONLY

# Memory Control (WEEK 3 Feature)
memory_limits:
  max_per_call_mb: 100               # default: 100 MB
  max_per_day_mb: 1000               # default: 1000 MB/day
  enforce: false                     # true=BLOCK, false=WARN ONLY

# Error Control (WEEK 3 Feature)
error_limits:
  max_consecutive: 3                 # default: 3 consecutive errors
  max_per_day: 10                    # default: 10 errors/day
  enforce: false                     # true=BLOCK, false=WARN ONLY
```

**Why:**
- Consolidated under one section
- Grouped by feature type
- Cleaner syntax with nested keys
- WEEK labels clearly show phasing
- Shorter inline comments (key: value # note format)

---

### Recommendation 3: Add Monitoring Section

**New Section to Add:**
```yaml
# ─────────────────────────────────────────────────────────
# SECTION 6: LOGGING & MONITORING
# Observability and debugging settings
# ─────────────────────────────────────────────────────────

logging:
  enable_memory_metrics: true        # Log memory usage after each execution
  enable_performance_metrics: true   # Log response time and success rate
  enable_quota_warnings: true        # Log quota threshold alerts
  log_level: "info"                  # debug, info, warn, error
```

**Why:**
- Dedicated section for observability
- Easy to enable/disable all logging at once
- Supports future monitoring features
- Clear what information gets logged

---

### Recommendation 4: Improve System Prompt Documentation

**Current (Line 58):**
```yaml
system_prompt: |
  You are {{name}}.
  ...
```

**Recommended:**
```yaml
# System Prompt: Defines agent personality, instructions, and behavior guidelines
# Template variables available: {{name}}, {{role}}, {{description}}, {{backstory}}
# This section is CRITICAL for agent behavior - update carefully
system_prompt: |
  You are {{name}}.
  ...
```

**Why:**
- Explains what system prompt is for
- Documents available template variables
- Warns about importance
- Makes the section more discoverable

---

## 📝 Complete Recommended YAML Structure

See below for the full reorganized file:

```yaml
# ─────────────────────────────────────────────────────────
# SECTION 1: AGENT IDENTITY
# What is this agent? Basic metadata for identification
# ─────────────────────────────────────────────────────────

id: hello-agent
name: Hello Agent
role: Friendly Assistant
description: A simple and friendly assistant that greets users and provides helpful responses

backstory: |
  You are a warm and welcoming assistant. Your role is to greet users, understand their needs,
  and provide helpful, friendly responses. You keep your answers concise and friendly.

# ─────────────────────────────────────────────────────────
# SECTION 2: MODEL & EXECUTION CONFIGURATION
# How should the agent behave? Core execution settings
# ─────────────────────────────────────────────────────────

temperature: 0.7                     # 0.0=deterministic, 1.0=creative
is_terminal: true                    # Is this agent a terminal node? (ends workflow)

# Model Configuration: Primary with automatic fallback to backup
# If primary fails, system automatically tries backup model
primary:
  model: gemma3:1b
  provider: ollama
  provider_url: http://localhost:11434

backup:
  model: deepseek-r1:1.5b
  provider: ollama
  provider_url: http://localhost:11434

# ─────────────────────────────────────────────────────────
# SECTION 3: TOOLS & CAPABILITIES
# What can this agent do? Available functions/tools
# ─────────────────────────────────────────────────────────

tools: []

# ─────────────────────────────────────────────────────────
# SECTION 4: SYSTEM PROMPT & INSTRUCTIONS
# How should the agent think? Detailed behavior guidance
# CRITICAL: This defines agent personality and behavior
# Template variables: {{name}}, {{role}}, {{description}}, {{backstory}}
# ─────────────────────────────────────────────────────────

system_prompt: |
  You are {{name}}.
  Role: {{role}}
  Description: {{description}}

  Backstory: {{backstory}}

  CRITICAL INSTRUCTIONS FOR MEMORY AND CONVERSATION:

  1. LISTEN AND REMEMBER:
     - Pay CLOSE attention to what the user tells you about themselves
     - Store their name, preferences, interests, and personal details
     - Use this information in ALL future responses

  2. WHEN USER MENTIONS THEIR NAME:
     - ALWAYS acknowledge it directly: "Got it, your name is [name]!"
     - Use their actual name in future messages
     - Never ask "What's your name?" if they already told you

  3. WHEN ASKED "What's my name?" or similar:
     - If you know their name from earlier: Answer directly with their name
     - Example: User says "Tôi tên gì?" → You say "Your name is Phan Minh Tài"
     - Do NOT say "I don't know" if they already told you

  4. CONVERSATION MEMORY:
     - Refer back to previous information: "As you mentioned before..."
     - Keep context of the entire conversation
     - Build on what the user shared earlier

  5. TONE & STYLE:
     - Be friendly, warm, and engaging
     - Be concise but thorough
     - Always use the user's name when appropriate

# ─────────────────────────────────────────────────────────
# SECTION 5: RESOURCE & QUOTA CONTROLS
# Limits and guardrails for token usage, cost, errors, memory
# All fields have sensible defaults if not specified
# ─────────────────────────────────────────────────────────

cost_limits:
  max_tokens_per_call: 1000          # Max tokens per API call (default: 1000)
  max_tokens_per_day: 50000          # Max tokens per 24h (default: 50,000)
  max_cost_per_day_usd: 10.0         # Max USD cost per 24h (default: $10)
  alert_threshold: 0.80              # Warn when usage exceeds 80% of limit
  enforce: false                     # true=BLOCK if exceeded, false=WARN ONLY

memory_limits:                        # WEEK 3: Memory quota enforcement
  max_per_call_mb: 100               # Max memory per execution (default: 100 MB)
  max_per_day_mb: 1000               # Max memory per 24h (default: 1000 MB/day)
  enforce: false                     # true=BLOCK if exceeded, false=WARN ONLY

error_limits:                         # WEEK 3: Error rate enforcement
  max_consecutive: 3                 # Max consecutive errors (default: 3)
  max_per_day: 10                    # Max errors per 24h (default: 10)
  enforce: false                     # true=BLOCK if exceeded, false=WARN ONLY

# ─────────────────────────────────────────────────────────
# SECTION 6: LOGGING & MONITORING
# Observability and debugging settings
# ─────────────────────────────────────────────────────────

logging:
  enable_memory_metrics: true        # Log memory usage after each execution
  enable_performance_metrics: true   # Log response time and success rate
  enable_quota_warnings: true        # Log quota threshold alerts
  log_level: "info"                  # debug, info, warn, error
```

---

## 🎁 Benefits of Recommended Structure

### 1. **Clear Mental Model**
```
Section 1: WHAT IS THIS AGENT?
Section 2: HOW DOES IT BEHAVE?
Section 3: WHAT CAN IT DO?
Section 4: HOW SHOULD IT THINK?
Section 5: WHAT ARE ITS LIMITS?
Section 6: WHAT SHOULD WE LOG?
```

### 2. **Easier Maintenance**
- Adding new WEEK features: Add to appropriate section (Section 5 for quotas, Section 6 for logging)
- Finding settings: Know which section to look in
- Consistent format: All sections follow same pattern

### 3. **Better Scalability**
- **Small crew (1 agent):** Simple config, easy to understand
- **Medium crew (5-10 agents):** Can copy template, consistent across all
- **Large crew (50+ agents):** Standard structure makes automation easier

### 4. **Documentation Ready**
- Each section can be documented separately
- Section headers explain purpose
- Inline comments explain individual fields
- Easy to generate reference docs from structure

### 5. **Hierarchical Organization**
```
hello-agent.yaml
├─ Identity (WHO)
├─ Execution (HOW)
├─ Capabilities (WHAT)
├─ Behavior (THINK)
├─ Controls (LIMITS)
└─ Monitoring (LOG)
```

---

## 🔄 Migration Path

### Step 1: Backup Current File
```bash
cp examples/00-hello-crew/config/agents/hello-agent.yaml \
   examples/00-hello-crew/config/agents/hello-agent.yaml.backup
```

### Step 2: Apply Recommended Structure
Replace current file with reorganized version (see complete YAML above)

### Step 3: Verify Functionality
```bash
cd examples/00-hello-crew
go run main.go
# Test that agent still works identically
```

### Step 4: Update Other Agent Configs
Apply same structure to all agents in `config/agents/*.yaml`

### Step 5: Create Structure Template
Save as `config/agents/_template.yaml` for new agents

---

## 📋 Summary of Changes

| Aspect | Current | Recommended | Benefit |
|--------|---------|-------------|---------|
| **Sections** | 6 (scattered) | 6 (logical) | Better organization |
| **System Prompt Position** | Last (line 58) | 4th (middle) | Higher visibility |
| **Cost Comments** | 22 lines of comments | 5 lines + nested keys | More concise |
| **WEEK Labels** | In comments only | In section headers | More discoverable |
| **Monitoring** | Missing | New section | Enables observability |
| **Memory Control** | Missing | New section | Supports WEEK 3 |
| **Error Control** | Missing | New section | Supports WEEK 3 |
| **Section Clarity** | Implicit | Explicit headers | Self-documenting |

---

## ✅ Recommendations Approved For Implementation?

The recommended structure provides:
1. ✅ **Clearer organization** - logical grouping of related fields
2. ✅ **Better readability** - section headers explain purpose
3. ✅ **More standard** - follows YAML best practices
4. ✅ **Scalable** - easy to add new features (WEEK 4+)
5. ✅ **Well-documented** - inline comments for clarity
6. ✅ **Future-proof** - space for upcoming metrics and monitoring

Would you like me to:
1. **Apply these changes** to hello-agent.yaml now?
2. **Create a comparison document** showing current vs. new side-by-side?
3. **Generate template files** for creating new agents with recommended structure?
4. **Update other agent configs** to match this structure?

---

**Generated:** Dec 23, 2025
**Status:** Analysis Complete - Ready for Implementation
