import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts"
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { UsagePoint } from "@/api/analytics"

const chartConfig = {
  count: {
    label: "Использования",
    color: "var(--chart-1)",
  },
} satisfies ChartConfig

interface UsageChartProps {
  title: string
  data: UsagePoint[]
  emptyLabel?: string
}

export function UsageChart({ title, data, emptyLabel = "Пока нет данных за этот период" }: UsageChartProps) {
  // Backend даёт day как "2026-04-14T00:00:00Z" — оставляем ISO, для tick'ов формат короткий.
  const formattedData = data.map((p) => ({
    day: p.day.slice(0, 10), // "2026-04-14"
    count: p.count,
  }))

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {formattedData.length === 0 ? (
          <div className="flex h-[120px] items-center justify-center text-sm text-muted-foreground">
            {emptyLabel}
          </div>
        ) : (
          <ChartContainer config={chartConfig} className="h-[240px] w-full">
            <AreaChart data={formattedData} margin={{ left: 12, right: 12 }}>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="day"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                tickFormatter={(v) => v.slice(5)} // "04-14"
              />
              <YAxis tickLine={false} axisLine={false} tickMargin={8} allowDecimals={false} />
              <ChartTooltip cursor={false} content={<ChartTooltipContent indicator="line" />} />
              <Area
                dataKey="count"
                type="monotone"
                fill="var(--color-count)"
                fillOpacity={0.3}
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
