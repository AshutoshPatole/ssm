package ssh

import (
	"fmt"
	"github.com/AshutoshPatole/ssm-v2/internal/configuration"
	"github.com/AshutoshPatole/ssm-v2/internal/store"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"time"
)

const PORT = 22

func InitSSHConnection(user, password, host, group, environment, alias string, setupDotFiles bool) {
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
		logrus.Fatal(color.InRed(err.Error()))
	}
	defer func(client *ssh.Client) {
		err := client.Close()
		if err != nil {
			logrus.Fatalf(color.InRed("Failed to close SSH connection"))
		}
	}(client)

	AddPublicKeys(client)
	store.Save(group, environment, host, user, alias)
	if setupDotFiles {
		configuration.Setup(client, user)
	}
}
