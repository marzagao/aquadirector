package tuya

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

func GenerateNonce() ([16]byte, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nonce, fmt.Errorf("generating nonce: %w", err)
	}
	return nonce, nil
}

func ComputeHMAC(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func VerifyHMAC(key, data, expected []byte) bool {
	computed := ComputeHMAC(key, data)
	return hmac.Equal(computed, expected)
}

// DeriveSessionKey derives the v3.5 session key from the local key and
// the client/device nonces exchanged during handshake.
//
// Algorithm:
//
//	tmpKey = XOR(clientNonce[:16], deviceNonce[:16])
//	sealed = AES-GCM-Seal(key=localKey, nonce=clientNonce[:12], plaintext=tmpKey)
//	sessionKey = sealed[:16]
func DeriveSessionKey(localKey, clientNonce, deviceNonce []byte) ([]byte, error) {
	if len(localKey) != 16 || len(clientNonce) < 16 || len(deviceNonce) < 16 {
		return nil, &CryptoError{Op: "derive_session_key", Err: fmt.Errorf("invalid input lengths: key=%d client=%d device=%d", len(localKey), len(clientNonce), len(deviceNonce))}
	}

	tmpKey := xorBytes(clientNonce[:16], deviceNonce[:16])

	block, err := aes.NewCipher(localKey)
	if err != nil {
		return nil, &CryptoError{Op: "derive_session_key", Err: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &CryptoError{Op: "derive_session_key", Err: err}
	}

	// Seal appends ciphertext+tag to dst. With 16-byte plaintext,
	// sealed = ciphertext(16) + tag(16) = 32 bytes.
	sealed := gcm.Seal(nil, clientNonce[:12], tmpKey, nil)
	if len(sealed) < 16 {
		return nil, &CryptoError{Op: "derive_session_key", Err: fmt.Errorf("sealed output too short: %d", len(sealed))}
	}

	sessionKey := make([]byte, 16)
	copy(sessionKey, sealed[:16])
	return sessionKey, nil
}

// GCMEncrypt encrypts plaintext with AES-GCM. Returns ciphertext||tag concatenated.
func GCMEncrypt(key, nonce, plaintext, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, &CryptoError{Op: "gcm_encrypt", Err: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &CryptoError{Op: "gcm_encrypt", Err: err}
	}

	return gcm.Seal(nil, nonce, plaintext, aad), nil
}

// GCMDecrypt decrypts ciphertext||tag with AES-GCM. The ciphertext parameter
// must include the 16-byte authentication tag appended.
func GCMDecrypt(key, nonce, ciphertextWithTag, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, &CryptoError{Op: "gcm_decrypt", Err: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &CryptoError{Op: "gcm_decrypt", Err: err}
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertextWithTag, aad)
	if err != nil {
		return nil, &CryptoError{Op: "gcm_decrypt", Err: err}
	}

	return plaintext, nil
}

func xorBytes(a, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}
