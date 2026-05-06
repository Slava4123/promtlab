import { useMemo, useState } from "react"
import { Check, Sparkles, Zap, Crown, Loader2, type LucideIcon } from "lucide-react"
import { PageLayout } from "@/components/layout/page-layout"
import { Skeleton } from "@/components/ui/skeleton"
import { DowngradePreviewDialog } from "@/components/subscription/downgrade-preview-dialog"
import {
  useCheckout,
  useDowngrade,
  useDowngradePreview,
  usePlans,
} from "@/hooks/use-subscription"
import { useAuthStore } from "@/stores/auth-store"
import type { Plan, PlanID } from "@/api/types"

// basePlanId — приводит pro_yearly → pro, max_yearly → max. Нужно для
// единообразного маппинга иконок/цветов/описаний и для сравнения с текущим
// планом юзера (чтобы Pro monthly и Pro yearly считались "тем же тарифом"
// при выборе подходящего таба).
function basePlanId(id: PlanID): "free" | "pro" | "max" {
  if (id === "pro" || id === "pro_yearly") return "pro"
  if (id === "max" || id === "max_yearly") return "max"
  return "free"
}

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

type Billing = "monthly" | "yearly"

// Русское склонение для слова «сохранённый» — единственный/мн.ч. родительный.
// 1 → сохранённый, 2-4 → сохранённых, 5+ / 11-14 → сохранённых.
function plurExec(n: number): string {
  const mod10 = n % 10, mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return "сохранённый"
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return "сохранённых"
  return "сохранённых"
}

// Форматируем число с разделителем разрядов (10 000, 1 000). Если поле
// отсутствует в ответе backend'а (старый бинарник без свежих миграций) —
// показываем «—» вместо краша через toLocaleString на undefined.
function formatNumber(value: number | undefined | null): string {
  if (typeof value !== "number" || !Number.isFinite(value)) return "—"
  return value.toLocaleString("ru-RU")
}

