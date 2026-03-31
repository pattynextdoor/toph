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

	LastToolName string
	LastText     string
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

	switch e.Type {
	case EventToolUse:
		s.Status = StatusActive
		s.LastToolName = e.ToolName
		s.ToolCounts[e.ToolName]++
	case EventAssistantText:
		s.Status = StatusActive
		s.TotalInputTokens += e.InputTokens
		s.TotalOutputTokens += e.OutputTokens
		s.TotalCacheRead += e.CacheReadInputTokens
		s.TotalCacheWrite += e.CacheCreationInputTokens
		if e.Text != "" {
			s.LastText = e.Text
		}
	case EventUserMessage:
		s.Status = StatusActive
	case EventError:
		s.Status = StatusError
	}
}
