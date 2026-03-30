package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme defines the color palette and base styles for all toph UI panels.
// Colors follow the design spec: terminal-default background, soft blues for
// focus, green/amber/red for status, and distinct hues for event types.
type Theme struct {
	BorderNormal color.Color
	BorderFocus  color.Color

	Active  color.Color
	Waiting color.Color
	Error   color.Color
	Idle    color.Color

	ToolUse   color.Color
	FileWrite color.Color
	Subagent  color.Color
	UserMsg   color.Color
	System    color.Color
	TextDim   color.Color

	ProgressLow  color.Color
	ProgressMid  color.Color
	ProgressHigh color.Color

	PanelNormal lipgloss.Style
	PanelFocus  lipgloss.Style
	Title       lipgloss.Style
	StatusBar   lipgloss.Style
}

// DefaultTheme returns the standard toph color theme. All hex values match the
// design spec in CLAUDE.md — rounded borders, soft blue focus ring, green/amber/red
// status indicators, and event-type colors for the activity feed.
func DefaultTheme() *Theme {
	t := &Theme{
		BorderNormal: lipgloss.Color("#585858"),
		BorderFocus:  lipgloss.Color("#87AFFF"),

		Active:  lipgloss.Color("#87D787"),
		Waiting: lipgloss.Color("#FFD787"),
		Error:   lipgloss.Color("#FF8787"),
		Idle:    lipgloss.Color("#6C6C6C"),

		ToolUse:   lipgloss.Color("#87D7D7"),
		FileWrite: lipgloss.Color("#87D787"),
		Subagent:  lipgloss.Color("#D7AFFF"),
		UserMsg:   lipgloss.Color("#87AFFF"),
		System:    lipgloss.Color("#6C6C6C"),
		TextDim:   lipgloss.Color("#6C6C6C"),

		ProgressLow:  lipgloss.Color("#87D787"),
		ProgressMid:  lipgloss.Color("#FFD787"),
		ProgressHigh: lipgloss.Color("#FF8787"),
	}

	t.PanelNormal = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderNormal).
		Padding(0, 1)

	t.PanelFocus = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocus).
		Padding(0, 1)

	t.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.BorderFocus)

	t.StatusBar = lipgloss.NewStyle().
		Foreground(t.TextDim)

	return t
}
