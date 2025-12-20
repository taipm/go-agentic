# 🔐 Security & CI/CD Guidelines

## Overview

This document explains the security measures and CI/CD setup for the go-agentic project.

---

## 🛡️ Security Principles

### 1. **Never Commit Secrets**

**What NOT to do:**
```bash
# ❌ WRONG - Never commit these files
git add .env
git add .env.local
git commit -m "Add configuration"
```

**What TO do:**
```bash
# ✅ CORRECT - Use .env.example as template
cp .env.example .env
# Edit .env with YOUR actual API keys
# .env is in .gitignore, won't be committed
```

### 2. **API Key Management**

#### Local Development
```bash
# 1. Copy the example file
cp .env.example .env

# 2. Add your real API key
echo "OPENAI_API_KEY=sk-your-actual-key-here" >> .env

# 3. Never commit this file
# (It's already in .gitignore)
```

#### GitHub Repository
```bash
# Secrets are stored in GitHub, not in code
# Repository Settings → Secrets and variables → Actions

# Use in workflows:
# env:
#   OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

### 3. **Environment File Structure**

**File: `.env.example`** (✅ COMMIT THIS)
```env
# API Configuration
OPENAI_API_KEY=sk-placeholder-never-real-keys

# Application Settings
APP_ENV=development
APP_PORT=8080

# Database (if applicable)
# DB_HOST=localhost
# DB_PORT=5432
```

**File: `.env`** (❌ NEVER COMMIT THIS)
```env
OPENAI_API_KEY=sk-your-real-key-here
APP_ENV=development
APP_PORT=8080
```

### 4. **What Gets Checked**

#### Automated Secret Detection
- ✅ TruffleHog - Finds leaked credentials
- ✅ git-secrets - Prevents accidental commits
- ✅ Custom regex patterns - Detects API keys, tokens
- ✅ .env file presence check - Prevents .env commits

#### Hardcoded Values Detection
```bash
# These patterns trigger alerts:
- sk-* (OpenAI API keys)
- ghp_* (GitHub personal tokens)
- password=* (Database credentials)
- OPENAI_API_KEY= (with actual keys)
```

---

## 🔄 CI/CD Workflows

### 1. **Security Checks** (ci.yml)
Runs on: Push to main/develop, PRs

**Steps:**
1. 🔍 **Secret Detection**
   - Uses TruffleHog to scan git history
   - Searches for hardcoded API keys
   - Checks for GitHub tokens

2. 🔐 **Environment File Validation**
   - Ensures `.env` not in repo
   - Verifies `.env.example` exists and is safe
   - Checks `.gitignore` configuration

3. ✅ **Approval:** Must pass before moving to build

### 2. **Build & Test** (ci.yml)
Runs on: Merged PRs to main/develop

**Steps:**
1. 🔨 Build core library
2. 🔨 Build go-crewai example
3. 🔨 Build it-support example
4. 🧪 Run unit tests
5. 📊 Collect coverage metrics

### 3. **Integration Tests** (integration-tests.yml)
Runs on: Push to main/develop, PRs

**Steps:**
1. 🧪 Run mock integration tests (no API calls)
2. 🔍 Verify config files exist
3. ✅ Type safety checks (go vet, staticcheck)
4. 📊 Upload coverage reports

### 4. **Secrets Audit** (secrets-check.yml)
Runs on: Push to main/develop, PRs, Daily at 2 AM UTC

**Steps:**
1. 🔐 Scan for exposed credentials
2. 🔍 Check .env.example safety
3. 🔔 Reminder for API key rotation
4. ✅ Verify GitHub token protection

---

## 📋 Setup Instructions

### Step 1: Create GitHub Secrets
1. Go to **Repository Settings → Secrets and variables → Actions**
2. Click **New repository secret**
3. Add these secrets:
   - `OPENAI_API_KEY` - Your actual API key

```bash
Name: OPENAI_API_KEY
Value: sk-your-actual-key-here
```

### Step 2: Create .env.example
```bash
# In your repository root
cp examples/it-support/.env.example ./.env.example

# Or create from scratch
cat > .env.example << 'EOF'
# API Configuration
OPENAI_API_KEY=sk-placeholder-key

# Application Settings
APP_ENV=development
APP_PORT=8080
EOF

