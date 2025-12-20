# Security Guidelines for go-agentic

This document outlines security best practices for developers working with the go-agentic library, particularly when handling API keys and sensitive credentials.

## Environment Variables and API Keys

### Critical: Never Commit Secrets

The repository is configured to **never** commit environment files containing secrets:

- `.env` files are excluded from git tracking
- `.env.local` and `.env.*.local` files are excluded
- All API keys must be loaded from environment variables at runtime

### Setup for Development

1. **Copy the template file:**
   ```bash
   cp examples/it-support/.env.example examples/it-support/.env
   ```

2. **Add your API key:**
   ```bash
   # Edit examples/it-support/.env
   OPENAI_API_KEY=sk-proj-your-actual-api-key-here
   ```

3. **Verify it's excluded from git:**
   ```bash
   git status
   # Should NOT show your .env file
   ```

### Supported API Providers

Currently, go-agentic integrates with:

- **OpenAI**: Requires `OPENAI_API_KEY` environment variable
  - Get your key from: https://platform.openai.com/account/api-keys
  - Keys should be kept secret and rotated regularly
  - Never share keys in pull requests, issues, or documentation

## Secret Scanning and Prevention

### Pre-commit Checks

To prevent accidental secret commits, configure git hooks:

```bash
# Install a secret scanner (example using detect-secrets)
pip install detect-secrets

# Scan your repository
detect-secrets scan --baseline .secrets.baseline

# Before committing, verify no secrets are staged
detect-secrets audit .secrets.baseline
```

### If You Accidentally Commit a Secret

1. **Immediately revoke the exposed key:**
   - For OpenAI: Visit https://platform.openai.com/account/api-keys and delete the exposed key

2. **Remove from git history:**
   ```bash
   # Option 1: Use git-filter-branch (careful!)
   git filter-branch --force --index-filter \
     "git rm --cached --ignore-unmatch examples/it-support/.env" \
     --prune-empty --tag-name-filter cat -- --all

   # Option 2: Use BFG Repo-Cleaner (recommended)
   bfg --delete-files .env
   ```

3. **Force push to update remote:**
   ```bash
   git push --force-with-lease origin main
   ```

4. **Notify collaborators** to re-pull the cleaned history

## Code Review Checklist

When reviewing PRs, ensure:

- [ ] No `.env` files in the diff
- [ ] No hardcoded API keys in code
- [ ] No API keys in test fixtures or example code
- [ ] No secrets in commit messages
- [ ] Configuration uses environment variables, not defaults

## Testing with Sensitive Data

### Mock External Services

Use mocking for tests involving API calls:

```go
// Good: Mock the OpenAI client
mockClient := &mockOpenAIClient{
    // Mock implementation
}
executor := &TeamExecutor{
    client: mockClient, // Inject mock instead of real client
}

// Bad: Real API calls in tests
executor := NewTeamExecutor(team, os.Getenv("OPENAI_API_KEY"))
```

### Test Fixtures

- Never include real API keys in test files
- Use placeholder values like `sk-proj-test-key-here`
- Keep test fixtures in memory, not committed files

## Deployment Security

### Environment Variables in CI/CD

Configure secrets in your CI/CD platform (GitHub Actions, GitLab CI, etc.):

```yaml
# GitHub Actions Example
env:
  OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

### Secrets Management

For production deployments:

- Use platform-provided secrets management (AWS Secrets Manager, Azure Key Vault, etc.)
- Rotate API keys regularly (quarterly minimum)
- Audit key access and usage
- Use service-specific keys with minimal required permissions
- Consider key versioning strategies

## Reporting Security Issues

If you discover a security vulnerability:

1. **Do NOT** open a public issue
2. **Do NOT** commit or push the vulnerability
3. **Contact maintainers privately** with:
   - Description of the vulnerability
   - Steps to reproduce
   - Suggested fix if available

See the repository's security policy for responsible disclosure details.

## Resources

- [OpenAI API Security](https://platform.openai.com/docs/guides/production-best-practices)
- [OWASP: Secrets Management](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [GitHub: Removing Sensitive Data](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository)
- [Git Secrets](https://github.com/awslabs/git-secrets)
- [Detect Secrets](https://github.com/Yelp/detect-secrets)

## Summary

**Golden Rules:**

1. ✅ Always use `.env.example` as a template
2. ✅ Load secrets from environment variables
3. ✅ Check `.gitignore` before committing
4. ✅ Review diffs for accidental secrets
5. ✅ Rotate keys if exposure is suspected
6. ❌ Never hardcode secrets
7. ❌ Never commit `.env` files
8. ❌ Never share keys in public repositories
