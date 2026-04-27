package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/leeozaka/gommits/internal/models"
)

// sqlFileRe matches paths like DBA/<anything>/<timestamp12>_(up|down).sql
// The DBA/ prefix is case-insensitive; the suffix casing is kept per convention.
var sqlFileRe = regexp.MustCompile(`(?i)^DBA/[^/]+/(\d{12})_(up|down)\.sql$`)

// AggregateDBAEntries collects unique .sql files from the DBA/ subtree across all commits,
// sorts up-scripts ascending and down-scripts descending by their filename timestamp prefix,
// and returns them as sequenced DBAEntry slices ready for the Excel writer.
//
// Each entry path is formatted as "<year>/INCxxxxxxx/<basename>" where INCxxxxxxx is a
// literal placeholder the user replaces manually after export.
//
// Files that do not match the expected naming pattern are silently dropped.
func AggregateDBAEntries(commits []models.CommitInfo, year int) (up, down []models.DBAEntry) {
	type sqlFile struct {
		basename  string
		timestamp string // 12-digit prefix for sorting
	}

	seenUp := make(map[string]bool)
	seenDown := make(map[string]bool)

	var upFiles, downFiles []sqlFile

	for _, commit := range commits {
		files := commit.RawFiles
		if len(files) == 0 {
			files = commit.Files
		}
		for _, f := range files {
			normalized := filepath.ToSlash(f)
			m := sqlFileRe.FindStringSubmatch(normalized)
			if m == nil {
				continue
			}
			ts := m[1]
			direction := m[2] // "up" or "down" (already lowercased by regex engine? no — need toLower)
			// The (?i) flag only covers "DBA"; capture group 2 is literal from filename.
			// Filenames are expected to be lowercase "up"/"down", but normalise just in case.
			basename := filepath.Base(normalized)

			if strings.EqualFold(direction, "up") {
				if !seenUp[basename] {
					seenUp[basename] = true
					upFiles = append(upFiles, sqlFile{basename: basename, timestamp: ts})
				}
			} else {
				if !seenDown[basename] {
					seenDown[basename] = true
					downFiles = append(downFiles, sqlFile{basename: basename, timestamp: ts})
				}
			}
		}
	}

	sort.Slice(upFiles, func(i, j int) bool {
		return upFiles[i].timestamp < upFiles[j].timestamp // ascending
	})
	sort.Slice(downFiles, func(i, j int) bool {
		return downFiles[i].timestamp > downFiles[j].timestamp // descending
	})

	for i, f := range upFiles {
		up = append(up, models.DBAEntry{
			Sequence: i + 1,
			Path:     fmt.Sprintf("%d/INCxxxxxxx/%s", year, f.basename),
		})
	}
	for i, f := range downFiles {
		down = append(down, models.DBAEntry{
			Sequence: i + 1,
			Path:     fmt.Sprintf("%d/INCxxxxxxx/%s", year, f.basename),
		})
	}

	return up, down
}
