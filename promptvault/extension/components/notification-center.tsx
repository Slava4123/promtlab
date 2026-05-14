import { useState } from "react"
import { Bell, Check, Loader2, Mail, X, ShieldCheck, Shield, Eye } from "lucide-react"
import { formatRelativeDate } from "@pv/shared/utils/format-date"
import {
  useMyInvitations,
  useAcceptInvitation,
  useDeclineInvitation,
} from "../hooks/use-invitations"
import { useToast } from "./ui/toaster"
import { cn } from "../lib/utils"
import type { TeamRole, TeamInvitation } from "../lib/types"

const ROLE_META: Record<TeamRole, { label: string; icon: React.ComponentType<{ className?: string }>; color: string }> = {
  owner: { label: "Владелец", icon: ShieldCheck, color: "text-amber-500" },
  editor: { label: "Редактор", icon: Shield, color: "text-(--color-brand)" },
  viewer: { label: "Просмотр", icon: Eye, color: "text-(--color-muted-foreground)" },
}

// Колокольчик с непрочитанным счётчиком + popover со списком приглашений.
// Только pending invitations — accepted/declined не показываем.
export function NotificationCenter() {
  const [open, setOpen] = useState(false)
  const { toast } = useToast()
  const invitations = useMyInvitations()
  const acceptMut = useAcceptInvitation()
  const declineMut = useDeclineInvitation()

  const pending = (invitations.data ?? []).filter((i) => i.status === "pending")
  const count = pending.length

  async function handleAccept(inv: TeamInvitation) {
    try {
      await acceptMut.mutateAsync(inv.id)
      toast({ title: `Вы в команде «${inv.team_name}»`, variant: "success" })
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
    } catch {
      toast({ title: "Не удалось отклонить", variant: "error" })
    }
  }

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="relative rounded-md p-1.5 text-(--color-muted-foreground) hover:bg-(--color-muted) hover:text-(--color-foreground)"
        aria-label={count > 0 ? `Уведомления (${count})` : "Уведомления"}
        title={count > 0 ? `Приглашений: ${count}` : "Нет уведомлений"}
      >
        <Bell className="h-4 w-4" />
        {count > 0 && (
          <span className="absolute -right-0.5 -top-0.5 flex h-3.5 min-w-3.5 items-center justify-center rounded-full bg-(--color-primary) px-1 text-[9px] font-bold text-(--color-primary-foreground)">
            {count}
          </span>
        )}
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} aria-hidden />
          <div className="absolute right-0 top-full z-50 mt-1 w-72 rounded-lg border border-(--color-border) bg-(--color-background) shadow-xl">
            <header className="border-b border-(--color-border) px-3 py-2">
              <h3 className="text-xs font-semibold">Уведомления</h3>
            </header>

            <div className="max-h-80 overflow-y-auto">
              {invitations.isPending ? (
                <div className="flex justify-center py-6">
                  <Loader2 className="h-4 w-4 animate-spin text-(--color-muted-foreground)" />
                </div>
              ) : count === 0 ? (
                <div className="flex flex-col items-center gap-2 py-8 text-center">
                  <Mail className="h-6 w-6 text-(--color-muted-foreground)/40" />
                  <p className="text-[11px] text-(--color-muted-foreground)">
                    Нет новых уведомлений
                  </p>
                </div>
              ) : (
                <ul>
                  {pending.map((inv) => {
                    const meta = ROLE_META[inv.role]
                    const Icon = meta.icon
                    const busy = acceptMut.isPending || declineMut.isPending
                    return (
                      <li
                        key={inv.id}
                        className="border-b border-(--color-border) px-3 py-2.5 last:border-0"
                      >
                        <div className="text-[11px] font-medium">
                          Приглашение в «{inv.team_name}»
                        </div>
                        <div className="mt-0.5 flex items-center gap-1.5 text-[10px] text-(--color-muted-foreground)">
                          <span>от {inv.inviter_name}</span>
                          <span>•</span>
                          <span>{formatRelativeDate(inv.created_at)}</span>
                        </div>
                        <div className="mt-1.5 flex items-center gap-2">
                          <span className={cn("flex items-center gap-1 text-[10px]", meta.color)}>
                            <Icon className="h-3 w-3" />
                            {meta.label}
                          </span>
                          <div className="ml-auto flex gap-1">
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
                      </li>
                    )
                  })}
                </ul>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
