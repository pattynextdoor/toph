package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pattynextdoor/toph/internal/data"
)

func TestJSONLSource_DetectNewLines(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create an empty JSONL file so discoverAndWatch picks it up.
	jsonlPath := filepath.Join(projectDir, "session-001.jsonl")
	if err := os.WriteFile(jsonlPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	src := NewJSONLSource(tmpDir)
	events := make(chan data.Event, 100)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go src.Start(ctx, events)
	// Give the watcher time to initialize.
	time.Sleep(200 * time.Millisecond)

	// Append a new JSONL line — should trigger a write event.
	f, err := os.OpenFile(jsonlPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(`{"type":"user","uuid":"u1","timestamp":"2026-03-30T09:21:18.876Z","sessionId":"session-001","cwd":"/dev/toph","gitBranch":"main","message":{"role":"user","content":"test"}}` + "\n")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	select {
	case e := <-events:
		if e.Type != data.EventUserMessage {
			t.Errorf("expected user event, got %s", e.Type)
		}
		if e.SessionID != "session-001" {
			t.Errorf("expected session-001, got %s", e.SessionID)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for event")
	}

	cancel()
	src.Stop()
}

func TestJSONLSource_Backfill(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-populate a JSONL file before starting the source.
	jsonlPath := filepath.Join(projectDir, "session-002.jsonl")
	line := `{"type":"user","uuid":"u1","timestamp":"2026-03-30T09:00:00.000Z","sessionId":"session-002","cwd":"/dev/toph","gitBranch":"main","message":{"role":"user","content":"pre-existing"}}` + "\n"
	if err := os.WriteFile(jsonlPath, []byte(line), 0644); err != nil {
		t.Fatal(err)
	}

	src := NewJSONLSource(tmpDir)
	events := make(chan data.Event, 100)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go src.Start(ctx, events)

	// Should receive the backfilled event.
	select {
	case e := <-events:
		if e.SessionID != "session-002" {
			t.Errorf("expected session-002, got %s", e.SessionID)
		}
		if e.Type != data.EventUserMessage {
			t.Errorf("expected user event, got %s", e.Type)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for backfill event")
	}

	cancel()
	src.Stop()
}

func TestJSONLSource_BackfillLimit(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write 250 lines — backfill should only emit the last 200.
	jsonlPath := filepath.Join(projectDir, "session-003.jsonl")
	f, err := os.Create(jsonlPath)
	if err != nil {
		t.Fatal(err)
	}
	for i := range 250 {
		_ = i
		f.WriteString(`{"type":"user","uuid":"u1","timestamp":"2026-03-30T09:00:00.000Z","sessionId":"session-003","cwd":"/dev/toph","gitBranch":"main","message":{"role":"user","content":"line"}}` + "\n")
	}
	f.Close()

	src := NewJSONLSource(tmpDir)
	events := make(chan data.Event, 300)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go src.Start(ctx, events)

	// Drain events with a short delay to let backfill complete.
	time.Sleep(500 * time.Millisecond)
	cancel()
	src.Stop()

	count := len(events)
	if count != 200 {
		t.Errorf("expected 200 backfilled events, got %d", count)
	}
}

func TestJSONLSource_ProjectFromPath(t *testing.T) {
	src := NewJSONLSource("/home/user/.claude/projects")

	tests := []struct {
		path    string
		want    string
	}{
		{"/home/user/.claude/projects/my-project/session.jsonl", "my-project"},
		{"/home/user/.claude/projects/another/sub/file.jsonl", "another"},
	}

	for _, tt := range tests {
		got := src.projectFromPath(tt.path)
		if got != tt.want {
			t.Errorf("projectFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestJSONLSource_NewDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	src := NewJSONLSource(tmpDir)
	events := make(chan data.Event, 100)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go src.Start(ctx, events)
	time.Sleep(200 * time.Millisecond)

	// Create a new project directory and JSONL file after the source started.
	projectDir := filepath.Join(tmpDir, "new-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Give watcher time to pick up the new directory.
	time.Sleep(200 * time.Millisecond)

	jsonlPath := filepath.Join(projectDir, "session-new.jsonl")
	if err := os.WriteFile(jsonlPath, []byte(`{"type":"user","uuid":"u1","timestamp":"2026-03-30T10:00:00.000Z","sessionId":"session-new","cwd":"/dev/toph","gitBranch":"main","message":{"role":"user","content":"hello"}}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	select {
	case e := <-events:
		if e.SessionID != "session-new" {
			t.Errorf("expected session-new, got %s", e.SessionID)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for event from new directory")
	}

	cancel()
	src.Stop()
}
