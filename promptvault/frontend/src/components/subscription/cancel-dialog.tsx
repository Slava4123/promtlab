import { useState } from "react"
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
import { Textarea } from "@/components/ui/textarea"
import type { CancelReason } from "@/api/types"

interface CancelDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (input: { reason?: CancelReason; other_text?: string }) => void
  isPending?: boolean
}

const REASONS: { value: CancelReason; label: string }[] = [
  { value: "too_expensive", label: "Слишком дорого" },
  { value: "not_using", label: "Не использую" },
  { value: "missing_feature", label: "Нет нужной функции" },
  { value: "found_alternative", label: "Нашёл альтернативу" },
  { value: "other", label: "Другая причина" },
]

export function CancelDialog({ open, onOpenChange, onConfirm, isPending = false }: CancelDialogProps) {
  const [reason, setReason] = useState<CancelReason | "">("")
  const [otherText, setOtherText] = useState("")

  const handleConfirm = () => {
    onConfirm({
      reason: reason || undefined,
      other_text: reason === "other" ? otherText.trim() : undefined,
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-destructive/10">
              <AlertTriangle className="size-5 text-destructive" aria-hidden="true" />
            </div>
            <DialogTitle>Отменить подписку?</DialogTitle>
          </div>
          <DialogDescription>
            Доступ сохранится до конца оплаченного периода, потом аккаунт перейдёт на Free.
            Можно вместо отмены поставить на паузу — оставшиеся дни не сгорят.
          </DialogDescription>
        </DialogHeader>

        <fieldset className="space-y-2" disabled={isPending}>
          <legend className="text-sm font-medium text-foreground">
            Поможет улучшить продукт — почему уходите? (необязательно)
          </legend>
          {REASONS.map((r) => (
            <label
              key={r.value}
              className="flex cursor-pointer items-center gap-3 rounded-md border border-border bg-muted/20 px-3 py-2 text-sm transition-colors hover:bg-muted/40 has-[:checked]:border-brand has-[:checked]:bg-brand/5"
            >
              <input
                type="radio"
                name="cancel-reason"
                value={r.value}
                checked={reason === r.value}
                onChange={() => setReason(r.value)}
                className="h-4 w-4 cursor-pointer accent-brand"
              />
              <span>{r.label}</span>
            </label>
          ))}
          {reason === "other" && (
            <Textarea
              value={otherText}
              onChange={(e) => setOtherText(e.target.value)}
              placeholder="Расскажите подробнее — что помогло бы остаться?"
              maxLength={500}
              rows={3}
              className="mt-1"
            />
          )}
        </fieldset>

        <DialogFooter>
          <DialogClose render={<Button variant="outline" disabled={isPending} />}>
            Не отменять
          </DialogClose>
          <Button variant="destructive" onClick={handleConfirm} disabled={isPending}>
            {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />}
            Отменить подписку
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
