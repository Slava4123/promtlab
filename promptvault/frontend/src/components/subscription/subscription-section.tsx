import { useState } from "react"
import { CreditCard, ExternalLink } from "lucide-react"
import { useNavigate } from "react-router"
import { Button } from "@/components/ui/button"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { Skeleton } from "@/components/ui/skeleton"
import { PlanBadge } from "./plan-badge"
import { UsageMeters } from "./usage-meters"
import { useSubscription, useUsage, useCancelSubscription, useSetAutoRenew } from "@/hooks/use-subscription"
import { useAuthStore } from "@/stores/auth-store"

export function SubscriptionSection() {
  const navigate = useNavigate()
  const planId = useAuthStore((s) => s.user?.plan_id ?? "free") as "free" | "pro" | "max"
  const { data: subscription, isLoading: subLoading } = useSubscription()
  const { data: usage, isLoading: usageLoading } = useUsage()
  const cancelMutation = useCancelSubscription()
  const autoRenewMutation = useSetAutoRenew()
  const [cancelOpen, setCancelOpen] = useState(false)

  const isLoading = subLoading || usageLoading

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

          {subscription?.cancel_at_period_end && (
            <div className="rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-200">
              Подписка будет отменена{" "}
              {new Date(subscription.current_period_end).toLocaleDateString(
                "ru-RU",
              )}
            </div>
          )}

          {subscription && !subscription.cancel_at_period_end && (
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

          <div className="flex gap-2">
            {planId === "free" ? (
              <Button size="sm" onClick={() => navigate("/pricing")}>
                Перейти на Pro
              </Button>
            ) : (
              !subscription?.cancel_at_period_end && (
                <Button variant="outline" size="sm" onClick={() => setCancelOpen(true)}>
                  Отменить подписку
                </Button>
              )
            )}
            <Button variant="ghost" size="sm" onClick={() => navigate("/pricing")}>
              Все тарифы <ExternalLink className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={cancelOpen}
        onOpenChange={setCancelOpen}
        title="Отменить подписку?"
        description="Вы сохраните доступ до конца оплаченного периода. После этого аккаунт перейдёт на бесплатный план."
        confirmLabel="Отменить подписку"
        variant="destructive"
        onConfirm={() => {
          cancelMutation.mutate(undefined, {
            onSettled: () => setCancelOpen(false),
          })
        }}
        isPending={cancelMutation.isPending}
      />
    </section>
  )
}
