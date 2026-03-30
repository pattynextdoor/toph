package model

import tea "charm.land/bubbletea/v2"

// Panel indices for focus cycling. These map 1:1 to the number keys 1-5 and
// to the Tab/Shift+Tab cycle order.
const (
	PanelSessions = iota
	PanelDetail
	PanelActivity
	PanelMetrics
	PanelTools
	PanelCount
)

func isQuit(msg tea.KeyPressMsg) bool {
	return msg.String() == "q" || msg.String() == "ctrl+c"
}

func isTab(msg tea.KeyPressMsg) bool {
	return msg.String() == "tab"
}

func isShiftTab(msg tea.KeyPressMsg) bool {
	return msg.String() == "shift+tab"
}

func isUp(msg tea.KeyPressMsg) bool {
	k := msg.String()
	return k == "k" || k == "up"
}

func isDown(msg tea.KeyPressMsg) bool {
	k := msg.String()
	return k == "j" || k == "down"
}

func isJumpToPanel(msg tea.KeyPressMsg) (int, bool) {
	switch msg.String() {
	case "1":
		return PanelSessions, true
	case "2":
		return PanelDetail, true
	case "3":
		return PanelActivity, true
	case "4":
		return PanelMetrics, true
	case "5":
		return PanelTools, true
	}
	return 0, false
}
