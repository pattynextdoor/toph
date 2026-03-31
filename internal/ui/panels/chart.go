package panels

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// barChars are block elements from empty to full (8 levels per cell row).
var barChars = []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

const levelsPerRow = 8 // number of distinct fill levels per character cell

// renderChart renders a multi-row bar chart that fills the given height.
// Each column represents one data point. Columns fill from the bottom up
// using Unicode block elements, giving (height * 8) levels of vertical
// resolution. This is the btop-style area chart.
func renderChart(values []int, width, height int, col, dimCol color.Color) string {
	if width <= 0 || height <= 0 || len(values) == 0 {
		return ""
	}

	// Resample to fit width
	sampled := resample(values, width)

	// Find max — if no data, show an empty chart frame
	max := 0
	nonZero := 0
	for _, v := range sampled {
		if v > max {
			max = v
		}
		if v > 0 {
			nonZero++
		}
	}

	if max == 0 || nonZero < 2 {
		return "" // not enough data
	}

	totalLevels := height * levelsPerRow

	// For each column, compute how many levels it fills (0 to totalLevels)
	colLevels := make([]int, len(sampled))
	for i, v := range sampled {
		colLevels[i] = v * totalLevels / max
		if colLevels[i] > totalLevels {
			colLevels[i] = totalLevels
		}
	}

	barStyle := lipgloss.NewStyle().Foreground(col)
	emptyStyle := lipgloss.NewStyle().Foreground(dimCol)

	// Render row by row, top to bottom
	var rows []string
	for row := 0; row < height; row++ {
		// This row represents levels from rowBottom to rowTop
		rowBottom := (height - 1 - row) * levelsPerRow
		var b strings.Builder

		for _, level := range colLevels {
			fillInRow := level - rowBottom
			if fillInRow <= 0 {
				// Column doesn't reach this row at all
				b.WriteRune(' ')
			} else if fillInRow >= levelsPerRow {
				// Column completely fills this row
				b.WriteRune('█')
			} else {
				// Partial fill — use the appropriate block character
				b.WriteRune(barChars[fillInRow])
			}
		}

		line := b.String()
		// Style: filled chars get the main color, but we render the whole
		// line at once. Use the bar color for non-space chars.
		// Simple approach: render the full line in bar color — spaces are invisible anyway.
		hasContent := false
		for _, r := range line {
			if r != ' ' {
				hasContent = true
				break
			}
		}
		if hasContent {
			rows = append(rows, barStyle.Render(line))
		} else {
			rows = append(rows, emptyStyle.Render(line))
		}
	}

	return strings.Join(rows, "\n")
}

// renderBrailleChart renders a single-row sparkline chart (kept for sparklines).
func renderBrailleChart(values []int, width int, col color.Color) string {
	if width <= 0 || len(values) == 0 {
		return ""
	}

	sampled := resample(values, width)

	max := 0
	nonZero := 0
	for _, v := range sampled {
		if v > max {
			max = v
		}
		if v > 0 {
			nonZero++
		}
	}
	if max == 0 || nonZero < 2 {
		return ""
	}

	bars := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇'}

	var b strings.Builder
	for _, v := range sampled {
		level := v * (len(bars) - 1) / max
		if level >= len(bars) {
			level = len(bars) - 1
		}
		b.WriteRune(bars[level])
	}

	return lipgloss.NewStyle().Foreground(col).Render(b.String())
}

// resample maps an input slice to a target length by averaging adjacent values.
func resample(values []int, targetLen int) []int {
	if len(values) <= targetLen {
		result := make([]int, targetLen)
		offset := targetLen - len(values)
		copy(result[offset:], values)
		return result
	}

	result := make([]int, targetLen)
	ratio := float64(len(values)) / float64(targetLen)
	for i := 0; i < targetLen; i++ {
		start := int(float64(i) * ratio)
		end := int(float64(i+1) * ratio)
		if end > len(values) {
			end = len(values)
		}
		sum := 0
		count := 0
		for j := start; j < end; j++ {
			sum += values[j]
			count++
		}
		if count > 0 {
			result[i] = sum / count
		}
	}
	return result
}
