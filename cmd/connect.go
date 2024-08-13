// Package cmd /*
package cmd

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AshutoshPatole/ssm-v2/internal/ssh"
	"github.com/AshutoshPatole/ssm-v2/internal/store"
	"github.com/TwiN/go-color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

type serverOption struct {
	Label       string
	Environment string
	HostName    string
	User        string
	IP          string
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
		user, host, err := ListToConnectServers(args[0], filterEnvironment)
		if err != nil {
			log.Fatal(err)
		}
		ConnectToServer(user, host)
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().StringVarP(&filterEnvironment, "filter", "f", "", "filter list by environment")
}

func ListToConnectServers(group, environment string) (string, string, error) {
	var config store.Config

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalln(err)
	}

	selectedEnvName := ""
	selectedHostName := ""
	selectedHostIP := ""

	var serverOptions []serverOption
	user := ""

	for _, grp := range config.Groups {
		if grp.Name == group {
			for _, env := range grp.Environment {
				if environment != "" {
					if environment == env.Name {
						for _, server := range env.Servers {
							serverOption := serverOption{
								Label:       fmt.Sprintf("%s (%s)", server.Alias, env.Name),
								Environment: env.Name,
								HostName:    server.HostName,
								IP:          server.IP,
								User:        server.User,
							}
							serverOptions = append(serverOptions, serverOption)
						}
					}
				} else {
					for _, server := range env.Servers {
						serverOption := serverOption{
							Label:       fmt.Sprintf("%s (%s)", server.Alias, env.Name),
							Environment: env.Name,
							HostName:    server.HostName,
							IP:          server.IP,
							User:        server.User,
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
		return "", "", err
	}

	// Extract environment name from the selected option
	for _, serverOption := range serverOptions {
		if serverOption.Label == selectedHostName {
			selectedEnvName = serverOption.Environment
			selectedHostName = strings.Split(serverOption.HostName, " (")[0]
			selectedHostIP = serverOption.IP
			user = serverOption.User
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
		return user, selectedHostIP, nil
	} else {
		fmt.Println(color.InRed("Aborted! Bad Request"))
		return "", "", err
	}
}

func ConnectToServer(user, host string) {
	ssh.Connect(user, host)
}
