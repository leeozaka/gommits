package git

import (
	"os/exec"
	"strings"

	"github.com/leeozaka/gommits/internal/models"
)

func execGit(path string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", path}, args...)
	cmd := exec.Command("git", fullArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func refExists(path, ref string) bool {
	_, err := execGit(path, "rev-parse", "--verify", ref)
	return err == nil
}

func IsGitRepo(path string) bool {
	output, err := execGit(path, "rev-parse", "--is-inside-work-tree")
	return err == nil && output == "true"
}

func GetCurrentBranch(path string) (string, error) {
	return execGit(path, "rev-parse", "--abbrev-ref", "HEAD")
}

func GatherCommits(path, authorInput, parentBranch string, currentBranchOnly bool) ([]models.CommitInfo, string, error) {
	currentBranch, err := GetCurrentBranch(path)
	if err != nil {
		return nil, "", err
	}

	args := []string{"log",
		"--pretty=format:%H|%an|%ae|%ad|%s",
		"--author=" + authorInput,
	}

	if currentBranchOnly {
		args = append(args, getCommitRange(path, currentBranch, parentBranch))
	} else {
		args = append(args, "--all")
	}

	output, err := execGit(path, args...)
	if err != nil {
		return nil, "", err
	}

	commits, err := parseCommits(path, output)
	return commits, currentBranch, err
}

func getCommitRange(path, currentBranch, parentBranch string) string {
	if !refExists(path, parentBranch) {
		if refExists(path, "origin/"+parentBranch) {
			parentBranch = "origin/" + parentBranch
		} else {
			return currentBranch
		}
	}

	mergeBase, err := execGit(path, "merge-base", currentBranch, parentBranch)
	if err != nil {
		return currentBranch
	}
	return mergeBase + ".." + currentBranch
}

func parseCommits(path, output string) ([]models.CommitInfo, error) {
	lines := strings.Split(output, "\n")
	results := make([]models.CommitInfo, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		files, err := GetChangedFiles(path, parts[0])
		if err != nil {
			return nil, err
		}

		results = append(results, models.CommitInfo{
			Hash:    parts[0],
			Author:  parts[1],
			Email:   parts[2],
			Date:    parts[3],
			Message: parts[4],
			Files:   files,
		})
	}

	return results, nil
}

func GetChangedFiles(path, commitHash string) ([]string, error) {
	output, err := execGit(path, "show", "--name-only", "--pretty=", commitHash)
	if err != nil {
		return nil, err
	}

	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

func DetectDefaultBranch(path string) string {
	for _, branch := range []string{"main", "master", "trunk", "development", "dev"} {
		if refExists(path, branch) {
			return branch
		}
		if refExists(path, "origin/"+branch) {
			return "origin/" + branch
		}
	}

	if output, err := execGit(path, "remote", "show", "origin"); err == nil {
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "HEAD branch:") {
				if parts := strings.SplitN(line, ":", 2); len(parts) > 1 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}

	if output, err := execGit(path, "branch"); err == nil && output != "" {
		if lines := strings.Split(output, "\n"); len(lines) > 0 {
			if branch := strings.TrimSpace(strings.TrimPrefix(lines[0], "*")); branch != "" {
				return branch
			}
		}
	}

	return "main"
}
