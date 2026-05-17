import { Calendar } from "lucide-react"
import { Card } from "@/components/ui/card"
import type { UsagePoint } from "@/api/analytics"

interface ActivityHeatmapProps {
  points: UsagePoint[]
}

// ActivityHeatmap — GitHub-style 4-week grid (28 cells, 7 cols × 4 rows).
// Opacity по count, нормализуется по max в наборе.
// Берёт последние 28 точек из data.usage_per_day.
export function ActivityHeatmap({ points }: ActivityHeatmapProps) {
  if (points.length === 0) {
    return (
      <Card className="p-4">
        <div className="mb-2 flex items-center gap-2">
          <Calendar className="size-[18px] text-violet-500" aria-hidden="true" />
          <h3 className="text-sm font-semibold">Активность 4 недели</h3>
        </div>
        <p className="text-xs text-muted-foreground">Пока нет активности — создайте промпт</p>
      </Card>
    )
  }

  const slice = points.slice(-28)
  const max = Math.max(...slice.map((p) => p.count), 1)

  return (
    <Card className="p-4">
      <div className="mb-3 flex items-center gap-2">
        <Calendar className="size-[18px] text-violet-500" aria-hidden="true" />
        <h3 className="text-sm font-semibold">Активность 4 недели</h3>
      </div>
      <div className="grid grid-cols-7 gap-1.5">
        {slice.map((p) => {
          const opacity = p.count === 0 ? 0.08 : 0.2 + (p.count / max) * 0.8
          return (
            <span
              key={p.day}
              data-cell
              title={`${p.day}: ${p.count}`}
              className="aspect-square rounded-sm bg-violet-500"
              style={{ opacity }}
            />
          )
        })}
      </div>
    </Card>
  )
}
