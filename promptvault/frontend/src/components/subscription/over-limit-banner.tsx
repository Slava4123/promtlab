// Persistent banner для юзеров с превышением лимита текущего плана.
// Показывается на всех защищённых страницах (через AppLayout), пока юзер
// не нажмёт «×» — тогда баннер скрывается до следующего календарного дня.
//
// Логика тихая: banner не появляется, пока какая-то метрика не превысила
// лимит. Никаких approaching-warning'ов на 80% — это излишне для платных
// юзеров и ведёт к alert-fatigue. Только реальный over-limit.

import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { AlertTriangle, ArrowRight, Trash2, X } from "lucide-react"

import { Button } from "@/components/ui/button"
import { useUsage } from "@/hooks/use-subscription"
import type { QuotaInfo, UsageSummary } from "@/api/types"

// Какие категории usage показывать. Ключ — поле в UsageSummary, label —
// в винительном падеже (для текста «X промптов»).
const FIELDS: Array<{ key: keyof UsageSummary; label: (n: number) => string }> = [
  { key: "prompts", label: (n) => `${n} ${plur(n, "промпт", "промпта", "промптов")}` },
  { key: "collections", label: (n) => `${n} ${plur(n, "коллекция", "коллекции", "коллекций")}` },
  { key: "teams", label: (n) => `${n} ${plur(n, "команда", "команды", "команд")}` },
  // Phase 16-Y: share_links убран — на share-ссылки больше нет квот (TTL вместо).
  { key: "chains", label: (n) => `${n} ${plur(n, "цепочка", "цепочки", "цепочек")}` },
]

function plur(n: number, one: string, few: string, many: string): string {
  const m10 = n % 10
  const m100 = n % 100
  if (m10 === 1 && m100 !== 11) return one
  if (m10 >= 2 && m10 <= 4 && (m100 < 10 || m100 >= 20)) return few
  return many
}

function todayKey(): string {
  // YYYY-MM-DD по локальному времени — banner снова появится завтра.
  return new Date().toISOString().slice(0, 10)
}

const STORAGE_PREFIX = "pv_overlimit_dismissed_"

function isDismissed(): boolean {
  try {
    return localStorage.getItem(STORAGE_PREFIX + todayKey()) === "1"
  } catch {
    return false
  }
}

function markDismissed() {
  try {
    localStorage.setItem(STORAGE_PREFIX + todayKey(), "1")
  } catch {
    /* private mode / quota — баннер вернётся при следующем открытии страницы */
  }
}

function pickTargetPlan(currentPlan: string): "pro" | "max" | null {
  if (currentPlan === "free") return "pro"
  if (currentPlan === "pro" || currentPlan === "pro_yearly") return "max"
  return null
}

export function OverLimitBanner() {
  const navigate = useNavigate()
  const { data: usage } = useUsage()
  const [dismissed, setDismissed] = useState(isDismissed)

  if (dismissed || !usage) return null

  // Берём plan_id из usage, а не из auth-store: auth-store на холодный
  // mount ещё может не успеть загрузить user, что давало race — Pro юзер
  // видел кнопку «Получить Pro» вместо «Перейти на Max».
  const planId = usage.plan_id ?? "free"

  const overages = FIELDS
    .map(({ key, label }) => {
      const info = usage[key] as QuotaInfo | undefined
      if (!info || info.limit <= 0 || info.used <= info.limit) return null
      return { label: label(info.used - info.limit), used: info.used, limit: info.limit }
    })
    .filter((x): x is NonNullable<typeof x> => x !== null)

  if (overages.length === 0) return null

  const target = pickTargetPlan(String(planId))
  const typeWord = plur(overages.length, "тип", "типа", "типов")
  const summary = overages.length === 1
    ? `${overages[0].label} сверх лимита текущего плана`
    : `${overages.length} ${typeWord} ресурсов сверх лимита текущего плана`

  const handleDismiss = () => {
    markDismissed()
    setDismissed(true)
  }

  return (
    <div
      role="alert"
      aria-live="polite"
      className="flex flex-wrap items-center gap-3 border-b border-amber-500/30 bg-amber-500/10 px-4 py-2.5 text-sm text-amber-900 dark:text-amber-200"
    >
      <AlertTriangle className="h-4 w-4 shrink-0" aria-hidden="true" />
      <span className="flex-1 min-w-0">{summary}.{" "}
        <span className="text-muted-foreground">
          Создавать новые получится после удаления лишних.
        </span>
      </span>
      <div className="flex items-center gap-2">
        {target ? (
          <Button size="sm" onClick={() => navigate("/pricing")}>
            {planId === "free" ? "Получить" : "Перейти на"} {target === "max" ? "Max" : "Pro"}
            <ArrowRight className="ml-1 h-3.5 w-3.5" />
          </Button>
        ) : (
          <Button size="sm" variant="outline" onClick={() => navigate("/trash")}>
            <Trash2 className="mr-1 h-3.5 w-3.5" />
            Освободить место
          </Button>
        )}
        <button
          type="button"
          onClick={handleDismiss}
          aria-label="Скрыть до завтра"
          className="rounded-md p-1 text-amber-900/70 transition-colors hover:bg-amber-500/15 hover:text-amber-900 dark:text-amber-200/70 dark:hover:text-amber-200"
        >
          <X className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}
