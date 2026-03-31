package source

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/pattynextdoor/toph/internal/data"
)

const defaultHookPort = 7891

// HookPayload is the JSON structure POSTed by Claude Code hooks.
type HookPayload struct {
	Event      string          `json:"event"`
	SessionID  string          `json:"session_id"`
	CWD        string          `json:"cwd"`
	ToolName   string          `json:"tool_name,omitempty"`
	Input      json.RawMessage `json:"input,omitempty"`
	Output     json.RawMessage `json:"output,omitempty"`
	StopReason string          `json:"stop_reason,omitempty"`
	AgentID    string          `json:"agent_id,omitempty"`
}

// HookSource receives real-time events from Claude Code hooks via HTTP POST.
// It listens on 127.0.0.1 only (never 0.0.0.0) for security.
type HookSource struct {
	port     int
	server   *http.Server
	listener net.Listener
}

// NewHookSource creates a HookSource. Pass 0 for the default port (7891).
func NewHookSource(port int) *HookSource {
	if port == 0 {
		port = defaultHookPort
	}
	return &HookSource{port: port}
}

func (s *HookSource) Name() string { return "hooks" }

// Start binds the HTTP server and blocks until ctx is cancelled or an error occurs.
// If the default port is in use, it tries the next port once.
func (s *HookSource) Start(ctx context.Context, events chan<- data.Event) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/hook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		var payload HookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		for _, e := range hookToEvents(payload) {
			events <- e
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		// Port in use — try next port once
		s.port++
		addr = fmt.Sprintf("127.0.0.1:%d", s.port)
		s.listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("hooks: cannot bind to %s: %w", addr, err)
		}
	}

	slog.Debug("hooks server listening", "addr", addr)

	go func() {
		<-ctx.Done()
		s.server.Close()
	}()

	if err := s.server.Serve(s.listener); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop shuts down the HTTP server.
func (s *HookSource) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Port returns the actual port the server is (or will be) listening on.
func (s *HookSource) Port() int { return s.port }

// hookToEvents maps a Claude Code hook payload to normalized toph events.
func hookToEvents(p HookPayload) []data.Event {
	now := time.Now()
	base := data.Event{
		Timestamp: now,
		SessionID: p.SessionID,
		AgentID:   p.AgentID,
		CWD:       p.CWD,
	}

	switch p.Event {
	case "PreToolUse":
		e := base
		e.Type = data.EventToolUse
		e.ToolName = p.ToolName
		if len(p.Input) > 0 {
			e.ToolInput = summarizeToolInput(p.ToolName, p.Input)
		}
		return []data.Event{e}

	case "PostToolUse":
		e := base
		e.Type = data.EventToolResult
		e.ToolName = p.ToolName
		return []data.Event{e}

	case "Stop":
		e := base
		e.Type = data.EventSessionEnd
		e.StopReason = p.StopReason
		return []data.Event{e}

	case "SubagentStart":
		e := base
		e.Type = data.EventSubagentStart
		return []data.Event{e}

	case "SubagentStop":
		e := base
		e.Type = data.EventSubagentEnd
		return []data.Event{e}

	case "Notification":
		e := base
		e.Type = data.EventProgress
		return []data.Event{e}

	default:
		slog.Debug("hooks: unknown event type", "event", p.Event)
		return nil
	}
}
