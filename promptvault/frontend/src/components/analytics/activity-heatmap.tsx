import { useMemo } from "react"
import { Calendar } from "lucide-react"
import { Card } from "@/components/ui/card"
import { formatDayShort } from "@/lib/date-format"
import { pluralizeRu } from "@/lib/pluralize"
import { cn } from "@/lib/utils"
import type { UsagePoint } from "@/api/analytics"

interface ActivityHeatmapProps {
  points: UsagePoint[]
}

// Дизайн-цифры выровнены под defaults react-activity-calendar (источник: context7,
// /grubersjoe/react-activity-calendar, benchmark 80.5) — индустриальная норма для
// GitHub-style heatmap: блок 12px, gap 4px, radius 3px. На full-row layout
// (col-span-6, ~1300px) 53 × 16 = 848px помещается с запасом.
const WEEKS = 53
const DAYS_IN_WEEK = 7
const BLOCK = 12 // px — размер ячейки
const GAP = 4 // px — расстояние между ячейками
const RADIUS = 3 // px — скругление

const monthFmt = new Intl.DateTimeFormat("ru-RU", { month: "short" })

// Все 7 дней с короткими русскими подписями — обычная GitHub-конвенция
// (Mon/Wed/Fri) сбивала пользователей, потому что выглядела как «только три
// дня». С полной разметкой сетка читается без лишних предположений.
const WEEKDAY_LABELS = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"]

// 5 уровней насыщенности — GitHub-style ramp. Tailwind не парсит динамические
// классы, поэтому статический lookup. Light/dark варианты обеспечивают читаемость
// на обеих темах: tier 0 — нейтральный фон, дальше явное нарастание violet
// с заметным контрастом уже на первом уровне (раньше violet-200 сливался
// с фоном — заполненные дни казались пустыми).
const TIER_CLASS: Record<0 | 1 | 2 | 3 | 4, string> = {
  0: "bg-foreground/[0.06] dark:bg-foreground/[0.08]",
  1: "bg-violet-300 dark:bg-violet-900",
  2: "bg-violet-500 dark:bg-violet-700",
  3: "bg-violet-600 dark:bg-violet-500",
  4: "bg-violet-800 dark:bg-violet-300",
}

interface Cell {
  date: Date
  iso: string
  count: number
}

// ActivityHeatmap — GitHub-style 52-week grid с месячными подписями сверху,
// днями недели слева и легендой «Меньше / Больше». Окно: сегодня и 52 недели
// назад. Padding нулями для отсутствующих дней — пользователь всегда видит
// полную сетку независимо от выбранного range.
export function ActivityHeatmap({ points }: ActivityHeatmapProps) {
  const { cells, monthLabels, total, max } = useMemo(() => buildGrid(points), [points])

  return (
    <Card className="p-5 sm:p-6">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <Calendar className="size-[18px] text-violet-500" aria-hidden="true" />
          <h3 className="text-sm font-semibold">Активность за год</h3>
        </div>
        <span className="text-sm font-medium tabular-nums text-muted-foreground">
          {total.toLocaleString("ru")}{" "}
          {pluralizeRu(total, "использование", "использования", "использований")} за год
        </span>
      </div>

      {total === 0 ? (
        <EmptyState cells={cells} monthLabels={monthLabels} />
      ) : (
        <Grid cells={cells} monthLabels={monthLabels} max={max} />
      )}

      <Legend />
    </Card>
  )
}

interface GridProps {
  cells: Cell[]
  monthLabels: { col: number; label: string }[]
  max: number
}

