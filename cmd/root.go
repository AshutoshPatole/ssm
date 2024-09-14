// Package cmd /*
package cmd

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TwiN/go-color"
	goversion "github.com/caarlos0/go-version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	verbose     bool
	showVersion bool

	version   = "0.0.0"
	commit    = ""
	treeState = ""
	date      = ""
	builtBy   = ""
	debugFile *os.File
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ssm",
	Short: "Simple SSH Manager with additional capabilities",
	Long: `SSM (Simple SSH Manager) is a versatile command-line tool for managing SSH connections and user authentication.
It simplifies the management of SSH profiles with commands to register users, import configurations, connect to remote servers, and synchronize settings across devices`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupFileLogging()
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Debugln("Debug level enabled")
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

//go:embed art.txt
var asciiArt string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ssm.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "toggle debug logs")
	rootCmd.PersistentFlags().BoolVar(&showVersion, "version", false, "Show version")

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
		viper.SetConfigName(".ssm")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logrus.Debugf("Using config file: %s", viper.ConfigFileUsed())
	} else {
		logrus.Errorf("Error [.ssm.yaml]: %v", err)
	}
}

func buildVersion(version, commit, date, builtBy, treeState string) goversion.Info {
	return goversion.GetVersionInfo(
		goversion.WithAppDetails("ssm", "Simple SSH Manager", "https://github.com/AshutoshPatole/ssm"),
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

func setupFileLogging() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Errorf("Failed to get home directory: %v", err)
		return
	}

	logDir := filepath.Join(homeDir, ".ssm")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logrus.Errorf("Failed to create log directory: %v", err)
		return
	}

	logFile := filepath.Join(logDir, "ssm_debug.log")
	debugFile, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Errorf("Failed to open log file: %v", err)
		return
	}

	logrus.SetOutput(io.MultiWriter(os.Stdout, debugFile))
}
