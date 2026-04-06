package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leeozaka/gommits/internal/models"
)

type authorScreen struct {
	textInput textinput.Model
}

func newAuthorScreen() ScreenModel {
	ti := textinput.New()
	ti.Placeholder = "Enter author name or email"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	return &authorScreen{textInput: ti}
}

func newAuthorScreenWithValue(value string) ScreenModel {
	ti := textinput.New()
	ti.Placeholder = "Enter author name or email"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.SetValue(value)
	return &authorScreen{textInput: ti}
}

func (s *authorScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEnter:
			author := s.textInput.Value()
			if author == "" {
				return s, errorCmd(fmt.Errorf("author name cannot be empty"), "author input")
			}
			return s, func() tea.Msg {
				return NavigateMsg{
					To:   models.OptionsScreen,
					Data: NavigateData{Author: author},
				}
			}

		case tea.KeyRunes:
			if string(keyMsg.Runes) == "b" {
				return s, func() tea.Msg {
					return NavigateMsg{To: models.DirectoryScreen}
				}
			}
		}
	}

	var cmd tea.Cmd
	s.textInput, cmd = s.textInput.Update(msg)
	return s, cmd
}

func (s *authorScreen) View(width, height int) string {
	return s.textInput.View() + "\n\n" + modifyHelpText("continue", true, true, false)
}
