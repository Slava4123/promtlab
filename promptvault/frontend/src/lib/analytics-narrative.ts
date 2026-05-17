import type { PersonalDashboard, Insight } from "@/api/analytics"
import { formatRange } from "@/api/analytics"

export interface NarrativeSegments {
  summary: string
  topModel: string | null
  streak: string | null
  actionHint: string | null
}

interface StreakInput {
  current?: number
  longest?: number
}

// buildNarrative — template-функция для AI-style summary без LLM-вызовов.
// Принцип «без AI на нашей стороне» из CLAUDE.md: текст детерминирован.
// Каждый сегмент опциональный — может быть null если данных нет.
export function buildNarrative(
  data: PersonalDashboard,
  insights: Insight[] | null,
): NarrativeSegments {
  // streak может быть передан через data.streak (для тестов / future use).
  // На /analytics странице сегмент заполняется отдельно через useStreak hook +
  // buildStreakSegment() — текущий поток сохраняется.
  const streakData = (data as PersonalDashboard & { streak?: StreakInput }).streak
  return {
    summary: buildSummary(data),
    topModel: buildTopModel(data),
    streak: streakData ? buildStreakSegment(streakData.current) : null,
    actionHint: buildActionHint(insights),
  }
}

// buildStreakSegment — вынесено для использования из analytics.tsx (useStreak hook).
// Возвращает null если current === 0 (uninformative noise).
export function buildStreakSegment(current: number | undefined | null): string | null {
  if (!current || current <= 0) return null
  return `streak ${current} ${pluralStreak(current)}`
}

function pluralStreak(n: number): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return "день"
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return "дня"
  return "дней"
}

function buildSummary(data: PersonalDashboard): string {
  const period = formatRange(data.range)
  const uses = data.totals_current.uses
  if (uses === 0) {
    return `За ${period} пока тихо — самое время попробовать новые промпты`
  }
  const prev = data.totals_previous.uses
  const deltaText = formatDelta(uses, prev)
  return `За ${period}: ${uses.toLocaleString("ru")} использований${deltaText}`
}

function buildTopModel(data: PersonalDashboard): string | null {
  if (data.usage_by_model.length === 0) return null
  const total = data.usage_by_model.reduce((s, r) => s + r.uses, 0)
  if (total === 0) return null
  const top = [...data.usage_by_model].sort((a, b) => b.uses - a.uses)[0]
  // skip uninformative cases: empty model name OR 100% with single model
  if (!top.model) return null
  const pct = Math.round((top.uses / total) * 100)
  if (pct === 100 && data.usage_by_model.length === 1) return null
  return `топ-модель ${top.model} (${pct}%)`
}

function buildActionHint(insights: Insight[] | null): string | null {
  if (!insights || insights.length === 0) return null
  const parts: string[] = []
  for (const ins of insights) {
    const count = Array.isArray(ins.payload) ? ins.payload.length : 0
    if (count === 0) continue
    if (ins.type === "unused_prompts") parts.push(`${count} забытых`)
    else if (ins.type === "possible_duplicates") parts.push(`${count} дубликата`)
    else if (ins.type === "orphan_tags") parts.push(`${count} orphan-тегов`)
    else if (ins.type === "empty_collections") parts.push(`${count} пустых коллекций`)
  }
  if (parts.length === 0) return null
  return `${parts.join(" и ")} ждут уборки`
}

function formatDelta(current: number, previous: number): string {
  if (previous === 0) return ""
  const pct = Math.round(((current - previous) / previous) * 100)
  if (pct === 0) return " (без изменений)"
  return pct > 0 ? ` ↑${pct}%` : ` ↓${Math.abs(pct)}%`
}
