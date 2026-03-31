// Package crypto provides encryption utilities for secure credential storage
package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// KeyFileName is the name of the master key file
	KeyFileName = ".invitsm-master-key"
	// KeySize is 32 bytes (256 bits) for AES-256
	KeySize = 32
)

// MasterKeyManager handles the master encryption key
type MasterKeyManager struct {
	keyPath string
}

// NewMasterKeyManager creates a new master key manager
func NewMasterKeyManager(configDir string) *MasterKeyManager {
	return &MasterKeyManager{
		keyPath: filepath.Join(configDir, KeyFileName),
	}
}

// GetOrCreateMasterKey retrieves or generates the master key
func (m *MasterKeyManager) GetOrCreateMasterKey() ([]byte, error) {
	// Check if key already exists
	if _, err := os.Stat(m.keyPath); err == nil {
		return m.loadKey()
	}

	// Generate new random key
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	// Save with secure permissions
	if err := m.saveKey(key); err != nil {
		return nil, err
	}

	return key, nil
}

// loadKey reads the master key from disk
func (m *MasterKeyManager) loadKey() ([]byte, error) {
	data, err := os.ReadFile(m.keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read master key: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode master key: %w", err)
	}

	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid master key size: expected %d, got %d", KeySize, len(key))
	}

	return key, nil
}

// saveKey writes the master key with secure permissions (0400)
func (m *MasterKeyManager) saveKey(key []byte) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(m.keyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Encode and save with restricted permissions
	encoded := base64.StdEncoding.EncodeToString(key)
	if err := os.WriteFile(m.keyPath, []byte(encoded), 0400); err != nil {
		return fmt.Errorf("failed to save master key: %w", err)
	}

	return nil
}

// KeyExists checks if the master key file exists
func (m *MasterKeyManager) KeyExists() bool {
	_, err := os.Stat(m.keyPath)
	return !os.IsNotExist(err)
}

// GetKeyPath returns the path to the master key file
func (m *MasterKeyManager) GetKeyPath() string {
	return m.keyPath
}

// DeleteKey removes the master key (use with caution)
func (m *MasterKeyManager) DeleteKey() error {
	if err := os.Remove(m.keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete master key: %w", err)
	}
	return nil
}
