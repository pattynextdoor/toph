"use client";

import { useRef, useMemo } from "react";
import { Canvas, useFrame } from "@react-three/fiber";
import * as THREE from "three";

// =============================================================================
// TUNING CONSTANTS
// =============================================================================
const PLANE_SIZE = 20;
const PLANE_SEGMENTS = 128;
const WAVE_SPEED = 0.8;
const WAVE_AMPLITUDE = 0.15;
const WAVE_FREQUENCY = 3.0;
const WAVE_DECAY = 0.3;
const RING_COUNT = 5;
const COLOR_PRIMARY = "#87AFFF";
const COLOR_SECONDARY = "#87D787";
const OPACITY = 0.12;

// =============================================================================
// Seismic Plane — concentric ripples radiating from center
// =============================================================================
function SeismicPlane() {
  const meshRef = useRef<THREE.Mesh>(null);
  const materialRef = useRef<THREE.ShaderMaterial>(null);

  const shaderArgs = useMemo(
    () => ({
      uniforms: {
        uTime: { value: 0 },
        uColor1: { value: new THREE.Color(COLOR_PRIMARY) },
        uColor2: { value: new THREE.Color(COLOR_SECONDARY) },
        uOpacity: { value: OPACITY },
        uWaveSpeed: { value: WAVE_SPEED },
        uWaveAmplitude: { value: WAVE_AMPLITUDE },
        uWaveFrequency: { value: WAVE_FREQUENCY },
        uWaveDecay: { value: WAVE_DECAY },
        uRingCount: { value: RING_COUNT },
      },
      vertexShader: `
        varying vec2 vUv;
        varying float vElevation;
        uniform float uTime;
        uniform float uWaveSpeed;
        uniform float uWaveAmplitude;
        uniform float uWaveFrequency;
        uniform float uWaveDecay;

        void main() {
          vUv = uv;
          vec3 pos = position;

          // Distance from center
          float dist = length(pos.xz);

          // Concentric ripples radiating outward
          float wave = sin(dist * uWaveFrequency - uTime * uWaveSpeed) * uWaveAmplitude;

          // Decay with distance
          wave *= exp(-dist * uWaveDecay);

          // Add secondary ripple at different frequency
          wave += sin(dist * uWaveFrequency * 1.7 - uTime * uWaveSpeed * 0.7) * uWaveAmplitude * 0.3 * exp(-dist * uWaveDecay * 1.2);

          pos.y += wave;
          vElevation = wave;

          gl_Position = projectionMatrix * modelViewMatrix * vec4(pos, 1.0);
        }
      `,
      fragmentShader: `
        varying vec2 vUv;
        varying float vElevation;
        uniform vec3 uColor1;
        uniform vec3 uColor2;
        uniform float uOpacity;
        uniform float uTime;

        void main() {
          // Distance from center in UV space
          float dist = length(vUv - 0.5) * 2.0;

          // Ring pattern
          float ring = abs(sin(dist * 12.0 - uTime * 0.5)) * 0.5 + 0.5;

          // Color based on elevation and distance
          vec3 color = mix(uColor1, uColor2, vElevation * 3.0 + 0.5);

          // Fade at edges
          float fade = 1.0 - smoothstep(0.3, 1.0, dist);

          // Combine with ring pattern for grid-like effect
          float alpha = uOpacity * fade * (0.3 + ring * 0.7);

          // Add bright lines at wave peaks
          float peak = smoothstep(0.06, 0.08, abs(vElevation)) * 0.5;
          alpha += peak * fade * 0.15;

          gl_FragColor = vec4(color, alpha);
        }
      `,
      transparent: true,
      depthWrite: false,
      side: THREE.DoubleSide,
      blending: THREE.AdditiveBlending,
    }),
    []
  );

  useFrame(({ clock }) => {
    if (materialRef.current) {
      materialRef.current.uniforms.uTime.value = clock.getElapsedTime();
    }
  });

  return (
    <mesh ref={meshRef} rotation={[-Math.PI / 2.5, 0, 0]} position={[0, -1, 0]}>
      <planeGeometry args={[PLANE_SIZE, PLANE_SIZE, PLANE_SEGMENTS, PLANE_SEGMENTS]} />
      <shaderMaterial ref={materialRef} args={[shaderArgs]} />
    </mesh>
  );
}

// =============================================================================
// Scene
// =============================================================================
function Scene() {
  return (
    <>
      <SeismicPlane />
    </>
  );
}

// =============================================================================
// Exported Component — positioned absolutely behind the hero
// =============================================================================
export function SeismicBackground() {
  return (
    <div className="absolute inset-0 z-0 opacity-60 pointer-events-none">
      <Canvas
        camera={{ position: [0, 3, 6], fov: 60 }}
        dpr={[1, 1.5]}
        gl={{ antialias: false, alpha: true }}
        style={{ background: "transparent" }}
      >
        <Scene />
      </Canvas>
    </div>
  );
}
