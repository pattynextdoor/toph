"use client";

import { useEffect, useRef, useState } from "react";

// =============================================================================
// TUNING CONSTANTS
// =============================================================================
const TRUNK_WIDTH = 2.5;
const BRANCH_WIDTH = 1.5;
const ENERGY_SPEED = 0.3;
const PULSE_LENGTH = 0.22;
const TRUNK_COLOR = "#87AFFF";
const PANEL_COLORS: Record<string, string> = {
  sessions: "#87D787",
  activity: "#87D7D7",
  detail:   "#FFD787",
  metrics:  "#87AFFF",
  tools:    "#87D7D7",
};
const BASE_GLOW = 0.15;
const PULSE_GLOW = 0.65;
const GLOW_BLUR = 10;
const BURST_BRIGHTNESS = 0.8;
const BURST_DECAY = 0.97;
const PANEL_IDS = ["sessions", "activity", "detail", "metrics", "tools"];

interface BranchTarget {
  x: number;
  y: number;
  color: string;
  id: string;
}

// =============================================================================
// Component
// =============================================================================
export function ParticleBridge({ events }: { events: { tool: string; id: number }[] }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const animFrameRef = useRef<number>(0);
  const phaseRef = useRef(0);
  const burstRef = useRef(0);
  const lastEventIdRef = useRef(-1);
  const targetsRef = useRef<BranchTarget[]>([]);
  const sourceRef = useRef<{ x: number; y: number }>({ x: 0, y: 0 });
  const [reducedMotion, setReducedMotion] = useState(false);

  useEffect(() => {
    setReducedMotion(window.matchMedia("(prefers-reduced-motion: reduce)").matches);
  }, []);

  // Trigger burst on new events
  useEffect(() => {
    if (events.length === 0) return;
    const latest = events[events.length - 1];
    if (latest.id <= lastEventIdRef.current) return;
    lastEventIdRef.current = latest.id;
    burstRef.current = BURST_BRIGHTNESS;
  }, [events]);

  useEffect(() => {
    if (reducedMotion) return;
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // Canvas is inside: <div class="absolute..."> inside <div class="hidden lg:block"> inside <div class="relative">
    // We need the "relative" div which contains both the canvas wrapper and the data-panel/data-source elements
    const containerEl = canvas.closest("[data-bridge-container]") || canvas.parentElement?.parentElement?.parentElement;
    if (!containerEl) return;

    const computePositions = () => {
      const containerRect = containerEl.getBoundingClientRect();

      // Source: right edge center of the JSONL pane
      const sourceEl = containerEl.querySelector("[data-source='jsonl']");
      if (sourceEl) {
        const r = sourceEl.getBoundingClientRect();
        sourceRef.current = {
          x: r.right - containerRect.left,
          y: r.top + r.height / 2 - containerRect.top,
        };
      }

      // Targets: left edge center of each panel
      const targets: BranchTarget[] = [];
      for (const id of PANEL_IDS) {
        const el = containerEl.querySelector(`[data-panel='${id}']`);
        if (el) {
          const r = el.getBoundingClientRect();
          targets.push({
            x: r.left - containerRect.left,
            y: r.top + r.height / 2 - containerRect.top,
            color: PANEL_COLORS[id] || TRUNK_COLOR,
            id,
          });
        }
      }
      targetsRef.current = targets;
    };

    const resizeCanvas = () => {
      const rect = containerEl.getBoundingClientRect();
      const dpr = window.devicePixelRatio;
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      canvas.style.width = `${rect.width}px`;
      canvas.style.height = `${rect.height}px`;
      computePositions();
    };

    resizeCanvas();
    window.addEventListener("resize", resizeCanvas);
    // Recompute after layout settles (content might not be rendered yet)
    const recomputeTimer = setTimeout(() => {
      resizeCanvas();
      computePositions();
    }, 300);

    const animate = () => {
      const w = canvas.width / window.devicePixelRatio;
      const h = canvas.height / window.devicePixelRatio;
      ctx.clearRect(0, 0, w, h);

      phaseRef.current = (phaseRef.current + ENERGY_SPEED * 0.016) % 1;
      burstRef.current *= BURST_DECAY;
      if (burstRef.current < 0.01) burstRef.current = 0;

      const phase = phaseRef.current;
      const burstMult = 1 + burstRef.current * 3;
      const src = sourceRef.current;
      const targets = targetsRef.current;

      if (targets.length === 0 || src.x === 0) {
        animFrameRef.current = requestAnimationFrame(animate);
        return;
      }

      // Split point: midway between source and nearest target
      const avgTargetX = targets.reduce((s, t) => s + t.x, 0) / targets.length;
      const splitX = src.x + (avgTargetX - src.x) * 0.45;
      const splitY = src.y;

      // Draw trunk: source → split
      drawEnergy(ctx, src.x, src.y, splitX, splitY, TRUNK_COLOR, TRUNK_WIDTH, phase, burstMult);

      // Draw branches: split → each panel
      for (let i = 0; i < targets.length; i++) {
        const t = targets[i];
        const branchPhase = (phase + i * 0.07) % 1;
        drawBranch(ctx, splitX, splitY, t.x, t.y, t.color, BRANCH_WIDTH, branchPhase, burstMult);
      }

      // Split node glow
      const nodeAlpha = Math.min((BASE_GLOW * 4 + burstRef.current) * burstMult, 1);
      const grad = ctx.createRadialGradient(splitX, splitY, 0, splitX, splitY, 8);
      grad.addColorStop(0, hexAlpha(TRUNK_COLOR, nodeAlpha));
      grad.addColorStop(0.6, hexAlpha(TRUNK_COLOR, nodeAlpha * 0.3));
      grad.addColorStop(1, hexAlpha(TRUNK_COLOR, 0));
      ctx.fillStyle = grad;
      ctx.fillRect(splitX - 10, splitY - 10, 20, 20);

      animFrameRef.current = requestAnimationFrame(animate);
    };

    animFrameRef.current = requestAnimationFrame(animate);

    return () => {
      cancelAnimationFrame(animFrameRef.current);
      clearTimeout(recomputeTimer);
      window.removeEventListener("resize", resizeCanvas);
    };
  }, [reducedMotion]);

  if (reducedMotion) return null;

  return (
    <div className="absolute inset-0 z-10 pointer-events-none overflow-hidden">
      <canvas ref={canvasRef} className="w-full h-full" />
    </div>
  );
}

