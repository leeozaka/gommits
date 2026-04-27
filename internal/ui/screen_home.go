package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leeozaka/gommits/internal/models"
)

type homeScreen struct{}

func newHomeScreen() ScreenModel {
	return homeScreen{}
}

func (s homeScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && msg.Type == tea.KeyEnter {
		return s, func() tea.Msg {
			return NavigateMsg{To: models.DirectoryScreen}
		}
	}
	return s, nil
}

func (s homeScreen) View(width, height int) string {
	var content string
	content += "Welcome to Gommits App!\n\n"
	content += "This application helps you analyze Git commits and export changed files.\n\n"
	content += highlightStyle.Render("Features:\n")
	content += "• Find commits by specific authors\n"
	content += "• View detailed commit information\n"
	content += "• Export changed files to Excel\n"
	content += "• Stylized terminal output\n\n"
	content += modifyHelpText("start", false, true, false)
	return content
}
