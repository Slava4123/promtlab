import { PieChart, Pie, Cell, ResponsiveContainer } from "recharts"
import { PieChart as PieIcon } from "lucide-react"
import { Card } from "@/components/ui/card"
import { pluralizeRu } from "@/lib/pluralize"
import type { ModelUsageRow } from "@/api/analytics"
import { colorFor, labelFor, DEFAULT_COLOR } from "./model-colors"

interface ModelsDonutProps {
  data: ModelUsageRow[]
}

// ModelsDonut — donut chart для распределения по моделям.
// Top-6 + «Другие» хвост. Используем shared MODEL_COLORS palette.
// Donut крупный (140px) + total в центре — Tremor-style паттерн «значение в
// центре, легенда сбоку». Раньше donut был 90px без центрального значения,
// карточка выглядела пустой при том что в ней живёт ключевая метрика.
export function ModelsDonut({ data }: ModelsDonutProps) {
  const total = data.reduce((s, r) => s + r.uses, 0)

  if (total === 0) {
    return (
      <Card className="p-5">
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
    <Card className="min-w-0 p-5">
      <div className="mb-4 flex items-center gap-2">
        <PieIcon className="size-[18px] text-muted-foreground" aria-hidden="true" />
        <h3 className="text-sm font-semibold">Модели</h3>
      </div>
      {/* flex-wrap: на mobile (<sm) donut и легенда становятся колонками,
          иначе legend сжимается до 0 и контент вылазит за карточку. */}
      <div className="flex flex-wrap items-center gap-5">
        {/* Donut с total в центре. relative + absolute overlay — стандартный
            Tremor паттерн (recharts не даёт прямого center-label API). */}
        <div className="relative size-[140px] shrink-0">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={display.map((r) => ({ name: r.model, value: r.uses }))}
                dataKey="value"
                innerRadius="68%"
                outerRadius="100%"
                paddingAngle={2}
                startAngle={90}
                endAngle={-270}
                strokeWidth={0}
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
          <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center">
            <span className="text-xl font-bold tabular-nums leading-none">
              {total.toLocaleString("ru")}
            </span>
            <span className="mt-1 text-[10px] uppercase tracking-wide text-muted-foreground">
              {pluralizeRu(total, "запрос", "запроса", "запросов")}
            </span>
          </div>
        </div>
        <ul className="flex-1 space-y-1.5 text-xs">
          {display.map((row) => {
            const pct = Math.round((row.uses / total) * 100)
            const color = row.model === "__other__" ? DEFAULT_COLOR : colorFor(row.model)
            const label = row.model === "__other__" ? "Другие" : labelFor(row.model)
            return (
              <li key={row.model} className="flex items-center gap-2">
                <span
                  className="size-2.5 shrink-0 rounded-sm"
                  style={{ backgroundColor: color }}
                  aria-hidden="true"
                />
                <span className="flex-1 truncate">{label}</span>
                <span className="font-medium tabular-nums text-muted-foreground">{pct}%</span>
              </li>
            )
          })}
        </ul>
      </div>
    </Card>
  )
}
