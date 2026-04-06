import { Check, Sparkles, Zap, Crown } from "lucide-react"

const plans = [
  {
    name: "Free",
    price: "0",
    period: "навсегда",
    description: "Для знакомства с платформой",
    icon: Zap,
    color: "#6366f1",
    features: [
      "До 50 промптов",
      "3 коллекции",
      "5 AI-запросов в день",
      "1 команда (до 3 участников)",
      "Версионирование промптов",
    ],
    limits: [
      "Без экспорта",
    ],
    current: true,
  },
  {
    name: "Pro",
    price: "599",
    period: "в месяц",
    description: "Для активной работы с промптами",
    icon: Sparkles,
    color: "#8b5cf6",
    popular: true,
    features: [
      "До 500 промптов",
      "Безлимитные коллекции",
      "100 AI-запросов в день",
      "5 команд (до 10 участников)",
      "Версионирование промптов",
      "Экспорт в JSON/Markdown",
      "Приоритетная поддержка",
    ],
    limits: [],
  },
  {
    name: "Max",
    price: "1 299",
    period: "в месяц",
    description: "Максимум возможностей для команд",
    icon: Crown,
    color: "#f59e0b",
    features: [
      "Безлимитные промпты",
      "Безлимитные коллекции",
      "Безлимитные AI-запросы",
      "Безлимитные команды",
      "Версионирование промптов",
      "Экспорт в JSON/Markdown",
      "Приоритетная поддержка",
      "API-доступ (скоро)",
    ],
    limits: [],
  },
]

export default function Pricing() {
  return (
    <div className="mx-auto max-w-[64rem] space-y-8">
      {/* Header */}
      <div className="text-center">
        <h1 className="text-2xl font-bold tracking-tight">Тарифы</h1>
        <p className="mt-1.5 text-[0.85rem] text-muted-foreground">
          Выберите план, который подходит вам
        </p>
      </div>

      {/* Plans grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {plans.map((plan) => {
          const Icon = plan.icon
          return (
            <div
              key={plan.name}
              className={`relative flex flex-col rounded-2xl border p-6 transition-all ${
                plan.popular
                  ? "border-violet-500/30 shadow-lg shadow-violet-500/5"
                  : "border-border"
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-violet-600 px-3 py-0.5 text-[0.7rem] font-medium text-white">
                  Популярный
                </div>
              )}

              {/* Icon + name */}
              <div className="mb-4 flex items-center gap-3">
                <div
                  className="flex h-10 w-10 items-center justify-center rounded-xl"
                  style={{
                    background: `${plan.color}15`,
                    boxShadow: `inset 0 0 0 1px ${plan.color}25`,
                  }}
                >
                  <Icon className="h-5 w-5" style={{ color: plan.color }} />
                </div>
                <div>
                  <h3 className="text-[0.95rem] font-semibold text-foreground">{plan.name}</h3>
                  <p className="text-[0.72rem] text-muted-foreground">{plan.description}</p>
                </div>
              </div>

              {/* Price */}
              <div className="mb-5">
                <div className="flex items-baseline gap-1">
                  <span className="text-3xl font-bold tracking-tight text-foreground">{plan.price} ₽</span>
                  <span className="text-[0.8rem] text-muted-foreground">/ {plan.period}</span>
                </div>
              </div>

              {/* Features */}
              <ul className="mb-6 flex-1 space-y-2.5">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-2 text-[0.8rem]">
                    <Check className="mt-0.5 h-3.5 w-3.5 shrink-0" style={{ color: plan.color }} />
                    <span className="text-foreground">{feature}</span>
                  </li>
                ))}
                {plan.limits.map((limit) => (
                  <li key={limit} className="flex items-start gap-2 text-[0.8rem] text-muted-foreground">
                    <span className="mt-0.5 h-3.5 w-3.5 shrink-0 text-center">—</span>
                    <span>{limit}</span>
                  </li>
                ))}
              </ul>

              {/* Button */}
              <button
                disabled
                className={`flex h-11 w-full items-center justify-center rounded-lg text-[0.85rem] font-medium transition-all disabled:cursor-not-allowed disabled:opacity-60 ${
                  plan.current
                    ? "border border-border bg-muted/30 text-muted-foreground"
                    : plan.popular
                      ? "text-white"
                      : "border border-border bg-card text-foreground"
                }`}
                style={
                  !plan.current && plan.popular
                    ? { background: "linear-gradient(135deg, #7c3aed, #6d28d9)" }
                    : undefined
                }
              >
                {plan.current ? "Текущий план" : "Скоро"}
              </button>
            </div>
          )
        })}
      </div>

      {/* Footer note */}
      <div className="text-center">
        <p className="text-[0.75rem] text-muted-foreground">
          Оплата через ЮKassa. Подписку можно отменить в любой момент.
        </p>
      </div>
    </div>
  )
}
