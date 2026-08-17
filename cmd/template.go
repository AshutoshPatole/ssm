// Package cmd /*
package cmd

import (
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"ssm-v2/internal/store"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "A brief description of your command",
	Long:  ``,
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
