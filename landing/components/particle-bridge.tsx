"use client";

import { useEffect, useRef, useState, useCallback } from "react";

// =============================================================================
// TUNING CONSTANTS
// =============================================================================
const AMBIENT_COUNT = 25;
const BURST_COUNT = 4;
const PARTICLE_SPEED = 2.5; // pixels per frame
const AMBIENT_SPEED = 0.8;
const PARTICLE_RADIUS = 2.5;
const AMBIENT_RADIUS = 1.2;
const TRAIL_ALPHA = 0.92; // lower = longer trails (0-1)
const GLOW_SIZE = 12;

const TOOL_COLORS: Record<string, string> = {
  Edit: "#87D7D7",
  Bash: "#87D787",
  Read: "#71717a",
  Write: "#87D787",
  Glob: "#87D7D7",
  Grep: "#87D7D7",
  Agent: "#D7AFFF",
  TaskCreate: "#FFD787",
};

interface Particle {
  x: number;
  y: number;
  targetX: number;
  startX: number;
  startY: number;
  targetY: number;
  progress: number;
  speed: number;
  color: string;
  radius: number;
  opacity: number;
  isBurst: boolean;
}

// =============================================================================
// Component
// =============================================================================
export function ParticleBridge({ events }: { events: { tool: string; id: number }[] }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const particlesRef = useRef<Particle[]>([]);
  const lastEventIdRef = useRef(-1);
  const animFrameRef = useRef<number>(0);
  const [reducedMotion, setReducedMotion] = useState(false);

  // Get the bridge gap coordinates (called on resize)
  const getBridgeBounds = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return null;
    const rect = canvas.getBoundingClientRect();
    return {
      width: rect.width,
      height: rect.height,
      // Particles flow from left ~15% to right ~85% of the canvas
      startX: rect.width * 0.05,
      endX: rect.width * 0.95,
    };
  }, []);

  // Spawn ambient particles
  const spawnAmbients = useCallback(() => {
    const bounds = getBridgeBounds();
    if (!bounds) return;

    const ambients: Particle[] = [];
    for (let i = 0; i < AMBIENT_COUNT; i++) {
      const startX = bounds.startX;
      const endX = bounds.endX;
      const y = bounds.height * (0.15 + Math.random() * 0.7);
      const targetY = y + (Math.random() - 0.5) * 40;

      ambients.push({
        x: startX + Math.random() * (endX - startX),
        y,
        startX,
        startY: y,
        targetX: endX,
        targetY,
        progress: Math.random(),
        speed: AMBIENT_SPEED * (0.5 + Math.random() * 1.0),
        color: "#87AFFF",
        radius: AMBIENT_RADIUS * (0.6 + Math.random() * 0.8),
        opacity: 0.15 + Math.random() * 0.1,
        isBurst: false,
      });
    }
    particlesRef.current = ambients;
  }, [getBridgeBounds]);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    if (mq.matches) {
      setReducedMotion(true);
      return;
    }
    spawnAmbients();
  }, [spawnAmbients]);

  // Spawn burst particles when new events arrive
  useEffect(() => {
    if (reducedMotion || events.length === 0) return;
    const latest = events[events.length - 1];
    if (latest.id <= lastEventIdRef.current) return;
    lastEventIdRef.current = latest.id;

    const bounds = getBridgeBounds();
    if (!bounds) return;

    const color = TOOL_COLORS[latest.tool] || "#87AFFF";

    for (let i = 0; i < BURST_COUNT; i++) {
      const startY = bounds.height * (0.2 + Math.random() * 0.6);
      const targetY = bounds.height * (0.15 + Math.random() * 0.7);

      particlesRef.current.push({
        x: bounds.startX,
        y: startY,
        startX: bounds.startX,
        startY,
        targetX: bounds.endX,
        targetY,
        progress: 0,
        speed: PARTICLE_SPEED * (0.7 + Math.random() * 0.6),
        color,
        radius: PARTICLE_RADIUS * (0.8 + Math.random() * 0.4),
        opacity: 0.7 + Math.random() * 0.3,
        isBurst: true,
      });
    }
  }, [events, reducedMotion, getBridgeBounds]);

  // Animation loop
  useEffect(() => {
    if (reducedMotion) return;
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const resizeCanvas = () => {
      const parent = canvas.parentElement;
      if (!parent) return;
      const rect = parent.getBoundingClientRect();
      canvas.width = rect.width * window.devicePixelRatio;
      canvas.height = rect.height * window.devicePixelRatio;
      ctx.scale(window.devicePixelRatio, window.devicePixelRatio);
      canvas.style.width = `${rect.width}px`;
      canvas.style.height = `${rect.height}px`;
    };

    resizeCanvas();
    window.addEventListener("resize", resizeCanvas);

    const animate = () => {
      const w = canvas.width / window.devicePixelRatio;
      const h = canvas.height / window.devicePixelRatio;

      // Fade previous frame for trail effect
      ctx.fillStyle = `rgba(9, 9, 11, ${TRAIL_ALPHA})`;
      ctx.fillRect(0, 0, w, h);

      const particles = particlesRef.current;

      for (let i = particles.length - 1; i >= 0; i--) {
        const p = particles[i];
        p.progress += p.speed / (p.targetX - p.startX);

        if (p.isBurst && p.progress > 1) {
          particles.splice(i, 1);
          continue;
        }

        // Wrap ambient particles
        if (!p.isBurst && p.progress > 1) {
          p.progress = 0;
          p.startY = h * (0.15 + Math.random() * 0.7);
          p.targetY = p.startY + (Math.random() - 0.5) * 40;
        }

        const t = p.progress;
        // Smooth easing
        const easedT = t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;

        // Horizontal: linear left to right
        p.x = p.startX + (p.targetX - p.startX) * easedT;

        // Vertical: interpolate with a slight arc
        const arcHeight = -30 * (p.isBurst ? 1.5 : 0.5);
        const arc = Math.sin(t * Math.PI) * arcHeight;
        p.y = p.startY + (p.targetY - p.startY) * t + arc;

        // Fade in/out
        let alpha = p.opacity;
        if (t < 0.1) alpha *= t / 0.1;
        if (t > 0.8) alpha *= (1 - t) / 0.2;

        // Draw glow
        if (p.isBurst) {
          const gradient = ctx.createRadialGradient(p.x, p.y, 0, p.x, p.y, GLOW_SIZE);
          gradient.addColorStop(0, p.color + Math.round(alpha * 80).toString(16).padStart(2, "0"));
          gradient.addColorStop(1, p.color + "00");
          ctx.fillStyle = gradient;
          ctx.fillRect(p.x - GLOW_SIZE, p.y - GLOW_SIZE, GLOW_SIZE * 2, GLOW_SIZE * 2);
        }

        // Draw core
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.radius, 0, Math.PI * 2);
        ctx.fillStyle = p.color + Math.round(alpha * 255).toString(16).padStart(2, "0");
        ctx.fill();
      }

      animFrameRef.current = requestAnimationFrame(animate);
    };

    animFrameRef.current = requestAnimationFrame(animate);

    return () => {
      cancelAnimationFrame(animFrameRef.current);
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
