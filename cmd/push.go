package cmd

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"log"
	"os"
	"ssm-v2/internal/store"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		user, err := store.LoginUser(userEmail, userPassword)
		if err != nil {
			fmt.Println(err)
			return
		}

		userId := user["user_id"].(string)

		upload(userId)
	},
}

func init() {
	syncCmd.AddCommand(pushCmd)
}

func upload(documentID string) {

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

	configurations := client.Collection("configurations")
	//configurations.
	_, err = configurations.Doc(documentID).Set(context.Background(), map[string]interface{}{
		"ssm_yaml": ssmYaml,
		"public":   publicKey,
		"private":  privateKey,
	})
	if err != nil {
		logrus.Fatalf("error adding configuration: %v", err)
	}
	logrus.Info("Configuration uploaded with reference %s", documentID)
}

func readFileAsBytes(path string) ([]byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	filePath := fmt.Sprintf("%s/%s", homeDir, path)

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}
	return fileContent, nil
}
