package panels

import (
	"fmt"
	"image/color"
	"math"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/pattynextdoor/toph/internal/data"
	"github.com/pattynextdoor/toph/internal/ui"
)

// DetailPanel renders info about the selected session (Panel 2).
type DetailPanel struct {
	theme   *ui.Theme
	focused bool
	frame   int // incremented each render for animation
}

// NewDetailPanel creates a new detail panel.
func NewDetailPanel(theme *ui.Theme) *DetailPanel {
	return &DetailPanel{theme: theme}
}

func (p *DetailPanel) SetFocused(f bool) { p.focused = f }
func (p *DetailPanel) Focused() bool     { return p.focused }

// Render draws the detail panel for the given session.
// If session is nil, shows a placeholder.
// ProgressRenderer renders a progress bar at a given percent and width.
type ProgressRenderer func(percent float64, width int) string

func (p *DetailPanel) Render(session *data.Session, width, height int, renderProgress ProgressRenderer) string {
	p.frame++

	style := p.theme.PanelNormal
	if p.focused {
		style = p.theme.PanelFocus
	}

	innerW := width - 4
	innerH := height - 2
	if innerW < 1 || innerH < 1 {
		return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render("")
	}

	title := ui.GradientText("DETAIL", p.theme.BorderFocus, p.theme.Subagent)
	var lines []string
	lines = append(lines, title)

	if session == nil {
		lines = append(lines, lipgloss.NewStyle().Foreground(p.theme.Idle).Render("Select a session"))
		return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render(strings.Join(lines, "\n"))
	}

	session.RLock()
	defer session.RUnlock()

	dimStyle := lipgloss.NewStyle().Foreground(p.theme.TextDim)
	valStyle := lipgloss.NewStyle().Foreground(p.theme.BorderFocus)

	// Status line
	var statusIcon string
	var statusColor = p.theme.Active
	switch session.Status {
	case data.StatusActive:
		statusIcon = "● active"
		statusColor = p.theme.Active
	case data.StatusWaiting:
		statusIcon = "◐ waiting"
		// Pulse between bright amber and dim at ~1Hz
		pulse := (math.Sin(float64(p.frame)*0.15) + 1) / 2
		if pulse > 0.5 {
			statusColor = p.theme.Waiting
		} else {
			statusColor = p.theme.Idle
		}
	case data.StatusIdle:
		statusIcon = "○ idle"
		statusColor = p.theme.Idle
	case data.StatusError:
		statusIcon = "✕ error"
		statusColor = p.theme.Error
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(statusColor).Render(statusIcon))

	// Key-value rows with consistent 5-char label column
	label := func(l string) string { return dimStyle.Render(fmt.Sprintf("%-5s", l)) }

	// Working directory
	if session.CWD != "" {
		lines = append(lines, label("dir")+valStyle.Render(shortenPath(session.CWD, innerW-5)))
	}

	// Git branch + model on one line if both present
	if session.GitBranch != "" && session.Model != "" {
		model := shortenModel(session.Model, innerW)
		lines = append(lines, label("git")+valStyle.Render(session.GitBranch)+
			dimStyle.Render("  ")+lipgloss.NewStyle().Foreground(p.theme.Subagent).Render(model))
	} else {
		if session.GitBranch != "" {
			lines = append(lines, label("git")+valStyle.Render(session.GitBranch))
		}
		if session.Model != "" {
			lines = append(lines, label("mod")+lipgloss.NewStyle().Foreground(p.theme.Subagent).Render(shortenModel(session.Model, innerW)))
		}
	}

	// Age + last tool
	age := formatDuration(time.Since(session.StartedAt))
	if session.LastToolName != "" {
		lines = append(lines, label("age")+dimStyle.Render(age)+
			dimStyle.Render("  ")+label("last")+lipgloss.NewStyle().Foreground(p.theme.ToolUse).Render(session.LastToolName))
	} else {
		lines = append(lines, label("age")+dimStyle.Render(age))
	}

	// Tokens + context bar on one line
	totalIn := session.TotalInputTokens
	totalOut := session.TotalOutputTokens
	if totalIn > 0 || totalOut > 0 {
		lines = append(lines, label("tok")+dimStyle.Render(fmt.Sprintf("%s in / %s out",
			formatTokens(totalIn), formatTokens(totalOut))))

		ctxMax := data.ContextWindowSize(session.Model)
		pct := float64(totalIn+totalOut) / float64(ctxMax)
		if pct > 1.0 {
			pct = 1.0
		}
		// Use Bubbles progress bar if available, fall back to ASCII
		barWidth := innerW - 10 // room for "ctx " prefix and " XX%" suffix
		if barWidth < 6 {
			barWidth = 6
		}
		if renderProgress != nil {
			lines = append(lines, dimStyle.Render("ctx ")+renderProgress(pct, barWidth)+
				dimStyle.Render(fmt.Sprintf(" %d%%", int(pct*100))))
		} else {
			lines = append(lines, renderContextBar(totalIn+totalOut, ctxMax, barWidth, p.theme))
		}
	}

	// Subagent list — no header, just inline
	if len(session.Subagents) > 0 {
		maxAgents := innerH - len(lines)
		if maxAgents > len(session.Subagents) {
			maxAgents = len(session.Subagents)
		}
		for i := 0; i < maxAgents; i++ {
			sa := session.Subagents[i]
			connector := "├"
			if i == maxAgents-1 {
				connector = "└"
			}

			var icon string
			var iconColor color.Color
			switch sa.Status {
			case data.StatusActive:
				icon = "●"
				iconColor = p.theme.Active
			case data.StatusIdle:
				icon = "○"
				iconColor = p.theme.Idle
			case data.StatusError:
				icon = "✕"
				iconColor = p.theme.Error
			default:
				icon = "○"
				iconColor = p.theme.Idle
			}

			agentType := sa.Type
			if agentType == "" {
				agentType = "agent"
			}

			iconStr := lipgloss.NewStyle().Foreground(iconColor).Render(icon)
			typeStr := lipgloss.NewStyle().Foreground(p.theme.Subagent).Render(agentType)
			desc := sa.Description
			maxDesc := innerW - 7 - len(agentType)
			if maxDesc > 0 && len(desc) > maxDesc {
				desc = desc[:maxDesc-3] + "..."
			}
			line := dimStyle.Render(" "+connector+" ") + iconStr + " " + typeStr
			if desc != "" && maxDesc > 0 {
				line += dimStyle.Render(": "+desc)
			}
			lines = append(lines, line)
		}
	}

	// Truncate to fit
	if len(lines) > innerH {
		lines = lines[:innerH]
	}

	return style.Width(width - 2).Height(height - 2).MaxWidth(width).MaxHeight(height).Render(strings.Join(lines, "\n"))
}

