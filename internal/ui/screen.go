package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leeozaka/gommits/internal/git"
	"github.com/leeozaka/gommits/internal/models"
)

type ScreenModel interface {
	Update(msg tea.Msg) (ScreenModel, tea.Cmd)
	View(width, height int) string
}

type NavigateMsg struct {
	To   models.Screen
	Data NavigateData
}

type NavigateData struct {
	Directory    string
	Author       string
	Branch       string
	ParentBranch string
	MaxCommits   int
	GitService   git.GitService
	MessageStyle lipgloss.Style
	Message      string
}
