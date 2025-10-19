package skylib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
)

// Hash functions

// MD5 computes MD5 hash
func CryptoMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// SHA1 computes SHA1 hash
func CryptoSHA1(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}

// SHA256 computes SHA256 hash
func CryptoSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// SHA512 computes SHA512 hash
func CryptoSHA512(data []byte) string {
	hash := sha512.Sum512(data)
	return hex.EncodeToString(hash[:])
}

// AES-GCM encryption

// AESGCMEncrypt encrypts data with AES-GCM
func CryptoAESGCMEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	// TODO: Fill nonce with random data

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// AESGCMDecrypt decrypts data with AES-GCM
func CryptoAESGCMDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// HMAC functions

// HMACSHA256 computes HMAC-SHA256
func CryptoHMACSHA256(key, message []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

// HMACSHA512 computes HMAC-SHA512
func CryptoHMACSHA512(key, message []byte) string {
	h := hmac.New(sha512.New, key)
	h.Write(message)
	return hex.EncodeToString(h.Sum(nil))
}

// ChaCha20-Poly1305 encryption

// ChaCha20Encrypt encrypts with ChaCha20-Poly1305
func CryptoChaCha20Encrypt(key, plaintext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	// TODO: Fill nonce with crypto random

	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// ChaCha20Decrypt decrypts with ChaCha20-Poly1305
func CryptoChaCha20Decrypt(key, ciphertext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aead.Open(nil, nonce, ciphertext, nil)
}

// Key derivation functions

// PBKDF2 derives a key using PBKDF2
func CryptoPBKDF2(password, salt []byte, iterations, keyLen int) []byte {
	return pbkdf2.Key(password, salt, iterations, keyLen, sha256.New)
}

// PBKDF2SHA512 derives a key using PBKDF2-SHA512
func CryptoPBKDF2SHA512(password, salt []byte, iterations, keyLen int) []byte {
	return pbkdf2.Key(password, salt, iterations, keyLen, sha512.New)
}
