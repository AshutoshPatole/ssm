package ssh

import (
	"fmt"
	"github.com/TwiN/go-color"
	"golang.org/x/crypto/ssh"
	"os"
)

func AddPublicKeys(session *ssh.Session) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return false
	}
	publicKeyPath := home + "/.ssh/id_ed25519.pub"
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		fmt.Println(color.InRed("Could not read public key " + publicKeyPath))
		return false
	}
	session.Stderr = os.Stderr
	session.Stdout = os.Stdout
	command := fmt.Sprintf("mkdir -p ~/.ssh/; chmod 700 -R ~/.ssh; echo '%s' >> ~/.ssh/authorized_keys; chmod 600 ~/.ssh/authorized_keys", publicKey)
	err = session.Run(command)
	if err != nil {
		fmt.Println(color.InRed("Could not add public key " + publicKeyPath))
		return false
	}
	fmt.Println(color.InGreen("Successfully added public key " + publicKeyPath))
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)
	return true
}