function planFeatures(plan: Plan): string[] {
  const features: string[] = []
  const base = basePlanId(plan.id)
  features.push(`До ${formatNumber(plan.max_prompts)} промптов`)
  features.push(`До ${formatNumber(plan.max_collections)} коллекций`)
  const teams = plan.max_teams ?? 0
  features.push(
    `${formatNumber(plan.max_teams)} ${teams === 1 ? "команда" : "команд"} (до ${formatNumber(plan.max_team_members)} участников)`,
  )
  // Phase 16-Y: квот на share-ссылки больше нет — они живут по TTL (30 дней
  // default, до 1 года). Анти-абуз — общий per-user rate-limit на /api.
  features.push(`Публичные ссылки на промпты с TTL 30 дней`)
  features.push(`${formatNumber(plan.max_ext_uses_daily)} вставок/день (расширение)`)
  features.push(`${formatNumber(plan.max_mcp_uses_daily)} MCP-вызовов/день`)
  // Phase 16: chains только при включённом флаге — синхронно с App.tsx и
  // app-sidebar.tsx, чтобы dark launch не объявлял фичу заранее.
  if (import.meta.env.VITE_CHAINS_ENABLED === "true") {
    features.push(
      `Цепочки: до ${formatNumber(plan.max_chains)} штук × ${formatNumber(plan.max_steps_per_chain)} шагов`,
    )
    if (plan.max_saved_executions > 0) {
      const word = plurExec(plan.max_saved_executions)
      features.push(`История запусков цепочек: ${formatNumber(plan.max_saved_executions)} ${word}`)
    }
    if (base === "max") {
      features.push("Условные ветвления в цепочках")
    }
  }
  // Phase 14: retention аналитики и флагманские фичи.
  if (base === "free") {
    features.push("Аналитика: 7 дней истории")
  } else if (base === "pro") {
    features.push("Аналитика: 90 дней истории + экспорт в CSV")
    features.push("Активность команды и история промптов")
  } else if (base === "max") {
    features.push("Аналитика: 365 дней истории + экспорт в CSV + API")
    features.push("Умные инсайты: забытые, популярные, дубликаты")
    features.push("Брендинг публичных ссылок (логотип команды)")
  }
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
// Free карточка для Pro/Max юзера = downgrade («Перейти на Free»);
// для Free юзера = no-op (кнопка disabled, текст «Текущий план», см. JSX),
// fallback «Остаться на Free» — только если состояние неожиданное.
function ctaLabel(plan: Plan, currentPlanId: PlanID): string {
  if (plan.price_kop === 0) {
    return currentPlanId === "free" ? "Остаться на Free" : "Перейти на Free"
  }
  const perDay = dailyPrice(plan.price_kop, plan.period_days)
  return `Получить ${plan.name} за ${perDay}₽/день`
}

// yearlyAnchor — для yearly планов показываем зачёркнутую цену monthly×12
// и процент экономии. Если monthly-аналог не найден (unlikely — планы
// приходят одним списком), возвращаем null и анкор не рендерится.
function yearlyAnchor(plan: Plan, plans: Plan[]): { wasKop: number; savedPct: number } | null {
  if (plan.period_days < 300) return null
  const base = basePlanId(plan.id)
  const monthly = plans.find((p) => p.id === base)
  if (!monthly || monthly.price_kop === 0) return null
  const wasKop = monthly.price_kop * 12
  if (wasKop <= plan.price_kop) return null
  const savedPct = Math.round(((wasKop - plan.price_kop) / wasKop) * 100)
  return { wasKop, savedPct }
}

export default function Pricing() {
  const { data: plans, isLoading, error } = usePlans()
  const checkout = useCheckout()
  const downgrade = useDowngrade()
  const currentPlanId = useAuthStore((s) => s.user?.plan_id ?? "free") as PlanID

  // Билинг-цикл: если текущий план юзера yearly — открываем сразу yearly-таб.
  const [billing, setBilling] = useState<Billing>(() =>
    currentPlanId.endsWith("_yearly") ? "yearly" : "monthly",
  )

  // M-10: downgrade preview. Открываем диалог после refetch — чтобы показать
  // конкретные warnings до того, как юзер подтвердит.
  const [downgradeOpen, setDowngradeOpen] = useState(false)
  const downgradePreview = useDowngradePreview("free")

  // Фильтруем планы под выбранный цикл: free всегда, остальные — по суффиксу
  // `_yearly` в ID. Раньше различали по period_days >= 300, но это ломалось
  // в dev-режиме когда BILLING_FAST_DEV=true и period_days приходилось
  // снижать до 1-2 для быстрых проверок expire/renewal — yearly-планы
  // тогда не попадали ни в одну вкладку.
  const visiblePlans = useMemo(() => {
    if (!plans) return []
    const isYearly = billing === "yearly"
    return plans.filter((p) => {
      if (p.price_kop === 0) return true
      const planIsYearly = p.id.endsWith("_yearly")
      return isYearly ? planIsYearly : !planIsYearly
    })
  }, [plans, billing])

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
          {/* Billing toggle: Monthly | Yearly −10% */}
          <div className="mb-6 flex justify-center">
            <div
              role="tablist"
              aria-label="Период оплаты"
              className="inline-flex rounded-full border border-border bg-muted/30 p-1 text-[0.8rem]"
            >
              <button
                role="tab"
                aria-selected={billing === "monthly"}
                onClick={() => setBilling("monthly")}
                className={`rounded-full px-4 py-1.5 font-medium transition-colors ${
                  billing === "monthly"
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                Ежемесячно
              </button>
              <button
                role="tab"
                aria-selected={billing === "yearly"}
                onClick={() => setBilling("yearly")}
                className={`flex items-center gap-2 rounded-full px-4 py-1.5 font-medium transition-colors ${
                  billing === "yearly"
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                Ежегодно
                <span className="rounded-full bg-emerald-500/15 px-2 py-0.5 text-[0.65rem] font-semibold text-emerald-600 dark:text-emerald-400">
                  −10%
                </span>
              </button>
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {visiblePlans.map((plan) => {
              const base = basePlanId(plan.id)
              const Icon = planIcons[base] ?? Zap
              const color = planColors[base] ?? "#6366f1"
              const isPopular = base === "pro"
              const isBestValue = base === "max"
              const isCurrent = currentPlanId === plan.id
              const features = planFeatures(plan)
              const perDay = dailyPrice(plan.price_kop, plan.period_days)
              const anchor = yearlyAnchor(plan, plans)

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
                        {planDescriptions[base] ?? ""}
                      </p>
                    </div>
                  </div>

                  <div className="mb-5">
                    <div className="flex items-baseline gap-1">
                      <span className="text-3xl font-bold tracking-tight text-foreground">
                        {formatPrice(plan.price_kop)} ₽
                      </span>
                      <span className="text-[0.8rem] text-muted-foreground">
                        /{" "}
                        {plan.period_days === 0
                          ? "навсегда"
                          : plan.period_days >= 300
                            ? "в год"
                            : "в месяц"}
                      </span>
                    </div>
                    {anchor && (
                      <p className="mt-1 text-[0.72rem]">
                        <span className="text-muted-foreground line-through">
                          {formatPrice(anchor.wasKop)} ₽
                        </span>
                        <span className="ml-2 font-medium text-emerald-600 dark:text-emerald-400">
                          экономия {anchor.savedPct}%
                        </span>
                      </p>
                    )}
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
                        // Если юзер уже на Free — ничего не делаем (кнопка disabled),
                        // иначе показываем preview перед downgrade.
                        if (currentPlanId === "free") return
                        setDowngradeOpen(true)
                        downgradePreview.refetch()
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
                      ctaLabel(plan, currentPlanId)
                    )}
                  </button>
                </div>
              )
            })}
          </div>

          <div className="mt-8 rounded-2xl border border-border bg-card/50 p-6">
            <h2 className="mb-3 text-[0.95rem] font-semibold text-foreground">
              Доступно во всех тарифах
            </h2>
            <ul className="grid gap-2 text-[0.8rem] text-muted-foreground sm:grid-cols-2">
              <li>История версий промптов с откатом</li>
              <li>Корзина: восстановление 30 дней</li>
              <li>Поиск с поддержкой опечаток (русский + английский)</li>
              <li>Серии активности и достижения</li>
              <li>Импорт/экспорт промптов</li>
              <li>Темы интерфейса (светлая/тёмная)</li>
              <li>До 5 API-ключей с ограничением прав</li>
              <li>Вход через GitHub, Google, Yandex</li>
            </ul>
          </div>

          <div className="mt-8 rounded-2xl border border-border bg-card/50 p-6">
            <h2 className="mb-3 text-[0.95rem] font-semibold text-foreground">
              Что вы получаете на Pro
            </h2>
            <div className="grid gap-3 sm:grid-cols-3">
              <div
                className="rounded-lg border p-3"
                style={{ borderColor: `${planColors.pro}50`, background: `${planColors.pro}08` }}
              >
                <p className="mb-1 text-[0.7rem] uppercase tracking-wide" style={{ color: planColors.pro }}>
                  MCP для Claude / Cursor / Cline
                </p>
                <p className="text-lg font-semibold text-foreground">Свои промпты — в любом AI-клиенте</p>
                <p className="mt-1 text-[0.72rem] text-muted-foreground">
                  Подключите PromptVault через MCP и вставляйте промпты одной командой.
                </p>
              </div>
              <div className="rounded-lg border border-border/60 p-3">
                <p className="mb-1 text-[0.7rem] uppercase tracking-wide text-muted-foreground">Команды</p>
                <p className="text-lg font-semibold text-foreground">Общая библиотека</p>
                <p className="mt-1 text-[0.72rem] text-muted-foreground">
                  Делитесь промптами внутри команды, роли owner / editor / viewer.
                </p>
              </div>
              <div className="rounded-lg border border-border/60 p-3">
                <p className="mb-1 text-[0.7rem] uppercase tracking-wide text-muted-foreground">Большие лимиты хранения</p>
                <p className="text-lg font-semibold text-foreground">Промпты, коллекции, теги</p>
                <p className="mt-1 text-[0.72rem] text-muted-foreground">
                  История версий каждого промпта, публичные ссылки, расширение для браузера.
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

      <DowngradePreviewDialog
        open={downgradeOpen}
        onOpenChange={setDowngradeOpen}
        preview={downgradePreview.data}
        isLoading={downgradePreview.isFetching}
        isPending={downgrade.isPending}
        onConfirm={() => {
          downgrade.mutate(undefined, {
            onSettled: () => setDowngradeOpen(false),
          })
        }}
      />
    </PageLayout>
  )
}
