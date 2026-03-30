const features = [
  {
    title: "Zero config",
    description:
      "Just run it. toph finds your Claude Code sessions automatically.",
  },
  {
    title: "Live",
    description:
      "Real-time activity feed, token tracking, cost estimation. 30fps.",
  },
  {
    title: "Beautiful",
    description:
      "Screenshot-worthy on first launch. Dark theme, smooth animations.",
  },
];

export function Features() {
  return (
    <div className="max-w-[960px] mx-auto mt-20 px-6 lg:px-0">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-px bg-zinc-800">
        {features.map((feature) => (
          <div key={feature.title} className="bg-zinc-950 px-9 py-10">
            <h3 className="text-[22px] font-semibold text-zinc-50">
              {feature.title}
            </h3>
            <p className="mt-4 text-[15px] leading-relaxed text-zinc-600">
              {feature.description}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}
