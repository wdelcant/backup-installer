# Security Policy

## 🔐 Security Architecture

### Encryption

- **Algorithm**: AES-256-GCM
- **Key Size**: 256 bits
- **Key Storage**: `~/.config/invitsm-backup/.invitsm-master-key`
- **Key Permissions**: 0400 (owner read-only)

### Sensitive Data Protection

| Data Type | Protection |
|-----------|-----------|
| Database passwords | Encrypted with AES-256-GCM |
| Webhook tokens | Encrypted with AES-256-GCM |
| Master key | Stored outside repo, 0400 permissions |
| Config file | 0600 permissions, gitignored |

### File Permissions

```
~/.config/invitsm-backup/.invitsm-master-key  →  0400
./config/config.yaml                          →  0600
./config/                                     →  0700
./scripts/pipeline.sh                         →  0755
./logs/                                       →  0755
```

## 🚨 Reporting a Vulnerability

If you discover a security vulnerability, please report it privately by opening an issue and marking it as confidential.

**DO NOT** create a public issue for security vulnerabilities.

## ✅ Security Checklist for Contributors

Before submitting a PR, ensure:

- [ ] No hardcoded credentials in code
- [ ] No real IP addresses or hostnames in examples
- [ ] All sensitive fields are encrypted
- [ ] File permissions are set correctly
- [ ] Security audit passes: `make audit`
- [ ] No sensitive data in logs

## 🔍 Security Audit

Run the security audit before every commit:

```bash
make audit
```

This checks for:
- Database credentials
- Connection strings
- API tokens
- Real hostnames/IPs
- Generated files

## 📦 Dependencies

We use Dependabot to automatically update dependencies. Security updates are prioritized.

## 🛡️ Best Practices

1. **Principle of Least Privilege**: Database users should have minimal required permissions
2. **Defense in Depth**: Multiple security layers (encryption, permissions, gitignore)
3. **Secure Defaults**: All security features enabled by default
4. **No Secrets in Logs**: Sensitive data is redacted from logs

## 🔄 Security Updates

Security patches are released as soon as possible. Watch the repository for security advisories.

---

**Last Updated**: March 2026