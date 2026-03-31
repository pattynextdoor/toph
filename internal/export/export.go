package export

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pattynextdoor/toph/internal/data"
)

// Snapshot represents the full dashboard state at a point in time.
type Snapshot struct {
	Timestamp  string         `json:"timestamp"`
	Sessions   []Session      `json:"sessions"`
	ToolCounts map[string]int `json:"tool_counts"`
	TotalCost  float64        `json:"total_cost_usd"`
}

// Session is the JSON-serializable representation of a monitored session.
type Session struct {
	ID            string  `json:"id"`
	Project       string  `json:"project"`
	CWD           string  `json:"cwd"`
	GitBranch     string  `json:"git_branch"`
	Model         string  `json:"model"`
	Status        string  `json:"status"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	CacheRead     int     `json:"cache_read_tokens"`
	CacheWrite    int     `json:"cache_write_tokens"`
	Cost          float64 `json:"cost_usd"`
	LastTool      string  `json:"last_tool"`
	SubagentCount int     `json:"subagent_count"`
}

// modelPricing holds per-million-token prices for a model.
type modelPricing struct {
	Input, CacheRead, CacheWrite, Output float64
}

var pricing = map[string]modelPricing{
	"claude-opus-4-6":   {15.0, 1.50, 18.75, 75.0},
	"claude-sonnet-4-6": {3.0, 0.30, 3.75, 15.0},
	"claude-haiku-4-5":  {0.80, 0.08, 1.00, 4.0},
}

var defaultPricing = modelPricing{3.0, 0.30, 3.75, 15.0}

// Run aggregates session data from the manager and prints a JSON snapshot to stdout.
func Run(manager *data.Manager) error {
	sessions := manager.Sessions()
	toolCounts := manager.ToolCounts()

	var exported []Session
	var totalCost float64

	for _, s := range sessions {
		s.RLock()
		cost := estimateCost(s.Model, s.TotalInputTokens, s.TotalOutputTokens, s.TotalCacheRead, s.TotalCacheWrite)
		totalCost += cost

		exported = append(exported, Session{
			ID:            s.ID,
			Project:       s.Project,
			CWD:           s.CWD,
			GitBranch:     s.GitBranch,
			Model:         s.Model,
			Status:        s.Status.String(),
			InputTokens:   s.TotalInputTokens,
			OutputTokens:  s.TotalOutputTokens,
			CacheRead:     s.TotalCacheRead,
			CacheWrite:    s.TotalCacheWrite,
			Cost:          cost,
			LastTool:      s.LastToolName,
			SubagentCount: len(s.Subagents),
		})
		s.RUnlock()
	}

	snapshot := Snapshot{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Sessions:   exported,
		ToolCounts: toolCounts,
		TotalCost:  totalCost,
	}

	out, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func estimateCost(model string, input, output, cacheRead, cacheWrite int) float64 {
	p, ok := pricing[model]
	if !ok {
		p = defaultPricing
	}
	nonCached := input - cacheRead
	if nonCached < 0 {
		nonCached = 0
	}
	cost := float64(nonCached) / 1_000_000 * p.Input
	cost += float64(cacheRead) / 1_000_000 * p.CacheRead
	cost += float64(cacheWrite) / 1_000_000 * p.CacheWrite
	cost += float64(output) / 1_000_000 * p.Output
	return cost
}
