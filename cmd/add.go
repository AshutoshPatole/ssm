// Package cmd /*
package cmd

import (
	"errors"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"ssm-v2/internal/ssh"
)

var (
	username                 string
	group                    string
	alias                    string
	environment              string
	setupDotFiles            bool
	allowedEnvironmentValues = []string{"dev", "staging", "prod"}
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long:  ``,
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

	ssh.InitSSHConnection(username, password, host, group, environment, alias, setupDotFiles)
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&username, "username", "u", "root", "Username to use")
	addCmd.Flags().StringVarP(&group, "group", "g", "", "Group to use")
	addCmd.Flags().StringVarP(&alias, "alias", "a", "", "Alias to use")
	addCmd.Flags().StringVarP(&environment, "environment", "e", "dev", "Environment to use")
	addCmd.Flags().BoolVarP(&setupDotFiles, "dotfiles", "d", false, "Configure the dotfiles")
	_ = addCmd.MarkFlagRequired("group")
	_ = addCmd.MarkFlagRequired("alias")
}
