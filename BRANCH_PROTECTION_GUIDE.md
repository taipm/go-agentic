# 🛡️ Branch Protection Rules Configuration Guide

**Project**: go-agentic  
**Last Updated**: 2025-12-21  
**Version**: 1.0

---

## 📋 Overview

Branch protection rules are critical security controls that enforce code quality and prevent accidental or malicious changes to important branches. This guide documents the recommended configuration for the `main` and `develop` branches.

### Why Branch Protection?
- **Prevents Direct Pushes**: Code must go through pull requests and reviews
- **Enforces Status Checks**: Tests, linting, and security checks must pass
- **Requires Reviews**: Code must be reviewed before merging
- **Protects History**: Prevents force pushes and deletions
- **Audit Trail**: All changes are documented

---

## 🔧 Recommended Configuration

### For `main` Branch (Production)

| Setting | Recommended | Purpose |
|---------|-------------|---------|
| **Require status checks to pass** | ✅ YES | Code quality gates |
| **Require branches to be up to date** | ✅ YES | No stale merges |
| **Require pull request reviews** | ✅ YES | Code review requirement |
| **Required review count** | 1 | Allow flexibility |
| **Dismiss stale reviews** | ✅ YES | Force re-review of changes |
| **Require review from CODEOWNERS** | ⚠️ Optional | If CODEOWNERS file exists |
| **Require up-to-date PRs** | ✅ YES | Test latest main |
| **Include administrators** | ✅ YES | Apply to all |
| **Restrict who can push** | ✅ YES | Limit to maintainers |
| **Allow force pushes** | ❌ NO | Protect commit history |
| **Allow deletions** | ❌ NO | Prevent branch deletion |

### For `develop` Branch (Development)

| Setting | Recommended | Purpose |
|---------|-------------|---------|
| **Require status checks to pass** | ✅ YES | Code quality gates |
| **Require branches to be up to date** | ✅ YES | No stale merges |
| **Require pull request reviews** | ✅ YES | Code review requirement |
| **Required review count** | 1 | Slightly more flexible |
| **Dismiss stale reviews** | ✅ YES | Force re-review |
| **Require up-to-date PRs** | ✅ YES | Test latest develop |
| **Include administrators** | ✅ YES | Apply to all |
| **Allow force pushes** | ❌ NO | Protect history |
| **Allow deletions** | ❌ NO | Prevent deletion |

---

## 📌 Required Status Checks

The following workflow checks **MUST** pass before merging:

```
✅ security-checks       (ci.yml)
✅ lint                  (ci.yml)
✅ build-matrix          (ci.yml)  [for multiple Go versions]
✅ dependency-check      (ci.yml)  [nancy vulnerability scan]
✅ integration-tests     (integration-tests.yml)
✅ type-safety-check     (integration-tests.yml)
```

**Why These Checks**:
- **security-checks**: Detects hardcoded secrets using TruffleHog and custom patterns
- **lint**: Enforces code style (gofmt, goimports, golangci-lint)
- **build-matrix**: Verifies code compiles on multiple Go versions
- **dependency-check**: Scans for dependency vulnerabilities
- **integration-tests**: Runs integration tests on actual code paths
- **type-safety-check**: Ensures type safety with staticcheck and go vet

---

## 🚀 Setup Instructions

### Step 1: Navigate to Branch Protection Settings

1. Go to your GitHub repository: `https://github.com/taipm/go-agentic`
2. Click **Settings** (⚙️ gear icon at top right)
3. In left sidebar, click **Branches**
4. You should see section: **Branch protection rules**

### Step 2: Add Rule for `main` Branch

1. Click **Add rule**
2. **Branch name pattern**: Type `main`
3. Click **Create** to start configuration

### Step 3: Configure Main Branch Settings

#### Status Checks
1. Check: **Require status checks to pass before merging**
2. Check: **Require branches to be up to date before merging**
3. Under "Status checks that are required":
   - Search for and select: `security-checks`
   - Search for and select: `lint`
   - Search for and select: `build-matrix`
   - Search for and select: `dependency-check`
   - Search for and select: `integration-tests`
   - Search for and select: `type-safety-check`

#### Pull Request Reviews
1. Check: **Require pull request reviews before merging**
2. Set **Required number of reviewers**: `1`
3. Check: **Dismiss stale pull request approvals when new commits are pushed**
4. Uncheck: **Require review from Code Owners** (optional if no CODEOWNERS)

#### Protect History
1. Check: **Include administrators** (so rules apply to everyone)
2. Uncheck: **Allow force pushes** (protect commit history)
3. Uncheck: **Allow deletions** (prevent branch deletion)

#### Additional Settings
1. Check: **Require status checks to pass before merging**
2. Check: **Require branches to be up to date before merging**

3. Click **Save changes** at bottom

### Step 4: Add Rule for `develop` Branch (Optional)

Repeat Steps 2-3 with same settings, but use `develop` as branch name pattern.

### Step 5: Verify Configuration

After saving, you should see:
```
✅ Protection rule for main
   - 6 required status checks
   - 1 required review
   - Force pushes disabled
   - Deletions disabled
```

---

## ✅ Verification Checklist

After setting up branch protection, verify each setting:

### Main Branch Verification

- [ ] **Branch pattern**: `main` (exact match)

- [ ] **Status checks enabled** and these are required:
  - [ ] `security-checks`
  - [ ] `lint`
  - [ ] `build-matrix`
  - [ ] `dependency-check`
  - [ ] `integration-tests`
  - [ ] `type-safety-check`

- [ ] **Require up to date**: ✅ YES

- [ ] **Pull request reviews**:
  - [ ] Required: ✅ YES
  - [ ] Number of reviewers: `1`
  - [ ] Dismiss stale reviews: ✅ YES

