import { useState, useEffect, useRef, useCallback } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { Plus, Search, Star, FileText, Loader2, Pin, Clock, Flame } from "lucide-react"
import { toast } from "sonner"

import { EmptyState } from "@/components/ui/empty-state"
import { PromptCard, PromptCardSkeleton } from "@/components/prompts/prompt-card"
import { Skeleton } from "@/components/ui/skeleton"
import { DismissibleBanner } from "@/components/hints/dismissible-banner"
import { UsePromptDialog } from "@/components/prompts/use-prompt-dialog"
import { usePrompts, useToggleFavorite, useTogglePin, useIncrementUsage, useDeletePrompt, usePinnedPrompts, useRecentPrompts } from "@/hooks/use-prompts"
import { useRestoreItem } from "@/hooks/use-trash"
import { useTags } from "@/hooks/use-tags"
import { useCollections } from "@/hooks/use-collections"
import { useStreak } from "@/hooks/use-streaks"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { hasVariables } from "@/lib/template/parse"
import { captureException } from "@/lib/sentry"
import type { Prompt } from "@/api/types"

export default function Dashboard() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const team = useWorkspaceStore((s) => s.team)
  const teamId = team?.teamId ?? null
  const teamName = team?.teamName ?? null
  const collectionId = searchParams.get("collection_id") ? Number(searchParams.get("collection_id")) : undefined
  const [search, setSearch] = useState("")
  const [debouncedSearch, setDebouncedSearch] = useState("")
  const [favoriteOnly, setFavoriteOnly] = useState(false)
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([])
  const [tagsExpanded, setTagsExpanded] = useState(false)

  // Debounce search
  useEffect(() => {
    const t = setTimeout(() => setDebouncedSearch(search), 300)
    return () => clearTimeout(t)
  }, [search])

  const { data: tags } = useTags(teamId)
  const { data: collections, isLoading: collectionsLoading } = useCollections(teamId)
  const { data: streak, isLoading: streakLoading } = useStreak()

  const {
    data,
    isLoading,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
  } = usePrompts({
    q: debouncedSearch || undefined,
    favorite: favoriteOnly || undefined,
    collection_id: collectionId,
    tag_ids: selectedTagIds.length > 0 ? selectedTagIds : undefined,
    team_id: teamId,
  })

  const toggleFav = useToggleFavorite()
  const togglePin = useTogglePin()
  const incrementUsage = useIncrementUsage()
  const deletePrompt = useDeletePrompt()

  const { data: pinnedData, error: pinnedError } = usePinnedPrompts(teamId)
  const { data: recentData, error: recentError } = useRecentPrompts(teamId)
  const restoreItem = useRestoreItem()
  const [usePromptModal, setUsePromptModal] = useState<Prompt | null>(null)

  const handleDelete = useCallback(
    (id: number) => {
      if (deletePrompt.isPending) return
      deletePrompt.mutate(id, {
        onSuccess: () => {
          toast("Промпт перемещён в корзину", {
            action: {
              label: "Отменить",
              onClick: () => restoreItem.mutate({ type: "prompt", id }),
            },
          })
        },
        onError: (e) => toast.error(e instanceof Error ? e.message : "Ошибка удаления"),
      })
    },
    [deletePrompt, restoreItem],
  )

  const handleUse = useCallback(
    async (prompt: Prompt) => {
      if (hasVariables(prompt.content)) {
        setUsePromptModal(prompt)
        return
      }
      try {
        await navigator.clipboard.writeText(prompt.content)
        incrementUsage.mutate(prompt.id)
        toast.success("Скопировано")
      } catch (err) {
        captureException(err instanceof Error ? err : new Error(String(err)), { tags: { feature: "clipboard-copy" } })
        toast.error("Не удалось скопировать")
      }
    },
    [incrementUsage],
  )

  // Infinite scroll — IntersectionObserver
  const sentinelRef = useRef<HTMLDivElement>(null)
  const handleIntersect = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
        fetchNextPage()
      }
    },
    [hasNextPage, isFetchingNextPage, fetchNextPage],
  )

  useEffect(() => {
    const el = sentinelRef.current
    if (!el) return
    const observer = new IntersectionObserver(handleIntersect, { rootMargin: "200px" })
    observer.observe(el)
    return () => observer.disconnect()
  }, [handleIntersect])

  // Flatten pages
  const allItems = data?.pages.flatMap((p) => p.items) ?? []
  const total = data?.pages[0]?.total ?? 0
  const usageCount = allItems.reduce((s, p) => s + p.usage_count, 0)

  return (
    <div className="mx-auto max-w-[64rem] space-y-5">
      {/* Header */}
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="min-w-0">
          <h1 className="text-2xl font-bold tracking-tight">
            {teamName ? `Промпты — ${teamName}` : "Промпты"}
          </h1>
          <p className="mt-0.5 text-[0.8rem] text-muted-foreground">
            {teamName ? "Командная библиотека промптов" : "Ваша библиотека промптов"}
          </p>
        </div>
        <button
          onClick={() => navigate("/prompts/new")}
          className="flex h-11 items-center gap-1.5 rounded-lg bg-violet-600 px-3.5 text-[0.8rem] font-medium text-white shadow-lg shadow-violet-600/10 transition-[color,background-color,transform,box-shadow] hover:bg-violet-500 hover:shadow-violet-500/20 active:scale-[0.97]"
        >
          <Plus className="h-3.5 w-3.5" />
          Новый
        </button>
      </div>

      {/* Stats — показываем skeleton пока грузим промпты/коллекции вместо фальшивых нулей (X-9). */}
      <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-4">
        <div className="rounded-xl border border-border bg-card px-3.5 py-2.5">
          <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">Всего</p>
          {isLoading ? (
            <Skeleton className="mt-1 h-5 w-10" />
          ) : (
            <p className="mt-0.5 text-lg font-bold tabular-nums text-foreground">{total}</p>
          )}
        </div>
        <div className="rounded-xl border border-yellow-500/15 bg-yellow-500/[0.03] px-3.5 py-2.5">
          <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">Избранное</p>
          {isLoading ? (
            <Skeleton className="mt-1 h-5 w-8" />
          ) : (
            <p className="mt-0.5 text-lg font-bold tabular-nums text-yellow-400">{allItems.filter(p => p.favorite).length}</p>
          )}
        </div>
        <div className="rounded-xl border border-violet-500/15 bg-violet-500/[0.03] px-3.5 py-2.5">
          <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">Использований</p>
          {isLoading ? (
            <Skeleton className="mt-1 h-5 w-12" />
          ) : (
            <p className="mt-0.5 text-lg font-bold tabular-nums text-violet-400">{usageCount}</p>
          )}
        </div>
        <div className="rounded-xl border border-border bg-card px-3.5 py-2.5">
          <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">Коллекции</p>
          {collectionsLoading ? (
            <Skeleton className="mt-1 h-5 w-8" />
          ) : (
            <p className="mt-0.5 text-lg font-bold tabular-nums text-foreground">{collections?.length ?? 0}</p>
          )}
        </div>
      </div>

      {/* Streak — skeleton пока грузим, иначе мигает "0 дней подряд" на свежем заходе (X-9). */}
      {streakLoading ? (
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 rounded-xl border border-orange-500/15 bg-orange-500/[0.03] px-4 py-3">
          <div className="flex items-center gap-2">
            <Flame className="h-5 w-5 text-orange-400/60" />
            <Skeleton className="h-6 w-8" />
            <Skeleton className="h-4 w-20" />
          </div>
          <Skeleton className="h-3 w-48" />
        </div>
      ) : (
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 rounded-xl border border-orange-500/15 bg-orange-500/[0.03] px-4 py-3">
          <div className="flex items-center gap-2">
            <Flame className={`h-5 w-5 text-orange-400${streak?.active_today ? " animate-pulse" : ""}`} />
            <span className="text-xl font-bold tabular-nums text-orange-400">{streak?.current_streak ?? 0}</span>
            <span className="text-sm font-medium text-orange-400/80">
              {(streak?.current_streak ?? 0) === 1 ? "день" : (streak?.current_streak ?? 0) < 5 ? "дня" : "дней"} подряд
            </span>
          </div>
          <span className="text-xs text-muted-foreground">
            {!streak || streak.current_streak === 0
              ? "Используйте промпт, чтобы начать серию"
              : streak.active_today
                ? streak.current_streak < 7 ? "Так держать!" : "Вы на огне!"
                : "Используйте промпты, чтобы не потерять серию"}
          </span>
          {streak && streak.longest_streak > streak.current_streak && (
            <span className="ml-auto text-xs text-muted-foreground">Рекорд: {streak.longest_streak}</span>
          )}
        </div>
      )}

      {/* M-13: education banners — показываются пока юзер не закроет × */}
      <DismissibleBanner
        id="dashboard_extension"
        title="Расширение для Chrome и Firefox"
        description="Вставляйте любимые промпты прямо в ChatGPT/Claude/любой веб-чат одной клавишей."
        cta={{ label: "Открыть в настройках", onClick: () => navigate("/settings/integrations") }}
        tone="violet"
      />
      <DismissibleBanner
        id="dashboard_mcp"
        title="MCP-сервер для Claude Code и Claude Desktop"
        description="Сохраняйте промпты из переписки с Claude одной командой и вытягивайте их обратно в любой чат."
        cta={{ label: "Как подключить", onClick: () => navigate("/settings/integrations") }}
        tone="emerald"
      />

      {/* Pinned section */}
      {pinnedError && (
        <p className="text-[0.75rem] text-red-400">Не удалось загрузить закреплённые промпты</p>
      )}
      {pinnedData && pinnedData.items.length > 0 && (
        <div className="space-y-2">
          <div className="flex items-center gap-1.5">
            <Pin className="h-3.5 w-3.5 text-violet-400" />
            <h2 className="text-[0.8rem] font-semibold text-foreground">Закреплённые</h2>
            <span className="text-[0.7rem] text-muted-foreground">{pinnedData.total}</span>
          </div>
          <div className="overflow-hidden -mx-4 px-4">
            <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-thin">
              {pinnedData.items.map((prompt) => (
                <button
                  key={prompt.id}
                  onClick={() => navigate(`/prompts/${prompt.id}`)}
                  className="group flex min-w-[140px] max-w-[200px] shrink-0 flex-col gap-1.5 rounded-lg border border-violet-500/15 bg-violet-500/[0.03] p-2.5 text-left transition-[transform,box-shadow,border-color] hover:-translate-y-0.5 hover:border-violet-500/25 hover:shadow-md"
              >
                <p className="line-clamp-1 text-[0.78rem] font-medium text-foreground">{prompt.title}</p>
                {prompt.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {prompt.tags.slice(0, 3).map((tag) => (
                      <span
                        key={tag.id}
                        className="rounded-full px-1.5 py-px text-[0.6rem]"
                        style={{ backgroundColor: tag.color + "20", color: tag.color }}
                      >
                        {tag.name}
                      </span>
                    ))}
                  </div>
                )}
              </button>
            ))}
            </div>
          </div>
        </div>
      )}

      {/* Recent section */}
      {recentError && (
        <p className="text-[0.75rem] text-red-400">Не удалось загрузить недавние промпты</p>
      )}
      {recentData && recentData.items.length > 0 && (
        <div className="space-y-2">
          <div className="flex items-center gap-1.5">
            <Clock className="h-3.5 w-3.5 text-muted-foreground" />
            <h2 className="text-[0.8rem] font-semibold text-foreground">Недавние</h2>
            <button
              onClick={() => navigate("/history")}
              className="ml-auto shrink-0 text-[0.7rem] text-muted-foreground transition-colors hover:text-foreground/70"
            >
              Все
            </button>
          </div>
          <div className="overflow-hidden -mx-4 px-4">
            <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-thin">
              {recentData.items.map((prompt) => (
                <button
                  key={prompt.id}
                  onClick={() => navigate(`/prompts/${prompt.id}`)}
                  className="group flex min-w-[140px] max-w-[200px] shrink-0 flex-col gap-1.5 rounded-lg border border-border bg-card p-2.5 text-left transition-[transform,box-shadow,border-color] hover:-translate-y-0.5 hover:border-border/80 hover:shadow-md"
                >
                  <p className="line-clamp-1 text-[0.78rem] font-medium text-foreground">{prompt.title}</p>
                  {prompt.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {prompt.tags.slice(0, 3).map((tag) => (
                        <span
                          key={tag.id}
                          className="rounded-full px-1.5 py-px text-[0.6rem]"
                          style={{ backgroundColor: tag.color + "20", color: tag.color }}
                        >
                          {tag.name}
                        </span>
                      ))}
                    </div>
                  )}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Search + Chips */}
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <input
            id="prompt-search"
            type="text"
            placeholder="Поиск..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-11 w-full rounded-lg border border-border bg-muted/20 pl-8 pr-12 text-[0.8rem] text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-violet-500/25 focus:bg-muted/30 focus:ring-1 focus:ring-violet-500/10"
          />
          {!search && <kbd className="absolute right-2.5 top-1/2 hidden -translate-y-1/2 rounded border border-border bg-muted/30 px-1 py-px text-[9px] text-muted-foreground sm:inline">⌘K</kbd>}
        </div>
        <div className="flex gap-1">
          <button
            onClick={() => setFavoriteOnly(false)}
            className={`flex h-11 items-center gap-1 rounded-lg border px-2.5 text-[0.72rem] font-medium transition-colors ${
              !favoriteOnly
                ? "border-violet-500/20 bg-violet-500/10 text-violet-300"
                : "border-border text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            Все
          </button>
          <button
            onClick={() => setFavoriteOnly(true)}
            className={`flex h-11 items-center gap-1 rounded-lg border px-2.5 text-[0.72rem] font-medium transition-colors ${
              favoriteOnly
                ? "border-violet-500/20 bg-violet-500/10 text-violet-300"
                : "border-border text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            <Star className={`h-3 w-3 ${favoriteOnly ? "fill-yellow-500 text-yellow-500" : ""}`} />
            Избранное
          </button>
        </div>
      </div>

      {/* Tag filters */}
      {tags && tags.length > 0 && (
        <div className="space-y-1.5">
          <div className={`relative flex flex-wrap gap-1 overflow-hidden transition-[max-height] ${tagsExpanded ? "" : "max-h-[60px]"}`}>
            {tags.map((tag) => {
              const isActive = selectedTagIds.includes(tag.id)
              const color = tag.color || "#6366f1"
              return (
                <button
                  key={tag.id}
                  onClick={() =>
                    setSelectedTagIds((prev) =>
                      isActive ? prev.filter((id) => id !== tag.id) : [...prev, tag.id],
                    )
                  }
                  className={`rounded-full px-2.5 py-1 text-[0.72rem] font-medium transition-colors ${
                    isActive ? "ring-1" : "hover:opacity-100"
                  }`}
                  style={{
                    backgroundColor: color + (isActive ? "40" : "18"),
                    color: color + (isActive ? "ff" : "cc"),
                    ...(isActive ? { boxShadow: `inset 0 0 0 1px ${color}60` } : {}),
                  }}
                >
                  {tag.name}
                </button>
              )
            })}
            {!tagsExpanded && tags.length > 15 && (
              <div className="pointer-events-none absolute inset-x-0 bottom-0 h-5 bg-gradient-to-t from-background to-transparent" />
            )}
          </div>
          <div className="flex gap-2">
            {tags.length > 15 && (
              <button
                onClick={() => setTagsExpanded(!tagsExpanded)}
                className="text-[0.7rem] text-muted-foreground transition-colors hover:text-foreground/70"
              >
                {tagsExpanded ? "Свернуть" : `Ещё ${tags.length - 15}+`}
              </button>
            )}
            {selectedTagIds.length > 0 && (
              <button
                onClick={() => setSelectedTagIds([])}
                className="text-[0.7rem] text-muted-foreground transition-colors hover:text-foreground/70"
              >
                Сбросить
              </button>
            )}
          </div>
        </div>
      )}

      {/* Cards */}
      {isLoading ? (
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <PromptCardSkeleton key={i} />
          ))}
        </div>
      ) : allItems.length === 0 ? (
        <EmptyState
          icon={<FileText className="h-7 w-7 text-violet-400/60" />}
          title={debouncedSearch ? "Ничего не найдено" : "Пока нет промптов"}
          description={debouncedSearch ? "Попробуйте другой запрос" : "Создайте первый промпт или выберите готовый из каталога шаблонов"}
          action={!debouncedSearch ? (
            <div className="flex flex-wrap justify-center gap-2">
              <button
                onClick={() => navigate("/prompts/new")}
                className="flex h-11 items-center gap-1.5 rounded-lg bg-violet-600 px-4 text-[0.8rem] font-medium text-white shadow-lg shadow-violet-600/10 transition-[color,background-color,transform,box-shadow] hover:bg-violet-500 active:scale-[0.97]"
              >
                <Plus className="h-3.5 w-3.5" />
                Создать промпт
              </button>
              <button
                onClick={() => navigate("/welcome")}
                className="flex h-11 items-center gap-1.5 rounded-lg border border-border bg-card px-4 text-[0.8rem] font-medium text-foreground transition-colors hover:bg-muted/50 active:scale-[0.97]"
              >
                Выбрать из шаблонов
              </button>
            </div>
          ) : undefined}
        />
      ) : (
        <>
          <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {allItems.map((prompt) => (
              <PromptCard
                key={prompt.id}
                prompt={prompt}
                onToggleFavorite={(id) => toggleFav.mutate(id)}
                onTogglePin={(id, teamWide) => togglePin.mutate({ id, teamWide })}
                onClick={(id) => navigate(`/prompts/${id}`)}
                onUse={handleUse}
                onDelete={handleDelete}
              />
            ))}
            {isFetchingNextPage &&
              Array.from({ length: 3 }).map((_, i) => (
                <PromptCardSkeleton key={`skel-${i}`} />
              ))
            }
          </div>

          {/* Sentinel for infinite scroll */}
          <div ref={sentinelRef} className="flex justify-center py-4">
            {isFetchingNextPage && (
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            )}
            {!hasNextPage && allItems.length > 18 && (
              <p className="text-[0.75rem] text-muted-foreground">Все промпты загружены</p>
            )}
          </div>
        </>
      )}

      {usePromptModal && (
        <UsePromptDialog
          prompt={usePromptModal}
          open
          onOpenChange={(o) => !o && setUsePromptModal(null)}
        />
      )}
    </div>
  )
}
