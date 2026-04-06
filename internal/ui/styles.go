package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1).
			Align(lipgloss.Center).
			Width(60)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#2D3748")).
			Padding(0, 1).
			Align(lipgloss.Center).
			Width(60)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#E53E3E")).
			Padding(0, 1).
			Align(lipgloss.Center).
			Width(60)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#38A169")).
			Padding(0, 1).
			Align(lipgloss.Center).
			Width(60)

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9E9E9E"))

	commitHashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2D3748")).
			Bold(true)

	commitAuthorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#38A169"))

	commitFilesStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4"))
)
