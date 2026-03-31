package panels

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/pattynextdoor/toph/internal/data"
	"github.com/pattynextdoor/toph/internal/ui"
)

// MetricsPanel renders aggregated token usage and cost info (Panel 4).
type MetricsPanel struct {
	theme   *ui.Theme
	focused bool

	// Animated accumulators — displayed values ease toward real values
	displayIn   float64
	displayOut  float64
	displayCost float64
}

// NewMetricsPanel creates a new metrics panel.
func NewMetricsPanel(theme *ui.Theme) *MetricsPanel {
	return &MetricsPanel{theme: theme}
}

func (p *MetricsPanel) SetFocused(f bool) { p.focused = f }
func (p *MetricsPanel) Focused() bool     { return p.focused }

// Model pricing per million tokens.
type modelPricing struct {
	Input      float64
	CacheRead  float64
	CacheWrite float64
	Output     float64
}

var pricing = map[string]modelPricing{
	"claude-opus-4-6":   {15.0, 1.50, 18.75, 75.0},
	"claude-sonnet-4-6": {3.0, 0.30, 3.75, 15.0},
	"claude-haiku-4-5":  {0.80, 0.08, 1.00, 4.0},
}

var defaultPricing = modelPricing{3.0, 0.30, 3.75, 15.0}

// Render draws the metrics panel using aggregated session data.
func (p *MetricsPanel) Render(sessions []*data.Session, burnRate float64, burnHistory [data.MetricsHistorySize]int, width, height int) string {
	style := p.theme.PanelNormal
	if p.focused {
		style = p.theme.PanelFocus
	}

	innerW := width - 4
	innerH := height - 2
	if innerH < 1 || innerW < 1 {
		return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render("")
	}

	title := ui.GradientText("METRICS", p.theme.BorderFocus, p.theme.Subagent)
	dimStyle := lipgloss.NewStyle().Foreground(p.theme.TextDim)
	valStyle := lipgloss.NewStyle().Foreground(p.theme.BorderFocus).Bold(true)
	greenStyle := lipgloss.NewStyle().Foreground(p.theme.Active)
	amberStyle := lipgloss.NewStyle().Foreground(p.theme.Waiting)

	var lines []string
	lines = append(lines, title)

	if len(sessions) == 0 {
		lines = append(lines, dimStyle.Render("No sessions"))
		return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render(strings.Join(lines, "\n"))
	}

	// Aggregate across all sessions
	var totalIn, totalOut, totalCacheRead, totalCacheWrite int
	var model string
	for _, s := range sessions {
		s.RLock()
		totalIn += s.TotalInputTokens
		totalOut += s.TotalOutputTokens
		totalCacheRead += s.TotalCacheRead
		totalCacheWrite += s.TotalCacheWrite
		if s.Model != "" {
			model = s.Model
		}
		s.RUnlock()
	}

	// Animate token accumulators toward real values
	p.displayIn = ui.AnimatedAccumulator(p.displayIn, float64(totalIn))
	p.displayOut = ui.AnimatedAccumulator(p.displayOut, float64(totalOut))

	// Tokens: styled values with dim labels (using animated display values)
	lines = append(lines, fmt.Sprintf("%s %s  %s %s",
		dimStyle.Render("in"),
		greenStyle.Render(formatTokens(int(p.displayIn))),
		dimStyle.Render("out"),
		amberStyle.Render(formatTokens(int(p.displayOut)))))

	// Burn rate
	if burnRate > 0 {
		lines = append(lines, fmt.Sprintf("%s %s",
			dimStyle.Render("rate"),
			lipgloss.NewStyle().Foreground(p.theme.Subagent).Bold(true).Render(fmt.Sprintf("%.0f tok/s", burnRate))))
	}

	// Cache ratio with visual indicator
	if totalIn > 0 && totalCacheRead > 0 {
		cacheRatio := float64(totalCacheRead) / float64(totalIn+totalCacheRead) * 100
		ratioColor := p.theme.Active
		if cacheRatio < 50 {
			ratioColor = p.theme.Waiting
		}
		lines = append(lines, fmt.Sprintf("%s %s %s",
			dimStyle.Render("cache"),
			lipgloss.NewStyle().Foreground(ratioColor).Bold(true).Render(fmt.Sprintf("%.0f%%", cacheRatio)),
			dimStyle.Render(formatTokens(totalCacheRead))))
	}

	// Cost with rate (animated)
	cost := estimateCost(model, totalIn, totalOut, totalCacheRead, totalCacheWrite)
	p.displayCost = ui.AnimatedAccumulator(p.displayCost, cost)
	prefix := ""
	if _, ok := pricing[model]; !ok {
		prefix = "~"
	}
	costStr := fmt.Sprintf("%s %s%s",
		dimStyle.Render("cost"),
		dimStyle.Render(prefix),
		valStyle.Render(fmt.Sprintf("$%.2f", p.displayCost)))
	if burnRate > 0 {
		pr, ok := pricing[model]
		if !ok {
			pr = defaultPricing
		}
		costPerHr := burnRate * pr.Output / 1_000_000 * 3600
		costStr += dimStyle.Render(fmt.Sprintf("  $%.0f/hr", costPerHr))
	}
	lines = append(lines, costStr)

	// Session count
	lines = append(lines, fmt.Sprintf("%s %s",
		dimStyle.Render("sessions"),
		valStyle.Render(fmt.Sprintf("%d", len(sessions)))))

	// Throughput chart — fill ALL remaining vertical space with a btop-style
	// multi-row area chart. This is the visual anchor of the panel.
	chartHeight := innerH - len(lines) - 1 // -1 for chart label
	if chartHeight >= 2 {
		histSlice := make([]int, data.MetricsHistorySize)
		for i, v := range burnHistory {
			histSlice[i] = v
		}
		chartWidth := innerW
		chart := renderChart(histSlice, chartWidth, chartHeight, p.theme.Active, p.theme.TextDim)
		if chart != "" {
			labelLeft := dimStyle.Render("throughput")
			labelRight := dimStyle.Render("5m")
			gap := innerW - lipgloss.Width(labelLeft) - lipgloss.Width(labelRight)
			if gap < 1 {
				gap = 1
			}
			lines = append(lines, fmt.Sprintf("%s%*s%s", labelLeft, gap, "", labelRight))
			lines = append(lines, chart)
		}
	}

	if len(lines) > innerH {
		lines = lines[:innerH]
	}

	return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render(strings.Join(lines, "\n"))
}

func estimateCost(model string, input, output, cacheRead, cacheWrite int) float64 {
	p, ok := pricing[model]
	if !ok {
		p = defaultPricing
	}

	nonCachedInput := input - cacheRead
	if nonCachedInput < 0 {
		nonCachedInput = 0
	}

	cost := float64(nonCachedInput) / 1_000_000 * p.Input
	cost += float64(cacheRead) / 1_000_000 * p.CacheRead
	cost += float64(cacheWrite) / 1_000_000 * p.CacheWrite
	cost += float64(output) / 1_000_000 * p.Output

	return cost
}
