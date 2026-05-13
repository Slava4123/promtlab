import { useState } from "react"
import { Link as LinkIcon, Copy, X, Loader2 } from "lucide-react"
import { Button } from "../ui/button"
import { useShareLink, useCreateShareLink, useDeactivateShareLink } from "../../hooks/use-share"
import { useToast } from "../ui/toaster"

interface ShareDialogProps {
  promptId: number
  open: boolean
  onClose: () => void
}

// Share-link dialog для промпта. State machine:
//   no-link → "Создать ссылку" button
//   active-link → копирование URL + view_count + двойной-клик деактивации
export function ShareDialog({ promptId, open, onClose }: ShareDialogProps) {
  const { toast } = useToast()
  const shareQuery = useShareLink(open ? promptId : null)
  const createMut = useCreateShareLink(promptId)
  const deactivateMut = useDeactivateShareLink(promptId)
  const [confirming, setConfirming] = useState(false)

  if (!open) return null

  const link = shareQuery.data

  async function copy(url: string) {
    try {
      await navigator.clipboard.writeText(url)
      toast({ title: "Ссылка скопирована", variant: "success", durationMs: 1500 })
    } catch {
      toast({ title: "Не удалось скопировать", variant: "error" })
    }
  }

  async function create() {
    try {
      const newLink = await createMut.mutateAsync()
      await copy(newLink.url)
    } catch (err) {
      toast({
        title: "Не удалось создать ссылку",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  async function deactivate() {
    if (!confirming) {
      setConfirming(true)
      setTimeout(() => setConfirming(false), 3000)
      return
    }
    try {
      await deactivateMut.mutateAsync()
      toast({ title: "Ссылка деактивирована", variant: "info" })
      setConfirming(false)
    } catch {
      toast({ title: "Не удалось деактивировать", variant: "error" })
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal
      aria-labelledby="share-title"
    >
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} />
      <div className="relative w-full max-w-sm rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
        <div className="flex items-center gap-2 mb-3">
          <LinkIcon className="h-4 w-4 text-(--color-brand)" />
          <h3 id="share-title" className="flex-1 text-sm font-semibold">Публичная ссылка</h3>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md p-1 text-(--color-muted-foreground) hover:bg-(--color-muted)"
            aria-label="Закрыть"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {shareQuery.isPending ? (
          <div className="flex justify-center py-6">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : link && link.is_active ? (
          <div className="space-y-3">
            <div className="rounded-md border border-(--color-border) bg-(--color-muted)/30 px-2 py-2 font-mono text-[10px] truncate">
              {link.url}
            </div>
            <div className="flex gap-2">
              <Button
                type="button"
                size="sm"
                onClick={() => copy(link.url)}
                className="flex-1 gap-1.5"
              >
                <Copy className="h-3.5 w-3.5" />
                Скопировать
              </Button>
              <Button
                type="button"
                size="sm"
                variant={confirming ? "destructive" : "outline"}
                onClick={deactivate}
                disabled={deactivateMut.isPending}
              >
                {confirming ? "Подтвердить" : "Деактивировать"}
              </Button>
            </div>
            <div className="flex items-center justify-between text-[10px] text-(--color-muted-foreground)">
              <span>Просмотров: {link.view_count}</span>
              <span>Создана: {new Date(link.created_at).toLocaleDateString("ru-RU")}</span>
            </div>
          </div>
        ) : (
          <div className="space-y-3">
            <p className="text-xs text-(--color-muted-foreground)">
              Создаст публичную ссылку. Любой, у кого есть ссылка, сможет просмотреть промпт без аккаунта.
            </p>
            <Button
              type="button"
              variant="brand"
              size="sm"
              onClick={create}
              disabled={createMut.isPending}
              className="w-full gap-1.5"
            >
              <LinkIcon className="h-3.5 w-3.5" />
              {createMut.isPending ? "Создаю…" : "Создать ссылку"}
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
