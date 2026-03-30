"use client";

import { useEffect, useRef, useState } from "react";
import { NumberTicker } from "@/components/ui/number-ticker";

// ---------------------------------------------------------------------------
// Data
// ---------------------------------------------------------------------------

const sessions = [
  { icon: "●", name: "api-server", branch: "feat/oauth", colorVar: "var(--terminal-active)", pulseClass: "" },
  { icon: "◐", name: "auth-flow", branch: "fix/session", colorVar: "var(--terminal-waiting)", pulseClass: "animate-pulse-amber" },
  { icon: "○", name: "docs", branch: "main", colorVar: "var(--terminal-idle)", pulseClass: "" },
] as const;

const toolBars = [
  { name: "Bash", filled: 8, empty: 2, count: 34 },
  { name: "Edit", filled: 6, empty: 4, count: 28 },
  { name: "Read", filled: 4, empty: 6, count: 19 },
  { name: "Glob", filled: 2, empty: 8, count: 12 },
] as const;

const allEvents = [
  { time: "14:23", tool: "Edit", file: "src/auth.ts" },
  { time: "14:23", tool: "Bash", file: "npm test" },
  { time: "14:24", tool: "Read", file: "package.json" },
  { time: "14:24", tool: "Glob", file: "src/**/*.ts" },
  { time: "14:25", tool: "Edit", file: "src/middleware.ts" },
  { time: "14:25", tool: "Bash", file: "npm run build" },
  { time: "14:26", tool: "Read", file: "tsconfig.json" },
  { time: "14:26", tool: "Edit", file: "src/routes/api.ts" },
  { time: "14:27", tool: "Bash", file: "git status" },
  { time: "14:27", tool: "Glob", file: "tests/**/*.test.ts" },
] as const;

const toolSymbols: Record<string, string> = {
  Edit: "✎",
  Bash: "▶",
  Read: "◉",
  Glob: "◎",
};

