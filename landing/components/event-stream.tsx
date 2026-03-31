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
  Edit: "\u270e",
  Bash: "\u25b6",
  Read: "\u25c9",
  Write: "\u270e",
  Glob: "\u25ce",
  Grep: "\u25ce",
  Agent: "\u25c8",
  TaskCreate: "\u2726",
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
// Shared event hook — both panes use the same event stream
// =============================================================================
function useEventStream() {
  const [events, setEvents] = useState<StreamEvent[]>([]);
  const [reducedMotion, setReducedMotion] = useState(false);
  const indexRef = useRef(0);
  const idRef = useRef(0);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    if (mq.matches) {
      setReducedMotion(true);
      const staticEvents = EVENT_POOL.slice(0, 10).map((e, i) => ({
        ...e,
        id: i,
        time: `14:${String(23 + Math.floor(i / 3)).padStart(2, "0")}:${String((i * 7) % 60).padStart(2, "0")}`,
      }));
      setEvents(staticEvents);
    }
  }, []);

  useEffect(() => {
    if (reducedMotion) return;

    const seed = EVENT_POOL.slice(0, 4).map((e, i) => ({
      ...e,
      id: idRef.current++,
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
        if (updated.length > 12) return updated.slice(updated.length - 12);
        return updated;
      });
    }, 1800);

    return () => clearInterval(interval);
  }, [reducedMotion]);

  return events;
}

// =============================================================================
// Left pane: Raw JSONL event stream
// =============================================================================
function RawStream({ events }: { events: StreamEvent[] }) {
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [events]);

  return (
    <div className="flex-1 min-w-0 rounded-lg border border-zinc-800 bg-zinc-900/30 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-zinc-800 px-4 py-2">
        <div className="flex items-center gap-2">
          <div className="h-2 w-2 rounded-full bg-green-500/70 animate-pulse" />
          <span className="font-mono text-[11px] text-zinc-500 uppercase tracking-wider">
            JSONL stream
          </span>
        </div>
        <span className="font-mono text-[11px] text-zinc-600">
          ~/.claude/projects/
        </span>
      </div>

      {/* Events */}
      <div
        ref={scrollRef}
        className="h-[340px] overflow-hidden px-3 py-2"
        style={{
          mask: "linear-gradient(to bottom, black 85%, transparent 100%)",
          WebkitMask: "linear-gradient(to bottom, black 85%, transparent 100%)",
        }}
      >
        {events.map((event) => (
          <div
            key={event.id}
            className="py-1 font-mono text-[11px] text-zinc-500 leading-relaxed"
          >
            <span className="text-zinc-700">{"{"}</span>
            <span className="text-zinc-600">&quot;type&quot;</span>
            <span className="text-zinc-700">:</span>
            <span className="text-zinc-500">&quot;tool_use&quot;</span>
            <span className="text-zinc-700">,</span>
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
// Right pane: toph's rendered view of the same events
// =============================================================================
function TophView({ events }: { events: StreamEvent[] }) {
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [events]);

  return (
    <div className="flex-1 min-w-0 rounded-lg border border-zinc-800 bg-zinc-950 overflow-hidden">
      {/* Terminal chrome */}
      <div className="flex items-center gap-2 border-b border-zinc-800 px-4 py-2">
        <div className="h-2 w-2 rounded-full bg-red-500/70" />
        <div className="h-2 w-2 rounded-full bg-yellow-500/70" />
        <div className="h-2 w-2 rounded-full bg-green-500/70" />
        <span className="ml-2 font-mono text-[11px] text-zinc-500">toph</span>
        <div className="flex-1" />
        <span className="font-mono text-[11px] text-zinc-600 uppercase tracking-wider">
          Activity
        </span>
      </div>

      {/* toph-style event list */}
      <div
        ref={scrollRef}
        className="h-[340px] overflow-hidden px-3 py-2"
        style={{
          mask: "linear-gradient(to bottom, black 85%, transparent 100%)",
          WebkitMask: "linear-gradient(to bottom, black 85%, transparent 100%)",
        }}
      >
        {events.map((event) => (
          <div
            key={event.id}
            className="flex items-center gap-2.5 py-1.5 font-mono text-[12px]"
          >
            <span className="text-zinc-600 shrink-0">{event.time.slice(0, 5)}</span>
            <span className="shrink-0" style={{ color: event.color }}>
              {TOOL_ICONS[event.tool] || "\u25cb"}
            </span>
            <span className="shrink-0" style={{ color: event.color }}>
              {event.tool}
            </span>
            <span className="text-zinc-400 truncate">{event.file}</span>
            <span
              className="ml-auto shrink-0 text-[10px]"
              style={{ color: SESSION_COLORS[event.session] || "var(--terminal-idle)" }}
            >
              {event.session}
            </span>
          </div>
        ))}
      </div>
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
      <div className="max-w-5xl mx-auto px-6 lg:px-16">
        <h2 className="font-heading text-3xl font-bold text-zinc-50 text-center mb-4">
          Live event capture
        </h2>
        <p className="text-center text-zinc-500 text-base max-w-2xl mx-auto mb-12">
          Every tool call your agents run is captured as JSONL. toph watches these
          logs and renders them in real-time.
        </p>
      </div>

      {/* Side-by-side: raw stream → toph view */}
      <div className="max-w-5xl mx-auto px-6 lg:px-16">
        <div className="flex flex-col lg:flex-row gap-4 items-stretch">
          {/* Left: Raw JSONL */}
          <RawStream events={events} />

          {/* Arrow connector — desktop only */}
          <div className="hidden lg:flex items-center justify-center shrink-0 px-2">
            <div className="flex flex-col items-center gap-1">
              <span className="font-mono text-[10px] text-zinc-600 uppercase tracking-widest">toph</span>
              <span className="text-zinc-600 text-xl">&rarr;</span>
            </div>
          </div>

          {/* Mobile arrow */}
          <div className="flex lg:hidden items-center justify-center py-1">
            <div className="flex items-center gap-2">
              <span className="font-mono text-[10px] text-zinc-600 uppercase tracking-widest">toph</span>
              <span className="text-zinc-600 text-lg">&darr;</span>
            </div>
          </div>

          {/* Right: toph's view */}
          <TophView events={events} />
        </div>

        {/* Labels */}
        <div className="flex flex-col lg:flex-row mt-3 gap-4">
          <div className="flex-1 text-center">
            <span className="font-mono text-[11px] text-zinc-600">Raw JSONL data from ~/.claude/</span>
          </div>
          <div className="hidden lg:block shrink-0 w-[60px]" />
          <div className="flex-1 text-center">
            <span className="font-mono text-[11px] text-zinc-600">What toph shows you</span>
          </div>
        </div>
      </div>
    </section>
  );
}
