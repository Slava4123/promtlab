import { SubscriptionSection } from "@/components/subscription/subscription-section"
import { SectionHeader } from "./_section-header"

export default function SettingsSubscriptionPage() {
  return (
    <section>
      <SectionHeader title="Подписка" description="Тариф, оплата, история платежей" />
      <SubscriptionSection />
    </section>
  )
}
