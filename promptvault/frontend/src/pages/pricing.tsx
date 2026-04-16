import { Check, Sparkles, Zap, Crown, Loader2, type LucideIcon } from "lucide-react"
import { PageLayout } from "@/components/layout/page-layout"
import { Skeleton } from "@/components/ui/skeleton"
import { usePlans, useCheckout, useDowngrade } from "@/hooks/use-subscription"
import { useAuthStore } from "@/stores/auth-store"
import type { Plan } from "@/api/types"

const planIcons: Record<string, LucideIcon> = {
  free: Zap,
  pro: Sparkles,
  max: Crown,
}

const planColors: Record<string, string> = {
  free: "#6366f1",
  pro: "#8b5cf6",
  max: "#f59e0b",
}

const planDescriptions: Record<string, string> = {
  free: "Для знакомства с платформой",
  pro: "Для активной работы с промптами",
  max: "Максимум возможностей для команд",
}

function formatLimit(value: number, suffix: string = ""): string {
  return value === -1 ? "Безлимит" : `${value}${suffix}`
}

function planFeatures(plan: Plan): string[] {
  const features: string[] = []
  features.push(
    plan.max_prompts === -1
      ? "Безлимитные промпты"
      : `До ${plan.max_prompts} промптов`,
  )
  features.push(
    plan.max_collections === -1
      ? "Безлимитные коллекции"
      : `${plan.max_collections} коллекции`,
  )
  features.push(
    plan.max_ai_requests_daily === -1
      ? "Безлимитные AI-запросы"
      : plan.ai_requests_is_total
        ? `${plan.max_ai_requests_daily} AI-запросов всего`
        : `${plan.max_ai_requests_daily} AI-запросов в день`,
  )
  features.push(
    plan.max_teams === -1
      ? "Безлимитные команды"
      : `${plan.max_teams} ${plan.max_teams === 1 ? "команда" : "команд"} (до ${formatLimit(plan.max_team_members)} участников)`,
  )
  features.push(
    plan.max_share_links === -1
      ? "Безлимитный шаринг"
      : `${plan.max_share_links} публичных ссылок`,
  )
  features.push(
    plan.max_ext_uses_daily === -1
      ? "Безлимитные вставки (расширение)"
      : `${plan.max_ext_uses_daily} вставок/день (расширение)`,
  )
  features.push(
    plan.max_mcp_uses_daily === -1
      ? "Безлимитные MCP-вызовы"
      : `${plan.max_mcp_uses_daily} MCP-вызовов/день`,
  )
  if (plan.features.includes("priority_support")) {
    features.push("Приоритетная поддержка")
  }
  return features
}

function formatPrice(priceKop: number): string {
  if (priceKop === 0) return "0"
  return (priceKop / 100).toLocaleString("ru-RU")
}

// dailyPrice — цена в рублях за день для anchor-копи в CTA/ROI блоке.
function dailyPrice(priceKop: number, periodDays: number): number {
  if (priceKop === 0 || periodDays === 0) return 0
  return Math.round(priceKop / 100 / periodDays)
}

// ctaLabel — value-ориентированный CTA вместо generic "Перейти на Pro".
function ctaLabel(plan: Plan): string {
  if (plan.price_kop === 0) return "Остаться на Free"
  const perDay = dailyPrice(plan.price_kop, plan.period_days)
  return `Получить ${plan.name} за ${perDay}₽/день`
}

