import { useState, useMemo } from "react"
import { Link, useParams, useNavigate } from "react-router-dom"
import { ArrowLeft, Download } from "lucide-react"
import { Button, buttonVariants } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useAuthStore } from "@/stores/auth-store"
import { useTeam } from "@/hooks/use-teams"
import { useTeamAnalytics } from "@/hooks/use-analytics"
import { computeDelta, downloadAnalyticsCSV, formatRange, type AnalyticsRange } from "@/api/analytics"
import { MetricCard } from "@/components/analytics/metric-card"
import { UsageChart } from "@/components/analytics/usage-chart"
import { TopPromptsTable } from "@/components/analytics/top-prompts-table"
import { ContributorsLeaderboard } from "@/components/analytics/contributors-leaderboard"
import { ModelSegmentationChart } from "@/components/analytics/model-segmentation-chart"
import { RangePicker } from "@/components/analytics/range-picker"
import { UpgradeGate } from "@/components/analytics/upgrade-gate"
import { toast } from "sonner"
import { ApiError } from "@/api/client"

// Phase 14 C.3: /teams/:slug/analytics
export default function TeamAnalyticsPage() {
  const { slug = "" } = useParams()
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const planId = user?.plan_id ?? "free"
  const isPaid = planId.startsWith("pro") || planId.startsWith("max")

  const { data: team, isLoading: teamLoading, error: teamError } = useTeam(slug)
  const [range, setRange] = useState<AnalyticsRange>("7d")
  const { data, isLoading, isError } = useTeamAnalytics(team?.id, range)

  const totalUses = useMemo(
    () => data?.usage_per_day.reduce((s, p) => s + p.count, 0) ?? 0,
    [data],
  )
  const totalCreated = useMemo(
    () => data?.prompts_created_per_day.reduce((s, p) => s + p.count, 0) ?? 0,
    [data],
  )
  const totalEdited = useMemo(
    () => data?.prompts_updated_per_day.reduce((s, p) => s + p.count, 0) ?? 0,
    [data],
  )

  async function handleExport() {
    if (!team) return
    try {
      await downloadAnalyticsCSV("team", range, team.id)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Не удалось скачать CSV")
    }
  }

  if (teamError instanceof ApiError && teamError.status === 403) {
    toast.error("Нет доступа к команде")
    navigate("/teams")
    return null
  }

  if (!isPaid) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="mb-4 text-2xl font-bold">Аналитика команды</h1>
        <UpgradeGate
          title="Аналитика команды — на тарифе Pro"
          description="Подробная статистика использования команды, топ промптов и контрибьюторов."
          targetPlan="Pro"
        />
      </div>
    )
  }

  return (
    <div className="container mx-auto space-y-6 px-4 py-8">
      {/* Header */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <Link
            to={`/teams/${slug}`}
            className={buttonVariants({ variant: "ghost", size: "sm" })}
          >
            <ArrowLeft className="size-4" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold">Аналитика: {team?.name ?? slug}</h1>
            <p className="text-sm text-muted-foreground">Метрики команды</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <RangePicker value={range} onChange={setRange} planId={planId} />
          <Button variant="outline" size="sm" onClick={handleExport} disabled={!team}>
            <Download className="size-4" />
            CSV
          </Button>
        </div>
      </div>

      {(isLoading || teamLoading) ? (
        <div className="grid gap-4 sm:grid-cols-3">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} className="h-28 w-full" />
          ))}
        </div>
      ) : isError ? (
        <p className="text-destructive">Не удалось загрузить аналитику команды.</p>
      ) : data ? (
        <>
          <div className="grid gap-4 sm:grid-cols-3">
            <MetricCard
              title="Использований"
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
              title="Обновлений"
              value={totalEdited.toLocaleString("ru")}
              subtitle={`версий за ${formatRange(data.range)}`}
              delta={computeDelta(data.totals_current.updated, data.totals_previous.updated)}
            />
          </div>

          <div className="grid gap-4 lg:grid-cols-2">
            <UsageChart title="Использование промптов команды" data={data.usage_per_day} />
            <UsageChart title="Создание промптов" data={data.prompts_created_per_day} />
          </div>

          <div className="grid gap-4 lg:grid-cols-2">
            <TopPromptsTable title="Топ промптов команды" prompts={data.top_prompts} />
            <ContributorsLeaderboard contributors={data.contributors} />
          </div>

          <ModelSegmentationChart data={data.usage_by_model} />
        </>
      ) : null}
    </div>
  )
}
