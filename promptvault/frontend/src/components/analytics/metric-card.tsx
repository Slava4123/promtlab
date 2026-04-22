import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { cn } from "@/lib/utils"

interface MetricCardProps {
  title: string
  value: string | number
  subtitle?: string
  trend?: "up" | "down" | "neutral"
  trendValue?: string
  className?: string
}

export function MetricCard({ title, value, subtitle, trend, trendValue, className }: MetricCardProps) {
  return (
    <Card className={className}>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-semibold tabular-nums">{value}</div>
        {(subtitle || trend) && (
          <div className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
            {trend && (
              <span
                className={cn(
                  "inline-flex items-center font-medium",
                  trend === "up" && "text-emerald-600 dark:text-emerald-400",
                  trend === "down" && "text-rose-600 dark:text-rose-400",
                )}
              >
                {trend === "up" ? "↑" : trend === "down" ? "↓" : "—"} {trendValue}
              </span>
            )}
            {subtitle && <span>{subtitle}</span>}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