function Grid({ cells, monthLabels, max }: GridProps) {
  const cellTrack = `repeat(${WEEKS}, ${BLOCK}px)`
  return (
    <div className="overflow-x-auto pb-1">
      <div className="flex gap-2" style={{ minWidth: WEEKS * (BLOCK + GAP) }}>
        {/* Колонка с лейблами дней недели — все 7 дней (Пн-Вс) с компактными
            русскими подписями. Выравнивание под ячейки сетки через одинаковую
            высоту и gap. */}
        <div
          className="flex shrink-0 flex-col"
          style={{ gap: `${GAP}px`, paddingTop: BLOCK + GAP + 2 }}
        >
          {WEEKDAY_LABELS.map((label, i) => (
            <div
              key={i}
              className="flex items-center text-xs leading-none text-muted-foreground"
              style={{ height: BLOCK }}
            >
              {label}
            </div>
          ))}
        </div>

        <div className="flex-1">
          {/* Месячные метки — выровнены над первой колонкой соответствующего месяца */}
          <div
            className="mb-1.5 grid"
            style={{
              gridTemplateColumns: cellTrack,
              columnGap: `${GAP}px`,
              height: BLOCK + 2,
            }}
          >
            {monthLabels.map(({ col, label }) => (
              <div
                key={`${col}-${label}`}
                className="text-xs font-medium leading-none text-muted-foreground"
                style={{ gridColumnStart: col + 1 }}
              >
                {label}
              </div>
            ))}
          </div>

          {/* Контейнер для сетки и разделителей. Положение relative нужно
              чтобы абсолютно позиционированные разделители месяцев попадали
              в правильные координаты колонок. */}
          <div className="relative">
            {/* Вертикальные разделители — тонкие линии в начале каждого
                нового месяца. Без них «расстояние между метками месяцев»
                выглядит произвольным; разделитель явно показывает границу
                и делает скан-чтение календаря «месяц-в-неделях» понятным. */}
            <div className="pointer-events-none absolute inset-0">
              {monthLabels.slice(1).map(({ col }) => (
                <div
                  key={col}
                  className="absolute top-0 bottom-0 w-px bg-foreground/[0.06]"
                  style={{ left: col * (BLOCK + GAP) - GAP / 2 }}
                  aria-hidden="true"
                />
              ))}
            </div>

            {/* Сетка ячеек: 53 колонки × 7 строк, заполняем по колонкам сверху вниз */}
            <div
              className="relative grid grid-flow-col"
              style={{
                gridTemplateColumns: cellTrack,
                gridTemplateRows: `repeat(${DAYS_IN_WEEK}, ${BLOCK}px)`,
                gap: `${GAP}px`,
              }}
            >
              {cells.map((c) => {
                const tier = tierFor(c.count, max)
                const label = ariaLabelFor(c.iso, c.count)
                return (
                  <span
                    key={c.iso}
                    data-cell
                    data-day={c.iso}
                    data-tier={tier}
                    aria-label={label}
                    title={label}
                    className={cn(
                      "transition-transform duration-150",
                      "hover:scale-125 hover:ring-2 hover:ring-violet-400 hover:ring-offset-1",
                      "hover:ring-offset-background hover:relative hover:z-10",
                      TIER_CLASS[tier],
                    )}
                    style={{
                      width: BLOCK,
                      height: BLOCK,
                      borderRadius: RADIUS,
                    }}
                  />
                )
              })}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function EmptyState({
  cells,
  monthLabels,
}: {
  cells: Cell[]
  monthLabels: { col: number; label: string }[]
}) {
  // При total=0 рисуем ту же сетку, но все ячейки tier=0 (без max-нормализации,
  // иначе деление на ноль). UX: юзер всё равно видит полный год, чтобы понять
  // visual placeholder, а пустая подсказка снизу подскажет первый шаг.
  return (
    <>
      <Grid cells={cells} monthLabels={monthLabels} max={1} />
      <p className="mt-4 text-sm text-muted-foreground">
        Пока нет активности — создайте промпт
      </p>
    </>
  )
}

function Legend() {
  return (
    <div className="mt-4 flex items-center justify-end gap-2 border-t border-foreground/5 pt-3 text-xs text-muted-foreground">
      <span>Меньше</span>
      <div className="flex" style={{ gap: `${GAP}px` }}>
        {([0, 1, 2, 3, 4] as const).map((tier) => (
          <span
            key={tier}
            className={TIER_CLASS[tier]}
            style={{
              width: BLOCK,
              height: BLOCK,
              borderRadius: RADIUS,
            }}
            aria-hidden="true"
          />
        ))}
      </div>
      <span>Больше</span>
    </div>
  )
}

function tierFor(count: number, max: number): 0 | 1 | 2 | 3 | 4 {
  if (count <= 0 || max <= 0) return 0
  const ratio = count / max
  if (ratio <= 0.25) return 1
  if (ratio <= 0.5) return 2
  if (ratio <= 0.75) return 3
  return 4
}

function ariaLabelFor(iso: string, count: number): string {
  const noun = pluralizeRu(count, "использование", "использования", "использований")
  return `${formatDayShort(iso)}: ${count} ${noun}`
}

function buildGrid(points: UsagePoint[]) {
  // Нормализуем ключ к 10-char ISO date: backend возвращает p.day как полный
  // RFC3339 timestamp "2025-05-20T00:00:00Z", а сравниваем мы с
  // d.toISOString().slice(0, 10) ниже. Без slice здесь Map.get всегда вернёт
  // undefined и все ячейки получат count=0.
  const byDay = new Map(points.map((p) => [p.day.slice(0, 10), p.count]))
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  // ISO-неделя: Пн=0, Вс=6. Учитываем, что Date.getDay() возвращает Вс=0.
  const todayWeekday = (today.getDay() + 6) % 7

  // Левый-верхний угол: понедельник 52 недели назад относительно текущей недели.
  // Это даёт ровно (52*7 + todayWeekday + 1) дней до today включительно.
  const startDate = new Date(today)
  startDate.setDate(today.getDate() - (52 * DAYS_IN_WEEK + todayWeekday))

  const cells: Cell[] = []
  for (let col = 0; col < WEEKS; col++) {
    for (let row = 0; row < DAYS_IN_WEEK; row++) {
      const d = new Date(startDate)
      d.setDate(startDate.getDate() + col * DAYS_IN_WEEK + row)
      if (d > today) continue // обрезаем будущее (последняя неполная колонка)
      const iso = d.toISOString().slice(0, 10)
      cells.push({ date: d, iso, count: byDay.get(iso) ?? 0 })
    }
  }

  // Метки месяцев: переход с одного месяца на другой в первом дне колонки.
  const monthLabels: { col: number; label: string }[] = []
  let prevMonth = -1
  for (let col = 0; col < WEEKS; col++) {
    const firstIdx = col * DAYS_IN_WEEK
    const cell = cells[firstIdx] // если будущая колонка обрезана, undefined
    if (!cell) break
    const m = cell.date.getMonth()
    if (m !== prevMonth) {
      monthLabels.push({ col, label: monthFmt.format(cell.date) })
      prevMonth = m
    }
  }

  const total = cells.reduce((s, c) => s + c.count, 0)
  const max = Math.max(...cells.map((c) => c.count), 0)

  return { cells, monthLabels, total, max }
}
