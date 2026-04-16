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
	commitSeparator  = "---COMMIT_SEP---"
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

	logFmt := commitSeparator + "\n" + LogFormat

	args := []string{"log",
		"--pretty=format:" + logFmt,
		"--name-only",
	}

	if authorInput != "" {
		args = append(args, "--author="+authorInput)
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

	commits := parseCommits(output)
	return commits, currentBranch, nil
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

func parseCommits(output string) []models.CommitInfo {
	if output == "" {
		return nil
	}

	blocks := strings.Split(output, commitSeparator)
	results := make([]models.CommitInfo, 0, len(blocks))

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		lines := strings.SplitN(block, "\n", 2)
		metaLine := strings.TrimSpace(lines[0])
		if metaLine == "" {
			continue
		}

		parts := strings.SplitN(metaLine, GitDelimiter, LogFieldCount)
		if len(parts) < LogFieldCount {
			continue
		}

		var files []string
		if len(lines) > 1 {
			for _, f := range strings.Split(strings.TrimSpace(lines[1]), "\n") {
				f = strings.TrimSpace(f)
				if f != "" {
					files = append(files, f)
				}
			}
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

	return results
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

func PathExistsInRef(repoPath, ref, targetPath string) bool {
	targetPath = strings.ReplaceAll(targetPath, "\\", "/")
	output, err := execGit(repoPath, "cat-file", "-t", ref+":"+targetPath)
	return err == nil && strings.TrimSpace(output) != ""
}
