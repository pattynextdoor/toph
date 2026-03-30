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
	mu       sync.RWMutex
	sessions map[string]*Session
	feed     *RingBuffer
}

// NewManager creates a Manager with an empty session map and a 1,000-event ring buffer.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		feed:     NewRingBuffer(ActivityBufferSize),
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
	m.feed.Push(e)
}

// Sessions returns all known sessions sorted by most recently updated first.
func (m *Manager) Sessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result
}

// ActiveSessions returns sessions whose UpdatedAt is within the given threshold
// of the current time, sorted by most recently updated first.
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
	sort.Slice(result, func(i, j int) bool {
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
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