// shortenModel strips the "claude-" prefix from model names when space is tight.
func shortenModel(model string, availWidth int) string {
	if availWidth < 28 && strings.HasPrefix(model, "claude-") {
		return model[7:] // "claude-opus-4-6" → "opus-4-6"
	}
	return model
}

// shortenPath trims a path to fit within maxLen by showing the last components.
func shortenPath(p string, maxLen int) string {
	if len(p) <= maxLen {
		return p
	}
	// Show last 2 components
	base := filepath.Base(p)
	parent := filepath.Base(filepath.Dir(p))
	short := filepath.Join(parent, base)
	if len(short) <= maxLen {
		return ".../" + short
	}
	return base
}

// renderContextBar builds a colored progress bar showing context window usage.
// Format: `ctx [====----] 8%`
// Color transitions: green (0-50%), amber (50-80%), red (80-100%).
func renderContextBar(filled, total int, barWidth int, theme *ui.Theme) string {
	if total <= 0 {
		return ""
	}
	pct := float64(filled) / float64(total) * 100
	if pct > 100 {
		pct = 100
	}

	filledCells := int(pct / 100 * float64(barWidth))
	if filledCells > barWidth {
		filledCells = barWidth
	}
	emptyCells := barWidth - filledCells

	var barColor color.Color
	switch {
	case pct >= 80:
		barColor = theme.ProgressHigh
	case pct >= 50:
		barColor = theme.ProgressMid
	default:
		barColor = theme.ProgressLow
	}

	filledStyle := lipgloss.NewStyle().Foreground(barColor)
	emptyStyle := lipgloss.NewStyle().Foreground(theme.TextDim)
	dimStyle := lipgloss.NewStyle().Foreground(theme.TextDim)

	bar := filledStyle.Render(strings.Repeat("=", filledCells)) +
		emptyStyle.Render(strings.Repeat("-", emptyCells))

	return dimStyle.Render("ctx ") + bar + dimStyle.Render(fmt.Sprintf(" %d%%", int(pct)))
}

// formatTokens formats a token count in a compact way (e.g., "142K", "1.2M").
func formatTokens(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1000000)
}
