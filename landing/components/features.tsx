const features = [
  {
    label: "01",
    title: "Zero config",
    description:
      "Point it at your terminal. toph auto-discovers Claude Code sessions from ~/.claude/projects/ — no setup, no config files, no environment variables. Just run it.",
  },
  {
    label: "02",
    title: "Live",
    description:
      "Activity feed updates every 2.5 seconds. Token counts tick up in real-time. Context fill meters animate as your agent thinks. 30fps rendering via Bubble Tea.",
  },
  {
    label: "03",
    title: "Beautiful",
    description:
      "Screenshot-worthy on first launch. Dark theme with color-coded panels, braille sparklines, and spring-animated meters. Built for the terminal, designed for r/unixporn.",
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
            <div key={feature.label}>
              <span className="font-mono text-sm text-zinc-600">
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
