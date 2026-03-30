export function Footer() {
  return (
    <footer className="border-t border-zinc-800 px-6 lg:px-16 py-6">
      <div className="flex flex-col items-center gap-2 lg:flex-row lg:justify-between">
        <span className="font-mono text-xs text-zinc-700">toph</span>
        <span className="font-mono text-xs text-zinc-700">
          <a
            href="https://github.com/pattynextdoor/toph"
            className="transition-colors hover:text-zinc-400"
            aria-label="View toph on GitHub"
          >
            GitHub
          </a>
          {" "}· MIT · Built with Go + Bubble Tea
        </span>
      </div>
    </footer>
  );
}
