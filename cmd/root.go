// Package cmd /*
package cmd

import (
	"embed"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ssm",
	Short: "A brief description of your application",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.Debugln("Debug level enabled")
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			logrus.SetLevel(logrus.InfoLevel)
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

//go:embed .env.production
var envFile embed.FS

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ssm-v2.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "toggle debug logs")

	// Note: Not sure if this is right method, but I am finding it difficult
	// to handle .env files and the firebase config file
	data, err := envFile.ReadFile(".env.production")
	if err != nil {
		logrus.Errorf("error reading embedded env file: %v", err)
	}

	// Write the embedded data to a temporary file and load it with dotenv
	tempFile, err := os.CreateTemp("", ".env")
	if err != nil {
		logrus.Errorf("error creating temporary file: %v", err)
	}
	defer func(tempFile *os.File) {
		_ = tempFile.Close()
	}(tempFile)

	if _, err := tempFile.Write(data); err != nil {
		logrus.Errorf("error writing to temporary file: %v", err)
	}

	err = godotenv.Load(tempFile.Name())
	if err != nil {
		return
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ssm-v2" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ssm.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}
}
