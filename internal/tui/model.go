// Package tui provides the Bubble Tea TUI for the backup installer wizard
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wdelcant/backup-installer/internal/config"
	"github.com/wdelcant/backup-installer/internal/crypto"
	"github.com/wdelcant/backup-installer/internal/pipeline"
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

	// Form inputs - Source DB
	sourceInputs []textinput.Model

	// Form inputs - Target DB
	targetInputs []textinput.Model

	// Form inputs - Schedule
	scheduleInputs []textinput.Model

	// Form inputs - Retention
	retentionInputs []textinput.Model

	// Form inputs - Webhook
	webhookInputs []textinput.Model

	// Mode selection
	selectedMode int

	// Webhook enabled
	webhookEnabled bool
}

// NewModel creates a new TUI model
func NewModel(configManager *config.Manager, keyManager *crypto.MasterKeyManager, encryptor *crypto.Encryptor) Model {
	m := Model{
		step:            StepWelcome,
		isFirstRun:      true,
		configManager:   configManager,
		keyManager:      keyManager,
		encryptor:       encryptor,
		config:          &config.Config{},
		installProgress: 0,
		installStatus:   "",
	}

	// Initialize source DB inputs
	m.sourceInputs = make([]textinput.Model, 5)
	for i := range m.sourceInputs {
		t := textinput.New()
		t.CharLimit = 100
		t.Width = 40

		switch i {
		case 0: // Host
			t.Placeholder = "localhost"
			t.Focus()
		case 1: // Port
			t.Placeholder = "5432"
			t.SetValue("5432")
		case 2: // Database
			t.Placeholder = "myapp_production"
		case 3: // Username
			t.Placeholder = "postgres"
		case 4: // Password
			t.Placeholder = "********"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		m.sourceInputs[i] = t
	}

	// Initialize target DB inputs
	m.targetInputs = make([]textinput.Model, 5)
	for i := range m.targetInputs {
		t := textinput.New()
		t.CharLimit = 100
		t.Width = 40

		switch i {
		case 0: // Host
			t.Placeholder = "localhost"
			t.Focus()
		case 1: // Port
			t.Placeholder = "5432"
			t.SetValue("5432")
		case 2: // Database
			t.Placeholder = "myapp_qa"
		case 3: // Username
			t.Placeholder = "postgres"
		case 4: // Password
			t.Placeholder = "********"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		m.targetInputs[i] = t
	}

	// Initialize schedule inputs
	m.scheduleInputs = make([]textinput.Model, 2)
	for i := range m.scheduleInputs {
		t := textinput.New()
		t.CharLimit = 50
		t.Width = 40

		switch i {
		case 0: // Cron expression
			t.Placeholder = "0 2 * * *"
			t.SetValue("0 2 * * *")
			t.Focus()
		case 1: // Timezone
			t.Placeholder = "America/Santiago"
			t.SetValue("America/Santiago")
		}
		m.scheduleInputs[i] = t
	}

	// Initialize retention inputs (GFS policy)
	m.retentionInputs = make([]textinput.Model, 4)
	for i := range m.retentionInputs {
		t := textinput.New()
		t.CharLimit = 50
		t.Width = 40

		switch i {
		case 0: // Backup path
			t.Placeholder = "./backups"
			t.SetValue("./backups")
			t.Focus()
		case 1: // Son (daily backups)
			t.Placeholder = "7"
			t.SetValue("7")
		case 2: // Father (weekly backups)
			t.Placeholder = "4"
			t.SetValue("4")
		case 3: // Grandfather (monthly backups)
			t.Placeholder = "12"
			t.SetValue("12")
		}
		m.retentionInputs[i] = t
	}

	// Initialize webhook inputs
	m.webhookInputs = make([]textinput.Model, 2)
	for i := range m.webhookInputs {
		t := textinput.New()
		t.CharLimit = 200
		t.Width = 40

		switch i {
		case 0: // URL
			t.Placeholder = "https://n8n.example.com/webhook/..."
			t.Focus()
		case 1: // Token
			t.Placeholder = "optional-token"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		m.webhookInputs[i] = t
	}

	return m
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// First, let the focused input handle the key
		// This allows typing in the inputs
		cmd := m.updateFocusedInput(msg)

		// Then handle special keys for navigation
		if msg.String() == "tab" || msg.String() == "shift+tab" ||
			msg.String() == "enter" || msg.String() == "esc" ||
			msg.String() == "up" || msg.String() == "down" {
			return m.handleKeyPress(msg)
		}

		return m, cmd

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

	// Update focused inputs based on current step
	switch m.step {
	case StepDBSource:
		for i := range m.sourceInputs {
			var cmd tea.Cmd
			m.sourceInputs[i], cmd = m.sourceInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	case StepDBTarget:
		for i := range m.targetInputs {
			var cmd tea.Cmd
			m.targetInputs[i], cmd = m.targetInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	case StepSchedule:
		for i := range m.scheduleInputs {
			var cmd tea.Cmd
			m.scheduleInputs[i], cmd = m.scheduleInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	case StepRetention:
		for i := range m.retentionInputs {
			var cmd tea.Cmd
			m.retentionInputs[i], cmd = m.retentionInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	case StepWebhook:
		for i := range m.webhookInputs {
			var cmd tea.Cmd
			m.webhookInputs[i], cmd = m.webhookInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
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

// focusInput focuses the input at the given index for the current step
func (m *Model) focusInput(index int) {
	// Blur all inputs first
	switch m.step {
	case StepDBSource:
		for i := range m.sourceInputs {
			if i == index {
				m.sourceInputs[i].Focus()
			} else {
				m.sourceInputs[i].Blur()
			}
		}
	case StepDBTarget:
		for i := range m.targetInputs {
			if i == index {
				m.targetInputs[i].Focus()
			} else {
				m.targetInputs[i].Blur()
			}
		}
	case StepSchedule:
		for i := range m.scheduleInputs {
			if i == index {
				m.scheduleInputs[i].Focus()
			} else {
				m.scheduleInputs[i].Blur()
			}
		}
	case StepRetention:
		for i := range m.retentionInputs {
			if i == index {
				m.retentionInputs[i].Focus()
			} else {
				m.retentionInputs[i].Blur()
			}
		}
	case StepWebhook:
		for i := range m.webhookInputs {
			if i == index {
				m.webhookInputs[i].Focus()
			} else {
				m.webhookInputs[i].Blur()
			}
		}
	}
}

// updateFocusedInput passes the key message to the currently focused input
func (m *Model) updateFocusedInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch m.step {
	case StepDBSource:
		if m.cursor >= 0 && m.cursor < len(m.sourceInputs) {
			m.sourceInputs[m.cursor], cmd = m.sourceInputs[m.cursor].Update(msg)
		}
	case StepDBTarget:
		if m.cursor >= 0 && m.cursor < len(m.targetInputs) {
			m.targetInputs[m.cursor], cmd = m.targetInputs[m.cursor].Update(msg)
		}
	case StepSchedule:
		if m.cursor >= 0 && m.cursor < len(m.scheduleInputs) {
			m.scheduleInputs[m.cursor], cmd = m.scheduleInputs[m.cursor].Update(msg)
		}
	case StepRetention:
		if m.cursor >= 0 && m.cursor < len(m.retentionInputs) {
			m.retentionInputs[m.cursor], cmd = m.retentionInputs[m.cursor].Update(msg)
		}
	case StepWebhook:
		if m.cursor >= 0 && m.cursor < len(m.webhookInputs) {
			m.webhookInputs[m.cursor], cmd = m.webhookInputs[m.cursor].Update(msg)
		}
	}

	return cmd
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
			m.cursor = 0
			m.focusInput(0)
		} else if m.step == StepWelcome {
			// Exit from welcome screen
			m.quitting = true
			return m, tea.Quit
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
			m.focusInput(0)
		}

	case StepDBSource:
		// Handle form navigation
		switch msg.String() {
		case "enter":
			// Validate and move to next step
			m.config.Source.Host = m.sourceInputs[0].Value()
			m.config.Source.Port = parsePort(m.sourceInputs[1].Value())
			m.config.Source.Database = m.sourceInputs[2].Value()
			m.config.Source.Username = m.sourceInputs[3].Value()
			m.config.Source.Password = m.sourceInputs[4].Value()
			// Ir a la pregunta de si quiere configurar destino
			m.step = StepMode
			m.selectedMode = 0 // Por defecto NO
		case "tab", "down":
			if m.cursor < len(m.sourceInputs)-1 {
				m.cursor++
				m.focusInput(m.cursor)
			}
		case "shift+tab", "up":
			if m.cursor > 0 {
				m.cursor--
				m.focusInput(m.cursor)
			}
		}
		return m, textinput.Blink

	case StepMode:
		switch msg.String() {
		case "enter":
			// selectedMode: 0 = NO, 1 = SI
			if m.selectedMode == 1 {
				m.config.Target.Enabled = true
				m.step = StepDBTarget
				m.focusInput(0)
			} else {
				m.config.Target.Enabled = false
				m.step = StepSchedule
				m.focusInput(0)
			}
		case "up", "down", "tab":
			// Toggle entre SI y NO
			if m.selectedMode == 0 {
				m.selectedMode = 1
			} else {
				m.selectedMode = 0
			}
		}

	case StepDBTarget:
		switch msg.String() {
		case "enter":
			m.config.Target.Host = m.targetInputs[0].Value()
			m.config.Target.Port = parsePort(m.targetInputs[1].Value())
			m.config.Target.Database = m.targetInputs[2].Value()
			m.config.Target.Username = m.targetInputs[3].Value()
			m.config.Target.Password = m.targetInputs[4].Value()
			m.step = StepSchedule
			m.focusInput(0)
		case "tab", "down":
			if m.cursor < len(m.targetInputs)-1 {
				m.cursor++
				m.focusInput(m.cursor)
			}
		case "shift+tab", "up":
			if m.cursor > 0 {
				m.cursor--
				m.focusInput(m.cursor)
			}
		}
		return m, textinput.Blink

	case StepSchedule:
		switch msg.String() {
		case "enter":
			m.config.Schedule.CronExpression = m.scheduleInputs[0].Value()
			m.config.Schedule.Timezone = m.scheduleInputs[1].Value()
			m.step = StepRetention
			m.focusInput(0)
		case "tab", "down":
			if m.cursor < len(m.scheduleInputs)-1 {
				m.cursor++
				m.focusInput(m.cursor)
			}
		case "shift+tab", "up":
			if m.cursor > 0 {
				m.cursor--
				m.focusInput(m.cursor)
			}
		}
		return m, textinput.Blink

	case StepRetention:
		switch msg.String() {
		case "enter":
			m.config.Storage.LocalPath = m.retentionInputs[0].Value()
			m.config.Storage.Retention.Enabled = true
			m.config.Storage.Retention.Son = parseInt(m.retentionInputs[1].Value())
			m.config.Storage.Retention.Father = parseInt(m.retentionInputs[2].Value())
			m.config.Storage.Retention.Grandfather = parseInt(m.retentionInputs[3].Value())
			m.step = StepWebhook
			m.focusInput(0)
		case "tab", "down":
			if m.cursor < len(m.retentionInputs)-1 {
				m.cursor++
				m.focusInput(m.cursor)
			}
		case "shift+tab", "up":
			if m.cursor > 0 {
				m.cursor--
				m.focusInput(m.cursor)
			}
		}
		return m, textinput.Blink

	case StepWebhook:
		switch msg.String() {
		case "enter":
			m.config.Webhook.URL = m.webhookInputs[0].Value()
			if m.webhookInputs[1].Value() != "" {
				m.config.Webhook.Headers["Authorization"] = m.webhookInputs[1].Value()
			}
			m.config.Webhook.Enabled = m.webhookEnabled && m.config.Webhook.URL != ""
			m.step = StepSummary
		case "s":
			// Skip webhook
			m.webhookEnabled = false
			m.step = StepSummary
		case "tab", "down":
			if m.cursor < len(m.webhookInputs)-1 {
				m.cursor++
				m.focusInput(m.cursor)
			}
		case "shift+tab", "up":
			if m.cursor > 0 {
				m.cursor--
				m.focusInput(m.cursor)
			}
		}
		return m, textinput.Blink

	case StepSummary:
		switch msg.String() {
		case "enter":
			m.step = StepInstalling
			return m, m.runInstallation()
		case "esc":
			m.step = StepWebhook
			m.focusInput(0)
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
		// Step 1: Save configuration
		if err := m.configManager.Save(m.config); err != nil {
			return installCompleteMsg{err: fmt.Errorf("failed to save config: %w", err)}
		}

		// Step 2: Generate pipeline script (CRITICAL - must succeed)
		if err := m.generatePipeline(); err != nil {
			return installCompleteMsg{err: fmt.Errorf("failed to generate pipeline: %w", err)}
		}

		// Step 3: Install crontab (OPTIONAL - don't fail entire installation)
		if err := m.installCrontab(); err != nil {
			// Log warning but continue - user can install crontab manually later
			// This prevents hanging when crontab command is not available or blocked
			fmt.Fprintf(os.Stderr, "⚠️  Warning: crontab installation failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "   You can install crontab manually later.\n")
		}

		return installCompleteMsg{err: nil}
	}
}

// generatePipeline generates the backup pipeline script
func (m Model) generatePipeline() error {
	scriptsDir := filepath.Join(m.baseDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return err
	}

	scriptPath := filepath.Join(scriptsDir, "pipeline.sh")

	// Generate script content using pipeline generator
	generator := pipeline.NewGenerator(m.baseDir)
	scriptContent, err := generator.Generate(m.config)
	if err != nil {
		return err
	}

	// Write script to file
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		return err
	}

	return nil
}

// installCrontab installs the backup cronjob
func (m Model) installCrontab() error {
	scriptPath := filepath.Join(m.baseDir, "scripts", "pipeline.sh")
	logsDir := filepath.Join(m.baseDir, "logs")

	// Ensure logs directory exists
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// Create cronjob line
	cronJob := fmt.Sprintf("%s %s >> %s/cron.log 2>&1",
		m.config.Schedule.CronExpression,
		scriptPath,
		logsDir)

	// Try to install crontab
	// First, try to get existing crontab
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()

	var newCrontab string
	if err != nil {
		// No existing crontab, create new one
		newCrontab = cronJob
	} else {
		// Has existing crontab, append our job
		existing := strings.TrimSpace(string(output))
		// Remove any existing backup-installer cronjobs
		lines := strings.Split(existing, "\n")
		var filteredLines []string
		for _, line := range lines {
			if !strings.Contains(line, "pipeline.sh") && !strings.Contains(line, "backup-installer") {
				filteredLines = append(filteredLines, line)
			}
		}
		filteredLines = append(filteredLines, cronJob)
		newCrontab = strings.Join(filteredLines, "\n")
	}

	// Install the crontab
	cmd = exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(newCrontab + "\n")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install crontab: %w", err)
	}

	return nil
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
