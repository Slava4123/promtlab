import { ExternalLink, X, Zap } from "lucide-react"
import { Button } from "../ui/button"
import { useQuotaStore } from "../../stores/quota-store"

// Глобальный модал — показывается когда bg-client получает 402.
// Подключается в AppShell.
export function QuotaExceededDialog() {
  const { open, message, quotaType, used, limit, plan, dismiss } = useQuotaStore()

  if (!open) return null

  async function openUpgrade() {
    const { getSettings } = await import("../../lib/storage")
    const { openWebPage } = await import("../../lib/utils")
    const { apiBase } = await getSettings()
    openWebPage(apiBase, "/pricing?source=quota_exceeded&from=extension")
    dismiss()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={dismiss} />
      <div className="relative w-full max-w-sm rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
        <div className="flex items-start gap-3">
          <div className="mt-0.5 flex h-8 w-8 items-center justify-center rounded-full bg-amber-500/15">
            <Zap className="h-4 w-4 text-amber-500" />
          </div>
          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-semibold">Лимит исчерпан</h3>
            {quotaType && (
              <p className="mt-0.5 text-[10px] text-(--color-muted-foreground)">
                {quotaType}{plan && ` • ${plan}`}
              </p>
            )}
          </div>
          <button
            type="button"
            onClick={dismiss}
            className="rounded-md p-1 text-(--color-muted-foreground) hover:bg-(--color-muted)"
            aria-label="Закрыть"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {message && (
          <p className="mt-3 text-xs text-(--color-muted-foreground)">{message}</p>
        )}

        {used !== null && limit !== null && (
          <div className="mt-3 rounded-md border border-(--color-border) bg-(--color-muted)/30 p-2">
            <div className="flex items-center justify-between text-[10px]">
              <span>Использовано</span>
              <span className="font-mono">
                {used} / {limit < 0 ? "∞" : limit}
              </span>
            </div>
          </div>
        )}

        <div className="mt-4 flex justify-end gap-2">
          <Button type="button" variant="outline" size="sm" onClick={dismiss}>
            Понятно
          </Button>
          <Button type="button" size="sm" onClick={openUpgrade} className="gap-1.5">
            <ExternalLink className="h-3.5 w-3.5" />
            Обновить тариф
          </Button>
        </div>
      </div>
    </div>
  )
}
