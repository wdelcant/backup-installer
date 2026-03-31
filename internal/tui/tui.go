package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/invitsm/invitsm-backup-installer/internal/config"
	"github.com/invitsm/invitsm-backup-installer/internal/crypto"
)

// StartWizard initializes and runs the TUI wizard
func StartWizard(isFirstRun bool, baseDir string, configManager *config.Manager, keyManager *crypto.MasterKeyManager, encryptor *crypto.Encryptor) error {
	// Initialize model with default config
	model := Model{
		step:          StepWelcome,
		isFirstRun:    isFirstRun,
		baseDir:       baseDir,
		configManager: configManager,
		keyManager:    keyManager,
		encryptor:     encryptor,
		config:        config.DefaultConfig(),

		// Default form values
		dbSourcePort:   "5432",
		dbTargetPort:   "5432",
		cronExpression: "0 2 * * *",
		timezone:       "America/Santiago",
		retentionDays:  7,
		backupPath:     "/opt/invitsm/backups",
		webhookEnabled: false,
		restoreDelay:   30,
	}

	// Run the TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
