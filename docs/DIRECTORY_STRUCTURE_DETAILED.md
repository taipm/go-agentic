# 📂 CẤU TRÚC THƯ MỤC CHI TIẾT SAU TÁCH (Visual Tree)

## 🎯 TỔNG QUAN CẤU TRÚC

```
go-agentic-monorepo/
│
├─ 📦 PHẦN 1: GO-CREWAI (CORE LIBRARY - LÕIVERSION)
│  │
│  ├── go.mod                                # module: github.com/taipm/go-crewai
│  ├── go.sum                                # Dependencies lock
│  ├── LICENSE                               # MIT or similar
│  │
│  ├── 🔧 CORE SOURCE CODE (2,384 lines - immutable)
│  │   ├── types.go                      [84]    # Data types
│  │   ├── agent.go                      [234]   # Agent engine
│  │   ├── crew.go                       [398]   # Orchestrator
│  │   ├── config.go                     [169]   # Config loader
│  │   ├── http.go                       [187]   # HTTP server
│  │   ├── streaming.go                  [54]    # SSE handler
│  │   ├── html_client.go                [252]   # Web UI template
│  │   ├── report.go                     [696]   # HTML report gen
│  │   └── tests.go                      [316]   # Test utilities
│  │
│  ├── 📚 DOCUMENTATION
│  │   ├── README.md                          # Library homepage
│  │   ├── QUICKSTART.md                      # Get started in 5 min
│  │   ├── ARCHITECTURE.md                    # System design
│  │   ├── API_REFERENCE.md                   # All exported types/funcs
│  │   │
│  │   └── docs/
│  │       ├── LIBRARY_INTRO.md               # Why go-crewai?
│  │       ├── STREAMING_GUIDE.md             # SSE real-time events
│  │       ├── TOOLS_DOCUMENTATION.md         # How to write tools
│  │       ├── CONFIG_SCHEMA.md               # YAML schema reference
│  │       ├── PLUGIN_GUIDE.md                # Extension points
│  │       ├── BEST_PRACTICES.md              # Do's and Don'ts
│  │       ├── TROUBLESHOOTING.md             # Common issues
│  │       └── MIGRATION.md                   # From Python CrewAI
│  │
│  ├── 📋 TEMPLATE EXAMPLES (Not runnable - for reference)
│  │   ├── examples/
│  │   │   ├── README.md                      # How to use templates
│  │   │   │
│  │   │   ├── minimal-main.go.template      # Minimal example
│  │   │   ├── full-main.go.template         # Full-featured
│  │   │   │
│  │   │   ├── crew.yaml.template            # Minimal crew
│  │   │   ├── crew-advanced.yaml.template   # Advanced crew
│  │   │   │
│  │   │   ├── agents/
│  │   │   │   ├── agent1.yaml.template      # Simple agent
│  │   │   │   ├── agent2.yaml.template      # Agent with system prompt
│  │   │   │   └── agent3.yaml.template      # Agent with tools
│  │   │   │
│  │   │   └── tools/
│  │   │       ├── simple_tools.go.template
│  │   │       ├── external_api_tools.go.template
│  │   │       └── database_tools.go.template
│  │   │
│  │   └── sample_project/                   # Template project structure
│  │       ├── README.md
│  │       ├── main.go
│  │       ├── crew.go
│  │       ├── tools.go
│  │       ├── config/
│  │       │   ├── crew.yaml
│  │       │   └── agents/
│  │       │       ├── agent1.yaml
│  │       │       └── agent2.yaml
│  │       └── go.mod.template
│  │
│  └── 🧪 TESTS
│      ├── *_test.go                         # Unit tests
│      ├── testdata/
│      │   ├── sample_crews.yaml
│      │   ├── sample_agents.yaml
│      │   └── mock_responses.json
│      └── integration/
│          └── integration_test.go


├─ 🚀 PHẦN 2: GO-AGENTIC-EXAMPLES (EXAMPLE APPLICATIONS)
│  │
│  ├── go.mod                                # module: github.com/taipm/go-agentic-examples
│  ├── go.sum
│  ├── README.md                             # Overview of all examples
│  ├── LICENSE
│  │
│  ├── 📖 DOCUMENTATION
│  │   ├── EXAMPLES_INDEX.md                 # Index: What examples available?
│  │   ├── QUICK_START.md                    # "Copy & Run" guide
│  │   ├── PATTERNS.md                       # Recurring patterns
│  │   │
│  │   └── examples/
│  │       ├── IT_SUPPORT.md                 # Detailed walkthrough
│  │       │   ├── Architecture
│  │       │   ├── Tools explained
│  │       │   ├── Config explained
│  │       │   └── Extension guide
│  │       │
│  │       ├── CUSTOMER_SERVICE.md           # Detailed walkthrough
│  │       ├── RESEARCH_ASSISTANT.md         # Detailed walkthrough
│  │       └── DATA_ANALYSIS.md              # Detailed walkthrough
│  │
│  ├── 📁 EXAMPLE 1: IT SUPPORT SYSTEM
│  │   │
│  │   └── it-support/
│  │       │
│  │       ├── cmd/
│  │       │   ├── main.go                   # Entry point
│  │       │   └── server.go                 # Optional: server mode
│  │       │
│  │       ├── internal/
│  │       │   ├── crew.go                   # CreateITSupportCrew()
│  │       │   ├── tools.go                  # Define all 8+ IT tools
│  │       │   │   ├── cpu_tool.go
│  │       │   │   ├── memory_tool.go
│  │       │   │   ├── disk_tool.go
│  │       │   │   ├── network_tool.go
│  │       │   │   ├── process_tool.go
│  │       │   │   ├── service_tool.go
│  │       │   │   └── system_info_tool.go
│  │       │   │
│  │       │   ├── config.go                 # Config loading
│  │       │   ├── handlers.go               # HTTP handlers
│  │       │   └── reporters.go              # Custom report formatting
│  │       │
│  │       ├── config/
│  │       │   ├── crew.yaml                 # IT crew definition
│  │       │   │   └─ entry_point: orchestrator
│  │       │   │   └─ agents: [orchestrator, clarifier, executor]
│  │       │   │   └─ routing: [signals → targets]
│  │       │   │
│  │       │   └── agents/
│  │       │       ├── orchestrator.yaml     # Routing agent
│  │       │       │   └─ Role: Analyze & route
│  │       │       │   └─ System prompt (Vietnamese)
│  │       │       │   └─ Handoff targets: [clarifier, executor]
│  │       │       │
│  │       │       ├── clarifier.yaml        # Info gathering
│  │       │       │   └─ Role: Ask clarifying questions
│  │       │       │   └─ No tools
│  │       │       │
│  │       │       └── executor.yaml         # Problem solving
│  │       │           └─ Role: Run diagnostics
│  │       │           └─ Tools: All 8+ IT tools
│  │       │           └─ IsTerminal: true
│  │       │
│  │       ├── web/                          # Optional: Web UI
│  │       │   ├── index.html
│  │       │   ├── styles.css
│  │       │   └── client.js                 # SSE client
│  │       │
│  │       ├── tests/
│  │       │   ├── crew_test.go
│  │       │   ├── tools_test.go
│  │       │   ├── integration_test.go
│  │       │   └── fixtures/
│  │       │       └── sample_responses.json
│  │       │
│  │       ├── .env.example                  # OPENAI_API_KEY=...
│  │       ├── Makefile                      # make run, make test, etc
│  │       ├── README.md                     # IT Support specific docs
│  │       │   ├── What is IT Support?
│  │       │   ├── How to run
│  │       │   ├── Configuration
│  │       │   ├── Custom tools
│  │       │   ├── Examples
│  │       │   └── Troubleshooting
│  │       │
│  │       └── demo.sh                       # Interactive demo script
│  │
│  ├── 📁 EXAMPLE 2: CUSTOMER SERVICE SYSTEM
│  │   │
│  │   └── customer-service/
│  │       │
│  │       ├── cmd/
│  │       │   ├── main.go
│  │       │   └── server.go
│  │       │
│  │       ├── internal/
│  │       │   ├── crew.go                   # CreateCustomerServiceCrew()
│  │       │   ├── tools.go                  # CRM, ticket, FAQ, email tools
│  │       │   │   ├── crm_tool.go
│  │       │   │   ├── ticket_tool.go
│  │       │   │   ├── faq_tool.go
│  │       │   │   ├── email_tool.go
│  │       │   │   └── knowledge_base_tool.go
│  │       │   │
│  │       │   ├── config.go
│  │       │   ├── handlers.go
│  │       │   └── formatters.go
│  │       │
│  │       ├── config/
│  │       │   ├── crew.yaml
│  │       │   └── agents/
│  │       │       ├── intake.yaml            # Receive customer inquiry
│  │       │       ├── analyzer.yaml          # Analyze & search KB
│  │       │       ├── resolver.yaml          # Create ticket/response
│  │       │       └── escalation.yaml        # Handle complex cases
│  │       │
│  │       ├── web/
│  │       │   ├── index.html
│  │       │   └── client.js
│  │       │
│  │       ├── tests/
│  │       ├── .env.example
│  │       ├── Makefile
│  │       ├── README.md
│  │       └── demo.sh
│  │
│  ├── 📁 EXAMPLE 3: RESEARCH ASSISTANT SYSTEM
│  │   │
│  │   └── research-assistant/
│  │       │
│  │       ├── cmd/
│  │       │   ├── main.go
│  │       │   └── server.go
│  │       │
│  │       ├── internal/
│  │       │   ├── crew.go                   # CreateResearchCrew()
│  │       │   ├── tools.go                  # Search, paper analysis, synthesis
│  │       │   │   ├── web_search_tool.go
│  │       │   │   ├── paper_analyzer_tool.go
│  │       │   │   ├── citation_tool.go
│  │       │   │   ├── synthesis_tool.go
│  │       │   │   └── export_tool.go
│  │       │   │
│  │       │   ├── config.go
│  │       │   ├── handlers.go
│  │       │   └── report_formatter.go
│  │       │
│  │       ├── config/
│  │       │   ├── crew.yaml
│  │       │   └── agents/
│  │       │       ├── researcher.yaml        # Find sources
│  │       │       ├── analyst.yaml           # Analyze findings
│  │       │       ├── synthesizer.yaml       # Synthesize report
│  │       │       └── editor.yaml            # Polish output
│  │       │
│  │       ├── web/
│  │       │   ├── index.html
│  │       │   └── report_template.html
│  │       │
│  │       ├── tests/
│  │       ├── .env.example
│  │       ├── Makefile
│  │       ├── README.md
│  │       └── demo.sh
│  │
│  ├── 📁 EXAMPLE 4: DATA ANALYSIS SYSTEM
│  │   │
│  │   └── data-analysis/
│  │       │
│  │       ├── cmd/
│  │       │   ├── main.go
│  │       │   └── server.go
│  │       │
│  │       ├── internal/
│  │       │   ├── crew.go                   # CreateDataAnalysisCrew()
│  │       │   ├── tools.go                  # Load, process, visualize
│  │       │   │   ├── data_loader_tool.go
│  │       │   │   ├── cleaner_tool.go
│  │       │   │   ├── analyzer_tool.go
│  │       │   │   ├── visualizer_tool.go
│  │       │   │   └── exporter_tool.go
│  │       │   │
│  │       │   ├── config.go
│  │       │   ├── handlers.go
│  │       │   └── chart_generator.go
│  │       │
│  │       ├── config/
│  │       │   ├── crew.yaml
│  │       │   └── agents/
│  │       │       ├── loader.yaml            # Load data
│  │       │       ├── analyzer.yaml          # Analyze patterns
│  │       │       ├── visualizer.yaml        # Create charts
│  │       │       └── reporter.yaml          # Generate report
│  │       │
│  │       ├── web/
│  │       │   ├── index.html
│  │       │   └── chart_template.html
│  │       │
│  │       ├── data/                          # Sample data
│  │       │   ├── sample.csv
│  │       │   ├── sample.json
│  │       │   └── README.md
│  │       │
│  │       ├── tests/
│  │       ├── .env.example
│  │       ├── Makefile
│  │       ├── README.md
│  │       └── demo.sh
│  │
│  ├── 🔧 SHARED UTILITIES (Optional)
│  │   └── internal/shared/
│  │       ├── logger.go                     # Logging helpers
│  │       ├── env.go                        # .env loading
│  │       ├── constants.go                  # Shared constants
│  │       ├── validators.go                 # Input validation
│  │       └── formatters.go                 # Output formatting
│  │
│  └── 🧪 SHARED TEST FIXTURES
│      └── testdata/
│          ├── sample_crew_configs.yaml
│          ├── sample_responses.json
│          └── mock_data/
│              ├── it_support_scenarios.yaml
│              ├── customer_service_conversations.yaml
│              ├── research_queries.yaml
│              └── datasets.csv


├─ 📚 ROOT LEVEL DOCUMENTATION
│  │
│  ├── README.md                             # Main project overview
│  │   ├─ What is go-agentic?
│  │   ├─ Architecture overview
│  │   ├─ Quick start (link to examples)
│  │   ├─ Directory structure
│  │   └─ Contributing guidelines
│  │
│  ├── CONTRIBUTING.md                       # How to contribute
│  │   ├─ Development setup
│  │   ├─ Code style
│  │   ├─ Testing requirements
│  │   ├─ Commit conventions
│  │   └─ PR process
│  │
│  ├── LICENSE                               # MIT License
│  │
│  ├── 🗂️ ARCHITECTURE DOCS
│  │   ├── ARCHITECTURE_SPLIT.md             # ← THIS FILE (phase 1)
│  │   ├── DIRECTORY_STRUCTURE_DETAILED.md   # ← THIS FILE (phase 2)
│  │   │
│  │   ├── LIBRARY_VS_EXAMPLES.md             # Differences explained
│  │   └── MIGRATION_FROM_MONOLITH.md        # How to migrate
│  │
│  ├── 🚀 GETTING STARTED
│  │   ├── INSTALLATION.md                   # How to install
│  │   ├── QUICKSTART.md                     # 5-minute setup
│  │   └── EXAMPLES_GUIDE.md                 # Running examples
│  │
│  ├── 📖 GUIDES
│  │   ├── BUILDING_CUSTOM_AGENTS.md         # How to build custom crew
│  │   ├── WRITING_CUSTOM_TOOLS.md           # How to write tools
│  │   ├── CONFIGURATION_GUIDE.md            # YAML configuration
│  │   └── DEPLOYMENT.md                     # Production deployment
│  │
│  └── 📊 PROJECT MANAGEMENT
│      ├── ROADMAP.md                        # Future plans
│      ├── CHANGELOG.md                      # Release notes
│      └── DEVELOPMENT.md                    # Development workflow


└─ 🔧 GIT CONFIGURATION
    ├── .gitignore                           # Ignore rules
    ├── .github/
    │   ├── workflows/
    │   │   ├── test.yml                    # Run all tests on PR
    │   │   ├── build.yml                   # Build verification
    │   │   ├── lint.yml                    # Code quality checks
    │   │   └── release.yml                 # Automated releases
    │   │
    │   └── ISSUE_TEMPLATE/
    │       ├── bug_report.md
    │       └── feature_request.md
    │
    ├── .gitmodules                          # Submodules (if any)
    └── CODEOWNERS                           # Code ownership
```

