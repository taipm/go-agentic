# Comparison: go-agentic vs Best Practice Frameworks

**Purpose:** Show exactly how go-agentic compares to industry leaders and what needs to change.

---

## 1. Tool Definition Pattern

### Anthropic SDK (Python) - BEST PRACTICE ⭐

```python
from anthropic import beta_tool

@beta_tool
def get_weather(location: str) -> str:
    """Lookup the weather for a given city

    Args:
        location: The city and state, e.g. San Francisco, CA
    """
    return json.dumps({"location": location, "temperature": "72°F"})

# Usage:
runner = client.beta.messages.tool_runner(
    tools=[get_weather],
    messages=[...],
)
```

**DX Score:** 9.5/10 ⭐

**Why it's great:**
- ✅ Single decorator
- ✅ Type hints are the contract
- ✅ No manual validation code
- ✅ Framework handles everything
- ✅ ~10 LOC total

---

### go-agentic (Current) ❌

```go
// Step 1: Define tool function
func GetWeather(ctx context.Context, args map[string]interface{}) (string, error) {
    location, ok := args["location"].(string)  // ← Manual type assertion
    if !ok {
        return "", fmt.Errorf("location must be string")  // ← Manual validation
    }
    // ... implementation
    return weather, nil
}

// Step 2: Create Tool struct with manual JSON schema
toolsMap["GetWeather"] = &core.Tool{
    Name: "GetWeather",
    Description: "Lookup the weather",
    Parameters: map[string]interface{}{  // ← Hand-written JSON schema
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type": "string",
                "description": "The city and state",
            },
        },
        "required": []string{"location"},
    },
    Func: core.ToolHandler(GetWeather),
}

// Step 3: Reference in YAML
// agents/agent.yaml:
// tools:
//   - GetWeather  ← Must match key in map exactly

// Step 4: Pass to executor
executor, _ := core.NewCrewExecutorFromConfig(apiKey, "config", toolsMap)
```

**DX Score:** 6.5/10 ❌

**Problems:**
- ❌ 4 places to update
- ❌ Manual type assertions
- ❌ Manual validation code (boilerplate)
- ❌ Hand-written JSON schemas (error-prone)
- ❌ Name mismatch → silent failure
- ❌ ~45 LOC total
- ❌ No fail-fast validation

---

### LangChain (Python) - BEST PRACTICE ⭐

```python
from langchain_core.tools import tool

@tool
def get_weather(city: str) -> str:
    """Get weather for a given city."""
    return f"It's sunny in {city}!"

# Usage:
tools = [get_weather]
agent = create_agent(tools=tools)
```

**DX Score:** 8.5/10 ⭐

**Why it's great:**
- ✅ Minimal decorator
- ✅ Type hints become contracts
- ✅ ~6 LOC total
- ✅ Single registration method
- ✅ Works with LangGraph

---

### FastAPI (Python) - BEST PRACTICE ⭐

```python
from fastapi import FastAPI
from pydantic import BaseModel

class WeatherParams(BaseModel):
    location: str
    unit: str = "fahrenheit"

@app.post("/weather/")
def get_weather(params: WeatherParams) -> dict:
    return {"location": params.location, "temp": "72F"}

# Automatic benefits:
# ✅ Type validation
# ✅ OpenAPI docs
# ✅ Error responses (422)
# ✅ JSON schema generated
```

**DX Score:** 9/10 ⭐

**Why it's great:**
- ✅ Pydantic models (DRY)
- ✅ Automatic validation
- ✅ Type safe
- ✅ Self-documenting
- ✅ Clear error messages

---

## 2. Parameter Handling

### Anthropic SDK - Auto from Type Hints

```python
@beta_tool
def book_hotel(
    location: str,
    checkin_date: str,
    nights: int,
    guests: int = 1,
) -> str:
    """Book a hotel room"""
    return f"Booked {nights} nights in {location}"
```

**Generated schema (automatic):**
```json
{
  "type": "object",
  "properties": {
    "location": {"type": "string"},
    "checkin_date": {"type": "string"},
    "nights": {"type": "integer"},
    "guests": {"type": "integer", "default": 1}
  },
  "required": ["location", "checkin_date", "nights"]
}
```

