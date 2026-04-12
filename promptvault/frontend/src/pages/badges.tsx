import { Loader2 } from "lucide-react"

import { PageLayout } from "@/components/layout/page-layout"
import { BadgeCard } from "@/components/badges/badge-card"
import { useBadges } from "@/hooks/use-badges"
import type { Badge, BadgeCategory } from "@/api/types"

const CATEGORY_LABELS: Record<BadgeCategory, string> = {
  personal: "Личные",
  team: "Командные",
  milestone: "Общие",
  streak: "Регулярность",
}

// Порядок отображения категорий.
const CATEGORY_ORDER: BadgeCategory[] = ["personal", "team", "milestone", "streak"]

function groupByCategory(badges: Badge[]): Map<BadgeCategory, Badge[]> {
  const groups = new Map<BadgeCategory, Badge[]>()
  for (const cat of CATEGORY_ORDER) {
    groups.set(cat, [])
  }
  for (const b of badges) {
    const arr = groups.get(b.category)
    if (arr) arr.push(b)
  }
  return groups
}

export default function BadgesPage() {
  const { data, isLoading, error } = useBadges()

  if (isLoading) {
    return (
      <PageLayout title="Достижения" description="Разблокируй бейджи, развивая свою коллекцию промптов.">
        <div className="flex h-40 items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      </PageLayout>
    )
  }

  if (error || !data) {
    return (
      <PageLayout title="Достижения">
        <p className="text-sm text-muted-foreground">
          Не удалось загрузить достижения. Попробуйте обновить страницу.
        </p>
      </PageLayout>
    )
  }

  const groups = groupByCategory(data.items)

  return (
    <PageLayout
      title="Достижения"
      description={`Разблокировано ${data.total_unlocked} из ${data.total_count}`}
    >
      <div className="space-y-8">
        {CATEGORY_ORDER.map((cat) => {
          const items = groups.get(cat) ?? []
          if (items.length === 0) return null
          return (
            <section key={cat}>
              <h2 className="mb-3 text-[0.72rem] font-medium uppercase tracking-wider text-muted-foreground">
                {CATEGORY_LABELS[cat]}
              </h2>
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {items.map((b) => (
                  <BadgeCard key={b.id} badge={b} />
                ))}
              </div>
            </section>
          )
        })}
      </div>
    </PageLayout>
  )
}
