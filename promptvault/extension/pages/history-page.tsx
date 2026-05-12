import { useNavigate } from "react-router-dom"
import { ArrowLeft, History as HistoryIcon, Loader2 } from "lucide-react"
import { useInfiniteQuery } from "@tanstack/react-query"
import { Button } from "../components/ui/button"
import { sendBg } from "../lib/bg-client"
import { dateGroupLabel, formatTime } from "@pv/shared/utils/format-date"
import type { UsageHistoryItem } from "../lib/api"

const PAGE_SIZE = 50

export function HistoryPage() {
  const navigate = useNavigate()
  const query = useInfiniteQuery({
    queryKey: ["usage-history"],
    initialPageParam: 0,
    queryFn: ({ pageParam }) =>
      sendBg({ type: "api.listUsageHistory", limit: PAGE_SIZE, offset: pageParam as number }),
    getNextPageParam: (last, all) => {
      if (!last.has_more) return undefined
      return all.length * PAGE_SIZE
    },
    staleTime: 30_000,
  })

  if (query.isPending) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-(--color-muted-foreground)" />
      </div>
    )
  }

  const allItems = query.data?.pages.flatMap((p) => p.items) ?? []

  // Группировка по date label (Сегодня / Вчера / На неделе / ...)
  const grouped = new Map<string, UsageHistoryItem[]>()
  for (const item of allItems) {
    const label = dateGroupLabel(item.used_at)
    const arr = grouped.get(label) ?? []
    arr.push(item)
    grouped.set(label, arr)
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-(--color-border) p-2">
        <Button type="button" variant="ghost" size="icon" onClick={() => navigate(-1)} aria-label="Назад">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="flex-1 text-sm font-semibold">История использования</h2>
      </div>

      <div className="flex-1 overflow-y-auto p-3">
        {allItems.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-12 text-center">
            <HistoryIcon className="h-10 w-10 text-(--color-muted-foreground)/40" />
            <p className="text-sm font-medium">История пуста</p>
            <p className="max-w-xs text-[10px] text-(--color-muted-foreground)">
              Здесь появятся использованные промпты после вставки на AI-сайт.
            </p>
          </div>
        ) : (
          <div className="space-y-4">
            {Array.from(grouped.entries()).map(([label, items]) => (
              <section key={label}>
                <h3 className="mb-1.5 text-[10px] font-medium uppercase tracking-wide text-(--color-muted-foreground)">
                  {label}
                </h3>
                <ul className="space-y-0.5">
                  {items.map((item) => (
                    <li key={item.id}>
                      <button
                        type="button"
                        onClick={() => navigate(`/prompts/${item.prompt_id}`)}
                        className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs hover:bg-(--color-muted)/40"
                      >
                        <span className="w-10 shrink-0 font-mono text-[10px] text-(--color-muted-foreground)">
                          {formatTime(item.used_at)}
                        </span>
                        <span className="flex-1 truncate">
                          {item.prompt?.title ?? `Промпт #${item.prompt_id}`}
                        </span>
                        {item.prompt?.model && (
                          <span className="rounded bg-(--color-muted) px-1.5 py-0.5 text-[9px] text-(--color-muted-foreground)">
                            {item.prompt.model}
                          </span>
                        )}
                      </button>
                    </li>
                  ))}
                </ul>
              </section>
            ))}

            {query.hasNextPage && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => query.fetchNextPage()}
                disabled={query.isFetchingNextPage}
                className="w-full"
              >
                {query.isFetchingNextPage ? "Загружаю…" : "Показать ещё"}
              </Button>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
