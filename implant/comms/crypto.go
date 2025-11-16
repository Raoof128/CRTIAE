package comms

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// CryptoEngine handles all encryption/decryption operations for C2 communications
//
// BLUE TEAM DETECTION:
// - High entropy buffers in memory indicate encrypted payloads
// - ChaCha20 usage is uncommon in benign applications
// - Memory scanning for ChaCha20 constants can detect this implementation
// - Look for frequent calls to crypto libraries from unsigned processes
type CryptoEngine struct {
	symmetricKey []byte // 32-byte key for ChaCha20-Poly1305
	aead         cipher.AEAD
	rsaPublicKey *rsa.PublicKey
}

// NewCryptoEngine initializes the cryptographic engine with a symmetric key
// If no key provided, generates a random 32-byte key
func NewCryptoEngine(key []byte) (*CryptoEngine, error) {
	var symmetricKey []byte

	if len(key) == 0 {
		// Generate random 32-byte key for ChaCha20-Poly1305
		symmetricKey = make([]byte, chacha20poly1305.KeySize)
		if _, err := rand.Read(symmetricKey); err != nil {
			return nil, err
		}
	} else if len(key) != chacha20poly1305.KeySize {
		return nil, errors.New("invalid key size: must be 32 bytes for ChaCha20-Poly1305")
	} else {
		symmetricKey = key
	}

	// Create ChaCha20-Poly1305 AEAD cipher
	// Rationale: ChaCha20 provides high performance and resistance to timing attacks
	// Poly1305 provides authenticated encryption (integrity + confidentiality)
	aead, err := chacha20poly1305.NewX(symmetricKey)
	if err != nil {
		return nil, err
	}

	return &CryptoEngine{
		symmetricKey: symmetricKey,
		aead:         aead,
	}, nil
}

// SetRSAPublicKey configures RSA public key for asymmetric key exchange
// Used during initial beacon registration to securely exchange symmetric keys
func (c *CryptoEngine) SetRSAPublicKey(pemEncodedKey string) error {
	block, _ := pem.Decode([]byte(pemEncodedKey))
	if block == nil {
		return errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("not an RSA public key")
	}

	c.rsaPublicKey = rsaPub
	return nil
}

// Encrypt encrypts plaintext using ChaCha20-Poly1305 AEAD
// Returns base64-encoded ciphertext with prepended nonce
//
// Format: [24-byte nonce][ciphertext+tag]
//
// OPSEC Note: Each encryption uses a unique random nonce to prevent
// pattern analysis. However, this increases bandwidth slightly.
func (c *CryptoEngine) Encrypt(plaintext []byte) (string, error) {
	// Generate random nonce (24 bytes for XChaCha20-Poly1305)
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// Encrypt and authenticate
	// The Seal method appends the ciphertext and authentication tag
	ciphertext := c.aead.Seal(nil, nonce, plaintext, nil)

	// Prepend nonce to ciphertext for transmission
	// Receiver will extract nonce from first 24 bytes
	combined := append(nonce, ciphertext...)

	// Base64 encode for safe transmission over text protocols
	return base64.StdEncoding.EncodeToString(combined), nil
}

// Decrypt decrypts base64-encoded ciphertext using ChaCha20-Poly1305
// Expects format: [24-byte nonce][ciphertext+tag]
// Returns error if authentication tag verification fails (tampered data)
func (c *CryptoEngine) Decrypt(encodedCiphertext string) ([]byte, error) {
	// Decode base64
	combined, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, err
	}

	// Extract nonce and ciphertext
	nonceSize := c.aead.NonceSize()
	if len(combined) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce := combined[:nonceSize]
	ciphertext := combined[nonceSize:]

	// Decrypt and verify authentication tag
	// Open returns error if tag verification fails (data tampered)
	plaintext, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed: authentication tag mismatch")
	}

	return plaintext, nil
}

// EncryptWithRSA encrypts data using RSA-OAEP for key exchange
// Used only during initial handshake to transmit symmetric key
//
// OPSEC: RSA is slower, so we only use it for key exchange,
// then switch to ChaCha20 for all subsequent communications
func (c *CryptoEngine) EncryptWithRSA(plaintext []byte) (string, error) {
	if c.rsaPublicKey == nil {
		return "", errors.New("RSA public key not set")
	}

	// Encrypt using RSA-OAEP with SHA-256
	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		c.rsaPublicKey,
		plaintext,
		nil,
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// GenerateAESKey generates a random AES-256 key (fallback option)
// Included for compatibility with servers that prefer AES-GCM
func GenerateAESKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptAESGCM provides AES-GCM encryption as alternative to ChaCha20
// Some environments may have hardware-accelerated AES (AES-NI)
func EncryptAESGCM(plaintext, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	combined := append(nonce, ciphertext...)

	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptAESGCM decrypts AES-GCM encrypted data
func DecryptAESGCM(encodedCiphertext string, key []byte) ([]byte, error) {
	combined, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aead.NonceSize()
	if len(combined) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce := combined[:nonceSize]
	ciphertext := combined[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GetSymmetricKey returns the current symmetric key (for key rotation scenarios)
func (c *CryptoEngine) GetSymmetricKey() []byte {
	return c.symmetricKey
}

// RotateKey generates and sets a new symmetric key
// OPSEC: Periodic key rotation limits exposure if keys are compromised
func (c *CryptoEngine) RotateKey() error {
	newKey := make([]byte, chacha20poly1305.KeySize)
	if _, err := rand.Read(newKey); err != nil {
		return err
	}

	aead, err := chacha20poly1305.NewX(newKey)
	if err != nil {
		return err
	}

	c.symmetricKey = newKey
	c.aead = aead
	return nil
}
