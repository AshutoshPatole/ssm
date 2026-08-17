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

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push your configuration to the cloud",
	Long:  `Upload your local configuration files to the cloud storage for easy synchronization across devices.`,
	Run: func(cmd *cobra.Command, args []string) {
		userPassword, err := ssh.AskPassword()
		if err != nil {
			logrus.Errorf("Error reading password: %v", err)
			return
		}
		user, err := store.LoginUser(userEmail, userPassword)
		if err != nil {
			fmt.Println(err)
			return
		}

		uidVal, ok := user["user_id"]
		if !ok {
			logrus.Error("Error: user_id claim not found in user credentials")
			return
		}
		userId, ok := uidVal.(string)
		if !ok || userId == "" {
			logrus.Error("Error: invalid user_id format")
			return
		}

		upload(userId, userPassword)
	},
}

func init() {
	syncCmd.AddCommand(pushCmd)
}

// upload encrypts and uploads the user's configuration files to Firestore
func upload(documentID string, userPassword string) {
	if err := store.InitFirebaseOnce(); err != nil {
		logrus.Errorf("Failed to initialize Firebase: %v", err)
		return
	}

	client, err := store.App.Firestore(context.Background())
	if err != nil {
		logrus.Errorf("Error getting Firestore client: %v", err)
		return
	}
	defer func(client *firestore.Client) {
		_ = client.Close()
	}(client)

	ssmYaml, _ := readFileAsBytes(".ssm.yaml")
	publicKey, _ := readFileAsBytes(filepath.Join(".ssh", "id_ed25519.pub"))
	privateKey, _ := readFileAsBytes(filepath.Join(".ssh", "id_ed25519"))
	zshrc, _ := readFileAsBytes(".zshrc")
	bashrc, _ := readFileAsBytes(".bashrc")
	tmux, _ := readFileAsBytes(".tmux.conf")
	sshConfig, _ := readFileAsBytes(filepath.Join(".ssh", "config"))

	key := security.GenerateEncryptionKey(userPassword)

	payload := map[string]interface{}{}
	if len(ssmYaml) > 0 {
		payload["ssm_yaml"] = security.EncryptData(ssmYaml, key)
	}
	if len(publicKey) > 0 {
		payload["public"] = security.EncryptData(publicKey, key)
	}
	if len(privateKey) > 0 {
		payload["private"] = security.EncryptData(privateKey, key)
	}
	if len(zshrc) > 0 {
		payload["zshrc"] = security.EncryptData(zshrc, key)
	}
	if len(bashrc) > 0 {
		payload["bashrc"] = security.EncryptData(bashrc, key)
	}
	if len(tmux) > 0 {
		payload["tmux"] = security.EncryptData(tmux, key)
	}
	if len(sshConfig) > 0 {
		payload["ssh_config"] = security.EncryptData(sshConfig, key)
	}

	configurations := client.Collection("configurations")
	_, err = configurations.Doc(documentID).Set(context.Background(), payload)
	if err != nil {
		logrus.Errorf("Error adding configuration: %v", err)
		return
	}
	logrus.Infof("Configuration successfully uploaded with reference ID: %s", documentID)
}

// readFileAsBytes reads the content of a file and returns it as a byte slice
func readFileAsBytes(relPath string) ([]byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %w", err)
	}

	filePath := filepath.Join(homeDir, relPath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []byte{}, nil
	}
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	return fileContent, nil
}
