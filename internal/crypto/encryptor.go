package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const (
	// EncryptedPrefix marks encrypted values
	EncryptedPrefix = "ENC[AES256_GCM]:"
)

// Encryptor handles AES-256-GCM encryption/decryption
type Encryptor struct {
	masterKey []byte
}

// NewEncryptor creates a new encryptor with the given master key
func NewEncryptor(masterKey []byte) *Encryptor {
	return &Encryptor{
		masterKey: masterKey,
	}
}

// Encrypt encrypts a plaintext string using AES-256-GCM
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return EncryptedPrefix + encoded, nil
}

// Decrypt decrypts an encrypted string
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Check if it's encrypted
	if !strings.HasPrefix(ciphertext, EncryptedPrefix) {
		// Not encrypted, return as-is
		return ciphertext, nil
	}

	// Remove prefix and decode
	encoded := strings.TrimPrefix(ciphertext, EncryptedPrefix)
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// IsEncrypted checks if a string is encrypted
func (e *Encryptor) IsEncrypted(value string) bool {
	return strings.HasPrefix(value, EncryptedPrefix)
}
