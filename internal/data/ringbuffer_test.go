package data

import (
	"fmt"
	"testing"
	"time"
)

func TestRingBuffer_BasicOps(t *testing.T) {
	rb := NewRingBuffer(3)

	if rb.Len() != 0 {
		t.Fatal("new buffer should be empty")
	}

	e1 := Event{Type: EventToolUse, ToolName: "Read", Timestamp: time.Now()}
	e2 := Event{Type: EventToolUse, ToolName: "Edit", Timestamp: time.Now()}
	e3 := Event{Type: EventToolUse, ToolName: "Bash", Timestamp: time.Now()}
	e4 := Event{Type: EventToolUse, ToolName: "Glob", Timestamp: time.Now()}

	rb.Push(e1)
	rb.Push(e2)
	rb.Push(e3)

	if rb.Len() != 3 {
		t.Fatalf("expected len 3, got %d", rb.Len())
	}

	rb.Push(e4)
	if rb.Len() != 3 {
		t.Fatalf("expected len 3 after overflow, got %d", rb.Len())
	}

	items := rb.All()
	if items[0].ToolName != "Edit" {
		t.Errorf("expected oldest to be Edit, got %s", items[0].ToolName)
	}
	if items[2].ToolName != "Glob" {
		t.Errorf("expected newest to be Glob, got %s", items[2].ToolName)
	}
}

func TestRingBuffer_Slice(t *testing.T) {
	rb := NewRingBuffer(100)
	for i := 0; i < 10; i++ {
		rb.Push(Event{Type: EventToolUse, ToolName: fmt.Sprintf("tool-%d", i)})
	}

	slice := rb.Slice(3, 6)
	if len(slice) != 3 {
		t.Fatalf("expected slice len 3, got %d", len(slice))
	}
	if slice[0].ToolName != "tool-3" {
		t.Errorf("expected tool-3, got %s", slice[0].ToolName)
	}
}
