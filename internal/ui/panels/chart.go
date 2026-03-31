package panels

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Braille dot patterns for 2-high charts. Each braille cell is 2 columns x 4 rows
// of dots. For a single-column chart we use the left column only:
//
//	Row 0: ⠁ (dot 1)
//	Row 1: ⠂ (dot 2)
//	Row 2: ⠄ (dot 3)
//	Row 3: ⡀ (dot 7)
//
// For a 2-row chart (8 vertical levels), the top row uses dots 1-4 of one cell
// and the bottom row uses dots 1-4 of the cell below.

// renderBrailleChart renders a single-row braille sparkline chart from a slice
// of values. Each value maps to a braille character with 1-8 levels of fill.
// Width is the number of braille characters to render.
func renderBrailleChart(values []int, width int, col color.Color) string {
	if width <= 0 || len(values) == 0 {
		return ""
	}

	// Resample values to fit the desired width
	sampled := resample(values, width)

	// Find max for scaling
	max := 1
	for _, v := range sampled {
		if v > max {
			max = v
		}
	}

	// Braille characters for 8 vertical levels (empty to full)
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
		// Pad left with zeros
		result := make([]int, targetLen)
		offset := targetLen - len(values)
		copy(result[offset:], values)
		return result
	}

	// Downsample by averaging
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
