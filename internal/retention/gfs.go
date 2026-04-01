// Package retention implements GFS (Grandfather-Father-Son) retention policy
package retention

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/wdelcant/backup-installer/internal/config"
)

// GFSType represents the type of backup in GFS hierarchy
type GFSType int

const (
	// Son represents daily backups (kept for N days)
	Son GFSType = 1
	// Father represents weekly backups (kept for N weeks)
	Father GFSType = 2
	// Grandfather represents monthly backups (kept for N months)
	Grandfather GFSType = 3
)

// BackupFile represents a backup file with its metadata
type BackupFile struct {
	Path         string
	Name         string
	Size         int64
	ModTime      time.Time
	Type         GFSType
	DatabaseName string
}

// Manager handles GFS retention policy
type Manager struct {
	retention config.RetentionConfig
}

// NewManager creates a new GFS retention manager
func NewManager(retention config.RetentionConfig) *Manager {
	return &Manager{
		retention: retention,
	}
}

// ClassifyBackups classifies backup files according to GFS policy
func (m *Manager) ClassifyBackups(backupDir string) ([]BackupFile, []BackupFile, error) {
	// Find all backup files
	files, err := m.findBackupFiles(backupDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find backup files: %w", err)
	}

	// Classify each file
	for i := range files {
		files[i].Type = m.classifyBackup(files[i].ModTime)
	}

	// Determine which files to keep
	toKeep := m.applyPolicy(files)

	return files, toKeep, nil
}

// findBackupFiles finds all backup files in the directory
func (m *Manager) findBackupFiles(backupDir string) ([]BackupFile, error) {
	var files []BackupFile

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	// Regex to match backup files: database_YYYYMMDD_HHMMSS.sql.gz
	backupRegex := regexp.MustCompile(`^(\w+)_(\d{8})_(\d{6})\.sql\.gz$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := backupRegex.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, BackupFile{
			Path:         filepath.Join(backupDir, entry.Name()),
			Name:         entry.Name(),
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			DatabaseName: matches[1],
		})
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// classifyBackup determines the GFS type of a backup based on its date
func (m *Manager) classifyBackup(backupTime time.Time) GFSType {
	// Check if it's the first day of the month (Grandfather)
	if backupTime.Day() == 1 {
		return Grandfather
	}

	// Check if it's Sunday (Father)
	if backupTime.Weekday() == time.Sunday {
		return Father
	}

	// Otherwise it's a daily backup (Son)
	return Son
}

// applyPolicy applies GFS retention policy and returns files to keep
func (m *Manager) applyPolicy(files []BackupFile) []BackupFile {
	if !m.retention.Enabled {
		// If GFS is disabled, keep all files
		return files
	}

	var toKeep []BackupFile
	sonCount := 0
	fatherCount := 0
	grandfatherCount := 0

	now := time.Now()

	for _, file := range files {
		keep := false

		switch file.Type {
		case Son:
			// Keep daily backups within retention period
			if sonCount < m.retention.Son && m.isWithinRetention(file.ModTime, now, m.retention.Son) {
				sonCount++
				keep = true
			}
		case Father:
			// Keep weekly backups within retention period
			if fatherCount < m.retention.Father && m.isWithinRetention(file.ModTime, now, m.retention.Father*7) {
				fatherCount++
				keep = true
			}
		case Grandfather:
			// Keep monthly backups within retention period
			if grandfatherCount < m.retention.Grandfather && m.isWithinRetention(file.ModTime, now, m.retention.Grandfather*30) {
				grandfatherCount++
				keep = true
			}
		}

		if keep {
			toKeep = append(toKeep, file)
		}
	}

	return toKeep
}

// isWithinRetention checks if a backup is within the retention period
func (m *Manager) isWithinRetention(backupTime, now time.Time, days int) bool {
	retentionCutoff := now.AddDate(0, 0, -days)
	return backupTime.After(retentionCutoff) || backupTime.Equal(retentionCutoff)
}

// Cleanup removes old backups according to GFS policy
func (m *Manager) Cleanup(backupDir string) error {
	files, toKeep, err := m.ClassifyBackups(backupDir)
	if err != nil {
		return err
	}

	// Create a set of files to keep
	keepSet := make(map[string]bool)
	for _, file := range toKeep {
		keepSet[file.Path] = true
	}

	// Remove files not in the keep set
	var removed int
	var freedBytes int64

	for _, file := range files {
		if !keepSet[file.Path] {
			if err := os.Remove(file.Path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", file.Name, err)
			}
			removed++
			freedBytes += file.Size
		}
	}

	if removed > 0 {
		fmt.Printf("🗑️  Removed %d old backup(s), freed %s\n", removed, formatBytes(freedBytes))
	}

	return nil
}

// GetStatistics returns statistics about backups
func (m *Manager) GetStatistics(backupDir string) (*Statistics, error) {
	files, _, err := m.ClassifyBackups(backupDir)
	if err != nil {
		return nil, err
	}

	stats := &Statistics{
		TotalFiles:       len(files),
		TotalSize:        0,
		SonCount:         0,
		FatherCount:      0,
		GrandfatherCount: 0,
	}

	for _, file := range files {
		stats.TotalSize += file.Size
		switch file.Type {
		case Son:
			stats.SonCount++
		case Father:
			stats.FatherCount++
		case Grandfather:
			stats.GrandfatherCount++
		}
	}

	return stats, nil
}

// Statistics holds backup statistics
type Statistics struct {
	TotalFiles       int
	TotalSize        int64
	SonCount         int
	FatherCount      int
	GrandfatherCount int
}

// formatBytes formats bytes to human-readable string
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
