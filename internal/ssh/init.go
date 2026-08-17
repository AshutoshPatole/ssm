package ssh

import (
	"fmt"
	"net"
	"time"

	"github.com/AshutoshPatole/ssm/internal/configuration"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const PORT = 22

// InitSSHConnection attempts to establish an SSH connection using multiple methods.
// It first tries a standard SSH connection, and if that fails, attempts using a custom dialer.
// Once connected, it handles key setup and optional dotfile configuration.
func InitSSHConnection(user, password, host, group, environment, alias string, setupDotFiles bool) {
	if client, err := trySSHConnection(user, password, host); err == nil {
		logrus.Debug("SSH connection successful")
		handleSuccessfulConnection(client, user, setupDotFiles)
		return
	} else {
		logrus.Debugf("Standard SSH connection failed: %v, trying alternative method", err)
	}

	if client, err := trySSHWithCustomDialer(user, password, host); err == nil {
		logrus.Debug("SSH connection with custom dialer successful")
		handleSuccessfulConnection(client, user, setupDotFiles)
		return
	} else {
		logrus.Fatal(color.InRed("All connection attempts failed"))
	}
}

// trySSHConnection attempts to establish a standard SSH connection using password authentication.
// It returns an SSH client if successful, or an error if the connection fails.
func trySSHConnection(user, password, host string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	return ssh.Dial("ssh", fmt.Sprintf("%s:%d", host, PORT), config)
}

// trySSHWithCustomDialer attempts to establish an SSH connection using a custom network dialer.
// This method provides more control over the connection parameters and can be more reliable
// in certain network environments. It returns an SSH client if successful, or an error if the
// connection fails.
func trySSHWithCustomDialer(user, password, host string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 10 * time.Second,
	}

	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", host, PORT))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, host, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}

// handleSuccessfulConnection performs post-connection setup tasks including
// adding public keys and optionally configuring dotfiles. It ensures proper
// cleanup by closing the client connection when done.
func handleSuccessfulConnection(client *ssh.Client, user string, setupDotFiles bool) {
	defer func(client *ssh.Client) {
		err := client.Close()
		if err != nil {
			logrus.Fatalf(color.InRed("Failed to close SSH connection"))
		}
	}(client)

	AddPublicKeys(client)
	if setupDotFiles {
		configuration.Setup(client, user)
	}
}
