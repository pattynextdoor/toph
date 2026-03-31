# toph Landing Page Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the toph landing page — a split-screen hero with animated terminal mockup + centered features section. Two viewports, dark-only, Next.js static export.

**Architecture:** Next.js app in `landing/` with shadcn/ui for base components and magicui for Terminal, AnimatedGridPattern, NumberTicker, BorderBeam, ShimmerButton. Single `page.tsx` renders both viewports. The animated terminal mockup is a custom `<TophDemo />` React component with choreographed state cycling.

**Tech Stack:** Next.js 16+, TypeScript, Tailwind CSS v4, shadcn/ui (zinc theme), magicui, Geist fonts (built-in), Lucide icons

**Design docs to reference:**
- `docs/plans/2026-03-30-landing-page-design.md` — Full design spec with ASCII mockups
- `design-system/toph/MASTER.md` — Color palette, typography, spacing
- `design-system/toph/pages/landing.md` — Layout overrides, animation budget

---

### Task 1: Scaffold Next.js project

**Files:**
- Create: `landing/` (entire directory via create-next-app)

**Step 1: Create the Next.js app**

```bash
cd /Users/patty/dev/toph
npx create-next-app@latest landing --yes
```

This creates a Next.js project with TypeScript, Tailwind CSS, ESLint, App Router, and Geist fonts pre-configured.

**Step 2: Configure static export**

Edit `landing/next.config.ts`:

```typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
};

export default nextConfig;
```

**Step 3: Verify it runs**

```bash
cd landing && npm run dev
```

Expected: Dev server starts at localhost:3000, shows default Next.js page.

**Step 4: Commit**

```bash
cd /Users/patty/dev/toph
git add landing/
git commit -m "🔧 chore: scaffold Next.js landing page project"
```

---

### Task 2: Initialize shadcn/ui + install magicui components

**Files:**
- Modify: `landing/components.json` (created by shadcn init)
- Create: `landing/components/ui/` (shadcn component files)
- Create: `landing/lib/utils.ts`

**Step 1: Initialize shadcn/ui**

```bash
cd /Users/patty/dev/toph/landing
npx shadcn@latest init
```

When prompted:
- Style: **New York**
- Base color: **Zinc** (matches our design system)
- CSS variables: **Yes**

**Step 2: Install magicui components**

```bash
npx shadcn@latest add @magicui/terminal
npx shadcn@latest add @magicui/animated-grid-pattern
npx shadcn@latest add @magicui/number-ticker
npx shadcn@latest add @magicui/border-beam
npx shadcn@latest add @magicui/shimmer-button
```

**Step 3: Verify components installed**

```bash
ls landing/components/ui/
```

Expected: `terminal.tsx`, `animated-grid-pattern.tsx`, `number-ticker.tsx`, `border-beam.tsx`, `shimmer-button.tsx` (plus any dependencies like `utils.ts`).

**Step 4: Commit**

```bash
cd /Users/patty/dev/toph
git add landing/
git commit -m "🔧 chore: add shadcn/ui and magicui components"
```

---

### Task 3: Set up global styles and dark theme

**Files:**
- Modify: `landing/app/globals.css`
- Modify: `landing/app/layout.tsx`

**Step 1: Configure globals.css for dark-only zinc theme**

