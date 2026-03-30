package data

import (
	"testing"
	"time"
)

func TestSessionUpdateFromEvent(t *testing.T) {
	s := NewSession("test-123", "my-project")

	s.UpdateFromEvent(Event{
		Type:      EventToolUse,
		Timestamp: time.Now(),
		ToolName:  "Bash",
	})

	if s.Status != StatusActive {
		t.Errorf("expected status Active, got %s", s.Status)
	}
	if s.ToolCounts["Bash"] != 1 {
		t.Errorf("expected Bash count 1, got %d", s.ToolCounts["Bash"])
	}

	s.UpdateFromEvent(Event{
		Type:         EventAssistantText,
		Timestamp:    time.Now(),
		InputTokens:  1000,
		OutputTokens: 500,
		Model:        "claude-opus-4-6",
	})

	if s.TotalInputTokens != 1000 {
		t.Errorf("expected 1000 input tokens, got %d", s.TotalInputTokens)
	}
	if s.TotalOutputTokens != 500 {
		t.Errorf("expected 500 output tokens, got %d", s.TotalOutputTokens)
	}
	if s.Model != "claude-opus-4-6" {
		t.Errorf("expected model claude-opus-4-6, got %s", s.Model)
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		et   EventType
		want string
	}{
		{EventToolUse, "tool_use"},
		{EventUserMessage, "user"},
		{EventAssistantText, "assistant"},
	}
	for _, tt := range tests {
		if got := tt.et.String(); got != tt.want {
			t.Errorf("EventType(%d).String() = %q, want %q", tt.et, got, tt.want)
		}
	}
}
