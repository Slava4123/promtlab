import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { ModelUsageRow } from "@/api/analytics"

interface ModelSegmentationChartProps {
  data: ModelUsageRow[]
  title?: string
}

// Palette для известных семейств моделей. Неузнанные модели получают серый.
// Хранение palette'а здесь (не в theme.css) — эти цвета специфичны только
// для этого chart'а и не должны утекать в другие компоненты.
const MODEL_COLORS: Array<[RegExp, string]> = [
  [/^claude/i, "#cc7a3e"], // оранж-коричневый как Anthropic brand
  [/^gpt/i, "#10a37f"],     // зелёный как OpenAI
  [/deepseek/i, "#4a7fff"], // синий
  [/gemini|google/i, "#8ab4f8"],
  [/llama|meta/i, "#0668e1"],
  [/mistral/i, "#ff7000"],
]
const DEFAULT_COLOR = "#94a3b8" // серый для «Без модели» и неопознанных

function colorFor(model: string): string {
  for (const [re, color] of MODEL_COLORS) {
    if (re.test(model)) return color
  }
  return DEFAULT_COLOR
}

// Backend агрегирует строки с пустой `prompts.model` под пустой строкой
// (через COALESCE(model_used,'') в analytics_repo.go). Для пользователя
// показываем расшифровку, а не голое "" или "Без модели", чтобы не было
// неоднозначности — это промпты, в которых пользователь не выбрал модель
// при создании, не legacy-данные и не баг.
function labelFor(model: string): string {
  return model === "" ? "Модель не указана" : model
}

const UNKNOWN_MODEL_HINT =
  "Промпты, в которых при создании не указана target-модель в редакторе"

// ModelSegmentationChart — простая горизонтальная полоса без тяжёлых recharts-
// зависимостей. Показывает долю каждой модели в общем использовании + список
// подписей с процентами.
export function ModelSegmentationChart({ data, title = "Использование по моделям" }: ModelSegmentationChartProps) {
  const total = data.reduce((sum, r) => sum + r.uses, 0)

  if (total === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{title}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">Пока нет данных. Используйте промпты, чтобы увидеть распределение.</p>
        </CardContent>
      </Card>
    )
  }

  // Top-6 моделей + «Другие» для хвоста.
  const sorted = [...data].sort((a, b) => b.uses - a.uses)
  const top = sorted.slice(0, 6)
  const tail = sorted.slice(6)
  const tailTotal = tail.reduce((sum, r) => sum + r.uses, 0)
  const display: ModelUsageRow[] = tailTotal > 0
    ? [...top, { model: "__other__", uses: tailTotal }]
    : top

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {/* Полоса-сегмент */}
        <div className="flex h-3 overflow-hidden rounded-full bg-muted">
          {display.map((row) => {
            const pct = (row.uses / total) * 100
            const color = row.model === "__other__" ? DEFAULT_COLOR : colorFor(row.model)
            const labelText = row.model === "__other__" ? "Другие" : labelFor(row.model)
            const hint = row.model === "" ? ` — ${UNKNOWN_MODEL_HINT}` : ""
            return (
              <div
                key={row.model}
                style={{ width: `${pct}%`, backgroundColor: color }}
                title={`${labelText}: ${row.uses} (${pct.toFixed(1)}%)${hint}`}
              />
            )
          })}
        </div>

        {/* Легенда */}
        <ul className="grid gap-1.5 sm:grid-cols-2">
          {display.map((row) => {
            const pct = (row.uses / total) * 100
            const color = row.model === "__other__" ? DEFAULT_COLOR : colorFor(row.model)
            const label = row.model === "__other__" ? "Другие" : labelFor(row.model)
            const isUnknown = row.model === ""
            return (
              <li
                key={row.model}
                className="flex items-center gap-2 text-xs"
                title={isUnknown ? UNKNOWN_MODEL_HINT : undefined}
              >
                <span
                  className="size-2.5 rounded-full"
                  style={{ backgroundColor: color }}
                  aria-hidden
                />
                <span className="flex-1 truncate">
                  {label}
                  {isUnknown && (
                    <span
                      className="ml-1 cursor-help text-muted-foreground/70"
                      aria-label={UNKNOWN_MODEL_HINT}
                    >
                      ⓘ
                    </span>
                  )}
                </span>
                <span className="tabular-nums text-muted-foreground">
                  {row.uses.toLocaleString("ru")} · {pct.toFixed(0)}%
                </span>
              </li>
            )
          })}
        </ul>
      </CardContent>
    </Card>
  )
}