- [ ] **Push restrictions**:
  - [ ] Include administrators: ✅ YES
  - [ ] Allow force pushes: ❌ NO
  - [ ] Allow deletions: ❌ NO

### Test the Protection

1. Try to push directly to `main`:
   ```bash
   git push origin main
   ```
   **Expected**: Should be rejected with message about branch protection

2. Try to merge without PR:
   **Expected**: GitHub should prevent direct merge

3. Create a test PR and verify:
   - Status checks must pass ✅
   - Review must be approved ✅
   - Then merge should be allowed ✅

---

## 🔄 How Branch Protection Works

### Scenario 1: Developer Pushes Directly to Main
```
❌ REJECTED
Reason: "Branch protection: Cannot push directly to main"
Action: Must create a pull request instead
```

### Scenario 2: PR Created but Tests Failing
```
❌ CANNOT MERGE
Status: Security check failed (secrets detected) ❌
Action: Fix the issue and re-run tests
```

### Scenario 3: PR Created, Tests Pass, but No Review
```
⏳ WAITING
Status: Tests ✅ | Review ⏳ (needs 1)
Action: Request review from team member
```

### Scenario 4: PR Created, Tests Pass, Reviewed and Approved
```
✅ READY TO MERGE
Status: Tests ✅ | Review ✅
Action: Developer can now merge the PR
```

### Scenario 5: Stale PR with New Commits
```
⏳ RE-REVIEW NEEDED
Status: Old review dismissed (new commits pushed)
Action: Reviewer must approve again
```

---

## 🚨 Bypassing Branch Protection (Emergency Only)

**Important**: Branch protection can only be bypassed by repository administrators.

### When to Bypass
- Critical production hotfix needed immediately
- Security incident requiring urgent patch
- Infrastructure emergency

### How to Bypass (Admin Only)
1. Go to pull request
2. Look for **Merge without waiting for required reviews** option
3. This only appears for administrators
4. Document the bypass reason

**After bypass**: Create incident report and review why protection couldn't wait.

---

## 📊 Status Checks Detail

### security-checks (from ci.yml)
```yaml
Jobs:
  - TruffleHog secret detection
  - Hardcoded API key scanning
  - .env file enforcement
  - .env.example validation

Fail Conditions:
  ❌ Any secrets detected
  ❌ .env files found in commit
  ❌ API keys in code
```

### lint (from ci.yml)
```yaml
Jobs:
  - gofmt code style check
  - goimports organization
  - golangci-lint static analysis

Fail Conditions:
  ❌ Code format issues
  ❌ Unused imports
  ❌ Linting violations
```

### build-matrix (from ci.yml)
```yaml
Matrix:
  - Go 1.20
  - Go 1.21

Fail Conditions:
  ❌ Compilation errors
  ❌ Type errors
  ❌ Build failures
```

### dependency-check (from ci.yml)
```yaml
Tool: nancy (vulnerability scanner)

Fail Conditions:
  ❌ Known vulnerabilities found
  ❌ Outdated dependencies
```

### integration-tests (from integration-tests.yml)
```yaml
Tests:
  - Core library tests
  - Example builds (no real API calls)
  - Config validation
  - Coverage reporting

Fail Conditions:
  ❌ Test failures
  ❌ Example build failures
```

### type-safety-check (from integration-tests.yml)
```yaml
Tools:
  - staticcheck analysis
  - go vet verification

Fail Conditions:
  ❌ Type safety issues
  ❌ Suspicious constructs
```

---

## 🔗 Related Documentation

- **[CICD_SECURITY_AUDIT.md](CICD_SECURITY_AUDIT.md)** - Complete CI/CD audit
- **[SECURITY.md](.github/SECURITY.md)** - Security guidelines
- **[SECRET_ROTATION_PROCEDURES.md](SECRET_ROTATION_PROCEDURES.md)** - Secret rotation guide
- **[.github/workflows/ci.yml](.github/workflows/ci.yml)** - Main CI pipeline

---

## 📝 Troubleshooting

### PR Cannot Merge Despite Passing Checks

**Possible Causes**:
1. Administrator hasn't approved yet
2. Stale reviews were dismissed (need re-approval)
3. Branch is out of date with main
4. A required check is still running

**Solution**:
1. Check status at bottom of PR page
2. Request review from team member
3. Click "Update branch" if behind main
4. Wait for all checks to complete

### Getting Blocked by Branch Protection

**Example Error**:
```
fatal: unable to read from remote repository
Branch protection: Cannot push directly to main
```

**Solution**:
1. Create a feature branch instead:
   ```bash
   git checkout -b feature/my-feature
   git push origin feature/my-feature
   ```
2. Open a pull request on GitHub
3. Get it reviewed and approved
4. Merge through GitHub UI

### Status Check Not Showing in List

**Possible Causes**:
1. Workflow hasn't run yet
2. Job name doesn't exactly match
3. Workflow file not committed

**Solution**:
1. Push a commit to `main` (or any branch)
2. Let workflows run to completion
3. Status checks should then appear in branch protection settings

---

## ✅ Sign-Off & Approval

**Configuration Version**: 1.0  
**Created**: 2025-12-21  
**Based on**: CICD_SECURITY_AUDIT.md recommendations  
**Target Branches**: `main`, `develop`  

**Recommended Actions**:
- [ ] Administrator follows setup instructions
- [ ] Verification checklist completed
- [ ] Test push/PR behavior
- [ ] Document in team procedures
- [ ] Next review: 2025-12-28

---

**Important**: Branch protection rules are critical for project safety. Once configured, they cannot be accidentally bypassed by developers. This protects the entire team's work.
