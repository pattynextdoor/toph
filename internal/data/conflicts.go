package data

import (
	"sync"
	"time"
)

const conflictWindow = 5 * time.Minute

// FileTouch records a session touching a file.
type FileTouch struct {
	SessionID string
	Timestamp time.Time
}

// ConflictTracker detects when multiple sessions modify the same file
// within a time window. It is safe for concurrent use.
type ConflictTracker struct {
	mu      sync.RWMutex
	touches map[string][]FileTouch // file path -> list of touches
}

// NewConflictTracker creates an empty ConflictTracker.
func NewConflictTracker() *ConflictTracker {
	return &ConflictTracker{
		touches: make(map[string][]FileTouch),
	}
}

// RecordTouch records a session touching a file. Returns true if this
// creates a conflict (another session touched the same file recently).
func (ct *ConflictTracker) RecordTouch(filePath, sessionID string, ts time.Time) bool {
	if filePath == "" {
		return false
	}

	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Prune old touches for this file
	cutoff := ts.Add(-conflictWindow)
	existing := ct.touches[filePath]
	var active []FileTouch
	for _, t := range existing {
		if t.Timestamp.After(cutoff) {
			active = append(active, t)
		}
	}

	// Check if any OTHER session touched this file recently
	conflict := false
	for _, t := range active {
		if t.SessionID != sessionID {
			conflict = true
			break
		}
	}

	// Record this touch
	active = append(active, FileTouch{SessionID: sessionID, Timestamp: ts})
	ct.touches[filePath] = active

	return conflict
}

// Conflicts returns all files currently touched by 2+ sessions.
// Keys are file paths, values are the distinct session IDs involved.
func (ct *ConflictTracker) Conflicts() map[string][]string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	cutoff := time.Now().Add(-conflictWindow)
	result := make(map[string][]string)

	for path, touches := range ct.touches {
		sessions := make(map[string]bool)
		for _, t := range touches {
			if t.Timestamp.After(cutoff) {
				sessions[t.SessionID] = true
			}
		}
		if len(sessions) >= 2 {
			var ids []string
			for id := range sessions {
				ids = append(ids, id)
			}
			result[path] = ids
		}
	}
	return result
}

// ConflictCount returns the number of files currently in conflict.
func (ct *ConflictTracker) ConflictCount() int {
	return len(ct.Conflicts())
}

// IsConflicted returns true if the given file path is currently touched
// by multiple sessions.
func (ct *ConflictTracker) IsConflicted(filePath string) bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	cutoff := time.Now().Add(-conflictWindow)
	touches := ct.touches[filePath]
	sessions := make(map[string]bool)
	for _, t := range touches {
		if t.Timestamp.After(cutoff) {
			sessions[t.SessionID] = true
		}
	}
	return len(sessions) >= 2
}
