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
	Short: "A brief description of your command",
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
