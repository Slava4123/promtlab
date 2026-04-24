import { useState, useMemo } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { Loader2, ArrowLeft, Ban, Play, KeyRound, Check } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { ActionDialog } from "@/components/admin/action-dialog"
import { useBadges } from "@/hooks/use-badges"
import {
  useAdminUserDetail,
  useFreezeUser,
  useUnfreezeUser,
  useResetPassword,
  useGrantBadge,
  useRevokeBadge,
} from "@/hooks/admin/use-admin-users"
import { cn } from "@/lib/utils"

type DialogType = "freeze" | "unfreeze" | "reset-password" | "revoke-badge" | null

// RevokeTarget содержит badge_id + title + icon для читаемого dialog description.
interface RevokeTarget {
  id: string
  title: string
  icon: string
}

export default function AdminUserDetailPage() {
  const { id } = useParams<{ id: string }>()
  const userId = Number(id)
  const navigate = useNavigate()

  const { data: user, isLoading } = useAdminUserDetail(userId)
  const { data: badgesCatalog } = useBadges()

  const freeze = useFreezeUser()
  const unfreeze = useUnfreezeUser()
  const resetPw = useResetPassword()
  const grantBadge = useGrantBadge()
  const revokeBadge = useRevokeBadge()

  const [dialog, setDialog] = useState<DialogType>(null)
  const [revokeTarget, setRevokeTarget] = useState<RevokeTarget | null>(null)

  // Set unlocked badge_id'шек для быстрого O(1) lookup'а в рендере.
  const unlockedSet = useMemo(
    () => new Set(user?.unlocked_badge_ids ?? []),
    [user?.unlocked_badge_ids],
  )

  if (isLoading) {
    return (
      <div className="flex h-40 items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
      </div>
    )
  }
  if (!user) {
    return (
      <div className="rounded-lg border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">
        Пользователь не найден
      </div>
    )
  }

  const handleFreeze = async () => {
    await freeze.mutateAsync(userId)
    toast.success("Пользователь заморожен")
  }
  const handleUnfreeze = async () => {
    await unfreeze.mutateAsync(userId)
    toast.success("Пользователь активирован")
  }
  const handleResetPassword = async (totpCode?: string) => {
    if (!totpCode) return
    await resetPw.mutateAsync({ id: userId, totpCode })
    toast.success("Код сброса отправлен на email")
  }
  const handleRevokeBadge = async (totpCode?: string) => {
    if (!totpCode || !revokeTarget) return
    await revokeBadge.mutateAsync({
      userId,
      badgeId: revokeTarget.id,
      totpCode,
    })
    toast.success("Бейдж отозван")
    setRevokeTarget(null)
  }

  const handleGrantBadge = async (badgeId: string) => {
    try {
      await grantBadge.mutateAsync({ userId, badgeId })
      toast.success("Бейдж выдан")
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Ошибка")
    }
  }

  // Каталог всех 11 бейджей — нужен для отображения title/icon.
  const allBadges = badgesCatalog?.items ?? []

  return (
    <div className="space-y-5">
      <div className="min-w-0">
        <button
          onClick={() => navigate("/admin/users")}
          className="mb-2 flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-3 w-3" />
          К списку
        </button>
        <h2 className="truncate text-xl font-semibold">{user.name}</h2>
        <p className="truncate text-sm text-muted-foreground">
          {user.email}
          {user.username && ` · @${user.username}`}
        </p>
      </div>

      {/* Status row — wrap чтобы не вылезать на мобилке */}
      <div className="flex flex-wrap items-center gap-2 text-xs">
        <span className="rounded-md bg-muted/30 px-2 py-1">
          Роль: <strong>{user.role}</strong>
        </span>
        <span
          className={
            user.status === "frozen"
              ? "rounded-md bg-destructive/15 px-2 py-1 text-destructive"
              : "rounded-md bg-emerald-500/15 px-2 py-1 text-emerald-400"
          }
        >
          Статус: <strong>{user.status}</strong>
        </span>
        <span className="rounded-md bg-muted/30 px-2 py-1">
          Тариф: <strong>{user.tier}</strong>
        </span>
        <span className="rounded-md bg-muted/30 px-2 py-1">
          Email {user.email_verified ? "✓" : "×"}
        </span>
      </div>

      {/* Stats — 2 col на мобилке, 4 на десктопе */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <StatCard label="Промпты" value={user.prompt_count} />
        <StatCard label="Коллекции" value={user.collection_count} />
        <StatCard label="Бейджи" value={user.badge_count} />
        <StatCard label="Использований" value={user.total_usage} />
      </div>

      {/* Linked providers */}
      {user.linked_providers.length > 0 && (
        <div className="rounded-xl border border-border p-4">
          <h3 className="mb-2 text-xs uppercase tracking-wider text-muted-foreground">
            Связанные аккаунты
          </h3>
          <div className="flex flex-wrap gap-2">
            {user.linked_providers.map((p) => (
              <span key={p} className="rounded-md bg-muted/30 px-2 py-1 text-xs">
                {p}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Actions — кнопки wrap на мобилке */}
      <div className="space-y-3 rounded-xl border border-border p-4">
        <h3 className="text-xs uppercase tracking-wider text-muted-foreground">
          Действия
        </h3>
        <div className="flex flex-wrap gap-2">
          {user.status === "active" ? (
            <Button
              variant="destructive"
              size="sm"
              onClick={() => setDialog("freeze")}
            >
              <Ban className="mr-1.5 h-3.5 w-3.5" />
              Заморозить
            </Button>
          ) : (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setDialog("unfreeze")}
            >
              <Play className="mr-1.5 h-3.5 w-3.5" />
              Разморозить
            </Button>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={() => setDialog("reset-password")}
          >
            <KeyRound className="mr-1.5 h-3.5 w-3.5" />
            Сброс пароля
          </Button>
        </div>
      </div>

      {/* Badges — grid 1-col на мобилке, 2-col на десктопе.
          Unlocked бейджи visually выделены, кнопки disabled правильно. */}
      {allBadges.length > 0 && (
        <div className="space-y-3 rounded-xl border border-border p-4">
          <h3 className="text-xs uppercase tracking-wider text-muted-foreground">
            Управление бейджами
          </h3>
          <div className="grid gap-1.5 sm:grid-cols-2">
            {allBadges.map((b) => {
              const isUnlocked = unlockedSet.has(b.id)
              return (
                <div
                  key={b.id}
                  className={cn(
                    "flex items-center justify-between gap-2 rounded-md border px-2 py-1.5",
                    isUnlocked
                      ? "border-violet-500/40 bg-violet-500/5"
                      : "border-border",
                  )}
                >
                  <div className="flex min-w-0 items-center gap-2 text-xs">
                    <span className="text-base leading-none">{b.icon}</span>
                    <span className="truncate">{b.title}</span>
                    {isUnlocked && (
                      <Check className="h-3 w-3 shrink-0 text-violet-400" />
                    )}
                  </div>
                  <div className="flex shrink-0 gap-1">
                    <Button
                      variant="ghost"
                      size="xs"
                      disabled={isUnlocked}
                      onClick={() => handleGrantBadge(b.id)}
                      title={isUnlocked ? "Уже разблокирован" : "Выдать"}
                    >
                      Выдать
                    </Button>
                    <Button
                      variant="ghost"
                      size="xs"
                      disabled={!isUnlocked}
                      onClick={() => {
                        setRevokeTarget({ id: b.id, title: b.title, icon: b.icon })
                        setDialog("revoke-badge")
                      }}
                      title={!isUnlocked ? "Нечего забирать" : "Забрать"}
                    >
                      Забрать
                    </Button>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Dialogs */}
      <ActionDialog
        open={dialog === "freeze"}
        onOpenChange={(o) => !o && setDialog(null)}
        title="Заморозить пользователя?"
        description={`${user.email} не сможет войти, пока вы не снимете заморозку.`}
        confirmLabel="Заморозить"
        onConfirm={handleFreeze}
      />
      <ActionDialog
        open={dialog === "unfreeze"}
        onOpenChange={(o) => !o && setDialog(null)}
        title="Разморозить пользователя?"
        description={`${user.email} сможет войти снова.`}
        confirmLabel="Разморозить"
        onConfirm={handleUnfreeze}
      />
      <ActionDialog
        open={dialog === "reset-password"}
        onOpenChange={(o) => !o && setDialog(null)}
        title="Отправить код сброса пароля?"
        description="Пользователь получит email с кодом. Текущий пароль перестанет работать после сброса."
        confirmLabel="Отправить"
        requireTOTP
        onConfirm={handleResetPassword}
      />
      <ActionDialog
        open={dialog === "revoke-badge"}
        onOpenChange={(o) => {
          if (!o) {
            setDialog(null)
            setRevokeTarget(null)
          }
        }}
        title="Забрать бейдж?"
        description={
          revokeTarget
            ? `Бейдж "${revokeTarget.icon} ${revokeTarget.title}" будет удалён у пользователя.`
            : ""
        }
        confirmLabel="Забрать"
        requireTOTP
        onConfirm={handleRevokeBadge}
      />
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-xl border border-border p-3">
      <p className="text-[0.7rem] text-muted-foreground">{label}</p>
      <p className="mt-1 text-xl font-semibold tabular-nums">{value}</p>
    </div>
  )
}
