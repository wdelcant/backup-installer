// Package logo provides ASCII art logos for the INVITSM Backup Installer
package logo

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Logo styles
var (
	primaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D4AA")). // Verde turquesa INVITSM
			Bold(true)

	secondaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")) // Gris

	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")). // Naranja/ámbar
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00D4AA")).
			Padding(1, 2).
			Align(lipgloss.Center)
)

// Logo variations
const (
	// LogoClassic - Versión clásica completa
	LogoClassic = `
    ██╗███╗   ██╗██╗   ██╗██╗████████╗███████╗███╗   ███╗
    ██║████╗  ██║██║   ██║██║╚══██╔══╝██╔════╝████╗ ████║
    ██║██╔██╗ ██║██║   ██║██║   ██║   ███████╗██╔████╔██║
    ██║██║╚██╗██║╚██╗ ██╔╝██║   ██║   ╚════██║██║╚██╔╝██║
    ██║██║ ╚████║ ╚████╔╝ ██║   ██║   ███████║██║ ╚═╝ ██║
    ╚═╝╚═╝  ╚═══╝  ╚═══╝  ╚═╝   ╚═╝   ╚══════╝╚═╝     ╚═╝`

	// LogoCompact - Versión compacta
	LogoCompact = `
    ██╗  ██╗██╗   ██╗██╗████████╗
    ██║  ██║██║   ██║██║╚══██╔══╝
    ███████║██║   ██║██║   ██║
    ██╔══██║╚██╗ ██╔╝██║   ██║
    ██║  ██║ ╚████╔╝ ██║   ██║
    ╚═╝  ╚═╝  ╚═══╝  ╚═╝   ╚═╝`

	// LogoMini - Versión mínima para headers
	LogoMini = `InvITSM`
)

// Rendered returns the full rendered logo with styling
func Rendered(version string) string {
	logo := primaryStyle.Render(LogoClassic)
	subtitle := secondaryStyle.Render("B A C K U P   I N S T A L L E R")
	ver := accentStyle.Render(fmt.Sprintf("v%s", version))

	content := fmt.Sprintf("%s\n\n%s\n%s", logo, subtitle, ver)
	return boxStyle.Render(content)
}

// RenderedCompact returns a compact version
func RenderedCompact(version string) string {
	logo := primaryStyle.Render(LogoCompact)
	subtitle := secondaryStyle.Render("Backup Installer")
	ver := accentStyle.Render(version)

	content := fmt.Sprintf("%s\n\n%s   %s", logo, subtitle, ver)
	return boxStyle.Render(content)
}

// Header returns a simple header for sub-screens
func Header(title string) string {
	logo := accentStyle.Render(LogoMini)
	sep := secondaryStyle.Render(" │ ")
	titleStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Render(title)

	return fmt.Sprintf("%s%s%s", logo, sep, titleStyled)
}

// Title returns just the styled logo text
func Title() string {
	return primaryStyle.Render(LogoMini)
}

// WelcomeScreen returns the full welcome screen
func WelcomeScreen(version string) string {
	var b strings.Builder

	b.WriteString(Rendered(version))
	b.WriteString("\n\n")

	// Welcome message
	welcome := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E5E7EB")).
		Align(lipgloss.Center).
		Width(60).
		Render("Bienvenido al configurador de backups automáticos para tu servidor INVITSM.")

	b.WriteString(welcome)
	b.WriteString("\n\n")

	// Requirements box
	reqBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4B5563")).
		Padding(1, 2).
		Width(58)

	reqContent := secondaryStyle.Render(`Requisitos:
  • Acceso a PostgreSQL (origen y destino)
  • Cron disponible en el sistema
  • [Opcional] URL de webhook n8n`)

	b.WriteString(reqBox.Render(reqContent))
	b.WriteString("\n\n")

	// CTA
	cta := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA")).
		Bold(true).
		Align(lipgloss.Center).
		Width(60).
		Render("[Enter] Comenzar configuración")

	b.WriteString(cta)

	return b.String()
}

// SuccessScreen returns the completion screen
func SuccessScreen(version string) string {
	var b strings.Builder

	// Success logo variant
	successLogo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true).
		Render(`
    ██╗███╗   ██╗██╗   ██╗██╗████████╗███████╗███╗   ███╗
    ██║████╗  ██║██║   ██║██║╚══██╔══╝██╔════╝████╗ ████║
    ██║██╔██╗ ██║██║   ██║██║   ██║   ███████╗██╔████╔██║
    ██║██║╚██╗██║╚██╗ ██╔╝██║   ██║   ╚════██║██║╚██╔╝██║
    ██║██║ ╚████║ ╚████╔╝ ██║   ██║   ███████║██║ ╚═╝ ██║
    ╚═╝╚═╝  ╚═══╝  ╚═══╝  ╚═╝   ╚═╝   ╚══════╝╚═╝     ╚═╝`)

	b.WriteString(successLogo)
	b.WriteString("\n\n")

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true).
		Align(lipgloss.Center).
		Width(60).
		Render("🎉 Instalación Completada")

	b.WriteString(title)
	b.WriteString("\n\n")

	return b.String()
}

// LoadingAnimation returns a loading spinner frame
func LoadingAnimation(frame int) string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinners[frame%len(spinners)]

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA")).
		Render(fmt.Sprintf("%s Procesando...", spinner))
}

// Footer returns a footer with helpful info
func Footer(text string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Align(lipgloss.Center).
		Width(60).
		Render(text)
}
