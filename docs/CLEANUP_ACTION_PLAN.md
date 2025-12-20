# 🔧 CLEANUP & SPLIT EXECUTION PLAN

## ⚡ QUICK ANSWER: YES, CÓ VẤNĐỀ

**Core library hiện tại = 85% tối ưu**

Vấn đề:
```
🚨 example_it_support.go (539 lines) KHÔNG NÊN ở trong core
🚨 cmd/main.go, cmd/test.go là IT-specific (KHÔNG NÊN ở core)

Khi di chuyển những files này:
✅ Core = 2,384 lines (100% pure)
✅ Examples = Separate package
✅ Clean separation
```

---

## 🎯 MỤC TIÊU CUỐI CÙNG

```
┌─────────────────────────────────────────┐
│ After Cleanup & Split:                  │
│                                         │
│ go-crewai/                              │
│ ├── types.go          [84]   ✓ Core   │
│ ├── agent.go         [234]   ✓ Core   │
│ ├── crew.go          [398]   ✓ Core   │
│ ├── config.go        [169]   ✓ Core   │
│ ├── http.go          [187]   ✓ Core   │
│ ├── streaming.go      [54]   ✓ Core   │
│ ├── html_client.go   [252]   ✓ Core   │
│ ├── report.go        [696]   ✓ Core   │
│ └── tests.go         [316]   ✓ Core   │
│    ─────────────────────────────────   │
│    TOTAL: 2,384 lines (100% pure)     │
│                                         │
│ go-agentic-examples/                    │
│ ├── it-support/      [539 LOC from ... │
│ ├── customer-service/                   │
│ ├── research-assistant/                 │
│ └── data-analysis/                      │
│    All import go-crewai library         │
└─────────────────────────────────────────┘
```

---

## 📋 STEP-BY-STEP CLEANUP

### PHASE 1: Move IT Support Example Out

#### Step 1.1: Create go-agentic-examples/ structure
```bash
# Create directory structure
mkdir -p go-agentic-examples/it-support/{cmd,internal,config/agents,tests,web}
mkdir -p go-agentic-examples/{customer-service,research-assistant,data-analysis}

# Create go.mod
cat > go-agentic-examples/go.mod << 'EOF'
module github.com/taipm/go-agentic-examples

go 1.25.2

require (
    github.com/taipm/go-crewai v1.0.0
    github.com/openai/openai-go/v3 v3.14.0
    gopkg.in/yaml.v3 v3.0.1
)

replace github.com/taipm/go-crewai => ../go-crewai
EOF
```

#### Step 1.2: Extract & Move IT Support Code
```bash
# CURRENT: example_it_support.go contains:
# 1. CreateITSupportCrew() function
# 2. createITSupportTools() + 8 IT tools
# 3. Tool implementations (GetCPUUsage, GetMemoryUsage, etc.)

# ACTION: Split into 2 files

# File 1: go-agentic-examples/it-support/internal/crew.go
# Contains: CreateITSupportCrew() + crew configuration
cp go-crewai/example_it_support.go go-agentic-examples/it-support/internal/crew.go
# Then EDIT: Remove tool implementations, keep only CreateITSupportCrew()

# File 2: go-agentic-examples/it-support/internal/tools.go
# Contains: createITSupportTools() + all IT tool implementations
# Create: go-agentic-examples/it-support/internal/tools.go
# Then EDIT: Move tool implementations from crew.go

# Change package: crewai → main (or it-support package)
```

#### Step 1.3: Move cmd files
```bash
# Move: go-crewai/cmd/main.go → go-agentic-examples/it-support/cmd/main.go
mv go-crewai/cmd/main.go go-agentic-examples/it-support/cmd/main.go

# EDIT: Update imports in main.go
# From: package main
#       import "github.com/taipm/go-crewai" (local)
# To:   package main
#       import "github.com/taipm/go-crewai" (remote)
#       import "./internal" (local it-support package)

# Move config: go-crewai/config/*.yaml → go-agentic-examples/it-support/config/
mv go-crewai/config/* go-agentic-examples/it-support/config/
```

