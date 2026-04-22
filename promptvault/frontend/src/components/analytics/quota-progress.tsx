import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import type { QuotaInfo } from "@/api/analytics"

interface QuotaProgressProps {
  title: string
  quota: QuotaInfo
  // Форматер отображения значения (например "5 из 50" или "5 / 50 в день").
  format?: (used: number, limit: number) => string
}

function defaultFormat(used: number, limit: number): string {
  return `${used.toLocaleString("ru")} / ${limit.toLocaleString("ru")}`
}

export function QuotaProgress({ title, quota, format = defaultFormat }: QuotaProgressProps) {
  const pct = Math.min(100, (quota.used / Math.max(1, quota.limit)) * 100)
  const nearLimit = pct >= 80

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between text-sm font-medium text-muted-foreground">
          <span>{title}</span>
          <span className={nearLimit ? "text-amber-600 dark:text-amber-400" : undefined}>
            {format(quota.used, quota.limit)}
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Progress value={pct} />
      </CardContent>
    </Card>
  )
}