// =============================================================================
// Drawing
// =============================================================================

function drawEnergy(
  ctx: CanvasRenderingContext2D,
  x1: number, y1: number, x2: number, y2: number,
  color: string, width: number, phase: number, burstMult: number,
) {
  // Base dim line
  ctx.beginPath();
  ctx.moveTo(x1, y1);
  ctx.lineTo(x2, y2);
  ctx.strokeStyle = hexAlpha(color, BASE_GLOW * burstMult);
  ctx.lineWidth = width;
  ctx.stroke();

  // Glow
  ctx.save();
  ctx.filter = `blur(${GLOW_BLUR}px)`;
  ctx.beginPath();
  ctx.moveTo(x1, y1);
  ctx.lineTo(x2, y2);
  ctx.strokeStyle = hexAlpha(color, BASE_GLOW * 0.5 * burstMult);
  ctx.lineWidth = width * 3;
  ctx.stroke();
  ctx.restore();

  // Pulse
  drawPulseLine(ctx, x1, y1, x2, y2, color, width, phase, burstMult);
}

function drawBranch(
  ctx: CanvasRenderingContext2D,
  x1: number, y1: number, x2: number, y2: number,
  color: string, width: number, phase: number, burstMult: number,
) {
  const cpX = x1 + (x2 - x1) * 0.4;
  const cpY = y1 + (y2 - y1) * 0.2;

  // Base dim curve
  ctx.beginPath();
  ctx.moveTo(x1, y1);
  ctx.quadraticCurveTo(cpX, cpY, x2, y2);
  ctx.strokeStyle = hexAlpha(color, BASE_GLOW * burstMult);
  ctx.lineWidth = width;
  ctx.stroke();

  // Glow
  ctx.save();
  ctx.filter = `blur(${GLOW_BLUR}px)`;
  ctx.beginPath();
  ctx.moveTo(x1, y1);
  ctx.quadraticCurveTo(cpX, cpY, x2, y2);
  ctx.strokeStyle = hexAlpha(color, BASE_GLOW * 0.3 * burstMult);
  ctx.lineWidth = width * 2.5;
  ctx.stroke();
  ctx.restore();

  // Pulse
  drawPulseBezier(ctx, x1, y1, cpX, cpY, x2, y2, color, width, phase, burstMult);

  // Target node glow
  const pulseAt = ((phase + PULSE_LENGTH / 2) % 1);
  if (pulseAt > 0.85) {
    const arrivalAlpha = (pulseAt - 0.85) / 0.15 * PULSE_GLOW * burstMult * 0.6;
    const g = ctx.createRadialGradient(x2, y2, 0, x2, y2, 10);
    g.addColorStop(0, hexAlpha(color, arrivalAlpha));
    g.addColorStop(1, hexAlpha(color, 0));
    ctx.fillStyle = g;
    ctx.fillRect(x2 - 12, y2 - 12, 24, 24);
  }
}

