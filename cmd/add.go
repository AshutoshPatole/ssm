// Package cmd /*
package cmd

import (
	"errors"
	"fmt"

	"github.com/AshutoshPatole/ssm/internal/security"
	"github.com/AshutoshPatole/ssm/internal/ssh"
	"github.com/AshutoshPatole/ssm/internal/store"
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

You can specify the hostname, username, group, environment, and other optional settings such as alias and dotfiles configuration.

Example usage:
		ssm add example.com -u myuser -g mygroup -a myalias -e prod -d

This will add a server with hostname example.com, username myuser, group mygroup, alias myalias, environment prod, and setup dotfiles.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			logrus.Debug("No hostname provided")
			return errors.New("requires hostname of the machine. Please provide a hostname")
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
			logrus.Fatalf("Invalid environment value. Allowed values are: %v", allowedEnvironmentValues)
		}
		logrus.Debugf("Adding server with hostname: %s", args[0])
		addServer(args[0])
	},
}

func addServer(host string) {
	logrus.Debugf("Attempting to add server: %s", host)
	fmt.Println("Please enter the password for the server:")
	password, err := ssh.AskPassword()
	if err != nil {
		logrus.Debugf("Error reading password: %v", err)
		logrus.Fatal("Error reading password. Please try again.")
	}

	if rdpConnectionString {
		logrus.Debug("RDP connection string detected")
		credentialKey := fmt.Sprintf("%s_%s_%s", group, environment, host)
		logrus.Debugf("Generated credential key: %s", credentialKey)
		err = security.StoreCredentials(credentialKey, password)
		if err != nil {
			logrus.Debugf("Error storing credential: %v", err)
			logrus.Fatalln("Error storing credential: " + err.Error())
		}
		logrus.Debug("Saving RDP connection details")
		store.Save(group, environment, host, username, alias, credentialKey, true)
		fmt.Println("RDP connection details saved successfully!")
	} else {
		logrus.Debug("Saving SSH connection details")
		store.Save(group, environment, host, username, alias, "", false)
		logrus.Debug("Initializing SSH connection")
		ssh.InitSSHConnection(username, password, host, group, environment, alias, setupDotFiles)
		fmt.Println("SSH connection details saved and initialized successfully!")
	}
	fmt.Printf("Server %s added to group %s with alias %s in %s environment.\n", host, group, alias, environment)
}

func init() {
	logrus.Debug("Initializing add command")
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&username, "username", "u", "root", "Username to use for the connection")
	addCmd.Flags().StringVarP(&group, "group", "g", "", "Group to categorize the server")
	addCmd.Flags().StringVarP(&alias, "alias", "a", "", "Alias for easy reference to the server")
	addCmd.Flags().StringVarP(&environment, "environment", "e", "dev", "Environment of the server (dev/staging/prod)")
	addCmd.Flags().BoolVarP(&setupDotFiles, "dotfiles", "d", false, "Configure the dotfiles on the server")
	addCmd.Flags().BoolVarP(&rdpConnectionString, "rdp", "r", false, "Flag to indicate it's an RDP connection instead of SSH")
	_ = addCmd.MarkFlagRequired("group")
	_ = addCmd.MarkFlagRequired("alias")
}
