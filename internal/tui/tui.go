package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wdelcant/backup-installer/internal/config"
	"github.com/wdelcant/backup-installer/internal/crypto"
)

// Version is the application version, set by main
var Version = "dev"

// StartWizard initializes and runs the TUI wizard
func StartWizard(isFirstRun bool, baseDir string, configManager *config.Manager, keyManager *crypto.MasterKeyManager, encryptor *crypto.Encryptor) error {
	// Initialize model with NewModel constructor
	model := NewModel(configManager, keyManager, encryptor)
	model.isFirstRun = isFirstRun
	model.baseDir = baseDir

	// Run the TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
