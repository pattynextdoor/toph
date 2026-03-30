package data

import "time"

// EventType classifies normalized events flowing through the system.
type EventType int

const (
	EventToolUse EventType = iota
	EventToolResult
	EventAssistantText
	EventUserMessage
	EventSystemMessage
	EventProgress
	EventError
	EventSessionStart
	EventSessionEnd
	EventSubagentStart
	EventSubagentEnd
)

func (e EventType) String() string {
	switch e {
	case EventToolUse:
		return "tool_use"
	case EventToolResult:
		return "tool_result"
	case EventAssistantText:
		return "assistant"
	case EventUserMessage:
		return "user"
	case EventSystemMessage:
		return "system"
	case EventProgress:
		return "progress"
	case EventError:
		return "error"
	case EventSessionStart:
		return "session_start"
	case EventSessionEnd:
		return "session_end"
	case EventSubagentStart:
		return "subagent_start"
	case EventSubagentEnd:
		return "subagent_end"
	default:
		return "unknown"
	}
}

// Event is the normalized representation of any activity detected by a Source.
// All sources (JSONL, hooks, process) emit Events on a shared channel.
type Event struct {
	Type      EventType
	Timestamp time.Time
	SessionID string
	AgentID   string

	ToolName   string
	ToolInput  string
	Text       string
	Model      string
	StopReason string

	InputTokens              int
	OutputTokens             int
	CacheCreationInputTokens int
	CacheReadInputTokens     int
}
