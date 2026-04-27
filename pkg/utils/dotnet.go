package utils

import (
	"os"
	"path/filepath"
	"sort"
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

		resolved[i].RawFiles = commit.Files
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

// AggregateDotnetEntries collects unique service-level project groups from resolved commits,
// determines if each is "edit" or "NEW", and returns sequenced entries with branch-prefixed paths.
// Excludes infrastructure/shared projects and test projects. Sorts edits first, then NEW.
func AggregateDotnetEntries(commits []models.CommitInfo, branch string, existsInParent func(string) bool) []models.DotnetEntry {
	serviceGroups := make(map[string]bool)

	for _, commit := range commits {
		for _, project := range commit.Files {
			if project == "" {
				continue
			}
			parent := filepath.Dir(project)
			if parent == "." || parent == "" {
				parent = project
			}
			if shouldExclude(parent) {
				continue
			}
			serviceGroups[parent] = true
		}
	}

	groups := make([]string, 0, len(serviceGroups))
	for g := range serviceGroups {
		groups = append(groups, g)
	}
	sort.Strings(groups)

	type classified struct {
		path      string
		entryType string
	}

	var edits, news []classified
	for _, group := range groups {
		entryType := ""
		if existsInParent != nil {
			if existsInParent(group) {
				entryType = "edit"
			} else {
				entryType = "NEW"
			}
		}

		entryPath := filepath.ToSlash(group)
		if branch != "" {
			entryPath = branch + "/" + entryPath
		}

		c := classified{path: entryPath, entryType: entryType}
		if entryType == "NEW" {
			news = append(news, c)
		} else {
			edits = append(edits, c)
		}
	}

	ordered := append(edits, news...)
	entries := make([]models.DotnetEntry, len(ordered))
	for i, c := range ordered {
		entries[i] = models.DotnetEntry{
			Sequence: i + 1,
			Path:     c.path,
			Type:     c.entryType,
		}
	}

	return entries
}

var excludedSegments = []string{
	"Core",
	"Infrastructure",
	"Migrations",
}

func shouldExclude(path string) bool {
	normalized := filepath.ToSlash(path)
	segments := strings.Split(normalized, "/")

	for _, seg := range segments {
		if strings.Contains(strings.ToLower(seg), ".tests") || strings.HasSuffix(strings.ToLower(seg), "tests") {
			return true
		}
		for _, excluded := range excludedSegments {
			if strings.EqualFold(seg, excluded) {
				return true
			}
		}
	}

	return false
}
