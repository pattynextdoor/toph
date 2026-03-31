"use client";

import { useEffect, useRef, useState } from "react";

// ---------------------------------------------------------------------------
// Data — matches the real toph dashboard layout
// ---------------------------------------------------------------------------

const sessions = [
  { name: "tp", branch: "main", age: "2m", active: true },
  { name: "toph", branch: "main", age: "3m", active: true },
  { name: "acropora", branch: "main", age: "6m", active: true },
  { name: "toph", branch: "main", age: "19h8m", active: true },
] as const;

// Braille sparkline patterns per session (visual only)
const sparklines = ["⣀⣤⣶⣿⣷⣤⡄", "⡀⣀⣤⣴⣶⣤⣀", "⣤⣶⣿⣿⣶⣤⣀", "⡀⡀⣀⣤⣶⣿⣷"] as const;

type ToolName = "Read" | "Edit" | "Bash" | "Glob" | "Skill" | "Agent";

const toolGlyphs: Record<ToolName, string> = {
  Read: "◇",
  Edit: "◆",
  Bash: "▶",
  Glob: "⊙",
  Skill: "✦",
  Agent: "✦",
};

const toolColorVars: Record<ToolName, string> = {
  Read: "var(--terminal-read)",
  Edit: "var(--terminal-edit)",
  Bash: "var(--terminal-bash)",
  Glob: "var(--terminal-search)",
  Skill: "var(--terminal-agent)",
  Agent: "var(--terminal-agent)",
};

interface ActivityEvent {
  time: string;
  session: string;
  tool: ToolName;
  detail: string;
  count?: number;
}

const allEvents: ActivityEvent[] = [
  { time: "04:42:55", session: "089eb8", tool: "Read", detail: "toph/TODOS.md" },
  { time: "04:43:19", session: "089eb8", tool: "Edit", detail: "toph/TODOS.md", count: 2 },
  { time: "04:43:22", session: "089eb8", tool: "Read", detail: "toph/TODOS.md" },
  { time: "04:43:29", session: "089eb8", tool: "Edit", detail: "toph/TODOS.md" },
  { time: "04:43:32", session: "089eb8", tool: "Read", detail: "toph/TODOS.md" },
  { time: "04:43:37", session: "089eb8", tool: "Bash", detail: "git add TODOS.md && git commit" },
  { time: "04:43:17", session: "089eb8", tool: "Read", detail: "toph/CLAUDE.md", count: 2 },
  { time: "04:43:42", session: "089eb8", tool: "Edit", detail: "toph/CLAUDE.md", count: 2 },
  { time: "04:43:46", session: "089eb8", tool: "Bash", detail: "git add CLAUDE.md && git commit" },
  { time: "04:44:18", session: "089eb8", tool: "Read", detail: "data/*.go", count: 3 },
  { time: "04:44:28", session: "089eb8", tool: "Edit", detail: "data/ringbuffer_test.go" },
  { time: "04:44:31", session: "089eb8", tool: "Bash", detail: "go test ./internal/data/ -run TestRingBufferClear" },
  { time: "04:44:41", session: "089eb8", tool: "Edit", detail: "ringbuffer.go", count: 2 },
  { time: "04:44:52", session: "089eb8", tool: "Bash", detail: "go build ./...", count: 3 },
  { time: "04:45:21", session: "089eb8", tool: "Read", detail: "model/model.go" },
  { time: "04:45:30", session: "089eb8", tool: "Edit", detail: "model/model.go" },
  { time: "04:45:38", session: "089eb8", tool: "Bash", detail: "go build ./...", count: 2 },
];

const pooledEvents = allEvents;

const toolBarData = [
  { name: "Read", count: 41, pct: 43 },
  { name: "Bash", count: 19, pct: 20 },
  { name: "Edit", count: 18, pct: 19 },
  { name: "Skill", count: 5, pct: 5 },
] as const;

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

/** Orbiting dot indicator for active sessions */
function OrbitDot({ color }: { color: string }) {
  return (
    <span className="relative inline-flex items-center justify-center w-3 h-3 shrink-0">
      <span
        className="absolute w-1.5 h-1.5 rounded-full"
        style={{ backgroundColor: color, opacity: 0.3 }}
      />
      <span
        className="animate-orbit absolute w-1 h-1 rounded-full"
        style={{ backgroundColor: color }}
      />
    </span>
  );
}

