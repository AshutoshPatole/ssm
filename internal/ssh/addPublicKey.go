package ssh

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"os"
)

func AddPublicKeys(client *ssh.Client) bool {

	session, err := client.NewSession()
	if err != nil {
		logrus.Fatal("Failed to create SSH session:", err)
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
	publicKeyPath := home + "/.ssh/id_ed25519.pub"
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		logrus.Error("Could not read public key:", publicKeyPath)
		return false
	}

	command := fmt.Sprintf("mkdir -p ~/.ssh/; chmod 700 -R ~/.ssh; echo '%s' >> ~/.ssh/authorized_keys; chmod 600 ~/.ssh/authorized_keys", publicKey)
	err = session.Run(command)
	if err != nil {
		logrus.Error("Could not add public key:", publicKeyPath, err)
		return false
	}
	return true
}
