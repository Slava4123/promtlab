import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts"
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { DEFAULT_USAGE_CHART_CONFIG } from "@/components/analytics/usage-chart-config"
import { formatDayShort } from "@/lib/date-format"
import type { UsagePoint } from "@/api/analytics"

interface UsageChartProps {
  title: string
  data: UsagePoint[]
  emptyLabel?: string
  chartConfig?: ChartConfig
}

export function UsageChart({
  title,
  data,
  emptyLabel = "Пока нет данных за этот период",
  chartConfig = DEFAULT_USAGE_CHART_CONFIG,
}: UsageChartProps) {
  // Backend даёт day как "2026-04-14T00:00:00Z" — оставляем ISO, для tick'ов формат короткий.
  const formattedData = data.map((p) => ({
    day: p.day.slice(0, 10), // "2026-04-14"
    count: p.count,
  }))

  const total = formattedData.reduce((s, p) => s + p.count, 0)

  return (
    <Card className="min-w-0">
      <CardHeader className="flex flex-row items-center justify-between gap-2 space-y-0">
        <CardTitle className="text-base">{title}</CardTitle>
        {total > 0 && (
          <span className="text-sm font-medium tabular-nums text-muted-foreground">
            {total.toLocaleString("ru")} всего
          </span>
        )}
      </CardHeader>
      <CardContent className="pb-2">
        {formattedData.length === 0 ? (
          <div className="flex h-[120px] items-center justify-center text-sm text-muted-foreground">
            {emptyLabel}
          </div>
        ) : (
          // aspect-auto переопределяет default `aspect-video` от shadcn
          // ChartContainer. Без override на mobile (≤375px) chart форсил
          // min-width = height × 16/9 ≈ 427px и вылазил за viewport.
          <ChartContainer config={chartConfig} className="aspect-auto h-[240px] w-full">
            {/* margin без top — иначе сверху графика появлялся лишний воздух
                ~8px, при этом снизу был «честный» 16px от Card.py-4. На глаз
                это читалось как «график прижат к низу». Сейчас recharts даёт
                симметричный padding 5px сверху и снизу chart-зоны. */}
            <AreaChart data={formattedData} margin={{ left: 12, right: 12, top: 0, bottom: 0 }}>
              <defs>
                {/* Линейный градиент даёт классический analytics-вид «насыщенный
                    верх → почти прозрачный низ», вместо плоской 30% заливки.
                    stopColor подцепляет --color-count, который ChartContainer
                    выставляет из chartConfig. */}
                <linearGradient id="usageFill" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--color-count)" stopOpacity={0.55} />
                  <stop offset="95%" stopColor="var(--color-count)" stopOpacity={0.05} />
                </linearGradient>
              </defs>
              <CartesianGrid vertical={false} strokeDasharray="3 3" className="stroke-foreground/10" />
              <XAxis
                dataKey="day"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                tickFormatter={formatDayShort}
              />
              <YAxis
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                allowDecimals={false}
                width={32}
              />
              <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="line" />} />
              <Area
                dataKey="count"
                type="natural"
                fill="url(#usageFill)"
                stroke="var(--color-count)"
                strokeWidth={2}
              />
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}
