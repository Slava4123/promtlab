import { useNavigate } from "react-router-dom"
import { AlertTriangle, ArrowRight, LifeBuoy, Trash2, type LucideIcon } from "lucide-react"
import { toast } from "sonner"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { useQuotaStore } from "@/stores/quota-store"
import { useAuthStore } from "@/stores/auth-store"
import { usePlans } from "@/hooks/use-subscription"
import { PlanBadge } from "./plan-badge"
import type { Plan } from "@/api/types"

// Человеко-читаемое название ресурса для заголовка диалога.
// Должно покрывать ВСЕ quotaType, что отдаёт backend (см. quota.go newQuotaExceeded calls).
const quotaLabels: Record<string, string> = {
  prompts: "промптов",
  collections: "коллекций",
  teams: "команд",
  team_members: "участников команды",
  // Phase 16-Y: share_links / daily_shares УДАЛЕНЫ — share больше не имеет квот.
  ext_daily: "вставок через расширение на сегодня",
  mcp_daily: "MCP-вызовов на сегодня",
  chains: "цепочек",
  chain_steps: "шагов в цепочке",
  // Pack T: team-pool квоты (миграция 000070). Лимит на ресурсы внутри одной
  // команды по плану owner'а — отличается от personal-лимита текущего юзера.
  team_prompts: "промптов в команде",
  team_collections: "коллекций в команде",
  team_chains: "цепочек в команде",
}

// Форма существительного для headline ("500 X на Pro"). Отличается от quotaLabel
// тем что использует «в день», а не «на сегодня» — для headline это точнее.
const quotaHeadlineNoun: Record<string, string> = {
  prompts: "промптов",
  collections: "коллекций",
  teams: "команд",
  team_members: "участников в команде",
  ext_daily: "вставок в день через расширение",
  mcp_daily: "MCP-вызовов в день",
  chains: "цепочек",
  chain_steps: "шагов в цепочке",
}

// Per-resource маркетинговые данные. Числа лимитов больше НЕ хранятся здесь —
// тянутся через usePlans() из /api/plans (live из БД). Это убирает рассинхрон
// между миграциями (000046, 000067, …) и фронтом — раньше при апдейте лимита
// в SQL нужно было ручками править и хардкод тут.
//
// maxAction — что Max-юзер может сделать сам, чтобы вписаться в лимит.
// Для дневных лимитов (ext_daily, mcp_daily) maxAction = null — юзер ничего
// не может, счётчик сбросится в полночь UTC. В этом случае primary-кнопка
// не показывается, остаётся только «Связаться с поддержкой».
type ResourceMeta = {
  detail: string
  maxAction: { label: string; href: string; icon: LucideIcon } | null
}

const resourceMeta: Record<string, ResourceMeta> = {
  prompts: {
    detail: "Храните всю библиотеку без ограничений.",
    maxAction: { label: "Удалить лишние промпты", href: "/dashboard", icon: Trash2 },
  },
  collections: {
    detail: "Группируйте промпты по командам, клиентам, проектам.",
    maxAction: { label: "Удалить лишние коллекции", href: "/collections", icon: Trash2 },
  },
  teams: {
    detail: "Разделяйте промпты между проектами и клиентами.",
    maxAction: { label: "Удалить лишние команды", href: "/teams", icon: Trash2 },
  },
  team_members: {
    detail: "Max подходит для агентств и студий с большими командами.",
    maxAction: { label: "Открыть команды", href: "/teams", icon: ArrowRight },
  },
  // Phase 16-Y: share_links / daily_shares УДАЛЕНЫ — share больше не имеет
  // квот, ссылки живут по TTL (30 дней default).
  ext_daily: {
    detail: "Вставляйте промпты прямо в ChatGPT, Claude, Gemini, Perplexity.",
    maxAction: null,
  },
  mcp_daily: {
    detail: "Используйте промпты в Claude Desktop, Cursor, Windsurf, Cline.",
    maxAction: null,
  },
  chains: {
    detail: "Стройте многошаговые пайплайны и переиспользуйте их в любом AI-клиенте.",
    maxAction: { label: "Удалить лишние цепочки", href: "/chains", icon: Trash2 },
  },
  chain_steps: {
    detail: "Сложные сценарии — исследование → анализ → черновик → проверка без обрыва.",
    maxAction: { label: "Открыть цепочки", href: "/chains", icon: ArrowRight },
  },
}

