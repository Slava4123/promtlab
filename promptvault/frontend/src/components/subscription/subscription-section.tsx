import { useState } from "react"
import { AlertCircle, CreditCard, ExternalLink, PauseCircle } from "lucide-react"
import { useNavigate } from "react-router"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { PlanBadge } from "./plan-badge"
import { UsageMeters } from "./usage-meters"
import { PauseDialog } from "./pause-dialog"
import { CancelDialog } from "./cancel-dialog"
import {
  useSubscription,
  useUsage,
  useCancelSubscription,
  usePauseSubscription,
  useResumeSubscription,
  useSetAutoRenew,
} from "@/hooks/use-subscription"
import { useAuthStore } from "@/stores/auth-store"
import type { PlanID } from "@/api/types"

function formatDateRu(iso: string): string {
  return new Date(iso).toLocaleDateString("ru-RU", {
    day: "numeric",
    month: "long",
    year: "numeric",
  })
}

export function SubscriptionSection() {
  const navigate = useNavigate()
  const planId = useAuthStore((s) => s.user?.plan_id ?? "free") as PlanID
  const { data: subscription, isLoading: subLoading } = useSubscription()
  const { data: usage, isLoading: usageLoading } = useUsage()
  const cancelMutation = useCancelSubscription()
  const pauseMutation = usePauseSubscription()
  const resumeMutation = useResumeSubscription()
  const autoRenewMutation = useSetAutoRenew()
  const [cancelOpen, setCancelOpen] = useState(false)
  const [pauseOpen, setPauseOpen] = useState(false)

  const isLoading = subLoading || usageLoading
  const isPaused = subscription?.status === "paused"
  const isActivePaid =
    subscription?.status === "active" &&
    subscription?.plan?.price_kop !== undefined &&
    subscription.plan.price_kop > 0

  return (
    <section className="space-y-4">
      <div className="flex items-center gap-2">
        <CreditCard className="h-5 w-5 text-muted-foreground" />
        <h2 className="text-lg font-semibold">Подписка</h2>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-3/4" />
        </div>
      ) : (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <PlanBadge planId={planId} />
              {subscription && (
                <span className="text-sm text-muted-foreground">
                  до{" "}
                  {new Date(subscription.current_period_end).toLocaleDateString(
                    "ru-RU",
                    { day: "numeric", month: "long", year: "numeric" },
                  )}
                </span>
              )}
            </div>
          </div>

          {subscription?.status === "past_due" && (
            <div
              role="alert"
              aria-live="polite"
              className="flex items-start gap-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800 dark:border-red-800 dark:bg-red-900/20 dark:text-red-200"
            >
              <AlertCircle aria-hidden="true" className="mt-0.5 h-4 w-4 shrink-0" />
              <div className="flex-1 space-y-2">
                <div>
                  <span className="font-medium">Не удалось продлить подписку.</span>{" "}
                  Возможно, недостаточно средств или карта истекла. Мы попробуем списать
                  ещё раз через 24 часа (не более 3 попыток). Доступ сохраняется до{" "}
                  {new Date(subscription.current_period_end).toLocaleDateString("ru-RU", {
                    day: "numeric",
                    month: "long",
                  })}
                  .
                </div>
                <Button size="sm" variant="outline" onClick={() => navigate("/pricing")}>
                  Обновить способ оплаты
                </Button>
              </div>
            </div>
          )}

          {subscription?.cancel_at_period_end && (
            <div className="rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-200">
              Подписка будет отменена {formatDateRu(subscription.current_period_end)}
            </div>
          )}

          {isPaused && subscription?.paused_until && (
            <div
              role="status"
              aria-live="polite"
              className="flex items-start gap-3 rounded-md border border-sky-200 bg-sky-50 p-3 text-sm text-sky-900 dark:border-sky-800 dark:bg-sky-900/20 dark:text-sky-200"
            >
              <PauseCircle aria-hidden="true" className="mt-0.5 h-4 w-4 shrink-0" />
              <div className="flex-1 space-y-2">
                <div>
                  <span className="font-medium">Подписка приостановлена.</span>{" "}
                  Возобновим автоматически {formatDateRu(subscription.paused_until)}.
                  Оставшиеся дни сохранятся.
                </div>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => resumeMutation.mutate()}
                  disabled={resumeMutation.isPending}
                >
                  Возобновить сейчас
                </Button>
              </div>
            </div>
          )}

          {subscription && !subscription.cancel_at_period_end && !isPaused && (
            <label className="flex items-start gap-3 rounded-md border border-border bg-muted/20 p-3 text-sm">
              <input
                type="checkbox"
                checked={subscription.auto_renew}
                disabled={autoRenewMutation.isPending}
                onChange={(e) => autoRenewMutation.mutate(e.target.checked)}
                className="mt-0.5 h-4 w-4 cursor-pointer accent-brand"
              />
              <span className="flex-1">
                <span className="font-medium text-foreground">Автопродление</span>
                <span className="ml-2 text-muted-foreground">
                  {subscription.auto_renew
                    ? "Подписка продлится автоматически за 2 дня до окончания."
                    : "Подписка истечёт в конце периода — потребуется ручная оплата."}
                </span>
              </span>
            </label>
          )}

          {usage && <UsageMeters usage={usage} />}

          <div className="flex flex-wrap gap-2">
            {planId === "free" ? (
              <Button size="sm" onClick={() => navigate("/pricing")}>
                Получить Pro за 19₽/день
              </Button>
            ) : (
              !subscription?.cancel_at_period_end && !isPaused && (
                <>
                  {isActivePaid && (
                    <Button variant="outline" size="sm" onClick={() => setPauseOpen(true)}>
                      Приостановить
                    </Button>
                  )}
                  <Button variant="outline" size="sm" onClick={() => setCancelOpen(true)}>
                    Отменить подписку
                  </Button>
                </>
              )
            )}
            <Button variant="ghost" size="sm" onClick={() => navigate("/pricing")}>
              Все тарифы <ExternalLink className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      )}

      <CancelDialog
        open={cancelOpen}
        onOpenChange={setCancelOpen}
        onConfirm={(input) => {
          cancelMutation.mutate(input, {
            onSettled: () => setCancelOpen(false),
          })
        }}
        isPending={cancelMutation.isPending}
      />

      <PauseDialog
        open={pauseOpen}
        onOpenChange={setPauseOpen}
        onConfirm={(months) => {
          pauseMutation.mutate(months, {
            onSettled: () => setPauseOpen(false),
          })
        }}
        isPending={pauseMutation.isPending}
      />
    </section>
  )
}
