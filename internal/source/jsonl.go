package source

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pattynextdoor/toph/internal/data"
)

// JSONLSource watches ~/.claude/projects/ (or a configurable base dir) for
// JSONL file changes via fsnotify and emits parsed events through a channel.
// It handles directory discovery, backfilling recent history on startup, and
// incremental reading of new lines as they're appended.
type JSONLSource struct {
	baseDir string
	watcher *fsnotify.Watcher
	mu      sync.Mutex
	offsets map[string]int64 // tracks read position per file
}

// NewJSONLSource creates a new JSONL source watching the given base directory.
// The baseDir should be the Claude Code projects directory
// (typically ~/.claude/projects/).
func NewJSONLSource(baseDir string) *JSONLSource {
	return &JSONLSource{
		baseDir: baseDir,
		offsets: make(map[string]int64),
	}
}

func (s *JSONLSource) Name() string { return "jsonl" }

// Start begins watching for JSONL changes. It blocks until ctx is cancelled
// or the watcher encounters an unrecoverable error.
func (s *JSONLSource) Start(ctx context.Context, events chan<- data.Event) error {
	var err error
	s.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	s.discoverAndWatch()
	s.backfill(events)

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-s.watcher.Events:
			if !ok {
				return nil
			}
			s.handleFSEvent(ev, events)
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return nil
			}
			slog.Debug("fsnotify error", "error", err)
		}
	}
}

// Stop closes the filesystem watcher.
func (s *JSONLSource) Stop() error {
	if s.watcher != nil {
		return s.watcher.Close()
	}
	return nil
}

// discoverAndWatch scans baseDir for existing project directories and sets up
// watchers on each one. It watches:
//   - baseDir itself (to detect new project directories)
//   - Each project directory (to detect new JSONL files)
//   - Each session's subagents/ directory (to detect subagent files)
func (s *JSONLSource) discoverAndWatch() {
	// Watch the base directory for new project dirs.
	if err := s.watcher.Add(s.baseDir); err != nil {
		slog.Debug("failed to watch base dir", "path", s.baseDir, "error", err)
		return
	}

	// Walk one level: project dirs under baseDir.
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		slog.Debug("failed to read base dir", "path", s.baseDir, "error", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projectPath := filepath.Join(s.baseDir, entry.Name())
		s.watchProjectDir(projectPath)
	}
}

// watchProjectDir adds a watch on a project directory and any subagent
// subdirectories within it.
func (s *JSONLSource) watchProjectDir(projectPath string) {
	if err := s.watcher.Add(projectPath); err != nil {
		slog.Debug("failed to watch project dir", "path", projectPath, "error", err)
		return
	}

	// Look for session subdirectories that contain subagents/ dirs.
	// Structure: {baseDir}/{project}/{sessionId}/subagents/
	sessionEntries, err := os.ReadDir(projectPath)
	if err != nil {
		return
	}
	for _, se := range sessionEntries {
		if !se.IsDir() {
			continue
		}
		subagentDir := filepath.Join(projectPath, se.Name(), "subagents")
		if info, err := os.Stat(subagentDir); err == nil && info.IsDir() {
			if err := s.watcher.Add(subagentDir); err != nil {
				slog.Debug("failed to watch subagent dir", "path", subagentDir, "error", err)
			}
		}
	}
}

// backfill reads the last ~200 lines from every existing .jsonl file and emits
// events. This provides immediate dashboard content on startup.
// backfillThreshold defines how recently a JSONL file must have been modified
// to warrant backfilling its contents. Older files just get their offset
// recorded so we can detect new writes, but we don't flood the dashboard with
// historical events.
const backfillThreshold = 10 * time.Minute

func (s *JSONLSource) backfill(events chan<- data.Event) {
	matches, err := filepath.Glob(filepath.Join(s.baseDir, "*", "*.jsonl"))
	if err != nil {
		slog.Debug("backfill glob failed", "error", err)
	}

	// Also pick up subagent JSONL files.
	subMatches, err := filepath.Glob(filepath.Join(s.baseDir, "*", "*", "subagents", "*.jsonl"))
	if err != nil {
		slog.Debug("backfill subagent glob failed", "error", err)
	}
	matches = append(matches, subMatches...)

	cutoff := time.Now().Add(-backfillThreshold)
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.ModTime().After(cutoff) {
			// Recently active file: backfill last 200 lines.
			s.readTail(path, 200, events)
		} else {
			// Old file: just record offset so we catch future writes.
			s.recordOffset(path, info.Size())
		}
	}
}

