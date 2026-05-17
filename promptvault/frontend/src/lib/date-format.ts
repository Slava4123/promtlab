// Russian-locale Intl formatter — "7 мая", "1 дек." pattern.
// Cached as module-level constant: instantiating Intl.DateTimeFormat per call
// would slow down chart renders that call this 30+ times per frame.
const fmt = new Intl.DateTimeFormat("ru-RU", { day: "numeric", month: "short" })

/**
 * formatDayShort — преобразует ISO-дату в короткий русский формат "D MMM".
 *
 * - "2026-05-07" → "7 мая"
 * - "2026-12-01" → "1 дек."
 * - "" / неверная дата → ""
 *
 * Используется в analytics charts (UsageChart tick formatter, ActivityHeatmap
 * tooltip) — Wave 2 правит x-axis с "05-07" на читаемое "7 мая".
 */
export function formatDayShort(iso: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ""
  return fmt.format(d)
}
