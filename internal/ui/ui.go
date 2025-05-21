package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leeozaka/gommits/internal/git"
	"github.com/leeozaka/gommits/internal/models"
	"github.com/leeozaka/gommits/pkg/utils"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1).
			Align(lipgloss.Center).
			Width(60)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#2D3748")).
			Padding(0, 1).
			Align(lipgloss.Center).
			Width(60)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#E53E3E")).
			Padding(0, 1).
			Align(lipgloss.Center).
			Width(60)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#38A169")).
			Padding(0, 1).
			Align(lipgloss.Center).
			Width(60)

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9E9E9E"))

	commitHashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2D3748")).
			Bold(true)

	commitAuthorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#38A169"))

	commitFilesStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4"))
)

type screen int

const (
	homeScreen screen = iota
	directoryScreen
	authorScreen
	optionsScreen
	resultsScreen
	exportScreen
)

type fetchCommitsMsg struct {
	commits []models.CommitInfo
	branch  string
	err     error
}

func fetchCommitsCmd(dir, author string, maxCommits int, currentBranchOnly bool, parentBranch string) tea.Cmd {
	return func() tea.Msg {
		commits, branch, err := git.GatherCommits(dir, author, parentBranch, currentBranchOnly)
		if err == nil && maxCommits > 0 && len(commits) > maxCommits {
			commits = commits[:maxCommits]
		}
		return fetchCommitsMsg{commits: commits, branch: branch, err: err}
	}
}

type exportCSVMsg struct {
	path string
	err  error
}

func exportCSVCmd(commits []models.CommitInfo, path string) tea.Cmd {
	return func() tea.Msg {
		err := utils.ExportToCSV(commits, path)
		return exportCSVMsg{path: path, err: err}
	}
}

type resetToHomeMsg struct{}

func resetToHomeCmd() tea.Cmd {
	return func() tea.Msg {
		return resetToHomeMsg{}
	}
}

