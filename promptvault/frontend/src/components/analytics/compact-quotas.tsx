import { Card } from "@/components/ui/card"
import type { UsageSummary } from "@/api/analytics"

interface CompactQuotasProps {
  quotas: UsageSummary | undefined
}

// CompactQuotas — однострочный footer с тремя ключевыми quota:
// Промпты, Коллекции, MCP-вызовы сегодня.
// Заменяет 3 больших QuotaProgress блока, не занимает prime real-estate.
export function CompactQuotas({ quotas }: CompactQuotasProps) {
  if (!quotas) return null

  const items = [
    { label: "Промпты", used: quotas.prompts.used, limit: quotas.prompts.limit },
    { label: "Коллекции", used: quotas.collections.used, limit: quotas.collections.limit },
    { label: "MCP сегодня", used: quotas.mcp_uses_today.used, limit: quotas.mcp_uses_today.limit },
  ]

  return (
    <Card className="px-5 py-4">
      <div className="mb-3 flex items-center gap-2">
        <span className="text-[11px] uppercase tracking-wide text-muted-foreground">Лимиты</span>
      </div>
      {/* Grid вместо flex: на mobile столбиком, на ≥md в три равные колонки.
          Старый flex с gap-6 на узких экранах схлопывался в вертикаль с
          centered текстом и узкими прогресс-барами. Grid даёт стабильную
          раскладку и одинаковую ширину bar'ам. */}
      <div className="grid gap-4 md:grid-cols-3">
        {items.map((item) => {
          const pct = item.limit > 0 ? (item.used / item.limit) * 100 : 0
          const isHigh = pct >= 90
          const isMid = pct >= 75 && pct < 90
          const barColor = isHigh ? "bg-rose-500" : isMid ? "bg-amber-500" : "bg-violet-500"
          return (
            <div key={item.label} className="space-y-1.5">
              <div className="flex items-baseline justify-between gap-2">
                <span className="text-xs font-medium text-foreground/80">{item.label}</span>
                <span className="font-mono text-xs tabular-nums text-muted-foreground">
                  {item.used.toLocaleString("ru")} / {item.limit.toLocaleString("ru")}
                </span>
              </div>
              <div className="h-2 overflow-hidden rounded-full bg-foreground/15">
                <div
                  className={`h-full rounded-full transition-[width] duration-300 ${barColor}`}
                  // min-width 6px: при квоте 7/10000 = 0.07% реальная ширина
                  // была 1px (~невидимо). Минимальный показ — чтобы юзер
                  // видел «активность есть, просто далеко от лимита».
                  // При 0 — 0 (полностью пустой trek).
                  style={{
                    width: pct > 0 ? `max(${Math.min(pct, 100)}%, 6px)` : "0",
                  }}
                />
              </div>
            </div>
          )
        })}
      </div>
    </Card>
  )
}
