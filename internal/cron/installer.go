// Package cron handles crontab installation and management
package cron

import (
	"fmt"
	"os/exec"
	"strings"
)

// Installer manages crontab entries
type Installer struct {
	scriptPath string
	cronExpr   string
}

// NewInstaller creates a new cron installer
func NewInstaller(scriptPath, cronExpr string) *Installer {
	return &Installer{
		scriptPath: scriptPath,
		cronExpr:   cronExpr,
	}
}

// Install adds the cron job
func (i *Installer) Install() error {
	// Get current crontab
	current, err := i.getCurrentCrontab()
	if err != nil {
		return err
	}

	// Check if job already exists
	if i.jobExists(current) {
		return nil // Already installed
	}

	// Add new job
	newEntry := fmt.Sprintf("%s %s >> /var/log/invitsm-backup.log 2>&1", i.cronExpr, i.scriptPath)
	newCrontab := current + newEntry + "\n"

	// Install new crontab
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(newCrontab)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install crontab: %w", err)
	}

	return nil
}

// Uninstall removes the cron job
func (i *Installer) Uninstall() error {
	current, err := i.getCurrentCrontab()
	if err != nil {
		return err
	}

	// Remove our job
	lines := strings.Split(current, "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.Contains(line, i.scriptPath) {
			newLines = append(newLines, line)
		}
	}

	newCrontab := strings.Join(newLines, "\n")

	// Install new crontab (or remove if empty)
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(newCrontab)
	if err := cmd.Run(); err != nil {
		// If crontab is empty, this might fail - that's OK
		if !strings.Contains(err.Error(), "no crontab") {
			return fmt.Errorf("failed to update crontab: %w", err)
		}
	}

	return nil
}

// getCurrentCrontab gets the current user's crontab
func (i *Installer) getCurrentCrontab() (string, error) {
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// No crontab installed - that's OK
			return "", nil
		}
		return "", fmt.Errorf("failed to get crontab: %w", err)
	}
	return string(output), nil
}

// jobExists checks if the job is already in crontab
func (i *Installer) jobExists(crontab string) bool {
	return strings.Contains(crontab, i.scriptPath)
}

// Verify checks if the cron job is installed
func (i *Installer) Verify() (bool, error) {
	current, err := i.getCurrentCrontab()
	if err != nil {
		return false, err
	}
	return i.jobExists(current), nil
}
