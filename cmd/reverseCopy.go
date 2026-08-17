// Package cmd /*
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	ssh2 "github.com/AshutoshPatole/ssm/internal/ssh"
	"github.com/TwiN/go-color"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var filterByEnvironment string

// reverseCopyCmd represents the command to download files from a remote machine
var reverseCopyCmd = &cobra.Command{
	Use:     "reverse-copy",
	Short:   "Download files from a remote machine",
	Aliases: []string{"rcp"},
	Long:    `Download files or directories from a remote machine. The default location for saving is the current working directory.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 1 {
			fmt.Println(color.InYellow("Usage: ssm reverse-copy group-name\nYou can also pass an environment using -e (optional)"))
			os.Exit(1)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Debug("Initiating reverse-copy command")
		user, host, _, isRDP, err := ListToConnectServers(args[0], filterByEnvironment)
		if err != nil {
			logrus.Error("Failed to retrieve server list: ", err)
			return
		}

		if isRDP {
			fmt.Println(color.InRed("Reverse copy operation is not supported for Windows machines (RDP connections)."))
			return
		}

		logrus.Debug("Establishing SSH connection for ", user, "@", host)
		client, _ := ssh2.NewSSHClient(user, host)

		files, err := ListFiles(client, ".", false)
		if err != nil {
			log.Fatalf("Failed to retrieve file list: %v", err)
		}

		logrus.Debug("Launching interactive file selection interface")
		p := tea.NewProgram(initialModel(client, files))

		if _, err := p.Run(); err != nil {
			log.Fatalf("Error running interactive interface: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(reverseCopyCmd)
}

// FileInfo represents information about a file or directory on the remote server
type FileInfo struct {
	Name  string
	IsDir bool
	Path  string
}

// ListFiles retrieves a list of files and directories from the specified remote directory
func ListFiles(client *ssh.Client, remoteDir string, showHidden bool) ([]FileInfo, error) {
	logrus.Debug("Retrieving file list from directory: ", remoteDir)
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	lsCommand := "ls -l"
	if showHidden {
		lsCommand += "A"
	}
	output, err := session.Output(lsCommand + " " + remoteDir + " | grep -v '^total' | awk '{print $1, substr($0, index($0,$9))}'")
	if err != nil {
		return nil, fmt.Errorf("failed to execute remote file listing command: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var files []FileInfo
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		isDir := parts[0][0] == 'd'
		if !showHidden && strings.HasPrefix(parts[1], ".") {
			continue
		}
		files = append(files, FileInfo{Name: parts[1], IsDir: isDir, Path: filepath.Join(remoteDir, parts[1])})
	}

	logrus.Debugf("Retrieved %d files/directories", len(files))
	return files, nil
}

// DownloadFile downloads a file or directory from the remote server to the local machine
func DownloadFile(client *ssh.Client, remoteFile, localFile string, isDir bool) error {
	logrus.Debug("Initiating download for: ", remoteFile)
	if isDir {
		return tarAndDownloadDir(client, remoteFile, localFile)
	}
	return downloadSingleFile(client, remoteFile, localFile)
}

// downloadSingleFile downloads a single file from the remote server
func downloadSingleFile(client *ssh.Client, remoteFile, localFile string) error {
	logrus.Debug("Downloading individual file: ", remoteFile)
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	remoteFileReader, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to establish stdout pipe: %w", err)
	}

	if err := session.Start(fmt.Sprintf("cat %s", remoteFile)); err != nil {
		return fmt.Errorf("failed to initiate remote file reading: %w", err)
	}

	localFileWriter, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	if _, err := io.Copy(localFileWriter, remoteFileReader); err != nil {
		return fmt.Errorf("failed to transfer file contents: %w", err)
	}

	logrus.Debug("File successfully downloaded: ", localFile)
	return session.Wait()
}

// tarAndDownloadDir compresses a remote directory and downloads it as a tar.gz file
func tarAndDownloadDir(client *ssh.Client, remoteDir, localFile string) error {
	logrus.Debug("Compressing and downloading directory: ", remoteDir)
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	dirName := filepath.Base(remoteDir)
	tarCmd := fmt.Sprintf("tar -czf /tmp/dir.tar.gz -C %s --transform 's,^,%s/,' .", remoteDir, dirName)
	output, err := session.CombinedOutput(tarCmd)
	if err != nil {
		return fmt.Errorf("failed to compress remote directory: %w, output: %s", err, string(output))
	}

	logrus.Debug("Directory compressed, initiating download of tar file")
	return downloadSingleFile(client, "/tmp/dir.tar.gz", localFile)
}

// downloadMsg represents the result of a file download operation
type downloadMsg struct {
	success bool
	err     error
	file    string
}

// model represents the state of the interactive file selection interface
type model struct {
	client         *ssh.Client
	files          []FileInfo
	selected       map[int]struct{}
	cursor         int
	status         string
	downloading    bool
	directoryStack []string
	scrollOffset   int
	windowHeight   int
	showHidden     bool
}

// initialModel creates and initializes a new model for the interactive interface
func initialModel(client *ssh.Client, files []FileInfo) model {
	_, h, _ := term.GetSize(int(os.Stdout.Fd()))
	logrus.Debug("Initializing interface model with window height: ", h)
	return model{
		client:         client,
		files:          files,
		selected:       make(map[int]struct{}),
		cursor:         0,
		status:         "Select files to download",
		directoryStack: []string{"."},
		windowHeight:   h,
		showHidden:     false,
	}
}

// Init initializes the model (required by tea.Model interface)
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the model accordingly
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		logrus.Debug("Window dimensions updated to: ", msg.Height)
		m.windowHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			logrus.Debug("User initiated program exit")
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
				if m.cursor >= m.scrollOffset+m.getViewportHeight() {
					m.scrollOffset = m.cursor - m.getViewportHeight() + 1
				}
			}

		case " ":
			// Toggle selection status of the file or directory
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
				logrus.Debug("File deselected: ", m.files[m.cursor].Name)
			} else {
				m.selected[m.cursor] = struct{}{}
				logrus.Debug("File selected: ", m.files[m.cursor].Name)
			}

		case "enter":
			selectedFile := m.files[m.cursor]
			if selectedFile.IsDir {
				// Navigate into the selected directory
				currentDir := m.directoryStack[len(m.directoryStack)-1]
				newDir := filepath.Join(currentDir, selectedFile.Name)
				logrus.Debug("Navigating to directory: ", newDir)
				files, err := ListFiles(m.client, newDir, m.showHidden)
				if err != nil {
					m.status = fmt.Sprintf("Failed to retrieve file list: %v", err)
					return m, nil
				}
				m.directoryStack = append(m.directoryStack, newDir)
				m.files = files
				m.cursor = 0
				m.selected = make(map[int]struct{})
				m.status = fmt.Sprintf("Entered directory: %s", selectedFile.Name)
			}

		case "backspace":
			// Navigate back to the parent directory
			if len(m.directoryStack) > 1 {
				m.directoryStack = m.directoryStack[:len(m.directoryStack)-1]
				prevDir := m.directoryStack[len(m.directoryStack)-1]
				logrus.Debug("Returning to directory: ", prevDir)
				files, err := ListFiles(m.client, prevDir, m.showHidden)
				if err != nil {
					m.status = fmt.Sprintf("Failed to retrieve file list: %v", err)
					return m, nil
				}
				m.files = files
				m.cursor = 0
				m.selected = make(map[int]struct{})
				m.status = fmt.Sprintf("Returned to directory: %s", prevDir)
			}

		case "d":
			if !m.downloading {
				m.downloading = true
				m.status = "Downloading selected files... "
				logrus.Debug("Initiating file download process")
				return m, downloadFiles(m.client, m.selectedFiles())
			}

		case "a":
			// Toggle show hidden files
			m.showHidden = !m.showHidden
			currentDir := m.directoryStack[len(m.directoryStack)-1]
			files, err := ListFiles(m.client, currentDir, m.showHidden)
			if err != nil {
				m.status = fmt.Sprintf("Failed to retrieve file list: %v", err)
				return m, nil
			}
			m.files = files
			m.cursor = 0
			m.selected = make(map[int]struct{})
			if m.showHidden {
				m.status = "Showing hidden files"
			} else {
				m.status = "Hiding hidden files"
			}
		}

	case downloadMsg:
		if msg.success {
			m.status = "Successfully downloaded " + msg.file
			m.cursor = 0
			m.selected = make(map[int]struct{})
			logrus.Debug("Download completed successfully: ", msg.file)
		} else {
			m.status = "Failed to download " + msg.file + ": " + msg.err.Error()
			logrus.Error("Download operation failed: ", msg.err)
		}
		m.downloading = false
	}

	return m, nil
}

// getViewportHeight calculates the available height for displaying files
func (m model) getViewportHeight() int {
	// Reserve space for header, status, and instructions
	reservedLines := 10
	return max(m.windowHeight-reservedLines, 1)
}

// View generates the text-based user interface for file selection
func (m model) View() string {
	s := "Files:\n\n"

	// Determine the range of files to display
	startIdx := m.scrollOffset
	endIdx := min(startIdx+m.getViewportHeight(), len(m.files))

	for i := startIdx; i < endIdx; i++ {
		file := m.files[i]
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checkbox := " "
		if _, ok := m.selected[i]; ok {
			checkbox = "🗸"
		}

		fileType := ""
		if file.IsDir {
			fileType = "🗁 "
		} else {
			fileType = "  "
		}

		s += fmt.Sprintf("%s [%s] %s %s\n", cursor, checkbox, fileType, file.Name)
	}

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	instructionsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	s += "\n" + statusStyle.Render(m.status)
	s += "\n\n" + instructionsStyle.Render(
		"Press 'space' to select/deselect files\n"+
			"Press 'enter' to navigate into directory\n"+
			"Press 'backspace' to navigate back to previous directory\n"+
			"Press 'd' to download selected files\n"+
			"Press 'a' to toggle hidden files\n"+
			"Press 'q' to quit",
	)

	return s
}

// selectedFiles returns a slice of FileInfo for all selected files
func (m model) selectedFiles() []FileInfo {
	var files []FileInfo
	for i := range m.selected {
		files = append(files, m.files[i])
	}
	logrus.Debugf("Total files selected for download: %d", len(files))
	return files
}

// downloadFiles initiates the download process for selected files
func downloadFiles(client *ssh.Client, files []FileInfo) tea.Cmd {
	return func() tea.Msg {
		for _, file := range files {
			localFile := "./" + strings.TrimPrefix(file.Name, ".")
			if file.IsDir {
				localFile += ".tar.gz"
			}
			logrus.Debug("Initiating download for: ", file.Name)
			err := DownloadFile(client, file.Path, localFile, file.IsDir)
			if err != nil {
				logrus.Error("Download failed for file: ", file.Name, ", error: ", err)
				return downloadMsg{success: false, err: err, file: file.Name}
			}
		}
		logrus.Debug("All selected files downloaded successfully")
		return downloadMsg{success: true, file: ""}
	}
}
