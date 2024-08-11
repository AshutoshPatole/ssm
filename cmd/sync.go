package cmd

import (
	"github.com/spf13/cobra"
)

var (
	userEmail string
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync command lets you upload or download your public, private keys and the ssm configuration file",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.PersistentFlags().StringVarP(&userEmail, "email", "e", "", "email address")
	_ = syncCmd.MarkPersistentFlagRequired("email")
}