#### Step 1.4: Delete from core library
```bash
# Remove files that should not be in core
rm go-crewai/example_it_support.go      # ✓ Moved to examples
rm go-crewai/cmd/main.go                # ✓ Moved to examples
rm go-crewai/cmd/test.go                # ✓ Need to check what this is
rm -rf go-crewai/config                 # ✓ Moved to examples
rm -rf go-crewai/cmd                    # ✓ Directory now empty
```

#### Step 1.5: Verify core library structure
```bash
cd go-crewai
ls -la
# Expected:
# ├── types.go
# ├── agent.go
# ├── crew.go
# ├── config.go              ← Struct definitions only, no loading example
# ├── http.go
# ├── streaming.go
# ├── html_client.go
# ├── report.go
# ├── tests.go
# ├── go.mod
# ├── go.sum
# ├── docs/
# ├── examples/              ← Templates only
# └── tests/                 ← Test files

# Files that should NOT be here:
# ✗ example_it_support.go
# ✗ cmd/
# ✗ config/
```

---

### PHASE 2: Create IT Support Example

#### Step 2.1: Create cmd/main.go for IT support
```bash
# File: go-agentic-examples/it-support/cmd/main.go

cat > go-agentic-examples/it-support/cmd/main.go << 'EOF'
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    
    "github.com/taipm/go-crewai"
    "../internal"  // IT Support specific
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY not set")
    }
    
    // Create IT Support crew
    crew := internal.CreateITSupportCrew()
    
    // Create executor
    executor := crewai.NewCrewExecutor(crew, apiKey)
    
    // Execute
    ctx := context.Background()
    response, err := executor.Execute(ctx, "Check system health")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response.Content)
}
EOF
```

#### Step 2.2: Create internal/crew.go for IT support
```bash
# File: go-agentic-examples/it-support/internal/crew.go

# This file should contain:
# - package declaration
# - CreateITSupportCrew() function
# - Agent definitions specific to IT support
# - References to tools (but NOT tool implementations)

# The file should import:
# - "github.com/taipm/go-crewai" (the core library)
```

#### Step 2.3: Create internal/tools.go for IT support
```bash
# File: go-agentic-examples/it-support/internal/tools.go

# This file should contain:
# - createITSupportTools() function
# - All 8 IT-specific tool implementations:
#   - GetCPUUsage()
#   - GetMemoryUsage()
#   - GetDiskSpace()
#   - GetSystemInfo()
#   - GetRunningProcesses()
#   - PingHost()
#   - CheckServiceStatus()
#   - ResolveDNS()

# The file should import:
# - "github.com/taipm/go-crewai" (for Tool struct)
```

#### Step 2.4: Move config files
```bash
# Copy YAML configs to examples
cp go-crewai/config/crew.yaml go-agentic-examples/it-support/config/
cp go-crewai/config/agents/* go-agentic-examples/it-support/config/agents/
```

#### Step 2.5: Create go.mod for IT support
```bash
# File: go-agentic-examples/it-support/go.mod

cat > go-agentic-examples/it-support/go.mod << 'EOF'
module github.com/taipm/go-agentic-examples/it-support

go 1.25.2

require (
    github.com/taipm/go-crewai v1.0.0
    github.com/openai/openai-go/v3 v3.14.0
    gopkg.in/yaml.v3 v3.0.1
)

replace github.com/taipm/go-crewai => ../../go-crewai
EOF
```

#### Step 2.6: Create README.md for IT support
```bash
# File: go-agentic-examples/it-support/README.md
# Content: How to run IT support example, what it does, etc.
```

---

### PHASE 3: Update Core Library

#### Step 3.1: Update go-crewai/go.mod
```bash
# File: go-crewai/go.mod

cat > go-crewai/go.mod << 'EOF'
module github.com/taipm/go-crewai

go 1.25.2

require (
    github.com/openai/openai-go/v3 v3.14.0
    gopkg.in/yaml.v3 v3.0.1
)
EOF

# Run: go mod tidy
cd go-crewai
go mod tidy
```

#### Step 3.2: Verify core library compiles
```bash
cd go-crewai
go build ./...
go test ./...
# Should pass with 0 errors
```

#### Step 3.3: Create go-crewai docs
```bash
# Create: go-crewai/docs/README.md
# Create: go-crewai/docs/API.md
# Create: go-crewai/docs/CONFIG.md
# etc.
```

