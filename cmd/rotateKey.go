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
	Short: "Rotate the encryption key for added security",
	Long: `The rotate-key command allows you to rotate the ssh key used by the servers.
This is an important security practice that helps protect your servers by regularly changing the key.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(color.InBold(color.InYellow("RDP Connections will be skipped")))
		if rotateAll {
			rotateKeysForAll()
		} else if rotateGroup != "" {
			rotateKeysForGroup(rotateGroup)
		} else {
			cmd.Printf("Please specify --all or --group\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(rotateKeyCmd)
	rotateKeyCmd.Flags().BoolVar(&rotateAll, "all", false, "Rotate keys for all servers")
	rotateKeyCmd.Flags().StringVar(&rotateGroup, "group", "", "Rotate keys for a specific group")
	rotateKeyCmd.Flags().StringVar(&privateKeyPath, "private-key", "", "Path to the Ed25519 private key")
	rotateKeyCmd.Flags().StringVar(&publicKeyPath, "public-key", "", "Path to the Ed25519 public key")
	rotateKeyCmd.MarkFlagRequired("private-key")
	rotateKeyCmd.MarkFlagRequired("public-key")
}

func rotateKeysForAll() {
	var config store.Config
	if err := viper.Unmarshal(&config); err != nil {
		logrus.Fatalf("Failed to unmarshal config: %v", err)
	}

	for _, group := range config.Groups {
		for _, env := range group.Environment {
			for _, server := range env.Servers {
				if !server.IsRDP {
					logrus.Infof("Rotating key for %s (%s) in group %s, environment %s", server.Alias, server.HostName, group.Name, env.Name)
					if err := rotateKeyForServer(server, group.Name, env.Name); err != nil {
						logrus.Errorf("Failed to rotate key for %s: %v", server.HostName, err)
					}
				}
			}
		}
	}

	if err := viper.WriteConfig(); err != nil {
		logrus.Fatalf("Failed to write updated config: %v", err)
	}
}

func rotateKeysForGroup(group string) {
	var config store.Config
	if err := viper.Unmarshal(&config); err != nil {
		logrus.Fatalf("Failed to unmarshal config: %v", err)
	}

	groupFound := false
	for _, g := range config.Groups {
		if g.Name == group {
			groupFound = true
			for _, env := range g.Environment {
				for _, server := range env.Servers {
					if !server.IsRDP {
						logrus.Infof("Rotating key for %s (%s) in group %s, environment %s", server.Alias, server.HostName, g.Name, env.Name)
						if err := rotateKeyForServer(server, g.Name, env.Name); err != nil {
							logrus.Errorf("Failed to rotate key for %s: %v", server.HostName, err)
						}
					}
				}
			}
			break
		}
	}

	if !groupFound {
		logrus.Errorf("Group %s not found", group)
		return
	}

	if err := viper.WriteConfig(); err != nil {
		logrus.Fatalf("Failed to write updated config: %v", err)
	}
}

func rotateKeyForServer(server store.Server, groupName, envName string) error {
	client, err := ssh.NewSSHClient(server.User, server.HostName)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer client.Close()

	err = ssh.RotateKeys(client, server.User, privateKeyPath, publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to rotate keys: %w", err)
	}

	if err := store.UpdateKeyRotationTime(groupName, envName, server.HostName); err != nil {
		return fmt.Errorf("failed to update key rotation time: %w", err)
	}

	return nil
}
