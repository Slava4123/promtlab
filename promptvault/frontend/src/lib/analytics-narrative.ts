import type { PersonalDashboard, Insight } from "@/api/analytics"
import { formatRange } from "@/api/analytics"

export interface NarrativeSegments {
  summary: string
  topModel: string | null
  streak: string | null
  actionHint: string | null
}

// buildNarrative — template-функция для AI-style summary без LLM-вызовов.
// Принцип «без AI на нашей стороне» из CLAUDE.md: текст детерминирован.
// Каждый сегмент опциональный — может быть null если данных нет.
export function buildNarrative(
  data: PersonalDashboard,
  insights: Insight[] | null,
): NarrativeSegments {
  return {
    summary: buildSummary(data),
    topModel: buildTopModel(data),
    streak: null, // заполняется в narrative-banner.tsx через useStreak hook
    actionHint: buildActionHint(insights),
  }
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
  const pct = Math.round((top.uses / total) * 100)
  const name = top.model === "" ? "Без модели" : top.model
  return `топ-модель ${name} (${pct}%)`
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
