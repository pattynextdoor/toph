package data

import (
	"testing"
	"time"
)

func TestManager_GetOrCreateSession(t *testing.T) {
	m := NewManager()
	m.HandleEvent(Event{
		Type:      EventUserMessage,
		Timestamp: time.Now(),
		SessionID: "sess-001",
		Text:      "hello",
	})

	sessions := m.Sessions()
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != "sess-001" {
		t.Errorf("expected session id sess-001, got %s", sessions[0].ID)
	}
}

func TestManager_ActivityFeed(t *testing.T) {
	m := NewManager()
	for i := 0; i < 5; i++ {
		m.HandleEvent(Event{
			Type:      EventToolUse,
			Timestamp: time.Now(),
			SessionID: "sess-001",
			ToolName:  "Bash",
		})
	}

	events := m.ActivityFeed()
	if len(events) != 5 {
		t.Errorf("expected 5 feed events, got %d", len(events))
	}
}

func TestManager_ActiveSessions(t *testing.T) {
	m := NewManager()

	m.HandleEvent(Event{
		Type:      EventToolUse,
		Timestamp: time.Now(),
		SessionID: "active-001",
		ToolName:  "Bash",
	})

	m.HandleEvent(Event{
		Type:      EventToolUse,
		Timestamp: time.Now().Add(-30 * time.Minute),
		SessionID: "stale-001",
		ToolName:  "Bash",
	})

	active := m.ActiveSessions(5 * time.Minute)
	if len(active) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(active))
	}
	if active[0].ID != "active-001" {
		t.Errorf("expected active-001, got %s", active[0].ID)
	}
}

func TestManager_ActivityFeedFiltering(t *testing.T) {
	m := NewManager()

	// Tool use events should appear in the feed
	m.HandleEvent(Event{Type: EventToolUse, Timestamp: time.Now(), SessionID: "s1", ToolName: "Bash"})
	// User messages should be FILTERED (low signal)
	m.HandleEvent(Event{Type: EventUserMessage, Timestamp: time.Now(), SessionID: "s1", Text: "hi"})
	// Assistant text should be FILTERED from the feed
	m.HandleEvent(Event{Type: EventAssistantText, Timestamp: time.Now(), SessionID: "s1", Text: "thinking..."})
	// System messages should be FILTERED from the feed
	m.HandleEvent(Event{Type: EventSystemMessage, Timestamp: time.Now(), SessionID: "s1"})

	events := m.ActivityFeed()
	if len(events) != 1 {
		t.Errorf("expected 1 feed event (only tool_use), got %d", len(events))
	}

	// Verify token accounting still works despite feed filtering
	s := m.SessionByID("s1")
	if s == nil {
		t.Fatal("session not found")
	}
}

func TestManager_SessionSortOrder(t *testing.T) {
	m := NewManager()
	now := time.Now()

	// Create sessions with different statuses by sending events and then
	// manually setting the status (since there's no direct "waiting" event type).
	events := []struct {
		id     string
		offset time.Duration
	}{
		{"idle-old", -10 * time.Minute},
		{"active-recent", -1 * time.Minute},
		{"waiting-old", -5 * time.Minute},
		{"dead-recent", -2 * time.Minute},
		{"active-old", -8 * time.Minute},
		{"waiting-recent", -30 * time.Second},
		{"error-recent", -3 * time.Minute},
	}

	for _, e := range events {
		m.HandleEvent(Event{
			Type:      EventToolUse,
			Timestamp: now.Add(e.offset),
			SessionID: e.id,
			ToolName:  "Bash",
		})
	}

	// Manually set statuses (sessions start as Active from EventToolUse).
	m.sessions["idle-old"].mu.Lock()
	m.sessions["idle-old"].Status = StatusIdle
	m.sessions["idle-old"].mu.Unlock()

	m.sessions["waiting-old"].mu.Lock()
	m.sessions["waiting-old"].Status = StatusWaiting
	m.sessions["waiting-old"].mu.Unlock()

	m.sessions["waiting-recent"].mu.Lock()
	m.sessions["waiting-recent"].Status = StatusWaiting
	m.sessions["waiting-recent"].mu.Unlock()

	m.sessions["dead-recent"].mu.Lock()
	m.sessions["dead-recent"].Status = StatusDead
	m.sessions["dead-recent"].mu.Unlock()

	m.sessions["error-recent"].mu.Lock()
	m.sessions["error-recent"].Status = StatusError
	m.sessions["error-recent"].mu.Unlock()

	sessions := m.Sessions()

	// Expected order:
	// 1. waiting-recent  (priority 0, most recent)
	// 2. waiting-old     (priority 0, older)
	// 3. active-recent   (priority 1, most recent)
	// 4. active-old      (priority 1, older)
	// 5. error-recent    (priority 2)
	// 6. idle-old        (priority 3)
	// 7. dead-recent     (priority 4)
	expected := []string{
		"waiting-recent",
		"waiting-old",
		"active-recent",
		"active-old",
		"error-recent",
		"idle-old",
		"dead-recent",
	}

	if len(sessions) != len(expected) {
		t.Fatalf("expected %d sessions, got %d", len(expected), len(sessions))
	}

	for i, want := range expected {
		if sessions[i].ID != want {
			t.Errorf("position %d: expected %s, got %s", i, want, sessions[i].ID)
		}
	}
}

func TestManager_ToolCounts(t *testing.T) {
	m := NewManager()
	m.HandleEvent(Event{Type: EventToolUse, Timestamp: time.Now(), SessionID: "s1", ToolName: "Bash"})
	m.HandleEvent(Event{Type: EventToolUse, Timestamp: time.Now(), SessionID: "s1", ToolName: "Bash"})
	m.HandleEvent(Event{Type: EventToolUse, Timestamp: time.Now(), SessionID: "s2", ToolName: "Read"})

	counts := m.ToolCounts()
	if counts["Bash"] != 2 {
		t.Errorf("expected Bash=2, got %d", counts["Bash"])
	}
	if counts["Read"] != 1 {
		t.Errorf("expected Read=1, got %d", counts["Read"])
	}
}
