package ssh

import (
	"errors"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
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
		logrus.Error(color.InRed("SSH connection not implemented for Windows systems yet"))
		return
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
