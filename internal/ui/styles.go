package ui

import "github.com/charmbracelet/lipgloss"

var (
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))
	greaterStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("63")).
			Align(lipgloss.Center)
)
