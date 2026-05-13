import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Plus, Users, ArrowLeft } from "lucide-react"
import { ListSkeleton } from "../../components/list-skeleton"
import { Button } from "../../components/ui/button"
import { Input } from "../../components/ui/input"
import { Label } from "../../components/ui/label"
import { Textarea } from "../../components/ui/textarea"
import { useToast } from "../../components/ui/toaster"
import { useTeams, useCreateTeam } from "../../hooks/use-teams-crud"
import { useWorkspaceStore } from "../../stores/workspace-store"
import type { TeamRole } from "../../lib/types"

// Перевод роли для badge на карточке команды. До этого выводили raw enum-значение
// от backend (OWNER/EDITOR/VIEWER) — это англицизм. Mirror ROLE_META в detail-page.
const ROLE_LABELS: Record<TeamRole, string> = {
  owner: "Владелец",
  editor: "Редактор",
  viewer: "Просмотр",
}

export function TeamsIndexPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const teamsQuery = useTeams()
  const createMut = useCreateTeam()
  const setTeam = useWorkspaceStore((s) => s.setTeam)
  const currentTeam = useWorkspaceStore((s) => s.team)
  const [createOpen, setCreateOpen] = useState(false)
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")

  async function handleCreate() {
    if (!name.trim()) {
      toast({ title: "Введите название", variant: "error" })
      return
    }
    try {
      const team = await createMut.mutateAsync({
        name: name.trim(),
        description: description.trim(),
      })
      toast({ title: "Команда создана", variant: "success" })
      setCreateOpen(false)
      setName("")
      setDescription("")
      setTeam(team.slug, team.id, team.name)
      navigate(`/teams/${team.slug}`)
    } catch (err) {
      toast({
        title: "Не удалось создать",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  if (teamsQuery.isPending) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
          <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <h2 className="flex-1 text-sm font-semibold">Команды</h2>
        </div>
        <div className="flex-1 overflow-y-auto p-3">
          <ListSkeleton count={3} showSubtitle showBadge />
        </div>
      </div>
    )
  }

  const teams = teamsQuery.data ?? []

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Команды</h2>
        <Button type="button" variant="brand" size="sm" onClick={() => setCreateOpen(true)} className="gap-1.5">
          <Plus className="h-3.5 w-3.5" />
          Создать
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {/* Personal workspace */}
        <button
          type="button"
          onClick={() => {
            useWorkspaceStore.getState().clearTeam()
            navigate("/")
          }}
          className={
            "flex w-full items-center gap-2 rounded-md border p-2.5 text-left transition-colors " +
            (currentTeam === null
              ? "border-(--color-brand) bg-(--color-brand-muted)"
              : "border-(--color-border) bg-(--color-card) hover:bg-(--color-muted)/40")
          }
        >
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-(--color-muted) text-sm font-semibold">
            Я
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-xs font-medium">Личное</div>
            <div className="text-[10px] text-(--color-muted-foreground)">Ваши промпты</div>
          </div>
          {currentTeam === null && <span className="text-[10px] text-(--color-brand)">текущее</span>}
        </button>

        {teams.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <Users className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Команд пока нет</p>
            <p className="text-[10px] text-(--color-muted-foreground)">
              Создайте команду чтобы делиться промптами с коллегами.
            </p>
          </div>
        ) : (
          teams.map((team) => (
            <button
              key={team.id}
              type="button"
              onClick={() => {
                setTeam(team.slug, team.id, team.name)
                navigate(`/teams/${team.slug}`)
              }}
              className={
                "flex w-full items-center gap-2 rounded-md border p-2.5 text-left transition-colors " +
                (currentTeam?.teamId === team.id
                  ? "border-(--color-brand) bg-(--color-brand-muted)"
                  : "border-(--color-border) bg-(--color-card) hover:bg-(--color-muted)/40")
              }
            >
              <div className="flex h-8 w-8 items-center justify-center rounded-md bg-(--color-brand-muted) text-sm font-semibold text-(--color-brand)">
                {team.name.charAt(0).toUpperCase()}
              </div>
              <div className="flex-1 min-w-0">
                <div className="truncate text-xs font-medium">{team.name}</div>
                {team.description && (
                  <div className="truncate text-[10px] text-(--color-muted-foreground)">
                    {team.description}
                  </div>
                )}
              </div>
              {team.role && (
                <span className="rounded bg-(--color-muted) px-1.5 py-0.5 text-[9px] uppercase tracking-wide">
                  {ROLE_LABELS[team.role as TeamRole] ?? team.role}
                </span>
              )}
            </button>
          ))
        )}
      </div>

      {/* Create dialog */}
      {createOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div
            className="absolute inset-0 bg-black/50 backdrop-blur-sm"
            onClick={() => setCreateOpen(false)}
          />
          <div className="relative w-full max-w-sm rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
            <h3 className="mb-3 text-sm font-semibold">Новая команда</h3>
            <div className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="team-name">Название</Label>
                <Input
                  id="team-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Например: AI Squad"
                  autoFocus
                />
              </div>
              <div className="space-y-1">
                <Label htmlFor="team-desc">Описание (опционально)</Label>
                <Textarea
                  id="team-desc"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  rows={2}
                  placeholder="О чём команда работает"
                />
              </div>
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <Button type="button" variant="outline" size="sm" onClick={() => setCreateOpen(false)}>
                Отмена
              </Button>
              <Button type="button" variant="brand" size="sm" onClick={handleCreate} disabled={createMut.isPending}>
                Создать
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
