import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { ArrowLeft, Check, ExternalLink, Loader2, Sparkles } from "lucide-react"
import { useQuery } from "@tanstack/react-query"
import { Button } from "../components/ui/button"
import { sendBg } from "../lib/bg-client"
import { qk } from "../lib/query-keys"
import { cn } from "../lib/utils"
import type { Plan } from "../lib/types"

const PLAN_FEATURES: Record<string, string[]> = {
  free: ["До 50 промптов", "Личное пространство", "Поиск Cmd+K", "Базовые цепочки"],
  pro: [
    "До 1000 промптов",
    "Команды (до 10 человек)",
    "Smart Insights",
    "До 5 цепочек × 10 шагов",
    "Email-поддержка",
  ],
  max: [
    "Безлимит промптов",
    "Безлимит команд",
    "100 цепочек × 50 шагов",
    "Условные цепочки (ветвление)",
    "Постоянные публичные ссылки",
    "Приоритетная поддержка",
  ],
}

export function PricingPage() {
  const navigate = useNavigate()
  const [yearly, setYearly] = useState(false)
  const plansQuery = useQuery({
    queryKey: qk.plans,
    queryFn: () => sendBg({ type: "api.listPlans" }),
    staleTime: 5 * 60_000,
  })
  const subQuery = useQuery({
    queryKey: qk.subscription,
    queryFn: () => sendBg({ type: "api.getCurrentSubscription" }),
    staleTime: 60_000,
  })

  const currentPlan = subQuery.data?.plan_id ?? "free"

  async function openCheckout(planId: string) {
    const { getSettings } = await import("../lib/storage")
    const { openWebPage } = await import("../lib/utils")
    const { apiBase } = await getSettings()
    openWebPage(apiBase, `/pricing?plan=${planId}&from=extension`)
  }

  if (plansQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  // Группируем планы: показываем только base (free/pro/max), скрываем _yearly variants
  // — toggle ниже переключает на _yearly.
  const allPlans = plansQuery.data ?? []
  const displayPlans = allPlans
    .filter((p) => !p.id.endsWith("_yearly"))
    .sort((a, b) => a.sort_order - b.sort_order)

  function planFor(basePlan: Plan): Plan {
    if (!yearly) return basePlan
    const yearlyId = `${basePlan.id}_yearly`
    return allPlans.find((p) => p.id === yearlyId) ?? basePlan
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Тарифы</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {/* Billing toggle */}
        <div className="flex items-center justify-center gap-2 rounded-md bg-(--color-muted)/40 p-1 text-xs">
          <button
            type="button"
            onClick={() => setYearly(false)}
            className={cn(
              "flex-1 rounded px-3 py-1.5 transition-colors",
              !yearly && "bg-(--color-background) text-(--color-foreground) shadow-sm",
            )}
          >
            Месяц
          </button>
          <button
            type="button"
            onClick={() => setYearly(true)}
            className={cn(
              "flex-1 rounded px-3 py-1.5 transition-colors",
              yearly && "bg-(--color-background) text-(--color-foreground) shadow-sm",
            )}
          >
            Год{" "}
            <span className="ml-1 text-[9px] text-emerald-500">−10%</span>
          </button>
        </div>

        {displayPlans.map((basePlan) => {
          const plan = planFor(basePlan)
          const isCurrent = plan.id === currentPlan
          const pricePerMonth = yearly ? Math.round(plan.price_kop / 12 / 100) : plan.price_kop / 100
          const features = PLAN_FEATURES[basePlan.id] ?? []
          return (
            <div
              key={plan.id}
              className={cn(
                "rounded-md border p-3",
                isCurrent
                  ? "border-(--color-brand) bg-(--color-brand-muted)"
                  : "border-(--color-border) bg-(--color-card)",
              )}
            >
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-1.5">
                    <h3 className="text-sm font-semibold">{plan.name}</h3>
                    {basePlan.id === "pro" && (
                      <Sparkles className="h-3.5 w-3.5 text-(--color-brand)" />
                    )}
                    {isCurrent && (
                      <span className="rounded bg-(--color-brand-muted) px-1.5 py-0.5 text-[9px] uppercase tracking-wide text-(--color-brand)">
                        Текущий
                      </span>
                    )}
                  </div>
                  <div className="mt-1 flex items-baseline gap-1">
                    <span className="text-xl font-bold">{pricePerMonth}₽</span>
                    <span className="text-[10px] text-(--color-muted-foreground)">/ мес</span>
                    {yearly && plan.price_kop > 0 && (
                      <span className="ml-2 text-[10px] text-(--color-muted-foreground)">
                        ({(plan.price_kop / 100).toLocaleString("ru-RU")}₽ в год)
                      </span>
                    )}
                  </div>
                </div>
              </div>
              <ul className="mt-3 space-y-1">
                {features.map((f, i) => (
                  <li key={i} className="flex items-start gap-1.5 text-[10px]">
                    <Check className="mt-0.5 h-3 w-3 shrink-0 text-emerald-500" />
                    <span>{f}</span>
                  </li>
                ))}
              </ul>
              {!isCurrent && plan.price_kop > 0 && (
                <Button
                  type="button"
                  onClick={() => openCheckout(plan.id)}
                  className="mt-3 w-full gap-1.5"
                  size="sm"
                >
                  <ExternalLink className="h-3 w-3" />
                  Выбрать {plan.name}
                </Button>
              )}
            </div>
          )
        })}

        <p className="pt-2 text-center text-[10px] text-(--color-muted-foreground)">
          Оплата проходит на promtlabs.ru через T-Bank.
          <br />
          Все тарифы можно отменить в любой момент.
        </p>
      </div>
    </div>
  )
}
