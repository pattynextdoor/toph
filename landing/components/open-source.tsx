export function OpenSource() {
  return (
    <section className="px-6 lg:px-16 py-24 border-t border-zinc-800/50">
      <div className="max-w-2xl mx-auto text-center">
        <h2 className="font-heading text-3xl font-bold text-zinc-50 mb-4">
          Open source
        </h2>
        <p className="text-base text-zinc-500 leading-relaxed">
          Built with Go and Bubble Tea. MIT licensed. Contributions welcome.
        </p>
        <div className="mt-8 flex items-center justify-center gap-6">
          <a
            href="https://github.com/pattynextdoor/toph"
            className="inline-flex items-center gap-2 rounded-lg border border-zinc-800 bg-zinc-900/50 px-5 py-2.5 font-mono text-sm text-zinc-300 transition-colors hover:border-zinc-700 hover:text-zinc-100 cursor-pointer"
            aria-label="View toph on GitHub"
          >
            View on GitHub &#8599;
          </a>
        </div>
        <div className="mt-8 flex items-center justify-center gap-4 font-mono text-xs text-zinc-600">
          <span>Go</span>
          <span className="text-zinc-800">&middot;</span>
          <span>Bubble Tea</span>
          <span className="text-zinc-800">&middot;</span>
          <span>MIT License</span>
        </div>
      </div>
    </section>
  );
}
