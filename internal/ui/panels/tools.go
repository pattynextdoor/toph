package panels

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/pattynextdoor/toph/internal/ui"
)

// ToolsPanel renders the tool usage bar chart (Panel 5), showing horizontal
// bars for each tool sorted by call frequency. Bars scale proportionally
// to the most-used tool.
type ToolsPanel struct {
	theme   *ui.Theme
	focused bool
}

// NewToolsPanel creates a ToolsPanel bound to the given theme.
func NewToolsPanel(theme *ui.Theme) *ToolsPanel {
	return &ToolsPanel{theme: theme}
}

func (p *ToolsPanel) SetFocused(f bool) { p.focused = f }
func (p *ToolsPanel) Focused() bool     { return p.focused }

// Render draws the tool frequency bar chart into the given width x height box.
func (p *ToolsPanel) Render(toolCounts map[string]int, width, height int) string {
	style := p.theme.PanelNormal
	if p.focused {
		style = p.theme.PanelFocus
	}

	innerW := width - 4
	innerH := height - 2
	if innerW < 1 || innerH < 1 {
		return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render("")
	}

	title := p.theme.Title.Render("TOOLS")
	var lines []string
	lines = append(lines, title)

	if len(toolCounts) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(p.theme.Idle).Render("No tool calls yet"))
		return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render(strings.Join(lines, "\n"))
	}

	type toolEntry struct {
		name  string
		count int
	}
	var entries []toolEntry
	for name, count := range toolCounts {
		entries = append(entries, toolEntry{name, count})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].name < entries[j].name // stable tiebreaker
	})

	maxCount := entries[0].count

	// Find the longest tool name (capped at 8 chars to save space).
	maxNameLen := 0
	for _, e := range entries {
		n := len(e.name)
		if n > maxNameLen {
			maxNameLen = n
		}
	}
	if maxNameLen > 8 {
		maxNameLen = 8
	}

	// label = "Name  N " — dynamic name width + count
	labelWidth := maxNameLen + 1 + 3 + 1 // name + space + count(3) + space

	barWidth := innerW - labelWidth
	if barWidth < 3 {
		barWidth = 3
	}

	// Reserve space for label, then give remaining to bar.
	// Use half the inner width as max bar length to guarantee no overflow.
	maxBarLen := (innerW - labelWidth) / 2
	if maxBarLen < 2 {
		maxBarLen = 2
	}

	maxRows := innerH - 1
	nameFmt := fmt.Sprintf("%%-%ds", maxNameLen)
	barStyle := lipgloss.NewStyle().Foreground(p.theme.ToolUse)
	for i, e := range entries {
		if i >= maxRows {
			break
		}
		name := e.name
		if len(name) > maxNameLen {
			name = name[:maxNameLen]
		}
		barLen := e.count * maxBarLen / maxCount
		if barLen < 1 {
			barLen = 1
		}
		label := fmt.Sprintf(nameFmt+" %3d %s", name, e.count,
			barStyle.Render(strings.Repeat("=", barLen)))
		lines = append(lines, label)
	}

	return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render(strings.Join(lines, "\n"))
}
