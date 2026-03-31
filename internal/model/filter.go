package model

import (
	"strings"

	"github.com/pattynextdoor/toph/internal/data"
)

// matchesFilter returns true if s contains the filter substring
// (case-insensitive).
func matchesFilter(s string, filter string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(filter))
}

// filterSessions returns only sessions whose Project, GitBranch, or ID
// matches the filter text. Returns the original slice when filter is empty.
func filterSessions(sessions []*data.Session, filter string) []*data.Session {
	if filter == "" {
		return sessions
	}
	var result []*data.Session
	for _, s := range sessions {
		if matchesFilter(s.Project, filter) || matchesFilter(s.GitBranch, filter) || matchesFilter(s.ID, filter) {
			result = append(result, s)
		}
	}
	return result
}

// filterEvents returns only events whose ToolName, ToolInput, Text, or
// SessionID matches the filter text. Returns the original slice when filter
// is empty.
func filterEvents(events []data.Event, filter string) []data.Event {
	if filter == "" {
		return events
	}
	var result []data.Event
	for _, e := range events {
		if matchesFilter(e.ToolName, filter) || matchesFilter(e.ToolInput, filter) || matchesFilter(e.Text, filter) || matchesFilter(e.SessionID, filter) {
			result = append(result, e)
		}
	}
	return result
}
