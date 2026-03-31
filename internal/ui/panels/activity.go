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

// ActivityPanel renders the real-time event stream (Panel 3). It maintains
// scroll state and auto-scrolls to the bottom unless the user scrolls up,
// at which point a "new" indicator appears.
type ActivityPanel struct {
	theme        *ui.Theme
	focused      bool
	autoScroll   bool
	offset       int
	sessionCount int // set by caller to control compact mode
}

// groupedEvent represents one or more consecutive events collapsed into a
// single display line. When Count > 1, the events share the same Type,
// ToolName, and SessionID, and occurred within groupWindow of each other.
type groupedEvent struct {
	Representative data.Event // most recent event in the group
	Count          int
	AllSameDetail  bool   // true when every event in the group has identical detail
	CommonDetail   string // the shared detail (when AllSameDetail) or empty
}

// groupWindow is the maximum gap between consecutive events before a new
// group starts, even if type+tool match.
const groupWindow = 60 * time.Second

// NewActivityPanel creates an ActivityPanel bound to the given theme.
func NewActivityPanel(theme *ui.Theme) *ActivityPanel {
	return &ActivityPanel{theme: theme, autoScroll: true}
}

func (p *ActivityPanel) SetFocused(f bool) { p.focused = f }
func (p *ActivityPanel) Focused() bool     { return p.focused }

// ScrollUp moves the viewport upward, disabling auto-scroll.
func (p *ActivityPanel) ScrollUp(amount int) {
	p.offset += amount
	p.autoScroll = false
}

// ScrollDown moves the viewport downward. Re-enables auto-scroll when
// the offset reaches zero (i.e. the user has scrolled back to the bottom).
func (p *ActivityPanel) ScrollDown(amount int) {
	p.offset -= amount
	if p.offset <= 0 {
		p.offset = 0
		p.autoScroll = true
	}
}

// ScrollToBottom jumps to the newest events and re-enables auto-scroll.
func (p *ActivityPanel) ScrollToBottom() {
	p.offset = 0
	p.autoScroll = true
}

// ScrollToTop jumps to the oldest events in the buffer.
func (p *ActivityPanel) ScrollToTop() {
	p.autoScroll = false
	p.offset = 999999
}

// SetSessionCount tells the panel how many active sessions exist so it can
// decide whether to show session IDs (multi-session) or hide them (single).
func (p *ActivityPanel) SetSessionCount(n int) { p.sessionCount = n }

// Render draws the event feed into the given width x height box.
func (p *ActivityPanel) Render(events []data.Event, width, height int) string {
	style := p.theme.PanelNormal
	if p.focused {
		style = p.theme.PanelFocus
	}

	innerW := width - 4
	innerH := height - 2
	if innerW < 1 || innerH < 1 {
		return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render("")
	}

	title := p.theme.Title.Render("ACTIVITY")
	visibleLines := innerH - 1

	// Group consecutive same-tool events before pagination so scroll
	// offsets operate on display lines, not raw events.
	groups := groupEvents(events)

	maxOffset := len(groups) - visibleLines
	if maxOffset < 0 {
		maxOffset = 0
	}
	if p.offset > maxOffset {
		p.offset = maxOffset
	}

	end := len(groups) - p.offset
	start := end - visibleLines
	if start < 0 {
		start = 0
	}

	var lines []string
	lines = append(lines, title)

	if len(groups) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(p.theme.Idle).Render("Waiting for events..."))
	}

	for _, g := range groups[start:end] {
		lines = append(lines, p.renderGroupedEvent(g, innerW))
	}

	if !p.autoScroll && p.offset > 0 {
		indicator := lipgloss.NewStyle().Foreground(p.theme.Waiting).Render(fmt.Sprintf("  ↓ %d new", p.offset))
		lines = append(lines, indicator)
	}

	content := strings.Join(lines, "\n")
	return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render(content)
}