---

## 🎯 FILE COUNT & SIZE ANALYSIS

### go-crewai/ (Library)
```
Core Files:
  - types.go              84 lines      ~3 KB
  - agent.go             234 lines      ~10 KB
  - crew.go              398 lines      ~15 KB
  - config.go            169 lines      ~7 KB
  - http.go              187 lines      ~8 KB
  - streaming.go          54 lines      ~2 KB
  - html_client.go       252 lines      ~11 KB
  - report.go            696 lines      ~25 KB
  - tests.go             316 lines      ~12 KB
  ─────────────────────────────────────
  TOTAL              2,384 lines      ~93 KB

Documentation:     ~15 files          ~100 KB
Templates:         ~10 files          ~30 KB
Tests:             ~5 files           ~20 KB
─────────────────────────────────────
Total Package:     ~30 files          ~243 KB
```

### go-agentic-examples/ (Examples)
```
IT Support:        ~800 lines          ~30 KB
Customer Service:  ~700 lines          ~25 KB
Research:          ~750 lines          ~28 KB
Data Analysis:     ~800 lines          ~30 KB
─────────────────────────────────────
App Code:        3,050 lines         ~113 KB

Tests:           ~500 lines          ~20 KB
Config:          ~200 lines          ~8 KB
Web/UI:          ~300 lines          ~15 KB
─────────────────────────────────────
Total Examples:  ~70 files           ~350 KB
```

