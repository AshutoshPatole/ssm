package cmd

import (
	"fmt"
	"ssm-v2/internal/store"

	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pull called")
		if store.App == nil {
			fmt.Println("Firebase app is not initialized. Please run the sync command first.")
			return
		}
		//test(store.App)
	},
}

func init() {
	syncCmd.AddCommand(pullCmd)
}
