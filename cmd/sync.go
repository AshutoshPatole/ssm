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
	Short: "Synchronize your SSH keys and SSM configuration file",
	Long: `The sync command allows you to upload or download your SSH public and private keys,
as well as the SSM configuration file and other dot files. This ensures that your SSH setup is consistent
across different machines and provides a backup of your essential SSH-related files.`,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.PersistentFlags().StringVarP(&userEmail, "email", "e", "", "User's email address for authentication")
	_ = syncCmd.MarkPersistentFlagRequired("email")
}
