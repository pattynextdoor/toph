package source

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/pattynextdoor/toph/internal/data"
)

// ParseLine parses a single JSONL line from a Claude Code log and returns
// zero or more normalized Events. Unknown record types and invalid JSON
// are silently skipped (returns nil).
func ParseLine(line []byte, project string) []data.Event {
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return nil
	}

	// First pass: extract just the record type to decide how to decode.
	var header struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &header); err != nil {
		return nil
	}

	switch header.Type {
	case "user":
		return parseUser(line, project)
	case "assistant":
		return parseAssistant(line, project)
	case "system":
		return parseSystem(line, project)
	default:
		// file-history-snapshot, progress, queue-operation, last-prompt, etc.
		return nil
	}
}

// ParseBytes parses multiple newline-delimited JSONL records, returning all
// events across every line.
func ParseBytes(b []byte, project string) []data.Event {
	var events []data.Event
	for _, line := range bytes.Split(b, []byte("\n")) {
		events = append(events, ParseLine(line, project)...)
	}
	return events
}

// ---------------------------------------------------------------------------
// Internal record types matching Claude Code's JSONL structure
// ---------------------------------------------------------------------------

type baseRecord struct {
	Type       string `json:"type"`
	SessionID  string `json:"sessionId"`
	CWD        string `json:"cwd"`
	Version    string `json:"version"`
	GitBranch  string `json:"gitBranch"`
	Timestamp  string `json:"timestamp"`
	UUID       string `json:"uuid"`
	ParentUUID string `json:"parentUuid"`
	AgentID    string `json:"agentId"`
}

type userRecord struct {
	baseRecord
	Message struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type assistantRecord struct {
	baseRecord
	Message struct {
		Model      string          `json:"model"`
		ID         string          `json:"id"`
		Type       string          `json:"type"`
		Role       string          `json:"role"`
		Content    json.RawMessage `json:"content"`
		StopReason string          `json:"stop_reason"`
		Usage      struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

type contentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ---------------------------------------------------------------------------
// Per-type parsers
// ---------------------------------------------------------------------------

func parseUser(line []byte, project string) []data.Event {
	var rec userRecord
	if err := json.Unmarshal(line, &rec); err != nil {
		return nil
	}

	text := extractUserText(rec.Message.Content)
	if len(text) > 100 {
		text = text[:100]
	}

	return []data.Event{{
		Type:      data.EventUserMessage,
		Timestamp: parseTimestamp(rec.Timestamp),
		SessionID: rec.SessionID,
		AgentID:   rec.AgentID,
		Project:   project,
		CWD:       rec.CWD,
		GitBranch: rec.GitBranch,
		Text:      text,
	}}
}

func parseAssistant(line []byte, project string) []data.Event {
	var rec assistantRecord
	if err := json.Unmarshal(line, &rec); err != nil {
		return nil
	}

	ts := parseTimestamp(rec.Timestamp)
	var events []data.Event

	// Parse content blocks for tool_use events.
	var blocks []contentBlock
	if err := json.Unmarshal(rec.Message.Content, &blocks); err == nil {
		for _, b := range blocks {
			if b.Type == "tool_use" {
				events = append(events, data.Event{
					Type:      data.EventToolUse,
					Timestamp: ts,
					SessionID: rec.SessionID,
					AgentID:   rec.AgentID,
					Project:   project,
					CWD:       rec.CWD,
					GitBranch: rec.GitBranch,
					ToolName:  b.Name,
					ToolInput: summarizeToolInput(b.Name, b.Input),
				})
			}
		}
	}

	// Always emit an assistant text event with usage data.
	events = append(events, data.Event{
		Type:                     data.EventAssistantText,
		Timestamp:                ts,
		SessionID:                rec.SessionID,
		AgentID:                  rec.AgentID,
		Project:                  project,
		CWD:                      rec.CWD,
		GitBranch:                rec.GitBranch,
		Model:                    rec.Message.Model,
		StopReason:               rec.Message.StopReason,
		InputTokens:              rec.Message.Usage.InputTokens,
		OutputTokens:             rec.Message.Usage.OutputTokens,
		CacheCreationInputTokens: rec.Message.Usage.CacheCreationInputTokens,
		CacheReadInputTokens:     rec.Message.Usage.CacheReadInputTokens,
		Text:                     extractAssistantText(rec.Message.Content),
	})

	return events
}

func parseSystem(line []byte, project string) []data.Event {
	var rec baseRecord
	if err := json.Unmarshal(line, &rec); err != nil {
		return nil
	}
	return []data.Event{{
		Type:      data.EventSystemMessage,
		Timestamp: parseTimestamp(rec.Timestamp),
		SessionID: rec.SessionID,
		AgentID:   rec.AgentID,
		Project:   project,
		CWD:       rec.CWD,
		GitBranch: rec.GitBranch,
	}}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// parseTimestamp tries RFC3339Nano, then millisecond format, then falls back
// to time.Now().
func parseTimestamp(s string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02T15:04:05.000Z", s); err == nil {
		return t
	}
	return time.Now()
}

// extractUserText pulls a plain string from the user message content field.
// Content may be a JSON string or an array of content blocks.
func extractUserText(raw json.RawMessage) string {
	// Try plain string first.
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	// Try array of blocks (less common for user messages but possible).
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				return b.Text
			}
		}
	}
	return ""
}

// extractAssistantText concatenates text blocks from an assistant content array.
func extractAssistantText(raw json.RawMessage) string {
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}
	var text string
	for _, b := range blocks {
		if b.Type == "text" {
			if text != "" {
				text += "\n"
			}
			text += b.Text
		}
	}
	return text
}

// summarizeToolInput extracts the most relevant field from a tool's input
// JSON and truncates it to 60 characters.
func summarizeToolInput(toolName string, raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}

	// Pick the most informative field per tool.
	key := ""
	switch toolName {
	case "Bash":
		key = "command"
	case "Read", "Write", "Edit":
		key = "file_path"
	case "Glob":
		key = "pattern"
	case "Grep":
		key = "pattern"
	default:
		// For unknown tools, try command, then file_path, then first key.
		for _, k := range []string{"command", "file_path", "pattern"} {
			if _, ok := m[k]; ok {
				key = k
				break
			}
		}
		if key == "" {
			// Fall back to first key alphabetically.
			for k := range m {
				key = k
				break
			}
		}
	}

	if key == "" {
		return ""
	}

	val, ok := m[key]
	if !ok {
		return ""
	}

	var s string
	if err := json.Unmarshal(val, &s); err != nil {
		return string(val)
	}

	if len(s) > 60 {
		s = s[:60]
	}
	return s
}
