"use client";

import { useEffect, useRef, useState } from "react";

// =============================================================================
// Event data — realistic Claude Code tool calls
// =============================================================================
const EVENT_POOL = [
  { session: "api-server", tool: "Edit", file: "src/auth/oauth.ts", color: "var(--terminal-tool)" },
  { session: "api-server", tool: "Bash", file: "npm test -- --watch", color: "var(--terminal-active)" },
  { session: "api-server", tool: "Read", file: "src/middleware/cors.ts", color: "#71717a" },
  { session: "auth-flow", tool: "Edit", file: "src/session/store.ts", color: "var(--terminal-tool)" },
  { session: "auth-flow", tool: "Glob", file: "src/**/*.test.ts", color: "var(--terminal-tool)" },
  { session: "api-server", tool: "Write", file: "src/routes/callback.ts", color: "var(--terminal-active)" },
  { session: "tests", tool: "Bash", file: "go test ./internal/...", color: "var(--terminal-active)" },
  { session: "api-server", tool: "Grep", file: "handleToken src/", color: "var(--terminal-tool)" },
  { session: "deploy", tool: "Bash", file: "docker build -t api .", color: "var(--terminal-active)" },
  { session: "auth-flow", tool: "Read", file: "package.json", color: "#71717a" },
  { session: "tests", tool: "Edit", file: "internal/parser_test.go", color: "var(--terminal-tool)" },
  { session: "api-server", tool: "Agent", file: "Explore codebase structure", color: "var(--terminal-subagent)" },
  { session: "deploy", tool: "Read", file: ".github/workflows/ci.yml", color: "#71717a" },
  { session: "api-server", tool: "Edit", file: "src/auth/refresh.ts", color: "var(--terminal-tool)" },
  { session: "auth-flow", tool: "Bash", file: "npm run typecheck", color: "var(--terminal-active)" },
  { session: "tests", tool: "Bash", file: "go vet ./...", color: "var(--terminal-active)" },
  { session: "api-server", tool: "TaskCreate", file: "Wire OAuth callback", color: "var(--terminal-waiting)" },
  { session: "deploy", tool: "Edit", file: "Dockerfile", color: "var(--terminal-tool)" },
];

const SESSION_COLORS: Record<string, string> = {
  "api-server": "var(--terminal-active)",
  "auth-flow": "var(--terminal-waiting)",
  "tests": "var(--terminal-tool)",
  "deploy": "var(--terminal-subagent)",
};

const TOOL_ICONS: Record<string, string> = {
  Edit: "\u270e", Bash: "\u25b6", Read: "\u25c9", Write: "\u270e",
  Glob: "\u25ce", Grep: "\u25ce", Agent: "\u25c8", TaskCreate: "\u2726",
};

interface StreamEvent {
  id: number;
  time: string;
  session: string;
  tool: string;
  file: string;
  color: string;
}

// =============================================================================
// Shared event stream hook
// =============================================================================
function useEventStream() {
  const [events, setEvents] = useState<StreamEvent[]>([]);
  const indexRef = useRef(0);
  const idRef = useRef(0);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    if (mq.matches) {
      const staticEvents = EVENT_POOL.slice(0, 10).map((e, i) => ({
        ...e, id: i,
        time: `14:${String(23 + Math.floor(i / 3)).padStart(2, "0")}:${String((i * 7) % 60).padStart(2, "0")}`,
      }));
      setEvents(staticEvents);
      return;
    }

    const seed = EVENT_POOL.slice(0, 4).map((e, i) => ({
      ...e, id: idRef.current++,
      time: `14:23:${String(i * 12).padStart(2, "0")}`,
    }));
    setEvents(seed);
    indexRef.current = 4;

    const interval = setInterval(() => {
      const next = EVENT_POOL[indexRef.current % EVENT_POOL.length];
      indexRef.current++;
      const now = new Date();
      const time = `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}:${String(now.getSeconds()).padStart(2, "0")}`;
      setEvents((prev) => {
        const updated = [...prev, { ...next, id: idRef.current++, time }];
        return updated.length > 14 ? updated.slice(updated.length - 14) : updated;
      });
    }, 1800);

    return () => clearInterval(interval);
  }, []);

  return events;
}

