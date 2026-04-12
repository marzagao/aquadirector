package tuya

import (
	"bytes"
	"testing"
)

func TestGenerateNonce(t *testing.T) {
	n1, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce: %v", err)
	}
	n2, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce: %v", err)
	}
	if n1 == n2 {
		t.Error("two nonces should not be identical")
	}
	if n1 == [16]byte{} {
		t.Error("nonce should not be all zeros")
	}
}

func TestComputeAndVerifyHMAC(t *testing.T) {
	key := []byte("0123456789abcdef")
	data := []byte("test message")

	mac := ComputeHMAC(key, data)
	if len(mac) != 32 {
		t.Fatalf("HMAC length = %d, want 32", len(mac))
	}

	if !VerifyHMAC(key, data, mac) {
		t.Error("VerifyHMAC should return true for correct HMAC")
	}

	if VerifyHMAC(key, []byte("wrong message"), mac) {
		t.Error("VerifyHMAC should return false for wrong data")
	}

	if VerifyHMAC([]byte("wrongkey12345678"), data, mac) {
		t.Error("VerifyHMAC should return false for wrong key")
	}
}

func TestGCMEncryptDecrypt(t *testing.T) {
	key := []byte("0123456789abcdef") // 16 bytes = AES-128
	nonce := make([]byte, 12)
	for i := range nonce {
		nonce[i] = byte(i)
	}
	plaintext := []byte(`{"dps":{"1":100,"2":263}}`)
	aad := []byte{0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00, 0x24}

	encrypted, err := GCMEncrypt(key, nonce, plaintext, aad)
	if err != nil {
		t.Fatalf("GCMEncrypt: %v", err)
	}

	// encrypted should be len(plaintext) + 16 (tag)
	if len(encrypted) != len(plaintext)+16 {
		t.Fatalf("encrypted length = %d, want %d", len(encrypted), len(plaintext)+16)
	}

	decrypted, err := GCMDecrypt(key, nonce, encrypted, aad)
	if err != nil {
		t.Fatalf("GCMDecrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestGCMDecryptTampered(t *testing.T) {
	key := []byte("0123456789abcdef")
	nonce := make([]byte, 12)
	plaintext := []byte("hello")

	encrypted, err := GCMEncrypt(key, nonce, plaintext, nil)
	if err != nil {
		t.Fatalf("GCMEncrypt: %v", err)
	}

	// Tamper with ciphertext
	encrypted[0] ^= 0xFF

	_, err = GCMDecrypt(key, nonce, encrypted, nil)
	if err == nil {
		t.Error("expected error for tampered ciphertext")
	}
}

func TestDeriveSessionKey(t *testing.T) {
	localKey := []byte("0123456789abcdef")
	clientNonce := []byte("abcdefghijklmnop")
	deviceNonce := []byte("ABCDEFGHIJKLMNOP")

	key, err := DeriveSessionKey(localKey, clientNonce, deviceNonce)
	if err != nil {
		t.Fatalf("DeriveSessionKey: %v", err)
	}

	if len(key) != 16 {
		t.Fatalf("session key length = %d, want 16", len(key))
	}

	// Same inputs should produce same key (deterministic)
	key2, err := DeriveSessionKey(localKey, clientNonce, deviceNonce)
	if err != nil {
		t.Fatalf("DeriveSessionKey 2: %v", err)
	}
	if !bytes.Equal(key, key2) {
		t.Error("same inputs should produce same session key")
	}

	// Different nonces should produce different key
	key3, err := DeriveSessionKey(localKey, deviceNonce, clientNonce)
	if err != nil {
		t.Fatalf("DeriveSessionKey 3: %v", err)
	}
	if bytes.Equal(key, key3) {
		t.Error("different nonces should produce different session key")
	}
}

func TestDeriveSessionKeyInvalidInputs(t *testing.T) {
	_, err := DeriveSessionKey([]byte("short"), []byte("0123456789abcdef"), []byte("0123456789abcdef"))
	if err == nil {
		t.Error("expected error for short local key")
	}
}

func TestXorBytes(t *testing.T) {
	a := []byte{0x01, 0x02, 0x03, 0x04}
	b := []byte{0x10, 0x20, 0x30, 0x40}
	got := xorBytes(a, b)
	want := []byte{0x11, 0x22, 0x33, 0x44}
	if !bytes.Equal(got, want) {
		t.Errorf("xorBytes = %v, want %v", got, want)
	}
}
