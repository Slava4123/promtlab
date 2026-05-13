import { useEffect, useMemo, useRef } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { ArrowLeft, Tag as TagIcon, Loader2 } from "lucide-react"
import { Button } from "../components/ui/button"
import { PromptList } from "../components/prompt-list"
import { PromptListSkeleton } from "../components/prompt-list-skeleton"
import { EmptyState } from "../components/empty-state"
import { useTags } from "../hooks/use-tags-crud"
import { useInfinitePromptList } from "../hooks/use-prompts"
import { useWorkspace } from "../hooks/use-workspace"

// Список промптов с конкретным тегом. Использует filter.tagIds на backend.
export function TagDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const tagId = id ? Number(id) : null
  const teamId = useWorkspace().workspaceId
  const scrollRef = useRef<HTMLDivElement>(null)

  const tagsQuery = useTags()
  const tag = useMemo(
    () => tagsQuery.data?.find((t) => t.id === tagId) ?? null,
    [tagsQuery.data, tagId],
  )

  const list = useInfinitePromptList(tagId !== null, {
    teamId,
    tagIds: tagId !== null ? [tagId] : undefined,
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
  const isPending = list.isPending || tagsQuery.isPending

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => navigate(-1)}
          aria-label="Назад"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="flex flex-1 items-center gap-2 min-w-0">
          <TagIcon
            className="h-4 w-4 shrink-0"
            style={{ color: tag?.color ?? "var(--color-muted-foreground)" }}
          />
          <h2 className="truncate text-sm font-semibold">
            {tag?.name ?? "Тег"}
          </h2>
          {tag && (
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
            title="С этим тегом пока нет промптов"
            description="Откройте промпт и добавьте к нему этот тег."
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