// renderGroupedEvent renders a groupedEvent. Single events (Count == 1)
// render identically to the old per-event path. Groups show a count badge
// like "Bash ×3" with the shared detail (if all identical) or no detail.
func (p *ActivityPanel) renderGroupedEvent(g groupedEvent, width int) string {
	if g.Count == 1 {
		return p.renderEvent(g.Representative, width)
	}
	e := g.Representative

	ts := e.Timestamp.Format("15:04:05")
	tsStyle := lipgloss.NewStyle().Foreground(p.theme.TextDim)

	var typeLabel string
	var typeColor = p.theme.TextDim

	switch e.Type {
	case data.EventToolUse:
		typeLabel = shortenToolName(e.ToolName)
		typeColor = p.theme.ToolUse
	case data.EventToolResult:
		typeLabel = "result"
		typeColor = p.theme.ToolUse
	case data.EventSubagentStart:
		typeLabel = "agent+"
		typeColor = p.theme.Subagent
	case data.EventSubagentEnd:
		typeLabel = "agent-"
		typeColor = p.theme.Subagent
	case data.EventError:
		typeLabel = "ERROR"
		typeColor = p.theme.Error
	default:
		typeLabel = e.Type.String()
	}

	if e.Conflicted {
		typeColor = p.theme.Error
		typeLabel = "!" + typeLabel
	}

	// Add count badge: "Bash ×3"
	typeLabel = fmt.Sprintf("%s ×%d", typeLabel, g.Count)

	age := time.Since(e.Timestamp)
	typeColor = ageColor(typeColor, age, p.theme)
	labelStyle := lipgloss.NewStyle().Foreground(typeColor)

	var prefix string
	if p.sessionCount > 1 {
		sessID := e.SessionID
		if len(sessID) > 6 {
			sessID = sessID[:6]
		}
		prefix = fmt.Sprintf("%s %s %s",
			tsStyle.Render(ts), tsStyle.Render(sessID), labelStyle.Render(typeLabel))
	} else {
		prefix = fmt.Sprintf("%s %s",
			tsStyle.Render(ts), labelStyle.Render(typeLabel))
	}

	usedWidth := lipgloss.Width(prefix)
	remaining := width - usedWidth - 1

	// For groups with identical details, show the shared detail.
	// Otherwise omit detail — the count badge is informative enough.
	detail := ""
	if g.AllSameDetail {
		detail = shortenDetail(g.CommonDetail)
	}

	if remaining > 4 && detail != "" {
		if len(detail) > remaining {
			detail = detail[:remaining-3] + "..."
		}
		detailColor := ageColor(p.theme.TextDim, age, p.theme)
		return prefix + " " + lipgloss.NewStyle().Foreground(detailColor).Render(detail)
	}
	return prefix
}

// renderEvent formats a single event line. In single-session mode the format
// is compact: "HH:MM:SS ToolName detail". In multi-session mode the session
// ID prefix is included. Lines are hard-truncated to fit within width.
func (p *ActivityPanel) renderEvent(e data.Event, width int) string {
	ts := e.Timestamp.Format("15:04:05")
	tsStyle := lipgloss.NewStyle().Foreground(p.theme.TextDim)

	var typeLabel string
	var typeColor = p.theme.TextDim

	switch e.Type {
	case data.EventToolUse:
		typeLabel = shortenToolName(e.ToolName)
		typeColor = p.theme.ToolUse
	case data.EventToolResult:
		typeLabel = "result"
		typeColor = p.theme.ToolUse
	case data.EventSubagentStart:
		typeLabel = "agent+"
		typeColor = p.theme.Subagent
	case data.EventSubagentEnd:
		typeLabel = "agent-"
		typeColor = p.theme.Subagent
	case data.EventError:
		typeLabel = "ERROR"
		typeColor = p.theme.Error
	default:
		typeLabel = e.Type.String()
	}

	// Override color for file conflicts — red with warning prefix
	if e.Conflicted {
		typeColor = p.theme.Error
		typeLabel = "!" + typeLabel
	}

	// Age the color: fade to dim over 2 minutes
	age := time.Since(e.Timestamp)
	typeColor = ageColor(typeColor, age, p.theme)

	labelStyle := lipgloss.NewStyle().Foreground(typeColor)

	// Build prefix: include session ID only with multiple sessions
	var prefix string
	if p.sessionCount > 1 {
		sessID := e.SessionID
		if len(sessID) > 6 {
			sessID = sessID[:6]
		}
		prefix = fmt.Sprintf("%s %s %s",
			tsStyle.Render(ts), tsStyle.Render(sessID), labelStyle.Render(typeLabel))
	} else {
		prefix = fmt.Sprintf("%s %s",
			tsStyle.Render(ts), labelStyle.Render(typeLabel))
	}

	// Measure rendered prefix width (accounts for ANSI escape codes)
	usedWidth := lipgloss.Width(prefix)
	remaining := width - usedWidth - 1 // -1 for the space separator

	// Prefer tool input (file path, command) as the detail
	detail := e.ToolInput
	if detail == "" {
		detail = e.Text
	}

	// Shorten file paths: show just the last 2 path components
	detail = shortenDetail(detail)

	if remaining > 4 && detail != "" {
		if len(detail) > remaining {
			detail = detail[:remaining-3] + "..."
		}
		detailColor := ageColor(p.theme.TextDim, age, p.theme)
		return prefix + " " + lipgloss.NewStyle().Foreground(detailColor).Render(detail)
	}
	return prefix
}

