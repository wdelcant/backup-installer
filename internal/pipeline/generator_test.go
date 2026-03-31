package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wdelcant/backup-installer/internal/config"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator("/tmp/test")
	if g == nil {
		t.Error("NewGenerator should not return nil")
	}
	if g.baseDir != "/tmp/test" {
		t.Errorf("Expected baseDir to be /tmp/test, got %s", g.baseDir)
	}
}

func TestGeneratorWithTemplate(t *testing.T) {
	// Create temp directory with template
	tempDir := t.TempDir()
	templateDir := filepath.Join(tempDir, "internal", "pipeline", "templates")
	os.MkdirAll(templateDir, 0755)

	// Create a simple test template
	templateContent := `#!/bin/bash
# Test pipeline script
# Source: {{.SourceHost}}:{{.SourcePort}}/{{.SourceDB}}
# Target: {{.TargetHost}}:{{.TargetPort}}/{{.TargetDB}}
# Backup dir: {{.BackupDir}}

echo "Backup started"
`
	templatePath := filepath.Join(templateDir, "pipeline.sh.tmpl")
	os.WriteFile(templatePath, []byte(templateContent), 0644)

	// Create generator
	g := NewGenerator(tempDir)

	// Test config
	cfg := &config.Config{
		Source: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "test_db",
			Username: "test_user",
			Password: "test_pass",
		},
		Target: config.TargetConfig{
			Enabled:          true,
			Host:             "qa.example.com",
			Port:             5432,
			Database:         "qa_db",
			Username:         "qa_user",
			Password:         "qa_pass",
			RestoreDelayMins: 30,
		},
		Storage: config.StorageConfig{
			LocalPath:     "/tmp/backups",
			RetentionDays: 7,
		},
		Webhook: config.WebhookConfig{
			Enabled: false,
		},
	}

	// Generate script
	scriptPath, err := g.Generate(cfg)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Read generated script
	scriptBytes, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read generated script: %v", err)
	}
	script := string(scriptBytes)

	// Verify script content
	if script == "" {
		t.Error("Generated script should not be empty")
	}

	// Check for bash shebang
	if len(script) < 11 || script[:11] != "#!/bin/bash" {
		t.Error("Script should start with #!/bin/bash")
	}

	// Check that template variables were replaced
	if !contains(script, "localhost") {
		t.Error("Script should contain source host")
	}

	if !contains(script, "test_db") {
		t.Error("Script should contain source database")
	}

	if !contains(script, "/tmp/backups") {
		t.Error("Script should contain backup directory")
	}
}

func TestGeneratorMissingTemplate(t *testing.T) {
	// Create temp directory without template
	tempDir := t.TempDir()

	g := NewGenerator(tempDir)
	cfg := &config.Config{
		Source: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "test",
		},
		Storage: config.StorageConfig{
			LocalPath: "/tmp/backups",
		},
	}

	_, err := g.Generate(cfg)
	if err == nil {
		t.Error("Generate should fail when template is missing")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsInternal(s, substr))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
