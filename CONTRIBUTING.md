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

## 🔄 Conventional Commits & Versionado Automático

Este proyecto usa [Conventional Commits](https://www.conventionalcommits.org/) para el versionado automático semántico.

### Tipos de Commit

| Tipo | Descripción | Versión |
|------|-------------|---------|
| `fix:` | Corrección de bug | Patch (1.0.0 → 1.0.1) |
| `feat:` | Nueva funcionalidad | Minor (1.0.0 → 1.1.0) |
| `BREAKING CHANGE:` | Cambio incompatible | Major (1.0.0 → 2.0.0) |
| `docs:` | Documentación | - |
| `style:` | Formato de código | - |
| `refactor:` | Refactorización | - |
| `perf:` | Performance | - |
| `test:` | Tests | - |
| `chore:` | Mantenimiento | - |
| `ci:` | CI/CD | - |

### Ejemplos

```bash
# Bug fix (genera v1.0.1)
git commit -m "fix: resolve cron parsing error"

# Nueva feature (genera v1.1.0)
git commit -m "feat: add webhook notifications"

# Breaking change (genera v2.0.0)
git commit -m "feat: new auth system

BREAKING CHANGE: old tokens no longer work"

# Con scope
git commit -m "feat(tui): add progress bar"
```

## 📦 Release Process (Automático)

1. Hacer commits siguiendo Conventional Commits
2. Mergear a `main`
3. GitHub Actions detecta los cambios y crea el release automáticamente
4. Los binarios se compilan para todas las plataformas

**Nota**: Si usas `git push` sin commits tipo `feat:` o `fix:`, no se generará un nuevo release.

---

**Questions?** Open an issue!