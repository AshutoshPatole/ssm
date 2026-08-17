// Package cmd /*
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AshutoshPatole/ssm/internal/ssh"
	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverOption represents a single server option with its attributes
type serverOption struct {
	Label         string
	Environment   string
	HostName      string
	User          string
	IP            string
	IsRDP         bool
	CredentialKey string
}

var filterEnvironment string

// connectCmd represents the connect command for initiating server connections
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to the servers",
	Long: `
Connect to servers using the following syntax:
ssm connect group-name

You can also specify which environments to list:
ssm connect group-name -e ppd
	`,
	Aliases: []string{"c", "con"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 1 {
			fmt.Println(color.InYellow("Usage: ssm connect group-name\nYou can also pass environment using -e (optional)"))
			os.Exit(1)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Debugf("Executing connect command with args: %v", args)
		user, host, credentialKey, isRDP, err := ListToConnectServers(args[0], filterEnvironment)
		if err != nil {
			logrus.Fatalf("Error listing servers: %v", err)
		}
		if isRDP {
			logrus.Debug("Connecting to RDP server")
			ConnectToServerRDP(user, host, credentialKey)
		}
		logrus.Debug("Connecting to SSH server")
		ConnectToServer(user, host)
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().StringVarP(&filterEnvironment, "filter", "f", "", "Filter server list by environment")
}

// ListToConnectServers retrieves and displays a list of servers for connection
func ListToConnectServers(group, environment string) (string, string, string, bool, error) {
	logrus.Debugf("Listing servers for group: %s, environment: %s", group, environment)
	var config store.Config

	if err := viper.Unmarshal(&config); err != nil {
		logrus.Fatalf("Failed to unmarshal configuration: %v", err)
	}

	selectedEnvName := ""
	selectedHostName := ""
	selectedHostIP := ""

	var serverOptions []serverOption
	user := ""
	isRDP := false
	credentialKey := ""

	// Populate server options based on group and environment filters
	for _, grp := range config.Groups {
		if grp.Name == group {
			for _, env := range grp.Environment {
				if environment != "" {
					if environment == env.Name {
						for _, server := range env.Servers {
							serverOption := serverOption{
								Label:         fmt.Sprintf("%s (%s)", server.Alias, env.Name),
								Environment:   env.Name,
								HostName:      server.HostName,
								IP:            server.IP,
								User:          server.User,
								IsRDP:         server.IsRDP,
								CredentialKey: server.Password,
							}
							serverOptions = append(serverOptions, serverOption)
						}
					}
				} else {
					for _, server := range env.Servers {
						serverOption := serverOption{
							Label:         fmt.Sprintf("%s (%s)", server.Alias, env.Name),
							Environment:   env.Name,
							HostName:      server.HostName,
							IP:            server.IP,
							User:          server.User,
							IsRDP:         server.IsRDP,
							CredentialKey: server.Password,
						}
						serverOptions = append(serverOptions, serverOption)
					}
				}
			}
		}
	}
	labels := make([]string, len(serverOptions))
	for i, serverOption := range serverOptions {
		labels[i] = serverOption.Label
	}

	logrus.Debugf("Found %d server options", len(serverOptions))

	fmt.Println(color.InGreen(selectedEnvName))
	prompt := &survey.Select{
		Message: "Select server",
		Options: labels,
	}
	err := survey.AskOne(prompt, &selectedHostName)
	if err != nil {
		logrus.Errorf("Failed to select server: %v", err)
		return "", "", "", false, err
	}

	// Extract server details from the selected option
	for _, serverOption := range serverOptions {
		if serverOption.Label == selectedHostName {
			selectedEnvName = serverOption.Environment
			selectedHostName = strings.Split(serverOption.HostName, " (")[0]
			selectedHostIP = serverOption.IP
			user = serverOption.User
			isRDP = serverOption.IsRDP
			credentialKey = serverOption.CredentialKey
			break
		}
	}
	if selectedHostName != "" && user != "" && selectedEnvName != "" {
		longestLabelLength := 12
		colonWidth := 2
		fmt.Println(color.InGreen(fmt.Sprintf("%-*s: %*s", longestLabelLength, "Host", colonWidth, selectedHostName)))
		fmt.Println(color.InGreen(fmt.Sprintf("%-*s: %*s", longestLabelLength, "IP Address", colonWidth, selectedHostIP)))
		fmt.Println(color.InGreen(fmt.Sprintf("%-*s: %*s", longestLabelLength, "User", colonWidth, user)))
		fmt.Println(color.InGreen(fmt.Sprintf("%-*s: %*s", longestLabelLength, "Environment", colonWidth, selectedEnvName)))
		rdpStatus := "No"
		if isRDP {
			rdpStatus = "Yes"
		}
		fmt.Println(color.InGreen(fmt.Sprintf("%-*s: %*s", longestLabelLength, "RDP", colonWidth, rdpStatus)))
		//ssh.Connect(user, selectedHostIP)
		logrus.Debugf("Selected server: %s (%s)", selectedHostName, selectedHostIP)
		return user, selectedHostIP, credentialKey, isRDP, nil
	} else {
		fmt.Println(color.InRed("Aborted! Bad Request"))
		logrus.Error("Failed to select a valid server")
		return "", "", "", false, fmt.Errorf("invalid server selection")
	}
}

// ConnectToServer initiates an SSH connection to the specified server
func ConnectToServer(user, host string) {
	logrus.Debugf("Connecting to server: %s@%s", user, host)
	ssh.Connect(user, host)
}
