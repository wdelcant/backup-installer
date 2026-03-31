package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/invitsm/invitsm-backup-installer/internal/logo"
)

// viewWelcome renders the welcome screen
func (m Model) viewWelcome() string {
	var content string

	// Use logo's welcome screen
	content += logo.WelcomeScreen("1.0.0")
	content += "\n\n"

	// Keyboard shortcuts
	shortcuts := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 2).
		Width(60)

	shortcutsContent := fmt.Sprintf("%s Comenzar configuración\n%s Salir sin instalar",
		lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render("[Enter]"),
		lipgloss.NewStyle().Foreground(colorSecondary).Render("[Esc]"))

	content += shortcuts.Render(shortcutsContent)

	return content
}

// viewSecurity renders the security setup screen
func (m Model) viewSecurity() string {
	content := logo.Header("Configuración de Seguridad")
	content += "\n\n"

	content += "Se generará una clave maestra para proteger tus credenciales.\n\n"

	// Security info box
	infoBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorInfo).
		Padding(1, 2).
		Width(60)

	infoContent := fmt.Sprintf("📍 Ubicación: %s\n\n🔐 Algoritmo: AES-256-GCM\n\n⚠️  Las credenciales se almacenarán encriptadas y solo serán accesibles desde este servidor.",
		lipgloss.NewStyle().Foreground(colorText).Render(m.keyManager.GetKeyPath()))

	content += infoBox.Render(infoContent)
	content += "\n\n"

	// Continue button
	content += buttonStyle.Render("[Enter] Generar clave y continuar")
	content += "\n\n"
	content += helpStyle.Render("[Esc] Volver")

	return content
}

// viewDBSource renders the source database configuration screen
func (m Model) viewDBSource() string {
	content := logo.Header("Base de Datos ORIGEN (Producción)")
	content += "\n\n"

	// Form fields
	content += RenderLabel("Host") + RenderInput(m.dbSourceHost, m.cursor == 0) + "\n"
	content += RenderLabel("Puerto") + RenderInput(m.dbSourcePort, m.cursor == 1) + "\n"
	content += RenderLabel("Base de datos") + RenderInput(m.dbSourceDatabase, m.cursor == 2) + "\n"
	content += RenderLabel("Usuario") + RenderInput(m.dbSourceUsername, m.cursor == 3) + "\n"
	content += RenderLabel("Contraseña") + RenderInput("••••••••", m.cursor == 4) + "\n"

	content += "\n"
	content += helpStyle.Render("[Tab/Shift+Tab] Navegar  [Enter] Continuar  [Esc] Volver")

	return content
}

// viewMode renders the backup mode selection screen
func (m Model) viewMode() string {
	content := logo.Header("Modo de Backup")
	content += "\n\n"

	modes := []struct {
		id          int
		title       string
		description string
	}{
		{0, "Solo Backup", "Guarda backups .sql.gz localmente"},
		{1, "Backup + Restore en QA/Dev", "Backup prod → Restaura automáticamente en QA"},
		{2, "Backup + Restore Manual", "Te avisa cuando está listo para restaurar"},
	}

	for i, mode := range modes {
		selected := m.cursor == i
		style := lipgloss.NewStyle()
		if selected {
			style = style.Foreground(colorPrimary).Bold(true)
		}

		checkbox := "○"
		if selected {
			checkbox = "●"
		}

		content += fmt.Sprintf("%s %s\n", style.Render(checkbox+" "+mode.title), style.Render(mode.description))
		content += "\n"
	}

	content += "\n"
	content += buttonStyle.Render("[Enter] Seleccionar")
	content += "\n\n"
	content += helpStyle.Render("[↑/↓] Cambiar opción  [Esc] Volver")

	return content
}

// viewDBTarget renders the target database configuration screen
func (m Model) viewDBTarget() string {
	content := logo.Header("Base de Datos DESTINO (QA/Dev)")
	content += "\n\n"

	// Warning box
	warningBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2).
		Width(60)

	warningContent := "⚠️  ADVERTENCIA: Esta base de datos será SOBRESCRITA completamente durante cada restore automático."
	content += warningBox.Render(warningContent)
	content += "\n\n"

	// Form fields
	content += RenderLabel("Host") + RenderInput(m.dbTargetHost, m.cursor == 0) + "\n"
	content += RenderLabel("Puerto") + RenderInput(m.dbTargetPort, m.cursor == 1) + "\n"
	content += RenderLabel("Base de datos") + RenderInput(m.dbTargetDatabase, m.cursor == 2) + "\n"
	content += RenderLabel("Usuario") + RenderInput(m.dbTargetUsername, m.cursor == 3) + "\n"
	content += RenderLabel("Contraseña") + RenderInput("••••••••", m.cursor == 4) + "\n"

	content += "\n"
	content += helpStyle.Render("[Tab/Shift+Tab] Navegar  [Enter] Continuar  [Esc] Volver")

	return content
}

