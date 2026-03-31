"use client";

import { useEffect, useRef, useState } from "react";

interface AnimateInProps {
  children: React.ReactNode;
  className?: string;
  delay?: number;       // ms delay before animation starts
  duration?: number;    // ms duration
  from?: "bottom" | "left" | "right" | "scale"; // entrance direction
  once?: boolean;       // only animate once (default true)
  threshold?: number;   // IntersectionObserver threshold
}

export function AnimateIn({
  children,
  className = "",
  delay = 0,
  duration = 600,
  from = "bottom",
  once = true,
  threshold = 0.15,
}: AnimateInProps) {
  const ref = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const motionOk = !window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (!motionOk) {
      setVisible(true);
      return;
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setVisible(true);
          if (once && ref.current) observer.unobserve(ref.current);
        }
      },
      { threshold }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, [once, threshold]);

  const transforms: Record<string, string> = {
    bottom: "translateY(24px)",
    left: "translateX(-24px)",
    right: "translateX(24px)",
    scale: "scale(0.95) translateY(12px)",
  };

  return (
    <div
      ref={ref}
      className={className}
      style={{
        opacity: visible ? 1 : 0,
        transform: visible ? "none" : transforms[from],
        transition: `opacity ${duration}ms ease-out ${delay}ms, transform ${duration}ms ease-out ${delay}ms`,
      }}
    >
      {children}
    </div>
  );
}
