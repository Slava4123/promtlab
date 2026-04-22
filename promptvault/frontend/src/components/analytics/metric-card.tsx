import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { cn } from "@/lib/utils"

interface MetricCardProps {
  title: string
  value: string | number
  subtitle?: string
  // delta в процентах: >0 → ↑зелёный, <0 → ↓красный, 0 → ≡серый, null → «—» (нет базы).
  delta?: number | null
  className?: string
}

export function MetricCard({ title, value, subtitle, delta, className }: MetricCardProps) {
  return (
    <Card className={className}>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-semibold tabular-nums">{value}</div>
        {(subtitle || delta !== undefined) && (
          <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
            {delta !== undefined && <DeltaBadge delta={delta} />}
            {subtitle && <span>{subtitle}</span>}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function DeltaBadge({ delta }: { delta: number | null }) {
  if (delta === null) {
    return <span className="inline-flex items-center font-medium text-muted-foreground">—</span>
  }
  if (delta === 0) {
    return <span className="inline-flex items-center font-medium text-muted-foreground">≡ 0%</span>
  }
  const up = delta > 0
  return (
    <span
      className={cn(
        "inline-flex items-center font-medium",
        up ? "text-emerald-600 dark:text-emerald-400" : "text-rose-600 dark:text-rose-400",
      )}
      aria-label={`изменение ${up ? "рост" : "падение"} ${Math.abs(delta)} процентов`}
    >
      {up ? "↑" : "↓"} {up ? "+" : ""}{delta}%
    </span>
  )
}
