package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"log"
)

func GenerateEncryptionKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func EncryptData(data []byte, key []byte) string {
	if len(data) == 0 {
		return ""
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("error creating cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("error creating GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return hex.EncodeToString(ciphertext)
}
