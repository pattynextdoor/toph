# toph Landing Page Design

## Brand Voice

**Confident & minimal.** Short declarative sentences. No exclamation marks. The product speaks through the animated terminal mockup; the copy stays out of the way. Generous whitespace. The page could be printed on a single sheet and still work.

## Tech Stack

- **Next.js** (static export → deploy to GitHub Pages or Vercel)
- **shadcn/ui** + **Tailwind CSS** (base styling)
- **magicui** (cherry-picked components: Terminal, AnimatedGridPattern, NumberTicker, BorderBeam, ShimmerButton)
- **TypeScript**

Project lives at `web/` inside the toph repo.

## Page Structure

Two viewports. Split-screen hero (Design C) → centered features (Design B). The layout shift on scroll feels like turning a page.

---

### Viewport 1 — Split-Screen Hero (100vh)

**Background:** `#09090b` (Tailwind `zinc-950`).

**Layout: 38% left column / 62% right column**

The split creates asymmetric tension — the eye is pulled toward the terminal. No other CLI tool landing page uses this layout.

#### Left Column (38%, vertically centered, padding: 0 48px)

1. **Nav (absolute, top: 32px):** `toph` in Geist Mono 14px `zinc-500` + ShimmerButton "GitHub" (small, right-aligned within column)

2. **Headline:** "btop for AI agents." — Geist Sans, 56-64px, weight 700, `zinc-50`, line-height 1.0, letter-spacing -0.03em. Two lines: "btop" / "for AI agents."

3. **Subtext (margin-top: 24px):** "See what your agents are doing. Real-time terminal dashboard." — Geist Sans 18px, `zinc-500`, line-height 1.6, max-width 28ch.

4. **Install command (margin-top: 40px):**
   ```
   $ brew install pattynextdoor/tap/toph
   ```
   Geist Mono 13px, `zinc-300` on `zinc-900/50` bg, `1px solid zinc-800` border, 6px radius. Copy icon in `zinc-500`. `$` prompt character in `zinc-600`.

5. **Scroll hint (absolute, bottom: 32px):** "↓ scroll" in Geist Mono 11px, `zinc-700`. Disappears after first scroll.

#### Right Column (62%, vertically centered)

