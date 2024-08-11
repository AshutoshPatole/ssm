package cmd

import (
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Register or sign-in your user",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
