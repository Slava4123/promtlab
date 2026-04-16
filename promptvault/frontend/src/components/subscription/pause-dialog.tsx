import { useState } from "react"
import { Loader2, Pause } from "lucide-react"
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

interface PauseDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (months: 1 | 2 | 3) => void
  isPending?: boolean
}

type Months = 1 | 2 | 3

// Возвращает дату resume = today + N месяцев — для превью в UI ("вернёмся 15 мая").
function computeResumeDate(months: Months): string {
  const d = new Date()
  d.setMonth(d.getMonth() + months)
  return d.toLocaleDateString("ru-RU", { day: "numeric", month: "long", year: "numeric" })
}

export function PauseDialog({ open, onOpenChange, onConfirm, isPending = false }: PauseDialogProps) {
  const [months, setMonths] = useState<Months>(1)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-brand/10">
              <Pause className="size-5 text-brand" />
            </div>
            <DialogTitle>Приостановить подписку</DialogTitle>
          </div>
          <DialogDescription>
            Подписка замораживается — оставшиеся дни сохранятся. На время паузы аккаунт работает
            как на Free. В конце периода подписка возобновится автоматически; можно возобновить
            раньше в любой момент.
          </DialogDescription>
        </DialogHeader>

        <fieldset className="space-y-2" disabled={isPending}>
          <legend className="sr-only">На какой срок приостановить</legend>
          {([1, 2, 3] as Months[]).map((m) => (
            <label
              key={m}
              className="flex cursor-pointer items-start gap-3 rounded-md border border-border bg-muted/20 p-3 text-sm transition-colors hover:bg-muted/40 has-[:checked]:border-brand has-[:checked]:bg-brand/5"
            >
              <input
                type="radio"
                name="pause-months"
                value={m}
                checked={months === m}
                onChange={() => setMonths(m)}
                className="mt-0.5 h-4 w-4 cursor-pointer accent-brand"
              />
              <span className="flex-1">
                <span className="font-medium text-foreground">
                  {m === 1 ? "1 месяц" : `${m} месяца`}
                </span>
                <span className="ml-2 text-muted-foreground">
                  возобновится {computeResumeDate(m)}
                </span>
              </span>
            </label>
          ))}
        </fieldset>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline" disabled={isPending}>
              Отмена
            </Button>
          </DialogClose>
          <Button onClick={() => onConfirm(months)} disabled={isPending}>
            {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />}
            Приостановить
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
