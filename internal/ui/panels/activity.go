package panels

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/pattynextdoor/toph/internal/data"
	"github.com/pattynextdoor/toph/internal/ui"
)

// ActivityPanel renders the real-time event stream (Panel 3). It maintains
// scroll state and auto-scrolls to the bottom unless the user scrolls up,
// at which point a "new" indicator appears.
type ActivityPanel struct {
	theme      *ui.Theme
	focused    bool
	autoScroll bool
	offset     int
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

// Render draws the event feed into the given width x height box.
func (p *ActivityPanel) Render(events []data.Event, width, height int) string {
	style := p.theme.PanelNormal
	if p.focused {
		style = p.theme.PanelFocus
	}

	innerW := width - 4
	innerH := height - 2
	if innerW < 1 || innerH < 1 {
		return style.Width(width - 2).Height(height - 2).Render("")
	}

	title := p.theme.Title.Render("ACTIVITY")
	visibleLines := innerH - 1

	maxOffset := len(events) - visibleLines
	if maxOffset < 0 {
		maxOffset = 0
	}
	if p.offset > maxOffset {
		p.offset = maxOffset
	}

	end := len(events) - p.offset
	start := end - visibleLines
	if start < 0 {
		start = 0
	}

	var lines []string
	lines = append(lines, title)

	if len(events) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(p.theme.Idle).Render("Waiting for events..."))
	}

	for _, e := range events[start:end] {
		lines = append(lines, p.renderEvent(e, innerW))
	}

	if !p.autoScroll && p.offset > 0 {
		indicator := lipgloss.NewStyle().Foreground(p.theme.Waiting).Render(fmt.Sprintf("  ↓ %d new", p.offset))
		lines = append(lines, indicator)
	}

	content := strings.Join(lines, "\n")
	return style.Width(width - 2).Height(height - 2).Render(content)
}

// renderEvent formats a single event line: [time] [session] [type] detail
func (p *ActivityPanel) renderEvent(e data.Event, width int) string {
	ts := e.Timestamp.Format("15:04:05")
	tsStyle := lipgloss.NewStyle().Foreground(p.theme.TextDim)

	var typeLabel string
	var typeColor = p.theme.TextDim

	switch e.Type {
	case data.EventToolUse:
		typeLabel = e.ToolName
		typeColor = p.theme.ToolUse
	case data.EventToolResult:
		typeLabel = "result"
		typeColor = p.theme.ToolUse
	case data.EventAssistantText:
		typeLabel = "think"
		typeColor = p.theme.TextDim
	case data.EventUserMessage:
		typeLabel = "user"
		typeColor = p.theme.UserMsg
	case data.EventSystemMessage:
		typeLabel = "system"
		typeColor = p.theme.System
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

	labelStyle := lipgloss.NewStyle().Foreground(typeColor)

	sessID := e.SessionID
	if len(sessID) > 6 {
		sessID = sessID[:6]
	}

	prefix := fmt.Sprintf("%s %s %-8s",
		tsStyle.Render(ts), tsStyle.Render(sessID), labelStyle.Render(typeLabel))

	detail := e.Text
	if e.ToolInput != "" && detail == "" {
		detail = e.ToolInput
	}
	maxDetail := width - 28
	if maxDetail > 0 && len(detail) > maxDetail {
		detail = detail[:maxDetail-3] + "..."
	}

	if detail != "" {
		return prefix + " " + lipgloss.NewStyle().Foreground(p.theme.TextDim).Render(detail)
	}
	return prefix
}
