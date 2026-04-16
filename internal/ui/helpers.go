package ui

import (
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/leeozaka/gommits/internal/git"
	"github.com/leeozaka/gommits/internal/models"
	"github.com/leeozaka/gommits/pkg/utils"
)

type authorResult struct {
	commits []models.CommitInfo
	branch  string
	err     error
}

func fetchCommitsCmd(svc git.GitService, dir, author string, maxCommits int, currentBranchOnly bool, parentBranch string, dotnetMode bool) tea.Cmd {
	return func() tea.Msg {
		authors := splitAuthors(author)

		var allCommits []models.CommitInfo
		var branch string
		var err error

		if len(authors) == 0 {
			allCommits, branch, err = svc.GatherCommits(dir, "", parentBranch, currentBranchOnly)
		} else if len(authors) == 1 {
			allCommits, branch, err = svc.GatherCommits(dir, authors[0], parentBranch, currentBranchOnly)
		} else {
			results := make([]authorResult, len(authors))
			var wg sync.WaitGroup
			wg.Add(len(authors))

			for i, a := range authors {
				go func(idx int, authorName string) {
					defer wg.Done()
					c, b, e := svc.GatherCommits(dir, authorName, parentBranch, currentBranchOnly)
					results[idx] = authorResult{commits: c, branch: b, err: e}
				}(i, a)
			}

			wg.Wait()

			seen := make(map[string]bool)
			for _, r := range results {
				if r.err != nil {
					err = r.err
					break
				}
				if r.branch != "" {
					branch = r.branch
				}
				for _, c := range r.commits {
					if !seen[c.Hash] {
						seen[c.Hash] = true
						allCommits = append(allCommits, c)
					}
				}
			}
		}

		if err == nil && maxCommits > 0 && len(allCommits) > maxCommits {
			allCommits = allCommits[:maxCommits]
		}
		if err == nil && dotnetMode {
			allCommits = utils.ResolveProjects(dir, allCommits)
		}
		return models.FetchCommitsMsg{
			Commits:      allCommits,
			Branch:       branch,
			ParentBranch: parentBranch,
			DotnetMode:   dotnetMode,
			Err:          err,
		}
	}
}

func splitAuthors(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func exportExcelCmd(svc git.GitService, commits []models.CommitInfo, repoPath string) tea.Cmd {
	return func() tea.Msg {
		repoName := svc.GetRepositoryName(repoPath)
		err := utils.ExportToExcel(commits, repoPath, repoName)
		return models.ExportExcelMsg{Path: repoPath, Err: err}
	}
}

func exportDotnetExcelCmd(svc git.GitService, commits []models.CommitInfo, repoPath, branch, parentBranch string) tea.Cmd {
	return func() tea.Msg {
		repoName := svc.GetRepositoryName(repoPath)

		existsInParent := func(path string) bool {
			if svc.PathExistsInRef(repoPath, parentBranch, path) {
				return true
			}
			return svc.PathExistsInRef(repoPath, "origin/"+parentBranch, path)
		}

		entries := utils.AggregateDotnetEntries(commits, branch, existsInParent)
		err := utils.ExportDotnetExcel(entries, repoPath, repoName)
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
