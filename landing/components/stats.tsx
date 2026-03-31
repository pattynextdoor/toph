"use client";

import { NumberTicker } from "@/components/ui/number-ticker";

const stats = [
  { value: 5, label: "panels", suffix: "" },
  { value: 9, label: "hook events", suffix: "" },
  { value: 30, label: "fps rendering", suffix: "" },
  { value: 0, label: "config needed", suffix: "" },
];

export function Stats() {
  return (
    <section className="px-6 lg:px-16 py-24 border-t border-zinc-800/50">
      <div className="max-w-4xl mx-auto grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
        {stats.map((stat) => (
          <div key={stat.label}>
            <div className="font-heading text-4xl font-bold text-zinc-50">
              {stat.value > 0 ? (
                <>
                  <NumberTicker
                    value={stat.value}
                    className="font-heading text-4xl font-bold text-zinc-50"
                  />
                  {stat.suffix && <span>{stat.suffix}</span>}
                </>
              ) : (
                <span>0</span>
              )}
            </div>
            <p className="font-mono text-xs text-zinc-500 mt-2 uppercase tracking-wider">
              {stat.label}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}
