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
        <div className="flex-1" />
        <Footer />
      </section>
    </main>
  );
}
