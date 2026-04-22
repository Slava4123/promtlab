import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import type { Insight } from "@/api/analytics"
import { ApiError } from "@/api/client"
import { useRefreshInsights } from "@/hooks/use-analytics"
import { toast } from "sonner"
import { Loader2, RefreshCw, TrendingUp, TrendingDown, Archive, Copy, Hash, FolderOpen } from "lucide-react"

const ICONS: Record<string, typeof TrendingUp> = {
  unused_prompts: Archive,
  trending: TrendingUp,
  declining: TrendingDown,
  most_edited: Hash,
  possible_duplicates: Copy,
  orphan_tags: Hash,
  empty_collections: FolderOpen,
}

// Бэкенд всегда отдаёт unused_prompts/trending/declining; оставшиеся 4 типа
// скрыты за ANALYTICS_EXPERIMENTAL_INSIGHTS и не приходят до M8 — LABELS
// держим заранее, чтобы UI не падал когда бэк включит их.
const LABELS: Record<string, string> = {
  unused_prompts: "Забытые промпты",
  trending: "Популярные",
  declining: "Спадающие",
  most_edited: "Много правок",
  possible_duplicates: "Возможные дубликаты",
  orphan_tags: "Теги без промптов",
  empty_collections: "Пустые коллекции",
}

interface InsightsPanelProps {
  insights: Insight[]
}

export function InsightsPanel({ insights }: InsightsPanelProps) {
  const refresh = useRefreshInsights()

  async function handleRefresh() {
    try {
      await refresh.mutateAsync()
      toast.success("Инсайты обновлены")
    } catch (err) {
      if (err instanceof ApiError && err.status === 429) {
        toast.error("Инсайты можно обновлять не чаще одного раза в час")
      } else {
        toast.error("Не удалось обновить инсайты")
      }
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          Умные инсайты
          <Badge variant="outline">Max</Badge>
          <Button
            variant="ghost"
            size="sm"
            className="ml-auto"
            onClick={handleRefresh}
            disabled={refresh.isPending}
            aria-label="Обновить инсайты"
          >
            {refresh.isPending ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <RefreshCw className="size-4" />
            )}
            <span className="ml-1 text-xs">Обновить</span>
          </Button>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {insights.length === 0 ? (
          <div className="py-8 text-center">
            <p className="text-base font-medium text-muted-foreground">Пока нет инсайтов</p>
            <p className="mt-1 text-sm text-muted-foreground">
              Инсайты пересчитываются раз в сутки. Возвращайтесь завтра.
            </p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {insights.map((ins) => {
              const Icon = ICONS[ins.type] ?? Archive
              const label = LABELS[ins.type] ?? ins.type
              const items = Array.isArray(ins.payload) ? ins.payload : []
              return (
                <div key={ins.type} className="rounded-lg border p-3">
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <Icon className="size-4" />
                    {label}
                    <Badge variant="secondary" className="ml-auto">
                      {items.length}
                    </Badge>
                  </div>
                  {items.length > 0 && (
                    <ul className="mt-2 space-y-1 text-sm text-muted-foreground">
                      {items.slice(0, 3).map((it: { prompt_id?: number; title?: string }, idx) => (
                        <li key={idx} className="truncate">
                          {it.title ?? `#${it.prompt_id ?? idx}`}
                        </li>
                      ))}
                      {items.length > 3 && (
                        <li className="text-xs">…ещё {items.length - 3}</li>
                      )}
                    </ul>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
