package ui

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// GradientText renders text with a color gradient from left to right.
func GradientText(text string, from, to color.Color) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	if len(runes) == 1 {
		return lipgloss.NewStyle().Foreground(from).Render(text)
	}

	r1, g1, b1, _ := from.RGBA()
	r2, g2, b2, _ := to.RGBA()

	var b strings.Builder
	for i, ch := range runes {
		t := float64(i) / float64(len(runes)-1)
		r := uint8((float64(r1>>8)*(1-t) + float64(r2>>8)*t))
		g := uint8((float64(g1>>8)*(1-t) + float64(g2>>8)*t))
		bl := uint8((float64(b1>>8)*(1-t) + float64(b2>>8)*t))
		col := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, bl))
		b.WriteString(lipgloss.NewStyle().Foreground(col).Bold(true).Render(string(ch)))
	}
	return b.String()
}

// AnimatedAccumulator smoothly transitions a displayed number toward a target.
// It returns the new display value, moving toward target by a fraction each frame.
func AnimatedAccumulator(current, target float64) float64 {
	if current == 0 {
		return target // first frame, snap to value
	}
	diff := target - current
	if math.Abs(diff) < 0.5 {
		return target // close enough, snap
	}
	// Ease toward target: move 15% of the remaining distance per frame
	return current + diff*0.15
}

// PulseAlpha returns a value between 0.0 and 1.0 that oscillates smoothly
// at the given frequency. frame should be incremented each render.
func PulseAlpha(frame int, speed float64) float64 {
	return (math.Sin(float64(frame)*speed) + 1) / 2
}
