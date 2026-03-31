// Package pipeline generates backup pipeline scripts
package pipeline

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/wdelcant/backup-installer/internal/config"
)

// Generator creates pipeline scripts from templates
type Generator struct {
	baseDir string
}

// NewGenerator creates a new pipeline generator
func NewGenerator(baseDir string) *Generator {
	return &Generator{
		baseDir: baseDir,
	}
}

// PipelineData holds template data
type PipelineData struct {
	Version        string
	Generated      string
	ConfigPath     string
	BackupDir      string
	LogDir         string
	WebhookURL     string
	WebhookEnabled string

	SourceHost string
	SourcePort int
	SourceDB   string
	SourceUser string
	SourcePass string

	TargetEnabled string
	TargetHost    string
	TargetPort    int
	TargetDB      string
	TargetUser    string
	TargetPass    string
	RestoreDelay  int

	RetentionDays int
}

// Generate creates the pipeline script
func (g *Generator) Generate(cfg *config.Config) (string, error) {
	// Read template
	tmplPath := filepath.Join(g.baseDir, "internal", "pipeline", "templates", "pipeline.sh.tmpl")
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Parse template
	tmpl, err := template.New("pipeline").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare data
	data := PipelineData{
		Version:        "1.0.0",
		Generated:      time.Now().Format(time.RFC3339),
		ConfigPath:     filepath.Join(g.baseDir, "config", "config.yaml"),
		BackupDir:      cfg.Storage.LocalPath,
		LogDir:         filepath.Join(g.baseDir, "logs"),
		WebhookURL:     cfg.Webhook.URL,
		WebhookEnabled: "false",

		SourceHost: cfg.Source.Host,
		SourcePort: cfg.Source.Port,
		SourceDB:   cfg.Source.Database,
		SourceUser: cfg.Source.Username,
		SourcePass: cfg.Source.Password,

		TargetEnabled: "false",
		TargetHost:    cfg.Target.Host,
		TargetPort:    cfg.Target.Port,
		TargetDB:      cfg.Target.Database,
		TargetUser:    cfg.Target.Username,
		TargetPass:    cfg.Target.Password,
		RestoreDelay:  cfg.Target.RestoreDelayMins,

		RetentionDays: cfg.Storage.RetentionDays,
	}

	if cfg.Webhook.Enabled {
		data.WebhookEnabled = "true"
	}

	if cfg.Target.Enabled {
		data.TargetEnabled = "true"
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Write script
	scriptPath := filepath.Join(g.baseDir, "scripts", "pipeline.sh")
	scriptsDir := filepath.Dir(scriptPath)
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create scripts directory: %w", err)
	}

	if err := os.WriteFile(scriptPath, buf.Bytes(), 0755); err != nil {
		return "", fmt.Errorf("failed to write script: %w", err)
	}

	return scriptPath, nil
}
