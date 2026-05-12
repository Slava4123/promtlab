import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { ArrowLeft, Plus, Tag as TagIcon, X, Loader2 } from "lucide-react"
import { Button } from "../components/ui/button"
import { Input } from "../components/ui/input"
import { ConfirmDialog } from "../components/ui/confirm-dialog"
import { useToast } from "../components/ui/toaster"
import { useTags, useCreateTag, useDeleteTag } from "../hooks/use-tags-crud"
import { useWorkspaceStore } from "../stores/workspace-store"
import { cn } from "../lib/utils"

const TAG_COLORS = [
  "#8b5cf6",
  "#3b82f6",
  "#06b6d4",
  "#10b981",
  "#84cc16",
  "#eab308",
  "#f59e0b",
  "#ef4444",
  "#ec4899",
  "#a855f7",
]

export function TagsPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const team = useWorkspaceStore((s) => s.team)
  const tagsQuery = useTags()
  const createMut = useCreateTag()
  const deleteMut = useDeleteTag()
  const [name, setName] = useState("")
  const [color, setColor] = useState(TAG_COLORS[0])
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function handleCreate() {
    const trimmed = name.trim()
    if (!trimmed) {
      toast({ title: "Введите имя тега", variant: "error" })
      return
    }
    try {
      await createMut.mutateAsync({
        name: trimmed,
        color,
        team_id: team?.teamId ?? null,
      })
      setName("")
      toast({ title: "Тег создан", variant: "success" })
    } catch (err) {
      toast({
        title: "Не удалось создать",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  async function handleDelete() {
    if (deleteId === null) return
    try {
      await deleteMut.mutateAsync(deleteId)
      toast({ title: "Тег удалён", variant: "info" })
    } catch {
      toast({ title: "Не удалось удалить", variant: "error" })
    } finally {
      setDeleteId(null)
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Теги</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        {/* Create */}
        <div className="rounded-md border border-(--color-border) bg-(--color-card) p-2 space-y-2">
          <div className="flex gap-1.5">
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Имя тега"
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault()
                  handleCreate()
                }
              }}
            />
            <Button
              type="button"
              size="sm"
              onClick={handleCreate}
              disabled={createMut.isPending || !name.trim()}
              className="gap-1"
            >
              <Plus className="h-3.5 w-3.5" />
              Создать
            </Button>
          </div>
          <div className="flex gap-1.5">
            {TAG_COLORS.map((c) => (
              <button
                key={c}
                type="button"
                onClick={() => setColor(c)}
                className={cn(
                  "h-5 w-5 rounded-full transition-transform",
                  color === c && "ring-2 ring-offset-1 ring-offset-(--color-background) scale-110",
                )}
                style={{ backgroundColor: c }}
                aria-label={c}
              />
            ))}
          </div>
        </div>

        {/* List */}
        {tagsQuery.isPending ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : (tagsQuery.data ?? []).length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <TagIcon className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Тегов пока нет</p>
            <p className="text-[10px] text-(--color-muted-foreground)">
              Создайте теги чтобы быстро находить промпты
            </p>
          </div>
        ) : (
          <ul className="flex flex-wrap gap-1.5">
            {(tagsQuery.data ?? []).map((t) => (
              <li
                key={t.id}
                className="group inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs"
                style={{
                  backgroundColor: `${t.color}22`,
                  color: t.color,
                }}
              >
                <span>{t.name}</span>
                <button
                  type="button"
                  onClick={() => setDeleteId(t.id)}
                  className="opacity-0 transition-opacity group-hover:opacity-100 hover:text-(--color-destructive)"
                  aria-label="Удалить"
                >
                  <X className="h-3 w-3" />
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      <ConfirmDialog
        open={deleteId !== null}
        title="Удалить тег?"
        description="Тег будет убран со всех промптов."
        confirmLabel="Удалить"
        variant="destructive"
        onConfirm={handleDelete}
        onClose={() => setDeleteId(null)}
      />
    </div>
  )
}
