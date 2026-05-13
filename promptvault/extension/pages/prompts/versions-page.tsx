import { useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { ArrowLeft, Loader2, RotateCcw, ChevronRight } from "lucide-react"
import ReactDiffViewer from "react-diff-viewer-continued"
import { Button } from "../../components/ui/button"
import { useToast } from "../../components/ui/toaster"
import { useVersions, useRevertVersion } from "../../hooks/use-versions"
import { usePrompt } from "../../hooks/use-prompts"
import { formatDateTime, formatRelativeDate } from "@pv/shared/utils/format-date"
import { cn } from "../../lib/utils"
import type { PromptVersion } from "../../lib/types"

// Страница истории версий промпта.
//  - Список версий слева (timeline), активная подсвечена
//  - Right pane — side-by-side diff между активной версией и текущим content
//  - Кнопка "Восстановить" создаёт новую версию с content активной
export function VersionsPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { toast } = useToast()
  const promptId = id ? Number(id) : null
  const versionsQuery = useVersions(promptId)
  const promptQuery = usePrompt(promptId)
  const revertMut = useRevertVersion(promptId ?? 0)
  const [activeId, setActiveId] = useState<number | null>(null)

  if (versionsQuery.isPending || promptQuery.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  const versions = versionsQuery.data?.items ?? []
  const prompt = promptQuery.data
  const active: PromptVersion | undefined = activeId
    ? versions.find((v) => v.id === activeId)
    : versions[0]

  async function handleRevert(version: PromptVersion) {
    if (!confirm(`Восстановить версию #${version.version_number}? Создастся новая версия с этим содержимым.`)) {
      return
    }
    try {
      await revertMut.mutateAsync(version.id)
      toast({ title: "Версия восстановлена", variant: "success" })
      navigate(`/prompts/${promptId}`)
    } catch (err) {
      toast({
        title: "Не удалось восстановить",
        description: err instanceof Error ? err.message : undefined,
        variant: "error",
      })
    }
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 truncate text-sm font-semibold">
          История: {prompt?.title ?? "промпт"}
        </h2>
      </div>

      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Timeline */}
        <div className="border-b border-(--color-border) max-h-[35%] overflow-y-auto">
          {versions.length === 0 ? (
            <div className="p-3 text-xs text-(--color-muted-foreground)">
              История пуста.
            </div>
          ) : (
            <ul className="divide-y divide-(--color-border)">
              {versions.map((v) => (
                <li key={v.id}>
                  <button
                    type="button"
                    onClick={() => setActiveId(v.id)}
                    className={cn(
                      "flex w-full items-center gap-2 px-3 py-2 text-left transition-colors",
                      active?.id === v.id
                        ? "bg-(--color-brand-muted)"
                        : "hover:bg-(--color-muted)/40",
                    )}
                  >
                    <span className="rounded bg-(--color-muted) px-1.5 py-0.5 text-[10px] font-mono">
                      v{v.version_number}
                    </span>
                    <div className="flex-1 min-w-0">
                      <div className="truncate text-xs font-medium">{v.title}</div>
                      <div className="flex items-center gap-1 text-[10px] text-(--color-muted-foreground)">
                        <span>{formatRelativeDate(v.created_at)}</span>
                        {v.changed_by_name && (
                          <>
                            <span>•</span>
                            <span className="truncate">{v.changed_by_name}</span>
                          </>
                        )}
                      </div>
                      {v.change_note && (
                        <div className="mt-0.5 truncate text-[10px] text-(--color-muted-foreground)">
                          {v.change_note}
                        </div>
                      )}
                    </div>
                    <ChevronRight className="h-3 w-3 text-(--color-muted-foreground)" />
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Diff */}
        <div className="flex-1 overflow-y-auto">
          {active && prompt ? (
            <div className="flex flex-col h-full">
              <div className="flex items-center justify-between border-b border-(--color-border) p-2">
                <div className="text-xs">
                  <div className="font-semibold">Версия #{active.version_number}</div>
                  <div className="text-(--color-muted-foreground)">
                    {formatDateTime(active.created_at)}
                  </div>
                </div>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  onClick={() => handleRevert(active)}
                  disabled={revertMut.isPending}
                  className="gap-1.5"
                >
                  <RotateCcw className="h-3.5 w-3.5" />
                  Восстановить
                </Button>
              </div>
              {/* Title diff */}
              <div className="border-b border-(--color-border) p-2 text-xs">
                <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground) mb-1">
                  Название
                </div>
                <ReactDiffViewer
                  oldValue={active.title}
                  newValue={prompt.title}
                  splitView={false}
                  hideLineNumbers
                  useDarkTheme
                />
              </div>
              {/* Content diff */}
              <div className="p-2 text-xs flex-1 overflow-y-auto">
                <div className="text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground) mb-1">
                  Содержимое
                </div>
                <ReactDiffViewer
                  oldValue={active.content}
                  newValue={prompt.content}
                  splitView={false}
                  hideLineNumbers
                  useDarkTheme
                />
              </div>
            </div>
          ) : (
            <div className="flex h-full items-center justify-center text-xs text-(--color-muted-foreground)">
              Выберите версию для сравнения
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
