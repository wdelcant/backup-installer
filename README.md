# 🗄️ Backup Installer

Instalador TUI interactivo para configurar backups automáticos de PostgreSQL con encriptación AES-256-GCM.

## ✨ Características

- **🔐 Seguridad**: Credenciales encriptadas con AES-256-GCM
- **🎨 TUI Interactiva**: Wizard fácil de usar con Bubble Tea
- **⏰ Programable**: Cron jobs configurables
- **🔄 Pipeline**: Backup → Restore automático en QA/Dev
- **📡 Webhooks**: Notificaciones a n8n
- **🐧 Multiplataforma**: Linux, macOS, Windows

## 📋 Requisitos

- Go 1.21+ (solo para desarrollo)
- PostgreSQL client tools (`pg_dump`, `psql`)
- Cron disponible en el sistema (Linux/macOS)

## 🚀 Instalación Rápida

### Opción 1: One-liner (Recomendado)

```bash
curl -fsSL https://raw.githubusercontent.com/wdelcant/backup-installer/main/install.sh | bash
```

### Opción 2: Descarga Manual

Descarga el binario para tu plataforma desde [releases](https://github.com/wdelcant/backup-installer/releases):

```bash
# Linux AMD64
wget https://github.com/wdelcant/backup-installer/releases/latest/download/backup-installer-linux-amd64
chmod +x backup-installer-linux-amd64
sudo mv backup-installer-linux-amd64 /usr/local/bin/backup-installer

# Linux ARM64
wget https://github.com/wdelcant/backup-installer/releases/latest/download/backup-installer-linux-arm64
chmod +x backup-installer-linux-arm64
sudo mv backup-installer-linux-arm64 /usr/local/bin/backup-installer

# macOS AMD64 (Intel)
wget https://github.com/wdelcant/backup-installer/releases/latest/download/backup-installer-darwin-amd64
chmod +x backup-installer-darwin-amd64
sudo mv backup-installer-darwin-amd64 /usr/local/bin/backup-installer

# macOS ARM64 (Apple Silicon)
wget https://github.com/wdelcant/backup-installer/releases/latest/download/backup-installer-darwin-arm64
chmod +x backup-installer-darwin-arm64
sudo mv backup-installer-darwin-arm64 /usr/local/bin/backup-installer
```

### Opción 3: Compilar desde el código fuente

```bash
# Clonar el repositorio
git clone https://github.com/wdelcant/backup-installer.git
cd backup-installer

# Compilar
make build

# Ejecutar instalador
sudo ./bin/backup-installer
```

## 🎯 Uso

### Ejecutar el instalador

```bash
backup-installer
```

El wizard te guiará para configurar:
1. **Base de datos origen** (producción)
2. **Base de datos destino** (QA/Dev - opcional)
3. **Horario de ejecución** (expresión cron)
4. **Retención de backups**
5. **Notificaciones webhook** (opcional)

### Comandos Makefile (desarrollo)

```bash
make build        # Compila el binario
make install      # Ejecuta el wizard de instalación
make run-now      # Ejecuta backup manual
make logs         # Ver logs en tiempo real
make uninstall    # Remueve configuración
make clean        # Limpia archivos de build
make test         # Ejecuta tests
```

### Después de Instalar

El instalador generará:

1. **Configuración**: `./config/config.yaml` (encriptado)
2. **Scripts**: `./scripts/pipeline.sh`
3. **Cron job**: Programado automáticamente
4. **Logs**: `./logs/pipeline-YYYY-MM-DD.log`

## 🔐 Seguridad

- **Clave maestra**: `~/.config/backup-installer/.master-key`
- **Permisos**: 0400 (solo owner lectura)
- **Algoritmo**: AES-256-GCM
- **Config**: 0600 (solo owner lectura/escritura)

## 📁 Estructura

```
backup-installer/
├── cmd/installer/          # Entry point
├── internal/
│   ├── logo/               # ASCII art
│   ├── crypto/             # Encriptación
│   ├── config/             # Configuración
│   ├── tui/                # Interfaz Bubble Tea
│   ├── pipeline/           # Generador scripts
│   ├── cron/               # Instalador cron
│   └── webhook/            # Notificaciones
├── config/                 # Configuración (gitignore)
├── scripts/                # Scripts generados (gitignore)
├── logs/                   # Logs (gitignore)
└── bin/                    # Binario (gitignore)
```

## 🔔 Webhook n8n

El webhook envía payloads JSON a n8n con:

```json
{
  "event": "backup.pipeline.completed",
  "timestamp": "2026-03-31T02:35:17Z",
  "pipeline": {
    "status": "success",
    "duration_seconds": 2100
  },
  "backup": {
    "status": "success",
    "file": {
      "size_bytes": 128450560,
      "size_human": "122.5 MB"
    }
  },
  "restore": {
    "status": "success",
    "target": {
      "database": "your_development_db"
    }
  }
}
```

## 🧹 Desinstalar

```bash
# Remover crontab y scripts
rm ~/.config/backup-installer/.master-key
rm -rf ./config ./scripts
```

## 🛠️ Desarrollo

```bash
# Clonar repositorio
git clone https://github.com/wdelcant/backup-installer.git
cd backup-installer

# Instalar dependencias
go mod download

# Ejecutar en modo desarrollo
make dev

# Correr tests
make test

# Build para producción
make build
```

## 📝 Licencia

MIT

---

**Versión**: 1.0.0  
**Repositorio**: [github.com/wdelcant/backup-installer](https://github.com/wdelcant/backup-installer)
