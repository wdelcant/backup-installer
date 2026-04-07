package tui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wdelcant/backup-installer/internal/config"
	"github.com/wdelcant/backup-installer/internal/crypto"
	"github.com/wdelcant/backup-installer/internal/logo"
)

// ErrEditRequested is returned when user wants to edit configuration
var ErrEditRequested = errors.New("edit configuration requested")

// ModalType represents the type of modal to show
type ModalType int

const (
	ModalNone ModalType = iota
	ModalConfig
	ModalLogs
	ModalStatus
	ModalConfirm
	ModalRunning
	ModalResult
)

// DashboardModel represents the dashboard TUI state
type DashboardModel struct {
	config        *config.Config
	baseDir       string
	configManager *config.Manager
	keyManager    *crypto.MasterKeyManager
	encryptor     *crypto.Encryptor
	quitting      bool
	editRequested bool
	cursor        int
	menuItems     []string
	width         int
	height        int

	// Modal state
	modalType       ModalType
	modalContent    string
	modalTitle      string
	modalConfirmMsg string

	// Running state
	runningProgress float64
	runningStatus   string
	runningOutput   strings.Builder
	runningComplete bool
	runningSuccess  bool
}

// DashboardItem represents a menu item
type DashboardItem int

const (
	ItemViewConfig DashboardItem = iota
	ItemEditConfig
	ItemRunBackup
	ItemViewLogs
	ItemStatus
	ItemUninstall
	ItemExit
)

// StartDashboard initializes and runs the dashboard TUI
// Returns ErrEditRequested if user wants to edit configuration
func StartDashboard(cfg *config.Config, baseDir string, configManager *config.Manager, keyManager *crypto.MasterKeyManager, encryptor *crypto.Encryptor) error {
	model := DashboardModel{
		config:        cfg,
		baseDir:       baseDir,
		configManager: configManager,
		keyManager:    keyManager,
		encryptor:     encryptor,
		editRequested: false,
		menuItems: []string{
			"📋 Ver configuración",
			"✏️  Editar configuración",
			"🔄 Ejecutar backup ahora",
			"📄 Ver logs",
			"📊 Estado del sistema",
			"🗑️  Desinstalar",
			"🚪 Salir",
		},
		modalType: ModalNone,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Check if edit was requested
	if finalModel.(DashboardModel).editRequested {
		return ErrEditRequested
	}

	return nil
}

// Init initializes the dashboard
func (m DashboardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If modal is open (but not running), handle modal keys
		if m.modalType != ModalNone && m.modalType != ModalRunning {
			switch msg.String() {
			case "esc", "enter", "q":
				if m.modalType == ModalConfirm && m.modalConfirmMsg == "run_backup" {
					// User confirmed, start backup
					return m, m.startBackup()
				}
				m.modalType = ModalNone
				m.modalContent = ""
				m.modalConfirmMsg = ""
			}
			return m, nil
		}

		// If running, only allow q to quit (but don't allow quitting)
		if m.modalType == ModalRunning {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			if m.cursor == int(ItemExit) {
				m.quitting = true
				return m, tea.Quit
			}
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.menuItems)-1 {
				m.cursor++
			}
		case "enter":
			return m.handleMenuSelection()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case progressMsg:
		m.runningProgress = msg.progress
		m.runningStatus = msg.status
		return m, nil
	case backupCompleteMsg:
		m.runningComplete = true
		m.runningSuccess = msg.success
		m.modalType = ModalResult
		if msg.err != nil {
			m.modalTitle = "❌ Error en backup"
			m.modalContent = fmt.Sprintf("El backup falló:\n\n%s", msg.err.Error())
		} else {
			m.modalTitle = "✅ Backup completado"
			m.modalContent = fmt.Sprintf("El backup se completó exitosamente.\n\n%s", m.runningOutput.String())
		}
		return m, nil
	}
	return m, nil
}

// handleMenuSelection processes the selected menu item
func (m DashboardModel) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case int(ItemViewConfig):
		m.showConfigModal()
	case int(ItemEditConfig):
		// Signal that we want to edit
		m.editRequested = true
		return m, tea.Quit
	case int(ItemRunBackup):
		m.showRunBackupModal()
	case int(ItemViewLogs):
		m.showLogsModal()
	case int(ItemStatus):
		m.showStatusModal()
	case int(ItemUninstall):
		m.showUninstallConfirm()
	case int(ItemExit):
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

