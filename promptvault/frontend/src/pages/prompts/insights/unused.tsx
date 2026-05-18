import { useState } from "react"
import { Link } from "react-router-dom"
import { toast } from "sonner"

import { InsightPromptRow } from "@/components/prompts/insights/insight-prompt-row"
import { PageLayout } from "@/components/layout/page-layout"
import { Button } from "@/components/ui/button"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { useUnusedPrompts } from "@/hooks/use-prompt-insights"
import { useDeletePrompt } from "@/hooks/use-prompts"

// Страница "Забытые промпты" — список промптов без использования 30+ дней.
// Per-row actions: Открыть (Link к редактору) + Удалить (soft-delete в корзину
// с undo-toast). Подтверждение через ConfirmDialog (паттерн проекта,
// см. trash.tsx) — НЕ window.confirm(). Удаление мягкое (30 дней в корзине),
// поэтому "Удалить навсегда" не используется.

type Deleting = { id: number; title: string } | null

export default function UnusedInsightsPage() {
  const { data, isLoading, isError } = useUnusedPrompts()
  const deletePrompt = useDeletePrompt()
  const [deleting, setDeleting] = useState<Deleting>(null)

  const handleConfirmDelete = () => {
    if (!deleting) return
    deletePrompt.mutate(deleting.id, {
      onSuccess: () => {
        toast.success("Промпт перемещён в корзину")
        setDeleting(null)
      },
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка удаления"),
    })
  }

  return (
    <PageLayout
      title="Забытые промпты"
      description="Промпты без использования 30+ дней. Подумайте о том, чтобы удалить или обновить."
      maxWidth="md"
    >
      {isLoading && <p className="text-sm text-muted-foreground">Загружаем…</p>}
      {isError && (
        <p className="text-sm text-destructive">Не удалось загрузить список.</p>
      )}
      {!isLoading && data && data.length === 0 && (
        <p className="text-sm text-muted-foreground">
          Нет забытых промптов — всё используется.
        </p>
      )}

      {data && data.length > 0 && (
        <ul className="space-y-2">
          {data.map((p) => (
            <li key={p.prompt_id}>
              <InsightPromptRow
                promptID={p.prompt_id}
                title={p.title}
                uses={p.uses}
                showUses={false}
                actions={
                  <>
                    <Button asChild variant="ghost" size="sm">
                      <Link to={`/prompts/${p.prompt_id}`}>Открыть</Link>
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setDeleting({ id: p.prompt_id, title: p.title })}
                    >
                      Удалить
                    </Button>
                  </>
                }
              />
            </li>
          ))}
        </ul>
      )}

      <ConfirmDialog
        open={!!deleting}
        onOpenChange={(open) => {
          if (!open) setDeleting(null)
        }}
        title="Удалить промпт?"
        description={
          deleting
            ? `«${deleting.title}» переместится в корзину и будет храниться 30 дней.`
            : ""
        }
        variant="destructive"
        confirmLabel="Удалить"
        onConfirm={handleConfirmDelete}
        isPending={deletePrompt.isPending}
      />
    </PageLayout>
  )
}
