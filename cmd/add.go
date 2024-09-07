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
			logrus.Debug("No hostname provided")
			return errors.New(color.InRed("Requires hostname of the machine"))
		}
		logrus.Debugf("Hostname provided: %s", args[0])
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		validEnvironment := false
		for _, env := range allowedEnvironmentValues {
			if env == environment {
				validEnvironment = true
				break
			}
		}
		if !validEnvironment {
			logrus.Debugf("Invalid environment value: %s", environment)
			logrus.Fatalln(color.InRed("Invalid environment value"))
		}
		logrus.Debugf("Adding server with hostname: %s", args[0])
		addServer(args[0])
	},
}

func addServer(host string) {
	logrus.Debugf("Attempting to add server: %s", host)
	password, err := ssh.AskPassword()
	if err != nil {
		logrus.Debugf("Error reading password: %v", err)
		logrus.Fatal(color.InRed("Error reading password"))
	}

	if rdpConnectionString {
		logrus.Debug("RDP connection string detected")
		credentialKey := fmt.Sprintf("%s_%s_%s", group, environment, host)
		logrus.Debugf("Generated credential key: %s", credentialKey)
		err = security.StoreCredentials(credentialKey, password)
		if err != nil {
			logrus.Debugf("Error storing credential: %v", err)
			logrus.Fatalln(color.InRed("Error storing credential: " + err.Error()))
		}
		logrus.Debug("Saving RDP connection details")
		store.Save(group, environment, host, username, alias, credentialKey, true)
	} else {
		logrus.Debug("Saving SSH connection details")
		store.Save(group, environment, host, username, alias, "", false)
		logrus.Debug("Initializing SSH connection")
		ssh.InitSSHConnection(username, password, host, group, environment, alias, setupDotFiles)
	}
}

func init() {
	logrus.Debug("Initializing add command")
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