// viewSchedule renders the schedule configuration screen
func (m Model) viewSchedule() string {
	content := logo.Header("Horario de Ejecución")
	content += "\n\n"

	content += RenderLabel("Expresión Cron") + RenderInput(m.cronExpression, true) + "\n"
	content += "\n"
	content += RenderLabel("Zona horaria") + RenderInput(m.timezone, m.cursor == 1) + "\n"

	content += "\n\n"
	content += helpStyle.Render("Ejemplos:\n")
	content += helpStyle.Render("  0 2 * * *     → Diario a las 02:00\n")
	content += helpStyle.Render("  0 */4 * * *   → Cada 4 horas\n")
	content += helpStyle.Render("  0 2 * * 1     → Semanal (lunes a las 02:00)")
	content += "\n\n"
	content += buttonStyle.Render("[Enter] Continuar")
	content += "\n\n"
	content += helpStyle.Render("[Esc] Volver")

	return content
}

// viewRetention renders the retention configuration screen
func (m Model) viewRetention() string {
	content := logo.Header("Retención y Almacenamiento")
	content += "\n\n"

	content += RenderLabel("Directorio local") + RenderInput(m.backupPath, m.cursor == 0) + "\n"
	content += RenderLabel("Días a mantener") + RenderInput(fmt.Sprintf("%d", m.retentionDays), m.cursor == 1) + "\n"

	content += "\n\n"
	content += helpStyle.Render("Los backups más antiguos se eliminarán automáticamente.\n")
	content += "\n"
	content += buttonStyle.Render("[Enter] Continuar")
	content += "\n\n"
	content += helpStyle.Render("[Esc] Volver")

	return content
}

// viewWebhook renders the webhook configuration screen
func (m Model) viewWebhook() string {
	content := logo.Header("Notificaciones Webhook (n8n)")
	content += "\n\n"

	if m.webhookEnabled {
		content += RenderLabel("URL Webhook") + RenderInput(m.webhookURL, m.cursor == 0) + "\n"
		content += RenderLabel("Token") + RenderInput("••••••••", m.cursor == 1) + "\n"
	} else {
		content += "¿Quieres configurar notificaciones webhook a n8n?\n\n"
		content += helpStyle.Render("Recibirás notificaciones con el estado de cada backup.\n")
	}

	content += "\n\n"
	content += buttonStyle.Render("[Enter] Continuar")
	content += "   "
	content += lipgloss.NewStyle().Foreground(colorSecondary).Render("[s] Saltar")
	content += "\n\n"
	content += helpStyle.Render("[Esc] Volver")

	return content
}

// viewSummary renders the configuration summary screen
func (m Model) viewSummary() string {
	content := logo.Header("Resumen de Configuración")
	content += "\n\n"

	summaryBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Width(60)

	summary := fmt.Sprintf(
		"ORIGEN:  %s @ %s:%s\n"+
			"DESTINO: %s ( %s )\n"+
			"SCHEDULE: %s\n"+
			"RETENCIÓN: %d días\n"+
			"WEBHOOK: %s",
		m.config.Source.Database,
		m.config.Source.Host,
		fmt.Sprintf("%d", m.config.Source.Port),
		func() string {
			if m.config.Target.Enabled {
				return m.config.Target.Database
			}
			return "No configurado"
		}(),
		func() string {
			if m.config.Target.Enabled {
				return "Backup + Restore"
			}
			return "Solo Backup"
		}(),
		m.config.Schedule.CronExpression,
		m.config.Storage.RetentionDays,
		func() string {
			if m.config.Webhook.Enabled {
				return "Sí configurado"
			}
			return "No configurado"
		}(),
	)

	content += summaryBox.Render(summary)
	content += "\n\n"
	content += buttonStyle.Render("[Enter] Instalar configuración")
	content += "\n\n"
	content += helpStyle.Render("[Esc] Volver y modificar")

	return content
}

// viewInstalling renders the installation progress screen
func (m Model) viewInstalling() string {
	content := logo.Header("Instalando...")
	content += "\n\n"

	// Progress bar
	progressBar := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Width(50).
		Render(fmt.Sprintf("[%5.1f%%]", m.installProgress))

	content += progressBar
	content += "\n\n"
	content += m.installStatus
	content += "\n\n"
	content += helpStyle.Render("Por favor espera...")

	return content
}

// viewSuccess renders the success screen
func (m Model) viewSuccess() string {
	content := logo.SuccessScreen("1.0.0")
	content += "\n\n"

	// Success details
	details := []string{
		"✅ Dependencias verificadas",
		"✅ Configuración guardada",
		"✅ Script de pipeline generado",
		"✅ Crontab instalado",
	}

	for _, detail := range details {
		content += detail + "\n"
	}

	content += "\n"
	content += "Próxima ejecución: Mañana 02:00\n"
	content += "\n\n"
	content += buttonStyle.Render("[Enter] Salir")
	content += "   "
	content += lipgloss.NewStyle().Foreground(colorAccent).Render("[r] Ejecutar backup ahora")

	return content
}
