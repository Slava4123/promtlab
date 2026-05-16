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
  FileText,
  Layers,
  Cpu,
  Tag as TagIcon,
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

// Read-only page для просмотра промпта. Layout mirror'ит editor-page
// (Label + content card на каждом section) — это даёт визуальную
// консистентность между «просмотр» и «редактирование» вместо плотного
// потока разнородной мета-информации, как было до Phase 16-Y.
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
  const description = prompt.description?.trim() ?? ""

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
    try {
      await deleteMut.mutateAsync(prompt!.id)
      toast({ title: "Удалён", description: "Можно восстановить из корзины", variant: "info" })
      void qc.invalidateQueries({ queryKey: qk.prompts })
      setDeleteOpen(false)
      navigate("/")
    } catch (err) {
      toast({
        title: "Не удалось удалить",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  return (
    <div className="flex h-full flex-col">
      {/* Sticky header — действия над промптом */}
      <div className="sticky top-0 z-10 flex items-center gap-1 border-b border-(--color-border) bg-(--color-background)/95 p-2 backdrop-blur">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => togglePin.mutate(prompt.id)}
          aria-label={isPinned ? "Открепить" : "Закрепить"}
          className={cn(isPinned && "text-(--color-brand)")}
        >
          <Pin className={cn("h-4 w-4", isPinned && "fill-current")} />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => toggleFav.mutate(prompt.id)}
          aria-label={prompt.favorite ? "Убрать из избранного" : "В избранное"}
          className={cn(prompt.favorite && "text-amber-500")}
        >
          <Star className={cn("h-4 w-4", prompt.favorite && "fill-current")} />
        </Button>
        <div className="flex-1" />
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => setShareOpen(true)}
          aria-label="Поделиться"
        >
          <Share2 className="h-3.5 w-3.5" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => navigate(`/prompts/${prompt.id}/edit`)}
          aria-label="Редактировать"
        >
          <Edit3 className="h-3.5 w-3.5" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => setDeleteOpen(true)}
          aria-label="Удалить"
          className="text-(--color-destructive)"
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        {/* Title block */}
        <header className="space-y-2">
          <h1 className="text-lg font-semibold leading-tight">{prompt.title}</h1>
          {prompt.tags.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {prompt.tags.map((t) => (
                <Link
                  key={t.id}
                  to={`/tags/${t.id}`}
                  className="rounded-md px-2 py-0.5 text-[10px] font-medium transition-opacity hover:opacity-80"
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
        </header>

        {/* Description — only если задано */}
        {description && (
          <section className="space-y-1">
            <Label icon={<FileText className="h-3 w-3" />}>Описание</Label>
            <p className="rounded-md border border-(--color-border)/60 bg-(--color-muted)/20 p-2.5 text-xs leading-relaxed text-(--color-foreground)/90 whitespace-pre-wrap">
              {description}
            </p>
          </section>
        )}

        {/* Content — main prompt body */}
        <section className="space-y-1">
          <Label icon={<Cpu className="h-3 w-3" />}>Промпт</Label>
          <div className="rounded-md border border-(--color-border) bg-(--color-card) p-3 font-mono text-xs leading-relaxed whitespace-pre-wrap break-words">
            {renderTemplate(prompt.content, {})}
          </div>
        </section>

        {/* Metadata badges */}
        {(prompt.collections.length > 0 || prompt.model) && (
          <section className="space-y-1.5">
            <Label icon={<Layers className="h-3 w-3" />}>Метаданные</Label>
            <div className="flex flex-wrap gap-1.5">
              {prompt.model && (
                <span className="inline-flex items-center gap-1 rounded-md border border-(--color-border) bg-(--color-card) px-2 py-1 text-[10px]">
                  <Cpu className="h-3 w-3 text-(--color-muted-foreground)" />
                  <span className="text-(--color-muted-foreground)">Модель:</span>
                  <span className="font-medium">{prompt.model}</span>
                </span>
              )}
              {prompt.collections.map((c) => (
                <Link
                  key={c.id}
                  to={`/collections/${c.id}`}
                  className="inline-flex items-center gap-1 rounded-md border border-(--color-border) bg-(--color-card) px-2 py-1 text-[10px] transition-colors hover:bg-(--color-muted)/40"
                >
                  <Layers className="h-3 w-3" style={{ color: c.color }} />
                  <span className="font-medium">{c.name}</span>
                </Link>
              ))}
            </div>
          </section>
        )}

        {/* Variables hint — если в content есть {{переменные}} */}
        {variables.length > 0 && (
          <section className="rounded-md border border-(--color-brand)/20 bg-(--color-brand-muted) p-2.5">
            <div className="flex items-center gap-1.5">
              <TagIcon className="h-3 w-3 text-(--color-brand)" />
              <span className="text-[10px] font-medium text-(--color-brand)">
                Переменные ({variables.length})
              </span>
            </div>
            <p className="mt-1 text-[10px] leading-relaxed text-(--color-foreground)/80">
              При вставке откроется форма заполнения:{" "}
              {variables.map((v, i) => (
                <span key={v}>
                  <code className="rounded bg-(--color-brand)/10 px-1 py-0.5 text-[9px]">
                    {`{{${v}}}`}
                  </code>
                  {i < variables.length - 1 && " "}
                </span>
              ))}
            </p>
          </section>
        )}
      </div>

      {/* Footer actions */}
      <div className="flex items-center gap-2 border-t border-(--color-border) p-2">
        <Button
          type="button"
          variant="brand"
          onClick={useNow}
          className="flex-1 gap-1.5"
          disabled={insert.submitting}
        >
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

// Label — uppercase mini-header с опциональной иконкой. Mirror'ит editor's
// "Название/Промпт/Модель/..." стиль и группирует визуально каждую секцию.
function Label({ icon, children }: { icon?: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-1 text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
      {icon}
      <span>{children}</span>
    </div>
  )
}
