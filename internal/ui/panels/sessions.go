package panels

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/pattynextdoor/toph/internal/data"
	"github.com/pattynextdoor/toph/internal/ui"
)

// SessionsPanel renders the session list (Panel 1) with status indicators,
// project names, git branches, and session ages. Sessions are displayed in
// the order provided — the caller is responsible for sorting by actionability
// (e.g. permission-waiting first).
type SessionsPanel struct {
	theme    *ui.Theme
	Selected int
	focused  bool
}

// NewSessionsPanel creates a SessionsPanel bound to the given theme.
func NewSessionsPanel(theme *ui.Theme) *SessionsPanel {
	return &SessionsPanel{theme: theme}
}

func (p *SessionsPanel) SetFocused(f bool) { p.focused = f }
func (p *SessionsPanel) Focused() bool     { return p.focused }
func (p *SessionsPanel) MoveUp()           { if p.Selected > 0 { p.Selected-- } }
func (p *SessionsPanel) MoveDown(max int)  { if p.Selected < max-1 { p.Selected++ } }

// SelectedSession returns the currently selected session from the provided
// list, or nil if the list is empty or selection is out of bounds.
func (p *SessionsPanel) SelectedSession(sessions []*data.Session) *data.Session {
	if len(sessions) == 0 || p.Selected >= len(sessions) {
		return nil
	}
	return sessions[p.Selected]
}

// Render draws the session list into the given width x height box.
func (p *SessionsPanel) Render(sessions []*data.Session, width, height int) string {
	style := p.theme.PanelNormal
	if p.focused {
		style = p.theme.PanelFocus
	}

	innerW := width - 4
	innerH := height - 2
	if innerW < 1 || innerH < 1 {
		return style.Width(width - 2).Height(height - 2).Render("")
	}

	title := p.theme.Title.Render("SESSIONS")
	var lines []string
	lines = append(lines, title)

	if len(sessions) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(p.theme.Idle).Render("No sessions found"))
	}

	for i, s := range sessions {
		if i >= innerH-1 {
			break
		}
		lines = append(lines, p.renderSessionRow(s, i == p.Selected, innerW))
	}

	content := strings.Join(lines, "\n")
	return style.Width(width - 2).Height(height - 2).Render(content)
}

// renderSessionRow formats a single session as: icon project branch age
func (p *SessionsPanel) renderSessionRow(s *data.Session, selected bool, width int) string {
	var icon string
	var iconColor color.Color
	switch s.Status {
	case data.StatusActive:
		icon = "●"
		iconColor = p.theme.Active
	case data.StatusWaiting:
		icon = "◐"
		iconColor = p.theme.Waiting
	case data.StatusIdle:
		icon = "○"
		iconColor = p.theme.Idle
	case data.StatusError:
		icon = "✕"
		iconColor = p.theme.Error
	default:
		icon = "○"
		iconColor = p.theme.Idle
	}

	iconStyle := lipgloss.NewStyle().Foreground(iconColor)

	project := humanizeProject(s.Project)
	if project == "" {
		maxLen := 8
		if len(s.ID) < maxLen {
			maxLen = len(s.ID)
		}
		project = s.ID[:maxLen]
	}

	branch := s.GitBranch
	if branch == "" {
		branch = "-"
	}

	age := formatDuration(time.Since(s.UpdatedAt))

	row := fmt.Sprintf("%s %s %s %s",
		iconStyle.Render(icon),
		project,
		lipgloss.NewStyle().Foreground(p.theme.TextDim).Render(branch),
		lipgloss.NewStyle().Foreground(p.theme.TextDim).Render(age),
	)

	if selected {
		row = lipgloss.NewStyle().
			Background(lipgloss.Color("#303030")).
			Width(width).
			Render(row)
	}
	return row
}

// humanizeProject converts a Claude Code project dir name (e.g.,
// "-Users-patty-dev-toph") into a short readable name (e.g., "toph").
// It takes the last path segment after splitting on hyphens that look
// like path separators.
func humanizeProject(name string) string {
	if name == "" {
		return ""
	}
	// Claude Code encodes paths as: -Users-patty-dev-toph
	// Split on "-" and take the last non-empty segment
	parts := strings.Split(strings.TrimLeft(name, "-"), "-")
	if len(parts) == 0 {
		return name
	}
	return parts[len(parts)-1]
}

// formatDuration returns a compact human-readable duration string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
