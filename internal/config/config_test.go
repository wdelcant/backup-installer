package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wdelcant/backup-installer/internal/crypto"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", config.Version)
	}

	if config.Source.Port != 5432 {
		t.Errorf("Expected default source port 5432, got %d", config.Source.Port)
	}

	if config.Target.Port != 5432 {
		t.Errorf("Expected default target port 5432, got %d", config.Target.Port)
	}

	if config.Schedule.CronExpression != "0 2 * * *" {
		t.Errorf("Expected default cron '0 2 * * *', got %s", config.Schedule.CronExpression)
	}

	if config.Storage.RetentionDays != DefaultRetentionDays {
		t.Errorf("Expected default retention %d, got %d", DefaultRetentionDays, config.Storage.RetentionDays)
	}
}

func TestManagerSaveAndLoad(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create encryptor with test key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	encryptor := crypto.NewEncryptor(key)

	// Create manager
	manager := NewManager(tempDir, encryptor)

	// Create test config with expected values
	expectedSourcePassword := "test_password"
	expectedTargetPassword := "target_password"

	config := &Config{
		Version: "1.0",
		Source: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "test_db",
			Username: "test_user",
			Password: expectedSourcePassword,
		},
		Target: TargetConfig{
			Enabled:  true,
			Host:     "localhost",
			Port:     5433,
			Database: "test_target",
			Username: "target_user",
			Password: expectedTargetPassword,
		},
		Schedule: ScheduleConfig{
			CronExpression: "0 2 * * *",
			Timezone:       "America/Santiago",
		},
		Storage: StorageConfig{
			LocalPath:     "/tmp/backups",
			RetentionDays: 7,
			Compression:   "gzip",
		},
	}

	// Save config
	err := manager.Save(config)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if !manager.Exists() {
		t.Error("Config file should exist after save")
	}

	// Load config
	loadedConfig, err := manager.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded config matches original
	if loadedConfig.Source.Host != config.Source.Host {
		t.Errorf("Source host mismatch: got %s, want %s", loadedConfig.Source.Host, config.Source.Host)
	}

	if loadedConfig.Source.Password != expectedSourcePassword {
		t.Errorf("Source password should be decrypted correctly: got %s, want %s", loadedConfig.Source.Password, expectedSourcePassword)
	}

	if loadedConfig.Target.Password != expectedTargetPassword {
		t.Errorf("Target password should be decrypted correctly: got %s, want %s", loadedConfig.Target.Password, expectedTargetPassword)
	}

	if loadedConfig.Schedule.CronExpression != config.Schedule.CronExpression {
		t.Errorf("Cron expression mismatch: got %s, want %s", loadedConfig.Schedule.CronExpression, config.Schedule.CronExpression)
	}
}

func TestManagerExists(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, nil)

	// Should not exist initially
	if manager.Exists() {
		t.Error("Config should not exist initially")
	}

	// Create the config file
	configPath := filepath.Join(tempDir, "config", ConfigFileName)
	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, []byte("test"), 0644)

	// Should exist now
	if !manager.Exists() {
		t.Error("Config should exist after creating file")
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"5432", 5432},
		{"80", 80},
		{"443", 443},
		{"", 5432},      // default
		{"abc", 5432},   // invalid, should return default
		{"-1", 5432},    // invalid port
		{"99999", 5432}, // port too high
	}

	for _, tt := range tests {
		result := parsePort(tt.input)
		if result != tt.expected {
			t.Errorf("parsePort(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"7", 7},
		{"30", 30},
		{"", 7},    // default
		{"-5", 7},  // negative, should return default
		{"abc", 7}, // invalid, should return default
	}

	for _, tt := range tests {
		result := parseInt(tt.input)
		if result != tt.expected {
			t.Errorf("parseInt(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}
