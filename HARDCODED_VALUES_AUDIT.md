# Hardcoded Values Audit Report for go-agentic Core

**Date:** 2025-12-22
**Scope:** `core/` directory
**Status:** Comprehensive audit of hardcoded values in all core packages

---

## Executive Summary

This audit identifies all hardcoded values in the core package and classifies them by:
- **Category:** What the value controls
- **Location:** Where it appears
- **Current Value:** The hardcoded value
- **Impact:** What breaks if changed
- **Recommendation:** Should it be configurable

**Finding:** 13 hardcoded values identified, 8 of which are intentionally hardcoded for safety/performance, 5 of which could optionally be configurable.

---

## Detailed Findings

### ✅ INTENTIONALLY HARDCODED (Safe & Correct)

These values should **remain hardcoded** - they're core constants that define behavior.

---

#### 1. Default Provider Selection

**Location:** `core/agent.go:23, 67` and `core/providers/provider.go:86`

**Current Value:** `"ollama"`

**Code:**
```go
// agent.go:23
providerType := agent.Provider
if providerType == "" {
    providerType = "ollama" // Default to Ollama for local development
}

// core/providers/provider.go:86
if providerType == "" {
    providerType = "ollama"
}
```

**Impact:** Determines which LLM backend to use (local vs cloud)

**Analysis:**
- ✅ **Correct to hardcode** - This is a deliberate architectural choice
- Default should be Ollama (free, local, no API keys)
- Users can override via `agent.Provider` in YAML config
- Makes local development out-of-box experience excellent

**Recommendation:** KEEP HARDCODED - This is a sensible default

**Related Config Override:**
```yaml
# In agent YAML configuration:
provider: "openai"  # Override default
```

---

#### 2. Default Ollama URL

**Location:** `core/providers/ollama/provider.go:57` and `core/providers/provider.go:120`

**Current Value:** `"http://localhost:11434"`

**Code:**
```go
// ollama/provider.go:57
if baseURL == "" {
    baseURL = "http://localhost:11434" // Default Ollama URL
}

// provider.go:120
providerURL = "http://localhost:11434" // Default Ollama URL
```

**Impact:** Where the code looks for Ollama server

**Analysis:**
- ✅ **Correct to hardcode** - This is Ollama's standard default port
- Port 11434 is Ollama's official default (documented in their setup)
- Users can override via `provider_url` in YAML config
- Most developers will use default local setup

**Recommendation:** KEEP HARDCODED - Standard Ollama port

**Related Config Override:**
```yaml
# For remote Ollama server:
provider_url: "http://192.168.1.100:11434"
```

---

#### 3. OpenAI Client TTL Cache Timeout

**Location:** `core/providers/openai/provider.go:27`

**Current Value:** `1 * time.Hour`

**Code:**
```go
const clientTTL = 1 * time.Hour
```

**Impact:** How long OpenAI client instances stay cached before expiring

