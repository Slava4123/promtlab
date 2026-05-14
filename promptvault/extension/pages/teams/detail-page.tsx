import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import {
  ArrowLeft,
  Users,
  UserPlus,
  Trash2,
  Loader2,
  Edit2,
  ShieldCheck,
  Shield,
  Eye,
} from "lucide-react"
import { Button } from "../../components/ui/button"
import { Input } from "../../components/ui/input"
import { Label } from "../../components/ui/label"
import { ConfirmDialog } from "../../components/ui/confirm-dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../components/ui/select"
import { useToast } from "../../components/ui/toaster"
import {
  useTeam,
  useUpdateTeam,
  useDeleteTeam,
  useInviteTeamMember,
  useRemoveTeamMember,
  useUpdateTeamMemberRole,
} from "../../hooks/use-teams-crud"
import { useWorkspace } from "../../hooks/use-workspace"
import type { TeamRole } from "../../lib/types"
import { cn } from "../../lib/utils"

const ROLE_META: Record<TeamRole, { label: string; icon: React.ComponentType<{ className?: string }>; color: string }> = {
  owner: { label: "Владелец", icon: ShieldCheck, color: "text-amber-500" },
  editor: { label: "Редактор", icon: Shield, color: "text-(--color-brand)" },
  viewer: { label: "Просмотр", icon: Eye, color: "text-(--color-muted-foreground)" },
}

