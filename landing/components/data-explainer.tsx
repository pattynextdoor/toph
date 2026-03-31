import { FileJson, Webhook, Activity } from "lucide-react";

const sources = [
  {
    title: "JSONL Logs",
    description:
      "Session files at ~/.claude/projects/ with every message, tool call, and usage record. toph watches them with fsnotify.",
    detail: "~/.claude/projects/{hash}/{session}.jsonl",
    icon: FileJson,
  },
  {
    title: "Hooks",
    description:
      "Claude Code's hook system fires events for tool use, permissions, errors, and session lifecycle. Opt-in for richer real-time data.",
    detail: "PreToolUse \u00b7 PostToolUse \u00b7 Stop",
    icon: Webhook,
  },
  {
    title: "Process Scan",
    description:
      "Detects running claude processes for CPU and memory usage. Works as a fallback even without log access.",
    detail: "ps aux | grep claude",
    icon: Activity,
  },
];

export function DataExplainer() {
  return (
    <section className="px-6 lg:px-16 py-24 border-t border-zinc-800/50">
      <div className="max-w-4xl mx-auto">
        <h2 className="font-heading text-3xl font-bold text-zinc-50 text-center mb-4">
          Under the hood
        </h2>
        <p className="text-center text-zinc-500 text-base max-w-2xl mx-auto mb-16">
          Your Claude Code sessions generate rich JSONL logs — tool calls, token
          usage, cost data, subagent activity. toph reads these in real-time, no
          API keys required.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {sources.map((source) => {
            const Icon = source.icon;
            return (
              <div
                key={source.title}
                className="rounded-lg border border-[#87AFFF]/10 bg-[#87AFFF]/[0.03] p-6"
              >
                <div className="mb-4 inline-flex items-center justify-center rounded-md border border-[#87AFFF]/15 bg-[#87AFFF]/[0.06] p-2.5">
                  <Icon className="h-5 w-5 text-[#87AFFF]" strokeWidth={1.5} />
                </div>
                <h3 className="font-sans text-lg font-medium text-zinc-100">
                  {source.title}
                </h3>
                <p className="text-sm text-zinc-400 mt-2 leading-relaxed">
                  {source.description}
                </p>
                <p className="font-mono text-xs text-[#87AFFF]/50 mt-4">
                  {source.detail}
                </p>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
