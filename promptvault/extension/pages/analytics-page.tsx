import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { ArrowLeft, BarChart3, Loader2, TrendingUp, Sparkles, RefreshCw } from "lucide-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Button } from "../components/ui/button"
import { useToast } from "../components/ui/toaster"
import { sendBg } from "../lib/bg-client"
import { cn } from "../lib/utils"
import { ApiError } from "../lib/types"
import type { AnalyticsRange } from "../lib/api"

const RANGES: { id: AnalyticsRange; label: string }[] = [
  { id: "7d", label: "7д" },
  { id: "30d", label: "30д" },
  { id: "90d", label: "90д" },
  { id: "365d", label: "1г" },
]

export function AnalyticsPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const qc = useQueryClient()
  const [range, setRange] = useState<AnalyticsRange>("30d")

  const analyticsQuery = useQuery({
    queryKey: ["analytics", "personal", range],
    queryFn: () => sendBg({ type: "api.getPersonalAnalytics", range }),
    staleTime: 60_000,
  })

  const insightsQuery = useQuery({
    queryKey: ["analytics", "insights"],
    queryFn: () => sendBg({ type: "api.getInsights" }),
    staleTime: 5 * 60_000,
    retry: false,
  })

  const refreshMut = useMutation({
    mutationFn: () => sendBg({ type: "api.refreshInsights" }),
    onSuccess: (data) => {
      qc.setQueryData(["analytics", "insights"], data)
      toast({ title: "Инсайты обновлены", variant: "success" })
    },
    onError: (err: Error) => {
      const isRateLimited = err instanceof ApiError && err.code === "rate_limited"
      toast({
        title: isRateLimited ? "Лимит обновлений — раз в час" : "Не удалось обновить",
        description: !isRateLimited ? err.message : undefined,
        variant: "error",
      })
    },
  })

  if (analyticsQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  const data = analyticsQuery.data
  const totals = data?.totals
  const usageByDay = data?.usage_by_day ?? []
  const topPrompts = data?.top_prompts ?? []
  const insights = insightsQuery.data?.items ?? []

  // Max value for chart scaling
  const maxUsage = Math.max(1, ...usageByDay.map((d) => d.count))

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Аналитика</h2>
        <div className="flex gap-0.5 rounded-md bg-(--color-muted)/40 p-0.5">
          {RANGES.map((r) => (
            <button
              key={r.id}
              type="button"
              onClick={() => setRange(r.id)}
              className={cn(
                "px-2 py-0.5 text-[10px] rounded transition-colors",
                range === r.id
                  ? "bg-(--color-background) text-(--color-foreground) shadow-sm"
                  : "text-(--color-muted-foreground) hover:text-(--color-foreground)",
              )}
            >
              {r.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        {!data ? (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <BarChart3 className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Нет данных</p>
            <p className="text-[10px] text-(--color-muted-foreground)">
              Используйте промпты на AI-сайтах, чтобы появилась статистика.
            </p>
          </div>
        ) : (
          <>
            {/* Totals */}
            <section className="grid grid-cols-3 gap-2">
              <MetricCard label="Использований" value={totals?.uses ?? 0} />
              <MetricCard label="Создано" value={totals?.created ?? 0} />
              <MetricCard label="Просмотров share" value={totals?.share_views ?? 0} />
            </section>

            {/* Usage chart (simple SVG bar chart — без recharts чтобы не раздувать chunk) */}
            {usageByDay.length > 0 && (
              <section>
                <h3 className="mb-2 flex items-center gap-1.5 text-xs font-semibold">
                  <TrendingUp className="h-3.5 w-3.5" />
                  Использований по дням
                </h3>
                <div className="rounded-md border border-(--color-border) bg-(--color-card) p-2">
                  <div className="flex items-end gap-0.5 h-24">
                    {usageByDay.slice(-30).map((d) => (
                      <div
                        key={d.date}
                        className="flex-1 rounded-t bg-(--color-primary)/60 hover:bg-(--color-primary)"
                        style={{ height: `${(d.count / maxUsage) * 100}%`, minHeight: "2px" }}
                        title={`${d.date}: ${d.count}`}
                      />
                    ))}
                  </div>
                  <div className="mt-1 flex justify-between text-[9px] text-(--color-muted-foreground)">
                    <span>{usageByDay[0]?.date.slice(5) ?? ""}</span>
                    <span>{usageByDay[usageByDay.length - 1]?.date.slice(5) ?? ""}</span>
                  </div>
                </div>
              </section>
            )}

            {/* Top prompts */}
            {topPrompts.length > 0 && (
              <section>
                <h3 className="mb-2 text-xs font-semibold">Топ промптов</h3>
                <ul className="space-y-1">
                  {topPrompts.slice(0, 10).map((p, i) => (
                    <li
                      key={p.id}
                      className="flex items-center gap-2 rounded-md border border-(--color-border) bg-(--color-card) px-2 py-1.5 text-xs hover:bg-(--color-muted)/40 cursor-pointer"
                      onClick={() => navigate(`/prompts/${p.id}`)}
                    >
                      <span className="w-5 text-center text-[10px] font-mono text-(--color-muted-foreground)">
                        #{i + 1}
                      </span>
                      <span className="flex-1 truncate">{p.title}</span>
                      <span className="rounded bg-(--color-primary)/15 px-1.5 py-0.5 text-[9px] font-mono text-(--color-primary)">
                        {p.usage_count}×
                      </span>
                    </li>
                  ))}
                </ul>
              </section>
            )}

            {/* Smart Insights (Max-only) */}
            <section>
              <div className="mb-2 flex items-center justify-between">
                <h3 className="flex items-center gap-1.5 text-xs font-semibold">
                  <Sparkles className="h-3.5 w-3.5 text-(--color-primary)" />
                  Smart Insights
                </h3>
                {!insightsQuery.isError && insights.length > 0 && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => refreshMut.mutate()}
                    disabled={refreshMut.isPending}
                    className="h-6 text-[10px] gap-1 px-2"
                  >
                    <RefreshCw className={cn("h-3 w-3", refreshMut.isPending && "animate-spin")} />
                    Обновить
                  </Button>
                )}
              </div>
              {insightsQuery.isPending ? (
                <div className="rounded-md border border-(--color-border) bg-(--color-card) p-3 text-center">
                  <Loader2 className="mx-auto h-4 w-4 animate-spin text-(--color-muted-foreground)" />
                </div>
              ) : insightsQuery.isError ? (
                <div className="rounded-md border border-(--color-border) bg-(--color-muted)/30 p-3 text-[10px] text-(--color-muted-foreground)">
                  Smart Insights доступны на тарифе <strong>Max</strong>. Откройте /pricing чтобы апгрейднуться.
                </div>
              ) : insights.length === 0 ? (
                <div className="rounded-md border border-(--color-border) bg-(--color-muted)/30 p-3 text-[10px] text-(--color-muted-foreground)">
                  Insights пока пусты — используйте больше промптов.
                </div>
              ) : (
                <ul className="space-y-1.5">
                  {insights.map((ins, i) => (
                    <li
                      key={i}
                      className="rounded-md border border-(--color-border) bg-(--color-card) p-2 text-xs"
                    >
                      <div className="font-medium">{ins.title}</div>
                      <p className="mt-0.5 text-[10px] text-(--color-muted-foreground)">
                        {ins.description}
                      </p>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </>
        )}
      </div>
    </div>
  )
}

function MetricCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-md border border-(--color-border) bg-(--color-card) p-2">
      <div className="text-[10px] uppercase tracking-wide text-(--color-muted-foreground)">
        {label}
      </div>
      <div className="mt-1 text-lg font-bold">{value.toLocaleString("ru-RU")}</div>
    </div>
  )
}
