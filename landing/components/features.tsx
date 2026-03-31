const features = [
  {
    label: "01",
    title: "Sessions",
    description: (
      <>
        Every running Claude Code session across your machine. Active, idle,
        waiting for permission — toph sees them all by watching{" "}
        <code>~/.claude/projects/</code> for filesystem vibrations.
      </>
    ),
  },
  {
    label: "02",
    title: "Token flow",
    description: (
      <>
        <code>input_tokens</code>, <code>output_tokens</code>,{" "}
        <code>cache_read</code>, burn rate. toph parses every assistant response
        to track context window fill and estimated cost in real-time.
      </>
    ),
  },
  {
    label: "03",
    title: "Tool calls",
    description: (
      <>
        <code>Bash</code>, <code>Edit</code>, <code>Read</code>,{" "}
        <code>Write</code>, <code>Grep</code> — every tool invocation your
        agents make, streamed into a live activity feed. See what your agents are
        doing the moment they do it.
      </>
    ),
  },
];

export function Features() {
  return (
    <section className="px-6 lg:px-16 py-24">
      <div className="max-w-5xl mx-auto">
        <h2 className="font-heading text-3xl font-bold text-zinc-50 text-center mb-16">
          What toph senses
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {features.map((feature) => (
            <div key={feature.label} className="relative pl-5 border-l border-[#87AFFF]/20">
              <span className="font-mono text-sm text-[#87AFFF]">
                {feature.label}
              </span>
              <h3 className="text-xl font-semibold text-zinc-50 mt-2">
                {feature.title}
              </h3>
              <p className="text-sm text-zinc-400 leading-relaxed mt-3">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
