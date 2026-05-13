import { useMemo, useState } from "react"
import { Gauge, ExternalLink } from "lucide-react"
import { useUsageSummary } from "../hooks/use-usage-summary"
import { useSettings } from "../hooks/use-settings"
import { openWebPage } from "../lib/utils"
import { cn } from "../lib/utils"
import type { UsageSummary, QuotaInfo, PlanID } from "../lib/types"

type ResourceKey = "prompts" | "collections" | "chains" | "ext_uses_today" | "mcp_uses_today"

const RESOURCE_LABELS: Record<ResourceKey, string> = {
  prompts: "Промпты",
  collections: "Коллекции",
  chains: "Цепочки",
  ext_uses_today: "Вставки сегодня",
  mcp_uses_today: "MCP сегодня",
}

const PLAN_LABELS: Record<PlanID, string> = {
  free: "Free",
  pro: "Pro",
  pro_yearly: "Pro (год)",
  max: "Max",
  max_yearly: "Max (год)",
}

function ratio({ used, limit }: QuotaInfo): number {
  // limit < 0 → ∞ (no limit), не считаем горячим.
  // limit === 0 → feature недоступна на тарифе — это НЕ 100% used,
  // это «не применимо». Раньше возвращали 1, и квота-индикатор показывал
  // 100% оранжевым у каждого нового free-юзера, кто никогда не использовал
  // ext_uses_today или mcp_uses_today. Семантический баг.
  if (limit <= 0) return 0
  return Math.min(used / limit, 1)
}

function format({ used, limit }: QuotaInfo): string {
  if (limit < 0) return `${used} / ∞`
  if (limit === 0) return "недоступно"
  return `${used} / ${limit}`
}

function severity(r: number): "ok" | "warn" | "danger" {
  if (r >= 0.95) return "danger"
  if (r >= 0.8) return "warn"
  return "ok"
}

// Компактный индикатор квот для header. Показывает «самый горячий» ресурс
// (с наибольшим использованием) + popover с деталями.
export function QuotaIndicator() {
  const usage = useUsageSummary()
  const settings = useSettings()
  const [open, setOpen] = useState(false)

  const hottest = useMemo(() => findHottest(usage.data), [usage.data])

  if (usage.isPending || !usage.data || !hottest) return null

  const sev = severity(hottest.r)
  const ringColor =
    sev === "danger"
      ? "text-(--color-destructive)"
      : sev === "warn"
        ? "text-(--color-warning)"
        : "text-(--color-muted-foreground)"
  // Контекст для screen-reader и tooltip: «N% использовано» вместо просто «N%».
  const usedLabel = `${Math.round(hottest.r * 100)}% использовано`

  function openPricing() {
    if (!settings) return
    openWebPage(settings.apiBase, "/pricing?source=quota_bar&from=extension")
    setOpen(false)
  }

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className={cn(
          "flex items-center gap-1 rounded-md px-1.5 py-1 text-[10px] font-medium transition-colors hover:bg-(--color-muted)",
          ringColor,
        )}
        aria-label={`${RESOURCE_LABELS[hottest.key]}: ${usedLabel}`}
        title={`${RESOURCE_LABELS[hottest.key]}: ${format(hottest.info)} (${usedLabel})`}
      >
        <Gauge className="h-3.5 w-3.5" />
        <span className="tabular-nums">{Math.round(hottest.r * 100)}%</span>
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} aria-hidden />
          <div className="absolute right-0 top-full z-50 mt-1 w-60 rounded-lg border border-(--color-border) bg-(--color-background) p-2.5 shadow-xl">
            <div className="mb-2 flex items-center justify-between">
              <div className="text-[10px] uppercase tracking-wide text-(--color-muted-foreground)">
                Использование
              </div>
              <span className="rounded bg-(--color-muted) px-1.5 py-0.5 text-[9px] font-medium">
                {PLAN_LABELS[usage.data.plan_id] ?? usage.data.plan_id}
              </span>
            </div>

            <ul className="space-y-1.5">
              {(Object.keys(RESOURCE_LABELS) as ResourceKey[]).map((key) => {
                const info = usage.data![key]
                if (!info) return null
                const r = ratio(info)
                const s = severity(r)
                return (
                  <li key={key} className="space-y-0.5">
                    <div className="flex items-center justify-between text-[11px]">
                      <span className="text-(--color-foreground)">{RESOURCE_LABELS[key]}</span>
                      <span className="font-mono text-(--color-muted-foreground)">
                        {format(info)}
                      </span>
                    </div>
                    <div className="h-1 overflow-hidden rounded-full bg-(--color-muted)">
                      <div
                        className={cn(
                          "h-full transition-all duration-(--duration-normal) ease-(--ease-out)",
                          s === "danger"
                            ? "bg-(--color-destructive)"
                            : s === "warn"
                              ? "bg-(--color-warning)"
                              : "bg-(--color-brand)",
                        )}
                        style={{ width: `${Math.round(r * 100)}%` }}
                      />
                    </div>
                  </li>
                )
              })}
            </ul>

            {usage.data.plan_id !== "max" && usage.data.plan_id !== "max_yearly" && (
              <button
                type="button"
                onClick={openPricing}
                className="mt-2.5 flex w-full items-center justify-center gap-1 rounded-md bg-(--color-brand) px-2 py-1.5 text-[10px] font-medium text-(--color-brand-foreground) shadow-[var(--brand-shadow)] hover:bg-[color-mix(in_oklch,var(--color-brand)_92%,black)]"
              >
                <ExternalLink className="h-3 w-3" />
                Повысить тариф
              </button>
            )}
          </div>
        </>
      )}
    </div>
  )
}

function findHottest(
  data: UsageSummary | undefined,
): { key: ResourceKey; info: QuotaInfo; r: number } | null {
  if (!data) return null
  let best: { key: ResourceKey; info: QuotaInfo; r: number } | null = null
  for (const key of Object.keys(RESOURCE_LABELS) as ResourceKey[]) {
    const info = data[key]
    if (!info) continue
    // Skip ресурсы с limit <= 0: limit < 0 = ∞ (Max-тариф), limit === 0 =
    // feature недоступна. Ни один из них не «горячий» — показывать их
    // 100% оранжевым было багом.
    if (info.limit <= 0) continue
    const r = ratio(info)
    if (!best || r > best.r) best = { key, info, r }
  }
  return best
}