const toolColors: Record<string, string> = {
  Edit: "var(--terminal-tool)",
  Bash: "var(--terminal-active)",
  Read: "", // uses text-zinc-500 class instead
  Glob: "var(--terminal-tool)",
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function ToolBarRow({ name, filled, empty, count }: (typeof toolBars)[number]) {
  return (
    <div className="flex items-center gap-2">
      <span className="w-8 text-right text-zinc-400">{name}</span>
      <span>
        <span style={{ color: "var(--terminal-tool)" }}>{"█".repeat(filled)}</span>
        <span className="text-zinc-700">{"░".repeat(empty)}</span>
      </span>
      <span className="text-zinc-500">{count}</span>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function TophDemo() {
  const [visibleEvents, setVisibleEvents] = useState<typeof allEvents[number][]>(() =>
    allEvents.slice(0, 4) as unknown as typeof allEvents[number][]
  );
  const [contextWidth, setContextWidth] = useState(74);
  const [reducedMotion, setReducedMotion] = useState(false);
  const eventIndexRef = useRef(4); // next event to add

  // Detect prefers-reduced-motion on mount
  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    if (mq.matches) {
      setReducedMotion(true);
      // Set static "most appealing" state
      setVisibleEvents(allEvents.slice(0, 6) as unknown as typeof allEvents[number][]);
      setContextWidth(78);
    }
  }, []);

  // Context meter animation (74% → 78%)
  useEffect(() => {
    if (reducedMotion) return;
    const timer = setTimeout(() => setContextWidth(78), 500);
    return () => clearTimeout(timer);
  }, [reducedMotion]);

  // Activity feed cycling
  useEffect(() => {
    if (reducedMotion) return;

    const interval = setInterval(() => {
      const nextIndex = eventIndexRef.current % allEvents.length;
      eventIndexRef.current += 1;

      setVisibleEvents((prev) => {
        const next = [...prev, allEvents[nextIndex]];
        // Keep max 6 visible — drop oldest
        if (next.length > 6) return next.slice(next.length - 6);
        return next;
      });
    }, 2500);

    return () => clearInterval(interval);
  }, [reducedMotion]);

  // ------------------------------------------------------------------
  // Panel label helper
  // ------------------------------------------------------------------
  const panelLabel = (text: string) => (
    <div className="text-[11px] uppercase tracking-wider text-zinc-500 mb-2">{text}</div>
  );

  // ------------------------------------------------------------------
  // Render
  // ------------------------------------------------------------------
  return (
    <div className="bg-zinc-950 font-mono text-[12px] text-zinc-300 w-full">
      {/* Top row */}
      <div className="grid grid-cols-[40fr_60fr] md:grid-cols-[22fr_36fr_42fr] border-b border-zinc-800">
        {/* Panel 1 — Sessions */}
        <div className="p-3 border-r border-zinc-800">
          {panelLabel("Sessions")}
          <div className="space-y-2">
            {sessions.map((s) => (
              <div key={s.name}>
                <div className="flex items-center gap-2">
                  <span className={s.pulseClass} style={{ color: s.colorVar }}>
                    {s.icon}
                  </span>
                  <span className="text-zinc-200">{s.name}</span>
                </div>
                <div className="ml-5 text-zinc-500 text-[11px]">{s.branch}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Panel 2 — Detail */}
        <div className="hidden md:block p-3 border-r border-zinc-800">
          {panelLabel("Detail")}
          <div className="space-y-1">
            <div>
              <span className="text-zinc-500">session: </span>
              <span className="text-zinc-200">api-server</span>
            </div>
            <div>
              <span className="text-zinc-500">status: </span>
              <span style={{ color: "var(--terminal-active)" }}>active</span>
            </div>
            <div>
              <span className="text-zinc-500">branch: </span>
              <span className="text-zinc-200">feat/oauth</span>
            </div>
            <div>
              <span className="text-zinc-500">tokens: </span>
              <NumberTicker
                value={14832}
                startValue={14200}
                className="text-zinc-200 text-[12px] font-mono"
              />
            </div>
            <div>
              <span className="text-zinc-500">cost: </span>
              <span className="text-zinc-200">$1.42</span>
            </div>
            <div>
              <span className="text-zinc-500">context: </span>
              <span className="text-zinc-400 text-[11px]">{contextWidth}%</span>
            </div>
            <div className="bg-zinc-800 h-2 rounded-full overflow-hidden mt-1">
              <div
                className="h-full rounded-full"
                style={{
                  backgroundColor: "var(--terminal-active)",
                  width: `${contextWidth}%`,
                  transition: reducedMotion ? "none" : "all 2000ms ease-out",
                }}
              />
            </div>
          </div>

          {/* Tool usage sub-section */}
          <div className="mt-3">
            {panelLabel("Tools")}
            <div className="space-y-1">
              {toolBars.map((t) => (
                <ToolBarRow key={t.name} {...t} />
              ))}
            </div>
          </div>
        </div>

        {/* Panel 3 — Activity Feed */}
        <div className="p-3">
          {panelLabel("Activity")}
          <div className="space-y-1 overflow-hidden">
            {visibleEvents.map((ev, i) => {
              const colorStyle = toolColors[ev.tool]
                ? { color: toolColors[ev.tool] }
                : undefined;
              const colorClass = ev.tool === "Read" ? "text-zinc-500" : "";

              return (
                <div key={`${ev.time}-${ev.tool}-${ev.file}-${i}`} className="flex items-center gap-2">
                  <span className="text-zinc-500 shrink-0">{ev.time}</span>
                  <span className={colorClass} style={colorStyle}>
                    {toolSymbols[ev.tool]}
                  </span>
                  <span className={colorClass} style={colorStyle}>
                    {ev.file}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Bottom row */}
      <div className="hidden md:grid grid-cols-2">
        {/* Panel 4 — Metrics */}
        <div className="p-3 border-r border-zinc-800">
          {panelLabel("Metrics")}
          <div className="space-y-1">
            <div>
              <span className="text-zinc-500">tokens: </span>
              <NumberTicker
                value={14832}
                startValue={14200}
                className="text-zinc-200 text-[12px] font-mono"
              />
            </div>
            <div>
              <span className="text-zinc-500">cost: </span>
              <span className="text-zinc-200">$1.42</span>
            </div>
            <div>
              <span className="text-zinc-500">burn rate: </span>
              <span className="text-zinc-200">~420 tok/min</span>
            </div>
            <div>
              <span className="text-zinc-500">sessions: </span>
              <span className="text-zinc-200">3 active</span>
            </div>
          </div>
        </div>

        {/* Panel 5 — Tools */}
        <div className="p-3">
          {panelLabel("Tools")}
          <div className="space-y-1">
            {toolBars.map((t) => (
              <ToolBarRow key={t.name} {...t} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
