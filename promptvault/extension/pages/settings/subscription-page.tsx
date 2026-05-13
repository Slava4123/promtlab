import { useNavigate } from "react-router-dom"
import { ArrowLeft, CreditCard, Loader2, Pause, Play, XCircle, ExternalLink } from "lucide-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { useToast } from "../../components/ui/toaster"
import { sendBg } from "../../lib/bg-client"
import { qk } from "../../lib/query-keys"
import { formatDate } from "@pv/shared/utils/format-date"

const PLAN_LABELS: Record<string, string> = {
  free: "Free",
  pro: "Pro",
  pro_yearly: "Pro (год)",
  max: "Max",
  max_yearly: "Max (год)",
}

export function SubscriptionPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const qc = useQueryClient()

  const subQuery = useQuery({
    queryKey: qk.subscription,
    queryFn: () => sendBg({ type: "api.getCurrentSubscription" }),
    staleTime: 60_000,
  })
  const usageQuery = useQuery({
    queryKey: qk.usage,
    queryFn: () => sendBg({ type: "api.getUsageSummary" }),
    staleTime: 60_000,
  })

  const cancelMut = useMutation({
    mutationFn: () => sendBg({ type: "api.cancelSubscription" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.subscription })
      toast({ title: "Подписка отменена", description: "Действует до конца периода", variant: "info" })
    },
    onError: (err: Error) =>
      toast({ title: "Не удалось отменить", description: err.message, variant: "error" }),
  })
  const pauseMut = useMutation({
    mutationFn: () => sendBg({ type: "api.pauseSubscription" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.subscription })
      toast({ title: "Подписка приостановлена", variant: "info" })
    },
  })
  const resumeMut = useMutation({
    mutationFn: () => sendBg({ type: "api.resumeSubscription" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: qk.subscription })
      toast({ title: "Подписка возобновлена", variant: "success" })
    },
  })

  async function openUpgrade() {
    const { apiBase } = await import("../../lib/storage").then((m) => m.getSettings())
    const { openWebPage } = await import("../../lib/utils")
    openWebPage(apiBase, "/pricing?source=extension")
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Подписка</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        {subQuery.isPending ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : (
          <>
            {/* Current plan */}
            <section className="rounded-md border border-(--color-brand)/30 bg-gradient-to-br from-(--color-primary)/10 to-transparent p-3">
              <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                Текущий тариф
              </div>
              <div className="mt-1 flex items-center gap-2">
                <CreditCard className="h-5 w-5 text-(--color-brand)" />
                <span className="text-lg font-bold">
                  {subQuery.data
                    ? PLAN_LABELS[subQuery.data.plan_id] ?? subQuery.data.plan_id
                    : "Free"}
                </span>
                {subQuery.data && (
                  <span
                    className={
                      "rounded px-1.5 py-0.5 text-[10px] " +
                      (subQuery.data.status === "active"
                        ? "bg-emerald-500/10 text-emerald-500"
                        : subQuery.data.status === "paused"
                          ? "bg-amber-500/10 text-amber-500"
                          : "bg-(--color-muted) text-(--color-muted-foreground)")
                    }
                  >
                    {subQuery.data.status}
                  </span>
                )}
              </div>
              {subQuery.data?.current_period_end && (
                <div className="mt-1 text-[10px] text-(--color-muted-foreground)">
                  Действует до {formatDate(subQuery.data.current_period_end)}
                </div>
              )}
            </section>

            {/* Usage meters */}
            {usageQuery.data && (
              <section className="space-y-2">
                <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                  Использование
                </div>
                <UsageMeter
                  label="Промпты"
                  used={usageQuery.data.prompts.used}
                  limit={usageQuery.data.prompts.limit}
                />
                <UsageMeter
                  label="Коллекции"
                  used={usageQuery.data.collections.used}
                  limit={usageQuery.data.collections.limit}
                />
                <UsageMeter
                  label="Цепочки"
                  used={usageQuery.data.chains.used}
                  limit={usageQuery.data.chains.limit}
                />
                <UsageMeter
                  label="Вставок сегодня"
                  used={usageQuery.data.ext_uses_today.used}
                  limit={usageQuery.data.ext_uses_today.limit}
                />
                <UsageMeter
                  label="MCP-вызовов сегодня"
                  used={usageQuery.data.mcp_uses_today.used}
                  limit={usageQuery.data.mcp_uses_today.limit}
                />
              </section>
            )}

            {/* Actions */}
            <section className="space-y-2">
              <Button type="button" onClick={openUpgrade} className="w-full gap-1.5">
                <ExternalLink className="h-3.5 w-3.5" />
                {subQuery.data && subQuery.data.plan_id !== "free"
                  ? "Управление тарифом"
                  : "Перейти на Pro / Max"}
              </Button>
              {subQuery.data && subQuery.data.status === "active" && (
                <>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => pauseMut.mutate()}
                    disabled={pauseMut.isPending}
                    className="w-full gap-1.5"
                  >
                    <Pause className="h-3.5 w-3.5" />
                    Приостановить
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => cancelMut.mutate()}
                    disabled={cancelMut.isPending}
                    className="w-full gap-1.5 text-(--color-destructive)"
                  >
                    <XCircle className="h-3.5 w-3.5" />
                    Отменить подписку
                  </Button>
                </>
              )}
              {subQuery.data?.status === "paused" && (
                <Button
                  type="button"
                  onClick={() => resumeMut.mutate()}
                  disabled={resumeMut.isPending}
                  className="w-full gap-1.5"
                >
                  <Play className="h-3.5 w-3.5" />
                  Возобновить
                </Button>
              )}
            </section>
          </>
        )}
      </div>
    </div>
  )
}

function UsageMeter({ label, used, limit }: { label: string; used: number; limit: number }) {
  const pct = limit > 0 ? Math.min(100, (used / limit) * 100) : 0
  const warning = pct >= 90
  return (
    <div className="space-y-0.5">
      <div className="flex items-center justify-between text-[10px]">
        <span>{label}</span>
        <span className={warning ? "text-amber-500" : "text-(--color-muted-foreground)"}>
          {used} / {limit < 0 ? "∞" : limit}
        </span>
      </div>
      <div className="h-1 overflow-hidden rounded-full bg-(--color-muted)">
        <div
          className={"h-full transition-all " + (warning ? "bg-amber-500" : "bg-(--color-primary)")}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  )
}
