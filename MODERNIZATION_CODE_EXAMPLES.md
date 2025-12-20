# go-agentic Examples: Modernization Code Examples

This document shows concrete before/after examples for the recommended modernizations.

---

## 1. Consolidating Environment Loading

### BEFORE (Repeated in 4 files)
```go
// /examples/it-support/main.go
func loadEnvFile() error {
	// Try to find .env file in current directory or parent directories
	paths := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join(filepath.Dir(os.Args[0]), ".env"),
		filepath.Join(filepath.Dir(os.Args[0]), "..", ".env"),
	}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				// Skip comments and empty lines
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				// Parse key=value
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					os.Setenv(key, value)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("no .env file found")
}

// /examples/customer-service/main.go
func loadEnvFile() error {
	data, err := os.ReadFile(".env")
	if err != nil {
		return fmt.Errorf("no .env file found")
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
	return nil
}

// Repeated in data-analysis/main.go and research-assistant/main.go...
```

**Lines saved by duplication**: ~50 lines × 4 = 200 lines

### AFTER (Shared utility)
```go
// /examples/shared/env.go
package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadEnv loads environment variables from .env file
// Searches in current directory, parent directories, and alongside executable
func LoadEnv() error {
	paths := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join(filepath.Dir(os.Args[0]), ".env"),
		filepath.Join(filepath.Dir(os.Args[0]), "..", ".env"),
	}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					os.Setenv(key, value)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("no .env file found")
}

// GetAPIKey retrieves and validates OpenAI API key
func GetAPIKey() (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	return apiKey, nil
}
```

### Usage in all examples:
```go
// /examples/it-support/main.go
import "github.com/taipm/go-agentic/examples/shared"

func main() {
	if err := shared.LoadEnv(); err != nil {
		fmt.Printf("Note: %v\n", err)
	}

	apiKey, err := shared.GetAPIKey()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	// ... rest of main
}
```

**Reduction**: -200 lines of duplicated code

---

## 2. Parameter Validation Integration

### BEFORE (Manual validation, inconsistent)
```go
// From customer-service/example_customer_service.go
{
	Name:        "IssueRefund",
	Description: "Process a refund for a customer",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"account_id": map[string]interface{}{
				"type":        "string",
				"description": "The customer account ID",
			},
			"amount": map[string]interface{}{
				"type":        "string",
				"description": "The refund amount",
			},
		},
		"required": []string{"account_id", "amount"},
	},
	Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
		accountID, ok := args["account_id"].(string)
		if !ok {
			return "", fmt.Errorf("invalid account_id parameter")
		}
		amount, _ := args["amount"].(string)  // ⚠️ UNSAFE - not validating!
		_ = amount
		refundID := fmt.Sprintf("REF-%06d", rand.Intn(999999))
		return fmt.Sprintf("Refund processed!...", refundID, accountID, amount), nil
	},
}
```

**Problems**:
- ❌ `amount` is not validated (unsafe cast `_`)
- ❌ No error handling for missing required parameter
- ❌ Inconsistent with other tools

### AFTER (Using library validation)
```go
// Create helper in /examples/shared/tools.go
package shared

import (
	"fmt"
	agentic "github.com/taipm/go-agentic"
)

// ValidateRequired validates that all required parameters are present
func ValidateRequired(args map[string]interface{}, required []string) error {
	for _, param := range required {
		if _, ok := args[param]; !ok {
			return fmt.Errorf("missing required parameter: %s", param)
		}
	}
	return nil
}

// GetString safely retrieves a string parameter
func GetString(args map[string]interface{}, key string) (string, error) {
	val, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing parameter: %s", key)
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be string, got %T", key, val)
	}
	return str, nil
}

// ---

// In example file
{
	Name:        "IssueRefund",
	Description: "Process a refund for a customer",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"account_id": map[string]interface{}{
				"type":        "string",
				"description": "The customer account ID",
			},
			"amount": map[string]interface{}{
				"type":        "string",
				"description": "The refund amount",
			},
		},
		"required": []string{"account_id", "amount"},
	},
	Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
		// Validate all required parameters
		if err := shared.ValidateRequired(args, []string{"account_id", "amount"}); err != nil {
			return "", err
		}

		// Get and validate parameters safely
		accountID, err := shared.GetString(args, "account_id")
		if err != nil {
			return "", err
		}

		amount, err := shared.GetString(args, "amount")
		if err != nil {
			return "", err
		}

		refundID := fmt.Sprintf("REF-%06d", rand.Intn(999999))
		return fmt.Sprintf("Refund processed!...", refundID, accountID, amount), nil
	},
}
```

