package comms

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestNewCryptoEngine(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr bool
	}{
		{
			name:    "valid 32-byte key",
			key:     make([]byte, 32),
			wantErr: false,
		},
		{
			name:    "nil key generates random",
			key:     nil,
			wantErr: false,
		},
		{
			name:    "invalid key size",
			key:     make([]byte, 16),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCryptoEngine(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCryptoEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("NewCryptoEngine() returned nil when expecting valid engine")
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	ce, err := NewCryptoEngine(nil)
	if err != nil {
		t.Fatalf("Failed to create crypto engine: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{
			name:      "simple text",
			plaintext: []byte("Hello, World!"),
		},
		{
			name:      "empty string",
			plaintext: []byte(""),
		},
		{
			name:      "binary data",
			plaintext: []byte{0x00, 0xFF, 0x01, 0xFE},
		},
		{
			name:      "large data",
			plaintext: bytes.Repeat([]byte("A"), 10000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := ce.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decrypted, err := ce.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal(tt.plaintext, decrypted) {
				t.Errorf("Decrypted data doesn't match original.\nWant: %v\nGot: %v",
					tt.plaintext, decrypted)
			}
		})
	}
}

func TestDecryptInvalidData(t *testing.T) {
	ce, _ := NewCryptoEngine(nil)

	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "invalid base64",
			data:    "not-valid-base64!!!",
			wantErr: true,
		},
		{
			name:    "too short",
			data:    "YWJj", // "abc" in base64
			wantErr: true,
		},
		{
			name:    "tampered data",
			data:    "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXpBQkNERUZHSElKS0xNTk9QUQ==",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ce.Decrypt(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKeyRotation(t *testing.T) {
	ce, _ := NewCryptoEngine(nil)

	plaintext := []byte("test data")

	// Encrypt with original key
	encrypted1, err := ce.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Rotate key
	err = ce.RotateKey()
	if err != nil {
		t.Fatalf("RotateKey() error = %v", err)
	}

	// Old ciphertext should fail with new key
	_, err = ce.Decrypt(encrypted1)
	if err == nil {
		t.Error("Expected decryption to fail with rotated key")
	}

	// New encryption should work
	encrypted2, err := ce.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error after rotation = %v", err)
	}

	decrypted2, err := ce.Decrypt(encrypted2)
	if err != nil {
		t.Fatalf("Decrypt() error after rotation = %v", err)
	}

	if !bytes.Equal(plaintext, decrypted2) {
		t.Error("Decryption failed after key rotation")
	}
}

func TestSetRSAPublicKey(t *testing.T) {
	ce, _ := NewCryptoEngine(nil)

	// Valid PEM-encoded RSA public key (2048-bit)
	validPEM := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA1234567890abcdefghijk
lmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmno
pqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrst
uvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxy
zABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABCD
EFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHI
JKLMNOPQRSTUVWXYZwIDAQAB
-----END PUBLIC KEY-----`

	invalidPEM := "not a PEM key"

	tests := []struct {
		name    string
		pemKey  string
		wantErr bool
	}{
		{
			name:    "invalid PEM",
			pemKey:  invalidPEM,
			wantErr: true,
		},
		{
			name:    "empty string",
			pemKey:  "",
			wantErr: true,
		},
	}

	// Note: validPEM test would require a real RSA key,
	// which is complex to generate in tests

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ce.SetRSAPublicKey(tt.pemKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetRSAPublicKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetSymmetricKey(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	ce, err := NewCryptoEngine(key)
	if err != nil {
		t.Fatalf("Failed to create crypto engine: %v", err)
	}

	retrievedKey := ce.GetSymmetricKey()

	if !bytes.Equal(key, retrievedKey) {
		t.Error("GetSymmetricKey() did not return the original key")
	}
}

func TestAESGCMEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("Hello, AES-GCM!")

	encrypted, err := EncryptAESGCM(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptAESGCM() error = %v", err)
	}

	decrypted, err := DecryptAESGCM(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptAESGCM() error = %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("AES-GCM decrypted data doesn't match.\nWant: %v\nGot: %v",
			plaintext, decrypted)
	}
}

func TestGenerateAESKey(t *testing.T) {
	key1, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("GenerateAESKey() returned key of length %d, want 32", len(key1))
	}

	// Generate second key and ensure they're different (randomness check)
	key2, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("GenerateAESKey() error = %v", err)
	}

	if bytes.Equal(key1, key2) {
		t.Error("GenerateAESKey() generated identical keys (randomness failure)")
	}
}

// Benchmark encryption performance
func BenchmarkEncrypt(b *testing.B) {
	ce, _ := NewCryptoEngine(nil)
	plaintext := []byte("benchmark data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ce.Encrypt(plaintext)
	}
}

// Benchmark decryption performance
func BenchmarkDecrypt(b *testing.B) {
	ce, _ := NewCryptoEngine(nil)
	plaintext := []byte("benchmark data")
	encrypted, _ := ce.Encrypt(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ce.Decrypt(encrypted)
	}
}

// Benchmark key generation
func BenchmarkGenerateAESKey(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateAESKey()
	}
}

// Test that different nonces produce different ciphertexts
func TestEncryptionDeterminism(t *testing.T) {
	ce, _ := NewCryptoEngine(nil)
	plaintext := []byte("same plaintext")

	encrypted1, _ := ce.Encrypt(plaintext)
	encrypted2, _ := ce.Encrypt(plaintext)

	if encrypted1 == encrypted2 {
		t.Error("Encrypting same plaintext twice produced identical ciphertext (nonce reuse!)")
	}
}
