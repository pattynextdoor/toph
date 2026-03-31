"use client";

import { useEffect, useState } from "react";
import { InstallCommand } from "@/components/install-command";
import { TophDemo } from "@/components/toph-demo";
import { useReducedMotion } from "@/hooks/use-reduced-motion";
import { AnimatedGridPattern } from "@/components/ui/animated-grid-pattern";
import { BorderBeam } from "@/components/ui/border-beam";

// Stagger delays (ms) for left column elements
const STAGGER = {
  nav: 100,
  headline: 200,
  subtext: 400,
  install: 550,
  scroll: 700,
  // Right column
  terminal: 300,
  chrome: 200,
  panels: 500,
};

const DURATION = 700;

export function Hero() {
  const reducedMotion = useReducedMotion();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // Trigger entrance after first paint
    const timer = setTimeout(() => setMounted(true), 50);
    return () => clearTimeout(timer);
  }, []);

  // If reduced motion, show everything immediately
  const show = reducedMotion || mounted;

  const enter = (delay: number, from: "bottom" | "left" | "scale" = "bottom") => {
    if (reducedMotion) return {};
    const transforms: Record<string, string> = {
      bottom: "translateY(20px)",
      left: "translateX(-20px)",
      scale: "scale(0.96)",
    };
    return {
      style: {
        opacity: show ? 1 : 0,
        transform: show ? "none" : transforms[from],
        transition: `opacity ${DURATION}ms ease-out ${delay}ms, transform ${DURATION}ms ease-out ${delay}ms`,
      },
    };
  };

  return (
    <section className="min-h-screen grid lg:grid-cols-[38fr_62fr]">
      {/* Left Column */}
      <div className="relative flex flex-col justify-center px-6 lg:px-12 py-20 text-center lg:text-left">
        {/* Nav */}
        <nav
          className="absolute top-8 left-6 right-6 lg:left-12 lg:right-12 flex flex-row justify-between items-center"
          {...enter(STAGGER.nav)}
        >
          <span className="font-mono text-sm text-zinc-500">toph</span>
          <a
            href="https://github.com/pattynextdoor/toph"
            aria-label="View toph on GitHub"
            className="font-mono text-sm text-zinc-500 transition-colors duration-200 hover:text-zinc-300 cursor-pointer"
          >
            GitHub &#8599;
          </a>
        </nav>

        {/* Headline */}
        <h1 className="font-heading text-5xl lg:text-[64px] font-bold text-zinc-50 leading-[1.05] tracking-[-0.03em]">
          <span className="block" {...enter(STAGGER.headline, "left")}>btop</span>
          <span className="block text-zinc-400" {...enter(STAGGER.headline + 100, "left")}>for AI agents.</span>
        </h1>

        {/* Subtext */}
        <p
          className="mt-8 font-sans text-base text-zinc-500 leading-relaxed max-w-[32ch] mx-auto lg:mx-0"
          {...enter(STAGGER.subtext)}
        >
          See what your agents are doing. Real-time terminal dashboard.
        </p>

        {/* Install Command */}
        <div className="mt-8 overflow-hidden max-w-md mx-auto lg:mx-0" {...enter(STAGGER.install)}>
          <InstallCommand />
        </div>

        {/* Scroll hint */}
        <span
          className="absolute bottom-8 left-6 lg:left-12 font-mono text-[11px] text-zinc-700 hidden lg:block"
          {...enter(STAGGER.scroll)}
        >
          &#8595; scroll
        </span>
      </div>

      {/* Right Column */}
      <div className="relative flex items-center justify-center overflow-hidden border-l border-white/[0.04] bg-[#0d0d10] p-8 lg:p-12 min-h-[400px]">
        {/* Animated grid pattern */}
        {!reducedMotion && (
          <AnimatedGridPattern
            numSquares={30}
            maxOpacity={0.15}
            duration={3}
            repeatDelay={1}
            className="inset-0 h-full w-full opacity-30"
          />
        )}

        {/* Terminal mockup — scales in */}
        <div
          className="relative z-10 w-full max-w-[900px] overflow-hidden rounded-xl border border-zinc-800 bg-zinc-950"
          aria-label="toph terminal dashboard showing session monitoring"
          {...enter(STAGGER.terminal, "scale")}
        >
          {/* Title bar */}
          <div className="flex items-center gap-2 border-b border-zinc-800 px-4 py-3" {...enter(STAGGER.chrome)}>
            <div className="h-3 w-3 rounded-full bg-red-500/70" />
            <div className="h-3 w-3 rounded-full bg-yellow-500/70" />
            <div className="h-3 w-3 rounded-full bg-green-500/70" />
            <span className="ml-2 font-mono text-xs text-zinc-500">toph</span>
          </div>

          {/* Dashboard interior — fades in after chrome */}
          <div {...enter(STAGGER.panels)}>
            <TophDemo />
          </div>

          {/* BorderBeam glow */}
          {!reducedMotion && (
            <BorderBeam
              duration={4}
              colorFrom="#87AFFF"
              colorTo="#87AFFF"
              size={300}
            />
          )}
        </div>
      </div>
    </section>
  );
}
