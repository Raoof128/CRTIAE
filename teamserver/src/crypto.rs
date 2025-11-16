// Cryptography module for teamserver
//
// Mirrors the implant's crypto functionality for symmetric encryption

use chacha20poly1305::{
    aead::{Aead, KeyInit},
    XChaCha20Poly1305, XNonce,
};
use base64::{Engine as _, engine::general_purpose};
use rand::Rng;

/// Generate a random 32-byte encryption key
pub fn generate_key() -> Vec<u8> {
    let mut rng = rand::thread_rng();
    (0..32).map(|_| rng.gen::<u8>()).collect()
}

/// Encrypt data using ChaCha20-Poly1305
/// Returns base64-encoded string with prepended nonce
pub fn encrypt(key: &[u8], plaintext: &[u8]) -> Result<String, String> {
    if key.len() != 32 {
        return Err("Key must be 32 bytes".to_string());
    }

    let cipher = XChaCha20Poly1305::new(key.into());

    // Generate random nonce
    let mut rng = rand::thread_rng();
    let nonce_bytes: [u8; 24] = rng.gen();
    let nonce = XNonce::from_slice(&nonce_bytes);

    // Encrypt
    let ciphertext = cipher
        .encrypt(nonce, plaintext)
        .map_err(|e| format!("Encryption failed: {:?}", e))?;

    // Prepend nonce to ciphertext
    let mut combined = nonce_bytes.to_vec();
    combined.extend_from_slice(&ciphertext);

    // Base64 encode
    Ok(general_purpose::STANDARD.encode(combined))
}

/// Decrypt base64-encoded ciphertext
pub fn decrypt(key: &[u8], encoded_ciphertext: &str) -> Result<Vec<u8>, String> {
    if key.len() != 32 {
        return Err("Key must be 32 bytes".to_string());
    }

    // Decode base64
    let combined = general_purpose::STANDARD.decode(encoded_ciphertext)
        .map_err(|e| format!("Base64 decode failed: {}", e))?;

    if combined.len() < 24 {
        return Err("Ciphertext too short".to_string());
    }

    // Extract nonce and ciphertext
    let (nonce_bytes, ciphertext) = combined.split_at(24);
    let nonce = XNonce::from_slice(nonce_bytes);

    // Decrypt
    let cipher = XChaCha20Poly1305::new(key.into());
    let plaintext = cipher
        .decrypt(nonce, ciphertext)
        .map_err(|e| format!("Decryption failed: {:?}", e))?;

    Ok(plaintext)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_encrypt_decrypt() {
        let key = generate_key();
        let plaintext = b"Hello, World!";

        let encrypted = encrypt(&key, plaintext).unwrap();
        let decrypted = decrypt(&key, &encrypted).unwrap();

        assert_eq!(plaintext.to_vec(), decrypted);
    }
}
