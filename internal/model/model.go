package model

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pattynextdoor/toph/internal/data"
	"github.com/pattynextdoor/toph/internal/ui"
	"github.com/pattynextdoor/toph/internal/ui/panels"
)

// tickInterval is the target frame time for the 30fps render loop. Events that
// arrive between ticks are buffered in pendingEvents and flushed into the
// Manager on each tick — this batches rapid-fire JSONL events into a single
// render pass, preventing flicker during bursts.
const tickInterval = 33 * time.Millisecond

// tickMsg is the internal message sent by tea.Tick at ~30fps to drive the
// render loop.
type tickMsg time.Time

// EventMsg wraps a data.Event so it can be sent into the Bubble Tea program
// from external goroutines (e.g. the JSONL source). It is exported because
// main.go needs to construct these when bridging source channels into
// tea.Program.Send.
type EventMsg data.Event

// Model is the root Bubble Tea model. It owns the data Manager (session state
// + activity ring buffer), all panel instances, and the 30fps tick loop.
// Incoming events are buffered between ticks and flushed to the Manager once
// per frame.
type Model struct {
	manager   *data.Manager
	theme     *ui.Theme
	layout    ui.Layout
	statusBar *ui.StatusBar

	sessions *panels.SessionsPanel
	detail   *panels.DetailPanel
	activity *panels.ActivityPanel
	tools    *panels.ToolsPanel

	focusedPanel  int
	width         int
	height        int
	ready         bool
	pendingEvents []data.Event
}

// New creates a root Model wired to the given Manager. All panels start
// unfocused; the sessions panel will receive focus once the first tick fires.
func New(manager *data.Manager) Model {
	theme := ui.DefaultTheme()
	return Model{
		manager:   manager,
		theme:     theme,
		statusBar: ui.NewStatusBar(theme),
		sessions:  panels.NewSessionsPanel(theme),
		detail:    panels.NewDetailPanel(theme),
		activity:  panels.NewActivityPanel(theme),
		tools:     panels.NewToolsPanel(theme),
	}
}

// Init kicks off the 30fps tick loop. The first WindowSizeMsg from Bubble Tea
// will set m.ready = true and trigger the initial layout computation.
func (m Model) Init() tea.Cmd {
	return tick()
}

// Update handles all incoming messages: window resizes, tick-driven flushes,
// buffered events from sources, and keyboard input.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout = ui.ComputeLayout(m.width, m.height)
		m.ready = true
		return m, nil

	case tickMsg:
		// Flush buffered events into the Manager once per frame.
		for _, e := range m.pendingEvents {
			m.manager.HandleEvent(e)
		}
		m.pendingEvents = m.pendingEvents[:0]
		return m, tick()

	case EventMsg:
		// Buffer incoming events; they'll be flushed on the next tick.
		m.pendingEvents = append(m.pendingEvents, data.Event(msg))
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey processes keyboard input — quit, panel focus cycling, and
// panel-specific navigation (scroll, cursor movement).
func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if isQuit(msg) {
		return m, tea.Quit
	}

	if panel, ok := isJumpToPanel(msg); ok {
		m.setFocus(panel)
		return m, nil
	}
	if isTab(msg) {
		m.setFocus((m.focusedPanel + 1) % PanelCount)
		return m, nil
	}
	if isShiftTab(msg) {
		m.setFocus((m.focusedPanel - 1 + PanelCount) % PanelCount)
		return m, nil
	}

	// Route navigation keys to the currently focused panel.
	switch m.focusedPanel {
	case PanelSessions:
		if isUp(msg) {
			m.sessions.MoveUp()
		}
		if isDown(msg) {
			m.sessions.MoveDown(len(m.manager.Sessions()))
		}
	case PanelActivity:
		switch {
		case isUp(msg):
			m.activity.ScrollUp(1)
		case isDown(msg):
			m.activity.ScrollDown(1)
		case msg.String() == "G":
			m.activity.ScrollToBottom()
		case msg.String() == "g":
			m.activity.ScrollToTop()
		}
	}

	return m, nil
}

// setFocus updates the focused panel index and propagates focus state to each
// panel so they can adjust their border styling.
func (m *Model) setFocus(panel int) {
	m.focusedPanel = panel
	m.sessions.SetFocused(panel == PanelSessions)
	m.detail.SetFocused(panel == PanelDetail)
	m.activity.SetFocused(panel == PanelActivity)
	m.tools.SetFocused(panel == PanelTools)
}

// View composes all panels into the final dashboard layout. It returns a
// "Terminal too small" message if the window is below the minimum usable size,
// or an init message while waiting for the first WindowSizeMsg.
func (m Model) View() tea.View {
	if !m.ready {
		return tea.NewView("Initializing toph...")
	}

	if m.layout.TooSmall {
		return tea.NewView(lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			"Terminal too small (min 60x20)"))
	}

	l := m.layout
	allSessions := m.manager.Sessions()
	feedEvents := m.manager.ActivityFeed()
	toolCounts := m.manager.ToolCounts()
	activeSessions := m.manager.ActiveSessions(data.DefaultActiveThreshold)

	// Render panels.
	sessionsView := m.sessions.Render(allSessions, l.LeftWidth, l.SessionsHeight)
	selectedSession := m.sessions.SelectedSession(allSessions)
	detailView := m.detail.Render(selectedSession, l.LeftWidth, l.DetailHeight)
	activityView := m.activity.Render(feedEvents, l.RightWidth, l.ActivityHeight)
	toolsView := m.tools.Render(toolCounts, l.LeftWidth, l.ToolsHeight)

	// Metrics panel — placeholder until Phase 2.
	metricsStyle := m.theme.PanelNormal
	if m.focusedPanel == PanelMetrics {
		metricsStyle = m.theme.PanelFocus
	}
	metricsView := metricsStyle.
		Width(l.RightWidth - 2).Height(l.MetricsHeight - 2).
		Render(m.theme.Title.Render("METRICS") + "\n" +
			lipgloss.NewStyle().Foreground(m.theme.Idle).Render("Collecting..."))

	// Compose columns: left (sessions/detail/tools), right (activity/metrics).
	leftCol := lipgloss.JoinVertical(lipgloss.Left, sessionsView, detailView, toolsView)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, activityView, metricsView)
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)

	statusBar := m.statusBar.Render(m.width, len(activeSessions), "jsonl", true)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, mainView, statusBar))
}

// tick returns a Cmd that fires a tickMsg after the tick interval, driving the
// 30fps render loop.
func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
