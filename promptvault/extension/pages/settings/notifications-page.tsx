import { useMemo } from "react"
import { useNavigate } from "react-router-dom"
import {
  AlertTriangle,
  ArrowLeft,
  Bell,
  Check,
  ExternalLink,
  Loader2,
  Mail,
  Shield,
  ShieldCheck,
  Eye,
  X,
} from "lucide-react"
import { Button } from "../../components/ui/button"
import { useToast } from "../../components/ui/toaster"
import {
  useMyInvitations,
  useAcceptInvitation,
  useDeclineInvitation,
} from "../../hooks/use-invitations"
import { useUsageSummary } from "../../hooks/use-usage-summary"
import { useSettings } from "../../hooks/use-settings"
import { useNotificationsReadStore } from "../../stores/notifications-read-store"
import { openWebPage } from "../../lib/utils"
import { formatRelativeDate } from "@pv/shared/utils/format-date"
import {
  QUOTA_KEYS,
  quotaByKey,
  type QuotaInfo,
  type QuotaKey,
  type TeamInvitation,
  type TeamRole,
} from "../../lib/types"

// Полный центр уведомлений: приглашения в команды + over-limit warnings.
// OverLimitBanner был убран из AppShell — все «срочные» сигналы стекаются
// сюда. Email-настройки (Smart Insights и пр.) — теперь только в веб-приложении.

const QUOTA_LABELS: Record<string, string> = {
  prompts: "Промпты",
  collections: "Коллекции",
  chains: "Цепочки",
  teams: "Команды",
  ext_uses_today: "Вставки сегодня",
  mcp_uses_today: "MCP-вызовы сегодня",
}

const ROLE_META: Record<TeamRole, { label: string; icon: React.ComponentType<{ className?: string }>; color: string }> = {
  owner: { label: "Владелец", icon: ShieldCheck, color: "text-amber-500" },
  editor: { label: "Редактор", icon: Shield, color: "text-(--color-primary)" },
  viewer: { label: "Просмотр", icon: Eye, color: "text-(--color-muted-foreground)" },
}

