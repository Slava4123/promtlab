import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { Plus, Edit2, Trash2, ArrowLeft, FolderOpen } from "lucide-react"
import { ListSkeleton } from "../components/list-skeleton"
import { Button } from "../components/ui/button"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { ConfirmDialog } from "../components/ui/confirm-dialog"
import { useToast } from "../components/ui/toaster"
import {
  useCollections,
  useCreateCollection,
  useUpdateCollection,
  useDeleteCollection,
} from "../hooks/use-collections-crud"
import { useWorkspaceStore } from "../stores/workspace-store"
import { cn } from "../lib/utils"
import type { CollectionDTO } from "../lib/types"
import {
  COLLECTION_ICON_OPTIONS,
  CollectionIcon,
} from "../lib/collection-icons"

const COLORS = [
  "#a78bfa",
  "#60a5fa",
  "#22d3ee",
  "#34d399",
  "#fbbf24",
  "#fb923c",
  "#f87171",
  "#f472b6",
]

interface FormState {
  id?: number
  name: string
  description: string
  color: string
  icon: string
}

const emptyForm: FormState = {
  name: "",
  description: "",
  color: COLORS[0],
  icon: COLLECTION_ICON_OPTIONS[0].value,
}

export function CollectionsPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const team = useWorkspaceStore((s) => s.team)
  const collectionsQuery = useCollections()
  const createMut = useCreateCollection()
  const updateMut = useUpdateCollection()
  const deleteMut = useDeleteCollection()
  const [form, setForm] = useState<FormState | null>(null)
  const [deleteId, setDeleteId] = useState<number | null>(null)

  async function handleSave() {
    if (!form || !form.name.trim()) {
      toast({ title: "Введите название", variant: "error" })
      return
    }
    try {
      if (form.id) {
        await updateMut.mutateAsync({
          id: form.id,
          body: {
            name: form.name.trim(),
            description: form.description.trim(),
            color: form.color,
            icon: form.icon,
          },
        })
        toast({ title: "Коллекция обновлена", variant: "success" })
      } else {
        await createMut.mutateAsync({
          name: form.name.trim(),
          description: form.description.trim(),
          color: form.color,
          icon: form.icon,
          team_id: team?.teamId ?? null,
        })
        toast({ title: "Коллекция создана", variant: "success" })
      }
      setForm(null)
    } catch (err) {
      toast({
        title: "Не удалось сохранить",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  async function handleDelete() {
    if (deleteId === null) return
    try {
      await deleteMut.mutateAsync(deleteId)
      toast({ title: "Удалено", description: "Можно вернуть из корзины", variant: "info" })
    } catch {
      toast({ title: "Не удалось удалить", variant: "error" })
    } finally {
      setDeleteId(null)
    }
  }

  function openEdit(c: CollectionDTO) {
    setForm({
      id: c.id,
      name: c.name,
      description: "",
      color: c.color ?? COLORS[0],
      icon: c.icon ?? COLLECTION_ICON_OPTIONS[0].value,
    })
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Коллекции</h2>
        <Button type="button" variant="brand" size="sm" onClick={() => setForm(emptyForm)} className="gap-1.5">
          <Plus className="h-3.5 w-3.5" />
          Создать
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-3">
        {collectionsQuery.isPending ? (
          <ListSkeleton count={4} showSubtitle showBadge />
        ) : (collectionsQuery.data ?? []).length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <FolderOpen className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Коллекций пока нет</p>
            <p className="text-[10px] text-(--color-muted-foreground)">
              Создайте первую коллекцию чтобы группировать промпты
            </p>
          </div>
        ) : (
          <ul className="grid grid-cols-2 gap-2">
            {(collectionsQuery.data ?? []).map((c) => (
              <li
                key={c.id}
                className="group relative flex flex-col gap-1.5 rounded-md border border-(--color-border) bg-(--color-card) p-2.5"
              >
                <button
                  type="button"
                  onClick={() => navigate(`/collections/${c.id}`)}
                  className="text-left"
                >
                  <div className="flex items-center gap-2">
                    <span
                      className="flex h-7 w-7 items-center justify-center rounded-md"
                      style={{ backgroundColor: `${c.color ?? COLORS[0]}22` }}
                    >
                      <CollectionIcon icon={c.icon} color={c.color} size={16} />
                    </span>
                    <span className="flex-1 truncate text-xs font-medium">{c.name}</span>
                  </div>
                  <p className="mt-0.5 text-[10px] text-(--color-muted-foreground)">
                    {c.prompts_count ?? 0} промптов
                  </p>
                </button>
                <div className="absolute right-1 top-1 flex gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
                  <button
                    type="button"
                    onClick={() => openEdit(c)}
                    className="rounded p-1 text-(--color-muted-foreground) hover:bg-(--color-muted) hover:text-(--color-foreground)"
                    aria-label="Редактировать"
                  >
                    <Edit2 className="h-3 w-3" />
                  </button>
                  <button
                    type="button"
                    onClick={() => setDeleteId(c.id)}
                    className="rounded p-1 text-(--color-muted-foreground) hover:bg-(--color-muted) hover:text-(--color-destructive)"
                    aria-label="Удалить"
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Edit/create dialog */}
      {form && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={() => setForm(null)} />
          <div className="relative w-full max-w-sm rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
            <h3 className="mb-3 text-sm font-semibold">
              {form.id ? "Редактировать" : "Новая коллекция"}
            </h3>
            <div className="space-y-3">
              <div className="space-y-1">
                <Label htmlFor="col-name">Название</Label>
                <Input
                  id="col-name"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="Например: Рабочие промпты"
                  autoFocus
                />
              </div>
              <div className="space-y-1">
                <Label>Иконка</Label>
                <div className="grid grid-cols-8 gap-1">
                  {COLLECTION_ICON_OPTIONS.map((opt) => {
                    const Icon = opt.Icon
                    return (
                      <button
                        key={opt.value}
                        type="button"
                        onClick={() => setForm({ ...form, icon: opt.value })}
                        title={opt.label}
                        className={cn(
                          "flex h-8 w-8 items-center justify-center rounded transition-colors",
                          form.icon === opt.value
                            ? "ring-2 ring-(--color-primary) bg-(--color-muted)/60"
                            : "bg-(--color-muted)/30 hover:bg-(--color-muted)",
                        )}
                      >
                        <Icon
                          width={14}
                          height={14}
                          style={form.icon === opt.value ? { color: form.color } : undefined}
                        />
                      </button>
                    )
                  })}
                </div>
              </div>
              <div className="space-y-1">
                <Label>Цвет</Label>
                <div className="flex gap-1.5">
                  {COLORS.map((color) => (
                    <button
                      key={color}
                      type="button"
                      onClick={() => setForm({ ...form, color })}
                      className={cn(
                        "h-6 w-6 rounded-full transition-transform",
                        form.color === color &&
                          "ring-2 ring-offset-2 ring-offset-(--color-background) scale-110",
                      )}
                      style={{
                        backgroundColor: color,
                        boxShadow: form.color === color ? `0 0 0 2px ${color}` : undefined,
                      }}
                      aria-label={color}
                    />
                  ))}
                </div>
              </div>
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <Button type="button" variant="outline" size="sm" onClick={() => setForm(null)}>
                Отмена
              </Button>
              <Button
                type="button"
                variant="brand"
                size="sm"
                onClick={handleSave}
                disabled={createMut.isPending || updateMut.isPending}
              >
                Сохранить
              </Button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deleteId !== null}
        title="Удалить коллекцию?"
        description="Промпты останутся, но без этой коллекции. Можно вернуть из корзины 30 дней."
        confirmLabel="Удалить"
        variant="destructive"
        onConfirm={handleDelete}
        onClose={() => setDeleteId(null)}
      />
    </div>
  )
}
