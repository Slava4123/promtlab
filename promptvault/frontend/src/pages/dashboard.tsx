import { useState, useEffect, useRef, useCallback } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { Plus, Search, Star, FileText, Loader2 } from "lucide-react"

import { PromptCard, PromptCardSkeleton } from "@/components/prompts/prompt-card"
import { usePrompts, useToggleFavorite } from "@/hooks/use-prompts"
import { useTags } from "@/hooks/use-tags"
import { useCollections } from "@/hooks/use-collections"
import { useWorkspaceStore } from "@/stores/workspace-store"

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
  const { data: collections } = useCollections(teamId)

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
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">
            {teamName ? `Промпты — ${teamName}` : "Промпты"}
          </h1>
          <p className="mt-0.5 text-[0.8rem] text-muted-foreground">
            {teamName ? "Командная библиотека промптов" : "Ваша библиотека AI-промптов"}
          </p>
        </div>
        <button
          onClick={() => navigate("/prompts/new")}
          className="flex h-11 items-center gap-1.5 rounded-lg bg-violet-600 px-3.5 text-[0.8rem] font-medium text-white shadow-lg shadow-violet-600/10 transition-all hover:bg-violet-500 hover:shadow-violet-500/20 active:scale-[0.97]"
        >
          <Plus className="h-3.5 w-3.5" />
          Новый
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-4">
        <div className="rounded-xl border border-border bg-card px-3.5 py-2.5">
          <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">Всего</p>
          <p className="mt-0.5 text-lg font-bold tabular-nums text-foreground">{total}</p>
        </div>
        <div className="rounded-xl border border-yellow-500/15 bg-yellow-500/[0.03] px-3.5 py-2.5">
          <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">Избранное</p>
          <p className="mt-0.5 text-lg font-bold tabular-nums text-yellow-400">{allItems.filter(p => p.favorite).length}</p>
        </div>
        <div className="rounded-xl border border-violet-500/15 bg-violet-500/[0.03] px-3.5 py-2.5">
          <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">Использований</p>
          <p className="mt-0.5 text-lg font-bold tabular-nums text-violet-400">{usageCount}</p>
        </div>
        <div className="rounded-xl border border-border bg-card px-3.5 py-2.5">
          <p className="text-[0.65rem] font-medium uppercase tracking-wider text-muted-foreground">Коллекции</p>
          <p className="mt-0.5 text-lg font-bold tabular-nums text-foreground">{collections?.length ?? 0}</p>
        </div>
      </div>

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
            className="h-11 w-full rounded-lg border border-border bg-muted/20 pl-8 pr-12 text-[0.8rem] text-foreground outline-none transition-all placeholder:text-muted-foreground focus:border-violet-500/25 focus:bg-muted/30 focus:ring-1 focus:ring-violet-500/10"
          />
          {!search && <kbd className="absolute right-2.5 top-1/2 hidden -translate-y-1/2 rounded border border-border bg-muted/30 px-1 py-px text-[9px] text-muted-foreground sm:inline">⌘K</kbd>}
        </div>
        <div className="flex gap-1">
          <button
            onClick={() => setFavoriteOnly(false)}
            className={`flex h-11 items-center gap-1 rounded-lg border px-2.5 text-[0.72rem] font-medium transition-all ${
              !favoriteOnly
                ? "border-violet-500/20 bg-violet-500/10 text-violet-300"
                : "border-border text-muted-foreground hover:bg-muted hover:text-foreground"
            }`}
          >
            Все
          </button>
          <button
            onClick={() => setFavoriteOnly(true)}
            className={`flex h-11 items-center gap-1 rounded-lg border px-2.5 text-[0.72rem] font-medium transition-all ${
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
          <div className={`relative flex flex-wrap gap-1 overflow-hidden transition-all ${tagsExpanded ? "" : "max-h-[60px]"}`}>
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
                  className={`rounded-full px-2.5 py-1 text-[0.72rem] font-medium transition-all ${
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
        <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <PromptCardSkeleton key={i} />
          ))}
        </div>
      ) : allItems.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-violet-500/[0.08] ring-1 ring-violet-500/10">
            <FileText className="h-7 w-7 text-violet-400/60" />
          </div>
          <p className="text-base font-medium text-muted-foreground">
            {debouncedSearch ? "Ничего не найдено" : "Пока нет промптов"}
          </p>
          <p className="mt-1 text-sm text-muted-foreground">
            {debouncedSearch ? "Попробуйте другой запрос" : "Создайте первый промпт для вашей библиотеки"}
          </p>
          {!debouncedSearch && (
            <button
              onClick={() => navigate("/prompts/new")}
              className="mt-5 flex h-11 items-center gap-1.5 rounded-lg bg-violet-600 px-4 text-[0.8rem] font-medium text-white shadow-lg shadow-violet-600/10 transition-all hover:bg-violet-500 active:scale-[0.97]"
            >
              <Plus className="h-3.5 w-3.5" />
              Создать промпт
            </button>
          )}
        </div>
      ) : (
        <>
          <div className="grid gap-2.5 sm:grid-cols-2 lg:grid-cols-3">
            {allItems.map((prompt) => (
              <PromptCard
                key={prompt.id}
                prompt={prompt}
                onToggleFavorite={(id) => toggleFav.mutate(id)}
                onClick={(id) => navigate(`/prompts/${id}`)}
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
    </div>
  )
}
