// Package tui provides the Bubble Tea TUI for the backup installer wizard
package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/invitsm/invitsm-backup-installer/internal/config"
	"github.com/invitsm/invitsm-backup-installer/internal/crypto"
)

// Step represents a wizard step
type Step int

const (
	StepWelcome Step = iota
	StepSecurity
	StepDBSource
	StepMode
	StepDBTarget
	StepSchedule
	StepRetention
	StepWebhook
	StepSummary
	StepInstalling
	StepSuccess
)

// String returns the string representation of a step
func (s Step) String() string {
	return [...]string{
		"welcome",
		"security",
		"db_source",
		"mode",
		"db_target",
		"schedule",
		"retention",
		"webhook",
		"summary",
		"installing",
		"success",
	}[s]
}

// Model represents the TUI application state
type Model struct {
	// State
	step       Step
	isFirstRun bool
	baseDir    string
	width      int
	height     int
	quitting   bool
	showHelp   bool

	// Dependencies
	configManager *config.Manager
	keyManager    *crypto.MasterKeyManager
	encryptor     *crypto.Encryptor

	// Configuration being built
	config *config.Config

	// UI state
	cursor          int
	errorMsg        string
	successMsg      string
	installProgress float64
	installStatus   string

	// Form fields (will be populated in each step's view)
	dbSourceHost     string
	dbSourcePort     string
	dbSourceDatabase string
	dbSourceUsername string
	dbSourcePassword string

	dbTargetEnabled  bool
	dbTargetHost     string
	dbTargetPort     string
	dbTargetDatabase string
	dbTargetUsername string
	dbTargetPassword string
	restoreDelay     int

	cronExpression string
	timezone       string

	retentionDays int
	backupPath    string

	webhookEnabled bool
	webhookURL     string
	webhookToken   string
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case progressMsg:
		m.installProgress = msg.progress
		m.installStatus = msg.status
		return m, nil

	case installCompleteMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			m.step = StepSuccess
		}
		return m, nil
	}

	return m, nil
}

// View renders the current step
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	switch m.step {
	case StepWelcome:
		return m.viewWelcome()
	case StepSecurity:
		return m.viewSecurity()
	case StepDBSource:
		return m.viewDBSource()
	case StepMode:
		return m.viewMode()
	case StepDBTarget:
		return m.viewDBTarget()
	case StepSchedule:
		return m.viewSchedule()
	case StepRetention:
		return m.viewRetention()
	case StepWebhook:
		return m.viewWebhook()
	case StepSummary:
		return m.viewSummary()
	case StepInstalling:
		return m.viewInstalling()
	case StepSuccess:
		return m.viewSuccess()
	default:
		return "Unknown step\n"
	}
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		if m.step > StepWelcome {
			m.step--
		}
		return m, nil
	}

	// Step-specific handling
	switch m.step {
	case StepWelcome:
		switch msg.String() {
		case "enter":
			m.step = StepSecurity
		}

	case StepSecurity:
		switch msg.String() {
		case "enter":
			m.step = StepDBSource
		}

	case StepDBSource:
		// Handle form navigation
		switch msg.String() {
		case "enter":
			// Validate and move to next step
			m.step = StepMode
		case "tab", "down":
			m.cursor++
		case "shift+tab", "up":
			m.cursor--
		}

	case StepMode:
		switch msg.String() {
		case "enter":
			if m.config.Target.Enabled {
				m.step = StepDBTarget
			} else {
				m.step = StepSchedule
			}
		case "up", "down":
			m.cursor++
		}

	case StepDBTarget:
		switch msg.String() {
		case "enter":
			m.step = StepSchedule
		}

	case StepSchedule:
		switch msg.String() {
		case "enter":
			m.step = StepRetention
		}

	case StepRetention:
		switch msg.String() {
		case "enter":
			m.step = StepWebhook
		}

	case StepWebhook:
		switch msg.String() {
		case "enter":
			m.step = StepSummary
		case "s":
			// Skip webhook
			m.step = StepSummary
		}

	case StepSummary:
		switch msg.String() {
		case "enter":
			m.step = StepInstalling
			return m, m.runInstallation()
		case "esc":
			m.step = StepWebhook
		}

	case StepSuccess:
		switch msg.String() {
		case "enter":
			m.quitting = true
			return m, tea.Quit
		case "r":
			// Run backup now
			m.step = StepInstalling
			return m, m.runInstallation()
		}
	}

	return m, nil
}

// runInstallation starts the installation process
func (m Model) runInstallation() tea.Cmd {
	return func() tea.Msg {
		// Simulate installation progress
		progress := 0.0
		for progress < 100 {
			progress += 10
			// Send progress updates
		}
		return installCompleteMsg{err: nil}
	}
}

// progressMsg is sent during installation
type progressMsg struct {
	progress float64
	status   string
}

// installCompleteMsg is sent when installation finishes
type installCompleteMsg struct {
	err error
}
