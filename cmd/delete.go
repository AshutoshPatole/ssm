// Package cmd /*
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/TwiN/go-color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	serverToDelete string
	cleanConfig    bool
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a server from the configuration",
	Long: `Delete a server from the configuration. This command will remove a server by its name
and can optionally clean up empty groups and environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteServer()
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&serverToDelete, "server", "s", "", "server to delete")
	deleteCmd.Flags().BoolVarP(&cleanConfig, "clean-config", "c", false, "Clean unused groups")
	_ = deleteCmd.MarkFlagRequired("serverToDelete")
}

func cleanConfiguration(config *store.Config) {
	for gi := len(config.Groups) - 1; gi >= 0; gi-- {
		group := &config.Groups[gi]
		for ei := len(group.Environment) - 1; ei >= 0; ei-- {
			env := &group.Environment[ei]
			if len(env.Servers) == 0 {
				group.Environment = append(group.Environment[:ei], group.Environment[ei+1:]...)
				fmt.Printf(color.InCyan("Removed empty environment: %s\n"), env.Name)
			} else {
				fmt.Println(color.InCyan("No empty environments to clean"))
			}
		}
		if len(group.Environment) == 0 {
			config.Groups = append(config.Groups[:gi], config.Groups[gi+1:]...)
			fmt.Printf(color.InCyan("Removed empty group: %s\n"), group.Name)
		}
	}
}

func deleteServer() {
	if serverToDelete == "" {
		fmt.Println(color.InRed("Error: Server name to delete is required"))
		fmt.Println(color.InYellow("Usage: ssm delete -s <server_name>"))
		return
	}

	var config store.Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf(color.InRed("Error: Failed to load configuration: %v\n"), err)
		return
	}
	serverFound := false
	for gi, grp := range config.Groups {
		for ei, env := range grp.Environment {
			for si, srv := range env.Servers {
				if srv.HostName == serverToDelete {
					fmt.Printf(color.InBlackOverYellow("Server '%s' found in environment '%s' of group '%s'\n"), srv.HostName, env.Name, grp.Name)
					reader := bufio.NewReader(os.Stdin)
					fmt.Print(color.InYellow("Are you sure you want to delete this server? (y/n): "))
					response, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf(color.InRed("Error reading input: %v\n"), err)
						return
					}

					response = strings.TrimSpace(response)
					if response == "y" || response == "yes" {
						serverFound = true
						config.Groups[gi].Environment[ei].Servers = append(env.Servers[:si], env.Servers[si+1:]...)
						fmt.Println(color.InGreen("Server deleted successfully!"))
					} else {
						fmt.Println(color.InYellow("Server deletion aborted."))
					}
					break
				}
			}
		}
	}

	if !serverFound {
		fmt.Printf(color.InRed("Server '%s' was not found in the configuration\n"), serverToDelete)
		return
	}

	if cleanConfig {
		fmt.Println(color.InGreenOverBlack("Cleaning configuration..."))
		cleanConfiguration(&config)
	}

	viper.Set("groups", config.Groups)
	if err := viper.WriteConfig(); err != nil {
		fmt.Printf(color.InRed("Error: Failed to write configuration: %v\n"), err)
		return
	}
	fmt.Println(color.InGreen("Configuration updated successfully"))
}
