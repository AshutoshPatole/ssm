// Package cmd /*
package cmd

import (
	"fmt"
	"os"

	"github.com/AshutoshPatole/ssm/internal/ssh"
	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
		logrus.Errorf("Error reading file %s: %v", filePath, err)
		return
	}
	var config store.Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		logrus.Errorf("Error parsing YAML config: %v", err)
		return
	}

	if !allGroup && groupName == "" {
		logrus.Error("Please specify a group name with --group or use --all to import all groups")
		return
	}

	var groupsToImport []store.Group
	if allGroup {
		groupsToImport = config.Groups
	} else {
		found := false
		for _, g := range config.Groups {
			if g.Name == groupName {
				groupsToImport = append(groupsToImport, g)
				found = true
				break
			}
		}
		if !found {
			logrus.Errorf("Group '%s' not found in %s", groupName, filePath)
			return
		}
	}

	for _, group := range groupsToImport {
		fmt.Println("Importing group:", group.Name)
		for _, environment := range group.Environment {
			for _, host := range environment.Servers {
				fmt.Printf("Enter password for server %s (%s@%s):\n", host.Alias, host.User, host.HostName)
				newPassword, err := ssh.AskPassword()
				if err != nil {
					logrus.Errorf("Skipping server %s: %v", host.HostName, err)
					continue
				}
				store.Save(group.Name, environment.Name, host.HostName, host.User, host.Alias, "", host.IsRDP)
				if !host.IsRDP {
					ssh.InitSSHConnection(host.User, newPassword, host.HostName, group.Name, environment.Name, host.Alias, setupDotFile)
				}
			}
		}
	}
}