✅ **Result:** Zero manual work, schema always matches code

---

### go-agentic Current - Manual JSON Schema

```go
func BookHotel(ctx context.Context, args map[string]interface{}) (string, error) {
    location, _ := args["location"].(string)
    checkinDate, _ := args["checkin_date"].(string)
    nights, _ := args["nights"].(float64)
    guests, ok := args["guests"].(float64)
    if !ok {
        guests = 1
    }
    // ...
}

toolsMap["BookHotel"] = &Tool{
    // ...
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type": "string",
            },
            "checkin_date": map[string]interface{}{
                "type": "string",
            },
            "nights": map[string]interface{}{
                "type": "number",
            },
            "guests": map[string]interface{}{
                "type": "number",
                "default": 1,
            },
        },
        "required": []string{"location", "checkin_date", "nights"},
    },
}
```

❌ **Problem:**
- Schema and code can diverge
- Easy to forget to update schema when function changes
- Type conversions manual (float64 vs int for nights)
- Verbose and error-prone

---

### go-agentic Future - Struct-Based (Like FastAPI)

```go
type BookHotelParams struct {
    Location    string `json:"location" description:"City and state"`
    CheckinDate string `json:"checkin_date" description:"YYYY-MM-DD format"`
    Nights      int    `json:"nights" description:"Number of nights" minimum:"1"`
    Guests      int    `json:"guests" default:"1" description:"Number of guests"`
}

func BookHotel(ctx context.Context, params BookHotelParams) (string, error) {
    // params.Nights is already an int
    // All parameters validated by framework
    return fmt.Sprintf("Booked %d nights in %s", params.Nights, params.Location), nil
}

// Framework automatically:
// ✅ Generates JSON schema from struct
// ✅ Validates before calling function
// ✅ Type-safe parameter conversion
// ✅ Clear error messages
```

✅ **Result:** Single source of truth (struct = schema)

---

## 3. Error Handling

### Best Practice (Anthropic SDK)

```python
# Tool returns error
@tool
def get_user(user_id: str) -> dict:
    user = db.get_user(user_id)  # May raise exception
    if not user:
        raise ValueError(f"User {user_id} not found")  # Framework sees this
    return user

# Framework:
# 1. Catches error
# 2. Formats error message
# 3. Sends to tool_runner
# 4. tool_runner sends to Claude
# 5. Claude sees error and retries
```

✅ **Result:** Automatic error propagation, LLM can retry

---

### go-agentic Current ❌

```go
func GetUser(ctx context.Context, args map[string]interface{}) (string, error) {
    userID, _ := args["user_id"].(string)
    user, err := db.GetUser(userID)
    if err != nil {
        return "", err  // Returns error
    }
    return user, nil
}

// In executor:
result, err := tool.Execute(...)
if err != nil {
    log.Printf("Tool error: %v", err)  // ← Logs but doesn't send to LLM
    // Workflow continues
}
```

❌ **Problem:**
- Error logged but not sent to LLM
- LLM doesn't know tool failed
- LLM can't retry
- Silent failure

---

### go-agentic Future ✅

```go
func GetUser(ctx context.Context, params GetUserParams) (string, error) {
    user, err := db.GetUser(params.UserID)
    if err != nil {
        return "", err
    }
    return user, nil
}

// In executor:
result, err := tool.Execute(...)
if err != nil {
    // ✅ Send error to LLM in history
    history.Add(Message{
        Role: "system",
        Content: fmt.Sprintf("Tool '%s' failed: %v\nPlease try again with different parameters.", toolName, err),
    })
    // LLM sees error and retries automatically
}
```

✅ **Result:** Errors propagated, LLM can retry

---

## 4. Registration Method

### LangChain (Simple) ⭐

```python
tools = [get_weather, book_hotel, get_user]
agent = create_agent(tools=tools)
```

✅ Simple list of functions

---

### go-agentic Current (Complex) ❌

