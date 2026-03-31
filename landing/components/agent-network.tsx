"use client";

import { useRef, useMemo } from "react";
import { Canvas, useFrame } from "@react-three/fiber";
import { OrbitControls, Float, Html } from "@react-three/drei";
import { EffectComposer, Bloom } from "@react-three/postprocessing";
import * as THREE from "three";

// =============================================================================
// TUNING CONSTANTS — adjust these for visual refinement
// =============================================================================
const PARTICLE_COUNT = 800;
const EDGE_PARTICLE_COUNT = 120;
const NODE_BASE_RADIUS = 0.12;
const NODE_GLOW_RADIUS = 0.28;
const CAMERA_DISTANCE = 10;
const CAMERA_FOV = 40;
const AUTO_ROTATE_SPEED = 0.12;
const BLOOM_INTENSITY = 2.5;
const BLOOM_LUMINANCE_THRESHOLD = 0.1;
const BLOOM_LUMINANCE_SMOOTHING = 0.95;
const PARTICLE_OPACITY = 0.35;
const PARTICLE_SIZE = 0.025;
const EDGE_PARTICLE_SPEED = 0.3;
const BG_COLOR = "#09090b";

// Node data: spread further apart for visual breathing room
const NODES = [
  { pos: [0, 0.3, 0] as [number, number, number], color: "#87D787", label: "api-server", size: 1.3, status: "coding", detail: "14.8K tokens", icon: "●" },
  { pos: [3.8, 1.8, -1.5] as [number, number, number], color: "#FFD787", label: "auth-flow", size: 1.0, status: "waiting", detail: "permission", icon: "◐" },
  { pos: [-3.5, 1.2, 2.0] as [number, number, number], color: "#6C6C6C", label: "docs", size: 0.75, status: "idle", detail: "2m ago", icon: "○" },
  { pos: [2.5, -2.2, 2.5] as [number, number, number], color: "#87D7D7", label: "tests", size: 0.95, status: "running", detail: "npm test", icon: "●" },
  { pos: [-2.5, -1.8, -2.8] as [number, number, number], color: "#D7AFFF", label: "deploy", size: 0.9, status: "subagent", detail: "3 agents", icon: "◉" },
];

// Edges between nodes (index pairs)
const EDGES: [number, number][] = [
  [0, 1], [0, 2], [0, 3], [0, 4], [1, 3], [2, 4],
];

// =============================================================================
// Ambient Particles — floating in the background
// =============================================================================
function AmbientParticles() {
  const ref = useRef<THREE.Points>(null);

  const positions = useMemo(() => {
    const arr = new Float32Array(PARTICLE_COUNT * 3);
    for (let i = 0; i < PARTICLE_COUNT; i++) {
      arr[i * 3] = (Math.random() - 0.5) * 20;
      arr[i * 3 + 1] = (Math.random() - 0.5) * 16;
      arr[i * 3 + 2] = (Math.random() - 0.5) * 20;
    }
    return arr;
  }, []);

  const sizes = useMemo(() => {
    const arr = new Float32Array(PARTICLE_COUNT);
    for (let i = 0; i < PARTICLE_COUNT; i++) {
      arr[i] = PARTICLE_SIZE * (0.5 + Math.random() * 1.5);
    }
    return arr;
  }, []);

  useFrame(({ clock }) => {
    if (!ref.current) return;
    const t = clock.getElapsedTime() * 0.05;
    ref.current.rotation.y = t;
    ref.current.rotation.x = t * 0.3;
  });

  return (
    <points ref={ref}>
      <bufferGeometry>
        <bufferAttribute
          attach="attributes-position"
          args={[positions, 3]}
        />
        <bufferAttribute
          attach="attributes-size"
          args={[sizes, 1]}
        />
      </bufferGeometry>
      <pointsMaterial
        size={PARTICLE_SIZE}
        color="#87AFFF"
        transparent
        opacity={PARTICLE_OPACITY}
        sizeAttenuation
        depthWrite={false}
        blending={THREE.AdditiveBlending}
      />
    </points>
  );
}

