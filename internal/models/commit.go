package models

type CommitInfo struct {
	Hash    string
	Author  string
	Email   string
	Date    string
	Message string
	Files   []string
}

type DotnetEntry struct {
	Sequence int
	Path     string
	Type     string // "edit", "NEW", or ""
}