git add .env.example
git commit -m "docs: Add environment template"
```

### Step 3: Verify .gitignore
```bash
# Ensure .env files are ignored
echo ".env" >> .gitignore
echo ".env.local" >> .gitignore
git add .gitignore
git commit -m "chore: Ensure .env files are ignored"
```

### Step 4: Push Workflows
```bash
# The workflow files are already in .github/workflows/
# Just push them
git add .github/
git commit -m "ci: Add GitHub Actions workflows"
git push
```

---

## 🚨 What Happens If a Secret is Detected

### If committed to git:
```
❌ CI Pipeline FAILS
📊 TruffleHog scan detects secret
🔒 Build is blocked
📧 Maintainers are notified
```

### Recovery steps:
```bash
# 1. Rotate the exposed key immediately
# 2. Remove the commit from history
git log --oneline  # Find the commit
git revert <commit-hash>  # Safe way to remove

# 3. Force push (be careful!)
git push --force-with-lease

# 4. Update GitHub secret with new key
# Go to Settings → Secrets and rotate it
```

---

## ✅ Best Practices

### Development

**✅ DO:**
```bash
# Use .env for local development
source .env
export OPENAI_API_KEY

# Test with dummy key in CI
export OPENAI_API_KEY=sk-test-key

# Use environment variables
API_KEY="${OPENAI_API_KEY}"
```

**❌ DON'T:**
```bash
# Don't hardcode in Go
const apiKey = "sk-..."

# Don't put in YAML
openai_key: sk-...

# Don't commit .env
git add .env

# Don't use generic names
.env → .env  ❌ (will be committed)
```

### Code Review

**When reviewing PRs:**

1. ✅ Check for hardcoded secrets
   ```bash
   # Look for patterns like:
   sk- (OpenAI)
   ghp_ (GitHub)
   password=...
   ```

2. ✅ Verify .env files aren't added
   ```bash
   # Check PR diff for .env
   ```

3. ✅ Review .env.example changes
   ```bash
   # Ensure only placeholders are added
   ```

### CI/CD Maintenance

**Monthly:**
- Review GitHub secrets
- Check API key rotation policy
- Update security scanning tools

**Weekly:**
- Monitor CI/CD failures
- Review merged PRs for security issues
- Check dependency vulnerabilities

---

## 🔍 Workflow Details

### ci.yml - Main CI Pipeline

```yaml
Security Checks (MUST PASS)
  ├─ TruffleHog scan
  ├─ API key detection
  ├─ GitHub token detection
  └─ .env file presence check
       ↓
Lint (MUST PASS)
  ├─ gofmt check
  ├─ goimports check
  └─ golangci-lint
       ↓
Build & Test (MUST PASS)
  ├─ Build core library
  ├─ Build go-crewai
  ├─ Build it-support
  ├─ Run unit tests
  └─ Upload coverage
```

### secrets-check.yml - Secrets Audit

```yaml
Runs: On push, PRs, Daily 2 AM UTC

Steps:
  ├─ git-secrets scan
  ├─ Verify .env.example safety
  ├─ Check GitHub tokens
  ├─ Check database credentials
  ├─ Validate environment structure
  └─ API key rotation reminder
```

### integration-tests.yml - Type Safety

```yaml
Steps:
  ├─ Type safety checks
  ├─ Configuration validation
  ├─ staticcheck analysis
  └─ go vet verification
```

---

## 📞 Security Issues

If you find a security vulnerability:

1. **DO NOT** create a public GitHub issue
2. **DO** email with details
3. **DO** include:
   - What vulnerability was found
   - How to reproduce it
   - Suggested fix (if you have one)

---

## 📚 References

- [GitHub Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [OWASP Secure Coding Practices](https://owasp.org/www-community/attacks/Sensitive_Data_Exposure)
- [git-secrets Tool](https://github.com/awslabs/git-secrets)
- [TruffleHog Documentation](https://github.com/trufflesecurity/trufflehog)

---

## ✨ Summary

| Layer | Protection | Tool |
|-------|-----------|------|
| **Code** | No hardcoded secrets | TruffleHog, grep |
| **Files** | .env not committed | .gitignore |
| **Repository** | Secrets stored safely | GitHub Secrets |
| **CI/CD** | Validation on every push | GitHub Actions |
| **Audit** | Daily rotation reminders | Scheduled workflows |

**Status:** 🟢 All security measures in place
