import { Navigate } from "react-router-dom"
import { useAuthStore } from "@/stores/auth-store"

import { LandingHeader } from "./landing/components/landing-header"
import { HeroSection } from "./landing/sections/hero"
import { CompatibilitySection } from "./landing/sections/compatibility"
import { ProductDemoSection } from "./landing/sections/product-demo"
import { FeatureDeepDivesSection } from "./landing/sections/feature-deep-dives"
import { IntegrationsRailSection } from "./landing/sections/integrations-rail"
import { TeamsSection } from "./landing/sections/teams-section"
import { AdvantagesBentoSection } from "./landing/sections/advantages-bento"
import { BetaBannerSection } from "./landing/sections/beta-banner"
import { LandingFooter } from "./landing/sections/landing-footer"

export default function Landing() {
  const { isAuthenticated, isLoading } = useAuthStore()

  if (isLoading) return null
  if (isAuthenticated) return <Navigate to="/dashboard" replace />

  return (
    <div className="dark min-h-screen bg-background text-foreground overflow-x-hidden">
      <LandingHeader />
      <main>
        <HeroSection />
        <CompatibilitySection />
        <ProductDemoSection />
        <FeatureDeepDivesSection />
        <IntegrationsRailSection />
        <TeamsSection />
        <AdvantagesBentoSection />
        <BetaBannerSection />
      </main>
      <LandingFooter />
    </div>
  )
}