function drawPulseLine(
  ctx: CanvasRenderingContext2D,
  x1: number, y1: number, x2: number, y2: number,
  color: string, width: number, phase: number, burstMult: number,
) {
  const segs = 16;
  for (let i = 0; i < segs; i++) {
    const t0 = ((phase + PULSE_LENGTH * i / segs) % 1 + 1) % 1;
    const t1 = ((phase + PULSE_LENGTH * (i + 1) / segs) % 1 + 1) % 1;
    const brightness = Math.sin((i / segs) * Math.PI) * PULSE_GLOW * burstMult;

    ctx.save();
    ctx.filter = `blur(${GLOW_BLUR * 0.4}px)`;
    ctx.beginPath();
    ctx.moveTo(x1 + (x2 - x1) * t0, y1 + (y2 - y1) * t0);
    ctx.lineTo(x1 + (x2 - x1) * t1, y1 + (y2 - y1) * t1);
    ctx.strokeStyle = hexAlpha(color, brightness);
    ctx.lineWidth = width * 2;
    ctx.stroke();
    ctx.restore();
  }
}

function drawPulseBezier(
  ctx: CanvasRenderingContext2D,
  x1: number, y1: number, cpX: number, cpY: number, x2: number, y2: number,
  color: string, width: number, phase: number, burstMult: number,
) {
  const segs = 14;
  for (let i = 0; i < segs; i++) {
    const t0 = ((phase + PULSE_LENGTH * i / segs) % 1 + 1) % 1;
    const t1 = ((phase + PULSE_LENGTH * (i + 1) / segs) % 1 + 1) % 1;
    const brightness = Math.sin((i / segs) * Math.PI) * PULSE_GLOW * burstMult;

    const px0 = bz(x1, cpX, x2, t0), py0 = bz(y1, cpY, y2, t0);
    const px1 = bz(x1, cpX, x2, t1), py1 = bz(y1, cpY, y2, t1);

    ctx.save();
    ctx.filter = `blur(${GLOW_BLUR * 0.4}px)`;
    ctx.beginPath();
    ctx.moveTo(px0, py0);
    ctx.lineTo(px1, py1);
    ctx.strokeStyle = hexAlpha(color, brightness);
    ctx.lineWidth = width * 1.8;
    ctx.stroke();
    ctx.restore();
  }
}

function bz(p0: number, cp: number, p1: number, t: number): number {
  const mt = 1 - t;
  return mt * mt * p0 + 2 * mt * t * cp + t * t * p1;
}

function hexAlpha(hex: string, alpha: number): string {
  return hex + Math.round(Math.max(0, Math.min(1, alpha)) * 255).toString(16).padStart(2, "0");
}
