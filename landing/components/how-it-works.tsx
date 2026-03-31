const steps = [
  {
    number: "1",
    title: "Install",
    command: "brew install pattynextdoor/tap/toph",
    description: "Single binary. No dependencies.",
  },
  {
    number: "2",
    title: "Run your agents",
    command: "claude code",
    description:
      "Start Claude Code sessions as usual. Nothing changes about your workflow.",
  },
  {
    number: "3",
    title: "Watch",
    command: "toph",
    description:
      "That's it. toph finds your sessions automatically and shows you everything.",
  },
];

export function HowItWorks() {
  return (
    <section className="px-6 lg:px-16 py-24 border-t border-zinc-800/50">
      <div className="max-w-3xl mx-auto">
        <h2 className="font-heading text-3xl font-bold text-zinc-50 text-center mb-16">
          Get started
        </h2>
        <div className="space-y-12">
          {steps.map((step) => (
            <div
              key={step.number}
              className="flex flex-col md:flex-row md:gap-8 gap-2"
            >
              <span className="font-mono text-4xl font-bold text-zinc-800 md:w-16 shrink-0">
                {step.number}
              </span>
              <div>
                <h3 className="font-sans text-xl font-semibold text-zinc-50">
                  {step.title}
                </h3>
                <p className="font-mono text-sm text-zinc-400 mt-1">
                  <span className="text-zinc-600">$ </span>
                  {step.command}
                </p>
                <p className="text-sm text-zinc-500 mt-2">
                  {step.description}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
