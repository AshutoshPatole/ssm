package cmd

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"ssm-v2/internal/store"

	"context"
	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pull called")
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
		panic(err)
	}
	defer func(client *firestore.Client) {
		err := client.Close()
		if err != nil {
			panic(err)
		}
	}(client)

	uid := fetchUID()
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
		panic(err)
	}
	yaml := document.Data()["ssm_yaml"].([]byte)
	publicKey := document.Data()["public"].([]byte)
	privateKey := document.Data()["private"].([]byte)
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

func fetchUID() string {
	userMap, err := store.LoginUser(userEmail, userPassword)
	if err != nil {
		panic(err)
	}
	userId := userMap["user_id"].(string)
	return userId
}
