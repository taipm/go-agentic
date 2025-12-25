# Quick Start: Tool Improvements Implementation

**For:** Developers ready to start implementing
**Time:** Read this in 5 minutes, then start coding
**Prerequisites:** Go 1.18+, git, go-agentic cloned

---

## 🚀 Get Started in 3 Steps

### Step 1: Create the Coercion Utility (30 min)

**Copy-paste ready! Just create this file:**

📁 **`core/tools/coercion.go`**

```go
package tools

import (
	"fmt"
	"strconv"
	"strings"
)

// CoerceToString converts various types to string safely.
func CoerceToString(v interface{}) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val)), nil
		}
		return fmt.Sprintf("%v", val), nil
	case int64:
		return fmt.Sprintf("%d", val), nil
	case int:
		return fmt.Sprintf("%d", val), nil
	case int32:
		return fmt.Sprintf("%d", val), nil
	case bool:
		return fmt.Sprintf("%v", val), nil
	case nil:
		return "", fmt.Errorf("cannot coerce nil to string")
	default:
		return fmt.Sprintf("%v", val), nil
	}
}

// CoerceToInt converts various types to int safely.
func CoerceToInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case float64:
		return int(val), nil
	case int64:
		return int(val), nil
	case int32:
		return int(val), nil
	case int:
		return val, nil
	case string:
		var i int
		_, err := fmt.Sscanf(val, "%d", &i)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as integer: %w", val, err)
		}
		return i, nil
	case nil:
		return 0, fmt.Errorf("cannot coerce nil to int")
	default:
		return 0, fmt.Errorf("cannot coerce %T to int", v)
	}
}

// CoerceToBool converts various types to bool safely.
func CoerceToBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "true", "yes", "1", "on", "enabled":
			return true, nil
		case "false", "no", "0", "off", "disabled":
			return false, nil
		default:
			return false, fmt.Errorf("cannot parse %q as boolean", val)
		}
	case float64:
		return val != 0, nil
	case int, int32, int64:
		return val != 0, nil
	case nil:
		return false, fmt.Errorf("cannot coerce nil to bool")
	default:
		return false, fmt.Errorf("cannot coerce %T to bool", v)
	}
}

// CoerceToFloat converts various types to float64 safely.
func CoerceToFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int, int32, int64:
		return float64(val.(int64)), nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse %q as float: %w", val, err)
		}
		return f, nil
	case nil:
		return 0, fmt.Errorf("cannot coerce nil to float")
	default:
		return 0, fmt.Errorf("cannot coerce %T to float", v)
	}
}

// MustGetString extracts and coerces a required string parameter.
func MustGetString(args map[string]interface{}, key string) (string, error) {
	val, exists := args[key]
	if !exists {
		return "", fmt.Errorf("required parameter %q missing", key)
	}

	result, err := CoerceToString(val)
	if err != nil {
		return "", fmt.Errorf("parameter %q: %w", key, err)
	}
	return result, nil
}

// MustGetInt extracts and coerces a required int parameter.
func MustGetInt(args map[string]interface{}, key string) (int, error) {
	val, exists := args[key]
	if !exists {
		return 0, fmt.Errorf("required parameter %q missing", key)
	}

	result, err := CoerceToInt(val)
	if err != nil {
		return 0, fmt.Errorf("parameter %q: %w", key, err)
	}
	return result, nil
}

// MustGetBool extracts and coerces a required bool parameter.
func MustGetBool(args map[string]interface{}, key string) (bool, error) {
	val, exists := args[key]
	if !exists {
		return false, fmt.Errorf("required parameter %q missing", key)
	}

	result, err := CoerceToBool(val)
	if err != nil {
		return false, fmt.Errorf("parameter %q: %w", key, err)
	}
	return result, nil
}

// OptionalGetString extracts an optional string parameter with default value.
func OptionalGetString(args map[string]interface{}, key, defaultVal string) string {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToString(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// OptionalGetInt extracts an optional int parameter with default value.
func OptionalGetInt(args map[string]interface{}, key string, defaultVal int) int {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToInt(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// OptionalGetBool extracts an optional bool parameter with default value.
func OptionalGetBool(args map[string]interface{}, key string, defaultVal bool) bool {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToBool(val)
	if err != nil {
		return defaultVal
	}
	return result
}

// OptionalGetFloat extracts an optional float parameter with default value.
func OptionalGetFloat(args map[string]interface{}, key string, defaultVal float64) float64 {
	val, exists := args[key]
	if !exists {
		return defaultVal
	}

	result, err := CoerceToFloat(val)
	if err != nil {
		return defaultVal
	}
	return result
}
```