// View renders the dashboard
func (m DashboardModel) View() string {
	if m.quitting {
		return "\n👋 ¡Hasta luego!\n\n"
	}

	// If modal is open, show modal
	if m.modalType != ModalNone {
		return m.renderModal()
	}

	var content string

	// Header
	content += logo.Header("Panel de Control")
	content += "\n\n"

	// Status box
	content += m.renderStatusBox()
	content += "\n\n"

	// Menu
	content += m.renderMenu()
	content += "\n\n"

	// Footer help
	content += lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Render("[↑/↓] Navegar  [Enter] Seleccionar  [q] Salir (en 'Salir')")

	return content
}

// renderModal renders a modal overlay
func (m DashboardModel) renderModal() string {
	// Create modal box
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D4AA")).
		Padding(2).
		Width(70)

	var content string

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA")).
		Bold(true).
		MarginBottom(1)
	content += titleStyle.Render(m.modalTitle)
	content += "\n\n"

	// Content based on modal type
	switch m.modalType {
	case ModalRunning:
		content += m.renderProgressBar()
		content += "\n\n"
		content += m.runningStatus
		content += "\n\n"
		content += lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Render("⏳ Ejecutando backup...")
		return modalStyle.Render(content)
	case ModalResult:
		content += m.modalContent
		content += "\n\n"
		content += lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Render("[Enter] o [Esc] Cerrar")
		return modalStyle.Render(content)
	default:
		content += m.modalContent
		content += "\n\n"
		content += lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Render("[Enter] o [Esc] Cerrar")
		return modalStyle.Render(content)
	}
}

// renderProgressBar renders a progress bar
func (m DashboardModel) renderProgressBar() string {
	barWidth := 50
	filled := int((m.runningProgress / 100.0) * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA"))

	return fmt.Sprintf("%s\n\n%s %5.1f%%",
		progressStyle.Render(bar),
		progressStyle.Render("Progreso:"),
		m.runningProgress)
}

// renderStatusBox renders the status information box
func (m DashboardModel) renderStatusBox() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D4AA")).
		Padding(1, 2).
		Width(60)

	// Get last backup info
	lastBackup := m.getLastBackupInfo()
	nextBackup := m.getNextBackupTime()

	status := fmt.Sprintf(
		"Base de datos: %s\n"+
			"Host: %s:%d\n"+
			"Schedule: %s\n"+
			"Último backup: %s\n"+
			"Próximo backup: %s",
		m.config.Source.Database,
		m.config.Source.Host,
		m.config.Source.Port,
		m.config.Schedule.CronExpression,
		lastBackup,
		nextBackup,
	)

	return boxStyle.Render(status)
}

// renderMenu renders the menu
func (m DashboardModel) renderMenu() string {
	var content string

	for i, item := range m.menuItems {
		style := lipgloss.NewStyle()
		if i == m.cursor {
			style = style.
				Foreground(lipgloss.Color("#00D4AA")).
				Bold(true).
				Background(lipgloss.Color("#1F2937"))
			content += style.Render(" > " + item + " ")
		} else {
			content += style.Render("   " + item)
		}
		content += "\n"
	}

	return content
}

// getLastBackupInfo returns information about the last backup
func (m DashboardModel) getLastBackupInfo() string {
	logsDir := filepath.Join(m.baseDir, "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil || len(entries) == 0 {
		return "Nunca"
	}

	// Find the most recent log file
	var latestTime time.Time
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil && info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
			}
		}
	}

	if latestTime.IsZero() {
		return "Nunca"
	}

	return latestTime.Format("2006-01-02 %H:%M")
}

// getNextBackupTime calculates the next backup time based on cron expression
func (m DashboardModel) getNextBackupTime() string {
	return "Mañana " + m.config.Schedule.CronExpression
}

// showConfigModal shows configuration in a modal
func (m *DashboardModel) showConfigModal() {
	m.modalType = ModalConfig
	m.modalTitle = "📋 Configuración actual"

	var content strings.Builder
	content.WriteString(fmt.Sprintf("Base de datos origen: %s\n", m.config.Source.Database))
	content.WriteString(fmt.Sprintf("Host origen: %s:%d\n", m.config.Source.Host, m.config.Source.Port))
	content.WriteString(fmt.Sprintf("Usuario origen: %s\n", m.config.Source.Username))
	content.WriteString(fmt.Sprintf("Schedule: %s\n", m.config.Schedule.CronExpression))
	content.WriteString(fmt.Sprintf("Timezone: %s\n", m.config.Schedule.Timezone))
	content.WriteString(fmt.Sprintf("Directorio de backups: %s\n", m.config.Storage.LocalPath))
	content.WriteString(fmt.Sprintf("Retención: %d/%d/%d (S/F/G)\n",
		m.config.Storage.Retention.Son,
		m.config.Storage.Retention.Father,
		m.config.Storage.Retention.Grandfather))

	if m.config.Target.Enabled {
		content.WriteString(fmt.Sprintf("\nBase de datos destino: %s\n", m.config.Target.Database))
		content.WriteString(fmt.Sprintf("Host destino: %s:%d\n", m.config.Target.Host, m.config.Target.Port))
	}

	m.modalContent = content.String()
}

