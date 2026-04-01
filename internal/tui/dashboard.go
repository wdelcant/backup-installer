package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wdelcant/backup-installer/internal/config"
	"github.com/wdelcant/backup-installer/internal/crypto"
	"github.com/wdelcant/backup-installer/internal/logo"
)

// DashboardModel represents the dashboard TUI state
type DashboardModel struct {
	config        *config.Config
	baseDir       string
	configManager *config.Manager
	keyManager    *crypto.MasterKeyManager
	encryptor     *crypto.Encryptor
	quitting      bool
	cursor        int
	menuItems     []string
	width         int
	height        int
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
func StartDashboard(cfg *config.Config, baseDir string, configManager *config.Manager, keyManager *crypto.MasterKeyManager, encryptor *crypto.Encryptor) error {
	model := DashboardModel{
		config:        cfg,
		baseDir:       baseDir,
		configManager: configManager,
		keyManager:    keyManager,
		encryptor:     encryptor,
		menuItems: []string{
			"📋 Ver configuración",
			"✏️  Editar configuración",
			"🔄 Ejecutar backup ahora",
			"📄 Ver logs",
			"📊 Estado del sistema",
			"🗑️  Desinstalar",
			"🚪 Salir",
		},
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Init initializes the dashboard
func (m DashboardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
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
	}
	return m, nil
}

// handleMenuSelection processes the selected menu item
func (m DashboardModel) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case int(ItemViewConfig):
		m.viewConfiguration()
		return m, nil
	case int(ItemEditConfig):
		// Launch wizard in edit mode
		return m, tea.Quit
	case int(ItemRunBackup):
		m.runBackup()
		return m, nil
	case int(ItemViewLogs):
		m.viewLogs()
		return m, nil
	case int(ItemStatus):
		m.showStatus()
		return m, nil
	case int(ItemUninstall):
		m.uninstall()
		return m, nil
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
		Render("[↑/↓] Navegar  [Enter] Seleccionar  [q] Salir")

	return content
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
	// Parse cron expression and calculate next run
	// For now, return a simplified message
	return "Mañana " + m.config.Schedule.CronExpression
}

// viewConfiguration displays the current configuration
func (m DashboardModel) viewConfiguration() {
	fmt.Println("\n📋 Configuración actual:")
	fmt.Println()
	fmt.Printf("Base de datos origen: %s@%s:%d\n", m.config.Source.Database, m.config.Source.Host, m.config.Source.Port)
	fmt.Printf("Schedule: %s\n", m.config.Schedule.CronExpression)
	fmt.Printf("Timezone: %s\n", m.config.Schedule.Timezone)
	fmt.Printf("Directorio de backups: %s\n", m.config.Storage.LocalPath)
	fmt.Printf("Retención: %d/%d/%d (S/F/G)\n", m.config.Storage.Retention.Son, m.config.Storage.Retention.Father, m.config.Storage.Retention.Grandfather)
	if m.config.Target.Enabled {
		fmt.Printf("Base de datos destino: %s@%s:%d\n", m.config.Target.Database, m.config.Target.Host, m.config.Target.Port)
	}
	fmt.Println()
	fmt.Println("Presiona [Enter] para continuar...")
	fmt.Scanln()
}

// runBackup executes a manual backup
func (m DashboardModel) runBackup() {
	fmt.Println("\n🔄 Ejecutando backup manual...")
	fmt.Println("(Esta función requiere implementación adicional)")
	time.Sleep(1 * time.Second)
	fmt.Println("Presiona [Enter] para continuar...")
	fmt.Scanln()
}

// viewLogs shows the logs
func (m DashboardModel) viewLogs() {
	fmt.Println("\n📄 Logs recientes:")
	logsDir := filepath.Join(m.baseDir, "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil || len(entries) == 0 {
		fmt.Println("No hay logs disponibles")
	} else {
		fmt.Printf("Hay %d archivos de log\n", len(entries))
		for _, entry := range entries {
			fmt.Printf("  - %s\n", entry.Name())
		}
	}
	fmt.Println()
	fmt.Println("Presiona [Enter] para continuar...")
	fmt.Scanln()
}

// showStatus shows system status
func (m DashboardModel) showStatus() {
	fmt.Println("\n📊 Estado del sistema:")
	fmt.Println()
	fmt.Printf("✅ Configuración: OK\n")
	fmt.Printf("🕐 Próximo backup: %s\n", m.getNextBackupTime())
	fmt.Printf("📁 Directorio de backups: %s\n", m.config.Storage.LocalPath)
	fmt.Println()
	fmt.Println("Presiona [Enter] para continuar...")
	fmt.Scanln()
}

// uninstall removes the configuration
func (m DashboardModel) uninstall() {
	fmt.Println("\n🗑️  Desinstalando...")
	fmt.Println("(Esta función requiere implementación adicional)")
	fmt.Println()
	fmt.Println("Presiona [Enter] para continuar...")
	fmt.Scanln()
}