export default function Pricing() {
  const { data: plans, isLoading, error } = usePlans()
  const checkout = useCheckout()
  const downgrade = useDowngrade()
  const currentPlanId = useAuthStore((s) => s.user?.plan_id ?? "free")

  return (
    <PageLayout
      title="Тарифы"
      description="Выберите план, который подходит вам"
    >
      {isLoading && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-96 rounded-2xl" />
          ))}
        </div>
      )}

      {error && (
        <div className="text-center text-sm text-destructive">
          Не удалось загрузить тарифы
        </div>
      )}

      {plans && (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {plans.map((plan) => {
              const Icon = planIcons[plan.id] ?? Zap
              const color = planColors[plan.id] ?? "#6366f1"
              const isPopular = plan.id === "pro"
              const isBestValue = plan.id === "max"
              const isCurrent = currentPlanId === plan.id
              const features = planFeatures(plan)
              const perDay = dailyPrice(plan.price_kop, plan.period_days)

              return (
                <div
                  key={plan.id}
                  className={`relative flex flex-col rounded-2xl border p-6 transition-colors ${
                    isPopular
                      ? "border-violet-500/30 shadow-lg shadow-violet-500/5"
                      : "border-border"
                  }`}
                >
                  {isPopular && (
                    <div className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-violet-600 px-3 py-0.5 text-[0.7rem] font-medium text-white">
                      Популярный
                    </div>
                  )}
                  {isBestValue && (
                    <div className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-amber-500 px-3 py-0.5 text-[0.7rem] font-medium text-white">
                      Лучшая цена
                    </div>
                  )}

                  <div className="mb-4 flex items-center gap-3">
                    <div
                      className="flex h-10 w-10 items-center justify-center rounded-xl"
                      style={{
                        background: `${color}15`,
                        boxShadow: `inset 0 0 0 1px ${color}25`,
                      }}
                    >
                      <Icon className="h-5 w-5" style={{ color }} />
                    </div>
                    <div>
                      <h3 className="text-[0.95rem] font-semibold text-foreground">
                        {plan.name}
                      </h3>
                      <p className="text-[0.72rem] text-muted-foreground">
                        {planDescriptions[plan.id] ?? ""}
                      </p>
                    </div>
                  </div>

                  <div className="mb-5">
                    <div className="flex items-baseline gap-1">
                      <span className="text-3xl font-bold tracking-tight text-foreground">
                        {formatPrice(plan.price_kop)} ₽
                      </span>
                      <span className="text-[0.8rem] text-muted-foreground">
                        / {plan.period_days === 0 ? "навсегда" : "в месяц"}
                      </span>
                    </div>
                    {perDay > 0 && (
                      <p className="mt-1 text-[0.72rem] text-muted-foreground">
                        ≈ {perDay}₽ в день — дешевле чашки кофе
                      </p>
                    )}
                  </div>

                  <ul className="mb-6 flex-1 space-y-2.5">
                    {features.map((feature) => (
                      <li
                        key={feature}
                        className="flex items-start gap-2 text-[0.8rem]"
                      >
                        <Check
                          className="mt-0.5 h-3.5 w-3.5 shrink-0"
                          style={{ color }}
                        />
                        <span className="text-foreground">{feature}</span>
                      </li>
                    ))}
                  </ul>

                  <button
                    disabled={isCurrent || checkout.isPending || downgrade.isPending}
                    onClick={() => {
                      if (isCurrent) return
                      if (plan.id === "free") {
                        downgrade.mutate()
                      } else {
                        checkout.mutate(plan.id)
                      }
                    }}
                    className={`flex h-11 w-full items-center justify-center rounded-lg text-[0.85rem] font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-60 ${
                      isCurrent
                        ? "border border-border bg-muted/30 text-muted-foreground"
                        : isPopular
                          ? "text-white"
                          : "border border-border bg-card text-foreground hover:bg-muted/50"
                    }`}
                    style={
                      !isCurrent && isPopular
                        ? { background: "var(--brand-gradient)" }
                        : undefined
                    }
                  >
                    {isCurrent ? (
                      "Текущий план"
                    ) : checkout.isPending || downgrade.isPending ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      ctaLabel(plan)
                    )}
                  </button>
                </div>
              )
            })}
          </div>

          <div className="mt-8 rounded-2xl border border-border bg-card/50 p-6">
            <h2 className="mb-3 text-[0.95rem] font-semibold text-foreground">
              Как ПромтЛаб Pro сравнивается с ChatGPT Plus?
            </h2>
            <div className="grid gap-3 sm:grid-cols-3">
              <div className="rounded-lg border border-border/60 p-3">
                <p className="mb-1 text-[0.7rem] uppercase tracking-wide text-muted-foreground">ChatGPT Plus</p>
                <p className="text-lg font-semibold text-foreground">~1 800 ₽/мес</p>
                <p className="mt-1 text-[0.72rem] text-muted-foreground">
                  $20 + сложности оплаты из РФ
                </p>
              </div>
              <div
                className="rounded-lg border p-3"
                style={{ borderColor: `${planColors.pro}50`, background: `${planColors.pro}08` }}
              >
                <p className="mb-1 text-[0.7rem] uppercase tracking-wide" style={{ color: planColors.pro }}>
                  ПромтЛаб Pro
                </p>
                <p className="text-lg font-semibold text-foreground">599 ₽/мес</p>
                <p className="mt-1 text-[0.72rem] text-muted-foreground">
                  В 3× дешевле + MCP + расширение + AI-улучшение
                </p>
              </div>
              <div className="rounded-lg border border-border/60 p-3">
                <p className="mb-1 text-[0.7rem] uppercase tracking-wide text-muted-foreground">Экономия</p>
                <p className="text-lg font-semibold text-emerald-500">~1 200 ₽/мес</p>
                <p className="mt-1 text-[0.72rem] text-muted-foreground">
                  14 400 ₽ в год — полтора месяца Max
                </p>
              </div>
            </div>
          </div>

          <div className="text-center">
            <p className="text-[0.75rem] text-muted-foreground">
              Оплата через Т-Банк. Подписку можно отменить в любой момент.
              {" "}Оплачивая, вы принимаете{" "}
              <a href="/legal/offer" target="_blank" rel="noopener noreferrer" className="underline hover:text-foreground">
                публичную оферту
              </a>.
            </p>
          </div>
        </>
      )}
    </PageLayout>
  )
}
