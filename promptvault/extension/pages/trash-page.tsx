import { useState } from "react"
import { useNavigate } from "react-router-dom"
import {
  ArrowLeft,
  Loader2,
  RotateCcw,
  Trash2,
  Tag as TagIcon,
  FolderOpen,
  FileText,
} from "lucide-react"
import { Button } from "../components/ui/button"
import { ConfirmDialog } from "../components/ui/confirm-dialog"
import { useToast } from "../components/ui/toaster"
import {
  useTrash,
  useRestoreTrashPrompt,
  useRestoreTrashCollection,
  usePermanentDeletePrompt,
  usePermanentDeleteCollection,
  useEmptyTrash,
} from "../hooks/use-trash"
import { formatRelativeDate } from "@pv/shared/utils/format-date"
import { pluralAfterDo } from "@pv/shared/utils/plural"

export function TrashPage() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const trashQuery = useTrash()
  const restorePrompt = useRestoreTrashPrompt()
  const restoreCol = useRestoreTrashCollection()
  const deletePrompt = usePermanentDeletePrompt()
  const deleteCol = usePermanentDeleteCollection()
  const emptyMut = useEmptyTrash()
  const [emptyOpen, setEmptyOpen] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState<{ type: "prompt" | "collection"; id: number } | null>(null)

  if (trashQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  const trash = trashQuery.data
  const prompts = trash?.prompts.items ?? []
  const collections = trash?.collections ?? []
  const totalItems = prompts.length + collections.length

  async function handleRestorePrompt(id: number) {
    try {
      await restorePrompt.mutateAsync(id)
      toast({ title: "Восстановлено", variant: "success" })
    } catch {
      toast({ title: "Не удалось восстановить", variant: "error" })
    }
  }
  async function handleRestoreCollection(id: number) {
    try {
      await restoreCol.mutateAsync(id)
      toast({ title: "Коллекция восстановлена", variant: "success" })
    } catch {
      toast({ title: "Не удалось восстановить", variant: "error" })
    }
  }

  async function handlePermanentDelete() {
    if (!confirmDelete) return
    try {
      if (confirmDelete.type === "prompt") {
        await deletePrompt.mutateAsync(confirmDelete.id)
      } else {
        await deleteCol.mutateAsync(confirmDelete.id)
      }
      toast({ title: "Удалено навсегда", variant: "info" })
    } catch {
      toast({ title: "Не удалось удалить", variant: "error" })
    } finally {
      setConfirmDelete(null)
    }
  }

  async function handleEmpty() {
    try {
      await emptyMut.mutateAsync()
      toast({ title: "Корзина очищена", variant: "info" })
      setEmptyOpen(false)
    } catch {
      toast({ title: "Не удалось очистить", variant: "error" })
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Корзина</h2>
        {totalItems > 0 && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => setEmptyOpen(true)}
            className="text-(--color-destructive) gap-1.5"
          >
            <Trash2 className="h-3.5 w-3.5" />
            Очистить
          </Button>
        )}
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        <p className="text-[10px] text-(--color-muted-foreground)">
          Удалённые элементы хранятся 30 дней. Потом удаляются автоматически.
        </p>

        {prompts.length > 0 && (
          <section>
            <h3 className="mb-2 flex items-center gap-1.5 text-xs font-semibold text-(--color-muted-foreground)">
              <FileText className="h-3.5 w-3.5" />
              Промпты ({prompts.length})
            </h3>
            <ul className="space-y-1.5">
              {prompts.map((p) => (
                <li
                  key={p.id}
                  className="rounded-md border border-(--color-border) bg-(--color-card) p-2 text-xs"
                >
                  <div className="font-medium truncate">{p.title}</div>
                  <div className="mt-0.5 flex items-center gap-1 text-[10px] text-(--color-muted-foreground)">
                    <span>Удалён {formatRelativeDate(p.deleted_at)}</span>
                    <span>•</span>
                    <span>Осталось {pluralAfterDo(p.days_left, "день", "дня", "дней")}</span>
                  </div>
                  <div className="mt-1.5 flex gap-1">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleRestorePrompt(p.id)}
                      className="h-6 text-[10px] px-2 gap-1"
                    >
                      <RotateCcw className="h-3 w-3" />
                      Вернуть
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => setConfirmDelete({ type: "prompt", id: p.id })}
                      className="h-6 text-[10px] px-2 text-(--color-destructive) gap-1"
                    >
                      <Trash2 className="h-3 w-3" />
                      Удалить навсегда
                    </Button>
                  </div>
                </li>
              ))}
            </ul>
          </section>
        )}

        {collections.length > 0 && (
          <section>
            <h3 className="mb-2 flex items-center gap-1.5 text-xs font-semibold text-(--color-muted-foreground)">
              <FolderOpen className="h-3.5 w-3.5" />
              Коллекции ({collections.length})
            </h3>
            <ul className="space-y-1.5">
              {collections.map((c) => (
                <li
                  key={c.id}
                  className="rounded-md border border-(--color-border) bg-(--color-card) p-2 text-xs"
                >
                  <div className="font-medium truncate">{c.name}</div>
                  <div className="mt-0.5 text-[10px] text-(--color-muted-foreground)">
                    Осталось {pluralAfterDo(c.days_left, "день", "дня", "дней")}
                  </div>
                  <div className="mt-1.5 flex gap-1">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleRestoreCollection(c.id)}
                      className="h-6 text-[10px] px-2 gap-1"
                    >
                      <RotateCcw className="h-3 w-3" />
                      Вернуть
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => setConfirmDelete({ type: "collection", id: c.id })}
                      className="h-6 text-[10px] px-2 text-(--color-destructive) gap-1"
                    >
                      <Trash2 className="h-3 w-3" />
                      Удалить навсегда
                    </Button>
                  </div>
                </li>
              ))}
            </ul>
          </section>
        )}

        {totalItems === 0 && (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <TagIcon className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Корзина пуста</p>
            <p className="text-[10px] text-(--color-muted-foreground)">
              Удалённые промпты и коллекции попадают сюда на 30 дней.
            </p>
          </div>
        )}
      </div>

      <ConfirmDialog
        open={emptyOpen}
        title="Очистить корзину?"
        description={`Будут навсегда удалены ${totalItems} элементов. Действие необратимо.`}
        confirmLabel="Очистить всё"
        variant="destructive"
        onConfirm={handleEmpty}
        onClose={() => setEmptyOpen(false)}
      />
      <ConfirmDialog
        open={confirmDelete !== null}
        title="Удалить навсегда?"
        description="Действие необратимо. Элемент исчезнет навсегда."
        confirmLabel="Удалить"
        variant="destructive"
        onConfirm={handlePermanentDelete}
        onClose={() => setConfirmDelete(null)}
      />
    </div>
  )
}
