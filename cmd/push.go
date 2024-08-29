package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

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
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize Firebase here
		if err := store.InitFirebase(); err != nil {
			logrus.Fatalln("Failed to initialize Firebase:", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		userPassword, _ := ssh.AskPassword()
		user, err := store.LoginUser(userEmail, userPassword)
		if err != nil {
			fmt.Println(err)
			return
		}

		userId := user["user_id"].(string)

		upload(userId, userPassword)
	},
}

func init() {
	syncCmd.AddCommand(pushCmd)
}

func upload(documentID string, userPassword string) {

	client, err := store.App.Firestore(context.Background())
	if err != nil {
		log.Fatalf("error getting Firestore client: %v", err)
	}
	defer func(client *firestore.Client) {
		err := client.Close()
		if err != nil {
			return
		}
	}(client)

	ssmYaml, _ := readFileAsBytes(".ssm.yaml")
	publicKey, _ := readFileAsBytes(".ssh/id_ed25519.pub")
	privateKey, _ := readFileAsBytes(".ssh/id_ed25519")
	zshrc, _ := readFileAsBytes(".zshrc")
	bashrc, _ := readFileAsBytes(".bashrc")
	tmux, _ := readFileAsBytes(".tmux.conf")
	sshConfig, _ := readFileAsBytes(".ssh/config")

	key := security.GenerateEncryptionKey(userPassword)

	encryptedSSMYaml := security.EncryptData(ssmYaml, key)
	encryptedPublicKey := security.EncryptData(publicKey, key)
	encryptedPrivateKey := security.EncryptData(privateKey, key)
	encryptedZshrc := security.EncryptData(zshrc, key)
	encryptedBashrc := security.EncryptData(bashrc, key)
	encryptedTmux := security.EncryptData(tmux, key)
	encryptedSSHConfig := security.EncryptData(sshConfig, key)

	configurations := client.Collection("configurations")
	//configurations.
	_, err = configurations.Doc(documentID).Set(context.Background(), map[string]interface{}{
		"ssm_yaml":   encryptedSSMYaml,
		"public":     encryptedPublicKey,
		"private":    encryptedPrivateKey,
		"zshrc":      encryptedZshrc,
		"bashrc":     encryptedBashrc,
		"tmux":       encryptedTmux,
		"ssh_config": encryptedSSHConfig,
	})
	if err != nil {
		logrus.Fatalf("error adding configuration: %v", err)
	}
	logrus.Infof("Configuration uploaded with reference %s", documentID)
}

func readFileAsBytes(path string) ([]byte, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	filePath := fmt.Sprintf("%s/%s", homeDir, path)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []byte{}, nil
	}
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	return fileContent, nil
}
