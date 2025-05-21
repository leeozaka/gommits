package models

type CommitInfo struct {
	Hash    string
	Author  string
	Email   string
	Date    string
	Message string
	Files   []string
}
