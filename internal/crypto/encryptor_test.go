package crypto

import (
	"testing"
)

func TestEncryptor(t *testing.T) {
	// Generate a test key
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	encryptor := NewEncryptor(key)

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple text",
			plaintext: "hello world",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "special characters",
			plaintext: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "unicode text",
			plaintext: "Hello 世界 🌍 ñáéíóú",
		},
		{
			name:      "long text",
			plaintext: "This is a very long text that should still work correctly with the encryption and decryption process without any issues whatsoever",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := encryptor.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Verify ciphertext is different from plaintext
			if ciphertext == tt.plaintext && tt.plaintext != "" {
				t.Error("Ciphertext should be different from plaintext")
			}

			// Decrypt
			decrypted, err := encryptor.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// Verify decrypted matches original
			if decrypted != tt.plaintext {
				t.Errorf("Decrypted text doesn't match: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptorDifferentKeys(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 1 // Different key

	encryptor1 := NewEncryptor(key1)
	encryptor2 := NewEncryptor(key2)

	plaintext := "test message"
	ciphertext, err := encryptor1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Try to decrypt with different key
	_, err = encryptor2.Decrypt(ciphertext)
	if err == nil {
		t.Error("Decrypting with wrong key should fail")
	}
}
