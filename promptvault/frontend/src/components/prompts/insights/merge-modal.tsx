import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { pluralizeRu } from "@/lib/pluralize"
import type { DuplicatePair } from "@/api/prompt-insights"

// MergeModal — диалог выбора, какой промпт оставить при объединении дубликатов.
// Юзер видит обе версии бок о бок (title + счётчик использований) и кликает
// «Оставить ‹title›» на той, что хочет сохранить. Другая уходит в trash
// (восстановимо 30 дней). Теги/коллекции НЕ переносятся — поэтому warning
// в описании (см. backend usecases/prompt_insights.MergePrompts).
interface MergeModalProps {
  pair: DuplicatePair
  open: boolean
  onClose: () => void
  onMerge: (args: { keepID: number; mergeID: number }) => void
}

export function MergeModal({ pair, open, onClose, onMerge }: MergeModalProps) {
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Объединить дубликаты</DialogTitle>
          <DialogDescription>
            Похожесть {Math.round(pair.similarity * 100)}%. Выберите, какой промпт
            оставить — второй уйдёт в корзину (можно восстановить за 30 дней).
            Теги и коллекции не переносятся.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-3 sm:grid-cols-2">
          {[pair.prompt_a, pair.prompt_b].map((p, idx) => {
            const other = idx === 0 ? pair.prompt_b : pair.prompt_a
            return (
              <div key={p.prompt_id} className="space-y-2 rounded-md border p-3">
                <p className="text-sm font-medium">{p.title}</p>
                <p className="text-xs text-muted-foreground tabular-nums">
                  {p.uses}{" "}
                  {pluralizeRu(p.uses, "использование", "использования", "использований")}
                </p>
                <Button
                  size="sm"
                  className="w-full"
                  onClick={() => onMerge({ keepID: p.prompt_id, mergeID: other.prompt_id })}
                >
                  Оставить «{p.title}»
                </Button>
              </div>
            )
          })}
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            Отмена
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
