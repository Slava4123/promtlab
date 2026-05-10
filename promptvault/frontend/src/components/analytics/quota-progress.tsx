import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { cn } from "@/lib/utils"
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

// Минимальный визуальный штрих, когда используется хоть что-то, но pct
// настолько мал (≈ 0.2% при 1/500), что бар выглядел бы пустым. 2% — это
// тонкая, но различимая засечка в начале шкалы.
const MIN_VISIBLE_PCT = 2

function computeBarValue(used: number, limit: number): number {
  if (limit <= 0) return 0
  const pct = (used / limit) * 100
  if (used > 0 && pct > 0 && pct < MIN_VISIBLE_PCT) return MIN_VISIBLE_PCT
  return Math.min(100, pct)
}

export function QuotaProgress({ title, quota, format = defaultFormat }: QuotaProgressProps) {
  const rawPct = quota.limit > 0 ? (quota.used / quota.limit) * 100 : 0
  const isOver = quota.used > quota.limit && quota.limit > 0
  const nearLimit = !isOver && rawPct >= 80
  const barValue = computeBarValue(quota.used, quota.limit)

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center justify-between gap-2 text-sm font-medium text-muted-foreground">
          <span className="truncate">{title}</span>
          <span className="flex items-center gap-2">
            {isOver && (
              <span
                className="rounded-full bg-rose-100 px-2 py-0.5 text-[0.65rem] font-semibold text-rose-700 dark:bg-rose-950/60 dark:text-rose-300"
                aria-label="Лимит превышен"
              >
                Превышено
              </span>
            )}
            <span
              className={cn(
                "tabular-nums",
                isOver && "text-rose-600 dark:text-rose-400",
                nearLimit && "text-amber-600 dark:text-amber-400",
              )}
            >
              {format(quota.used, quota.limit)}
            </span>
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Progress
          value={barValue}
          aria-label={`${title}: ${format(quota.used, quota.limit)}`}
        />
      </CardContent>
    </Card>
  )
}
