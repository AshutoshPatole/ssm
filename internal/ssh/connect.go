package ssh

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func Connect(user, server string) {
	homeDir, _ := os.UserHomeDir()
	privateKey := filepath.Join(homeDir, ".ssh", "id_ed25519")

	_, err := os.Stat(privateKey)
	if os.IsNotExist(err) {
		logrus.Fatal("ED25519 private key does not exist on the local system")
	}
	var sshCmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		sshCmd = exec.Command("ssh", "-X", user+"@"+server)
	case "darwin":
		sshCmd = exec.Command("ssh", user+"@"+server)
	case "windows":
		sshCmd = exec.Command("ssh", user+"@"+server)
	default:
		logrus.Error("Unsupported operating system")
		return
	}

	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	err = sshCmd.Start()
	if err != nil {
		logrus.Fatal("Failed to start SSH command:", err)
	}

	err = sshCmd.Wait()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			logrus.Infof("SSH session exited with code: %d", exitErr.ExitCode())
		} else {
			logrus.Errorf("SSH command wait failed: %v", err)
		}
	}
}

func NewSSHClient(user, host string) (*ssh.Client, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	privateKey := filepath.Join(homeDir, ".ssh", "id_ed25519")

	if _, err := os.Stat(privateKey); os.IsNotExist(err) {
		return nil, fmt.Errorf("ED25519 private key does not exist at %s", privateKey)
	}
	key, err := os.ReadFile(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Parse the private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback:   ssh.InsecureIgnoreHostKey(),
		HostKeyAlgorithms: []string{ssh.KeyAlgoRSA, ssh.KeyAlgoDSA, ssh.KeyAlgoED25519, ssh.KeyAlgoECDSA256, ssh.KeyAlgoECDSA384, ssh.KeyAlgoECDSA521},
		Timeout:           5 * time.Second,
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(host, "22"), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect via SSH: %w", err)
	}

	return client, nil
}


