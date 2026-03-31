package panels

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/pattynextdoor/toph/internal/data"
	"github.com/pattynextdoor/toph/internal/ui"
)

// DetailPanel renders info about the selected session (Panel 2).
type DetailPanel struct {
	theme   *ui.Theme
	focused bool
}

// NewDetailPanel creates a new detail panel.
func NewDetailPanel(theme *ui.Theme) *DetailPanel {
	return &DetailPanel{theme: theme}
}

func (p *DetailPanel) SetFocused(f bool) { p.focused = f }
func (p *DetailPanel) Focused() bool     { return p.focused }

// Render draws the detail panel for the given session.
// If session is nil, shows a placeholder.
func (p *DetailPanel) Render(session *data.Session, width, height int) string {
	style := p.theme.PanelNormal
	if p.focused {
		style = p.theme.PanelFocus
	}

	innerW := width - 4
	innerH := height - 2
	if innerW < 1 || innerH < 1 {
		return style.Width(width - 2).Height(height - 2).Render("")
	}

	title := p.theme.Title.Render("DETAIL")
	var lines []string
	lines = append(lines, title)

	if session == nil {
		lines = append(lines, lipgloss.NewStyle().Foreground(p.theme.Idle).Render("Select a session"))
		return style.Width(width - 2).Height(height - 2).Render(strings.Join(lines, "\n"))
	}

	session.RLock()
	defer session.RUnlock()

	dimStyle := lipgloss.NewStyle().Foreground(p.theme.TextDim)
	valStyle := lipgloss.NewStyle().Foreground(p.theme.BorderFocus)

	// Status line
	var statusIcon string
	var statusColor = p.theme.Active
	switch session.Status {
	case data.StatusActive:
		statusIcon = "● active"
		statusColor = p.theme.Active
	case data.StatusWaiting:
		statusIcon = "◐ waiting"
		statusColor = p.theme.Waiting
	case data.StatusIdle:
		statusIcon = "○ idle"
		statusColor = p.theme.Idle
	case data.StatusError:
		statusIcon = "✕ error"
		statusColor = p.theme.Error
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon))
	lines = append(lines, "")

	// Working directory
	cwd := session.CWD
	if cwd != "" {
		// Show just the last 2 path components for brevity
		cwd = shortenPath(cwd, innerW-6)
		lines = append(lines, dimStyle.Render("dir  ")+valStyle.Render(cwd))
	}

	// Git branch
	if session.GitBranch != "" {
		lines = append(lines, dimStyle.Render("git  ")+valStyle.Render(session.GitBranch))
	}

	// Model
	if session.Model != "" {
		lines = append(lines, dimStyle.Render("model")+lipgloss.NewStyle().Foreground(p.theme.Subagent).Render(" "+session.Model))
	}

	// Duration
	duration := time.Since(session.StartedAt)
	lines = append(lines, dimStyle.Render("age  ")+dimStyle.Render(" "+formatDuration(duration)))

	// Last tool
	if session.LastToolName != "" {
		lines = append(lines, dimStyle.Render("last ")+lipgloss.NewStyle().Foreground(p.theme.ToolUse).Render(" "+session.LastToolName))
	}

	lines = append(lines, "")

	// Token summary
	totalIn := session.TotalInputTokens
	totalOut := session.TotalOutputTokens
	if totalIn > 0 || totalOut > 0 {
		lines = append(lines, dimStyle.Render(fmt.Sprintf("tokens %s in / %s out",
			formatTokens(totalIn), formatTokens(totalOut))))
	}

	// Truncate to fit
	if len(lines) > innerH {
		lines = lines[:innerH]
	}

	return style.Width(width - 2).Height(height - 2).Render(strings.Join(lines, "\n"))
}

// shortenPath trims a path to fit within maxLen by showing the last components.
func shortenPath(p string, maxLen int) string {
	if len(p) <= maxLen {
		return p
	}
	// Show last 2 components
	base := filepath.Base(p)
	parent := filepath.Base(filepath.Dir(p))
	short := filepath.Join(parent, base)
	if len(short) <= maxLen {
		return ".../" + short
	}
	return base
}

// formatTokens formats a token count in a compact way (e.g., "142K", "1.2M").
func formatTokens(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}
