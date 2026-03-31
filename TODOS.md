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

### Desktop notification on permission-wait
**What:** Fire macOS notification (`osascript`) when a session enters 'waiting for permission' state.
**Why:** #1 pain point for multi-session users — permission prompts get buried in other terminals.
**Effort:** S | **Priority:** P2 | **Depends on:** Session state machine

### Inline sparklines per session
**What:** 8-char braille sparkline next to each session showing token burn pattern over last 10 minutes.
**Why:** Turns session list from a status board into a heartbeat monitor. Screenshot gold.
**Requires:** Per-session token history ring buffer (60 samples at 10s intervals).
**Effort:** M | **Priority:** P2 | **Depends on:** Metrics calculation

### Activity feed color aging
**What:** Events fade from bright to dim as they age. 3s ago = bright cyan, 5m ago = dim gray.
**Why:** Eye naturally tracks what's fresh. Lip Gloss color interpolation handles this.
**Effort:** S | **Priority:** P2 | **Depends on:** Activity feed panel

### File conflict heatmap
**What:** Highlight files touched by 2+ agents simultaneously in red in the activity feed.
**Why:** Nobody else does this. Multi-agent is new and file conflicts are a real pain. Differentiating feature.
**Requires:** Track file paths per session, cross-reference within 5-min window.
**Effort:** M | **Priority:** P2 | **Depends on:** Multi-session event tracking

## P3 — Phase 3+ features

### toph export --json
**What:** Dump current dashboard state as JSON for scripting/piping to jq.
**Why:** Enables composability with Unix ecosystem. `toph export | jq .sessions[].cost` for CI/CD cost tracking.
**Effort:** S | **Priority:** P3 | **Depends on:** Stable data model
