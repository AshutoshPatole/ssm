// Package cmd /*
package cmd

import (
	"bufio"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"os"
	"ssm-v2/internal/store"
	"strings"

	"github.com/spf13/cobra"
)

var serverToDelete string

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteServer()
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&serverToDelete, "server", "s", "", "server to delete")
	_ = deleteCmd.MarkFlagRequired("serverToDelete")
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

					fmt.Println(color.InBlackOverRed(srv.HostName + " found in " + env.Name + " in " + grp.Name))
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("Are you sure you want to delete this server? (y/n): ")
					response, err := reader.ReadString('\n')
					if err != nil {
						log.Fatalln(err)
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
		fmt.Println(color.InRed("Server " + serverToDelete + " was not found in configuration"))
	}
	if serverFound {
		viper.Set("groups", config.Groups)
		if err := viper.WriteConfig(); err != nil {
			log.Fatalln(err)
		}
	}
}
