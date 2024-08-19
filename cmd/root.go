// Package cmd /*
package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TwiN/go-color"
	goversion "github.com/caarlos0/go-version"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	verbose     bool
	showVersion bool
)

var (
	version   = "0.1.2"
	commit    = ""
	treeState = ""
	date      = ""
	builtBy   = ""
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ssm",
	Short: "Simple SSH Manager with additional capabilities",
	Long: `SSM (Simple SSH Manager) is a versatile command-line tool for managing SSH connections and user authentication. 
It simplifies the management of SSH profiles with commands to register users, import configurations, connect to remote servers, and synchronize settings across devices`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.Debugln("Debug level enabled")
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			logrus.SetLevel(logrus.InfoLevel)
		}
	},
	Version: buildVersion(version, commit, date, builtBy, treeState).String(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	isUpdateAvailable, _, latestVersion := CheckForUpdates()
	if isUpdateAvailable {
		fmt.Println(color.InGreen("New Update Available"), color.InBold(color.InGreen(latestVersion)))
		return
	}
}

//go:embed .env.production
var envFile embed.FS

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ssm-v2.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "toggle debug logs")
	rootCmd.PersistentFlags().BoolVar(&showVersion, "version", false, "Show version")
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

		// generate .ssm.yaml if it is not present
		configName := ".ssm.yaml"
		configFile := filepath.Join(home, configName)

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			file, err := os.Create(configFile)
			cobra.CheckErr(err)
			defer func(file *os.File) {
				_ = file.Close()
			}(file)
		}

		// Search config in home directory with name ".ssm" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(configName)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}
}

//go:embed art.txt
var asciiArt string

//func buildVersion(version, commit, date, builtBy, treeState string) {
//
//	fmt.Println(asciiArt)
//	fmt.Printf("Version: %s\n", version)
//	fmt.Printf("Commit: %s\n", commit)
//	fmt.Printf("Built by: %s\n", builtBy)
//	fmt.Printf("TreeState: %s\n", treeState)
//	fmt.Printf("Build Date: %s\n", date)
//
//}

func buildVersion(version, commit, date, builtBy, treeState string) goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails("ssm", "Simple SSH Manager", "https://github.com/AshutoshPatole/ssm-v2"),
		goversion.WithASCIIName(asciiArt),
		func(i *goversion.Info) {
			if commit != "" {
				i.GitCommit = commit
			}
			if treeState != "" {
				i.GitTreeState = treeState
			}
			if date != "" {
				i.BuildDate = date
			}
			if version != "" {
				i.GitVersion = version
			}
			if builtBy != "" {
				i.BuiltBy = builtBy
			}
		},
	)
}
