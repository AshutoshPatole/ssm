package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"

	"golang.org/x/crypto/pbkdf2"
)

// Salt used for PBKDF2 key derivation
const KeySalt = "SSM_PBKDF2_SALT_v1"

func GenerateEncryptionKey(password string) []byte {
	return pbkdf2.Key([]byte(password), []byte(KeySalt), 100000, 32, sha256.New)
}

func EncryptData(data []byte, key []byte) string {
	if len(data) == 0 {
		return ""
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("error creating cipher: %v", err)
		return ""
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("error creating GCM: %v", err)
		return ""
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("error generating nonce: %v", err)
		return ""
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return hex.EncodeToString(ciphertext)
}
