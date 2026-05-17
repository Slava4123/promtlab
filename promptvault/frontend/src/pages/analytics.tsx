import { useState, useMemo } from "react"
import { Loader2, Download } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuthStore } from "@/stores/auth-store"
import { usePersonalAnalytics, useInsights } from "@/hooks/use-analytics"
import { computeDelta, downloadAnalyticsCSV, formatRange, type AnalyticsRange } from "@/api/analytics"
import { pluralizeRu } from "@/lib/pluralize"
import { MetricCard } from "@/components/analytics/metric-card"
import { UsageChart } from "@/components/analytics/usage-chart"
import { createUsageChartConfig } from "@/components/analytics/usage-chart-config"
import { TopPromptsTable } from "@/components/analytics/top-prompts-table"
import { QuotaProgress } from "@/components/analytics/quota-progress"
import { RangePicker } from "@/components/analytics/range-picker"
import { UpgradeGate } from "@/components/analytics/upgrade-gate"
import { InsightsPanel } from "@/components/analytics/insights-panel"
import { InsightsLockedCard } from "@/components/analytics/insights-locked-card"
import { ModelSegmentationChart } from "@/components/analytics/model-segmentation-chart"
import { toast } from "sonner"

// Per-instance chartConfig: лейбл tooltip должен соответствовать смыслу графика.
// Без явного config все инстансы UsageChart используют дефолтный "Использования",
// что давало неверный tooltip на графике создания промптов.
const CREATED_PER_DAY_CONFIG = createUsageChartConfig("Создано")

