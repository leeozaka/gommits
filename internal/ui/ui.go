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

	toastStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#38A169")).
			Padding(1, 3).
			Margin(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#2F855A")).
			Bold(true).
			Align(lipgloss.Center)

	toastErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#E53E3E")).
			Padding(1, 3).
			Margin(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#C53030")).
			Bold(true).
			Align(lipgloss.Center)
)

func fetchCommitsCmd(dir, author string, maxCommits int, currentBranchOnly bool, parentBranch string) tea.Cmd {
	return func() tea.Msg {
		commits, branch, err := git.GatherCommits(dir, author, parentBranch, currentBranchOnly)
		if err == nil && maxCommits > 0 && len(commits) > maxCommits {
			commits = commits[:maxCommits]
		}
		return models.FetchCommitsMsg{Commits: commits, Branch: branch, Err: err}
	}
}

func exportExcelCmd(commits []models.CommitInfo, repoPath string) tea.Cmd {
	return func() tea.Msg {
		err := utils.ExportToExcel(commits, repoPath)
		return models.ExportExcelMsg{Path: repoPath, Err: err}
	}
}

func resetToHomeCmd() tea.Cmd {
	return func() tea.Msg {
		return models.ResetToHomeMsg{}
	}
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return models.TickMsg(t)
	})
}

func showToastCmd(message string, toastType models.ToastType, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return models.ShowToastMsg{Message: message, Type: toastType, Duration: duration}
	}
}

func hideToastCmd(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return models.HideToastMsg{}
	})
}

