import type { ChartConfig } from "@/components/ui/chart"

// createUsageChartConfig — фабрика chartConfig для UsageChart с заданным
// лейблом значения. Лейбл показывается в tooltip как имя series.
// Передавайте свой лейбл для каждого инстанса, чтобы tooltip соответствовал
// содержанию (например, "Использования" для usage_per_day vs "Создано"
// для prompts_created_per_day).
//
// Файл отдельно от usage-chart.tsx из-за react-refresh/only-export-components:
// hot-reload требует, чтобы файл с компонентами не экспортировал утилиты.
export function createUsageChartConfig(valueLabel: string): ChartConfig {
  return {
    count: {
      label: valueLabel,
      color: "var(--chart-1)",
    },
  } satisfies ChartConfig
}

export const DEFAULT_USAGE_CHART_CONFIG = createUsageChartConfig("Использования")
