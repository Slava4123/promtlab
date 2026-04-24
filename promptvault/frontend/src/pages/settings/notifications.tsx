import { useState } from "react"
import { Bell, Loader2 } from "lucide-react"
import { toast } from "sonner"

import { useAuthStore } from "@/stores/auth-store"
import { useSetInsightEmails } from "@/hooks/use-settings"
import { SectionHeader } from "./_section-header"

// Phase 14 M-10: opt-in по ФЗ-152. Default false; пользователь включает
// явно здесь. Backend в insight_notifications.up.sql добавил колонку
// users.insight_emails_enabled.
export default function SettingsNotificationsPage() {
  const user = useAuthStore((s) => s.user)
  const fetchMe = useAuthStore((s) => s.fetchMe)
  const mutation = useSetInsightEmails()

  // Optimistic override закрывает gap между mutation success и fetchMe refresh.
  // null → используем server value (user.insight_emails_enabled).
  const [optimistic, setOptimistic] = useState<boolean | null>(null)
  const enabled = optimistic ?? !!user?.insight_emails_enabled

  if (!user) return null

  const plan = user.plan_id ?? "free"
  const isMax = plan.startsWith("max")

  async function handleToggle(next: boolean) {
    setOptimistic(next)
    try {
      await mutation.mutateAsync(next)
      await fetchMe()
      toast.success(
        next
          ? "Email-уведомления по Smart Insights включены"
          : "Email-уведомления отключены",
      )
    } catch {
      // toast уже показан хуком.
    } finally {
      setOptimistic(null)
    }
  }

  return (
    <div className="space-y-6">
      <SectionHeader
        icon={Bell}
        title="Уведомления"
        description="Как и когда мы можем писать вам на email."
      />

      <div className="rounded-lg border p-4">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <p className="font-medium">Smart Insights digest</p>
            <p className="text-sm text-muted-foreground">
              Раз в неделю будем присылать краткую сводку изменений в инсайтах:
              неиспользуемые промпты, растущие, возможные дубликаты. Доступно
              на тарифе Max.
            </p>
            {!isMax && (
              <p className="text-xs text-amber-600 dark:text-amber-400">
                Инсайты рассчитываются только для тарифа Max. Переключатель
                можно включить заранее — письма начнут приходить после апгрейда.
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              По ФЗ-152 это явное согласие на маркетинговые письма. Выключить
              можно в любой момент.
            </p>
          </div>
          <label className="flex shrink-0 cursor-pointer items-center gap-2">
            <span className="text-sm">
              {enabled ? "Вкл." : "Выкл."}
            </span>
            <input
              type="checkbox"
              className="size-4"
              checked={enabled}
              disabled={mutation.isPending}
              onChange={(e) => handleToggle(e.target.checked)}
              aria-label="Email-уведомления по Smart Insights"
            />
            {mutation.isPending && <Loader2 className="size-4 animate-spin" />}
          </label>
        </div>
      </div>
    </div>
  )
}
