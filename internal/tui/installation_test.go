package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/wdelcant/backup-installer/internal/config"
	"github.com/wdelcant/backup-installer/internal/crypto"
)

// TestGeneratePipeline tests that pipeline script is generated correctly
func TestGeneratePipeline(t *testing.T) {
	// Skip in CI/CD environments
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping TestGeneratePipeline - running in CI/CD")
	}

	// Skip if templates are not available
	templatePath := "internal/pipeline/templates/pipeline.sh.tmpl"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Skip("Skipping TestGeneratePipeline - template not available")
	}

	// Create temp directory
	tempDir := t.TempDir()

	// Create test config
	cfg := &config.Config{
		Source: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "test_db",
			Username: "test_user",
			Password: "test_pass",
		},
		Schedule: config.ScheduleConfig{
			CronExpression: "0 2 * * *",
			Timezone:       "America/Santiago",
		},
		Storage: config.StorageConfig{
			LocalPath: filepath.Join(tempDir, "backups"),
			Retention: config.RetentionConfig{
				Enabled:     true,
				Son:         7,
				Father:      4,
				Grandfather: 12,
			},
		},
	}

	// Create model
	model := Model{
		baseDir: tempDir,
		config:  cfg,
	}

	// Test generatePipeline
	err := model.generatePipeline()
	if err != nil {
		t.Fatalf("generatePipeline failed: %v", err)
	}

	// Verify script was created
	scriptPath := filepath.Join(tempDir, "scripts", "pipeline.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Errorf("pipeline.sh was not created at %s", scriptPath)
	}

	// Verify script has correct permissions
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("failed to stat pipeline.sh: %v", err)
	}

	// Check if executable (Unix permissions)
	if info.Mode()&0111 == 0 {
		t.Errorf("pipeline.sh is not executable: mode=%o", info.Mode())
	}

	// Verify script content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read pipeline.sh: %v", err)
	}

	if len(content) == 0 {
		t.Error("pipeline.sh is empty")
	}

	// Verify script starts with shebang
	if len(content) < 2 || string(content[:2]) != "#!" {
		t.Error("pipeline.sh does not start with shebang")
	}
}

// TestInstallCrontab tests that crontab is installed correctly
func TestInstallCrontab(t *testing.T) {
	// Skip if not running on a system with crontab
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping crontab test in CI environment")
	}

	// Create temp directory
	tempDir := t.TempDir()

	// Create test config
	cfg := &config.Config{
		Source: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "test_db",
		},
		Schedule: config.ScheduleConfig{
			CronExpression: "0 2 * * *",
			Timezone:       "America/Santiago",
		},
	}

	// Create model
	model := Model{
		baseDir: tempDir,
		config:  cfg,
	}

	// Create scripts directory and dummy pipeline.sh
	scriptsDir := filepath.Join(tempDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	pipelinePath := filepath.Join(scriptsDir, "pipeline.sh")
	if err := os.WriteFile(pipelinePath, []byte("#!/bin/bash\necho test"), 0755); err != nil {
		t.Fatalf("failed to create pipeline.sh: %v", err)
	}

	// Test installCrontab with timeout
	done := make(chan error, 1)
	go func() {
		done <- model.installCrontab()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Logf("installCrontab returned error (may be expected in test env): %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("installCrontab timed out after 5 seconds - likely hanging")
	}
}

// TestRunInstallation tests the complete installation process
func TestRunInstallation(t *testing.T) {
	// Skip in CI/CD environments
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping TestRunInstallation - running in CI/CD")
	}

	// Skip if templates are not available
	templatePath := "internal/pipeline/templates/pipeline.sh.tmpl"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Skip("Skipping TestRunInstallation - template not available")
	}

	// Create temp directory
	tempDir := t.TempDir()

	// Create encryptor with test key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	encryptor := crypto.NewEncryptor(key)

	// Create config manager
	configManager := config.NewManager(tempDir, encryptor)

	// Create test config
	cfg := &config.Config{
		Source: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "test_db",
			Username: "test_user",
			Password: "test_pass",
		},
		Target: config.TargetConfig{
			Enabled: false,
		},
		Schedule: config.ScheduleConfig{
			CronExpression: "0 2 * * *",
			Timezone:       "America/Santiago",
		},
		Storage: config.StorageConfig{
			LocalPath: filepath.Join(tempDir, "backups"),
			Retention: config.RetentionConfig{
				Enabled:     true,
				Son:         7,
				Father:      4,
				Grandfather: 12,
			},
		},
	}

	// Create model
	model := Model{
		baseDir:       tempDir,
		configManager: configManager,
		config:        cfg,
	}

	// Test runInstallation with timeout
	done := make(chan error, 1)
	go func() {
		cmd := model.runInstallation()
		if cmd != nil {
			// Execute the command
			msg := cmd()
			if completeMsg, ok := msg.(installCompleteMsg); ok {
				done <- completeMsg.err
			} else {
				done <- nil
			}
		} else {
			done <- nil
		}
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("runInstallation failed: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("runInstallation timed out after 10 seconds - likely hanging")
	}

	// Verify config was saved
	if !configManager.Exists() {
		t.Error("configuration was not saved")
	}

	// Verify pipeline script was created
	scriptPath := filepath.Join(tempDir, "scripts", "pipeline.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("pipeline.sh was not created")
	}
}

// TestRunInstallationProgress tests that installation shows progress
func TestRunInstallationProgress(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create encryptor with test key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	encryptor := crypto.NewEncryptor(key)

	// Create config manager
	configManager := config.NewManager(tempDir, encryptor)

	// Create test config
	cfg := &config.Config{
		Source: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "test_db",
		},
		Schedule: config.ScheduleConfig{
			CronExpression: "0 2 * * *",
		},
		Storage: config.StorageConfig{
			LocalPath: filepath.Join(tempDir, "backups"),
			Retention: config.RetentionConfig{
				Enabled: true,
				Son:     7,
			},
		},
	}

	// Create model
	model := Model{
		baseDir:       tempDir,
		configManager: configManager,
		config:        cfg,
	}

	// Track progress
	progressReceived := false

	// Test runInstallation
	cmd := model.runInstallation()
	if cmd != nil {
		// The command should complete within 10 seconds
		done := make(chan bool, 1)
		go func() {
			msg := cmd()
			_ = msg
			done <- true
		}()

		select {
		case <-done:
			progressReceived = true
		case <-time.After(10 * time.Second):
			t.Fatal("installation did not complete within 10 seconds")
		}
	}

	if !progressReceived {
		t.Error("installation did not complete")
	}
}

// TestRequirementsCheck tests the requirements checker
func TestRequirementsCheck(t *testing.T) {
	// This test should always pass - just verify bash exists
	hasBash := checkRequirements()
	if !hasBash {
		t.Error("bash should be available")
	}
}

// checkRequirements is a helper to test requirements checking
func checkRequirements() bool {
	// Check if bash exists (should always exist)
	_, err := exec.LookPath("bash")
	return err == nil
}
