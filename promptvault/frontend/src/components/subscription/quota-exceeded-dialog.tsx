import { useNavigate } from "react-router"
import { AlertTriangle } from "lucide-react"
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
import { PlanBadge } from "./plan-badge"
import { useAuthStore } from "@/stores/auth-store"

const quotaLabels: Record<string, string> = {
  prompts: "промптов",
  collections: "коллекций",
  ai_daily: "AI-запросов на сегодня",
  ai_total: "AI-запросов",
  teams: "команд",
  team_members: "участников команды",
  share_links: "публичных ссылок",
  ext_daily: "вставок через расширение на сегодня",
  mcp_daily: "MCP-вызовов на сегодня",
}

export function QuotaExceededDialog() {
  const { open, quotaType, message, dismiss } = useQuotaStore()
  const navigate = useNavigate()
  const planId = useAuthStore((s) => s.user?.plan_id ?? "free")

  const resource = quotaType ? (quotaLabels[quotaType] ?? quotaType) : ""

  return (
    <Dialog open={open} onOpenChange={(v) => !v && dismiss()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-amber-100 dark:bg-amber-900/30">
            <AlertTriangle className="h-6 w-6 text-amber-600 dark:text-amber-400" />
          </div>
          <DialogTitle className="text-center">Лимит исчерпан</DialogTitle>
          <DialogDescription className="text-center">
            {message || `Вы достигли лимита ${resource} на текущем плане.`}
          </DialogDescription>
        </DialogHeader>

        <div className="flex justify-center">
          <PlanBadge planId={planId as "free" | "pro" | "max"} />
        </div>

        <DialogFooter className="flex-col gap-2 sm:flex-col">
          <Button
            className="w-full"
            onClick={() => {
              dismiss()
              navigate("/pricing")
            }}
          >
            Обновить план
          </Button>
          <Button variant="ghost" className="w-full" onClick={dismiss}>
            Позже
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
