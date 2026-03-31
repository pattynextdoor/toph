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

	title := p.theme.Title.Render("METRICS")
	dimStyle := lipgloss.NewStyle().Foreground(p.theme.TextDim)

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

	// Row 1: tokens in/out + burn rate
	burnStr := ""
	if burnRate > 0 {
		burnStr = fmt.Sprintf("  %s", lipgloss.NewStyle().Foreground(p.theme.Subagent).Render(fmt.Sprintf("%.0f tok/s", burnRate)))
	}
	lines = append(lines, fmt.Sprintf("%s in  %s out%s",
		lipgloss.NewStyle().Foreground(p.theme.Active).Render(formatTokens(totalIn)),
		lipgloss.NewStyle().Foreground(p.theme.Waiting).Render(formatTokens(totalOut)),
		burnStr))

	// Row 2: cache ratio
	if totalIn > 0 && totalCacheRead > 0 {
		cacheRatio := float64(totalCacheRead) / float64(totalIn+totalCacheRead) * 100
		lines = append(lines, dimStyle.Render(fmt.Sprintf("cache %s", formatTokens(totalCacheRead)))+
			lipgloss.NewStyle().Foreground(p.theme.Active).Render(fmt.Sprintf(" %.0f%% hit", cacheRatio)))
	}

	// Row 3: cost + cost rate + sessions
	cost := estimateCost(model, totalIn, totalOut, totalCacheRead, totalCacheWrite)
	prefix := ""
	if _, ok := pricing[model]; !ok {
		prefix = "~"
	}
	costLine := fmt.Sprintf("cost %s%s",
		dimStyle.Render(prefix),
		lipgloss.NewStyle().Foreground(p.theme.Active).Render(fmt.Sprintf("$%.2f", cost)))
	// Cost rate: extrapolate from burn rate to $/hr using output pricing
	if burnRate > 0 {
		pr, ok := pricing[model]
		if !ok {
			pr = defaultPricing
		}
		costPerSec := burnRate * pr.Output / 1_000_000
		costPerHr := costPerSec * 3600
		costLine += dimStyle.Render(fmt.Sprintf("  $%.0f/hr", costPerHr))
	}
	lines = append(lines, costLine)

	// Burn rate chart — use remaining vertical space
	chartHeight := innerH - len(lines)
	if chartHeight >= 1 {
		// Convert history array to slice
		histSlice := make([]int, data.MetricsHistorySize)
		for i, v := range burnHistory {
			histSlice[i] = v
		}
		chartWidth := innerW
		chart := renderBrailleChart(histSlice, chartWidth, p.theme.Active)
		if chart != "" {
			lines = append(lines, dimStyle.Render("tok/s")+" "+dimStyle.Render("5m"))
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
