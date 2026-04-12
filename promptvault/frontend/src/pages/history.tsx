import { useRef, useEffect, useCallback, useMemo } from "react"
import { useNavigate } from "react-router-dom"
import { Clock, FileText, Loader2 } from "lucide-react"

import { useHistory } from "@/hooks/use-history"
import { PageLayout } from "@/components/layout/page-layout"
import { useWorkspaceStore } from "@/stores/workspace-store"
import type { UsageLogEntry } from "@/api/types"

function groupByDay(items: UsageLogEntry[]) {
  const groups: { label: string; items: UsageLogEntry[] }[] = []
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const yesterday = new Date(today.getTime() - 86400000)
  const weekAgo = new Date(today.getTime() - 7 * 86400000)

  let currentLabel = ""
  let currentItems: UsageLogEntry[] = []

  for (const item of items) {
    const d = new Date(item.used_at)
    const itemDay = new Date(d.getFullYear(), d.getMonth(), d.getDate())

    let label: string
    if (itemDay.getTime() === today.getTime()) {
      label = "Сегодня"
    } else if (itemDay.getTime() === yesterday.getTime()) {
      label = "Вчера"
    } else if (itemDay.getTime() > weekAgo.getTime()) {
      label = d.toLocaleDateString("ru-RU", { weekday: "long" })
      label = label.charAt(0).toUpperCase() + label.slice(1)
    } else {
      label = d.toLocaleDateString("ru-RU", { day: "numeric", month: "long", year: "numeric" })
    }

    if (label !== currentLabel) {
      if (currentItems.length > 0) {
        groups.push({ label: currentLabel, items: currentItems })
      }
      currentLabel = label
      currentItems = [item]
    } else {
      currentItems.push(item)
    }
  }
  if (currentItems.length > 0) {
    groups.push({ label: currentLabel, items: currentItems })
  }

  return groups
}

export default function History() {
  const navigate = useNavigate()
  const team = useWorkspaceStore((s) => s.team)
  const teamId = team?.teamId ?? null

  const {
    data,
    isLoading,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
  } = useHistory(teamId)

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

  const allItems = useMemo(() => data?.pages.flatMap((p) => p.items) ?? [], [data])
  const groups = useMemo(() => groupByDay(allItems), [allItems])

  return (
    <PageLayout
      title="История"
      description="Хронология использования промптов"
    >
      {/* Content */}
      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : allItems.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-muted/30 ring-1 ring-border">
            <Clock className="h-7 w-7 text-muted-foreground/60" />
          </div>
          <p className="text-base font-medium text-muted-foreground">Пока нет истории</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Начните использовать промпты — они появятся здесь
          </p>
        </div>
      ) : (
        <div className="space-y-6">
          {groups.map((group) => (
            <div key={group.label}>
              <h2 className="mb-2 text-[0.75rem] font-semibold uppercase tracking-wider text-muted-foreground">
                {group.label}
              </h2>
              <div className="space-y-2">
                {group.items.map((entry) => (
                  <button
                    key={entry.id}
                    onClick={() => navigate(`/prompts/${entry.prompt_id}`)}
                    className="group flex w-full items-center gap-3 rounded-xl border border-border bg-card px-3.5 py-3 text-left transition-[transform,box-shadow] hover:-translate-y-0.5 hover:border-violet-500/15 hover:shadow-md"
                  >
                    <span className="shrink-0 text-[0.7rem] tabular-nums text-muted-foreground">
                      {new Date(entry.used_at).toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" })}
                    </span>
                    <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-violet-500/[0.08] ring-1 ring-violet-500/10">
                      <FileText className="h-3 w-3 text-violet-400" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-[0.8rem] font-medium text-foreground">
                        {entry.prompt.title}
                      </p>
                      {entry.prompt.tags.length > 0 && (
                        <div className="mt-0.5 flex gap-1">
                          {entry.prompt.tags.slice(0, 3).map((tag) => (
                            <span
                              key={tag.id}
                              className="rounded-full px-1.5 py-px text-[0.58rem]"
                              style={{ backgroundColor: tag.color + "18", color: tag.color + "cc" }}
                            >
                              {tag.name}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </button>
                ))}
              </div>
            </div>
          ))}

          {/* Sentinel */}
          <div ref={sentinelRef} className="flex justify-center py-4">
            {isFetchingNextPage && (
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            )}
            {!hasNextPage && allItems.length > 20 && (
              <p className="text-[0.75rem] text-muted-foreground">Вся история загружена</p>
            )}
          </div>
        </div>
      )}
    </PageLayout>
  )
}