```go
// Way 1: Using loader
loader := core.NewLoader("./config")
loader.RegisterTool("GetWeather", GetWeather)
loader.RegisterTool("BookHotel", BookHotel)
loader.RegisterTool("GetUser", GetUser)
crew, _ := loader.LoadCrew()

// Way 2: Using map
toolsMap := map[string]*Tool{
    "GetWeather": {...},
    "BookHotel": {...},
    "GetUser": {...},
}
executor, _ := core.NewCrewExecutorFromConfig(apiKey, "config", toolsMap)

// Way 3: YAML
// agents/agent.yaml:
// tools:
//   - GetWeather
//   - BookHotel
//   - GetUser
```

❌ Multiple ways, unclear which is "correct"

---

### go-agentic Future (Simple) ✅

```go
registry := core.NewToolRegistry()
registry.Add(GetWeather)
registry.Add(BookHotel)
registry.Add(GetUser)

// Or:
tools := core.NewToolRegistry().
    Add(GetWeather).
    Add(BookHotel).
    Add(GetUser).
    Build()

executor, _ := core.NewCrewExecutorFromConfig(apiKey, "config", tools)
```

✅ Single, obvious method

---

## 5. Validation

### Best Practice (FastAPI)

```python
class GetWeatherParams(BaseModel):
    location: str = Field(..., min_length=1, max_length=100)
    unit: str = Field(default="fahrenheit", pattern="^(celsius|fahrenheit)$")

@app.post("/")
def get_weather(params: GetWeatherParams):
    # params already validated by FastAPI
    return weather
```

**Automatic validation:**
- ✅ location is not empty (min_length=1)
- ✅ location is max 100 chars
- ✅ unit is one of valid values
- ✅ Clear 422 error if invalid

---

### go-agentic Current ❌

```go
func GetWeather(ctx context.Context, args map[string]interface{}) (string, error) {
    location, ok := args["location"].(string)
    if !ok {
        return "", fmt.Errorf("location must be string")
    }
    if len(location) == 0 {
        return "", fmt.Errorf("location cannot be empty")
    }
    if len(location) > 100 {
        return "", fmt.Errorf("location must be < 100 chars")
    }

    unit, ok := args["unit"].(string)
    if !ok {
        unit = "fahrenheit"
    }
    if unit != "celsius" && unit != "fahrenheit" {
        return "", fmt.Errorf("unit must be celsius or fahrenheit")
    }
    // ... 20+ LOC of validation before actual logic
}
```

❌ Problems:
- Lots of boilerplate (20+ LOC)
- Easy to make mistakes
- Validation code mixed with business logic
- No single schema definition

---

### go-agentic Future ✅

```go
type GetWeatherParams struct {
    Location string `json:"location" minLength:"1" maxLength:"100" description:"City"`
    Unit     string `json:"unit" enum:"celsius,fahrenheit" default:"fahrenheit"`
}

func GetWeather(ctx context.Context, params GetWeatherParams) (string, error) {
    // Framework already validated:
    // ✅ location is not empty
    // ✅ location is < 100 chars
    // ✅ unit is valid

    // Just implement business logic
    return getWeatherData(params.Location, params.Unit), nil
}
```

✅ Result:
- No validation boilerplate
- Single source of truth (struct)
- Clear contracts

---

## 6. Configuration Validation

### Best Practice (gRPC)

```protobuf
service QuizService {
  rpc AskQuestion(QuestionRequest) returns (Question) {}
  rpc RecordAnswer(AnswerRequest) returns (AnswerResponse) {}
}

message QuestionRequest {
  string quiz_id = 1;
  int32 question_number = 2;
}
```

**Compile time:**
- ✅ Protoc validates schema syntax
- ✅ Generates server interface
- ✅ Type checking

---

### go-agentic Current ❌

```go
// agents/teacher.yaml
tools:
  - GetQuizStatus
  - RecordAnswer
  - WriteReport

// If you forget to register RecordAnswer in Go:
// No error!
// Tool just silently unavailable
// LLM can't find it
// Workflow fails mysteriously
```

❌ **Problem:** No validation that tools exist

---

### go-agentic Future ✅

