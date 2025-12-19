# CI/CD Pipeline

## Overview

The go-agentic project uses GitHub Actions for comprehensive continuous integration and deployment (CI/CD).

## Pipeline Status

[![Tests](https://github.com/taipm/go-agentic/actions/workflows/tests.yml/badge.svg)](https://github.com/taipm/go-agentic/actions/workflows/tests.yml)
[![Security Scan](https://github.com/taipm/go-agentic/actions/workflows/security.yml/badge.svg)](https://github.com/taipm/go-agentic/actions/workflows/security.yml)
[![Code Quality](https://github.com/taipm/go-agentic/actions/workflows/quality.yml/badge.svg)](https://github.com/taipm/go-agentic/actions/workflows/quality.yml)
[![Dependency Check](https://github.com/taipm/go-agentic/actions/workflows/dependencies.yml/badge.svg)](https://github.com/taipm/go-agentic/actions/workflows/dependencies.yml)
[![Build](https://github.com/taipm/go-agentic/actions/workflows/build.yml/badge.svg)](https://github.com/taipm/go-agentic/actions/workflows/build.yml)

## Workflows

### 1. Tests (`tests.yml`)
**Triggers**: Push to main, Pull requests

**Checks**:
- ✓ Download and verify dependencies
- ✓ Build library
- ✓ Run unit tests with race detection
- ✓ Build all examples
- ✓ Coverage analysis

**Go Versions**: 1.25.5

```bash
go test -v -race -cover ./...
```

### 2. Security (`security.yml`)
**Triggers**: Push to main, Pull requests

**Checks**:
- ✓ Static security analysis with gosec
- ✓ Search for exposed credentials
- ✓ Verify .gitignore coverage
- ✓ Check git history for sensitive files
- ✓ SARIF report upload to GitHub CodeQL

**Tools**:
- gosec: Security scanner for Go
- Custom patterns for API keys

### 3. Code Quality (`quality.yml`)
**Triggers**: Push to main, Pull requests

**Checks**:
- ✓ golangci-lint: Multi-linter framework
- ✓ go vet: Official Go analyzer
- ✓ go fmt: Code formatting verification
- ✓ go mod tidy: Dependency cleanup
- ✓ go mod verify: Integrity check

**Linters**:
- gofmt, govet, errcheck, staticcheck, and more

### 4. Dependencies (`dependencies.yml`)
**Triggers**: Push to main, Pull requests, Daily at 00:00 UTC

**Checks**:
- ✓ govulncheck: Vulnerability detection
- ✓ License verification
- ✓ Version pinning validation
- ✓ go.sum integrity verification
- ✓ Outdated package detection

**Frequency**: Daily automated scan

### 5. Build (`build.yml`)
**Triggers**: Push to main, Pull requests

**Builds**:
- ✓ Library compilation
- ✓ All 4 examples
- ✓ Artifact upload

**Caching**: Go modules cached for speed

### 6. Release (`release.yml`)
**Triggers**: Git tags matching `v*`

**Actions**:
- ✓ Build verification
- ✓ Final security checks
- ✓ Test execution
- ✓ GitHub Release creation
- ✓ Pre-release marking

## Protection Rules

### Branch Protection (main)

The main branch has the following protection rules:

1. **Require status checks to pass before merging**:
   - Tests must pass
   - Security scan must pass
   - Code quality must pass
   - Build must succeed

2. **Require branches to be up to date before merging**: Yes

3. **Require code reviews**: Recommended (1+ reviews)

4. **Require CODEOWNERS review**: Yes

## Security Checks Details

### Hardcoded Secrets Detection
```bash
grep -r "OPENAI_API_KEY\s*=" --include="*.go"
grep -r "aws_access_key\|azure_.*_key" --include="*.go"
```

### .gitignore Verification
- `.env` files are ignored
- Sensitive patterns are ignored
- All except: examples, go-agentic, docs, etc.

### Dependency Scanning
- Runs daily
- Checks for known vulnerabilities
- Verifies version constraints
- Validates go.sum integrity

## Performance

### Build Times
- Tests: ~2 minutes
- Security scan: ~1 minute
- Code quality: ~1 minute
- Dependencies: ~30 seconds
- Build: ~1 minute

**Total**: ~5-6 minutes per commit

### Caching Strategy
- Go modules cached per go.sum
- Artifacts uploaded for releases
- Cache invalidation on dependency changes

## Artifacts

### Test Reports
Generated on every test run:
- Coverage reports
- Test logs
- Race detection results

### Build Artifacts
Created for every build:
- Compiled examples
- Binary artifacts
- Available for 90 days

### Release Artifacts
Created on version tags:
- GitHub Release page
- Release notes
- Pre-release marking

## Debugging CI Failures

### Check Logs
1. Go to GitHub Actions tab
2. Click the failed workflow
3. Expand the step that failed
4. Review logs

### Common Issues

**Tests fail**:
```bash
# Run locally
cd go-agentic
go test -v -race ./...
```

**Security scan fails**:
```bash
# Install and run gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...
```

**Linting fails**:
```bash
# Install and run golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run
```

**Build fails**:
```bash
# Check dependencies
go mod verify
go mod tidy

# Try building
go build ./...
```

## GitHub Actions Configuration

All workflows are in `.github/workflows/`:
- `tests.yml`
- `security.yml`
- `quality.yml`
- `dependencies.yml`
- `build.yml`
- `release.yml`

### Secrets Required
None - all checks use open-source tools

### Environment Variables
- `GITHUB_TOKEN`: Automatic (used for releases)

## Best Practices

### For Developers
1. Run tests locally before pushing: `go test -v ./...`
2. Format code: `go fmt ./...`
3. Verify linting: `golangci-lint run`
4. Update dependencies: `go mod tidy`

### For Reviewers
1. Ensure all CI checks pass
2. Review the security scan results
3. Check code quality metrics
4. Verify test coverage

### For Maintainers
1. Keep workflows updated
2. Monitor dependency updates
3. Review security advisories
4. Update Go version when released

## Maintenance

### Workflow Updates
Workflows are reviewed and updated:
- On Go releases
- On tool updates
- On security advisories
- Monthly for best practices

### Dependency Updates
- golangci-lint: Auto-updated by Dependabot
- gosec: Auto-updated by Dependabot
- Go: Manual updates (currently 1.25.5)

## Future Enhancements

- [ ] Code coverage reporting to Codecov
- [ ] Performance benchmarking
- [ ] Integration test matrix
- [ ] API documentation generation
- [ ] Automated release notes
- [ ] Docker image building

---

For more information, see [SECURITY.md](SECURITY.md)
