# toph

"btop for AI agents." A beautiful terminal dashboard that monitors AI coding agent activity in real-time.

Named after Toph Beifong from Avatar: The Last Airbender — she's blind but "sees" everything through earthbending vibrations. toph sees your AI agents through vibrations in the filesystem.

## Tech Stack

- **Go** + **Bubble Tea** (TUI framework, Model-Update-View)
- **Lip Gloss** (CSS-like terminal styling)
- **Harmonica** (spring-based animations)
- **Bubbles** (pre-built components: viewport, list, spinner, progress, table)
- **fsnotify** (file system watching)
- **Wish** (SSH server for remote dashboard access — Phase 4)

## Architecture

```
~/.claude/projects/     ─── fsnotify watcher ───┐
                                                 ├──→ Event Bus (Go channels) ──→ Bubble Tea Model ──→ View
Claude Code hooks ──→ HTTP server (localhost) ───┘
                                                      30fps tick-based rendering
```

### Source Interface (define from day 1)

```go
type Source interface {
    Name() string
    Start(ctx context.Context, events chan<- Event) error
    Stop() error
}
```

Three implementations for v0.1:
- `JSONLSource` — watches `~/.claude/projects/` JSONL files (zero-config, primary)
- `HookSource` — HTTP server on localhost for Claude Code hooks (opt-in, richer data)
- `ProcessSource` — `ps` scan for running claude processes (supplementary)

### Key Design Decisions

1. **Zero-config first run:** Just run `toph`. It finds `~/.claude/projects/` automatically.
2. **JSONL-first, hooks-optional:** File watching works without any Claude Code config. Hooks are a power-user enhancement.
3. **Read-only:** toph observes. It never creates, kills, or modifies sessions.
4. **Claude Code first:** No Aider/Codex/Cursor in v0.1. Source interface enables this later.
5. **Beautiful by default:** First screenshot should make someone want to install it.
6. **Respect the terminal:** Terminal default bg. Support NO_COLOR. Work in 256-color and truecolor.
7. **toph owns hooks:** Replaces acropora's hook config in `~/.claude/settings.json`. `toph setup` manages this.
8. **Tail + backfill on startup:** Read last ~200 JSONL lines instantly, backfill full history in background goroutine.
9. **30fps tick-based rendering:** `tea.Tick(33ms)` batches events, renders once per tick. No flicker during bursts.
10. **1,000 event ring buffer:** Activity feed capped at 1,000 events (~500KB). Old events drop off.

## Data Sources

### Primary: Claude Code JSONL Logs (zero-config)

Location: `~/.claude/projects/{project-hash}/{sessionId}.jsonl`
Subagents: `~/.claude/projects/{project-hash}/{sessionId}/subagents/agent-{agentId}.jsonl`
Subagent metadata: `agent-{agentId}.meta.json` (contains `agentType`)

**IMPORTANT — Actual JSONL record structure (verified from real logs):**

Record types: `file-history-snapshot`, `progress`, `system`, `user`, `assistant`, `last-prompt`, `queue-operation`

Common fields on most records:
```go
type BaseRecord struct {
    Type       string `json:"type"`       // "user", "assistant", "progress", "system", etc.
    SessionID  string `json:"sessionId"`
    CWD        string `json:"cwd"`
    Version    string `json:"version"`
    GitBranch  string `json:"gitBranch"`
    Timestamp  string `json:"timestamp"`
    UUID       string `json:"uuid"`
    ParentUUID string `json:"parentUuid"`
}
```

Assistant records (where token/usage data lives):
```go
type AssistantRecord struct {
    BaseRecord
    Message struct {
        Model   string          `json:"model"`   // "claude-opus-4-6", "claude-sonnet-4-6", etc.
        ID      string          `json:"id"`
        Type    string          `json:"type"`     // "message"
        Role    string          `json:"role"`     // "assistant"
        Content json.RawMessage `json:"content"`  // Array of content blocks
        Usage   struct {
            InputTokens              int    `json:"input_tokens"`
            OutputTokens             int    `json:"output_tokens"`
            CacheCreationInputTokens int    `json:"cache_creation_input_tokens"`
            CacheReadInputTokens     int    `json:"cache_read_input_tokens"`
            ServiceTier              string `json:"service_tier"`
        } `json:"usage"`
        StopReason string `json:"stop_reason"`
    } `json:"message"`
}
```

