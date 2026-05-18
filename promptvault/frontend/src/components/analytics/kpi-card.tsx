import { ArrowUp, ArrowDown, type LucideIcon } from "lucide-react"
import { Card } from "@/components/ui/card"
import { Sparkline } from "./sparkline"
import { cn } from "@/lib/utils"

interface KpiCardProps {
  label: string
  value: string | number
  delta?: number | null
  sparkline?: number[]
  icon: LucideIcon
  className?: string
}

// KpiCard — расширение MetricCard: добавлены icon и sparkline.
// Layout: label сверху (uppercase muted) + icon справа, value крупно,
// delta inline с ArrowUp/Down, sparkline снизу.
// Цвета delta: raw emerald/rose для консистентности с metric-card (см. CLAUDE.md).
export function KpiCard({ label, value, delta, sparkline, icon: Icon, className }: KpiCardProps) {
  const trend = !delta ? "neutral" : delta > 0 ? "up" : "down"

  return (
    <Card className={cn("p-4", className)}>
      <div className="mb-1.5 flex items-center justify-between">
        <span className="text-[11px] uppercase tracking-wide text-muted-foreground">{label}</span>
        <Icon className="size-4 text-muted-foreground" aria-hidden="true" />
      </div>
      <div className="flex items-baseline gap-2">
        <span className="text-2xl font-bold tabular-nums">{value}</span>
        <DeltaInline delta={delta} />
      </div>
      {sparkline && sparkline.length > 0 && (
        <div className="mt-2">
          <Sparkline points={sparkline} trend={trend} />
        </div>
      )}
    </Card>
  )
}

function DeltaInline({ delta }: { delta: number | null | undefined }) {
  if (delta === null) return <span className="text-xs text-muted-foreground">—</span>
  if (delta === undefined || delta === 0)
    return (
      <span className="rounded-md bg-foreground/[0.04] px-1.5 py-0.5 text-xs font-medium text-muted-foreground">
        0%
      </span>
    )
  const up = delta > 0
  return (
    <span
      className={cn(
        // Pill-бейдж по Tremor-паттерну (bg-emerald-100 / dark bg-emerald-500/15):
        // явный visual marker лучше чистого цвета текста, его легче заметить
        // в потоке KPI-карточек.
        "inline-flex items-center gap-0.5 rounded-md px-1.5 py-0.5 text-xs font-medium",
        up
          ? "bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400"
          : "bg-rose-100 text-rose-700 dark:bg-rose-500/15 dark:text-rose-400",
      )}
    >
      {up ? <ArrowUp className="size-3" /> : <ArrowDown className="size-3" />}
      {Math.abs(delta)}%
    </span>
  )
}
