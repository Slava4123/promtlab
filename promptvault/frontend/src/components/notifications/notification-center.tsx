// Объединённый центр уведомлений: приглашения в команды + over-limit warnings.
// Рендерится в шапке AppLayout вместо отдельного «Приглашения» popover'а.
//
// Read-state хранится в localStorage (см. lib/notifications), поэтому Pro/Max
// юзер с over-limit'ом видит уведомление один раз — потом нажимает «Прочитано»
// и оно исчезает до тех пор, пока usage не изменится (id уведомления включает
// used+limit — изменение цифр пересоздаёт уведомление).

import { useEffect, useMemo, useState } from "react"
import { Link, useNavigate } from "react-router-dom"
import { AlertTriangle, Bell, Check, ChevronRight, Mail, X } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Popover, PopoverTrigger, PopoverContent } from "@/components/ui/popover"
import { RoleBadge } from "@/components/teams/role-badge"
import {
  useAcceptInvitation,
  useDeclineInvitation,
  useMyInvitations,
} from "@/hooks/use-teams"
import { useUsage } from "@/hooks/use-subscription"
import {
  buildNotifications,
  clearOldReads,
  isRead,
  markAllRead,
  markRead,
  type Notification,
} from "@/lib/notifications"

export function NotificationCenter() {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)

  const { data: invitations } = useMyInvitations()
  const { data: usage } = useUsage()
  const acceptInvitation = useAcceptInvitation()
  const declineInvitation = useDeclineInvitation()

  // Бамп заставляет реагировать на маркировку «прочитано» (localStorage без событий).
  const [readBump, setReadBump] = useState(0)
  const bumpRead = () => setReadBump((x) => x + 1)

  const allNotifications = useMemo(
    () => buildNotifications(invitations, usage),
    [invitations, usage],
  )

  // unread = всё что не помечено в localStorage И живо в текущем списке.
  // readBump в deps намеренный — bumpRead() заставляет useMemo пере-вычислить
  // фильтр после mark-as-read (localStorage без событий, ESLint не видит implicit
  // dep через isRead → localStorage).
  /* eslint-disable-next-line react-hooks/exhaustive-deps -- readBump trigger для re-evaluate filter после localStorage mark-as-read. */
  const unread = useMemo(() => allNotifications.filter((n) => !isRead(n.id)), [allNotifications, readBump])

  // Чистим устаревшие read-id (когда usage изменился, старый quota-id больше
  // не появляется — мы из localStorage его удалим).
  useEffect(() => {
    clearOldReads(new Set(allNotifications.map((n) => n.id)))
  }, [allNotifications])

  const unreadCount = unread.length
  // Если среди непрочитанных есть quota_over — бейдж колокольчика амбер
  // (warning), иначе бренд-фиолетовый. Помогает юзеру понять, что там срочно,
  // ещё до открытия popover'а.
  const hasUrgentUnread = unread.some((n) => n.kind === "quota_over")
  const visible: Notification[] = open ? allNotifications : []  // popover-only render

  const handleAccept = (id: number) => {
    acceptInvitation.mutate(id, {
      onSuccess: () => {
        toast.success("Вы присоединились к команде")
        markRead(`invitation-${id}`)
        bumpRead()
      },
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }
  const handleDecline = (id: number) => {
    declineInvitation.mutate(id, {
      onSuccess: () => {
        toast.success("Приглашение отклонено")
        markRead(`invitation-${id}`)
        bumpRead()
      },
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }
  const handleMarkRead = (n: Notification) => {
    markRead(n.id)
    bumpRead()
  }
  const handleMarkAllRead = () => {
    markAllRead(allNotifications.map((n) => n.id))
    bumpRead()
  }

  return (
    // MJ-33: переход с custom popover (fixed inset-0 backdrop + absolute panel)
    // на Base UI Popover. Дает focus-trap, Esc to close, корректные ARIA-роли,
    // outside-click handling из коробки. Кастомный backdrop был источником
    // багов с z-index при модалках поверх.
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          aria-label="Уведомления"
          className="relative flex h-11 w-11 cursor-pointer items-center justify-center rounded-lg border border-border bg-muted/20 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          <Bell className="h-4 w-4" />
          {unreadCount > 0 && (
            <span
              className={`absolute -right-1 -top-1 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[9px] font-bold ${
                hasUrgentUnread
                  ? "bg-amber-500 text-amber-950"
                  : "bg-brand text-brand-foreground"
              }`}
            >
              {unreadCount}
            </span>
          )}
        </button>
      </PopoverTrigger>
      <PopoverContent
        align="end"
        sideOffset={8}
        className="w-[min(22rem,calc(100vw-1rem))] p-0 sm:w-[22rem]"
      >
          <div>
            <div className="flex items-center justify-between gap-2 border-b border-border px-4 py-3">
              <p className="text-[0.85rem] font-medium text-foreground">Уведомления</p>
              <div className="flex items-center gap-2">
                {unreadCount > 0 && (
                  <button
                    onClick={handleMarkAllRead}
                    className="text-[0.7rem] text-muted-foreground transition-colors hover:text-foreground"
                  >
                    Прочитать все
                  </button>
                )}
              </div>
            </div>

            {visible.length === 0 ? (
              <div className="flex flex-col items-center px-4 py-8 text-center">
                <Mail aria-hidden="true" className="mb-2 h-6 w-6 text-muted-foreground/60" />
                <p className="text-[0.78rem] text-muted-foreground">Нет уведомлений</p>
              </div>
            ) : (
              <div className="max-h-[420px] overflow-y-auto">
                {visible.map((n) => {
                  const read = isRead(n.id)
                  // quota_over — это warning state: тонкая амбер-полоска слева
                  // и предупреждающая иконка рядом с заголовком. Прочитанное
                  // приглушаем opacity, чтобы не отвлекало.
                  const isWarning = n.kind === "quota_over"
                  return (
                    <div
                      key={n.id}
                      className={`relative border-b border-border px-4 py-3 last:border-0 ${
                        read ? "opacity-60" : ""
                      } ${isWarning ? "border-l-2 border-l-amber-500/70 bg-amber-500/[0.04]" : ""}`}
                    >
                      <div className="mb-1.5 flex items-start justify-between gap-2">
                        <div className="flex items-start gap-1.5">
                          {isWarning && (
                            <AlertTriangle
                              aria-hidden="true"
                              className="mt-[1px] h-3.5 w-3.5 shrink-0 text-amber-500"
                            />
                          )}
                          <p className="text-[0.78rem] font-medium leading-snug text-foreground">
                            {n.title}
                          </p>
                        </div>
                        {!read && (
                          <button
                            onClick={() => handleMarkRead(n)}
                            aria-label="Скрыть уведомление"
                            title="Скрыть"
                            className="-mt-0.5 -mr-1 rounded-md p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                          >
                            <X className="h-3.5 w-3.5" />
                          </button>
                        )}
                      </div>
                      <p className="mb-2 text-[0.7rem] leading-relaxed text-muted-foreground">
                        {n.body}
                      </p>

                      {n.kind === "invitation" && n.invitation && (
                        <div className="flex items-center gap-2">
                          <RoleBadge role={n.invitation.role} interactive={false} />
                          <div className="ml-auto flex gap-2">
                            <button
                              onClick={() => handleAccept(n.invitation!.id)}
                              disabled={acceptInvitation.isPending}
                              className="flex h-7 items-center gap-1 rounded-lg px-3 text-[0.72rem] font-medium text-brand-foreground [background:var(--brand-gradient)] transition-colors active:scale-[0.97]"
                            >
                              <Check className="h-3 w-3" />
                              Принять
                            </button>
                            <button
                              onClick={() => handleDecline(n.invitation!.id)}
                              disabled={declineInvitation.isPending}
                              className="flex h-7 items-center gap-1 rounded-lg border border-border px-3 text-[0.72rem] text-muted-foreground transition-colors hover:text-foreground"
                            >
                              <X className="h-3 w-3" />
                              Отклонить
                            </button>
                          </div>
                        </div>
                      )}

                      {n.kind === "quota_over" && n.cta && (
                        <div className="flex flex-wrap items-center gap-2">
                          <Button
                            asChild
                            size="sm"
                            variant="outline"
                            className="h-7 text-[0.72rem]"
                          >
                            <Link
                              to={n.cta.href}
                              onClick={() => { setOpen(false); handleMarkRead(n) }}
                            >
                              {n.cta.label}
                              <ChevronRight className="ml-1 h-3 w-3" />
                            </Link>
                          </Button>
                          {!read && (
                            <Button
                              size="sm"
                              variant="ghost"
                              className="h-7 text-[0.72rem] text-muted-foreground"
                              onClick={() => handleMarkRead(n)}
                            >
                              Скрыть
                            </Button>
                          )}
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            )}

            {unreadCount > 0 && (
              <div className="border-t border-border px-4 py-2">
                <button
                  onClick={() => { setOpen(false); navigate("/pricing") }}
                  className="text-[0.7rem] text-muted-foreground transition-colors hover:text-foreground"
                >
                  Посмотреть тарифы →
                </button>
              </div>
            )}
          </div>
      </PopoverContent>
    </Popover>
  )
}
