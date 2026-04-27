package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leeozaka/gommits/internal/git"
	"github.com/leeozaka/gommits/internal/models"
)

type resultsScreen struct {
	gitService   git.GitService
	commits      []models.CommitInfo
	directory    string
	branch       string
	parentBranch string
	showFiles    bool
	dotnetMode   bool
}

func newResultsScreen(svc git.GitService, commits []models.CommitInfo, directory, branch, parentBranch string, showFiles, dotnetMode bool) ScreenModel {
	return &resultsScreen{
		gitService:   svc,
		commits:      commits,
		directory:    directory,
		branch:       branch,
		parentBranch: parentBranch,
		showFiles:    showFiles,
		dotnetMode:   dotnetMode,
	}
}

func (s *resultsScreen) Update(msg tea.Msg) (ScreenModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEnter:
			if s.dotnetMode {
				return s, exportDotnetExcelCmd(s.gitService, s.commits, s.directory, s.branch, s.parentBranch)
			}
			return s, exportExcelCmd(s.gitService, s.commits, s.directory)

		case tea.KeyRunes:
			if string(keyMsg.Runes) == "b" {
				return s, func() tea.Msg {
					return NavigateMsg{To: models.OptionsScreen}
				}
			}
		}
	}
	return s, nil
}

func (s *resultsScreen) View(width, height int) string {
	var content strings.Builder

	if len(s.commits) == 0 {
		content.WriteString("No commits found for this author.\n\n")
	} else {
		content.WriteString(fmt.Sprintf("Found %d commits:\n\n", len(s.commits)))

		availableHeight := height - 15
		if availableHeight < 10 {
			availableHeight = 10
		}

		linesPerCommit := 5
		if s.showFiles {
			linesPerCommit = 7
		}

		maxDisplayCommits := availableHeight / linesPerCommit
		if maxDisplayCommits < 1 {
			maxDisplayCommits = 1
		}
		if maxDisplayCommits > 5 {
			maxDisplayCommits = 5
		}

		displayCount := len(s.commits)
		if displayCount > maxDisplayCommits {
			displayCount = maxDisplayCommits
		}

		for i := 0; i < displayCount; i++ {
			c := s.commits[i]
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

			if s.showFiles && len(c.Files) > 0 {
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

		if len(s.commits) > displayCount {
			content.WriteString(dimmedStyle.Render(fmt.Sprintf("...and %d more commits\n", len(s.commits)-displayCount)))
		}
	}
	content.WriteString("\n")
	content.WriteString("Press " + highlightStyle.Render("Enter") + " to export to Excel.\n")
	content.WriteString(modifyHelpText("", true, true, false))

	return content.String()
}
