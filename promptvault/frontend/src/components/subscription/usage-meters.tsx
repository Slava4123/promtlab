import type { UsageSummary, QuotaInfo } from "@/api/types"

interface UsageMetersProps {
  usage: UsageSummary
  className?: string
}

// key → [базовый label, тип лимита]
// "daily" — дневной, "total" — общий, "active" — количество активных
const resourceConfig: Record<string, { label: string; type: "total" | "daily" | "active" }> = {
  prompts: { label: "Промпты", type: "total" },
  collections: { label: "Коллекции", type: "total" },
  teams: { label: "Команды", type: "total" },
  share_links: { label: "Публичные ссылки", type: "active" },
  ext_uses_today: { label: "Расширение", type: "daily" },
  mcp_uses_today: { label: "MCP", type: "daily" },
}

function getSuffix(key: string): string {
  const cfg = resourceConfig[key]
  if (!cfg) return ""

  if (cfg.type === "daily") return " (сегодня)"
  if (cfg.type === "active") return " (активные)"
  return ""
}

function Meter({ resourceKey, label, info }: { resourceKey: string; label: string; info: QuotaInfo }) {
  if (info.limit <= 0) return null

  const pct = Math.min((info.used / info.limit) * 100, 100)
  const color =
    pct >= 90 ? "bg-red-500" : pct >= 75 ? "bg-amber-500" : "bg-emerald-500"

  const suffix = getSuffix(resourceKey)

  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted-foreground">
          {label}{suffix}
        </span>
        <span className="tabular-nums font-medium">
          {info.used} / {info.limit}
        </span>
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-muted/40">
        <div className={`h-full rounded-full ${color} transition-all`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  )
}

export function UsageMeters({ usage, className = "" }: UsageMetersProps) {
  const keys = Object.keys(resourceConfig)

  return (
    <div className={`space-y-3 ${className}`}>
      {keys.map((key) => {
        const info = usage[key as keyof UsageSummary] as QuotaInfo | undefined
        if (!info || typeof info !== "object" || !("used" in info)) return null
        return <Meter key={key} resourceKey={key} label={resourceConfig[key].label} info={info} />
      })}
    </div>
  )
}
