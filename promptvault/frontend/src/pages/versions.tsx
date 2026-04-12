import { useState, useRef, useCallback } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { ArrowLeft, History, RotateCcw, Loader2, FileText } from "lucide-react"
import { EmptyState } from "@/components/ui/empty-state"
import { toast } from "sonner"

import { usePrompt } from "@/hooks/use-prompts"
import { useVersions, useRevertVersion } from "@/hooks/use-versions"
import { VersionDiff } from "@/components/prompts/version-diff"
import type { PromptVersion } from "@/api/types"

function formatDate(dateStr: string) {
  const d = new Date(dateStr)
  return d.toLocaleString("ru-RU", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

export default function Versions() {
  const navigate = useNavigate()
  const { id } = useParams()
  const promptId = Number(id)

  const { data: prompt, isLoading: loadingPrompt } = usePrompt(promptId)
  const { data, isLoading: loadingVersions, fetchNextPage, hasNextPage, isFetchingNextPage } = useVersions(promptId)
  const revertVersion = useRevertVersion()
  const [selected, setSelected] = useState<PromptVersion | null>(null)

  const versions = data?.pages.flatMap((p) => p.items) ?? []
  const totalVersions = data?.pages[0]?.total ?? 0

  const isLoading = loadingPrompt || loadingVersions

  // Infinite scroll observer
  const observerRef = useRef<IntersectionObserver | null>(null)
  const loadMoreRef = useCallback((node: HTMLDivElement | null) => {
    if (observerRef.current) observerRef.current.disconnect()
    if (!node) return
    observerRef.current = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
        fetchNextPage()
      }
    })
    observerRef.current.observe(node)
  }, [hasNextPage, isFetchingNextPage, fetchNextPage])

  const handleRevert = async (version: PromptVersion) => {
    if (revertVersion.isPending) return
    try {
      await revertVersion.mutateAsync({ promptId, versionId: version.id })
      toast.success(`Откат к версии ${version.version_number}`)
      setSelected(null)
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Ошибка отката")
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-[72rem]">
      {/* Header */}
      <div className="mb-8 flex items-center gap-3">
        <button
          onClick={() => navigate(-1)}
          className="flex h-8 w-8 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
        </button>
        <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-violet-500/[0.08] ring-1 ring-violet-500/10">
          <History className="h-4 w-4 text-violet-400" />
        </div>
        <div>
          <h1 className="text-lg font-bold tracking-tight text-foreground">
            История версий
          </h1>
          <p className="text-[0.75rem] text-muted-foreground">
            {prompt?.title}
          </p>
        </div>
      </div>

      {versions.length === 0 ? (
        <EmptyState
          icon={<FileText className="h-7 w-7 text-muted-foreground/40" />}
          title="Нет сохранённых версий"
          description="Версии создаются автоматически при каждом обновлении промпта"
        />
      ) : (
        <div className="flex flex-col gap-6 lg:grid lg:grid-cols-[280px_1fr]">
          {/* Timeline */}
          <div className="space-y-1">
            <div className="mb-3 text-[0.75rem] font-medium text-muted-foreground">
              {totalVersions} {totalVersions === 1 ? "версия" : totalVersions < 5 ? "версии" : "версий"}
            </div>
            <div className="flex flex-col gap-1">
              {versions.map((v) => (
                <button
                  key={v.id}
                  onClick={() => setSelected(v)}
                  className={`flex w-full flex-col gap-1 rounded-lg px-3 py-2.5 text-left transition-colors ${
                    selected?.id === v.id
                      ? "bg-violet-500/[0.08] ring-1 ring-violet-500/15"
                      : "hover:bg-foreground/[0.04]"
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <span className={`text-[0.8rem] font-semibold tabular-nums ${
                      selected?.id === v.id ? "text-violet-400" : "text-foreground"
                    }`}>
                      v{v.version_number}
                    </span>
                    <span className="whitespace-nowrap text-[0.7rem] text-muted-foreground">
                      {formatDate(v.created_at)}
                    </span>
                  </div>
                  {v.change_note && (
                    <span className="line-clamp-2 text-[0.72rem] text-muted-foreground">
                      {v.change_note}
                    </span>
                  )}
                  {v.title && (
                    <span className="line-clamp-1 text-[0.7rem] text-muted-foreground">
                      {v.title}
                    </span>
                  )}
                </button>
              ))}
              {hasNextPage && (
                <div ref={loadMoreRef} className="flex justify-center py-3">
                  {isFetchingNextPage && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
                </div>
              )}
            </div>
          </div>

          {/* Diff area */}
          <div className="min-w-0">
            {selected ? (
              <div className="space-y-4">
                <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                  <div className="text-[0.82rem] text-foreground">
                    <span className="font-medium text-violet-400">Версия {selected.version_number}</span>
                    {selected.change_note && (
                      <span className="ml-2 text-muted-foreground"> — {selected.change_note}</span>
                    )}
                  </div>
                  <button
                    onClick={() => handleRevert(selected)}
                    disabled={revertVersion.isPending}
                    className="flex h-8 w-fit items-center gap-1.5 rounded-lg px-3.5 text-[0.78rem] font-medium text-amber-400 transition-colors hover:bg-amber-500/10 disabled:opacity-50"
                    style={{ border: "1px solid rgba(245,158,11,0.15)" }}
                  >
                    {revertVersion.isPending ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : (
                      <RotateCcw className="h-3.5 w-3.5" />
                    )}
                    <span className="sm:inline">Откатить</span>
                  </button>
                </div>

                {/* Title diff (если отличается) */}
                {prompt && selected.title !== prompt.title && (
                  <div className="space-y-1.5">
                    <span className="text-[0.75rem] font-medium text-muted-foreground">Название</span>
                    <VersionDiff
                      oldValue={selected.title}
                      newValue={prompt.title}
                      oldTitle={`v${selected.version_number}`}
                      newTitle="Текущая"
                    />
                  </div>
                )}

                {/* Content diff */}
                {prompt && (
                  <div className="space-y-1.5">
                    <span className="text-[0.75rem] font-medium text-muted-foreground">Содержимое</span>
                    <VersionDiff
                      oldValue={selected.content}
                      newValue={prompt.content}
                      oldTitle={`v${selected.version_number}`}
                      newTitle="Текущая"
                    />
                  </div>
                )}
              </div>
            ) : (
              <div className="hidden flex-col items-center justify-center py-20 text-center lg:flex">
                <History className="h-8 w-8 text-muted-foreground/50" />
                <p className="mt-3 text-sm text-muted-foreground">Выберите версию для сравнения</p>
                <p className="mt-1 text-[0.75rem] text-muted-foreground">
                  Будет показан diff между выбранной версией и текущим состоянием
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