type tickMsg time.Time

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type model struct {
	currentScreen     screen
	textInput         textinput.Model
	directory         string
	author            string
	message           string
	messageStyle      lipgloss.Style
	commits           []models.CommitInfo
	branch            string
	csvPath           string
	maxCommits        int
	showFiles         bool
	currentBranchOnly bool
	parentBranch      string
	quitting          bool
	width             int
	height            int
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter path to Git repository"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return model{
		currentScreen:     homeScreen,
		textInput:         ti,
		messageStyle:      infoStyle,
		message:           "Welcome to Gommits App!",
		maxCommits:        0,
		showFiles:         true,
		currentBranchOnly: true,
		parentBranch:      "main",
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			switch m.currentScreen {
			case homeScreen:
				m.currentScreen = directoryScreen
				m.textInput.Placeholder = "Enter path to Git repository"
				m.textInput.SetValue("")
				m.message = "Please enter the path to a Git repository"
				m.messageStyle = infoStyle

			case directoryScreen:
				dir := m.textInput.Value()
				if dir == "" {
					dir = "."
				}
				absDir, err := filepath.Abs(dir)
				if err != nil {
					m.message = fmt.Sprintf("Error: %v", err)
					m.messageStyle = errorStyle
					return m, cmd
				}

				if !git.IsGitRepo(absDir) {
					m.message = fmt.Sprintf("Error: %s is not a Git repository", absDir)
					m.messageStyle = errorStyle
					return m, cmd
				}

				branchName, err := git.GetCurrentBranch(absDir)
				if err != nil {
					m.message = fmt.Sprintf("Error getting branch name: %v", err)
					m.messageStyle = errorStyle
					return m, cmd
				}
				m.branch = branchName
				m.directory = absDir

				m.parentBranch = git.DetectDefaultBranch(absDir)

				m.currentScreen = authorScreen
				m.textInput.Placeholder = "Enter author name or email"
				m.textInput.SetValue("")
				m.message = "Please enter the author name or email to filter commits"
				m.messageStyle = infoStyle

			case authorScreen:
				author := m.textInput.Value()
				if author == "" {
					m.message = "Error: Author name cannot be empty"
					m.messageStyle = errorStyle
					return m, cmd
				}

				m.author = author
				m.currentScreen = optionsScreen
				m.textInput.Placeholder = "Enter maximum number of commits (0 for no limit)"
				m.textInput.SetValue("0")
				m.message = "Configure additional options"
				m.messageStyle = infoStyle
				m.currentBranchOnly = true

			case optionsScreen:
				if m.message == "Enter parent branch name for comparison" {
					parentBranch := m.textInput.Value()
					if parentBranch != "" {
						m.parentBranch = parentBranch
					}
					m.textInput.Placeholder = "Enter maximum number of commits (0 for no limit)"
					m.textInput.SetValue("0")
					m.message = "Configure additional options"
					m.messageStyle = infoStyle
					return m, cmd
				}

				maxCommitsStr := m.textInput.Value()
				maxCommits := 0
				if maxCommitsStr != "" {
					fmt.Sscanf(maxCommitsStr, "%d", &maxCommits)
					if maxCommits < 0 {
						maxCommits = 0
					}
				}
				m.maxCommits = maxCommits

				m.message = fmt.Sprintf("Fetching commits for author '%s' in %s...", m.author, m.directory)
				m.messageStyle = infoStyle

				return m, fetchCommitsCmd(m.directory, m.author, m.maxCommits, m.currentBranchOnly, m.parentBranch)

			case resultsScreen:
				m.currentScreen = exportScreen
				m.textInput.Placeholder = "Enter path for CSV export"
				defaultPath := filepath.Join(m.directory, "changed_files.csv")
				m.textInput.SetValue(defaultPath)
				m.message = "Enter the path where you want to save the CSV file"
				m.messageStyle = infoStyle

			case exportScreen:
				csvPath := m.textInput.Value()
				if csvPath == "" {
					csvPath = filepath.Join(m.directory, "changed_files.csv")
				}

				m.csvPath = csvPath
				m.message = fmt.Sprintf("Exporting commits to %s...", csvPath)
				m.messageStyle = infoStyle

				return m, exportCSVCmd(m.commits, csvPath)
			}

		case tea.KeyTab:
			if m.currentScreen == directoryScreen {
				m.textInput.SetValue(".")
			} else if m.currentScreen == optionsScreen {
				if msg.Alt {
					m.showFiles = !m.showFiles
				} else {
					m.currentBranchOnly = !m.currentBranchOnly
				}
			}

		case tea.KeyRunes:
			key := string(msg.Runes)
			if m.currentScreen == optionsScreen && key == "p" {
				m.textInput.Placeholder = "Enter parent branch name for comparison"
				m.textInput.SetValue(m.parentBranch)
				m.message = "Enter parent branch name for comparison"
				m.messageStyle = infoStyle
			} else if key == "b" && m.currentScreen != homeScreen {
				switch m.currentScreen {
				case directoryScreen:
					m.currentScreen = homeScreen
					m.message = "Welcome to Gommits App!"
					m.messageStyle = infoStyle
				case authorScreen:
					m.currentScreen = directoryScreen
					m.textInput.Placeholder = "Enter path to Git repository"
					m.textInput.SetValue(m.directory)
					m.message = "Please enter the path to a Git repository"
					m.messageStyle = infoStyle
				case optionsScreen:
					m.currentScreen = authorScreen
					m.textInput.Placeholder = "Enter author name or email"
					m.textInput.SetValue(m.author)
					m.message = "Please enter the author name or email to filter commits"
					m.messageStyle = infoStyle
				case resultsScreen:
					m.currentScreen = optionsScreen
					m.textInput.Placeholder = "Enter maximum number of commits (0 for no limit)"
					m.textInput.SetValue(fmt.Sprintf("%d", m.maxCommits))
					m.message = "Configure additional options"
					m.messageStyle = infoStyle
				case exportScreen:
					m.currentScreen = resultsScreen
					m.message = fmt.Sprintf("Found %d commits in branch '%s'", len(m.commits), m.branch)
					m.messageStyle = successStyle
				}
			}
		}

	case fetchCommitsMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Error: %v", msg.err)
			m.messageStyle = errorStyle
			return m, nil
		}

		m.commits = msg.commits
		m.branch = msg.branch
		m.currentScreen = resultsScreen
		m.message = fmt.Sprintf("Found %d commits in branch '%s'", len(m.commits), m.branch)
		m.messageStyle = successStyle
		return m, nil

	case exportCSVMsg:
		if msg.err != nil {
			m.message = fmt.Sprintf("Error exporting CSV: %v", msg.err)
			m.messageStyle = errorStyle
			return m, cmd
		}

		m.message = fmt.Sprintf("Successfully exported %d commits to %s", len(m.commits), msg.path)
		m.messageStyle = successStyle

		return m, tickCmd(3 * time.Second)

	case tickMsg:
		return m, resetToHomeCmd()

	case resetToHomeMsg:
		m.currentScreen = homeScreen
		m.textInput.Placeholder = "Enter path to Git repository"
		m.textInput.SetValue("")
		m.message = "Welcome to Gommits App!"
		m.messageStyle = infoStyle

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(lipgloss.Place(m.width, 3, lipgloss.Center, lipgloss.Center, titleStyle.Render("Gommits - Commit Analyzer")))
	s.WriteString("\n")

	s.WriteString(lipgloss.Place(m.width, 2, lipgloss.Center, lipgloss.Center, m.messageStyle.Render(m.message)))
	s.WriteString("\n\n")

	var content strings.Builder
	switch m.currentScreen {
	case homeScreen:
		content.WriteString("Welcome to Gommits App!\n\n")
		content.WriteString("This application helps you analyze Git commits and export changed files.\n\n")
		content.WriteString(highlightStyle.Render("Features:\n"))
		content.WriteString("• Find commits by specific authors\n")
		content.WriteString("• View detailed commit information\n")
		content.WriteString("• Export changed files to CSV\n")
		content.WriteString("• Stylized terminal output\n\n")
		content.WriteString(modifiedHelpText("start", false, true, false))
	case directoryScreen:
		content.WriteString(m.textInput.View() + "\n\n")
		content.WriteString(modifiedHelpText("continue", true, true, true))
	case authorScreen:
		content.WriteString(m.textInput.View() + "\n\n")
		content.WriteString(modifiedHelpText("continue", true, true, false))
	case exportScreen:
		content.WriteString(m.textInput.View() + "\n\n")
		content.WriteString(modifiedHelpText("save CSV", true, true, false))
	case optionsScreen:
		content.WriteString(m.textInput.View() + "\n\n")
		content.WriteString("Press " + highlightStyle.Render("Enter") + " to fetch commits.\n")
		content.WriteString("Press " + highlightStyle.Render("Tab") + " to toggle current branch only (" + boolToYesNo(m.currentBranchOnly) + ").\n")
		content.WriteString("Press " + highlightStyle.Render("Alt+Tab") + " to toggle show files (" + boolToYesNo(m.showFiles) + ").\n")
		content.WriteString("Press " + highlightStyle.Render("P") + " to edit parent branch (" + m.parentBranch + ").\n")
		content.WriteString(modifiedHelpText("", true, true, false))

	case resultsScreen:
		if len(m.commits) == 0 {
			content.WriteString("No commits found for this author.\n\n")
		} else {
			content.WriteString(fmt.Sprintf("Found %d commits:\n\n", len(m.commits)))

			displayCount := len(m.commits)
			if displayCount > 5 {
				displayCount = 5
			}

			for i := 0; i < displayCount; i++ {
				c := m.commits[i]
				content.WriteString(commitHashStyle.Render(fmt.Sprintf("Commit: %s", c.Hash)))
				content.WriteString("\n")
				content.WriteString(fmt.Sprintf("  Author: %s", commitAuthorStyle.Render(c.Author)))
				content.WriteString("\n")
				content.WriteString(fmt.Sprintf("  Date: %s", c.Date))
				content.WriteString("\n")
				content.WriteString(fmt.Sprintf("  Message: %s", c.Message))
				content.WriteString("\n")

				if m.showFiles && len(c.Files) > 0 {
					fileCount := len(c.Files)
					if fileCount > 3 {
						content.WriteString(fmt.Sprintf("  Files: %s\n", commitFilesStyle.Render(
							fmt.Sprintf("%s and %d more...", strings.Join(c.Files[:3], ", "), fileCount-3))))
					} else {
						content.WriteString(fmt.Sprintf("  Files: %s\n", commitFilesStyle.Render(strings.Join(c.Files, ", "))))
					}
				}
				content.WriteString("\n")
			}

			if len(m.commits) > 5 {
				content.WriteString(dimmedStyle.Render(fmt.Sprintf("...and %d more commits\n\n", len(m.commits)-5)))
			}
		}
		content.WriteString("\n")
		content.WriteString("Press " + highlightStyle.Render("Enter") + " to export to CSV.\n")
		content.WriteString(modifiedHelpText("", true, true, false))
	}

	contentPlaceHeight := m.height - 8 - 3
	if contentPlaceHeight < 5 {
		contentPlaceHeight = 5
	}
	s.WriteString(lipgloss.Place(m.width, contentPlaceHeight, lipgloss.Center, lipgloss.Center, content.String()))

	footerText := "Navigation: " +
		highlightStyle.Render("Enter") + " to proceed, " +
		highlightStyle.Render("B") + " for back, " +
		highlightStyle.Render("Esc/Ctrl+C") + " to quit"
	s.WriteString("\n\n")
	s.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Center, dimmedStyle.Render(footerText)))

	return s.String()
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func modifiedHelpText(enterAction string, includeBack bool, includeQuit bool, showTabHint bool) string {
	var parts []string
	if enterAction != "" {
		parts = append(parts, highlightStyle.Render("Enter")+" to "+enterAction)
	}
	if includeBack {
		parts = append(parts, highlightStyle.Render("B")+" for back")
	}
	if includeQuit {
		parts = append(parts, highlightStyle.Render("Esc")+" to quit")
	}

	var finalHelp string
	if len(parts) > 0 {
		finalHelp = "Press " + strings.Join(parts, ", ") + ".\n"
	}

	if showTabHint {
		finalHelp += dimmedStyle.Render("Hint: Press Tab to use current directory (.).") + "\n"
	}
	return finalHelp
}

func StartUI() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
