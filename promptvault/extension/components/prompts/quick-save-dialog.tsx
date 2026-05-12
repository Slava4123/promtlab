import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import { Save, X, ExternalLink } from "lucide-react"
import { Button } from "../ui/button"
import { Input } from "../ui/input"
import { Label } from "../ui/label"
import { Textarea } from "../ui/textarea"
import { useToast } from "../ui/toaster"
import { useCreatePrompt } from "../../hooks/use-prompts-crud"
import { useWorkspaceStore } from "../../stores/workspace-store"

const PENDING_KEY = "pv.pendingCapture"

interface PendingCapture {
  content: string
  sourceUrl: string
  capturedAt: number
}

// Quick-save dialog для выделенного текста из context menu.
// Маунтится в AppShell, слушает chrome.storage.session на pendingCapture.
export function QuickSaveDialog() {
  const navigate = useNavigate()
  const { toast } = useToast()
  const team = useWorkspaceStore((s) => s.team)
  const createMut = useCreatePrompt()
  const [open, setOpen] = useState(false)
  const [title, setTitle] = useState("")
  const [content, setContent] = useState("")
  const [sourceUrl, setSourceUrl] = useState("")

  // Загружаем pending capture при mount + listener на изменения storage.
  useEffect(() => {
    async function loadPending() {
      const data = await chrome.storage.session?.get(PENDING_KEY)
      const pending = data?.[PENDING_KEY] as PendingCapture | undefined
      if (pending && pending.content.trim()) {
        setContent(pending.content)
        setSourceUrl(pending.sourceUrl)
        setOpen(true)
        // Auto-suggest title from first 60 chars
        const firstLine = pending.content.split("\n")[0]
        setTitle(firstLine.length > 60 ? `${firstLine.slice(0, 57)}…` : firstLine)
        // Очищаем pending, чтобы при следующем open не было дубликата.
        chrome.storage.session?.remove(PENDING_KEY)
      }
    }
    void loadPending()

    function onChanged(
      changes: { [key: string]: chrome.storage.StorageChange },
      area: chrome.storage.AreaName,
    ) {
      if (area === "session" && PENDING_KEY in changes) {
        void loadPending()
      }
    }
    chrome.storage.onChanged.addListener(onChanged)
    return () => chrome.storage.onChanged.removeListener(onChanged)
  }, [])

  if (!open) return null

  async function handleSave() {
    if (!title.trim() || !content.trim()) {
      toast({ title: "Заполните название и текст", variant: "error" })
      return
    }
    try {
      const saved = await createMut.mutateAsync({
        title: title.trim(),
        content: content.trim(),
        description: sourceUrl ? `Источник: ${sourceUrl}` : "",
        team_id: team?.teamId ?? null,
      })
      toast({ title: "Промпт сохранён", variant: "success" })
      setOpen(false)
      navigate(`/prompts/${saved.id}`)
    } catch (err) {
      toast({
        title: "Не удалось сохранить",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={() => setOpen(false)} />
      <div className="relative w-full max-w-md rounded-lg border border-(--color-border) bg-(--color-background) p-4 shadow-xl">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold">Сохранить как промпт</h3>
          <button
            type="button"
            onClick={() => setOpen(false)}
            className="rounded-md p-1 text-(--color-muted-foreground) hover:bg-(--color-muted)"
            aria-label="Закрыть"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        {sourceUrl && (
          <a
            href={sourceUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-2 inline-flex items-center gap-1 text-[10px] text-(--color-muted-foreground) hover:underline"
          >
            <ExternalLink className="h-3 w-3" />
            {new URL(sourceUrl).hostname}
          </a>
        )}
        <div className="mt-3 space-y-3">
          <div className="space-y-1">
            <Label htmlFor="qs-title">Название</Label>
            <Input
              id="qs-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              autoFocus
              maxLength={100}
            />
          </div>
          <div className="space-y-1">
            <Label htmlFor="qs-content">Содержимое</Label>
            <Textarea
              id="qs-content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={6}
              className="font-mono text-xs"
            />
            <p className="text-[10px] text-(--color-muted-foreground)">
              {content.length.toLocaleString("ru-RU")} симв
            </p>
          </div>
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <Button type="button" variant="outline" size="sm" onClick={() => setOpen(false)}>
            Отмена
          </Button>
          <Button
            type="button"
            size="sm"
            onClick={handleSave}
            disabled={createMut.isPending || !title.trim() || !content.trim()}
            className="gap-1.5"
          >
            <Save className="h-3.5 w-3.5" />
            Сохранить
          </Button>
        </div>
      </div>
    </div>
  )
}
