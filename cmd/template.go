// Package cmd /*
package cmd

import (
	"os"

	"github.com/AshutoshPatole/ssm/internal/store"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Generate and save a template YAML configuration file.",
	Long: `This command generates a template YAML configuration file with example SSH groups, environments, and server settings.
The template is then saved to your home directory as '.ssm-template.yaml'.
This file can be used as a starting point for importing your SSH profiles.`,

	Run: func(cmd *cobra.Command, args []string) {
		saveTemplate()
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
}

// saveTemplate saves the template YAML to a file
func saveTemplate() {
	content := `
groups:
  - name: atlanta
    user: root
    environment:
      - name: dev|staging|prod
        servers:
          - hostname: chn-mit-test
            alias: test
            user: root
  - name: chennai
    user: root
    environment:
      - name: prod|dev|staging
        servers:
          - hostname: chn-mit-chennai
            alias: bb
            user: root
          - hostname: chn-mit-second
            ip: 10.100.15.xx
            alias: aa
            user: root
`
	var data store.Config

	err := yaml.Unmarshal([]byte(content), &data)
	if err != nil {
		logrus.Fatal("something went wrong in unmarshalling", err.Error())
	}

	d, err := yaml.Marshal(data)
	if err != nil {
		logrus.Fatal("something went wrong in marshalling", err.Error())
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal("something went wrong in getting home dir", err.Error())
	}

	wErr := os.WriteFile(homeDir+"/.ssm-template.yaml", d, 0644)
	if wErr != nil {
		logrus.Fatal("something went wrong in writing file", wErr.Error())
	}

	logrus.Print(color.InGreen("File saved at " + homeDir + "/.ssm-template.yaml"))
}
