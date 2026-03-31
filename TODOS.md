# toph — TODOS

## P1 — Must resolve before/during Phase 1

### Active session detection heuristic
**What:** Define how toph distinguishes active sessions from historical ones in `~/.claude/projects/`.
**Why:** The directory contains JSONL files for ALL past sessions across 15+ projects. Without a heuristic, toph shows hundreds of dead sessions.
**Candidate heuristics:**
- mtime-based: JSONL file modified within last N minutes (5? 10?)
- Process cross-reference: `ps aux | grep claude` to find PIDs, match to session CWD
- File lock check: Claude Code may hold a lock on active session files
- Hybrid: mtime as fast filter, process scan as confirmation
**Effort:** M | **Priority:** P1 | **Depends on:** JSONL format understanding

### Graceful shutdown with context.Context
**What:** Wire `context.Context` from `main()` through all goroutines. Cancel on SIGINT/SIGTERM.
**Why:** Multiple goroutines (fsnotify watchers, HTTP server, periodic scanner, backfill parser) need coordinated cleanup. Without context cancellation, goroutine leaks on exit.
**Effort:** S | **Priority:** P1 | **Depends on:** Nothing

### 30fps tick-based render throttling
**What:** Use `tea.Tick(33ms)` to batch events and render once per tick instead of per-event.
**Why:** During bursts of JSONL events (50+ tool calls), per-event rendering causes terminal flicker. Tick-based rendering collects all events since last tick and renders once.
**Pattern:** `tea.Tick(33ms)` fires `RenderTickMsg` → `Update()` flushes pending events → `View()` renders.
**Effort:** S | **Priority:** P1 | **Depends on:** Bubble Tea model structure

## P2 — Phase 2 enhancements

### Debug logging to file
**What:** `slog` output to `~/.config/toph/toph.log`. `--debug` flag enables verbose mode.
**Why:** toph is a TUI — can't log to stdout. Without file logging, debugging user-reported issues is impossible.
**Effort:** S | **Priority:** P2 | **Depends on:** Nothing

### Session auto-naming from git branch
**What:** Extract `gitBranch` from JSONL records to display `dossier/feat-oauth` instead of UUIDs.
**Why:** Instant readability boost. The data already exists in every JSONL record.
**Effort:** S | **Priority:** P2 | **Depends on:** JSONL parser

## Completed

- **Desktop notification on permission-wait** (P2) — `internal/notify/notify.go`
- **Inline sparklines per session** (P2) — `internal/ui/panels/sessions.go`
- **Activity feed color aging** (P2) — `internal/ui/panels/activity.go`
- **File conflict heatmap** (P2) — `internal/data/conflicts.go` + activity panel
- **toph export --json** (P3) — `internal/export/export.go`

