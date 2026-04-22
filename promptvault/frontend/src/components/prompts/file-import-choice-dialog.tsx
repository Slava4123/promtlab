import { FileText } from "lucide-react"
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

export type FileImportChoice = "replace" | "prepend" | "append"

interface FileImportChoiceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  filename: string
  charCount: number
  /** Вызывается с выбранной стратегией. cancel = закрытие диалога без вызова. */
  onChoose: (choice: FileImportChoice) => void
}

/**
 * Показывается когда в content редактора уже есть текст и пользователь
 * инициировал импорт файла. 3 варианта вставки + отмена. "Заменить" —
 * destructive, default focus. "В начало" и "В конец" — brand. "Отмена" — outline.
 *
 * Узкий use-case: не расширяем общий ConfirmDialog, а пишем свой — проще,
 * проще тестировать, не тянет семантику "подтвердить destructive операцию".
 */
export function FileImportChoiceDialog({
  open,
  onOpenChange,
  filename,
  charCount,
  onChoose,
}: FileImportChoiceDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-brand/10">
              <FileText className="size-5 text-brand" />
            </div>
            <div className="space-y-1">
              <DialogTitle>Вставка файла {filename}</DialogTitle>
              <DialogDescription>
                В редакторе уже есть текст. Куда вставить содержимое из файла
                ({charCount.toLocaleString("ru-RU")} {pluralSymbols(charCount)})?
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>
        <DialogFooter className="flex-col gap-2 sm:flex-col">
          <Button
            variant="destructive-solid"
            autoFocus
            onClick={() => onChoose("replace")}
          >
            Заменить
          </Button>
          <Button variant="brand" onClick={() => onChoose("prepend")}>
            Вставить в начало
          </Button>
          <Button variant="brand" onClick={() => onChoose("append")}>
            Вставить в конец
          </Button>
          <DialogClose render={<Button variant="outline" />}>Отмена</DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function pluralSymbols(n: number): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod100 >= 11 && mod100 <= 14) return "символов"
  if (mod10 === 1) return "символ"
  if (mod10 >= 2 && mod10 <= 4) return "символа"
  return "символов"
}
