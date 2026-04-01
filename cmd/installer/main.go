package main

import (
	"fmt"
	"os"
	"path/filepath"

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
		}
	}

	// Default: run the installer wizard
	if err := runWizard(); err != nil {
		fmt.Fprintf(os.Stderr, "\nError running installer: %v\n", err)
		os.Exit(1)
	}
}

// runWizard runs the interactive TUI installer
func runWizard() error {
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

	// Check if this is first run (no master key)
	isFirstRun := !keyManager.KeyExists()

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

	// Start TUI
	if err := tui.StartWizard(isFirstRun, baseDir, configManager, keyManager, encryptor); err != nil {
		return err
	}

	fmt.Println("\n✅ Installation completed successfully!")
	return nil
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
  --config, -c     Show current configuration
  --run, -r        Run backup manually
  --status, -s     Show backup status
  --uninstall, -U  Uninstall backup configuration

Examples:
  # Run installer
  %s

  # Update to latest version
  %s --update

  # Check for updates
  %s --check-update

For more information: https://github.com/wdelcant/%s
`, appName, appName, appName, appName, appName, appName)
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
	// TODO: Check if cron job is installed

	return nil
}
