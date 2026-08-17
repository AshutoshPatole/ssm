package security

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	password := "supersecret123"
	key := GenerateEncryptionKey(password)
	originalData := []byte("hello world! testing encryption and decryption")

	encrypted := EncryptData(originalData, key)
	if encrypted == "" {
		t.Fatalf("expected non-empty encrypted string")
	}

	decrypted, err := DecryptData(encrypted, key)
	if err != nil {
		t.Fatalf("unexpected decryption error: %v", err)
	}

	if !bytes.Equal(originalData, decrypted) {
		t.Fatalf("decrypted data does not match original: got %q, want %q", string(decrypted), string(originalData))
	}
}

func TestNonceRandomness(t *testing.T) {
	password := "supersecret123"
	key := GenerateEncryptionKey(password)
	data := []byte("identical payload")

	encrypted1 := EncryptData(data, key)
	encrypted2 := EncryptData(data, key)

	if encrypted1 == encrypted2 {
		t.Fatalf("two encryptions of identical payload produced identical ciphertext (static nonce detected)")
	}
}

func TestDecryptBoundsCheck(t *testing.T) {
	password := "supersecret123"
	key := GenerateEncryptionKey(password)

	// Short hex string (less than nonce length of 12 bytes = 24 hex chars)
	shortData := "01020304"
	_, err := DecryptData(shortData, key)
	if err == nil {
		t.Fatalf("expected error for short ciphertext, got nil")
	}
}
