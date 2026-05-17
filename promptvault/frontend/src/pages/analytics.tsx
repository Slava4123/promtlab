import { useState, useMemo } from "react"
import { Loader2, Download, Activity, FileText, Eye, Trophy } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuthStore } from "@/stores/auth-store"
import { usePersonalAnalytics, useInsights } from "@/hooks/use-analytics"
import { useStreak } from "@/hooks/use-streaks"
import { computeDelta, downloadAnalyticsCSV, type AnalyticsRange } from "@/api/analytics"
import { UsageChart } from "@/components/analytics/usage-chart"
import { TopPromptsTable } from "@/components/analytics/top-prompts-table"
import { RangePicker } from "@/components/analytics/range-picker"
import { UpgradeGate } from "@/components/analytics/upgrade-gate"
import { InsightsPanel } from "@/components/analytics/insights-panel"
import { InsightsLockedCard } from "@/components/analytics/insights-locked-card"
import { NarrativeBanner } from "@/components/analytics/narrative-banner"
import { KpiCard } from "@/components/analytics/kpi-card"
import { ActivityHeatmap } from "@/components/analytics/activity-heatmap"
import { ModelsDonut } from "@/components/analytics/models-donut"
import { StreakTracker } from "@/components/analytics/streak-tracker"
import { CompactQuotas } from "@/components/analytics/compact-quotas"
import { buildNarrative } from "@/lib/analytics-narrative"
import { toast } from "sonner"

// Phase 14 C.2 + analytics redesign 2026-05-17: /analytics — личный dashboard
// в формате Bento Grid. Three-state Pro Insights teaser сохранён:
//  - Free: 7-дневное окно, UpgradeGate Pro, без CSV, без Smart Insights
//  - Pro: до 90 дней, CSV export, 2 insight types (unused + duplicates)
//  - Max: до 365 дней, CSV, все 7 insight types
export default function AnalyticsPage() {
  const user = useAuthStore((s) => s.user)
  const planId = user?.plan_id ?? "free"
  const isMax = planId.startsWith("max")
  const isPaid = planId.startsWith("pro") || isMax

  const [range, setRange] = useState<AnalyticsRange>("7d")

  const { data, isLoading, isError } = usePersonalAnalytics(range)
  const insightsQuery = useInsights(isPaid)
  const streakQuery = useStreak()

  const usageSparkline = useMemo(
    () => data?.usage_per_day?.map((p) => p.count) ?? [],
    [data],
  )
  const createdSparkline = useMemo(
    () => data?.prompts_created_per_day?.map((p) => p.count) ?? [],
    [data],
  )
  const sharedSparkline = useMemo(
    () => data?.share_views_per_day?.map((p) => p.count) ?? [],
    [data],
  )
  const narrative = useMemo(
    () => (data ? buildNarrative(data, insightsQuery.data?.items ?? null) : null),
    [data, insightsQuery.data],
  )
  const streakSegment = streakQuery.data
    ? `streak ${streakQuery.data.current_streak} ${pluralStreak(streakQuery.data.current_streak)}`
    : null
  const narrativeFinal = narrative ? { ...narrative, streak: streakSegment } : null

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
    <div className="container mx-auto space-y-4 px-4 py-8">
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
          {isPaid && (
            <Button variant="outline" size="sm" onClick={handleExport}>
              <Download className="size-4" />
              CSV
            </Button>
          )}
        </div>
      </div>

      {isLoading || !data ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[0, 1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-28 w-full" />
          ))}
        </div>
      ) : (
        <>
          {/* AI Narrative Banner */}
          {narrativeFinal && <NarrativeBanner segments={narrativeFinal} />}

          {/* KPI Strip — 4 cards с sparklines */}
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <KpiCard
              label="Использования"
              value={data.totals_current.uses.toLocaleString("ru")}
              delta={computeDelta(data.totals_current.uses, data.totals_previous.uses)}
              sparkline={usageSparkline}
              icon={Activity}
            />
            <KpiCard
              label="Новых промптов"
              value={data.totals_current.created.toLocaleString("ru")}
              delta={computeDelta(data.totals_current.created, data.totals_previous.created)}
              sparkline={createdSparkline}
              icon={FileText}
            />
            <KpiCard
              label="Просмотров ссылок"
              value={data.totals_current.share_views.toLocaleString("ru")}
              delta={
                isPaid
                  ? computeDelta(data.totals_current.share_views, data.totals_previous.share_views)
                  : null
              }
              sparkline={isPaid ? sharedSparkline : undefined}
              icon={Eye}
            />
            {streakQuery.data ? (
              <StreakTracker
                current={streakQuery.data.current_streak}
                longest={streakQuery.data.longest_streak}
                activeToday={streakQuery.data.active_today}
              />
            ) : (
              <KpiCard
                label="Топ-промпт"
                value={data.top_prompts[0]?.uses?.toLocaleString("ru") ?? "—"}
                icon={Trophy}
              />
            )}
          </div>

          {/* Smart Insights three-state */}
          {!isPaid && (
            <UpgradeGate
              title="Подсказки — на тарифе Pro"
              description="Забытые промпты и дубликаты помогут навести порядок. Полный набор — в Max."
              targetPlan="Pro"
            />
          )}

          {isPaid && insightsQuery.isLoading && (
            <div className="flex items-center justify-center py-6">
              <Loader2 className="size-5 animate-spin text-muted-foreground" />
            </div>
          )}

          {isPaid && insightsQuery.data && (
            <section className="space-y-3">
              <h2 className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
                Стоит сделать сегодня
              </h2>
              <InsightsPanel insights={insightsQuery.data.items} />
              {!isMax && (
                <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-3">
                  <InsightsLockedCard title="Растёт" description="Промпты, использование которых выросло за 7 дней." />
                  <InsightsLockedCard title="Падает" description="Промпты, которые перестали активно использоваться." />
                  <InsightsLockedCard title="Часто правят" description="Топ промптов по количеству версий." />
                  <InsightsLockedCard title="Orphan-теги" description="Теги без промптов для уборки." />
                  <InsightsLockedCard title="Пустые коллекции" description="Коллекции без промптов." />
                </div>
              )}
            </section>
          )}

          {/* Bento Grid main charts */}
          <div className="grid gap-3 lg:grid-cols-6 lg:auto-rows-[90px]">
            <div className="lg:col-span-4 lg:row-span-3">
              <UsageChart title="Использование по дням" data={data.usage_per_day} />
            </div>
            <div className="lg:col-span-2 lg:row-span-3">
              <ActivityHeatmap points={data.usage_per_day} />
            </div>
            <div className="lg:col-span-2 lg:row-span-2">
              <ModelsDonut data={data.usage_by_model} />
            </div>
            <div className="lg:col-span-4 lg:row-span-2">
              <TopPromptsTable title="Топ-10 промптов" prompts={data.top_prompts} />
            </div>
          </div>

          {/* Compact Quotas */}
          <CompactQuotas quotas={data.quotas} />

          {/* Upgrade gate для Free — расширенная история */}
          {!isPaid && (
            <UpgradeGate
              title="Больше истории на Pro"
              description="До 90 дней на Pro, до 365 на Max. Плюс экспорт CSV и подробные метрики."
              targetPlan="Pro"
            />
          )}
        </>
      )}
    </div>
  )
}

function pluralStreak(n: number): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return "день"
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return "дня"
  return "дней"
}
