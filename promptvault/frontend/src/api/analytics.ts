import { api } from "./client"

// --- Response types (mirror backend analyticsuc.* structs) ---

export type AnalyticsRange = "7d" | "30d" | "90d" | "365d"

// formatRange — человекочитаемая подпись диапазона для UI. API возвращает
// технический ID ("7d", "30d" и т.д.), но показывать это юзеру некрасиво.
// Одна точка локализации — чтобы не дублировать map в каждой странице.
export function formatRange(r: AnalyticsRange | string): string {
  switch (r) {
    case "7d":
      return "7 дней"
    case "30d":
      return "30 дней"
    case "90d":
      return "90 дней"
    case "365d":
      return "365 дней"
    default:
      return r
  }
}

export interface UsagePoint {
  day: string // ISO date
  count: number
}

export interface PromptUsageRow {
  prompt_id: number
  title: string
  uses: number
}

export interface ContributorRow {
  user_id: number
  email: string
  name?: string
  prompts_created: number
  prompts_edited: number
  uses: number
}

export interface QuotaInfo {
  used: number
  limit: number
}

export interface UsageSummary {
  plan_id: string
  prompts: QuotaInfo
  collections: QuotaInfo
  teams: QuotaInfo
  share_links: QuotaInfo
  daily_shares_today: QuotaInfo
  ext_uses_today: QuotaInfo
  mcp_uses_today: QuotaInfo
}

// Totals — суммы за период для "+XX%/-YY%" индикатора в метриках.
export interface Totals {
  uses: number
  created: number
  updated: number
  share_views: number
}

// ModelUsageRow — сегментация использований по AI-модели.
// model === "" → "Без модели" (промпт без указанной model).
export interface ModelUsageRow {
  model: string
  uses: number
}

export interface PersonalDashboard {
  range: AnalyticsRange
  usage_per_day: UsagePoint[]
  top_prompts: PromptUsageRow[]
  prompts_created_per_day: UsagePoint[]
  prompts_updated_per_day: UsagePoint[]
  share_views_per_day: UsagePoint[]
  top_shared: PromptUsageRow[]
  quotas?: UsageSummary
  totals_current: Totals
  totals_previous: Totals
  usage_by_model: ModelUsageRow[]
}

export interface TeamDashboard {
  range: AnalyticsRange
  usage_per_day: UsagePoint[]
  top_prompts: PromptUsageRow[]
  prompts_created_per_day: UsagePoint[]
  prompts_updated_per_day: UsagePoint[]
  contributors: ContributorRow[]
  totals_current: Totals
  totals_previous: Totals
  usage_by_model: ModelUsageRow[]
}

// computeDelta — %-изменение от prev к current. Возвращает null если prev=0
// (нет базы для сравнения — UI покажет «—»).
export function computeDelta(current: number, previous: number): number | null {
  if (previous === 0) return null
  return Math.round(((current - previous) / previous) * 100)
}

export interface PromptAnalytics {
  prompt_id: number
  usage_per_day: UsagePoint[]
  share_views_per_day: UsagePoint[]
}

export interface Insight {
  type:
    | "unused_prompts"
    | "trending"
    | "declining"
    | "most_edited"
    | "possible_duplicates"
    | "orphan_tags"
    | "empty_collections"
  payload: unknown
  computed_at: string
}

export interface InsightsResponse {
  items: Insight[]
}

// --- API functions ---

export function fetchPersonalAnalytics(range: AnalyticsRange = "7d"): Promise<PersonalDashboard> {
  return api<PersonalDashboard>(`/analytics/personal?range=${range}`)
}

export function fetchTeamAnalytics(teamId: number, range: AnalyticsRange = "7d"): Promise<TeamDashboard> {
  return api<TeamDashboard>(`/analytics/teams/${teamId}?range=${range}`)
}

export function fetchPromptAnalytics(promptId: number): Promise<PromptAnalytics> {
  return api<PromptAnalytics>(`/analytics/prompts/${promptId}`)
}

export function fetchInsights(): Promise<InsightsResponse> {
  return api<InsightsResponse>("/analytics/insights")
}

// refreshInsights — force-пересчёт инсайтов (Max-only, rate-limit 1/час).
// Backend вернёт 429 если лимит исчерпан — ApiError подхватит статус.
export function refreshInsights(): Promise<InsightsResponse> {
  return api<InsightsResponse>("/analytics/insights/refresh", { method: "POST" })
}

// Export URL: не делаем fetch, а возвращаем готовую ссылку для скачивания
// (браузер сделает GET с auth-cookie). JWT-access-token передать через URL
// не можем — но наш fetch wrapper использует cookie-based session refresh.
// Для корректного download: используем programmatic GET с blob download.
export async function downloadAnalyticsCSV(
  scope: "personal" | "team",
  range: AnalyticsRange = "90d",
  teamId?: number,
  format: "csv" | "xlsx" = "csv",
): Promise<void> {
  const params = new URLSearchParams({ format, scope, range })
  if (scope === "team" && teamId) {
    params.set("team_id", String(teamId))
  }
  // Используем blob ответ — наш api<T> парсит JSON, здесь нужно raw.
  const url = `/api/analytics/export?${params.toString()}`
  const token = (await import("./client")).getAccessToken()
  const res = await fetch(url, {
    credentials: "include",
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  })
  if (!res.ok) {
    throw new Error(`Download failed: ${res.status}`)
  }
  const blob = await res.blob()
  const link = document.createElement("a")
  link.href = URL.createObjectURL(blob)
  link.download = `analytics-${scope}-${range}.csv`
  link.click()
  URL.revokeObjectURL(link.href)
}
