"use client";

import { useRef, useMemo, useCallback, useEffect, useState } from "react";
import { Canvas, useFrame, useThree } from "@react-three/fiber";
import { EffectComposer, Bloom } from "@react-three/postprocessing";
import * as THREE from "three";

// =============================================================================
// TUNING CONSTANTS
// =============================================================================
const AMBIENT_PARTICLE_COUNT = 60;
const BURST_PARTICLE_COUNT = 5;
const PARTICLE_SPEED = 1.2;
const AMBIENT_SPEED = 0.3;
const PARTICLE_SIZE = 3.0;
const AMBIENT_SIZE = 1.5;
const ARC_HEIGHT = 0.8;
const TRAIL_LENGTH = 0.15;
const BLOOM_INTENSITY = 2.0;
const BLOOM_THRESHOLD = 0.1;

// Color map for tool types
const TOOL_PARTICLE_COLORS: Record<string, string> = {
  Edit: "#87D7D7",
  Bash: "#87D787",
  Read: "#71717a",
  Write: "#87D787",
  Glob: "#87D7D7",
  Grep: "#87D7D7",
  Agent: "#D7AFFF",
  TaskCreate: "#FFD787",
};

// =============================================================================
// Particle system — manages both ambient and burst particles
// =============================================================================

interface Particle {
  x: number;
  y: number;
  progress: number; // 0 to 1 across the bridge
  speed: number;
  color: THREE.Color;
  size: number;
  opacity: number;
  isBurst: boolean;
  arcOffset: number; // randomize arc height per particle
  yOffset: number;   // vertical target variation
}