```go
executor, err := core.NewCrewExecutorFromConfig(apiKey, "config", toolsMap)
if err != nil {
    // Error: Tool 'RecordAnswer' referenced in teacher.yaml
    //        but not registered
    // Available tools: GetQuizStatus, WriteReport
    return
}
```

✅ **Result:** Fail-fast with clear error

---

## 7. Comparison Table

```
╔═══════════════════════╦═════════════╦══════════════╦═══════════════╗
║ Feature               ║ Current DX  ║ Best Practice║ Future go-ag  ║
╠═══════════════════════╬═════════════╬══════════════╬═══════════════╣
║ Tool Definition       ║ 4 places    ║ 1 place      ║ 1 place       ║
║ Parameters            ║ map[string] ║ Type hints   ║ Struct        ║
║ Validation            ║ Manual      ║ Auto         ║ Auto          ║
║ Schema Definition     ║ Hand-written║ Auto-gen     ║ Auto-gen      ║
║ Registration          ║ 2 ways      ║ 1 way        ║ 1 way         ║
║ Error Handling        ║ Silent log  ║ Propagate   ║ Propagate     ║
║ Validation Errors     ║ Manual      ║ Auto + clear ║ Auto + clear  ║
║ LOC per Tool          ║ 40+         ║ 4-10         ║ 10-15         ║
║ Configuration Validation ║ None     ║ Compile-time ║ Load-time     ║
║ Silent Failures       ║ Many        ║ None         ║ None          ║
║ Type Safety           ║ Low         ║ High         ║ High          ║
║ DX Score              ║ 6.5/10      ║ 8.5-9.5/10   ║ 8.5+/10       ║
╚═══════════════════════╩═════════════╩══════════════╩═══════════════╝
```

---

## Key Takeaways

### What Best Practices Show

1. **Struct-based parameters are standard**
   - Anthropic SDK: Type hints
   - FastAPI: Pydantic models
   - gRPC: Proto messages
   - ✅ go-agentic should use structs with tags

2. **Schemas should be auto-generated**
   - Type hints → JSON schema (Anthropic)
   - Pydantic models → OpenAPI (FastAPI)
   - Proto definitions → Code (gRPC)
   - ❌ Never hand-written JSON

3. **Single registration method**
   - LangChain: Pass list of functions
   - Anthropic: Pass functions to tool_runner
   - ❌ Never multiple confusing ways

4. **Validation should be automatic**
   - Framework validates before calling
   - Clear errors if invalid
   - ❌ Developer shouldn't write validation code

5. **Errors should propagate**
   - Tool errors sent back to caller (LLM)
   - Caller can retry
   - ❌ Never silent failures

6. **Configuration should be validated**
   - Fail at load time, not runtime
   - Clear error messages
   - ❌ Never silent mismatches

### What go-agentic Must Do

| Best Practice | go-agentic Status | Action Needed |
|---|---|---|
| Struct-based params | ❌ map[string] | Implement StructuredTool |
| Auto schemas | ❌ Hand-written | Schema generator |
| Single registration | ❌ 2-3 methods | Unified ToolRegistry |
| Auto validation | ❌ Manual code | Framework validation |
| Error propagation | ❌ Silent log | Send errors to history |
| Load-time validation | ❌ None | Config validator |

---

## Implementation Priority

Based on impact and best practice standards:

1. **P0 (Critical):** Struct-based parameters + schema generation
2. **P0 (Critical):** Error propagation to LLM
3. **P1 (High):** Load-time configuration validation
4. **P1 (High):** Unified tool registration
5. **P2 (Medium):** Documentation improvements
6. **P3 (Nice):** Unified routing system

---

## Next Steps

1. Implement struct-based tool parameters (Phase 1)
2. Add auto schema generation (Phase 2)
3. Implement fail-fast validation (Phase 3)
4. Add error propagation (Phase 4)
5. Update documentation and examples (Phase 5)
6. Refactor examples to show new pattern

See **DX_IMPROVEMENT_ROADMAP.md** for detailed implementation plan.

---

**Generated:** 2025-12-25
**Version:** 1.0
