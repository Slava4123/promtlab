import { PieChart, Pie, Cell, ResponsiveContainer } from "recharts"
import { PieChart as PieIcon } from "lucide-react"
import { Card } from "@/components/ui/card"
import type { ModelUsageRow } from "@/api/analytics"
import { colorFor, labelFor, DEFAULT_COLOR } from "./model-colors"

interface ModelsDonutProps {
  data: ModelUsageRow[]
}

// ModelsDonut — donut chart для распределения по моделям.
// Top-6 + «Другие» хвост. Используем shared MODEL_COLORS palette.
// Recharts PieChart с innerRadius=60% для donut эффекта.
export function ModelsDonut({ data }: ModelsDonutProps) {
  const total = data.reduce((s, r) => s + r.uses, 0)

  if (total === 0) {
    return (
      <Card className="p-4">
        <div className="mb-2 flex items-center gap-2">
          <PieIcon className="size-[18px] text-muted-foreground" aria-hidden="true" />
          <h3 className="text-sm font-semibold">Модели</h3>
        </div>
        <p className="text-xs text-muted-foreground">Пока нет данных</p>
      </Card>
    )
  }

  const sorted = [...data].sort((a, b) => b.uses - a.uses)
  const top = sorted.slice(0, 6)
  const tail = sorted.slice(6)
  const tailTotal = tail.reduce((s, r) => s + r.uses, 0)
  const display: ModelUsageRow[] =
    tailTotal > 0 ? [...top, { model: "__other__", uses: tailTotal }] : top

  return (
    <Card className="p-4">
      <div className="mb-3 flex items-center gap-2">
        <PieIcon className="size-[18px] text-muted-foreground" aria-hidden="true" />
        <h3 className="text-sm font-semibold">Модели</h3>
      </div>
      <div className="flex items-center gap-3">
        <div className="size-[90px] shrink-0">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={display.map((r) => ({ name: r.model, value: r.uses }))}
                dataKey="value"
                innerRadius="60%"
                outerRadius="100%"
                paddingAngle={2}
              >
                {display.map((r) => (
                  <Cell
                    key={r.model}
                    fill={r.model === "__other__" ? DEFAULT_COLOR : colorFor(r.model)}
                  />
                ))}
              </Pie>
            </PieChart>
          </ResponsiveContainer>
        </div>
        <ul className="flex-1 space-y-1 text-xs">
          {display.map((row) => {
            const pct = Math.round((row.uses / total) * 100)
            const color = row.model === "__other__" ? DEFAULT_COLOR : colorFor(row.model)
            const label = row.model === "__other__" ? "Другие" : labelFor(row.model)
            return (
              <li key={row.model} className="flex items-center gap-2">
                <span className="size-2 rounded-full" style={{ backgroundColor: color }} />
                <span className="flex-1 truncate">{label}</span>
                <span className="tabular-nums text-muted-foreground">{pct}%</span>
              </li>
            )
          })}
        </ul>
      </div>
    </Card>
  )
}
