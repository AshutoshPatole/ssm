package ssh

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func AddPublicKeys(client *ssh.Client) bool {
	session, err := client.NewSession()
	if err != nil {
		logrus.Error("Failed to create SSH session:", err)
		return false
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Error("Failed to get user home directory:", err)
		return false
	}
	publicKeyPath := filepath.Join(home, ".ssh", "id_ed25519.pub")
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		logrus.Error("Could not read public key:", publicKeyPath)
		return false
	}

	session.Stdin = bytes.NewReader(publicKey)
	command := "mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys"
	err = session.Run(command)
	if err != nil {
		logrus.Error("Could not add public key:", publicKeyPath, err)
		return false
	}
	return true
}
