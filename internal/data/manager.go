package data

import (
	"sort"
	"sync"
	"time"
)

const (
	// ActivityBufferSize is the max number of events kept in the global activity feed.
	ActivityBufferSize = 1000

	// DefaultActiveThreshold defines how recently a session must have been updated
	// to be considered "active."
	DefaultActiveThreshold = 5 * time.Minute
)

// Manager owns all session state and the global activity feed.
// It is the central data store that the Bubble Tea model reads from.
type Manager struct {
	mu            sync.RWMutex
	sessions      map[string]*Session
	feed          *RingBuffer
	conflicts     *ConflictTracker
	sampleCounter int
}

// NewManager creates a Manager with an empty session map and a 1,000-event ring buffer.
func NewManager() *Manager {
	return &Manager{
		sessions:  make(map[string]*Session),
		feed:      NewRingBuffer(ActivityBufferSize),
		conflicts: NewConflictTracker(),
	}
}

// HandleEvent routes an incoming event to the appropriate session (creating it
// if necessary) and appends the event to the global activity feed.
func (m *Manager) HandleEvent(e Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[e.SessionID]
	if !ok {
		sess = NewSession(e.SessionID, "")
		m.sessions[e.SessionID] = sess
	}

	sess.UpdateFromEvent(e)

	// Track file touches for conflict detection. When a file-touching tool
	// (Read, Write, Edit) is used, record it. If another session already
	// touched the same file within the conflict window, mark the event.
	if e.Type == EventToolUse && e.ToolInput != "" && isFileTool(e.ToolName) {
		if m.conflicts.RecordTouch(e.ToolInput, e.SessionID, e.Timestamp) {
			e.Conflicted = true
		}
	}

	// Only push high-signal events to the activity feed.
	// Assistant text and system messages are too frequent and noisy.
	if isActivityFeedEvent(e) {
		m.feed.Push(e)
	}
}

// Sessions returns all known sessions sorted by actionability: waiting first,
// then active, then idle/error/dead. Within each group, most recently updated
// sessions come first.
func (m *Manager) Sessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	sortSessionsByActionability(result)
	return result
}

// ActiveSessions returns sessions whose UpdatedAt is within the given threshold
// of the current time, sorted by actionability (same as Sessions).
func (m *Manager) ActiveSessions(threshold time.Duration) []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cutoff := time.Now().Add(-threshold)
	var result []*Session
	for _, s := range m.sessions {
		if s.UpdatedAt.After(cutoff) {
			result = append(result, s)
		}
	}
	sortSessionsByActionability(result)
	return result
}

// ActivityFeed returns all events currently in the ring buffer (oldest first).
func (m *Manager) ActivityFeed() []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.feed.All()
}

// ActivityFeedLast returns the most recent n events from the ring buffer.
func (m *Manager) ActivityFeedLast(n int) []Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.feed.Last(n)
}

// SessionByID returns the session with the given ID, or nil if not found.
func (m *Manager) SessionByID(id string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

// isActivityFeedEvent returns true for events worth showing in the activity
// feed. Filters out high-frequency low-signal events like assistant text
// (think) and system messages that would otherwise drown out tool calls and
// user messages.
func isActivityFeedEvent(e Event) bool {
	switch e.Type {
	case EventAssistantText, EventSystemMessage, EventUserMessage:
		return false
	default:
		return true
	}
}

// sortSessionsByActionability sorts sessions so that the most actionable ones
// appear first: waiting (needs human input) → active → error → idle → dead.
// Within each priority group, most recently updated sessions come first.
func sortSessionsByActionability(sessions []*Session) {
	sort.Slice(sessions, func(i, j int) bool {
		sessions[i].mu.RLock()
		pi := sessions[i].Status.StatusPriority()
		ti := sessions[i].UpdatedAt
		sessions[i].mu.RUnlock()

		sessions[j].mu.RLock()
		pj := sessions[j].Status.StatusPriority()
		tj := sessions[j].UpdatedAt
		sessions[j].mu.RUnlock()

		if pi != pj {
			return pi < pj
		}
		return ti.After(tj)
	})
}

// SessionWaitingInfo identifies a session that just transitioned to waiting.
type SessionWaitingInfo struct {
	ID      string
	Project string
}

// CheckSessionStates runs periodic state checks on all sessions, such as
// detecting permission-waiting timeouts and idle sessions. Called once per
// tick from the model. Returns info for sessions that just transitioned to
// StatusWaiting so the caller can fire desktop notifications.
func (m *Manager) CheckSessionStates() []SessionWaitingInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var waiting []SessionWaitingInfo
	for _, s := range m.sessions {
		if s.CheckStale() {
			s.mu.RLock()
			waiting = append(waiting, SessionWaitingInfo{
				ID:      s.ID,
				Project: s.Project,
			})
			s.mu.RUnlock()
		}
	}
	return waiting
}

// SampleTokenRates samples token burn rates for all sessions' sparklines.
// Call every tick; internally it only fires every ~300 ticks (10s at 30fps).
func (m *Manager) SampleTokenRates() {
	m.sampleCounter++
	if m.sampleCounter < 300 {
		return
	}
	m.sampleCounter = 0

	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, s := range m.sessions {
		s.SampleTokenRate()
	}
}

// SetSubagentMeta attaches metadata (type and description) from a .meta.json
// file to the appropriate subagent within a session. If the subagent hasn't
// been seen yet via events, it creates a placeholder entry.
func (m *Manager) SetSubagentMeta(sessionID, agentID, agentType, description string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[sessionID]
	if !ok {
		return
	}

	sess.mu.Lock()
	defer sess.mu.Unlock()

	for _, sa := range sess.Subagents {
		if sa.ID == agentID {
			sa.Type = agentType
			sa.Description = description
			return
		}
	}
	// Subagent not seen yet via events — create a placeholder.
	sess.Subagents = append(sess.Subagents, &Subagent{
		ID:          agentID,
		Type:        agentType,
		Description: description,
		Status:      StatusActive,
		StartedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
}

// ConflictCount returns the number of files currently in conflict
// (touched by 2+ sessions within the conflict window).
func (m *Manager) ConflictCount() int {
	return m.conflicts.ConflictCount()
}

// isFileTool returns true for tools that touch files on disk.
func isFileTool(name string) bool {
	switch name {
	case "Read", "Write", "Edit":
		return true
	}
	return false
}

// ToolCounts aggregates tool usage counts across all sessions.
func (m *Manager) ToolCounts() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totals := make(map[string]int)
	for _, s := range m.sessions {
		s.mu.RLock()
		for tool, count := range s.ToolCounts {
			totals[tool] += count
		}
		s.mu.RUnlock()
	}
	return totals
}
