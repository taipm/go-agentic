# Simple Chat Example - Extremely Minimalist Version

## 🚀 Why This Example is Super Simple

**main.go**: Only ~60 lines of code!

```go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"github.com/taipm/go-agentic"
	"gopkg.in/yaml.v3"
)

func main() {
	// 1. Load .env
	loadEnv()
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		fmt.Println("❌ Missing API key")
		os.Exit(1)
	}

	// 2. Parse crew.yaml
	var cfg struct {
		Crew struct {
			MaxRounds   int `yaml:"maxRounds"`
			MaxHandoffs int `yaml:"maxHandoffs"`
		} `yaml:"crew"`
		Agents []struct {
			ID          string  `yaml:"id"`
			Name        string  `yaml:"name"`
			Role        string  `yaml:"role"`
			Backstory   string  `yaml:"backstory"`
			Model       string  `yaml:"model"`
			Temperature float64 `yaml:"temperature"`
			IsTerminal  bool    `yaml:"isTerminal"`
		} `yaml:"agents"`
		Topics []string `yaml:"topics"`
	}

	data, _ := os.ReadFile("crew.yaml")
	yaml.Unmarshal(data, &cfg)

	// 3. Create agents
	agents := make([]*agentic.Agent, len(cfg.Agents))
	for i, a := range cfg.Agents {
		agents[i] = &agentic.Agent{
			ID: a.ID, Name: a.Name, Role: a.Role, Backstory: a.Backstory,
			Model: a.Model, Temperature: a.Temperature, IsTerminal: a.IsTerminal,
			Tools: []*agentic.Tool{},
		}
	}

	// 4. Run crew
	crew := &agentic.Crew{Agents: agents, MaxRounds: cfg.Crew.MaxRounds, MaxHandoffs: cfg.Crew.MaxHandoffs}
	executor := agentic.NewTeamExecutor(crew, key)

	// 5. Discuss topics
	fmt.Println("\n🤖 Multi-Agent Chat\n" + strings.Repeat("=", 50))
	for i, topic := range cfg.Topics {
		fmt.Printf("\n📌 Topic %d: %s\n", i+1, topic)
		if response, err := executor.Execute(context.Background(), topic); err == nil {
			fmt.Printf("✅ Result:\n%s\n", response)
		}
	}
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")
}

func loadEnv() {
	data, _ := os.ReadFile(".env")
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}
}
```

## ✅ What It Does

1. **Loads .env** - Gets your OpenAI API key
2. **Parses crew.yaml** - Reads agent definitions and topics
3. **Creates agents** - Builds Agent objects from config
4. **Runs executor** - Starts multi-agent conversations
5. **Prints results** - Shows agent conversations

## 🎯 Why It's Simple

✅ **Single file** - No separate modules or packages
✅ **Inline structs** - Config parsing right in main()
✅ **~60 lines** - Everything visible at once
✅ **No boilerplate** - No unnecessary abstractions
✅ **Clear logic flow** - Easy to follow from top to bottom
✅ **YAML config** - Customize without touching code

## 💡 Key Learning Points

For beginners learning go-agentic:

1. **How to structure basic multi-agent systems**
2. **Loading configuration from files**
3. **Creating agents dynamically**
4. **Executing team conversations**
5. **Handling real API integration**

## 🚀 How to Run

```bash
# Copy template
cp .env.example .env

# Edit .env with your API key
# Then:
go run main.go

# That's it! Agents will speak in Vietnamese
```

## 📊 Comparison

| Aspect | Version |
|--------|---------|
| Lines of code | ~60 |
| Main functions | 2 (main + loadEnv) |
| Type definitions | 0 (inline structs) |
| Helper functions | 1 |
| Files included | 1 (main.go) |
| Readability | ⭐⭐⭐⭐⭐ |
| Learning curve | Very gentle |

## 🔑 Core Concepts Demonstrated

**1. Configuration Management**
```go
yaml.Unmarshal(data, &cfg)  // Parse YAML
```

**2. Dynamic Agent Creation**
```go
agents[i] = &agentic.Agent{ ... }  // Create from config
```

**3. Team Execution**
```go
executor.Execute(context.Background(), topic)  // Run conversation
```

**4. Error Handling**
```go
if key == "" { os.Exit(1) }  // Basic validation
```

## 🎓 For Different Audiences

**Complete Beginners:**
- Read main() from top to bottom
- Understand basic flow
- No need to understand complex patterns

**Intermediate Developers:**
- See how configuration management works
- Learn about dynamic object creation
- Understand type assertions and inline structs

**Advanced Users:**
- Base for building production systems
- Extend with error handling
- Add more sophisticated features

## 🔧 How to Customize

**Change topics** (edit crew.yaml):
```yaml
topics:
  - "Your topic"
  - "Another topic"
```

**Change agent personality** (edit crew.yaml):
```yaml
agents:
  - id: "expert"
    name: "New Name"
    backstory: "New backstory in Vietnamese"
```

**Add more agents** (edit crew.yaml):
```yaml
agents:
  - id: "agent1"
    ...
  - id: "agent2"
    ...
  - id: "agent3"
    ...
```

## ❓ Why No Separate Files?

Other examples might split code into:
- `config.go` - Configuration handling
- `agents.go` - Agent creation
- `main.go` - Main logic

This example keeps everything in one file because:
✅ Easier to learn (single file)
✅ No package management overhead
✅ Can see entire flow at once
✅ Perfect for prototyping

## 🚀 Next Steps

After mastering this simple example:

1. **Add error handling** - Better error messages
2. **Add logging** - Track what's happening
3. **Split into files** - Organize for larger projects
4. **Add more features** - Tools, custom prompts, etc.
5. **Deploy** - Run on servers, in production

## 📚 Resources

- `crew.yaml` - Configuration file
- `GETTING_STARTED.md` - Setup guide
- `/SECURITY.md` - Security best practices
- Main README - Full documentation

## 🎉 Summary

This is the **simplest, most readable** go-agentic example possible while being completely functional. It demonstrates:

✅ How to use go-agentic
✅ Configuration management
✅ Multi-language support (Vietnamese)
✅ Agent personality through parameters
✅ Real API integration

Perfect for learning! 🚀
