package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/leeozaka/gommits/internal/git"
	"github.com/leeozaka/gommits/internal/models"
)

type optionsScreen struct {
	textInput         textinput.Model
	gitService        git.GitService
	directory         string
	author            string
	parentBranch      string
	currentBranchOnly bool
	showFiles         bool
	dotnetMode        bool
	editing           bool
	editingField      string
}

func newOptionsScreen(svc git.GitService, directory, author, parentBranch string) ScreenModel {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Width = 50
	ti.Blur()
	return &optionsScreen{
		textInput:         ti,
		gitService:        svc,
		directory:         directory,
		author:            author,
		parentBranch:      parentBranch,
		currentBranchOnly: true,
		showFiles:         true,
	}
}

func newOptionsScreenWithValues(svc git.GitService, directory, author, parentBranch string, currentBranchOnly, showFiles, dotnetMode bool) ScreenModel {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Width = 50
	ti.Blur()
	return &optionsScreen{
		textInput:         ti,
		gitService:        svc,
		directory:         directory,
		author:            author,
		parentBranch:      parentBranch,
		currentBranchOnly: currentBranchOnly,
		showFiles:         showFiles,
		dotnetMode:        dotnetMode,
	}
}

func (s *optionsScreen) startEditing(field, placeholder, value string) tea.Cmd {
	s.editing = true
	s.editingField = field
	s.textInput.Placeholder = placeholder
	s.textInput.SetValue(value)
	s.textInput.Focus()
	return textinput.Blink
}

func (s *optionsScreen) stopEditing() {
	s.editing = false
	s.editingField = ""
	s.textInput.Blur()
	s.textInput.SetValue("")
}

func (s *optionsScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}

	if s.editing {
		switch keyMsg.Type {
		case tea.KeyEnter:
			val := s.textInput.Value()
			switch s.editingField {
			case "parentBranch":
				if val != "" {
					s.parentBranch = val
				}
			case "maxCommits":
				maxCommits := 0
				if val != "" {
					fmt.Sscanf(val, "%d", &maxCommits)
					if maxCommits < 0 {
						maxCommits = 0
					}
				}
				s.stopEditing()
				return s, fetchCommitsCmd(s.gitService, s.directory, s.author, maxCommits, s.currentBranchOnly, s.parentBranch, s.dotnetMode)
			}
			s.stopEditing()
			return s, nil
		case tea.KeyEsc:
			s.stopEditing()
			return s, nil
		}

		var cmd tea.Cmd
		s.textInput, cmd = s.textInput.Update(msg)
		return s, cmd
	}

	switch keyMsg.Type {
	case tea.KeyEnter:
		return s, fetchCommitsCmd(s.gitService, s.directory, s.author, 0, s.currentBranchOnly, s.parentBranch, s.dotnetMode)

	case tea.KeyTab:
		if keyMsg.Alt {
			s.showFiles = !s.showFiles
		} else {
			s.currentBranchOnly = !s.currentBranchOnly
		}
		return s, nil

	case tea.KeyRunes:
		key := string(keyMsg.Runes)
		switch key {
		case "d":
			s.dotnetMode = !s.dotnetMode
		case "p":
			return s, s.startEditing("parentBranch", "Enter parent branch name", s.parentBranch)
		case "m":
			return s, s.startEditing("maxCommits", "Enter maximum number of commits (0 for no limit)", "0")
		case "b":
			return s, func() tea.Msg {
				return NavigateMsg{To: models.AuthorScreen}
			}
		}
	}

	return s, nil
}

func (s *optionsScreen) View(width, height int) string {
	var content string

	if s.editing {
		content += s.textInput.View() + "\n"
		content += dimmedStyle.Render("Press Enter to confirm, Esc to cancel.") + "\n\n"
		return content
	}

	content += "Press " + highlightStyle.Render("Enter") + " to fetch commits.\n"
	content += "Press " + highlightStyle.Render("M") + " to set max commits.\n"
	content += "Press " + highlightStyle.Render("P") + " to edit parent branch (" + s.parentBranch + ").\n"
	content += "Press " + highlightStyle.Render("Tab") + " to toggle current branch only (" + boolToYesNo(s.currentBranchOnly) + ").\n"
	content += "Press " + highlightStyle.Render("Alt+Tab") + " to toggle show files (" + boolToYesNo(s.showFiles) + ").\n"
	content += "Press " + highlightStyle.Render("D") + " to toggle dotnet project mode (" + boolToYesNo(s.dotnetMode) + ").\n"
	authorDisplay := s.author
	if authorDisplay == "" {
		authorDisplay = "all authors"
	}
	content += dimmedStyle.Render("Author filter: "+authorDisplay) + "\n"
	content += modifyHelpText("", true, true, false)
	return content
}