Tool use is in `Message.Content` array — blocks where `type == "tool_use"`:
```go
type ToolUseBlock struct {
    Type  string          `json:"type"`  // "tool_use"
    ID    string          `json:"id"`
    Name  string          `json:"name"`  // "Bash", "Edit", "Read", "Write", "Glob", "Grep", etc.
    Input json.RawMessage `json:"input"`
}
```

Progress records (hooks, agent progress, bash output):
```go
type ProgressRecord struct {
    BaseRecord
    Data struct {
        Type    string          `json:"type"`    // "hook_progress", "agent_progress", "bash_progress"
        Message json.RawMessage `json:"message"` // Nested structure varies by subtype
    } `json:"data"`
}
```

Subagent records also include:
```go
AgentID string `json:"agentId"`
Slug    string `json:"slug"`     // Human-readable name like "moonlit-wobbling-ember"
```

### Secondary: Claude Code Hooks (opt-in, richer real-time data)

toph runs an HTTP server on `127.0.0.1:7891` (MUST bind to localhost only, never 0.0.0.0).
`toph setup` configures Claude Code hooks to POST events to this server.
`toph setup --remove` reverses the configuration.

Hook events: SessionStart, SessionEnd, PreToolUse, PostToolUse, SubagentStart, SubagentStop, Stop, Notification, PermissionRequest, CwdChanged, FileChanged, and more.

### Tertiary: Process Scanning (supplementary)

`ps` to find running `claude` processes. Get PID, CPU%, memory. Works even without JSONL access. Helps confirm which sessions are truly active.

## Project Structure

```
toph/
├── cmd/
│   └── toph/
│       └── main.go           # Entry point, arg parsing, context setup
├── internal/
│   ├── config/
│   │   └── config.go         # CLI flags, future config file
│   ├── data/
│   │   ├── conflicts.go      # File conflict detection
│   │   ├── event.go          # Normalized event types
│   │   ├── metrics.go        # Token/cost calculations
│   │   ├── ringbuffer.go     # Activity feed ring buffer
│   │   └── session.go        # Session state model
│   ├── export/
│   │   └── export.go         # toph export --json command
│   ├── model/
│   │   ├── filter.go         # Session/event filtering
│   │   ├── keys.go           # Keybinding definitions
│   │   └── model.go          # Root Bubble Tea model (30fps tick loop)
│   ├── notify/
│   │   └── notify.go         # Desktop notifications (macOS)
│   ├── serve/
│   │   └── serve.go          # Wish SSH server for remote access
│   ├── setup/
│   │   └── setup.go          # toph setup hook configuration
│   ├── source/               # Source interface implementations
│   │   ├── source.go         # Source interface definition
│   │   ├── jsonl.go          # JSONLSource: fsnotify watcher + parser
│   │   ├── hooks.go          # HookSource: HTTP server on localhost
│   │   ├── parser.go         # JSONL record parser
│   │   └── process.go        # ProcessSource: ps-based detection
│   └── ui/
│       ├── panels/
│       │   ├── activity.go   # Panel 3: Activity feed (ring buffer)
│       │   ├── chart.go      # Braille chart rendering
│       │   ├── detail.go     # Panel 2: Session detail
│       │   ├── help.go       # Help overlay panel
│       │   ├── metrics.go    # Panel 4: Token/cost metrics
│       │   ├── sessions.go   # Panel 1: Session list
│       │   └── tools.go      # Panel 5: Tool usage chart
│       ├── animate.go        # Animation utilities
│       ├── layout.go         # Panel arrangement + responsive sizing
│       ├── statusbar.go      # Bottom status bar
│       └── theme.go          # Color palette, border styles
├── go.mod
├── go.sum
├── CLAUDE.md                  # This file
├── TODOS.md                   # Deferred work items
├── LICENSE                    # MIT
└── .goreleaser.yml            # Cross-platform releases
```

## Panel Layout

