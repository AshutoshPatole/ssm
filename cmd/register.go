package cmd

import (
	"fmt"
	"github.com/AshutoshPatole/ssm-v2/internal/ssh"
	"github.com/AshutoshPatole/ssm-v2/internal/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	email string
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a user for ssm",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize Firebase here
		if err := store.InitFirebase(); err != nil {
			logrus.Fatalln("Failed to initialize Firebase:", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		password, _ := ssh.AskPassword()
		user, err := store.RegisterUser(email, password)
		if err != nil {
			logrus.Fatalln(err)
		}
		fmt.Printf("User is registered with UID:  %s", user.UID)
	},
}

func init() {
	authCmd.AddCommand(registerCmd)
	registerCmd.Flags().StringVarP(&email, "email", "e", "", "email address")
	_ = registerCmd.MarkFlagRequired("email")
}
