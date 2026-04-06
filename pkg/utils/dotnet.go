package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/leeozaka/gommits/internal/models"
)

// ResolveProjects walks up from each changed file's directory to find the nearest .csproj,
// returning the relative directory path of that .csproj. Files without a matching .csproj are omitted.
// Results are deduplicated per commit.
func ResolveProjects(repoPath string, commits []models.CommitInfo) []models.CommitInfo {
	resolved := make([]models.CommitInfo, len(commits))
	copy(resolved, commits)

	cache := make(map[string]string) // dir -> project relative path ("" means none found)

	for i, commit := range resolved {
		seen := make(map[string]bool)
		var projects []string

		for _, file := range commit.Files {
			dir := filepath.Dir(file)
			project := findProject(repoPath, dir, cache)
			if project != "" && !seen[project] {
				seen[project] = true
				projects = append(projects, project)
			}
		}

		resolved[i].Files = projects
	}

	return resolved
}

func findProject(repoPath, relDir string, cache map[string]string) string {
	if result, ok := cache[relDir]; ok {
		return result
	}

	absRepo, _ := filepath.Abs(repoPath)
	current := relDir

	for {
		absDir := filepath.Join(absRepo, current)

		entries, err := os.ReadDir(absDir)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".csproj") {
					cache[relDir] = current
					return current
				}
			}
		}

		if current == "." || current == "" {
			break
		}

		current = filepath.Dir(current)
		if current == "." || current == "/" {
			// Check root level too
			absRoot := absRepo
			entries, err := os.ReadDir(absRoot)
			if err == nil {
				for _, e := range entries {
					if !e.IsDir() && strings.HasSuffix(e.Name(), ".csproj") {
						cache[relDir] = "."
						return "."
					}
				}
			}
			break
		}
	}

	cache[relDir] = ""
	return ""
}