#### Step 3.4: Create go-crewai/examples/ with templates
```bash
mkdir -p go-crewai/examples/{agents,tools}

# File: go-crewai/examples/README.md
# File: go-crewai/examples/minimal.go
# File: go-crewai/examples/crew.yaml.template
# File: go-crewai/examples/agents/agent1.yaml.template
```

---

### PHASE 4: Verify & Test

#### Step 4.1: Verify core library
```bash
cd go-crewai

# Check structure
ls -la
# Should have: types.go, agent.go, crew.go, config.go, http.go, streaming.go, html_client.go, report.go, tests.go, go.mod, go.sum, docs/, examples/, tests/

# Check compilation
go build ./...

# Check tests
go test ./...

# Check line count
wc -l *.go
# types.go should be ~84 lines
# agent.go should be ~234 lines
# Total ~2,384 lines (not including examples/)
```

#### Step 4.2: Verify IT support example
```bash
cd go-agentic-examples/it-support

# Check structure
ls -la
# Should have: cmd/, internal/, config/, tests/, go.mod, README.md, .env.example

# Check compilation
go build ./cmd/main.go

# Check it runs (with OPENAI_API_KEY set)
OPENAI_API_KEY=test go run ./cmd/main.go

# Check imports
grep -r "import" internal/
# Should import "github.com/taipm/go-crewai"
```

#### Step 4.3: Check no circular imports
```bash
# Verify: go-crewai does NOT import from go-agentic-examples
grep -r "go-agentic-examples" go-crewai/
# Should return NOTHING

# Verify: go-agentic-examples/it-support DOES import from go-crewai
grep -r "github.com/taipm/go-crewai" go-agentic-examples/
# Should return matches
```

---

## 📊 FILE MOVEMENT SUMMARY

### Files to MOVE OUT of core

```
go-crewai/example_it_support.go
├── extract CreateITSupportCrew() function
│   → go-agentic-examples/it-support/internal/crew.go
│
└── extract createITSupportTools() + tool implementations
    → go-agentic-examples/it-support/internal/tools.go

go-crewai/cmd/main.go
└── → go-agentic-examples/it-support/cmd/main.go (IT-specific)

go-crewai/config/
└── → go-agentic-examples/it-support/config/ (IT-specific config)
```

### Files to KEEP in core

```
go-crewai/types.go              ✓ Pure types
go-crewai/agent.go              ✓ Generic execution
go-crewai/crew.go               ✓ Generic orchestration
go-crewai/config.go             ✓ Generic YAML loader
go-crewai/http.go               ✓ Generic HTTP API
go-crewai/streaming.go          ✓ Generic SSE events
go-crewai/html_client.go        ✓ Generic web UI
go-crewai/report.go             ✓ Generic reporting
go-crewai/tests.go              ✓ Generic test utilities
go-crewai/docs/                 ✓ Library documentation
go-crewai/examples/             ✓ Template examples
go-crewai/tests/                ✓ Library test files
```

---

## ✅ CHECKLIST: CLEANUP

### Before Starting
- [ ] Backup current code (git commit -m "backup before cleanup")
- [ ] Read CORE_LIBRARY_ANALYSIS.md
- [ ] Understand what's core vs example

### Remove from Core
- [ ] Delete go-crewai/example_it_support.go
- [ ] Delete go-crewai/cmd/main.go
- [ ] Delete go-crewai/cmd/test.go
- [ ] Delete go-crewai/config/ directory
- [ ] Delete go-crewai/cmd/ directory (if empty)

### Create Examples Package
- [ ] Create go-agentic-examples/ directory structure
- [ ] Create go-agentic-examples/go.mod
- [ ] Create go-agentic-examples/it-support/{cmd,internal,config,tests}

### Move IT Support Code
- [ ] Split example_it_support.go into crew.go + tools.go
- [ ] Move to go-agentic-examples/it-support/internal/
- [ ] Update package declarations
- [ ] Update imports (use "github.com/taipm/go-crewai")

