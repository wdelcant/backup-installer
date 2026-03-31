# 🗄️ INVITSM Backup Installer

Instalador TUI interactivo para configurar backups automáticos de PostgreSQL con encriptación AES-256-GCM.

## ✨ Características

- **🔐 Seguridad**: Credenciales encriptadas con AES-256-GCM
- **🎨 TUI Interactiva**: Wizard fácil de usar con Bubble Tea
- **⏰ Programable**: Cron jobs configurables
- **🔄 Pipeline**: Backup → Restore automático en QA/Dev
- **📡 Webhooks**: Notificaciones a n8n
- **🐧 Multiplataforma**: Debian, Ubuntu, CentOS, Arch

## 📋 Requisitos

- Go 1.21+
- PostgreSQL client tools (`pg_dump`, `psql`)
- Cron disponible en el sistema

## 🚀 Instalación Rápida

### Opción 1: One-liner (Recomendado)

```bash
curl -fsSL https://raw.githubusercontent.com/invitsm/invitsm/main/tools/backup-installer/install.sh | bash
```

### Opción 2: Manual

```bash
# Clonar o copiar el proyecto
cd tools/backup-installer

# Compilar
make build

# Ejecutar instalador
sudo make install
```

## 🎯 Uso

### Comandos Principales

```bash
make build        # Compila el binario
make install      # Ejecuta el wizard de instalación
make run-now      # Ejecuta backup manual
make logs         # Ver logs en tiempo real
make uninstall    # Remueve configuración
make clean        # Limpia archivos de build
```

### Después de Instalar

El instalador generará:

1. **Configuración**: `./config/config.yaml` (encriptado)
2. **Scripts**: `./scripts/pipeline.sh`
3. **Cron job**: Programado automáticamente
4. **Logs**: `./logs/pipeline-YYYY-MM-DD.log`

## 🔐 Seguridad

- **Clave maestra**: `~/.config/invitsm-backup/.invitsm-master-key`
- **Permisos**: 0400 (solo owner lectura)
- **Algoritmo**: AES-256-GCM
- **Config**: 0600 (solo owner lectura/escritura)

## 📁 Estructura

```
tools/backup-installer/
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
make uninstall

# Opcional: eliminar clave maestra
rm ~/.config/invitsm-backup/.invitsm-master-key
```

## 🛠️ Desarrollo

```bash
# Run en modo desarrollo
make dev

# Correr tests
make test

# Build para producción
make build
```

## 📝 Licencia

Propietario - INVITSM

---

**Versión**: 1.0.0  
**Última actualización**: Marzo 2026