**Test it:**
```bash
cd /Users/taipm/GitHub/go-agentic
go test ./core/tools -v -run Coerce -run "CoerceToString|CoerceToInt" 2>&1 | head -20
```

---

### Step 2: Create the Validation Utility (45 min)

📁 **`core/tools/validation.go`**

See `IMPLEMENTATION_PLAN.md` section "Quick Win #2" for full code.

**Quick test:**
```bash
go test ./core/tools -v -run Validate 2>&1 | head -20
```

---

### Step 3: Add Per-Tool Timeout (30 min)

**In `core/types.go`, find the `Tool` struct and add:**

```go
type Tool struct {
    Name              string                      `json:"name"`
    Description       string                      `json:"description"`
    Parameters        map[string]interface{}      `json:"parameters,omitempty"`
    Func              ToolFunc                    `json:"-"`
    TimeoutSeconds    int                         `json:"timeout_seconds,omitempty"` // ← ADD THIS
}
```

**In `core/tools/executor.go`, find `ExecuteWithRetry()` and add:**

```go
func ExecuteWithRetry(ctx context.Context, tool *agenticcore.Tool, args map[string]interface{}) (string, error) {
    // Add per-tool timeout
    if tool.TimeoutSeconds > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, time.Duration(tool.TimeoutSeconds)*time.Second)
        defer cancel()
    }

    // ... rest of function
}
```

---

## ✅ Verify Everything Works

```bash
# Run all tests
go test ./core/tools -v

# Run an example
go run ./examples/01-quiz-exam/cmd/main.go
```

**Expected:** All tests pass ✓

---

## 📚 Next Steps (After Quick Wins)

Once Quick Wins are done:

1. **Builder Pattern** (See `IMPLEMENTATION_PLAN.md` → "Opportunity #1")
2. **Struct Schema** (See `IMPLEMENTATION_PLAN.md` → "Opportunity #2")
3. **Refactor examples** (See `IMPLEMENTATION_PLAN.md` → "Phase 3")

---

## 🎯 What You're Building

### Quick Win #1: Type Coercion
**Problem:** Repeat type assertion code in every tool
```go
// ❌ Old (10 lines)
var name string
switch v := args["name"].(type) {
case string:
    name = v
case float64:
    name = fmt.Sprintf("%v", v)
// ... more cases
}

// ✅ New (1 line)
name, err := tools.MustGetString(args, "name")
```

### Quick Win #2: Validation
**Problem:** Tool schemas can diverge from code
```go
// ✅ Now validates at load time
if err := tools.ValidateToolSchema(tool); err != nil {
    return nil, fmt.Errorf("tool config error: %w", err)
}
```

### Quick Win #3: Per-Tool Timeout
**Problem:** All tools use same timeout (inflexible)
```go
// ✅ Now each tool can have its own timeout
tools["FastTool"] = &Tool{
    TimeoutSeconds: 2,  // 2 seconds
}
tools["SlowTool"] = &Tool{
    TimeoutSeconds: 30, // 30 seconds
}
```

---

## 💡 Tips for Success

✅ **Do this:**
- Implement in order (Quick Wins first)
- Test after each step
- Refer to `IMPLEMENTATION_PLAN.md` for full code
- Run examples to verify

❌ **Don't do this:**
- Skip testing
- Implement all at once
- Forget to check for regressions

---

## 🆘 If You Get Stuck

1. **Can't find where to add code?** → Check `IMPLEMENTATION_PLAN.md` for exact file paths
2. **Test failures?** → Compare your code with code in `IMPLEMENTATION_PLAN.md`
3. **Not sure what to do next?** → Check `IMPLEMENTATION_CHECKLIST.md`

---

## 📞 Questions?

Check these files in this order:
1. `IMPLEMENTATION_PLAN.md` - Full details
2. `IMPLEMENTATION_CHECKLIST.md` - What to do
3. Example code in `examples/` - See it in action

---

**Ready to start? Begin with Step 1 above!** 🚀