// =============================================================================
// Glowing Node — core sphere + outer glow aura
// =============================================================================
function GlowNode({
  position, color, size = 1, label, status, detail, icon,
}: {
  position: [number, number, number]; color: string; size?: number;
  label: string; status: string; detail: string; icon: string;
}) {
  const meshRef = useRef<THREE.Mesh>(null);

  useFrame(({ clock }) => {
    if (!meshRef.current) return;
    const t = clock.getElapsedTime();
    meshRef.current.scale.setScalar(size * (1 + Math.sin(t * 2 + position[0]) * 0.06));
  });

  return (
    <Float speed={1.5} rotationIntensity={0} floatIntensity={0.3} floatingRange={[-0.1, 0.1]}>
      <group position={position}>
        {/* Core */}
        <mesh ref={meshRef}>
          <sphereGeometry args={[NODE_BASE_RADIUS * size, 32, 32]} />
          <meshBasicMaterial color={color} />
        </mesh>
        {/* Glow aura — tighter */}
        <mesh>
          <sphereGeometry args={[NODE_GLOW_RADIUS * size, 32, 32]} />
          <meshBasicMaterial
            color={color}
            transparent
            opacity={0.06}
            depthWrite={false}
            blending={THREE.AdditiveBlending}
          />
        </mesh>
        {/* Outer haze — tighter */}
        <mesh>
          <sphereGeometry args={[NODE_GLOW_RADIUS * size * 1.5, 16, 16]} />
          <meshBasicMaterial
            color={color}
            transparent
            opacity={0.02}
            depthWrite={false}
            blending={THREE.AdditiveBlending}
          />
        </mesh>
        {/* HTML Label */}
        <Html
          position={[0, -(NODE_BASE_RADIUS * size + 0.5), 0]}
          center
          distanceFactor={10}
          style={{ pointerEvents: "none", userSelect: "none" }}
        >
          <div className="flex flex-col items-center gap-0.5 whitespace-nowrap">
            <div className="flex items-center gap-1.5 font-mono text-[11px]">
              <span style={{ color }}>{icon}</span>
              <span className="text-zinc-200 font-medium">{label}</span>
            </div>
            <div className="font-mono text-[9px] text-zinc-500">
              {status} · {detail}
            </div>
          </div>
        </Html>
      </group>
    </Float>
  );
}

// =============================================================================
// Edge with flowing particles
// =============================================================================
function FlowingEdge({ from, to }: { from: [number, number, number]; to: [number, number, number] }) {
  const particlesRef = useRef<THREE.Points>(null);

  // Create a THREE.Line object imperatively to avoid JSX type conflicts
  const lineObj = useMemo(() => {
    const geo = new THREE.BufferGeometry();
    geo.setAttribute(
      "position",
      new THREE.BufferAttribute(new Float32Array([...from, ...to]), 3)
    );
    const mat = new THREE.LineBasicMaterial({
      color: "#87AFFF",
      transparent: true,
      opacity: 0.15,
      blending: THREE.AdditiveBlending,
    });
    return new THREE.Line(geo, mat);
  }, [from, to]);

  // Particle offsets along the edge (0-1 parameter)
  const offsets = useMemo(() => {
    const arr = new Float32Array(EDGE_PARTICLE_COUNT);
    for (let i = 0; i < EDGE_PARTICLE_COUNT; i++) {
      arr[i] = Math.random();
    }
    return arr;
  }, []);

  const particlePositions = useMemo(() => {
    return new Float32Array(EDGE_PARTICLE_COUNT * 3);
  }, []);

  const fromVec = useMemo(() => new THREE.Vector3(...from), [from]);
  const toVec = useMemo(() => new THREE.Vector3(...to), [to]);

  useFrame(({ clock }) => {
    if (!particlesRef.current) return;
    const t = clock.getElapsedTime();
    const posAttr = particlesRef.current.geometry.attributes.position;

    for (let i = 0; i < EDGE_PARTICLE_COUNT; i++) {
      const param = (offsets[i] + t * EDGE_PARTICLE_SPEED) % 1;
      const x = fromVec.x + (toVec.x - fromVec.x) * param;
      const y = fromVec.y + (toVec.y - fromVec.y) * param;
      const z = fromVec.z + (toVec.z - fromVec.z) * param;

      // Add slight perpendicular wobble
      const wobble = Math.sin(param * Math.PI * 4 + t * 3) * 0.05;
      (posAttr.array as Float32Array)[i * 3] = x + wobble;
      (posAttr.array as Float32Array)[i * 3 + 1] = y + wobble;
      (posAttr.array as Float32Array)[i * 3 + 2] = z;
    }
    posAttr.needsUpdate = true;
  });

  return (
    <>
      {/* Thin connecting line */}
      <primitive object={lineObj} />
      {/* Flowing particles along edge */}
      <points ref={particlesRef}>
        <bufferGeometry>
          <bufferAttribute
            attach="attributes-position"
            args={[particlePositions, 3]}
          />
        </bufferGeometry>
        <pointsMaterial
          size={0.04}
          color="#87AFFF"
          transparent
          opacity={0.6}
          sizeAttenuation
          depthWrite={false}
          blending={THREE.AdditiveBlending}
        />
      </points>
    </>
  );
}

