package cmd

import (
	"github.com/spf13/cobra"
	"ssm-v2/internal/store"
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
	err := store.InitFirebase()
	if err != nil {
		return
	}
}