type model struct {
	currentScreen     models.Screen
	textInput         textinput.Model
	directory         string
	author            string
	message           string
	messageStyle      lipgloss.Style
	commits           []models.CommitInfo
	branch            string
	maxCommits        int
	showFiles         bool
	currentBranchOnly bool
	parentBranch      string
	quitting          bool
	width             int
	height            int
	toast             models.Toast
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter path to Git repository"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	return model{
		currentScreen:     models.HomeScreen,
		textInput:         ti,
		message:           "Welcome to Gommits App!",
		messageStyle:      infoStyle,
		maxCommits:        0,
		showFiles:         true,
		currentBranchOnly: true,
		parentBranch:      "main",
		toast: models.Toast{
			Visible:  false,
			Opacity:  0.0,
			Position: 0.0,
			Duration: 3 * time.Second,
		},
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
			case models.HomeScreen:
				m.currentScreen = models.DirectoryScreen
				m.textInput.Placeholder = "Enter path to Git repository"
				m.textInput.SetValue("")
				m.message = "Please enter the path to a Git repository"
				m.messageStyle = infoStyle

			case models.DirectoryScreen:
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

				m.currentScreen = models.AuthorScreen
				m.textInput.Placeholder = "Enter author name or email"
				m.textInput.SetValue("")
				m.message = "Please enter the author name or email to filter commits"
				m.messageStyle = infoStyle

			case models.AuthorScreen:
				author := m.textInput.Value()
				if author == "" {
					m.message = "Error: Author name cannot be empty"
					m.messageStyle = errorStyle
					return m, cmd
				}

				m.author = author
				m.currentScreen = models.OptionsScreen
				m.textInput.Placeholder = "Enter maximum number of commits (0 for no limit)"
				m.textInput.SetValue("0")
				m.message = "Configure additional options"
				m.messageStyle = infoStyle
				m.currentBranchOnly = true

			case models.OptionsScreen:
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

			case models.ResultsScreen:
				m.message = "Exporting commits to Excel..."
				m.messageStyle = infoStyle

				return m, exportExcelCmd(m.commits, m.directory)
			}

		case tea.KeyTab:
			if m.currentScreen == models.DirectoryScreen {
				m.textInput.SetValue(".")
			} else if m.currentScreen == models.OptionsScreen {
				if msg.Alt {
					m.showFiles = !m.showFiles
				} else {
					m.currentBranchOnly = !m.currentBranchOnly
				}
			}

		case tea.KeyRunes:
			key := string(msg.Runes)
			if m.currentScreen == models.OptionsScreen && key == "p" {
				m.textInput.Placeholder = "Enter parent branch name for comparison"
				m.textInput.SetValue(m.parentBranch)
				m.message = "Enter parent branch name for comparison"
				m.messageStyle = infoStyle
			} else if key == "b" && m.currentScreen != models.HomeScreen {
				switch m.currentScreen {
				case models.DirectoryScreen:
					m.currentScreen = models.HomeScreen
					m.message = "Welcome to Gommits App!"
					m.messageStyle = infoStyle
				case models.AuthorScreen:
					m.currentScreen = models.DirectoryScreen
					m.textInput.Placeholder = "Enter path to Git repository"
					m.textInput.SetValue(m.directory)
					m.message = "Please enter the path to a Git repository"
					m.messageStyle = infoStyle
				case models.OptionsScreen:
					m.currentScreen = models.AuthorScreen
					m.textInput.Placeholder = "Enter author name or email"
					m.textInput.SetValue(m.author)
					m.message = "Please enter the author name or email to filter commits"
					m.messageStyle = infoStyle
				case models.ResultsScreen:
					m.currentScreen = models.OptionsScreen
					m.textInput.Placeholder = "Enter maximum number of commits (0 for no limit)"
					m.textInput.SetValue(fmt.Sprintf("%d", m.maxCommits))
					m.message = "Configure additional options"
					m.messageStyle = infoStyle
				}
			}
		}

	case models.FetchCommitsMsg:
		if msg.Err != nil {
			m.message = fmt.Sprintf("Error: %v", msg.Err)
			m.messageStyle = errorStyle
			return m, nil
		}

		m.commits = msg.Commits
		m.branch = msg.Branch
		m.currentScreen = models.ResultsScreen
		m.message = fmt.Sprintf("Found %d commits in branch '%s'", len(m.commits), m.branch)
		m.messageStyle = successStyle
		return m, nil

	case models.ExportExcelMsg:
		if msg.Err != nil {
			return m, showToastCmd("❌ Export failed", models.ToastError, 3*time.Second)
		}

		toastMessage := fmt.Sprintf("✅ Exported %d commits to Excel", len(m.commits))

		return m, showToastCmd(toastMessage, models.ToastSuccess, 3*time.Second)

	case models.ShowToastMsg:
		m.toast = models.Toast{
			Message:   msg.Message,
			Type:      msg.Type,
			Visible:   true,
			Opacity:   0.0,
			Position:  0.0,
			StartTime: time.Now(),
			Duration:  msg.Duration,
		}
		return m, tea.Batch(
			hideToastCmd(msg.Duration),
			tickCmd(50*time.Millisecond),
		)

	case models.HideToastMsg:
		m.toast.Visible = false
		m.toast.Opacity = 0.0
		m.toast.Position = 0.0
		return m, nil

	case models.TickMsg:
		if m.toast.Visible {
			elapsed := time.Since(m.toast.StartTime)
			if elapsed >= m.toast.Duration {
				m.toast.Visible = false
				m.toast.Opacity = 0.0
				m.toast.Position = 0.0
			} else {
				slideInDuration := 300 * time.Millisecond
				if elapsed < slideInDuration {
					slideProgress := float64(elapsed) / float64(slideInDuration)
					m.toast.Position = 1 - (1-slideProgress)*(1-slideProgress)*(1-slideProgress)
				} else {
					m.toast.Position = 1.0
				}

				fadeInDuration := 200 * time.Millisecond
				if elapsed < fadeInDuration {
					m.toast.Opacity = float64(elapsed) / float64(fadeInDuration)
				} else {
					fadeOutStart := m.toast.Duration - 500*time.Millisecond
					if elapsed >= fadeOutStart {
						fadeProgress := float64(elapsed-fadeOutStart) / float64(500*time.Millisecond)
						m.toast.Opacity = 1.0 - fadeProgress
					} else {
						m.toast.Opacity = 1.0
					}
				}

				return m, tickCmd(50 * time.Millisecond)
			}
		}
		return m, nil

	case models.ResetToHomeMsg:
		m.currentScreen = models.HomeScreen
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
	if m.toast.Visible && m.toast.Opacity > 0 {
		return m.renderMainContentWithToast()
	}

	return m.renderMainContent()
}

