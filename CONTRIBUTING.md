# Contributing to INVITSM Backup Installer

## 🔐 Security Guidelines

### Before Committing

**ALWAYS** run the security audit before committing:

```bash
make audit
```

This checks for:
- ❌ Hardcoded database credentials
- ❌ Real IP addresses or hostnames
- ❌ Database connection strings
- ❌ API tokens or secrets
- ❌ Generated files that should be in .gitignore

### What NOT to Commit

- [ ] `config/config.yaml` - Contains encrypted credentials
- [ ] `bin/backup-installer` - Compiled binary
- [ ] `*.log` files - May contain sensitive information
- [ ] `go.work` - Local development file
- [ ] Any file with real credentials

### What TO Commit

- [ ] Source code (`.go` files)
- [ ] Templates (`.tmpl` files)
- [ ] Configuration examples (`*.example.yaml`)
- [ ] Documentation (`.md` files)
- [ ] Build scripts (`Makefile`, `install.sh`)
- [ ] `go.mod` and `go.sum`

## 🛠️ Development Workflow

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/my-feature`
3. **Make your changes**
4. **Run security audit**: `make audit`
5. **Build and test**: `make build && make test`
6. **Commit with conventional commits**: `git commit -m "feat: add new feature"`
7. **Push and open PR**

## 📝 Code Style

- Use meaningful variable names
- Add comments for complex logic
- Follow Go best practices
- Keep functions small and focused
- Handle errors properly

## 🔒 Security Best Practices

1. **Never hardcode credentials** - Use configuration files
2. **Encrypt sensitive data** - Use the crypto package
3. **Validate all inputs** - Especially in TUI forms
4. **Use secure file permissions** - 0400/0600 for sensitive files
5. **Log carefully** - Never log passwords or tokens

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific test
go test ./internal/crypto -v

# Run security audit
make audit
```

## 📦 Release Process

1. Update version in `main.go`
2. Update `CHANGELOG.md`
3. Tag the release: `git tag v1.0.0`
4. Push tag: `git push origin v1.0.0`
5. Create GitHub release with binaries

---

**Questions?** Open an issue!