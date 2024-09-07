package ssh

import (
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
)

func AddPublicKeys(client *ssh.Client) bool {

	session, err := client.NewSession()
	if err != nil {
		log.Fatal(color.InRed(err.Error()))
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return false
	}
	publicKeyPath := home + "/.ssh/id_ed25519.pub"
	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		logrus.Println(color.InRed("Could not read public key " + publicKeyPath))
		return false
	}
	session.Stderr = os.Stderr
	session.Stdout = os.Stdout
	command := fmt.Sprintf("mkdir -p ~/.ssh/; chmod 700 -R ~/.ssh; echo '%s' >> ~/.ssh/authorized_keys; chmod 600 ~/.ssh/authorized_keys", publicKey)
	err = session.Run(command)
	if err != nil {
		logrus.Println(color.InRed("Could not add public key " + publicKeyPath))
		return false
	}
	//fmt.Println(color.InGreen("Successfully added public key " + publicKeyPath))
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)
	return true
}