### Move Configs
- [ ] Move config/crew.yaml → it-support/config/
- [ ] Move config/agents/*.yaml → it-support/config/agents/
- [ ] Update paths in code if needed

### Create cmd/main.go
- [ ] Create go-agentic-examples/it-support/cmd/main.go
- [ ] Make it import go-crewai library
- [ ] Test it compiles and runs

### Update Core Library
- [ ] Update go-crewai/go.mod (remove example dependencies)
- [ ] Run: cd go-crewai && go mod tidy
- [ ] Test: go build ./...
- [ ] Test: go test ./...
- [ ] Verify: 0 import errors

### Verify Structure
- [ ] go-crewai has 9 core files only
- [ ] go-crewai has no IT-specific code
- [ ] go-agentic-examples/it-support is complete
- [ ] No circular imports
- [ ] All imports use github.com/taipm/go-crewai

### Test Everything
- [ ] go-crewai compiles: ✓ go build ./...
- [ ] go-crewai tests pass: ✓ go test ./...
- [ ] IT support compiles: ✓ go build ./cmd/main.go
- [ ] IT support runs: ✓ OPENAI_API_KEY=test go run ./cmd/main.go
- [ ] No confusing files in core: ✓ ls -la go-crewai/

### Documentation
- [ ] Create go-crewai/docs/ with library docs
- [ ] Create go-agentic-examples/README.md
- [ ] Create go-agentic-examples/it-support/README.md
- [ ] Create MIGRATION.md for users
- [ ] Update root README.md

### Git Commit
- [ ] Commit with message: "Split: Move IT example to go-agentic-examples"
- [ ] Commit with message: "Cleanup: Remove example code from core library"

---

## 📈 BEFORE & AFTER

### BEFORE Cleanup
```
go-crewai/ = Confusing Mix (2,993 lines)
├── Core library code       (2,384 lines) 79%
├── IT Support example      (539 lines)   18%
└── IT entry points         (70 lines)    3%

Problem: Users don't know what's core!
```

### AFTER Cleanup
```
go-crewai/ = Pure Library (2,384 lines)
├── Core library code       (2,384 lines) 100%
└── No example code

go-agentic-examples/ = Clear Examples
├── it-support/             (539 lines)
├── customer-service/
├── research-assistant/
└── data-analysis/

Result: Crystal clear separation!
```

---

## 🎯 TIME ESTIMATE

- **Cleanup core library**: 30 minutes
  - Delete 3 files
  - Update go.mod
  - Test compilation
  
- **Create example structure**: 20 minutes
  - Create directories
  - Create go.mod
  - Create cmd/main.go

- **Move & refactor code**: 1 hour
  - Extract crew functions
  - Extract tool implementations
  - Update imports
  - Move configs

- **Testing & verification**: 30 minutes
  - Verify compilation
  - Verify no imports from examples
  - Check file structure

- **Documentation**: 30 minutes
  - Create README files
  - Update imports
  - Create migration guide

**Total: ~3 hours** for complete cleanup

---

## 🚀 AFTER CLEANUP: VERIFICATION

You'll be able to verify success by checking:

```bash
# 1. Core library builds clean
cd go-crewai
go build ./...  # ✅ Should succeed

# 2. Core has no example code
grep -r "GetCPUUsage\|IT Support\|example_it" go-crewai/
# ✅ Should return nothing

# 3. Core is ~2,384 lines
wc -l go-crewai/*.go | tail -1
# ✅ Should be around 2,384

# 4. Example has IT code
grep -r "GetCPUUsage" go-agentic-examples/it-support/
# ✅ Should find the function

# 5. Example imports core library
grep -r "github.com/taipm/go-crewai" go-agentic-examples/
# ✅ Should find imports

# 6. No circular imports
grep -r "go-agentic-examples" go-crewai/
# ✅ Should return nothing
```

---

## FINAL RESULT

**After this cleanup:**

✅ **go-crewai/** is a PERFECT core library
   - 2,384 lines, pure, minimal, comprehensive
   - No example code
   - 100% reusable
   - Production-ready

✅ **go-agentic-examples/** contains all examples
   - Each example is independent
   - All use go-crewai library
   - Easy to understand
   - Easy to extend

✅ **Clear separation of concerns**
   - Users know what to import
   - Users know what to copy
   - Users know how to extend

✅ **Ready for distribution**
   - go-crewai v1.0.0 → library
   - go-agentic-examples v1.0.0 → examples
   - Independent versioning