---

## 🔄 IMPORT STRUCTURE (After Split)

### go-crewai Package
```go
// No dependencies on examples
import (
    "github.com/openai/openai-go/v3"
    "gopkg.in/yaml.v3"
)
```

### go-agentic-examples/it-support
```go
// IT Support imports library
import (
    "github.com/taipm/go-crewai"  // ← Lõi library
    "github.com/openai/openai-go/v3"
)
```

### go-agentic-examples/customer-service
```go
// Customer Service imports library
import (
    "github.com/taipm/go-crewai"  // ← Lõi library
    "github.com/openai/openai-go/v3"
)
```

Same pattern for research-assistant/ and data-analysis/

---

## ✅ STRUCTURAL CHECKLIST

- [ ] go-crewai/ has all 9 core .go files
- [ ] go-crewai/ has comprehensive docs/ folder
- [ ] go-crewai/ has examples/ with templates
- [ ] go-crewai/go.mod points to github.com/taipm/go-crewai
- [ ] go-agentic-examples/ has 4 subdirectories (one per example)
- [ ] Each example has cmd/, internal/, config/, tests/
- [ ] Each example has .env.example and README.md
- [ ] Each example imports go-crewai as dependency
- [ ] go-agentic-examples/go.mod points to github.com/taipm/go-agentic-examples
- [ ] go-agentic-examples/go.mod has `replace` directive for local dev
- [ ] Root README.md explains both parts
- [ ] Root has ARCHITECTURE_SPLIT.md (this file)

