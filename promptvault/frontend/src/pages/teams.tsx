import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Plus, Users, Pencil, Trash2, Loader2 } from "lucide-react"
import { EmptyState } from "@/components/ui/empty-state"
import { toast } from "sonner"

import { useTeams, useCreateTeam, useUpdateTeam, useDeleteTeam } from "@/hooks/use-teams"
import { PageLayout } from "@/components/layout/page-layout"
import { RoleBadge } from "@/components/teams/role-badge"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
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
    <PageLayout
      title="Команды"
      description="Совместная работа над промптами"
      action={
        <Button variant="brand" size="sm" onClick={openCreate}>
          <Plus className="h-3.5 w-3.5" />
          Новая команда
        </Button>
      }
    >
      {/* List */}
      {isLoading ? (
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="rounded-xl border border-border bg-card p-5">
              <div className="mb-3 h-9 w-9 animate-pulse rounded-lg bg-muted/40" />
              <div className="mb-2 h-4 w-2/3 animate-pulse rounded-md bg-muted/40" />
              <div className="h-3 w-1/2 animate-pulse rounded-md bg-muted/30" />
            </div>
          ))}
        </div>
      ) : !teams || teams.length === 0 ? (
        <EmptyState
          icon={<Users className="h-7 w-7 text-brand-muted-foreground/60" />}
          title="Пока нет команд"
          description="В команде все промпты, коллекции и теги общие: меняет один — видят все. Есть роли owner/editor/viewer."
          action={
            <Button variant="brand" size="sm" onClick={openCreate}>
              <Plus className="h-3.5 w-3.5" />
              Создать команду
            </Button>
          }
        />
      ) : (
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {teams.map((t) => (
            <div
              key={t.id}
              className="group cursor-pointer rounded-xl border border-border bg-card p-5 transition-[transform,box-shadow,border-color] duration-200 hover:-translate-y-0.5 hover:border-brand/20 hover:shadow-[0_8px_32px_-8px_rgba(0,0,0,0.5),0_0_0_1px_rgba(139,92,246,0.1)]"
              onClick={() => navigate(`/teams/${t.slug}`)}
            >
              <div className="mb-3 flex items-start justify-between">
                <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-brand-muted ring-1 ring-inset ring-brand/15">
                  <Users className="h-4 w-4 text-brand-muted-foreground" />
                </div>
                <div className="flex items-center gap-1">
                  <RoleBadge role={t.role} />
                  {t.role === "owner" && (
                    <div className="flex gap-1 sm:opacity-0 sm:transition-opacity sm:group-hover:opacity-100">
                      <button
                        className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
                        onClick={(e) => { e.stopPropagation(); openEdit(t) }}
                        aria-label="Редактировать команду"
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button
                        className="rounded-md p-1 text-muted-foreground hover:bg-red-500/10 hover:text-red-400"
                        onClick={(e) => { e.stopPropagation(); confirmDelete(t.slug) }}
                        aria-label="Удалить команду"
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
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{editing ? "Редактировать команду" : "Новая команда"}</DialogTitle>
          </DialogHeader>

          <div className="space-y-2">
            <label className="text-[0.8rem] font-medium text-foreground">Название</label>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Например: Backend-разработка"
              autoFocus
              className="flex h-11 w-full rounded-lg border border-border bg-background px-3.5 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-brand/40 focus:ring-3 focus:ring-brand/10"
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
              className="flex w-full resize-none rounded-lg border border-border bg-background px-3.5 py-2.5 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-brand/40 focus:ring-3 focus:ring-brand/10"
            />
          </div>

          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setDialogOpen(false)}>
              Отмена
            </Button>
            <Button
              variant="brand"
              size="sm"
              onClick={handleSave}
              disabled={!name.trim()}
            >
              {(createTeam.isPending || updateTeam.isPending) && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
              {editing ? "Сохранить" : "Создать"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Удалить команду?"
        description="Коллекции команды станут личными"
        variant="destructive"
        confirmLabel="Удалить"
        onConfirm={handleDelete}
        isPending={deleteTeam.isPending}
      />
    </PageLayout>
  )
}
