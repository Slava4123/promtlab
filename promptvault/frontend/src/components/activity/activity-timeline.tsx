import { useEffect, useRef } from "react"
import { Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ActivityItem } from "./activity-item"
import { Separator } from "@/components/ui/separator"
import type { ActivityItem as ActivityItemData } from "@/api/activity"

interface ActivityTimelineProps {
  items: ActivityItemData[]
  hasMore: boolean
  isFetching: boolean
  onLoadMore: () => void
  // useIntersectionObserver: когда true, автоматически подгружает следующую страницу
  // при скролле к концу списка.
  autoLoad?: boolean
  // hasFilter=true → юзер применил фильтр; пустой результат означает «фильтр
  // ничего не нашёл», а не «в команде вообще нет активности». Empty state
  // должен отличать эти случаи, иначе юзер думает, что фильтр сломан.
  hasFilter?: boolean
  onClearFilter?: () => void
}

export function ActivityTimeline({
  items,
  hasMore,
  isFetching,
  onLoadMore,
  autoLoad = true,
  hasFilter = false,
  onClearFilter,
}: ActivityTimelineProps) {
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!autoLoad || !hasMore) return
    const sentinel = sentinelRef.current
    if (!sentinel) return
    const io = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && !isFetching) {
          onLoadMore()
        }
      },
      { rootMargin: "200px" },
    )
    io.observe(sentinel)
    return () => io.disconnect()
  }, [autoLoad, hasMore, isFetching, onLoadMore])

  if (items.length === 0 && !isFetching) {
    if (hasFilter) {
      return (
        <div className="py-12 text-center">
          <p className="text-base font-medium text-muted-foreground">Нет событий по выбранному фильтру</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Попробуйте выбрать другой тип события или сбросить фильтр.
          </p>
          {onClearFilter && (
            <Button variant="outline" size="sm" className="mt-4" onClick={onClearFilter}>
              Сбросить фильтр
            </Button>
          )}
        </div>
      )
    }
    return (
      <div className="py-12 text-center">
        <p className="text-base font-medium text-muted-foreground">В этой команде пока нет активности</p>
        <p className="mt-1 text-sm text-muted-foreground">
          Создайте промпт или пригласите участника, чтобы появились события.
        </p>
      </div>
    )
  }

  return (
    <div className="flex flex-col">
      {items.map((item, idx) => (
        <div key={item.id}>
          <ActivityItem item={item} />
          {idx < items.length - 1 && <Separator />}
        </div>
      ))}
      {hasMore && (
        <div ref={sentinelRef} className="flex justify-center py-4">
          {isFetching ? (
            <Loader2 className="size-5 animate-spin text-muted-foreground" />
          ) : (
            <Button variant="outline" size="sm" onClick={onLoadMore}>
              Загрузить ещё
            </Button>
          )}
        </div>
      )}
      {!hasMore && items.length > 0 && (
        <p className="py-4 text-center text-xs text-muted-foreground">Больше событий нет</p>
      )}
    </div>
  )
}
