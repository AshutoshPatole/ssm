// package cmd
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rotateAll   bool
	rotateGroup string
	rotateHost  string
)

// rotateKeyCmd represents the rotateKey command
var rotateKeyCmd = &cobra.Command{
	Use:   "rotate-key",
	Short: "Rotate the encryption key for added security",
	Long: `The rotate-key command allows you to rotate the ssh key used by the servers.
This is an important security practice that helps protect your servers by regularly changing the key.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if rotateAll {
			rotateKeysForAll()
		} else if rotateGroup != "" {
			rotateKeysForGroup(rotateGroup)
		} else if rotateHost != "" {
			rotateKeysForHost(rotateHost)
		} else {
			cmd.Printf("Please specify --all, --group, or --host\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(rotateKeyCmd)
}

func rotateKeysForAll() {
	// Implementation
}

func rotateKeysForGroup(group string) {
	// Implementation
}

func rotateKeysForHost(host string) {
	// Implementation
}