/** Time-gap separator line with diamond */
function GapSeparator() {
  return (
    <div className="flex items-center justify-center text-zinc-600 select-none py-0.5">
      <span className="text-[10px]">{"────────────── ◆ ──────────────"}</span>
    </div>
  );
}

/** Single tool bar row for the Tools panel */
function ToolBarRow({ name, count, pct }: { name: string; count: number; pct: number }) {
  const colorVar = toolColorVars[name as ToolName] || "var(--terminal-tool)";
  const barWidth = Math.max(pct * 1.5, 2); // scale for visual weight
  return (
    <div className="flex items-center gap-2">
      <span className="w-10 text-right text-zinc-400">{name}</span>
      <span className="text-zinc-500 w-5 text-right">{count}</span>
      <span className="text-zinc-600 w-7 text-right">{pct}%</span>
      <span className="flex-1 overflow-hidden">
        <span
          className="inline-block h-[2px] rounded-full"
          style={{ backgroundColor: colorVar, width: `${barWidth}%` }}
        />
      </span>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

export function TophDemo() {
  const [visibleEvents, setVisibleEvents] = useState<ActivityEvent[]>(() =>
    pooledEvents.slice(0, 8)
  );
  const [reducedMotion, setReducedMotion] = useState(false);
  const eventIndexRef = useRef(8);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    if (mq.matches) {
      setReducedMotion(true);
      setVisibleEvents(pooledEvents.slice(0, 10));
    }
  }, []);

  // Activity feed cycling
  useEffect(() => {
    if (reducedMotion) return;
    const interval = setInterval(() => {
      const nextIndex = eventIndexRef.current % pooledEvents.length;
      eventIndexRef.current += 1;
      setVisibleEvents((prev) => {
        const next = [...prev, pooledEvents[nextIndex]];
        if (next.length > 10) return next.slice(next.length - 10);
        return next;
      });
    }, 2500);
    return () => clearInterval(interval);
  }, [reducedMotion]);

  const panelLabel = (text: string) => (
    <div className="text-[11px] uppercase tracking-wider font-bold mb-2" style={{ color: "var(--terminal-border)" }}>
      {text}
    </div>
  );

  return (
    <div className="bg-zinc-950 font-mono text-[11px] leading-snug text-zinc-300 w-full">
      {/* ─── Top row: Sessions + Detail | Activity ─── */}
      <div className="grid grid-cols-[40fr_60fr] md:grid-cols-[35fr_65fr] border-b border-zinc-800">
        {/* Left column: Sessions + Detail stacked */}
        <div className="border-r border-zinc-800">
          {/* Sessions */}
          <div className="p-3 border-b border-zinc-800">
            {panelLabel("Sessions")}
            <div className="space-y-1">
              {sessions.map((s, i) => (
                <div key={`${s.name}-${i}`} className="flex items-center gap-1.5">
                  <OrbitDot color="var(--terminal-active)" />
                  <span className="text-zinc-200">{s.name}</span>
                  <span className="text-zinc-600">{s.branch}</span>
                  <span className="text-zinc-700 text-[10px] ml-auto font-mono tracking-tight">
                    {sparklines[i]}
                  </span>
                  <span className="text-zinc-500 ml-1">{s.age}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Detail */}
          <div className="p-3 border-b border-zinc-800">
            {panelLabel("Detail")}
            <div className="space-y-0.5">
              <div className="flex items-center gap-2">
                <span style={{ color: "var(--terminal-active)" }}>●</span>
                <span style={{ color: "var(--terminal-active)" }}>active</span>
              </div>
              <div>
                <span className="text-zinc-600">dir  </span>
                <span className="text-zinc-300">/Users/patty/dev/tp</span>
              </div>
              <div>
                <span className="text-zinc-600">git  </span>
                <span style={{ color: "var(--terminal-active)" }}>main</span>
                <span className="text-zinc-400 ml-2">claude-opus-4-6</span>
              </div>
              <div>
                <span className="text-zinc-600">age  </span>
                <span className="text-zinc-300">18s</span>
                <span className="text-zinc-600 ml-2">last </span>
                <span style={{ color: "var(--terminal-bash)" }} className="font-bold">Bash</span>
              </div>
              <div>
                <span className="text-zinc-600">tok  </span>
                <span className="text-zinc-300">84 in / 19.8K out</span>
              </div>
              {/* Context bar */}
              <div className="flex items-center gap-2 mt-1">
                <span className="text-zinc-600">ctx</span>
                <div className="flex-1 bg-zinc-800 h-2 rounded-sm overflow-hidden">
                  <div
                    className="h-full rounded-sm"
                    style={{
                      width: "9%",
                      background: "linear-gradient(90deg, var(--terminal-active), var(--terminal-waiting))",
                    }}
                  />
                </div>
                <span className="text-zinc-500 text-[10px]">9%</span>
              </div>
              {/* Subagent tree */}
              <div className="mt-1">
                <span className="text-zinc-600">└ </span>
                <span style={{ color: "var(--terminal-active)" }}>●</span>
                <span className="text-zinc-400 ml-1">Explore:</span>
                <span className="text-zinc-300 ml-1">Explore tp project</span>
              </div>
            </div>
          </div>

          {/* Tools (hidden on small screens) */}
          <div className="hidden md:block p-3">
            {panelLabel("Tools")}
            <div className="space-y-1">
              {toolBarData.map((t) => (
                <ToolBarRow key={t.name} name={t.name} count={t.count} pct={t.pct} />
              ))}
            </div>
          </div>
        </div>

        {/* Activity Feed */}
        <div className="p-3 flex flex-col">
          {panelLabel("Activity")}
          <div className="space-y-0.5 overflow-hidden flex-1">
            {visibleEvents.map((ev, i) => {
              const color = toolColorVars[ev.tool];
              const glyph = toolGlyphs[ev.tool];
              const showGap = i === 3 || i === 6; // visual rhythm breaks
              return (
                <div key={`${ev.time}-${ev.tool}-${i}`}>
                  {showGap && <GapSeparator />}
                  <div className="flex items-center gap-1.5">
                    <span className="text-zinc-600 shrink-0">{ev.time}</span>
                    <span className="text-zinc-600 shrink-0">{ev.session}</span>
                    <span style={{ color }} className="shrink-0">{glyph}</span>
                    <span style={{ color }} className="font-bold shrink-0">
                      {ev.tool}
                      {ev.count && ev.count > 1 ? ` \u00d7${ev.count}` : ""}
                    </span>
                    <span className="text-zinc-500 truncate">{ev.detail}</span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* ─── Bottom row: Metrics (hidden on mobile) ─── */}
      <div className="hidden md:grid grid-cols-2">
        <div className="p-3 border-r border-zinc-800">
          {panelLabel("Metrics")}
          <div className="space-y-0.5">
            <div>
              <span style={{ color: "var(--terminal-active)" }}>in 229</span>
              <span className="text-zinc-600 mx-2"> </span>
              <span style={{ color: "var(--terminal-waiting)" }}>out 34.3K</span>
              <span className="text-zinc-500 ml-2">1313 tok/s</span>
            </div>
            <div>
              <span className="text-zinc-600">cache </span>
              <span style={{ color: "var(--terminal-active)" }}>100%</span>
              <span className="text-zinc-400 ml-1">4.1M</span>
            </div>
            <div>
              <span className="text-zinc-600">cost  </span>
              <span style={{ color: "var(--terminal-active)" }}>$17.23</span>
              <span className="text-zinc-500 ml-2">$354/hr</span>
            </div>
            <div>
              <span className="text-zinc-600">sessions </span>
              <span className="text-zinc-200">2</span>
            </div>
          </div>
        </div>

        {/* Status bar style footer */}
        <div className="p-3 flex items-end">
          <div className="flex items-center gap-4 text-zinc-600 text-[10px]">
            <span><span className="text-zinc-400">tab</span> panels</span>
            <span><span className="text-zinc-400">j/k</span> navigate</span>
            <span><span className="text-zinc-400">?</span> help</span>
            <span><span className="text-zinc-400">q</span> quit</span>
            <span className="ml-auto">
              jsonl <span style={{ color: "var(--terminal-active)" }}>●</span> 30fps <span className="text-zinc-300">2 active</span>
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
