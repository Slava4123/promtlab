import { ExternalLink, X, Zap } from "lucide-react"
import { Button } from "../ui/button"
import { useQuotaStore } from "../../stores/quota-store"

// Маппинг технических quota_type'ов → читаемые названия. quotaType
// сейчас передаётся как 'unknown' (см. bg-client.ts), поэтому также пытаемся
// угадать тип по тексту backend-сообщения.
const QUOTA_LABELS: Record<string, string> = {
  prompts: "Промпты",
  collections: "Коллекции",
  chains: "Цепочки",
  teams: "Команды",
  ext_uses_today: "Вставки сегодня",
  mcp_uses_today: "MCP-вызовы сегодня",
  api_keys: "API-ключи",
  share_links: "Публичные ссылки",
  team_members: "Участники команды",
}

const PLAN_LABELS: Record<string, string> = {
  free: "Free",
  pro: "Pro",
  pro_yearly: "Pro (год)",
  max: "Max",
  max_yearly: "Max (год)",
}

// Fallback-метка когда ни quotaType, ни эвристика по тексту не сработали —
// показываем хоть что-то осмысленное вместо пустоты.
const QUOTA_FALLBACK_LABEL = "Лимит ресурса"

export function readableQuotaType(quotaType: string | null, message: string | null): string {
  if (quotaType && quotaType !== "unknown" && QUOTA_LABELS[quotaType]) {
    return QUOTA_LABELS[quotaType]
  }
  // Угадываем по сообщению: «Лимит цепочек исчерпан» → chains.
  if (message) {
    const m = message.toLowerCase()
    if (m.includes("цепоч")) return QUOTA_LABELS.chains
    if (m.includes("промпт")) return QUOTA_LABELS.prompts
    if (m.includes("коллекц")) return QUOTA_LABELS.collections
    if (m.includes("команд")) return QUOTA_LABELS.teams
    // "встав" покрывает «вставки/вставку/вставок/вставкой» — все формы.
    if (m.includes("встав") || m.includes("использовани")) return QUOTA_LABELS.ext_uses_today
    if (m.includes("mcp")) return QUOTA_LABELS.mcp_uses_today
    if (m.includes("api-ключ") || m.includes("api ключ")) return QUOTA_LABELS.api_keys
  }
  // Если backend менял copy — заметим в логах. Не флудим, если пришёл
  // совсем пустой контекст (quotaType=null && message=null).
  if (quotaType || message) {
    console.warn("[QuotaDialog] не распознан тип квоты", { quotaType, message })
  }
  return QUOTA_FALLBACK_LABEL
}

// Глобальный модал — показывается когда bg-client получает 402.
// Подключается в AppShell.
export function QuotaExceededDialog() {
  const { open, message, quotaType, used, limit, plan, dismiss } = useQuotaStore()

  if (!open) return null

  const readable = readableQuotaType(quotaType, message)
  const planLabel = plan ? PLAN_LABELS[plan] ?? plan : null

  async function openUpgrade() {
    const { getSettings } = await import("../../lib/storage")
    const { openWebPage } = await import("../../lib/utils")
    const { apiBase } = await getSettings()
    openWebPage(apiBase, "/pricing?source=quota_exceeded&from=extension")
    dismiss()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Modal backdrop. Полная клавиатурная навигация — на уровне dialog
          (Escape по document keydown), backdrop сам по себе не должен быть
          tab-focusable. ESLint правило слишком строгое для этого паттерна. */}
      {/* eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={dismiss} />
      <div className="relative w-full max-w-sm rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
        <div className="flex items-start gap-3">
          <div className="mt-0.5 flex h-8 w-8 items-center justify-center rounded-full bg-amber-500/15">
            <Zap className="h-4 w-4 text-amber-500" />
          </div>
          <div className="flex-1 min-w-0">
            <h3 className="text-sm font-semibold">Лимит исчерпан</h3>
            {(readable || planLabel) && (
              <p className="mt-0.5 text-[10px] text-(--color-muted-foreground)">
                {readable}
                {readable && planLabel && " • "}
                {planLabel && `тариф ${planLabel}`}
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