Replace `landing/app/globals.css` with the dark-only theme. The key customizations:
- Force dark mode (no light mode toggle)
- Set background to `zinc-950` (#09090b)
- Add custom terminal color CSS variables
- Add `@keyframes pulse-amber` for the waiting session indicator
- Add `prefers-reduced-motion` media query

Key CSS variables to define:
```css
:root {
  --terminal-active: #87D787;
  --terminal-waiting: #FFD787;
  --terminal-tool: #87D7D7;
  --terminal-error: #FF8787;
  --terminal-idle: #6C6C6C;
  --terminal-subagent: #D7AFFF;
  --terminal-border: #87AFFF;
  --surface-darker: #0d0d10;
}
```

**Step 2: Configure layout.tsx**

Update `landing/app/layout.tsx`:
- Set `<html lang="en" className="dark">` to force dark mode
- Metadata: title "toph — btop for AI agents", description "A terminal dashboard for AI coding agents. See what your agents are doing."
- OG metadata placeholder
- Keep Geist fonts (already configured by create-next-app)
- Set body className to `bg-zinc-950 text-zinc-50 antialiased`

**Step 3: Clear default page content**

Replace `landing/app/page.tsx` with a minimal placeholder:

```tsx
export default function Home() {
  return (
    <main className="min-h-screen bg-zinc-950">
      <p className="text-zinc-400 p-12">toph landing page</p>
    </main>
  );
}
```

**Step 4: Verify dark theme**

```bash
cd /Users/patty/dev/toph/landing && npm run dev
```

Expected: Dark background, zinc-400 placeholder text visible at localhost:3000.

**Step 5: Commit**

```bash
cd /Users/patty/dev/toph
git add landing/app/
git commit -m "🎨 style: configure dark-only zinc theme and layout"
```

---

### Task 4: Build the install-command component

**Files:**
- Create: `landing/components/install-command.tsx`

**Step 1: Build the component**

A code block showing `$ brew install pattynextdoor/tap/toph` with a copy-to-clipboard button.

Specs:
- Geist Mono font (use `font-mono` Tailwind class)
- `$` prompt in `zinc-600`, command text in `zinc-300`
- Background: `zinc-900/50`, border: `1px solid zinc-800`, border-radius: 6px
- Copy button: Lucide `Copy` icon in `zinc-500`, on click changes to `Check` icon for 1.5s
- `aria-label="Copy install command"` on the button
- `cursor-pointer` on the button

```tsx
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
        {copied ? <Check className="h-4 w-4 text-green-400" /> : <Copy className="h-4 w-4" />}
      </button>
    </div>
  );
}
```

**Step 2: Test in page.tsx**

Temporarily render `<InstallCommand />` in page.tsx to verify visually.

**Step 3: Commit**

```bash
git add landing/components/install-command.tsx
git commit -m "✨ feat(landing): add install command component with copy-to-clipboard"
```

---

### Task 5: Build the TophDemo animated terminal interior

**Files:**
- Create: `landing/components/toph-demo.tsx`

This is the most complex component. It's a custom React component that renders a simulated 5-panel toph dashboard with choreographed animations.

**Step 1: Build the static layout first**

Create `landing/components/toph-demo.tsx` with the 5-panel grid layout (no animation yet):

Panel structure:
- Top row: Sessions (22% width) | Detail (36%) | Activity Feed (42%)
- Bottom row: Metrics (50%) | Tools (50%)
- All text in `font-mono text-[12px]`
- Panel labels: `text-[11px] uppercase tracking-wider text-zinc-500`
- Panel borders: `border-zinc-800`

Static content:
- Sessions: 3 sessions with status icons (●, ◐, ○) and colored text
- Detail: session name, status, branch, token count, cost, context meter progress bar
- Activity Feed: 6 hardcoded events with tool name + file path, color-coded
- Metrics: token count, cost, burn rate
- Tools: horizontal bar chart (Bash, Edit, Read, Glob)

**Step 2: Add animation choreography**

Convert to a client component with `useEffect` and `useState`:

- **Activity feed:** Use `setInterval` (2500ms) to cycle through a predefined list of events, appending to a visible list (max 6 visible, oldest drop off)
- **Context meter:** Animate width from 72% → 78% using CSS transition (`transition-all duration-[2000ms]`)
- **Token counter:** Use magicui `NumberTicker` component, incrementing the value every few seconds
- **Amber pulsing:** CSS `animate-pulse-amber` on the waiting session (defined in globals.css)

The animation state resets every ~45 seconds via a master `useEffect` cleanup.

**Step 3: Add `prefers-reduced-motion` support**

Use `useMediaQuery` or a simple `window.matchMedia('(prefers-reduced-motion: reduce)')` check. When reduced motion is preferred, show the static layout with no cycling — frozen at the initial "most appealing" state.

**Step 4: Verify visually**

Render `<TophDemo />` in page.tsx. Check:
- 5-panel layout renders correctly
- Events cycle in the activity feed
- Context meter animates
- Amber session pulses
- Colors match design system

**Step 5: Commit**

```bash
git add landing/components/toph-demo.tsx
git commit -m "✨ feat(landing): add animated terminal dashboard mockup"
```

---

### Task 6: Build the hero section (split-screen VP1)

**Files:**
- Create: `landing/components/hero.tsx`

**Step 1: Build the split-screen layout**

Specs from design doc:
- Full viewport height: `min-h-screen`
- CSS Grid or Flexbox: left 38%, right 62%
- Left column: `px-12` (48px), flex column, justify-center, relative (for nav + scroll hint)
- Right column: `bg-[#0d0d10]`, `border-l border-white/[0.04]`, relative, overflow-hidden

**Left column contents (top to bottom):**
1. Nav (absolute, top-8): "toph" text + ShimmerButton "GitHub" linking to `https://github.com/pattynextdoor/toph`
2. Headline: "btop" / "for AI agents." — 56-64px, bold, zinc-50, tight line-height
3. Subtext: "See what your agents are doing. Real-time terminal dashboard." — 18px, zinc-500
4. `<InstallCommand />` — margin-top-10
5. Scroll hint (absolute, bottom-8): "↓ scroll" in mono 11px zinc-700

**Right column contents:**
1. `<AnimatedGridPattern />` filling the column at ~35% opacity (use `maxOpacity={0.15}` and CSS mask)
2. `<TophDemo />` wrapped in a `<Terminal />` from magicui — this gives us the window chrome (dots, title bar)
3. The terminal wrapper gets `<BorderBeam />` for the glow effect (parent needs `relative overflow-hidden`)
4. Terminal vertically centered: flex items-center, padded 48px on each side

**Step 2: Wire up magicui components**

```tsx
import { Terminal } from "@/components/ui/terminal";
import { AnimatedGridPattern } from "@/components/ui/animated-grid-pattern";
import { BorderBeam } from "@/components/ui/border-beam";
import { ShimmerButton } from "@/components/ui/shimmer-button";
```

For the Terminal component: note that magicui's Terminal has built-in animation sequencing (TypingAnimation, AnimatedSpan). We're NOT using that — we're using the Terminal as a shell/chrome wrapper and rendering our custom `<TophDemo />` as its children. We may need to set `sequence={false}` or just render TophDemo as a child div.

For BorderBeam: the parent Terminal div needs `relative overflow-hidden`. Set `duration={4}` for slow trace, and use `colorFrom="#87AFFF"` `colorTo="#87AFFF"` (accent blue) or gradient from transparent via accent.

**Step 3: Verify the hero**

```bash
cd /Users/patty/dev/toph/landing && npm run dev
```

Expected: Split-screen hero with animated terminal on the right, headline + install on the left.

**Step 4: Commit**

```bash
git add landing/components/hero.tsx
git commit -m "✨ feat(landing): add split-screen hero section with terminal mockup"
```

---

### Task 7: Build VP2 — positioning strip, features, footer

**Files:**
- Create: `landing/components/positioning.tsx`
- Create: `landing/components/features.tsx`
- Create: `landing/components/footer.tsx`

**Step 1: Build the positioning strip**

Full-width band at top of VP2:
- Background: `bg-[#0d0d10]`
- Borders: `border-y border-white/[0.03]`
- Padding: `py-7 px-16` (28px / 64px)
- Text: `font-mono text-[13px] tracking-[0.04em] text-center`
- "Not a macOS app. Not a web dashboard. Not a tmux hack." in `text-zinc-600`
- "A real terminal dashboard." in `text-zinc-400`

```tsx
export function Positioning() {
  return (
    <div className="border-y border-white/[0.03] bg-[#0d0d10] px-16 py-7 text-center font-mono text-[13px] tracking-[0.04em]">
      <span className="text-zinc-600">
        Not a macOS app. Not a web dashboard. Not a tmux hack.{" "}
      </span>
      <span className="text-zinc-400">A real terminal dashboard.</span>
    </div>
  );
}
```

**Step 2: Build the features grid**

Ruled 3-column grid:
- Outer container: `max-w-[960px] mx-auto mt-20 grid grid-cols-3 gap-px bg-zinc-800`
- Each cell: `bg-zinc-950 px-9 py-10` (36px / 40px)
- Bold title: `text-[22px] font-semibold text-zinc-50`
- Description: `mt-4 text-[15px] leading-relaxed text-zinc-600`

Three features:
1. "Zero config" — "Just run it. toph finds your Claude Code sessions automatically."
2. "Live" — "Real-time activity feed, token tracking, cost estimation. 30fps."
3. "Beautiful" — "Screenshot-worthy on first launch. Dark theme, smooth animations."

```tsx
const features = [
  { title: "Zero config", description: "Just run it. toph finds your Claude Code sessions automatically." },
  { title: "Live", description: "Real-time activity feed, token tracking, cost estimation. 30fps." },
  { title: "Beautiful", description: "Screenshot-worthy on first launch. Dark theme, smooth animations." },
];
```

**Step 3: Build the footer**

Split layout:
- `flex justify-between items-center`
- Border-top: `border-t border-zinc-800`
- Padding: `px-16 py-6` (64px / 24px)
- Left: "toph" in mono 12px zinc-700
- Right: "GitHub · MIT · Built with Go + Bubble Tea" in mono 12px zinc-700
- "GitHub" is a link to the repo with `hover:text-zinc-400 transition-colors`

**Step 4: Add fade-in on scroll for VP2**

Wrap VP2 content in a container that uses IntersectionObserver to add `opacity-100` when visible:

```tsx
"use client";

import { useEffect, useRef, useState } from "react";

export function FadeInOnScroll({ children }: { children: React.ReactNode }) {
  const ref = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) setVisible(true); },
      { threshold: 0.1 }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, []);

  return (
    <div
      ref={ref}
      className={`transition-opacity duration-200 ease-out ${visible ? "opacity-100" : "opacity-0"}`}
    >
      {children}
    </div>
  );
}
```

Note: respect `prefers-reduced-motion` — if reduced motion, set `visible` to `true` immediately (skip the fade).

**Step 5: Commit**

```bash
git add landing/components/positioning.tsx landing/components/features.tsx landing/components/footer.tsx
git commit -m "✨ feat(landing): add positioning strip, features grid, and footer"
```

---

### Task 8: Wire up page.tsx

**Files:**
- Modify: `landing/app/page.tsx`

**Step 1: Compose the full page**

```tsx
import { Hero } from "@/components/hero";
import { Positioning } from "@/components/positioning";
import { Features } from "@/components/features";
import { Footer } from "@/components/footer";
import { FadeInOnScroll } from "@/components/fade-in-on-scroll";

export default function Home() {
  return (
    <main>
      {/* Viewport 1: Split-screen hero */}
      <Hero />

      {/* Viewport 2: Centered features + footer */}
      <section className="relative min-h-screen flex flex-col">
        <FadeInOnScroll>
          <Positioning />
          <Features />
        </FadeInOnScroll>
        <div className="flex-1" /> {/* Spacer pushes footer down */}
        <Footer />
      </section>
    </main>
  );
}
```

**Step 2: Verify full page flow**

```bash
cd /Users/patty/dev/toph/landing && npm run dev
```

Check:
- VP1 split-screen renders with terminal animation
- Scrolling reveals VP2 with fade-in
- Positioning strip → Features → Footer all render
- The VP1→VP2 transition feels like "turning a page"

**Step 3: Commit**

```bash
git add landing/app/page.tsx
git commit -m "✨ feat(landing): wire up full page with both viewports"
```

---

### Task 9: Responsive breakpoints

**Files:**
- Modify: `landing/components/hero.tsx`
- Modify: `landing/components/features.tsx`
- Modify: `landing/components/toph-demo.tsx`
- Modify: `landing/components/positioning.tsx`
- Modify: `landing/components/footer.tsx`

**Step 1: Hero responsive behavior**

The split-screen hero needs to collapse below 900px (use Tailwind `lg:` breakpoint at 1024px as closest standard):

- Desktop (`lg:` and up): `grid grid-cols-[38fr_62fr]`
- Mobile/tablet (below `lg:`): `flex flex-col` — text section stacks above terminal, both centered
- Headline size: `text-4xl lg:text-6xl`
- Terminal section: `w-full lg:w-auto`, terminal mockup centered
- Nav: adjust positioning for stacked layout
- Hide scroll hint on mobile

**Step 2: Features responsive**

- Desktop: `grid-cols-3`
- Mobile: `grid-cols-1` — features stack vertically, each with bottom border
- Padding adjustments: `px-6 lg:px-16`

**Step 3: Terminal mockup mobile simplification**

Below 768px (`md:` breakpoint), hide Detail + Metrics + Tools panels. Show only:
- Sessions panel (left, 40%)
- Activity Feed (right, 60%)

Use a `useMediaQuery` hook or CSS `hidden md:block` on the panels that should hide.

**Step 4: Positioning + Footer responsive**

- Reduce padding on mobile: `px-6 lg:px-16`
- Footer stacks vertically on mobile: `flex-col gap-2 lg:flex-row`

**Step 5: Verify at all breakpoints**

Test at: 375px, 768px, 1024px, 1440px. Check:
- No horizontal scroll at any width
- Terminal readable at mobile widths
- Features stack cleanly on mobile
- Install command doesn't overflow

**Step 6: Commit**

```bash
git add landing/components/
git commit -m "💄 style(landing): add responsive breakpoints for mobile and tablet"
```

---

### Task 10: Accessibility and prefers-reduced-motion

**Files:**
- Modify: `landing/components/toph-demo.tsx`
- Modify: `landing/components/hero.tsx`
- Modify: `landing/components/fade-in-on-scroll.tsx` (if separate)
- Modify: `landing/app/globals.css`

**Step 1: Add reduced-motion support**

Create a hook `landing/hooks/use-reduced-motion.ts`:

```typescript
"use client";

import { useEffect, useState } from "react";

export function useReducedMotion() {
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    setReduced(mq.matches);
    const handler = (e: MediaQueryListEvent) => setReduced(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  return reduced;
}
```

**Step 2: Apply to components**

- `TophDemo`: If reduced motion, freeze at initial state (no interval cycling)
- `Hero`: If reduced motion, hide AnimatedGridPattern and BorderBeam
- `FadeInOnScroll`: If reduced motion, render children immediately visible (no opacity transition)
- ShimmerButton: magicui may handle this internally; verify

**Step 3: Add CSS fallback**

In `globals.css`:
```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

**Step 4: Verify focus states**

Tab through the page and verify:
- Install command copy button has visible focus ring
- GitHub ShimmerButton has visible focus ring
- Footer GitHub link has visible focus ring
- Tab order is logical: left column → right column → VP2

**Step 5: Add aria attributes**

- Terminal mockup container: `aria-label="toph terminal dashboard showing session monitoring"`
- Copy button: already has `aria-label`
- GitHub links: `aria-label="View toph on GitHub"`

**Step 6: Commit**

```bash
git add landing/
git commit -m "🔒️ fix(landing): add accessibility support and prefers-reduced-motion"
```

---

### Task 11: Metadata, OG image, and final polish

**Files:**
- Modify: `landing/app/layout.tsx`
- Create: `landing/public/og.png` (placeholder — real OG image comes from a screenshot later)

**Step 1: Set metadata**

In `landing/app/layout.tsx`:

```typescript
export const metadata: Metadata = {
  title: "toph — btop for AI agents",
  description: "A terminal dashboard for AI coding agents. See what your agents are doing. Real-time activity feed, token tracking, cost estimation. Zero config.",
  keywords: ["terminal", "dashboard", "AI", "coding agents", "Claude Code", "btop", "TUI", "Go", "Bubble Tea"],
  authors: [{ name: "pattynextdoor" }],
  openGraph: {
    title: "toph — btop for AI agents",
    description: "A terminal dashboard for AI coding agents. See what your agents are doing.",
    type: "website",
    url: "https://toph.dev",
    images: [{ url: "/og.png", width: 1200, height: 630 }],
  },
  twitter: {
    card: "summary_large_image",
    title: "toph — btop for AI agents",
    description: "A terminal dashboard for AI coding agents.",
    images: ["/og.png"],
  },
};
```

**Step 2: Create placeholder OG image**

For now, create a simple placeholder. The real OG image will be a screenshot of the animated terminal at its most appealing state.

**Step 3: Build and verify static export**

```bash
cd /Users/patty/dev/toph/landing
npm run build
```

Expected: Static export to `landing/out/` directory. No errors. Check output size is <500KB (excluding node_modules).

**Step 4: Run the pre-delivery checklist**

From the design doc, verify:
- [ ] No emojis used as icons (use Lucide SVG)
- [ ] `cursor-pointer` on all clickable elements
- [ ] Hover states with smooth transitions (200ms ease-out)
- [ ] Text contrast 4.5:1+ verified
- [ ] Focus states visible for keyboard navigation
- [ ] `prefers-reduced-motion` respected
- [ ] Responsive at 375px, 768px, 1024px, 1440px
- [ ] No horizontal scroll on mobile
- [ ] Terminal mockup readable at mobile widths
- [ ] `<meta>` description set

**Step 5: Commit**

```bash
cd /Users/patty/dev/toph
git add landing/
git commit -m "✨ feat(landing): add metadata, OG tags, and final polish"
```

---

## Task Dependency Graph

```
Task 1 (scaffold)
  └── Task 2 (shadcn + magicui)
       └── Task 3 (global styles)
            ├── Task 4 (install-command)
            ├── Task 5 (toph-demo) ←── most complex, can be done in parallel with 4
            └── Task 6 (hero) ←── depends on 4 + 5
                 └── Task 7 (VP2 components)
                      └── Task 8 (wire up page)
                           ├── Task 9 (responsive)
                           ├── Task 10 (accessibility)
                           └── Task 11 (metadata + polish)
```

Tasks 4 and 5 can be built in parallel. Tasks 9, 10, 11 can be done in parallel after Task 8.
