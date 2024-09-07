package ssh

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// RotateKeys rotates the SSH keys for the given client using the provided Ed25519 key pair
func RotateKeys(client *ssh.Client, user, privateKeyPath, publicKeyPath string) error {
	// Read the public key
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %v", err)
	}

	// Validate that the key is Ed25519
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %v", err)
	}

	if publicKey.Type() != ssh.KeyAlgoED25519 {
		return fmt.Errorf("public key is not Ed25519, got: %s", publicKey.Type())
	}

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("echo '%s' >> ~/.ssh/authorized_keys", string(publicKeyBytes))
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to add new public key: %v", err)
	}

	return nil
}
