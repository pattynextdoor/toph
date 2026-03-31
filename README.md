<p align="center">

<h3>toph</h3>

btop for AI agents. A terminal dashboard that monitors AI coding agent activity in real-time.

[![CI](https://img.shields.io/github/actions/workflow/status/pattynextdoor/toph/ci.yml?branch=main&style=flat-square&label=CI)](https://github.com/pattynextdoor/toph/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Built_with-Go-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)
[![Beta](https://img.shields.io/badge/Status-Beta-yellow?style=flat-square)](#status)

</p>

---

Your AI agents are working. Are you watching? See every tool call, token burn, and context fill across all your sessions -- one terminal, zero config, real-time.

<p align="center">
<img src="docs/demo.gif" width="800" alt="toph demo — live dashboard monitoring Claude Code sessions" />
</p>

## Install

```sh
brew install pattynextdoor/tap/toph
```

Or with Go:

```sh
go install github.com/pattynextdoor/toph/cmd/toph@latest
```

Or from source:

```sh
git clone https://github.com/pattynextdoor/toph.git
cd toph
go build -o toph ./cmd/toph
```

That's it. No config needed. On first run, toph discovers your Claude Code sessions from `~/.claude/projects/` automatically.

## Usage

**Watch your agents:**

```sh
toph                  # launch the dashboard
toph --debug          # enable debug logging to ~/.config/toph/toph.log
```

**Richer real-time data (optional):**

```sh
toph setup            # configure Claude Code hooks → POST to toph's local server
toph setup --remove   # undo hook configuration
```

**Remote access:**

```sh
toph serve            # start SSH server on port 2222
toph serve --port 3333
ssh -p 2222 localhost # view dashboard from another terminal
```

**Scripting:**

```sh
toph export           # dump current state as JSON
toph export | jq '.sessions[].cost_usd'
toph --version
```

## Keybindings

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Cycle panel focus |
| `j` / `k` / arrows | Navigate within panel |
| `1`-`5` | Jump to panel |
| `enter` | Drill into detail |
| `G` / `g` | Activity feed bottom / top |
| `/` | Filter sessions and events |
| `r` | Force refresh |
| `ctrl+l` | Clear activity feed |
| `?` | Help overlay |
| `q` / `ctrl+c` | Quit |

## What you see

| Panel | Shows |
|-------|-------|
| **Sessions** | Auto-detected agents with animated status, sparklines, git branch |
| **Detail** | Working directory, model, tokens, context fill meter, subagent tree |
| **Activity** | Real-time tool calls with file paths, color-coded and grouped |
| **Metrics** | Token burn rate, cache hit ratio, cost tracking, throughput chart |
| **Tools** | Frequency bar chart of tool usage across sessions |

## How it works

toph reads data from three sources:

| Source | Config needed | Latency | Data richness |
|--------|--------------|---------|---------------|
| JSONL logs | None | ~1s | Tool calls, tokens, cost |
| Hooks | `toph setup` | Real-time | All of the above + permissions, compaction |
| Process scan | None | ~5s | CPU, memory only |

Claude Code writes session logs to `~/.claude/projects/`. toph watches these with fsnotify -- no polling, no lag. For richer real-time data, `toph setup` configures Claude Code hooks to POST events to toph's local HTTP server on `127.0.0.1:7891`.

## Tech stack

- **Go** + **Bubble Tea** (TUI framework, Model-Update-View)
- **Lip Gloss** (CSS-like terminal styling)
- **Bubbles** (spinner, progress bar, viewport components)
- **Harmonica** (spring-based animations)
- **fsnotify** (file system watching)
- **Wish** (SSH server for remote dashboard access)

## Status

toph is in **beta**. The dashboard, JSONL watching, hook integration, SSH server, JSON export, file conflict detection, and desktop notifications are all working. Currently supports Claude Code only -- Aider and Codex support is planned.

## License

[MIT](LICENSE)
