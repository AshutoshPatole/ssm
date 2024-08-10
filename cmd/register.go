package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"ssm-v2/internal/store"
)

var (
	email    string
	password string
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
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
	registerCmd.Flags().StringVarP(&password, "password", "p", "", "password")
	_ = registerCmd.MarkFlagRequired("email")
	_ = registerCmd.MarkFlagRequired("password")

}
