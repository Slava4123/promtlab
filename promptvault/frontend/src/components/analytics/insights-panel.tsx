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
// формирует русскую плюрализованную подпись с реальным count, emptyDesc —
// текст для discovery-state (когда инсайтов этого типа нет, карточка всё
// равно рендерится приглушённой, чтобы юзер видел доступные категории).
const INSIGHT_META: Record<
  Insight["type"],
  {
    icon: LucideIcon
    tone: InsightTone
    title: string
    href: string
    descBuilder: (n: number) => string
    emptyDesc: string
    ctaLabel: string
    emptyCtaLabel: string
  }
> = {
  unused_prompts: {
    icon: AlertCircle,
    tone: "warning",
    title: "Забытые",
    href: "/prompts/insights/unused",
    descBuilder: (n) =>
      `${n} ${n === 1 ? "промпт не использовался" : "промптов не использовались"} 30+ дней`,
    emptyDesc: "Промпты, которые не использовались 30+ дней",
    ctaLabel: "Посмотреть",
    emptyCtaLabel: "Открыть",
  },
  possible_duplicates: {
    icon: Copy,
    tone: "info",
    title: "Дубликаты",
    href: "/prompts/insights/duplicates",
    descBuilder: (n) => `${n} ${n === 1 ? "пара" : "пары"} похожих промптов`,
    emptyDesc: "Похожие промпты, которые стоит объединить",
    ctaLabel: "Объединить",
    emptyCtaLabel: "Открыть",
  },
  trending: {
    icon: TrendingUp,
    tone: "success",
    title: "Растут",
    href: "/prompts/insights/trending",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} растут в использовании`,
    emptyDesc: "Промпты с ростом использования за неделю",
    ctaLabel: "Открыть",
    emptyCtaLabel: "Открыть",
  },
  declining: {
    icon: TrendingDown,
    tone: "warning",
    title: "Падают",
    href: "/prompts/insights/declining",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} используются всё реже`,
    emptyDesc: "Промпты, использование которых снизилось",
    ctaLabel: "Посмотреть",
    emptyCtaLabel: "Открыть",
  },
  most_edited: {
    icon: Archive,
    tone: "info",
    title: "Часто правят",
    href: "/prompts/insights/most-edited",
    descBuilder: (n) => `${n} ${n === 1 ? "промпт" : "промпта"} с большим числом версий`,
    emptyDesc: "Промпты с наибольшим числом редакций",
    ctaLabel: "Открыть",
    emptyCtaLabel: "Открыть",
  },
  orphan_tags: {
    icon: Hash,
    tone: "warning",
    title: "Теги без промптов",
    href: "/tags?filter=orphan",
    descBuilder: (n) => `${n} ${n === 1 ? "тег" : "тегов"} без промптов`,
    emptyDesc: "Теги, не привязанные к активным промптам",
    ctaLabel: "Очистить",
    emptyCtaLabel: "Открыть",
  },
  empty_collections: {
    icon: FolderOpen,
    tone: "warning",
    title: "Пустые коллекции",
    href: "/collections?filter=empty",
    descBuilder: (n) => `${n} ${n === 1 ? "коллекция" : "коллекций"} без промптов`,
    emptyDesc: "Коллекции без единого активного промпта",
    ctaLabel: "Очистить",
    emptyCtaLabel: "Открыть",
  },
}

// INSIGHT_ORDER — фиксированный визуальный порядок карточек. Используется
// когда showAll=true, чтобы юзер видел стабильную сетку категорий.
const INSIGHT_ORDER: Insight["type"][] = [
  "unused_prompts",
  "possible_duplicates",
  "trending",
  "declining",
  "most_edited",
  "orphan_tags",
  "empty_collections",
]

interface InsightsPanelProps {
  insights: Insight[]
  // showAll — рендерить все типы из INSIGHT_ORDER, заполняя empty-state
  // карточками для тех, по которым инсайтов нет. Discoverability: юзер видит
  // доступные категории умной аналитики целиком, даже если данных пока нет.
  // По умолчанию false (legacy behaviour: только cards с реальными данными).
  showAll?: boolean
  // allowedTypes — фильтр для discovery-режима (showAll=true): рендерим
  // только типы из этого списка (сохраняя порядок INSIGHT_ORDER), остальные
  // не показываем. Используется для Pro-юзеров, которым доступны только 2 типа
  // (unused_prompts + possible_duplicates по ADR-0008): backend может не
  // прислать пустой possible_duplicates, но фронт всё равно показывает empty
  // карточку, чтобы юзер видел доступную категорию. Не задано → все 7 типов.
  allowedTypes?: Insight["type"][]
}

export function InsightsPanel({ insights, showAll = false, allowedTypes }: InsightsPanelProps) {
  if (!showAll && insights.length === 0) {
    return <p className="text-sm text-muted-foreground">Пока нет инсайтов. Возвращайтесь завтра.</p>
  }

  if (showAll) {
    const byType = new Map(insights.map((ins) => [ins.type, ins]))
    const allowed = allowedTypes ? new Set(allowedTypes) : null
    const types = allowed ? INSIGHT_ORDER.filter((t) => allowed.has(t)) : INSIGHT_ORDER
    return (
      <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
        {types.map((type) => {
          const meta = INSIGHT_META[type]
          const ins = byType.get(type)
          const count = ins && Array.isArray(ins.payload) ? ins.payload.length : 0
          const isEmpty = !ins || count === 0
          return (
            <InsightActionCard
              key={type}
              tone={meta.tone}
              icon={meta.icon}
              title={meta.title}
              description={isEmpty ? meta.emptyDesc : meta.descBuilder(count)}
              href={meta.href}
              ctaLabel={isEmpty ? meta.emptyCtaLabel : meta.ctaLabel}
              count={isEmpty ? undefined : count}
              empty={isEmpty}
            />
          )
        })}
      </div>
    )
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
