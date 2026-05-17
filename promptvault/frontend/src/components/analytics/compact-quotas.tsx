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
    <Card className="flex items-center gap-6 px-4 py-3">
      <span className="text-[11px] uppercase tracking-wide text-muted-foreground">Лимиты:</span>
      {items.map((item) => {
        const pct = item.limit > 0 ? (item.used / item.limit) * 100 : 0
        const isHigh = pct >= 90
        const isMid = pct >= 75 && pct < 90
        const barColor = isHigh ? "bg-rose-500" : isMid ? "bg-amber-500" : "bg-violet-500"
        return (
          <div key={item.label} className="flex flex-1 items-center gap-2">
            <span className="text-xs text-muted-foreground">{item.label}</span>
            <div className="flex-1">
              <div className="h-1.5 overflow-hidden rounded-full bg-foreground/10">
                <div
                  className={`h-full ${barColor}`}
                  style={{ width: `${Math.min(pct, 100)}%` }}
                />
              </div>
            </div>
            <span className="font-mono text-xs tabular-nums">
              {item.used.toLocaleString("ru")}/{item.limit.toLocaleString("ru")}
            </span>
          </div>
        )
      })}
    </Card>
  )
}
