// Package version provides version checking and update functionality
package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	repoOwner = "wdelcant"
	repoName  = "backup-installer"
)

// ReleaseInfo holds information about the latest release
type ReleaseInfo struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

// Asset represents a release asset from GitHub API
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ReleaseInfoWithAssets extends ReleaseInfo with assets
type ReleaseInfoWithAssets struct {
	ReleaseInfo
	Assets []Asset `json:"assets"`
}

// Checker handles version checking
type Checker struct {
	currentVersion string
	httpClient     *http.Client
}

// NewChecker creates a new version checker
func NewChecker(currentVersion string) *Checker {
	return &Checker{
		currentVersion: currentVersion,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetLatestRelease fetches the latest release from GitHub
func (c *Checker) GetLatestRelease() (*ReleaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("backup-installer/%s", c.currentVersion))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// GetLatestReleaseWithAssets fetches the latest release with assets
func (c *Checker) GetLatestReleaseWithAssets() (*ReleaseInfoWithAssets, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("backup-installer/%s", c.currentVersion))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release ReleaseInfoWithAssets
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// IsUpdateAvailable checks if a newer version is available
func (c *Checker) IsUpdateAvailable() (bool, *ReleaseInfo, error) {
	// Skip check for dev versions
	if c.currentVersion == "dev" {
		return false, nil, nil
	}

	release, err := c.GetLatestRelease()
	if err != nil {
		return false, nil, err
	}

	latestVersion := normalizeVersion(release.TagName)
	currentVersion := normalizeVersion(c.currentVersion)

	// Simple string comparison (assumes semantic versioning)
	if latestVersion > currentVersion {
		return true, release, nil
	}

	return false, nil, nil
}

// normalizeVersion removes the 'v' prefix for comparison
func normalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

// RenderUpdateNotification renders a styled update notification
func RenderUpdateNotification(currentVersion, latestVersion string) string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)

	b.WriteString(titleStyle.Render("⚠️  Nueva versión disponible!"))
	b.WriteString("\n\n")

	// Version comparison
	currentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	arrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)

	latestStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true)

	b.WriteString(fmt.Sprintf("%s %s %s\n\n",
		currentStyle.Render(fmt.Sprintf("v%s", currentVersion)),
		arrowStyle.Render("→"),
		latestStyle.Render(latestVersion)))

	// Update command
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00D4AA")).
		Padding(1, 2)

	cmdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E5E7EB")).
		Bold(true)

	cmd := fmt.Sprintf("backup-installer --update\n\n# O manualmente:\ncurl -fsSL https://raw.githubusercontent.com/%s/%s/main/install.sh | bash",
		repoOwner, repoName)

	b.WriteString(boxStyle.Render(cmdStyle.Render(cmd)))

	return b.String()
}

// SelfUpdate performs a self-update (downloads and replaces the binary)
func SelfUpdate() error {
	// Get latest release with assets
	checker := NewChecker("")
	release, err := checker.GetLatestReleaseWithAssets()
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	// Determine binary name based on OS and arch
	binaryName := fmt.Sprintf("backup-installer-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	// Find the binary asset
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the binary
	fmt.Printf("Descargando %s...\n", binaryName)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create temporary file
	tmpFile := execPath + ".tmp"
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Copy content
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpFile)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	f.Close()

	// Make executable
	if err := os.Chmod(tmpFile, 0755); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to make executable: %w", err)
	}

	// Replace old binary
	if err := os.Rename(tmpFile, execPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("✅ Actualizado a %s exitosamente!\n", release.TagName)
	return nil
}
