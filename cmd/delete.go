// Package cmd /*
package cmd

import (
	"bufio"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"ssm-v2/internal/store"
	"strings"

	"github.com/spf13/cobra"
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
				logrus.Debugln("Removed empty environment: %s", env.Name)
			} else {
				logrus.Debugln("Nothing to clean")
			}
		}
		if len(group.Environment) == 0 {
			config.Groups = append(config.Groups[:gi], config.Groups[gi+1:]...)
			logrus.Debugln("Removed empty group: %s", group.Name)
		}
	}
}

func deleteServer() {
	if serverToDelete == "" {
		logrus.Fatal("server to delete is required")
	}

	var config store.Config
	if err := viper.Unmarshal(&config); err != nil {
		logrus.Fatal(err)
	}
	serverFound := false
	for gi, grp := range config.Groups {
		for ei, env := range grp.Environment {
			for si, srv := range env.Servers {
				if srv.HostName == serverToDelete {

					logrus.Info(color.InBlackOverRed(srv.HostName + " found in " + env.Name + " in " + grp.Name))
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("Are you sure you want to delete this server? (y/n): ")
					response, err := reader.ReadString('\n')
					if err != nil {
						logrus.Fatalln(err)
					}

					response = strings.TrimSpace(response)
					if response == "y" || response == "yes" {
						serverFound = true
						config.Groups[gi].Environment[ei].Servers = append(env.Servers[:si], env.Servers[si+1:]...)
						logrus.Info(color.InGreen("Server deleted successfully!"))
					} else {
						logrus.Info(color.InYellow("Server deletion aborted."))
					}
					break
				}
			}
		}
	}

	if !serverFound {
		fmt.Println(color.InRed("Server " + serverToDelete + " was not found in configuration"))
	}

	if cleanConfig {
		fmt.Println(color.InGreenOverBlack("Cleaning configuration"))
		cleanConfiguration(&config)
	}

	viper.Set("groups", config.Groups)
	if err := viper.WriteConfig(); err != nil {
		logrus.Fatal(err)
	}
}