// Phase 14 C.2: /analytics — личный dashboard.
// Уровни доступа:
//  - Free: 7-дневное окно, без CSV, без Smart Insights
//  - Pro: до 90 дней, CSV export, без Smart Insights
//  - Max: до 365 дней, CSV, Smart Insights
export default function AnalyticsPage() {
  const user = useAuthStore((s) => s.user)
  const planId = user?.plan_id ?? "free"
  const isMax = planId.startsWith("max")
  const isPaid = planId.startsWith("pro") || isMax

  const [range, setRange] = useState<AnalyticsRange>("7d")

  const { data, isLoading, isError } = usePersonalAnalytics(range)
  // Pricing iteration v3: Smart Insights теперь доступны и Pro (2 типа:
  // unused_prompts + possible_duplicates), и Max (полные 7 типов). Free
  // получает UpgradeGate. Backend гейтит per-type — фронт показывает
  // locked-карточки для Pro как teaser к Max.
  const insightsQuery = useInsights(isPaid)

  const totalUses = useMemo(
    () => data?.usage_per_day?.reduce((s, p) => s + p.count, 0) ?? 0,
    [data],
  )
  const totalCreated = useMemo(
    () => data?.prompts_created_per_day?.reduce((s, p) => s + p.count, 0) ?? 0,
    [data],
  )
  const totalShareViews = useMemo(
    () => data?.share_views_per_day?.reduce((s, p) => s + p.count, 0) ?? 0,
    [data],
  )

  async function handleExport() {
    try {
      await downloadAnalyticsCSV("personal", range)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Не удалось скачать CSV")
    }
  }

  if (isError) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="mb-4 text-2xl font-bold">Аналитика</h1>
        <p className="text-destructive">Не удалось загрузить данные. Попробуйте обновить страницу.</p>
      </div>
    )
  }

  return (
    <div className="container mx-auto space-y-6 px-4 py-8">
      {/* Header */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">Аналитика</h1>
          <p className="text-sm text-muted-foreground">
            Ваше использование промптов и публичных ссылок
          </p>
        </div>
        <div className="flex items-center gap-2">
          <RangePicker value={range} onChange={setRange} planId={planId} />
          {isPaid ? (
            <Button variant="outline" size="sm" onClick={handleExport}>
              <Download className="size-4" />
              CSV
            </Button>
          ) : null}
        </div>
      </div>

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[0, 1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-28 w-full" />
          ))}
        </div>
      ) : data ? (
        <>
          {/* Метрики */}
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <MetricCard
              title="Всего использований"
              value={totalUses.toLocaleString("ru")}
              subtitle={`за ${formatRange(data.range)}`}
              delta={computeDelta(data.totals_current.uses, data.totals_previous.uses)}
            />
            <MetricCard
              title="Новых промптов"
              value={totalCreated.toLocaleString("ru")}
              subtitle={`создано за ${formatRange(data.range)}`}
              delta={computeDelta(data.totals_current.created, data.totals_previous.created)}
            />
            <MetricCard
              title="Просмотров публичных ссылок"
              value={totalShareViews.toLocaleString("ru")}
              subtitle={isPaid ? `за ${formatRange(data.range)}` : "Доступно на Pro+"}
              delta={
                isPaid
                  ? computeDelta(data.totals_current.share_views, data.totals_previous.share_views)
                  : undefined
              }
            />
            <MetricCard
              title="Топ-промпт"
              value={data.top_prompts[0]?.uses.toLocaleString("ru") ?? "—"}
              subtitle={
                data.top_prompts[0]
                  ? `${pluralizeRu(data.top_prompts[0].uses, "использование", "использования", "использований")} · ${data.top_prompts[0].title}`
                  : "Нет использованных промптов"
              }
            />
          </div>

          {/* Графики */}
          <div className="grid gap-4 lg:grid-cols-2">
            <UsageChart title="Использование по дням" data={data.usage_per_day} />
            <UsageChart
              title="Создание промптов по дням"
              data={data.prompts_created_per_day}
              chartConfig={CREATED_PER_DAY_CONFIG}
            />
          </div>

          {/* Топ-таблицы */}
          <div className="grid gap-4 lg:grid-cols-2">
            <TopPromptsTable title="Топ-10 промптов по использованию" prompts={data.top_prompts} />
            <TopPromptsTable
              title="Топ промптов по просмотрам публичных ссылок"
              prompts={data.top_shared}
              metricLabel="Просмотров"
            />
          </div>

          {/* Segmentation по AI-моделям */}
          <ModelSegmentationChart data={data.usage_by_model} />

          {/* Квоты */}
          {data.quotas ? (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <QuotaProgress title="Промпты" quota={data.quotas.prompts} />
              <QuotaProgress title="Коллекции" quota={data.quotas.collections} />
              {/* Phase 16-Y: блок share-ссылок убран — на share больше нет квот. */}
              <QuotaProgress title="MCP-вызовов сегодня" quota={data.quotas.mcp_uses_today} />
            </div>
          ) : null}

          {/* Smart Insights — three-state (Free/Pro/Max).
              Free: UpgradeGate → Pro (минимальный teaser).
              Pro: 2 типа (unused + duplicates) от backend + 5 locked-карточек как teaser к Max.
              Max: 7 типов от backend, без locked-карточек. */}
          {!isPaid && (
            <UpgradeGate
              title="Подсказки — на тарифе Pro"
              description="Забытые промпты и дубликаты помогут навести порядок. Полный набор — в Max."
              targetPlan="Pro"
            />
          )}

          {isPaid && insightsQuery.isLoading && (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="size-5 animate-spin text-muted-foreground" />
            </div>
          )}

          {isPaid && insightsQuery.data && (
            <div className="space-y-4">
              <InsightsPanel insights={insightsQuery.data.items} />

              {/* Pro юзер видит locked-карточки для 5 Max-only типов */}
              {!isMax && (
                <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
                  <InsightsLockedCard
                    title="Растущая популярность"
                    description="Промпты, использование которых выросло за 7 дней."
                  />
                  <InsightsLockedCard
                    title="Падающая популярность"
                    description="Промпты, которые перестали активно использоваться."
                  />
                  <InsightsLockedCard
                    title="Самые редактируемые"
                    description="Топ промптов по количеству версий."
                  />
                  <InsightsLockedCard
                    title="Теги без промптов"
                    description="Orphan-теги для уборки."
                  />
                  <InsightsLockedCard
                    title="Пустые коллекции"
                    description="Коллекции без промптов."
                  />
                </div>
              )}
            </div>
          )}

          {/* Upgrade gate для Free — расширенная история / CSV */}
          {!isPaid && (
            <UpgradeGate
              title="Больше истории на Pro"
              description="До 90 дней на Pro, до 365 дней на Max. Плюс экспорт CSV и подробные метрики."
              targetPlan="Pro"
            />
          )}
        </>
      ) : null}
    </div>
  )
}