export function TeamDetailPage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const teamQuery = useTeam(slug ?? null)
  const updateMut = useUpdateTeam(slug ?? "")
  const deleteMut = useDeleteTeam()
  const inviteMut = useInviteTeamMember(slug ?? "")
  const removeMut = useRemoveTeamMember(slug ?? "")
  const updateRoleMut = useUpdateTeamMemberRole(slug ?? "")
  const { setWorkspaceId } = useWorkspace()

  const [inviteOpen, setInviteOpen] = useState(false)
  const [editName, setEditName] = useState("")
  const [editingName, setEditingName] = useState(false)
  const [inviteEmail, setInviteEmail] = useState("")
  const [inviteRole, setInviteRole] = useState<"editor" | "viewer">("editor")
  const [removeId, setRemoveId] = useState<number | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)

  if (teamQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  const team = teamQuery.data
  if (!team) {
    return (
      <div className="flex h-full items-center justify-center text-sm text-(--color-muted-foreground)">
        Команда не найдена
      </div>
    )
  }

  const isOwner = team.role === "owner"
  const canWrite = isOwner || team.role === "editor"

  async function handleInvite() {
    if (!inviteEmail.trim()) return
    try {
      await inviteMut.mutateAsync({ email: inviteEmail.trim(), role: inviteRole })
      toast({ title: "Приглашение отправлено", variant: "success" })
      setInviteEmail("")
      setInviteOpen(false)
    } catch (err) {
      toast({
        title: "Не удалось пригласить",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  async function handleRemove() {
    if (removeId === null) return
    try {
      await removeMut.mutateAsync(removeId)
      toast({ title: "Участник удалён", variant: "info" })
    } catch {
      toast({ title: "Не удалось удалить", variant: "error" })
    } finally {
      setRemoveId(null)
    }
  }

  async function handleRoleChange(memberId: number, role: TeamRole) {
    try {
      await updateRoleMut.mutateAsync({ memberId, role })
      toast({ title: "Роль обновлена", variant: "success" })
    } catch {
      toast({ title: "Не удалось обновить роль", variant: "error" })
    }
  }

  async function handleSaveName() {
    if (!editName.trim()) return
    try {
      await updateMut.mutateAsync({ name: editName.trim() })
      toast({ title: "Сохранено", variant: "success" })
      setEditingName(false)
    } catch {
      toast({ title: "Не удалось сохранить", variant: "error" })
    }
  }

  async function handleDeleteTeam() {
    try {
      await deleteMut.mutateAsync(team!.slug)
      toast({ title: "Команда удалена", variant: "info" })
      setWorkspaceId(null)
      navigate("/teams")
    } catch {
      toast({ title: "Не удалось удалить", variant: "error" })
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        {editingName ? (
          <div className="flex flex-1 items-center gap-1">
            <Input
              value={editName}
              onChange={(e) => setEditName(e.target.value)}
              className="h-7 text-sm"
              autoFocus
            />
            <Button type="button" variant="brand" size="sm" onClick={handleSaveName}>
              Сохранить
            </Button>
            <Button type="button" size="sm" variant="ghost" onClick={() => setEditingName(false)}>
              Отмена
            </Button>
          </div>
        ) : (
          <h2 className="flex-1 truncate text-sm font-semibold">{team.name}</h2>
        )}
        {isOwner && !editingName && (
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => {
              setEditName(team.name)
              setEditingName(true)
            }}
            aria-label="Редактировать"
          >
            <Edit2 className="h-3.5 w-3.5" />
          </Button>
        )}
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        {team.description && (
          <p className="text-xs text-(--color-muted-foreground)">{team.description}</p>
        )}

        {/* Members */}
        <section>
          <div className="mb-2 flex items-center justify-between">
            <h3 className="flex items-center gap-1.5 text-xs font-semibold">
              <Users className="h-3.5 w-3.5" />
              Участники ({team.members.length})
            </h3>
            {canWrite && (
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => setInviteOpen(true)}
                className="h-7 gap-1 text-[10px]"
              >
                <UserPlus className="h-3 w-3" />
                Пригласить
              </Button>
            )}
          </div>
          <ul className="space-y-1">
            {team.members.map((m) => {
              const meta = ROLE_META[m.role]
              const Icon = meta.icon
              return (
                <li
                  key={m.user_id}
                  className="flex items-center gap-2 rounded-md border border-(--color-border) bg-(--color-card) p-2 text-xs"
                >
                  <div className="flex h-7 w-7 items-center justify-center rounded-full bg-(--color-brand-muted) text-[10px] font-semibold text-(--color-brand)">
                    {(m.name ?? m.email).charAt(0).toUpperCase()}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="truncate font-medium">{m.name || m.email}</div>
                    {m.name && <div className="truncate text-[10px] text-(--color-muted-foreground)">{m.email}</div>}
                  </div>
                  <span className={cn("flex items-center gap-1 text-[10px]", meta.color)}>
                    <Icon className="h-3 w-3" />
                    {meta.label}
                  </span>
                  {isOwner && m.role !== "owner" && (
                    <>
                      <select
                        value={m.role}
                        onChange={(e) => handleRoleChange(m.user_id, e.target.value as TeamRole)}
                        className="h-6 rounded border border-(--color-border) bg-(--color-background) px-1 text-[10px]"
                      >
                        <option value="viewer">Просмотр</option>
                        <option value="editor">Редактор</option>
                      </select>
                      <button
                        type="button"
                        onClick={() => setRemoveId(m.user_id)}
                        className="rounded p-0.5 text-(--color-muted-foreground) hover:text-(--color-destructive)"
                        aria-label="Удалить"
                      >
                        <Trash2 className="h-3 w-3" />
                      </button>
                    </>
                  )}
                </li>
              )
            })}
          </ul>
        </section>

        {/* Quick links */}
        <section className="space-y-1">
          <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
            Разделы команды
          </div>
          <button
            type="button"
            onClick={() => navigate(`/teams/${slug}/activity`)}
            className="flex w-full items-center justify-between rounded-md border border-(--color-border) bg-(--color-card) px-3 py-2 text-xs hover:bg-(--color-muted)/40"
          >
            <span>Активность</span>
            <span className="text-(--color-muted-foreground)">→</span>
          </button>
        </section>

        {/* Danger zone */}
        {isOwner && (
          <section>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setDeleteOpen(true)}
              className="w-full text-(--color-destructive) border-(--color-destructive)/30 gap-1.5"
            >
              <Trash2 className="h-3.5 w-3.5" />
              Удалить команду
            </Button>
          </section>
        )}
      </div>

      {/* Invite dialog */}
      {inviteOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          {/* eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions -- modal backdrop */}
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={() => setInviteOpen(false)} />
          <div className="relative w-full max-w-sm rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
            <h3 className="mb-3 text-sm font-semibold">Пригласить участника</h3>
            <div className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="invite-email">Email</Label>
                <Input
                  id="invite-email"
                  type="email"
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                  placeholder="user@example.com"
                  autoFocus
                />
              </div>
              <div className="space-y-1">
                <Label htmlFor="invite-role">Роль</Label>
                <Select
                  value={inviteRole}
                  onValueChange={(v) => setInviteRole(v as "editor" | "viewer")}
                >
                  <SelectTrigger id="invite-role">
                    <SelectValue placeholder="Выберите роль" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="editor">
                      <div className="flex flex-col items-start text-left">
                        <span className="font-medium">Редактор</span>
                        <span className="text-[10px] text-(--color-muted-foreground)">
                          Может создавать и редактировать
                        </span>
                      </div>
                    </SelectItem>
                    <SelectItem value="viewer">
                      <div className="flex flex-col items-start text-left">
                        <span className="font-medium">Просмотр</span>
                        <span className="text-[10px] text-(--color-muted-foreground)">
                          Только чтение
                        </span>
                      </div>
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <Button type="button" variant="outline" size="sm" onClick={() => setInviteOpen(false)}>
                Отмена
              </Button>
              <Button type="button" variant="brand" size="sm" onClick={handleInvite} disabled={inviteMut.isPending}>
                Пригласить
              </Button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={removeId !== null}
        title="Удалить участника?"
        description="Участник потеряет доступ к промптам команды."
        confirmLabel="Удалить"
        variant="destructive"
        onConfirm={handleRemove}
        onClose={() => setRemoveId(null)}
      />
      <ConfirmDialog
        open={deleteOpen}
        title="Удалить команду?"
        description="Все промпты, коллекции и цепочки команды переместятся в корзину. Действие можно отменить в течение 30 дней."
        confirmLabel="Удалить"
        variant="destructive"
        onConfirm={handleDeleteTeam}
        onClose={() => setDeleteOpen(false)}
      />
    </div>
  )
}
