# ⚡ QUICK REFERENCE: STEP-BY-STEP SPLIT GUIDE

## 📋 All Commands in One Place

### PHASE 1: Backup (15 min)
```bash
cd /Users/taipm/GitHub/go-agentic

# Backup branch
git checkout -b backup/before-split
git push origin backup/before-split
git checkout feature/epic-4-cross-platform

# Create examples directories
mkdir -p go-agentic-examples/it-support/{cmd,internal,config/agents,tests,web}
mkdir -p go-agentic-examples/{customer-service,research-assistant,data-analysis}
```

---

### PHASE 2: Remove from Core (30 min)
```bash
cd /Users/taipm/GitHub/go-agentic

# Backup before deletion
mkdir -p /tmp/go-agentic-backup
cp go-crewai/example_it_support.go /tmp/go-agentic-backup/
cp go-crewai/cmd/main.go /tmp/go-agentic-backup/
cp go-crewai/cmd/test.go /tmp/go-agentic-backup/
cp -r go-crewai/config /tmp/go-agentic-backup/

# Delete from core
rm go-crewai/example_it_support.go
rm -rf go-crewai/cmd
rm -rf go-crewai/config
```

---

### PHASE 3: Create Examples Root (45 min)
```bash
cd /Users/taipm/GitHub/go-agentic

# Root go.mod
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

# Root README
cat > go-agentic-examples/README.md << 'EOF'
# 🚀 Go-Agentic Examples

Complete examples for go-crewai library.

## Examples

- **it-support/**: IT support system
- **customer-service/**: Customer service
- **research-assistant/**: Research assistant
- **data-analysis/**: Data analysis
EOF

# Create .gitignore
cat > go-agentic-examples/.gitignore << 'EOF'
*.exe
*.dylib
*.out
.env
.env.local
.vscode/
.idea/
*.swp
*.prof
coverage.out
EOF

touch go-agentic-examples/go.sum
```

---

### PHASE 4: Create IT Support (1 hour)

#### 4.1 Create IT Support go.mod
```bash
cat > go-agentic-examples/it-support/go.mod << 'EOF'
module github.com/taipm/go-agentic-examples/it-support

go 1.25.2

require (
    github.com/taipm/go-crewai v1.0.0
    github.com/openai/openai-go/v3 v3.14.0
    gopkg.in/yaml.v3 v3.0.1
)

replace github.com/taipm/go-crewai => ../../../go-crewai
EOF
```

#### 4.2 Copy IT Code (crew.go + tools.go)
See STEP_BY_STEP_EXECUTION.md Phase 4.2-4.3 for full content

#### 4.3 Create cmd/main.go
See STEP_BY_STEP_EXECUTION.md Phase 4.4 for full content

#### 4.4 Copy Configs
```bash
cp /tmp/go-agentic-backup/config/crew.yaml go-agentic-examples/it-support/config/
cp /tmp/go-agentic-backup/config/agents/*.yaml go-agentic-examples/it-support/config/agents/
```

#### 4.5 Create .env.example
```bash
cat > go-agentic-examples/it-support/.env.example << 'EOF'
OPENAI_API_KEY=sk-your-key-here
EOF
```

---

### PHASE 5: Update go.mod (30 min)
```bash
cd /Users/taipm/GitHub/go-agentic/go-crewai
cat > go.mod << 'EOF'
module github.com/taipm/go-crewai

go 1.25.2

require (
    github.com/openai/openai-go/v3 v3.14.0
    gopkg.in/yaml.v3 v3.0.1
)

require (
    github.com/tidwall/gjson v1.18.0 // indirect
    github.com/tidwall/match v1.1.1 // indirect
    github.com/tidwall/pretty v1.2.1 // indirect
    github.com/tidwall/sjson v1.2.5 // indirect
)
EOF

go mod tidy

# Examples
cd /Users/taipm/GitHub/go-agentic/go-agentic-examples
go mod download
go mod tidy

cd it-support
go mod tidy
```

---

