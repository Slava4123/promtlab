import { useState, useMemo } from "react"
import { Loader2, Download } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuthStore } from "@/stores/auth-store"
import { usePersonalAnalytics, useInsights } from "@/hooks/use-analytics"
import { downloadAnalyticsCSV, type AnalyticsRange } from "@/api/analytics"
import { MetricCard } from "@/components/analytics/metric-card"
import { UsageChart } from "@/components/analytics/usage-chart"
import { TopPromptsTable } from "@/components/analytics/top-prompts-table"
import { QuotaProgress } from "@/components/analytics/quota-progress"
import { RangePicker } from "@/components/analytics/range-picker"
import { UpgradeGate } from "@/components/analytics/upgrade-gate"
import { InsightsPanel } from "@/components/analytics/insights-panel"
import { toast } from "sonner"

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
  const insightsQuery = useInsights(isMax)

  const totalUses = useMemo(
    () => data?.usage_per_day.reduce((s, p) => s + p.count, 0) ?? 0,
    [data],
  )
  const totalCreated = useMemo(
    () => data?.prompts_created_per_day.reduce((s, p) => s + p.count, 0) ?? 0,
    [data],
  )
  const totalShareViews = useMemo(
    () => data?.share_views_per_day.reduce((s, p) => s + p.count, 0) ?? 0,
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
              subtitle={`за ${data.range}`}
            />
            <MetricCard
              title="Новых промптов"
              value={totalCreated.toLocaleString("ru")}
              subtitle={`создано за ${data.range}`}
            />
            <MetricCard
              title="Просмотров share-ссылок"
              value={totalShareViews.toLocaleString("ru")}
              subtitle={isPaid ? `за ${data.range}` : "Доступно на Pro+"}
            />
            <MetricCard
              title="Топ-промпт"
              value={data.top_prompts[0]?.uses.toLocaleString("ru") ?? "—"}
              subtitle={data.top_prompts[0]?.title ?? "—"}
            />
          </div>

          {/* Графики */}
          <div className="grid gap-4 lg:grid-cols-2">
            <UsageChart title="Использование по дням" data={data.usage_per_day} />
            <UsageChart title="Создание промптов по дням" data={data.prompts_created_per_day} />
          </div>

          {/* Топ-таблицы */}
          <div className="grid gap-4 lg:grid-cols-2">
            <TopPromptsTable title="Топ-10 промптов по использованию" prompts={data.top_prompts} />
            <TopPromptsTable
              title="Топ промптов по просмотрам share-ссылок"
              prompts={data.top_shared}
              metricLabel="Просмотров"
            />
          </div>

          {/* Квоты */}
          {data.quotas ? (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <QuotaProgress title="Промпты" quota={data.quotas.prompts} />
              <QuotaProgress title="Коллекции" quota={data.quotas.collections} />
              <QuotaProgress
                title="Share-ссылок сегодня"
                quota={data.quotas.daily_shares_today}
                format={(u, l) => (l === -1 ? `${u} / без лимита` : `${u} / ${l}`)}
              />
              <QuotaProgress title="MCP-вызовов сегодня" quota={data.quotas.mcp_uses_today} />
            </div>
          ) : null}

          {/* Smart Insights — Max only */}
          {isMax ? (
            insightsQuery.isLoading ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="size-5 animate-spin text-muted-foreground" />
              </div>
            ) : insightsQuery.data ? (
              <InsightsPanel insights={insightsQuery.data.items} />
            ) : null
          ) : (
            <UpgradeGate
              title="Smart Insights — на тарифе Max"
              description="Автоматически находим забытые, популярные и похожие промпты. Обновляется раз в сутки."
              targetPlan="Max"
            />
          )}

          {/* Upgrade gate для Free */}
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
