package ssh

import (
	"errors"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func Connect(user, server string) {
	homeDir, _ := os.UserHomeDir()
	privateKey := filepath.Join(homeDir, ".ssh/id_ed25519")

	_, err := os.Stat(privateKey)
	if os.IsNotExist(err) {
		logrus.Fatal(color.InRed("ED25519 private key does not exists on the local system"))
	}
	var sshCmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		sshCmd = exec.Command("ssh", user+"@"+server)
	case "darwin":
		sshCmd = exec.Command("ssh", user+"@"+server)
	case "windows":
		//logrus.Error(color.InRed("SSH connection not implemented for Windows systems yet"))
		//return
		sshCmd = exec.Command("ssh", user+"@"+server)
	default:
		logrus.Error(color.InRed("Unsupported operating system"))
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
		ok := errors.As(err, &exitErr)
		if !ok {
			logrus.Fatal("Failed to wait for SSH command:", err)
		}
		waitStatus := exitErr.Sys().(syscall.WaitStatus)
		logrus.Println("SSH session exited with:", waitStatus.ExitStatus())
	}
}

func NewSSHClient(user, host string) (*ssh.Client, error) {
	homeDir, _ := os.UserHomeDir()
	privateKey := filepath.Join(homeDir, ".ssh/id_ed25519")

	_, err := os.Stat(privateKey)
	if os.IsNotExist(err) {
		logrus.Fatal(color.InRed("ED25519 private key does not exists on the local system"))
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
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, 22), config)
	if err != nil {
		logrus.Fatal(color.InRed(err.Error()))
	}

	return client, nil
}
