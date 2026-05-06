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

// overItems — что произойдёт с over-limit ресурсами после downgrade.
//
// Soft-block модель (как Notion / Linear / Figma): сами данные не удаляются,
// все остаются доступны для чтения и использования. Блокируется только
// создание новых сверх лимита Free — пока юзер не сократит количество.
// Возврат на Pro мгновенно снимает ограничение, ничего восстанавливать
// не нужно.
//
// label — статический заголовок категории в номинативе мн. ч. (как table-header):
// рядом с counter'ом «+N» это читается естественно — «Промпты +5», а не
// «5 промптов» в склонённой форме. Общее объяснение «новые после удаления»
// вынесено в подпись под списком, чтобы не повторять в каждом пункте.
interface OverItem {
  count: number
  label: string
}
function overItems(p: DowngradePreview): OverItem[] {
  const out: OverItem[] = []
  if (p.over_prompts > 0) out.push({ count: p.over_prompts, label: "Промпты" })
  if (p.over_collections > 0) out.push({ count: p.over_collections, label: "Коллекции" })
  if (p.over_teams > 0) out.push({ count: p.over_teams, label: "Команды" })
  if (p.over_shares > 0) out.push({ count: p.over_shares, label: "Активные публичные ссылки" })
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
                ? "У вас сейчас больше ресурсов, чем разрешает Free. Сами данные не удаляются — мы только ограничим создание новых до тех пор, пока количество не уложится в лимит."
                : "Всё ваше содержимое помещается в лимиты Free — переход пройдёт без потерь."}
          </DialogDescription>
        </DialogHeader>

        {isLoading ? (
          <div className="flex items-center justify-center py-6">
            <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" aria-hidden="true" />
          </div>
        ) : hasOverages ? (
          <>
            <ul className="divide-y divide-amber-500/20 rounded-md border border-amber-500/30 bg-amber-50/50 text-sm text-amber-900 dark:bg-amber-900/10 dark:text-amber-200">
              {items.map(({ count, label }) => (
                <li key={label} className="flex items-baseline justify-between gap-3 px-3 py-2">
                  <span className="break-words">{label}</span>
                  <span className="shrink-0 font-mono text-[0.78rem] tabular-nums">+{count}</span>
                </li>
              ))}
            </ul>
            <p className="text-[0.78rem] leading-relaxed text-muted-foreground">
              Новые получится создавать после удаления лишних. Возврат на Pro в любой момент мгновенно снимет ограничения.
            </p>
          </>
        ) : null}

        <DialogFooter>
          <DialogClose
            render={
              <Button variant="outline" disabled={isPending} className="w-full sm:w-auto" />
            }
          >
            Отмена
          </DialogClose>
          <Button
            variant="destructive"
            onClick={onConfirm}
            disabled={isPending || isLoading}
            className="w-full sm:w-auto"
          >
            {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />}
            Перейти на Free
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