export function NotificationsPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const settings = useSettings()
  const invitations = useMyInvitations()
  const usage = useUsageSummary()
  const acceptMut = useAcceptInvitation()
  const declineMut = useDeclineInvitation()
  const readIds = useNotificationsReadStore((s) => s.ids)
  const markRead = useNotificationsReadStore((s) => s.markRead)
  const markAllReadStore = useNotificationsReadStore((s) => s.markAllRead)

  const readSet = new Set(readIds)

  // Фильтруем СРАЗУ — скрытые карточки полностью исчезают, не приглушаются.
  const pending: TeamInvitation[] = (invitations.data ?? [])
    .filter((i) => i.status === "pending")
    .filter((i) => !readSet.has(`invitation-${i.id}`))

  const overLimit = useMemo(() => {
    if (!usage.data) return []
    const items: Array<{ key: QuotaKey; info: QuotaInfo }> = []
    for (const key of QUOTA_KEYS) {
      const info = quotaByKey(usage.data, key)
      if (info.limit <= 0) continue
      if (info.used >= info.limit && !readSet.has(`quota-${key}`)) {
        items.push({ key, info })
      }
    }
    return items
    // readIds в deps — пересобрать после mark-as-read.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [usage.data, readIds])

  function markAllRead() {
    const ids: string[] = []
    for (const inv of pending) ids.push(`invitation-${inv.id}`)
    for (const o of overLimit) ids.push(`quota-${o.key}`)
    markAllReadStore(ids)
  }

  async function handleAccept(inv: TeamInvitation) {
    try {
      await acceptMut.mutateAsync(inv.id)
      toast({ title: `Вы в команде «${inv.team_name}»`, variant: "success" })
      markRead(`invitation-${inv.id}`)
      void invitations.refetch()
    } catch (err) {
      toast({
        title: "Не удалось принять",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  async function handleDecline(inv: TeamInvitation) {
    try {
      await declineMut.mutateAsync(inv.id)
      toast({ title: "Приглашение отклонено", variant: "info" })
      markRead(`invitation-${inv.id}`)
    } catch {
      toast({ title: "Не удалось отклонить", variant: "error" })
    }
  }

  function openUpgrade() {
    if (!settings) return
    openWebPage(settings.apiBase, "/pricing?source=notifications&from=extension")
  }

  function openEmailSettings() {
    if (!settings) return
    openWebPage(settings.apiBase, "/settings/notifications?from=extension")
  }

  const isPending = invitations.isPending || usage.isPending
  // Поскольку pending/overLimit уже отфильтрованы выше, totalUnread = их сумма.
  const totalUnread = pending.length + overLimit.length

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Уведомления</h2>
        {totalUnread > 0 && (
          <button
            type="button"
            onClick={markAllRead}
            className="text-[10px] text-(--color-muted-foreground) hover:text-(--color-foreground)"
          >
            Прочитать все
          </button>
        )}
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {isPending ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : pending.length === 0 && overLimit.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-12 text-center">
            <Bell className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Нет уведомлений</p>
            <p className="max-w-xs text-[10px] text-(--color-muted-foreground)">
              Приглашения в команды и превышения лимитов будут появляться здесь.
            </p>
          </div>
        ) : (
          <>
            {/* Over-limit warnings */}
            {overLimit.map((o) => {
              const id = `quota-${o.key}`
              return (
                <div
                  key={id}
                  className="relative rounded-md border border-l-2 border-l-amber-500/70 border-(--color-border) bg-amber-500/[0.04] p-3"
                >
                  <div className="flex items-start gap-2">
                    <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-amber-500" />
                    <div className="flex-1 min-w-0">
                      <div className="text-xs font-medium">
                        Лимит «{QUOTA_LABELS[o.key]}» исчерпан
                      </div>
                      <div className="mt-0.5 text-[10px] text-(--color-muted-foreground)">
                        Использовано {o.info.used} из {o.info.limit}. Обновите тариф,
                        чтобы продолжить.
                      </div>
                      <div className="mt-2 flex items-center gap-2">
                        <Button
                          type="button"
                          size="sm"
                          variant="outline"
                          onClick={openUpgrade}
                          className="h-7 gap-1 text-[10px]"
                        >
                          <ExternalLink className="h-3 w-3" />
                          Обновить тариф
                        </Button>
                        <button
                          type="button"
                          onClick={() => markRead(id)}
                          className="text-[10px] text-(--color-muted-foreground) hover:underline"
                        >
                          Скрыть
                        </button>
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() => markRead(id)}
                      className="rounded p-0.5 text-(--color-muted-foreground) hover:bg-(--color-muted)"
                      aria-label="Скрыть"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </div>
                </div>
              )
            })}

            {/* Pending invitations */}
            {pending.map((inv) => {
              const id = `invitation-${inv.id}`
              const meta = ROLE_META[inv.role]
              const Icon = meta.icon
              const busy = acceptMut.isPending || declineMut.isPending
              return (
                <div
                  key={id}
                  className="rounded-md border border-(--color-border) bg-(--color-card) p-3"
                >
                  <div className="text-xs font-medium">
                    Приглашение в команду «{inv.team_name}»
                  </div>
                  <div className="mt-0.5 flex items-center gap-1.5 text-[10px] text-(--color-muted-foreground)">
                    <span>от {inv.inviter_name}</span>
                    <span>•</span>
                    <span>{formatRelativeDate(inv.created_at)}</span>
                  </div>
                  <div className="mt-2 flex items-center gap-2">
                    <span className={`flex items-center gap-1 text-[10px] ${meta.color}`}>
                      <Icon className="h-3 w-3" />
                      {meta.label}
                    </span>
                    <div className="ml-auto flex gap-1.5">
                      <button
                        type="button"
                        onClick={() => handleAccept(inv)}
                        disabled={busy}
                        className="flex h-6 items-center gap-1 rounded-md bg-(--color-primary) px-2 text-[10px] font-medium text-(--color-primary-foreground) hover:opacity-90 disabled:opacity-50"
                      >
                        <Check className="h-3 w-3" />
                        Принять
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDecline(inv)}
                        disabled={busy}
                        className="flex h-6 items-center gap-1 rounded-md border border-(--color-border) px-2 text-[10px] text-(--color-muted-foreground) hover:text-(--color-foreground) disabled:opacity-50"
                      >
                        <X className="h-3 w-3" />
                        Отклонить
                      </button>
                    </div>
                  </div>
                </div>
              )
            })}
          </>
        )}

        {/* Email-настройки — deep-link в веб */}
        <div className="rounded-md border border-(--color-border) bg-(--color-card)/30 p-3">
          <div className="flex items-center gap-1.5">
            <Mail className="h-3 w-3 text-(--color-muted-foreground)" />
            <h3 className="text-[11px] font-medium">Email-рассылки</h3>
          </div>
          <p className="mt-1 text-[10px] text-(--color-muted-foreground)">
            Smart Insights, дайджесты и другие категории email-уведомлений настраиваются в веб-приложении.
          </p>
          <Button
            type="button"
            size="sm"
            variant="outline"
            onClick={openEmailSettings}
            className="mt-2 gap-1.5 w-full"
          >
            <ExternalLink className="h-3.5 w-3.5" />
            Настройки email
          </Button>
        </div>
      </div>
    </div>
  )
}
