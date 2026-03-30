package ui

// Layout holds the computed dimensions for every panel in the toph dashboard.
// Call ComputeLayout on each terminal resize to recalculate sizes. The split
// ratios (35/65 left/right, 40/40/20 left stack, 70/30 right stack) are tuned
// to match the panel layout spec in CLAUDE.md.
type Layout struct {
	Width  int
	Height int

	LeftWidth  int
	RightWidth int

	SessionsHeight int
	DetailHeight   int
	ToolsHeight    int

	ActivityHeight int
	MetricsHeight  int

	StatusBarHeight int

	// TooSmall is set when the terminal is below the minimum usable size
	// (60x20). The root model should render a "Terminal too small" message
	// instead of the full dashboard.
	TooSmall bool
}

const (
	minWidth  = 60
	minHeight = 20
)

// ComputeLayout divides the terminal into left/right columns and stacks panels
// within each column. The left column gets 35% width (sessions, detail, tools)
// and the right column gets the remainder (activity feed, metrics). Heights are
// split proportionally within each column.
func ComputeLayout(width, height int) Layout {
	l := Layout{
		Width:           width,
		Height:          height,
		StatusBarHeight: 1,
	}

	if width < minWidth || height < minHeight {
		l.TooSmall = true
		return l
	}

	usableHeight := height - l.StatusBarHeight

	// Left column: 35% width; right column: remainder.
	l.LeftWidth = width * 35 / 100
	l.RightWidth = width - l.LeftWidth

	// Left stack: sessions 40%, tools 20%, detail fills the rest.
	l.SessionsHeight = usableHeight * 40 / 100
	l.ToolsHeight = usableHeight * 20 / 100
	l.DetailHeight = usableHeight - l.SessionsHeight - l.ToolsHeight

	// Right stack: activity feed 70%, metrics fills the rest.
	l.ActivityHeight = usableHeight * 70 / 100
	l.MetricsHeight = usableHeight - l.ActivityHeight

	return l
}

// InnerWidth returns the content width inside a bordered panel, accounting for
// the left/right border characters (1 each) and left/right padding (1 each) = 4.
func (l Layout) InnerWidth(panelWidth int) int {
	w := panelWidth - 4 // 2 border chars + 2 padding chars
	if w < 0 {
		return 0
	}
	return w
}

// InnerHeight returns the content height inside a bordered panel, accounting
// for the top and bottom border lines = 2.
func (l Layout) InnerHeight(panelHeight int) int {
	h := panelHeight - 2 // top + bottom border
	if h < 0 {
		return 0
	}
	return h
}
