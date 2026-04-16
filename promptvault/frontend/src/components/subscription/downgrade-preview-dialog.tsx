import { AlertTriangle, Loader2 } from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import type { DowngradePreview } from "@/api/subscription"

interface DowngradePreviewDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  preview: DowngradePreview | undefined
  isLoading?: boolean
  isPending?: boolean
  onConfirm: () => void
}

function pluralizeRu(n: number, one: string, few: string, many: string): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return few
  return many
}

// overItems — превышения в человекочитаемом виде. Рендерятся только поля > 0.
function overItems(p: DowngradePreview): string[] {
  const out: string[] = []
  if (p.over_prompts > 0) {
    out.push(
      `${p.over_prompts} ${pluralizeRu(p.over_prompts, "промпт", "промпта", "промптов")} будут скрыты (можно восстановить при возврате на Pro)`,
    )
  }
  if (p.over_collections > 0) {
    out.push(
      `${p.over_collections} ${pluralizeRu(p.over_collections, "коллекция", "коллекции", "коллекций")} станут недоступны`,
    )
  }
  if (p.over_teams > 0) {
    out.push(
      `${p.over_teams} ${pluralizeRu(p.over_teams, "команда", "команды", "команд")} потеряете доступ к управлению`,
    )
  }
  if (p.over_shares > 0) {
    out.push(
      `${p.over_shares} ${pluralizeRu(p.over_shares, "публичная ссылка", "публичных ссылки", "публичных ссылок")} перестанут работать`,
    )
  }
  return out
}

export function DowngradePreviewDialog({
  open,
  onOpenChange,
  preview,
  isLoading = false,
  isPending = false,
  onConfirm,
}: DowngradePreviewDialogProps) {
  const items = preview ? overItems(preview) : []
  const hasOverages = items.length > 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-amber-500/15">
              <AlertTriangle className="size-5 text-amber-600 dark:text-amber-400" aria-hidden="true" />
            </div>
            <DialogTitle>Переход на Free</DialogTitle>
          </div>
          <DialogDescription>
            {isLoading
              ? "Проверяем, что изменится…"
              : hasOverages
                ? "При переходе на Free вы превысите лимиты. Вот что произойдёт:"
                : "Всё ваше содержимое помещается в лимиты Free — переход пройдёт без потерь."}
          </DialogDescription>
        </DialogHeader>

        {isLoading ? (
          <div className="flex items-center justify-center py-6">
            <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" aria-hidden="true" />
          </div>
        ) : hasOverages ? (
          <ul className="space-y-2 rounded-md border border-amber-500/30 bg-amber-50/50 p-3 text-sm text-amber-900 dark:bg-amber-900/10 dark:text-amber-200">
            {items.map((text) => (
              <li key={text} className="flex items-start gap-2">
                <span aria-hidden="true" className="mt-1.5 h-1 w-1 shrink-0 rounded-full bg-current opacity-60" />
                <span>{text}</span>
              </li>
            ))}
          </ul>
        ) : null}

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline" disabled={isPending}>
              Остаться на {preview?.current_plan_id === "free" ? "Free" : "текущем тарифе"}
            </Button>
          </DialogClose>
          <Button variant="destructive" onClick={onConfirm} disabled={isPending || isLoading}>
            {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />}
            Перейти на Free
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
