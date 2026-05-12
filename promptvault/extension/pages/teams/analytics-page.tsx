import { useNavigate, useParams } from "react-router-dom"
import { ArrowLeft, BarChart3, Loader2, TrendingUp, Users } from "lucide-react"
import { useQuery } from "@tanstack/react-query"
import { useState, useMemo } from "react"
import { Button } from "../../components/ui/button"
import { sendBg } from "../../lib/bg-client"
import { useTeam } from "../../hooks/use-teams-crud"
import { cn } from "../../lib/utils"
import { pluralAfterDo } from "@pv/shared/utils/plural"
import type { AnalyticsRange } from "../../lib/api"

const RANGES: { value: AnalyticsRange; label: string }[] = [
  { value: "7d", label: "7д" },
  { value: "30d", label: "30д" },
  { value: "90d", label: "90д" },
  { value: "365d", label: "год" },
]

export function TeamAnalyticsPage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const [range, setRange] = useState<AnalyticsRange>("30d")

  const teamQuery = useTeam(slug ?? null)
  const team = teamQuery.data

  const analytics = useQuery({
    queryKey: ["team-analytics", team?.id, range],
    queryFn: () =>
      sendBg({ type: "api.getTeamAnalytics", teamId: team!.id, range }),
    enabled: Boolean(team?.id),
    staleTime: 60_000,
  })

  const data = analytics.data
  const totals = data?.totals_current
  const totalsPrev = data?.totals_previous

  const usesDelta = useMemo(() => {
    if (!totals || !totalsPrev) return null
    if (totalsPrev.uses === 0) return null
    return Math.round(((totals.uses - totalsPrev.uses) / totalsPrev.uses) * 100)
  }, [totals, totalsPrev])

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 truncate text-sm font-semibold">
          Аналитика: {team?.name ?? "команда"}
        </h2>
      </div>

      {/* Range selector */}
      <div className="flex gap-1 border-b border-(--color-border) px-3 py-2">
        {RANGES.map((r) => (
          <button
            key={r.value}
            type="button"
            onClick={() => setRange(r.value)}
            className={cn(
              "rounded-md px-2 py-1 text-[10px] font-medium",
              range === r.value
                ? "bg-(--color-primary) text-(--color-primary-foreground)"
                : "text-(--color-muted-foreground) hover:bg-(--color-muted)",
            )}
          >
            {r.label}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {analytics.isPending || !data ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : (
          <>
            {/* Totals */}
            <section className="grid grid-cols-2 gap-2">
              <MetricCard
                icon={TrendingUp}
                label="Использований"
                value={totals?.uses ?? 0}
                delta={usesDelta}
              />
              <MetricCard
                icon={BarChart3}
                label="Создано"
                value={totals?.created ?? 0}
              />
            </section>

            {/* Top prompts */}
            {data.top_prompts.length > 0 && (
              <section className="space-y-1.5">
                <h3 className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                  Топ промптов
                </h3>
                <ol className="space-y-1">
                  {data.top_prompts.slice(0, 5).map((p, i) => (
                    <li
                      key={p.prompt_id}
                      className="flex items-center gap-2 rounded-md border border-(--color-border) bg-(--color-card) p-2 text-xs"
                    >
                      <span className="w-4 text-center text-[10px] font-medium text-(--color-muted-foreground)">
                        {i + 1}.
                      </span>
                      <span className="flex-1 truncate">{p.title}</span>
                      <span className="font-mono text-[10px] text-(--color-muted-foreground)">
                        {p.uses}×
                      </span>
                    </li>
                  ))}
                </ol>
              </section>
            )}

            {/* Contributors */}
            {data.contributors.length > 0 && (
              <section className="space-y-1.5">
                <h3 className="flex items-center gap-1.5 text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                  <Users className="h-3 w-3" />
                  Участники
                </h3>
                <ol className="space-y-1">
                  {data.contributors.slice(0, 8).map((c) => (
                    <li
                      key={c.user_id}
                      className="rounded-md border border-(--color-border) bg-(--color-card) p-2 text-xs"
                    >
                      <div className="flex items-center justify-between">
                        <span className="truncate font-medium">{c.name || c.email}</span>
                        <span className="font-mono text-[10px] text-(--color-muted-foreground)">
                          {c.uses}×
                        </span>
                      </div>
                      <div className="mt-0.5 flex gap-2 text-[9px] text-(--color-muted-foreground)">
                        <span>+{c.prompts_created} {pluralAfterDo(c.prompts_created, "промпт", "промпта", "промптов")}</span>
                        <span>~{c.prompts_edited} {pluralAfterDo(c.prompts_edited, "правка", "правки", "правок")}</span>
                      </div>
                    </li>
                  ))}
                </ol>
              </section>
            )}

            {/* Models */}
            {data.usage_by_model.length > 0 && (
              <section className="space-y-1.5">
                <h3 className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                  По моделям
                </h3>
                <ul className="space-y-0.5">
                  {data.usage_by_model.slice(0, 5).map((m) => (
                    <li
                      key={m.model || "unknown"}
                      className="flex items-center justify-between text-[11px]"
                    >
                      <span className="font-mono">{m.model || "—"}</span>
                      <span className="text-(--color-muted-foreground)">{m.uses}×</span>
                    </li>
                  ))}
                </ul>
              </section>
            )}

            {totals && totals.uses === 0 && (
              <p className="py-6 text-center text-[10px] text-(--color-muted-foreground)">
                За выбранный период использований не было.
              </p>
            )}
          </>
        )}
      </div>
    </div>
  )
}

interface MetricCardProps {
  icon: React.ComponentType<{ className?: string }>
  label: string
  value: number
  delta?: number | null
}

function MetricCard({ icon: Icon, label, value, delta }: MetricCardProps) {
  return (
    <div className="rounded-md border border-(--color-border) bg-(--color-card) p-3">
      <div className="flex items-center gap-1.5">
        <Icon className="h-3 w-3 text-(--color-primary)" />
        <span className="text-[10px] uppercase tracking-wide text-(--color-muted-foreground)">
          {label}
        </span>
      </div>
      <div className="mt-1 text-lg font-semibold">{value}</div>
      {delta !== null && delta !== undefined && (
        <div
          className={cn(
            "text-[10px]",
            delta > 0 ? "text-emerald-500" : delta < 0 ? "text-(--color-destructive)" : "text-(--color-muted-foreground)",
          )}
        >
          {delta > 0 ? "+" : ""}{delta}% к прошлому периоду
        </div>
      )}
    </div>
  )
}
