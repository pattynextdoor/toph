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

// gapThreshold is the minimum time between consecutive display items
// before a "── Xm ──" separator line is inserted.
const gapThreshold = 2 * time.Minute

// displayItem is either a grouped event or a time-gap separator.
type displayItem struct {
	group    *groupedEvent // non-nil for event lines
	gapLabel string        // non-empty for separator lines
}

// toolGlyph returns a Unicode icon for the given tool name.
func toolGlyph(toolName string) string {
	switch toolName {
	case "Bash":
		return "▶"
	case "Read":
		return "◇"
	case "Edit", "Write", "NotebookEdit":
		return "◆"
	case "Glob", "Grep":
		return "⊙"
	case "Agent", "Skill":
		return "✦"
	default:
		return "○"
	}
}

// toolSpecificColor returns a per-tool color so different tool types are
// visually distinct at a glance.
func toolSpecificColor(toolName string, theme *ui.Theme) color.Color {
	switch toolName {
	case "Bash":
		return theme.Waiting // amber #FFD787
	case "Read":
		return theme.UserMsg // blue #87AFFF
	case "Edit", "Write", "NotebookEdit":
		return theme.FileWrite // green #87D787
	case "Glob", "Grep":
		return lipgloss.Color("#D7AFD7") // magenta
	case "Agent", "Skill":
		return theme.Subagent // lavender #D7AFFF
	default:
		return theme.ToolUse // cyan fallback
	}
}

// formatGapDuration formats a duration into a compact human-readable string
// like "2m", "1h 23m", or "5h 21m".
func formatGapDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 && m > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if h > 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dm", m)
}

// buildDisplayItems interleaves grouped events with time-gap separators.
func buildDisplayItems(groups []groupedEvent) []displayItem {
	if len(groups) == 0 {
		return nil
	}
	items := make([]displayItem, 0, len(groups)+len(groups)/4)
	items = append(items, displayItem{group: &groups[0]})

	for i := 1; i < len(groups); i++ {
		prev := groups[i-1].Representative.Timestamp
		cur := groups[i].Representative.Timestamp
		gap := cur.Sub(prev)
		if gap < 0 {
			gap = -gap
		}
		if gap >= gapThreshold {
			items = append(items, displayItem{gapLabel: formatGapDuration(gap)})
		}
		items = append(items, displayItem{group: &groups[i]})
	}
	return items
}

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

	title := ui.GradientText("ACTIVITY", p.theme.BorderFocus, p.theme.Subagent)
	visibleLines := innerH - 1

	// Group consecutive same-tool events, then interleave time-gap
	// separators so the feed shows rhythm between bursts.
	groups := groupEvents(events)
	items := buildDisplayItems(groups)

	maxOffset := len(items) - visibleLines
	if maxOffset < 0 {
		maxOffset = 0
	}
	if p.offset > maxOffset {
		p.offset = maxOffset
	}

	end := len(items) - p.offset
	start := end - visibleLines
	if start < 0 {
		start = 0
	}

	var lines []string
	lines = append(lines, title)

	if len(items) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(p.theme.Idle).Render("Waiting for events..."))
	}

	for _, item := range items[start:end] {
		if item.group != nil {
			lines = append(lines, p.renderGroupedEvent(*item.group, innerW))
		} else {
			lines = append(lines, p.renderGapSeparator(item.gapLabel, innerW))
		}
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
// like "▶ Bash ×3" with the shared detail (if all identical) or no detail.
func (p *ActivityPanel) renderGroupedEvent(g groupedEvent, width int) string {
	if g.Count == 1 {
		return p.renderEvent(g.Representative, width)
	}
	e := g.Representative

	ts := e.Timestamp.Format("15:04:05")
	tsStyle := lipgloss.NewStyle().Foreground(p.theme.TextDim)

	typeLabel, typeColor := p.eventLabelAndColor(e)

	if e.Conflicted {
		typeColor = p.theme.Error
		typeLabel = "!" + typeLabel
	}

	// Add count badge: "▶ Bash ×3"
	typeLabel = fmt.Sprintf("%s ×%d", typeLabel, g.Count)

	age := time.Since(e.Timestamp)
	// Tool label keeps its per-tool color always (categorical, not temporal).
	// Bold signals freshness instead.
	labelStyle := lipgloss.NewStyle().Foreground(typeColor)
	if age < 5*time.Second {
		labelStyle = labelStyle.Bold(true)
	}

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

// eventLabelAndColor returns the glyph-prefixed label and per-tool color
// for an event. Centralises the mapping so renderEvent and
// renderGroupedEvent stay consistent.
func (p *ActivityPanel) eventLabelAndColor(e data.Event) (string, color.Color) {
	switch e.Type {
	case data.EventToolUse:
		name := shortenToolName(e.ToolName)
		glyph := toolGlyph(e.ToolName)
		return glyph + " " + name, toolSpecificColor(e.ToolName, p.theme)
	case data.EventToolResult:
		return "· result", p.theme.TextDim
	case data.EventSubagentStart:
		return "✦ agent+", p.theme.Subagent
	case data.EventSubagentEnd:
		return "✦ agent-", p.theme.Subagent
	case data.EventError:
		return "✗ ERROR", p.theme.Error
	default:
		return "○ " + e.Type.String(), p.theme.TextDim
	}
}

// renderGapSeparator draws a centered time-gap line like "── 5h 21m ──".
func (p *ActivityPanel) renderGapSeparator(label string, width int) string {
	middle := " " + label + " "
	sideLen := (width - len(middle)) / 2
	if sideLen < 2 {
		sideLen = 2
	}
	side := strings.Repeat("─", sideLen)
	line := side + middle + side
	// Trim to exact width
	if len(line) > width {
		line = line[:width]
	}
	return lipgloss.NewStyle().Foreground(p.theme.TextDim).Render(line)
}

// renderEvent formats a single event line. In single-session mode the format
// is compact: "HH:MM:SS ▶ Bash detail". In multi-session mode the session
// ID prefix is included. Lines are hard-truncated to fit within width.
func (p *ActivityPanel) renderEvent(e data.Event, width int) string {
	ts := e.Timestamp.Format("15:04:05")
	tsStyle := lipgloss.NewStyle().Foreground(p.theme.TextDim)

	typeLabel, typeColor := p.eventLabelAndColor(e)

	// Override color for file conflicts — red with warning prefix
	if e.Conflicted {
		typeColor = p.theme.Error
		typeLabel = "!" + typeLabel
	}

	age := time.Since(e.Timestamp)
	// Tool label keeps its per-tool color always (categorical, not temporal).
	// Bold signals freshness instead.
	labelStyle := lipgloss.NewStyle().Foreground(typeColor)
	if age < 5*time.Second {
		labelStyle = labelStyle.Bold(true)
	}

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
