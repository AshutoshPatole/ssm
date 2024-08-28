// Package cmd /*
package cmd

import (
	"errors"
	"fmt"

	"github.com/AshutoshPatole/ssm/internal/security"
	"github.com/AshutoshPatole/ssm/internal/ssh"
	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	username                 string
	group                    string
	alias                    string
	environment              string
	setupDotFiles            bool
	allowedEnvironmentValues = []string{"dev", "staging", "prod"}
	rdpConnectionString      bool
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new SSH server configuration to your profile.",
	Long: `This command allows you to add a new SSH server configuration to your profile.

You can specify the hostname, username, group, environment, and other optional settings such as alias and dotfiles configuration.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New(color.InRed("Requires hostname of the machine"))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		validEnvironment := false
		for _, env := range allowedEnvironmentValues {
			if env == environment {
				validEnvironment = true
			}
		}
		if !validEnvironment {
			logrus.Fatalln(color.InRed("Invalid environment value"))
		}
		addServer(args[0])
	},
}

func addServer(host string) {
	password, err := ssh.AskPassword()
	if err != nil {
		logrus.Fatal(color.InRed("Error reading password"))
	}

	if rdpConnectionString {
		credentialKey := fmt.Sprintf("%s_%s_%s", group, environment, host)
		err = security.StoreCredentials(credentialKey, password)
		if err != nil {
			logrus.Fatalln(color.InRed("Error storing credential: " + err.Error()))
		}
		store.Save(group, environment, host, username, alias, credentialKey, true)
	} else {
		store.Save(group, environment, host, username, alias, "", false)
		ssh.InitSSHConnection(username, password, host, group, environment, alias, setupDotFiles)
	}
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&username, "username", "u", "root", "Username to use")
	addCmd.Flags().StringVarP(&group, "group", "g", "", "Group to use")
	addCmd.Flags().StringVarP(&alias, "alias", "a", "", "Alias to use")
	addCmd.Flags().StringVarP(&environment, "environment", "e", "dev", "Environment to use")
	addCmd.Flags().BoolVarP(&setupDotFiles, "dotfiles", "d", false, "Configure the dotfiles")
	addCmd.Flags().BoolVarP(&rdpConnectionString, "rdp", "r", false, "Flag to indicate its RDP connection and not try SSH")
	_ = addCmd.MarkFlagRequired("group")
	_ = addCmd.MarkFlagRequired("alias")
}
