package models

import (
	"time"
)

type Screen int

const (
	HomeScreen Screen = iota
	DirectoryScreen
	AuthorScreen
	OptionsScreen
	ResultsScreen
)

type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
)

type Toast struct {
	Message   string
	Type      ToastType
	Visible   bool
	Opacity   float64
	Position  float64
	StartTime time.Time
	Duration  time.Duration
}

type FetchCommitsMsg struct {
	Commits      []CommitInfo
	Branch       string
	ParentBranch string
	DotnetMode   bool
	Err          error
}

type ExportExcelMsg struct {
	Path string
	Err  error
}

type ResetToHomeMsg struct{}

type ShowToastMsg struct {
	Message  string
	Type     ToastType
	Duration time.Duration
}

type HideToastMsg struct{}

type ErrorMsg struct {
	Err     error
	Context string
}

func NewError(err error, context string) ErrorMsg {
	return ErrorMsg{Err: err, Context: context}
}

type TickMsg time.Time
