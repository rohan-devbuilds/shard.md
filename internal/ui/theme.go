package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Title     lipgloss.Style
	Metadata  lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Error     lipgloss.Style
	Prompt    lipgloss.Style
	Assistant lipgloss.Style
	Tool      lipgloss.Style
	Border    lipgloss.Style
	Box       lipgloss.Style
}

func NewTheme(name string) Theme {
	intense := name != "minimal"
	borderColor := lipgloss.Color("240")
	accent := lipgloss.Color("44")
	if name == "dark" {
		accent = lipgloss.Color("141")
		borderColor = lipgloss.Color("238")
	}

	title := lipgloss.NewStyle().Foreground(accent)
	if intense {
		title = title.Bold(true)
	}

	return Theme{
		Title:     title,
		Metadata:  lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		Success:   lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		Error:     lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Prompt:    lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(intense),
		Assistant: lipgloss.NewStyle().Foreground(accent).Bold(intense),
		Tool:      lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(intense),
		Border:    lipgloss.NewStyle().Foreground(borderColor),
		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1),
	}
}
