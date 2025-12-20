# Simple Chat v2: Fluent Builder API Example

Ví dụ này minh họa cách sử dụng **Fluent Builder API** mới của go-agentic - một cách dễ đọc, gọn gàng để tạo agents và teams.

## ✨ Điểm Nổi Bật

### Fluent Builder API (Phase 1 Improvement)

**Trước (Verbose)**:
```go
agent := &agentic.Agent{
    ID: "expert",
    Name: "Chuyên Gia",
    Role: "Expert",
    Backstory: "...",
    Model: "gpt-4o-mini",
    Temperature: 0.7,
    IsTerminal: true,
    Tools: []*agentic.Tool{},
}
```

**Sau (Clean & Fluent)**:
```go
expert := agentic.NewAgent("expert", "Chuyên Gia").
    WithRole("Expert").
    WithBackstory("...").
    WithModel("gpt-4o-mini").
    WithTemperature(0.7).
    SetTerminal(true).
    Build()
```

### Lợi Ích

✅ **Dễ đọc**: Left-to-right fluent style
✅ **Gọn gàng**: Ít boilerplate
✅ **Type-safe**: Validation built-in
✅ **Chained**: Composition-friendly
✅ **Flexible**: Defaults overridable

## 🚀 Quick Start

### 1. Setup

```bash
cp .env.example .env
# Edit .env and add your OpenAI API key
```

### 2. Run

```bash
go run main.go
```

## 📝 Example Structure

```go
// 1. Create first agent using fluent builder
enthusiast := agentic.NewAgent("id", "Name").
    WithRole("Role description").
    WithBackstory("Multi-line backstory...").
    WithModel("gpt-4o-mini").
    WithTemperature(0.8).
    SetTerminal(false).
    WithHandoff("other-agent").
    Build()

// 2. Create second agent
expert := agentic.NewAgent("expert", "Expert").
    WithRole("Expert role").
    WithBackstory("Expert backstory...").
    SetTerminal(true).
    Build()

// 3. Create team using fluent API
team := agentic.NewTeam().
    AddAgents(enthusiast, expert).
    WithMaxRounds(4).
    WithMaxHandoffs(3).
    Build()

// 4. Execute
executor := agentic.NewTeamExecutor(team, apiKey)
response, _ := executor.Execute(ctx, "Your question")
```

## 🎯 Builder Methods Available

### AgentBuilder
```go
NewAgent(id, name)              // Create builder
.WithRole(role)                 // Set role
.WithBackstory(backstory)       // Set backstory
.WithModel(model)               // Set model (not hardcoded!)
.WithTemperature(temp)          // Set temperature
.WithSystemPrompt(prompt)       // Custom system prompt
.AddTool(tool)                  // Add single tool
.AddTools(tool1, tool2...)      // Add multiple tools
.SetTerminal(true/false)        // Mark as terminal
.WithHandoff(target)            // Add handoff target
.WithHandoffs(target1, ...)     // Add multiple targets
.Build()                        // Validate & return
```

### ToolBuilder
```go
NewTool(name, description)                  // Create builder
.NoParameters()                             // No params
.WithParameter(name, type, desc, required) // Add param
.WithParameters(map[...])                   // Set all params
.Handler(func)                              // Set handler
.Build()                                    // Validate & return
```

### TeamBuilder
```go
NewTeam()                       // Create builder
.AddAgent(agent)                // Add single agent
.AddAgents(a1, a2...)           // Add multiple agents
.WithMaxRounds(count)           // Max rounds
.WithMaxHandoffs(count)         // Max handoffs
.Build()                        // Validate & return
```

## ✅ Validation Built-In

Builders automatically validate:

```go
// ❌ These will panic with clear messages:
NewAgent("", "name").Build()                        // Missing ID
NewAgent("id", "name").Build()                      // Missing Role
NewAgent("id", "name").WithRole("r").Build()        // Missing Backstory
NewTool("", "desc").Handler(fn).Build()             // Missing Name
NewTool("name", "desc").Build()                     // Missing Handler
NewTeam().Build()                                   // No agents
NewTeam().AddAgent(agent).Build()                   // No terminal agent
```

## 🆚 Compared to YAML Config (simple-chat)

| Feature | Fluent v2 | YAML (simple-chat) |
|---------|-----------|------------------|
| **Type Safety** | ✅ Compile-time | ❌ Runtime |
| **IDE Support** | ✅ Full autocomplete | ❌ No autocomplete |
| **Validation** | ✅ Automatic | ❌ Manual |
| **Flexibility** | ✅ All Go features | ⚠️ Limited to YAML |
| **Learning Curve** | ✅ Familiar patterns | ⚠️ New schema |
| **Configuration** | ❌ Not supported | ✅ Easy to change |

**Best for**: Fluent v2 for complex logic, simple-chat YAML for simple static configs.

## 📊 Code Comparison

### Total Lines
- **Fluent v2**: ~60 lines (pure Go)
- **YAML**: ~80 lines (split across files)

### Boilerplate
- **Fluent v2**: None (builders hide it)
- **YAML**: 40+ lines of config structure

## 🔄 Phase 1 of UX Improvements

This example demonstrates **Phase 1** of the go-agentic UX improvements:

✅ **Phase 1: Fluent Builder API** (This example)
- AgentBuilder, ToolBuilder, TeamBuilder
- 50-80% boilerplate reduction
- Clean, readable code

⏭️ **Phase 2: Unified YAML Configuration** (Coming)
- Single team.yaml file
- 75% fewer config files

⏭️ **Phase 3: Declarative Routing** (Coming)
- Clear routing rules
- 80% less system prompt code

---

**Go-agentic Version**: v0.0.1-alpha.1 (Phase 1 compatible)
**Created**: December 20, 2025
