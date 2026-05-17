import {
  AlertCircle,
  Copy,
  TrendingUp,
  TrendingDown,
  Archive,
  Hash,
  FolderOpen,
  type LucideIcon,
} from "lucide-react"
import { InsightActionCard, type InsightTone } from "./insight-action-card"
import type { Insight } from "@/api/analytics"

// INSIGHT_META — единая таблица мапинга Insight.type → визуальные параметры
// карточки. Tone задаёт цветовую группу (warning/info/success), descBuilder
// формирует русскую плюрализованную подпись. Маппинг покрывает все 7 типов
// из Insight["type"]; для неизвестного типа InsightsPanel рендерит null.
const INSIGHT_META: Record<
  Insight["type"],
  {
    icon: LucideIcon
    tone: InsightTone
    title: string
    href: string
    descBuilder: (n: number) => string
    ctaLabel: string
  }
> = {
  unused_prompts: {
    icon: AlertCircle,
    tone: "warning",
    title: "Забытые",
    href: "/prompts?filter=unused",
    descBuilder: (n) =>
      `${n} ${n === 1 ? "промпт не использовался" : "промптов не использовались"} 30+ дней`,
    ctaLabel: "Посмотреть",
  },
  possible_duplicates: {
    icon: Copy,
    tone: "info",
    title: "Дубликаты",
    href: "/prompts?filter=duplicates",
    descBuilder: (n) => `${n} ${n === 1 ? "пара" : "пары"} похожих промптов`,
    ctaLabel: "Объединить",
  },
  trending: {
    icon: TrendingUp,
    tone: "success",
    title: "Растут",
    href: "/prompts?filter=trending",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} растут в использовании`,
    ctaLabel: "Открыть",
  },
  declining: {
    icon: TrendingDown,
    tone: "warning",
    title: "Падают",
    href: "/prompts?filter=declining",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} используются всё реже`,
    ctaLabel: "Посмотреть",
  },
  most_edited: {
    icon: Archive,
    tone: "info",
    title: "Часто правят",
    href: "/prompts?sort=most-edited",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} с большим числом версий`,
    ctaLabel: "Открыть",
  },
  orphan_tags: {
    icon: Hash,
    tone: "warning",
    title: "Orphan-теги",
    href: "/tags",
    descBuilder: (n) => `${n} ${n === 1 ? "тег" : "тегов"} без промптов`,
    ctaLabel: "Очистить",
  },
  empty_collections: {
    icon: FolderOpen,
    tone: "warning",
    title: "Пустые коллекции",
    href: "/collections",
    descBuilder: (n) => `${n} ${n === 1 ? "коллекция" : "коллекций"} без промптов`,
    ctaLabel: "Очистить",
  },
}

interface InsightsPanelProps {
  insights: Insight[]
}

export function InsightsPanel({ insights }: InsightsPanelProps) {
  if (insights.length === 0) {
    return <p className="text-sm text-muted-foreground">Пока нет инсайтов. Возвращайтесь завтра.</p>
  }
  return (
    <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
      {insights.map((ins) => {
        const meta = INSIGHT_META[ins.type]
        if (!meta) return null
        const count = Array.isArray(ins.payload) ? ins.payload.length : 0
        return (
          <InsightActionCard
            key={ins.type}
            tone={meta.tone}
            icon={meta.icon}
            title={meta.title}
            description={meta.descBuilder(count)}
            href={meta.href}
            ctaLabel={meta.ctaLabel}
            count={count}
          />
        )
      })}
    </div>
  )
}
