import { Hero } from "@/components/hero";
import { Positioning } from "@/components/positioning";
import { Features } from "@/components/features";
import { DataExplainer } from "@/components/data-explainer";
import { EventStream } from "@/components/event-stream";
import { HowItWorks } from "@/components/how-it-works";
import { Stats } from "@/components/stats";
import { OpenSource } from "@/components/open-source";
import { Footer } from "@/components/footer";
import { AnimateIn } from "@/components/animate-in";

export default function Home() {
  return (
    <main>
      <Hero />
      <AnimateIn>
        <Positioning />
      </AnimateIn>
      <AnimateIn from="scale">
        <Features />
      </AnimateIn>
      <AnimateIn from="bottom" delay={100}>
        <DataExplainer />
      </AnimateIn>
      <AnimateIn from="scale">
        <EventStream />
      </AnimateIn>
      <AnimateIn from="bottom">
        <HowItWorks />
      </AnimateIn>
      <AnimateIn from="scale" delay={50}>
        <Stats />
      </AnimateIn>
      <AnimateIn from="bottom">
        <OpenSource />
      </AnimateIn>
      <Footer />
    </main>
  );
}
