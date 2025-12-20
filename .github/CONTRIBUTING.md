# 👥 Contributing Guide

Thank you for contributing to go-agentic! This guide will help you get started.

---

## 🚀 Getting Started

### Prerequisites

```bash
# Go 1.20 or later
go version

# Git installed
git --version
```

### Setup Development Environment

```bash
# 1. Clone the repository
git clone https://github.com/taipm/go-agentic.git
cd go-agentic

# 2. Create your feature branch
git checkout -b feature/your-feature-name

# 3. Copy environment template
cp .env.example .env

# 4. Add your API key to .env
echo "OPENAI_API_KEY=sk-your-key-here" >> .env

# 5. Download dependencies
go mod download
cd core && go mod tidy
cd ../examples/it-support && go mod tidy
cd ../examples/go-crewai && go mod tidy
```

---

## 📋 Development Workflow

### 1. Before You Start

```bash
# Ensure your feature branch is up to date
git fetch origin
git rebase origin/main

# Create a new branch
git checkout -b feature/short-description
```

### 2. Writing Code

**Code Style:**
- Follow standard Go conventions
- Run `gofmt -s -w .` before committing
- Run `go vet ./...` to check for issues
- Keep functions small and focused

**Code Organization:**
```
core/                          # Core library
├── types.go                   # Core types
├── crew.go                    # Main orchestration
├── config.go                  # Configuration loading
├── agent.go                   # Agent execution
└── http.go                    # HTTP server

examples/
├── go-crewai/                # Full example with config loading
├── it-support/               # IT support example
└── research-assistant/       # Research example
```

**Comments & Documentation:**
- Add comments for public functions
- Explain complex logic
- Update README if adding features

### 3. Configuration Files

**Important:** Never commit `.env` files!

```bash
# ✅ DO: Use .env.example
cat > .env.example << 'EOF'
OPENAI_API_KEY=sk-placeholder
EOF

# ❌ DON'T: Commit .env with real keys
git add .env  # This will be rejected by CI!
```

### 4. Testing

```bash
# Run unit tests
cd core && go test -v ./...

# Run with coverage
go test -v -cover ./...

# Run specific test
go test -v -run TestAgentExecution ./...

# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 5. Before Committing

```bash
# 1. Format code
gofmt -s -w .

# 2. Run linters
go vet ./...
go fmt ./...

# 3. Run tests
go test ./...

# 4. Check for secrets
git secrets --scan

# 5. Verify .env not staged
git status | grep ".env"  # Should show nothing!

# 6. Verify only intended files staged
git diff --cached
```

---

## 📝 Commit Guidelines

### Commit Message Format

```
<type>: <subject>

<body>

<footer>
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

### Examples

```bash
# Feature
git commit -m "feat: Add signal-based routing to core library"

# Bug fix
git commit -m "fix: Resolve infinite loop in agent orchestration"

# Documentation
git commit -m "docs: Add API documentation for CrewExecutor"

# Config change
git commit -m "chore: Update crew.yaml routing configuration"
```

---

## 🔄 Pull Request Process

### 1. Create Your PR

```bash
# Push your branch
git push origin feature/your-feature-name

# Go to GitHub and create a pull request
```

### 2. PR Description Template

```markdown
## Description
Brief description of what this PR does

## Type of Change
- [ ] New feature
- [ ] Bug fix
- [ ] Documentation update
- [ ] Performance improvement

## Testing
- [ ] Unit tests added
- [ ] Integration tests pass
- [ ] Manual testing done

## Checklist
- [ ] Code follows style guidelines
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No API keys or secrets committed
- [ ] All tests pass
- [ ] CI/CD pipeline passes

## Related Issues
Closes #(issue)
```

### 3. CI/CD Checks

Your PR will automatically run:

```
✅ Security checks (no secrets)
✅ Code format checks (gofmt)
✅ Lint checks (golangci-lint)
✅ Build verification
✅ Unit tests
✅ Integration tests
```

