"use client";

import { useState } from "react";
import { Check, Copy } from "lucide-react";

export function InstallCommand() {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText("brew install pattynextdoor/tap/toph");
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className="flex items-center justify-between gap-3 rounded-md border border-zinc-800 bg-zinc-900/50 px-4 py-3 font-mono text-[13px]">
      <code>
        <span className="text-zinc-600">$ </span>
        <span className="text-zinc-300">brew install pattynextdoor/tap/toph</span>
      </code>
      <button
        onClick={handleCopy}
        className="cursor-pointer text-zinc-500 transition-colors duration-200 hover:text-zinc-300 focus:outline-none focus:ring-2 focus:ring-zinc-400 focus:ring-offset-2 focus:ring-offset-zinc-950 rounded-sm"
        aria-label="Copy install command"
      >
        {copied ? (
          <Check className="h-4 w-4 text-green-400" />
        ) : (
          <Copy className="h-4 w-4" />
        )}
      </button>
    </div>
  );
}
