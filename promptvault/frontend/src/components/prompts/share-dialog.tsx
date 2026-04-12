import { useState } from "react"
import { Copy, Link2, Link2Off, Loader2, Eye } from "lucide-react"
import { toast } from "sonner"

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { useShareLink, useCreateShareLink, useDeleteShareLink } from "@/hooks/use-share"
import { ApiError } from "@/api/client"

interface ShareDialogProps {
  promptId: number
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ShareDialog({ promptId, open, onOpenChange }: ShareDialogProps) {
  const { data: shareLink, isLoading, isError, error } = useShareLink(promptId)
  const createShare = useCreateShareLink()
  const deleteShare = useDeleteShareLink()
  const [confirming, setConfirming] = useState(false)

  const is404 = error instanceof ApiError && error.status === 404
  const hasLink = !isError && !!shareLink
  const showError = isError && !is404

  const handleCreate = () => {
    createShare.mutate(promptId, {
      onSuccess: () => toast.success("Ссылка создана", { description: "Скопируйте и отправьте" }),
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  const handleDeactivate = () => {
    if (!confirming) {
      setConfirming(true)
      return
    }
    deleteShare.mutate(promptId, {
      onSuccess: () => {
        toast.success("Ссылка деактивирована")
        setConfirming(false)
      },
      onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка"),
    })
  }

  const handleCopy = async () => {
    if (!shareLink?.url) return
    try {
      await navigator.clipboard.writeText(shareLink.url)
      toast.success("Ссылка скопирована", { description: "Отправьте получателю" })
    } catch (err) {
      console.error("clipboard write failed:", err)
      toast.error("Не удалось скопировать. Выделите и нажмите Ctrl+C")
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { onOpenChange(o); setConfirming(false) }}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Поделиться промптом</DialogTitle>
        </DialogHeader>

        {isLoading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : showError ? (
          <div className="py-4 text-center text-sm text-muted-foreground">
            Не удалось загрузить данные. Попробуйте закрыть и открыть снова.
          </div>
        ) : hasLink ? (
          <div className="space-y-4">
            {/* URL */}
            <div className="flex min-w-0 items-center gap-2 overflow-hidden rounded-lg border border-border bg-muted/50 p-3">
              <Link2 className="h-4 w-4 shrink-0 text-violet-400" />
              <span className="min-w-0 flex-1 truncate text-xs font-mono text-foreground">
                {shareLink.url}
              </span>
              <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0" onClick={handleCopy}>
                <Copy className="h-3.5 w-3.5" />
              </Button>
            </div>

            {/* Stats */}
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <div className="flex items-center gap-1.5">
                <Eye className="h-3.5 w-3.5" />
                <span>{shareLink.view_count} просмотров</span>
              </div>
              <span>·</span>
              <span>Создана {new Date(shareLink.created_at).toLocaleDateString("ru-RU")}</span>
            </div>

            {/* Actions */}
            <div className="flex flex-col gap-2 sm:flex-row">
              <Button className="w-full sm:flex-1" onClick={handleCopy}>
                <Copy className="mr-2 h-4 w-4" />
                Скопировать ссылку
              </Button>
              <Button
                variant={confirming ? "destructive" : "outline"}
                onClick={handleDeactivate}
                disabled={deleteShare.isPending}
              >
                {deleteShare.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <>
                    <Link2Off className="mr-2 h-4 w-4" />
                    {confirming ? "Подтвердить" : "Отключить"}
                  </>
                )}
              </Button>
            </div>
          </div>
        ) : (
          <div className="space-y-4 py-2">
            <p className="text-sm text-muted-foreground">
              Создайте публичную ссылку — любой сможет просмотреть этот промпт без регистрации.
            </p>
            <Button
              className="w-full"
              onClick={handleCreate}
              disabled={createShare.isPending}
            >
              {createShare.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <Link2 className="mr-2 h-4 w-4" />
              )}
              Создать ссылку
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
