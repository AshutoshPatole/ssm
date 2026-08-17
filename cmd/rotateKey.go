// package cmd
package cmd

import (
	"fmt"

	"github.com/AshutoshPatole/ssm/internal/ssh"
	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rotateAll      bool
	rotateGroup    string
	privateKeyPath string
	publicKeyPath  string
)

// rotateKeyCmd represents the rotateKey command
var rotateKeyCmd = &cobra.Command{
	Use:   "rotate-key",
	Short: "Enhance security by rotating SSH encryption keys",
	Long: `The rotate-key command enables you to update the SSH keys used by your servers.
This crucial security measure helps safeguard your infrastructure by periodically changing the authentication keys.
By rotating keys regularly, you minimize the risk of unauthorized access even if a key is compromised.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(color.InBold(color.InYellow("Note: RDP Connections will be excluded from the key rotation process")))
		if rotateAll {
			rotateKeysForAll()
		} else if rotateGroup != "" {
			rotateKeysForGroup(rotateGroup)
		} else {
			cmd.Printf("Please specify either --all to rotate keys for all servers or --group to rotate keys for a specific group\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(rotateKeyCmd)
	rotateKeyCmd.Flags().BoolVar(&rotateAll, "all", false, "Rotate SSH keys for all configured servers")
	rotateKeyCmd.Flags().StringVar(&rotateGroup, "group", "", "Rotate SSH keys for servers in a specific group")
	rotateKeyCmd.Flags().StringVar(&privateKeyPath, "private-key", "", "File path to the new Ed25519 private key")
	rotateKeyCmd.Flags().StringVar(&publicKeyPath, "public-key", "", "File path to the new Ed25519 public key")
	rotateKeyCmd.MarkFlagRequired("private-key")
	rotateKeyCmd.MarkFlagRequired("public-key")
}

func rotateKeysForAll() {
	var config store.Config
	if err := viper.Unmarshal(&config); err != nil {
		logrus.Fatalf("Failed to unmarshal configuration: %v", err)
	}

	for _, group := range config.Groups {
		for _, env := range group.Environment {
			for _, server := range env.Servers {
				if !server.IsRDP {
					logrus.Infof("Initiating key rotation for %s (%s) in group %s, environment %s", server.Alias, server.HostName, group.Name, env.Name)
					if err := rotateKeyForServer(server, group.Name, env.Name); err != nil {
						logrus.Errorf("Key rotation failed for %s: %v", server.HostName, err)
					}
				}
			}
		}
	}

	if err := viper.WriteConfig(); err != nil {
		logrus.Fatalf("Failed to save updated configuration: %v", err)
	}
}

func rotateKeysForGroup(group string) {
	var config store.Config
	if err := viper.Unmarshal(&config); err != nil {
		logrus.Fatalf("Failed to unmarshal configuration: %v", err)
	}

	groupFound := false
	for _, g := range config.Groups {
		if g.Name == group {
			groupFound = true
			for _, env := range g.Environment {
				for _, server := range env.Servers {
					if !server.IsRDP {
						logrus.Infof("Initiating key rotation for %s (%s) in group %s, environment %s", server.Alias, server.HostName, g.Name, env.Name)
						if err := rotateKeyForServer(server, g.Name, env.Name); err != nil {
							logrus.Errorf("Key rotation failed for %s: %v", server.HostName, err)
						}
					}
				}
			}
			break
		}
	}

	if !groupFound {
		logrus.Errorf("Specified group %s not found in configuration", group)
		return
	}

	if err := viper.WriteConfig(); err != nil {
		logrus.Fatalf("Failed to save updated configuration: %v", err)
	}
}

func rotateKeyForServer(server store.Server, groupName, envName string) error {
	client, err := ssh.NewSSHClient(server.User, server.HostName)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}
	defer client.Close()

	err = ssh.RotateKeys(client, server.User, privateKeyPath, publicKeyPath)
	if err != nil {
		return fmt.Errorf("key rotation process failed: %w", err)
	}

	if err := store.UpdateKeyRotationTime(groupName, envName, server.HostName); err != nil {
		return fmt.Errorf("failed to record key rotation timestamp: %w", err)
	}

	return nil
}