```
┌─────────────────────────┬──────────────────────────────┐
│ 1. SESSIONS             │ 3. ACTIVITY FEED             │
│                         │                              │
│ Session list with       │ Real-time event stream       │
│ status indicators       │ Color-coded by type          │
│ Sorted: waiting first   │ 1,000 event ring buffer      │
│                         │ Scrollable viewport          │
├─────────────────────────┤                              │
│ 2. DETAIL               │                              │
│                         ├──────────────────────────────┤
│ Selected session:       │ 4. METRICS                   │
│ - Current task/status   │                              │
│ - Subagent tree         │ Context fill ████████░░ 78%  │
│ - Context fill meter    │ Tokens: 142K / 200K          │
│ - Working directory     │ Cost today: $2.34            │
│ - Duration              │ Burn rate: ~420 tok/s        │
│                         │ Sessions: 3 active           │
├─────────────────────────┼──────────────────────────────┤
│ 5. TOOLS                │ STATUS BAR                   │
│ Bash ████████████ 34    │ [tab] panels  [/] filter     │
│ Edit ██████████   28    │ [q] quit  [?] help           │
│ Read ███████      19    │ Watching: ~/.claude/projects  │
│ Glob ████         12    │ Refresh: 30fps ● Connected   │
└─────────────────────────┴──────────────────────────────┘
```

### Panel 1 — Sessions (left, upper)
- Auto-detected from `~/.claude/projects/`
- Status icons: `●` active (green), `◐` waiting/permission (yellow, PULSING via Harmonica), `○` idle (dim), `✕` errored (red)
- Sorted by actionability: permission-waiting floats to top
- Each row: status icon + project name + git branch + session age + subagent count

### Panel 2 — Detail (left, lower)
- Selected session from Panel 1
- Current status description ("Writing to src/main.rs", "Running cargo test", "Thinking...")
- Subagent tree if subagents exist (indented, with status icons)
- Context fill meter (Harmonica spring easing on value change)
- Working directory, session duration, last tool used

### Panel 3 — Activity Feed (right, upper)
- Scrollable real-time events across ALL sessions
- Format: `[timestamp] [session] [event_type] description`
- Color coding: per-tool colors — Bash (amber), Read (blue), Edit/Write (green), Glob/Grep (magenta), Agent/Skill (lavender), errors (red)
- Per-tool Unicode glyphs: ▶ Bash, ◇ Read, ◆ Edit/Write, ⊙ Glob/Grep, ✦ Agent/Skill
- Consecutive same-tool events grouped with count badge (e.g., "▶ Bash ×3")
- Time gaps > 2 minutes shown as centered separator lines
- Auto-scrolls to bottom; user scroll-up disables auto-scroll, shows "↓ new" indicator
- 1,000 event ring buffer

### Panel 4 — Metrics (right, lower)
- Context window fill: progress bar + percentage + absolute numbers
- Token burn rate: tokens/sec averaged over last 30 seconds
- Estimated time to context limit based on burn rate
- Cost tracking: session + daily (model-aware pricing)
- Active session count

### Panel 5 — Tools (left, bottom)
- Horizontal bar chart of tool call frequency across all sessions
- Bars scale to terminal width, update in real-time

### Status Bar (bottom)
- Context-sensitive keybinding hints
- Data source indicator
- Refresh rate, connection status

## Visual Design

**Color palette:**
- Background: terminal default
- Panel borders: `#585858` (gray), focused: `#87AFFF` (soft blue)
- Active: `#87D787` green | Waiting: `#FFD787` amber (pulsing) | Error: `#FF8787` red | Idle: `#6C6C6C` dim
- Events: tool use `#87D7D7` cyan, file write `#87D787` green, subagent `#D7AFFF` lavender
- Progress bars: gradient `#87D787` → `#FFD787` → `#FF8787` as they fill
- Rounded corners (`╭╮╰╯`) via Lip Gloss

**Animations (Harmonica):**
- Context fill meter: spring easing on value change
- Permission-waiting sessions: pulsing amber glow
- Activity feed: smooth scroll

**Typography:**
- Panel titles: UPPERCASE, bold
- Sparklines: braille characters (`⡀⡄⡆⡇⣇⣧⣷⣿`)
- Progress bars: `█░` with gradient coloring

