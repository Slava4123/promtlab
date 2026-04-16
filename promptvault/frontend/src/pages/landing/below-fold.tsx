import { ProductDemoSection } from "./sections/product-demo"
import { FeatureDeepDivesSection } from "./sections/feature-deep-dives"
import { IntegrationsRailSection } from "./sections/integrations-rail"
import { TeamsSection } from "./sections/teams-section"
import { AdvantagesBentoSection } from "./sections/advantages-bento"
import { BetaBannerSection } from "./sections/beta-banner"

/**
 * P-13: Группируем все below-fold секции лендинга в один chunk, чтобы initial
 * bundle для первого экрана (Hero + Compatibility) был минимальным. Лениво
 * подгружается через IntersectionObserver когда юзер начинает скроллить.
 */
export default function BelowFold() {
  return (
    <>
      <ProductDemoSection />
      <FeatureDeepDivesSection />
      <IntegrationsRailSection />
      <TeamsSection />
      <AdvantagesBentoSection />
      <BetaBannerSection />
    </>
  )
}
