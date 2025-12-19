# Security Policy

## Reporting Security Vulnerabilities

If you discover a security vulnerability in go-agentic, please email security@example.com instead of using the issue tracker. Please include the following details:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## Security Practices

### Environment Variables
- Never commit `.env` files to the repository
- All `.env` files are in `.gitignore`
- Use environment variables for sensitive configuration
- Example: `OPENAI_API_KEY` must be set via environment, not in code

### API Keys and Credentials
- No hardcoded API keys, passwords, or tokens
- All credentials must come from environment variables
- Use `.env.example` as template for required variables

### Dependencies
- All dependencies are locked in `go.sum`
- Regular security audits via `govulncheck`
- Dependencies are verified on every build
- OpenAI SDK is kept up-to-date (v3.15.0)

### Code Review
- All code changes go through GitHub Actions checks:
  - Unit tests (race detection enabled)
  - Security scanning with gosec
  - Code quality checks with golangci-lint
  - Go fmt and go vet verification
  - Dependency vulnerability scanning

### CI/CD Security
- GitHub Actions workflows enforce security checks
- All commits must pass security, quality, and test gates
- Release process includes final security verification
- Automatic scanning of dependencies daily

### Secrets Management
- GitHub Secrets are used for sensitive data
- No credentials in workflow files
- Token rotation on regular schedule
- Audit logs reviewed for anomalies

## Security Checklist

Every commit must pass:
- ✅ No hardcoded secrets (gosec)
- ✅ No vulnerable dependencies (govulncheck)
- ✅ Code quality standards (golangci-lint)
- ✅ All tests pass (race-safe)
- ✅ Proper formatting (go fmt)
- ✅ Module integrity (go mod verify)

## Threat Model

### Addressed Risks
1. **Exposed credentials** - Mitigated with .gitignore and secret scanning
2. **Vulnerable dependencies** - Mitigated with govulncheck and daily scans
3. **Code injection** - Mitigated with static analysis and code review
4. **Configuration issues** - Mitigated with YAML validation and testing

### Out of Scope
- Runtime exploitation
- Supply chain attacks (dependencies assumed trusted)
- Physical security
- Social engineering

## Compliance

### Supported Go Versions
- Go 1.25.5 (current)
- Backward compatibility: Go 1.23+ recommended

### Dependency Policy
- OpenAI SDK: v3.15.0+
- YAML: v3.0.1+
- All indirect dependencies verified

## Version Support

| Version | Status | Security Updates |
|---------|--------|------------------|
| v0.0.1-alpha.1 | Active | Until v1.0.0 |

## Security Updates

Security updates will be released as soon as possible after discovery. Users should:
1. Enable GitHub notifications for releases
2. Subscribe to security advisories
3. Update dependencies regularly
4. Test in staging before production

## Questions?

For security questions or concerns, please reach out to the maintainers directly.
