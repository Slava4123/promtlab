import { useNavigate } from "react-router-dom"
import { AlertTriangle, ArrowRight, Loader2 } from "lucide-react"
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
import { useCheckout } from "@/hooks/use-subscription"
import { PlanBadge } from "./plan-badge"

// Человеко-читаемое название ресурса для заголовка.
const quotaLabels: Record<string, string> = {
  prompts: "промптов",
  collections: "коллекций",
  teams: "команд",
  team_members: "участников команды",
  share_links: "публичных ссылок",
  ext_daily: "вставок через расширение на сегодня",
  mcp_daily: "MCP-вызовов на сегодня",
}

// Per-quota value-prop: что юзер получит на Pro. Рекомендуем Pro (19₽/день).
type QuotaBenefit = { headline: string; detail: string; targetPlan: "pro" | "max" }

const quotaBenefits: Record<string, QuotaBenefit> = {
  prompts: {
    headline: "500 промптов на Pro, безлимит на Max",
    detail: "Храните всю библиотеку без ограничений — 19₽ в день.",
    targetPlan: "pro",
  },
  collections: {
    headline: "Безлимитные коллекции на Pro",
    detail: "Группируйте промпты как удобно — команды, клиенты, проекты.",
    targetPlan: "pro",
  },
  teams: {
    headline: "5 команд на Pro, безлимит на Max",
    detail: "Разделяйте промпты между проектами и клиентами.",
    targetPlan: "pro",
  },
  team_members: {
    headline: "До 10 участников в команде на Pro",
    detail: "Безлимит участников на Max — для агентств и студий.",
    targetPlan: "pro",
  },
  share_links: {
    headline: "10 публичных ссылок на Pro",
    detail: "Делитесь готовыми промптами — безлимит на Max.",
    targetPlan: "pro",
  },
  ext_daily: {
    headline: "30 вставок в день на Pro",
    detail: "Вставляйте промпты в ChatGPT/Claude/Gemini без ограничений.",
    targetPlan: "pro",
  },
  mcp_daily: {
    headline: "30 MCP-вызовов в день на Pro",
    detail: "Подключите Claude Desktop / Cursor / Windsurf через MCP.",
    targetPlan: "pro",
  },
}

export function QuotaExceededDialog() {
  const { open, quotaType, message, used, limit, plan, dismiss } = useQuotaStore()
  const navigate = useNavigate()
  const checkout = useCheckout()
  const planId = useAuthStore((s) => s.user?.plan_id ?? "free")

  const resource = quotaType ? (quotaLabels[quotaType] ?? quotaType) : ""
  const benefit = quotaType ? quotaBenefits[quotaType] : undefined
  const targetPlan = benefit?.targetPlan ?? "pro"
  const hasUsage = typeof used === "number" && typeof limit === "number" && limit > 0

  return (
    <Dialog open={open} onOpenChange={(v) => !v && dismiss()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-amber-100 dark:bg-amber-900/30">
            <AlertTriangle className="h-6 w-6 text-amber-600 dark:text-amber-400" />
          </div>
          <DialogTitle className="text-center">
            {hasUsage ? `Лимит ${resource}: ${used}/${limit}` : "Лимит исчерпан"}
          </DialogTitle>
          <DialogDescription className="text-center">
            {message || `Вы достигли лимита ${resource} на текущем плане.`}
          </DialogDescription>
        </DialogHeader>

        <div className="flex justify-center">
          <PlanBadge planId={(plan as "free" | "pro" | "max" | undefined) ?? (planId as "free" | "pro" | "max")} />
        </div>

        {benefit && (
          <div
            className="rounded-lg border p-4"
            style={{
              borderColor: targetPlan === "max" ? "#f59e0b50" : "#8b5cf650",
              background: targetPlan === "max" ? "#f59e0b08" : "#8b5cf608",
            }}
          >
            <p className="mb-1 text-[0.85rem] font-semibold text-foreground">{benefit.headline}</p>
            <p className="text-[0.78rem] text-muted-foreground">{benefit.detail}</p>
          </div>
        )}

        <DialogFooter className="flex-col gap-2 sm:flex-col">
          <Button
            className="w-full"
            disabled={checkout.isPending}
            onClick={() => {
              dismiss()
              // Пробуем прямой checkout для залогиненного юзера — меньше кликов до оплаты.
              // Если нет auth (нет plan_id) — обычный редирект на pricing.
              if (planId !== "free" || !planId) {
                navigate("/pricing")
                return
              }
              checkout.mutate(targetPlan)
            }}
          >
            {checkout.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <>
                Получить {targetPlan === "max" ? "Max" : "Pro"} за {targetPlan === "max" ? "42" : "19"}₽/день
                <ArrowRight className="ml-1.5 h-3.5 w-3.5" />
              </>
            )}
          </Button>
          <Button variant="ghost" className="w-full" onClick={() => { dismiss(); navigate("/pricing") }}>
            Сравнить тарифы
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
