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
		return style.Width(width - 2).Height(height - 2).Render("")
	}

	title := p.theme.Title.Render("TOOLS")
	var lines []string
	lines = append(lines, title)

	if len(toolCounts) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(p.theme.Idle).Render("No tool calls yet"))
		return style.Width(width - 2).Height(height - 2).Render(strings.Join(lines, "\n"))
	}

	type toolEntry struct {
		name  string
		count int
	}
	var entries []toolEntry
	for name, count := range toolCounts {
		entries = append(entries, toolEntry{name, count})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].count > entries[j].count })

	maxCount := entries[0].count

	// Compute label width from the widest entry so bars align.
	labelWidth := 0
	for _, e := range entries {
		l := len(fmt.Sprintf("%-6s %3d ", e.name, e.count))
		if l > labelWidth {
			labelWidth = l
		}
	}

	barWidth := innerW - labelWidth - 1
	if barWidth < 5 {
		barWidth = 5
	}

	maxRows := innerH - 1
	for i, e := range entries {
		if i >= maxRows {
			break
		}
		barLen := e.count * barWidth / maxCount
		if barLen < 1 {
			barLen = 1
		}
		bar := strings.Repeat("█", barLen)
		label := fmt.Sprintf("%-6s %3d ", e.name, e.count)
		lines = append(lines, label+lipgloss.NewStyle().Foreground(p.theme.ToolUse).Render(bar))
	}

	return style.Width(width - 2).Height(height - 2).Render(strings.Join(lines, "\n"))
}
