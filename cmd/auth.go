package cmd

import (
	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Register or sign-in your user",
	Long: `The auth command provides functionality for user authentication.
It allows users to register new accounts or sign in to existing ones.
This command initializes the Firebase authentication service for secure user management.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := store.InitFirebaseOnce(); err != nil {
			logrus.Fatalln("Failed to initialize Firebase:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
