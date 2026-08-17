package ssh

import (
	"fmt"
	"github.com/TwiN/go-color"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"time"
)

const PORT = 22

func InitSSHConnection(user, password, host string) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback:   ssh.InsecureIgnoreHostKey(),
		HostKeyAlgorithms: []string{ssh.KeyAlgoRSA, ssh.KeyAlgoDSA, ssh.KeyAlgoED25519, ssh.KeyAlgoECDSA256, ssh.KeyAlgoECDSA384, ssh.KeyAlgoECDSA521},
		Timeout:           5 * time.Second,
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, PORT), config)
	if err != nil {
		log.Fatal(color.InRed(err.Error()))
	}
	defer func(client *ssh.Client) {
		err := client.Close()
		if err != nil {
			log.Fatalf(color.InRed("Failed to close SSH connection"))
		}
	}(client)

	session, err := client.NewSession()
	if err != nil {
		log.Fatal(color.InRed(err.Error()))
	}

	AddPublicKeys(session)

	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

}
