import { useEffect, useRef } from "react"
import { AlertTriangle, X } from "lucide-react"
import { Button } from "./button"
import { cn } from "../../lib/utils"

interface ConfirmDialogProps {
  open: boolean
  title: string
  description?: string
  confirmLabel?: string
  cancelLabel?: string
  variant?: "default" | "destructive"
  onConfirm: () => void | Promise<void>
  onClose: () => void
}

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel = "Подтвердить",
  cancelLabel = "Отмена",
  variant = "default",
  onConfirm,
  onClose,
}: ConfirmDialogProps) {
  const cancelRef = useRef<HTMLButtonElement>(null)
  useEffect(() => {
    if (open) cancelRef.current?.focus()
  }, [open])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal
      aria-labelledby="confirm-title"
    >
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} aria-hidden />
      <div className="relative w-full max-w-sm rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
        <div className="flex items-start gap-3">
          {variant === "destructive" && (
            <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-(--color-destructive)" />
          )}
          <div className="flex-1 min-w-0">
            <h3 id="confirm-title" className="text-sm font-semibold">{title}</h3>
            {description && (
              <p className="mt-1 text-xs text-(--color-muted-foreground)">{description}</p>
            )}
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md p-1 text-(--color-muted-foreground) hover:bg-(--color-muted)"
            aria-label="Закрыть"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <Button ref={cancelRef} type="button" variant="outline" size="sm" onClick={onClose}>
            {cancelLabel}
          </Button>
          <Button
            type="button"
            size="sm"
            variant={variant === "destructive" ? "destructive" : "default"}
            onClick={onConfirm}
            className={cn(variant === "destructive" && "bg-(--color-destructive) text-white")}
          >
            {confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  )
}