// =============================================================================
// Scene
// =============================================================================
function Scene() {
  return (
    <>
      <ambientLight intensity={0.1} />

      {/* Nodes */}
      {NODES.map((node, i) => (
        <GlowNode
          key={i}
          position={node.pos}
          color={node.color}
          size={node.size}
          label={node.label}
          status={node.status}
          detail={node.detail}
          icon={node.icon}
        />
      ))}

      {/* Edges */}
      {EDGES.map(([a, b], i) => (
        <FlowingEdge key={i} from={NODES[a].pos} to={NODES[b].pos} />
      ))}

      {/* Ambient particles */}
      <AmbientParticles />

      {/* Camera controls */}
      <OrbitControls
        enableZoom={false}
        enablePan={false}
        autoRotate
        autoRotateSpeed={AUTO_ROTATE_SPEED}
        minPolarAngle={Math.PI / 3}
        maxPolarAngle={Math.PI / 1.5}
      />

      {/* Post-processing */}
      <EffectComposer>
        <Bloom
          intensity={BLOOM_INTENSITY}
          luminanceThreshold={BLOOM_LUMINANCE_THRESHOLD}
          luminanceSmoothing={BLOOM_LUMINANCE_SMOOTHING}
        />
      </EffectComposer>
    </>
  );
}

// =============================================================================
// Exported Component
// =============================================================================
export function AgentNetwork() {
  return (
    <section className="relative py-24 border-t border-zinc-800/50">
      <div className="max-w-5xl mx-auto px-6 lg:px-16">
        <h2 className="font-heading text-3xl font-bold text-zinc-50 text-center mb-4">
          Agent network
        </h2>
        <p className="text-center text-zinc-500 text-base max-w-2xl mx-auto mb-12">
          Sessions, subagents, and tool calls — visualized as a living network.
          Each node is an agent. Each pulse is data flowing.
        </p>
      </div>
      <div className="h-[600px] w-full" style={{ mask: "linear-gradient(to bottom, transparent 0%, black 8%, black 92%, transparent 100%)", WebkitMask: "linear-gradient(to bottom, transparent 0%, black 8%, black 92%, transparent 100%)" }}>
        <Canvas
          camera={{ position: [0, 2, CAMERA_DISTANCE], fov: CAMERA_FOV }}
          dpr={[1, 2]}
          gl={{ antialias: true, alpha: true }}
          style={{ background: "transparent" }}
        >
          <Scene />
        </Canvas>
      </div>
    </section>
  );
}
