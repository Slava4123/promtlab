import { useState } from "react"
import { toast } from "sonner"

import { PageLayout } from "@/components/layout/page-layout"
import { Button } from "@/components/ui/button"
import { MergeModal } from "@/components/prompts/insights/merge-modal"
import { useDuplicates, useMergePrompts } from "@/hooks/use-prompt-insights"
import type { DuplicatePair } from "@/api/prompt-insights"

// Страница "Возможные дубликаты" — пары похожих промптов (pg_trgm similarity).
// Клик "Объединить" → MergeModal с выбором какой оставить. Бэк (MergePrompts)
// переносит usage stats / теги / коллекции / share-ссылки на keepID, mergeID
// уходит в trash (восстановимо 30 дней). Подробности — usecases/prompt_insights.
export default function DuplicatesPage() {
  const { data, isLoading, isError } = useDuplicates()
  const merge = useMergePrompts()
  const [activePair, setActivePair] = useState<DuplicatePair | null>(null)

  const handleMerge = (args: { keepID: number; mergeID: number }) => {
    merge.mutate(args, {
      onSuccess: () => {
        toast.success("Дубликаты объединены")
        setActivePair(null)
      },
      onError: (e) =>
        toast.error(e instanceof Error ? e.message : "Ошибка объединения"),
    })
  }

  return (
    <PageLayout
      title="Возможные дубликаты"
      description="Похожие промпты — объедините, чтобы держать библиотеку чистой."
      maxWidth="md"
    >
      {isLoading && <p className="text-sm text-muted-foreground">Загружаем…</p>}
      {isError && (
        <p className="text-sm text-destructive">Не удалось загрузить список.</p>
      )}
      {!isLoading && data && data.length === 0 && (
        <p className="text-sm text-muted-foreground">Дубликатов не нашлось.</p>
      )}

      {data && data.length > 0 && (
        <ul className="space-y-2">
          {data.map((pair, i) => (
            <li
              key={`${pair.prompt_a.prompt_id}-${pair.prompt_b.prompt_id}-${i}`}
              className="rounded-md border px-3 py-2"
            >
              <div className="flex items-center justify-between gap-3">
                <div className="min-w-0 flex-1 space-y-0.5">
                  <p className="truncate text-sm font-medium">
                    {pair.prompt_a.title} ↔ {pair.prompt_b.title}
                  </p>
                  <p className="text-xs text-muted-foreground tabular-nums">
                    Сходство {Math.round(pair.similarity * 100)}%
                  </p>
                </div>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => setActivePair(pair)}
                >
                  Объединить
                </Button>
              </div>
            </li>
          ))}
        </ul>
      )}

      {activePair && (
        <MergeModal
          pair={activePair}
          open
          onClose={() => setActivePair(null)}
          onMerge={handleMerge}
        />
      )}
    </PageLayout>
  )
}
