package data

import (
	"sync"
	"time"
)

// SessionStatus represents the current state of a monitored session.
type SessionStatus int

const (
	StatusActive SessionStatus = iota
	StatusWaiting
	StatusIdle
	StatusError
	StatusDead
)

// StatusPriority returns a sort weight for the status. Lower values sort first,
// so waiting sessions (needing human attention) float to the top.
func (s SessionStatus) StatusPriority() int {
	switch s {
	case StatusWaiting:
		return 0
	case StatusActive:
		return 1
	case StatusError:
		return 2
	case StatusIdle:
		return 3
	case StatusDead:
		return 4
	default:
		return 5
	}
}

func (s SessionStatus) String() string {
	switch s {
	case StatusActive:
		return "active"
	case StatusWaiting:
		return "waiting"
	case StatusIdle:
		return "idle"
	case StatusError:
		return "error"
	case StatusDead:
		return "dead"
	default:
		return "unknown"
	}
}

// Session holds the accumulated state for a single Claude Code session,
// built up incrementally as Events arrive.
type Session struct {
	mu sync.RWMutex

	ID        string
	Project   string
	CWD       string
	GitBranch string
	Version   string
	Model     string

	Status    SessionStatus
	StartedAt time.Time
	UpdatedAt time.Time

	TotalInputTokens  int
	TotalOutputTokens int
	TotalCacheRead    int
	TotalCacheWrite   int
	ToolCounts        map[string]int

	Subagents []*Subagent

	LastToolName   string
	LastText       string
	LastStopReason string
	LastHadToolUse bool      // true if the last assistant message contained tool_use blocks
	LastEventAt    time.Time // wall-clock time when the last event arrived (for timeout detection)
}

// Subagent represents a child agent spawned by a session.
type Subagent struct {
	ID          string
	Type        string
	Description string
	Status      SessionStatus
	StartedAt   time.Time
	UpdatedAt   time.Time
}

// RLock acquires a read lock on the session for reading multiple fields
// atomically.
func (s *Session) RLock() { s.mu.RLock() }

// RUnlock releases the read lock.
func (s *Session) RUnlock() { s.mu.RUnlock() }

// ContextWindowSize returns the context window token limit for the given model.
// All current Claude models use 200K tokens.
func ContextWindowSize(model string) int {
	return 200_000
}

// NewSession creates a Session with sensible defaults.
func NewSession(id, project string) *Session {
	now := time.Now()
	return &Session{
		ID:         id,
		Project:    project,
		Status:     StatusActive,
		StartedAt:  now,
		UpdatedAt:  now,
		ToolCounts: make(map[string]int),
	}
}

// UpdateFromEvent applies an incoming event to mutate session state.
// Thread-safe via internal mutex.
func (s *Session) UpdateFromEvent(e Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.UpdatedAt = e.Timestamp
	if e.Model != "" {
		s.Model = e.Model
	}
	if e.Project != "" {
		s.Project = e.Project
	}
	if e.CWD != "" {
		s.CWD = e.CWD
	}
	if e.GitBranch != "" {
		s.GitBranch = e.GitBranch
	}

	s.LastEventAt = time.Now()

	// Track subagent state when the event belongs to a child agent.
	if e.AgentID != "" {
		sa := s.getOrCreateSubagent(e.AgentID)
		sa.UpdatedAt = e.Timestamp
		switch e.Type {
		case EventSubagentStart:
			sa.Status = StatusActive
		case EventSubagentEnd:
			sa.Status = StatusIdle
		case EventToolUse, EventAssistantText, EventUserMessage:
			sa.Status = StatusActive
		case EventError:
			sa.Status = StatusError
		}
	}

	switch e.Type {
	case EventToolUse:
		s.Status = StatusActive
		s.LastToolName = e.ToolName
		s.LastHadToolUse = true
		s.ToolCounts[e.ToolName]++
	case EventAssistantText:
		s.Status = StatusActive
		s.LastStopReason = e.StopReason
		s.TotalInputTokens += e.InputTokens
		s.TotalOutputTokens += e.OutputTokens
		s.TotalCacheRead += e.CacheReadInputTokens
		s.TotalCacheWrite += e.CacheCreationInputTokens
		if e.Text != "" {
			s.LastText = e.Text
		}
	case EventUserMessage:
		s.Status = StatusActive
		s.LastHadToolUse = false
		s.LastStopReason = ""
	case EventError:
		s.Status = StatusError
	}
}

// getOrCreateSubagent finds an existing subagent by ID or creates a new one.
// Must be called with s.mu held.
func (s *Session) getOrCreateSubagent(agentID string) *Subagent {
	for _, sa := range s.Subagents {
		if sa.ID == agentID {
			return sa
		}
	}
	sa := &Subagent{
		ID:        agentID,
		Status:    StatusActive,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.Subagents = append(s.Subagents, sa)
	return sa
}

// waitingTimeout is how long after the last event we wait before considering
// an active session with pending tool use to be "waiting for permission."
const waitingTimeout = 15 * time.Second

// idleTimeout is how long a session can go without any events before it
// transitions from Active to Idle.
const idleTimeout = 5 * time.Minute

// CheckStale updates session status based on inactivity:
//   - Active sessions with a pending tool_use and no events for >15s → Waiting
//   - Active sessions with no events for >5 minutes → Idle
//
// The waiting check fires first because it's more specific (the user likely
// needs to approve a tool call). Call this periodically (e.g., on every tick).
func (s *Session) CheckStale() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status != StatusActive {
		return
	}

	if s.LastEventAt.IsZero() {
		return
	}

	elapsed := time.Since(s.LastEventAt)

	// Check for permission-waiting first (more specific).
	if s.LastHadToolUse && elapsed > waitingTimeout {
		s.Status = StatusWaiting
		return
	}

	// Check for idle (general inactivity).
	if elapsed > idleTimeout {
		s.Status = StatusIdle
	}
}
