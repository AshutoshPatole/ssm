// Package cmd /*
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AshutoshPatole/ssm-v2/internal/ssh"
	"github.com/AshutoshPatole/ssm-v2/internal/store"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to the servers",
	Long: `
To connect to the servers use:
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
		user, host, _, isRDP, err := ListToConnectServers(args[0], filterEnvironment)
		if err != nil {
			log.Fatal(err)
		}
		if isRDP {
			logrus.Infoln("TODO: Combine both rdp and connect into one")
			return
		}
		ConnectToServer(user, host)
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().StringVarP(&filterEnvironment, "filter", "f", "", "filter list by environment")
}

func ListToConnectServers(group, environment string) (string, string, string, bool, error) {
	var config store.Config

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalln(err)
	}

	selectedEnvName := ""
	selectedHostName := ""
	selectedHostIP := ""

	var serverOptions []serverOption
	user := ""
	isRDP := false
	credentialKey := ""

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

	fmt.Println(color.InGreen(selectedEnvName))
	prompt := &survey.Select{
		Message: "Select server",
		Options: labels,
	}
	err := survey.AskOne(prompt, &selectedHostName)
	if err != nil {
		return "", "", "", false, err
	}

	// Extract environment name from the selected option
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

		//ssh.Connect(user, selectedHostIP)
		return user, selectedHostIP, credentialKey, isRDP, nil
	} else {
		fmt.Println(color.InRed("Aborted! Bad Request"))
		return "", "", "", false, err
	}
}

func ConnectToServer(user, host string) {
	ssh.Connect(user, host)
}
