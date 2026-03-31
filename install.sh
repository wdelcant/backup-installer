#!/bin/bash
# =============================================================================
# INVITSM Backup Installer - One-liner Installation Script
# =============================================================================
# Usage: curl -fsSL https://.../install.sh | bash
# =============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    log_error "No ejecutar como root. El script solicitará sudo cuando sea necesario."
    exit 1
fi

# Check for Go
if ! command -v go &> /dev/null; then
    log_error "Go no está instalado. Instalar desde https://golang.org/dl/"
    exit 1
fi

log_info "Go version: $(go version)"

# Check for PostgreSQL client
if ! command -v pg_dump &> /dev/null; then
    log_warn "PostgreSQL client no encontrado"
    log_info "Instalar con: sudo apt-get install postgresql-client"
    exit 1
fi

log_info "PostgreSQL client: $(pg_dump --version | head -1)"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

log_info "Building backup installer..."
make build

if [ ! -f "./bin/backup-installer" ]; then
    log_error "Build falló. Verificar errores arriba."
    exit 1
fi

log_info "✅ Build completado exitosamente"
echo ""
echo "Para ejecutar el instalador:"
echo "  sudo make install"
echo ""
echo "O ejecutar directamente:"
echo "  sudo ./bin/backup-installer"
echo ""