package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"cloud.google.com/go/firestore"
	"github.com/AshutoshPatole/ssm/internal/security"
	"github.com/AshutoshPatole/ssm/internal/ssh"
	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull your configurations from the cloud",
	Long:  `Retrieves and applies your stored configurations from the cloud, including SSH keys, shell configurations, and other settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		downloadConfigurations()
	},
}

func init() {
	syncCmd.AddCommand(pullCmd)
}

// downloadConfigurations retrieves user configurations from Firestore, decrypts them, and saves them to the local system
func downloadConfigurations() {
	if err := store.InitFirebaseOnce(); err != nil {
		logrus.Errorf("Failed to initialize Firebase: %v", err)
		return
	}

	userPassword, err := ssh.AskPassword()
	if err != nil {
		logrus.Errorf("Error reading password: %v", err)
		return
	}

	uid, err := fetchUID(userPassword)
	if err != nil {
		logrus.Errorf("Authentication failed: %v", err)
		return
	}

	client, err := store.App.Firestore(context.Background())
	if err != nil {
		logrus.Errorf("Failed to connect to Firestore: %v", err)
		return
	}
	defer func(client *firestore.Client) {
		_ = client.Close()
	}(client)

	logrus.Debugf("Fetching user configurations for UID: %s", uid)

	document, err := client.Collection("configurations").Doc(uid).Get(context.Background())
	if err != nil || !document.Exists() {
		logrus.Info("No configuration found for the current user")
		if err != nil {
			logrus.Debugf("Firestore error: %v", err)
		}
		return
	}

	logrus.Debugf("Found configuration for user with UID: %s", uid)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Errorf("Failed to get home directory: %v", err)
		return
	}

	dataMap := document.Data()
	key := security.GenerateEncryptionKey(userPassword)

	// Ensure .ssh directory exists
	sshDir := filepath.Join(userHomeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		logrus.Errorf("Failed to create .ssh directory: %v", err)
	}

	fileConfigs := []struct {
		mapKey      string
		relPath     string
		permissions os.FileMode
	}{
		{"ssm_yaml", ".ssm.yaml", 0644},
		{"public", filepath.Join(".ssh", "id_ed25519.pub"), 0644},
		{"private", filepath.Join(".ssh", "id_ed25519"), 0600},
		{"bashrc", ".bashrc", 0644},
		{"zshrc", ".zshrc", 0644},
		{"ssh_config", filepath.Join(".ssh", "config"), 0644},
		{"tmux", ".tmux.conf", 0644},
	}

	for _, fc := range fileConfigs {
		val, ok := dataMap[fc.mapKey]
		if !ok || val == nil {
			continue
		}
		encryptedStr, ok := val.(string)
		if !ok || encryptedStr == "" {
			continue
		}

		decrypted, err := security.DecryptData(encryptedStr, key)
		if err != nil {
			logrus.Errorf("Failed to decrypt %s: %v", fc.relPath, err)
			continue
		}
		if len(decrypted) == 0 {
			continue
		}

		fullPath := filepath.Join(userHomeDir, fc.relPath)
		if err := saveFile(fullPath, decrypted, fc.permissions); err != nil {
			logrus.Errorf("Failed to save %s: %v", fc.relPath, err)
		}
	}
}

// saveFile writes data to a file with specified permissions
func saveFile(filename string, data []byte, permission os.FileMode) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	err := os.WriteFile(filename, data, permission)
	if err != nil {
		return fmt.Errorf("failed to save file %s: %w", filename, err)
	}
	logrus.Infof("Successfully saved file: %s", filename)
	return nil
}

// fetchUID retrieves the user ID using the provided password
func fetchUID(userPassword string) (string, error) {
	userMap, err := store.LoginUser(userEmail, userPassword)
	if err != nil {
		return "", err
	}
	uidVal, ok := userMap["user_id"]
	if !ok {
		return "", fmt.Errorf("user_id not found in token claims")
	}
	uidStr, ok := uidVal.(string)
	if !ok || uidStr == "" {
		return "", fmt.Errorf("invalid user_id in token claims")
	}
	return uidStr, nil
}