## Keybindings

```
Tab / Shift+Tab    Cycle panel focus
j / k / ↑ / ↓     Navigate within panel
Enter              Select session
/                  Filter
Esc                Clear filter
r                  Force refresh
1-5                Jump to panel
?                  Help overlay
q / Ctrl+C         Quit
G / g              Activity feed bottom/top
Ctrl+L             Clear feed
```

## Token/Cost Calculation

Extract from `assistant` records → `message.usage`:
- Input cost = `(input_tokens - cache_read_input_tokens) * input_price + cache_read_input_tokens * cache_read_price + cache_creation_input_tokens * cache_write_price`
- Output cost = `output_tokens * output_price`

Model pricing (per million tokens):
| Model | Input | Cache Read | Cache Write | Output |
|-------|-------|------------|-------------|--------|
| claude-opus-4-6 | $15 | $1.50 | $18.75 | $75 |
| claude-sonnet-4-6 | $3 | $0.30 | $3.75 | $15 |
| claude-haiku-4-5 | $0.80 | $0.08 | $1.00 | $4 |
| Unknown model | Default to Sonnet pricing, show "~" prefix on cost |

## Implementation Phases

### Phase 1: Core (Week 1)
1. `go mod init github.com/pattynextdoor/toph`
2. Source interface definition (`internal/source/source.go`)
3. JSONL parser with correct nested struct parsing
4. JSONLSource: fsnotify watcher + tail-and-stream + backfill goroutine
5. Session detection (see TODOS.md for heuristic decision)
6. Root Bubble Tea model with 30fps tick loop
7. Lip Gloss panel layout (responsive to terminal size)
8. Panel 1 (Sessions) + Panel 3 (Activity Feed with ring buffer)
9. Status bar + keybindings
10. Graceful shutdown via context.Context

### Phase 2: Detail + Metrics (Week 2)
1. Panel 2 (Detail): session info, subagent tree
2. Panel 4 (Metrics): token counts, context fill meter, burn rate
3. Panel 5 (Tools): tool frequency bar chart
4. Harmonica animations on context fill meter
5. Permission-waiting detection + pulsing indicator
6. Session sorting by actionability
7. Color theme system

### Phase 3: Polish + Hooks (Week 3)
1. HookSource: HTTP server on `127.0.0.1:7891`
2. `toph setup` command (dry-run preview, modifies settings.json, `--remove` to undo)
3. ProcessSource: ps-based session confirmation
4. Cost calculation (model-aware pricing)
5. Responsive layout (terminal resize handling, minimum size check)
6. Help overlay (`?` key)
7. `--debug` flag + slog file logging
8. README with screenshots
9. goreleaser config

### Phase 4: Distribution (Week 4)
1. `go install` support
2. Homebrew tap
3. GitHub Actions CI
4. Wish SSH server with SSH key auth only
5. Demo GIF for README

## Error Handling

- Truncated JSONL lines: skip, retry on next fsnotify event
- Missing `~/.claude/projects/`: show "No Claude Code data found" in status bar
- Missing usage fields: show "N/A" in metrics
- Unknown model: default to Sonnet pricing, prefix cost with "~"
- Port in use (hooks server): try next port, show in status bar
- Terminal too small: show "Terminal too small" message
- fsnotify dropped events: periodic 30s re-scan as safety net
- Unknown JSONL record types: skip silently, log at debug level

## Security

- HTTP hook server binds to `127.0.0.1` only — never `0.0.0.0`
- Wish SSH server: SSH key auth only (Ed25519 minimum), no password auth
- SSH host key stored at `~/.config/toph/host_key`, mode 0600
- toph never logs or persists JSONL content beyond in-memory activity feed
- JSONL files may contain user prompts/code — toph displays but never transmits

## Reference Tools

Study for interaction patterns:
- **btop** — Panel layout, graph density, color theming, 30fps refresh
- **lazygit** — Panel focus, vim-style navigation, left-list/right-detail
- **yazi** — Speed, minimal chrome, keyboard-first
- **k9s** — Resource monitoring, status indicators, namespace filtering