**Benefits**:
- ✅ Consistent validation across all tools
- ✅ Clear error messages
- ✅ Type-safe parameter access
- ✅ DRY principle applied

---

## 3. Tool Definition Helpers

### BEFORE (Repetitive schema construction)
```go
// Repeated pattern in all examples
{
	Name: "SearchKnowledgeBase",
	Description: "Search the knowledge base for solutions to common customer issues",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query",
			},
		},
		"required": []string{"query"},
	},
	Handler: /* ... */,
}
```

### AFTER (Using helpers)
```go
// /examples/shared/toolhelpers.go
package shared

import (
	agentic "github.com/taipm/go-agentic"
	"context"
)

// StringParam creates a string parameter definition
func StringParam(name, description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
	}
}

// NumberParam creates a numeric parameter definition
func NumberParam(name, description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "number",
		"description": description,
	}
}

// BoolParam creates a boolean parameter definition
func BoolParam(name, description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
}

// NewTool creates a tool with simpler syntax
type ToolBuilder struct {
	name        string
	description string
	properties  map[string]map[string]interface{}
	required    []string
	handler     func(context.Context, map[string]interface{}) (string, error)
}

func Tool(name, description string) *ToolBuilder {
	return &ToolBuilder{
		name:       name,
		description: description,
		properties: make(map[string]map[string]interface{}),
	}
}

func (t *ToolBuilder) String(paramName, paramDesc string) *ToolBuilder {
	t.properties[paramName] = StringParam(paramName, paramDesc)
	return t
}

func (t *ToolBuilder) Number(paramName, paramDesc string) *ToolBuilder {
	t.properties[paramName] = NumberParam(paramName, paramDesc)
	return t
}

func (t *ToolBuilder) Required(params ...string) *ToolBuilder {
	t.required = append(t.required, params...)
	return t
}

func (t *ToolBuilder) Handler(h func(context.Context, map[string]interface{}) (string, error)) *ToolBuilder {
	t.handler = h
	return t
}

func (t *ToolBuilder) Build() *agentic.Tool {
	return &agentic.Tool{
		Name:        t.name,
		Description: t.description,
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": t.properties,
			"required":   t.required,
		},
		Handler: t.handler,
	}
}

// Usage in examples:
var tools = []*agentic.Tool{
	shared.Tool("SearchKnowledgeBase", "Search knowledge base for solutions").
		String("query", "The search query").
		Required("query").
		Handler(searchKBHandler).
		Build(),

	shared.Tool("CheckAccountStatus", "Check customer account status").
		String("account_id", "The customer account ID").
		Required("account_id").
		Handler(checkAccountHandler).
		Build(),
}
```

**Reduction**: -100+ lines of boilerplate

---

## 4. Error Handling Standardization

### BEFORE (Inconsistent)
```go
// IT Support - detailed errors
Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
	host, ok := args["host"].(string)
	if !ok {
		return "", fmt.Errorf("host parameter required")
	}
	cmd := exec.CommandContext(ctx, "ping", "-c", "4", host)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ping failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// Customer Service - inconsistent
Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
	accountID, ok := args["account_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid account_id parameter")  // Different message
	}
	amount, _ := args["amount"].(string)  // Not checked at all
	// ...
	return fmt.Sprintf("Refund processed!...", refundID, accountID, amount), nil
}
```

### AFTER (Consistent)
```go
// /examples/shared/errors.go
package shared

import (
	"fmt"
	"log/slog"
)

// ToolError represents a tool execution error
type ToolError struct {
	Code    string // "VALIDATION", "EXECUTION", "NETWORK", etc.
	Message string
	Details error
}

func (e *ToolError) Error() string {
	return e.Message
}

// NewValidationError creates a validation error
func NewValidationError(paramName string, details error) error {
	return &ToolError{
		Code:    "VALIDATION",
		Message: fmt.Sprintf("validation failed for parameter '%s': %v", paramName, details),
		Details: details,
	}
}

// LogToolError logs tool execution error
func LogToolError(toolName string, err error) {
	if err == nil {
		return
	}
	slog.Error("tool execution failed",
		"tool", toolName,
		"error", err.Error(),
	)
}

// ---

// In example file (using helpers + consistent error handling)
Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
	// Validate parameters
	host, err := shared.GetString(args, "host")
	if err != nil {
		return "", shared.NewValidationError("host", err)
	}

	// Execute
	cmd := exec.CommandContext(ctx, "ping", "-c", "4", host)
	output, err := cmd.CombinedOutput()
	if err != nil {
		shared.LogToolError("PingHost", err)
		return "", fmt.Errorf("ping failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
```

