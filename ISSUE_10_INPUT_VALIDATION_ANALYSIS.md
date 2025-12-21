# 📋 Issue #10: No Input Validation

**Project**: go-agentic Library
**Issue**: Missing input validation for user queries and agent parameters
**Files**:
- http.go:64-78 (StreamHandler query parameter)
- crew.go:120 (SetResumeAgent)
**Date**: 2025-12-22
**Status**: 🔍 **ANALYSIS IN PROGRESS**

---

## 🔴 Problem Statement

### Current Issues

The system accepts user input without validation, creating several attack vectors and reliability issues:

```go
// http.go:76-78 (UNSAFE)
if req.Query == "" {
    http.Error(w, "Query is required", http.StatusBadRequest)
    return
}
// ❌ PROBLEM 1: Only checks if empty, nothing else
// ❌ No length validation
// ❌ No character validation
// ❌ No SQL injection protection (if DB is added later)
// ❌ No rate limiting
```

```go
// crew.go:122-124 (UNSAFE)
func (ce *CrewExecutor) SetResumeAgent(agentID string) {
    ce.ResumeAgentID = agentID  // ❌ No validation!
}
// ❌ Accepts any string
// ❌ No check if agent actually exists
// ❌ No format validation
```

---

## 🎯 Failure Scenarios

### Scenario #1: Extremely Long Query (DoS Attack)
```go
// Attacker sends:
query := strings.Repeat("a", 100*1024*1024)  // 100MB string

// Current behavior:
// ✅ Check: Query is not empty → PASS
// ❌ Result: System tries to process 100MB query
// ❌ Memory exhaustion
// ❌ CPU spike from LLM processing massive input
// ❌ Potential timeout or crash
```

### Scenario #2: Invalid Agent ID Resume
```go
// Attacker sends:
POST /api/crew/stream
{
    "query": "what is this?",
    "resumeAgent": "invalid_agent_xyz"
}

// Current behavior in crew.go:122
SetResumeAgent("invalid_agent_xyz")
ce.ResumeAgentID = "invalid_agent_xyz"  // ❌ No validation

// Later in ExecuteStream:122
currentAgent = ce.findAgentByID("invalid_agent_xyz")
if currentAgent == nil {
    return fmt.Errorf("resume agent ... not found")
}
// ✅ Error is caught, but...
// ❌ Resources wasted validating invalid input
// ❌ No clear early error message
```

### Scenario #3: Special Characters in Query
```go
// Attacker sends query with:
// - Unicode bombs (multiple combining characters)
// - Zero-width characters
// - RTL/LTR marks
// - Control characters

query := "test\x00\x01\x02..."  // Null bytes and control chars

// Current behavior:
// ✅ Not empty → PASS
// ❌ Sent to LLM as-is
// ❌ Potential confusion/errors
// ❌ No sanitization
```

### Scenario #4: URL Encoding Bypass
```go
// Attacker sends:
// GET /api/crew/stream?q=%00%01%02

req.Query = r.URL.Query().Get("q")  // Already decoded!
// ❌ URL decoded but not validated
// ❌ Null bytes and control chars in string
// ❌ Not validated after decoding
```

### Scenario #5: JSON Injection in Query Parameter
```go
// Normal request:
GET /api/crew/stream?q={"query":"hello"}

// Attack:
GET /api/crew/stream?q={"query":"hello","_system":"break out"}

// Current behavior:
// ❌ No schema validation
// ❌ Extra fields accepted but ignored
// ❌ Could cause confusion if _system field matters
```

### Scenario #6: Extremely Deep JSON
```go
// Attacker sends nested JSON:
{
    "query": "test",
    "history": [
        {"role": "user", "content": "a", "nested": {
            "deep": {
                "deeper": {
                    "deepest": ...1000 levels...
                }
            }
        }}
    ]
}

// Current behavior:
// ❌ No depth limit
// ❌ Decoder might fail or use excessive memory
// ❌ Stack overflow possible (rare)
```

---

## 📊 Input Validation Requirements

### Query Parameter Validation

