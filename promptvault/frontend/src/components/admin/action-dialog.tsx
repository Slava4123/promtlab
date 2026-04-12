import { useState, useEffect } from "react"
import { Loader2, ShieldAlert } from "lucide-react"

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
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

interface ActionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  confirmLabel?: string
  /**
   * requireTOTP — когда true, показывается поле ввода TOTP кода (sudo mode).
   * onConfirm получает код или undefined (если requireTOTP=false).
   */
  requireTOTP?: boolean
  onConfirm: (totpCode?: string) => Promise<void> | void
}

/**
 * ActionDialog — универсальный подтверждающий диалог для destructive admin
 * actions. Поддерживает опциональный TOTP re-verification (sudo mode).
 *
 * UX:
 * - User нажимает action → диалог открывается с description объяснением
 * - Если requireTOTP — обязательное поле с 6-значным кодом
 * - "Подтвердить" disabled пока не введён code (когда required)
 * - При ошибке (backend возвращает 401 на неверный TOTP) показывается
 *   inline error; code input сбрасывается
 */
export function ActionDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel = "Подтвердить",
  requireTOTP = false,
  onConfirm,
}: ActionDialogProps) {
  const [code, setCode] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  // Сброс state при каждом открытии.
  useEffect(() => {
    if (open) {
      setCode("")
      setError("")
      setPending(false)
    }
  }, [open])

  const handleConfirm = async () => {
    if (requireTOTP && !code) return
    setPending(true)
    setError("")
    try {
      await onConfirm(requireTOTP ? code : undefined)
      onOpenChange(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка")
      if (requireTOTP) setCode("")
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-destructive/10">
              <ShieldAlert className="size-5 text-destructive" />
            </div>
            <div className="space-y-1">
              <DialogTitle>{title}</DialogTitle>
              <DialogDescription>{description}</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        {requireTOTP && (
          <div className="space-y-1.5 pt-2">
            <Label htmlFor="totp_verify" className="text-xs">
              Код из Authenticator (или backup-код)
            </Label>
            <Input
              id="totp_verify"
              inputMode="text"
              autoComplete="one-time-code"
              placeholder="000000"
              value={code}
              onChange={(e) => {
                setCode(e.target.value)
                if (error) setError("")
              }}
              className="text-center tracking-widest"
            />
            {error && <p className="text-xs text-destructive">{error}</p>}
          </div>
        )}

        {!requireTOTP && error && (
          <p className="pt-2 text-xs text-destructive">{error}</p>
        )}

        <DialogFooter>
          <DialogClose render={<Button variant="outline" />}>Отмена</DialogClose>
          <Button
            variant="destructive-solid"
            onClick={handleConfirm}
            disabled={pending || (requireTOTP && !code)}
          >
            {pending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
