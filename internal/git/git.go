package git

import (
	"os/exec"
	"strings"

	"github.com/leeozaka/gommits/internal/models"
)

func IsGitRepo(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

func GetCurrentBranch(path string) (string, error) {
	branchCmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(branchOutput)), nil
}

func GatherCommits(path, authorInput string, currentBranchOnly bool) ([]models.CommitInfo, string, error) {
	var results []models.CommitInfo

	branchCmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.CombinedOutput()
	if err != nil {
		return nil, "", err
	}
	currentBranch := strings.TrimSpace(string(branchOutput))

	args := []string{"-C", path, "log",
		"--pretty=format:%H|%an|%ae|%ad|%s",
		"--author=" + authorInput,
	}

	if currentBranchOnly {
		args = append(args, currentBranch)
	} else {
		args = append(args, "--all")
	}

	cmd := exec.Command("git", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "", err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}
		commitHash := parts[0]
		commitAuthor := parts[1]
		commitEmail := parts[2]
		commitDate := parts[3]
		commitMsg := parts[4]

		files, err := GetChangedFiles(path, commitHash)
		if err != nil {
			return nil, "", err
		}

		results = append(results, models.CommitInfo{
			Hash:    commitHash,
			Author:  commitAuthor,
			Email:   commitEmail,
			Date:    commitDate,
			Message: commitMsg,
			Files:   files,
		})
	}

	return results, currentBranch, nil
}

func GetChangedFiles(path, commitHash string) ([]string, error) {
	cmd := exec.Command("git", "-C", path, "show", "--name-only", "--pretty=", commitHash)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}
	return files, nil
}
