import { ReferralSection } from "@/components/settings/referral-section"
import { SectionHeader } from "./_section-header"

export default function SettingsReferralPage() {
  return (
    <section>
      <SectionHeader title="Рефералы" description="Приглашайте друзей и получайте бонусы" />
      <ReferralSection />
    </section>
  )
}
