package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wdelcant/backup-installer/internal/config"
	"github.com/wdelcant/backup-installer/internal/crypto"
	"github.com/wdelcant/backup-installer/internal/logo"
	"github.com/wdelcant/backup-installer/internal/tui"
	"github.com/wdelcant/backup-installer/internal/version"
)

// Version is set during build via ldflags
var Version = "dev"

const appName = "backup-installer"

func main() {
	// Parse arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			printVersion()
			return
		case "--help", "-h", "help":
			printHelp()
			return
		case "--update", "-u", "update":
			if err := runUpdate(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "--check-update", "check-update":
			if err := checkUpdate(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "--uninstall", "-U", "uninstall":
			if err := runUninstall(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "--config", "-c", "config":
			if err := showConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "--run", "-r", "run":
			if err := runBackup(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "--status", "-s", "status":
			if err := showStatus(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "--diagnose", "-d", "diagnose":
			if err := runDiagnose(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// Default: run the installer wizard
	if err := runWizard(false); err != nil {
		fmt.Fprintf(os.Stderr, "\nError running installer: %v\n", err)
		os.Exit(1)
	}
}

// runWizard runs the interactive TUI installer or dashboard
// forceWizard: if true, always show wizard even if config exists
func runWizard(forceWizard bool) error {
	// Get base directory (where the binary is located)
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	baseDir := filepath.Dir(execPath)

	// For development, use current directory if binary is in /tmp or similar
	if filepath.Base(baseDir) == "tmp" || filepath.Base(baseDir) == "go-build" {
		cwd, _ := os.Getwd()
		baseDir = cwd
	}

	// Initialize master key manager
	configDir := filepath.Join(baseDir, "config")
	keyManager := crypto.NewMasterKeyManager(configDir)

	// Get or create master key
	masterKey, err := keyManager.GetOrCreateMasterKey()
	if err != nil {
		return fmt.Errorf("error initializing master key: %w", err)
	}

	// Initialize encryptor
	encryptor := crypto.NewEncryptor(masterKey)

	// Initialize config manager
	configManager := config.NewManager(baseDir, encryptor)

	// Set version in TUI package
	tui.Version = Version

	// Check for updates before starting
	checker := version.NewChecker(Version)
	if hasUpdate, release, err := checker.IsUpdateAvailable(); err == nil && hasUpdate {
		fmt.Println()
		fmt.Println(version.RenderUpdateNotification(Version, release.TagName))
		fmt.Println()
	}

	// Show logo
	fmt.Println(logo.Rendered(Version))
	fmt.Println()

	// Main loop - keeps showing dashboard after wizard completes
	for {
		// Check if configuration already exists and we're not forcing wizard mode
		if configManager.Exists() && !forceWizard {
			// Load existing configuration
			existingConfig, err := configManager.Load()
			if err != nil {
				return fmt.Errorf("error loading existing config: %w", err)
			}

			// Show dashboard with existing configuration
			err = tui.StartDashboard(existingConfig, baseDir, configManager, keyManager, encryptor)
			if err == tui.ErrEditRequested {
				// User wants to edit configuration, show wizard
				fmt.Println("\n🔄 Reconfigurando...")
				fmt.Println()
				if err := tui.StartWizard(false, baseDir, configManager, keyManager, encryptor); err != nil {
					return err
				}
				fmt.Println("\n✅ Configuración actualizada exitosamente!")
				// Loop continues - will show dashboard again
				continue
			}
			if err != nil {
				return err
			}
			// User exited dashboard normally
			return nil
		}

		// No configuration exists OR force wizard mode, start wizard
		if err := tui.StartWizard(true, baseDir, configManager, keyManager, encryptor); err != nil {
			return err
		}

		fmt.Println("\n✅ Installation completed successfully!")
		// After wizard completes, loop continues
		// This will show dashboard on next iteration if config exists
		forceWizard = false
	}
}

// printVersion prints the current version
func printVersion() {
	fmt.Printf("%s version %s\n", appName, Version)
}

// printHelp prints the help message
func printHelp() {
	fmt.Printf(`%s - PostgreSQL Backup Installer

Usage:
  %s [command]

Commands:
  (no command)     Run the interactive installer wizard
  --version, -v    Show version information
  --help, -h       Show this help message
  --update, -u     Self-update to the latest version
  --check-update   Check if a new version is available
  --diagnose, -d   Run diagnostic check
  --config, -c     Show current configuration
  --run, -r        Run backup manually
  --status, -s     Show backup status
  --uninstall, -U  Uninstall backup configuration

Examples:
  # Run installer
  %s

  # Diagnose installation
  %s --diagnose

  # Update to latest version
  %s --update

  # Check for updates
  %s --check-update

For more information: https://github.com/wdelcant/%s
`, appName, appName, appName, appName, appName, appName, appName)
}

// runUpdate performs a self-update
func runUpdate() error {
	fmt.Println("🔍 Checking for updates...")

	checker := version.NewChecker(Version)
	hasUpdate, release, err := checker.IsUpdateAvailable()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !hasUpdate {
		fmt.Printf("✅ You are already on the latest version: %s\n", Version)
		return nil
	}

	fmt.Printf("🔄 Updating %s → %s\n\n", Version, release.TagName)
	return version.SelfUpdate()
}

// checkUpdate checks if an update is available
func checkUpdate() error {
	checker := version.NewChecker(Version)
	hasUpdate, release, err := checker.IsUpdateAvailable()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if hasUpdate {
		fmt.Println(version.RenderUpdateNotification(Version, release.TagName))
	} else {
		fmt.Printf("✅ You are on the latest version: %s\n", Version)
	}

	return nil
}

// runUninstall removes the backup configuration
func runUninstall() error {
	fmt.Println("🗑️  Uninstalling backup installer...")

	// Get base directory
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	baseDir := filepath.Dir(execPath)

	// Remove crontab
	// TODO: Implement crontab removal

	// Remove config
	configDir := filepath.Join(baseDir, "config")
	if err := os.RemoveAll(configDir); err != nil {
		return fmt.Errorf("failed to remove config: %w", err)
	}

	// Remove scripts
	scriptsDir := filepath.Join(baseDir, "scripts")
	if err := os.RemoveAll(scriptsDir); err != nil {
		return fmt.Errorf("failed to remove scripts: %w", err)
	}

	fmt.Println("✅ Uninstalled successfully!")
	fmt.Println("📝 Note: Backup files in ./backups were not removed")
	return nil
}

// showConfig displays the current configuration
func showConfig() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	baseDir := filepath.Dir(execPath)
	configPath := filepath.Join(baseDir, "config", "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("no configuration found. Run '%s' to configure", appName)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// runBackup runs a manual backup
func runBackup() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	baseDir := filepath.Dir(execPath)
	scriptPath := filepath.Join(baseDir, "scripts", "pipeline.sh")

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup script found. Run '%s' to configure first", appName)
	}

	fmt.Println("🔄 Running backup...")
	// TODO: Execute the script
	return nil
}

// showStatus shows the backup status
func showStatus() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	baseDir := filepath.Dir(execPath)
	configPath := filepath.Join(baseDir, "config", "config.yaml")
	logsDir := filepath.Join(baseDir, "logs")

	fmt.Println("📊 Backup Status")
	fmt.Println()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("❌ Not configured")
		return nil
	}

	fmt.Println("✅ Configuration: OK")

	// Check for recent logs
	entries, err := os.ReadDir(logsDir)
	if err != nil || len(entries) == 0 {
		fmt.Println("📁 No backups run yet")
	} else {
		fmt.Printf("📁 Logs: %d backup(s) found\n", len(entries))
	}

	// Check crontab
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("⚠️  Could not read crontab")
	} else {
		cronContent := string(output)
		if strings.Contains(cronContent, "backup") {
			fmt.Println("✅ Cron job: Installed")
		} else {
			fmt.Println("❌ Cron job: Not found")
		}
	}

	return nil
}

// runDiagnose performs a full diagnostic of the installation
func runDiagnose() error {
	fmt.Println("🔍 Diagnóstico de instalación")
	fmt.Println()

	// Get base directory
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %w", err)
	}

	baseDir := filepath.Dir(execPath)

	fmt.Printf("📂 Directorio base: %s\n", baseDir)
	fmt.Println()

	// Check binary
	fmt.Println("1️⃣  Verificando binario:")
	if _, err := os.Stat(execPath); err == nil {
		info, _ := os.Stat(execPath)
		fmt.Printf("   ✅ Binario encontrado: %s (%d bytes)\n", execPath, info.Size())
		fmt.Printf("   📅 Modificado: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("   ❌ Binario no encontrado: %s\n", execPath)
	}
	fmt.Println()

	// Check config directory
	fmt.Println("2️⃣  Verificando configuración:")
	configDir := filepath.Join(baseDir, "config")
	configPath := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configDir); err == nil {
		fmt.Printf("   ✅ Directorio de config existe: %s\n", configDir)

		// List files in config dir
		entries, err := os.ReadDir(configDir)
		if err == nil && len(entries) > 0 {
			fmt.Printf("   📁 Archivos en config (%d):\n", len(entries))
			for _, entry := range entries {
				fmt.Printf("      - %s\n", entry.Name())
			}
		} else if len(entries) == 0 {
			fmt.Printf("   ⚠️  Directorio de config vacío\n")
		}

		// Check config.yaml specifically
		if _, err := os.Stat(configPath); err == nil {
			info, _ := os.Stat(configPath)
			fmt.Printf("   ✅ config.yaml existe (%d bytes)\n", info.Size())
			fmt.Printf("   📅 Modificado: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("   ❌ config.yaml no existe\n")
		}
	} else {
		fmt.Printf("   ❌ Directorio de config no existe: %s\n", configDir)
	}
	fmt.Println()

	// Check scripts directory
	fmt.Println("3️⃣  Verificando scripts:")
	scriptsDir := filepath.Join(baseDir, "scripts")
	if _, err := os.Stat(scriptsDir); err == nil {
		fmt.Printf("   ✅ Directorio de scripts existe: %s\n", scriptsDir)
		entries, _ := os.ReadDir(scriptsDir)
		if len(entries) > 0 {
			fmt.Printf("   📁 Scripts encontrados (%d):\n", len(entries))
			for _, entry := range entries {
				fmt.Printf("      - %s\n", entry.Name())
			}
		}
	} else {
		fmt.Printf("   ❌ Directorio de scripts no existe\n")
	}
	fmt.Println()

	// Check crontab
	fmt.Println("4️⃣  Verificando crontab:")
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("   ⚠️  No se pudo leer crontab: %v\n", err)
	} else {
		cronContent := string(output)
		if strings.Contains(cronContent, "backup") {
			fmt.Printf("   ✅ Encontrado cronjob de backup:\n")
			lines := strings.Split(cronContent, "\n")
			for _, line := range lines {
				if strings.Contains(line, "backup") {
					fmt.Printf("      %s\n", line)
				}
			}
		} else {
			fmt.Printf("   ❌ No se encontró cronjob de backup\n")
		}
	}
	fmt.Println()

	// Check logs directory
	fmt.Println("5️⃣  Verificando logs:")
	logsDir := filepath.Join(baseDir, "logs")
	if _, err := os.Stat(logsDir); err == nil {
		fmt.Printf("   ✅ Directorio de logs existe: %s\n", logsDir)
		entries, _ := os.ReadDir(logsDir)
		if len(entries) > 0 {
			fmt.Printf("   📄 Archivos de log (%d):\n", len(entries))
			for _, entry := range entries {
				if !entry.IsDir() {
					info, _ := entry.Info()
					fmt.Printf("      - %s (%d bytes, %s)\n",
						entry.Name(),
						info.Size(),
						info.ModTime().Format("2006-01-02 15:04"))
				}
			}
		} else {
			fmt.Printf("   ⚠️  No hay archivos de log\n")
		}
	} else {
		fmt.Printf("   ⚠️  Directorio de logs no existe\n")
	}
	fmt.Println()

	// Summary
	fmt.Println("📋 Resumen:")
	fmt.Println("   Usa estos comandos para gestionar tu instalación:")
	fmt.Printf("   - %s --diagnose : Diagnóstico completo\n", appName)
	fmt.Printf("   - %s --config    : Ver configuración\n", appName)
	fmt.Printf("   - %s --status    : Ver estado\n", appName)
	fmt.Printf("   - %s --run       : Ejecutar backup manual\n", appName)
	fmt.Printf("   - %s --uninstall : Desinstalar\n", appName)
	fmt.Println()

	return nil
}
