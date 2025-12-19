# Go-Agentic Quick Reference

## 📋 Project Organization

```
.
├── go-agentic/              # Core library (pure Go)
│   ├── *.go                 # 10 source files
│   ├── go.mod + go.sum      # Module definition
│   ├── docs/                # Extended documentation
│   ├── LIBRARY_STRUCTURE.md # Detailed guide
│   ├── ARCHITECTURE.txt     # Visual diagrams
│   └── README.md            # Library README
│
├── examples/                # Example applications
│   ├── customer-service/
│   ├── data-analysis/
│   ├── it-support/
│   └── research-assistant/
│
├── LIBRARY_TREE.md          # Complete project tree
├── STRUCTURE.md             # Project structure guide
├── README.md                # Main README
└── go.mod + go.sum          # Root module (examples support)
```

## 🚀 Quick Commands

### Build Library
```bash
cd go-agentic
go build ./...
```

### Build Examples
```bash
cd examples/customer-service
go build -o customer-service-example ./main.go ./example_customer_service.go
```

### Run Example
```bash
export OPENAI_API_KEY=sk-...
./customer-service-example
```

## 📚 Documentation

| File | Purpose |
|------|---------|
| `go-agentic/README.md` | Library overview and features |
| `go-agentic/LIBRARY_STRUCTURE.md` | Detailed architecture (recommended read) |
| `go-agentic/ARCHITECTURE.txt` | Visual diagrams and data flows |
| `LIBRARY_TREE.md` | Complete project tree with descriptions |
| `STRUCTURE.md` | Project root organization |
| `examples/README.md` | Examples guide |

## 💻 Core Library Files

```
types.go       → Data types (Agent, Crew, Tool, etc.)
agent.go       → Agent execution with LLM
crew.go        → Multi-agent orchestration
config.go      → YAML configuration loading
http.go        → HTTP server with SSE
html_client.go → Web testing UI
streaming.go   → Server-Sent Events utilities
tests.go       → Test framework and scenarios
report.go      → HTML report generation
go.mod         → Module definition
go.sum         → Dependencies
```

## 🔌 Import Path

```go
import "github.com/taipm/go-agentic"

// Usage
crew := &agentic.Crew{...}
agent := &agentic.Agent{...}
executor := agentic.NewCrewExecutor(crew, apiKey)
```

## 📊 Statistics

- **Library Files**: 10 core Go files
- **Library Size**: ~2,565 lines of code
- **Documentation**: 16 files (3 main + 14 extended)
- **Examples**: 4 real-world use cases
- **Dependencies**: OpenAI SDK + yaml.v3

## ✅ Quick Checks

- [x] Library builds cleanly
- [x] All 4 examples build successfully
- [x] No breaking changes
- [x] Backward compatible
- [x] Production ready
- [x] Comprehensive documentation

## 🎯 Next Steps

1. **Understand Structure**
   - Read: `go-agentic/LIBRARY_STRUCTURE.md`
   - Review: `go-agentic/ARCHITECTURE.txt`

2. **Try Examples**
   - Choose an example in `examples/`
   - Read its README.md
   - Build and run it

3. **Use in Your Project**
   - Import: `github.com/taipm/go-agentic`
   - Create agents and tools
   - Execute crew on your tasks

## 📖 Full Documentation

For complete information, see:
- **Architecture Details**: `go-agentic/LIBRARY_STRUCTURE.md`
- **Visual Diagrams**: `go-agentic/ARCHITECTURE.txt`
- **Project Tree**: `LIBRARY_TREE.md`
- **Extended Docs**: `go-agentic/docs/` (14 files)

## 💡 Key Features

✅ Multi-agent orchestration
✅ Real-time SSE streaming
✅ Intelligent agent routing
✅ Web testing UI
✅ Configuration-driven setup
✅ Comprehensive testing framework
✅ Production-ready error handling
✅ Clean, professional code structure

---

For more details, see the comprehensive documentation in the `docs/` directory.
