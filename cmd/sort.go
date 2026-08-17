package cmd

import (
	"log"

	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	Use:   "sort",
	Short: "Sort all servers in the configuration file by alias name",
	Long: `The sort command reads the SSM configuration file and sorts all servers
in every group and environment alphabetically by their alias name (a->z).
This helps keep the configuration file organized and easy to navigate.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cfg store.Config
		err := viper.Unmarshal(&cfg)
		if err != nil {
			logrus.Fatal(err.Error())
		}

		store.SortConfig(&cfg)

		viper.Set("groups", cfg.Groups)
		err = viper.WriteConfig()
		if err != nil {
			log.Fatal(err.Error())
		}
		logrus.Info("Configuration sorted successfully")
	},
}

func init() {
	rootCmd.AddCommand(sortCmd)
}
