package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

// StatusBar renders the bottom status line with keybinding hints on the left
// and connection/source info on the right. It spans the full terminal width.
type StatusBar struct {
	theme *Theme
}

// NewStatusBar creates a StatusBar bound to the given theme.
func NewStatusBar(theme *Theme) *StatusBar {
	return &StatusBar{theme: theme}
}

// RenderFilter draws the status bar in filter mode. When editing is true a
// block cursor is shown after the text; when false the applied filter is
// displayed with a hint on how to clear it.
func (sb *StatusBar) RenderFilter(width int, filterText string, editing bool) string {
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(sb.theme.BorderFocus)
	descStyle := lipgloss.NewStyle().Foreground(sb.theme.TextDim)

	cursor := ""
	if editing {
		cursor = "\u2588" // block cursor
	}
	left := fmt.Sprintf("%s %s%s", keyStyle.Render("/"), filterText, cursor)

	var right string
	if editing {
		right = fmt.Sprintf("%s %s  %s %s",
			keyStyle.Render("esc"), descStyle.Render("clear"),
			keyStyle.Render("enter"), descStyle.Render("apply"),
		)
	} else {
		right = fmt.Sprintf("%s %s  %s %s",
			keyStyle.Render("/"), descStyle.Render("edit filter"),
			keyStyle.Render("esc"), descStyle.Render("clear"),
		)
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return fmt.Sprintf("%s%*s%s", left, gap, "", right)
}

// Render draws the status bar at the given width. activeCount is the number
// of currently active sessions, source is the data source label (e.g.
// "~/.claude/projects"), and connected indicates whether the source is live.
func (sb *StatusBar) Render(width int, activeCount int, source string, connected bool) string {
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(sb.theme.BorderFocus)
	descStyle := lipgloss.NewStyle().Foreground(sb.theme.TextDim)

	left := fmt.Sprintf("%s %s  %s %s  %s %s  %s %s",
		keyStyle.Render("tab"), descStyle.Render("panels"),
		keyStyle.Render("j/k"), descStyle.Render("navigate"),
		keyStyle.Render("?"), descStyle.Render("help"),
		keyStyle.Render("q"), descStyle.Render("quit"),
	)

	connIcon := "●"
	connColor := sb.theme.Active
	if !connected {
		connIcon = "○"
		connColor = sb.theme.Error
	}
	connStyle := lipgloss.NewStyle().Foreground(connColor)

	right := fmt.Sprintf("%s  %s %s  %d active",
		descStyle.Render(source),
		connStyle.Render(connIcon),
		descStyle.Render("30fps"),
		activeCount,
	)

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return fmt.Sprintf("%s%*s%s", left, gap, "", right)
}