// Маппинг quota_type из ответа backend → имя поля в Plan. Должен покрывать
// все ключи из resourceMeta. Если нового quota_type здесь нет, headline
// просто не отрендерится (return null) — degradate gracefully.
const QUOTA_TO_PLAN_FIELD: Record<string, keyof Plan> = {
  prompts: "max_prompts",
  collections: "max_collections",
  teams: "max_teams",
  team_members: "max_team_members",
  ext_daily: "max_ext_uses_daily",
  mcp_daily: "max_mcp_uses_daily",
  chains: "max_chains",
  chain_steps: "max_steps_per_chain",
}

// formatPlanLimit — резолв лимита из живого Plan и форматирование с разрядами.
// undefined plan / unknown quotaType → null (хедлайн не показываем).
function formatPlanLimit(plan: Plan | undefined, quotaType: string): string | null {
  if (!plan) return null
  const field = QUOTA_TO_PLAN_FIELD[quotaType]
  if (!field) return null
  const value = plan[field]
  if (typeof value !== "number") return null
  return value.toLocaleString("ru-RU")
}

const planCTA: Record<"pro" | "max", { label: string; pricePerDay: string }> = {
  pro: { label: "Pro", pricePerDay: "20" },
  max: { label: "Max", pricePerDay: "43" },
}

const SUPPORT_EMAIL = "slava0gpt@gmail.com"

// pickTargetPlan — какой следующий тариф предлагать как апсейл.
// null = юзер уже на Max, апселла нет — показываем «свяжитесь с поддержкой» CTA.
function pickTargetPlan(currentPlan: string): "pro" | "max" | null {
  if (currentPlan === "free") return "pro"
  if (currentPlan === "pro" || currentPlan === "pro_yearly") return "max"
  return null
}

// buildHeadline — динамический headline в карточке value-prop.
// Free юзер: "500 X на Pro, 10 000 на Max" (полная вилка).
// Pro юзер:  "10 000 X на Max" (без упоминания Pro — он уже там).
// Max юзер:  null (карточка не показывается).
//
// Числа берём из live-планов (usePlans) — миграции апдейтят БД, фронт
// тут же показывает актуальное значение. Если plans ещё грузятся (proLimit
// или maxLimit === null) — headline не показываем, чтобы не моргать «—».
function buildHeadline(
  quotaType: string,
  target: "pro" | "max" | null,
  proPlan: Plan | undefined,
  maxPlan: Plan | undefined,
): string | null {
  if (!resourceMeta[quotaType]) return null
  const noun = quotaHeadlineNoun[quotaType] ?? quotaType
  const proLimit = formatPlanLimit(proPlan, quotaType)
  const maxLimit = formatPlanLimit(maxPlan, quotaType)
  if (target === "max") {
    if (!maxLimit) return null
    return `${maxLimit} ${noun} на Max`
  }
  if (target === "pro") {
    if (!proLimit || !maxLimit) return null
    return `${proLimit} ${noun} на Pro, ${maxLimit} на Max`
  }
  return null
}

