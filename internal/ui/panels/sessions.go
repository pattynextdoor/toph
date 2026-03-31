package panels

import (
	"fmt"
	"image/color"
	"math"
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
	frame    int // incremented each render for animation
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
// spinnerFrame is the current frame of the Bubbles spinner for active sessions.
func (p *SessionsPanel) Render(sessions []*data.Session, width, height int, spinnerFrame string) string {
	p.frame++

	style := p.theme.PanelNormal
	if p.focused {
		style = p.theme.PanelFocus
	}

	innerW := width - 4
	innerH := height - 2
	if innerW < 1 || innerH < 1 {
		return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render("")
	}

	title := ui.GradientText("SESSIONS", p.theme.BorderFocus, p.theme.Subagent)
	var lines []string
	lines = append(lines, title)

	if len(sessions) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(p.theme.Idle).Render("No sessions found"))
	}

	for i, s := range sessions {
		if i >= innerH-1 {
			break
		}
		lines = append(lines, p.renderSessionRow(s, i == p.Selected, innerW, spinnerFrame))
	}

	content := strings.Join(lines, "\n")
	return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render(content)
}

// sparkChars maps token rate values to braille characters, from lowest (dot)
// to highest (full block). Eight levels give smooth visual resolution.
var sparkChars = []rune{'\u2840', '\u2844', '\u2846', '\u2847', '\u28C7', '\u28E7', '\u28F7', '\u28FF'}

// renderSparkline renders an 8-character braille sparkline from a token
// history array. Active sessions get green sparklines; all-zero histories
// render as dim dots so the column width stays stable.
func renderSparkline(history [data.SparklineSamples]int, theme *ui.Theme) string {
	// Find max for scaling.
	max := 1
	allZero := true
	for _, v := range history {
		if v > max {
			max = v
		}
		if v > 0 {
			allZero = false
		}
	}

	if allZero {
		return lipgloss.NewStyle().Foreground(theme.Idle).Render(string([]rune{
			sparkChars[0], sparkChars[0], sparkChars[0], sparkChars[0],
			sparkChars[0], sparkChars[0], sparkChars[0], sparkChars[0],
		}))
	}

	var result strings.Builder
	for _, v := range history {
		idx := v * (len(sparkChars) - 1) / max
		result.WriteRune(sparkChars[idx])
	}
	return lipgloss.NewStyle().Foreground(theme.Active).Render(result.String())
}

// renderSessionRow formats a single session as: icon project branch sparkline age
func (p *SessionsPanel) renderSessionRow(s *data.Session, selected bool, width int, spinnerFrame string) string {
	var icon string
	var iconColor color.Color
	switch s.Status {
	case data.StatusActive:
		icon = spinnerFrame // animated spinner from Bubbles
		iconColor = p.theme.Active
	case data.StatusWaiting:
		icon = "◐"
		// Pulse between bright amber and dim at ~1Hz
		pulse := (math.Sin(float64(p.frame)*0.15) + 1) / 2
		if pulse > 0.5 {
			iconColor = p.theme.Waiting
		} else {
			iconColor = p.theme.Idle
		}
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

	// Only render sparkline if there's enough horizontal space (icon + project
	// + branch + sparkline(8) + age + spaces need ~30+ cols).
	var row string
	if width > 30 {
		sparkline := renderSparkline(s.GetTokenHistory(), p.theme)
		row = fmt.Sprintf("%s %s %s %s %s",
			iconStyle.Render(icon),
			project,
			lipgloss.NewStyle().Foreground(p.theme.TextDim).Render(branch),
			sparkline,
			lipgloss.NewStyle().Foreground(p.theme.TextDim).Render(age),
		)
	} else {
		row = fmt.Sprintf("%s %s %s %s",
			iconStyle.Render(icon),
			project,
			lipgloss.NewStyle().Foreground(p.theme.TextDim).Render(branch),
			lipgloss.NewStyle().Foreground(p.theme.TextDim).Render(age),
		)
	}

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
