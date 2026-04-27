package models

type CommitInfo struct {
	Hash     string
	Author   string
	Email    string
	Date     string
	Message  string
	Files    []string
	RawFiles []string // original file list before ResolveProjects rewrites Files
}

type DotnetEntry struct {
	Sequence int
	Path     string
	Type     string // "edit", "NEW", or ""
}

type DBAEntry struct {
	Sequence int
	Path     string
}