function ParticleSystem({ events }: { events: { tool: string; id: number }[] }) {
  const pointsRef = useRef<THREE.Points>(null);
  const particlesRef = useRef<Particle[]>([]);
  const lastEventIdRef = useRef(-1);
  const { viewport } = useThree();

  // The bridge goes from ~-3 to ~3 in NDC-ish coordinates
  const bridgeLeft = -viewport.width / 2 * 0.15;
  const bridgeRight = viewport.width / 2 * 0.15;
  const bridgeWidth = bridgeRight - bridgeLeft;

  // Initialize ambient particles
  useEffect(() => {
    const ambients: Particle[] = [];
    for (let i = 0; i < AMBIENT_PARTICLE_COUNT; i++) {
      ambients.push({
        x: 0,
        y: 0,
        progress: Math.random(),
        speed: AMBIENT_SPEED * (0.5 + Math.random() * 1.0),
        color: new THREE.Color("#87AFFF").multiplyScalar(0.4),
        size: AMBIENT_SIZE * (0.5 + Math.random()),
        opacity: 0.15 + Math.random() * 0.15,
        isBurst: false,
        arcOffset: (Math.random() - 0.5) * 2,
        yOffset: (Math.random() - 0.5) * 3,
      });
    }
    particlesRef.current = ambients;
  }, []);

  // Spawn burst particles when new events arrive
  useEffect(() => {
    if (events.length === 0) return;
    const latest = events[events.length - 1];
    if (latest.id <= lastEventIdRef.current) return;
    lastEventIdRef.current = latest.id;

    const toolColor = TOOL_PARTICLE_COLORS[latest.tool] || "#87AFFF";

    for (let i = 0; i < BURST_PARTICLE_COUNT; i++) {
      particlesRef.current.push({
        x: 0,
        y: 0,
        progress: 0,
        speed: PARTICLE_SPEED * (0.8 + Math.random() * 0.4),
        color: new THREE.Color(toolColor),
        size: PARTICLE_SIZE * (0.8 + Math.random() * 0.4),
        opacity: 0.8 + Math.random() * 0.2,
        isBurst: true,
        arcOffset: (Math.random() - 0.5) * 1.5,
        yOffset: (Math.random() - 0.5) * 2,
      });
    }
  }, [events]);

  // Geometry buffers
  const maxParticles = AMBIENT_PARTICLE_COUNT + 200; // headroom for bursts
  const positions = useMemo(() => new Float32Array(maxParticles * 3), [maxParticles]);
  const colors = useMemo(() => new Float32Array(maxParticles * 3), [maxParticles]);
  const sizes = useMemo(() => new Float32Array(maxParticles), [maxParticles]);

  useFrame((_, delta) => {
    if (!pointsRef.current) return;

    const particles = particlesRef.current;
    let aliveCount = 0;

    for (let i = particles.length - 1; i >= 0; i--) {
      const p = particles[i];
      p.progress += p.speed * delta;

      if (p.isBurst && p.progress > 1) {
        // Remove dead burst particles
        particles.splice(i, 1);
        continue;
      }

      // Wrap ambient particles
      if (!p.isBurst && p.progress > 1) {
        p.progress = 0;
        p.yOffset = (Math.random() - 0.5) * 3;
        p.arcOffset = (Math.random() - 0.5) * 2;
      }

      // Calculate position along the arc
      const t = p.progress;
      const x = bridgeLeft + t * bridgeWidth;
      const arcY = Math.sin(t * Math.PI) * ARC_HEIGHT * (1 + p.arcOffset * 0.3);
      const y = p.yOffset + arcY;

      // Fade in/out
      let alpha = p.opacity;
      if (t < 0.1) alpha *= t / 0.1;
      if (t > 0.85) alpha *= (1 - t) / 0.15;

      positions[aliveCount * 3] = x;
      positions[aliveCount * 3 + 1] = y;
      positions[aliveCount * 3 + 2] = 0;
      colors[aliveCount * 3] = p.color.r * alpha;
      colors[aliveCount * 3 + 1] = p.color.g * alpha;
      colors[aliveCount * 3 + 2] = p.color.b * alpha;
      sizes[aliveCount] = p.size;
      aliveCount++;
    }

    // Zero out unused slots
    for (let i = aliveCount; i < maxParticles; i++) {
      positions[i * 3] = 0;
      positions[i * 3 + 1] = 0;
      positions[i * 3 + 2] = -100; // behind camera
      sizes[i] = 0;
    }

    const geo = pointsRef.current.geometry;
    geo.attributes.position.needsUpdate = true;
    geo.attributes.color.needsUpdate = true;
    geo.attributes.size.needsUpdate = true;
    geo.setDrawRange(0, aliveCount);
  });

  return (
    <points ref={pointsRef}>
      <bufferGeometry>
        <bufferAttribute attach="attributes-position" args={[positions, 3]} />
        <bufferAttribute attach="attributes-color" args={[colors, 3]} />
        <bufferAttribute attach="attributes-size" args={[sizes, 1]} />
      </bufferGeometry>
      <pointsMaterial
        size={PARTICLE_SIZE}
        vertexColors
        transparent
        opacity={1}
        sizeAttenuation={false}
        depthWrite={false}
        blending={THREE.AdditiveBlending}
      />
    </points>
  );
}

// =============================================================================
// Scene wrapper
// =============================================================================
function BridgeScene({ events }: { events: { tool: string; id: number }[] }) {
  return (
    <>
      <ParticleSystem events={events} />
      <EffectComposer>
        <Bloom
          intensity={BLOOM_INTENSITY}
          luminanceThreshold={BLOOM_THRESHOLD}
          luminanceSmoothing={0.9}
        />
      </EffectComposer>
    </>
  );
}

// =============================================================================
// Exported overlay component
// =============================================================================
export function ParticleBridge({ events }: { events: { tool: string; id: number }[] }) {
  const [reducedMotion, setReducedMotion] = useState(false);

  useEffect(() => {
    setReducedMotion(window.matchMedia("(prefers-reduced-motion: reduce)").matches);
  }, []);

  if (reducedMotion) return null;

  return (
    <div className="absolute inset-0 z-10 pointer-events-none">
      <Canvas
        camera={{ position: [0, 0, 10], fov: 50 }}
        dpr={[1, 1.5]}
        gl={{ antialias: false, alpha: true }}
        style={{ background: "transparent" }}
      >
        <BridgeScene events={events} />
      </Canvas>
    </div>
  );
}