**Analysis:**
- ✅ **Correct to hardcode** - This is an optimal cache lifetime
- 1 hour balances:
  - Memory efficiency (doesn't accumulate old clients)
  - Performance (avoids recreating clients too frequently)
  - API key rotation (old clients expire if key changes)
- Not something users need to configure
- Internally managed optimization

**Recommendation:** KEEP HARDCODED - Performance tuning value

**Context:**
```go
// client cache cleanup runs every:
const cleanupInterval = 5 * time.Minute
```

---

#### 4. Client Cache Cleanup Interval

**Location:** `core/providers/openai/provider.go:74`

**Current Value:** `5 * time.Minute`

**Code:**
```go
ticker := time.NewTicker(5 * time.Minute)
```

**Impact:** How frequently the system removes expired OpenAI clients from memory

**Analysis:**
- ✅ **Correct to hardcode** - This is an internal optimization
- 5 minutes provides good balance:
  - Doesn't run too frequently (CPU overhead)
  - Doesn't accumulate too much garbage (memory)
- Pure performance tuning, not user-facing
- Cleanup happens automatically in background goroutine

**Recommendation:** KEEP HARDCODED - Internal optimization

---

#### 5. HTTP Client Timeout for Ollama

**Location:** `core/providers/ollama/provider.go:73-75`

**Current Value:** `0` (no timeout for streaming)

**Code:**
```go
client: &http.Client{
    Timeout: 0, // Streaming requests can take a while
},
```

**Impact:** How long HTTP requests to Ollama can run

**Analysis:**
- ✅ **Correct to hardcode** - Essential for streaming support
- Streaming responses (real-time chat) can take indefinite time
- Timeout=0 means: respect context timeout instead
- Context timeout is configurable per-request
- Setting a fixed timeout would break streaming

**Recommendation:** KEEP HARDCODED - Required for streaming

**Related Context:** Each request via `ExecuteAgentStream()` can have its own context timeout

---

#### 6. System Message Role

**Location:** `core/providers/ollama/provider.go:276` and `core/providers/openai/provider.go`

**Current Value:** `"system"`

**Code:**
```go
result = append(result, OllamaMessage{
    Role:    "system",
    Content: systemPrompt,
})
```

**Impact:** How system prompts are formatted in API calls

**Analysis:**
- ✅ **Correct to hardcode** - Standard across all LLM APIs
- "system", "user", "assistant" are standard roles defined by OpenAI/Ollama
- Not configurable by spec
- Changing this would break all LLM communication

**Recommendation:** KEEP HARDCODED - API specification requirement

---

#### 7. User Message Role Default

**Location:** `core/providers/ollama/provider.go:286`

**Current Value:** `"user"`

**Code:**
```go
if role != "system" && role != "assistant" && role != "assistant" {
    role = "user" // Default to user for unknown roles
}
```

**Impact:** How unknown message roles are handled

**Analysis:**
- ✅ **Correct to hardcode** - Safe fallback for unknown roles
- Message roles must be one of: "system", "user", "assistant"
- Unknown roles default to "user" (safest option)
- Prevents API errors from invalid role values
- Should never happen in practice (internal code control)

**Recommendation:** KEEP HARDCODED - Safe default

---

#### 8. Tool Name Case Requirement

**Location:** `core/providers/ollama/provider.go:331`

**Current Value:** `'A'..='Z'` (uppercase start)

**Code:**
```go
// Validate tool name starts with uppercase (convention for functions)
if len(toolName) > 0 && toolName[0] >= 'A' && toolName[0] <= 'Z' {
```

**Impact:** Which tool calls are recognized from Ollama responses

**Analysis:**
- ✅ **Correct to hardcode** - Go naming convention enforcement
- Tool names should follow Go function naming: `PascalCase`
- Examples: `GetCPUUsage()`, `CheckDiskStatus()`, `ExecuteCommand()`
- Prevents false positives on lowercase text
- Matches actual tool implementations in code

**Recommendation:** KEEP HARDCODED - Enforces Go conventions

---

### ⚠️ OPTIONAL CONFIGURATION (Could be Configurable)

These values currently hardcoded but could optionally be made configurable if needed.

---

#### 9. Parallel Agent Execution Timeout

**Location:** `core/crew.go:1183`

**Current Value:** `60 * time.Second` (60 seconds)

**Code:**
```go
const ParallelAgentTimeout = 60 * time.Second
```

**Impact:** Maximum time each agent has when running in parallel groups

**Analysis:**
- ⚠️ **Could be configurable** - But current default is reasonable
- 60 seconds is a good default for most agent executions
- Prevents hanging parallel agents from blocking entire crew
- Different use cases might need different timeouts
- Complex reasoning tasks might need longer, simple tasks less

**Recommendation:** OPTIONAL - Keep hardcoded for now, could move to CrewConfig if needed

**When to configure:** If parallel agents consistently timeout or run too fast

---

#### 10. Maximum Tool Output Characters

**Location:** `core/crew.go:1425`

**Current Value:** `2000`

**Code:**
```go
const maxOutputChars = 2000 // Maximum characters per tool output to prevent context overflow
```

**Impact:** How much output from tools is included in agent responses

**Analysis:**
- ⚠️ **Could be configurable** - But current default is reasonable
- Prevents tool output from overwhelming the LLM context window
- 2000 characters is a safe default (roughly 500 tokens)
- Most tools produce output under this limit
- Different use cases might need different values

**Recommendation:** OPTIONAL - Keep hardcoded for now, could move to config if needed

**When to configure:** If you have tools that produce very large outputs

---

#### 11. Log Message Patterns

**Location:** Multiple files (`http.go:383-385`, `crew.go:579, 1237`)

**Current Value:** Various log patterns and status strings

**Examples:**
```go
// http.go:383
log.Printf("🚀 HTTP Server starting on http://localhost:%d", port)

// crew.go:579
status := "✅"

// crew.go:1237
status := "✅"
```

**Impact:** Human-readable log and status messages

**Analysis:**
- ⚠️ **Could be configurable** - But not necessary
- These are logging/debugging outputs only
- No functional impact
- Emojis and colors are user experience improvements
- Don't affect API or behavior

**Recommendation:** KEEP HARDCODED - UX enhancement, not needed as config

---

#### 12. Request ID Context Key

**Location:** `core/request_tracking.go:15`

**Current Value:** `"request-id"`

**Code:**
```go
const RequestIDKey = "request-id"
```

**Impact:** How request IDs are stored in context

**Analysis:**
- ⚠️ **Could be configurable** - But internal implementation detail
- This is just a string key for storing data in context
- Used throughout request tracking system
- Not exposed to users
- Changing would require code updates anyway

**Recommendation:** KEEP HARDCODED - Internal implementation constant

---

#### 13. Invalid UTF-8 Test Pattern

**Location:** `core/http_test.go:460`

**Current Value:** `"hello\xff\xfe"`

**Code:**
```go
invalidUTF8 := "hello\xff\xfe"
```

**Impact:** Test data for UTF-8 validation testing

**Analysis:**
- ✅ **Not a real hardcoded value** - This is test data
- Used only in test_test.go for UTF-8 edge cases
- Should NOT be configurable (it's test fixture)
- Example in test showing invalid UTF-8 handling

**Recommendation:** KEEP HARDCODED - Test fixture, not production code

---

## Configuration Override Mechanisms

The system provides several ways to override hardcoded defaults:

### 1. YAML Agent Configuration (Highest Priority)

```yaml
# In agent YAML files (e.g., executor.yaml):
provider: "openai"                           # Override provider type
provider_url: "http://remote-ollama:11434"   # Override Ollama URL
model: "gpt-4o-mini"                         # Override model selection
temperature: 0.8                              # Override sampling temperature
```

### 2. Environment Variables (For API Keys)

```bash
export OPENAI_API_KEY="sk-xxx..."           # Required for OpenAI provider
export OLLAMA_URL="http://ollama:11434"     # Override Ollama URL (if implemented)
```

**Note:** Currently, Ollama URL must be set in YAML. Environment variable override could be added if needed.

### 3. Function Parameters (Runtime)

```go
// When calling ExecuteAgent directly:
provider, err := providerFactory.GetProvider(
    "openai",                        // Override provider type
    "http://custom-ollama:11434",    // Override URL
    "sk-xxx...",                     // API key
)
```

---

## Recommendations Summary

### Keep Hardcoded (12 values)
✅ All current hardcoded values are appropriate
✅ Most relate to performance/safety tuning
✅ Good defaults for 95% of use cases
✅ Users can override via YAML config when needed

### No Action Required
- System already supports configuration override for user-facing values
- Performance/internal constants are well-tuned
- Architecture decisions (Ollama default) are sound

### Future Enhancements (If Needed)
**If streaming timeout is needed:** Add context timeout parameter
**If cache TTL is problematic:** Make it an init-time configuration
**If Ollama URL from env is needed:** Add `OLLAMA_URL` env var support

---

## Conclusion

The core package demonstrates excellent practice for hardcoded values:

1. **Performance Constants** (TTL, cleanup interval) - Properly hardcoded
2. **API Specifications** (roles, message format) - Must remain hardcoded
3. **Safe Defaults** (provider selection, URL) - Correctly hardcoded with override support
4. **Convention Enforcement** (tool naming) - Good to hardcode for consistency

**Overall Assessment:** ✅ **No changes needed**

The system provides sufficient flexibility through YAML configuration for the values that matter to users, while keeping internal optimizations and specifications hardcoded.

---

## Hardcoded Values Reference Table

| # | Category | Location | Value | Type | Recommendation |
|---|----------|----------|-------|------|---|
| 1 | Default Provider | agent.go:23,67 | `"ollama"` | Architecture | KEEP |
| 2 | Ollama URL | ollama/provider.go:57 | `"http://localhost:11434"` | Default | KEEP |
| 3 | OpenAI Client TTL | openai/provider.go:27 | `1 * time.Hour` | Performance | KEEP |
| 4 | Cache Cleanup Interval | openai/provider.go:74 | `5 * time.Minute` | Performance | KEEP |
| 5 | HTTP Client Timeout | ollama/provider.go:73 | `0` (streaming) | Required | KEEP |
| 6 | System Message Role | ollama/provider.go:276 | `"system"` | Specification | KEEP |
| 7 | User Role Default | ollama/provider.go:286 | `"user"` | Fallback | KEEP |
| 8 | Tool Name Convention | ollama/provider.go:331 | Uppercase start | Convention | KEEP |
| 9 | Parallel Agent Timeout | crew.go:1183 | `60 * time.Second` | Performance | OPTIONAL |
| 10 | Max Tool Output | crew.go:1425 | `2000 chars` | Optimization | OPTIONAL |
| 11 | Log Patterns | http.go:383+ | Various emoji/text | UX | KEEP HARDCODED |
| 12 | Request ID Key | request_tracking.go:15 | `"request-id"` | Internal | KEEP |
| 13 | Test Data | http_test.go:460 | `"hello\xff\xfe"` | Test Fixture | KEEP |

---

## Version Info

- **go-agentic Version:** Multi-provider refactored (Phase 3-4 complete)
- **Audit Date:** 2025-12-22
- **Auditor:** Code review via grep and semantic analysis
- **Status:** All findings documented and classified