### PHASE 6: Test & Verify (45 min)
```bash
# Test core
cd /Users/taipm/GitHub/go-agentic/go-crewai
go build ./...

# Test examples
cd /Users/taipm/GitHub/go-agentic/go-agentic-examples/it-support
go build ./cmd/main.go

# Check line counts
wc -l /Users/taipm/GitHub/go-agentic/go-crewai/*.go | tail -1

# Verify files deleted
test -f /Users/taipm/GitHub/go-agentic/go-crewai/example_it_support.go || echo "✓ Deleted"
test -d /Users/taipm/GitHub/go-agentic/go-crewai/cmd || echo "✓ Deleted"
test -d /Users/taipm/GitHub/go-agentic/go-crewai/config || echo "✓ Deleted"
```

---

### PHASE 7: Documentation (1 hour)

See STEP_BY_STEP_EXECUTION.md Phase 7 for:
- Updated root README.md
- CONTRIBUTING.md
- SPLIT_COMPLETE.md

---

### PHASE 8: Commit (15 min)
```bash
cd /Users/taipm/GitHub/go-agentic

git add -A

git commit -m "feat: split go-agentic into core library and examples

Split go-crewai into a pure, reusable core library and moved
example code to separate go-agentic-examples package.

Core Library (2,384 lines):
- types.go, agent.go, crew.go, config.go
- http.go, streaming.go, html_client.go
- report.go, tests.go
- 100% pure, no domain-specific code

Examples Package:
- IT Support system (complete, working)
- Structure for Customer Service, Research, Data Analysis

Removed from core:
- example_it_support.go (539 lines)
- cmd/ directory
- config/ directory (IT-specific)"

# Verify
git log -1 --stat
```

---

## ✅ Verification Checklist

```bash
# 1. Core library structure
ls -lah /Users/taipm/GitHub/go-agentic/go-crewai/*.go | wc -l
# Expected: 9 files

# 2. Core library size
wc -l /Users/taipm/GitHub/go-agentic/go-crewai/*.go | tail -1
# Expected: ~2,384 lines

# 3. Unwanted files deleted
test -f /Users/taipm/GitHub/go-agentic/go-crewai/example_it_support.go && echo "❌ Still exists" || echo "✅ Deleted"
test -d /Users/taipm/GitHub/go-agentic/go-crewai/cmd && echo "❌ Still exists" || echo "✅ Deleted"

# 4. Core compiles
cd /Users/taipm/GitHub/go-agentic/go-crewai && go build ./... && echo "✅ Success" || echo "❌ Failed"

# 5. Examples compiles
cd /Users/taipm/GitHub/go-agentic/go-agentic-examples/it-support && go build ./cmd/main.go && echo "✅ Success" || echo "❌ Failed"

# 6. No circular imports
grep -r "go-agentic-examples" /Users/taipm/GitHub/go-agentic/go-crewai/ || echo "✅ No imports found"

# 7. Git status clean
cd /Users/taipm/GitHub/go-agentic && git status
# Expected: nothing to commit, working tree clean
```

---

## ⏱️ Time Breakdown

| Phase | Task | Time |
|-------|------|------|
| 1 | Backup & Prepare | 15 min |
| 2 | Remove IT Code | 30 min |
| 3 | Create Examples | 45 min |
| 4 | Move IT Support | 60 min |
| 5 | Update go.mod | 30 min |
| 6 | Test & Verify | 45 min |
| 7 | Documentation | 60 min |
| 8 | Final Commit | 15 min |
| **TOTAL** | | **~5 hours** |

---

## 🚀 After Completion

```bash
# Review changes
git log -1 --stat
git show HEAD

# Test if needed
export OPENAI_API_KEY=your_key
cd go-agentic-examples/it-support
go run ./cmd/main.go -q "Check system"

# Push when ready
git push origin feature/epic-4-cross-platform

# Create tags (optional)
git tag v1.0.0-core
git tag v1.0.0-examples
git push origin --tags
```

---

## 📞 Troubleshooting

### Build fails for core library
```bash
cd go-crewai
go mod tidy
go clean
go build ./...
```

### Build fails for examples
```bash
cd go-agentic-examples/it-support
go mod download
go mod tidy
go build ./cmd/main.go
```

### Import errors
- Verify go.mod replace directive points to correct path
- Check relative paths in go.mod
- Run `go mod tidy`

---

## 📚 Full Documentation

For detailed explanations, see:
- **STEP_BY_STEP_EXECUTION.md**: Complete step-by-step guide with explanations
- **00_START_HERE.md**: Quick overview
- **CORE_ASSESSMENT_EXECUTIVE.md**: Why this split is needed

---

**Total Time: ~5 hours | Result: Perfect, distributable project ✅**

