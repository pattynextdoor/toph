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
// Component
// =============================================================================
export function EventStream() {
  const [events, setEvents] = useState<StreamEvent[]>([]);
  const [reducedMotion, setReducedMotion] = useState(false);
  const indexRef = useRef(0);
  const idRef = useRef(0);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    if (mq.matches) {
      setReducedMotion(true);
      // Show static events
      const staticEvents = EVENT_POOL.slice(0, 12).map((e, i) => ({
        ...e,
        id: i,
        time: `14:${String(23 + Math.floor(i / 3)).padStart(2, "0")}:${String((i * 7) % 60).padStart(2, "0")}`,
      }));
      setEvents(staticEvents);
    }
  }, []);

  useEffect(() => {
    if (reducedMotion) return;

    // Seed with a few initial events
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
        if (updated.length > 14) return updated.slice(updated.length - 14);
        return updated;
      });
    }, 1800);

    return () => clearInterval(interval);
  }, [reducedMotion]);

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [events]);

  return (
    <section className="relative py-24 border-t border-zinc-800/50 overflow-hidden">
      <div className="max-w-5xl mx-auto px-6 lg:px-16">
        <h2 className="font-heading text-3xl font-bold text-zinc-50 text-center mb-4">
          Live event capture
        </h2>
        <p className="text-center text-zinc-500 text-base max-w-2xl mx-auto mb-12">
          Every tool call, file edit, and command your agents run — captured in real-time.
          This is the raw data toph watches.
        </p>
      </div>

      {/* Event stream display */}
      <div className="max-w-4xl mx-auto px-6 lg:px-16">
        <div
          className="relative rounded-lg border border-zinc-800 bg-zinc-900/30 overflow-hidden"
          style={{
            mask: "linear-gradient(to bottom, transparent 0%, black 5%, black 90%, transparent 100%)",
            WebkitMask: "linear-gradient(to bottom, transparent 0%, black 5%, black 90%, transparent 100%)",
          }}
        >
          {/* Header bar */}
          <div className="flex items-center justify-between border-b border-zinc-800 px-4 py-2">
            <div className="flex items-center gap-2">
              <div className="h-2 w-2 rounded-full bg-green-500/70 animate-pulse" />
              <span className="font-mono text-[11px] text-zinc-500 uppercase tracking-wider">
                Event stream
              </span>
            </div>
            <span className="font-mono text-[11px] text-zinc-600">
              {events.length} events captured
            </span>
          </div>

          {/* Scrollable event list */}
          <div
            ref={scrollRef}
            className="h-[380px] overflow-hidden px-4 py-3 space-y-0"
          >
            {events.map((event) => (
              <div
                key={event.id}
                className="flex items-center gap-3 py-1.5 font-mono text-[12px] animate-in fade-in-0 slide-in-from-bottom-1 duration-300"
              >
                {/* Timestamp */}
                <span className="text-zinc-600 shrink-0 w-[70px]">{event.time}</span>

                {/* Session badge */}
                <span
                  className="shrink-0 w-[90px] truncate text-[11px]"
                  style={{ color: SESSION_COLORS[event.session] || "var(--terminal-idle)" }}
                >
                  {event.session}
                </span>

                {/* Tool icon + name */}
                <span className="shrink-0 w-[80px] flex items-center gap-1.5" style={{ color: event.color }}>
                  <span>{TOOL_ICONS[event.tool] || "\u25cb"}</span>
                  <span>{event.tool}</span>
                </span>

                {/* File/argument */}
                <span className="text-zinc-400 truncate">{event.file}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
