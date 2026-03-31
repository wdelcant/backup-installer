#!/bin/bash
# =============================================================================
# Backup Installer - One-liner Installation Script
# =============================================================================
# Usage: curl -fsSL https://raw.githubusercontent.com/wdelcant/backup-installer/main/install.sh | bash
# =============================================================================

set -e

# Configuration
REPO="wdelcant/backup-installer"
INSTALL_DIR="/usr/local/bin"
FALLBACK_DIR="$HOME/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# Detect OS
detect_os() {
    local os
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        CYGWIN*|MINGW*|MSYS*) os="windows";;
        *)
            log_error "Sistema operativo no soportado: $(uname -s)"
            exit 1
            ;;
    esac
    echo "$os"
}

# Detect architecture
detect_arch() {
    local arch
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64";;
        arm64|aarch64)  arch="arm64";;
        *)
            log_error "Arquitectura no soportada: $(uname -m)"
            exit 1
            ;;
    esac
    echo "$arch"
}

# Get latest release version
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep -o '"tag_name": "[^"]*"' | cut -d'"' -f4)
    if [ -z "$version" ]; then
        log_error "No se pudo obtener la última versión desde GitHub"
        exit 1
    fi
    echo "$version"
}

# Download file with progress
download_file() {
    local url="$1"
    local output="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$output" "$url" --progress-bar
    elif command -v wget >/dev/null 2>&1; then
        wget -q --show-progress -O "$output" "$url"
    else
        log_error "Se requiere curl o wget para descargar"
        exit 1
    fi
}

# Verify checksum
verify_checksum() {
    local binary="$1"
    local checksums_file="$2"
    local expected_binary="$3"

    if [ ! -f "$checksums_file" ]; then
        log_warn "No se encontró archivo de checksums, saltando verificación"
        return 0
    fi

    log_step "Verificando checksum..."

    local expected_hash
    expected_hash=$(grep "$expected_binary" "$checksums_file" 2>/dev/null | awk '{print $1}')

    if [ -z "$expected_hash" ]; then
        log_warn "No se encontró checksum para $expected_binary"
        return 0
    fi

    local actual_hash
    if command -v sha256sum >/dev/null 2>&1; then
        actual_hash=$(sha256sum "$binary" | awk '{print $1}')
    elif command -v shasum >/dev/null 2>&1; then
        actual_hash=$(shasum -a 256 "$binary" | awk '{print $1}')
    else
        log_warn "No se encontró sha256sum o shasum, saltando verificación"
        return 0
    fi

    if [ "$expected_hash" = "$actual_hash" ]; then
        log_info "Checksum verificado correctamente"
        return 0
    else
        log_error "Checksum no coincide!"
        log_error "  Esperado: $expected_hash"
        log_error "  Actual:   $actual_hash"
        return 1
    fi
}

# Main installation function
main() {
    echo ""
    log_info "Backup Installer - Script de Instalación"
    echo ""

    # Detect platform
    log_step "Detectando plataforma..."
    OS=$(detect_os)
    ARCH=$(detect_arch)
    log_info "OS: $OS, Arquitectura: $ARCH"

    # Get latest version
    log_step "Obteniendo última versión..."
    VERSION=$(get_latest_version)
    log_info "Versión: $VERSION"

    # Construct binary name
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="backup-installer-${OS}-${ARCH}.exe"
    else
        BINARY_NAME="backup-installer-${OS}-${ARCH}"
    fi

    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TMP_DIR"' EXIT

    # Download binary
    log_step "Descargando $BINARY_NAME..."
    BINARY_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"
    BINARY_PATH="$TMP_DIR/backup-installer"

    if ! download_file "$BINARY_URL" "$BINARY_PATH"; then
        log_error "Error descargando el binario"
        exit 1
    fi
    log_info "Descarga completada"

    # Download checksums
    log_step "Descargando checksums..."
    CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
    CHECKSUMS_PATH="$TMP_DIR/checksums.txt"

    if ! download_file "$CHECKSUMS_URL" "$CHECKSUMS_PATH" 2>/dev/null; then
        log_warn "No se pudo descargar checksums.txt"
        CHECKSUMS_PATH=""
    fi

    # Verify checksum
    if [ -n "$CHECKSUMS_PATH" ]; then
        if ! verify_checksum "$BINARY_PATH" "$CHECKSUMS_PATH" "$BINARY_NAME"; then
            exit 1
        fi
    fi

    # Make binary executable
    chmod +x "$BINARY_PATH"

    # Determine install location
    log_step "Instalando..."
    if [ -w "$INSTALL_DIR" ] || [ "$EUID" -eq 0 ]; then
        TARGET_DIR="$INSTALL_DIR"
        TARGET_NAME="backup-installer"
    else
        TARGET_DIR="$FALLBACK_DIR"
        TARGET_NAME="backup-installer"
        log_warn "No hay permisos de escritura en $INSTALL_DIR"
        log_info "Instalando en $TARGET_DIR"
        mkdir -p "$TARGET_DIR"
    fi

    TARGET_PATH="$TARGET_DIR/$TARGET_NAME"

    # Install binary
    if [ -w "$TARGET_DIR" ]; then
        mv "$BINARY_PATH" "$TARGET_PATH"
    else
        log_info "Requiere sudo para instalar en $TARGET_DIR"
        sudo mv "$BINARY_PATH" "$TARGET_PATH"
    fi

    log_info "Instalado en: $TARGET_PATH"

    # Check if target dir is in PATH
    if [[ ":$PATH:" != *":$TARGET_DIR:"* ]]; then
        log_warn "$TARGET_DIR no está en el PATH"
        echo ""
        echo "Agrega esto a tu ~/.bashrc o ~/.zshrc:"
        echo "  export PATH=\"$TARGET_DIR:\$PATH\""
        echo ""
    fi

    # Success message
    echo ""
    log_info "✅ Instalación completada exitosamente!"
    echo ""
    echo "Para ejecutar el instalador:"
    if [ "$TARGET_DIR" = "$INSTALL_DIR" ]; then
        echo "  backup-installer"
    else
        echo "  $TARGET_PATH"
        echo "  # o después de agregar $TARGET_DIR al PATH:"
        echo "  backup-installer"
    fi
    echo ""
}

# Run main function
main "$@"
