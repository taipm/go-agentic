# 🔐 Secret Rotation & Key Management Procedures

**Project**: go-agentic  
**Last Updated**: 2025-12-21  
**Version**: 1.0

---

## 📋 Overview

This document outlines the procedures for rotating secrets and managing API keys for the go-agentic project. Regular secret rotation is a critical security practice that reduces the risk of compromised credentials.

### Key Principles
- **Regular Rotation**: Minimize exposure window for stolen credentials
- **Immediate Action**: Rotate secrets if compromise is suspected
- **Audit Trail**: Document all rotations and personnel involved
- **Zero-Downtime**: Plan rotations to minimize service disruption
- **Automation**: Use GitHub Actions for automated rotation reminders

---

## 📅 Rotation Schedule

### Monthly Rotation (Every First Monday)
**Time**: 9:00 AM (local time)  
**Owner**: Lead Developer / Tech Lead

**Checklist**:
- [ ] Review GitHub Secrets (Settings → Secrets and variables → Actions)
- [ ] Verify OPENAI_API_KEY validity and usage
- [ ] Check for unused or expired secrets
- [ ] Review access logs in GitHub Actions
- [ ] Document any issues found
- [ ] No rotation needed if secrets are valid and secure

### Quarterly Rotation (Every 90 Days: Jan 1, Apr 1, Jul 1, Oct 1)
**Time**: 9:00 AM (local time)  
**Owner**: Tech Lead / Security Officer

**Checklist**:
- [ ] Rotate all API keys (OPENAI_API_KEY, etc.)
- [ ] Update .env.example if structure changed
- [ ] Audit all GitHub Actions using secrets
- [ ] Review CloudFlare/hosting credentials
- [ ] Check database connection strings
- [ ] Review third-party service credentials
- [ ] Update documentation if endpoints changed

### Annual Security Review (December 31)
**Time**: 10:00 AM (local time)  
**Owner**: Tech Lead / Security Officer / Team

**Checklist**:
- [ ] Complete audit of all secrets and credentials
- [ ] Review access permissions and GitHub organization settings
- [ ] Assess any security incidents or near-misses
- [ ] Update security documentation
- [ ] Review and confirm all team members' access levels
- [ ] Plan next year's rotation schedule

---

## 🔄 How to Rotate OPENAI_API_KEY

