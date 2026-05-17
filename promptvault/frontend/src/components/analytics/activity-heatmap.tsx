import { Calendar } from "lucide-react"
import { Card } from "@/components/ui/card"
import { formatDayShort } from "@/lib/date-format"
import { pluralizeRu } from "@/lib/pluralize"
import type { UsagePoint } from "@/api/analytics"

interface ActivityHeatmapProps {
  points: UsagePoint[]
}

const WINDOW_DAYS = 28

// ActivityHeatmap — GitHub-style 4-week grid (28 cells, 7 cols × 4 rows).
// Opacity по count, нормализуется по max в наборе.
// Окно фиксировано: сегодня минус 27 дней .. сегодня. Если данных меньше —
// недостающие дни добиваются нулями, чтобы пользователь всегда видел полную сетку.
export function ActivityHeatmap({ points }: ActivityHeatmapProps) {
  const cells = padToWindow(points, WINDOW_DAYS)
  const total = cells.reduce((s, c) => s + c.count, 0)
  const max = Math.max(...cells.map((c) => c.count), 1)

  if (total === 0) {
    return (
      <Card className="p-4">
        <div className="mb-3 flex items-center gap-2">
          <Calendar className="size-[18px] text-violet-500" aria-hidden="true" />
          <h3 className="text-sm font-semibold">Активность 4 недели</h3>
        </div>
        <div className="grid grid-cols-7 gap-1.5">
          {cells.map((c) => (
            <span
              key={c.day}
              data-cell
              data-day={c.day}
              aria-label={ariaLabelFor(c.day, c.count)}
              title={ariaLabelFor(c.day, c.count)}
              className="aspect-square rounded-sm bg-violet-500"
              style={{ opacity: 0.08 }}
            />
          ))}
        </div>
        <p className="mt-3 text-xs text-muted-foreground">
          Пока нет активности — создайте промпт
        </p>
      </Card>
    )
  }

  return (
    <Card className="p-4">
      <div className="mb-3 flex items-center gap-2">
        <Calendar className="size-[18px] text-violet-500" aria-hidden="true" />
        <h3 className="text-sm font-semibold">Активность 4 недели</h3>
      </div>
      <div className="grid grid-cols-7 gap-1.5">
        {cells.map((c) => {
          const opacity = c.count === 0 ? 0.08 : 0.2 + (c.count / max) * 0.8
          const label = ariaLabelFor(c.day, c.count)
          return (
            <span
              key={c.day}
              data-cell
              data-day={c.day}
              aria-label={label}
              title={label}
              className="aspect-square rounded-sm bg-violet-500"
              style={{ opacity }}
            />
          )
        })}
      </div>
    </Card>
  )
}

function ariaLabelFor(day: string, count: number): string {
  const noun = pluralizeRu(count, "использование", "использования", "использований")
  return `${formatDayShort(day)}: ${count} ${noun}`
}

// padToWindow — возвращает массив из `days` элементов, последний день — сегодня,
// первый — `today - (days - 1)`. Существующие points мерджатся по day (ISO UTC).
// Не мутирует входной массив.
function padToWindow(points: UsagePoint[], days: number): UsagePoint[] {
  const byDay = new Map(points.map((p) => [p.day, p.count]))
  const today = new Date()
  const out: UsagePoint[] = []
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(today.getDate() - i)
    const key = d.toISOString().slice(0, 10)
    out.push({ day: key, count: byDay.get(key) ?? 0 })
  }
  return out
}
