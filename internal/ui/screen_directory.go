package ui

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leeozaka/gommits/internal/git"
	"github.com/leeozaka/gommits/internal/models"
)

type directoryScreen struct {
	textInput  textinput.Model
	gitService git.GitService
}

func newDirectoryScreen(svc git.GitService) ScreenModel {
	ti := textinput.New()
	ti.Placeholder = "Enter path to Git repository"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	return &directoryScreen{textInput: ti, gitService: svc}
}

func newDirectoryScreenWithValue(svc git.GitService, value string) ScreenModel {
	ti := textinput.New()
	ti.Placeholder = "Enter path to Git repository"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.SetValue(value)
	return &directoryScreen{textInput: ti, gitService: svc}
}

func (s *directoryScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEnter:
			dir := s.textInput.Value()
			if dir == "" {
				dir = "."
			}
			absDir, err := filepath.Abs(dir)
			if err != nil {
				return s, errorCmd(err, "resolving directory path")
			}

			if !s.gitService.IsGitRepo(absDir) {
				return s, errorCmd(fmt.Errorf("%s is not a Git repository", absDir), "validating repository")
			}

			branchName, err := s.gitService.GetCurrentBranch(absDir)
			if err != nil {
				return s, errorCmd(err, "getting branch name")
			}

			parentBranch := s.gitService.DetectDefaultBranch(absDir)

			return s, func() tea.Msg {
				return NavigateMsg{
					To: models.AuthorScreen,
					Data: NavigateData{
						Directory:    absDir,
						Branch:       branchName,
						ParentBranch: parentBranch,
					},
				}
			}

		case tea.KeyTab:
			s.textInput.SetValue(".")
			return s, nil

		case tea.KeyRunes:
			if string(keyMsg.Runes) == "b" {
				return s, func() tea.Msg {
					return NavigateMsg{To: models.HomeScreen}
				}
			}
		}
	}

	var cmd tea.Cmd
	s.textInput, cmd = s.textInput.Update(msg)
	return s, cmd
}

func (s *directoryScreen) View(width, height int) string {
	return s.textInput.View() + "\n\n" + modifyHelpText("continue", true, true, true)
}
