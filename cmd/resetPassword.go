package cmd

import (
	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	resetEmail string
)

// resetPasswordCmd represents the resetPassword command
var resetPasswordCmd = &cobra.Command{
	Use:     "reset-password",
	Short:   "Reset password for a user",
	Long:    `This command sends a password reset email to the specified user.`,
	Aliases: []string{"rp", "reset"},
	Run: func(cmd *cobra.Command, args []string) {
		err := store.ResetPassword(resetEmail)
		if err != nil {
			logrus.Errorln(err)
		}
	},
}

func init() {
	authCmd.AddCommand(resetPasswordCmd)
	resetPasswordCmd.Flags().StringVarP(&resetEmail, "email", "e", "", "Email address of the user")
	resetPasswordCmd.MarkFlagRequired("email")
}
