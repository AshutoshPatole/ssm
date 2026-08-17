package cmd

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"log"
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

	configurations := client.Collection("configurations")
	//configurations.
	_, err = configurations.Doc(documentID).Create(context.Background(), map[string]interface{}{
		"something": "with nothing",
	})
	if err != nil {
		logrus.Fatalf("error adding configuration: %v", err)
	}
	logrus.Info("Configuration uploaded with reference %s", documentID)
}
