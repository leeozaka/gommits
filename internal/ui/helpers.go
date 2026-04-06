package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leeozaka/gommits/internal/git"
	"github.com/leeozaka/gommits/internal/models"
	"github.com/leeozaka/gommits/pkg/utils"
)

func fetchCommitsCmd(svc git.GitService, dir, author string, maxCommits int, currentBranchOnly bool, parentBranch string, dotnetMode bool) tea.Cmd {
	return func() tea.Msg {
		commits, branch, err := svc.GatherCommits(dir, author, parentBranch, currentBranchOnly)
		if err == nil && maxCommits > 0 && len(commits) > maxCommits {
			commits = commits[:maxCommits]
		}
		if err == nil && dotnetMode {
			commits = utils.ResolveProjects(dir, commits)
		}
		return models.FetchCommitsMsg{Commits: commits, Branch: branch, Err: err}
	}
}

func exportExcelCmd(svc git.GitService, commits []models.CommitInfo, repoPath string) tea.Cmd {
	return func() tea.Msg {
		repoName := svc.GetRepositoryName(repoPath)
		err := utils.ExportToExcel(commits, repoPath, repoName)
		return models.ExportExcelMsg{Path: repoPath, Err: err}
	}
}

func resetToHomeCmd() tea.Cmd {
	return func() tea.Msg {
		return models.ResetToHomeMsg{}
	}
}

func showToastCmd(message string, toastType models.ToastType, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return models.ShowToastMsg{Message: message, Type: toastType, Duration: duration}
	}
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func modifyHelpText(enterAction string, includeBack bool, includeQuit bool, showTabHint bool) string {
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
