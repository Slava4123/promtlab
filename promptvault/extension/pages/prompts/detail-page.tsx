import { Link, useNavigate, useParams } from "react-router-dom"
import { useQueryClient } from "@tanstack/react-query"
import {
  ArrowLeft,
  Edit3,
  Trash2,
  Share2,
  Copy,
  Send,
  Pin,
  Star,
  Loader2,
} from "lucide-react"
import { Button } from "../../components/ui/button"
import { useToast } from "../../components/ui/toaster"
import { ConfirmDialog } from "../../components/ui/confirm-dialog"
import { usePrompt } from "../../hooks/use-prompts"
import { useDeletePrompt } from "../../hooks/use-prompts-crud"
import { useToggleFavorite, useTogglePin } from "../../hooks/use-mutations"
import { useInsertPrompt } from "../../hooks/use-insert-prompt"
import { ShareDialog } from "../../components/prompts/share-dialog"
import { extractVariables, renderTemplate } from "@pv/shared/template"
import { useState } from "react"
import { cn } from "../../lib/utils"
import { qk } from "../../lib/query-keys"

// Страница просмотра промпта (read-only с действиями).
export function PromptDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const promptId = id ? Number(id) : null
  const promptQuery = usePrompt(promptId)
  const deleteMut = useDeletePrompt()
  const toggleFav = useToggleFavorite()
  const togglePin = useTogglePin()
  const insert = useInsertPrompt()
  const qc = useQueryClient()
  const [shareOpen, setShareOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)

  if (promptQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  const prompt = promptQuery.data
  if (!prompt) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex items-center gap-1 border-b border-(--color-border) p-2">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => navigate("/")}
            aria-label="Назад"
          >
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <h2 className="flex-1 text-sm font-semibold">Не найдено</h2>
        </div>
        <div className="flex flex-1 flex-col items-center justify-center gap-3 p-6 text-center">
          <p className="text-sm font-medium">Промпт не найден</p>
          <p className="text-[10px] text-(--color-muted-foreground)">
            Возможно, промпт удалён или вы вошли под другим аккаунтом.
          </p>
          <Button type="button" size="sm" onClick={() => navigate("/")}>
            К списку промптов
          </Button>
        </div>
      </div>
    )
  }

  const variables = extractVariables(prompt.content)
  const isPinned = prompt.pinned_personal || prompt.pinned_team

  async function copyContent() {
    try {
      await navigator.clipboard.writeText(prompt!.content)
      toast({ title: "Скопировано", variant: "success", durationMs: 1500 })
    } catch {
      toast({ title: "Не удалось скопировать", variant: "error" })
    }
  }

  async function useNow() {
    if (variables.length > 0) {
      navigate(`/prompts/${prompt!.id}/use`)
      return
    }
    await insert.insert(prompt!, prompt!.content)
  }

  async function handleDelete() {
    await deleteMut.mutateAsync(prompt!.id)
    toast({ title: "Удалён", description: "Можно восстановить из корзины", variant: "info" })
    void qc.invalidateQueries({ queryKey: qk.prompts })
    setDeleteOpen(false)
    navigate("/")
  }

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="flex items-center gap-1 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => togglePin.mutate(prompt.id)}
          aria-label="Закрепить"
          className={cn(isPinned && "text-(--color-brand)")}
        >
          <Pin className={cn("h-4 w-4", isPinned && "fill-current")} />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => toggleFav.mutate(prompt.id)}
          aria-label="В избранное"
          className={cn(prompt.favorite && "text-amber-500")}
        >
          <Star className={cn("h-4 w-4", prompt.favorite && "fill-current")} />
        </Button>
        <div className="flex-1" />
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => setShareOpen(true)}
          aria-label="Поделиться"
        >
          <Share2 className="h-3.5 w-3.5" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => navigate(`/prompts/${prompt.id}/edit`)}
          aria-label="Редактировать"
        >
          <Edit3 className="h-3.5 w-3.5" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => setDeleteOpen(true)}
          aria-label="Удалить"
          className="text-(--color-destructive)"
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-3 space-y-3">
        <h1 className="text-base font-semibold">{prompt.title}</h1>
        {prompt.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {prompt.tags.map((t) => (
              <Link
                key={t.id}
                to={`/tags/${t.id}`}
                className="rounded-md px-2 py-0.5 text-[10px]"
                style={{
                  backgroundColor: `${t.color}22`,
                  color: t.color,
                }}
              >
                {t.name}
              </Link>
            ))}
          </div>
        )}
        {prompt.collections.length > 0 && (
          <div className="flex flex-wrap gap-1 text-[10px] text-(--color-muted-foreground)">
            <span>Коллекции:</span>
            {prompt.collections.map((c) => (
              <Link key={c.id} to={`/collections/${c.id}`} className="hover:underline">
                {c.name}
              </Link>
            ))}
          </div>
        )}
        <div className="rounded-md border border-(--color-border) bg-(--color-card) p-3 text-xs whitespace-pre-wrap">
          {renderTemplate(prompt.content, {})}
        </div>
        {prompt.model && (
          <div className="text-[10px] text-(--color-muted-foreground)">
            Модель: {prompt.model}
          </div>
        )}
      </div>

      {/* Footer actions */}
      <div className="flex items-center gap-2 border-t border-(--color-border) p-2">
        <Button type="button" onClick={useNow} className="flex-1 gap-1.5" disabled={insert.submitting}>
          <Send className="h-3.5 w-3.5" />
          {variables.length > 0 ? "Заполнить и вставить" : "Вставить"}
        </Button>
        <Button type="button" variant="outline" size="icon" onClick={copyContent} aria-label="Скопировать">
          <Copy className="h-4 w-4" />
        </Button>
      </div>

      <ShareDialog
        promptId={prompt.id}
        open={shareOpen}
        onClose={() => setShareOpen(false)}
      />
      <ConfirmDialog
        open={deleteOpen}
        title="Удалить промпт?"
        description="Промпт переместится в корзину. Можно восстановить в течение 30 дней."
        confirmLabel="Удалить"
        variant="destructive"
        onConfirm={handleDelete}
        onClose={() => setDeleteOpen(false)}
      />
    </div>
  )
}