| Requirement | Reason | Type |
|------------|--------|------|
| **Non-empty** | Query is required | Functional |
| **Max length** | Prevent DoS/memory exhaustion | Security |
| **Min length** | Prevent noise/useless queries | Functional |
| **Printable chars** | Reject control characters | Security |
| **No null bytes** | Prevent string truncation | Security |
| **Valid UTF-8** | Handle encoding properly | Functional |
| **No extremely long tokens** | Prevent single token DoS | Security |

**Recommended Limits**:
```
Minimum: 1 character
Maximum: 10,000 characters (typical: 100-500)
Max tokens: ~3,000 (OpenAI limit)
Token size: max 2KB per token
```

### AgentID Validation

| Requirement | Reason | Type |
|------------|--------|------|
| **Non-empty** | ID must be specified | Functional |
| **Alphanumeric** | Valid identifier format | Functional |
| **Max length** | Prevent excessively long IDs | Security |
| **Exists** | Agent must be in crew | Functional |
| **No special chars** | Prevent injection | Security |

**Recommended Limits**:
```
Pattern: ^[a-zA-Z0-9_-]{1,128}$
Length: 1-128 characters
Format: alphanumeric, underscore, hyphen only
```

### History Validation

| Requirement | Reason | Type |
|------------|--------|------|
| **Max items** | Prevent memory explosion | Security |
| **Item size limit** | Individual message limit | Security |
| **Total size limit** | Entire history limit | Security |
| **Valid roles** | Only "user"/"assistant"/"system" | Functional |
| **Valid content** | UTF-8, no null bytes | Functional |

**Recommended Limits**:
```
Max messages: 1,000
Max per message: 100KB
Max total history: 10MB
Valid roles: user, assistant, system
```

---

## 🎯 Solutions Comparison

### Solution 1: Basic Length Validation Only

**Approach**: Add simple size checks to prevent obvious DoS

```go
const (
    MaxQueryLength = 10000
    MinQueryLength = 1
)

func validateQuery(query string) error {
    if len(query) < MinQueryLength || len(query) > MaxQueryLength {
        return fmt.Errorf("query length must be %d-%d characters",
            MinQueryLength, MaxQueryLength)
    }
    return nil
}

// In StreamHandler:
if err := validateQuery(req.Query); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

**Advantages**:
- ✅ Simple to implement
- ✅ Prevents obvious DoS attacks
- ✅ Minimal performance impact

**Disadvantages**:
- ❌ No character validation
- ❌ Doesn't catch special characters
- ❌ Limited protection
- ❌ No agent existence validation

**Breaking Changes**: NONE

---

### Solution 2: Comprehensive Input Validation

**Approach**: Validate all inputs thoroughly with sanitization

```go
type InputValidator struct {
    MaxQueryLen      int
    MinQueryLen      int
    MaxHistoryLen    int
    MaxMessageSize   int
    AllowedRoles     map[string]bool
}

func (v *InputValidator) ValidateQuery(query string) error {
    // 1. Length check
    if len(query) < v.MinQueryLen || len(query) > v.MaxQueryLen {
        return fmt.Errorf("query length out of bounds")
    }

    // 2. UTF-8 validation
    if !utf8.ValidString(query) {
        return fmt.Errorf("query contains invalid UTF-8")
    }

    // 3. Null byte check
    if strings.ContainsRune(query, '\x00') {
        return fmt.Errorf("query contains null bytes")
    }

    // 4. Control character check
    for _, r := range query {
        if unicode.IsControl(r) && r != '\n' && r != '\t' {
            return fmt.Errorf("query contains control characters")
        }
    }

    return nil
}

