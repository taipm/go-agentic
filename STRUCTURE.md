# Go-Agentic Project Structure

## Directory Organization

After reorganization, the project now has the following structure:

```
go-agentic/                          # Root directory
├── go-agentic/                      # Core library package
│   ├── types.go                    # Core data types (Agent, Crew, Tool, etc.)
│   ├── agent.go                    # Agent execution logic
│   ├── crew.go                     # Crew orchestration engine
│   ├── http.go                     # HTTP server with SSE streaming
│   ├── streaming.go                # Streaming utilities
│   ├── config.go                   # Configuration loading
│   ├── tests.go                    # Testing utilities
│   ├── report.go                   # Test report generation
│   ├── go.mod                      # Library module definition
│   └── go.sum                      # Library dependencies
│
├── examples/                        # Example applications (at root level)
│   ├── README.md                   # Examples documentation
│   ├── it-support/                 # IT Support crew
│   │   ├── main.go                # Interactive CLI
│   │   ├── example_it_support.go   # Crew definition & tools
│   │   ├── test.go                # Testing utilities
│   │   ├── config/                # Configuration files
│   │   └── it-support-example     # Built executable
│   │
│   ├── customer-service/           # Customer Service crew
│   │   ├── main.go                # Interactive CLI
│   │   ├── example_customer_service.go  # Crew definition & tools
│   │   └── customer-service-example    # Built executable
│   │
│   ├── data-analysis/              # Data Analysis crew
│   │   ├── main.go                # Interactive CLI
│   │   ├── example_data_analysis.go # Crew definition & tools
│   │   └── data-analysis-example   # Built executable
│   │
│   └── research-assistant/         # Research Assistant crew
│       ├── main.go                # Interactive CLI
│       ├── example_research_assistant.go # Crew definition & tools
│       └── research-assistant-example   # Built executable
│
├── go.mod                          # Root module (enables examples to build)
├── go.sum                          # Root dependencies
├── README.md                       # Main project documentation
└── STRUCTURE.md                    # This file
```

## Key Points

### Root-Level Examples
- All examples are now located in `./examples/[example-name]/`
- This allows cleaner organization and easier discovery
- Examples can be built independently from their own directories

### Module Structure
- `go-agentic/` directory contains the core library
- Root-level `go.mod` uses a `replace` directive to reference the local `go-agentic` module
- This setup allows examples to import the library using `github.com/taipm/go-agentic`

### Building Examples
```bash
cd examples/[example-name]
go build -o [example-name]-example ./main.go ./example_*.go
```

### Example Features
Each example follows the same pattern:

1. **main.go** - Entry point with CLI interface
   - Loads `.env` file for OPENAI_API_KEY
   - Creates crew and executor
   - Runs interactive loop

2. **example_*.go** - Crew definition
   - `createXxxCrew()` - Creates the crew with agents
   - `createXxxTools()` - Defines all tools for the crew
   - Tool implementations with context-based handlers

3. **Agent Pipeline**
   - Agents work together in defined sequence
   - Orchestration enables handoffs between agents
   - Terminal agents mark end of workflow

## Moving Examples to Root

This reorganization moved examples from `go-agentic/examples/` to `./examples/` for:
- Better discoverability in the project root
- Easier access when exploring the repository
- Cleaner separation between library and examples
- Improved user experience for new contributors

## Import Paths

All examples use absolute imports:
```go
import "github.com/taipm/go-agentic"
```

This works because the root-level `go.mod` defines a local `replace` directive:
```
replace github.com/taipm/go-agentic => ./go-agentic
```