// readTail reads a JSONL file, takes the last n lines, parses them, and emits
// events. It also records the file's current size as the offset for subsequent
// incremental reads via readNewLines.
func (s *JSONLSource) readTail(path string, n int, events chan<- data.Event) {
	f, err := os.Open(path)
	if err != nil {
		slog.Debug("readTail: failed to open", "path", path, "error", err)
		return
	}
	defer f.Close()

	// Read all content to get last n lines and record file size.
	content, err := io.ReadAll(f)
	if err != nil {
		slog.Debug("readTail: failed to read", "path", path, "error", err)
		return
	}

	// Record the file offset so readNewLines starts from the end.
	s.mu.Lock()
	s.offsets[path] = int64(len(content))
	s.mu.Unlock()

	// Split into lines and take the last n.
	lines := splitLines(content)
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	project := s.projectFromPath(path)
	for _, line := range lines {
		for _, ev := range ParseLine(line, project) {
			events <- ev
		}
	}
}

// recordOffset records a file's current size as the read offset without
// emitting any events. Used for old files that we don't want to backfill.
func (s *JSONLSource) recordOffset(path string, size int64) {
	s.mu.Lock()
	s.offsets[path] = size
	s.mu.Unlock()
}

// readNewLines reads any bytes appended to a JSONL file since the last read,
// parses complete lines, and emits events.
func (s *JSONLSource) readNewLines(path string, events chan<- data.Event) {
	s.mu.Lock()
	offset := s.offsets[path]
	s.mu.Unlock()

	f, err := os.Open(path)
	if err != nil {
		slog.Debug("readNewLines: failed to open", "path", path, "error", err)
		return
	}
	defer f.Close()

	// Get current file size.
	info, err := f.Stat()
	if err != nil {
		slog.Debug("readNewLines: failed to stat", "path", path, "error", err)
		return
	}

	// If the file is smaller than our offset, it was truncated — reset.
	if info.Size() < offset {
		offset = 0
	}

	// Nothing new to read.
	if info.Size() == offset {
		return
	}

	// Seek to where we left off and read new bytes.
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		slog.Debug("readNewLines: seek failed", "path", path, "error", err)
		return
	}

	newBytes, err := io.ReadAll(f)
	if err != nil {
		slog.Debug("readNewLines: read failed", "path", path, "error", err)
		return
	}

	// Update offset to new file position.
	newOffset := offset + int64(len(newBytes))
	s.mu.Lock()
	s.offsets[path] = newOffset
	s.mu.Unlock()

	project := s.projectFromPath(path)
	for _, line := range splitLines(newBytes) {
		for _, ev := range ParseLine(line, project) {
			events <- ev
		}
	}
}

// handleFSEvent routes fsnotify events to the appropriate handler.
// CREATE on directories adds them to the watcher; WRITE on .jsonl files
// triggers incremental reading.
func (s *JSONLSource) handleFSEvent(ev fsnotify.Event, events chan<- data.Event) {
	// New directory created — could be a new project dir or subagents dir.
	if ev.Has(fsnotify.Create) {
		info, err := os.Stat(ev.Name)
		if err != nil {
			return
		}
		if info.IsDir() {
			// Determine depth relative to baseDir to decide watch strategy.
			rel, err := filepath.Rel(s.baseDir, ev.Name)
			if err != nil {
				return
			}
			depth := len(strings.Split(rel, string(filepath.Separator)))
			switch {
			case depth == 1:
				// New project directory — watch it for new JSONL files.
				s.watchProjectDir(ev.Name)
			default:
				// Could be a session dir or subagents dir — watch it.
				if err := s.watcher.Add(ev.Name); err != nil {
					slog.Debug("failed to watch new dir", "path", ev.Name, "error", err)
				}
			}
			return
		}

		// New JSONL file created — read any initial content.
		if strings.HasSuffix(ev.Name, ".jsonl") {
			s.readNewLines(ev.Name, events)
		}
		return
	}

	// Existing JSONL file was written to — read new lines.
	if ev.Has(fsnotify.Write) && strings.HasSuffix(ev.Name, ".jsonl") {
		s.readNewLines(ev.Name, events)
	}
}

// projectFromPath extracts the project name from a JSONL file path by taking
// the first path component relative to baseDir.
func (s *JSONLSource) projectFromPath(path string) string {
	rel, err := filepath.Rel(s.baseDir, path)
	if err != nil {
		return "unknown"
	}
	parts := strings.SplitN(rel, string(filepath.Separator), 2)
	if len(parts) == 0 {
		return "unknown"
	}
	return parts[0]
}

// splitLines splits content by newlines, filtering out empty lines.
func splitLines(content []byte) [][]byte {
	var lines [][]byte
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 {
			// Copy the line since scanner reuses its buffer.
			cp := make([]byte, len(line))
			copy(cp, line)
			lines = append(lines, cp)
		}
	}
	return lines
}
