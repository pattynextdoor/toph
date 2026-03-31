# Claude Code Hook Events

toph can consume all 9 Claude Code hook event types for real-time monitoring. This document describes each event, when it fires, what data it provides, and how toph uses it.

## Setup

Run `toph setup` to configure Claude Code to send hook events to toph's HTTP server. This modifies `~/.claude/settings.json` to register toph as a hook handler.

```bash
toph setup          # preview + apply hook config
toph setup --remove # remove toph hooks
```

## Common Payload

Every hook event delivers this base payload via `stdin` as JSON:

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | Unique session identifier |
| `transcript_path` | string | Path to the session transcript file |
| `cwd` | string | Current working directory |
| `permission_mode` | string | `"ask"` or `"allow"` |
| `hook_event_name` | string | Name of the firing event |

## Events

### PreToolUse

Fires immediately **before** any tool is invoked.

| Extra Field | Type | Description |
|-------------|------|-------------|
| `tool_name` | string | Tool about to run (`"Bash"`, `"Edit"`, `"Read"`, etc.) |
| `tool_input` | object | Full input arguments to the tool |

**What toph does with it:** Updates the Activity Feed in real-time before the tool finishes. Shows "running..." status for long-running tools like Bash.

---

### PostToolUse

Fires immediately **after** a tool has finished executing.

| Extra Field | Type | Description |
|-------------|------|-------------|
| `tool_name` | string | Tool that ran |
| `tool_input` | object | Input that was passed |
| `tool_result` | string/object | Result/output returned by the tool |

**What toph does with it:** Completes the Activity Feed entry with result status. Updates tool frequency counts in the Tools panel. Tracks file modifications for the file heatmap.

---

### Stop

Fires when the **main agent** determines it has completed its task and is about to stop.

| Extra Field | Type | Description |
|-------------|------|-------------|
| `reason` | string | Why the agent believes it is done |

**What toph does with it:** Updates session status from "active" to "idle". Records session duration. Logs the stop reason in the Activity Feed.

---

### SubagentStop

Fires when a **subagent** (not the main agent) completes its portion of the task.

| Extra Field | Type | Description |
|-------------|------|-------------|
| `reason` | string | Subagent's stated reason for stopping |

**What toph does with it:** Updates the subagent tree in the Detail panel. Marks the subagent as completed. Useful for tracking multi-agent workflow progress.

---

### SessionStart

Fires at the beginning of a new Claude Code session.

| Extra Field | Type | Description |
|-------------|------|-------------|
| *(none)* | — | Only common fields |

**What toph does with it:** Adds a new session to the Sessions panel. Auto-detects the project directory and git branch. Starts the session timer.

---

### SessionEnd

Fires when a Claude Code session is ending.

| Extra Field | Type | Description |
|-------------|------|-------------|
| *(none)* | — | Only common fields |

**What toph does with it:** Marks the session as ended in the Sessions panel (dimmed). Records final token count and cost. Stops the session timer.

---

### UserPromptSubmit

Fires when a user submits a prompt to Claude (before Claude processes it).

| Extra Field | Type | Description |
|-------------|------|-------------|
| `user_prompt` | string | The text of the submitted prompt |

**What toph does with it:** Logs user prompts in the Activity Feed as context for the tool calls that follow. Helps understand what the agent is working on.

---

### PreCompact

Fires before Claude compacts/summarizes the conversation context (when the context window fills up).

| Extra Field | Type | Description |
|-------------|------|-------------|
| *(none)* | — | Only common fields |

**What toph does with it:** Updates the context fill meter — a compaction event means the context window was full. Resets the context percentage estimate after compaction completes.

---

### Notification

Fires when Claude Code emits a notification to the user.

| Extra Field | Type | Description |
|-------------|------|-------------|
| `message` | string | The notification content |

**What toph does with it:** Displays notifications in the Activity Feed. Permission-waiting notifications cause the session to sort to the top of the Sessions panel with a pulsing amber indicator.

## Summary

| Event | Fires When | Can Block? | toph Panel |
|-------|-----------|------------|------------|
| `PreToolUse` | Before tool runs | Yes | Activity, Tools |
| `PostToolUse` | After tool runs | No | Activity, Tools, Detail |
| `Stop` | Main agent stopping | Yes | Sessions |
| `SubagentStop` | Subagent stopping | Yes | Detail (subagent tree) |
| `SessionStart` | Session opens | No | Sessions |
| `SessionEnd` | Session closes | No | Sessions, Metrics |
| `UserPromptSubmit` | User sends prompt | No | Activity |
| `PreCompact` | Context compaction | No | Detail (context meter) |
| `Notification` | Agent notification | No | Activity, Sessions |

## Data Source Priority

toph uses multiple data sources in priority order:

1. **Hooks** (richest, real-time) — requires `toph setup`
2. **JSONL logs** (zero-config, slight delay) — works out of the box
3. **Process scan** (fallback) — CPU/memory only

Hooks provide the most detailed and timely data. JSONL log watching is the default zero-config experience. Both can run simultaneously — toph deduplicates events from overlapping sources.
