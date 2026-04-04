import { useState } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { ArrowLeft, Pencil, Trash2, UserPlus, Users, Loader2, AlertTriangle } from "lucide-react"
import { toast } from "sonner"

import { useTeam, useUpdateTeam, useDeleteTeam, useInviteMember, useTeamInvitations, useCancelInvitation, useUpdateMemberRole, useRemoveMember } from "@/hooks/use-teams"
import { useAuthStore } from "@/stores/auth-store"
import { RoleBadge } from "@/components/teams/role-badge"
import { MemberList } from "@/components/teams/member-list"
import { InviteDialog } from "@/components/teams/invite-dialog"
import type { TeamRole } from "@/api/types"

export default function TeamView() {
  const { slug = "" } = useParams()
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const { data: team, isLoading } = useTeam(slug)
  const updateTeam = useUpdateTeam()
  const deleteTeam = useDeleteTeam()
  const inviteMember = useInviteMember()
  const { data: pendingInvitations } = useTeamInvitations(slug)
  const cancelInvitation = useCancelInvitation()
  const updateMemberRole = useUpdateMemberRole()
  const removeMember = useRemoveMember()

  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [inviteOpen, setInviteOpen] = useState(false)
  const [editName, setEditName] = useState("")
  const [editDescription, setEditDescription] = useState("")

  const isOwner = team?.role === "owner"

  const openEdit = () => {
    if (!team) return
    setEditName(team.name)
    setEditDescription(team.description || "")
    setEditOpen(true)
  }

  const handleUpdate = async () => {
    if (!editName.trim() || !team) return
    try {
      await updateTeam.mutateAsync({ slug, name: editName, description: editDescription })
      toast.success("Команда обновлена")
      setEditOpen(false)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Ошибка")
    }
  }

  const handleDelete = () => {
    deleteTeam.mutate(slug, {
      onSuccess: () => { toast.success("Команда удалена"); navigate("/teams") },
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  const handleInvite = async (query: string, role: TeamRole) => {
    await inviteMember.mutateAsync({ slug, query, role })
    toast.success("Приглашение отправлено")
    setInviteOpen(false)
  }

  const handleCancelInvitation = (invitationId: number) => {
    cancelInvitation.mutate({ slug, invitationId }, {
      onSuccess: () => toast.success("Приглашение отменено"),
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  const handleChangeRole = (userId: number, role: TeamRole) => {
    updateMemberRole.mutate({ slug, userId, role }, {
      onSuccess: () => toast.success("Роль изменена"),
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  const handleRemove = (userId: number) => {
    const isSelf = userId === user?.id
    removeMember.mutate({ slug, userId }, {
      onSuccess: () => {
        if (isSelf) {
          toast.success("Вы покинули команду")
          navigate("/teams")
        } else {
          toast.success("Участник удалён")
        }
      },
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  if (isLoading) {
    return (
      <div className="mx-auto max-w-[64rem] space-y-5">
        <div className="h-5 w-24 animate-pulse rounded-md bg-white/[0.04]" />
        <div className="h-8 w-64 animate-pulse rounded-md bg-white/[0.04]" />
        <div className="h-4 w-48 animate-pulse rounded-md bg-white/[0.03]" />
        <div className="mt-8 space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-14 animate-pulse rounded-lg bg-white/[0.02]" />
          ))}
        </div>
      </div>
    )
  }

  if (!team) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <p className="text-base font-medium text-zinc-400">Команда не найдена</p>
        <button onClick={() => navigate("/teams")} className="mt-4 text-sm text-violet-400 hover:text-violet-300">
          Вернуться к командам
        </button>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-[64rem] space-y-6">
      {/* Back link */}
      <button
        onClick={() => navigate("/teams")}
        className="flex items-center gap-1.5 text-[0.8rem] text-zinc-500 transition-colors hover:text-zinc-300"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        Команды
      </button>

      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold tracking-tight">{team.name}</h1>
            <RoleBadge role={team.role} />
          </div>
          {team.description && (
            <p className="text-[0.85rem] text-zinc-500">{team.description}</p>
          )}
        </div>
        {isOwner && (
          <div className="flex gap-2">
            <button
              onClick={openEdit}
              className="flex h-8 items-center gap-1.5 rounded-lg px-3 text-[0.8rem] text-zinc-400 transition-all hover:text-zinc-200"
              style={{ border: "1px solid rgba(255,255,255,0.06)", background: "rgba(255,255,255,0.02)" }}
            >
              <Pencil className="h-3.5 w-3.5" />
              Редактировать
            </button>
            <button
              onClick={() => setDeleteOpen(true)}
              className="flex h-8 items-center gap-1.5 rounded-lg px-3 text-[0.8rem] text-red-400/70 transition-all hover:bg-red-500/10 hover:text-red-400"
              style={{ border: "1px solid rgba(239,68,68,0.1)" }}
            >
              <Trash2 className="h-3.5 w-3.5" />
              Удалить
            </button>
          </div>
        )}
      </div>

      {/* Members section */}
      <div
        className="rounded-xl p-5"
        style={{ border: "1px solid rgba(255,255,255,0.04)", background: "#0f0f12" }}
      >
        <div className="mb-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Users className="h-4 w-4 text-zinc-500" />
            <h2 className="text-[0.9rem] font-semibold text-white">Участники</h2>
            <span className="text-[0.75rem] text-zinc-600">{team.members.length}</span>
          </div>
          {isOwner && (
            <button
              onClick={() => setInviteOpen(true)}
              className="flex h-7 items-center gap-1.5 rounded-lg bg-violet-600 px-3 text-[0.75rem] font-medium text-white transition-all hover:bg-violet-500 active:scale-[0.97]"
            >
              <UserPlus className="h-3 w-3" />
              Пригласить
            </button>
          )}
        </div>

        <MemberList
          members={team.members}
          currentUserRole={team.role}
          currentUserId={user?.id ?? 0}
          onChangeRole={handleChangeRole}
          onRemove={handleRemove}
        />
      </div>

      {/* Invite Dialog */}
      {/* Pending Invitations */}
      {isOwner && pendingInvitations && pendingInvitations.length > 0 && (
        <div
          className="rounded-xl p-5"
          style={{ border: "1px solid rgba(251,191,36,0.1)", background: "#0f0f12" }}
        >
          <div className="mb-3 flex items-center gap-2">
            <UserPlus className="h-4 w-4 text-amber-500/70" />
            <h2 className="text-[0.9rem] font-semibold text-white">Ожидающие приглашения</h2>
            <span className="text-[0.75rem] text-zinc-600">{pendingInvitations.length}</span>
          </div>
          <div className="space-y-1">
            {pendingInvitations.map((inv) => (
              <div key={inv.id} className="flex items-center gap-3 rounded-lg px-3 py-2.5 hover:bg-white/[0.02]">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-amber-500/10 text-[0.75rem] font-medium text-amber-400">
                  {inv.name?.charAt(0).toUpperCase() || inv.email.charAt(0).toUpperCase()}
                </div>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-[0.8rem] font-medium text-white">{inv.name || inv.email}</p>
                  <p className="truncate text-[0.7rem] text-zinc-600">{inv.email}</p>
                </div>
                <RoleBadge role={inv.role} />
                <span className="text-[0.7rem] text-amber-500/60">Ожидает</span>
                <button
                  onClick={() => handleCancelInvitation(inv.id)}
                  className="rounded-md p-1 text-zinc-600 transition-colors hover:bg-red-500/10 hover:text-red-400"
                  title="Отменить приглашение"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      <InviteDialog
        open={inviteOpen}
        onClose={() => setInviteOpen(false)}
        onInvite={handleInvite}
        isPending={inviteMember.isPending}
      />

      {/* Edit Dialog */}
      {editOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setEditOpen(false)}>
          <div
            className="w-full max-w-md rounded-2xl p-6 space-y-4"
            style={{ border: "1px solid rgba(255,255,255,0.06)", background: "linear-gradient(145deg, #101015, #0d0d10)" }}
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="text-lg font-semibold text-white">Редактировать команду</h2>

            <div className="space-y-2">
              <label className="text-[0.8rem] font-medium text-zinc-300">Название</label>
              <input
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                autoFocus
                className="flex h-10 w-full rounded-lg px-3.5 text-sm text-white outline-none transition-all placeholder:text-zinc-600"
                style={{ border: "1px solid rgba(255,255,255,0.07)", background: "rgba(255,255,255,0.025)" }}
                onFocus={(e) => { e.target.style.borderColor = "rgba(139,92,246,0.4)"; e.target.style.boxShadow = "0 0 0 3px rgba(139,92,246,0.08)" }}
                onBlur={(e) => { e.target.style.borderColor = "rgba(255,255,255,0.07)"; e.target.style.boxShadow = "none" }}
                onKeyDown={(e) => e.key === "Enter" && handleUpdate()}
              />
            </div>

            <div className="space-y-2">
              <label className="text-[0.8rem] font-medium text-zinc-300">Описание</label>
              <textarea
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
                rows={2}
                className="flex w-full resize-none rounded-lg px-3.5 py-2.5 text-sm text-white outline-none transition-all placeholder:text-zinc-600"
                style={{ border: "1px solid rgba(255,255,255,0.07)", background: "rgba(255,255,255,0.025)" }}
                onFocus={(e) => { e.target.style.borderColor = "rgba(139,92,246,0.4)"; e.target.style.boxShadow = "0 0 0 3px rgba(139,92,246,0.08)" }}
                onBlur={(e) => { e.target.style.borderColor = "rgba(255,255,255,0.07)"; e.target.style.boxShadow = "none" }}
              />
            </div>

            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setEditOpen(false)}
                className="flex h-9 items-center rounded-lg px-4 text-[0.8rem] text-zinc-500 transition-all hover:text-zinc-300"
                style={{ border: "1px solid rgba(255,255,255,0.06)", background: "rgba(255,255,255,0.02)" }}
              >
                Отмена
              </button>
              <button
                onClick={handleUpdate}
                disabled={!editName.trim()}
                className="flex h-9 items-center gap-2 rounded-lg px-5 text-[0.8rem] font-medium text-white transition-all active:scale-[0.97] disabled:opacity-50"
                style={{ background: "linear-gradient(135deg, #7c3aed, #6d28d9)", boxShadow: "0 4px 16px -2px rgba(124,58,237,0.25)" }}
              >
                {updateTeam.isPending && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                Сохранить
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation */}
      {deleteOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setDeleteOpen(false)}>
          <div
            className="w-full max-w-sm rounded-2xl p-6 space-y-4"
            style={{ border: "1px solid rgba(239,68,68,0.12)", background: "linear-gradient(145deg, #101015, #0d0d10)" }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-500/10">
                <AlertTriangle className="h-5 w-5 text-red-400" />
              </div>
              <div>
                <h3 className="text-[0.9rem] font-semibold text-white">Удалить команду?</h3>
                <p className="text-[0.75rem] text-zinc-500">Все участники будут удалены, коллекции станут личными</p>
              </div>
            </div>
            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setDeleteOpen(false)}
                className="flex h-9 items-center rounded-lg px-4 text-[0.8rem] text-zinc-500 transition-all hover:text-zinc-300"
                style={{ border: "1px solid rgba(255,255,255,0.06)", background: "rgba(255,255,255,0.02)" }}
              >
                Отмена
              </button>
              <button
                onClick={handleDelete}
                className="flex h-9 items-center gap-2 rounded-lg px-4 text-[0.8rem] font-medium text-white transition-all active:scale-[0.97]"
                style={{ background: "linear-gradient(135deg, #dc2626, #b91c1c)", boxShadow: "0 4px 16px -2px rgba(220,38,38,0.25)" }}
              >
                {deleteTeam.isPending && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                Удалить
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
