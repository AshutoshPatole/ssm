package cmd

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the SSM configuration file in a text editor",
	Long: `The edit command opens the SSM configuration file in your preferred text editor.
On Linux/macOS it tries nvim, then vi, then nano.
On Windows it opens the file in notepad.`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			logrus.Fatal("No configuration file found")
		}

		editor, editorArgs := findEditor()
		editorArgs = append(editorArgs, configFile)

		proc := exec.Command(editor, editorArgs...)
		proc.Stdin = os.Stdin
		proc.Stdout = os.Stdout
		proc.Stderr = os.Stderr

		if err := proc.Run(); err != nil {
			logrus.Fatalf("Failed to open editor: %s", err.Error())
		}
	},
}

func findEditor() (string, []string) {
	if runtime.GOOS == "windows" {
		return "notepad", nil
	}

	editors := []string{"nvim", "vi", "nano"}
	for _, editor := range editors {
		if path, err := exec.LookPath(editor); err == nil {
			return path, nil
		}
	}

	logrus.Fatal("No suitable text editor found (tried nvim, vi, nano)")
	return "", nil
}

func init() {
	rootCmd.AddCommand(editCmd)
}
