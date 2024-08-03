// Package cmd /*
package cmd

import (
	"fmt"
	"github.com/TwiN/go-color"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	ssh2 "ssm-v2/internal/ssh"
	"strings"
)

var filterByEnvironment string

// reverseCopyCmd downloads file from remote machine
var reverseCopyCmd = &cobra.Command{
	Use:     "reverse-copy",
	Short:   "Download file from remote machine",
	Aliases: []string{"rcp"},
	Long:    `Download file from remote machine. Default location for saving is $HOME`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 1 {
			fmt.Println(color.InYellow("Usage: ssm reverse-copy group-name\nYou can also pass environment using -e (optional)"))
			os.Exit(1)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		user, host, err := ListToConnectServers(args[0], filterByEnvironment)
		if err != nil {
			return
		}

		client, _ := ssh2.NewSSHClient(user, host)

		files, err := ListFiles(client, ".")
		if err != nil {
			log.Fatalf("Failed to list files: %v", err)
		}

		p := tea.NewProgram(initialModel(client, files))

		if _, err := p.Run(); err != nil {
			log.Fatalf("Error running program: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(reverseCopyCmd)
}

// FileInfo holds the name and type of file
type FileInfo struct {
	Name  string
	IsDir bool
}

// ListFiles lists files in the given directory on the remote server
func ListFiles(client *ssh.Client, remoteDir string) ([]FileInfo, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	output, err := session.Output("ls -l " + remoteDir + " | awk '{print $1 \" \" $9}'")
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var files []FileInfo
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		isDir := parts[0][0] == 'd'
		files = append(files, FileInfo{Name: parts[1], IsDir: isDir})
	}

	return files, nil
}

// DownloadFile downloads a file or zips a directory from the remote server
func DownloadFile(client *ssh.Client, remoteFile, localFile string, isDir bool) error {
	if isDir {
		return tarAndDownloadDir(client, remoteFile, localFile)
	}
	return downloadSingleFile(client, remoteFile, localFile)
}

func downloadSingleFile(client *ssh.Client, remoteFile, localFile string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	remoteFileReader, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := session.Start(fmt.Sprintf("cat %s", remoteFile)); err != nil {
		return fmt.Errorf("failed to start cat command: %w", err)
	}

	localFileWriter, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	if _, err := io.Copy(localFileWriter, remoteFileReader); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return session.Wait()
}

func tarAndDownloadDir(client *ssh.Client, remoteDir, localFile string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer func(session *ssh.Session) {
		_ = session.Close()
	}(session)

	tarCmd := fmt.Sprintf("cd %s && tar -czf /tmp/dir.tar.gz .", remoteDir)
	output, err := session.CombinedOutput(tarCmd)
	if err != nil {
		return fmt.Errorf("failed to tar directory: %w, output: %s", err, string(output))
	}

	return downloadSingleFile(client, "/tmp/dir.tar.gz", localFile)
}

type downloadMsg struct {
	success bool
	err     error
	file    string
}

type model struct {
	client      *ssh.Client
	files       []FileInfo
	selected    map[int]struct{}
	cursor      int
	status      string
	downloading bool
}

func initialModel(client *ssh.Client, files []FileInfo) model {
	return model{client: client, files: files, selected: make(map[int]struct{}), cursor: 0, status: "Select files to download"}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}

		case "enter":
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		case "d":
			if !m.downloading {
				m.downloading = true
				m.status = "Downloading selected files..."
				return m, downloadFiles(m.client, m.selectedFiles())
			}
		}

	case downloadMsg:
		if msg.success {
			m.status = "Downloaded " + msg.file + " successfully"
		} else {
			m.status = "Failed to download " + msg.file + ": " + msg.err.Error()
		}
		m.downloading = false
	}

	return m, nil
}

func (m model) View() string {
	s := "Files:\n\n"

	for i, file := range m.files {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checkbox := " "
		if _, ok := m.selected[i]; ok {
			checkbox = "x"
		}

		fileType := ""
		if file.IsDir {
			fileType = "[DIR] "
		}

		s += fmt.Sprintf("%s [%s] %s%s\n", cursor, checkbox, fileType, file.Name)
	}

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	s += "\n" + statusStyle.Render(m.status)
	s += "\n\nPress 'd' to download selected files"
	s += "\nPress 'q' to quit"

	return s
}

func (m model) selectedFiles() []FileInfo {
	var files []FileInfo
	for i := range m.selected {
		files = append(files, m.files[i])
	}
	return files
}

func downloadFiles(client *ssh.Client, files []FileInfo) tea.Cmd {
	return func() tea.Msg {
		for _, file := range files {
			localFile := "./" + file.Name
			if file.IsDir {
				localFile += ".tar.gz"
			}
			err := DownloadFile(client, file.Name, localFile, file.IsDir)
			if err != nil {
				return downloadMsg{success: false, err: err, file: file.Name}
			}
		}
		return downloadMsg{success: true, file: ""}
	}
}
