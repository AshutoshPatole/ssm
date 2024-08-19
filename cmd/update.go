package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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

		fmt.Print("Do you want to download the update? (y/n): ")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(answer) == "y" {
			downloadUpdate(latestRelease)
		}
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var release GitHubRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return nil, err
	}

	return &release, nil
}

func downloadUpdate(release *GitHubRelease) {
	assetName := getAssetName()
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		logrus.Fatalf("No matching asset found for your system")
		return
	}

	upgrade(downloadURL, assetName)

	logrus.Debugf("Update downloaded as %s\n", assetName)
}

func getAssetName() string {
	operatingSystem := runtime.GOOS
	arch := runtime.GOARCH

	switch operatingSystem {
	case "darwin":
		if arch == "arm64" {
			return "ssm-v2_Darwin_arm64.tar.gz"
		}
		return "ssm-v2_Darwin_x86_64.tar.gz"
	case "linux":
		if arch == "arm64" {
			return "ssm-v2_Linux_arm64.tar.gz"
		} else if arch == "386" {
			return "ssm-v2_Linux_i386.tar.gz"
		}
		return "ssm-v2_Linux_x86_64.tar.gz"
	case "windows":
		if arch == "arm64" {
			return "ssm-v2_Windows_arm64.zip"
		}
		return "ssm-v2_Windows_x86_64.zip"
	default:
		logrus.Fatalf("Unsupported operating system: %s", operatingSystem)
		return ""
	}
}

func upgrade(downloadURL string, assetName string) {
	tempDir, err := os.MkdirTemp("", "ssm-update")
	if err != nil {
		logrus.Fatalf("Error creating temp directory: %s", err)
		return
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, assetName)

	if err := downloadFile(downloadURL, archivePath); err != nil {
		logrus.Fatalf("Error downloading update: %s", err)
		return
	}

	if runtime.GOOS == "windows" {
		if err := handleWindowsUpdate(archivePath); err != nil {
			logrus.Fatalf("Error updating on Windows: %s", err)
		}
		return
	}

	if runtime.GOOS == "linux" {
		binaryPath, err := extractAndGetBinary(archivePath, tempDir)
		if err != nil {
			logrus.Fatalf("Error extracting update: %s", err)
		}

		if err := installBinary(binaryPath); err != nil {
			logrus.Fatalf("Error installing update: %s", err)
		}
	}

	fmt.Println("Update successfully installed to /usr/local/bin")
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractAndGetBinary(archivePath, destDir string) (string, error) {
	if strings.HasSuffix(archivePath, ".tar.gz") {
		return extractTarGz(archivePath, destDir)
	} else if strings.HasSuffix(archivePath, ".zip") {
		return extractZip(archivePath, destDir)
	}
	return "", fmt.Errorf("unsupported archive format")
}

func extractTarGz(archivePath, destDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var binaryPath string
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		path := filepath.Join(destDir, header.Name)
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(path, 0755); err != nil {
				return "", err
			}
		} else if header.Typeflag == tar.TypeReg {
			outFile, err := os.Create(path)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", err
			}
			outFile.Close()

			if strings.HasPrefix(header.Name, "ssm") {
				binaryPath = path
			}
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("binary not found in archive")
	}

	return binaryPath, nil
}

func extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var binaryPath string
	for _, f := range r.File {
		path := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return "", err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return "", err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return "", err
		}

		if strings.HasPrefix(f.Name, "ssm") {
			binaryPath = path
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("binary not found in archive")
	}

	return binaryPath, nil
}

func installBinary(binaryPath string) error {
	commands := []string{"sudo rm -f /usr/local/bin/ssm", fmt.Sprintf("sudo mv %s /usr/local/bin/ssm", binaryPath), "sudo chmod +x /usr/local/bin/ssm"}
	for _, cmdStr := range commands {
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func handleWindowsUpdate(archivePath string) error {
	extractDir := filepath.Dir(archivePath)

	// Extract the ZIP file
	if _, err := extractZip(archivePath, extractDir); err != nil {
		return fmt.Errorf("failed to extract ZIP: %w", err)
	}

	// Find the binary
	var binaryPath string
	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), "ssm") && strings.HasSuffix(info.Name(), ".exe") {
			binaryPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error finding binary: %w", err)
	}
	if binaryPath == "" {
		return fmt.Errorf("binary not found in extracted files")
	}

	// Move the binary to a permanent location (e.g., C:\Program Files\ssm)
	installDir := filepath.Join(os.Getenv("ProgramFiles"), "ssm")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	newBinaryPath := filepath.Join(installDir, filepath.Base(binaryPath))
	if err := os.Rename(binaryPath, newBinaryPath); err != nil {
		return fmt.Errorf("failed to move binary: %w", err)
	}

	fmt.Printf("Update successfully installed to %s\n", newBinaryPath)
	fmt.Println("Please add it to your PATH variable.")

	return nil
}
