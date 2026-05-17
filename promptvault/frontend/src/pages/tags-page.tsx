import { useState } from "react"
import { useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import { PageLayout } from "@/components/layout/page-layout"
import { Button } from "@/components/ui/button"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { useTags, useDeleteTag } from "@/hooks/use-tags"
import { useOrphanTags } from "@/hooks/use-orphan-tags"
import type { Tag } from "@/api/types"

// /tags — минимальная страница управления тегами.
// Overlay `?filter=orphan` показывает теги без активных промптов
// (бэкенд B10: GET /api/tags/orphan). Это не отдельная вкладка/route —
// просто другой источник данных + поясняющий текст, чтобы юзер из
// Smart Insights мог сюда «провалиться» и сразу почистить мусор.
//
// Удаление через ConfirmDialog (паттерн проекта, см. unused.tsx и trash.tsx).
// Окрашенный кружок — мини-визуальная привязка к color из БД, чтобы строки
// не выглядели одинаково (теги обычно цветные в фильтрах промптов).

type Deleting = { id: number; name: string } | null

export default function TagsPage() {
  const [params] = useSearchParams()
  const isOrphan = params.get("filter") === "orphan"

  // teamId=null → личное пространство. Когда команда выбрана, юзер видит
  // её теги. /tags/orphan на бэке аналогично режется по team_id из ctx.
  const all = useTags(null)
  const orphan = useOrphanTags()
  const del = useDeleteTag()

  const [deleting, setDeleting] = useState<Deleting>(null)

  const items: Tag[] | undefined = isOrphan ? orphan.data : all.data
  const isLoading = isOrphan ? orphan.isLoading : all.isLoading
  const isError = isOrphan ? orphan.isError : all.isError

  const handleConfirmDelete = () => {
    if (!deleting) return
    del.mutate(deleting.id, {
      onSuccess: () => {
        toast.success("Тег удалён")
        setDeleting(null)
      },
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка удаления"),
    })
  }

  return (
    <PageLayout
      title={isOrphan ? "Теги без активных промптов" : "Теги"}
      description={
        isOrphan
          ? "Эти теги не привязаны ни к одному активному промпту — можно удалить."
          : "Метки для группировки промптов. Применяются в редакторе и фильтрах."
      }
      maxWidth="md"
    >
      {isLoading && <p className="text-sm text-muted-foreground">Загружаем…</p>}
      {isError && <p className="text-sm text-destructive">Не удалось загрузить теги.</p>}

      {!isLoading && items && items.length === 0 && (
        <p className="text-sm text-muted-foreground">
          {isOrphan
            ? "Нет «orphan»-тегов — все теги используются."
            : "Тегов пока нет. Создайте теги через редактор промпта."}
        </p>
      )}

      {items && items.length > 0 && (
        <ul className="space-y-2">
          {items.map((t) => (
            <li
              key={t.id}
              className="flex items-center justify-between gap-3 rounded-md border border-border bg-card px-3 py-2"
            >
              <span className="flex items-center gap-2 text-sm">
                <span
                  aria-hidden
                  className="inline-block h-2.5 w-2.5 rounded-full"
                  style={{ background: t.color || "#8b5cf6" }}
                />
                {t.name}
              </span>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setDeleting({ id: t.id, name: t.name })}
              >
                Удалить
              </Button>
            </li>
          ))}
        </ul>
      )}

      <ConfirmDialog
        open={!!deleting}
        onOpenChange={(open) => {
          if (!open) setDeleting(null)
        }}
        title="Удалить тег?"
        description={
          deleting
            ? `Тег «${deleting.name}» будет удалён. Это действие необратимо.`
            : ""
        }
        variant="destructive"
        confirmLabel="Удалить"
        onConfirm={handleConfirmDelete}
        isPending={del.isPending}
      />
    </PageLayout>
  )
}
