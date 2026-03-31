package panels

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/pattynextdoor/toph/internal/ui"
)

// HelpPanel renders a centered overlay showing all keybindings. It is toggled
// with the ? key and dismissed with ? or Esc.
type HelpPanel struct {
	theme   *ui.Theme
	Visible bool
}

// NewHelpPanel creates a HelpPanel bound to the given theme.
func NewHelpPanel(theme *ui.Theme) *HelpPanel {
	return &HelpPanel{theme: theme}
}

// Toggle flips the help overlay visibility.
func (p *HelpPanel) Toggle() {
	p.Visible = !p.Visible
}

// Hide dismisses the help overlay.
func (p *HelpPanel) Hide() {
	p.Visible = false
}

// keybinding pairs a key label with its description for rendering.
type keybinding struct {
	key  string
	desc string
}

var helpBindings = []keybinding{
	{"tab / shift+tab", "Cycle panel focus"},
	{"j / k / up / down", "Navigate within panel"},
	{"1-5", "Jump to panel"},
	{"enter", "Select session"},
	{"G / g", "Activity feed bottom/top"},
	{"/", "Filter"},
	{"esc", "Clear filter / dismiss"},
	{"r", "Force refresh"},
	{"ctrl+l", "Clear feed"},
	{"?", "Toggle this help"},
	{"q / ctrl+c", "Quit"},
}

// Render draws the help overlay as a centered box with rounded borders.
// The box uses the focus color to stand out from the underlying dashboard.
func (p *HelpPanel) Render(width, height int) string {
	// Find the longest key label for alignment.
	maxKeyLen := 0
	for _, b := range helpBindings {
		if len(b.key) > maxKeyLen {
			maxKeyLen = len(b.key)
		}
	}

	var rows []string
	for _, b := range helpBindings {
		keyStyle := lipgloss.NewStyle().
			Foreground(p.theme.BorderFocus).
			Bold(true).
			Width(maxKeyLen + 2)
		descStyle := lipgloss.NewStyle().
			Foreground(p.theme.TextDim)

		row := keyStyle.Render(b.key) + descStyle.Render(b.desc)
		rows = append(rows, row)
	}

	title := lipgloss.NewStyle().
		Foreground(p.theme.BorderFocus).
		Bold(true).
		Align(lipgloss.Center).
		Render("toph — keybindings")

	body := strings.Join(rows, "\n")
	content := title + "\n\n" + body

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.theme.BorderFocus).
		Padding(1, 3).
		Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
