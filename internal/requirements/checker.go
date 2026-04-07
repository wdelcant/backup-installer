// Package requirements provides system requirements checking
package requirements

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Requirement represents a system requirement
type Requirement struct {
	Name        string
	Description string
	Command     string
	PackageName string
	Installed   bool
	Optional    bool
}

// CheckResult holds the result of requirements check
type CheckResult struct {
	AllInstalled    bool
	MissingRequired []Requirement
	MissingOptional []Requirement
	Installed       []Requirement
}

// Check checks if all requirements are installed
func Check() *CheckResult {
	requirements := []Requirement{
		{
			Name:        "cron",
			Description: "Cron daemon for scheduled backups",
			Command:     "crontab",
			PackageName: "cron",
			Optional:    false,
		},
		{
			Name:        "pg_dump",
			Description: "PostgreSQL backup utility",
			Command:     "pg_dump",
			PackageName: "postgresql-client",
			Optional:    false,
		},
		{
			Name:        "psql",
			Description: "PostgreSQL client for restore operations",
			Command:     "psql",
			PackageName: "postgresql-client",
			Optional:    true,
		},
		{
			Name:        "gzip",
			Description: "Compression utility for backup files",
			Command:     "gzip",
			PackageName: "gzip",
			Optional:    false,
		},
		{
			Name:        "bash",
			Description: "Bash shell for running scripts",
			Command:     "bash",
			PackageName: "bash",
			Optional:    false,
		},
	}

	result := &CheckResult{
		AllInstalled:    true,
		MissingRequired: []Requirement{},
		MissingOptional: []Requirement{},
		Installed:       []Requirement{},
	}

	for _, req := range requirements {
		if commandExists(req.Command) {
			req.Installed = true
			result.Installed = append(result.Installed, req)
		} else {
			req.Installed = false
			if req.Optional {
				result.MissingOptional = append(result.MissingOptional, req)
			} else {
				result.MissingRequired = append(result.MissingRequired, req)
				result.AllInstalled = false
			}
		}
	}

	return result
}

// commandExists checks if a command exists in PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// RenderResult renders the check result with styling
func RenderResult(result *CheckResult) string {
	var sb strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA")).
		Bold(true)
	sb.WriteString(titleStyle.Render("🔍 Verificando requisitos del sistema"))
	sb.WriteString("\n\n")

	// Installed requirements
	if len(result.Installed) > 0 {
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
		sb.WriteString(successStyle.Render("✅ Requisitos instalados:\n"))
		for _, req := range result.Installed {
			sb.WriteString(fmt.Sprintf("   • %s - %s\n", req.Name, req.Description))
		}
		sb.WriteString("\n")
	}

	// Missing required
	if len(result.MissingRequired) > 0 {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
		sb.WriteString(errorStyle.Render("❌ Requisitos FALTANTES (críticos):\n"))
		for _, req := range result.MissingRequired {
			sb.WriteString(fmt.Sprintf("   • %s - %s\n", req.Name, req.Description))
			sb.WriteString(fmt.Sprintf("     Instalar: %s\n", getInstallCommand(req.PackageName)))
		}
		sb.WriteString("\n")
	}

	// Missing optional
	if len(result.MissingOptional) > 0 {
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
		sb.WriteString(warnStyle.Render("⚠️  Requisitos opcionales no instalados:\n"))
		for _, req := range result.MissingOptional {
			sb.WriteString(fmt.Sprintf("   • %s - %s\n", req.Name, req.Description))
			sb.WriteString(fmt.Sprintf("     Instalar (opcional): %s\n", getInstallCommand(req.PackageName)))
		}
		sb.WriteString("\n")
	}

	// Summary
	if result.AllInstalled {
		successBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#10B981")).
			Padding(1, 2)
		sb.WriteString(successBox.Render("✅ ¡Todos los requisitos están instalados!\n\nPodés continuar con la instalación."))
	} else {
		errorBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#EF4444")).
			Padding(1, 2)

		errorMsg := "❌ FALTAN REQUISITOS CRÍTICOS\n\n"
		errorMsg += "Por favor instalá los paquetes faltantes antes de continuar:\n\n"
		for _, req := range result.MissingRequired {
			errorMsg += fmt.Sprintf("  %s\n", getInstallCommand(req.PackageName))
		}
		errorMsg += "\nDespués de instalar, volvé a ejecutar: backup-installer"

		sb.WriteString(errorBox.Render(errorMsg))
	}

	return sb.String()
}

// getInstallCommand returns the install command for a package
func getInstallCommand(packageName string) string {
	// Detect package manager
	commands := []string{
		fmt.Sprintf("sudo apt-get install -y %s", packageName), // Debian/Ubuntu
		fmt.Sprintf("sudo yum install -y %s", packageName),     // RHEL/CentOS
		fmt.Sprintf("sudo dnf install -y %s", packageName),     // Fedora
		fmt.Sprintf("sudo pacman -S %s", packageName),          // Arch
		fmt.Sprintf("sudo apk add %s", packageName),            // Alpine
	}

	// Try to detect which package manager is available
	managers := map[string]string{
		"apt-get": "Debian/Ubuntu",
		"yum":     "RHEL/CentOS",
		"dnf":     "Fedora",
		"pacman":  "Arch",
		"apk":     "Alpine",
	}

	for manager, distro := range managers {
		if commandExists(manager) {
			return fmt.Sprintf("sudo %s install -y %s  # %s", manager, packageName, distro)
		}
	}

	// Return all options if no package manager detected
	var sb strings.Builder
	sb.WriteString("Comandos para diferentes distribuciones:\n")
	for _, cmd := range commands {
		sb.WriteString(fmt.Sprintf("  %s\n", cmd))
	}
	return sb.String()
}

// QuickCheck performs a quick check and returns true if all required are installed
func QuickCheck() bool {
	result := Check()
	return result.AllInstalled
}