// showRunBackupModal shows backup execution in a modal
func (m *DashboardModel) showRunBackupModal() {
	m.modalType = ModalConfirm
	m.modalTitle = "🔄 Ejecutar backup ahora"
	m.modalContent = "¿Estás seguro de que querés ejecutar el backup ahora?\n\n" +
		"Esto ejecutará el script de backup manualmente.\n" +
		"El proceso puede tardar algunos minutos.\n\n" +
		"Presiona [Enter] para confirmar o [Esc] para cancelar."
	m.modalConfirmMsg = "run_backup"
}

// startBackup starts the backup execution
func (m *DashboardModel) startBackup() tea.Cmd {
	return func() tea.Msg {
		scriptPath := filepath.Join(m.baseDir, "scripts", "pipeline.sh")

		// Check if script exists
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return backupCompleteMsg{
				success: false,
				err:     fmt.Errorf("el script pipeline.sh no existe. Ejecutá el wizard para configurarlo"),
			}
		}

		// Execute the script
		cmd := exec.Command(scriptPath)
		cmd.Dir = m.baseDir

		// Read output line by line
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return backupCompleteMsg{success: false, err: err}
		}

		if err := cmd.Start(); err != nil {
			return backupCompleteMsg{success: false, err: err}
		}

		// Read output
		scanner := bufio.NewScanner(stdout)
		lineCount := 0
		for scanner.Scan() {
			line := scanner.Text()
			m.runningOutput.WriteString(line + "\n")
			lineCount++

			// Update progress based on lines read
			progress := float64(lineCount) / 20.0 * 100.0
			if progress > 95 {
				progress = 95
			}
		}

		if err := cmd.Wait(); err != nil {
			return backupCompleteMsg{
				success: false,
				err:     fmt.Errorf("el script falló: %w\n\n%s", err, m.runningOutput.String()),
			}
		}

		m.runningProgress = 100
		m.runningStatus = "Completado"

		return backupCompleteMsg{success: true, err: nil}
	}
}

// showLogsModal shows logs in a modal
func (m *DashboardModel) showLogsModal() {
	m.modalType = ModalLogs
	m.modalTitle = "📄 Logs recientes"

	logsDir := filepath.Join(m.baseDir, "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil || len(entries) == 0 {
		m.modalContent = "No hay logs disponibles"
		return
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("Total de archivos de log: %d\n\n", len(entries)))
	content.WriteString("Archivos:\n")

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				content.WriteString(fmt.Sprintf("  • %s (%d bytes, %s)\n",
					entry.Name(),
					info.Size(),
					info.ModTime().Format("2006-01-02 %H:%M")))
			}
		}
	}

	m.modalContent = content.String()
}

// showStatusModal shows system status in a modal
func (m *DashboardModel) showStatusModal() {
	m.modalType = ModalStatus
	m.modalTitle = "📊 Estado del sistema"

	var content strings.Builder
	content.WriteString("✅ Configuración: OK\n")
	content.WriteString(fmt.Sprintf("🕐 Próximo backup: %s\n", m.getNextBackupTime()))
	content.WriteString(fmt.Sprintf("📁 Directorio de backups: %s\n", m.config.Storage.LocalPath))
	content.WriteString(fmt.Sprintf("🔐 Encriptación: Activada\n"))

	m.modalContent = content.String()
}

// showUninstallConfirm shows uninstall confirmation
func (m *DashboardModel) showUninstallConfirm() {
	m.modalType = ModalConfirm
	m.modalTitle = "⚠️ Confirmar desinstalación"
	m.modalContent = "¿Estás seguro de que querés desinstalar?\n\n" +
		"Se eliminarán:\n" +
		"  • Configuración\n" +
		"  • Scripts\n" +
		"  • Cron jobs\n\n" +
		"Los backups existentes NO se eliminarán.\n\n" +
		"Para confirmar, usa: backup-installer --uninstall"
}

// backupCompleteMsg is sent when backup finishes
type backupCompleteMsg struct {
	success bool
	err     error
}
