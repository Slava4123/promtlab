import { useEffect, useMemo, useRef } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { ArrowLeft, FolderOpen, Loader2 } from "lucide-react"
import { Button } from "../components/ui/button"
import { PromptList } from "../components/prompt-list"
import { PromptListSkeleton } from "../components/prompt-list-skeleton"
import { EmptyState } from "../components/empty-state"
import { useCollections } from "../hooks/use-collections-crud"
import { useInfinitePromptList } from "../hooks/use-prompts"
import { useWorkspaceStore } from "../stores/workspace-store"

// Список промптов в коллекции. Использует тот же useInfinitePromptList,
// что и dashboard, но с filter.collectionId.
export function CollectionDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const collectionId = id ? Number(id) : null
  const teamId = useWorkspaceStore((s) => s.team?.teamId ?? null)
  const scrollRef = useRef<HTMLDivElement>(null)

  const collectionsQuery = useCollections()
  const collection = useMemo(
    () => collectionsQuery.data?.find((c) => c.id === collectionId) ?? null,
    [collectionsQuery.data, collectionId],
  )

  const list = useInfinitePromptList(collectionId !== null, {
    teamId,
    collectionId,
  })

  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    const onScroll = () => {
      if (list.hasNextPage && !list.isFetchingNextPage) {
        const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 80
        if (nearBottom) void list.fetchNextPage()
      }
    }
    el.addEventListener("scroll", onScroll)
    return () => el.removeEventListener("scroll", onScroll)
  }, [list])

  const prompts = list.data?.pages.flatMap((p) => p.items) ?? []
  const isPending = list.isPending || collectionsQuery.isPending

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => navigate("/collections")}
          aria-label="Назад к коллекциям"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex flex-1 items-center gap-2 min-w-0">
          <FolderOpen
            className="h-4 w-4 shrink-0"
            style={{ color: collection?.color ?? "var(--color-muted-foreground)" }}
          />
          <h2 className="truncate text-sm font-semibold">
            {collection?.name ?? "Коллекция"}
          </h2>
          {collection && (
            <span className="text-[10px] text-(--color-muted-foreground)">
              {prompts.length}
              {list.hasNextPage ? "+" : ""}
            </span>
          )}
        </div>
      </div>

      <div ref={scrollRef} className="flex-1 overflow-y-auto p-3">
        {isPending && prompts.length === 0 ? (
          <PromptListSkeleton />
        ) : prompts.length === 0 ? (
          <EmptyState
            title="В коллекции пока пусто"
            description="Откройте промпт и добавьте его в эту коллекцию."
          />
        ) : (
          <>
            <PromptList
              prompts={prompts}
              onSelect={(p) => navigate(`/prompts/${p.id}`)}
              highlightedId={null}
              focusedId={null}
            />
            {list.isFetchingNextPage && (
              <div className="mt-3 flex justify-center">
                <Loader2 className="h-4 w-4 animate-spin text-(--color-muted-foreground)" />
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
