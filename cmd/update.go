package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "checks for latest update",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		CheckForUpdates()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func CheckForUpdates() {
	currentVersion, err := semver.NewVersion(version)
	if err != nil {
		logrus.Fatalf("Error parsing current version: %s\n", err)
	}

	latestRelease, err := getLatestRelease()
	if err != nil {
		logrus.Fatalf("Error fetching latest release: %s\n", err)
	}

	latestVersion, err := semver.NewVersion(latestRelease.TagName)
	if err != nil {
		logrus.Fatalf("Error parsing latest version: %s\n", err)
	}
	if latestVersion.GreaterThan(currentVersion) {
		fmt.Println(asciiArt)
		fmt.Printf("Current version: %s\n", currentVersion)
		fmt.Printf("New version available: %s\n", latestVersion)
		fmt.Println("https://github.com/AshutoshPatole/ssm-v2/releases")
	}
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var (
	owner = "AshutoshPatole"
	repo  = "ssm-v2"
)

func getLatestRelease() (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release GitHubRelease
	//body, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	fmt.Println("Error reading response body:", err)
	//	os.Exit(1)
	//}
	//
	//// Print the response body
	//fmt.Println(string(body))
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return nil, err
	}

	return &release, nil
}
