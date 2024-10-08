package cmd

import (
	"fmt"

	"github.com/AshutoshPatole/ssm/internal/ssh"
	"github.com/AshutoshPatole/ssm/internal/store"
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
