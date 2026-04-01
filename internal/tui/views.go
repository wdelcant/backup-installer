package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/wdelcant/backup-installer/internal/logo"
)

// viewWelcome renders the welcome screen
func (m Model) viewWelcome() string {
	var content string

	// Use logo's welcome screen
	content += logo.WelcomeScreen(Version)
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

	// Form fields with real inputs
	content += RenderLabel("Host") + m.sourceInputs[0].View() + "\n\n"
	content += RenderLabel("Puerto") + m.sourceInputs[1].View() + "\n\n"
	content += RenderLabel("Base de datos") + m.sourceInputs[2].View() + "\n\n"
	content += RenderLabel("Usuario") + m.sourceInputs[3].View() + "\n\n"
	content += RenderLabel("Contraseña") + m.sourceInputs[4].View() + "\n\n"

	content += helpStyle.Render("[Tab/Shift+Tab] Navegar  [Enter] Continuar  [Esc] Volver")

	return content
}

// viewMode renders the backup mode selection screen
func (m Model) viewMode() string {
	content := logo.Header("Configuración de Destino")
	content += "\n\n"

	content += "¿Deseás configurar una base de datos de destino para restore automático?\n\n"
	content += lipgloss.NewStyle().Foreground(colorTextMuted).Render(
		"Si configurás un destino, los backups de producción se restaurarán automáticamente en QA/Dev.") + "\n\n"

	options := []struct {
		title       string
		description string
	}{
		{"NO", "Solo backup - Guardar archivos .sql.gz localmente"},
		{"SÍ", "Backup + Restore - Restaurar automáticamente en otra base de datos"},
	}

	for i, opt := range options {
		selected := m.selectedMode == i
		style := lipgloss.NewStyle()
		if selected {
			style = style.Foreground(colorPrimary).Bold(true)
		}

		checkbox := "○"
		if selected {
			checkbox = "●"
		}

		content += fmt.Sprintf("%s %s\n", style.Render(checkbox+" "+opt.title), style.Render(opt.description))
		content += "\n"
	}

	content += "\n"
	content += buttonStyle.Render("[Enter] Continuar")
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

	// Form fields with real inputs
	content += RenderLabel("Host") + m.targetInputs[0].View() + "\n\n"
	content += RenderLabel("Puerto") + m.targetInputs[1].View() + "\n\n"
	content += RenderLabel("Base de datos") + m.targetInputs[2].View() + "\n\n"
	content += RenderLabel("Usuario") + m.targetInputs[3].View() + "\n\n"
	content += RenderLabel("Contraseña") + m.targetInputs[4].View() + "\n\n"

	content += helpStyle.Render("[Tab/Shift+Tab] Navegar  [Enter] Continuar  [Esc] Volver")

	return content
}

// viewSchedule renders the schedule configuration screen
func (m Model) viewSchedule() string {
	content := logo.Header("Horario de Ejecución")
	content += "\n\n"

	content += RenderLabel("Expresión Cron") + m.scheduleInputs[0].View() + "\n\n"
	content += RenderLabel("Zona horaria") + m.scheduleInputs[1].View() + "\n\n"

	content += helpStyle.Render("Ejemplos:\n")
	content += helpStyle.Render("  0 2 * * *     → Diario a las 02:00\n")
	content += helpStyle.Render("  0 */4 * * *   → Cada 4 horas\n")
	content += helpStyle.Render("  0 2 * * 1     → Semanal (lunes a las 02:00)")
	content += "\n\n"
	content += buttonStyle.Render("[Enter] Continuar")
	content += "\n\n"
	content += helpStyle.Render("[Tab/Shift+Tab] Navegar  [Esc] Volver")

	return content
}

// viewRetention renders the retention configuration screen
func (m Model) viewRetention() string {
	content := logo.Header("Retención y Almacenamiento")
	content += "\n\n"

	content += "Política de retención GFS (Grandfather-Father-Son)\n\n"
	content += lipgloss.NewStyle().Foreground(colorTextMuted).Render(
		"Esta política mantiene:\n• Backups diarios (Son)\n• Backups semanales (Father - domingos)\n• Backups mensuales (Grandfather - 1ro de mes)") + "\n\n"

	content += RenderLabel("Directorio local") + m.retentionInputs[0].View() + "\n\n"
	content += RenderLabel("Diarios (Son)") + m.retentionInputs[1].View() + "\n\n"
	content += RenderLabel("Semanales (Father)") + m.retentionInputs[2].View() + "\n\n"
	content += RenderLabel("Mensuales (Grandfather)") + m.retentionInputs[3].View() + "\n\n"

	content += "\n"
	content += helpStyle.Render("Ejemplo: 7/4/12 = 7 días + 4 semanas + 12 meses\n")
	content += "\n"
	content += buttonStyle.Render("[Enter] Continuar")
	content += "\n\n"
	content += helpStyle.Render("[Tab/Shift+Tab] Navegar  [Esc] Volver")

	return content
}

// viewWebhook renders the webhook configuration screen
func (m Model) viewWebhook() string {
	content := logo.Header("Notificaciones Webhook (n8n)")
	content += "\n\n"

	content += RenderLabel("URL Webhook") + m.webhookInputs[0].View() + "\n\n"
	content += RenderLabel("Token") + m.webhookInputs[1].View() + "\n\n"

	content += "\n\n"
	content += buttonStyle.Render("[Enter] Continuar")
	content += "   "
	content += lipgloss.NewStyle().Foreground(colorSecondary).Render("[s] Saltar")
	content += "\n\n"
	content += helpStyle.Render("[Tab/Shift+Tab] Navegar  [Esc] Volver")

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
		"ORIGEN:  %s @ %s:%d\n"+
			"DESTINO: %s ( %s )\n"+
			"SCHEDULE: %s\n"+
			"RETENCIÓN: %s\n"+
			"WEBHOOK: %s",
		m.config.Source.Database,
		m.config.Source.Host,
		m.config.Source.Port,
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
		fmt.Sprintf("%d/%d/%d (S/F/G)", m.config.Storage.Retention.Son, m.config.Storage.Retention.Father, m.config.Storage.Retention.Grandfather),
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
	content := logo.SuccessScreen(Version)
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