// =============================================================================
// Mini terminal window wrapper
// =============================================================================
function MiniWindow({ title, children, className = "", panelId }: { title: string; children: React.ReactNode; className?: string; panelId?: string }) {
  return (
    <div className={`rounded-lg border border-zinc-800 bg-zinc-950 overflow-hidden ${className}`} data-panel={panelId}>
      <div className="flex items-center gap-1.5 border-b border-zinc-800 px-3 py-1.5">
        <div className="h-1.5 w-1.5 rounded-full bg-zinc-700" />
        <div className="h-1.5 w-1.5 rounded-full bg-zinc-700" />
        <div className="h-1.5 w-1.5 rounded-full bg-zinc-700" />
        <span className="ml-1.5 font-mono text-[10px] text-zinc-500 uppercase tracking-wider">{title}</span>
      </div>
      <div className="p-2.5 font-mono text-[11px]">{children}</div>
    </div>
  );
}

// =============================================================================
// Left pane: Raw JSONL stream
// =============================================================================
function RawStream({ events }: { events: StreamEvent[] }) {
  const scrollRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
  }, [events]);

  return (
    <div className="w-full lg:w-[38%] shrink-0 rounded-lg border border-zinc-800 bg-zinc-900/30 overflow-hidden" data-source="jsonl">
      <div className="flex items-center justify-between border-b border-zinc-800 px-4 py-2">
        <div className="flex items-center gap-2">
          <div className="h-2 w-2 rounded-full bg-green-500/70 animate-pulse" />
          <span className="font-mono text-[11px] text-zinc-500 uppercase tracking-wider">JSONL stream</span>
        </div>
        <span className="font-mono text-[11px] text-zinc-600">~/.claude/projects/</span>
      </div>
      <div
        ref={scrollRef}
        className="h-[420px] overflow-hidden px-3 py-2"
        style={{ mask: "linear-gradient(to bottom, black 85%, transparent 100%)", WebkitMask: "linear-gradient(to bottom, black 85%, transparent 100%)" }}
      >
        {events.map((event) => (
          <div key={event.id} className="py-1 font-mono text-[11px] text-zinc-500 leading-relaxed">
            <span className="text-zinc-700">{"{"}</span>
            <span className="text-zinc-600">&quot;tool&quot;</span>
            <span className="text-zinc-700">:</span>
            <span style={{ color: event.color }}>&quot;{event.tool}&quot;</span>
            <span className="text-zinc-700">,</span>
            <span className="text-zinc-600">&quot;file&quot;</span>
            <span className="text-zinc-700">:</span>
            <span className="text-zinc-400">&quot;{event.file}&quot;</span>
            <span className="text-zinc-700">{"}"}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

// =============================================================================
// Right pane: 5 toph panels as mini windows
// =============================================================================
function TophPanels({ events }: { events: StreamEvent[] }) {
  // Derive panel data from the shared event stream
  const recentEvents = events.slice(-6);
  const totalTokens = 14200 + events.length * 48;
  const cost = (totalTokens * 0.000015).toFixed(2);

  // Count tools across all events
  const toolCounts: Record<string, number> = {};
  for (const e of events) {
    toolCounts[e.tool] = (toolCounts[e.tool] || 0) + 1;
  }
  const sortedTools = Object.entries(toolCounts).sort((a, b) => b[1] - a[1]).slice(0, 4);
  const maxCount = sortedTools[0]?.[1] || 1;

  // Unique sessions seen
  const sessionSet = new Set(events.map((e) => e.session));

  return (
    <div className="flex-1 min-w-0 grid grid-cols-2 gap-2.5">
      {/* Sessions */}
      <MiniWindow title="Sessions" panelId="sessions">
        <div className="space-y-1.5">
          {["api-server", "auth-flow", "tests", "deploy"].map((s) => {
            const active = sessionSet.has(s);
            const icon = s === "auth-flow" ? "\u25d0" : active ? "\u25cf" : "\u25cb";
            return (
              <div key={s} className="flex items-center gap-1.5">
                <span style={{ color: active ? (SESSION_COLORS[s] || "#6C6C6C") : "#3f3f46" }}>{icon}</span>
                <span className={active ? "text-zinc-300" : "text-zinc-600"}>{s}</span>
              </div>
            );
          })}
        </div>
      </MiniWindow>

      {/* Activity */}
      <MiniWindow title="Activity" className="row-span-2" panelId="activity">
        <div className="space-y-1">
          {recentEvents.map((event) => (
            <div key={event.id} className="flex items-center gap-1.5 text-[10px]">
              <span style={{ color: event.color }}>{TOOL_ICONS[event.tool] || "\u25cb"}</span>
              <span className="text-zinc-400 truncate">{event.file}</span>
            </div>
          ))}
        </div>
      </MiniWindow>

      {/* Detail */}
      <MiniWindow title="Detail" panelId="detail">
        <div className="space-y-1 text-[10px]">
          <div><span className="text-zinc-600">session </span><span className="text-zinc-300">api-server</span></div>
          <div><span className="text-zinc-600">tokens </span><span className="text-zinc-300">{totalTokens.toLocaleString()}</span></div>
          <div><span className="text-zinc-600">cost </span><span className="text-zinc-300">${cost}</span></div>
          <div className="pt-0.5">
            <span className="text-zinc-600">context </span>
            <div className="inline-block w-16 h-1.5 bg-zinc-800 rounded-full overflow-hidden align-middle ml-1">
              <div className="h-full rounded-full bg-[var(--terminal-active)]" style={{ width: `${Math.min(74 + events.length * 0.3, 95)}%`, transition: "width 1s ease-out" }} />
            </div>
          </div>
        </div>
      </MiniWindow>

      {/* Metrics */}
      <MiniWindow title="Metrics" panelId="metrics">
        <div className="space-y-1 text-[10px]">
          <div><span className="text-zinc-600">tokens </span><span className="text-zinc-300">{totalTokens.toLocaleString()}</span></div>
          <div><span className="text-zinc-600">cost </span><span className="text-zinc-300">${cost}</span></div>
          <div><span className="text-zinc-600">burn </span><span className="text-zinc-300">~420/min</span></div>
          <div><span className="text-zinc-600">sessions </span><span className="text-zinc-300">{sessionSet.size} active</span></div>
        </div>
      </MiniWindow>

      {/* Tools */}
      <MiniWindow title="Tools" panelId="tools">
        <div className="space-y-1">
          {sortedTools.map(([tool, count]) => (
            <div key={tool} className="flex items-center gap-1.5 text-[10px]">
              <span className="w-10 text-right text-zinc-400">{tool}</span>
              <div className="flex-1 h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                <div
                  className="h-full rounded-full bg-[var(--terminal-tool)]"
                  style={{ width: `${(count / maxCount) * 100}%`, transition: "width 0.5s ease-out" }}
                />
              </div>
              <span className="text-zinc-500 w-4 text-right">{count}</span>
            </div>
          ))}
        </div>
      </MiniWindow>
    </div>
  );
}

// =============================================================================
// Exported Component
// =============================================================================
export function EventStream() {
  const events = useEventStream();

  return (
    <section className="relative py-24 border-t border-zinc-800/50 overflow-hidden">
      <div className="max-w-6xl mx-auto px-6 lg:px-16">
        <h2 className="font-heading text-3xl font-bold text-zinc-50 text-center mb-4">
          Live event capture
        </h2>
        <p className="text-center text-zinc-500 text-base max-w-2xl mx-auto mb-12">
          Every tool call your agents run is captured as JSONL. toph watches these
          logs and renders them across five panels in real-time.
        </p>
      </div>

      <div className="max-w-6xl mx-auto px-6 lg:px-16">
        <div className="relative">
          <div className="flex flex-col lg:flex-row gap-4 items-stretch">
            {/* Left: Raw JSONL */}
            <RawStream events={events} />

            {/* Arrow — desktop */}
            <div className="hidden lg:flex items-center justify-center shrink-0 w-[60px]">
              <div className="text-zinc-600 text-2xl">&rarr;</div>
            </div>

            {/* Arrow — mobile */}
            <div className="flex lg:hidden items-center justify-center py-1">
              <div className="text-zinc-600 text-2xl">&darr;</div>
            </div>

            {/* Right: 5 toph panels */}
            <TophPanels events={events} />
          </div>
        </div>

        {/* Labels */}
        <div className="flex flex-col lg:flex-row mt-3 gap-4">
          <div className="w-full lg:w-[38%] shrink-0 text-center">
            <span className="font-mono text-[11px] text-zinc-600">Raw JSONL from ~/.claude/</span>
          </div>
          <div className="hidden lg:block shrink-0 w-[50px]" />
          <div className="flex-1 text-center">
            <span className="font-mono text-[11px] text-zinc-600">What toph shows you</span>
          </div>
        </div>
      </div>
    </section>
  );
}
