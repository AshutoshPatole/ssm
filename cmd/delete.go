// Package cmd /*
package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/AshutoshPatole/ssm/internal/store"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	serverToDelete string
	cleanConfig    bool
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a server from the configuration",
	Long: `Delete a server from the configuration. This command will remove a server by its IP address
and can optionally clean up empty groups and environments.

If no flags are provided, an interactive UI will be launched to select and delete servers.`,
	Run: func(cmd *cobra.Command, args []string) {
		if serverToDelete != "" {
			deleteServerNonInteractive()
		} else {
			runInteractiveDelete()
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVarP(&serverToDelete, "server", "s", "", "server to delete (hostname or IP)")
	deleteCmd.Flags().BoolVarP(&cleanConfig, "clean-config", "c", false, "Clean unused groups")
}

// Interactive UI Structures

type ServerItem struct {
	GroupIndex  int
	EnvIndex    int
	ServerIndex int
	Group       string
	Env         string
	Name        string
	IP          string
}

type deleteModel struct {
	items        []ServerItem
	selected     map[int]struct{}
	cursor       int
	quitting     bool
	confirming   bool
	status       string
	config       *store.Config
	scrollOffset int
	windowHeight int
}

func initialDeleteModel(config *store.Config) deleteModel {
	var items []ServerItem
	for gi, group := range config.Groups {
		for ei, env := range group.Environment {
			for si, server := range env.Servers {
				items = append(items, ServerItem{
					GroupIndex:  gi,
					EnvIndex:    ei,
					ServerIndex: si,
					Group:       group.Name,
					Env:         env.Name,
					Name:        server.HostName,
					IP:          server.IP,
				})
			}
		}
	}

	_, h, _ := term.GetSize(int(os.Stdout.Fd()))

	return deleteModel{
		items:        items,
		selected:     make(map[int]struct{}),
		config:       config,
		status:       "Select servers to delete",
		windowHeight: h,
	}
}

func (m deleteModel) Init() tea.Cmd {
	return nil
}

func (m deleteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height

	case tea.KeyMsg:
		if m.confirming {
			switch msg.String() {
			case "y", "Y":
				return m, deleteSelectedServers(m)
			case "n", "N", "esc", "q":
				m.confirming = false
				m.status = "Deletion cancelled"
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				if m.cursor >= m.scrollOffset+m.getViewportHeight() {
					m.scrollOffset = m.cursor - m.getViewportHeight() + 1
				}
			}

		case " ":
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		case "d":
			if len(m.selected) > 0 {
				m.confirming = true
				m.status = fmt.Sprintf("Delete %d selected servers? (y/n)", len(m.selected))
			} else {
				m.status = "No servers selected"
			}
		}
	}
	return m, nil
}

func (m deleteModel) getViewportHeight() int {
	reservedLines := 5 // Header + Status + Instructions
	return max(m.windowHeight-reservedLines, 1)
}

func (m deleteModel) View() string {
	if m.quitting {
		return ""
	}

	s := "Servers:\n\n"

	startIdx := m.scrollOffset
	endIdx := min(startIdx+m.getViewportHeight(), len(m.items))

	for i := startIdx; i < endIdx; i++ {
		item := m.items[i]
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checkbox := "[ ]"
		if _, ok := m.selected[i]; ok {
			checkbox = "[x]"
		}

		// Format: > [x] Group / Env / Name (IP)
		line := fmt.Sprintf("%s %s %s / %s / %s (%s)\n", cursor, checkbox, item.Group, item.Env, item.Name, item.IP)
		if m.cursor == i {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(line)
		} else {
			s += line
		}
	}

	s += "\n"
	if m.confirming {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(m.status)
	} else {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(m.status)
	}

	helpText := "Press 'space' to select, 'd' to delete, 'q' to quit"
	if m.confirming {
		helpText = "Press 'y' to confirm, 'n', 'q' or 'esc' to cancel"
	}

	s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(helpText)

	return s
}

// Logic to delete servers