func (m model) renderMainContent() string {
	var s strings.Builder

	s.WriteString(lipgloss.Place(m.width, 3, lipgloss.Center, lipgloss.Center, titleStyle.Render("Gommits - Commit Analyzer")))
	s.WriteString("\n")

	s.WriteString(lipgloss.Place(m.width, 2, lipgloss.Center, lipgloss.Center, m.messageStyle.Render(m.message)))
	s.WriteString("\n\n")

	var content strings.Builder
	switch m.currentScreen {
	case models.HomeScreen:
		content.WriteString("Welcome to Gommits App!\n\n")
		content.WriteString("This application helps you analyze Git commits and export changed files.\n\n")
		content.WriteString(highlightStyle.Render("Features:\n"))
		content.WriteString("• Find commits by specific authors\n")
		content.WriteString("• View detailed commit information\n")
		content.WriteString("• Export changed files to Excel\n")
		content.WriteString("• Stylized terminal output\n\n")
		content.WriteString(modifiedHelpText("start", false, true, false))
	case models.DirectoryScreen:
		content.WriteString(m.textInput.View() + "\n\n")
		content.WriteString(modifiedHelpText("continue", true, true, true))
	case models.AuthorScreen:
		content.WriteString(m.textInput.View() + "\n\n")
		content.WriteString(modifiedHelpText("continue", true, true, false))

	case models.OptionsScreen:
		content.WriteString(m.textInput.View() + "\n\n")
		content.WriteString("Press " + highlightStyle.Render("Enter") + " to fetch commits.\n")
		content.WriteString("Press " + highlightStyle.Render("Tab") + " to toggle current branch only (" + boolToYesNo(m.currentBranchOnly) + ").\n")
		content.WriteString("Press " + highlightStyle.Render("Alt+Tab") + " to toggle show files (" + boolToYesNo(m.showFiles) + ").\n")
		content.WriteString("Press " + highlightStyle.Render("P") + " to edit parent branch (" + m.parentBranch + ").\n")
		content.WriteString(modifiedHelpText("", true, true, false))

	case models.ResultsScreen:
		if len(m.commits) == 0 {
			content.WriteString("No commits found for this author.\n\n")
		} else {
			content.WriteString(fmt.Sprintf("Found %d commits:\n\n", len(m.commits)))

			availableHeight := m.height - 15
			if availableHeight < 10 {
				availableHeight = 10
			}

			linesPerCommit := 5
			if m.showFiles {
				linesPerCommit = 7
			}

			maxDisplayCommits := availableHeight / linesPerCommit
			if maxDisplayCommits < 1 {
				maxDisplayCommits = 1
			}
			if maxDisplayCommits > 5 {
				maxDisplayCommits = 5
			}

			displayCount := len(m.commits)
			if displayCount > maxDisplayCommits {
				displayCount = maxDisplayCommits
			}

			for i := 0; i < displayCount; i++ {
				c := m.commits[i]
				content.WriteString(commitHashStyle.Render(fmt.Sprintf("Commit: %s", c.Hash)))
				content.WriteString("\n")
				content.WriteString(fmt.Sprintf("  Author: %s", commitAuthorStyle.Render(c.Author)))
				content.WriteString("\n")
				content.WriteString(fmt.Sprintf("  Date: %s", c.Date))
				content.WriteString("\n")

				message := c.Message
				if len(message) > 60 {
					message = message[:57] + "..."
				}
				content.WriteString(fmt.Sprintf("  Message: %s", message))
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

			if len(m.commits) > displayCount {
				content.WriteString(dimmedStyle.Render(fmt.Sprintf("...and %d more commits\n", len(m.commits)-displayCount)))
			}
		}
		content.WriteString("\n")
		content.WriteString("Press " + highlightStyle.Render("Enter") + " to export to Excel.\n")
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

func (m model) renderMainContentWithToast() string {
	var s strings.Builder

	toastContent := m.renderToast()
	if toastContent != "" {
		verticalOffset := int(float64(2) * (1 - m.toast.Position))
		for i := 0; i < verticalOffset; i++ {
			s.WriteString("\n")
		}

		toastWidth := lipgloss.Width(toastContent)
		horizontalPadding := (m.width - toastWidth) / 2
		if horizontalPadding < 0 {
			horizontalPadding = 0
		}

		s.WriteString(strings.Repeat(" ", horizontalPadding))
		s.WriteString(toastContent)
		s.WriteString("\n\n")
	}

	s.WriteString(lipgloss.Place(m.width, 3, lipgloss.Center, lipgloss.Center, titleStyle.Render("Gommits - Commit Analyzer")))
	s.WriteString("\n")

	s.WriteString(lipgloss.Place(m.width, 2, lipgloss.Center, lipgloss.Center, m.messageStyle.Render(m.message)))
	s.WriteString("\n\n")

	var content strings.Builder
	switch m.currentScreen {
	case models.HomeScreen:
		content.WriteString("Welcome to Gommits App!\n\n")
		content.WriteString("This application helps you analyze Git commits and export changed files.\n\n")
		content.WriteString(highlightStyle.Render("Features:\n"))
		content.WriteString("• Find commits by specific authors\n")
		content.WriteString("• View detailed commit information\n")
		content.WriteString("• Export changed files to Excel\n")
		content.WriteString("• Stylized terminal output\n\n")
		content.WriteString(modifiedHelpText("start", false, true, false))
	case models.DirectoryScreen:
		content.WriteString(m.textInput.View() + "\n\n")
		content.WriteString(modifiedHelpText("continue", true, true, true))
	case models.AuthorScreen:
		content.WriteString(m.textInput.View() + "\n\n")
		content.WriteString(modifiedHelpText("continue", true, true, false))

	case models.OptionsScreen:
		content.WriteString(m.textInput.View() + "\n\n")
		content.WriteString("Press " + highlightStyle.Render("Enter") + " to fetch commits.\n")
		content.WriteString("Press " + highlightStyle.Render("Tab") + " to toggle current branch only (" + boolToYesNo(m.currentBranchOnly) + ").\n")
		content.WriteString("Press " + highlightStyle.Render("Alt+Tab") + " to toggle show files (" + boolToYesNo(m.showFiles) + ").\n")
		content.WriteString("Press " + highlightStyle.Render("P") + " to edit parent branch (" + m.parentBranch + ").\n")
		content.WriteString(modifiedHelpText("", true, true, false))

	case models.ResultsScreen:
		if len(m.commits) == 0 {
			content.WriteString("No commits found for this author.\n\n")
		} else {
			content.WriteString(fmt.Sprintf("Found %d commits:\n\n", len(m.commits)))

			availableHeight := m.height - 15
			if m.toast.Visible {
				availableHeight -= 4
			}
			if availableHeight < 10 {
				availableHeight = 10
			}

			linesPerCommit := 5
			if m.showFiles {
				linesPerCommit = 7
			}

			maxDisplayCommits := availableHeight / linesPerCommit
			if maxDisplayCommits < 1 {
				maxDisplayCommits = 1
			}
			if maxDisplayCommits > 5 {
				maxDisplayCommits = 5
			}

			displayCount := len(m.commits)
			if displayCount > maxDisplayCommits {
				displayCount = maxDisplayCommits
			}

			for i := 0; i < displayCount; i++ {
				c := m.commits[i]
				content.WriteString(commitHashStyle.Render(fmt.Sprintf("Commit: %s", c.Hash)))
				content.WriteString("\n")
				content.WriteString(fmt.Sprintf("  Author: %s", commitAuthorStyle.Render(c.Author)))
				content.WriteString("\n")
				content.WriteString(fmt.Sprintf("  Date: %s", c.Date))
				content.WriteString("\n")

				message := c.Message
				if len(message) > 60 {
					message = message[:57] + "..."
				}
				content.WriteString(fmt.Sprintf("  Message: %s", message))
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

			if len(m.commits) > displayCount {
				content.WriteString(dimmedStyle.Render(fmt.Sprintf("...and %d more commits\n", len(m.commits)-displayCount)))
			}
		}
		content.WriteString("\n")
		content.WriteString("Press " + highlightStyle.Render("Enter") + " to export to Excel.\n")
		content.WriteString(modifiedHelpText("", true, true, false))
	}

	contentPlaceHeight := m.height - 8 - 3
	if m.toast.Visible {
		contentPlaceHeight -= 4
	}
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

func (m model) renderToast() string {
	if !m.toast.Visible || m.toast.Opacity <= 0 {
		return ""
	}

	var style lipgloss.Style
	switch m.toast.Type {
	case models.ToastSuccess:
		style = toastStyle
	case models.ToastError:
		style = toastErrorStyle
	default:
		style = toastStyle
	}

	if m.toast.Opacity < 1.0 {
		bgR, bgG, bgB := 0x1A, 0x1A, 0x1A

		if m.toast.Type == models.ToastSuccess {
			fgR, fgG, fgB := 0x38, 0xA1, 0x69
			borderR, borderG, borderB := 0x2F, 0x85, 0x5A

			style = style.
				Background(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x",
					int(float64(fgR)*m.toast.Opacity+float64(bgR)*(1-m.toast.Opacity)),
					int(float64(fgG)*m.toast.Opacity+float64(bgG)*(1-m.toast.Opacity)),
					int(float64(fgB)*m.toast.Opacity+float64(bgB)*(1-m.toast.Opacity))))).
				BorderForeground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x",
					int(float64(borderR)*m.toast.Opacity+float64(bgR)*(1-m.toast.Opacity)),
					int(float64(borderG)*m.toast.Opacity+float64(bgG)*(1-m.toast.Opacity)),
					int(float64(borderB)*m.toast.Opacity+float64(bgB)*(1-m.toast.Opacity)))))
		} else {
			fgR, fgG, fgB := 0xE5, 0x3E, 0x3E
			borderR, borderG, borderB := 0xC5, 0x30, 0x30

			style = style.
				Background(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x",
					int(float64(fgR)*m.toast.Opacity+float64(bgR)*(1-m.toast.Opacity)),
					int(float64(fgG)*m.toast.Opacity+float64(bgG)*(1-m.toast.Opacity)),
					int(float64(fgB)*m.toast.Opacity+float64(bgB)*(1-m.toast.Opacity))))).
				BorderForeground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x",
					int(float64(borderR)*m.toast.Opacity+float64(bgR)*(1-m.toast.Opacity)),
					int(float64(borderG)*m.toast.Opacity+float64(bgG)*(1-m.toast.Opacity)),
					int(float64(borderB)*m.toast.Opacity+float64(bgB)*(1-m.toast.Opacity)))))
		}

		textR, textG, textB := 0xFA, 0xFA, 0xFA
		style = style.
			Foreground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x",
				int(float64(textR)*m.toast.Opacity+float64(bgR)*(1-m.toast.Opacity)),
				int(float64(textG)*m.toast.Opacity+float64(bgG)*(1-m.toast.Opacity)),
				int(float64(textB)*m.toast.Opacity+float64(bgB)*(1-m.toast.Opacity)))))
	}

	return style.Render(m.toast.Message)
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
