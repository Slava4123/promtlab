// Единственный источник истины для русских названий ресурсов из quota-системы.
// Импортируется из quota-exceeded-dialog (показ при 402) и notifications-page
// (показ при over-limit warning через UsageSummary). Раньше каждое место имело
// свою копию QUOTA_LABELS — drift привёл к отсутствию team_*/branding/insights.

import { addBreadcrumb } from "./sentry"
import type { PlanID, QuotaErrorType, QuotaKey } from "./types"

// LabelKey — объединение error-namespace (newQuotaExceeded) и usage-namespace
// (UsageSummary). Имена namespace отличаются: "ext_daily" (error) ↔
// "ext_uses_today" (usage), но семантически один ресурс — одна метка.
export type QuotaLabelKey = QuotaErrorType | QuotaKey

// Record<QuotaLabelKey, ...> даёт compile-time exhaustiveness: при добавлении
// нового quota_type в backend (и в QUOTA_ERROR_TYPES) TypeScript потребует
// добавить метку сюда, иначе билд упадёт.
export const QUOTA_LABELS: Record<QuotaLabelKey, string> = {
  // Personal — общие для error и usage namespace
  prompts: "Промпты",
  collections: "Коллекции",
  chains: "Цепочки",
  teams: "Команды",
  // Personal daily — два namespace, одна метка
  ext_daily: "Вставки сегодня",
  ext_uses_today: "Вставки сегодня",
  mcp_daily: "MCP-вызовы сегодня",
  mcp_uses_today: "MCP-вызовы сегодня",
  // Team-pool
  team_prompts: "Промпты команды",
  team_collections: "Коллекции команды",
  team_chains: "Цепочки команды",
  team_members: "Участники команды",
  chain_steps: "Шаги в цепочке",
  // Misc
  branding: "Брендинг команды",
  insights: "Smart Insights (Max)",
  export: "Экспорт CSV (Pro)",
}

export const PLAN_LABELS: Record<PlanID, string> = {
  free: "Free",
  pro: "Pro",
  pro_yearly: "Pro (год)",
  max: "Max",
  max_yearly: "Max (год)",
}

// Fallback-метка, когда ни quotaType, ни эвристика по тексту не сработали.
export const QUOTA_FALLBACK_LABEL = "Лимит ресурса"

function isKnownQuotaLabelKey(s: string): s is QuotaLabelKey {
  return s in QUOTA_LABELS
}

// quotaTypeLabel — для случаев, когда ключ заведомо QuotaKey (UsageSummary).
// notifications-page использует именно его.
export function quotaTypeLabel(key: QuotaLabelKey): string {
  return QUOTA_LABELS[key]
}

// readableQuotaType — для 402 quota dialog. Backend шлёт quota_type, но если
// его не распознали (новый ресурс / typo / null) — fall back на эвристику по
// тексту сообщения; в крайнем случае показываем QUOTA_FALLBACK_LABEL и
// добавляем Sentry breadcrumb, чтобы SRE заметил что нужен новый ключ.
export function readableQuotaType(
  quotaType: string | null,
  message: string | null,
): string {
  if (quotaType && quotaType !== "unknown" && isKnownQuotaLabelKey(quotaType)) {
    return QUOTA_LABELS[quotaType]
  }
  if (message) {
    const m = message.toLowerCase()
    if (m.includes("цепоч")) return QUOTA_LABELS.chains
    if (m.includes("промпт")) return QUOTA_LABELS.prompts
    if (m.includes("коллекц")) return QUOTA_LABELS.collections
    if (m.includes("команд")) return QUOTA_LABELS.teams
    // "встав" покрывает «вставки/вставку/вставок/вставкой» — все формы.
    if (m.includes("встав") || m.includes("использовани") || m.includes("расширен")) {
      return QUOTA_LABELS.ext_daily
    }
    if (m.includes("mcp")) return QUOTA_LABELS.mcp_daily
    if (m.includes("инсайт") || m.includes("insights")) return QUOTA_LABELS.insights
    if (m.includes("экспорт") || m.includes("export")) return QUOTA_LABELS.export
    if (m.includes("брендин") || m.includes("логотип")) return QUOTA_LABELS.branding
  }
  // Не флудим, если контекст совсем пустой (quotaType=null && message=null).
  if (quotaType || message) {
    addBreadcrumb(
      "quota.unknown_type",
      "fallback used",
      { quotaType, msg: message?.slice(0, 80) },
      "warning",
    )
  }
  return QUOTA_FALLBACK_LABEL
}
