import { useEffect, useRef } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { ArrowLeft, Activity, FileText, FolderOpen, Tag, Users, Loader2 } from "lucide-react"
import { useInfiniteQuery } from "@tanstack/react-query"
import { Button } from "../../components/ui/button"
import { sendBg } from "../../lib/bg-client"
import { formatRelativeDate } from "@pv/shared/utils/format-date"
import type { ActivityItem, ActivityResponse } from "../../lib/api"

// Маппинг event_type → иконка + цвет.
function eventIcon(eventType: string): React.ComponentType<{ className?: string }> {
  if (eventType.startsWith("prompt.")) return FileText
  if (eventType.startsWith("collection.")) return FolderOpen
  if (eventType.startsWith("tag.")) return Tag
  if (eventType.startsWith("member.") || eventType.startsWith("role.")) return Users
  return Activity
}

function eventLabel(eventType: string): string {
  const map: Record<string, string> = {
    "prompt.created": "создал промпт",
    "prompt.updated": "обновил промпт",
    "prompt.deleted": "удалил промпт",
    "collection.created": "создал коллекцию",
    "collection.updated": "обновил коллекцию",
    "collection.deleted": "удалил коллекцию",
    "tag.created": "создал тег",
    "tag.deleted": "удалил тег",
    "member.invited": "пригласил",
    "member.joined": "присоединился",
    "member.removed": "удалён из команды",
    "role.changed": "сменил роль",
    "share.created": "создал ссылку",
    "share.deactivated": "отозвал ссылку",
    "chain.created": "создал цепочку",
    "chain.updated": "обновил цепочку",
    "chain.deleted": "удалил цепочку",
    "chain.execution_started": "запустил цепочку",
    "chain.execution_completed": "завершил цепочку",
  }
  return map[eventType] ?? eventType
}

const PAGE_SIZE = 30

export function TeamActivityPage() {
  const { slug } = useParams<{ slug: string }>()
  const navigate = useNavigate()
  const scrollRef = useRef<HTMLDivElement>(null)

  const query = useInfiniteQuery<ActivityResponse>({
    queryKey: ["team-activity", slug],
    enabled: Boolean(slug),
    initialPageParam: 1,
    queryFn: ({ pageParam }) =>
      sendBg({
        type: "api.getTeamActivity",
        slug: slug ?? "",
        page: pageParam as number,
        pageSize: PAGE_SIZE,
      }),
    getNextPageParam: (last) => (last.has_more ? last.page + 1 : undefined),
    staleTime: 30_000,
  })

  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    const onScroll = () => {
      if (query.hasNextPage && !query.isFetchingNextPage) {
        const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 80
        if (nearBottom) void query.fetchNextPage()
      }
    }
    el.addEventListener("scroll", onScroll)
    return () => el.removeEventListener("scroll", onScroll)
  }, [query])

  const items: ActivityItem[] = query.data?.pages.flatMap((p) => p.items) ?? []

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">Активность команды</h2>
      </div>

      <div ref={scrollRef} className="flex-1 overflow-y-auto p-3">
        {query.isPending && items.length === 0 ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
          </div>
        ) : items.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-12 text-center">
            <Activity className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">Пока тишина</p>
            <p className="text-[10px] text-(--color-muted-foreground)">
              Создавайте промпты — события появятся здесь.
            </p>
          </div>
        ) : (
          <ul className="space-y-1.5">
            {items.map((item) => {
              const Icon = eventIcon(item.event_type)
              return (
                <li
                  key={item.id}
                  className="flex items-start gap-2 rounded-md border border-(--color-border) bg-(--color-card) p-2.5 text-xs"
                >
                  <Icon className="mt-0.5 h-3.5 w-3.5 shrink-0 text-(--color-muted-foreground)" />
                  <div className="flex-1 min-w-0 space-y-0.5">
                    <div>
                      <span className="font-medium">
                        {item.actor_name || item.actor_email || "—"}
                      </span>
                      <span className="text-(--color-muted-foreground)">
                        {" "}
                        {eventLabel(item.event_type)}
                      </span>
                      {item.target_label && (
                        <span> «{item.target_label}»</span>
                      )}
                    </div>
                    <div className="text-[10px] text-(--color-muted-foreground)">
                      {formatRelativeDate(item.created_at)}
                    </div>
                  </div>
                </li>
              )
            })}
            {query.isFetchingNextPage && (
              <li className="flex justify-center py-2">
                <Loader2 className="h-4 w-4 animate-spin text-(--color-muted-foreground)" />
              </li>
            )}
          </ul>
        )}
      </div>
    </div>
  )
}