### Step 1: Generate New API Key
1. Log in to [OpenAI Platform](https://platform.openai.com/account/api-keys)
2. Click "Create new secret key"
3. Give it a descriptive name: `go-agentic-prod-YYYYMMDD`
4. Copy the new key immediately (it won't be shown again)

### Step 2: Update GitHub Secrets
1. Go to GitHub repository settings
2. Navigate to **Settings → Secrets and variables → Actions**
3. Click "New repository secret"
4. Name: `OPENAI_API_KEY`
5. Value: Paste the new API key
6. Click "Add secret"

**GitHub UI Navigation**:
```
Repository → Settings (gear icon)
  → Secrets and variables (left sidebar)
  → Actions
  → New repository secret
```

### Step 3: Update Local Development Files
1. Update `.env` file in `/examples/it-support/`:
   ```bash
   OPENAI_API_KEY=sk-proj-NEW_KEY_HERE
   ```

2. Update `.env` file in `/examples/vector-search/`:
   ```bash
   OPENAI_API_KEY=sk-proj-NEW_KEY_HERE
   ```

3. **Important**: Never commit `.env` files. They are in `.gitignore`.

### Step 4: Verify New Key Works
1. In `/examples/it-support/`:
   ```bash
   go run ./cmd/main.go "What is your purpose?"
   ```

2. In `/examples/vector-search/`:
   ```bash
   go run ./cmd/main.go "Test query"
   ```

3. Verify both examples run without "OPENAI_API_KEY not found" errors

### Step 5: Revoke Old Key
1. Go back to [OpenAI Platform API Keys](https://platform.openai.com/account/api-keys)
2. Find the old key by creation date
3. Click the delete/trash icon next to old key
4. Confirm deletion

### Step 6: Document Rotation
Create a brief log entry:
```markdown
## YYYYMMDD - OPENAI_API_KEY Rotation
- **Date**: 2025-MM-DD HH:MM UTC
- **Reason**: Monthly rotation
- **Old Key**: sk-proj-...XXXX (last 4 digits)
- **New Key**: sk-proj-...YYYY (last 4 digits)
- **Rotated By**: [Your GitHub Username]
- **Verified**: ✅ Both examples tested successfully
- **Notes**: No issues found
```

---

## 🚨 Emergency Rotation (Suspected Compromise)

If you suspect a secret has been compromised:

### Immediate Actions (Within 5 Minutes)
1. **Revoke immediately** at the service provider (e.g., OpenAI)
2. **Update GitHub Secrets** with a temporary placeholder:
   ```
   OPENAI_API_KEY=sk-proj-TEMPORARY-PLACEHOLDER-WHILE-ROTATING
   ```
3. **Notify team lead** via Slack/email immediately

### Assessment Phase (Within 1 Hour)
1. Check GitHub Actions logs for any suspicious activity
2. Review API usage in OpenAI dashboard for unusual patterns
3. Check git history for accidental commits
4. Review `.env` files for unexpected changes
5. Document timeline of suspected compromise

### Recovery Phase (Within 2 Hours)
1. Generate new API key (follow "How to Rotate OPENAI_API_KEY" steps)
2. Update all services and GitHub Secrets
3. Run integration tests to verify new key works
4. Create incident report (see template below)

### Post-Incident
1. Review how secret was exposed
2. Implement additional safeguards if needed
3. Share findings with team
4. Update documentation

---

## 📝 Incident Report Template

```markdown
# Incident Report: Secret Compromise
**Date**: YYYY-MM-DD  
**Time**: HH:MM UTC  
**Secret**: [e.g., OPENAI_API_KEY]  
**Status**: RESOLVED / IN PROGRESS

## Timeline
- **HH:MM**: Initial detection/report
- **HH:MM**: Root cause identified
- **HH:MM**: Secret revoked
- **HH:MM**: New secret deployed
- **HH:MM**: Verification complete

## Root Cause
[Describe how secret was exposed - e.g., accidental commit, log exposure, etc.]

## Impact Assessment
- **Services Affected**: [List]
- **Data Exposed**: [Describe]
- **Duration**: [Time window]
- **Severity**: [LOW / MEDIUM / HIGH / CRITICAL]

## Remediation Steps
- [ ] Secret revoked at provider
- [ ] New secret generated
- [ ] All systems updated
- [ ] Integration tests passed
- [ ] Team notified
- [ ] Documentation updated

## Prevention Measures
[Describe changes made to prevent recurrence]

## Lessons Learned
[Key takeaways and improvements]

**Incident Closed By**: [Name]  
**Date Closed**: YYYY-MM-DD  
**Review Due Date**: YYYY-MM-DD (30 days later)
```

---

## 🔐 Other Secrets Management

### Database Credentials
If applicable, follow the same quarterly rotation schedule:
- Update `.env` files
- Update GitHub Secrets if stored there
- Update any connection pooling services
- Verify database connectivity after rotation

### API Keys for Third-Party Services
Apply the same rotation schedule:
- Stripe API keys (if applicable)
- Firebase credentials (if applicable)
- Any cloud provider credentials (AWS, GCP, Azure)

**For each service**:
1. Generate new credentials
2. Update GitHub Secrets
3. Update `.env` files
4. Test functionality
5. Revoke old credentials
6. Document rotation

---

## 🛡️ Security Best Practices

### DO ✅
- ✅ Store secrets in GitHub Secrets (Settings → Secrets)
- ✅ Use environment variables in CI/CD workflows
- ✅ Rotate secrets regularly (quarterly minimum)
- ✅ Use `.env.example` with placeholder values only
- ✅ Add all secret files to `.gitignore`
- ✅ Document rotation dates and responsible parties
- ✅ Use secret-specific names: `service-env-date` format
- ✅ Rotate immediately if compromise suspected

### DON'T ❌
- ❌ Store secrets in code repositories
- ❌ Commit `.env` files to git
- ❌ Share secrets via email or chat
- ❌ Use generic secret names
- ❌ Keep old secrets active during rotation
- ❌ Leave rotation logs in plaintext with actual key values
- ❌ Use the same secret across environments
- ❌ Ignore rotation reminders

---

## 📊 Rotation Log

Keep this log updated after each rotation:

| Date | Secret | Rotated By | Status | Notes |
|------|--------|-----------|--------|-------|
| 2025-12-21 | OPENAI_API_KEY | [Your Name] | ✅ Complete | Initial setup |
| | | | | |

---

## 🔗 Related Documentation

- **[SECURITY.md](.github/SECURITY.md)** - Security guidelines
- **[CICD_SECURITY_AUDIT.md](CICD_SECURITY_AUDIT.md)** - CI/CD security audit
- **.gitignore** - Files excluded from git
- **[.env.example](examples/it-support/.env.example)** - Environment template

---

## 📞 Quick Reference

### GitHub Secrets URL
```
https://github.com/taipm/go-agentic/settings/secrets/actions
```

### OpenAI API Keys
```
https://platform.openai.com/account/api-keys
```

### Rotation Reminder Automation
The `.github/workflows/secrets-check.yml` workflow runs daily at 2 AM UTC and includes:
- Daily secret detection scans
- API key rotation reminders for scheduled jobs
- Dependency vulnerability checks

---

## ✅ Approval & Sign-Off

**Document Version**: 1.0  
**Created**: 2025-12-21  
**Reviewed By**: Security Audit  
**Next Review**: 2025-12-28  
**Rotation Schedule**: Active (Monthly, Quarterly, Annual)

---

**Important**: This document should be reviewed and updated annually or whenever the project's secret management requirements change.