---

## 5. Structured Logging (Optional - Go 1.21+)

### BEFORE
```go
fmt.Printf("[INFO] Agent '%s' (ID: %s) using model '%s' with temperature %.1f\n",
	agent.Name, agent.ID, agent.Model, agent.Temperature)

fmt.Printf("Error: %v\n\n", err)

fmt.Println("🚀 go-agentic IT Support Crew v1.0")
```

### AFTER
```go
import "log/slog"

slog.Info("Agent execution started",
	"agent_name", agent.Name,
	"agent_id", agent.ID,
	"model", agent.Model,
	"temperature", agent.Temperature,
)

slog.Error("Tool execution failed",
	"tool_name", tool.Name,
	"error", err.Error(),
)

slog.Info("Interactive CLI started",
	"system", "IT Support Crew",
	"version", "v1.0",
)
```

**Benefits**:
- ✅ Machine-parseable logs
- ✅ Structured fields for filtering
- ✅ Better debugging and monitoring
- ✅ Production-ready logging

---

## 6. YAML Configuration Example

### Current (Code-based)
```go
// example_customer_service.go
func createCustomerServiceCrew() *agentic.Crew {
	tools := createCustomerServiceTools()

	classifier := &agentic.Agent{
		ID:          "classifier",
		Name:        "Issue Classifier",
		Role:        "Customer issue categorizer",
		Backstory:   "You are expert at understanding customer issues...",
		Model:       "gpt-4o-mini",
		Tools:       []*agentic.Tool{},
		Temperature: 0.7,
		IsTerminal:  false,
	}
	// ... more agents
}
```

### Alternative (YAML-based, showcasing library feature)
```yaml
# config/agents.yaml
version: "1.0"
description: "Customer Service Support System"
entry_point: "classifier"

agents:
  - id: classifier
    name: "Issue Classifier"
    role: "Customer issue categorizer"
    backstory: "You are expert at understanding customer issues and categorizing them. Determine the type of issue and decide which team should handle it."
    model: "gpt-4o-mini"
    temperature: 0.7
    tools: []
    is_terminal: false

  - id: resolver
    name: "Issue Resolver"
    role: "Problem solver"
    backstory: "You are expert at resolving customer issues. Use available tools to investigate and resolve customer problems."
    model: "gpt-4o-mini"
    temperature: 0.7
    tools:
      - "SearchKnowledgeBase"
      - "CheckAccountStatus"
      - "ViewOrderHistory"
      - "IssueRefund"
      - "ResetPassword"
      - "CreateTicket"
    is_terminal: false

  - id: responder
    name: "Response Specialist"
    role: "Customer response coordinator"
    backstory: "You specialize in communicating solutions to customers in a friendly and professional manner."
    model: "gpt-4o-mini"
    temperature: 0.7
    tools: []
    is_terminal: true

settings:
  max_handoffs: 5
  max_rounds: 10
  timeout_seconds: 300
  language: "en"
  organization: "Customer-Support"
```

### Usage
```go
// main.go
func main() {
	// Load from YAML
	team, err := agentic.LoadTeamConfig("config/agents.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Create team from config
	agents := createAgentsFromConfig(team)
	crew := &agentic.Team{
		Agents: agents,
		MaxRounds: team.Settings.MaxRounds,
		MaxHandoffs: team.Settings.MaxHandoffs,
	}

	executor := agentic.NewTeamExecutor(crew, apiKey)
	// ...
}
```

---

## Summary of Improvements

| Improvement | Before | After | Benefit |
|------------|--------|-------|---------|
| Environment Loading | 4 copies × 50 lines | 1 shared module | -200 lines |
| Tool Definition | Manual + boilerplate | Builder pattern | -100 lines |
| Parameter Validation | Manual, inconsistent | Shared helpers | 100% consistency |
| Error Handling | Inconsistent | Standardized | Better debugging |
| Logging | fmt.Printf | slog structured | Machine-readable |
| Configuration | Code-based | YAML alt | Better education |
| **TOTAL** | **~1200 lines** | **~850 lines** | **30% reduction** |

