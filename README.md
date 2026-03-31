# toph

btop for AI agents. A terminal dashboard that monitors AI coding agent activity in real-time.

Named after [Toph Beifong](https://avatar.fandom.com/wiki/Toph_Beifong) — she sees everything through vibrations. toph sees your agents through vibrations in the filesystem.

![toph demo](docs/demo.gif)

## What it does

toph watches your Claude Code sessions and shows you everything in a single terminal dashboard:

- **Sessions** — auto-detected running agents with status indicators
- **Activity feed** — real-time stream of tool calls, file edits, and commands
- **Detail** — token counts, cost tracking, context window fill meter
- **Metrics** — burn rate, session cost, daily totals
- **Tools** — frequency breakdown of Bash, Edit, Read, Glob, and more

Zero config. Just run `toph` and it finds your sessions automatically.

## Install

```bash
brew install pattynextdoor/tap/toph
```

## How it works

toph reads data from three sources:

1. **JSONL logs** — Claude Code writes session logs to `~/.claude/projects/`. toph watches these with fsnotify. Zero config.
2. **Hooks** — Claude Code's hook system provides 9 real-time event types for richer data. Opt-in via `toph setup`. See [Hook Events Reference](docs/hooks.md).
3. **Process scan** — Detects running `claude` processes for CPU/memory. Fallback when logs aren't available.

## Data sources

| Source | Config needed | Latency | Data richness |
|--------|--------------|---------|---------------|
| JSONL logs | None | ~1s | Tool calls, tokens, cost |
| Hooks | `toph setup` | Real-time | All of the above + permissions, prompts, compaction |
| Process scan | None | ~5s | CPU, memory only |

## Hook events

toph can consume all 9 Claude Code hook event types:

| Event | Fires when |
|-------|-----------|
| `PreToolUse` | Before any tool runs |
| `PostToolUse` | After any tool runs |
| `Stop` | Main agent stopping |
| `SubagentStop` | Subagent stopping |
| `SessionStart` | Session opens |
| `SessionEnd` | Session closes |
| `UserPromptSubmit` | User sends a prompt |
| `PreCompact` | Context window compaction |
| `Notification` | Agent notification |

Full details: [docs/hooks.md](docs/hooks.md)

## Tech stack

- **Go** + **Bubble Tea** (TUI framework)
- **Lip Gloss** (terminal styling)
- **fsnotify** (file watching)
- **Harmonica** (spring animations)

## License

MIT