func deleteSelectedServers(m deleteModel) tea.Cmd {
	// Perform deletion synchronously before quitting
	ipsToDelete := make(map[string]struct{})
	for idx := range m.selected {
		ipsToDelete[m.items[idx].IP] = struct{}{}
	}

	newGroups := []store.Group{}

	for _, grp := range m.config.Groups {
		newEnv := []store.Env{}
		for _, env := range grp.Environment {
			newServers := []store.Server{}
			for _, srv := range env.Servers {
				if _, deleteIt := ipsToDelete[srv.IP]; !deleteIt {
					newServers = append(newServers, srv)
				}
			}
			env.Servers = newServers
			if len(env.Servers) > 0 {
				newEnv = append(newEnv, env)
			}
		}
		grp.Environment = newEnv
		if len(grp.Environment) > 0 {
			newGroups = append(newGroups, grp)
		}
	}

	m.config.Groups = newGroups
	viper.Set("groups", m.config.Groups)

	if err := viper.WriteConfig(); err != nil {
		logrus.Error("Failed to write config:", err)
	}

	// Return the Quit command directly
	return tea.Quit
}

func runInteractiveDelete() {
	var config store.Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error: Failed to load configuration: %v\n", err)
		return
	}

	p := tea.NewProgram(initialDeleteModel(&config))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running interactive delete: %v\n", err)
	}
}

// Legacy / CLI Flag Mode

func deleteServerNonInteractive() {
	ipToDelete := resolveIP(serverToDelete)
	if ipToDelete == "" {
		fmt.Printf("Error: Unable to resolve '%s' to an IP address\n", serverToDelete)
		return
	}

	var config store.Config
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error: Failed to load configuration: %v\n", err)
		return
	}
	serverFound := false
	for gi, grp := range config.Groups {
		for ei, env := range grp.Environment {
			for si, srv := range env.Servers {
				if srv.IP == ipToDelete {
					fmt.Printf("Server '%s' with IP '%s' found in environment '%s' of group '%s'\n", serverToDelete, srv.IP, env.Name, grp.Name)
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("Are you sure you want to delete this server? (y/n): ")
					response, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf("Error reading input: %v\n", err)
						return
					}
					serverFound = true
					response = strings.TrimSpace(response)
					if response == "y" || response == "yes" {
						config.Groups[gi].Environment[ei].Servers = append(env.Servers[:si], env.Servers[si+1:]...)
						fmt.Println("Server deleted successfully!")
					} else {
						fmt.Println("Server deletion aborted.")
					}
					break
				}
			}
		}
	}

	if !serverFound {
		fmt.Printf("Server '%s' with IP '%s' was not found in the configuration\n", serverToDelete, ipToDelete)
		return
	}

	if cleanConfig {
		fmt.Println("Cleaning configuration...")
		cleanConfiguration(&config)
	}

	viper.Set("groups", config.Groups)
	if err := viper.WriteConfig(); err != nil {
		fmt.Printf("Error: Failed to write configuration: %v\n", err)
		return
	}
	fmt.Println("Configuration updated successfully")
}

func cleanConfiguration(config *store.Config) {
	for gi := len(config.Groups) - 1; gi >= 0; gi-- {
		group := &config.Groups[gi]
		for ei := len(group.Environment) - 1; ei >= 0; ei-- {
			env := &group.Environment[ei]
			if len(env.Servers) == 0 {
				group.Environment = append(group.Environment[:ei], group.Environment[ei+1:]...)
				fmt.Printf("Removed empty environment: %s\n", env.Name)
			}
		}
		if len(group.Environment) == 0 {
			config.Groups = append(config.Groups[:gi], config.Groups[gi+1:]...)
			fmt.Printf("Removed empty group: %s\n", group.Name)
		}
	}
}

func resolveIP(input string) string {
	ip := net.ParseIP(input)
	if ip != nil {
		return input // It's already an IP
	}
	ips, err := net.LookupIP(input)
	if err != nil || len(ips) == 0 {
		return "" // Unable to resolve
	}
	return ips[0].String()
}
