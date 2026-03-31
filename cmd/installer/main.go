package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/invitsm/invitsm-backup-installer/internal/config"
	"github.com/invitsm/invitsm-backup-installer/internal/crypto"
	"github.com/invitsm/invitsm-backup-installer/internal/logo"
	"github.com/invitsm/invitsm-backup-installer/internal/tui"
)

const (
	appName    = "invitsm-backup-installer"
	appVersion = "1.0.0"
)

func main() {
	// Get base directory (where the binary is located)
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
		os.Exit(1)
	}

	baseDir := filepath.Dir(execPath)

	// For development, use current directory if binary is in /tmp or similar
	if filepath.Base(baseDir) == "tmp" || filepath.Base(baseDir) == "go-build" {
		cwd, _ := os.Getwd()
		baseDir = cwd
	}

	// Initialize master key manager
	configDir := filepath.Join(baseDir, "config")
	keyManager := crypto.NewMasterKeyManager(configDir)

	// Check if this is first run (no master key)
	isFirstRun := !keyManager.KeyExists()

	// Get or create master key
	masterKey, err := keyManager.GetOrCreateMasterKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing master key: %v\n", err)
		os.Exit(1)
	}

	// Initialize encryptor
	encryptor := crypto.NewEncryptor(masterKey)

	// Initialize config manager
	configManager := config.NewManager(baseDir, encryptor)

	// Show logo
	fmt.Println(logo.Rendered(appVersion))
	fmt.Println()

	// Start TUI
	if err := tui.StartWizard(isFirstRun, baseDir, configManager, keyManager, encryptor); err != nil {
		fmt.Fprintf(os.Stderr, "\nError running installer: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Installation completed successfully!")
}