func (v *InputValidator) ValidateAgentID(agentID string) error {
    // Pattern: alphanumeric, underscore, hyphen only
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{1,128}$`, agentID)
    if !matched {
        return fmt.Errorf("invalid agent ID format")
    }
    return nil
}

func (v *InputValidator) ValidateHistory(history []Message) error {
    if len(history) > v.MaxHistoryLen {
        return fmt.Errorf("history exceeds maximum messages")
    }

    for i, msg := range history {
        if len(msg.Content) > v.MaxMessageSize {
            return fmt.Errorf("message %d exceeds size limit", i)
        }
        if !v.AllowedRoles[msg.Role] {
            return fmt.Errorf("message %d has invalid role: %s", i, msg.Role)
        }
    }
    return nil
}
```

**Advantages**:
- ✅ Comprehensive protection
- ✅ UTF-8 validation
- ✅ Special character detection
- ✅ Format validation
- ✅ Multiple validation layers
- ✅ Clear error messages

**Disadvantages**:
- ❌ More code
- ❌ Slightly more complex
- ❌ Regex performance overhead
- ❌ Still doesn't check agent existence

**Breaking Changes**: NONE

---

### Solution 3: Comprehensive + Agent Existence Check (Best)

**Approach**: Full validation including agent existence verification

```go
func (h *HTTPHandler) ValidateStreamRequest(req *StreamRequest, agent *Agent) error {
    // 1. Validate query
    if err := h.validator.ValidateQuery(req.Query); err != nil {
        return fmt.Errorf("invalid query: %w", err)
    }

    // 2. Validate history
    if err := h.validator.ValidateHistory(req.History); err != nil {
        return fmt.Errorf("invalid history: %w", err)
    }

    // 3. If resume agent specified, validate existence
    if req.ResumeAgent != "" {
        resumeAgent := h.executor.findAgentByID(req.ResumeAgent)
        if resumeAgent == nil {
            return fmt.Errorf("resume agent not found: %s", req.ResumeAgent)
        }
    }

    return nil
}

// In StreamHandler:
if err := h.ValidateStreamRequest(&req, h.executor.entryAgent); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

**Advantages**:
- ✅ All benefits of Solution 2
- ✅ Checks agent existence early
- ✅ Clear error messages
- ✅ Fail-fast approach
- ✅ Best user experience

**Disadvantages**:
- ❌ Requires passing crew to handler
- ❌ Agent lookup adds small performance cost
- ❌ Slightly more code

**Breaking Changes**: NONE

---

## 🏆 RECOMMENDATION: **Solution 3 - Comprehensive + Existence Check**

### Why This Is Best

1. **Security**: Comprehensive input validation blocks most attacks
2. **Reliability**: Early validation prevents wasted processing
3. **UX**: Clear error messages help developers
4. **Prevention**: Agent existence check prevents downstream errors
5. **Maintainability**: Clear separation of validation logic

---

## 📈 Implementation Plan: Solution 3

### Step 1: Create InputValidator Type

**File**: http.go (new section)

```go
type InputValidator struct {
    MaxQueryLen      int
    MinQueryLen      int
    MaxHistoryLen    int
    MaxMessageSize   int
    AllowedRoles     map[string]bool
}

func NewInputValidator() *InputValidator {
    return &InputValidator{
        MaxQueryLen:    10000,
        MinQueryLen:    1,
        MaxHistoryLen:  1000,
        MaxMessageSize: 100000,  // 100KB per message
        AllowedRoles: map[string]bool{
            "user":      true,
            "assistant": true,
            "system":    true,
        },
    }
}

// ValidateQuery checks query parameter
func (v *InputValidator) ValidateQuery(query string) error {
    // Length check
    if len(query) < v.MinQueryLen || len(query) > v.MaxQueryLen {
        return fmt.Errorf("query length must be %d-%d characters",
            v.MinQueryLen, v.MaxQueryLen)
    }

    // UTF-8 check
    if !utf8.ValidString(query) {
        return fmt.Errorf("query contains invalid UTF-8")
    }

    // Null byte check
    if strings.ContainsRune(query, '\x00') {
        return fmt.Errorf("query contains null bytes")
    }

    // Control character check (allow newline/tab)
    for _, r := range query {
        if unicode.IsControl(r) && r != '\n' && r != '\t' {
            return fmt.Errorf("query contains invalid control characters")
        }
    }

    return nil
}

// ValidateAgentID checks agent ID format
func (v *InputValidator) ValidateAgentID(agentID string) error {
    if agentID == "" {
        return fmt.Errorf("agent ID cannot be empty")
    }

    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{1,128}$`, agentID)
    if !matched {
        return fmt.Errorf("invalid agent ID format (alphanumeric, underscore, hyphen only)")
    }

    return nil
}

// ValidateHistory checks history slice
func (v *InputValidator) ValidateHistory(history []Message) error {
    if len(history) > v.MaxHistoryLen {
        return fmt.Errorf("history exceeds maximum %d messages", v.MaxHistoryLen)
    }

    totalSize := 0
    for i, msg := range history {
        // Role validation
        if !v.AllowedRoles[msg.Role] {
            return fmt.Errorf("message %d: invalid role '%s'", i, msg.Role)
        }

        // Content size validation
        contentSize := len([]byte(msg.Content))
        if contentSize > v.MaxMessageSize {
            return fmt.Errorf("message %d exceeds size limit (%d > %d bytes)",
                i, contentSize, v.MaxMessageSize)
        }

        totalSize += contentSize
    }

    return nil
}
```

### Step 2: Add Validator to HTTPHandler

**File**: http.go (modify HTTPHandler struct)

```go
type HTTPHandler struct {
    executor  *CrewExecutor
    mu        sync.RWMutex
    validator *InputValidator  // ← ADD THIS
}

func NewHTTPHandler(executor *CrewExecutor) *HTTPHandler {
    return &HTTPHandler{
        executor:  executor,
        validator: NewInputValidator(),  // ← INITIALIZE
    }
}
```

### Step 3: Validate Request in StreamHandler

**File**: http.go (modify StreamHandler)

```go
// After parsing req.Query, add validation:

// ✅ FIX for Issue #10: Comprehensive input validation
if err := h.validator.ValidateQuery(req.Query); err != nil {
    log.Printf("[INPUT ERROR] Invalid query: %v", err)
    http.Error(w, fmt.Sprintf("Invalid query: %v", err), http.StatusBadRequest)
    return
}

if err := h.validator.ValidateHistory(req.History); err != nil {
    log.Printf("[INPUT ERROR] Invalid history: %v", err)
    http.Error(w, fmt.Sprintf("Invalid history: %v", err), http.StatusBadRequest)
    return
}

// If resume agent specified, validate it exists
if req.ResumeAgent != "" {
    if err := h.validator.ValidateAgentID(req.ResumeAgent); err != nil {
        log.Printf("[INPUT ERROR] Invalid agent ID: %v", err)
        http.Error(w, fmt.Sprintf("Invalid agent ID: %v", err), http.StatusBadRequest)
        return
    }

    resumeAgent := h.executor.findAgentByID(req.ResumeAgent)
    if resumeAgent == nil {
        log.Printf("[INPUT ERROR] Resume agent not found: %s", req.ResumeAgent)
        http.Error(w, fmt.Sprintf("Resume agent not found: %s", req.ResumeAgent),
            http.StatusBadRequest)
        return
    }
}
```

### Step 4: Add Tests

**Tests to add**:
- `TestValidateQueryLength` - Too short/too long
- `TestValidateQueryUTF8` - Invalid UTF-8
- `TestValidateQueryNullBytes` - Null byte detection
- `TestValidateQueryControlChars` - Control character detection
- `TestValidateAgentIDFormat` - Valid/invalid formats
- `TestValidateHistory` - Valid/invalid history
- `TestValidateTotalHistorySize` - Size limits
- `TestStreamHandlerInputValidation` - Integration test

---

## ✅ Benefits

### Security
- ✅ Prevents DoS attacks (size limits)
- ✅ Blocks malformed input (UTF-8, control chars)
- ✅ Validates agent existence (prevents errors)
- ✅ Clear early error detection

### Reliability
- ✅ Fail-fast on invalid input
- ✅ No wasted processing
- ✅ Clear error messages
- ✅ Prevents downstream errors

### Operations
- ✅ Easy to adjust limits
- ✅ Clear validation logging
- ✅ Can track validation failures
- ✅ Easier debugging

---

## 📊 Break Changes Analysis

### Changes Required

| Component | Change | Type | Impact |
|-----------|--------|------|--------|
| HTTPHandler | Add validator field | Internal | None |
| StreamHandler | Add validation checks | Internal | None |
| SetResumeAgent | Add validation | Internal | None |
| Input limits | New constants | Config | Can adjust |

### Breaking Changes Count: **ZERO** ✅

- No API changes
- No new required parameters
- Validation limits can be adjusted if needed
- Fail-fast behavior is transparent to valid inputs

---

*Generated: 2025-12-22*
*Status*: 🔍 **ANALYSIS COMPLETE - AWAITING APPROVAL**
*Recommendation*: ✅ **Proceed with Solution 3 (Comprehensive + Existence Check)**
