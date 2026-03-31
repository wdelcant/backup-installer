// Package security provides security hardening utilities
package security

import (
	"fmt"
	"os"
	"path/filepath"
)

// ApplyHardening applies security hardening to the installation
func ApplyHardening(configPath, keyPath string) error {
	// Set secure permissions on config file (0600)
	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("failed to set config permissions: %w", err)
	}

	// Set secure permissions on key file (0400)
	if err := os.Chmod(keyPath, 0400); err != nil {
		return fmt.Errorf("failed to set key permissions: %w", err)
	}

	// Set secure permissions on config directory (0700)
	configDir := filepath.Dir(configPath)
	if err := os.Chmod(configDir, 0700); err != nil {
		return fmt.Errorf("failed to set config dir permissions: %w", err)
	}

	// Set secure permissions on key directory (0700)
	keyDir := filepath.Dir(keyPath)
	if err := os.Chmod(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to set key dir permissions: %w", err)
	}

	// Set secure permissions on scripts directory (0755)
	scriptsDir := filepath.Join(configDir, "..", "scripts")
	if _, err := os.Stat(scriptsDir); err == nil {
		if err := os.Chmod(scriptsDir, 0755); err != nil {
			return fmt.Errorf("failed to set scripts dir permissions: %w", err)
		}
	}

	// Set secure permissions on logs directory (0755)
	logsDir := filepath.Join(configDir, "..", "logs")
	if _, err := os.Stat(logsDir); err == nil {
		if err := os.Chmod(logsDir, 0755); err != nil {
			return fmt.Errorf("failed to set logs dir permissions: %w", err)
		}
	}

	return nil
}

// VerifyPermissions checks if file permissions are secure
func VerifyPermissions(path string, expectedMode os.FileMode) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	mode := info.Mode().Perm()
	return mode == expectedMode, nil
}

// ScrubSensitiveData removes sensitive data from logs
func ScrubSensitiveData(data map[string]interface{}) map[string]interface{} {
	scrubbed := make(map[string]interface{})
	for k, v := range data {
		switch k {
		case "password", "token", "secret", "key", "authorization":
			scrubbed[k] = "***REDACTED***"
		default:
			scrubbed[k] = v
		}
	}
	return scrubbed
}
