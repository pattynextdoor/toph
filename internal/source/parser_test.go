package source

import (
	"os"
	"testing"

	"github.com/pattynextdoor/toph/internal/data"
)

func TestParseLine_UserRecord(t *testing.T) {
	line := []byte(`{"type":"user","uuid":"u1","timestamp":"2026-03-30T09:21:18.876Z","sessionId":"sess-001","cwd":"/Users/patty/dev/toph","gitBranch":"main","version":"2.1.87","message":{"role":"user","content":"build the thing"}}`)

	events := ParseLine(line, "toph")
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	ev := events[0]
	if ev.Type != data.EventUserMessage {
		t.Errorf("expected EventUserMessage, got %v", ev.Type)
	}
	if ev.SessionID != "sess-001" {
		t.Errorf("expected session sess-001, got %s", ev.SessionID)
	}
	if ev.Text != "build the thing" {
		t.Errorf("expected 'build the thing', got %q", ev.Text)
	}
	if ev.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestParseLine_UserRecord_TruncatesLongText(t *testing.T) {
	// 120 character message should be truncated to 100.
	longText := ""
	for i := 0; i < 120; i++ {
		longText += "x"
	}
	line := []byte(`{"type":"user","uuid":"u2","timestamp":"2026-03-30T09:21:18.876Z","sessionId":"sess-001","message":{"role":"user","content":"` + longText + `"}}`)

	events := ParseLine(line, "toph")
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if len(events[0].Text) != 100 {
		t.Errorf("expected text truncated to 100 chars, got %d", len(events[0].Text))
	}
}

func TestParseLine_AssistantWithToolUse(t *testing.T) {
	line := []byte(`{"type":"assistant","uuid":"a1","timestamp":"2026-03-30T09:21:22.397Z","parentUuid":"u1","sessionId":"sess-001","cwd":"/Users/patty/dev/toph","gitBranch":"main","version":"2.1.87","message":{"model":"claude-opus-4-6","id":"msg_01X","type":"message","role":"assistant","content":[{"type":"text","text":"I will build it."},{"type":"tool_use","id":"tu_1","name":"Bash","input":{"command":"ls"}}],"stop_reason":"tool_use","usage":{"input_tokens":1500,"output_tokens":200,"cache_creation_input_tokens":100,"cache_read_input_tokens":800}}}`)

	events := ParseLine(line, "toph")
	if len(events) != 2 {
		t.Fatalf("expected 2 events (tool_use + assistant), got %d", len(events))
	}

	// First event should be the tool_use.
	toolEv := events[0]
	if toolEv.Type != data.EventToolUse {
		t.Errorf("expected EventToolUse, got %v", toolEv.Type)
	}
	if toolEv.ToolName != "Bash" {
		t.Errorf("expected ToolName=Bash, got %s", toolEv.ToolName)
	}
	if toolEv.ToolInput != "ls" {
		t.Errorf("expected ToolInput=ls, got %q", toolEv.ToolInput)
	}

	// Second event should be the assistant text with usage.
	textEv := events[1]
	if textEv.Type != data.EventAssistantText {
		t.Errorf("expected EventAssistantText, got %v", textEv.Type)
	}
	if textEv.Model != "claude-opus-4-6" {
		t.Errorf("expected model claude-opus-4-6, got %s", textEv.Model)
	}
	if textEv.InputTokens != 1500 {
		t.Errorf("expected 1500 input tokens, got %d", textEv.InputTokens)
	}
	if textEv.OutputTokens != 200 {
		t.Errorf("expected 200 output tokens, got %d", textEv.OutputTokens)
	}
	if textEv.CacheCreationInputTokens != 100 {
		t.Errorf("expected 100 cache creation tokens, got %d", textEv.CacheCreationInputTokens)
	}
	if textEv.CacheReadInputTokens != 800 {
		t.Errorf("expected 800 cache read tokens, got %d", textEv.CacheReadInputTokens)
	}
	if textEv.StopReason != "tool_use" {
		t.Errorf("expected stop_reason=tool_use, got %s", textEv.StopReason)
	}
	if textEv.Text != "I will build it." {
		t.Errorf("expected assistant text 'I will build it.', got %q", textEv.Text)
	}
}

func TestParseLine_InvalidJSON(t *testing.T) {
	events := ParseLine([]byte(`{not valid json`), "toph")
	if len(events) != 0 {
		t.Errorf("expected 0 events for invalid JSON, got %d", len(events))
	}
}

func TestParseLine_UnknownType(t *testing.T) {
	events := ParseLine([]byte(`{"type":"queue-operation","sessionId":"sess-001"}`), "toph")
	if len(events) != 0 {
		t.Errorf("expected 0 events for unknown type, got %d", len(events))
	}
}

func TestParseLine_EmptyLine(t *testing.T) {
	events := ParseLine([]byte(""), "toph")
	if len(events) != 0 {
		t.Errorf("expected 0 events for empty line, got %d", len(events))
	}
}

func TestParseFile(t *testing.T) {
	b, err := os.ReadFile("testdata/sample.jsonl")
	if err != nil {
		t.Fatalf("failed to read testdata/sample.jsonl: %v", err)
	}

	events := ParseBytes(b, "toph")

	// Expected events from sample.jsonl:
	// Line 1 (user): 1 EventUserMessage
	// Line 2 (assistant with tool_use): 1 EventToolUse + 1 EventAssistantText = 2
	// Line 3 (assistant, text only): 1 EventAssistantText
	// Line 4 (system): 1 EventSystemMessage
	// Total: 5
	if len(events) != 5 {
		t.Fatalf("expected 5 events from sample.jsonl, got %d", len(events))
	}

	// Verify event types in order.
	expected := []data.EventType{
		data.EventUserMessage,
		data.EventToolUse,
		data.EventAssistantText,
		data.EventAssistantText,
		data.EventSystemMessage,
	}
	for i, want := range expected {
		if events[i].Type != want {
			t.Errorf("event[%d]: expected %v, got %v", i, want, events[i].Type)
		}
	}

	// Verify the user message text.
	if events[0].Text != "build the thing" {
		t.Errorf("user event text: expected 'build the thing', got %q", events[0].Text)
	}

	// Verify the tool use event.
	if events[1].ToolName != "Bash" {
		t.Errorf("tool event: expected Bash, got %s", events[1].ToolName)
	}

	// All session IDs should be sess-001.
	for i, ev := range events {
		if ev.SessionID != "sess-001" {
			t.Errorf("event[%d]: expected sessionId sess-001, got %s", i, ev.SessionID)
		}
	}
}