---

## 🎨 Visual Summary: Package Organization

```
┌────────────────────────────────────────────────────────────┐
│                    go-agentic Root                         │
│                                                            │
│  ├─ go-crewai/          (Library - 2,384 LOC)            │
│  │  ├─ types.go                                          │
│  │  ├─ agent.go                                          │
│  │  ├─ crew.go                                           │
│  │  ├─ config.go                                         │
│  │  ├─ http.go                                           │
│  │  ├─ streaming.go                                      │
│  │  ├─ html_client.go                                    │
│  │  ├─ report.go                                         │
│  │  ├─ tests.go                                          │
│  │  ├─ docs/                                             │
│  │  ├─ examples/templates                                │
│  │  └─ go.mod                                            │
│  │                                                        │
│  └─ go-agentic-examples/     (Examples - 3,050 LOC)      │
│     ├─ it-support/                                       │
│     │  ├─ cmd/main.go                                    │
│     │  ├─ internal/crew.go, tools.go                     │
│     │  ├─ config/crew.yaml + agents/                     │
│     │  └─ tests/                                         │
│     │                                                    │
│     ├─ customer-service/         (Same structure)        │
│     ├─ research-assistant/       (Same structure)        │
│     ├─ data-analysis/            (Same structure)        │
│     │                                                    │
│     └─ go.mod (depends on go-crewai)                    │
│                                                           │
│  ├─ README.md             (Root overview)               │
│  ├─ ARCHITECTURE_SPLIT.md (← This strategic document)   │
│  └─ CONTRIBUTING.md                                      │
└────────────────────────────────────────────────────────┘
```

---

## 🚀 NEXT STEPS

1. **Review** this document with team
2. **Create** go-crewai/ and go-agentic-examples/ directories
3. **Move** files according to structure above
4. **Update** go.mod files with correct module names
5. **Test** each package independently
6. **Document** migration path for existing users
7. **Release** go-crewai v1.0.0 as library
8. **Release** go-agentic-examples v1.0.0

