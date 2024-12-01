package configuration

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

func Setup(client *ssh.Client, user string) {
	remoteHomeDir := fmt.Sprintf("/home/%s", user)
	if user == "root" {
		remoteHomeDir = "/root"
	}
	clone(client, remoteHomeDir)
	runInstallScript(client, remoteHomeDir)
}

func clone(client *ssh.Client, remoteHomeDir string) {
	url := "https://github.com/AshutoshPatole18/dotfiles.git"
	cloneDir := filepath.Join(remoteHomeDir, "dotfiles")

	command := fmt.Sprintf("git clone %s %s", url, cloneDir)
	runCommand(client, command)
}

func runInstallScript(client *ssh.Client, remoteHomeDir string) {
	installScriptPath := filepath.Join(remoteHomeDir, "dotfiles", "install.sh")

	command := fmt.Sprintf("bash %s", installScriptPath)
	runCommand(client, command)
}

func runCommand(client *ssh.Client, command string) {
	session, err := client.NewSession()
	if err != nil {
		logrus.Fatal("Failed to create session:", err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	logrus.Debug("Running command:", command)
	err = session.Run(command)
	if err != nil {
		logrus.Fatal("Failed to run command:", command, err)
	}
}
