package cmd

import (
	"github.com/spf13/cobra"
)

var (
	userEmail    string
	userPassword string
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync command lets you upload or download your public, private keys and the ssm configuration file",
	Long:  ``,
	//Run: func(cmd *cobra.Command, args []string) {
	//},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.PersistentFlags().StringVarP(&userEmail, "email", "e", "", "email address")
	syncCmd.PersistentFlags().StringVarP(&userPassword, "password", "p", "", "password")
	_ = syncCmd.MarkPersistentFlagRequired("email")
	_ = syncCmd.MarkPersistentFlagRequired("password")

}
