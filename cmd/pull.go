package cmd

import (
	"fmt"
	"os"
	"path"

	"cloud.google.com/go/firestore"
	"github.com/AshutoshPatole/ssm-v2/internal/security"
	"github.com/AshutoshPatole/ssm-v2/internal/ssh"
	"github.com/AshutoshPatole/ssm-v2/internal/store"
	"github.com/sirupsen/logrus"

	"context"

	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull your configurations from the cloud",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize Firebase here
		if err := store.InitFirebase(); err != nil {
			logrus.Fatalln("Failed to initialize Firebase:", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		if store.App == nil {
			fmt.Println("Firebase app is not initialized. Please run the sync command first.")
			return
		}
		downloadConfigurations()
	},
}

func init() {
	syncCmd.AddCommand(pullCmd)
}

func downloadConfigurations() {
	client, err := store.App.Firestore(context.Background())
	if err != nil {
		logrus.Fatal(err)
	}
	defer func(client *firestore.Client) {
		err := client.Close()
		if err != nil {
			logrus.Fatal(err)
		}
	}(client)

	userPassword, _ := ssh.AskPassword()
	uid := fetchUID(userPassword)

	logrus.Debugf("Fetching user configurations %s", uid)

	document, err := client.Collection("configurations").Doc(uid).Get(context.Background())
	if err != nil {
		logrus.Info("Did not found any configuration for current user")
		logrus.Debugf(err.Error())
		return
	}
	if document.Exists() {
		logrus.Debugf("Found configuration for current user %s", uid)
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}

	yamlEncrypted := document.Data()["ssm_yaml"].(string)
	publicKeyEncrypted := document.Data()["public"].(string)
	privateKeyEncrypted := document.Data()["private"].(string)

	key := security.GenerateEncryptionKey(userPassword)

	yaml, err := security.DecryptData(yamlEncrypted, key)
	if err != nil {
		logrus.Fatal(err)
	}

	publicKey, err := security.DecryptData(publicKeyEncrypted, key)
	if err != nil {
		logrus.Fatal(err)
	}

	privateKey, err := security.DecryptData(privateKeyEncrypted, key)
	if err != nil {
		logrus.Fatal(err)
	}

	// check if .ssh exists at home
	sshDir := userHomeDir + "/.ssh"
	_, err = os.Stat(sshDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(sshDir, 0755)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	saveFile(path.Join(userHomeDir+"/.ssm.yaml"), yaml, 0644)
	saveFile(path.Join(userHomeDir+"/.ssh/id_ed25519.pub"), publicKey, 0644)
	saveFile(path.Join(userHomeDir+"/.ssh/id_ed25519"), privateKey, 0600)
}

func saveFile(filename string, data []byte, permission uint32) {
	err := os.WriteFile(filename, data, os.FileMode(permission))
	if err != nil {
		logrus.Errorf("Failed to save file %s: %s", filename, err)
		return
	}
}

func fetchUID(userPassword string) string {
	userMap, err := store.LoginUser(userEmail, userPassword)
	if err != nil {
		logrus.Fatal(err)
	}
	userId := userMap["user_id"].(string)
	return userId
}
