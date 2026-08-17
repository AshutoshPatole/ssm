package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/AshutoshPatole/ssm-v2/internal/ssh"
	"github.com/TwiN/go-color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rdpFilterEnvironment string
)

// rdpCmd represents the rdp command
var rdpCmd = &cobra.Command{
	Use:   "rdp",
	Short: "Connect to Windows servers using RDP",
	Long: `
To connect to Windows servers using RDP:
ssm rdp group-name

You can also specify which environments to list:
ssm rdp group-name -f ppd
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 1 {
			logrus.Fatalln(color.InYellow("Usage: ssm rdp group-name\nYou can also pass environment using -e (optional)"))
			os.Exit(1)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		user, hostIP, err := ListToConnectServers(args[0], rdpFilterEnvironment)
		if err != nil {
			logrus.Fatalln(err)
		}

		rdpPassword, _ := ssh.AskPassword()
		logrus.Debugf("Connecting to Windows machine %s using %s user\n", hostIP, user)

		ConnectToServerRDP(user, hostIP, rdpPassword)
	},
}

func init() {
	rootCmd.AddCommand(rdpCmd)
	rdpCmd.Flags().StringVarP(&rdpFilterEnvironment, "filter", "f", "", "filter list by environment")
}

func ConnectToServerRDP(user, host, password string) {
	if runtime.GOOS != "linux" {
		logrus.Error("This function is only supported on Linux")
		return
	}

	_, err := exec.LookPath("xfreerdp")
	if err != nil {
		logrus.Fatalln(color.InRed("xfreerdp is not installed or not in PATH. Please install it and try again."))
		return
	}

	if os.Getenv("DISPLAY") == "" {
		logrus.Warnln(color.InYellow("DISPLAY environment variable is not set. X11 forwarding might not be configured correctly."))
		return
	}

	// Add a delay to ensure environment is set
	time.Sleep(2 * time.Second)

	var cmd *exec.Cmd

	cmd = exec.Command("xfreerdp",
		fmt.Sprintf("/u:%s", user),
		fmt.Sprintf("/p:%s", string(password)),
		fmt.Sprintf("/v:%s", host),
		"+clipboard",
		"/dynamic-resolution",
		"/compression-level:2",
		"/log-level:TRACE",
		"/log-filters:*:TRACE",
		"/scale:100",
		"/scale-desktop:100",
		"/f",
		"-grab-keyboard",
		"/disp",
		"/gdi:hw",
		"/sound",
		"/video")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println(color.InGreen("Attempting full connection..."))
	err = cmd.Run()
	if err != nil {
		logrus.Fatalf(color.InRed("RDP client exited with error:"), err)
	} else {
		logrus.Debugln("RDP client finished successfully")
	}
}
