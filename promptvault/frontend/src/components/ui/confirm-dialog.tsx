import * as React from "react"
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

interface ConfirmDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  icon?: React.ReactNode
  variant?: "destructive" | "brand"
  confirmLabel?: string
  cancelLabel?: string
  onConfirm: () => void
  isPending?: boolean
}

function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  icon,
  variant = "destructive",
  confirmLabel = "Подтвердить",
  cancelLabel = "Отмена",
  onConfirm,
  isPending = false,
}: ConfirmDialogProps) {
  const iconNode = icon ?? (
    <AlertTriangle className="size-5 text-destructive" />
  )

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-destructive/10">
              {iconNode}
            </div>
            <div className="space-y-1">
              <DialogTitle>{title}</DialogTitle>
              <DialogDescription>{description}</DialogDescription>
            </div>
          </div>
        </DialogHeader>
        <DialogFooter>
          <DialogClose render={<Button variant="outline" />}>
            {cancelLabel}
          </DialogClose>
          <Button
            variant={variant === "destructive" ? "destructive-solid" : "brand"}
            onClick={onConfirm}
            disabled={isPending}
          >
            {isPending && <Loader2 className="animate-spin" />}
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export { ConfirmDialog, type ConfirmDialogProps }
