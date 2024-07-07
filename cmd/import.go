// Package cmd /*
package cmd

import (
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"ssm-v2/internal/ssh"
	"ssm-v2/internal/store"
)

var (
	filePath     string
	groupName    string
	allGroup     bool
	setupDotFile bool
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import SSH configurations from a YAML file",
	Long:  `This command imports SSH configurations from a specified YAML file and sets up SSH connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		readFile()
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVarP(&filePath, "file", "f", "", "file path")
	importCmd.Flags().StringVarP(&groupName, "group", "g", "", "group name")
	importCmd.Flags().BoolVarP(&allGroup, "all", "a", false, "all groups")
	importCmd.Flags().BoolVarP(&setupDotFile, "setup-dot", "", false, "setup dot files in servers")
	_ = importCmd.MarkFlagRequired("file")
}

func readFile() {
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		logrus.Fatalln(err)
	}
	var config store.Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		logrus.Fatalln(err)
	}

	if !allGroup && groupName == "" {
		logrus.Fatal(color.InRed("Either specify one group name or use --all flag to import everything"))
	}

	if allGroup {
		for _, group := range config.Groups {
			fmt.Println("Importing ", group.Name)
			password, err := ssh.AskPassword()
			if err != nil {
				logrus.Fatalln(err)
				return
			}
			for _, environment := range group.Environment {
				for _, host := range environment.Servers {
					if host.User == group.User {
						ssh.InitSSHConnection(host.User, password, host.HostName, group.Name, environment.Name, host.Alias, setupDotFile)
					} else {
						logrus.Info("User %s for %s does not matches with group user %s", host.User, host.HostName, group.User)
						newPassword, _ := ssh.AskPassword()
						ssh.InitSSHConnection(host.User, newPassword, host.HostName, group.Name, environment.Name, host.Alias, setupDotFile)
					}
				}
			}
		}
	}
}
