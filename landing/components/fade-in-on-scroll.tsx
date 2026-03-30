"use client";

import { useEffect, useRef, useState } from "react";

export function FadeInOnScroll({ children }: { children: React.ReactNode }) {
  const ref = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    // Respect reduced motion
    const motionOk = !window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    if (!motionOk) {
      setVisible(true);
      return;
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) setVisible(true);
      },
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
