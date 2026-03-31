package source

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pattynextdoor/toph/internal/data"
)

// ProcessInfo holds resource usage for a detected claude process.
type ProcessInfo struct {
	PID    int
	CPU    float64
	Memory float64
}

// ProcessSource scans for running claude processes via `ps aux` on a timer.
// It emits session start/end events when processes appear or disappear.
type ProcessSource struct {
	interval time.Duration
	procs    map[int]ProcessInfo
}

// NewProcessSource creates a ProcessSource that scans every interval.
func NewProcessSource(interval time.Duration) *ProcessSource {
	return &ProcessSource{
		interval: interval,
		procs:    make(map[int]ProcessInfo),
	}
}

func (s *ProcessSource) Name() string { return "process" }

// Start runs an initial scan, then re-scans on each tick until ctx is cancelled.
func (s *ProcessSource) Start(ctx context.Context, events chan<- data.Event) error {
	s.scan(events)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.scan(events)
		}
	}
}

func (s *ProcessSource) Stop() error { return nil }

// scan runs `ps aux`, filters for claude processes, and emits events for
// newly appeared or disappeared PIDs compared to the previous scan.
func (s *ProcessSource) scan(events chan<- data.Event) {
	out, err := exec.Command("ps", "aux").Output()
	if err != nil {
		return
	}

	current := make(map[int]ProcessInfo)
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}
		cmd := strings.Join(fields[10:], " ")
		if !strings.Contains(cmd, "claude") || strings.Contains(cmd, "toph") {
			continue
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)

		current[pid] = ProcessInfo{PID: pid, CPU: cpu, Memory: mem}
	}

	now := time.Now()

	// Detect new processes
	for pid := range current {
		if _, existed := s.procs[pid]; !existed {
			events <- data.Event{
				Type:      data.EventSessionStart,
				Timestamp: now,
				SessionID: fmt.Sprintf("pid-%d", pid),
				Text:      fmt.Sprintf("claude process started (PID %d)", pid),
			}
		}
	}

	// Detect gone processes
	for pid := range s.procs {
		if _, exists := current[pid]; !exists {
			events <- data.Event{
				Type:      data.EventSessionEnd,
				Timestamp: now,
				SessionID: fmt.Sprintf("pid-%d", pid),
				Text:      fmt.Sprintf("claude process ended (PID %d)", pid),
			}
		}
	}

	s.procs = current
}