- **Background:** `#0d0d10` (slightly lighter than page bg), `border-left: 1px solid rgba(255,255,255,0.04)`
- **AnimatedGridPattern** filling the entire right column at ~35% opacity (stronger than page bg because it's contained in the terminal zone)
- **Animated terminal mockup** centered vertically, `margin: 0 48px`, `width: calc(100% - 96px)`
- `BorderBeam` on terminal chrome, slow 4s accent-blue trace
- Terminal aspect ratio: ~1.55:1
- Interior: 5-panel layout (Sessions 22% / Detail 36% / Activity Feed 42%, with Metrics + Tools below)
- All animation choreography same as before (events trickle, context meter creeps, amber pulses, NumberTicker counts)

#### VP1 Proportions

```
LEFT (38%)                    RIGHT (62%)
┌─────────────────────┐      ┌──────────────────────────────┐
│ toph    [GitHub] nav │      │                              │
│                     │      │  ░░ AnimatedGridPattern ░░   │
│                     │      │                              │
│ btop                │      │  ┌────────────────────────┐  │
│ for AI agents.      │      │  │ ● ● ● toph             │  │
│                     │      │  │ ┌─────┬──────┬────────┐│  │
│ See what your       │      │  │ │SESS │DETAIL│ACTIVITY││  │
│ agents are doing.   │      │  │ │● api│tokens│▸ Edit  ││  │
│ Real-time terminal  │      │  │ │◐ auth│cost │▸ Bash  ││  │
│ dashboard.          │      │  │ │○ docs│ctx% │▸ Read  ││  │
│                     │      │  │ ├─────┴──────┴────────┤│  │
│ $ brew install      │      │  │ │METRICS   │TOOLS     ││  │
│   .../toph  [copy]  │      │  │ └──────────┴──────────┘│  │
│                     │      │  └────────────────────────┘  │
│ ↓ scroll            │      │                              │
└─────────────────────┘      └──────────────────────────────┘
```

---

### Viewport 2 — Centered Features + Footer (100vh)

The layout breaks from the split and goes **full-width centered** — the contrast between VP1's rigid split and VP2's centered simplicity makes the scroll feel like turning a page.

**Positioning strip (full-width, top of VP2):**

Full-width band with `border-top: 1px solid rgba(255,255,255,0.03)` and `border-bottom: same`. Background: `#0d0d10`. Padding: 28px 64px.

> Not a macOS app. Not a web dashboard. Not a tmux hack. **A real terminal dashboard.**

Geist Mono 13px, letter-spacing 0.04em. "Not a..." in `zinc-600`. "A real terminal dashboard." in `zinc-400` (brighter). Centered.

**Features (3 columns, ruled grid, margin-top: 80px):**

Three columns separated by 1px `zinc-800` rules (CSS grid with `gap: 1px`, `background: zinc-800` on the grid, `background: zinc-950` on the cells). Each cell: 40px 36px padding.

| Bold word | Description |
|-----------|-------------|
| **Zero config** | Just run it. toph finds your Claude Code sessions automatically. |
| **Live** | Real-time activity feed, token tracking, cost estimation. 30fps. |
| **Beautiful** | Screenshot-worthy on first launch. Dark theme, smooth animations. |

- Bold word: Geist Sans, 22px, weight 600, `zinc-50`
- Description: Geist Sans, 15px, weight 400, `zinc-600`, line-height 1.65
- Max-width: 960px centered

**Footer (absolute bottom, full-width):**

Split footer: `toph` left, `GitHub · MIT · Built with Go + Bubble Tea` right. Geist Mono 12px, `zinc-700`. Border-top: 1px `zinc-800`. Padding: 24px 64px.

```
VP2:
┌──────────────────────────────────────────────────────┐
│ Not a macOS app. Not a web dashboard. Not a tmux     │
│ hack. A real terminal dashboard.                     │
│──────────────────────────────────────────────────────│
│                                                      │
│  Zero config    │  Live           │  Beautiful       │
│                 │                 │                  │
│  Just run it.   │  Real-time      │  Screenshot-     │
│  toph finds     │  activity feed, │  worthy on       │
│  your Claude    │  token tracking,│  first launch.   │
│  Code sessions  │  cost estim-    │  Dark theme,     │
│  automatically. │  ation. 30fps.  │  smooth anims.   │
│                                                      │
│                                                      │
│──────────────────────────────────────────────────────│
│ toph                 GitHub · MIT · Go + Bubble Tea  │
└──────────────────────────────────────────────────────┘
```

---

## Color Palette

| Token | Hex | Tailwind | Use |
|-------|-----|----------|-----|
| Background | `#09090b` | `zinc-950` | Page background |
| Headline | `#fafafa` | `zinc-50` | Name, feature bold words |
| Body | `#a1a1aa` | `zinc-400` | Feature descriptions |
| Muted | `#71717a` | `zinc-500` | Subheading, comparison, footer |
| Dim | `#52525b` | `zinc-600` | Footer links |
| Surface | `#18181b` | `zinc-900` | Code block background |
| Terminal active | `#87D787` | — | Green status, file writes |
| Terminal waiting | `#FFD787` | — | Amber pulsing status |
| Terminal tool use | `#87D7D7` | — | Cyan events |
| Terminal error | `#FF8787` | — | Red status |
| Terminal idle | `#6C6C6C` | — | Dim sessions |
| Border beam | `#87AFFF` | — | Animated terminal border |

## Typography

- **Headline font:** Inter or Geist Sans — clean geometric, available via `next/font`
- **Body font:** Same family, lighter weight
- **Code/terminal:** Geist Mono or JetBrains Mono
- No serif fonts. The entire page is sans-serif + monospace.

## Animated Terminal Mockup — Implementation Notes

The terminal interior is a single React component (`<TophDemo />`) with:

```typescript
// Choreographed state machine
const sessions = [
  { name: "api-server", status: "active", branch: "feat/oauth" },
  { name: "auth-flow", status: "waiting", branch: "fix/session" },
  { name: "docs", status: "idle", branch: "main" },
];

const events = [
  { tool: "Edit", file: "src/auth.ts", color: "cyan", delay: 0 },
  { tool: "Bash", file: "npm test", color: "green", delay: 2500 },
  { tool: "Read", file: "package.json", color: "dim", delay: 5000 },
  { tool: "Glob", file: "src/**/*.ts", color: "cyan", delay: 7500 },
  { tool: "Edit", file: "src/middleware.ts", color: "cyan", delay: 10000 },
  // ... cycles every ~45s
];
```

- Events append to a scrolling list with `transition: transform 300ms ease-out`
- Context meter uses `transition: width 2s cubic-bezier(0.34, 1.56, 0.64, 1)` (spring-like)
- Token `NumberTicker` increments smoothly
- Amber session pulses via CSS `@keyframes pulse { 0%, 100% { opacity: 1 } 50% { opacity: 0.5 } }` at 2s interval
- All text in the terminal uses monospace font

The mockup should look close enough to real toph output that someone familiar with terminal tools recognizes it as authentic — not a marketing illustration, but a product screenshot that happens to be animated.

## Responsive Behavior

- **Desktop (>1024px):** VP1 split-screen (38/62). Terminal shows full 5-panel layout. VP2 three-column ruled grid.
- **Tablet (768-1024px):** VP1 split collapses to stacked — text above, terminal below (centered, ~90% width). VP2 three columns, tighter gaps.
- **Mobile (<768px):** VP1 fully stacked — text centered, then terminal at ~95% width with simplified 2-panel view (Sessions + Activity only). VP2 features stacked vertically. Install command wraps gracefully.

The split-screen is a desktop-first editorial choice. Below 900px it degrades to a centered stack. The transition must feel intentional, not broken.

## Project Structure

```
web/
├── app/
│   ├── layout.tsx          # Root layout, fonts, metadata
│   ├── page.tsx            # Single page (hero + features)
│   └── globals.css         # Tailwind imports
├── components/
│   ├── hero.tsx            # Split-screen hero (38/62 layout)
│   ├── toph-demo.tsx       # Animated terminal interior (choreographed states)
│   ├── positioning.tsx     # Full-width positioning strip
│   ├── features.tsx        # Ruled 3-column feature grid
│   ├── install-command.tsx  # Code block with copy button
│   └── footer.tsx          # Split footer (logo left, links right)
├── lib/
│   └── utils.ts            # shadcn utils
├── next.config.ts          # Static export config
├── tailwind.config.ts
├── tsconfig.json
├── package.json
└── components.json         # shadcn config
```

## What This Page Does NOT Have

- No navigation bar (one page, nothing to navigate to)
- No dark/light toggle (dark only, always)
- No analytics (add later if needed)
- No cookie banner
- No testimonials (no users yet)
- No pricing (it's free/OSS)
- No changelog or blog
- No "star us on GitHub" badge (the ShimmerButton to GitHub is enough)

## Accessibility (from UI/UX Pro Max audit)

- **Contrast:** All text must meet 4.5:1 ratio against `#09090b` background. `zinc-400` (#a1a1aa) on `zinc-950` = 6.3:1. `zinc-500` (#71717a) on `zinc-950` = 4.6:1. Both pass.
- **Focus states:** Visible focus rings on install command block, copy button, GitHub button. Use `ring-2 ring-offset-2 ring-offset-zinc-950 ring-zinc-400`.
- **`prefers-reduced-motion`:** Disable all animations (terminal cycling, BorderBeam, AnimatedGridPattern, ShimmerButton shimmer). Show static terminal screenshot instead.
- **`aria-label`:** Copy button ("Copy install command"), GitHub button ("View toph on GitHub").
- **Tab order:** Install command → Copy button → GitHub CTA → Footer links.
- **No color as sole indicator:** Status icons in terminal use shapes (●, ◐, ○, ✕) alongside colors.
- **Alt text:** Terminal mockup should have `aria-label="toph terminal dashboard showing session monitoring"`.

## Animation Budget

| Element | Type | Duration | `prefers-reduced-motion` behavior |
|---------|------|----------|-----------------------------------|
| Terminal content | Scripted state cycle | 45s loop | Show static screenshot |
| BorderBeam | Border glow trace | Continuous | Hidden |
| AnimatedGridPattern | Subtle bg motion | Continuous | Hidden |
| Viewport 2 fade-in | Opacity transition | 200ms | Instant (no transition) |
| ShimmerButton | Light sweep | 3s loop | Static button |
| Copy tooltip | "Copied!" feedback | 1.5s | Still shown (functional) |

**Rule:** Max 1-2 animated elements per viewport (UX Pro Max guideline). Viewport 1 has the terminal + border as primary. Grid pattern is subtle enough to not count as a focal animation.

## Design System Reference

Full design system persisted at:
- `design-system/toph/MASTER.md` — Global source of truth (colors, typography, spacing, anti-patterns)
- `design-system/toph/pages/landing.md` — Landing page overrides (layout, animation budget, component config)

## Pre-Delivery Checklist

- [ ] No emojis used as icons (use Lucide SVG)
- [ ] `cursor-pointer` on all clickable elements
- [ ] Hover states with smooth transitions (200ms ease-out)
- [ ] Text contrast 4.5:1+ verified for all text/bg combinations
- [ ] Focus states visible for keyboard navigation
- [ ] `prefers-reduced-motion` respected for ALL animations
- [ ] Responsive at 375px, 768px, 1024px, 1440px
- [ ] No horizontal scroll on mobile
- [ ] Terminal mockup readable at mobile widths (simplified 2-panel view)
- [ ] Total page weight <500KB
- [ ] Lighthouse >95 all categories
- [ ] OG image set (terminal screenshot)
- [ ] `<meta>` description set

## Success Criteria

- Page loads in <1s on 3G
- Lighthouse score >95 on all categories
- Someone screenshots the page and shares it
- The animated terminal makes someone say "I need to install this"
- Total page weight <500KB (no videos, no large images)
