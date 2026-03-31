// Package config handles configuration management with encryption support
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wdelcant/backup-installer/internal/crypto"
	"gopkg.in/yaml.v3"
)

const (
	// ConfigFileName is the name of the configuration file
	ConfigFileName = "config.yaml"
	// DefaultRetentionDays is the default number of days to keep backups
	DefaultRetentionDays = 7
	// DefaultBackupPath is the default local backup directory
	DefaultBackupPath = "/opt/invitsm/backups"
)

// Config represents the complete backup configuration
type Config struct {
	Version   string    `yaml:"version" json:"version"`
	Generated time.Time `yaml:"generated" json:"generated"`

	Source   DatabaseConfig `yaml:"source" json:"source"`
	Target   TargetConfig   `yaml:"target" json:"target"`
	Schedule ScheduleConfig `yaml:"schedule" json:"schedule"`
	Storage  StorageConfig  `yaml:"storage" json:"storage"`
	Webhook  WebhookConfig  `yaml:"webhook" json:"webhook"`
	Logging  LoggingConfig  `yaml:"logging" json:"logging"`
}

// DatabaseConfig holds source database connection details
type DatabaseConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"-"` // Sensitive, will be encrypted
}

// TargetConfig holds target database configuration for restore
type TargetConfig struct {
	Enabled          bool   `yaml:"enabled" json:"enabled"`
	Host             string `yaml:"host" json:"host"`
	Port             int    `yaml:"port" json:"port"`
	Database         string `yaml:"database" json:"database"`
	Username         string `yaml:"username" json:"username"`
	Password         string `yaml:"password" json:"-"` // Sensitive
	RestoreDelayMins int    `yaml:"restore_delay_minutes" json:"restore_delay_minutes"`
	SafetyBackup     bool   `yaml:"create_safety_backup" json:"create_safety_backup"`
}

// ScheduleConfig holds backup schedule configuration
type ScheduleConfig struct {
	CronExpression string `yaml:"cron_expression" json:"cron_expression"`
	Timezone       string `yaml:"timezone" json:"timezone"`
	NextRun        string `yaml:"next_run" json:"next_run"`
}

// StorageConfig holds backup storage configuration
type StorageConfig struct {
	LocalPath              string   `yaml:"local_path" json:"local_path"`
	RetentionDays          int      `yaml:"retention_days" json:"retention_days"`
	Compression            string   `yaml:"compression" json:"compression"`
	AdditionalDestinations []string `yaml:"additional_destinations" json:"additional_destinations"`
}

// WebhookConfig holds n8n webhook configuration
type WebhookConfig struct {
	Enabled bool              `yaml:"enabled" json:"enabled"`
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Events  []string          `yaml:"events" json:"events"`
	Timeout int               `yaml:"timeout_seconds" json:"timeout_seconds"`
	Retries int               `yaml:"retry_attempts" json:"retry_attempts"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level     string `yaml:"level" json:"level"`
	Path      string `yaml:"path" json:"path"`
	Rotation  string `yaml:"rotation" json:"rotation"`
	MaxSizeMB int    `yaml:"max_size_mb" json:"max_size_mb"`
}

// Manager handles configuration loading and saving with encryption
type Manager struct {
	configPath string
	configDir  string
	encryptor  *crypto.Encryptor
}

// NewManager creates a new configuration manager
func NewManager(baseDir string, encryptor *crypto.Encryptor) *Manager {
	configDir := filepath.Join(baseDir, "config")
	return &Manager{
		configPath: filepath.Join(configDir, ConfigFileName),
		configDir:  configDir,
		encryptor:  encryptor,
	}
}

// Load reads and decrypts the configuration
func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Decrypt sensitive fields
	if err := m.decryptConfig(&config); err != nil {
		return nil, fmt.Errorf("failed to decrypt config: %w", err)
	}

	return &config, nil
}

// Save encrypts and writes the configuration
func (m *Manager) Save(config *Config) error {
	// Create directory if needed
	if err := os.MkdirAll(m.configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Encrypt sensitive fields
	if err := m.encryptConfig(config); err != nil {
		return fmt.Errorf("failed to encrypt config: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with secure permissions
	if err := os.WriteFile(m.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// encryptConfig encrypts all sensitive fields
func (m *Manager) encryptConfig(config *Config) error {
	if m.encryptor == nil {
		return nil
	}

	var err error
	config.Source.Password, err = m.encryptor.Encrypt(config.Source.Password)
	if err != nil {
		return err
	}

	if config.Target.Enabled {
		config.Target.Password, err = m.encryptor.Encrypt(config.Target.Password)
		if err != nil {
			return err
		}
	}

	if config.Webhook.Enabled {
		if token, ok := config.Webhook.Headers["Authorization"]; ok {
			encrypted, err := m.encryptor.Encrypt(token)
			if err != nil {
				return err
			}
			config.Webhook.Headers["Authorization"] = encrypted
		}
	}

	return nil
}

// decryptConfig decrypts all sensitive fields
func (m *Manager) decryptConfig(config *Config) error {
	if m.encryptor == nil {
		return nil
	}

	var err error
	config.Source.Password, err = m.encryptor.Decrypt(config.Source.Password)
	if err != nil {
		return err
	}

	if config.Target.Enabled {
		config.Target.Password, err = m.encryptor.Decrypt(config.Target.Password)
		if err != nil {
			return err
		}
	}

	if config.Webhook.Enabled {
		if token, ok := config.Webhook.Headers["Authorization"]; ok {
			decrypted, err := m.encryptor.Decrypt(token)
			if err != nil {
				return err
			}
			config.Webhook.Headers["Authorization"] = decrypted
		}
	}

	return nil
}

// Exists checks if the configuration file exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.configPath)
	return !os.IsNotExist(err)
}

// Delete removes the configuration file
func (m *Manager) Delete() error {
	if err := os.Remove(m.configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete config: %w", err)
	}
	return nil
}

// GetConfigPath returns the path to the configuration file
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Version:   "1.0",
		Generated: time.Now(),
		Source: DatabaseConfig{
			Port: 5432,
		},
		Target: TargetConfig{
			Enabled:          false,
			Port:             5432,
			RestoreDelayMins: 30,
			SafetyBackup:     true,
		},
		Schedule: ScheduleConfig{
			CronExpression: "0 2 * * *",
			Timezone:       "America/Santiago",
		},
		Storage: StorageConfig{
			LocalPath:     DefaultBackupPath,
			RetentionDays: DefaultRetentionDays,
			Compression:   "gzip",
		},
		Webhook: WebhookConfig{
			Enabled: false,
			Headers: make(map[string]string),
			Events:  []string{"pipeline_completed", "pipeline_failed"},
			Timeout: 30,
			Retries: 3,
		},
		Logging: LoggingConfig{
			Level:     "info",
			Path:      "./logs",
			Rotation:  "daily",
			MaxSizeMB: 100,
		},
	}
}

// ToJSON converts config to JSON (for webhook payloads)
func (c *Config) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}