export function QuotaExceededDialog() {
  const { open, quotaType, message, used, limit, plan, dismiss } = useQuotaStore()
  const navigate = useNavigate()
  const planId = useAuthStore((s) => s.user?.plan_id ?? "free")
  // Live планы из /api/plans (TanStack Query кэш 60min). Заменяет хардкод
  // proLimit/maxLimit из старой версии, синхронизированный с миграциями
  // вручную. Теперь — single source of truth: миграция апдейтит цифры,
  // фронт сам подтягивает.
  const { data: plans } = usePlans()
  const proPlan = plans?.find((p) => p.id === "pro")
  const maxPlan = plans?.find((p) => p.id === "max")

  const resourceLabel = quotaType ? (quotaLabels[quotaType] ?? quotaType) : ""
  const target = pickTargetPlan(planId)
  const meta = quotaType ? resourceMeta[quotaType] : undefined
  const headline = quotaType ? buildHeadline(quotaType, target, proPlan, maxPlan) : null
  const hasUsage = typeof used === "number" && typeof limit === "number" && limit > 0
  const isMaxUser = target === null

  // Phase 16-Z2: вместо inline-consent + direct-checkout ведём юзера на
  // /pricing?upgrade=<target>. Pricing-страница автоматически откроет
  // CheckoutConfirmDialog с нужным планом — там сводка, дата следующего
  // списания и чек-бокс согласия. Это даёт юзеру шанс ещё раз сравнить
  // тарифы перед окончательным согласием.
  const upgradeAndNavigate = (targetPlanId: string) => {
    dismiss()
    navigate(`/pricing?upgrade=${encodeURIComponent(targetPlanId)}`)
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && dismiss()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-amber-100 dark:bg-amber-900/30">
            <AlertTriangle className="h-6 w-6 text-amber-600 dark:text-amber-400" />
          </div>
          <DialogTitle className="text-center">
            {hasUsage ? `Лимит ${resourceLabel}: ${used}/${limit}` : "Лимит исчерпан"}
          </DialogTitle>
          <DialogDescription className="text-center">
            {isMaxUser
              ? meta?.maxAction
                ? `Вы на тарифе Max — это потолок. Удалите лишнее или свяжитесь с поддержкой для расширения лимитов.`
                : `Это дневной лимит — счётчик сбросится в полночь UTC. Если нужно больше прямо сейчас, свяжитесь с поддержкой.`
              : (message || `Вы достигли лимита ${resourceLabel} на текущем плане.`)}
          </DialogDescription>
        </DialogHeader>

        <div className="flex justify-center">
          <PlanBadge planId={(plan as "free" | "pro" | "max" | undefined) ?? (planId as "free" | "pro" | "max")} />
        </div>

        {meta && headline && (
          <div
            className="rounded-lg border p-4"
            style={{
              borderColor: target === "max" ? "#f59e0b50" : "#8b5cf650",
              background: target === "max" ? "#f59e0b08" : "#8b5cf608",
            }}
          >
            <p className="mb-1 text-[0.85rem] font-semibold text-foreground">{headline}</p>
            <p className="text-[0.78rem] text-muted-foreground">{meta.detail}</p>
          </div>
        )}

        <DialogFooter className="flex-col gap-2 sm:flex-col">
          {target ? (
            <Button
              className="w-full gap-1.5"
              onClick={() => upgradeAndNavigate(target)}
            >
              {planId === "free" ? "Получить" : "Перейти на"} {planCTA[target].label} за {planCTA[target].pricePerDay}₽/день
              <ArrowRight className="h-3.5 w-3.5" />
            </Button>
          ) : meta?.maxAction ? (
            // Max юзер на count-based лимите — даём конкретное действие
            // («Удалить лишние промпты», «Удалить лишние коллекции» и т.д.).
            <Button
              className="w-full gap-2"
              onClick={() => { dismiss(); navigate(meta.maxAction!.href) }}
            >
              <meta.maxAction.icon className="h-4 w-4" />
              {meta.maxAction.label}
            </Button>
          ) : null
          /* Max юзер на daily-лимите (daily_shares/ext_daily/mcp_daily) — primary
             кнопки нет, юзер не может ничего сделать сам, ждёт midnight UTC.
             Остаётся только «Связаться с поддержкой» ниже. */
          }

          {isMaxUser ? (
            // Без asChild + <a>: вложенный <a> ломал inline-flex кнопки и иконка
            // съезжала над текстом. Делаем чистый Button + ручной open mailto +
            // копируем email в буфер на случай отсутствия mail-клиента (типичный
            // кейс на десктопе без default handler — иначе клик ничего не делал).
            <Button
              variant="outline"
              className="w-full gap-2"
              onClick={async () => {
                const subject = encodeURIComponent("Запрос на расширение лимитов Max")
                try {
                  await navigator.clipboard.writeText(SUPPORT_EMAIL)
                  toast.success("Email скопирован в буфер обмена", { description: SUPPORT_EMAIL })
                } catch {
                  toast.info("Напишите нам", { description: SUPPORT_EMAIL, duration: 8000 })
                }
                window.location.href = `mailto:${SUPPORT_EMAIL}?subject=${subject}`
                dismiss()
              }}
            >
              <LifeBuoy className="h-4 w-4" />
              Связаться с поддержкой
            </Button>
          ) : (
            <Button variant="ghost" className="w-full" onClick={() => { dismiss(); navigate("/pricing") }}>
              Сравнить тарифы
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
