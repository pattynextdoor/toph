package data

import (
	"testing"
	"time"
)

func TestConflictTracker_NoConflict(t *testing.T) {
	ct := NewConflictTracker()
	now := time.Now()

	// Same session touching same file = no conflict
	if ct.RecordTouch("/src/main.go", "sess-1", now) {
		t.Error("expected no conflict for first touch")
	}
	if ct.RecordTouch("/src/main.go", "sess-1", now.Add(time.Second)) {
		t.Error("expected no conflict for same session")
	}
}

func TestConflictTracker_Conflict(t *testing.T) {
	ct := NewConflictTracker()
	now := time.Now()

	ct.RecordTouch("/src/main.go", "sess-1", now)
	if !ct.RecordTouch("/src/main.go", "sess-2", now.Add(time.Second)) {
		t.Error("expected conflict when second session touches same file")
	}

	if ct.ConflictCount() != 1 {
		t.Errorf("expected 1 conflict, got %d", ct.ConflictCount())
	}
}

func TestConflictTracker_Expiry(t *testing.T) {
	ct := NewConflictTracker()
	now := time.Now()

	// Touch from 6 minutes ago (outside window)
	ct.RecordTouch("/src/main.go", "sess-1", now.Add(-6*time.Minute))
	// New touch from different session
	if ct.RecordTouch("/src/main.go", "sess-2", now) {
		t.Error("expected no conflict -- first touch expired")
	}
}

func TestConflictTracker_EmptyPath(t *testing.T) {
	ct := NewConflictTracker()
	now := time.Now()

	if ct.RecordTouch("", "sess-1", now) {
		t.Error("expected no conflict for empty path")
	}
}

func TestConflictTracker_MultipleFiles(t *testing.T) {
	ct := NewConflictTracker()
	now := time.Now()

	ct.RecordTouch("/src/a.go", "sess-1", now)
	ct.RecordTouch("/src/b.go", "sess-1", now)
	ct.RecordTouch("/src/a.go", "sess-2", now.Add(time.Second))
	ct.RecordTouch("/src/b.go", "sess-2", now.Add(time.Second))

	if ct.ConflictCount() != 2 {
		t.Errorf("expected 2 conflicts, got %d", ct.ConflictCount())
	}

	if !ct.IsConflicted("/src/a.go") {
		t.Error("expected /src/a.go to be conflicted")
	}
	if !ct.IsConflicted("/src/b.go") {
		t.Error("expected /src/b.go to be conflicted")
	}
}