**If any check fails:**
1. Review the failure message
2. Fix the issues locally
3. Commit and push again
4. CI will automatically re-run

### 4. Code Review

- At least one maintainer review required
- Address feedback
- Request re-review after changes

---

## 🔐 Security Considerations

### API Keys & Secrets

**CRITICAL:** Never commit secrets!

```bash
# ✅ SAFE: Use environment variables
apiKey := os.Getenv("OPENAI_API_KEY")

# ❌ UNSAFE: Hardcoded keys
const apiKey = "sk-..."
openaiKey := "sk-..."
```

### Detecting Secrets

If you accidentally commit a secret:

```bash
# 1. Stop immediately
# 2. Rotate the exposed key
# 3. Remove the commit
git log --oneline  # Find the commit hash
git revert <hash>  # Safe removal
git push

# 4. Update the secret in GitHub Settings
```

### What the CI checks for:

- OpenAI API keys (sk-*)
- GitHub tokens (ghp_*)
- Database passwords
- .env file presence
- Hardcoded credentials

---

## 📚 Project Structure

```
go-agentic/
├── .github/
│   ├── workflows/           # GitHub Actions
│   ├── SECURITY.md         # Security guide
│   └── CONTRIBUTING.md     # This file
├── core/                    # Core library (THE main code)
│   ├── crew.go             # Orchestration engine
│   ├── agent.go            # Agent implementation
│   ├── config.go           # Config loading
│   └── types.go            # Type definitions
├── examples/
│   ├── go-crewai/          # Full example
│   └── it-support/         # IT support example
├── docs/                    # Documentation
├── README.md               # Project overview
└── .env.example            # Environment template
```

---

## 🎯 Areas to Contribute

### Core Library
- [ ] New agent types
- [ ] Additional routing strategies
- [ ] Performance optimizations
- [ ] New tools framework

### Examples
- [ ] Customer service example
- [ ] Data analysis example
- [ ] Research assistant example

### Documentation
- [ ] API documentation
- [ ] Tutorial guides
- [ ] Architecture diagrams
- [ ] Video walkthroughs

### Tests
- [ ] Unit test coverage
- [ ] Integration tests
- [ ] Performance benchmarks
- [ ] Edge case testing

---

## 🐛 Bug Reports

Found a bug? Please create an issue with:

1. **Title:** Clear, concise description
2. **Description:** What happened vs expected
3. **Steps to reproduce:** How to trigger the bug
4. **Environment:** Go version, OS, etc.
5. **Logs/Screenshots:** Any error messages

Example:
```markdown
## Bug: Infinite loop with vague input

### Description
When user inputs "Tôi không vào được Internet" (network issue),
the system gets stuck in a loop between Orchestrator and Clarifier.

### Steps to Reproduce
1. Start it-support server
2. Send request with input "Tôi không vào được Internet"
3. Observe infinite routing loop

### Environment
- Go 1.21
- OpenAI API (gpt-4o-mini)
- macOS 14.2
```

---

## 💡 Feature Requests

Have an idea? Create an issue with:

1. **Title:** Feature title
2. **Description:** What you want to do
3. **Use case:** Why it's needed
4. **Example:** How it would work

Example:
```markdown
## Feature: Agent memory system

### Description
Agents should retain context across multiple conversations

### Use Case
- Customer support agents remember customer history
- Research agents build on previous findings
- IT support agents learn from past tickets

### Example
```

---

## 📞 Questions?

- Check [README.md](../../README.md) for overview
- Review [SECURITY.md](./SECURITY.md) for secrets
- Look at existing code for examples
- Open a discussion issue for questions

---

## ✨ Thank You!

Your contributions help make go-agentic better! 🙏

**Remember:**
- ✅ No hardcoded secrets
- ✅ Code formatted with gofmt
- ✅ Tests passing
- ✅ PR description clear
- ✅ CI/CD pipeline passing

Happy coding! 🚀
