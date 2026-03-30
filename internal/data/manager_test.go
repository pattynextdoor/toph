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
