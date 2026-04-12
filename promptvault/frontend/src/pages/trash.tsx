import { useState } from "react"
import { Trash2, RotateCcw, FileText, FolderOpen } from "lucide-react"
import { EmptyState } from "@/components/ui/empty-state"
import { toast } from "sonner"

import { useTrash, useRestoreItem, usePermanentDelete, useEmptyTrash } from "@/hooks/use-trash"
import { PageLayout } from "@/components/layout/page-layout"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { Button } from "@/components/ui/button"
import type { TrashPrompt, TrashCollection } from "@/api/types"

type DeletingItem = { type: "prompt" | "collection" | "tag"; id: number; title: string } | null

export default function TrashPage() {
  const team = useWorkspaceStore((s) => s.team)
  const teamId = team?.teamId ?? null
  const { data, isLoading } = useTrash({ team_id: teamId })
  const restore = useRestoreItem()
  const permanentDelete = usePermanentDelete()
  const emptyTrash = useEmptyTrash()
  const [deleting, setDeleting] = useState<DeletingItem>(null)
  const [emptyConfirmOpen, setEmptyConfirmOpen] = useState(false)

  const handleRestore = (type: "prompt" | "collection" | "tag", id: number) => {
    restore.mutate({ type, id }, {
      onSuccess: () => toast.success("Восстановлено"),
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  const confirmDelete = (type: "prompt" | "collection" | "tag", id: number, title: string) => {
    setDeleting({ type, id, title })
  }

  const handlePermanentDelete = () => {
    if (!deleting) return
    permanentDelete.mutate({ type: deleting.type, id: deleting.id }, {
      onSuccess: () => { toast.success("Удалено навсегда"); setDeleting(null) },
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  const handleEmptyTrash = () => {
    emptyTrash.mutate(teamId, {
      onSuccess: (data) => { toast.success(`Удалено элементов: ${data.deleted}`); setEmptyConfirmOpen(false) },
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  const prompts = data?.prompts.items ?? []
  const collections = data?.collections ?? []
  const totalItems = prompts.length + collections.length

  return (
    <PageLayout
      title="Корзина"
      description="Удалённые элементы хранятся 30 дней, затем удаляются автоматически"
      action={totalItems > 0 ? (
        <Button
          variant="destructive"
          size="sm"
          onClick={() => setEmptyConfirmOpen(true)}
        >
          <Trash2 className="h-3.5 w-3.5" />
          Очистить корзину
        </Button>
      ) : undefined}
    >
      {isLoading ? (
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="rounded-xl border border-border bg-card p-4">
              <div className="mb-3 h-8 w-8 animate-pulse rounded-lg bg-foreground/[0.06]" />
              <div className="mb-2 h-4 w-2/3 animate-pulse rounded-md bg-foreground/[0.06]" />
              <div className="h-3 w-1/2 animate-pulse rounded-md bg-foreground/[0.04]" />
            </div>
          ))}
        </div>
      ) : totalItems === 0 ? (
        <EmptyState
          icon={<Trash2 className="h-7 w-7 text-muted-foreground/40" />}
          title="Корзина пуста"
          description="Удалённые промпты, коллекции и теги появятся здесь"
        />
      ) : (
        <div className="space-y-6">
          {/* Промпты */}
          {prompts.length > 0 && (
            <section>
              <p className="mb-2 px-1 text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">
                Промпты ({prompts.length})
              </p>
              <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {prompts.map((p) => (
                  <PromptTrashCard key={p.id} item={p} onRestore={handleRestore} onDelete={confirmDelete} />
                ))}
              </div>
            </section>
          )}

          {/* Коллекции */}
          {collections.length > 0 && (
            <section>
              <p className="mb-2 px-1 text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">
                Коллекции ({collections.length})
              </p>
              <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {collections.map((c) => (
                  <CollectionTrashCard key={c.id} item={c} onRestore={handleRestore} onDelete={confirmDelete} />
                ))}
              </div>
            </section>
          )}

        </div>
      )}

      {/* Delete confirmation */}
      <ConfirmDialog
        open={!!deleting}
        onOpenChange={(open) => { if (!open) setDeleting(null) }}
        title="Удалить навсегда?"
        description={deleting ? `«${deleting.title}» нельзя будет восстановить` : ""}
        variant="destructive"
        confirmLabel="Удалить навсегда"
        onConfirm={handlePermanentDelete}
        isPending={permanentDelete.isPending}
      />

      {/* Empty trash confirmation */}
      <ConfirmDialog
        open={emptyConfirmOpen}
        onOpenChange={setEmptyConfirmOpen}
        title="Очистить корзину?"
        description={`Все элементы (${totalItems}) будут удалены навсегда`}
        variant="destructive"
        confirmLabel="Очистить"
        onConfirm={handleEmptyTrash}
        isPending={emptyTrash.isPending}
      />
    </PageLayout>
  )
}

function PromptTrashCard({ item, onRestore, onDelete }: {
  item: TrashPrompt
  onRestore: (type: "prompt", id: number) => void
  onDelete: (type: "prompt", id: number, title: string) => void
}) {
  return (
    <div className="group rounded-xl border border-border bg-card p-4 opacity-70 transition-opacity hover:opacity-100">
      <div className="mb-3 flex items-center gap-2.5">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-brand-muted ring-1 ring-brand/10">
          <FileText className="h-3.5 w-3.5 text-brand-muted-foreground/60" />
        </div>
        <h3 className="min-w-0 flex-1 truncate text-[0.82rem] font-medium text-foreground">{item.title}</h3>
      </div>
      <p className="mb-3 line-clamp-2 text-[0.75rem] leading-relaxed text-muted-foreground">{item.content}</p>
      <div className="flex items-center justify-between">
        <span className="text-[10px] text-muted-foreground">
          {item.days_left > 0 ? `${item.days_left} дн. до удаления` : "Будет удалён"}
        </span>
        <div className="flex gap-1">
          <button
            onClick={() => onRestore("prompt", item.id)}
            className="flex h-7 items-center gap-1 rounded-md px-2 text-[0.72rem] text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            <RotateCcw className="h-3 w-3" />
            Восстановить
          </button>
          <button
            onClick={() => onDelete("prompt", item.id, item.title)}
            className="flex h-7 items-center gap-1 rounded-md px-2 text-[0.72rem] text-red-400/70 transition-colors hover:bg-red-500/10 hover:text-red-400"
            aria-label="Удалить навсегда"
          >
            <Trash2 className="h-3 w-3" />
          </button>
        </div>
      </div>
    </div>
  )
}

function CollectionTrashCard({ item, onRestore, onDelete }: {
  item: TrashCollection
  onRestore: (type: "collection", id: number) => void
  onDelete: (type: "collection", id: number, title: string) => void
}) {
  return (
    <div className="group rounded-xl border border-border bg-card p-4 opacity-70 transition-opacity hover:opacity-100">
      <div className="mb-3 flex items-center gap-2.5">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg ring-1 ring-border" style={{ backgroundColor: item.color + "15" }}>
          <FolderOpen className="h-3.5 w-3.5" style={{ color: item.color }} />
        </div>
        <h3 className="min-w-0 flex-1 truncate text-[0.82rem] font-medium text-foreground">{item.name}</h3>
      </div>
      {item.description && (
        <p className="mb-3 line-clamp-2 text-[0.75rem] leading-relaxed text-muted-foreground">{item.description}</p>
      )}
      <div className="flex items-center justify-between">
        <span className="text-[10px] text-muted-foreground">
          {item.days_left > 0 ? `${item.days_left} дн. до удаления` : "Будет удалён"}
        </span>
        <div className="flex gap-1">
          <button
            onClick={() => onRestore("collection", item.id)}
            className="flex h-7 items-center gap-1 rounded-md px-2 text-[0.72rem] text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            <RotateCcw className="h-3 w-3" />
            Восстановить
          </button>
          <button
            onClick={() => onDelete("collection", item.id, item.name)}
            className="flex h-7 items-center gap-1 rounded-md px-2 text-[0.72rem] text-red-400/70 transition-colors hover:bg-red-500/10 hover:text-red-400"
            aria-label="Удалить навсегда"
          >
            <Trash2 className="h-3 w-3" />
          </button>
        </div>
      </div>
    </div>
  )
}
