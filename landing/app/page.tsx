import { Hero } from "@/components/hero";
import { Positioning } from "@/components/positioning";
import { Features } from "@/components/features";
import { DataExplainer } from "@/components/data-explainer";
import { AgentNetwork } from "@/components/agent-network";
import { HowItWorks } from "@/components/how-it-works";
import { Stats } from "@/components/stats";
import { OpenSource } from "@/components/open-source";
import { Footer } from "@/components/footer";
import { FadeInOnScroll } from "@/components/fade-in-on-scroll";

export default function Home() {
  return (
    <main>
      <Hero />
      <FadeInOnScroll>
        <Positioning />
      </FadeInOnScroll>
      <Features />
      <DataExplainer />
      <AgentNetwork />
      <HowItWorks />
      <Stats />
      <OpenSource />
      <Footer />
    </main>
  );
}
