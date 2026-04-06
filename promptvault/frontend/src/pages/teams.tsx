import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Plus, Users, Pencil, Trash2, Loader2, AlertTriangle } from "lucide-react"
import { toast } from "sonner"

import { useTeams, useCreateTeam, useUpdateTeam, useDeleteTeam } from "@/hooks/use-teams"
import { RoleBadge } from "@/components/teams/role-badge"
import type { Team } from "@/api/types"

function plural(n: number, one: string, few: string, many: string) {
  const mod10 = n % 10, mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return few
  return many
}

export default function Teams() {
  const navigate = useNavigate()
  const { data: teams, isLoading } = useTeams()
  const createTeam = useCreateTeam()
  const updateTeam = useUpdateTeam()
  const deleteTeam = useDeleteTeam()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deletingSlug, setDeletingSlug] = useState<string | null>(null)
  const [editing, setEditing] = useState<Team | null>(null)
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")

  const openCreate = () => {
    setEditing(null)
    setName("")
    setDescription("")
    setDialogOpen(true)
  }

  const openEdit = (t: Team) => {
    setEditing(t)
    setName(t.name)
    setDescription(t.description || "")
    setDialogOpen(true)
  }

  const handleSave = async () => {
    if (!name.trim()) return
    try {
      if (editing) {
        await updateTeam.mutateAsync({ slug: editing.slug, name, description })
        toast.success("Команда обновлена")
      } else {
        await createTeam.mutateAsync({ name, description })
        toast.success("Команда создана")
      }
      setDialogOpen(false)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Ошибка")
    }
  }

  const confirmDelete = (slug: string) => {
    setDeletingSlug(slug)
    setDeleteDialogOpen(true)
  }

  const handleDelete = () => {
    if (!deletingSlug) return
    deleteTeam.mutate(deletingSlug, {
      onSuccess: () => { toast.success("Команда удалена"); setDeleteDialogOpen(false) },
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  return (
    <div className="mx-auto max-w-[64rem] space-y-5">
      {/* Header */}
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Команды</h1>
          <p className="mt-0.5 text-[0.8rem] text-muted-foreground">Совместная работа над промптами</p>
        </div>
        <button
          onClick={openCreate}
          className="flex h-8 items-center gap-1.5 rounded-lg bg-violet-600 px-3.5 text-[0.8rem] font-medium text-white shadow-lg shadow-violet-600/10 transition-all hover:bg-violet-500 hover:shadow-violet-500/20 active:scale-[0.97]"
        >
          <Plus className="h-3.5 w-3.5" />
          Новая команда
        </button>
      </div>

      {/* List */}
      {isLoading ? (
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="rounded-xl border border-border bg-card p-5">
              <div className="mb-3 h-9 w-9 animate-pulse rounded-lg bg-muted/40" />
              <div className="mb-2 h-4 w-2/3 animate-pulse rounded-md bg-muted/40" />
              <div className="h-3 w-1/2 animate-pulse rounded-md bg-muted/30" />
            </div>
          ))}
        </div>
      ) : !teams || teams.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-violet-500/[0.08] ring-1 ring-violet-500/10">
            <Users className="h-7 w-7 text-violet-400/60" />
          </div>
          <p className="text-base font-medium text-muted-foreground">Пока нет команд</p>
          <p className="mt-1 text-sm text-muted-foreground">Создайте команду для совместной работы над промптами</p>
          <button
            onClick={openCreate}
            className="mt-5 flex h-8 items-center gap-1.5 rounded-lg bg-violet-600 px-4 text-[0.8rem] font-medium text-white shadow-lg shadow-violet-600/10 transition-all hover:bg-violet-500 active:scale-[0.97]"
          >
            <Plus className="h-3.5 w-3.5" />
            Создать команду
          </button>
        </div>
      ) : (
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
          {teams.map((t) => (
            <div
              key={t.id}
              className="group cursor-pointer rounded-xl border border-border bg-card p-5 transition-all duration-200 hover:-translate-y-0.5 hover:border-violet-500/20 hover:shadow-[0_8px_32px_-8px_rgba(0,0,0,0.5),0_0_0_1px_rgba(139,92,246,0.1)]"
              onClick={() => navigate(`/teams/${t.slug}`)}
            >
              <div className="mb-3 flex items-start justify-between">
                <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-violet-500/[0.08] ring-1 ring-inset ring-violet-500/15">
                  <Users className="h-4 w-4 text-violet-400" />
                </div>
                <div className="flex items-center gap-1">
                  <RoleBadge role={t.role} />
                  {t.role === "owner" && (
                    <div className="flex gap-1 sm:opacity-0 sm:transition-opacity sm:group-hover:opacity-100">
                      <button
                        className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
                        onClick={(e) => { e.stopPropagation(); openEdit(t) }}
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button
                        className="rounded-md p-1 text-muted-foreground hover:bg-red-500/10 hover:text-red-400"
                        onClick={(e) => { e.stopPropagation(); confirmDelete(t.slug) }}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  )}
                </div>
              </div>
              <h3 className="mb-1 text-[0.85rem] font-medium text-foreground">{t.name}</h3>
              {t.description && (
                <p className="mb-3 text-[0.75rem] text-muted-foreground line-clamp-2">{t.description}</p>
              )}
              <div className="flex items-center gap-1.5 text-[0.7rem] text-muted-foreground">
                <Users className="h-3 w-3" />
                <span>{t.member_count} {plural(t.member_count, "участник", "участника", "участников")}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create/Edit Dialog */}
      {dialogOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setDialogOpen(false)}>
          <div
            className="w-full max-w-md rounded-2xl border border-border bg-card p-6 space-y-4"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="text-lg font-semibold text-foreground">{editing ? "Редактировать команду" : "Новая команда"}</h2>

            <div className="space-y-2">
              <label className="text-[0.8rem] font-medium text-foreground">Название</label>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Например: Backend-разработка"
                autoFocus
                className="flex h-11 w-full rounded-lg border border-border bg-background px-3.5 text-sm text-foreground outline-none transition-all placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
                onKeyDown={(e) => e.key === "Enter" && handleSave()}
              />
            </div>

            <div className="space-y-2">
              <label className="text-[0.8rem] font-medium text-foreground">Описание <span className="text-muted-foreground">(необязательно)</span></label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Чем занимается команда?"
                rows={2}
                className="flex w-full resize-none rounded-lg border border-border bg-background px-3.5 py-2.5 text-sm text-foreground outline-none transition-all placeholder:text-muted-foreground focus:border-violet-500/40 focus:ring-3 focus:ring-violet-500/10"
              />
            </div>

            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setDialogOpen(false)}
                className="flex h-9 items-center rounded-lg border border-border bg-card px-4 text-[0.8rem] text-muted-foreground transition-all hover:text-foreground"
              >
                Отмена
              </button>
              <button
                onClick={handleSave}
                disabled={!name.trim()}
                className="flex h-9 items-center gap-2 rounded-lg px-5 text-[0.8rem] font-medium text-white transition-all active:scale-[0.97] disabled:opacity-50"
                style={{ background: "linear-gradient(135deg, #7c3aed, #6d28d9)", boxShadow: "0 4px 16px -2px rgba(124,58,237,0.25)" }}
              >
                {(createTeam.isPending || updateTeam.isPending) && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                {editing ? "Сохранить" : "Создать"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {deleteDialogOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setDeleteDialogOpen(false)}>
          <div
            className="w-full max-w-sm rounded-2xl border border-red-500/15 bg-card p-6 space-y-4"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-500/10">
                <AlertTriangle className="h-5 w-5 text-red-400" />
              </div>
              <div>
                <h3 className="text-[0.9rem] font-semibold text-foreground">Удалить команду?</h3>
                <p className="text-[0.75rem] text-muted-foreground">Коллекции команды станут личными</p>
              </div>
            </div>
            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setDeleteDialogOpen(false)}
                className="flex h-9 items-center rounded-lg border border-border bg-card px-4 text-[0.8rem] text-muted-foreground transition-all hover:text-foreground"
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
