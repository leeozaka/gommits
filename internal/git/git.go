package git

import (
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/leeozaka/gommits/internal/models"
)

const (
	OriginPrefix     = "origin/"
	DefaultBranchRef = "main"
	GitDelimiter     = "|"
	LogFormat        = "%H" + GitDelimiter + "%an" + GitDelimiter + "%ae" + GitDelimiter + "%ad" + GitDelimiter + "%s"
	LogFieldCount    = 5
	HeadBranchPrefix = "HEAD branch:"
)

var defaultBranchCandidates = []string{"main", "master", "trunk", "development", "dev"}

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

func GetRepositoryName(path string) string {
	output, err := execGit(path, "remote", "get-url", "origin")
	if err != nil {
		return filepath.Base(path)
	}

	raw := strings.TrimSpace(output)
	raw = strings.TrimSuffix(raw, ".git")

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		if u, err := url.Parse(raw); err == nil {
			if name := filepath.Base(u.Path); name != "" && name != "." {
				return name
			}
		}
	}

	if idx := strings.LastIndex(raw, ":"); idx != -1 {
		sshPath := raw[idx+1:]
		if name := filepath.Base(sshPath); name != "" && name != "." {
			return name
		}
	}

	return filepath.Base(path)
}

func GatherCommits(path, authorInput, parentBranch string, currentBranchOnly bool) ([]models.CommitInfo, string, error) {
	currentBranch, err := GetCurrentBranch(path)
	if err != nil {
		return nil, "", err
	}

	args := []string{"log",
		"--pretty=format:" + LogFormat,
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
		if refExists(path, OriginPrefix+parentBranch) {
			parentBranch = OriginPrefix + parentBranch
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

		parts := strings.SplitN(line, GitDelimiter, LogFieldCount)
		if len(parts) < LogFieldCount {
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
	for _, branch := range defaultBranchCandidates {
		if refExists(path, branch) {
			return branch
		}
		if refExists(path, OriginPrefix+branch) {
			return OriginPrefix + branch
		}
	}

	if output, err := execGit(path, "remote", "show", "origin"); err == nil {
		for line := range strings.SplitSeq(output, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, HeadBranchPrefix) {
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

	return DefaultBranchRef
}
