package panels

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
	"github.com/pattynextdoor/toph/internal/ui"
)

// ToolsPanel renders the tool usage bar chart (Panel 5), showing horizontal
// bars for each tool sorted by call frequency.
type ToolsPanel struct {
	theme    *ui.Theme
	focused  bool
	progress progress.Model
}

// NewToolsPanel creates a ToolsPanel bound to the given theme.
func NewToolsPanel(theme *ui.Theme) *ToolsPanel {
	p := progress.New(
		progress.WithColors(theme.ToolUse),
		progress.WithoutPercentage(),
		progress.WithFillCharacters('━', '╌'),
	)
	return &ToolsPanel{theme: theme, progress: p}
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
	var total int
	for name, count := range toolCounts {
		entries = append(entries, toolEntry{name, count})
		total += count
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].name < entries[j].name
	})

	maxCount := entries[0].count

	// Find the longest tool name (capped at 8 chars)
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

	// Label takes: name + space + count(3) + space + percent(4) + space
	labelWidth := maxNameLen + 1 + 3 + 1 + 4 + 1
	barWidth := innerW - labelWidth
	if barWidth < 4 {
		barWidth = 4
	}
	// Cap at half width to avoid overflow from wide chars
	if barWidth > innerW/2 {
		barWidth = innerW / 2
	}

	maxRows := innerH - 1
	nameFmt := fmt.Sprintf("%%-%ds", maxNameLen)
	dimStyle := lipgloss.NewStyle().Foreground(p.theme.TextDim)
	countStyle := lipgloss.NewStyle().Foreground(p.theme.BorderFocus).Bold(true)

	p.progress.SetWidth(barWidth)

	for i, e := range entries {
		if i >= maxRows {
			break
		}
		name := e.name
		if len(name) > maxNameLen {
			name = name[:maxNameLen]
		}

		pct := float64(e.count) / float64(maxCount)
		pctOfTotal := float64(e.count) / float64(total) * 100

		bar := p.progress.ViewAs(pct)
		label := fmt.Sprintf("%s %s %s %s",
			dimStyle.Render(fmt.Sprintf(nameFmt, name)),
			countStyle.Render(fmt.Sprintf("%3d", e.count)),
			dimStyle.Render(fmt.Sprintf("%2.0f%%", pctOfTotal)),
			bar,
		)
		lines = append(lines, label)
	}

	return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render(strings.Join(lines, "\n"))
}