// ageColor fades a color toward TextDim based on event age.
// Events < 5s old keep their original color.
// Events > 2min old are fully dimmed.
// In between, we step through discrete brightness levels.
func ageColor(original color.Color, age time.Duration, theme *ui.Theme) color.Color {
	if age < 5*time.Second {
		return original
	}
	if age > 2*time.Minute {
		return theme.TextDim
	}
	// Between 5s and 2min: step down in 3 bands.
	if age < 30*time.Second {
		return original // still pretty fresh
	}
	if age < 1*time.Minute {
		// Slightly faded
		return lipgloss.Color("#808080")
	}
	// 1-2 minutes: quite faded
	return lipgloss.Color("#606060")
}

// groupEvents collapses consecutive events that share the same Type, ToolName,
// and SessionID (within groupWindow) into grouped display entries.
func groupEvents(events []data.Event) []groupedEvent {
	if len(events) == 0 {
		return nil
	}
	groups := make([]groupedEvent, 0, len(events))
	cur := groupedEvent{
		Representative: events[0],
		Count:          1,
		AllSameDetail:  true,
		CommonDetail:   eventDetail(events[0]),
	}

	for i := 1; i < len(events); i++ {
		e := events[i]
		gap := e.Timestamp.Sub(cur.Representative.Timestamp)
		if gap < 0 {
			gap = -gap
		}
		sameGroup := e.Type == cur.Representative.Type &&
			e.ToolName == cur.Representative.ToolName &&
			e.SessionID == cur.Representative.SessionID &&
			gap <= groupWindow

		if sameGroup {
			cur.Count++
			cur.Representative = e // keep the most recent
			if cur.AllSameDetail && eventDetail(e) != cur.CommonDetail {
				cur.AllSameDetail = false
			}
		} else {
			groups = append(groups, cur)
			cur = groupedEvent{
				Representative: e,
				Count:          1,
				AllSameDetail:  true,
				CommonDetail:   eventDetail(e),
			}
		}
	}
	groups = append(groups, cur)
	return groups
}

// eventDetail extracts the display detail for an event (same logic as renderEvent).
func eventDetail(e data.Event) string {
	d := e.ToolInput
	if d == "" {
		d = e.Text
	}
	return d
}

// shortenDetail trims verbose detail strings. For file paths it shows just
// the last 2 components (e.g., "landing/page.tsx" instead of
// "/Users/patty/dev/toph/landing/page.tsx").
func shortenDetail(s string) string {
	if len(s) == 0 {
		return s
	}
	// Detect file paths
	if s[0] == '/' || strings.HasPrefix(s, "~/") {
		parts := strings.Split(s, "/")
		// Take last 2 non-empty components
		var tail []string
		for i := len(parts) - 1; i >= 0 && len(tail) < 2; i-- {
			if parts[i] != "" {
				tail = append([]string{parts[i]}, tail...)
			}
		}
		if len(tail) > 0 {
			return strings.Join(tail, "/")
		}
	}
	return s
}

// shortenToolName abbreviates long MCP tool names like
// "mcp__playwright__browser_take_screenshot" to "take_screenshot".
// Short built-in names (Read, Bash, Edit, etc.) pass through unchanged.
func shortenToolName(name string) string {
	if len(name) <= 12 {
		return name
	}
	// MCP tools use double-underscore separators; take the last segment
	if idx := strings.LastIndex(name, "__"); idx >= 0 {
		return name[idx+2:]
	}
	// Fallback: just truncate
	return name[:12]
}
