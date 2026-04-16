package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leeozaka/gommits/internal/git"
	"github.com/leeozaka/gommits/internal/models"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

func errorCmd(err error, context string) tea.Cmd {
	return func() tea.Msg {
		return models.NewError(err, context)
	}
}

type model struct {
	activeScreen ScreenModel
	gitService   git.GitService
	toastManager ToastManager

	directory         string
	author            string
	branch            string
	parentBranch      string
	maxCommits        int
	showFiles         bool
	currentBranchOnly bool
	dotnetMode        bool
	commits           []models.CommitInfo

	message      string
	messageStyle lipgloss.Style
	quitting     bool
	width        int
	height       int
}

func initialModel() model {
	return model{
		activeScreen:      newHomeScreen(),
		gitService:        git.NewCLIGitService(),
		toastManager:      NewToastManager(),
		message:           "Welcome to Gommits App!",
		messageStyle:      infoStyle,
		showFiles:         true,
		currentBranchOnly: true,
		parentBranch:      git.DefaultBranchRef,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case models.ShowToastMsg, models.HideToastMsg, models.TickMsg:
		var cmd tea.Cmd
		m.toastManager, cmd = m.toastManager.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		if msg.Type == tea.KeyEsc {
			if opts, ok := m.activeScreen.(*optionsScreen); ok && opts.editing {
				var cmd tea.Cmd
				m.activeScreen, cmd = m.activeScreen.Update(msg)
				return m, cmd
			}
			m.quitting = true
			return m, tea.Quit
		}

		var cmd tea.Cmd
		m.activeScreen, cmd = m.activeScreen.Update(msg)
		return m, cmd

	case models.ErrorMsg:
		m.message = fmt.Sprintf("Error (%s): %v", msg.Context, msg.Err)
		m.messageStyle = errorStyle
		return m, nil

	case NavigateMsg:
		return m.handleNavigation(msg)

	case models.FetchCommitsMsg:
		if msg.Err != nil {
			return m, errorCmd(msg.Err, "fetching commits")
		}
		m.commits = msg.Commits
		m.branch = msg.Branch
		m.parentBranch = msg.ParentBranch
		m.dotnetMode = msg.DotnetMode
		m.message = fmt.Sprintf("Found %d commits in branch '%s'", len(m.commits), m.branch)
		m.messageStyle = successStyle
		m.activeScreen = newResultsScreen(m.gitService, m.commits, m.directory, m.branch, m.parentBranch, m.showFiles, m.dotnetMode)
		return m, nil

	case models.ExportExcelMsg:
		if msg.Err != nil {
			return m, showToastCmd("❌ Export failed", models.ToastError, 3*time.Second)
		}
		return m, showToastCmd(
			fmt.Sprintf("✅ Exported %d commits to Excel", len(m.commits)),
			models.ToastSuccess, 3*time.Second,
		)

	case models.ResetToHomeMsg:
		m.activeScreen = newHomeScreen()
		m.message = "Welcome to Gommits App!"
		m.messageStyle = infoStyle

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m model) handleNavigation(msg NavigateMsg) (model, tea.Cmd) {
	if msg.Data.Directory != "" {
		m.directory = msg.Data.Directory
	}
	m.author = msg.Data.Author
	if msg.Data.Branch != "" {
		m.branch = msg.Data.Branch
	}
	if msg.Data.ParentBranch != "" {
		m.parentBranch = msg.Data.ParentBranch
	}

	switch msg.To {
	case models.HomeScreen:
		m.activeScreen = newHomeScreen()
		m.message = "Welcome to Gommits App!"
		m.messageStyle = infoStyle

	case models.DirectoryScreen:
		m.activeScreen = newDirectoryScreenWithValue(m.gitService, m.directory)
		m.message = "Please enter the path to a Git repository"
		m.messageStyle = infoStyle

	case models.AuthorScreen:
		m.activeScreen = newAuthorScreenWithValue(m.author)
		m.message = "Enter author(s) to filter, or leave empty for all"
		m.messageStyle = infoStyle

	case models.OptionsScreen:
		m.activeScreen = newOptionsScreenWithValues(
			m.gitService, m.directory, m.author, m.parentBranch,
			m.currentBranchOnly, m.showFiles, m.dotnetMode,
		)
		m.message = "Configure additional options"
		m.messageStyle = infoStyle

	case models.ResultsScreen:
		m.activeScreen = newResultsScreen(m.gitService, m.commits, m.directory, m.branch, m.parentBranch, m.showFiles, m.dotnetMode)
		m.message = fmt.Sprintf("Found %d commits in branch '%s'", len(m.commits), m.branch)
		m.messageStyle = successStyle
	}

	return m, textinput.Blink
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(lipgloss.Place(m.width, 3, lipgloss.Center, lipgloss.Center, titleStyle.Render("Gommits - Commit Analyzer")))
	s.WriteString("\n")
	s.WriteString(lipgloss.Place(m.width, 2, lipgloss.Center, lipgloss.Center, m.messageStyle.Render(m.message)))
	s.WriteString("\n\n")

	content := m.activeScreen.View(m.width, m.height)

	contentPlaceHeight := m.height - 8 - 3
	if contentPlaceHeight < 5 {
		contentPlaceHeight = 5
	}
	s.WriteString(lipgloss.Place(m.width, contentPlaceHeight, lipgloss.Center, lipgloss.Center, content))

	footerText := "Navigation: " +
		highlightStyle.Render("Enter") + " to proceed, " +
		highlightStyle.Render("B") + " for back, " +
		highlightStyle.Render("Esc/Ctrl+C") + " to quit"
	s.WriteString("\n\n")
	s.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Center, dimmedStyle.Render(footerText)))

	screen := s.String()

	if m.toastManager.IsVisible() {
		bg := backgroundViewModel{content: screen}
		fg := toastViewModel{content: m.toastManager.View()}
		o := overlay.New(fg, bg, overlay.Center, overlay.Top, 0, 1)
		return o.View()
	}

	return screen
}

func StartUI() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
